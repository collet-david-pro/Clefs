package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createKeyPlanView crée la vue du plan de clés avec 2 vues
func createKeyPlanView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Plan de Clés", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Bouton d'action
	exportBtn := widget.NewButton("📄 Générer PDF du Plan", func() {
		generateKeyPlanPDF(app)
	})
	exportBtn.Importance = widget.HighImportance

	buttonsContainer := container.NewHBox(exportBtn)

	// Récupérer les données du plan de clés
	buildingsMap, err := db.GetKeyPlanData()
	if err != nil {
		return container.NewVBox(
			title,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Créer les deux vues
	roomsView := createRoomsToKeysView(buildingsMap)
	keysView := createKeysToRoomsView()

	// Créer les onglets
	tabs := container.NewAppTabs(
		container.NewTabItem("Portes -> Cles", container.NewVScroll(roomsView)),
		container.NewTabItem("Cles -> Portes", container.NewVScroll(keysView)),
	)

	header := container.NewBorder(nil, nil, nil, buttonsContainer, title)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		tabs,
	)

	return content
}

// createRoomsToKeysView crée la vue Portes → Clés
func createRoomsToKeysView(buildingsMap map[int]db.Building) fyne.CanvasObject {
	planBox := container.NewVBox()

	if len(buildingsMap) == 0 {
		planBox.Add(widget.NewLabel("Aucun bâtiment configuré"))
		return planBox
	}

	planBox.Add(widget.NewLabelWithStyle("📍 Plan des Portes et leurs Clés", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	planBox.Add(widget.NewLabel("Vue organisée par bâtiments et salles, montrant les clés qui ouvrent chaque porte."))
	planBox.Add(widget.NewSeparator())

	for _, building := range buildingsMap {
		// En-tête du bâtiment
		buildingLabel := widget.NewLabelWithStyle("🏢 "+building.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		planBox.Add(buildingLabel)

		if len(building.Rooms) == 0 {
			planBox.Add(widget.NewLabel("  Aucune salle"))
		} else {
			// Pour chaque salle
			for _, room := range building.Rooms {
				roomText := fmt.Sprintf("  🚪 %s", room.Name)
				if room.Type != "" {
					roomText += fmt.Sprintf(" (%s)", room.Type)
				}

				roomLabel := widget.NewLabelWithStyle(roomText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				planBox.Add(roomLabel)

				// Clés associées
				if len(room.Keys) == 0 {
					planBox.Add(widget.NewLabel("      Aucune clé associée"))
				} else {
					for _, key := range room.Keys {
						keyText := fmt.Sprintf("      🔑 %s - %s", key.Number, key.Description)
						keyLabel := widget.NewLabel(keyText)
						planBox.Add(keyLabel)
					}
				}
				planBox.Add(widget.NewLabel("")) // Espacement
			}
		}

		planBox.Add(widget.NewSeparator())
	}

	return planBox
}

// createKeysToRoomsView crée la vue Clés → Portes
func createKeysToRoomsView() fyne.CanvasObject {
	planBox := container.NewVBox()

	// Récupérer toutes les clés
	keys, err := db.GetAllKeys()
	if err != nil {
		planBox.Add(widget.NewLabel(fmt.Sprintf("Erreur: %v", err)))
		return planBox
	}

	if len(keys) == 0 {
		planBox.Add(widget.NewLabel("Aucune clé configurée"))
		return planBox
	}

	planBox.Add(widget.NewLabelWithStyle("🔑 Plan des Clés et leurs Portes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	planBox.Add(widget.NewLabel("Vue organisée par clés, montrant toutes les portes que chaque clé peut ouvrir."))
	planBox.Add(widget.NewSeparator())

	// Pour chaque clé
	for _, key := range keys {
		// En-tête de la clé
		keyHeader := fmt.Sprintf("🔑 %s - %s", key.Number, key.Description)
		keyLabel := widget.NewLabelWithStyle(keyHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		planBox.Add(keyLabel)

		// Informations supplémentaires
		infoText := fmt.Sprintf("   Quantité: %d (Réserve: %d)", key.QuantityTotal, key.QuantityReserve)
		if key.StorageLocation != "" {
			infoText += fmt.Sprintf(" | Stockage: %s", key.StorageLocation)
		}
		planBox.Add(widget.NewLabel(infoText))

		// Récupérer les salles associées
		rooms, err := db.GetRoomsForKey(key.ID)
		if err != nil {
			planBox.Add(widget.NewLabel(fmt.Sprintf("   Erreur: %v", err)))
		} else if len(rooms) == 0 {
			planBox.Add(widget.NewLabel("   Aucune porte associée"))
		} else {
			planBox.Add(widget.NewLabel("   Ouvre les portes suivantes:"))

			// Grouper par bâtiment
			buildingRooms := make(map[int][]db.Room)
			for _, room := range rooms {
				buildingRooms[room.BuildingID] = append(buildingRooms[room.BuildingID], room)
			}

			// Afficher par bâtiment
			for buildingID, roomList := range buildingRooms {
				building, err := db.GetBuildingByID(buildingID)
				if err == nil {
					planBox.Add(widget.NewLabel(fmt.Sprintf("      🏢 %s:", building.Name)))
					for _, room := range roomList {
						roomText := fmt.Sprintf("         🚪 %s", room.Name)
						if room.Type != "" {
							roomText += fmt.Sprintf(" (%s)", room.Type)
						}
						planBox.Add(widget.NewLabel(roomText))
					}
				}
			}
		}

		planBox.Add(widget.NewLabel("")) // Espacement
		planBox.Add(widget.NewSeparator())
	}

	return planBox
}

// generateKeyPlanPDF génère et enregistre le plan de clés en PDF
func generateKeyPlanPDF(app *App) {
	// Récupérer les données du plan de clés
	buildingsMap, err := db.GetKeyPlanData()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des données: %v", err))
		return
	}

	if len(buildingsMap) == 0 {
		app.showError("Aucune donnée", "Aucun bâtiment configuré.")
		return
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateKeyPlanPDF(buildingsMap)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename("plan_de_cles", 0)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Plan de clés enregistré : %s", filepath))
}
