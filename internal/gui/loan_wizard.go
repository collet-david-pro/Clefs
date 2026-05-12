package gui

import (
	"clefs/internal/business"
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type loanWizardState struct {
	borrowerID      int
	borrowerName    string
	selectedAccess  []int // IDs des accès cochés
	suggestedKeys   []db.Key
	finalKeys       []db.Key
	plannedReturn   *time.Time
	loanType        string // "ponctuel" ou "permanent"
}

// showLoanWizard ouvre l'assistant de prêt en 3 étapes.
func showLoanWizard(a *App) {
	state := &loanWizardState{loanType: "ponctuel"}
	var popup *widget.PopUp

	content := container.NewMax()

	// showStep est déclaré via un pointeur de fonction pour permettre la récursion
	var showStep func(int)
	showStep = func(step int) {
		content.Objects = nil
		switch step {
		case 1:
			content.Add(wizardStep1(a, state,
				func() { showStep(2) },
				func() { a.window.Canvas().Overlays().Remove(popup) },
			))
		case 2:
			content.Add(wizardStep2(a, state,
				func() { showStep(3) },
				func() { showStep(1) },
				func() { a.window.Canvas().Overlays().Remove(popup) },
			))
		case 3:
			content.Add(wizardStep3(a, state, popup,
				func() { showStep(2) },
				func() {
					a.window.Canvas().Overlays().Remove(popup)
					a.showDashboard()
				},
			))
		}
		content.Refresh()
	}

	showStep(1)
	popup = widget.NewModalPopUp(container.NewPadded(content), a.window.Canvas())
	popup.Resize(fyne.NewSize(700, 550))
	popup.Show()
}

// --- Étape 1 : Sélection du détenteur ---

func wizardStep1(a *App, state *loanWizardState, onNext, onCancel func()) fyne.CanvasObject {
	borrowers, _ := db.GetAllBorrowers()

	names := make([]string, len(borrowers))
	for i, b := range borrowers {
		names[i] = b.Name
	}

	borrowerSelect := widget.NewSelect(names, nil)
	if state.borrowerID > 0 {
		for i, b := range borrowers {
			if b.ID == state.borrowerID {
				borrowerSelect.SetSelectedIndex(i)
				break
			}
		}
	}

	newBorrowerBtn := widget.NewButton("➕ Nouveau détenteur", func() {
		showAddBorrowerDialog(a)
	})

	nextBtn := widget.NewButton("Suivant →", func() {
		idx := borrowerSelect.SelectedIndex()
		if idx < 0 {
			a.showError("Erreur", "Veuillez sélectionner un détenteur.")
			return
		}
		state.borrowerID = borrowers[idx].ID
		state.borrowerName = borrowers[idx].Name
		onNext()
	})
	nextBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", onCancel)

	return container.NewVBox(
		widget.NewLabelWithStyle("Étape 1 / 3 — Détenteur", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Sélectionnez le détenteur qui reçoit les clés :"),
		borrowerSelect,
		newBorrowerBtn,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, nextBtn),
	)
}

// --- Étape 2 : Sélection des accès requis ---

func wizardStep2(a *App, state *loanWizardState, onNext, onBack, onCancel func()) fyne.CanvasObject {
	accesses, _ := db.GetAllAccesses()
	buildings, _ := db.GetAllBorrowers() // pour les noms — on charge les bâtiments séparément
	_ = buildings
	bldgs, _ := db.GetAllBuildings()

	// Filtres
	bOptions := []string{"Tous"}
	for _, b := range bldgs {
		bOptions = append(bOptions, b.Name)
	}
	bFilter := widget.NewSelect(bOptions, nil)
	bFilter.SetSelectedIndex(0)

	// Map ID → coché
	checked := make(map[int]bool)
	for _, id := range state.selectedAccess {
		checked[id] = true
	}

	listBox := container.NewVBox()

	rebuildList := func() {
		listBox.Objects = nil
		for _, r := range accesses {
			r := r
			if bFilter.Selected != "Tous" {
				bName := ""
				for _, b := range bldgs {
					if b.ID == r.BuildingID {
						bName = b.Name
						break
					}
				}
				if bName != bFilter.Selected {
					continue
				}
			}
			bName := ""
			for _, b := range bldgs {
				if b.ID == r.BuildingID {
					bName = b.Name
					break
				}
			}
			label := fmt.Sprintf("%s  [%s%s]", r.Name, bName, func() string {
				if r.Floor != "" {
					return " — " + r.Floor
				}
				return ""
			}())
			chk := widget.NewCheck(label, func(v bool) {
				checked[r.ID] = v
			})
			chk.SetChecked(checked[r.ID])
			listBox.Add(chk)
		}
		listBox.Refresh()
	}
	bFilter.OnChanged = func(_ string) { rebuildList() }
	rebuildList()

	// Type de prêt
	loanTypeSelect := widget.NewSelect([]string{"ponctuel", "permanent"}, func(v string) {
		state.loanType = v
	})
	loanTypeSelect.SetSelected(state.loanType)

	// Date de retour prévue (saisie texte DD/MM/YYYY)
	returnEntry := widget.NewEntry()
	returnEntry.SetPlaceHolder("JJ/MM/AAAA (optionnel)")
	if state.plannedReturn != nil {
		returnEntry.SetText(state.plannedReturn.Format("02/01/2006"))
	}

	nextBtn := widget.NewButton("Suivant →", func() {
		// Collecter les accès cochés
		state.selectedAccess = nil
		for id, v := range checked {
			if v {
				state.selectedAccess = append(state.selectedAccess, id)
			}
		}
		sort.Ints(state.selectedAccess)

		if len(state.selectedAccess) == 0 {
			a.showError("Erreur", "Sélectionnez au moins un accès.")
			return
		}

		// Parser la date si fournie
		state.plannedReturn = nil
		if returnEntry.Text != "" {
			t, err := time.Parse("02/01/2006", returnEntry.Text)
			if err != nil {
				a.showError("Erreur", "Format de date invalide. Utilisez JJ/MM/AAAA.")
				return
			}
			state.plannedReturn = &t
		}

		// Calculer la suggestion de clés
		candidates, err := business.BuildAvailableKeysForAccesses(state.selectedAccess)
		if err != nil {
			a.showError("Erreur", fmt.Sprintf("Erreur lors du calcul: %v", err))
			return
		}
		result := business.SuggestKeys(state.selectedAccess, candidates)
		state.suggestedKeys = result.SelectedKeys
		state.finalKeys = make([]db.Key, len(result.SelectedKeys))
		copy(state.finalKeys, result.SelectedKeys)

		onNext()
	})
	nextBtn.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabelWithStyle("Étape 2 / 3 — Accès requis", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Filtrer par bâtiment :"), bFilter,
		),
		container.NewVScroll(listBox),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Type de prêt :"), loanTypeSelect,
			widget.NewLabel("Retour prévu :"), returnEntry,
		),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewButton("← Retour", onBack),
			widget.NewButton("Annuler", onCancel),
			nextBtn,
		),
	)
}

// --- Étape 3 : Proposition de clés + validation ---

func wizardStep3(a *App, state *loanWizardState, popup *widget.PopUp, onBack, onDone func()) fyne.CanvasObject {
	// Recalculer la suggestion pour l'affichage
	candidates, _ := business.BuildAvailableKeysForAccesses(state.selectedAccess)
	result := business.SuggestKeys(state.selectedAccess, candidates)

	// Bandeaux d'alerte
	alerts := container.NewVBox()
	if result.HasUncoverable {
		uncoverableNames := accessIDsToNames(result.UncoverableIDs)
		warn := widget.NewLabelWithStyle(
			"⚠️ Accès non couverts : "+strings.Join(uncoverableNames, ", "),
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
		)
		alerts.Add(container.NewPadded(warn))
	}

	// Vérifier redondances avec les clés déjà en possession du détenteur
	existingLoans, _ := db.GetActiveLoansByBorrowerID(state.borrowerID)
	if len(existingLoans) > 0 {
		accessesByKey := map[int][]db.Room{}
		var existingKeys []db.Key
		for _, l := range existingLoans {
			k, err := db.GetKeyByID(l.KeyID)
			if err != nil {
				continue
			}
			existingKeys = append(existingKeys, *k)
			rooms, _ := db.GetRoomsForKey(l.KeyID)
			accessesByKey[l.KeyID] = rooms
		}
		// Ajouter aussi les clés suggérées
		for _, k := range state.finalKeys {
			k := k
			rooms, _ := db.GetRoomsForKey(k.ID)
			existingKeys = append(existingKeys, k)
			accessesByKey[k.ID] = rooms
		}
		redundant := business.DetectRedundancies(existingKeys, accessesByKey)
		if len(redundant) > 0 {
			names := make([]string, len(redundant))
			for i, r := range redundant {
				names[i] = r.Name
			}
			warn := widget.NewLabelWithStyle(
				"ℹ️ Redondance détectée : "+strings.Join(names, ", "),
				fyne.TextAlignLeading, fyne.TextStyle{},
			)
			alerts.Add(container.NewPadded(warn))
		}
	}

	// Liste des clés finales (modifiable par checkboxes)
	allAvailable, _ := db.GetAvailableKeys()
	keyChecked := make(map[int]bool)
	for _, k := range state.finalKeys {
		keyChecked[k.ID] = true
	}

	keyList := container.NewVBox()
	for _, k := range allAvailable {
		k := k
		isSuggested := false
		for _, sk := range result.SelectedKeys {
			if sk.ID == k.ID {
				isSuggested = true
				break
			}
		}
		label := fmt.Sprintf("%s — %s", k.Number, k.Description)
		if isSuggested {
			label += " ✓"
		}
		chk := widget.NewCheck(label, func(v bool) {
			keyChecked[k.ID] = v
			// Mettre à jour finalKeys
			state.finalKeys = nil
			for id, checked := range keyChecked {
				if checked {
					for _, av := range allAvailable {
						if av.ID == id {
							state.finalKeys = append(state.finalKeys, av)
							break
						}
					}
				}
			}
		})
		chk.SetChecked(keyChecked[k.ID])
		keyList.Add(chk)
	}

	validateBtn := widget.NewButton("✅ Valider et imprimer le bon", func() {
		if len(state.finalKeys) == 0 {
			a.showError("Erreur", "Sélectionnez au moins une clé.")
			return
		}
		keyIDs := make([]int, len(state.finalKeys))
		for i, k := range state.finalKeys {
			keyIDs[i] = k.ID
		}
		if err := db.CreateMultipleLoans(keyIDs, state.borrowerID); err != nil {
			a.showError("Erreur", fmt.Sprintf("Erreur lors de la création des prêts: %v", err))
			return
		}

		// Générer le bon PDF
		go func() {
			borrower, _ := db.GetBorrowerByID(state.borrowerID)
			loans, _ := db.GetActiveLoansByBorrowerID(state.borrowerID)
			// Garder seulement les prêts qu'on vient de créer (les derniers)
			recentLoans := loans[:len(state.finalKeys)]
			if borrower != nil && len(recentLoans) > 0 {
				pdfData, err := pdf.GenerateBorrowerReceipt(borrower, recentLoans)
				if err == nil {
					filename := pdf.GenerateFilename("bon_remise_"+borrower.Name, 0)
					pdf.SavePDF(filename, pdfData)
				}
			}
		}()

		a.window.Canvas().Overlays().Remove(popup)
		a.showSuccess(fmt.Sprintf("✅ %d clé(s) remise(s) à %s.\nLe bon de remise a été généré dans le dossier documents/.",
			len(state.finalKeys), state.borrowerName))
		onDone()
	})
	validateBtn.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabelWithStyle("Étape 3 / 3 — Clés à remettre", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Détenteur : %s  |  %d accès demandé(s)", state.borrowerName, len(state.selectedAccess))),
		widget.NewSeparator(),
		alerts,
		widget.NewLabel("Clés proposées (modifiables) :"),
		container.NewVScroll(keyList),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewButton("← Retour", onBack),
			widget.NewButton("Annuler", func() { a.window.Canvas().Overlays().Remove(popup) }),
			validateBtn,
		),
	)
}

// accessIDsToNames convertit des IDs d'accès en noms lisibles
func accessIDsToNames(ids []int) []string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		room, err := db.GetRoomByID(id)
		if err == nil {
			names = append(names, room.Name)
		} else {
			names = append(names, fmt.Sprintf("Accès #%d", id))
		}
	}
	return names
}
