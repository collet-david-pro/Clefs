package db

import (
	"path/filepath"
	"testing"
)

// setupTestDB initialise une base vierge dans un répertoire temporaire propre
// au test, et referme la connexion à la fin. Le package s'appuyant sur le
// singleton DB, les tests ne doivent PAS utiliser t.Parallel().
func setupTestDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "clefs_test.db")
	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		if err := CloseDB(); err != nil {
			t.Errorf("CloseDB: %v", err)
		}
	})
	return dbPath
}

// mustCreateKey insère une clé et échoue le test en cas d'erreur.
func mustCreateKey(t *testing.T, number string, total, reserve int) *Key {
	t.Helper()
	k := &Key{Number: number, Description: "clé de test", QuantityTotal: total, QuantityReserve: reserve}
	if err := CreateKey(k, nil); err != nil {
		t.Fatalf("CreateKey(%s): %v", number, err)
	}
	return k
}

// mustCreateBorrower insère un détenteur et échoue le test en cas d'erreur.
func mustCreateBorrower(t *testing.T, name string) *Borrower {
	t.Helper()
	b := &Borrower{Name: name, Email: name + "@test.local"}
	if err := CreateBorrower(b); err != nil {
		t.Fatalf("CreateBorrower(%s): %v", name, err)
	}
	return b
}

// availabilityOf retourne la ligne de disponibilité d'une clé donnée.
func availabilityOf(t *testing.T, keyID int) KeyWithAvailability {
	t.Helper()
	kwas, err := GetKeysWithAvailability()
	if err != nil {
		t.Fatalf("GetKeysWithAvailability: %v", err)
	}
	for _, kwa := range kwas {
		if kwa.ID == keyID {
			return kwa
		}
	}
	t.Fatalf("clé %d absente de GetKeysWithAvailability", keyID)
	return KeyWithAvailability{}
}
