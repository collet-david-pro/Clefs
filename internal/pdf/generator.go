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

	// En-têtes du tableau
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 220, 255)

	pdf.CellFormat(25, 8, tr("Numéro"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(75, 8, tr("Description"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, tr("Total"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, tr("Réserve"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, tr("Sorties"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, tr("Dispo"), "1", 1, "C", true, 0, "")

	// Données
	pdf.SetFont("Arial", "", 9)

	for _, key := range keys {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Répéter en-têtes
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(200, 220, 255)
			pdf.CellFormat(25, 8, tr("Numéro"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(75, 8, tr("Description"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, tr("Total"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, tr("Réserve"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, tr("Sorties"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, tr("Dispo"), "1", 1, "C", true, 0, "")
			pdf.SetFont("Arial", "", 9)
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
