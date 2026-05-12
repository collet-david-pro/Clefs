package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createBackupsView crée la vue de gestion des sauvegardes
func createBackupsView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("💾 Gestion des Sauvegardes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Informations
	infoLabel := widget.NewLabel("Gérez vos sauvegardes : visualisez, restaurez ou supprimez les sauvegardes existantes.")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Bouton pour créer une nouvelle sauvegarde
	newBackupBtn := widget.NewButton("➕ Créer une Nouvelle Sauvegarde", func() {
		performQuickBackup(app)
		// Rafraîchir la vue
		app.showBackups()
	})
	newBackupBtn.Importance = widget.HighImportance

	header := container.NewVBox(
		title,
		infoLabel,
		widget.NewSeparator(),
		newBackupBtn,
		widget.NewSeparator(),
	)

	// Liste des sauvegardes
	backupsList := createBackupsList(app)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		backupsList,
	)

	return content
}

// createBackupsList crée la liste des sauvegardes
func createBackupsList(app *App) fyne.CanvasObject {
	// Récupérer les sauvegardes
	backups, err := db.ListBackups(app.dbPath)
	if err != nil {
		log.Printf("Erreur lors de la récupération des sauvegardes: %v", err)
		return widget.NewLabel("❌ Erreur lors du chargement des sauvegardes")
	}

	if len(backups) == 0 {
		emptyMsg := widget.NewLabel("📭 Aucune sauvegarde disponible")
		emptyMsg.Alignment = fyne.TextAlignCenter
		emptyInfo := widget.NewLabel("Créez votre première sauvegarde en cliquant sur le bouton ci-dessus.")
		emptyInfo.Alignment = fyne.TextAlignCenter
		emptyInfo.Wrapping = fyne.TextWrapWord
		return container.NewVBox(
			widget.NewSeparator(),
			emptyMsg,
			emptyInfo,
		)
	}

	// Créer le tableau des sauvegardes
	backupsContainer := container.NewVBox()

	// En-tête du tableau
	headerRow := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("📅 Date", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("🕐 Heure", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("📦 Taille", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("📝 Nom du Fichier", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("⚙️ Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	backupsContainer.Add(headerRow)
	backupsContainer.Add(widget.NewSeparator())

	// Lignes de données
	for _, backup := range backups {
		row := createBackupRow(backup, app)
		backupsContainer.Add(row)
		backupsContainer.Add(widget.NewSeparator())
	}

	return container.NewVScroll(backupsContainer)
}

// createBackupRow crée une ligne pour une sauvegarde
func createBackupRow(backup db.BackupInfo, app *App) fyne.CanvasObject {
	// Date
	dateLabel := widget.NewLabel(backup.ModTime.Format("02/01/2006"))

	// Heure
	timeLabel := widget.NewLabel(backup.ModTime.Format("15:04:05"))

	// Taille
	sizeLabel := widget.NewLabel(backup.SizeStr)

	// Nom du fichier
	nameLabel := widget.NewLabel(backup.Name)
	nameLabel.Wrapping = fyne.TextWrapOff

	// Actions
	restoreBtn := widget.NewButton("📥 Restaurer", func() {
		showRestoreConfirmDialog(app, backup)
	})
	restoreBtn.Importance = widget.MediumImportance

	deleteBtn := widget.NewButton("🗑️ Supprimer", func() {
		showDeleteBackupDialog(app, backup)
	})
	deleteBtn.Importance = widget.DangerImportance

	actions := container.NewHBox(restoreBtn, deleteBtn)

	row := container.NewGridWithColumns(5,
		dateLabel,
		timeLabel,
		sizeLabel,
		nameLabel,
		actions,
	)

	return row
}

// showRestoreConfirmDialog affiche la confirmation de restauration
func showRestoreConfirmDialog(app *App, backup db.BackupInfo) {
	message := fmt.Sprintf(
		"⚠️ ATTENTION : Cette action va remplacer votre base de données actuelle.\n\n"+
			"Sauvegarde à restaurer :\n"+
			"• Nom : %s\n"+
			"• Date : %s\n"+
			"• Taille : %s\n\n"+
			"Une sauvegarde de la base actuelle sera créée automatiquement avant la restauration.\n\n"+
			"Voulez-vous continuer ?",
		backup.Name,
		backup.ModTime.Format("02/01/2006 15:04:05"),
		backup.SizeStr,
	)

	app.showConfirm("Confirmer la Restauration", message, func() {
		// Effectuer la restauration
		err := db.RestoreDatabase(backup.Path, app.dbPath)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la restauration: %v", err))
			return
		}

		app.showSuccess(fmt.Sprintf(
			"✅ Base de données restaurée avec succès !\n\n"+
				"Sauvegarde restaurée : %s\n\n"+
				"L'application va se rafraîchir.",
			backup.Name,
		))

		// Rafraîchir l'affichage
		app.showDashboard()
	})
}

// showDeleteBackupDialog affiche la confirmation de suppression
func showDeleteBackupDialog(app *App, backup db.BackupInfo) {
	message := fmt.Sprintf(
		"🗑️ Êtes-vous sûr de vouloir supprimer cette sauvegarde ?\n\n"+
			"• Nom : %s\n"+
			"• Date : %s\n"+
			"• Taille : %s\n\n"+
			"⚠️ Cette action est irréversible !",
		backup.Name,
		backup.ModTime.Format("02/01/2006 15:04:05"),
		backup.SizeStr,
	)

	app.showConfirm("Confirmer la Suppression", message, func() {
		// Supprimer la sauvegarde
		err := db.DeleteBackup(backup.Path)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la suppression: %v", err))
			return
		}

		app.showSuccess(fmt.Sprintf(
			"✅ Sauvegarde supprimée avec succès !\n\n"+
				"Fichier supprimé : %s",
			backup.Name,
		))

		// Rafraîchir la vue
		app.showBackups()
	})
}
