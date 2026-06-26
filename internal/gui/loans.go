package gui

import (
	"clefs/internal/db"
	"clefs/internal/export"
	"clefs/internal/pdf"
	"fmt"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createActiveLoansView crée la vue des emprunts actifs
func createActiveLoansView(app *App) fyne.CanvasObject {
	csvBtn := widget.NewButton("📊 Exporter CSV", func() { exportLoansCSV(app) })
	title := widget.NewLabelWithStyle("Emprunts en Cours", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewBorder(nil, nil, title, csvBtn)

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
	borrowerNames := make([]string, 0, len(loansByBorrower))
	for name := range loansByBorrower {
		borrowerNames = append(borrowerNames, name)
	}
	sort.Strings(borrowerNames)

	// Créer la liste avec accordéon
	loansList := container.NewVBox()

	if len(loansByBorrower) == 0 {
		emptyCard := widget.NewCard("", "Aucun emprunt actif",
			widget.NewLabel("Il n'y a actuellement aucune clé empruntée."))
		loansList.Add(emptyCard)
	} else {
		for _, borrowerName := range borrowerNames {
			currentLoans := make([]db.LoanWithDetails, len(loansByBorrower[borrowerName]))
			copy(currentLoans, loansByBorrower[borrowerName])
			accordion := createBorrowerAccordion(app, borrowerName, currentLoans)
			loansList.Add(accordion)
			loansList.Add(widget.NewLabel(""))
		}
	}

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(loansList),
	)

	return content
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
			showReturnWithConditionDialog(app, l, func() { app.showActiveLoans() })
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

// exportLoansCSV exporte les emprunts en cours en CSV dans documents/.
func exportLoansCSV(app *App) {
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération: %v", err))
		return
	}
	headers := []string{"Clé", "Description", "Détenteur", "Date remise", "Retour prévu", "Durée (jours)", "Agent"}
	rows := make([][]string, len(loans))
	for i, l := range loans {
		planned := ""
		if l.PlannedReturnDate != nil {
			planned = l.PlannedReturnDate.Format("02/01/2006")
		}
		days := fmt.Sprintf("%d", int(time.Since(l.LoanDate).Hours()/24))
		rows[i] = []string{l.KeyNumber, l.KeyDescription, l.BorrowerName, l.LoanDate.Format("02/01/2006"), planned, days, l.CreatedBy}
	}
	filePath, err := export.SaveCSV(export.Filename("emprunts_en_cours"), headers, rows)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'export: %v", err))
		return
	}
	app.showSuccess(fmt.Sprintf("✅ Export CSV enregistré : %s", filePath))
}

// showReturnWithConditionDialog affiche un dialogue de retour avec saisie de l'état constaté.
func showReturnWithConditionDialog(app *App, loan db.LoanWithDetails, onDone func()) {
	conditionEntry := widget.NewEntry()
	conditionEntry.SetPlaceHolder("État constaté (optionnel) : bon état, rayé, etc.")

	var popup *widget.PopUp
	confirmBtn := widget.NewButton("Confirmer le retour", func() {
		if err := db.ReturnLoanWithCondition(loan.ID, conditionEntry.Text); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(popup)
		app.showSuccess(fmt.Sprintf("Clé %s retournée par %s.", loan.KeyNumber, loan.BorrowerName))
		onDone()
	})
	confirmBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popup)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Retour de clé", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Clé %s — empruntée par %s", loan.KeyNumber, loan.BorrowerName)),
		widget.NewSeparator(),
		widget.NewLabel("État constaté au retour :"),
		conditionEntry,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, confirmBtn),
	)
	popup = widget.NewModalPopUp(container.NewPadded(content), app.window.Canvas())
	popup.Show()
}
