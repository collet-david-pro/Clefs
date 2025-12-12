package gui

import (
	"clefs/internal/db"
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// createConfigView crée la vue de configuration avec sauvegarde/import
func createConfigView(app *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Section Sauvegarde/Restauration
	backupSection := createBackupSection(app)

	// Section Navigation vers les autres configurations
	navSection := createConfigNavigationSection(app)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		backupSection,
		widget.NewSeparator(),
		navSection,
	)

	return container.NewVScroll(content)
}

// createBackupSection crée la section de sauvegarde/restauration
func createBackupSection(app *App) fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle("💾 Sauvegarde et Restauration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Informations
	infoLabel := widget.NewLabel("Sauvegardez régulièrement votre base de données pour éviter toute perte de données.")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Bouton Sauvegarder
	backupBtn := widget.NewButton("💾 Sauvegarder la Base de Données", func() {
		showBackupDialog(app)
	})
	backupBtn.Importance = widget.HighImportance

	// Bouton Restaurer
	restoreBtn := widget.NewButton("📥 Importer/Restaurer une Sauvegarde", func() {
		showRestoreDialog(app)
	})
	restoreBtn.Importance = widget.MediumImportance

	// Bouton Sauvegarde Automatique
	autoBackupBtn := widget.NewButton("⚡ Sauvegarde Rapide", func() {
		performQuickBackup(app)
	})

	// Bouton Gérer les Sauvegardes
	manageBackupsBtn := widget.NewButton("📋 Gérer les Sauvegardes", func() {
		app.showBackups()
	})
	manageBackupsBtn.Importance = widget.MediumImportance

	// Bouton Importer depuis Python
	importPythonBtn := widget.NewButton("📥 Importer depuis Version Python", func() {
		showImportPythonDialog(app)
	})
	importPythonBtn.Importance = widget.MediumImportance

	// Section Version Démo
	demoTitle := widget.NewLabelWithStyle("🎮 Mode Démonstration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	demoInfo := widget.NewLabel("Remplissez la base de données avec des données de test pour découvrir l'application.")
	demoInfo.Wrapping = fyne.TextWrapWord

	// Bouton Version Démo
	demoBtn := widget.NewButton("🎮 Charger la Version Démo", func() {
		showLoadDemoDialog(app)
	})
	demoBtn.Importance = widget.MediumImportance

	// Section Danger Zone
	dangerTitle := widget.NewLabelWithStyle("⚠️ ZONE DANGEREUSE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dangerWarning := widget.NewLabel("Les actions ci-dessous sont irréversibles et suppriment toutes les données !")
	dangerWarning.Wrapping = fyne.TextWrapWord

	// Bouton Réinitialiser
	resetBtn := widget.NewButton("🗑️ RÉINITIALISER LA BASE DE DONNÉES", func() {
		showResetDatabaseDialog(app)
	})
	resetBtn.Importance = widget.DangerImportance

	buttons := container.NewVBox(
		backupBtn,
		restoreBtn,
		autoBackupBtn,
		manageBackupsBtn,
		importPythonBtn,
		widget.NewSeparator(),
		demoTitle,
		demoInfo,
		demoBtn,
		widget.NewSeparator(),
		dangerTitle,
		dangerWarning,
		resetBtn,
	)

	return container.NewVBox(
		sectionTitle,
		infoLabel,
		widget.NewSeparator(),
		buttons,
	)
}

// createConfigNavigationSection crée la section de navigation vers les autres configs
func createConfigNavigationSection(app *App) fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle("⚙️ Gestion des Données", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Boutons de navigation
	buildingsBtn := widget.NewButton("🏢 Gérer les Bâtiments", func() {
		app.showBuildings()
	})

	roomsBtn := widget.NewButton("🚪 Gérer les Salles", func() {
		app.showRooms()
	})

	keysBtn := widget.NewButton("🔑 Gérer les Clés", func() {
		app.showKeys()
	})

	borrowersBtn := widget.NewButton("👥 Gérer les Emprunteurs", func() {
		app.showBorrowers()
	})

	buttons := container.NewVBox(
		buildingsBtn,
		roomsBtn,
		keysBtn,
		borrowersBtn,
	)

	return container.NewVBox(
		sectionTitle,
		widget.NewSeparator(),
		buttons,
	)
}

// showBackupDialog affiche la boîte de dialogue de sauvegarde
func showBackupDialog(app *App) {
	// Créer le répertoire de sauvegarde
	dbPath := app.dbPath
	if err := db.CreateBackupDirectory(dbPath); err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la création du répertoire de sauvegarde: %v", err))
		return
	}

	// Nom de fichier par défaut
	defaultFilename := fmt.Sprintf("clefs_backup_%s.db", time.Now().Format("20060102_150405"))

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if writer == nil {
			return // Annulé
		}
		defer writer.Close()

		backupPath := writer.URI().Path()

		// Effectuer la sauvegarde
		err = db.BackupDatabase(dbPath, backupPath)
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur lors de la sauvegarde: %v", err))
			return
		}

		app.showSuccess(fmt.Sprintf("Base de données sauvegardée avec succès!\n\nEmplacement: %s", backupPath))
	}, app.window)

	saveDialog.SetFileName(defaultFilename)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".db"}))
	saveDialog.Show()
}

// showRestoreDialog affiche la boîte de dialogue de restauration
func showRestoreDialog(app *App) {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if reader == nil {
			return // Annulé
		}
		defer reader.Close()

		backupPath := reader.URI().Path()

		// Confirmer la restauration
		app.showConfirm("Confirmer la Restauration",
			"⚠️ ATTENTION: Cette action va remplacer votre base de données actuelle.\n\n"+
				"Une sauvegarde de la base actuelle sera créée automatiquement.\n\n"+
				"Voulez-vous continuer?",
			func() {
				// Effectuer la restauration
				err := db.RestoreDatabase(backupPath, app.dbPath)
				if err != nil {
					app.showError("Erreur", fmt.Sprintf("Erreur lors de la restauration: %v", err))
					return
				}

				app.showSuccess("Base de données restaurée avec succès!\n\nL'application va se rafraîchir.")

				// Rafraîchir l'affichage
				app.showDashboard()
			})
	}, app.window)

	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".db"}))
	openDialog.Show()
}

// performQuickBackup effectue une sauvegarde rapide
func performQuickBackup(app *App) {
	dbPath := app.dbPath

	// Créer le répertoire de sauvegarde
	if err := db.CreateBackupDirectory(dbPath); err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la création du répertoire de sauvegarde: %v", err))
		return
	}

	// Chemin de sauvegarde par défaut
	backupPath := db.GetDefaultBackupPath(dbPath)

	// Effectuer la sauvegarde
	err := db.BackupDatabase(dbPath, backupPath)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la sauvegarde: %v", err))
		return
	}

	// Extraire juste le nom du fichier pour l'affichage
	filename := filepath.Base(backupPath)
	app.showSuccess(fmt.Sprintf("✅ Sauvegarde rapide effectuée!\n\nFichier: %s", filename))
}

// showResetDatabaseDialog affiche le dialogue de réinitialisation avec 3 confirmations
func showResetDatabaseDialog(app *App) {
	// PREMIÈRE CONFIRMATION
	app.showConfirm("⚠️ Réinitialisation - Étape 1/3",
		"🚨 ATTENTION : Vous êtes sur le point de SUPPRIMER TOUTES LES DONNÉES !\n\n"+
			"Cela inclut :\n"+
			"• Toutes les clés\n"+
			"• Tous les emprunteurs\n"+
			"• Tous les emprunts\n"+
			"• Tous les bâtiments et salles\n\n"+
			"Une sauvegarde automatique sera créée avant la suppression.\n\n"+
			"Êtes-vous ABSOLUMENT SÛR de vouloir continuer ?",
		func() {
			// DEUXIÈME CONFIRMATION
			app.showConfirm("⚠️ Réinitialisation - Étape 2/3",
				"🔴 VRAIMENT ?\n\n"+
					"Cette action est IRRÉVERSIBLE !\n\n"+
					"Toutes vos données actuelles seront DÉFINITIVEMENT PERDUES.\n"+
					"Seule la sauvegarde automatique pourra les récupérer.\n\n"+
					"Voulez-vous VRAIMENT continuer ?",
				func() {
					// TROISIÈME CONFIRMATION
					app.showConfirm("⚠️ Réinitialisation - Étape 3/3 - DERNIÈRE CHANCE",
						"🛑 CONFIRMATION DÉFINITIVE\n\n"+
							"C'est votre DERNIÈRE CHANCE de reculer !\n\n"+
							"En cliquant sur 'Confirmer', vous acceptez de :\n"+
							"• Supprimer TOUTES les données de l'application\n"+
							"• Repartir avec une base de données vierge\n"+
							"• Perdre définitivement toutes les informations actuelles\n\n"+
							"⚠️ CETTE ACTION EST DÉFINITIVE !\n\n"+
							"Confirmez-vous la réinitialisation complète ?",
						func() {
							// Effectuer la réinitialisation
							performDatabaseReset(app)
						})
				})
		})
}

// performDatabaseReset effectue la réinitialisation de la base de données
func performDatabaseReset(app *App) {
	dbPath := app.dbPath

	// Effectuer la réinitialisation
	err := db.ResetDatabase(dbPath)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la réinitialisation: %v", err))
		return
	}

	app.showSuccess("✅ Base de données réinitialisée avec succès !\n\n" +
		"Une sauvegarde de vos anciennes données a été créée dans le dossier 'backups/'.\n\n" +
		"L'application va maintenant se rafraîchir avec une base vierge.")

	// Rafraîchir l'affichage
	app.showDashboard()
}

// showLoadDemoDialog affiche le dialogue pour charger la version démo
func showLoadDemoDialog(app *App) {
	app.showConfirm("Charger la Version Démo",
		"🎮 Voulez-vous charger des données de démonstration ?\n\n"+
			"Cela va ajouter :\n"+
			"• 5 bâtiments\n"+
			"• 12 salles/points d'accès\n"+
			"• 10 clés avec associations\n"+
			"• 8 emprunteurs\n"+
			"• 6 emprunts actifs\n\n"+
			"⚠️ Note : Les données existantes seront conservées.\n"+
			"Si vous voulez repartir de zéro, utilisez d'abord la réinitialisation.",
		func() {
			performLoadDemo(app)
		})
}

// performLoadDemo charge les données de démonstration
func performLoadDemo(app *App) {
	// Charger les données de démo
	err := db.GenerateDemoData()
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors du chargement des données de démo: %v", err))
		return
	}

	app.showSuccess("✅ Données de démonstration chargées avec succès !\n\n" +
		"Vous pouvez maintenant explorer toutes les fonctionnalités de l'application.\n\n" +
		"L'application va se rafraîchir pour afficher les nouvelles données.")

	// Rafraîchir l'affichage
	app.showDashboard()
}

// showImportPythonDialog affiche le dialogue d'importation depuis Python
func showImportPythonDialog(app *App) {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			app.showError("Erreur", fmt.Sprintf("Erreur: %v", err))
			return
		}
		if reader == nil {
			return // Annulé
		}
		defer reader.Close()

		pythonDBPath := reader.URI().Path()

		// Confirmer l'importation
		app.showConfirm("Confirmer l'Importation",
			"📥 Importer les données depuis la version Python ?\n\n"+
				"Cette action va :\n"+
				"• Créer une sauvegarde automatique de votre base actuelle\n"+
				"• Importer toutes les données de l'ancienne base Python\n"+
				"• Fusionner les données (les doublons seront ignorés)\n\n"+
				"⚠️ Cette opération peut prendre quelques instants.\n\n"+
				"Voulez-vous continuer ?",
			func() {
				// Effectuer l'importation
				err := db.ImportFromPythonDB(pythonDBPath, app.dbPath)
				if err != nil {
					app.showError("Erreur", fmt.Sprintf("Erreur lors de l'importation: %v", err))
					return
				}

				app.showSuccess("✅ Importation réussie !\n\n" +
					"Les données de la version Python ont été importées avec succès.\n\n" +
					"Une sauvegarde de votre base actuelle a été créée automatiquement.\n\n" +
					"L'application va se rafraîchir pour afficher les données importées.")

				// Rafraîchir l'affichage
				app.showDashboard()
			})
	}, app.window)

	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".db"}))
	openDialog.Show()
}
