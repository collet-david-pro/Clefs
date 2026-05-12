package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createModernDashboard crée un tableau de bord moderne avec des cards et statistiques
func createModernDashboard(app *App) fyne.CanvasObject {
	// En-tête simplifié
	titleLabel := widget.NewLabelWithStyle("Tableau de Bord", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	newLoanBtn := widget.NewButton("Nouvel emprunt", func() {
		app.showNewLoan()
	})
	newLoanBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButton("Rafraîchir", func() {
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
	topContent := container.NewVBox(
		container.NewPadded(header),
		widget.NewSeparator(),
		container.NewPadded(statsCards),
		widget.NewSeparator(),
	)

	// Utiliser un Border layout pour que le tableau prenne tout l'espace restant
	content := container.NewBorder(
		topContent,                     // Haut
		nil,                            // Bas
		nil,                            // Gauche
		nil,                            // Droite
		container.NewPadded(keysTable), // Centre (le tableau)
	)

	return content
}

// getStatistics récupère les statistiques pour le dashboard
func getStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	keys, _ := db.GetAllKeys()
	stats["totalKeys"] = len(keys)

	activeLoans, _ := db.GetAllActiveLoans()
	stats["activeLoans"] = len(activeLoans)

	availableKeys, _ := db.GetAvailableKeys()
	stats["availableKeys"] = len(availableKeys)

	borrowers, _ := db.GetAllBorrowers()
	stats["totalBorrowers"] = len(borrowers)

	overdueLoans, _ := db.GetOverdueLoans()
	stats["overdueLoans"] = len(overdueLoans)

	redundancies, _ := db.GetBorrowersWithRedundantAccesses()
	stats["redundancies"] = len(redundancies)

	return stats
}

// createStatisticsCards crée les cards de statistiques avec alertes retards et redondances
func createStatisticsCards(stats map[string]interface{}) fyne.CanvasObject {
	totalKeysLabel := widget.NewLabel(fmt.Sprintf("🔑 Total clés: %d", stats["totalKeys"]))
	activeLoansLabel := widget.NewLabel(fmt.Sprintf("📤 Emprunts actifs: %d", stats["activeLoans"]))
	availableKeysLabel := widget.NewLabel(fmt.Sprintf("✅ Disponibles: %d", stats["availableKeys"]))
	borrowersLabel := widget.NewLabel(fmt.Sprintf("👥 Détenteurs: %d", stats["totalBorrowers"]))

	row1 := container.NewHBox(
		totalKeysLabel, widget.NewSeparator(),
		activeLoansLabel, widget.NewSeparator(),
		availableKeysLabel, widget.NewSeparator(),
		borrowersLabel,
	)

	// Alertes — visibles seulement si > 0
	alertsBox := container.NewVBox()
	if n, _ := stats["overdueLoans"].(int); n > 0 {
		lbl := widget.NewLabelWithStyle(
			fmt.Sprintf("🔴 %d prêt(s) EN RETARD — vérifiez l'onglet Emprunts en Cours", n),
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
		)
		alertsBox.Add(lbl)
	}
	if n, _ := stats["redundancies"].(int); n > 0 {
		lbl := widget.NewLabel(fmt.Sprintf("⚠️ %d détenteur(s) avec des accès redondants", n))
		alertsBox.Add(lbl)
	}

	return container.NewVBox(container.NewCenter(row1), alertsBox)
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
					// Emprunteurs - Affichage du nombre uniquement
					count := len(key.BorrowerNames)
					var text string

					if count == 0 {
						text = "--"
					} else if count == 1 {
						text = "1 emprunt"
					} else {
						text = fmt.Sprintf("%d emprunts", count)
					}

					label := widget.NewLabel(text)
					label.Alignment = fyne.TextAlignCenter
					cellContainer.Add(label)

				case 4:
					// Actions avec positions fixes
					// On utilise une grille à 2 colonnes pour garantir que les boutons
					// restent à la même place (gauche pour Emprunter, droite pour Retourner)

					var borrowObj fyne.CanvasObject
					if key.AvailableCount > 0 {
						borrowBtn := widget.NewButton("Emprunter", func() {
							app.showNewLoan()
						})
						borrowBtn.Importance = widget.HighImportance
						borrowObj = borrowBtn
					} else {
						// Espace vide pour maintenir l'alignement
						borrowObj = layout.NewSpacer()
					}

					var returnObj fyne.CanvasObject
					if key.LoanedCount > 0 {
						returnBtn := widget.NewButton("Retourner", func() {
							k := key
							showReturnDialog(app, k.ID)
						})
						returnBtn.Importance = widget.MediumImportance
						returnObj = returnBtn
					} else {
						// Espace vide pour maintenir l'alignement
						returnObj = layout.NewSpacer()
					}

					// Conteneur grille 2 colonnes
					actions := container.NewGridWithColumns(2, borrowObj, returnObj)

					// On centre le tout, mais avec une largeur fixe suffisante si possible
					// ou on laisse le Grid gérer l'espace
					cellContainer.Add(actions)
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
