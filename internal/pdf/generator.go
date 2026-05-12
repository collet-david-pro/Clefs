package pdf

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"clefs/internal/db"

	"github.com/phpdave11/gofpdf"
)

// PDFTableColumn définit une colonne pour writePDFTableHeader.
type PDFTableColumn struct {
	Header string
	Width  float64
	Align  string
}

// writePDFTableHeader écrit une ligne d'en-tête de tableau et remet la police en corps normal.
func writePDFTableHeader(p *gofpdf.Fpdf, tr func(string) string, columns []PDFTableColumn) {
	p.SetFont("Arial", "B", 10)
	p.SetFillColor(200, 220, 255)
	for _, col := range columns {
		p.CellFormat(col.Width, 8, tr(col.Header), "1", 0, col.Align, true, 0, "")
	}
	p.Ln(8)
	p.SetFont("Arial", "", 9)
}

// pdfField écrit une paire libellé / valeur sur une ligne.
func pdfField(p *gofpdf.Fpdf, tr func(string) string, label, value string) {
	p.SetFont("Arial", "", 12)
	p.Cell(60, 8, tr(label))
	p.SetFont("Arial", "B", 12)
	p.Cell(0, 8, tr(value))
	p.Ln(8)
}

// pdfOutput finalise le document et retourne les octets.
func pdfOutput(p *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// pdfEngagement écrit le texte d'engagement et la zone de signature.
func pdfEngagement(p *gofpdf.Fpdf, tr func(string) string, name string, nKeys int) {
	p.SetFont("Arial", "", 11)
	suffix := "la clé mentionnée ci-dessus"
	pronom := "en prendre soin et à la restituer"
	if nKeys > 1 {
		suffix = fmt.Sprintf("les %d clés mentionnées ci-dessus", nKeys)
		pronom = "en prendre soin et à les restituer"
	}
	text := fmt.Sprintf("Je soussigné(e), %s, reconnais avoir reçu %s. "+
		"Je m'engage à %s à la fin de leur utilisation. "+
		"En cas de perte ou de dégradation, je suis conscient(e) que ma responsabilité pourra être engagée.",
		name, suffix, pronom)
	p.MultiCell(0, 6, tr(text), "", "", false)
	p.Ln(15)
	p.SetFont("Arial", "", 12)
	p.Cell(0, 8, tr("Signature de l'emprunteur :"))
	p.Ln(8)
	p.Line(80, p.GetY(), 180, p.GetY())
}

// GenerateLoanReceipt génère un reçu PDF pour un emprunt unique.
func GenerateLoanReceipt(loan *db.LoanWithDetails) ([]byte, error) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	tr := p.UnicodeTranslatorFromDescriptor("")

	p.SetFont("Arial", "B", 16)
	p.Cell(0, 10, tr("BON DE REMISE DE CLÉ"))
	p.Ln(12)
	p.SetFont("Arial", "", 9)
	p.Cell(0, 6, tr("Collège Victor Hugo — Chauny"))
	p.Ln(12)

	pdfField(p, tr, "Détenteur :", loan.BorrowerName)
	pdfField(p, tr, "Date de remise :", loan.LoanDate.Format("02/01/2006 à 15:04"))
	if loan.PlannedReturnDate != nil {
		pdfField(p, tr, "Retour prévu :", loan.PlannedReturnDate.Format("02/01/2006"))
	}
	if loan.CreatedBy != "" {
		pdfField(p, tr, "Agent :", loan.CreatedBy)
	}
	p.Ln(4)

	writePDFTableHeader(p, tr, []PDFTableColumn{
		{"N° clé", 35, "L"},
		{"Description", 155, "L"},
	})
	p.SetFont("Arial", "", 10)
	p.CellFormat(35, 6, tr(loan.KeyNumber), "1", 0, "L", false, 0, "")
	p.CellFormat(155, 6, tr(loan.KeyDescription), "1", 1, "L", false, 0, "")
	p.Ln(8)

	pdfEngagement(p, tr, loan.BorrowerName, 1)
	return pdfOutput(p)
}

// BorrowerReceiptOptions contient les données optionnelles V3 pour le bon de remise enrichi.
type BorrowerReceiptOptions struct {
	Agent         string
	PlannedReturn *time.Time
	LoanType      string
	Accesses      []db.Room // accès couverts par les clés remises
}

// GenerateBorrowerReceipt génère un bon de remise enrichi pour un détenteur.
func GenerateBorrowerReceipt(borrower *db.Borrower, loans []db.LoanWithDetails, opts ...BorrowerReceiptOptions) ([]byte, error) {
	var opt BorrowerReceiptOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	tr := p.UnicodeTranslatorFromDescriptor("")

	// En-tête
	p.SetFont("Arial", "B", 16)
	p.Cell(0, 10, tr("BON DE REMISE DE CLÉ(S)"))
	p.Ln(12)
	p.SetFont("Arial", "", 9)
	p.Cell(0, 6, tr("Collège Victor Hugo — Chauny"))
	p.Ln(12)

	// Informations détenteur
	loanType := opt.LoanType
	if loanType == "" {
		loanType = "ponctuel"
	}
	pdfField(p, tr, "Détenteur :", borrower.Name)
	pdfField(p, tr, "Date de remise :", time.Now().Format("02/01/2006 à 15:04"))
	pdfField(p, tr, "Type de prêt :", loanType)
	if opt.PlannedReturn != nil {
		pdfField(p, tr, "Retour prévu :", opt.PlannedReturn.Format("02/01/2006"))
	}
	if opt.Agent != "" {
		pdfField(p, tr, "Agent :", opt.Agent)
	}
	p.Ln(4)

	// Tableau des clés remises
	p.SetFont("Arial", "B", 11)
	p.Cell(0, 8, tr("Clés remises :"))
	p.Ln(8)
	writePDFTableHeader(p, tr, []PDFTableColumn{
		{"N° clé", 35, "L"},
		{"Description", 100, "L"},
		{"Date remise", 40, "C"},
	})
	p.SetFont("Arial", "", 10)
	for _, loan := range loans {
		if p.GetY() > 260 {
			p.AddPage()
			writePDFTableHeader(p, tr, []PDFTableColumn{
				{"N° clé", 35, "L"},
				{"Description", 100, "L"},
				{"Date remise", 40, "C"},
			})
		}
		desc := loan.KeyDescription
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		p.CellFormat(35, 6, tr(loan.KeyNumber), "1", 0, "L", false, 0, "")
		p.CellFormat(100, 6, tr(desc), "1", 0, "L", false, 0, "")
		p.CellFormat(40, 6, tr(loan.LoanDate.Format("02/01/2006")), "1", 1, "C", false, 0, "")
	}
	p.Ln(6)

	// Tableau des accès couverts (si fournis)
	if len(opt.Accesses) > 0 {
		if p.GetY() > 220 {
			p.AddPage()
		}
		p.SetFont("Arial", "B", 11)
		p.Cell(0, 8, tr("Accès couverts par ces clés :"))
		p.Ln(8)
		writePDFTableHeader(p, tr, []PDFTableColumn{
			{"Désignation", 100, "L"},
			{"Bâtiment", 55, "L"},
			{"Étage", 35, "C"},
		})
		p.SetFont("Arial", "", 10)
		for _, acc := range opt.Accesses {
			if p.GetY() > 270 {
				p.AddPage()
				writePDFTableHeader(p, tr, []PDFTableColumn{
					{"Désignation", 100, "L"},
					{"Bâtiment", 55, "L"},
					{"Étage", 35, "C"},
				})
			}
			p.CellFormat(100, 6, tr(acc.Name), "1", 0, "L", false, 0, "")
			p.CellFormat(55, 6, tr(acc.Notes), "1", 0, "L", false, 0, "") // building name passed via Notes field by caller
			p.CellFormat(35, 6, tr(acc.Floor), "1", 1, "C", false, 0, "")
		}
		p.Ln(6)
	}

	pdfEngagement(p, tr, borrower.Name, len(loans))
	return pdfOutput(p)
}

// GenerateKeyPlanPDF génère un PDF du plan de clés (Compact et Trié)
func GenerateKeyPlanPDF(buildingsMap map[int]db.Building) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr("Plan de Clés"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Généré le %s", time.Now().Format("02/01/2006 à 15:04"))))
	pdf.Ln(10)

	// Convertir la map en slice pour le tri
	var buildings []db.Building
	for _, b := range buildingsMap {
		buildings = append(buildings, b)
	}

	// Trier les bâtiments par nom
	sort.Slice(buildings, func(i, j int) bool {
		return strings.ToLower(buildings[i].Name) < strings.ToLower(buildings[j].Name)
	})

	// Pour chaque bâtiment
	for _, building := range buildings {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		// Nom du bâtiment
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(0, 7, tr(building.Name), "1", 1, "L", true, 0, "")

		if len(building.Rooms) == 0 {
			pdf.SetFont("Arial", "I", 9)
			pdf.Cell(0, 6, tr("  (Aucune salle)"))
			pdf.Ln(6)
		} else {
			// Trier les salles par nom
			sort.Slice(building.Rooms, func(i, j int) bool {
				return strings.ToLower(building.Rooms[i].Name) < strings.ToLower(building.Rooms[j].Name)
			})

			// Pour chaque salle
			for _, room := range building.Rooms {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}

				// Salle
				pdf.SetFont("Arial", "B", 10)
				roomText := fmt.Sprintf("• %s", room.Name)
				if room.Type != "" {
					roomText += fmt.Sprintf(" (%s)", room.Type)
				}
				pdf.Cell(80, 6, tr(roomText))

				// Clés associées (sur la même ligne si possible)
				if len(room.Keys) > 0 {
					// Trier les clés
					sort.Slice(room.Keys, func(i, j int) bool {
						return room.Keys[i].Number < room.Keys[j].Number
					})

					var keyTexts []string
					for _, key := range room.Keys {
						keyTexts = append(keyTexts, key.Number)
					}
					keysString := strings.Join(keyTexts, ", ")

					pdf.SetFont("Arial", "", 9)
					// Si la liste est trop longue, on la met à la ligne
					if len(keysString) > 60 {
						pdf.Ln(5)
						pdf.Cell(10, 5, "") // Indentation
						pdf.MultiCell(0, 5, tr("Clés : "+keysString), "", "L", false)
					} else {
						pdf.Cell(0, 6, tr(": "+keysString))
						pdf.Ln(6)
					}
				} else {
					pdf.SetFont("Arial", "I", 9)
					pdf.Cell(0, 6, tr(": Aucune clé"))
					pdf.Ln(6)
				}
			}
		}
		pdf.Ln(4) // Petit espace après chaque bâtiment
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateLoansReportPDF génère un rapport PDF des emprunts actifs
func GenerateLoansReportPDF(loans []db.LoanWithDetails) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Rapport des Clés Sorties"))
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Généré le %s", time.Now().Format("02/01/2006 à 15:04"))))
	pdf.Ln(8)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Nombre total d'emprunts actifs : %d", len(loans))))
	pdf.Ln(12)

	loansReportCols := []PDFTableColumn{
		{"Clé", 30, "C"}, {"Description", 60, "C"}, {"Emprunteur", 50, "C"}, {"Date", 40, "C"},
	}
	writePDFTableHeader(pdf, tr, loansReportCols)

	for _, loan := range loans {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			writePDFTableHeader(pdf, tr, loansReportCols)
		}

		pdf.CellFormat(30, 6, tr(loan.KeyNumber), "1", 0, "L", false, 0, "")

		// Tronquer la description si trop longue
		desc := loan.KeyDescription
		if len(desc) > 35 {
			desc = desc[:32] + "..."
		}
		pdf.CellFormat(60, 6, tr(desc), "1", 0, "L", false, 0, "")

		// Tronquer le nom si trop long
		name := loan.BorrowerName
		if len(name) > 25 {
			name = name[:22] + "..."
		}
		pdf.CellFormat(50, 6, tr(name), "1", 0, "L", false, 0, "")

		pdf.CellFormat(40, 6, tr(loan.LoanDate.Format("02/01/2006")), "1", 0, "C", false, 0, "")
		pdf.Ln(6)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateGlobalBorrowerReport génère un rapport PDF global groupé par emprunteur
func GenerateGlobalBorrowerReport(loansByBorrower map[string][]db.LoanWithDetails) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Rapport Global des Emprunts"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Généré le %s", time.Now().Format("02/01/2006 à 15:04"))))
	pdf.Ln(15)

	// Calculer le total
	totalLoans := 0
	for _, loans := range loansByBorrower {
		totalLoans += len(loans)
	}
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, tr(fmt.Sprintf("Total : %d emprunteurs, %d clés sorties", len(loansByBorrower), totalLoans)))
	pdf.Ln(12)

	// Pour chaque emprunteur (on pourrait trier les clés ici pour l'ordre alphabétique)
	// Note: Dans une map, l'ordre est aléatoire. Pour la production, il vaudrait mieux trier.

	pdf.SetFillColor(240, 240, 240)

	for borrower, loans := range loansByBorrower {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		// En-tête Emprunteur
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(230, 230, 250) // Lavande clair
		pdf.CellFormat(0, 10, tr(fmt.Sprintf("  %s (%d clés)", borrower, len(loans))), "1", 1, "L", true, 0, "")

		// Détails des clés
		pdf.SetFont("Arial", "", 10)

		// En-têtes colonnes
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(30, 7, tr("Clé"), "L", 0, "C", false, 0, "")
		pdf.CellFormat(90, 7, tr("Description"), "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, tr("Date d'emprunt"), "", 0, "C", false, 0, "")
		pdf.CellFormat(30, 7, tr("Durée"), "R", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 10)
		for _, loan := range loans {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}

			days := int(time.Since(loan.LoanDate).Hours() / 24)
			duration := fmt.Sprintf("%d jours", days)
			if days == 0 {
				duration = "Aujourd'hui"
			}

			pdf.CellFormat(30, 6, tr(loan.KeyNumber), "L", 0, "C", false, 0, "")

			// Tronquer description
			desc := loan.KeyDescription
			if len(desc) > 45 {
				desc = desc[:42] + "..."
			}
			pdf.CellFormat(90, 6, tr(desc), "", 0, "L", false, 0, "")
			pdf.CellFormat(40, 6, tr(loan.LoanDate.Format("02/01/2006")), "", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, tr(duration), "R", 1, "C", false, 0, "")
		}

		// Ligne de séparation bas de section
		pdf.CellFormat(0, 1, "", "T", 1, "", false, 0, "")
		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateKeyStockReport génère un bilan PDF du stock de clés
func GenerateKeyStockReport(keys []db.Key, loanCounts map[int]int) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Bilan du Stock de Clés"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Généré le %s", time.Now().Format("02/01/2006 à 15:04"))))
	pdf.Ln(15)

	stockCols := []PDFTableColumn{
		{"Numéro", 25, "C"}, {"Description", 75, "C"}, {"Total", 20, "C"},
		{"Réserve", 20, "C"}, {"Sorties", 25, "C"}, {"Dispo", 25, "C"},
	}
	writePDFTableHeader(pdf, tr, stockCols)

	for _, key := range keys {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			writePDFTableHeader(pdf, tr, stockCols)
		}

		borrowed := loanCounts[key.ID]
		available := key.QuantityTotal - key.QuantityReserve - borrowed

		// Alerte stock bas
		fill := false
		if available <= 0 {
			pdf.SetFillColor(255, 200, 200) // Rouge clair
			fill = true
		} else if available == 1 {
			pdf.SetFillColor(255, 240, 200) // Orange clair
			fill = true
		}

		pdf.CellFormat(25, 6, tr(key.Number), "1", 0, "L", fill, 0, "")

		desc := key.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		pdf.CellFormat(75, 6, tr(desc), "1", 0, "L", fill, 0, "")

		pdf.CellFormat(20, 6, fmt.Sprintf("%d", key.QuantityTotal), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", key.QuantityReserve), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", borrowed), "1", 0, "C", fill, 0, "")

		// Gras pour la disponibilité
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", available), "1", 1, "C", fill, 0, "")
		pdf.SetFont("Arial", "", 9)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
