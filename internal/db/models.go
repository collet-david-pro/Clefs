package db

import "time"

// Key représente une clé physique dans le système
type Key struct {
	ID              int    `db:"id"`
	Number          string `db:"number"`
	Description     string `db:"description"`
	QuantityTotal   int    `db:"quantity_total"`
	QuantityReserve int    `db:"quantity_reserve"`
	StorageLocation string `db:"storage_location"`
	Category        string `db:"category"` // simple, trousseau, badge, passe
	Notes           string `db:"notes"`
	Rooms           []Room
}

// Room représente un accès physique (porte, portail, zone)
type Room struct {
	ID         int    `db:"id"`
	Name       string `db:"name"`
	Type       string `db:"type"`
	BuildingID int    `db:"building_id"`
	Floor      string `db:"floor"`    // RDC, R+1, Sous-sol...
	Category   string `db:"category"` // salle de classe, local technique, bureau...
	Notes      string `db:"notes"`
	Building   Building
	Keys       []Key
}

// Borrower représente un détenteur de clé
type Borrower struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"` // permanent, contractuel, intervenant, entreprise
	Phone  string `db:"phone"`
	Loans  []Loan
}

// Loan représente un prêt de clé
type Loan struct {
	ID                int        `db:"id"`
	KeyID             int        `db:"key_id"`
	BorrowerID        int        `db:"borrower_id"`
	LoanDate          time.Time  `db:"loan_date"`
	ReturnDate        *time.Time `db:"return_date"`
	PlannedReturnDate *time.Time `db:"planned_return_date"`
	LoanType          string     `db:"loan_type"` // ponctuel, permanent
	ReturnedCondition string     `db:"returned_condition"`
	CreatedBy         string     `db:"created_by"`
	Key               Key
	Borrower          Borrower
}

// Building représente un bâtiment
type Building struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Rooms []Room
}

// KeyRoomAssociation représente la table d'association many-to-many
type KeyRoomAssociation struct {
	KeyID  int `db:"key_id"`
	RoomID int `db:"room_id"`
}

// KeyWithAvailability contient une clé avec ses informations de disponibilité
type KeyWithAvailability struct {
	Key
	LoanedCount    int
	AvailableCount int
	BorrowerNames  []string
}

// KeyWithCoverage contient une clé disponible et la liste des accès qu'elle couvre
// parmi ceux demandés (utilisé par l'algorithme de prêt par besoin)
type KeyWithCoverage struct {
	Key
	AvailableCount   int
	CoveredAccessIDs []int
}

// LoanWithDetails contient un prêt avec tous les détails dénormalisés
type LoanWithDetails struct {
	Loan
	KeyNumber        string
	KeyDescription   string
	BorrowerName     string
	BorrowerEmail    string
	PlannedReturnStr string // formaté pour affichage
}

// LoanFilters regroupe les critères de filtrage de l'historique des prêts
type LoanFilters struct {
	BorrowerID *int
	KeyID      *int
	BuildingID *int
	DateFrom   *time.Time
	DateTo     *time.Time
	Status     string // "active", "returned", "overdue", "" (tous)
}

// BorrowerWithKeys contient un détenteur avec ses clés actuelles et les accès cumulés
type BorrowerWithKeys struct {
	Borrower
	ActiveLoans     []LoanWithDetails
	CoveredAccesses []Room
	Redundancies    []Room
}

// RedundancyReport signale les accès couverts par plusieurs clés chez un même détenteur
type RedundancyReport struct {
	Borrower          Borrower
	Keys              []Key
	RedundantAccesses []Room
}

// InventoryAnomaly décrit une clé dont le nombre d'exemplaires sortis dépasse
// le stock utilisable déclaré (total - réserve), donnant un disponible négatif.
// Le sur-prêt étant volontairement autorisé, cette structure sert uniquement
// au signalement à l'écran (« erreur d'inventaire, vérifier le stock »).
type InventoryAnomaly struct {
	KeyID     int
	KeyNumber string
	Total     int
	Reserve   int
	Loaned    int
	Available int // toujours strictement négatif
}
