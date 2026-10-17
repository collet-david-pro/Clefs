package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	imgjpeg "image/jpeg"
	imgpng "image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// check vérifie qu'un document est un PDF valide et, si la variable
// PDF_PREVIEW_DIR est définie, l'y enregistre pour relecture visuelle.
func check(t *testing.T, name string, data []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || len(data) < 1000 {
		t.Fatalf("%s: sortie invalide (%d octets)", name, len(data))
	}
	if dir := os.Getenv("PDF_PREVIEW_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, name+".pdf"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// sampleReceipt fabrique un bon de n clés aux textes longs et accentués.
func sampleReceipt(name string, n int) Receipt {
	planned := time.Now().AddDate(0, 0, 30)
	r := Receipt{
		BorrowerName: name, BorrowerStatus: "contractuel", Date: time.Now(),
		LoanType: "permanent", PlannedReturn: &planned, Agent: "secretariat",
	}
	for i := 1; i <= n; i++ {
		r.Keys = append(r.Keys, ReceiptKey{
			Number: fmt.Sprintf("K-%03d", i), Description: "Passe partiel du bâtiment B — salles de sciences, laboratoire et réserve « produits »",
			Storage: "Armoire à clés n°2, crochet " + fmt.Sprint(i), LoanDate: time.Now(),
		})
	}
	return r
}

// TestReceiptFitsOnePage vérifie qu'un bon tient sur une seule page A4,
// quel que soit le nombre de clés (densité adaptée automatiquement).
func TestReceiptFitsOnePage(t *testing.T) {
	for _, n := range []int{1, 13, 30, 65} {
		r := sampleReceipt("Hélène Dupont-Lefèvre", n)
		if n == 13 {
			r.Keys[0].Number = "Clé Gymnase portail principal"
			r.Keys[1].Description = "Descriptionsansaucunespacequidépasselargementlalargeurdelacolonneprévue"
		}
		d := newBlankDoc(DefaultSettings(), "Bon de remise de clés")
		d.receipt(r, "")
		if d.PageNo() != 1 {
			t.Errorf("bon de %d clés : %d pages, attendu 1", n, d.PageNo())
		}
		data, err := d.output()
		check(t, fmt.Sprintf("bon_remise_%d_cles", n), data, err)
	}

	r := sampleReceipt("Bob", 3)
	r.Reissue, r.PlannedReturn = true, nil
	data, err := GenerateReceipt(DefaultSettings(), r)
	check(t, "bon_remise_reedite", data, err)

	if _, err := GenerateReceipt(DefaultSettings(), Receipt{BorrowerName: "Vide"}); err == nil {
		t.Error("un bon sans clé doit être refusé")
	}
}

// TestGenerateReceiptArchive vérifie l'archive : sommaire puis une page par bon.
func TestGenerateReceiptArchive(t *testing.T) {
	var receipts []Receipt
	for i := 1; i <= 40; i++ {
		r := sampleReceipt(fmt.Sprintf("Détenteur n°%d", i), 1+i%7)
		r.Reissue = true
		receipts = append(receipts, r)
	}
	d, err := receiptArchive(DefaultSettings(), receipts)
	if err != nil {
		t.Fatal(err)
	}
	// 40 détenteurs : sommaire sur 2 pages, puis un bon par page.
	if d.PageNo() != 42 {
		t.Errorf("archive de 40 bons : %d pages, attendu 42", d.PageNo())
	}
	data, err := d.output()
	check(t, "archive_bons", data, err)
	if _, err := GenerateReceiptArchive(DefaultSettings(), nil); err == nil {
		t.Error("une archive vide doit être refusée")
	}
}

// TestGenerateKeyPlan couvre plusieurs bâtiments, un bâtiment vide et des accès sans clé.
func TestGenerateKeyPlan(t *testing.T) {
	var buildings []PlanBuilding
	for b := 1; b <= 4; b++ {
		pb := PlanBuilding{Name: fmt.Sprintf("Bâtiment %c", 'A'+b-1)}
		for a := 1; a <= 18; a++ {
			pa := PlanAccess{Name: fmt.Sprintf("Salle %d%02d", b, a), Floor: "RDC", Category: "Salle de classe"}
			if a%5 != 0 {
				pa.Keys = []string{fmt.Sprintf("K%d", a), "PASSE-GÉNÉRAL", fmt.Sprintf("B%d", b)}
			}
			pb.Accesses = append(pb.Accesses, pa)
		}
		buildings = append(buildings, pb)
	}
	buildings = append(buildings, PlanBuilding{Name: "Gymnase"})
	data, err := GenerateKeyPlan(DefaultSettings(), buildings)
	check(t, "plan_cles", data, err)
}

// TestGenerateStockReport couvre les trois états de disponibilité.
func TestGenerateStockReport(t *testing.T) {
	rows := []StockRow{
		{Number: "K-01", Description: "Clé de la loge", Storage: "Loge", Total: 3, Reserve: 1, Loaned: 1, Available: 1},
		{Number: "K-02", Description: "Clé épuisée", Total: 2, Reserve: 0, Loaned: 2, Available: 0},
		{Number: "K-03", Description: "Clé en sur-prêt", Total: 1, Reserve: 0, Loaned: 3, Available: -2},
	}
	for i := 4; i <= 60; i++ {
		rows = append(rows, StockRow{Number: fmt.Sprintf("K-%02d", i), Description: "Clé de salle", Total: 4, Loaned: 1, Available: 3})
	}
	data, err := GenerateStockReport(DefaultSettings(), rows)
	check(t, "bilan_stock", data, err)
}

// TestGenerateInventoryReport couvre un inventaire avec et sans écart.
func TestGenerateInventoryReport(t *testing.T) {
	data, err := GenerateInventoryReport(DefaultSettings(), InventoryReport{Agent: "agent", Counted: 3, Changes: []InventoryRow{
		{Number: "K-01", Description: "Clé retrouvée", Loaned: 1, Counted: 3, OldTotal: 2, NewTotal: 4, OldReserve: 1, NewReserve: 1},
		{Number: "K-02", Description: "Clé cassée", Loaned: 0, Counted: 0, OldTotal: 2, NewTotal: 0, OldReserve: 1, NewReserve: 0},
	}})
	check(t, "inventaire", data, err)

	data, err = GenerateInventoryReport(DefaultSettings(), InventoryReport{Agent: "agent", Counted: 5})
	check(t, "inventaire_sans_ecart", data, err)
}

// TestCapitalizeAndSigned vérifie les petits formatages.
func TestCapitalizeAndSigned(t *testing.T) {
	if capitalize("écart") != "Écart" || capitalize("") != "" || capitalize("a") != "A" {
		t.Error("capitalize")
	}
	if signed(2) != "+2" || signed(-3) != "−3" || signed(0) != "0" {
		t.Error("signed")
	}
}

// TestSettings vérifie les réglages : texte d'engagement personnalisé,
// valeurs par défaut des champs vides, logo PNG/JPEG et logo invalide.
func TestSettings(t *testing.T) {
	s := Settings{Engagement: "{detenteur} reçoit {cles} et s'engage à {la_les} rendre."}.withDefaults()
	if got := s.engagementText("Alice", 1); got != "Alice reçoit la clé mentionnée ci-dessus et s'engage à la rendre." {
		t.Errorf("engagement (1 clé) : %q", got)
	}
	if got := s.engagementText("Bob", 3); got != "Bob reçoit les 3 clés mentionnées ci-dessus et s'engage à les rendre." {
		t.Errorf("engagement (3 clés) : %q", got)
	}
	if s.Establishment != "Collège Victor Hugo" || s.ReceiptTitle == "" {
		t.Errorf("valeurs par défaut non appliquées : %+v", s)
	}

	var png, jpg bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 300, 100)) // logo large, ratio 3:1
	for x := 0; x < 300; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{79, 70, 229, 255})
		}
	}
	if err := imgpng.Encode(&png, img); err != nil {
		t.Fatal(err)
	}
	if err := imgjpeg.Encode(&jpg, img, nil); err != nil {
		t.Fatal(err)
	}
	if f, _, _, err := ValidateLogo(png.Bytes()); err != nil || f != "PNG" {
		t.Errorf("logo PNG : %s, %v", f, err)
	}
	if f, _, _, err := ValidateLogo(jpg.Bytes()); err != nil || f != "JPG" {
		t.Errorf("logo JPEG : %s, %v", f, err)
	}
	if _, _, _, err := ValidateLogo([]byte("pas une image")); err == nil {
		t.Error("un logo invalide doit être refusé")
	}

	custom := Settings{
		AppName: "Gestion des clés", Establishment: "Lycée Jean Moulin", Address: "12 rue de la Paix, 75000 Paris",
		Logo: png.Bytes(), FooterNote: "Document à conserver", ReceiptTitle: "Fiche de prêt de clés",
		SignatureEstablishment: "Le gestionnaire", Signatory: "M. Martin, gestionnaire",
	}
	data, err := GenerateReceipt(custom, sampleReceipt("Hélène Dupont", 4))
	check(t, "bon_remise_personnalise", data, err)

	// Logo illisible : ignoré, le document est quand même produit.
	custom.Logo = []byte("corrompu")
	data, err = GenerateStockReport(custom, []StockRow{{Number: "K-1", Total: 1, Available: 1}})
	check(t, "bilan_logo_invalide", data, err)
}
