package app

import (
	"time"

	"clefs/internal/db"
)

// Les DTO (Data Transfer Objects) sont les structures échangées avec le
// frontend. Wails en génère les types TypeScript (frontend/wailsjs/go/models.ts).
// Les dates sont transmises en chaînes ISO 8601 (vides si absentes) : le
// frontend les formate en français.

// AppInfo décrit l'installation courante.
type AppInfo struct {
	Version      string `json:"version"`
	DataDir      string `json:"dataDir"`
	DBPath       string `json:"dbPath"`
	DocumentsDir string `json:"documentsDir"`
	BackupsDir   string `json:"backupsDir"`
	Agent        string `json:"agent"`
	IsEmpty      bool   `json:"isEmpty"`
}

// KeyDTO est une clé avec sa disponibilité calculée.
type KeyDTO struct {
	ID              int      `json:"id"`
	Number          string   `json:"number"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	StorageLocation string   `json:"storageLocation"`
	Notes           string   `json:"notes"`
	Total           int      `json:"total"`
	Reserve         int      `json:"reserve"`
	Loaned          int      `json:"loaned"`
	Available       int      `json:"available"` // Total − Réserve − Sorties, négatif en sur-prêt
	Holders         []string `json:"holders"`
	AccessIDs       []int    `json:"accessIds"`
}

// KeyInput est la saisie d'une clé (création si ID = 0).
type KeyInput struct {
	ID              int    `json:"id"`
	Number          string `json:"number"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	StorageLocation string `json:"storageLocation"`
	Notes           string `json:"notes"`
	Total           int    `json:"total"`
	Reserve         int    `json:"reserve"`
	AccessIDs       []int  `json:"accessIds"`
}

// KeyDetail est la fiche complète d'une clé.
type KeyDetail struct {
	Key         KeyDTO      `json:"key"`
	Accesses    []AccessDTO `json:"accesses"`
	ActiveLoans []LoanDTO   `json:"activeLoans"`
	History     []LoanDTO   `json:"history"`
}

// AccessDTO est un accès (porte, portail, zone), table rooms.
type AccessDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	BuildingID   int    `json:"buildingId"`
	BuildingName string `json:"buildingName"`
	Floor        string `json:"floor"`
	Category     string `json:"category"`
	Notes        string `json:"notes"`
	KeyIDs       []int  `json:"keyIds"`
}

// AccessInput est la saisie d'un accès (création si ID = 0).
type AccessInput struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	BuildingID int    `json:"buildingId"`
	Floor      string `json:"floor"`
	Category   string `json:"category"`
	Notes      string `json:"notes"`
}

// AccessDetail est la fiche complète d'un accès.
type AccessDetail struct {
	Access  AccessDTO   `json:"access"`
	Keys    []KeyDTO    `json:"keys"`
	Holders []HolderDTO `json:"holders"`
}

// HolderDTO indique qu'un détenteur peut ouvrir un accès avec une clé donnée.
type HolderDTO struct {
	BorrowerID   int    `json:"borrowerId"`
	BorrowerName string `json:"borrowerName"`
	KeyNumber    string `json:"keyNumber"`
}

// BuildingDTO est un bâtiment avec son nombre d'accès.
type BuildingDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AccessCount int    `json:"accessCount"`
}

// BorrowerDTO est un détenteur avec son nombre de clés détenues.
type BorrowerDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	Phone       string `json:"phone"`
	ActiveLoans int    `json:"activeLoans"`
}

// BorrowerInput est la saisie d'un détenteur (création si ID = 0).
type BorrowerInput struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
	Phone  string `json:"phone"`
}

// BorrowerDetail est la fiche complète d'un détenteur.
type BorrowerDetail struct {
	Borrower        BorrowerDTO `json:"borrower"`
	ActiveLoans     []LoanDTO   `json:"activeLoans"`
	CoveredAccesses []AccessDTO `json:"coveredAccesses"`
	Redundancies    []AccessDTO `json:"redundancies"`
	History         []LoanDTO   `json:"history"`
}

// LoanDTO est un prêt avec les informations utiles à l'affichage.
type LoanDTO struct {
	ID                int    `json:"id"`
	KeyID             int    `json:"keyId"`
	KeyNumber         string `json:"keyNumber"`
	KeyDescription    string `json:"keyDescription"`
	BorrowerID        int    `json:"borrowerId"`
	BorrowerName      string `json:"borrowerName"`
	LoanDate          string `json:"loanDate"`
	ReturnDate        string `json:"returnDate"`
	PlannedReturnDate string `json:"plannedReturnDate"`
	LoanType          string `json:"loanType"`
	Condition         string `json:"condition"`
	CreatedBy         string `json:"createdBy"`
	Status            string `json:"status"` // active, returned, overdue
	Days              int    `json:"days"`   // durée du prêt (jusqu'au retour ou à aujourd'hui)
}

// AnomalyDTO signale une clé en sur-prêt (disponible négatif).
type AnomalyDTO struct {
	KeyID     int    `json:"keyId"`
	KeyNumber string `json:"keyNumber"`
	Total     int    `json:"total"`
	Reserve   int    `json:"reserve"`
	Loaned    int    `json:"loaned"`
	Available int    `json:"available"`
}

// RedundancyDTO signale un détenteur dont plusieurs clés ouvrent les mêmes accès.
type RedundancyDTO struct {
	BorrowerID   int      `json:"borrowerId"`
	BorrowerName string   `json:"borrowerName"`
	KeyNumbers   []string `json:"keyNumbers"`
	Accesses     []string `json:"accesses"`
}

// Dashboard regroupe les indicateurs et alertes du tableau de bord.
type Dashboard struct {
	KeyTypes          int             `json:"keyTypes"`
	KeyCopies         int             `json:"keyCopies"`
	LoanedCopies      int             `json:"loanedCopies"`
	AvailableCopies   int             `json:"availableCopies"`
	Borrowers         int             `json:"borrowers"`
	BorrowersWithKeys int             `json:"borrowersWithKeys"`
	Accesses          int             `json:"accesses"`
	Overdue           []LoanDTO       `json:"overdue"`
	Anomalies         []AnomalyDTO    `json:"anomalies"`
	Redundancies      []RedundancyDTO `json:"redundancies"`
	RecentLoans       []LoanDTO       `json:"recentLoans"`
}

// SuggestRequest est la demande de calcul du trousseau (prêt par besoin).
type SuggestRequest struct {
	BorrowerID     int   `json:"borrowerId"`
	AccessIDs      []int `json:"accessIds"`
	ExcludedKeyIDs []int `json:"excludedKeyIds"` // clés retirées à la main : jamais re-proposées
	ManualKeyIDs   []int `json:"manualKeyIds"`   // clés ajoutées à la main : toujours conservées
}

// SuggestedKey est une clé du trousseau calculé.
type SuggestedKey struct {
	Key       KeyDTO   `json:"key"`
	Suggested bool     `json:"suggested"` // proposée par l'algorithme
	Manual    bool     `json:"manual"`    // ajoutée à la main
	Covers    []string `json:"covers"`    // accès demandés que cette clé ouvre
}

// Suggestion est le trousseau calculé pour un besoin.
type Suggestion struct {
	Keys         []SuggestedKey `json:"keys"`
	Uncoverable  []AccessDTO    `json:"uncoverable"`
	Redundancies []string       `json:"redundancies"` // accès déjà couverts par les clés du détenteur
}

// CreateLoanRequest est la validation d'une remise de clés.
type CreateLoanRequest struct {
	BorrowerID    int    `json:"borrowerId"`
	KeyIDs        []int  `json:"keyIds"`
	PlannedReturn string `json:"plannedReturn"` // AAAA-MM-JJ, vide si non renseigné
	LoanType      string `json:"loanType"`
}

// LoanResult est le compte rendu d'une remise de clés.
type LoanResult struct {
	Count        int          `json:"count"`
	BorrowerName string       `json:"borrowerName"`
	ReceiptPath  string       `json:"receiptPath"`
	ReceiptError string       `json:"receiptError"`
	Anomalies    []AnomalyDTO `json:"anomalies"`
}

// HistoryFilter est le filtre de l'historique (0 ou "" = pas de filtre).
type HistoryFilter struct {
	BorrowerID int    `json:"borrowerId"`
	KeyID      int    `json:"keyId"`
	From       string `json:"from"` // AAAA-MM-JJ
	To         string `json:"to"`   // AAAA-MM-JJ
	Status     string `json:"status"`
}

// WhoHasWhat est une ligne de la vue « Qui a quoi ? ».
type WhoHasWhat struct {
	Borrower     BorrowerDTO `json:"borrower"`
	Loans        []LoanDTO   `json:"loans"`
	Accesses     []string    `json:"accesses"`
	Redundancies []string    `json:"redundancies"`
}

// PlanBuilding est un bâtiment du plan de clés.
type PlanBuilding struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	Accesses []PlanAccess `json:"accesses"`
}

// PlanAccess est un accès du plan de clés avec les clés qui l'ouvrent.
type PlanAccess struct {
	Access AccessDTO `json:"access"`
	Keys   []KeyDTO  `json:"keys"`
}

// BackupDTO est une sauvegarde du dossier backups/.
type BackupDTO struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size string `json:"size"`
	Date string `json:"date"`
}

// DatabaseFileInfo est l'aperçu d'une base à importer (assistant d'import).
type DatabaseFileInfo struct {
	Path        string `json:"path"`
	Kind        string `json:"kind"`  // v3, legacy, invalid
	Label       string `json:"label"` // libellé lisible de la version détectée
	Buildings   int    `json:"buildings"`
	Accesses    int    `json:"accesses"`
	Keys        int    `json:"keys"`
	Borrowers   int    `json:"borrowers"`
	Loans       int    `json:"loans"`
	ActiveLoans int    `json:"activeLoans"`
}

// ImportResult est le compte rendu d'un import de base.
type ImportResult struct {
	Kind      string `json:"kind"`
	Buildings int    `json:"buildings"`
	Accesses  int    `json:"accesses"`
	Keys      int    `json:"keys"`
	Borrowers int    `json:"borrowers"`
	Loans     int    `json:"loans"`
}

// CSVImportReport est le compte rendu d'un import CSV de détenteurs.
type CSVImportReport struct {
	Created  int      `json:"created"`
	Problems []string `json:"problems"`
}

// --- conversions ---

// isoDate formate une date en ISO 8601, ou "" si nil.
func isoDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// toKeyDTO convertit une clé avec disponibilité.
func toKeyDTO(k db.KeyWithAvailability, accessIDs []int) KeyDTO {
	holders := k.BorrowerNames
	if holders == nil {
		holders = []string{}
	}
	if accessIDs == nil {
		accessIDs = []int{}
	}
	return KeyDTO{
		ID: k.ID, Number: k.Number, Description: k.Description, Category: k.Category,
		StorageLocation: k.StorageLocation, Notes: k.Notes,
		Total: k.QuantityTotal, Reserve: k.QuantityReserve,
		Loaned: k.LoanedCount, Available: k.AvailableCount,
		Holders: holders, AccessIDs: accessIDs,
	}
}

// toAccessDTO convertit un accès.
func toAccessDTO(r db.Room, buildingName string, keyIDs []int) AccessDTO {
	if keyIDs == nil {
		keyIDs = []int{}
	}
	return AccessDTO{
		ID: r.ID, Name: r.Name, Type: r.Type, BuildingID: r.BuildingID, BuildingName: buildingName,
		Floor: r.Floor, Category: r.Category, Notes: r.Notes, KeyIDs: keyIDs,
	}
}

// toLoanDTO convertit un prêt et calcule son statut.
func toLoanDTO(l db.LoanWithDetails, now time.Time) LoanDTO {
	status := "active"
	end := now
	if l.ReturnDate != nil {
		status = "returned"
		end = *l.ReturnDate
	} else if l.PlannedReturnDate != nil && now.After(*l.PlannedReturnDate) {
		status = "overdue"
	}
	days := int(end.Sub(l.LoanDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	loanType := l.LoanType
	if loanType == "" {
		loanType = "ponctuel"
	}
	return LoanDTO{
		ID: l.ID, KeyID: l.KeyID, KeyNumber: l.KeyNumber, KeyDescription: l.KeyDescription,
		BorrowerID: l.BorrowerID, BorrowerName: l.BorrowerName,
		LoanDate: l.LoanDate.Format(time.RFC3339), ReturnDate: isoDate(l.ReturnDate),
		PlannedReturnDate: isoDate(l.PlannedReturnDate), LoanType: loanType,
		Condition: l.ReturnedCondition, CreatedBy: l.CreatedBy, Status: status, Days: days,
	}
}

// toLoanDTOs convertit une liste de prêts (jamais nil, pour un tableau JSON).
func toLoanDTOs(loans []db.LoanWithDetails) []LoanDTO {
	now := time.Now()
	out := make([]LoanDTO, 0, len(loans))
	for _, l := range loans {
		out = append(out, toLoanDTO(l, now))
	}
	return out
}

// toAnomalyDTOs convertit les anomalies d'inventaire.
func toAnomalyDTOs(as []db.InventoryAnomaly) []AnomalyDTO {
	out := make([]AnomalyDTO, 0, len(as))
	for _, a := range as {
		out = append(out, AnomalyDTO{
			KeyID: a.KeyID, KeyNumber: a.KeyNumber, Total: a.Total,
			Reserve: a.Reserve, Loaned: a.Loaned, Available: a.Available,
		})
	}
	return out
}

// toBorrowerDTO convertit un détenteur.
func toBorrowerDTO(b db.Borrower, activeLoans int) BorrowerDTO {
	status := b.Status
	if status == "" {
		status = "permanent"
	}
	return BorrowerDTO{ID: b.ID, Name: b.Name, Email: b.Email, Status: status, Phone: b.Phone, ActiveLoans: activeLoans}
}
