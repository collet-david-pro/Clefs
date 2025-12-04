package gui

import (
	"clefs/internal/db"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createKeysView crée la vue de gestion des clés
func createKeysView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gérer les Clés", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("➕ Ajouter une Clé", func() {
		showAddKeyDialog(app)
	})
	addBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, addBtn, title)

	// Récupérer les clés
	keys, err := db.GetAllKeys()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Créer la liste des clés
	keysList := createKeysListView(keys, app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(keysList),
	)

	return content
}

// createKeysListView crée la liste des clés
func createKeysListView(keys []db.Key, app *App) fyne.CanvasObject {
	list := container.NewVBox()

	for _, key := range keys {
		k := key // Capture
		
		// Récupérer le nombre d'emprunts actifs
		loanCount, _ := db.GetKeyActiveLoanCount(k.ID)
		
		// Récupérer les salles associées
		rooms, _ := db.GetRoomsForKey(k.ID)
		roomsText := "Aucune salle"
		if len(rooms) > 0 {
			roomsText = ""
			for i, room := range rooms {
				if i > 0 {
					roomsText += ", "
				}
				roomsText += room.Name
			}
		}

		keyInfo := container.NewVBox(
			widget.NewLabelWithStyle(fmt.Sprintf("Clé %s", k.Number), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("Description: %s", k.Description)),
			widget.NewLabel(fmt.Sprintf("Quantité totale: %d | Réserve: %d", k.QuantityTotal, k.QuantityReserve)),
			widget.NewLabel(fmt.Sprintf("Emplacement: %s", k.StorageLocation)),
			widget.NewLabel(fmt.Sprintf("Salles: %s", roomsText)),
			widget.NewLabel(fmt.Sprintf("Emprunts actifs: %d", loanCount)),
		)

		editBtn := widget.NewButton("✏️ Modifier", func() {
			showEditKeyDialog(app, k.ID)
		})

		deleteBtn := widget.NewButton("🗑️ Supprimer", func() {
			app.showConfirm("Confirmer la suppression",
				fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la clé %s?", k.Number),
				func() {
					err := db.DeleteKey(k.ID)
					if err != nil {
						app.showError("Erreur", fmt.Sprintf("Erreur lors de la suppression: %v", err))
						return
					}
					app.showSuccess("Clé supprimée avec succès!")
					app.showKeys()
				})
		})
		deleteBtn.Importance = widget.DangerImportance

		actions := container.NewHBox(editBtn, deleteBtn)

		keyCard := container.NewBorder(nil, nil, nil, actions, keyInfo)
		list.Add(keyCard)
		// Séparateur léger entre les éléments
		if k.ID != keys[len(keys)-1].ID {
			list.Add(widget.NewSeparator())
		}
	}

	return list
}

// showAddKeyDialog affiche la boîte de dialogue pour ajouter une clé
func showAddKeyDialog(app *App) {
	// Récupérer les bâtiments et salles
	buildings, err := db.GetAllBuildings()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des bâtiments: %v", err))
		return
	}

	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Numéro de la clé")

	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Description")

	totalEntry := widget.NewEntry()
	totalEntry.SetPlaceHolder("1")
	totalEntry.SetText("1")

	reserveEntry := widget.NewEntry()
	reserveEntry.SetPlaceHolder("0")
	reserveEntry.SetText("0")

	storageEntry := widget.NewEntry()
	storageEntry.SetPlaceHolder("Emplacement de stockage")

	// Sélection des salles
	roomCheckboxes := make(map[int]*widget.Check)
	roomsBox := container.NewVBox()

	for _, building := range buildings {
		buildingLabel := widget.NewLabelWithStyle(building.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		roomsBox.Add(buildingLabel)

		rooms, _ := db.GetRoomsByBuildingID(building.ID)
		for _, room := range rooms {
			r := room
			checkbox := widget.NewCheck(r.Name, nil)
			roomCheckboxes[r.ID] = checkbox
			roomsBox.Add(checkbox)
		}
	}

	form := container.NewVBox(
		widget.NewLabel("Numéro de la clé:"),
		numberEntry,
		widget.NewLabel("Description:"),
		descEntry,
		widget.NewLabel("Quantité totale:"),
		totalEntry,
		widget.NewLabel("Quantité en réserve:"),
		reserveEntry,
		widget.NewLabel("Emplacement de stockage:"),
		storageEntry,
		widget.NewSeparator(),
		widget.NewLabel("Salles associées:"),
		container.NewVScroll(roomsBox),
	)

	var dialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(dialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if numberEntry.Text == "" {
			app.showError("Erreur", "Le numéro de la clé est requis.")
			return
		}

		total, err := strconv.Atoi(totalEntry.Text)
		if err != nil || total < 1 {
			app.showError("Erreur", "La quantité totale doit être un nombre positif.")
			return
		}

		reserve, err := strconv.Atoi(reserveEntry.Text)
		if err != nil || reserve < 0 {
			app.showError("Erreur", "La quantité en réserve doit être un nombre positif ou zéro.")
			return
		}

		// Récupérer les salles sélectionnées
		var selectedRoomIDs []int
		for roomID, checkbox := range roomCheckboxes {
			if checkbox.Checked {
				selectedRoomIDs = append(selectedRoomIDs, roomID)
			}
		}

		key := &db.Key{
			Number:          numberEntry.Text,
			Description:     descEntry.Text,
			QuantityTotal:   total,
			QuantityReserve: reserve,
			StorageLocation: storageEntry.Text,
		}

		err = db.CreateKey(key, selectedRoomIDs)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Clé créée avec succès!")
		app.showKeys()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Ajouter une Clé", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(600, 600))
	dialog.Show()
}

// showEditKeyDialog affiche la boîte de dialogue pour modifier une clé
func showEditKeyDialog(app *App, keyID int) {
	// Récupérer la clé
	key, err := db.GetKeyByID(keyID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération de la clé: %v", err))
		return
	}

	// Récupérer les salles actuelles
	currentRooms, _ := db.GetRoomsForKey(keyID)
	currentRoomIDs := make(map[int]bool)
	for _, room := range currentRooms {
		currentRoomIDs[room.ID] = true
	}

	// Récupérer les bâtiments et salles
	buildings, err := db.GetAllBuildings()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des bâtiments: %v", err))
		return
	}

	numberEntry := widget.NewEntry()
	numberEntry.SetText(key.Number)

	descEntry := widget.NewEntry()
	descEntry.SetText(key.Description)

	totalEntry := widget.NewEntry()
	totalEntry.SetText(strconv.Itoa(key.QuantityTotal))

	reserveEntry := widget.NewEntry()
	reserveEntry.SetText(strconv.Itoa(key.QuantityReserve))

	storageEntry := widget.NewEntry()
	storageEntry.SetText(key.StorageLocation)

	// Sélection des salles
	roomCheckboxes := make(map[int]*widget.Check)
	roomsBox := container.NewVBox()

	for _, building := range buildings {
		buildingLabel := widget.NewLabelWithStyle(building.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		roomsBox.Add(buildingLabel)

		rooms, _ := db.GetRoomsByBuildingID(building.ID)
		for _, room := range rooms {
			r := room
			checkbox := widget.NewCheck(r.Name, nil)
			if currentRoomIDs[r.ID] {
				checkbox.Checked = true
			}
			roomCheckboxes[r.ID] = checkbox
			roomsBox.Add(checkbox)
		}
	}

	form := container.NewVBox(
		widget.NewLabel("Numéro de la clé:"),
		numberEntry,
		widget.NewLabel("Description:"),
		descEntry,
		widget.NewLabel("Quantité totale:"),
		totalEntry,
		widget.NewLabel("Quantité en réserve:"),
		reserveEntry,
		widget.NewLabel("Emplacement de stockage:"),
		storageEntry,
		widget.NewSeparator(),
		widget.NewLabel("Salles associées:"),
		container.NewVScroll(roomsBox),
	)

	var dialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(dialog)
	})

	saveBtn := widget.NewButton("Enregistrer", func() {
		if numberEntry.Text == "" {
			app.showError("Erreur", "Le numéro de la clé est requis.")
			return
		}

		total, err := strconv.Atoi(totalEntry.Text)
		if err != nil || total < 1 {
			app.showError("Erreur", "La quantité totale doit être un nombre positif.")
			return
		}

		reserve, err := strconv.Atoi(reserveEntry.Text)
		if err != nil || reserve < 0 {
			app.showError("Erreur", "La quantité en réserve doit être un nombre positif ou zéro.")
			return
		}

		// Récupérer les salles sélectionnées
		var selectedRoomIDs []int
		for roomID, checkbox := range roomCheckboxes {
			if checkbox.Checked {
				selectedRoomIDs = append(selectedRoomIDs, roomID)
			}
		}

		key.Number = numberEntry.Text
		key.Description = descEntry.Text
		key.QuantityTotal = total
		key.QuantityReserve = reserve
		key.StorageLocation = storageEntry.Text

		err = db.UpdateKey(key, selectedRoomIDs)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la modification: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Clé modifiée avec succès!")
		app.showKeys()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Modifier la Clé", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(600, 600))
	dialog.Show()
}
