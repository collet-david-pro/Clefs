package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	// Créer la liste avec accordéon
	loansList := container.NewVBox()

	if len(loansByBorrower) == 0 {
		emptyCard := widget.NewCard("", "Aucun emprunt actif",
			widget.NewLabel("Il n'y a actuellement aucune clé empruntée."))
		loansList.Add(emptyCard)
	} else {
		for borrowerName, borrowerLoans := range loansByBorrower {
			// Créer une copie locale pour éviter les problèmes de closure
			currentLoans := make([]db.LoanWithDetails, len(borrowerLoans))
			copy(currentLoans, borrowerLoans)

			// Créer l'accordéon pour cet emprunteur
			accordion := createBorrowerAccordion(app, borrowerName, currentLoans)
			loansList.Add(accordion)
			loansList.Add(widget.NewLabel("")) // Espacement
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

	// Boutons d'action
	loansReportBtn := widget.NewButton("📊 Générer Rapport des Clés Sorties", func() {
		generateLoansReportPDF(app)
	})
	loansReportBtn.Importance = widget.HighImportance

	globalReportBtn := widget.NewButton("📄 Générer Rapport Global par Emprunteur", func() {
		generateGlobalBorrowerReportPDF(app)
	})

	buttonsContainer := container.NewHBox(loansReportBtn, globalReportBtn)

	header := container.NewBorder(nil, nil, nil, buttonsContainer, title)

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

	// Créer l'affichage groupé par clé
	reportContent := createLoansReportByKey(loans, app)

	content := container.NewBorder(
		container.NewVBox(header, infoLabel, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewVScroll(reportContent),
	)

	return content
}

// createLoansReportByKey crée l'affichage groupé par clé avec accordéon
func createLoansReportByKey(loans []db.LoanWithDetails, app *App) fyne.CanvasObject {
	// Grouper par clé
	loansByKey := make(map[string][]db.LoanWithDetails)
	keyInfo := make(map[string]string) // Pour stocker la description de chaque clé

	for _, loan := range loans {
		loansByKey[loan.KeyNumber] = append(loansByKey[loan.KeyNumber], loan)
		keyInfo[loan.KeyNumber] = loan.KeyDescription
	}

	// Créer la liste avec accordéons
	list := container.NewVBox()

	if len(loansByKey) == 0 {
		emptyCard := widget.NewCard("", "Aucun emprunt actif",
			widget.NewLabel("Il n'y a actuellement aucune clé empruntée."))
		list.Add(emptyCard)
	} else {
		for keyNumber, keyLoans := range loansByKey {
			// Créer une copie locale
			currentLoans := make([]db.LoanWithDetails, len(keyLoans))
			copy(currentLoans, keyLoans)
			currentKeyNumber := keyNumber
			currentKeyDesc := keyInfo[keyNumber]

			// Créer l'accordéon pour cette clé
			accordion := createKeyLoansAccordion(app, currentKeyNumber, currentKeyDesc, currentLoans)
			list.Add(accordion)
			list.Add(widget.NewLabel("")) // Espacement
		}
	}

	return list
}

// createKeyLoansAccordion crée un accordéon pour une clé dans le rapport
func createKeyLoansAccordion(app *App, keyNumber string, keyDesc string, loans []db.LoanWithDetails) *widget.Accordion {
	// Créer le contenu détaillé
	detailsContent := container.NewVBox()

	// Informations de la clé
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("📝 %s", keyDesc)))
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("📊 %d emprunt(s) actif(s)", len(loans))))
	detailsContent.Add(widget.NewSeparator())

	// Liste des emprunteurs
	for _, loan := range loans {
		l := loan // Capture

		// Calculer la durée
		days := int(time.Since(l.LoanDate).Hours() / 24)
		durationText := fmt.Sprintf("%d jour(s)", days)
		if days == 0 {
			durationText = "Aujourd'hui"
		}

		borrowerInfo := container.NewVBox(
			widget.NewLabelWithStyle(
				fmt.Sprintf("👤 %s", l.BorrowerName),
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			),
			widget.NewLabel(fmt.Sprintf("   📅 Emprunté le: %s (%s)",
				l.LoanDate.Format("02/01/2006"), durationText)),
		)

		returnBtn := widget.NewButton("↩️ Retourner", func() {
			app.showConfirm("Confirmer le retour",
				fmt.Sprintf("Confirmer le retour de la clé %s empruntée par %s?", l.KeyNumber, l.BorrowerName),
				func() {
					err := db.ReturnLoan(l.ID)
					if err != nil {
						app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
						return
					}
					app.showSuccess("Clé retournée avec succès!")
					app.showLoansReport()
				})
		})
		returnBtn.Importance = widget.MediumImportance

		borrowerRow := container.NewBorder(nil, nil, nil, returnBtn, borrowerInfo)
		detailsContent.Add(borrowerRow)
		detailsContent.Add(widget.NewSeparator())
	}

	// Créer l'item d'accordéon
	title := fmt.Sprintf("🔑 %s - %d emprunteur(s)", keyNumber, len(loans))

	accordionItem := widget.NewAccordionItem(title, detailsContent)

	// Créer l'accordéon
	accordion := widget.NewAccordion(accordionItem)

	return accordion
}

// generateLoansReportPDF génère et enregistre le rapport des clés sorties
func generateLoansReportPDF(app *App) {
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

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename("rapport_cles_sorties", 0)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Rapport enregistré : %s", filepath))
}

// generateBorrowerReceiptPDF génère et enregistre un reçu groupé pour un emprunteur
func generateBorrowerReceiptPDF(app *App, loans []db.LoanWithDetails) {
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

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename(fmt.Sprintf("recu_emprunteur_%s", borrower.Name), 0)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Reçu enregistré : %s", filepath))
}

// generateGlobalBorrowerReportPDF génère et enregistre le rapport global par emprunteur
func generateGlobalBorrowerReportPDF(app *App) {
	// Récupérer les emprunts actifs
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunts: %v", err))
		return
	}

	if len(loans) == 0 {
		app.showError("Aucun emprunt", "Aucun emprunt actif à afficher.")
		return
	}

	// Grouper par emprunteur
	loansByBorrower := make(map[string][]db.LoanWithDetails)
	for _, loan := range loans {
		loansByBorrower[loan.BorrowerName] = append(loansByBorrower[loan.BorrowerName], loan)
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateGlobalBorrowerReport(loansByBorrower)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename("rapport_global_emprunteurs", 0)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Rapport enregistré : %s", filepath))
}

// createBorrowerAccordion crée un accordéon pour un emprunteur
func createBorrowerAccordion(app *App, borrowerName string, loans []db.LoanWithDetails) *widget.Accordion {
	// Créer le contenu détaillé (qui sera caché/affiché)
	detailsContent := container.NewVBox()

	// Liste des clés avec détails
	for _, loan := range loans {
		l := loan // Capture

		// Calculer la durée
		days := int(time.Since(l.LoanDate).Hours() / 24)
		durationText := fmt.Sprintf("%d jour(s)", days)
		if days == 0 {
			durationText = "Aujourd'hui"
		}

		keyInfo := container.NewVBox(
			widget.NewLabelWithStyle(
				fmt.Sprintf("🔑 %s", l.KeyNumber),
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			),
			widget.NewLabel(fmt.Sprintf("   %s", l.KeyDescription)),
			widget.NewLabel(fmt.Sprintf("   📅 Emprunté le: %s (%s)",
				l.LoanDate.Format("02/01/2006"), durationText)),
		)

		returnBtn := widget.NewButton("↩️ Retourner", func() {
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

		keyRow := container.NewBorder(nil, nil, nil, returnBtn, keyInfo)
		detailsContent.Add(keyRow)
		detailsContent.Add(widget.NewSeparator())
	}

	// Bouton pour générer le reçu groupé
	generateReceiptBtn := widget.NewButton("📄 Générer PDF du Reçu", func() {
		generateBorrowerReceiptPDF(app, loans)
	})
	generateReceiptBtn.Importance = widget.HighImportance
	detailsContent.Add(generateReceiptBtn)

	// Créer l'item d'accordéon
	accordionItem := widget.NewAccordionItem(
		fmt.Sprintf("👤 %s - %d clé(s)", borrowerName, len(loans)),
		detailsContent,
	)

	// Créer l'accordéon
	accordion := widget.NewAccordion(accordionItem)

	return accordion
}
