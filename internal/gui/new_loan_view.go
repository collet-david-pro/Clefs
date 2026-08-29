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
	excludedKeyIDs := make(map[int]bool) // clés explicitement retirées — exclues du recalcul
	manualKeyIDs := make(map[int]bool)   // clés ajoutées manuellement — toujours conservées

	// --- Widgets persistants mis à jour dynamiquement ---
	trousseauBox := container.NewVBox()
	trousseauTitle := widget.NewLabelWithStyle("Trousseau calculé", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	alertBox := container.NewVBox()

	// doValidate est défini plus bas — le bouton sera recréé avec cette fonction
	// après que toutes les variables sont en scope.
	// On utilise un pointeur de fonction pour éviter le problème de forward reference.
	var doValidate func()
	validateBtn := widget.NewButton("Valider et imprimer le bon", func() {
		if doValidate != nil {
			doValidate()
		}
	})
	validateBtn.Importance = widget.HighImportance
	validateBtn.Disable()

	// Champ date retour et type de prêt
	returnEntry := widget.NewEntry()
	returnEntry.SetPlaceHolder("JJ/MM/AAAA (optionnel)")
	loanTypeSelect := widget.NewSelect([]string{"ponctuel", "permanent"}, nil)
	loanTypeSelect.SetSelected("ponctuel")

	// --- Fonction de recalcul du trousseau ---
	var recalculate func()
	recalculate = func() {
		var accessIDs []int
		for id, checked := range checkedAccesses {
			if checked {
				accessIDs = append(accessIDs, id)
			}
		}
		sort.Ints(accessIDs)

		trousseauBox.Objects = nil
		alertBox.Objects = nil
		suggestedKeyIDs = make(map[int]bool)

		if selectedBorrowerID == 0 || len(accessIDs) == 0 {
			finalKeys = nil
			validateBtn.Disable()
			validateBtn.Refresh()
			trousseauBox.Add(widget.NewLabel("Sélectionnez un détenteur et des portes."))
			trousseauBox.Refresh()
			alertBox.Refresh()
			return
		}

		// Calcul greedy en excluant les clés explicitement retirées
		candidates, _ := business.BuildAvailableKeysForAccesses(accessIDs)
		var filteredCandidates []db.KeyWithCoverage
		for _, c := range candidates {
			if !excludedKeyIDs[c.ID] {
				filteredCandidates = append(filteredCandidates, c)
			}
		}
		result := business.SuggestKeys(accessIDs, filteredCandidates)

		// Construire le trousseau final :
		// 1. clés suggérées par l'algo (hors exclues)
		// 2. clés ajoutées manuellement (toujours conservées)
		keySet := make(map[int]db.Key)
		for _, k := range result.SelectedKeys {
			keySet[k.ID] = k
			suggestedKeyIDs[k.ID] = true
		}
		for id := range manualKeyIDs {
			// Retrouver la clé dans les candidats disponibles
			for _, c := range candidates {
				if c.ID == id {
					keySet[id] = c.Key
					break
				}
			}
		}
		finalKeys = nil
		for _, k := range keySet {
			finalKeys = append(finalKeys, k)
		}
		// Tri stable par numéro
		sort.Slice(finalKeys, func(i, j int) bool {
			return finalKeys[i].Number < finalKeys[j].Number
		})

		// Alerte accès non couverts
		if result.HasUncoverable {
			names := accessIDsToNames(result.UncoverableIDs)
			alertBox.Add(widget.NewLabelWithStyle(
				"Accès non couverts par aucune clé disponible : "+strings.Join(names, ", "),
				fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
			))
		}

		// Redondances avec les clés déjà détenues par ce détenteur
		existingLoans, _ := db.GetActiveLoansByBorrowerID(selectedBorrowerID)
		if len(existingLoans) > 0 {
			accessesByKey := map[int][]db.Room{}
			var allKeys []db.Key
			for _, l := range existingLoans {
				k, err := db.GetKeyByID(l.KeyID)
				if err != nil {
					continue
				}
				allKeys = append(allKeys, *k)
				rooms, _ := db.GetRoomsForKey(l.KeyID)
				accessesByKey[l.KeyID] = rooms
			}
			for _, k := range finalKeys {
				kc := k
				rooms, _ := db.GetRoomsForKey(kc.ID)
				allKeys = append(allKeys, kc)
				accessesByKey[kc.ID] = rooms
			}
			redundant := business.DetectRedundancies(allKeys, accessesByKey)
			if len(redundant) > 0 {
				rnames := make([]string, len(redundant))
				for i, r := range redundant {
					rnames[i] = r.Name
				}
				alertBox.Add(widget.NewLabel("Redondance détectée : " + strings.Join(rnames, ", ")))
			}
		}

		// Afficher les cards du trousseau
		for i := range finalKeys {
			k := finalKeys[i]
			keyID := k.ID
			rooms, _ := db.GetRoomsForKey(k.ID)
			coveredNames := make([]string, len(rooms))
			for j, r := range rooms {
				coveredNames[j] = r.Name
			}
			isSuggested := suggestedKeyIDs[k.ID]
			card := loanKeyCard(k, coveredNames, isSuggested, func() {
				// Marquer comme exclue (si suggérée) ou simplement retirer (si manuelle)
				if suggestedKeyIDs[keyID] {
					excludedKeyIDs[keyID] = true
				}
				delete(manualKeyIDs, keyID)
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
		validateBtn.Refresh()

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
				// Changer les accès remet les exclusions à zéro — nouvelle situation
				excludedKeyIDs = make(map[int]bool)
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
		// Si cette clé était exclue, on lève l'exclusion
		delete(excludedKeyIDs, k.ID)
		// Marquer comme manuelle — sera conservée même si l'algo ne la choisit pas
		manualKeyIDs[k.ID] = true
		recalculate()
	})

	manualSection := container.NewBorder(nil, nil, nil, addKeyBtn, addKeySelect)

	// --- Bouton valider ---
	cancelBtn := widget.NewButton("Annuler", func() { a.showDashboard() })

	doValidate = func() {
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
		nKeys := len(finalKeys) // capture avant la goroutine
		go func() {
			borrower, _ := db.GetBorrowerByID(selectedBorrowerID)
			loans, _ := db.GetActiveLoansByBorrowerID(selectedBorrowerID)
			// Les prêts sont triés par loan_date croissant — les N derniers sont ceux qu'on vient de créer
			var recentLoans []db.LoanWithDetails
			if len(loans) >= nKeys {
				recentLoans = loans[len(loans)-nKeys:]
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
		// Afficher la confirmation PUIS naviguer au dashboard quand l'utilisateur clique OK.
		// Ne pas appeler showDashboard() dans le même call stack que le clic bouton
		// car Fyne détruirait la vue courante pendant le traitement de l'événement.
		confirmBox := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("%d clé(s) remise(s) à %s.\nLe bon de remise est généré dans documents/.", len(finalKeys), borrowerName)),
		)
		// Le prêt vient peut-être de faire passer une clé en sur-prêt : le
		// signaler dans la confirmation, sans remettre le prêt en cause.
		if anomalies, err := db.CheckInventoryAnomalies(); err == nil {
			loaned := make(map[int]bool, len(keyIDs))
			for _, id := range keyIDs {
				loaned[id] = true
			}
			for _, an := range anomalies {
				if loaned[an.KeyID] {
					warn := widget.NewLabel(fmt.Sprintf(
						"⚠ Erreur d'inventaire, vérifier le stock : clé n° %s (%d sortie(s) pour %d utilisable(s)).",
						an.KeyNumber, an.Loaned, an.Total-an.Reserve))
					warn.Wrapping = fyne.TextWrapWord
					confirmBox.Add(warn)
				}
			}
		}
		var p *widget.PopUp
		p = widget.NewModalPopUp(container.NewVBox(
			confirmBox,
			widget.NewButton("OK", func() {
				a.window.Canvas().Overlays().Remove(p)
				a.showDashboard()
			}),
		), a.window.Canvas())
		p.Show()
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
