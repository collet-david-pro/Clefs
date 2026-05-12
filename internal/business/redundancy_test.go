package business

import (
	"clefs/internal/db"
	"testing"
)

func makeRoom(id int, name string) db.Room {
	return db.Room{ID: id, Name: name}
}

func TestDetectRedundancies_NoRedundancy(t *testing.T) {
	// Deux clés sans accès en commun
	keys := []db.Key{
		{ID: 1, Number: "K001"},
		{ID: 2, Number: "K002"},
	}
	accessesByKey := map[int][]db.Room{
		1: {makeRoom(10, "Salle A"), makeRoom(20, "Salle B")},
		2: {makeRoom(30, "Salle C")},
	}
	result := DetectRedundancies(keys, accessesByKey)
	if len(result) != 0 {
		t.Errorf("attendu 0 redondance, obtenu %d: %v", len(result), result)
	}
}

func TestDetectRedundancies_OneRedundancy(t *testing.T) {
	// K001 et K002 ouvrent toutes deux la Salle A
	keys := []db.Key{
		{ID: 1, Number: "K001"},
		{ID: 2, Number: "K002"},
	}
	accessesByKey := map[int][]db.Room{
		1: {makeRoom(10, "Salle A"), makeRoom(20, "Salle B")},
		2: {makeRoom(10, "Salle A"), makeRoom(30, "Salle C")},
	}
	result := DetectRedundancies(keys, accessesByKey)
	if len(result) != 1 {
		t.Fatalf("attendu 1 redondance, obtenu %d", len(result))
	}
	if result[0].ID != 10 {
		t.Errorf("mauvaise salle redondante: %d", result[0].ID)
	}
}

func TestDetectRedundancies_MultipleRedundancies(t *testing.T) {
	keys := []db.Key{
		{ID: 1, Number: "K001"},
		{ID: 2, Number: "K002"},
	}
	accessesByKey := map[int][]db.Room{
		1: {makeRoom(10, "A"), makeRoom(20, "B"), makeRoom(30, "C")},
		2: {makeRoom(10, "A"), makeRoom(20, "B")},
	}
	result := DetectRedundancies(keys, accessesByKey)
	if len(result) != 2 {
		t.Fatalf("attendu 2 redondances, obtenu %d", len(result))
	}
}

func TestDetectRedundancies_SingleKey(t *testing.T) {
	// Une seule clé — impossible d'avoir des redondances
	keys := []db.Key{{ID: 1, Number: "K001"}}
	accessesByKey := map[int][]db.Room{
		1: {makeRoom(10, "A"), makeRoom(20, "B")},
	}
	result := DetectRedundancies(keys, accessesByKey)
	if len(result) != 0 {
		t.Errorf("attendu 0 redondance avec une seule clé, obtenu %d", len(result))
	}
}

func TestDetectRedundancies_EmptyKeys(t *testing.T) {
	result := DetectRedundancies([]db.Key{}, map[int][]db.Room{})
	if len(result) != 0 {
		t.Errorf("attendu 0 redondance, obtenu %d", len(result))
	}
}

func TestDetectRedundancies_SortedByName(t *testing.T) {
	// Vérifie que le résultat est trié par nom
	keys := []db.Key{
		{ID: 1, Number: "K001"},
		{ID: 2, Number: "K002"},
	}
	accessesByKey := map[int][]db.Room{
		1: {makeRoom(30, "Salle Z"), makeRoom(10, "Salle A")},
		2: {makeRoom(30, "Salle Z"), makeRoom(10, "Salle A")},
	}
	result := DetectRedundancies(keys, accessesByKey)
	if len(result) != 2 {
		t.Fatalf("attendu 2 redondances, obtenu %d", len(result))
	}
	if result[0].Name > result[1].Name {
		t.Errorf("résultat non trié: %s > %s", result[0].Name, result[1].Name)
	}
}

func TestBuildBorrowerWithKeys_NoLoans(t *testing.T) {
	borrower := db.Borrower{ID: 1, Name: "Jean Dupont"}
	result, err := BuildBorrowerWithKeys(borrower, nil)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(result.CoveredAccesses) != 0 {
		t.Errorf("attendu 0 accès couverts, obtenu %d", len(result.CoveredAccesses))
	}
	if len(result.Redundancies) != 0 {
		t.Errorf("attendu 0 redondance, obtenu %d", len(result.Redundancies))
	}
}
