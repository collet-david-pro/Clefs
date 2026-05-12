package db

// SQLiteStore implémente l'interface Store en s'appuyant sur la variable globale DB.
// C'est le pont entre l'interface et les fonctions du package db existantes.
type SQLiteStore struct{}

func NewSQLiteStore() *SQLiteStore { return &SQLiteStore{} }

// --- Bâtiments ---
func (s *SQLiteStore) GetAllBuildings() ([]Building, error)           { return GetAllBuildings() }
func (s *SQLiteStore) GetBuildingByID(id int) (*Building, error)      { return GetBuildingByID(id) }
func (s *SQLiteStore) CreateBuilding(b *Building) error               { return CreateBuilding(b) }
func (s *SQLiteStore) UpdateBuilding(b *Building) error               { return UpdateBuilding(b) }
func (s *SQLiteStore) DeleteBuilding(id int) error                    { return DeleteBuilding(id) }

// --- Accès ---
func (s *SQLiteStore) GetAllAccesses() ([]Room, error)                        { return GetAllAccesses() }
func (s *SQLiteStore) GetAccessesByBuilding(id int) ([]Room, error)           { return GetAccessesByBuilding(id) }
func (s *SQLiteStore) GetAccessesByFloor(bid int, floor string) ([]Room, error) {
	rows, err := DB.Query(roomSelectSQL+` WHERE building_id = ? AND floor = ? ORDER BY name`, bid, floor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRooms(rows)
}
func (s *SQLiteStore) GetAccessesByCategory(cat string) ([]Room, error) {
	rows, err := DB.Query(roomSelectSQL+` WHERE category = ? ORDER BY name`, cat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRooms(rows)
}
func (s *SQLiteStore) CreateAccess(r *Room) error  { return CreateRoom(r) }
func (s *SQLiteStore) UpdateAccess(r *Room) error  { return UpdateRoom(r) }
func (s *SQLiteStore) DeleteAccess(id int) error   { return DeleteRoom(id) }

// --- Clés ---
func (s *SQLiteStore) GetAllKeys() ([]Key, error)                                { return GetAllKeys() }
func (s *SQLiteStore) GetKeyByID(id int) (*Key, error)                           { return GetKeyByID(id) }
func (s *SQLiteStore) GetKeysWithAvailability() ([]KeyWithAvailability, error)   { return GetKeysWithAvailability() }
func (s *SQLiteStore) GetKeysForAccess(id int) ([]Key, error)                    { return GetKeysForAccess(id) }
func (s *SQLiteStore) GetAvailableKeysForAccesses(ids []int) ([]KeyWithCoverage, error) {
	// Délègue à la logique business — retourne toutes les clés dispo avec leur couverture
	available, err := GetAvailableKeys()
	if err != nil {
		return nil, err
	}
	reqSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		reqSet[id] = struct{}{}
	}
	var result []KeyWithCoverage
	for _, k := range available {
		rooms, err := GetRoomsForKey(k.ID)
		if err != nil {
			continue
		}
		var covered []int
		for _, r := range rooms {
			if _, ok := reqSet[r.ID]; ok {
				covered = append(covered, r.ID)
			}
		}
		count, _ := GetActiveLoanCount(k.ID)
		result = append(result, KeyWithCoverage{
			Key:              k,
			AvailableCount:   k.QuantityTotal - k.QuantityReserve - count,
			CoveredAccessIDs: covered,
		})
	}
	return result, nil
}
func (s *SQLiteStore) GetRoomsForKey(id int) ([]Room, error)       { return GetRoomsForKey(id) }
func (s *SQLiteStore) GetKeyHistory(id int) ([]LoanWithDetails, error) { return GetKeyHistory(id) }
func (s *SQLiteStore) CreateKey(k *Key, ids []int) error           { return CreateKey(k, ids) }
func (s *SQLiteStore) UpdateKey(k *Key, ids []int) error           { return UpdateKey(k, ids) }
func (s *SQLiteStore) DeleteKey(id int) error                      { return DeleteKey(id) }

// --- Détenteurs ---
func (s *SQLiteStore) GetAllBorrowers() ([]Borrower, error)             { return GetAllBorrowers() }
func (s *SQLiteStore) GetBorrowerByID(id int) (*Borrower, error)        { return GetBorrowerByID(id) }
func (s *SQLiteStore) GetBorrowerHistory(id int) ([]LoanWithDetails, error) { return GetBorrowerHistory(id) }
func (s *SQLiteStore) GetBorrowerWithCurrentKeys(id int) (*BorrowerWithKeys, error) {
	b, err := GetBorrowerByID(id)
	if err != nil {
		return nil, err
	}
	loans, err := GetActiveLoansByBorrowerID(id)
	if err != nil {
		return nil, err
	}
	// Union des accès couverts + redondances calculées inline
	seen := map[int]struct{}{}
	accessCount := map[int]int{}
	accessMap := map[int]Room{}
	var covered []Room
	for _, l := range loans {
		rooms, _ := GetRoomsForKey(l.KeyID)
		for _, r := range rooms {
			accessCount[r.ID]++
			accessMap[r.ID] = r
			if _, ok := seen[r.ID]; !ok {
				seen[r.ID] = struct{}{}
				covered = append(covered, r)
			}
		}
	}
	var redundant []Room
	for id, cnt := range accessCount {
		if cnt > 1 {
			redundant = append(redundant, accessMap[id])
		}
	}
	return &BorrowerWithKeys{
		Borrower:        *b,
		ActiveLoans:     loans,
		CoveredAccesses: covered,
		Redundancies:    redundant,
	}, nil
}
func (s *SQLiteStore) CreateBorrower(b *Borrower) error { return CreateBorrower(b) }
func (s *SQLiteStore) UpdateBorrower(b *Borrower) error { return UpdateBorrower(b) }
func (s *SQLiteStore) DeleteBorrower(id int) error      { return DeleteBorrower(id) }

// --- Prêts ---
func (s *SQLiteStore) GetAllActiveLoans() ([]LoanWithDetails, error) { return GetAllActiveLoans() }
func (s *SQLiteStore) GetLoanByID(id int) (*LoanWithDetails, error)  { return GetLoanByID(id) }
func (s *SQLiteStore) GetActiveLoansByKeyID(id int) ([]LoanWithDetails, error) {
	return GetActiveLoansByKeyID(id)
}
func (s *SQLiteStore) GetActiveLoansByBorrowerID(id int) ([]LoanWithDetails, error) {
	return GetActiveLoansByBorrowerID(id)
}
func (s *SQLiteStore) GetOverdueLoans() ([]LoanWithDetails, error)              { return GetOverdueLoans() }
func (s *SQLiteStore) GetLoanHistory(f LoanFilters) ([]LoanWithDetails, error)  { return GetLoanHistory(f) }
func (s *SQLiteStore) GetActiveLoanCount(id int) (int, error)                   { return GetActiveLoanCount(id) }
func (s *SQLiteStore) CreateLoan(keyID, borrowerID int) error                   { return CreateLoan(keyID, borrowerID) }
func (s *SQLiteStore) CreateMultipleLoans(keyIDs []int, borrowerID int) error   { return CreateMultipleLoans(keyIDs, borrowerID) }
func (s *SQLiteStore) ReturnLoan(id int) error                                  { return ReturnLoan(id) }

// --- Redondances ---
func (s *SQLiteStore) GetBorrowersWithRedundantAccesses() ([]RedundancyReport, error) {
	return GetBorrowersWithRedundantAccesses()
}

// Vérification à la compilation que SQLiteStore satisfait Store
var _ Store = (*SQLiteStore)(nil)
