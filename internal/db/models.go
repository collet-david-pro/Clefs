package db

import (
	"time"
)

// Key représente une clé dans le système
type Key struct {
	ID              int       `db:"id"`
	Number          string    `db:"number"`
	Description     string    `db:"description"`
	QuantityTotal   int       `db:"quantity_total"`
	QuantityReserve int       `db:"quantity_reserve"`
	StorageLocation string    `db:"storage_location"`
	Rooms           []Room    // Relation many-to-many
}

// Room représente une salle/pièce
type Room struct {
	ID         int      `db:"id"`
	Name       string   `db:"name"`
	Type       string   `db:"type"`
	BuildingID int      `db:"building_id"`
	Building   Building // Relation
	Keys       []Key    // Relation many-to-many
}

// Borrower représente un emprunteur
type Borrower struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Loans []Loan // Relation
}

// Loan représente un emprunt de clé
type Loan struct {
	ID         int        `db:"id"`
	KeyID      int        `db:"key_id"`
	BorrowerID int        `db:"borrower_id"`
	LoanDate   time.Time  `db:"loan_date"`
	ReturnDate *time.Time `db:"return_date"`
	Key        Key        // Relation
	Borrower   Borrower   // Relation
}

// Building représente un bâtiment
type Building struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Rooms []Room // Relation
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

// LoanWithDetails contient un emprunt avec tous les détails
type LoanWithDetails struct {
	Loan
	KeyNumber       string
	KeyDescription  string
	BorrowerName    string
	BorrowerEmail   string
}
