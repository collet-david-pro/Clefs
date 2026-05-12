package gui

import (
	"clefs/internal/db"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// showNewLoanDialog affiche le dialogue de création d'un nouvel emprunt
func showNewLoanDialog(app *App) {
	showNewLoanDialogWithKey(app, 0)
}

// showNewLoanDialogWithKey affiche le dialogue de création d'un emprunt pour une clé donnée (0 = aucune présélection)
func showNewLoanDialogWithKey(app *App, preselectedKeyID int) {
	borrowers, err := db.GetAllBorrowers()
	if err != nil || len(borrowers) == 0 {
		app.showError("Erreur", "Impossible de charger les emprunteurs. Veuillez d'abord ajouter des emprunteurs.")
		return
	}

	keys, err := db.GetAvailableKeys()
	if err != nil || len(keys) == 0 {
		app.showError("Aucune clé disponible", "Il n'y a aucune clé disponible pour un emprunt.")
		return
	}

	// Construire les options emprunteurs
	borrowerNames := make([]string, len(borrowers))
	for i, b := range borrowers {
		borrowerNames[i] = b.Name
	}

	// Construire les options clés
	keyOptions := make([]string, len(keys))
	for i, k := range keys {
		keyOptions[i] = fmt.Sprintf("%s — %s", k.Number, k.Description)
	}

	borrowerSelect := widget.NewSelect(borrowerNames, nil)
	keySelect := widget.NewSelect(keyOptions, nil)

	// Présélectionner la clé si fournie
	if preselectedKeyID > 0 {
		for i, k := range keys {
			if k.ID == preselectedKeyID {
				keySelect.SetSelectedIndex(i)
				break
			}
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Emprunteur", borrowerSelect),
		widget.NewFormItem("Clé", keySelect),
	)

	var popup *widget.PopUp
	confirmBtn := widget.NewButton("Valider", func() {
		if borrowerSelect.SelectedIndex() < 0 {
			app.showError("Erreur", "Veuillez sélectionner un emprunteur.")
			return
		}
		if keySelect.SelectedIndex() < 0 {
			app.showError("Erreur", "Veuillez sélectionner une clé.")
			return
		}
		borrower := borrowers[borrowerSelect.SelectedIndex()]
		key := keys[keySelect.SelectedIndex()]

		if err := db.CreateLoan(key.ID, borrower.ID); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création du prêt: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(popup)
		app.showSuccess(fmt.Sprintf("Clé %s remise à %s avec succès.", key.Number, borrower.Name))
		app.showDashboard()
	})
	confirmBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popup)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Nouvel Emprunt", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		container.NewHBox(cancelBtn, confirmBtn),
	)

	popup = widget.NewModalPopUp(container.NewPadded(content), app.window.Canvas())
	popup.Show()
}

// showReturnDialog affiche le dialogue de retour des clés empruntées pour une clé donnée
func showReturnDialog(app *App, keyID int) {
	loans, err := db.GetActiveLoansByKeyID(keyID)
	if err != nil || len(loans) == 0 {
		app.showError("Erreur", "Aucun emprunt actif pour cette clé.")
		return
	}

	// Construire les options de retour
	loanOptions := make([]string, len(loans))
	for i, l := range loans {
		days := int(db.GetLoanDuration(l.LoanDate))
		loanOptions[i] = fmt.Sprintf("%s — %d jour(s)", l.BorrowerName, days)
	}

	loanSelect := widget.NewSelect(loanOptions, nil)
	if len(loans) == 1 {
		loanSelect.SetSelectedIndex(0)
	}

	var popup *widget.PopUp
	confirmBtn := widget.NewButton("Confirmer le retour", func() {
		if loanSelect.SelectedIndex() < 0 {
			app.showError("Erreur", "Veuillez sélectionner un emprunt.")
			return
		}
		loan := loans[loanSelect.SelectedIndex()]
		app.showConfirm("Confirmer le retour",
			fmt.Sprintf("Confirmer le retour de la clé %s par %s ?", loan.KeyNumber, loan.BorrowerName),
			func() {
				if err := db.ReturnLoan(loan.ID); err != nil {
					app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
					return
				}
				app.window.Canvas().Overlays().Remove(popup)
				app.showSuccess(fmt.Sprintf("Clé %s retournée par %s.", loan.KeyNumber, loan.BorrowerName))
				app.showDashboard()
			})
	})
	confirmBtn.Importance = widget.MediumImportance

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popup)
	})

	key, _ := db.GetKeyByID(keyID)
	title := "Retour de clé"
	if key != nil {
		title = fmt.Sprintf("Retour — Clé %s", key.Number)
	}

	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%d emprunt(s) actif(s)", len(loans))),
		widget.NewFormItem("Emprunt à retourner", loanSelect).Widget,
		container.NewHBox(cancelBtn, confirmBtn),
	)

	popup = widget.NewModalPopUp(container.NewPadded(content), app.window.Canvas())
	popup.Show()
}

