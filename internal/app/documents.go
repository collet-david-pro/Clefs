package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"clefs/internal/db"
	"clefs/internal/export"
	"clefs/internal/pdf"
)

// documents.go : PDF et CSV produits dans documents/, import CSV des détenteurs.
// Chaque méthode d'export retourne le chemin absolu du fichier créé.

// frDate formate une date ISO 8601 en JJ/MM/AAAA ("" si vide).
func frDate(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Local().Format("02/01/2006")
}

// statusLabel retourne le libellé français d'un statut de prêt.
func statusLabel(s string) string {
	switch s {
	case "returned":
		return "Retourné"
	case "overdue":
		return "En retard"
	default:
		return "En cours"
	}
}

// saveCSV enregistre un CSV Excel France dans documents/ et retourne son chemin absolu.
func (a *App) saveCSV(prefix string, headers []string, rows [][]string) (string, error) {
	path, err := export.SaveCSV(export.Filename(prefix), headers, rows)
	if err != nil {
		return "", err
	}
	return a.absPath(path), nil
}

// savePDF enregistre un PDF dans documents/ et retourne son chemin absolu.
func (a *App) savePDF(prefix string, data []byte, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("génération du document : %w", err)
	}
	path, err := pdf.SavePDF(pdf.GenerateFilename(prefix, 0), data)
	if err != nil {
		return "", err
	}
	return a.absPath(path), nil
}

// stockRows convertit les clés en lignes du bilan du stock.
func stockRows(keys []KeyDTO) []pdf.StockRow {
	rows := make([]pdf.StockRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, pdf.StockRow{
			Number: k.Number, Description: k.Description, Storage: k.StorageLocation,
			Total: k.Total, Reserve: k.Reserve, Loaned: k.Loaned, Available: k.Available,
		})
	}
	return rows
}

// ExportKeyStockPDF génère le bilan PDF du stock de clés.
func (a *App) ExportKeyStockPDF() (string, error) {
	defer a.read()()
	keys, err := listKeys()
	if err != nil {
		return "", err
	}
	data, err := pdf.GenerateStockReport(pdfSettings(), stockRows(keys))
	return a.savePDF("bilan_cles", data, err)
}

// ExportKeyPlanPDF génère le plan de clés en PDF.
func (a *App) ExportKeyPlanPDF() (string, error) {
	defer a.read()()
	plan, err := keyPlan()
	if err != nil {
		return "", err
	}
	buildings := make([]pdf.PlanBuilding, 0, len(plan))
	for _, b := range plan {
		pb := pdf.PlanBuilding{Name: b.Name}
		for _, pa := range b.Accesses {
			keys := make([]string, 0, len(pa.Keys))
			for _, k := range pa.Keys {
				keys = append(keys, k.Number)
			}
			pb.Accesses = append(pb.Accesses, pdf.PlanAccess{Name: pa.Access.Name, Floor: pa.Access.Floor, Category: pa.Access.Category, Keys: keys})
		}
		buildings = append(buildings, pb)
	}
	data, err := pdf.GenerateKeyPlan(pdfSettings(), buildings)
	return a.savePDF("plan_cles", data, err)
}

// ExportKeysCSV exporte l'inventaire des clés.
func (a *App) ExportKeysCSV() (string, error) {
	defer a.read()()
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		return "", err
	}
	headers := []string{"Numéro", "Description", "Catégorie", "Total", "Réserve", "Sorties", "Disponibles", "Emplacement", "Notes"}
	rows := make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{
			k.Number, k.Description, k.Category,
			strconv.Itoa(k.QuantityTotal), strconv.Itoa(k.QuantityReserve),
			strconv.Itoa(k.LoanedCount), strconv.Itoa(k.AvailableCount),
			k.StorageLocation, k.Notes,
		}
	}
	return a.saveCSV("inventaire_cles", headers, rows)
}

// borrowersCSVHeaders est l'ordre des colonnes de l'import/export et du
// modèle des détenteurs (le nom est obligatoire, les autres facultatifs).
var borrowersCSVHeaders = []string{"Nom", "Email", "Statut", "Téléphone"}

// ExportBorrowersCSV exporte la liste des détenteurs.
func (a *App) ExportBorrowersCSV() (string, error) {
	defer a.read()()
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return "", err
	}
	rows := make([][]string, len(borrowers))
	for i, b := range borrowers {
		rows[i] = []string{b.Name, b.Email, borrowerStatusOrDefault(b.Status), b.Phone}
	}
	return a.saveCSV("detenteurs", borrowersCSVHeaders, rows)
}

// borrowerStatusOrDefault retourne le statut, ou « permanent » s'il est vide.
func borrowerStatusOrDefault(s string) string {
	if s == "" {
		return "permanent"
	}
	return s
}

// DownloadBorrowersTemplate génère un modèle CSV à remplir puis à importer.
func (a *App) DownloadBorrowersTemplate() (string, error) {
	defer a.read()()
	rows := [][]string{{"Dupont Jean", "jean.dupont@exemple.fr", "permanent", "0102030405"}}
	return a.saveCSV("modele_detenteurs", borrowersCSVHeaders, rows)
}

// loanCSVRows convertit des prêts en lignes CSV.
func loanCSVRows(loans []LoanDTO) [][]string {
	rows := make([][]string, len(loans))
	for i, l := range loans {
		rows[i] = []string{
			l.KeyNumber, l.KeyDescription, l.BorrowerName, frDate(l.LoanDate),
			frDate(l.PlannedReturnDate), frDate(l.ReturnDate), statusLabel(l.Status),
			l.LoanType, strconv.Itoa(l.Days), l.Condition, l.CreatedBy,
		}
	}
	return rows
}

// loanCSVHeaders est l'en-tête commun des exports de prêts.
var loanCSVHeaders = []string{"Clé", "Description", "Détenteur", "Date remise", "Retour prévu", "Date retour", "Statut", "Type", "Durée (jours)", "État au retour", "Agent"}

// ExportActiveLoansCSV exporte les prêts en cours.
func (a *App) ExportActiveLoansCSV() (string, error) {
	defer a.read()()
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return "", err
	}
	return a.saveCSV("emprunts_en_cours", loanCSVHeaders, loanCSVRows(toLoanDTOs(loans)))
}

// ExportHistoryCSV exporte l'historique filtré.
func (a *App) ExportHistoryCSV(f HistoryFilter) (string, error) {
	defer a.read()()
	filters, err := toLoanFilters(f)
	if err != nil {
		return "", err
	}
	loans, err := db.GetLoanHistory(filters)
	if err != nil {
		return "", err
	}
	return a.saveCSV("historique", loanCSVHeaders, loanCSVRows(toLoanDTOs(loans)))
}

// ImportBorrowersCSV ouvre la sélection d'un fichier CSV (format du modèle)
// et crée les détenteurs. Les lignes invalides sont ignorées et listées dans
// le compte rendu sans faire échouer tout l'import. Annulation → Created = -1.
func (a *App) ImportBorrowersCSV() (CSVImportReport, error) {
	path, err := a.chooseFile("Importer des détenteurs (CSV)", "Fichier CSV (*.csv)", "*.csv")
	if err != nil {
		return CSVImportReport{}, err
	}
	if path == "" {
		return CSVImportReport{Created: -1, Problems: []string{}}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return CSVImportReport{}, fmt.Errorf("lecture du fichier impossible : %w", err)
	}
	rows, problems, err := parseBorrowersCSV(data)
	if err != nil {
		return CSVImportReport{}, err
	}
	defer a.read()()
	report := CSVImportReport{Problems: problems}
	for _, r := range rows {
		b := &db.Borrower{Name: r.Name, Email: r.Email, Status: r.Status, Phone: r.Phone}
		if err := db.CreateBorrower(b); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("ligne %d (%v)", r.line, err))
			continue
		}
		report.Created++
	}
	return report, nil
}

// csvBorrower est une ligne valide du CSV des détenteurs.
type csvBorrower struct {
	line                       int
	Name, Email, Status, Phone string
}

// parseBorrowersCSV lit un CSV de détenteurs : BOM UTF-8 et séparateur
// point-virgule tolérés (format de nos exports), ligne d'en-tête « Nom »
// ignorée, lignes vides sautées. Retourne les lignes valides et la liste des
// lignes rejetées.
func parseBorrowersCSV(data []byte) ([]csvBorrower, []string, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	cr := csv.NewReader(bytes.NewReader(data))
	cr.Comma = ';'
	cr.FieldsPerRecord = -1
	records, err := cr.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("CSV invalide : %w", err)
	}
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("le fichier est vide")
	}
	var rows []csvBorrower
	problems := []string{}
	for i, rec := range records {
		line := i + 1
		cell := func(n int) string {
			if n < len(rec) {
				return strings.TrimSpace(rec[n])
			}
			return ""
		}
		if i == 0 && strings.EqualFold(cell(0), "Nom") {
			continue
		}
		if strings.TrimSpace(strings.Join(rec, "")) == "" {
			continue
		}
		if cell(0) == "" {
			problems = append(problems, fmt.Sprintf("ligne %d (nom manquant)", line))
			continue
		}
		status := strings.ToLower(cell(2))
		if status == "" {
			status = "permanent"
		}
		if !contains(borrowerStatuses, status) {
			problems = append(problems, fmt.Sprintf("ligne %d (statut « %s » inconnu)", line, cell(2)))
			continue
		}
		rows = append(rows, csvBorrower{line: line, Name: cell(0), Email: cell(1), Status: status, Phone: cell(3)})
	}
	return rows, problems, nil
}
