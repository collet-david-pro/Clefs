package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createV3SourceDatabase fabrique une base 3.x complète (schéma courant, en
// mode WAL) avec un prêt actif, un prêt rendu et une clé en sur-prêt, puis la
// referme. Doit être appelé AVANT setupTestDB (singleton DB).
func createV3SourceDatabase(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "clefs_v3.db")
	if err := InitDB(path); err != nil {
		t.Fatalf("InitDB source: %v", err)
	}
	b := &Building{Name: "Bâtiment A"}
	if err := CreateBuilding(b); err != nil {
		t.Fatalf("CreateBuilding: %v", err)
	}
	r := &Room{Name: "Salle 101", BuildingID: b.ID, Floor: "RDC"}
	if err := CreateRoom(r); err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	k := &Key{Number: "K-01", QuantityTotal: 1, Category: "passe"}
	if err := CreateKey(k, []int{r.ID}); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	alice := &Borrower{Name: "Alice", Status: "contractuel"}
	bob := &Borrower{Name: "Bob"}
	for _, br := range []*Borrower{alice, bob} {
		if err := CreateBorrower(br); err != nil {
			t.Fatalf("CreateBorrower: %v", err)
		}
	}
	// Deux prêts actifs pour un exemplaire : sur-prêt volontaire.
	ids, err := CreateMultipleLoans([]int{k.ID}, alice.ID, LoanOptions{})
	if err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}
	if _, err := CreateMultipleLoans([]int{k.ID}, bob.ID, LoanOptions{}); err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}
	// Un troisième prêt, rendu.
	returned, err := CreateMultipleLoans([]int{k.ID}, bob.ID, LoanOptions{})
	if err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}
	if err := ReturnLoan(returned[0]); err != nil {
		t.Fatalf("ReturnLoan: %v", err)
	}
	_ = ids
	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB source: %v", err)
	}
	return path
}

// TestInspectDatabase vérifie la détection des trois cas : base 3.x, base
// ancienne (V1/V2) et fichier quelconque.
func TestInspectDatabase(t *testing.T) {
	v3Path := createV3SourceDatabase(t)
	setupTestDB(t)

	info, err := InspectDatabase(v3Path)
	if err != nil {
		t.Fatalf("InspectDatabase(v3): %v", err)
	}
	if info.Kind != DatabaseKindV3 {
		t.Errorf("base 3.x détectée comme %q", info.Kind)
	}
	if info.Keys != 1 || info.Borrowers != 2 || info.Loans != 3 || info.ActiveLoans != 2 || info.Rooms != 1 || info.Buildings != 1 {
		t.Errorf("aperçu inattendu: %+v", info)
	}

	legacy, err := InspectDatabase(createFakeV2Database(t))
	if err != nil {
		t.Fatalf("InspectDatabase(v2): %v", err)
	}
	if legacy.Kind != DatabaseKindLegacy {
		t.Errorf("base V2 détectée comme %q", legacy.Kind)
	}
	if legacy.ActiveLoans != 1 || legacy.Loans != 2 {
		t.Errorf("aperçu V2 inattendu: %+v", legacy)
	}

	text := filepath.Join(t.TempDir(), "notes.db")
	if err := os.WriteFile(text, []byte("ceci n'est pas une base"), 0644); err != nil {
		t.Fatal(err)
	}
	other, err := InspectDatabase(text)
	if err != nil {
		t.Fatalf("InspectDatabase(texte): %v", err)
	}
	if other.Kind != DatabaseKindInvalid {
		t.Errorf("fichier texte détecté comme %q", other.Kind)
	}

	if _, err := InspectDatabase(filepath.Join(t.TempDir(), "absent.db")); err == nil {
		t.Error("un fichier absent doit retourner une erreur")
	}
}

// TestImportFromV3 vérifie qu'une base 3.x est reprise sans perte dans une
// nouvelle installation, sur-prêt compris, et que la source n'est pas modifiée.
func TestImportFromV3(t *testing.T) {
	v3Path := createV3SourceDatabase(t)
	before, err := os.ReadFile(v3Path)
	if err != nil {
		t.Fatal(err)
	}
	dbPath := setupTestDB(t)

	empty, err := IsDatabaseEmpty()
	if err != nil || !empty {
		t.Fatalf("nouvelle base: IsDatabaseEmpty = %v, %v", empty, err)
	}

	summary, err := ImportFromV3(v3Path, dbPath)
	if err != nil {
		t.Fatalf("ImportFromV3: %v", err)
	}
	if summary.Keys != 1 || summary.Borrowers != 2 || summary.Loans != 3 || summary.Associations != 1 {
		t.Errorf("résumé inattendu: %+v", summary)
	}

	empty, err = IsDatabaseEmpty()
	if err != nil || empty {
		t.Errorf("après import: IsDatabaseEmpty = %v, %v", empty, err)
	}
	anomalies, err := CheckInventoryAnomalies()
	if err != nil {
		t.Fatal(err)
	}
	if len(anomalies) != 1 || anomalies[0].Available != -1 {
		t.Errorf("le sur-prêt doit être conservé à l'import: %+v", anomalies)
	}
	borrowers, err := GetAllBorrowers()
	if err != nil {
		t.Fatal(err)
	}
	if borrowers[0].Name != "Alice" || borrowers[0].Status != "contractuel" {
		t.Errorf("détenteur mal repris: %+v", borrowers[0])
	}
	rooms, err := GetAllAccesses()
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 || rooms[0].Floor != "RDC" {
		t.Errorf("accès mal repris: %+v", rooms)
	}

	after, err := os.ReadFile(v3Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("la base source a été modifiée par l'import")
	}

	// Pas de copie intermédiaire laissée dans backups/.
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(dbPath), "backups", "import_source_*"))
	if len(matches) != 0 {
		t.Errorf("copie intermédiaire non supprimée: %v", matches)
	}
}

// TestImportFromV3RejectsLegacy vérifie qu'une base V2 n'est pas restaurée
// telle quelle (elle doit passer par ImportFromV2).
func TestImportFromV3RejectsLegacy(t *testing.T) {
	dbPath := setupTestDB(t)
	if _, err := ImportFromV3(createFakeV2Database(t), dbPath); err == nil {
		t.Error("une base V2 ne doit pas être acceptée par ImportFromV3")
	}
}

// TestBackupPathWithApostrophe vérifie qu'une sauvegarde fonctionne dans un
// dossier dont le nom contient une apostrophe.
func TestBackupPathWithApostrophe(t *testing.T) {
	dbPath := setupTestDB(t)
	mustCreateKey(t, "K-01", 1, 0)
	backup := filepath.Join(t.TempDir(), "dossier d'essai", "sauvegarde.db")
	if err := BackupDatabase(dbPath, backup); err != nil {
		t.Fatalf("BackupDatabase: %v", err)
	}
	if _, err := os.Stat(backup); err != nil {
		t.Errorf("sauvegarde absente: %v", err)
	}
}

// TestV4FieldsPersisted vérifie que les champs que la V3 affichait sans les
// enregistrer sont désormais conservés : statut et téléphone du détenteur,
// catégorie et notes de la clé, retour prévu, type et agent du prêt.
func TestV4FieldsPersisted(t *testing.T) {
	setupTestDB(t)

	b := &Borrower{Name: "Alice", Status: "intervenant", Phone: "0102030405"}
	if err := CreateBorrower(b); err != nil {
		t.Fatal(err)
	}
	got, err := GetBorrowerByID(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "intervenant" || got.Phone != "0102030405" {
		t.Errorf("création détenteur: %+v", got)
	}
	b.Status, b.Phone = "entreprise", "0607080910"
	if err := UpdateBorrower(b); err != nil {
		t.Fatal(err)
	}
	got, _ = GetBorrowerByID(b.ID)
	if got.Status != "entreprise" || got.Phone != "0607080910" {
		t.Errorf("modification détenteur: %+v", got)
	}

	k := &Key{Number: "K-01", QuantityTotal: 2, Category: "badge", Notes: "bleu"}
	if err := CreateKey(k, nil); err != nil {
		t.Fatal(err)
	}
	gk, _ := GetKeyByID(k.ID)
	if gk.Category != "badge" || gk.Notes != "bleu" {
		t.Errorf("création clé: %+v", gk)
	}
	k.Category, k.Notes = "", "rouge"
	if err := UpdateKey(k, nil); err != nil {
		t.Fatal(err)
	}
	gk, _ = GetKeyByID(k.ID)
	if gk.Category != "simple" || gk.Notes != "rouge" {
		t.Errorf("modification clé (catégorie vide → simple): %+v", gk)
	}

	past := time.Now().AddDate(0, 0, -3)
	if _, err := CreateMultipleLoans([]int{k.ID}, b.ID, LoanOptions{
		PlannedReturnDate: &past, LoanType: "permanent", CreatedBy: "agent",
	}); err != nil {
		t.Fatal(err)
	}
	loans, err := GetActiveLoansByBorrowerID(b.ID)
	if err != nil || len(loans) != 1 {
		t.Fatalf("prêts actifs: %v, %v", loans, err)
	}
	l := loans[0]
	if l.PlannedReturnDate == nil || l.PlannedReturnDate.Format("2006-01-02") != past.Format("2006-01-02") {
		t.Errorf("retour prévu non enregistré: %v", l.PlannedReturnDate)
	}
	if l.LoanType != "permanent" || l.CreatedBy != "agent" {
		t.Errorf("type/agent non enregistrés: %+v", l.Loan)
	}
	overdue, err := GetOverdueLoans()
	if err != nil || len(overdue) != 1 {
		t.Errorf("le prêt échu doit apparaître en retard: %v, %v", overdue, err)
	}

	// Sans options : type « ponctuel » par défaut, pas de retour prévu.
	if _, err := CreateMultipleLoans([]int{k.ID}, b.ID, LoanOptions{}); err != nil {
		t.Fatal(err)
	}
	loans, _ = GetActiveLoansByBorrowerID(b.ID)
	if loans[1].LoanType != "ponctuel" || loans[1].PlannedReturnDate != nil {
		t.Errorf("valeurs par défaut: %+v", loans[1].Loan)
	}
}
