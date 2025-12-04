package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createModernDashboard crée un tableau de bord moderne avec des cards et statistiques
func createModernDashboard(app *App) fyne.CanvasObject {
	// En-tête simplifié
	titleLabel := widget.NewLabelWithStyle("Tableau de Bord", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	newLoanBtn := widget.NewButton("➕ Nouvel Emprunt", func() {
		showNewLoanDialog(app)
	})
	newLoanBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButton("🔄 Rafraîchir", func() {
		app.showDashboard()
	})

	headerButtons := container.NewHBox(newLoanBtn, refreshBtn)
	header := container.NewBorder(nil, nil, titleLabel, headerButtons)

	// Récupérer les statistiques
	stats := getStatistics()

	// Créer les cards de statistiques simplifiées
	statsCards := createStatisticsCards(stats)

	// Récupérer les clés avec disponibilité
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		log.Printf("Erreur lors de la récupération des clés: %v", err)
		return container.NewVBox(
			header,
			widget.NewLabel("Erreur lors du chargement des données"),
		)
	}

	// Créer le tableau simplifié
	keysTable := createSimpleKeysTable(keys, app)

	// Layout principal simplifié
	content := container.NewVBox(
		container.NewPadded(header),
		widget.NewSeparator(),
		container.NewPadded(statsCards),
		widget.NewSeparator(),
		container.NewPadded(keysTable),
	)

	return container.NewScroll(content)
}

// getStatistics récupère les statistiques pour le dashboard
func getStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Récupérer le nombre total de clés
	keys, _ := db.GetAllKeys()
	stats["totalKeys"] = len(keys)

	// Récupérer les emprunts actifs
	activeLoans, _ := db.GetAllActiveLoans()
	stats["activeLoans"] = len(activeLoans)

	// Récupérer les clés disponibles
	availableKeys, _ := db.GetAvailableKeys()
	stats["availableKeys"] = len(availableKeys)

	// Récupérer le nombre d'emprunteurs
	borrowers, _ := db.GetAllBorrowers()
	stats["totalBorrowers"] = len(borrowers)

	return stats
}

// createStatisticsCards crée les cards de statistiques simplifiées
func createStatisticsCards(stats map[string]interface{}) fyne.CanvasObject {
	// Créer des labels simples pour les statistiques
	totalKeysLabel := widget.NewLabel(fmt.Sprintf("🔑 Total des Clés: %d", stats["totalKeys"]))
	activeLoansLabel := widget.NewLabel(fmt.Sprintf("📤 Emprunts Actifs: %d", stats["activeLoans"]))
	availableKeysLabel := widget.NewLabel(fmt.Sprintf("✅ Clés Disponibles: %d", stats["availableKeys"]))
	borrowersLabel := widget.NewLabel(fmt.Sprintf("👥 Emprunteurs: %d", stats["totalBorrowers"]))

	// Conteneur horizontal pour les stats
	statsContainer := container.NewHBox(
		totalKeysLabel,
		widget.NewSeparator(),
		activeLoansLabel,
		widget.NewSeparator(),
		availableKeysLabel,
		widget.NewSeparator(),
		borrowersLabel,
	)

	return container.NewCenter(statsContainer)
}

// createStatsCard crée une card de statistique stylisée
func createStatsCard(title string, value string, colorName fyne.ThemeColorName) fyne.CanvasObject {
	valueLabel := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{})

	content := container.NewVBox(
		container.NewCenter(valueLabel),
		container.NewCenter(titleLabel),
	)

	card := widget.NewCard("", "", content)
	return card
}

// createSimpleKeysTable crée un tableau simple et lisible des clés
func createSimpleKeysTable(keys []db.KeyWithAvailability, app *App) fyne.CanvasObject {
	if len(keys) == 0 {
		emptyLabel := widget.NewLabelWithStyle(
			"Aucune clé dans l'inventaire",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		)
		return container.NewCenter(emptyLabel)
	}

	// Headers avec style
	headers := []string{"Numéro", "Description", "Disponibilité", "Emprunteurs", "Actions"}

	table := widget.NewTable(
		func() (int, int) {
			return len(keys) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			return container.NewMax(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cellContainer := cell.(*fyne.Container)
			cellContainer.Objects = nil

			if id.Row == 0 {
				// En-têtes avec style
				label := widget.NewLabelWithStyle(
					headers[id.Col],
					fyne.TextAlignCenter,
					fyne.TextStyle{Bold: true},
				)
				cellContainer.Add(container.NewCenter(label))
			} else {
				key := keys[id.Row-1]
				switch id.Col {
				case 0:
					// Numéro avec badge
					label := widget.NewLabelWithStyle(
						key.Number,
						fyne.TextAlignCenter,
						fyne.TextStyle{Bold: true},
					)
					cellContainer.Add(container.NewCenter(label))

				case 1:
					// Description
					label := widget.NewLabel(key.Description)
					label.Wrapping = fyne.TextWrapWord
					cellContainer.Add(label)

				case 2:
					// Disponibilité simple avec texte coloré
					usable := key.QuantityTotal - key.QuantityReserve
					availText := fmt.Sprintf("%d / %d", key.AvailableCount, usable)

					availLabel := widget.NewLabel(availText)
					if key.AvailableCount > 0 {
						availLabel.Importance = widget.SuccessImportance
					} else {
						availLabel.Importance = widget.DangerImportance
					}
					cellContainer.Add(container.NewCenter(availLabel))

				case 3:
					// Emprunteurs - Affichage optimisé sur une ligne
					borrowersText := "--"
					if len(key.BorrowerNames) > 0 {
						if len(key.BorrowerNames) <= 3 {
							// 1-3 emprunteurs : affichage avec séparateurs
							borrowersText = ""
							for i, name := range key.BorrowerNames {
								if i > 0 {
									borrowersText += " | "
								}
								borrowersText += name
							}
						} else {
							// 4+ emprunteurs : affichage compact
							borrowersText = fmt.Sprintf("%s, %s et %d autre(s)",
								key.BorrowerNames[0],
								key.BorrowerNames[1],
								len(key.BorrowerNames)-2)
						}
					}
					label := widget.NewLabel(borrowersText)
					label.Wrapping = fyne.TextWrapWord
					cellContainer.Add(label)

				case 4:
					// Actions avec icônes
					actions := container.NewHBox()

					if key.AvailableCount > 0 {
						borrowBtn := widget.NewButton("Emprunter", func() {
							k := key
							showNewLoanDialogWithKey(app, k.ID)
						})
						borrowBtn.Importance = widget.HighImportance
						actions.Add(borrowBtn)
					}

					if key.LoanedCount > 0 {
						returnBtn := widget.NewButton("Retourner", func() {
							k := key
							showReturnDialog(app, k.ID)
						})
						returnBtn.Importance = widget.MediumImportance
						actions.Add(returnBtn)
					}

					cellContainer.Add(container.NewCenter(actions))
				}
			}
		},
	)

	// Définir les largeurs de colonnes optimisées
	table.SetColumnWidth(0, 120) // Numéro
	table.SetColumnWidth(1, 350) // Description
	table.SetColumnWidth(2, 150) // Disponibilité
	table.SetColumnWidth(3, 300) // Emprunteurs (augmenté de 200 à 300)
	table.SetColumnWidth(4, 180) // Actions

	// Retourner le tableau dans un conteneur scrollable
	return container.NewScroll(table)
}

// showKeyDetails affiche les détails d'une clé
func showKeyDetails(app *App, keyID int) {
	// Récupérer les détails de la clé
	key, err := db.GetKeyByID(keyID)
	if err != nil {
		app.showError("Erreur", "Impossible de charger les détails de la clé")
		return
	}

	// Récupérer les emprunts actifs
	loans, _ := db.GetActiveLoansByKeyID(keyID)

	// Créer le contenu des détails
	detailsContent := container.NewVBox(
		widget.NewLabelWithStyle("Numéro:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(key.Number),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Description:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(key.Description),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Quantités:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Total: %d | Réserve: %d", key.QuantityTotal, key.QuantityReserve)),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Lieu de stockage:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(key.StorageLocation),
	)

	// Ajouter les emprunts actifs s'il y en a
	if len(loans) > 0 {
		detailsContent.Add(widget.NewSeparator())
		detailsContent.Add(widget.NewLabelWithStyle("Emprunts actifs:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		for _, loan := range loans {
			loanText := fmt.Sprintf("• %s - depuis le %s",
				loan.BorrowerName,
				loan.LoanDate.Format("02/01/2006"),
			)
			detailsContent.Add(widget.NewLabel(loanText))
		}
	}

	// Créer la popup
	var dialog *widget.PopUp

	closeBtn := widget.NewButton("Fermer", func() {
		app.window.Canvas().Overlays().Remove(dialog)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("📋 Détails de la Clé", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewScroll(detailsContent),
		widget.NewSeparator(),
		container.NewCenter(closeBtn),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(500, 400))
	dialog.Show()
}
