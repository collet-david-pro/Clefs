package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// createActiveLoansView crée la vue des emprunts actifs
func createActiveLoansView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Emprunts en Cours", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Récupérer les emprunts actifs
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return container.NewVBox(
			title,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Grouper par emprunteur
	loansByBorrower := make(map[string][]db.LoanWithDetails)
	for _, loan := range loans {
		loansByBorrower[loan.BorrowerName] = append(loansByBorrower[loan.BorrowerName], loan)
	}

	// Créer la liste
	loansList := container.NewVBox()

	if len(loansByBorrower) == 0 {
		loansList.Add(widget.NewLabel("Aucun emprunt actif"))
	} else {
		for borrowerName, borrowerLoans := range loansByBorrower {
			// En-tête de l'emprunteur
			borrowerLabel := widget.NewLabelWithStyle(
				fmt.Sprintf("%s (%d clé(s))", borrowerName, len(borrowerLoans)),
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			)

			// Bouton pour générer le reçu
			receiptBtn := widget.NewButton("📄 Reçu", func() {
				generateBorrowerReceiptFromLoans(app, borrowerLoans)
			})

			borrowerHeader := container.NewBorder(nil, nil, nil, receiptBtn, borrowerLabel)
			loansList.Add(borrowerHeader)

			// Liste des clés empruntées
			for _, loan := range borrowerLoans {
				l := loan // Capture

				loanText := fmt.Sprintf("  • Clé %s - %s (depuis le %s)",
					l.KeyNumber,
					l.KeyDescription,
					l.LoanDate.Format("02/01/2006"))

				loanLabel := widget.NewLabel(loanText)

				returnBtn := widget.NewButton("Retourner", func() {
					app.showConfirm("Confirmer le retour",
						fmt.Sprintf("Confirmer le retour de la clé %s?", l.KeyNumber),
						func() {
							err := db.ReturnLoan(l.ID)
							if err != nil {
								app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
								return
							}
							app.showSuccess("Clé retournée avec succès!")
							app.showActiveLoans()
						})
				})
				returnBtn.Importance = widget.MediumImportance

				loanRow := container.NewBorder(nil, nil, nil, returnBtn, loanLabel)
				loansList.Add(loanRow)
			}

			loansList.Add(widget.NewSeparator())
		}
	}

	content := container.NewBorder(
		title,
		nil,
		nil,
		nil,
		container.NewVScroll(loansList),
	)

	return content
}

// createLoansReportView crée la vue du rapport des emprunts
func createLoansReportView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Rapport des Clés Sorties", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Bouton pour exporter en PDF
	exportBtn := widget.NewButton("📄 Exporter en PDF", func() {
		exportLoansReportPDF(app)
	})
	exportBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, exportBtn, title)

	// Récupérer les emprunts actifs
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Informations générales
	infoLabel := widget.NewLabel(fmt.Sprintf("Généré le %s | Total: %d emprunt(s) actif(s)",
		time.Now().Format("02/01/2006 à 15:04"),
		len(loans)))

	// Créer le tableau
	reportTable := createLoansReportTable(loans)

	content := container.NewBorder(
		container.NewVBox(header, infoLabel, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewVScroll(reportTable),
	)

	return content
}

// createLoansReportTable crée le tableau du rapport
func createLoansReportTable(loans []db.LoanWithDetails) fyne.CanvasObject {
	// En-têtes
	headers := container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Clé", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Description", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Emprunteur", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Date", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Lignes
	rows := container.NewVBox()
	for _, loan := range loans {
		row := container.NewGridWithColumns(4,
			widget.NewLabel(loan.KeyNumber),
			widget.NewLabel(loan.KeyDescription),
			widget.NewLabel(loan.BorrowerName),
			widget.NewLabel(loan.LoanDate.Format("02/01/2006")),
		)
		rows.Add(row)
		rows.Add(widget.NewSeparator())
	}

	return container.NewVBox(headers, widget.NewSeparator(), rows)
}

// exportLoansReportPDF exporte le rapport en PDF
func exportLoansReportPDF(app *App) {
	// Récupérer les emprunts actifs
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunts: %v", err))
		return
	}

	if len(loans) == 0 {
		app.showError("Aucun emprunt", "Aucun emprunt actif à exporter.")
		return
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateLoansReportPDF(loans)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Sauvegarder le fichier
	filename := fmt.Sprintf("rapport_cles_sorties_%s.pdf", time.Now().Format("20060102"))

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write(pdfData)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de l'écriture du fichier: %v", err))
			return
		}

		app.showSuccess("Rapport PDF généré avec succès!")
	}, app.window)

	saveDialog.SetFileName(filename)
	saveDialog.Show()
}

// generateBorrowerReceiptFromLoans génère un reçu pour un emprunteur à partir de ses emprunts
func generateBorrowerReceiptFromLoans(app *App, loans []db.LoanWithDetails) {
	if len(loans) == 0 {
		return
	}

	// Récupérer l'emprunteur
	borrower, err := db.GetBorrowerByID(loans[0].BorrowerID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération de l'emprunteur: %v", err))
		return
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateBorrowerReceipt(borrower, loans)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Sauvegarder le fichier
	filename := fmt.Sprintf("bon_de_sortie_cles_%s_%s.pdf",
		borrower.Name,
		time.Now().Format("20060102"))

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write(pdfData)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de l'écriture du fichier: %v", err))
			return
		}

		app.showSuccess("Reçu PDF généré avec succès!")
	}, app.window)

	saveDialog.SetFileName(filename)
	saveDialog.Show()
}
