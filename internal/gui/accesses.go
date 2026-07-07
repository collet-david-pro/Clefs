package gui

import (
	"clefs/internal/db"
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// accessesCompactMode mémorise le mode d'affichage de la liste des accès
// (compact = ligne dépliable, détaillé = card complète), conservé entre deux
// rafraîchissements de la vue. Même principe que keysCompactMode.
var accessesCompactMode bool

// createAccessesView crée la vue de gestion des accès (ex-salles) avec filtres enrichis.
func createAccessesView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gérer les Accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("➕ Ajouter un Accès", func() {
		showAddAccessDialog(a)
	})
	addBtn.Importance = widget.HighImportance

	toggleBtn := widget.NewButton(accessesToggleLabel(), func() {
		accessesCompactMode = !accessesCompactMode
		a.showAccesses()
	})

	header := container.NewBorder(nil, nil, nil, container.NewHBox(toggleBtn, addBtn), title)

	buildings, _ := db.GetAllBuildings()
	allAccesses, _ := db.GetAllAccesses()

	// --- Filtres ---
	buildingOptions := []string{"Tous les bâtiments"}
	for _, b := range buildings {
		buildingOptions = append(buildingOptions, b.Name)
	}

	// Collecter étages et catégories uniques
	floorSet := map[string]struct{}{}
	categorySet := map[string]struct{}{}
	for _, r := range allAccesses {
		if r.Floor != "" {
			floorSet[r.Floor] = struct{}{}
		}
		if r.Category != "" {
			categorySet[r.Category] = struct{}{}
		}
	}
	floorOptions := []string{"Tous les étages"}
	for f := range floorSet {
		floorOptions = append(floorOptions, f)
	}
	sort.Strings(floorOptions[1:])

	categoryOptions := []string{"Toutes les catégories"}
	for c := range categorySet {
		categoryOptions = append(categoryOptions, c)
	}
	sort.Strings(categoryOptions[1:])

	buildingFilter := widget.NewSelect(buildingOptions, nil)
	buildingFilter.SetSelectedIndex(0)
	floorFilter := widget.NewSelect(floorOptions, nil)
	floorFilter.SetSelectedIndex(0)
	categoryFilter := widget.NewSelect(categoryOptions, nil)
	categoryFilter.SetSelectedIndex(0)

	listContainer := container.NewVBox()

	refresh := func() {
		listContainer.Objects = nil
		filtered := filterAccesses(allAccesses, buildings,
			buildingFilter.Selected, floorFilter.Selected, categoryFilter.Selected)

		if len(filtered) == 0 {
			listContainer.Add(widget.NewLabel("Aucun accès trouvé pour ces filtres."))
		} else if accessesCompactMode {
			listContainer.Add(createAccessesCompactView(filtered, buildings, a))
		} else {
			for _, r := range filtered {
				listContainer.Add(accessDetailBlock(r, buildings, a))
			}
		}
		listContainer.Refresh()
	}

	buildingFilter.OnChanged = func(_ string) { refresh() }
	floorFilter.OnChanged = func(_ string) { refresh() }
	categoryFilter.OnChanged = func(_ string) { refresh() }
	refresh()

	filters := container.NewGridWithColumns(3,
		buildingFilter, floorFilter, categoryFilter,
	)

	return container.NewBorder(
		container.NewVBox(header, filters, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(listContainer),
	)
}

// filterAccesses applique les filtres bâtiment/étage/catégorie à une liste
// d.accès. Une chaîne de filtre vide ou "Tous..." désactive le critère.
func filterAccesses(accesses []db.Room, buildings []db.Building, bFilter, fFilter, cFilter string) []db.Room {
	var result []db.Room
	for _, r := range accesses {
		if bFilter != "Tous les bâtiments" {
			bName := ""
			for _, b := range buildings {
				if b.ID == r.BuildingID {
					bName = b.Name
					break
				}
			}
			if bName != bFilter {
				continue
			}
		}
		if fFilter != "Tous les étages" && r.Floor != fFilter {
			continue
		}
		if cFilter != "Toutes les catégories" && r.Category != cFilter {
			continue
		}
		result = append(result, r)
	}
	return result
}

// accessesToggleLabel retourne le libellé du bouton de bascule compact/détaillé
// (il annonce le mode vers lequel le clic fera basculer).
func accessesToggleLabel() string {
	if accessesCompactMode {
		return "🗂 Vue détaillée"
	}
	return "📄 Vue compacte"
}

// buildingNameFor retourne le nom du bâtiment d'un accès, ou "" s'il est
// introuvable.
func buildingNameFor(r db.Room, buildings []db.Building) string {
	for _, b := range buildings {
		if b.ID == r.BuildingID {
			return b.Name
		}
	}
	return ""
}

// accessDetailBlock construit le bloc de détail complet d'un accès (card avec
// méta-données, nombre de clés et actions Modifier/Supprimer), réutilisé par la
// vue détaillée et par le contenu déplié de la vue compacte.
func accessDetailBlock(r db.Room, buildings []db.Building, a *App) fyne.CanvasObject {
	bName := buildingNameFor(r, buildings)
	keys, _ := db.GetKeysForAccess(r.ID)
	return accessCard(r, bName, len(keys),
		func() { showEditAccessDialog(a, r.ID) },
		func() {
			a.showConfirm("Supprimer", fmt.Sprintf("Supprimer l'accès %q ?", r.Name), func() {
				if err := db.DeleteRoom(r.ID); err != nil {
					a.showError("Erreur", err.Error())
					return
				}
				a.showAccesses()
			})
		},
	)
}

// createAccessesCompactView crée la liste compacte des accès : un accordéon dont
// chaque entrée affiche, repliée, le nom de l'accès et son nombre de clés, et
// dépliée, le détail complet (identique à la vue détaillée). Réutilise le même
// mécanisme d'expand/collapse que la vue clés.
func createAccessesCompactView(accesses []db.Room, buildings []db.Building, a *App) fyne.CanvasObject {
	acc := widget.NewAccordion()
	acc.MultiOpen = true
	for _, r := range accesses {
		r := r
		keys, _ := db.GetKeysForAccess(r.ID)
		keyCount := len(keys)
		summary := fmt.Sprintf("%s  —  %d clé(s)", r.Name, keyCount)
		if bName := buildingNameFor(r, buildings); bName != "" {
			summary = fmt.Sprintf("%s  ·  %s  —  %d clé(s)", r.Name, bName, keyCount)
		}
		acc.Append(widget.NewAccordionItem(summary, accessDetailBlock(r, buildings, a)))
	}
	return acc
}

// showAddAccessDialog affiche le formulaire d'ajout d'un accès
func showAddAccessDialog(a *App) {
	buildings, err := db.GetAllBuildings()
	if err != nil || len(buildings) == 0 {
		a.showError("Erreur", "Veuillez d'abord créer au moins un bâtiment.")
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("ex: Salle B12, Portail nord...")
	typeEntry := widget.NewEntry()
	typeEntry.SetPlaceHolder("ex: Salle de classe, Local technique...")
	floorEntry := widget.NewEntry()
	floorEntry.SetPlaceHolder("ex: RDC, R+1, Sous-sol...")
	categoryEntry := widget.NewEntry()
	categoryEntry.SetPlaceHolder("ex: salle de classe, bureau, accès extérieur...")
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Accès sensible, alarme, PMR...")
	notesEntry.SetMinRowsVisible(2)

	bNames := make([]string, len(buildings))
	for i, b := range buildings {
		bNames[i] = b.Name
	}
	buildingSelect := widget.NewSelect(bNames, nil)
	buildingSelect.SetSelectedIndex(0)

	form := widget.NewForm(
		widget.NewFormItem("Désignation *", nameEntry),
		widget.NewFormItem("Bâtiment *", buildingSelect),
		widget.NewFormItem("Type", typeEntry),
		widget.NewFormItem("Étage / Niveau", floorEntry),
		widget.NewFormItem("Catégorie", categoryEntry),
		widget.NewFormItem("Observations", notesEntry),
	)

	var popup *widget.PopUp
	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			a.showError("Erreur", "La désignation est obligatoire.")
			return
		}
		idx := buildingSelect.SelectedIndex()
		if idx < 0 {
			a.showError("Erreur", "Veuillez sélectionner un bâtiment.")
			return
		}
		room := &db.Room{
			Name:       nameEntry.Text,
			Type:       typeEntry.Text,
			BuildingID: buildings[idx].ID,
			Floor:      floorEntry.Text,
			Category:   categoryEntry.Text,
			Notes:      notesEntry.Text,
		}
		if err := db.CreateRoom(room); err != nil {
			a.showError("Erreur", err.Error())
			return
		}
		a.window.Canvas().Overlays().Remove(popup)
		a.showAccesses()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() { a.window.Canvas().Overlays().Remove(popup) })

	content := container.NewVBox(
		widget.NewLabelWithStyle("Nouvel Accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		container.NewHBox(cancelBtn, saveBtn),
	)
	popup = widget.NewModalPopUp(container.NewPadded(content), a.window.Canvas())
	popup.Show()
}

// showEditAccessDialog affiche le formulaire de modification d'un accès
func showEditAccessDialog(a *App, roomID int) {
	buildings, _ := db.GetAllBuildings()
	room, err := db.GetRoomByID(roomID)
	if err != nil {
		a.showError("Erreur", "Accès introuvable.")
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(room.Name)
	typeEntry := widget.NewEntry()
	typeEntry.SetText(room.Type)
	floorEntry := widget.NewEntry()
	floorEntry.SetText(room.Floor)
	categoryEntry := widget.NewEntry()
	categoryEntry.SetText(room.Category)
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(room.Notes)
	notesEntry.SetMinRowsVisible(2)

	bNames := make([]string, len(buildings))
	selectedIdx := 0
	for i, b := range buildings {
		bNames[i] = b.Name
		if b.ID == room.BuildingID {
			selectedIdx = i
		}
	}
	buildingSelect := widget.NewSelect(bNames, nil)
	buildingSelect.SetSelectedIndex(selectedIdx)

	form := widget.NewForm(
		widget.NewFormItem("Désignation *", nameEntry),
		widget.NewFormItem("Bâtiment *", buildingSelect),
		widget.NewFormItem("Type", typeEntry),
		widget.NewFormItem("Étage / Niveau", floorEntry),
		widget.NewFormItem("Catégorie", categoryEntry),
		widget.NewFormItem("Observations", notesEntry),
	)

	var popup *widget.PopUp
	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			a.showError("Erreur", "La désignation est obligatoire.")
			return
		}
		idx := buildingSelect.SelectedIndex()
		room.Name = nameEntry.Text
		room.Type = typeEntry.Text
		room.BuildingID = buildings[idx].ID
		room.Floor = floorEntry.Text
		room.Category = categoryEntry.Text
		room.Notes = notesEntry.Text
		if err := db.UpdateRoom(room); err != nil {
			a.showError("Erreur", err.Error())
			return
		}
		a.window.Canvas().Overlays().Remove(popup)
		a.showAccesses()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() { a.window.Canvas().Overlays().Remove(popup) })

	content := container.NewVBox(
		widget.NewLabelWithStyle("Modifier l'Accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		container.NewHBox(cancelBtn, saveBtn),
	)
	popup = widget.NewModalPopUp(container.NewPadded(content), a.window.Canvas())
	popup.Show()
}
