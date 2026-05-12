package gui

import (
	"clefs/internal/db"
	"clefs/internal/export"
	"clefs/internal/pdf"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	header := container.NewBorder(nil, nil, nil, container.NewHBox(csvBtn, addBtn), title)

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
