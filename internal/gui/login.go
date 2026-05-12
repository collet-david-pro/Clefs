package gui

import (
	"clefs/internal/db"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// showLoginDialog affiche la fenêtre d'identification au démarrage.
// onConfirm est appelé avec le nom de l'agent une fois validé.
func showLoginDialog(a *App, onConfirm func(username string)) {
	borrowers, _ := db.GetAllBorrowers()

	// Construire la liste des noms connus + option saisie libre
	names := make([]string, 0, len(borrowers)+1)
	for _, b := range borrowers {
		names = append(names, b.Name)
	}
	names = append(names, "Autre...")

	selectWidget := widget.NewSelect(names, nil)
	if len(names) > 0 {
		selectWidget.SetSelectedIndex(0)
	}

	freeEntry := widget.NewEntry()
	freeEntry.SetPlaceHolder("Saisissez votre nom...")
	freeEntry.Hide()

	selectWidget.OnChanged = func(val string) {
		if val == "Autre..." {
			freeEntry.Show()
		} else {
			freeEntry.Hide()
		}
	}

	var popup *widget.PopUp

	confirmBtn := widget.NewButton("Commencer", func() {
		name := selectWidget.Selected
		if name == "Autre..." {
			name = freeEntry.Text
		}
		if name == "" {
			name = "Agent"
		}
		a.window.Canvas().Overlays().Remove(popup)
		onConfirm(name)
	})
	confirmBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("🔑 Gestionnaire de Clés", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Collège Victor Hugo — Chauny"),
		widget.NewSeparator(),
		widget.NewLabel("Qui êtes-vous ?"),
		selectWidget,
		freeEntry,
		confirmBtn,
	)

	popup = widget.NewModalPopUp(container.NewPadded(content), a.window.Canvas())
	popup.Show()
}
