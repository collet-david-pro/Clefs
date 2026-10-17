package app

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"clefs/internal/db"
)

// setupApp ouvre une base vierge dans un dossier temporaire, qui devient le
// répertoire courant (documents/ y est créé). Pas de t.Parallel() : la couche
// db repose sur un singleton.
func setupApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	dbPath := filepath.Join(dir, "clefs.db")
	if err := db.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = db.CloseDB() })
	return New(dir, dbPath)
}

// seed crée un bâtiment, trois accès, deux clés et un détenteur :
// K-PASSE ouvre A et B, K-C ouvre C.
func seed(t *testing.T, a *App) (accA, accB, accC, keyPasse, keyC, alice int) {
	t.Helper()
	must := func(id int, err error) int {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	b := must(a.SaveBuilding(0, "Principal"))
	accA = must(a.SaveAccess(AccessInput{Name: "Salle A", BuildingID: b}))
	accB = must(a.SaveAccess(AccessInput{Name: "Salle B", BuildingID: b}))
	accC = must(a.SaveAccess(AccessInput{Name: "Salle C", BuildingID: b}))
	keyPasse = must(a.SaveKey(KeyInput{Number: "K-PASSE", Total: 1, AccessIDs: []int{accA, accB}}))
	keyC = must(a.SaveKey(KeyInput{Number: "K-C", Total: 2, AccessIDs: []int{accC}}))
	alice = must(a.SaveBorrower(BorrowerInput{Name: "Alice", Status: "contractuel"}))
	return
}

// TestValidateKeyInput vérifie les règles de saisie reprises de la V3.
func TestValidateKeyInput(t *testing.T) {
	cases := []struct {
		in KeyInput
		ok bool
	}{
		{KeyInput{Number: "K1", Total: 1}, true},
		{KeyInput{Number: " ", Total: 1}, false},
		{KeyInput{Number: "K1", Total: 0}, true}, // toutes perdues ou cassées
		{KeyInput{Number: "K1", Total: -1}, false},
		{KeyInput{Number: "K1", Total: 2, Reserve: -1}, false},
		{KeyInput{Number: "K1", Total: 2, Reserve: 3}, false},
		{KeyInput{Number: "K1", Total: 2, Reserve: 2}, true},
	}
	for _, c := range cases {
		if err := validateKeyInput(c.in); (err == nil) != c.ok {
			t.Errorf("validateKeyInput(%+v) = %v, attendu ok=%v", c.in, err, c.ok)
		}
	}
}

// TestDuplicateKeyNumber vérifie le message lisible sur numéro en double.
func TestDuplicateKeyNumber(t *testing.T) {
	a := setupApp(t)
	if _, err := a.SaveKey(KeyInput{Number: "K1", Total: 1}); err != nil {
		t.Fatal(err)
	}
	_, err := a.SaveKey(KeyInput{Number: "K1", Total: 1})
	if err == nil || !strings.Contains(err.Error(), "déjà") {
		t.Errorf("doublon: %v", err)
	}
}

// TestSuggestKeys vérifie le trousseau minimal, l'exclusion et l'ajout manuel.
func TestSuggestKeys(t *testing.T) {
	a := setupApp(t)
	accA, accB, accC, keyPasse, keyC, alice := seed(t, a)

	s, err := a.SuggestKeys(SuggestRequest{BorrowerID: alice, AccessIDs: []int{accA, accB, accC}})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Keys) != 2 || len(s.Uncoverable) != 0 {
		t.Fatalf("trousseau attendu de 2 clés: %+v", s)
	}

	// K-PASSE retirée à la main : A et B deviennent non couvrables.
	s, _ = a.SuggestKeys(SuggestRequest{BorrowerID: alice, AccessIDs: []int{accA, accB, accC}, ExcludedKeyIDs: []int{keyPasse}})
	if len(s.Keys) != 1 || s.Keys[0].Key.ID != keyC || len(s.Uncoverable) != 2 {
		t.Errorf("exclusion: %+v", s)
	}

	// K-PASSE ajoutée à la main pour C seul : elle est conservée.
	s, _ = a.SuggestKeys(SuggestRequest{BorrowerID: alice, AccessIDs: []int{accC}, ManualKeyIDs: []int{keyPasse}})
	if len(s.Keys) != 2 {
		t.Errorf("ajout manuel: %+v", s)
	}
}

// TestCreateLoanAndReturn vérifie un prêt complet : enregistrement des
// options, bon PDF, sur-prêt signalé sans blocage, retour et réédition.
func TestCreateLoanAndReturn(t *testing.T) {
	a := setupApp(t)
	_, _, _, keyPasse, _, alice := seed(t, a)
	bob, _ := a.SaveBorrower(BorrowerInput{Name: "Bob / Entreprise"})

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	res, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyPasse}, PlannedReturn: tomorrow, LoanType: "permanent"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 1 || res.ReceiptError != "" || len(res.Anomalies) != 0 {
		t.Fatalf("résultat: %+v", res)
	}
	if _, err := os.Stat(res.ReceiptPath); err != nil {
		t.Errorf("bon de remise absent: %v", err)
	}

	loans, _ := a.ListActiveLoans()
	if len(loans) != 1 || loans[0].LoanType != "permanent" || loans[0].PlannedReturnDate == "" || loans[0].Status != "active" || loans[0].CreatedBy == "" {
		t.Errorf("prêt enregistré: %+v", loans)
	}

	// Deuxième prêt du seul exemplaire : accepté mais signalé.
	res, err = a.CreateLoan(CreateLoanRequest{BorrowerID: bob, KeyIDs: []int{keyPasse}})
	if err != nil {
		t.Fatalf("le sur-prêt doit être accepté: %v", err)
	}
	if len(res.Anomalies) != 1 || res.Anomalies[0].Available != -1 {
		t.Errorf("sur-prêt non signalé: %+v", res.Anomalies)
	}
	if !strings.Contains(filepath.Base(res.ReceiptPath), "Bob_-_Entreprise") {
		t.Errorf("nom de fichier non assaini: %s", res.ReceiptPath)
	}

	if _, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyPasse}, PlannedReturn: "2000-01-01"}); err == nil {
		t.Error("une date de retour passée doit être refusée")
	}

	if err := a.DeleteKey(keyPasse); err == nil {
		t.Error("une clé prêtée ne doit pas pouvoir être supprimée")
	}
	if err := a.DeleteBorrower(alice); err == nil {
		t.Error("un détenteur qui a des clés ne doit pas pouvoir être supprimé")
	}

	path, err := a.ReissueReceipt(alice)
	if err != nil || path == "" {
		t.Errorf("réédition: %v", err)
	}

	loans, _ = a.ListActiveLoans()
	ids := []int{}
	for _, l := range loans {
		ids = append(ids, l.ID)
	}
	if err := a.ReturnLoans(ids, "bon état"); err != nil {
		t.Fatal(err)
	}
	hist, err := a.GetHistory(HistoryFilter{Status: "returned"})
	if err != nil || len(hist) != 2 || hist[0].Condition != "bon état" {
		t.Errorf("historique après retour: %+v, %v", hist, err)
	}
	if err := a.DeleteBorrower(alice); err != nil {
		t.Errorf("suppression après retour: %v", err)
	}
}

// TestDashboard vérifie les indicateurs principaux.
func TestDashboard(t *testing.T) {
	a := setupApp(t)
	_, _, _, _, keyC, alice := seed(t, a)
	if _, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyC}}); err != nil {
		t.Fatal(err)
	}
	d, err := a.GetDashboard()
	if err != nil {
		t.Fatal(err)
	}
	if d.KeyTypes != 2 || d.KeyCopies != 3 || d.LoanedCopies != 1 || d.AvailableCopies != 2 || d.BorrowersWithKeys != 1 || d.Accesses != 3 {
		t.Errorf("tableau de bord: %+v", d)
	}
	if len(d.RecentLoans) != 1 {
		t.Errorf("prêts récents: %+v", d.RecentLoans)
	}
}

// TestParseBorrowersCSV vérifie la lecture du format de nos exports.
func TestParseBorrowersCSV(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("Nom;Email;Statut;Téléphone\n"+
		"Dupont Jean;jean@x.fr;permanent;0102\n"+
		";;;\n"+
		";sans@nom.fr;;\n"+
		"Martin;;inconnu;\n"+
		"Durand;;Contractuel;\n")...)
	rows, problems, err := parseBorrowersCSV(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1].Status != "contractuel" {
		t.Errorf("lignes valides: %+v", rows)
	}
	if len(problems) != 2 {
		t.Errorf("lignes rejetées: %v", problems)
	}
}

// TestImportWizardLegacyNeedsEmptyBase vérifie qu'une base V1/V2 n'est pas
// mélangée à des données existantes.
func TestImportWizardLegacyNeedsEmptyBase(t *testing.T) {
	a := setupApp(t)
	if err := a.LoadDemoData(); err != nil {
		t.Fatal(err)
	}
	if err := a.LoadDemoData(); err == nil {
		t.Error("la démo ne doit se charger que dans une base vide")
	}
	info, err := a.InspectDatabaseFile(a.dbPath)
	if err == nil {
		t.Errorf("la base ouverte ne doit pas pouvoir s'importer elle-même: %+v", info)
	}
}

// TestSafeFileName vérifie l'assainissement des noms de fichiers.
func TestSafeFileName(t *testing.T) {
	if got := safeFileName(` Dupont: "Jean" / Paul? `); got != "Dupont-_Jean_-_Paul" {
		t.Errorf("safeFileName = %q", got)
	}
	if got := safeFileName("  "); got != "sans_nom" {
		t.Errorf("safeFileName vide = %q", got)
	}
}

// TestApplyInventory vérifie l'inventaire de bout en bout : contrôles de
// saisie, mise à jour du stock et rapport PDF.
func TestApplyInventory(t *testing.T) {
	a := setupApp(t)
	_, _, _, keyPasse, keyC, alice := seed(t, a)
	if _, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyC}}); err != nil {
		t.Fatal(err)
	}

	if _, err := a.ApplyInventory(nil); err == nil {
		t.Error("un inventaire vide doit être refusé")
	}
	if _, err := a.ApplyInventory([]InventoryItem{{KeyID: keyC, InStock: -1}}); err == nil {
		t.Error("un stock négatif doit être refusé")
	}
	if _, err := a.ApplyInventory([]InventoryItem{{KeyID: keyC, InStock: 1}, {KeyID: keyC, InStock: 2}}); err == nil {
		t.Error("une clé comptée deux fois doit être refusée")
	}

	// K-C : 2 au total, 1 prêtée ; on en compte 3 en stock → total 4.
	// K-PASSE : 1 au total, aucune prêtée ; cassée → 0 en stock, total 0.
	res, err := a.ApplyInventory([]InventoryItem{{KeyID: keyC, InStock: 3}, {KeyID: keyPasse, InStock: 0}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Counted != 2 || res.Changed != 2 || res.Delta != 1 || res.ReportError != "" {
		t.Errorf("compte rendu: %+v", res)
	}
	if _, err := os.Stat(res.ReportPath); err != nil {
		t.Errorf("rapport d'inventaire absent: %v", err)
	}
	keys, _ := a.ListKeys()
	for _, k := range keys {
		if k.ID == keyC && (k.Total != 4 || k.Available != 3) {
			t.Errorf("K-C après inventaire: %+v", k)
		}
		if k.ID == keyPasse && (k.Total != 0 || k.Available != 0) {
			t.Errorf("K-PASSE après inventaire: %+v", k)
		}
	}

	// Inventaire sans écart : rapport quand même produit.
	res, err = a.ApplyInventory([]InventoryItem{{KeyID: keyC, InStock: 3}})
	if err != nil || res.Changed != 0 || res.ReportPath == "" {
		t.Errorf("inventaire sans écart: %+v, %v", res, err)
	}
}

// TestPDFExports vérifie la génération des documents depuis la base.
func TestPDFExports(t *testing.T) {
	a := setupApp(t)
	_, _, _, _, keyC, alice := seed(t, a)
	if _, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyC}}); err != nil {
		t.Fatal(err)
	}
	for name, export := range map[string]func() (string, error){
		"bilan": a.ExportKeyStockPDF, "plan": a.ExportKeyPlanPDF,
		"réédition": func() (string, error) { return a.ReissueReceipt(alice) },
		"archive":   a.ExportAllReceipts,
	} {
		path, err := export()
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if info, err := os.Stat(path); err != nil || info.Size() < 1000 {
			t.Errorf("%s: fichier absent ou vide (%v)", name, err)
		}
	}
}

// TestSettings vérifie les paramètres : valeurs d'origine, enregistrement,
// contrôles, logo, agent, et conservation après réinitialisation.
func TestSettings(t *testing.T) {
	a := setupApp(t)
	s, err := a.GetSettings()
	if err != nil || s.Establishment != "Collège Victor Hugo" || s.DefaultLoanType != "ponctuel" {
		t.Fatalf("valeurs d'origine: %+v, %v", s, err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 40, 20))); err != nil {
		t.Fatal(err)
	}
	s.Establishment = "  Lycée Jean Moulin  "
	s.Address = "Paris"
	s.AppName = "" // vide → valeur d'origine
	s.LogoDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	s.DefaultLoanType = "permanent"
	s.DefaultReturnDays = 30
	s.AgentName = "Secrétariat"
	saved, err := a.SaveSettings(s)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Establishment != "Lycée Jean Moulin" || saved.AppName != "Gestionnaire de Clés" {
		t.Errorf("nettoyage: %+v", saved)
	}
	got, _ := a.GetSettings()
	if got.Establishment != "Lycée Jean Moulin" || got.LogoDataURL == "" || got.DefaultReturnDays != 30 || got.DefaultLoanType != "permanent" {
		t.Errorf("relecture: %+v", got)
	}

	for name, bad := range map[string]Settings{
		"logo":  {LogoDataURL: "data:image/png;base64,AAAA"},
		"durée": {DefaultReturnDays: 400},
		"long":  {Establishment: strings.Repeat("x", 200)},
	} {
		if _, err := a.SaveSettings(bad); err == nil {
			t.Errorf("%s : valeur invalide acceptée", name)
		}
	}

	// L'agent des Paramètres est enregistré sur les prêts.
	_, _, _, _, keyC, alice := seed(t, a)
	if _, err := a.CreateLoan(CreateLoanRequest{BorrowerID: alice, KeyIDs: []int{keyC}}); err != nil {
		t.Fatal(err)
	}
	loans, _ := a.ListActiveLoans()
	if loans[0].CreatedBy != "Secrétariat" {
		t.Errorf("agent: %q", loans[0].CreatedBy)
	}

	if path, err := a.PreviewReceipt(got); err != nil || path == "" {
		t.Errorf("aperçu du bon: %v", err)
	}

	// Les paramètres survivent à la réinitialisation de la base.
	if err := a.ResetDatabase(); err != nil {
		t.Fatal(err)
	}
	if after, _ := a.GetSettings(); after.Establishment != "Lycée Jean Moulin" {
		t.Errorf("paramètres perdus à la réinitialisation: %+v", after)
	}
}
