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

// createKeysListView crée la liste des clés avec accordéon
func createKeysListView(keys []db.Key, app *App) fyne.CanvasObject {
	list := container.NewVBox()

	for _, key := range keys {
		k := key // Capture

		// Créer l'accordéon pour cette clé
		accordion := createKeyAccordion(app, k)
		list.Add(accordion)
		list.Add(widget.NewLabel("")) // Espacement
	}

	return list
}

// createKeyAccordion crée un accordéon pour une clé
func createKeyAccordion(app *App, key db.Key) *widget.Accordion {
	activeLoans, err := db.GetActiveLoansByKeyID(key.ID)
	if err != nil {
		activeLoans = []db.LoanWithDetails{}
	}

	rooms, err := db.GetRoomsForKey(key.ID)
	if err != nil {
		rooms = []db.Room{}
	}
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

	// Calculer la disponibilité
	borrowed := len(activeLoans)
	available := key.QuantityTotal - key.QuantityReserve - borrowed

	// Créer le contenu détaillé
	detailsContent := container.NewVBox()

	// Informations de la clé
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("📝 Description: %s", key.Description)))
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("📦 Quantité totale: %d | Réserve: %d", key.QuantityTotal, key.QuantityReserve)))
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("📍 Emplacement: %s", key.StorageLocation)))
	detailsContent.Add(widget.NewLabel(fmt.Sprintf("🏢 Salles: %s", roomsText)))

	// Statut de disponibilité avec couleur
	statusText := fmt.Sprintf("✅ Disponibles: %d | 🔴 Sorties: %d", available, borrowed)
	if available <= 0 {
		statusText = fmt.Sprintf("⚠️ STOCK ÉPUISÉ | 🔴 Sorties: %d", borrowed)
	}
	detailsContent.Add(widget.NewLabelWithStyle(statusText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	detailsContent.Add(widget.NewSeparator())

	// Liste des emprunts actifs
	if len(activeLoans) > 0 {
		detailsContent.Add(widget.NewLabelWithStyle("📋 Emprunts en cours:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

		for _, loan := range activeLoans {
			l := loan // Capture

			// Calculer la durée
			days := int(db.GetLoanDuration(l.LoanDate))
			durationText := fmt.Sprintf("%d jour(s)", days)
			if days == 0 {
				durationText = "Aujourd'hui"
			}

			loanInfo := container.NewVBox(
				widget.NewLabel(fmt.Sprintf("   👤 %s", l.BorrowerName)),
				widget.NewLabel(fmt.Sprintf("   📅 Depuis le: %s (%s)",
					l.LoanDate.Format("02/01/2006"), durationText)),
			)

			returnBtn := widget.NewButton("↩️ Retourner", func() {
				app.showConfirm("Confirmer le retour",
					fmt.Sprintf("Confirmer le retour de la clé %s empruntée par %s?", key.Number, l.BorrowerName),
					func() {
						err := db.ReturnLoan(l.ID)
						if err != nil {
							app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
							return
						}
						app.showSuccess("Clé retournée avec succès!")
						app.showKeys()
					})
			})
			returnBtn.Importance = widget.MediumImportance

			loanRow := container.NewBorder(nil, nil, nil, returnBtn, loanInfo)
			detailsContent.Add(loanRow)
			detailsContent.Add(widget.NewSeparator())
		}
	} else {
		detailsContent.Add(widget.NewLabel("✅ Aucun emprunt actif pour cette clé"))
		detailsContent.Add(widget.NewSeparator())
	}

	// Boutons d'action
	editBtn := widget.NewButton("✏️ Modifier", func() {
		showEditKeyDialog(app, key.ID)
	})

	deleteBtn := widget.NewButton("🗑️ Supprimer", func() {
		app.showConfirm("Confirmer la suppression",
			fmt.Sprintf("Êtes-vous sûr de vouloir supprimer la clé %s?", key.Number),
			func() {
				err := db.DeleteKey(key.ID)
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
	detailsContent.Add(actions)

	// Créer l'item d'accordéon
	title := fmt.Sprintf("🔑 %s - %s", key.Number, key.Description)
	if borrowed > 0 {
		title = fmt.Sprintf("🔑 %s - %s (%d sortie(s))", key.Number, key.Description, borrowed)
	}

	accordionItem := widget.NewAccordionItem(title, detailsContent)

	// Créer l'accordéon
	accordion := widget.NewAccordion(accordionItem)

	return accordion
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
