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
	SelectedKeys      []db.Key // clés à remettre
	UncoverableIDs    []int    // accès qu'aucune clé disponible ne couvre
	HasUncoverable    bool
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

func toSet(ids []int) map[int]struct{} {
	s := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

func setToSlice(s map[int]struct{}) []int {
	out := make([]int, 0, len(s))
	for id := range s {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

func intersect(ids []int, set map[int]struct{}) map[int]struct{} {
	result := make(map[int]struct{})
	for _, id := range ids {
		if _, ok := set[id]; ok {
			result[id] = struct{}{}
		}
	}
	return result
}

func intersectCount(ids []int, set map[int]struct{}) int {
	count := 0
	for _, id := range ids {
		if _, ok := set[id]; ok {
			count++
		}
	}
	return count
}

func removeCandidate(candidates []db.KeyWithCoverage, keyID int) []db.KeyWithCoverage {
	out := make([]db.KeyWithCoverage, 0, len(candidates)-1)
	for _, c := range candidates {
		if c.ID != keyID {
			out = append(out, c)
		}
	}
	return out
}
