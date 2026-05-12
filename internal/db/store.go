package db

// Store définit toutes les opérations de persistance de l'application.
// Cette interface permet les tests unitaires via des implémentations mock.
type Store interface {
	// Bâtiments
	GetAllBuildings() ([]Building, error)
	GetBuildingByID(id int) (*Building, error)
	CreateBuilding(b *Building) error
	UpdateBuilding(b *Building) error
	DeleteBuilding(id int) error

	// Accès (salles/portes)
	GetAllAccesses() ([]Room, error)
	GetAccessesByBuilding(buildingID int) ([]Room, error)
	GetAccessesByFloor(buildingID int, floor string) ([]Room, error)
	GetAccessesByCategory(category string) ([]Room, error)
	CreateAccess(r *Room) error
	UpdateAccess(r *Room) error
	DeleteAccess(id int) error

	// Clés
	GetAllKeys() ([]Key, error)
	GetKeyByID(id int) (*Key, error)
	GetKeysWithAvailability() ([]KeyWithAvailability, error)
	GetKeysForAccess(accessID int) ([]Key, error)
	GetAvailableKeysForAccesses(accessIDs []int) ([]KeyWithCoverage, error)
	GetRoomsForKey(keyID int) ([]Room, error)
	GetKeyHistory(keyID int) ([]LoanWithDetails, error)
	CreateKey(k *Key, accessIDs []int) error
	UpdateKey(k *Key, accessIDs []int) error
	DeleteKey(id int) error

	// Détenteurs
	GetAllBorrowers() ([]Borrower, error)
	GetBorrowerByID(id int) (*Borrower, error)
	GetBorrowerWithCurrentKeys(id int) (*BorrowerWithKeys, error)
	GetBorrowerHistory(borrowerID int) ([]LoanWithDetails, error)
	CreateBorrower(b *Borrower) error
	UpdateBorrower(b *Borrower) error
	DeleteBorrower(id int) error

	// Prêts
	GetAllActiveLoans() ([]LoanWithDetails, error)
	GetLoanByID(id int) (*LoanWithDetails, error)
	GetActiveLoansByKeyID(keyID int) ([]LoanWithDetails, error)
	GetActiveLoansByBorrowerID(borrowerID int) ([]LoanWithDetails, error)
	GetOverdueLoans() ([]LoanWithDetails, error)
	GetLoanHistory(filters LoanFilters) ([]LoanWithDetails, error)
	GetActiveLoanCount(keyID int) (int, error)
	CreateLoan(keyID, borrowerID int) error
	CreateMultipleLoans(keyIDs []int, borrowerID int) error
	ReturnLoan(loanID int) error

	// Redondances
	GetBorrowersWithRedundantAccesses() ([]RedundancyReport, error)
}
