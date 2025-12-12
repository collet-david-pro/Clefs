package gui

import (
	"clefs/internal/db"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createRoomsView crée la vue de gestion des salles
func createRoomsView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gérer les Points d'Accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("➕ Ajouter un Point d'Accès", func() {
		showAddRoomDialog(app)
	})
	addBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, addBtn, title)

	// Récupérer les bâtiments avec leurs salles
	buildings, err := db.GetAllBuildings()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Créer la liste des salles par bâtiment
	roomsList := createRoomsListView(buildings, app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(roomsList),
	)

	return content
}

// createRoomsListView crée la liste des salles groupées par bâtiment
func createRoomsListView(buildings []db.Building, app *App) fyne.CanvasObject {
	list := container.NewVBox()

	for _, building := range buildings {
		b := building // Capture

		// En-tête du bâtiment
		buildingLabel := widget.NewLabelWithStyle(b.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		list.Add(buildingLabel)

		// Récupérer les salles du bâtiment
		rooms, err := db.GetRoomsByBuildingID(b.ID)
		if err != nil {
			continue
		}

		if len(rooms) == 0 {
			list.Add(widget.NewLabel("  Aucune salle"))
		} else {
			for _, room := range rooms {
				r := room // Capture

				roomText := fmt.Sprintf("  %s", r.Name)
				if r.Type != "" {
					roomText += fmt.Sprintf(" (%s)", r.Type)
				}

				roomLabel := widget.NewLabel(roomText)

				editBtn := widget.NewButton("✏️", func() {
					showEditRoomDialog(app, r.ID)
				})
				editBtn.Importance = widget.LowImportance

				deleteBtn := widget.NewButton("🗑️", func() {
					// Vérifier si des clés sont associées
					keys, _ := db.GetKeysForRoom(r.ID)
					if len(keys) > 0 {
						app.showError("Impossible de supprimer", "Cette salle est associée à des clés.")
						return
					}

					app.showConfirm("Confirmer la suppression",
						fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la salle %s?", r.Name),
						func() {
							err := db.DeleteRoom(r.ID)
							if err != nil {
								app.showError("Erreur", fmt.Sprintf("Erreur lors de la suppression: %v", err))
								return
							}
							app.showSuccess("Salle supprimée avec succès!")
							app.showRooms()
						})
				})
				deleteBtn.Importance = widget.DangerImportance

				actions := container.NewHBox(editBtn, deleteBtn)

				roomRow := container.NewBorder(nil, nil, nil, actions, roomLabel)
				list.Add(roomRow)
			}
		}

		list.Add(widget.NewSeparator())
	}

	return list
}

// showAddRoomDialog affiche la boîte de dialogue pour ajouter une salle
func showAddRoomDialog(app *App) {
	// Récupérer les bâtiments
	buildings, err := db.GetAllBuildings()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des bâtiments: %v", err))
		return
	}

	if len(buildings) == 0 {
		app.showError("Aucun bâtiment", "Veuillez d'abord créer un bâtiment.")
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom de la salle")

	typeEntry := widget.NewEntry()
	typeEntry.SetPlaceHolder("Type (ex: Bureau, Salle de classe)")

	// Sélection du bâtiment
	buildingOptions := make([]string, len(buildings))
	buildingMap := make(map[string]int)
	for i, b := range buildings {
		buildingOptions[i] = b.Name
		buildingMap[b.Name] = b.ID
	}

	buildingSelect := widget.NewSelect(buildingOptions, nil)
	if len(buildingOptions) > 0 {
		buildingSelect.SetSelected(buildingOptions[0])
	}

	form := container.NewVBox(
		widget.NewLabel("Nom de la salle:"),
		nameEntry,
		widget.NewLabel("Type:"),
		typeEntry,
		widget.NewLabel("Bâtiment:"),
		buildingSelect,
	)

	var popupDialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom de la salle est requis.")
			return
		}

		if buildingSelect.Selected == "" {
			app.showError("Erreur", "Veuillez sélectionner un bâtiment.")
			return
		}

		room := &db.Room{
			Name:       nameEntry.Text,
			Type:       typeEntry.Text,
			BuildingID: buildingMap[buildingSelect.Selected],
		}

		err := db.CreateRoom(room)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Salle créée avec succès!")
		app.showRooms()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Ajouter un Point d'Accès", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(400, 300))
	popupDialog.Show()
}

// showEditRoomDialog affiche la boîte de dialogue pour modifier une salle
func showEditRoomDialog(app *App, roomID int) {
	// Récupérer la salle
	rooms, err := db.GetAllRooms()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération de la salle: %v", err))
		return
	}

	var room *db.Room
	for _, r := range rooms {
		if r.ID == roomID {
			room = &r
			break
		}
	}

	if room == nil {
		app.showError("Erreur", "Salle non trouvée.")
		return
	}

	// Récupérer les bâtiments
	buildings, err := db.GetAllBuildings()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des bâtiments: %v", err))
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(room.Name)

	typeEntry := widget.NewEntry()
	typeEntry.SetText(room.Type)

	// Sélection du bâtiment
	buildingOptions := make([]string, len(buildings))
	buildingMap := make(map[string]int)
	var currentBuildingName string

	for i, b := range buildings {
		buildingOptions[i] = b.Name
		buildingMap[b.Name] = b.ID
		if b.ID == room.BuildingID {
			currentBuildingName = b.Name
		}
	}

	buildingSelect := widget.NewSelect(buildingOptions, nil)
	buildingSelect.SetSelected(currentBuildingName)

	form := container.NewVBox(
		widget.NewLabel("Nom de la salle:"),
		nameEntry,
		widget.NewLabel("Type:"),
		typeEntry,
		widget.NewLabel("Bâtiment:"),
		buildingSelect,
	)

	var popupDialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom de la salle est requis.")
			return
		}

		if buildingSelect.Selected == "" {
			app.showError("Erreur", "Veuillez sélectionner un bâtiment.")
			return
		}

		room.Name = nameEntry.Text
		room.Type = typeEntry.Text
		room.BuildingID = buildingMap[buildingSelect.Selected]

		err := db.UpdateRoom(room)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la modification: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Salle modifiée avec succès!")
		app.showRooms()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Modifier le Point d'Accès", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(400, 300))
	popupDialog.Show()
}
