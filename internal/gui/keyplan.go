package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

// createRoomsToKeysView crée la vue Portes → Clés (Compacte et Triée).
// Le rendu s'appuie sur les cards et badges de ui_kit.go : une card par
// bâtiment regroupant ses accès, chaque accès listant ses clés sous forme de
// pastilles colorées.
func createRoomsToKeysView(buildingsMap map[int]db.Building) fyne.CanvasObject {
	planBox := container.NewVBox()

	if len(buildingsMap) == 0 {
		planBox.Add(widget.NewLabel("Aucun bâtiment configuré"))
		return container.NewPadded(planBox)
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
		buildingTitle := canvas.NewText("🏢 "+building.Name, colorText)
		buildingTitle.TextStyle = fyne.TextStyle{Bold: true}
		buildingTitle.TextSize = 16

		cardContent := container.NewVBox(buildingTitle)

		if len(building.Rooms) == 0 {
			empty := canvas.NewText("Aucune salle", colorTextMuted)
			empty.TextSize = 12
			cardContent.Add(empty)
		} else {
			// Trier les salles par nom
			sort.Slice(building.Rooms, func(i, j int) bool {
				return strings.ToLower(building.Rooms[i].Name) < strings.ToLower(building.Rooms[j].Name)
			})

			for _, room := range building.Rooms {
				roomName := room.Name
				if room.Type != "" {
					roomName += fmt.Sprintf(" (%s)", room.Type)
				}
				nameText := canvas.NewText(roomName, colorText)
				nameText.TextStyle = fyne.TextStyle{Bold: true}
				nameText.TextSize = 13

				var keysObj fyne.CanvasObject
				if len(room.Keys) == 0 {
					keysObj = badge("Aucune clé", colorDangerSoft, colorDanger)
				} else {
					sort.Slice(room.Keys, func(i, j int) bool {
						return room.Keys[i].Number < room.Keys[j].Number
					})
					keyBadges := container.NewHBox()
					for _, key := range room.Keys {
						keyBadges.Add(badge(key.Number, colorPrimarySoft, colorPrimary))
					}
					keysObj = keyBadges
				}

				row := container.NewBorder(nil, nil,
					container.NewHBox(coloredDot(colorPrimary, 8), nameText),
					nil, keysObj)
				cardContent.Add(container.NewPadded(row))
			}
		}
		planBox.Add(card(cardContent))
	}

	return container.NewPadded(planBox)
}

// createKeysToRoomsView crée la vue Clés → Portes (Compacte et Triée).
// Chaque clé est présentée dans une card : en-tête (numéro + description) et
// pastilles des accès qu'elle ouvre, dans l'esprit visuel de ui_kit.go.
func createKeysToRoomsView() fyne.CanvasObject {
	planBox := container.NewVBox()

	// Récupérer toutes les clés
	keys, err := db.GetAllKeys()
	if err != nil {
		planBox.Add(widget.NewLabel(fmt.Sprintf("Erreur: %v", err)))
		return container.NewPadded(planBox)
	}

	if len(keys) == 0 {
		planBox.Add(widget.NewLabel("Aucune clé configurée"))
		return container.NewPadded(planBox)
	}

	// Trier les clés par numéro
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Number < keys[j].Number
	})

	for _, key := range keys {
		keyTitle := canvas.NewText(fmt.Sprintf("🔑 %s — %s", key.Number, key.Description), colorText)
		keyTitle.TextStyle = fyne.TextStyle{Bold: true}
		keyTitle.TextSize = 15

		cardContent := container.NewVBox(keyTitle)

		rooms, err := db.GetRoomsForKey(key.ID)
		switch {
		case err != nil:
			cardContent.Add(badge("Erreur de chargement", colorDangerSoft, colorDanger))
		case len(rooms) == 0:
			cardContent.Add(badge("Aucune porte", colorWarningSoft, colorWarning))
		default:
			sort.Slice(rooms, func(i, j int) bool {
				return strings.ToLower(rooms[i].Name) < strings.ToLower(rooms[j].Name)
			})
			roomBadges := container.NewVBox()
			line := container.NewHBox()
			for i, room := range rooms {
				line.Add(badge(room.Name, colorPrimarySoft, colorPrimary))
				// Répartir en lignes de 4 pastilles pour éviter un débordement.
				if (i+1)%4 == 0 {
					roomBadges.Add(line)
					line = container.NewHBox()
				}
			}
			if len(line.Objects) > 0 {
				roomBadges.Add(line)
			}
			cardContent.Add(roomBadges)
		}

		planBox.Add(card(cardContent))
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
