package pdf

import (
	"bytes"
	"fmt"
	"time"

	"clefs/internal/db"

	"github.com/phpdave11/gofpdf"
)

// GenerateLoanReceipt génère un reçu PDF pour un emprunt
func GenerateLoanReceipt(loan *db.LoanWithDetails) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Bon de Sortie de Clé"))
	pdf.Ln(15)

	// Détails de l'emprunt
	pdf.SetFont("Arial", "", 12)

	pdf.Cell(70, 10, tr("Numéro de la clé :"))
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, tr(loan.KeyNumber))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(70, 10, tr("Description :"))
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, tr(loan.KeyDescription))
	pdf.Ln(8)

	pdf.Cell(70, 10, tr("Emprunté par :"))
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, tr(loan.BorrowerName))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(70, 10, tr("Date d'emprunt :"))
	pdf.Cell(0, 10, tr(loan.LoanDate.Format("02/01/2006 à 15:04")))
	pdf.Ln(15)

	// Texte d'engagement
	pdf.SetFont("Arial", "", 11)
	text := fmt.Sprintf("Je soussigné(e), %s, reconnais avoir reçu la clé mentionnée ci-dessus. "+
		"Je m'engage à en prendre soin et à la restituer à la fin de son utilisation. "+
		"En cas de perte ou de dégradation, je suis conscient(e) que ma responsabilité "+
		"pourra être engagée.", loan.BorrowerName)

	pdf.MultiCell(0, 6, tr(text), "", "", false)
	pdf.Ln(20)

	// Signature
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, tr("Signature de l'emprunteur :"))
	pdf.Ln(8)
	pdf.Line(80, pdf.GetY(), 180, pdf.GetY())

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateBorrowerReceipt génère un reçu PDF pour tous les emprunts d'un emprunteur
func GenerateBorrowerReceipt(borrower *db.Borrower, loans []db.LoanWithDetails) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Bon de Sortie de Clés"))
	pdf.Ln(15)

	// Détails de l'emprunteur
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(70, 10, tr("Emprunté par :"))
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, tr(borrower.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(70, 10, tr("Date :"))
	pdf.Cell(0, 10, tr(time.Now().Format("02/01/2006 à 15:04")))
	pdf.Ln(8)

	pdf.Cell(70, 10, tr("Nombre de clés :"))
	pdf.Cell(0, 10, tr(fmt.Sprintf("%d", len(loans))))
	pdf.Ln(12)

	// Ligne de séparation
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(8)

	// Liste des clés
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, tr("Liste des clés empruntées :"))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	for i, loan := range loans {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		text := fmt.Sprintf("%d. %s - %s (%s)",
			i+1,
			loan.KeyNumber,
			loan.KeyDescription,
			loan.LoanDate.Format("02/01/2006"))

		pdf.Cell(0, 7, tr(text))
		pdf.Ln(7)
	}

	pdf.Ln(10)

	// Texte d'engagement
	pdf.SetFont("Arial", "", 11)
	text := fmt.Sprintf("Je soussigné(e), %s, reconnais avoir reçu les %d clé(s) mentionnée(s) ci-dessus. "+
		"Je m'engage à en prendre soin et à les restituer à la fin de leur utilisation. "+
		"En cas de perte ou de dégradation, je suis conscient(e) que ma responsabilité "+
		"pourra être engagée.", borrower.Name, len(loans))

	pdf.MultiCell(0, 6, tr(text), "", "", false)
	pdf.Ln(20)

	// Signature
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, tr("Signature de l'emprunteur :"))
	pdf.Ln(8)
	pdf.Line(80, pdf.GetY(), 180, pdf.GetY())

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateKeyPlanPDF génère un PDF du plan de clés
func GenerateKeyPlanPDF(buildings map[int]db.Building) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Titre
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr("Plan de Clés"))
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Généré le %s", time.Now().Format("02/01/2006 à 15:04"))))
	pdf.Ln(12)

	// Pour chaque bâtiment
	for _, building := range buildings {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		// Nom du bâtiment
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 8, tr(building.Name))
		pdf.Ln(8)

		// Pour chaque salle
		for _, room := range building.Rooms {
			if pdf.GetY() > 260 {
				pdf.AddPage()
			}

			pdf.SetFont("Arial", "B", 11)
			roomText := fmt.Sprintf("  %s", room.Name)
			if room.Type != "" {
				roomText += fmt.Sprintf(" (%s)", room.Type)
			}
			pdf.Cell(0, 6, tr(roomText))
			pdf.Ln(6)

			// Clés associées
			if len(room.Keys) > 0 {
				pdf.SetFont("Arial", "", 10)
				for _, key := range room.Keys {
					pdf.Cell(10, 5, "")
					keyText := fmt.Sprintf("• Clé %s - %s", key.Number, key.Description)
					pdf.Cell(0, 5, tr(keyText))
					pdf.Ln(5)
				}
			} else {
				pdf.SetFont("Arial", "I", 10)
				pdf.Cell(10, 5, "")
				pdf.Cell(0, 5, tr("Aucune clé associée"))
				pdf.Ln(5)
			}
			pdf.Ln(3)
		}
		pdf.Ln(5)
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

	// En-têtes du tableau
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(30, 7, tr("Clé"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 7, tr("Description"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 7, tr("Emprunteur"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 7, tr("Date"), "1", 0, "C", false, 0, "")
	pdf.Ln(7)

	// Données
	pdf.SetFont("Arial", "", 9)
	for _, loan := range loans {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Répéter les en-têtes
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(30, 7, tr("Clé"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(60, 7, tr("Description"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(50, 7, tr("Emprunteur"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 7, tr("Date"), "1", 0, "C", false, 0, "")
			pdf.Ln(7)
			pdf.SetFont("Arial", "", 9)
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
