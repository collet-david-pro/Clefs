// Package business contient la logique métier pure de l'application,
// indépendante de l'interface graphique.
//
// Deux problèmes y sont traités :
//   - loan_advisor.go : "prêt par besoin" — étant donné des portes à ouvrir,
//     proposer la plus petite combinaison de clés qui les couvre toutes
//     (problème du Set Cover, résolu par un algorithme glouton/greedy).
//   - redundancy.go : détecter qu'un détenteur possède des clés se recouvrant
//     (une clé dont tous les accès sont déjà couverts par les autres).
//
// La plupart des fonctions reçoivent leurs données en paramètre, ce qui les
// rend testables unitairement (cf. loan_advisor_test.go, redundancy_test.go).
package business

import (
	"clefs/internal/db"
	"errors"
	"sort"
)

// ErrAccessesUncoverable est retourné quand certains accès demandés ne peuvent
// être couverts par aucune clé disponible.
var ErrAccessesUncoverable = errors.New("certains accès ne peuvent pas être couverts avec les clés disponibles")

// SuggestionResult est le résultat de SuggestKeys.
type SuggestionResult struct {
	SelectedKeys   []db.Key // clés à remettre
	UncoverableIDs []int    // accès qu'aucune clé disponible ne couvre
	HasUncoverable bool
}

// SuggestKeys retourne la combinaison minimale de clés disponibles pour couvrir
// tous les accès demandés (algorithme greedy Set Cover).
//
// À chaque itération, on choisit la clé disponible qui couvre le plus d'accès
// pas encore couverts. En cas d'ex-æquo, on préfère la clé avec le moins
// d'accès "bonus" (minimiser la surface d'accès non demandée).
func SuggestKeys(requiredAccessIDs []int, availableKeys []db.KeyWithCoverage) SuggestionResult {
	remaining := toSet(requiredAccessIDs)
	// copie locale pour ne pas modifier le slice appelant
	candidates := make([]db.KeyWithCoverage, len(availableKeys))
	copy(candidates, availableKeys)

	var selected []db.Key

	for len(remaining) > 0 {
		best, bestCoverage := findBestKey(candidates, remaining)
		if best == nil || len(bestCoverage) == 0 {
			// Aucune clé ne couvre les accès restants
			return SuggestionResult{
				SelectedKeys:   selected,
				UncoverableIDs: setToSlice(remaining),
				HasUncoverable: true,
			}
		}
		selected = append(selected, best.Key)
		for id := range bestCoverage {
			delete(remaining, id)
		}
		candidates = removeCandidate(candidates, best.ID)
	}

	return SuggestionResult{SelectedKeys: selected}
}

// findBestKey trouve la clé candidate qui couvre le plus d'accès restants.
// En cas d'égalité, préfère celle avec le moins d'accès hors périmètre.
func findBestKey(candidates []db.KeyWithCoverage, remaining map[int]struct{}) (*db.KeyWithCoverage, map[int]struct{}) {
	var best *db.KeyWithCoverage
	var bestCoverage map[int]struct{}
	bestScore := -1
	bestBonus := int(^uint(0) >> 1) // max int

	for i := range candidates {
		c := &candidates[i]
		if c.AvailableCount <= 0 {
			continue
		}
		coverage := intersect(c.CoveredAccessIDs, remaining)
		score := len(coverage)
		if score == 0 {
			continue
		}
		bonus := len(c.CoveredAccessIDs) - score // accès couverts non demandés
		if score > bestScore || (score == bestScore && bonus < bestBonus) {
			best = c
			bestCoverage = coverage
			bestScore = score
			bestBonus = bonus
		}
	}
	return best, bestCoverage
}

// BuildAvailableKeysForAccesses construit la liste des KeyWithCoverage à partir
// des accès demandés en interrogeant la base de données.
func BuildAvailableKeysForAccesses(accessIDs []int) ([]db.KeyWithCoverage, error) {
	allKeys, err := db.GetAvailableKeys()
	if err != nil {
		return nil, err
	}

	required := toSet(accessIDs)
	var result []db.KeyWithCoverage

	for _, key := range allKeys {
		rooms, err := db.GetRoomsForKey(key.ID)
		if err != nil {
			continue
		}
		var covered []int
		for _, r := range rooms {
			if _, ok := required[r.ID]; ok {
				covered = append(covered, r.ID)
			}
		}
		// Même si cette clé ne couvre aucun des accès demandés, on l'inclut
		// pour permettre à l'appelant de savoir quelles clés existent.
		// SuggestKeys filtrera celles sans couverture utile.
		count, err := db.GetActiveLoanCount(key.ID)
		if err != nil {
			continue
		}
		available := key.QuantityTotal - key.QuantityReserve - count
		if available <= 0 {
			continue
		}
		result = append(result, db.KeyWithCoverage{
			Key:              key,
			AvailableCount:   available,
			CoveredAccessIDs: covered,
		})
	}

	// Trier par nombre d'accès couverts décroissant pour aider le greedy
	sort.Slice(result, func(i, j int) bool {
		ci := intersectCount(result[i].CoveredAccessIDs, required)
		cj := intersectCount(result[j].CoveredAccessIDs, required)
		return ci > cj
	})

	return result, nil
}

// --- helpers internes ---
// Petites fonctions ensemblistes basées sur map[int]struct{} (un "set" d'IDs).
// struct{} est utilisé comme valeur car il n'occupe aucune mémoire.

// toSet convertit un slice d'IDs en ensemble (déduplique au passage).
func toSet(ids []int) map[int]struct{} {
	s := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

// setToSlice retourne les IDs d'un ensemble sous forme de slice trié.
func setToSlice(s map[int]struct{}) []int {
	out := make([]int, 0, len(s))
	for id := range s {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

// intersect retourne les IDs présents à la fois dans ids et dans set.
func intersect(ids []int, set map[int]struct{}) map[int]struct{} {
	result := make(map[int]struct{})
	for _, id := range ids {
		if _, ok := set[id]; ok {
			result[id] = struct{}{}
		}
	}
	return result
}

// intersectCount compte les IDs de ids présents dans set (sans allouer de map).
func intersectCount(ids []int, set map[int]struct{}) int {
	count := 0
	for _, id := range ids {
		if _, ok := set[id]; ok {
			count++
		}
	}
	return count
}

// removeCandidate retourne une copie de candidates privée de la clé keyID
// (utilisé après qu'une clé a été sélectionnée par le greedy).
func removeCandidate(candidates []db.KeyWithCoverage, keyID int) []db.KeyWithCoverage {
	out := make([]db.KeyWithCoverage, 0, len(candidates)-1)
	for _, c := range candidates {
		if c.ID != keyID {
			out = append(out, c)
		}
	}
	return out
}
