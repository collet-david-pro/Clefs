package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createDashboard crée la vue du tableau de bord
func createDashboard(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Tableau de Bord", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.TextStyle.Bold = true

	// Bouton pour créer un nouvel emprunt
	newLoanBtn := widget.NewButton("➕ Nouvel Emprunt", func() {
		showNewLoanDialog(app)
	})
	newLoanBtn.Importance = widget.HighImportance

	header := container.NewBorder(nil, nil, nil, newLoanBtn, title)

	// Récupérer les clés avec disponibilité
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		log.Printf("Erreur lors de la récupération des clés: %v", err)
		return container.NewVBox(
			header,
			widget.NewLabel("Erreur lors du chargement des données"),
		)
	}

	// Créer le tableau
	table := createKeysTable(keys, app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(table),
	)

	return content
}

// createKeysTable crée le tableau des clés pour le tableau de bord
func createKeysTable(keys []db.KeyWithAvailability, app *App) fyne.CanvasObject {
	if len(keys) == 0 {
		return widget.NewLabel("Aucune clé disponible")
	}

	// Créer un tableau avec widget.Table pour un meilleur alignement
	table := widget.NewTable(
		func() (int, int) {
			return len(keys) + 1, 5 // +1 pour l'en-tête, 5 colonnes
		},
		func() fyne.CanvasObject {
			// Template pour les cellules
			return container.NewMax(widget.NewLabel(""))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cellContainer := cell.(*fyne.Container)
			cellContainer.Objects = nil // Vider le container

			if id.Row == 0 {
				// En-têtes
				var label *widget.Label
				switch id.Col {
				case 0:
					label = widget.NewLabelWithStyle("Numéro", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				case 1:
					label = widget.NewLabelWithStyle("Description", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				case 2:
					label = widget.NewLabelWithStyle("Disponibilité", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
				case 3:
					label = widget.NewLabelWithStyle("Emprunté Par", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				case 4:
					label = widget.NewLabelWithStyle("Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
				}
				cellContainer.Add(label)
			} else {
				// Données
				key := keys[id.Row-1]
				switch id.Col {
				case 0:
					// Numéro
					label := widget.NewLabelWithStyle(key.Number, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
					cellContainer.Add(label)
				case 1:
					// Description
					label := widget.NewLabel(key.Description)
					cellContainer.Add(label)
				case 2:
					// Disponibilité
					usable := key.QuantityTotal - key.QuantityReserve
					availText := fmt.Sprintf("%d / %d", key.AvailableCount, usable)
					label := widget.NewLabel(availText)
					if key.AvailableCount > 0 {
						label.Importance = widget.SuccessImportance
					} else {
						label.Importance = widget.DangerImportance
					}
					cellContainer.Add(container.NewCenter(label))
				case 3:
					// Emprunteurs
					borrowersText := "--"
					if len(key.BorrowerNames) > 0 {
						borrowersText = ""
						for i, name := range key.BorrowerNames {
							if i > 0 {
								borrowersText += ", "
							}
							borrowersText += name
						}
					}
					label := widget.NewLabel(borrowersText)
					label.Wrapping = fyne.TextWrapWord
					cellContainer.Add(label)
				case 4:
					// Actions
					actions := container.NewHBox()
					if key.AvailableCount > 0 {
						borrowBtn := widget.NewButton("Emprunter", func() {
							k := key // Capture de la variable
							showNewLoanDialogWithKey(app, k.ID)
						})
						borrowBtn.Importance = widget.HighImportance
						actions.Add(borrowBtn)
					}
					if key.LoanedCount > 0 {
						returnBtn := widget.NewButton("Retourner", func() {
							k := key // Capture de la variable
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

	// Définir les largeurs de colonnes
	table.SetColumnWidth(0, 100) // Numéro
	table.SetColumnWidth(1, 300) // Description
	table.SetColumnWidth(2, 120) // Disponibilité
	table.SetColumnWidth(3, 250) // Emprunté Par
	table.SetColumnWidth(4, 250) // Actions

	return table
}

// showNewLoanDialog affiche la boîte de dialogue pour créer un nouvel emprunt
func showNewLoanDialog(app *App) {
	// Récupérer les clés disponibles
	availableKeys, err := db.GetAvailableKeys()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des clés: %v", err))
		return
	}

	if len(availableKeys) == 0 {
		app.showError("Aucune clé disponible", "Toutes les clés sont actuellement empruntées.")
		return
	}

	// Récupérer les emprunteurs
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunteurs: %v", err))
		return
	}

	if len(borrowers) == 0 {
		app.showError("Aucun emprunteur", "Veuillez d'abord créer un emprunteur.")
		return
	}

	// Créer le formulaire amélioré
	showLoanFormImproved(app, availableKeys, borrowers, nil)
}

// showNewLoanDialogWithKey affiche la boîte de dialogue pour créer un emprunt avec une clé présélectionnée
func showNewLoanDialogWithKey(app *App, keyID int) {
	// Récupérer les clés disponibles
	availableKeys, err := db.GetAvailableKeys()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des clés: %v", err))
		return
	}

	// Récupérer les emprunteurs
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunteurs: %v", err))
		return
	}

	if len(borrowers) == 0 {
		app.showError("Aucun emprunteur", "Veuillez d'abord créer un emprunteur.")
		return
	}

	// Créer le formulaire amélioré avec la clé présélectionnée
	preselectedKeys := []int{keyID}
	showLoanFormImproved(app, availableKeys, borrowers, preselectedKeys)
}

// showLoanForm affiche le formulaire d'emprunt
func showLoanForm(app *App, availableKeys []db.Key, borrowers []db.Borrower, preselectedKeys []int) {
	// Sélection des clés (multi-sélection)
	keyCheckboxes := make(map[int]*widget.Check)
	keySelectionBox := container.NewVBox()

	for _, key := range availableKeys {
		k := key // Capture de la variable
		checkbox := widget.NewCheck(fmt.Sprintf("%s - %s", k.Number, k.Description), nil)

		// Présélectionner si nécessaire
		if preselectedKeys != nil {
			for _, preselectedID := range preselectedKeys {
				if k.ID == preselectedID {
					checkbox.Checked = true
					break
				}
			}
		}

		keyCheckboxes[k.ID] = checkbox
		keySelectionBox.Add(checkbox)
	}

	// Sélection de l'emprunteur
	borrowerOptions := make([]string, len(borrowers))
	borrowerMap := make(map[string]int)
	for i, b := range borrowers {
		borrowerOptions[i] = b.Name
		borrowerMap[b.Name] = b.ID
	}

	borrowerSelect := widget.NewSelect(borrowerOptions, nil)
	if len(borrowerOptions) > 0 {
		borrowerSelect.SetSelected(borrowerOptions[0])
	}

	// Formulaire
	form := container.NewVBox(
		widget.NewLabel("Sélectionnez les clés à emprunter:"),
		container.NewVScroll(keySelectionBox),
		widget.NewSeparator(),
		widget.NewLabel("Emprunteur:"),
		borrowerSelect,
	)

	// Boutons
	var dialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(dialog)
	})

	confirmBtn := widget.NewButton("Créer l'emprunt", func() {
		// Récupérer les clés sélectionnées
		var selectedKeyIDs []int
		for keyID, checkbox := range keyCheckboxes {
			if checkbox.Checked {
				selectedKeyIDs = append(selectedKeyIDs, keyID)
			}
		}

		if len(selectedKeyIDs) == 0 {
			app.showError("Erreur", "Veuillez sélectionner au moins une clé.")
			return
		}

		if borrowerSelect.Selected == "" {
			app.showError("Erreur", "Veuillez sélectionner un emprunteur.")
			return
		}

		borrowerID := borrowerMap[borrowerSelect.Selected]

		// Créer les emprunts
		err := db.CreateMultipleLoans(selectedKeyIDs, borrowerID)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création de l'emprunt: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Emprunt créé avec succès!")
		app.showDashboard() // Rafraîchir
	})
	confirmBtn.Importance = widget.HighImportance

	buttons := container.NewHBox(cancelBtn, confirmBtn)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Nouvel Emprunt", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		buttons,
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(600, 500))
	dialog.Show()
}

// showReturnDialog affiche la boîte de dialogue pour retourner une clé
func showReturnDialog(app *App, keyID int) {
	// Récupérer les emprunts actifs pour cette clé
	loans, err := db.GetActiveLoansByKeyID(keyID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunts: %v", err))
		return
	}

	if len(loans) == 0 {
		app.showError("Erreur", "Aucun emprunt actif pour cette clé.")
		return
	}

	// Si un seul emprunt, retourner directement
	if len(loans) == 1 {
		app.showConfirm("Confirmer le retour",
			fmt.Sprintf("Confirmer le retour de la clé %s empruntée par %s?",
				loans[0].KeyNumber, loans[0].BorrowerName),
			func() {
				err := db.ReturnLoan(loans[0].ID)
				if err != nil {
					app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
					return
				}
				app.showSuccess("Clé retournée avec succès!")
				app.showDashboard()
			})
		return
	}

	// Plusieurs emprunts : afficher une liste de sélection
	showReturnSelectionDialog(app, loans)
}

// showReturnSelectionDialog affiche la sélection d'emprunt à retourner
func showReturnSelectionDialog(app *App, loans []db.LoanWithDetails) {
	loanOptions := make([]string, len(loans))
	loanMap := make(map[string]int)

	for i, loan := range loans {
		option := fmt.Sprintf("%s - %s (%s)", loan.KeyNumber, loan.BorrowerName, loan.LoanDate.Format("02/01/2006"))
		loanOptions[i] = option
		loanMap[option] = loan.ID
	}

	loanSelect := widget.NewSelect(loanOptions, nil)
	if len(loanOptions) > 0 {
		loanSelect.SetSelected(loanOptions[0])
	}

	var dialog *widget.PopUp

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(dialog)
	})

	confirmBtn := widget.NewButton("Retourner", func() {
		if loanSelect.Selected == "" {
			app.showError("Erreur", "Veuillez sélectionner un emprunt.")
			return
		}

		loanID := loanMap[loanSelect.Selected]
		err := db.ReturnLoan(loanID)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors du retour: %v", err))
			return
		}

		app.window.Canvas().Overlays().Remove(dialog)
		app.showSuccess("Clé retournée avec succès!")
		app.showDashboard()
	})
	confirmBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Sélectionner l'emprunt à retourner", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Plusieurs emprunts actifs pour cette clé:"),
		loanSelect,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, confirmBtn),
	)

	dialog = widget.NewModalPopUp(content, app.window.Canvas())
	dialog.Resize(fyne.NewSize(500, 250))
	dialog.Show()
}
