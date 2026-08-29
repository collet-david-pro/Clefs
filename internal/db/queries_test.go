package db

import "testing"

// TestLoanLifecycle vérifie le cycle complet : création de prêt, décompte de
// disponibilité, puis retour.
func TestLoanLifecycle(t *testing.T) {
	setupTestDB(t)
	k := mustCreateKey(t, "K-01", 3, 1)
	b := mustCreateBorrower(t, "Alice")

	if _, err := CreateMultipleLoans([]int{k.ID}, b.ID); err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}

	count, err := GetActiveLoanCount(k.ID)
	if err != nil {
		t.Fatalf("GetActiveLoanCount: %v", err)
	}
	if count != 1 {
		t.Errorf("prêts actifs = %d, attendu 1", count)
	}

	// 3 au total - 1 en réserve - 1 prêtée = 1 disponible
	kwa := availabilityOf(t, k.ID)
	if kwa.AvailableCount != 1 {
		t.Errorf("disponible = %d, attendu 1", kwa.AvailableCount)
	}
	if kwa.LoanedCount != 1 {
		t.Errorf("prêtées = %d, attendu 1", kwa.LoanedCount)
	}

	loans, err := GetActiveLoansByBorrowerID(b.ID)
	if err != nil {
		t.Fatalf("GetActiveLoansByBorrowerID: %v", err)
	}
	if len(loans) != 1 {
		t.Fatalf("prêts du détenteur = %d, attendu 1", len(loans))
	}
	if loans[0].KeyNumber != "K-01" || loans[0].BorrowerName != "Alice" {
		t.Errorf("détails du prêt inattendus : clé %q, détenteur %q", loans[0].KeyNumber, loans[0].BorrowerName)
	}

	if err := ReturnLoan(loans[0].Loan.ID); err != nil {
		t.Fatalf("ReturnLoan: %v", err)
	}
	count, _ = GetActiveLoanCount(k.ID)
	if count != 0 {
		t.Errorf("prêts actifs après retour = %d, attendu 0", count)
	}
	if avail := availabilityOf(t, k.ID).AvailableCount; avail != 2 {
		t.Errorf("disponible après retour = %d, attendu 2", avail)
	}
}

// TestCreateMultipleLoansSeveralKeys vérifie qu'un prêt groupé crée bien un
// enregistrement par clé, atomiquement.
func TestCreateMultipleLoansSeveralKeys(t *testing.T) {
	setupTestDB(t)
	k1 := mustCreateKey(t, "K-01", 2, 0)
	k2 := mustCreateKey(t, "K-02", 2, 0)
	b := mustCreateBorrower(t, "Bob")

	loanIDs, err := CreateMultipleLoans([]int{k1.ID, k2.ID}, b.ID)
	if err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}
	if len(loanIDs) != 2 {
		t.Fatalf("IDs retournés = %d, attendu 2", len(loanIDs))
	}
	loans, err := GetActiveLoansByBorrowerID(b.ID)
	if err != nil {
		t.Fatalf("GetActiveLoansByBorrowerID: %v", err)
	}
	if len(loans) != 2 {
		t.Errorf("prêts créés = %d, attendu 2", len(loans))
	}
	// Les IDs retournés doivent correspondre exactement aux prêts en base
	got := map[int]bool{}
	for _, l := range loans {
		got[l.Loan.ID] = true
	}
	for _, id := range loanIDs {
		if !got[id] {
			t.Errorf("ID retourné %d absent des prêts actifs", id)
		}
	}
}

// TestOverloanAllowed verrouille un comportement voulu : prêter plus
// d'exemplaires que le stock déclaré doit RÉUSSIR (le contrôle d'inventaire
// est informatif, jamais bloquant). Le disponible devient alors négatif.
func TestOverloanAllowed(t *testing.T) {
	setupTestDB(t)
	k := mustCreateKey(t, "K-01", 1, 0)
	alice := mustCreateBorrower(t, "Alice")
	bob := mustCreateBorrower(t, "Bob")

	if _, err := CreateMultipleLoans([]int{k.ID}, alice.ID); err != nil {
		t.Fatalf("premier prêt: %v", err)
	}
	// Deuxième prêt du même exemplaire : doit passer sans erreur.
	if _, err := CreateMultipleLoans([]int{k.ID}, bob.ID); err != nil {
		t.Fatalf("sur-prêt refusé alors qu'il doit être autorisé: %v", err)
	}

	kwa := availabilityOf(t, k.ID)
	if kwa.AvailableCount != -1 {
		t.Errorf("disponible = %d, attendu -1 (sur-prêt)", kwa.AvailableCount)
	}
	if kwa.LoanedCount != 2 {
		t.Errorf("prêtées = %d, attendu 2", kwa.LoanedCount)
	}
}

// TestUpdateKeyQuantityTotal vérifie l'ajustement rapide du stock, y compris
// le passage sous le nombre d'exemplaires déjà prêtés.
func TestUpdateKeyQuantityTotal(t *testing.T) {
	setupTestDB(t)
	k := mustCreateKey(t, "K-01", 5, 0)
	b := mustCreateBorrower(t, "Alice")
	if _, err := CreateMultipleLoans([]int{k.ID}, b.ID); err != nil {
		t.Fatalf("CreateMultipleLoans: %v", err)
	}

	// Réduire le stock sous le nombre prêté doit être accepté en base :
	// c'est l'alerte d'inventaire, côté interface, qui signale l'anomalie.
	if err := UpdateKeyQuantityTotal(k.ID, 0); err != nil {
		t.Fatalf("UpdateKeyQuantityTotal: %v", err)
	}
	got, err := GetKeyByID(k.ID)
	if err != nil {
		t.Fatalf("GetKeyByID: %v", err)
	}
	if got.QuantityTotal != 0 {
		t.Errorf("stock total = %d, attendu 0", got.QuantityTotal)
	}
	if avail := availabilityOf(t, k.ID).AvailableCount; avail != -1 {
		t.Errorf("disponible = %d, attendu -1", avail)
	}
}

// TestCheckInventoryAnomalies vérifie la détection des clés en sur-prêt et
// l'absence de faux positifs sur un inventaire cohérent.
func TestCheckInventoryAnomalies(t *testing.T) {
	setupTestDB(t)
	sain := mustCreateKey(t, "K-01", 3, 1)
	surpret := mustCreateKey(t, "K-02", 1, 0)
	alice := mustCreateBorrower(t, "Alice")
	bob := mustCreateBorrower(t, "Bob")

	if _, err := CreateMultipleLoans([]int{sain.ID, surpret.ID}, alice.ID); err != nil {
		t.Fatalf("prêts initiaux: %v", err)
	}
	anomalies, err := CheckInventoryAnomalies()
	if err != nil {
		t.Fatalf("CheckInventoryAnomalies: %v", err)
	}
	if len(anomalies) != 0 {
		t.Fatalf("faux positif : %+v", anomalies)
	}

	// Deuxième sortie du seul exemplaire de K-02 → sur-prêt
	if _, err := CreateMultipleLoans([]int{surpret.ID}, bob.ID); err != nil {
		t.Fatalf("sur-prêt: %v", err)
	}
	anomalies, err = CheckInventoryAnomalies()
	if err != nil {
		t.Fatalf("CheckInventoryAnomalies: %v", err)
	}
	if len(anomalies) != 1 {
		t.Fatalf("anomalies = %d, attendu 1 (%+v)", len(anomalies), anomalies)
	}
	a := anomalies[0]
	if a.KeyNumber != "K-02" || a.Loaned != 2 || a.Available != -1 || a.Total != 1 {
		t.Errorf("anomalie inattendue : %+v", a)
	}
}
