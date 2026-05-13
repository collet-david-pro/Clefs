package gui

import (
	"clefs/internal/db"
	"clefs/internal/export"
	"clefs/internal/pdf"
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

	stockReportBtn := widget.NewButton("📦 Générer Bilan des Clés", func() {
		generateKeyStockReportPDF(app)
	})

	csvBtn := widget.NewButton("📊 Exporter CSV", func() {
		exportKeysCSV(app)
	})

	header := container.NewBorder(nil, nil, nil, container.NewHBox(csvBtn, stockReportBtn, addBtn), title)

	// Récupérer les clés avec disponibilité
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

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

// createKeysListView crée la liste des clés avec accordéon
// createKeysListView crée la liste des clés sous forme de cards plates.
func createKeysListView(keys []db.KeyWithAvailability, app *App) fyne.CanvasObject {
	list := container.NewVBox()
	for _, k := range keys {
		k := k
		card := keyCard(k,
			func() { // Emprunter — ouvre la vue prêt unifiée
				app.showNewLoan()
			},
			func() { // Retour
				showReturnDialog(app, k.ID)
			},
		)
		// Boutons modifier / supprimer en pied de card
		editBtn := widget.NewButton("Modifier", func() { showEditKeyDialog(app, k.ID) })
		deleteBtn := widget.NewButton("Supprimer", func() {
			app.showConfirm("Supprimer",
				fmt.Sprintf("Supprimer la clé %s ?", k.Number),
				func() {
					if err := db.DeleteKey(k.ID); err != nil {
						app.showError("Erreur", err.Error())
						return
					}
					app.showKeys()
				})
		})
		deleteBtn.Importance = widget.DangerImportance
		actions := container.NewHBox(editBtn, deleteBtn)
		list.Add(container.NewVBox(card, container.NewPadded(actions), widget.NewSeparator()))
	}
	if len(list.Objects) == 0 {
		list.Add(widget.NewLabel("Aucune clé enregistrée."))
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

	// Champs du formulaire en haut (hauteur fixe)
	topFields := container.NewVBox(
		widget.NewLabel("Numéro de la clé :"),
		numberEntry,
		widget.NewLabel("Description :"),
		descEntry,
		widget.NewLabel("Quantité totale :"),
		totalEntry,
		widget.NewLabel("Quantité en réserve :"),
		reserveEntry,
		widget.NewLabel("Emplacement de stockage :"),
		storageEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Accès associés :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
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
		if err = db.CreateKey(key, selectedRoomIDs); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Clé créée avec succès!")
		app.showKeys()
	})
	saveBtn.Importance = widget.HighImportance

	// Layout : champs en haut (fixe) + liste des accès scrollable + boutons en bas
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Ajouter une Clé", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			topFields,
		),
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(cancelBtn, saveBtn),
		),
		nil, nil,
		container.NewVScroll(roomsBox),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(600, 700))
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

	topFields := container.NewVBox(
		widget.NewLabel("Numéro de la clé :"),
		numberEntry,
		widget.NewLabel("Description :"),
		descEntry,
		widget.NewLabel("Quantité totale :"),
		totalEntry,
		widget.NewLabel("Quantité en réserve :"),
		reserveEntry,
		widget.NewLabel("Emplacement de stockage :"),
		storageEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Accès associés :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
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
		if err = db.UpdateKey(key, selectedRoomIDs); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la modification: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Clé modifiée avec succès!")
		app.showKeys()
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Modifier la Clé", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			topFields,
		),
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(cancelBtn, saveBtn),
		),
		nil, nil,
		container.NewVScroll(roomsBox),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(600, 700))
	dialog.Show()
}

// generateKeyStockReportPDF génère et enregistre le bilan du stock de clés
func generateKeyStockReportPDF(app *App) {
	// Récupérer toutes les clés
	keys, err := db.GetAllKeys()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des clés: %v", err))
		return
	}

	// Récupérer les comptes d'emprunts pour chaque clé
	loanCounts := make(map[int]int)
	for _, key := range keys {
		count, _ := db.GetActiveLoanCount(key.ID)
		loanCounts[key.ID] = count
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateKeyStockReport(keys, loanCounts)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename("bilan_cles", 0)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Bilan enregistré : %s", filepath))
}

func exportKeysCSV(app *App) {
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des clés: %v", err))
		return
	}
	headers := []string{"Numéro", "Description", "Catégorie", "Total", "Réserve", "Sorties", "Disponibles", "Emplacement", "Notes"}
	rows := make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{
			k.Number, k.Description, k.Category,
			strconv.Itoa(k.QuantityTotal), strconv.Itoa(k.QuantityReserve),
			strconv.Itoa(k.LoanedCount), strconv.Itoa(k.AvailableCount),
			k.StorageLocation, k.Notes,
		}
	}
	filePath, err := export.SaveCSV(export.Filename("inventaire_cles"), headers, rows)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'export: %v", err))
		return
	}
	app.showSuccess(fmt.Sprintf("✅ Export CSV enregistré : %s", filePath))
}
