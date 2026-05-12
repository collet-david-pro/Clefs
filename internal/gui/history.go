package gui

import (
	"clefs/internal/db"
	"clefs/internal/export"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createHistoryView crée la vue de l'historique complet des prêts avec filtres.
func createHistoryView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Historique des Prêts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	keys, _ := db.GetAllKeys()
	borrowers, _ := db.GetAllBorrowers()

	// --- Filtres ---
	keyOptions := []string{"Toutes les clés"}
	for _, k := range keys {
		keyOptions = append(keyOptions, fmt.Sprintf("%s — %s", k.Number, k.Description))
	}

	borrowerOptions := []string{"Tous les détenteurs"}
	for _, b := range borrowers {
		borrowerOptions = append(borrowerOptions, b.Name)
	}

	statusOptions := []string{"Tous", "En cours", "Retournés", "En retard"}

	keySelect := widget.NewSelect(keyOptions, nil)
	keySelect.SetSelectedIndex(0)
	borrowerSelect := widget.NewSelect(borrowerOptions, nil)
	borrowerSelect.SetSelectedIndex(0)
	statusSelect := widget.NewSelect(statusOptions, nil)
	statusSelect.SetSelectedIndex(0)

	// Date début / fin (texte DD/MM/YYYY)
	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("Du (JJ/MM/AAAA)")
	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("Au (JJ/MM/AAAA)")

	listContainer := container.NewVBox()
	countLabel := widget.NewLabel("")

	applyFilters := func() {
		filters := db.LoanFilters{}

		if keySelect.SelectedIndex() > 0 {
			k := keys[keySelect.SelectedIndex()-1]
			filters.KeyID = &k.ID
		}
		if borrowerSelect.SelectedIndex() > 0 {
			b := borrowers[borrowerSelect.SelectedIndex()-1]
			filters.BorrowerID = &b.ID
		}
		switch statusSelect.Selected {
		case "En cours":
			filters.Status = "active"
		case "Retournés":
			filters.Status = "returned"
		case "En retard":
			filters.Status = "overdue"
		}
		if fromEntry.Text != "" {
			t, err := time.Parse("02/01/2006", fromEntry.Text)
			if err == nil {
				filters.DateFrom = &t
			}
		}
		if toEntry.Text != "" {
			t, err := time.Parse("02/01/2006", toEntry.Text)
			if err == nil {
				filters.DateTo = &t
			}
		}

		loans, err := db.GetLoanHistory(filters)
		listContainer.Objects = nil

		if err != nil {
			listContainer.Add(widget.NewLabel(fmt.Sprintf("Erreur : %v", err)))
			listContainer.Refresh()
			return
		}

		countLabel.SetText(fmt.Sprintf("%d prêt(s) trouvé(s)", len(loans)))

		if len(loans) == 0 {
			listContainer.Add(widget.NewLabel("Aucun prêt ne correspond aux filtres."))
		} else {
			for _, l := range loans {
				l := l
				status, statusColor := loanStatus(l)

				dateStr := l.LoanDate.Format("02/01/2006")
				returnStr := "—"
				if l.ReturnDate != nil {
					returnStr = l.ReturnDate.Format("02/01/2006")
				}
				plannedStr := "—"
				if l.PlannedReturnDate != nil {
					plannedStr = l.PlannedReturnDate.Format("02/01/2006")
				}
				duration := fmt.Sprintf("%d j", int(time.Since(l.LoanDate).Hours()/24))
				if l.ReturnDate != nil {
					duration = fmt.Sprintf("%d j", int(l.ReturnDate.Sub(l.LoanDate).Hours()/24))
				}
				agent := l.CreatedBy
				if agent == "" {
					agent = "—"
				}

				info := fmt.Sprintf("🔑 %-8s  👤 %-20s  📅 %-10s  🔚 %-10s  ⏰ %-10s  ⌛ %-5s  👷 %s",
					l.KeyNumber, l.BorrowerName, dateStr, returnStr, plannedStr, duration, agent)

				statusLabel := widget.NewLabelWithStyle(status, fyne.TextAlignLeading, fyne.TextStyle{Bold: statusColor == "red"})

				row := container.NewVBox(
					container.NewHBox(statusLabel, widget.NewLabel(info)),
					widget.NewSeparator(),
				)
				listContainer.Add(row)
			}
		}
		listContainer.Refresh()
	}

	// currentFilters garde les filtres actifs pour l'export CSV
	var currentFilters db.LoanFilters

	searchBtn := widget.NewButton("🔍 Rechercher", func() { applyFilters() })
	searchBtn.Importance = widget.HighImportance
	resetBtn := widget.NewButton("↺ Réinitialiser", func() {
		keySelect.SetSelectedIndex(0)
		borrowerSelect.SetSelectedIndex(0)
		statusSelect.SetSelectedIndex(0)
		fromEntry.SetText("")
		toEntry.SetText("")
		applyFilters()
	})
	csvBtn := widget.NewButton("📊 Exporter CSV", func() { exportHistoryCSV(a, currentFilters) })

	// Mettre à jour currentFilters à chaque recherche
	origApply := applyFilters
	applyFilters = func() {
		origApply()
		currentFilters = db.LoanFilters{}
		if keySelect.SelectedIndex() > 0 {
			k := keys[keySelect.SelectedIndex()-1]
			currentFilters.KeyID = &k.ID
		}
		if borrowerSelect.SelectedIndex() > 0 {
			b := borrowers[borrowerSelect.SelectedIndex()-1]
			currentFilters.BorrowerID = &b.ID
		}
		switch statusSelect.Selected {
		case "En cours":
			currentFilters.Status = "active"
		case "Retournés":
			currentFilters.Status = "returned"
		case "En retard":
			currentFilters.Status = "overdue"
		}
	}

	applyFilters()

	filters := container.NewVBox(
		container.NewGridWithColumns(3, keySelect, borrowerSelect, statusSelect),
		container.NewGridWithColumns(2, fromEntry, toEntry),
		container.NewHBox(searchBtn, resetBtn, csvBtn, countLabel),
		widget.NewSeparator(),
	)

	return container.NewBorder(
		container.NewVBox(title, filters),
		nil, nil, nil,
		container.NewVScroll(listContainer),
	)
}

func exportHistoryCSV(a *App, filters db.LoanFilters) {
	loans, err := db.GetLoanHistory(filters)
	if err != nil {
		a.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération: %v", err))
		return
	}
	headers := []string{"Clé", "Description", "Détenteur", "Date remise", "Retour prévu", "Retour réel", "Statut", "Agent"}
	rows := make([][]string, len(loans))
	for i, l := range loans {
		status, _ := loanStatus(l)
		planned, returned := "—", "—"
		if l.PlannedReturnDate != nil {
			planned = l.PlannedReturnDate.Format("02/01/2006")
		}
		if l.ReturnDate != nil {
			returned = l.ReturnDate.Format("02/01/2006")
		}
		rows[i] = []string{l.KeyNumber, l.KeyDescription, l.BorrowerName, l.LoanDate.Format("02/01/2006"), planned, returned, status, l.CreatedBy}
	}
	filePath, err := export.SaveCSV(export.Filename("historique_prets"), headers, rows)
	if err != nil {
		a.showError("Erreur", fmt.Sprintf("Erreur lors de l'export: %v", err))
		return
	}
	a.showSuccess(fmt.Sprintf("✅ Export CSV enregistré : %s", filePath))
}

// loanStatus retourne le libellé de statut et une couleur indicative ("red", "orange", "")
func loanStatus(l db.LoanWithDetails) (string, string) {
	if l.ReturnDate != nil {
		return "✅ Retourné", ""
	}
	if l.PlannedReturnDate != nil && time.Now().After(*l.PlannedReturnDate) {
		return "🔴 En retard", "red"
	}
	return "📤 En cours", ""
}
