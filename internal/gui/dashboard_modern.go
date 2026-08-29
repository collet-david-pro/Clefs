package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createModernDashboard crée le tableau de bord : en-tête avec actions,
// rangée de tuiles de statistiques, bandeaux d'alerte éventuels, puis le
// tableau des clés qui occupe l'espace restant.
func createModernDashboard(app *App) fyne.CanvasObject {
	newLoanBtn := widget.NewButtonWithIcon("Nouvel emprunt", theme.ContentAddIcon(), func() {
		app.showNewLoan()
	})
	newLoanBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButtonWithIcon("Rafraîchir", theme.ViewRefreshIcon(), func() {
		app.showDashboard()
	})

	header := pageHeader("Tableau de bord", "Vue d'ensemble du parc de clés",
		container.NewHBox(refreshBtn, newLoanBtn))

	// Récupérer les statistiques
	stats := getStatistics()
	statsRow := createStatisticsCards(stats)

	// Récupérer les clés avec disponibilité
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		log.Printf("Erreur lors de la récupération des clés: %v", err)
		return container.NewVBox(
			header,
			widget.NewLabel("Erreur lors du chargement des données"),
		)
	}

	keysTable := createSimpleKeysTable(keys, app)
	tableTitle := canvas.NewText("Inventaire des clés", colorText)
	tableTitle.TextStyle = fyne.TextStyle{Bold: true}
	tableTitle.TextSize = 16

	topContent := container.NewVBox(
		header,
		widget.NewLabel(""),
		statsRow,
		widget.NewLabel(""),
		tableTitle,
	)

	// Le tableau (dans une carte) prend tout l'espace vertical restant.
	content := container.NewBorder(
		topContent,
		nil, nil, nil,
		card(keysTable),
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

	anomalies, _ := db.CheckInventoryAnomalies()
	stats["inventoryAnomalies"] = anomalies

	return stats
}

// createStatisticsCards crée la rangée de tuiles de statistiques (4 colonnes)
// suivie des éventuels bandeaux d'alerte (retards, redondances).
func createStatisticsCards(stats map[string]interface{}) fyne.CanvasObject {
	tiles := container.NewGridWithColumns(4,
		statTile(fmt.Sprintf("%d", stats["totalKeys"]), "Clés au total", colorPrimary),
		statTile(fmt.Sprintf("%d", stats["activeLoans"]), "Emprunts actifs", colorWarning),
		statTile(fmt.Sprintf("%d", stats["availableKeys"]), "Disponibles", colorSuccess),
		statTile(fmt.Sprintf("%d", stats["totalBorrowers"]), "Détenteurs", colorTextMuted),
	)

	box := container.NewVBox(tiles)

	// Alertes — affichées en bandeaux colorés seulement si > 0.
	if anomalies, _ := stats["inventoryAnomalies"].([]db.InventoryAnomaly); len(anomalies) > 0 {
		box.Add(widget.NewLabel(""))
		box.Add(inventoryAlertBanner(anomalies))
	}
	if n, _ := stats["overdueLoans"].(int); n > 0 {
		icon := widget.NewIcon(theme.WarningIcon())
		msg := canvas.NewText(
			fmt.Sprintf("%d prêt(s) en retard — voir « Emprunts en cours »", n), colorDanger)
		msg.TextStyle = fyne.TextStyle{Bold: true}
		box.Add(widget.NewLabel(""))
		box.Add(coloredCard(container.NewHBox(icon, msg), colorDangerSoft, colorDanger))
	}
	if n, _ := stats["redundancies"].(int); n > 0 {
		icon := widget.NewIcon(theme.InfoIcon())
		msg := canvas.NewText(
			fmt.Sprintf("%d détenteur(s) avec des accès redondants", n), colorWarning)
		box.Add(widget.NewLabel(""))
		box.Add(coloredCard(container.NewHBox(icon, msg), colorWarningSoft, colorWarning))
	}

	return box
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
