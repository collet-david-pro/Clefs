package business

import (
	"clefs/internal/db"
	"sort"
)

// DetectRedundancies retourne la liste des accès couverts par plus d'une clé
// dans la collection fournie (clés d'un même détenteur).
func DetectRedundancies(keys []db.Key, accessesByKey map[int][]db.Room) []db.Room {
	accessCount := make(map[int]int)
	accessMap := make(map[int]db.Room)

	for _, key := range keys {
		for _, access := range accessesByKey[key.ID] {
			accessCount[access.ID]++
			accessMap[access.ID] = access
		}
	}

	var redundant []db.Room
	for id, count := range accessCount {
		if count > 1 {
			redundant = append(redundant, accessMap[id])
		}
	}

	sort.Slice(redundant, func(i, j int) bool {
		return redundant[i].Name < redundant[j].Name
	})

	return redundant
}

// BuildBorrowerWithKeys construit un BorrowerWithKeys complet pour un détenteur :
// ses prêts actifs, l'union de tous ses accès couverts, et les redondances.
func BuildBorrowerWithKeys(borrower db.Borrower, loans []db.LoanWithDetails) (db.BorrowerWithKeys, error) {
	result := db.BorrowerWithKeys{
		Borrower:    borrower,
		ActiveLoans: loans,
	}

	if len(loans) == 0 {
		return result, nil
	}

	accessesByKey := make(map[int][]db.Room)
	var keys []db.Key

	for _, loan := range loans {
		key, err := db.GetKeyByID(loan.KeyID)
		if err != nil {
			continue
		}
		keys = append(keys, *key)

		rooms, err := db.GetRoomsForKey(loan.KeyID)
		if err != nil {
			continue
		}
		accessesByKey[loan.KeyID] = rooms
	}

	// Union de tous les accès (dédupliqués)
	seen := make(map[int]struct{})
	for _, rooms := range accessesByKey {
		for _, r := range rooms {
			if _, ok := seen[r.ID]; !ok {
				seen[r.ID] = struct{}{}
				result.CoveredAccesses = append(result.CoveredAccesses, r)
			}
		}
	}
	sort.Slice(result.CoveredAccesses, func(i, j int) bool {
		return result.CoveredAccesses[i].Name < result.CoveredAccesses[j].Name
	})

	// Redondances
	result.Redundancies = DetectRedundancies(keys, accessesByKey)

	return result, nil
}
