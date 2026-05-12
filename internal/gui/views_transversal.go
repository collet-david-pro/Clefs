package gui

import (
	"clefs/internal/business"
	"clefs/internal/db"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createWhoHasWhatView — "Qui a quoi ?" : détenteurs avec leurs clés et accès cumulés.
func createWhoHasWhatView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Qui a quoi ?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Détenteurs avec leurs clés actuelles et accès couverts.")

	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return errorView(title, err)
	}

	list := container.NewVBox()

	for _, b := range borrowers {
		b := b
		loans, err := db.GetActiveLoansByBorrowerID(b.ID)
		if err != nil || len(loans) == 0 {
			continue
		}

		bwk, err := business.BuildBorrowerWithKeys(b, loans)
		if err != nil {
			continue
		}

		// Construire le résumé
		keyNames := make([]string, len(loans))
		for i, l := range loans {
			keyNames[i] = l.KeyNumber
		}

		accessNames := make([]string, len(bwk.CoveredAccesses))
		for i, r := range bwk.CoveredAccesses {
			accessNames[i] = r.Name
		}

		redondanceStr := ""
		if len(bwk.Redundancies) > 0 {
			redNames := make([]string, len(bwk.Redundancies))
			for i, r := range bwk.Redundancies {
				redNames[i] = r.Name
			}
			redondanceStr = "  [REDONDANCE : " + strings.Join(redNames, ", ") + "]"
		}

		headerText := fmt.Sprintf("%s — %d clé(s) : %s%s",
			b.Name, len(loans), strings.Join(keyNames, ", "), redondanceStr)

		accessLabel := widget.NewLabel("Accès couverts : " + strings.Join(accessNames, ", "))
		accessLabel.Wrapping = fyne.TextWrapWord

		detailBox := container.NewVBox(accessLabel)
		if len(bwk.Redundancies) > 0 {
			redNames := make([]string, len(bwk.Redundancies))
			for i, r := range bwk.Redundancies {
				redNames[i] = r.Name
			}
			detailBox.Add(widget.NewLabelWithStyle(
				"Redondances : "+strings.Join(redNames, ", "),
				fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
			))
		}

		list.Add(widget.NewAccordion(widget.NewAccordionItem(headerText, detailBox)))
	}

	if len(list.Objects) == 0 {
		list.Add(widget.NewLabel("Aucun détenteur n'a de clé en ce moment."))
	}

	return container.NewBorder(
		container.NewVBox(title, subtitle, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(list),
	)
}

// createKeyToAccessView — "Quelle clé ouvre quelle porte ?" : sélectionner un accès → clés associées.
func createKeyToAccessView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Quelle clé pour quelle porte ?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	accesses, _ := db.GetAllAccesses()
	buildings, _ := db.GetAllBuildings()

	options := make([]string, len(accesses))
	for i, r := range accesses {
		bName := ""
		for _, b := range buildings {
			if b.ID == r.BuildingID {
				bName = b.Name
				break
			}
		}
		options[i] = fmt.Sprintf("%s [%s]", r.Name, bName)
	}

	resultBox := container.NewVBox()
	var accessSelect *widget.Select

	onAccessChange := func(selected string) {
		resultBox.Objects = nil
		idx := 0
		for i, o := range options {
			if o == selected {
				idx = i
				break
			}
		}
		if idx >= len(accesses) {
			resultBox.Refresh()
			return
		}
		r := accesses[idx]
		keys, _ := db.GetKeysForAccess(r.ID)
		if len(keys) == 0 {
			resultBox.Add(widget.NewLabel("Aucune clé n'est associée à cet accès."))
		} else {
			for _, k := range keys {
				count, _ := db.GetActiveLoanCount(k.ID)
				available := k.QuantityTotal - k.QuantityReserve - count
				status := "✅ Disponible"
				if available <= 0 {
					status = "🔴 Indisponible"
				}
				resultBox.Add(widget.NewLabel(fmt.Sprintf(
					"🔑 %s — %s  |  Stock: %d  |  Sorties: %d  |  %s",
					k.Number, k.Description, k.QuantityTotal, count, status,
				)))
				resultBox.Add(widget.NewSeparator())
			}
		}
		resultBox.Refresh()
	}

	accessSelect = widget.NewSelect(options, onAccessChange)

	return container.NewBorder(
		container.NewVBox(title, accessSelect, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(resultBox),
	)
}

// createKeysByBuildingView — "Quelles clés dans ce bâtiment ?"
func createKeysByBuildingView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Clés par bâtiment", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	buildings, _ := db.GetAllBuildings()
	options := make([]string, len(buildings))
	for i, b := range buildings {
		options[i] = b.Name
	}

	resultBox := container.NewVBox()
	var buildingSelect *widget.Select

	onBuildingChange := func(selected string) {
		resultBox.Objects = nil
		idx := buildingSelect.SelectedIndex()
		if idx < 0 || idx >= len(buildings) {
			resultBox.Refresh()
			return
		}
		bld := buildings[idx]
		accesses, _ := db.GetAccessesByBuilding(bld.ID)

		keySet := map[int]db.Key{}
		for _, r := range accesses {
			keys, _ := db.GetKeysForAccess(r.ID)
			for _, k := range keys {
				keySet[k.ID] = k
			}
		}

		if len(keySet) == 0 {
			resultBox.Add(widget.NewLabel("Aucune clé associée à ce bâtiment."))
		} else {
			resultBox.Add(widget.NewLabel(fmt.Sprintf("%d clé(s) pour %s :", len(keySet), bld.Name)))
			resultBox.Add(widget.NewSeparator())
			for _, k := range keySet {
				count, _ := db.GetActiveLoanCount(k.ID)
				available := k.QuantityTotal - k.QuantityReserve - count
				resultBox.Add(widget.NewLabel(fmt.Sprintf(
					"🔑 %s — %s  |  Dispo: %d/%d",
					k.Number, k.Description, available, k.QuantityTotal,
				)))
			}
		}
		resultBox.Refresh()
	}

	buildingSelect = widget.NewSelect(options, onBuildingChange)

	return container.NewBorder(
		container.NewVBox(title, buildingSelect, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(resultBox),
	)
}

// createAvailableKeysView — "Clés disponibles" avec cards et filtre bâtiment.
func createAvailableKeysView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Clés disponibles", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	keys, _ := db.GetKeysWithAvailability()
	buildings, _ := db.GetAllBuildings()

	bOptions := []string{"Tous les bâtiments"}
	for _, b := range buildings {
		bOptions = append(bOptions, b.Name)
	}
	bFilter := widget.NewSelect(bOptions, nil)
	bFilter.SetSelectedIndex(0)

	listBox := container.NewVBox()

	rebuild := func() {
		listBox.Objects = nil
		shown := 0
		for _, k := range keys {
			k := k
			if bFilter.Selected != "Tous les bâtiments" {
				rooms, _ := db.GetRoomsForKey(k.ID)
				found := false
				for _, r := range rooms {
					for _, b := range buildings {
						if b.ID == r.BuildingID && b.Name == bFilter.Selected {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					continue
				}
			}
			card := keyCard(k,
				func() { a.showNewLoan() },
				func() { showReturnDialog(a, k.ID) },
			)
			listBox.Add(card)
			shown++
		}
		if shown == 0 {
			listBox.Add(widget.NewLabel("Aucune clé trouvée."))
		}
		listBox.Refresh()
	}

	bFilter.OnChanged = func(_ string) { rebuild() }
	rebuild()

	return container.NewBorder(
		container.NewVBox(title,
			widget.NewLabel(fmt.Sprintf("Dernière mise à jour : %s", time.Now().Format("15:04:05"))),
			bFilter,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewVScroll(listBox),
	)
}

// errorView retourne une vue d'erreur simple
func errorView(title fyne.CanvasObject, err error) fyne.CanvasObject {
	return container.NewVBox(title, widget.NewLabel(fmt.Sprintf("Erreur : %v", err)))
}
