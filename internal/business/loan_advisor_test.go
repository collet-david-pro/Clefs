package business

import (
	"clefs/internal/db"
	"testing"
)

// makeKey construit une KeyWithCoverage de test
func makeKey(id int, number string, available int, coveredIDs []int) db.KeyWithCoverage {
	return db.KeyWithCoverage{
		Key:              db.Key{ID: id, Number: number},
		AvailableCount:   available,
		CoveredAccessIDs: coveredIDs,
	}
}

func TestSuggestKeys_ExactCoverage(t *testing.T) {
	// Une clé couvre exactement les accès demandés
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10, 20, 30}),
	}
	result := SuggestKeys([]int{10, 20, 30}, keys)

	if result.HasUncoverable {
		t.Fatalf("ne devrait pas avoir d'accès non couvrable")
	}
	if len(result.SelectedKeys) != 1 {
		t.Fatalf("attendu 1 clé, obtenu %d", len(result.SelectedKeys))
	}
	if result.SelectedKeys[0].Number != "K001" {
		t.Errorf("mauvaise clé sélectionnée: %s", result.SelectedKeys[0].Number)
	}
}

func TestSuggestKeys_MinimalCombination(t *testing.T) {
	// K001 couvre {10,20}, K002 couvre {30}, K003 couvre {10,20,30}
	// Résultat attendu : K003 seule (1 clé suffit)
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10, 20}),
		makeKey(2, "K002", 1, []int{30}),
		makeKey(3, "K003", 1, []int{10, 20, 30}),
	}
	result := SuggestKeys([]int{10, 20, 30}, keys)

	if result.HasUncoverable {
		t.Fatal("ne devrait pas avoir d'accès non couvrable")
	}
	if len(result.SelectedKeys) != 1 {
		t.Fatalf("attendu 1 clé (greedy doit choisir K003), obtenu %d", len(result.SelectedKeys))
	}
	if result.SelectedKeys[0].Number != "K003" {
		t.Errorf("mauvaise clé sélectionnée: %s", result.SelectedKeys[0].Number)
	}
}

func TestSuggestKeys_TwoKeysRequired(t *testing.T) {
	// Aucune clé seule ne couvre tout — besoin de 2
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10, 20}),
		makeKey(2, "K002", 1, []int{30, 40}),
	}
	result := SuggestKeys([]int{10, 20, 30, 40}, keys)

	if result.HasUncoverable {
		t.Fatal("ne devrait pas avoir d'accès non couvrable")
	}
	if len(result.SelectedKeys) != 2 {
		t.Fatalf("attendu 2 clés, obtenu %d", len(result.SelectedKeys))
	}
}

func TestSuggestKeys_UncoverableAccess(t *testing.T) {
	// L'accès 99 n'est couvert par aucune clé
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10, 20}),
	}
	result := SuggestKeys([]int{10, 20, 99}, keys)

	if !result.HasUncoverable {
		t.Fatal("devrait signaler un accès non couvrable")
	}
	if len(result.UncoverableIDs) != 1 || result.UncoverableIDs[0] != 99 {
		t.Errorf("accès non couvrable attendu: [99], obtenu: %v", result.UncoverableIDs)
	}
	// Les clés qui couvrent partiellement sont quand même sélectionnées
	if len(result.SelectedKeys) != 1 {
		t.Errorf("attendu 1 clé partielle sélectionnée, obtenu %d", len(result.SelectedKeys))
	}
}

func TestSuggestKeys_NoKeysAvailable(t *testing.T) {
	result := SuggestKeys([]int{10, 20}, []db.KeyWithCoverage{})

	if !result.HasUncoverable {
		t.Fatal("devrait signaler des accès non couvrables quand il n'y a aucune clé")
	}
	if len(result.SelectedKeys) != 0 {
		t.Errorf("ne devrait sélectionner aucune clé, obtenu %d", len(result.SelectedKeys))
	}
}

func TestSuggestKeys_EmptyRequest(t *testing.T) {
	// Aucun accès demandé — résultat vide, pas d'erreur
	result := SuggestKeys([]int{}, []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10, 20}),
	})

	if result.HasUncoverable {
		t.Fatal("aucun accès demandé, ne devrait pas être non couvrable")
	}
	if len(result.SelectedKeys) != 0 {
		t.Errorf("aucune clé ne devrait être sélectionnée, obtenu %d", len(result.SelectedKeys))
	}
}

func TestSuggestKeys_MinimizesBonus(t *testing.T) {
	// K001 couvre {10} — 0 accès bonus
	// K002 couvre {10, 20, 30} — 2 accès bonus (20 et 30 non demandés)
	// On demande uniquement {10} : doit choisir K001 (moins de surface)
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 1, []int{10}),
		makeKey(2, "K002", 1, []int{10, 20, 30}),
	}
	result := SuggestKeys([]int{10}, keys)

	if result.HasUncoverable {
		t.Fatal("ne devrait pas avoir d'accès non couvrable")
	}
	if len(result.SelectedKeys) != 1 {
		t.Fatalf("attendu 1 clé, obtenu %d", len(result.SelectedKeys))
	}
	if result.SelectedKeys[0].Number != "K001" {
		t.Errorf("devrait choisir K001 (moins d'accès bonus), obtenu %s", result.SelectedKeys[0].Number)
	}
}

func TestSuggestKeys_UnavailableKeyIgnored(t *testing.T) {
	// K001 disponible=0 → doit être ignorée, K002 couvre tout
	keys := []db.KeyWithCoverage{
		makeKey(1, "K001", 0, []int{10, 20}), // non disponible
		makeKey(2, "K002", 1, []int{10, 20}),
	}
	result := SuggestKeys([]int{10, 20}, keys)

	if result.HasUncoverable {
		t.Fatal("ne devrait pas avoir d'accès non couvrable")
	}
	if len(result.SelectedKeys) != 1 || result.SelectedKeys[0].Number != "K002" {
		t.Errorf("devrait choisir K002, obtenu %v", result.SelectedKeys)
	}
}
