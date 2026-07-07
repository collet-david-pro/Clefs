package gui

import (
	"bytes"
	"clefs/internal/db"
	"clefs/internal/export"
	"clefs/internal/pdf"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// createBorrowersView crée la vue de gestion des emprunteurs
func createBorrowersView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gérer les Emprunteurs", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("➕ Ajouter un Emprunteur", func() {
		showAddBorrowerDialog(app)
	})
	addBtn.Importance = widget.HighImportance

	csvBtn := widget.NewButton("📊 Exporter CSV", func() { exportBorrowersCSV(app) })

	templateBtn := widget.NewButton("📥 Modèle CSV", func() { downloadBorrowersTemplateCSV(app) })

	importBtn := widget.NewButton("📤 Importer CSV", func() { showImportBorrowersCSV(app) })

	header := container.NewBorder(nil, nil, nil, container.NewHBox(templateBtn, importBtn, csvBtn, addBtn), title)

	// Récupérer les emprunteurs
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return container.NewVBox(
			header,
			widget.NewLabel(fmt.Sprintf("Erreur: %v", err)),
		)
	}

	// Créer la liste des emprunteurs
	borrowersList := createBorrowersListView(borrowers, app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(borrowersList),
	)

	return content
}

// createBorrowersListView crée la liste des emprunteurs
func createBorrowersListView(borrowers []db.Borrower, app *App) fyne.CanvasObject {
	list := container.NewVBox()

	for _, borrower := range borrowers {
		b := borrower // Capture

		// Récupérer le nombre d'emprunts actifs
		loanCount, _ := db.GetBorrowerActiveLoanCount(b.ID)

		statusStr := b.Status
		if statusStr == "" {
			statusStr = "permanent"
		}
		borrowerInfo := container.NewVBox(
			widget.NewLabelWithStyle(b.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("Statut: %s  |  Email: %s  |  Tél: %s", statusStr, b.Email, b.Phone)),
			widget.NewLabel(fmt.Sprintf("Emprunts actifs: %d", loanCount)),
		)

		actions := container.NewHBox()

		if loanCount > 0 {
			receiptBtn := widget.NewButton("📄 Reçu", func() {
				generateBorrowerReceipt(app, b.ID)
			})
			actions.Add(receiptBtn)
		}

		editBtn := widget.NewButton("✏️ Modifier", func() {
			showEditBorrowerDialog(app, b.ID)
		})
		actions.Add(editBtn)

		deleteBtn := widget.NewButton("🗑️ Supprimer", func() {
			if loanCount > 0 {
				app.showError("Impossible de supprimer", "Cet emprunteur a des emprunts actifs.")
				return
			}
			app.showConfirm("Confirmer la suppression",
				fmt.Sprintf("Êtes-vous sûr de vouloir supprimer %s?", b.Name),
				func() {
					err := db.DeleteBorrower(b.ID)
					if err != nil {
						app.showError("Erreur", fmt.Sprintf("Erreur lors de la suppression: %v", err))
						return
					}
					app.showSuccess("Emprunteur supprimé avec succès!")
					app.showBorrowers()
				})
		})
		deleteBtn.Importance = widget.DangerImportance
		actions.Add(deleteBtn)

		borrowerCard := container.NewBorder(nil, nil, nil, actions, borrowerInfo)
		list.Add(borrowerCard)
		// Séparateur seulement entre les éléments, pas après le dernier
		if b.ID != borrowers[len(borrowers)-1].ID {
			list.Add(widget.NewSeparator())
		}
	}

	return list
}

// showAddBorrowerDialog affiche la boîte de dialogue pour ajouter un emprunteur
func showAddBorrowerDialog(app *App) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom Prénom")
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email (optionnel)")
	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Téléphone (optionnel)")
	statusSelect := widget.NewSelect(
		[]string{"permanent", "contractuel", "intervenant", "entreprise"},
		nil,
	)
	statusSelect.SetSelected("permanent")

	form := widget.NewForm(
		widget.NewFormItem("Nom *", nameEntry),
		widget.NewFormItem("Statut *", statusSelect),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Téléphone", phoneEntry),
	)

	var popupDialog *widget.PopUp
	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom est requis.")
			return
		}
		borrower := &db.Borrower{
			Name:   nameEntry.Text,
			Email:  emailEntry.Text,
			Phone:  phoneEntry.Text,
			Status: statusSelect.Selected,
		}
		if err := db.CreateBorrower(borrower); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la création: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Détenteur créé avec succès!")
		app.showBorrowers()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Ajouter un Détenteur", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)
	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(420, 300))
	popupDialog.Show()
}

// showEditBorrowerDialog affiche la boîte de dialogue pour modifier un emprunteur
func showEditBorrowerDialog(app *App, borrowerID int) {
	borrower, err := db.GetBorrowerByID(borrowerID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération: %v", err))
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(borrower.Name)
	emailEntry := widget.NewEntry()
	emailEntry.SetText(borrower.Email)
	phoneEntry := widget.NewEntry()
	phoneEntry.SetText(borrower.Phone)
	statusSelect := widget.NewSelect(
		[]string{"permanent", "contractuel", "intervenant", "entreprise"},
		nil,
	)
	status := borrower.Status
	if status == "" {
		status = "permanent"
	}
	statusSelect.SetSelected(status)

	form := widget.NewForm(
		widget.NewFormItem("Nom *", nameEntry),
		widget.NewFormItem("Statut *", statusSelect),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Téléphone", phoneEntry),
	)

	var popupDialog *widget.PopUp
	saveBtn := widget.NewButton("Enregistrer", func() {
		if nameEntry.Text == "" {
			app.showError("Erreur", "Le nom est requis.")
			return
		}
		borrower.Name = nameEntry.Text
		borrower.Email = emailEntry.Text
		borrower.Phone = phoneEntry.Text
		borrower.Status = statusSelect.Selected
		if err := db.UpdateBorrower(borrower); err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la modification: %v", err))
			return
		}
		app.window.Canvas().Overlays().Remove(popupDialog)
		app.showSuccess("Détenteur modifié avec succès!")
		app.showBorrowers()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Annuler", func() {
		app.window.Canvas().Overlays().Remove(popupDialog)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Modifier le Détenteur", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, saveBtn),
	)
	popupDialog = widget.NewModalPopUp(content, app.window.Canvas())
	popupDialog.Resize(fyne.NewSize(420, 300))
	popupDialog.Show()
}

// generateBorrowerReceipt génère un reçu PDF pour un emprunteur
func generateBorrowerReceipt(app *App, borrowerID int) {
	// Récupérer l'emprunteur
	borrower, err := db.GetBorrowerByID(borrowerID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération de l'emprunteur: %v", err))
		return
	}

	// Récupérer les emprunts actifs
	loans, err := db.GetActiveLoansByBorrowerID(borrowerID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération des emprunts: %v", err))
		return
	}

	if len(loans) == 0 {
		app.showError("Erreur", "Aucun emprunt actif pour cet emprunteur.")
		return
	}

	// Générer le PDF
	pdfData, err := pdf.GenerateBorrowerReceipt(borrower, loans)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Sauvegarder le fichier
	filename := fmt.Sprintf("bon_de_sortie_cles_%s_%s.pdf",
		borrower.Name,
		time.Now().Format("20060102"))

	// Demander où sauvegarder
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write(pdfData)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de l'écriture du fichier: %v", err))
			return
		}

		app.showSuccess("Reçu PDF généré avec succès!")
	}, app.window)

	saveDialog.SetFileName(filename)
	saveDialog.Show()
}

// borrowersCSVHeaders est l'ordre des colonnes attendu pour l'import/export et
// le modèle des détenteurs. Il correspond aux champs de db.Borrower saisissables
// dans showAddBorrowerDialog (le nom est obligatoire, les autres facultatifs).
var borrowersCSVHeaders = []string{"Nom", "Email", "Statut", "Téléphone"}

// downloadBorrowersTemplateCSV génère dans documents/ un modèle CSV vide (juste
// les en-têtes et une ligne d'exemple) que l'utilisateur peut remplir puis
// réimporter via showImportBorrowersCSV.
func downloadBorrowersTemplateCSV(app *App) {
	rows := [][]string{
		{"Dupont Jean", "jean.dupont@exemple.fr", "permanent", "0102030405"},
	}
	filePath, err := export.SaveCSV(export.Filename("modele_detenteurs"), borrowersCSVHeaders, rows)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du modèle: %v", err))
		return
	}
	app.showSuccess(fmt.Sprintf("✅ Modèle CSV enregistré : %s", filePath))
}

// showImportBorrowersCSV ouvre un sélecteur de fichier puis crée en masse les
// détenteurs décrits dans le CSV choisi (au format du modèle). Les lignes
// invalides sont ignorées et résumées à la fin sans faire échouer tout l'import.
func showImportBorrowersCSV(app *App) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		importBorrowersFromCSV(app, reader)
	}, app.window)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	fileDialog.Show()
}

// importBorrowersFromCSV lit le CSV fourni et crée les détenteurs ligne par
// ligne. Il tolère le BOM UTF-8 et le séparateur point-virgule produits par
// l'export, saute une éventuelle ligne d'en-tête, et affiche un résumé
// détaillant les lignes créées et les lignes rejetées.
func importBorrowersFromCSV(app *App, r io.Reader) {
	data, err := io.ReadAll(r)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Lecture du fichier impossible: %v", err))
		return
	}
	// Retirer un éventuel BOM UTF-8 en tête (présent dans nos exports).
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	cr := csv.NewReader(bytes.NewReader(data))
	cr.Comma = ';'
	cr.FieldsPerRecord = -1 // tolérer un nombre de colonnes variable
	records, err := cr.ReadAll()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("CSV invalide: %v", err))
		return
	}
	if len(records) == 0 {
		app.showError("Erreur", "Le fichier est vide.")
		return
	}

	validStatuses := map[string]bool{
		"permanent": true, "contractuel": true, "intervenant": true, "entreprise": true,
	}

	created := 0
	var problems []string
	for i, rec := range records {
		lineNum := i + 1
		// Sauter une ligne d'en-tête (première ligne dont la 1re cellule vaut "Nom").
		if i == 0 && len(rec) > 0 && strings.EqualFold(strings.TrimSpace(rec[0]), "Nom") {
			continue
		}
		// Ignorer les lignes totalement vides.
		if isBlankRecord(rec) {
			continue
		}
		name := ""
		if len(rec) > 0 {
			name = strings.TrimSpace(rec[0])
		}
		if name == "" {
			problems = append(problems, fmt.Sprintf("ligne %d (nom manquant)", lineNum))
			continue
		}
		email, status, phone := "", "permanent", ""
		if len(rec) > 1 {
			email = strings.TrimSpace(rec[1])
		}
		if len(rec) > 2 && strings.TrimSpace(rec[2]) != "" {
			status = strings.ToLower(strings.TrimSpace(rec[2]))
			if !validStatuses[status] {
				problems = append(problems, fmt.Sprintf("ligne %d (statut « %s » inconnu)", lineNum, status))
				continue
			}
		}
		if len(rec) > 3 {
			phone = strings.TrimSpace(rec[3])
		}
		b := &db.Borrower{Name: name, Email: email, Status: status, Phone: phone}
		if err := db.CreateBorrower(b); err != nil {
			problems = append(problems, fmt.Sprintf("ligne %d (%v)", lineNum, err))
			continue
		}
		created++
	}

	summary := fmt.Sprintf("%d détenteur(s) créé(s).", created)
	if len(problems) > 0 {
		summary += fmt.Sprintf("\n%d ligne(s) invalide(s) :\n  • %s",
			len(problems), strings.Join(problems, "\n  • "))
	}
	if created > 0 {
		app.showInfo("Import terminé", summary)
		app.showBorrowers()
		return
	}
	app.showInfo("Import terminé", summary)
}

// isBlankRecord indique si un enregistrement CSV ne contient que des cellules
// vides (utile pour ignorer les lignes vides d'un fichier).
func isBlankRecord(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// exportBorrowersCSV exporte la liste des détenteurs en CSV dans documents/.
func exportBorrowersCSV(app *App) {
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la récupération: %v", err))
		return
	}
	headers := []string{"Nom", "Email", "Statut", "Téléphone"}
	rows := make([][]string, len(borrowers))
	for i, b := range borrowers {
		rows[i] = []string{b.Name, b.Email, b.Status, b.Phone}
	}
	filePath, err := export.SaveCSV(export.Filename("detenteurs"), headers, rows)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'export: %v", err))
		return
	}
	app.showSuccess(fmt.Sprintf("✅ Export CSV enregistré : %s", filePath))
}
