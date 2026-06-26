package gui

import (
	"clefs/internal/db"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
