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

// createNewLoanView crée la vue de prêt unifiée en une seule page.
// Tout est visible d'un coup : détenteur, portes à cocher, trousseau calculé, validation.
func createNewLoanView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Nouvel emprunt", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// --- État local ---
	borrowers, _ := db.GetAllBorrowers()
	accesses, _ := db.GetAllAccesses()
	buildings, _ := db.GetAllBuildings()

	// Map buildingID → nom
	bldgName := make(map[int]string)
	for _, b := range buildings {
		bldgName[b.ID] = b.Name
	}

	var selectedBorrowerID int
	checkedAccesses := make(map[int]bool)
	var finalKeys []db.Key
	var suggestedKeyIDs map[int]bool

	// --- Widgets persistants mis à jour dynamiquement ---
	trousseauBox := container.NewVBox()
	trousseauTitle := widget.NewLabelWithStyle("Trousseau calculé", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	alertBox := container.NewVBox()

	validateBtn := widget.NewButton("Valider et imprimer le bon", nil)
	validateBtn.Importance = widget.HighImportance
	validateBtn.Disable()

	// Champ date retour et type de prêt
	returnEntry := widget.NewEntry()
	returnEntry.SetPlaceHolder("JJ/MM/AAAA (optionnel)")
	loanTypeSelect := widget.NewSelect([]string{"ponctuel", "permanent"}, nil)
	loanTypeSelect.SetSelected("ponctuel")

	// --- Fonction de recalcul du trousseau ---
	// Déclaré via var pour permettre la récursion dans les closures de boutons
	var recalculate func()
	recalculate = func() {
		// Collecter les accès cochés
		var accessIDs []int
		for id, checked := range checkedAccesses {
			if checked {
				accessIDs = append(accessIDs, id)
			}
		}
		sort.Ints(accessIDs)

		trousseauBox.Objects = nil
		alertBox.Objects = nil
		finalKeys = nil
		suggestedKeyIDs = make(map[int]bool)

		if selectedBorrowerID == 0 || len(accessIDs) == 0 {
			validateBtn.Disable()
			trousseauBox.Add(widget.NewLabel("Sélectionnez un détenteur et des portes."))
			trousseauBox.Refresh()
			alertBox.Refresh()
			return
		}

		// Calcul algorithme greedy
		candidates, _ := business.BuildAvailableKeysForAccesses(accessIDs)
		result := business.SuggestKeys(accessIDs, candidates)

		// Alertes
		if result.HasUncoverable {
			names := accessIDsToNames(result.UncoverableIDs)
			alertBox.Add(widget.NewLabelWithStyle(
				"Accès non couverts par aucune clé disponible : "+strings.Join(names, ", "),
				fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
			))
		}

		// Redondances avec les clés déjà détenues
		existingLoans, _ := db.GetActiveLoansByBorrowerID(selectedBorrowerID)
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
			for _, k := range result.SelectedKeys {
				kc := k
				rooms, _ := db.GetRoomsForKey(kc.ID)
				existingKeys = append(existingKeys, kc)
				accessesByKey[kc.ID] = rooms
			}
			redundant := business.DetectRedundancies(existingKeys, accessesByKey)
			if len(redundant) > 0 {
				rnames := make([]string, len(redundant))
				for i, r := range redundant {
					rnames[i] = r.Name
				}
				alertBox.Add(widget.NewLabel("Redondance détectée : " + strings.Join(rnames, ", ")))
			}
		}

		// Remplir le trousseau
		finalKeys = make([]db.Key, len(result.SelectedKeys))
		copy(finalKeys, result.SelectedKeys)
		for _, k := range result.SelectedKeys {
			suggestedKeyIDs[k.ID] = true
		}

		for i := range finalKeys {
			k := finalKeys[i]
			keyID := k.ID // capture immuable
			rooms, _ := db.GetRoomsForKey(k.ID)
			coveredNames := make([]string, len(rooms))
			for j, r := range rooms {
				coveredNames[j] = r.Name
			}
			card := loanKeyCard(k, coveredNames, suggestedKeyIDs[k.ID], func() {
				newKeys := make([]db.Key, 0, len(finalKeys)-1)
				for _, fk := range finalKeys {
					if fk.ID != keyID {
						newKeys = append(newKeys, fk)
					}
				}
				finalKeys = newKeys
				recalculate()
			})
			trousseauBox.Add(card)
		}

		if len(finalKeys) == 0 {
			trousseauBox.Add(widget.NewLabel("Aucune clé disponible pour ces accès."))
			validateBtn.Disable()
		} else {
			validateBtn.Enable()
		}

		trousseauBox.Refresh()
		alertBox.Refresh()
	}

	// --- Sélection du détenteur ---
	borrowerNames := make([]string, len(borrowers))
	for i, b := range borrowers {
		borrowerNames[i] = b.Name
	}
	var borrowerSelect *widget.Select
	borrowerSelect = widget.NewSelect(borrowerNames, func(_ string) {
		idx := borrowerSelect.SelectedIndex()
		if idx >= 0 {
			selectedBorrowerID = borrowers[idx].ID
		}
		recalculate()
	})
	borrowerSelect.PlaceHolder = "Sélectionner un détenteur..."

	newBorrowerBtn := widget.NewButton("+ Nouveau détenteur", func() {
		showAddBorrowerDialog(a)
	})

	detenteurRow := container.NewBorder(nil, nil, nil, newBorrowerBtn, borrowerSelect)

	// --- Filtres sur les accès ---
	bOptions := []string{"Tous les bâtiments"}
	for _, b := range buildings {
		bOptions = append(bOptions, b.Name)
	}
	bFilter := widget.NewSelect(bOptions, nil)
	bFilter.SetSelectedIndex(0)

	// Liste des accès (reconstruite selon filtre)
	accessListBox := container.NewVBox()

	rebuildAccessList := func() {
		accessListBox.Objects = nil
		for _, r := range accesses {
			r := r
			bn := bldgName[r.BuildingID]
			if bFilter.Selected != "Tous les bâtiments" && bn != bFilter.Selected {
				continue
			}
			row := accessCheckRow(r, bn, checkedAccesses[r.ID], func(v bool) {
				checkedAccesses[r.ID] = v
				recalculate()
			})
			accessListBox.Add(row)
		}
		accessListBox.Refresh()
	}

	bFilter.OnChanged = func(_ string) { rebuildAccessList() }
	rebuildAccessList()

	// --- Section ajout manuel de clé ---
	allAvailable, _ := db.GetAvailableKeys()
	addKeySelect := widget.NewSelect(func() []string {
		opts := make([]string, len(allAvailable))
		for i, k := range allAvailable {
			opts[i] = fmt.Sprintf("%s — %s", k.Number, k.Description)
		}
		return opts
	}(), nil)
	addKeySelect.PlaceHolder = "Ajouter une clé manuellement..."

	addKeyBtn := widget.NewButton("Ajouter", func() {
		idx := addKeySelect.SelectedIndex()
		if idx < 0 {
			return
		}
		k := allAvailable[idx]
		// Vérifier si déjà dans le trousseau
		for _, fk := range finalKeys {
			if fk.ID == k.ID {
				return
			}
		}
		finalKeys = append(finalKeys, k)
		recalculate()
	})

	manualSection := container.NewBorder(nil, nil, nil, addKeyBtn, addKeySelect)

	// --- Bouton valider ---
	cancelBtn := widget.NewButton("Annuler", func() { a.showDashboard() })

	validateBtn.OnTapped = func() {
		if selectedBorrowerID == 0 {
			a.showError("Erreur", "Veuillez sélectionner un détenteur.")
			return
		}
		if len(finalKeys) == 0 {
			a.showError("Erreur", "Le trousseau est vide.")
			return
		}

		// Parser date retour
		var plannedReturn *time.Time
		if returnEntry.Text != "" {
			t, err := time.Parse("02/01/2006", returnEntry.Text)
			if err != nil {
				a.showError("Erreur", "Format de date invalide. Utilisez JJ/MM/AAAA.")
				return
			}
			plannedReturn = &t
		}

		keyIDs := make([]int, len(finalKeys))
		for i, k := range finalKeys {
			keyIDs[i] = k.ID
		}
		if err := db.CreateMultipleLoans(keyIDs, selectedBorrowerID); err != nil {
			a.showError("Erreur", fmt.Sprintf("Erreur lors de la création du prêt : %v", err))
			return
		}

		// Générer le bon PDF en arrière-plan
		go func() {
			borrower, _ := db.GetBorrowerByID(selectedBorrowerID)
			loans, _ := db.GetActiveLoansByBorrowerID(selectedBorrowerID)
			// Garder uniquement les prêts tout juste créés (les N derniers)
			var recentLoans []db.LoanWithDetails
			if len(loans) >= len(finalKeys) {
				recentLoans = loans[:len(finalKeys)]
			} else {
				recentLoans = loans
			}
			if borrower != nil && len(recentLoans) > 0 {
				// Collecter les accès couverts pour le bon
				var allAccesses []db.Room
				seen := map[int]struct{}{}
				for _, k := range finalKeys {
					rooms, _ := db.GetRoomsForKey(k.ID)
					for _, r := range rooms {
						if _, ok := seen[r.ID]; !ok {
							seen[r.ID] = struct{}{}
							// Passer le nom du bâtiment via le champ Notes (utilisé par le PDF)
							r.Notes = bldgName[r.BuildingID]
							allAccesses = append(allAccesses, r)
						}
					}
				}
				opts := pdf.BorrowerReceiptOptions{
					Agent:         "Administrateur",
					PlannedReturn: plannedReturn,
					LoanType:      loanTypeSelect.Selected,
					Accesses:      allAccesses,
				}
				pdfData, err := pdf.GenerateBorrowerReceipt(borrower, recentLoans, opts)
				if err == nil {
					filename := pdf.GenerateFilename("bon_remise_"+borrower.Name, 0)
					pdf.SavePDF(filename, pdfData)
				}
			}
		}()

		borrowerName := ""
		for _, b := range borrowers {
			if b.ID == selectedBorrowerID {
				borrowerName = b.Name
				break
			}
		}
		a.showSuccess(fmt.Sprintf("%d clé(s) remise(s) à %s.\nLe bon de remise est généré dans documents/.", len(finalKeys), borrowerName))
		a.showDashboard()
	}

	// --- Assemblage final ---
	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Détenteur", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			detenteurRow,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Portes / zones d'accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			bFilter,
		),
		nil, nil, nil,
		container.NewVScroll(accessListBox),
	)

	rightPanel := container.NewBorder(
		container.NewVBox(
			trousseauTitle,
			alertBox,
		),
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Ajouter une clé manuellement", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			manualSection,
			widget.NewSeparator(),
			container.NewGridWithColumns(2,
				widget.NewLabel("Retour prévu :"), returnEntry,
				widget.NewLabel("Type de prêt :"), loanTypeSelect,
			),
			widget.NewSeparator(),
			container.NewHBox(cancelBtn, validateBtn),
		),
		nil, nil,
		container.NewVScroll(trousseauBox),
	)

	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.5)

	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		nil, nil, nil,
		split,
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
