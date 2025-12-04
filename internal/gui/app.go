package gui

import (
	"clefs/internal/db"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// App représente l'application principale
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	content *fyne.Container
	dbPath  string
}

// NewApp crée une nouvelle instance de l'application
func NewApp(dbPath string) *App {
	a := app.New()
	w := a.NewWindow("Gestionnaire de Clés")
	w.Resize(fyne.NewSize(1200, 800))

	return &App{
		fyneApp: a,
		window:  w,
		dbPath:  dbPath,
	}
}

// Run démarre l'application
func (a *App) Run() {
	// Créer le menu de navigation
	menu := a.createMenu()

	// Afficher le tableau de bord par défaut
	a.showDashboard()

	// Layout principal avec menu à gauche et contenu à droite
	mainContent := container.NewBorder(nil, nil, menu, nil, a.content)

	a.window.SetContent(mainContent)
	a.window.ShowAndRun()
}

// createMenu crée le menu de navigation
func (a *App) createMenu() fyne.CanvasObject {
	dashboardBtn := widget.NewButton("📊 Tableau de Bord", func() {
		a.showDashboard()
	})
	dashboardBtn.Importance = widget.HighImportance

	activeLoansBtn := widget.NewButton("📋 Emprunts en Cours", func() {
		a.showActiveLoans()
	})

	reportsBtn := widget.NewButton("📄 Rapport des Clés", func() {
		a.showLoansReport()
	})

	keyPlanBtn := widget.NewButton("🗺️ Plan de Clés", func() {
		a.showKeyPlan()
	})

	separator1 := widget.NewSeparator()

	configBtn := widget.NewButton("⚙️ Configuration", func() {
		a.showConfig()
	})

	separator2 := widget.NewSeparator()

	helpBtn := widget.NewButton("📖 Mode d'Emploi", func() {
		a.showHelp()
	})

	aboutBtn := widget.NewButton("ℹ️ À Propos", func() {
		a.showAbout()
	})

	separator3 := widget.NewSeparator()

	// Bouton Quitter
	quitBtn := widget.NewButton("🚪 Quitter", func() {
		a.quit()
	})
	quitBtn.Importance = widget.WarningImportance

	menuBox := container.NewVBox(
		dashboardBtn,
		activeLoansBtn,
		reportsBtn,
		keyPlanBtn,
		separator1,
		configBtn,
		separator2,
		helpBtn,
		aboutBtn,
		separator3,
		quitBtn,
	)

	return container.NewVScroll(menuBox)
}

// setContent met à jour le contenu principal
func (a *App) setContent(content fyne.CanvasObject) {
	a.content = container.NewMax(content)

	// Recréer le layout principal
	menu := a.createMenu()
	mainContent := container.NewBorder(nil, nil, menu, nil, a.content)
	a.window.SetContent(mainContent)
}

// showDashboard affiche le tableau de bord
func (a *App) showDashboard() {
	content := createDashboard(a)
	a.setContent(content)
}

// showKeys affiche la gestion des clés
func (a *App) showKeys() {
	content := createKeysView(a)
	a.setContent(content)
}

// showBorrowers affiche la gestion des emprunteurs
func (a *App) showBorrowers() {
	content := createBorrowersView(a)
	a.setContent(content)
}

// showBuildings affiche la gestion des bâtiments
func (a *App) showBuildings() {
	content := createBuildingsView(a)
	a.setContent(content)
}

// showRooms affiche la gestion des salles
func (a *App) showRooms() {
	content := createRoomsView(a)
	a.setContent(content)
}

// showActiveLoans affiche les emprunts actifs
func (a *App) showActiveLoans() {
	content := createActiveLoansView(a)
	a.setContent(content)
}

// showLoansReport affiche le rapport des emprunts
func (a *App) showLoansReport() {
	content := createLoansReportView(a)
	a.setContent(content)
}

// showKeyPlan affiche le plan de clés
func (a *App) showKeyPlan() {
	content := createKeyPlanView(a)
	a.setContent(content)
}

// showConfig affiche la page de configuration
func (a *App) showConfig() {
	content := createConfigView(a)
	a.setContent(content)
}

// showBackups affiche la gestion des sauvegardes
func (a *App) showBackups() {
	content := createBackupsView(a)
	a.setContent(content)
}

// showAbout affiche la page À propos
func (a *App) showAbout() {
	content := createAboutView()
	a.setContent(content)
}

// showHelp affiche le mode d'emploi
func (a *App) showHelp() {
	content := createHelpView()
	a.setContent(content)
}

// quit ferme l'application proprement
func (a *App) quit() {
	a.showConfirm("Quitter l'Application",
		"Êtes-vous sûr de vouloir quitter ?",
		func() {
			// Fermer la base de données proprement
			if err := db.CloseDB(); err != nil {
				log.Printf("Erreur lors de la fermeture de la base de données: %v", err)
			}

			// Quitter l'application
			a.fyneApp.Quit()
		})
}

// showError affiche un message d'erreur
func (a *App) showError(title, message string) {
	var errorPopup *widget.PopUp

	errorPopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel(message),
			widget.NewButton("OK", func() {
				a.window.Canvas().Overlays().Remove(errorPopup)
			}),
		),
		a.window.Canvas(),
	)
	errorPopup.Show()
}

// showSuccess affiche un message de succès
func (a *App) showSuccess(message string) {
	var successPopup *widget.PopUp

	successPopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel(message),
			widget.NewButton("OK", func() {
				a.window.Canvas().Overlays().Remove(successPopup)
			}),
		),
		a.window.Canvas(),
	)
	successPopup.Show()
}

// showConfirm affiche une boîte de dialogue de confirmation
func (a *App) showConfirm(title, message string, onConfirm func()) {
	var confirmPopup *widget.PopUp

	confirmPopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel(message),
			container.NewHBox(
				widget.NewButton("Annuler", func() {
					a.window.Canvas().Overlays().Remove(confirmPopup)
				}),
				widget.NewButton("Confirmer", func() {
					a.window.Canvas().Overlays().Remove(confirmPopup)
					onConfirm()
				}),
			),
		),
		a.window.Canvas(),
	)
	confirmPopup.Show()
}

// refreshCurrentView rafraîchit la vue actuelle
func (a *App) refreshCurrentView() {
	// Cette méthode sera appelée après des modifications pour rafraîchir l'affichage
	// Pour l'instant, on recharge simplement le tableau de bord
	a.showDashboard()
}

// Initialize initialise l'application et la base de données
func Initialize(dbPath string) (*App, error) {
	// Initialiser la base de données
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données: %v", err)
		return nil, err
	}

	// Créer l'application
	app := NewApp(dbPath)
	return app, nil
}
