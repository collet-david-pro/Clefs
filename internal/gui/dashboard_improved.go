package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// showLoanFormImproved affiche le formulaire d'emprunt amélioré avec recherche
func showLoanFormImproved(app *App, availableKeys []db.Key, borrowers []db.Borrower, preselectedKeys []int) {
	// Sélection de l'emprunteur avec recherche
	borrowerOptions := make([]string, len(borrowers))
	borrowerMap := make(map[string]int)
	for i, b := range borrowers {
		borrowerOptions[i] = b.Name
		borrowerMap[b.Name] = b.ID
	}

	borrowerSelect := widget.NewSelect(borrowerOptions, nil)
	borrowerSelect.PlaceHolder = "Sélectionner un emprunteur..."
	if len(borrowerOptions) > 0 {
		borrowerSelect.SetSelected(borrowerOptions[0])
	}

	// Champ de recherche pour les clés
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher une clé (numéro ou description)...")

	// Sélection des clés (multi-sélection) avec checkboxes
	keyCheckboxes := make(map[int]*widget.Check)
	allKeys := availableKeys // Garder une copie de toutes les clés
	
	keySelectionBox := container.NewVBox()

	// Fonction pour mettre à jour l'affichage des clés
	updateKeyDisplay := func(query string) {
		keySelectionBox.Objects = nil
		
		for _, key := range allKeys {
			k := key // Capture de la variable
			
			// Filtrer par recherche
			if query == "" ||
				strings.Contains(strings.ToLower(k.Number), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(k.Description), strings.ToLower(query)) {
				
				// Créer ou récupérer la checkbox
				checkbox, exists := keyCheckboxes[k.ID]
				if !exists {
					checkbox = widget.NewCheck(fmt.Sprintf("%s - %s", k.Number, k.Description), nil)
					
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
				}
				
				keySelectionBox.Add(checkbox)
			}
		}
		keySelectionBox.Refresh()
	}

	// Initialiser l'affichage
	updateKeyDisplay("")

	// Mettre à jour lors de la recherche
	searchEntry.OnChanged = func(query string) {
		updateKeyDisplay(query)
	}

	// Scroll pour les clés
	keyScroll := container.NewVScroll(keySelectionBox)
	keyScroll.SetMinSize(fyne.NewSize(550, 300))

	// Compteur de clés sélectionnées
	selectedCountLabel := widget.NewLabel("0 clé(s) sélectionnée(s)")
	selectedCountLabel.TextStyle.Bold = true

	// Mettre à jour le compteur
	updateSelectedCount := func() {
		count := 0
		for _, checkbox := range keyCheckboxes {
			if checkbox.Checked {
				count++
			}
		}
		selectedCountLabel.SetText(fmt.Sprintf("%d clé(s) sélectionnée(s)", count))
	}

	// Ajouter l'événement OnChanged à toutes les checkboxes
	for _, checkbox := range keyCheckboxes {
		cb := checkbox
		cb.OnChanged = func(bool) {
			updateSelectedCount()
		}
	}

	// Formulaire
	form := container.NewVBox(
		widget.NewLabelWithStyle("Emprunteur:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		borrowerSelect,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Clés à emprunter:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		searchEntry,
		keyScroll,
		container.NewHBox(selectedCountLabel),
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
	dialog.Resize(fyne.NewSize(650, 600))
	dialog.Show()
}

// Remplacer showLoanForm par la version améliorée
func init() {
	// Cette fonction sera appelée au démarrage
	log.Println("Module dashboard amélioré chargé")
}
