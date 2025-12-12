package gui

import (
	"clefs/internal/db"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createBuildingsView crée la vue de gestion des bâtiments
func createBuildingsView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gérer les Bâtiments", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("➕ Ajouter un Bâtiment", func() {
		showAddBuildingDialog(app)
	})
	addBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, addBtn, title)

	// Récupérer les bâtiments
	buildings, err := db.GetAllBuildings()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Créer la liste des bâtiments
	buildingsList := createBuildingsListView(buildings, app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(buildingsList),
	)

	return content
}

// createBuildingsListView crée la liste des bâtiments
func createBuildingsListView(buildings []db.Building, app *App) fyne.CanvasObject {
	list := container.NewVBox()

	for _, building := range buildings {
		b := building // Capture

		// Récupérer le nombre de salles
		rooms, _ := db.GetRoomsByBuildingID(b.ID)
		roomCount := len(rooms)

		buildingInfo := container.NewVBox(
			widget.NewLabelWithStyle(b.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("Nombre de salles: %d", roomCount)),
		)

		editBtn := widget.NewButton("✏️ Modifier", func() {
			showEditBuildingDialog(app, b.ID)
		})

		deleteBtn := widget.NewButton("🗑️ Supprimer", func() {
			if roomCount > 0 {
				app.showError("Impossible de supprimer", "Ce bâtiment contient des salles.")
				return
			}
			app.showConfirm("Confirmer la suppression",
				fmt.Sprintf("Êtes-vous sûr de vouloir supprimer le bâtiment %s?", b.Name),
				func() {
					err := db.DeleteBuilding(b.ID)
					if err != nil {
						app.showError("Erreur", fmt.Sprintf("Erreur lors de la suppression: %v", err))
						return
					}
					app.showSuccess("Bâtiment supprimé avec succès!")
					app.showBuildings()
				})
		})
		deleteBtn.Importance = widget.DangerImportance

		actions := container.NewHBox(editBtn, deleteBtn)

		buildingCard := container.NewBorder(nil, nil, nil, actions, buildingInfo)
		list.Add(buildingCard)
		// Séparateur seulement entre les éléments
		if b.ID != buildings[len(buildings)-1].ID {
			list.Add(widget.NewSeparator())
		}
	}

	return list
}

// showAddBuildingDialog affiche la boîte de dialogue pour ajouter un bâtiment
func showAddBuildingDialog(app *App) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom du bâtiment")

	form := container.NewVBox(
		widget.NewLabel("Nom du bâtiment:"),
		nameEntry,
	)

	var popupDialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom du bâtiment est requis.")
			return
		}

		building := &db.Building{
			Name: nameEntry.Text,
		}

		err := db.CreateBuilding(building)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Bâtiment créé avec succès!")
		app.showBuildings()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Ajouter un Bâtiment", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(400, 200))
	popupDialog.Show()
}

// showEditBuildingDialog affiche la boîte de dialogue pour modifier un bâtiment
func showEditBuildingDialog(app *App, buildingID int) {
	// Récupérer le bâtiment
	building, err := db.GetBuildingByID(buildingID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération du bâtiment: %v", err))
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(building.Name)

	form := container.NewVBox(
		widget.NewLabel("Nom du bâtiment:"),
		nameEntry,
	)

	var popupDialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom du bâtiment est requis.")
			return
		}

		building.Name = nameEntry.Text

		err := db.UpdateBuilding(building)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la modification: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Bâtiment modifié avec succès!")
		app.showBuildings()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Modifier le Bâtiment", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(400, 200))
	popupDialog.Show()
}
