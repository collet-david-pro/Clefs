package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"sort"
	"strings"

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

// createRoomsToKeysView crée la vue Portes → Clés (Compacte et Triée)
func createRoomsToKeysView(buildingsMap map[int]db.Building) fyne.CanvasObject {
	planBox := container.NewVBox()

	if len(buildingsMap) == 0 {
		planBox.Add(widget.NewLabel("Aucun bâtiment configuré"))
		return planBox
	}

	// Convertir la map en slice pour le tri
	var buildings []db.Building
	for _, b := range buildingsMap {
		buildings = append(buildings, b)
	}

	// Trier les bâtiments par nom
	sort.Slice(buildings, func(i, j int) bool {
		return strings.ToLower(buildings[i].Name) < strings.ToLower(buildings[j].Name)
	})

	for _, building := range buildings {
		// En-tête du bâtiment (Compact)
		buildingLabel := widget.NewLabelWithStyle("🏢 "+building.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		planBox.Add(buildingLabel)

		if len(building.Rooms) == 0 {
			planBox.Add(widget.NewLabel("  (Aucune salle)"))
		} else {
			// Trier les salles par nom
			sort.Slice(building.Rooms, func(i, j int) bool {
				return strings.ToLower(building.Rooms[i].Name) < strings.ToLower(building.Rooms[j].Name)
			})

			// Pour chaque salle
			for _, room := range building.Rooms {
				// Construction de la ligne salle + clés
				var textBuilder strings.Builder
				textBuilder.WriteString(fmt.Sprintf("  • %s", room.Name))
				if room.Type != "" {
					textBuilder.WriteString(fmt.Sprintf(" (%s)", room.Type))
				}
				textBuilder.WriteString(" : ")

				if len(room.Keys) == 0 {
					textBuilder.WriteString("Aucune clé")
				} else {
					// Trier les clés par numéro
					sort.Slice(room.Keys, func(i, j int) bool {
						return room.Keys[i].Number < room.Keys[j].Number
					})

					var keyTexts []string
					for _, key := range room.Keys {
						keyTexts = append(keyTexts, fmt.Sprintf("%s", key.Number))
					}
					textBuilder.WriteString(strings.Join(keyTexts, ", "))
				}

				// Affichage compact sur une ligne
				label := widget.NewLabel(textBuilder.String())
				label.Wrapping = fyne.TextWrapWord
				planBox.Add(label)
			}
		}
		// Petit séparateur discret entre bâtiments
		planBox.Add(widget.NewSeparator())
	}

	return container.NewPadded(planBox)
}

// createKeysToRoomsView crée la vue Clés → Portes (Compacte et Triée)
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

	// Trier les clés par numéro
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Number < keys[j].Number
	})

	// Pour chaque clé
	for _, key := range keys {
		// En-tête de la clé
		keyHeader := fmt.Sprintf("🔑 %s - %s", key.Number, key.Description)

		// Récupérer les salles associées
		rooms, err := db.GetRoomsForKey(key.ID)
		var roomsText string

		if err != nil {
			roomsText = "Erreur de chargement"
		} else if len(rooms) == 0 {
			roomsText = "Aucune porte"
		} else {
			// Trier les salles par nom
			sort.Slice(rooms, func(i, j int) bool {
				return strings.ToLower(rooms[i].Name) < strings.ToLower(rooms[j].Name)
			})

			var roomNames []string
			for _, room := range rooms {
				roomNames = append(roomNames, room.Name)
			}
			roomsText = strings.Join(roomNames, ", ")
		}

		// Affichage compact : Clé en gras, liste des portes en dessous
		keyLabel := widget.NewLabelWithStyle(keyHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		planBox.Add(keyLabel)

		roomsLabel := widget.NewLabel("   -> Ouvre : " + roomsText)
		roomsLabel.Wrapping = fyne.TextWrapWord
		planBox.Add(roomsLabel)

		planBox.Add(widget.NewSeparator())
	}

	return container.NewPadded(planBox)
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
