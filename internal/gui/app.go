package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// App représente l'application principale
type App struct {
	fyneApp     fyne.App
	window      fyne.Window
	content     *fyne.Container
	dbPath      string
	currentView string // vue active pour refreshCurrentView
}

// NewApp crée une nouvelle instance de l'application
func NewApp(dbPath string) *App {
	a := app.New()
	ApplySimpleTheme(a)

	w := a.NewWindow("🔑 Gestionnaire de Clés")
	w.Resize(fyne.NewSize(1400, 900))
	w.CenterOnScreen()

	return &App{
		fyneApp: a,
		window:  w,
		dbPath:  dbPath,
	}
}

// Run démarre l'application directement sur le dashboard
func (a *App) Run() {
	a.showDashboard()
	a.window.ShowAndRun()
}

// createMenu crée la barre de navigation latérale
func (a *App) createMenu() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gestionnaire de Clés", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	sep := widget.NewSeparator

	// Section Prêts
	sectionPrets := widget.NewLabelWithStyle("Prêts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dashboardBtn := widget.NewButton("Tableau de bord", func() { a.showDashboard() })
	dashboardBtn.Importance = widget.HighImportance
	newLoanBtn := widget.NewButton("Nouvel emprunt", func() { a.showNewLoan() })
	newLoanBtn.Importance = widget.HighImportance
	activeLoansBtn := widget.NewButton("Emprunts en cours", func() { a.showActiveLoans() })
	historyBtn := widget.NewButton("Historique", func() { a.showHistory() })

	// Section Référentiels
	sectionRef := widget.NewLabelWithStyle("Référentiels", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	keysBtn := widget.NewButton("Clés", func() { a.showKeys() })
	borrowersBtn := widget.NewButton("Détenteurs", func() { a.showBorrowers() })
	accessesBtn := widget.NewButton("Accès", func() { a.showAccesses() })
	buildingsBtn := widget.NewButton("Bâtiments", func() { a.showBuildings() })

	// Section Consultation
	sectionConsult := widget.NewLabelWithStyle("Consultation", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	whoBtn := widget.NewButton("Qui a quoi ?", func() { a.showWhoHasWhat() })
	buildingViewBtn := widget.NewButton("Par bâtiment", func() { a.showKeysByBuilding() })
	redundBtn := widget.NewButton("Redondances", func() { a.showRedundancies() })
	keyPlanBtn := widget.NewButton("Plan de clés", func() { a.showKeyPlan() })

	// Section Application
	sectionApp := widget.NewLabelWithStyle("Application", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	configBtn := widget.NewButton("Configuration", func() { a.showConfig() })
	helpBtn := widget.NewButton("Aide", func() { a.showHelp() })
	aboutBtn := widget.NewButton("A propos", func() { a.showAbout() })
	quitBtn := widget.NewButton("Quitter", func() { a.quit() })
	quitBtn.Importance = widget.DangerImportance

	menuBox := container.NewVBox(
		container.NewPadded(title),
		sep(),
		container.NewPadded(container.NewVBox(
			sectionPrets,
			dashboardBtn,
			newLoanBtn,
			activeLoansBtn,
			historyBtn,
		)),
		sep(),
		container.NewPadded(container.NewVBox(
			sectionRef,
			keysBtn,
			borrowersBtn,
			accessesBtn,
			buildingsBtn,
		)),
		sep(),
		container.NewPadded(container.NewVBox(
			sectionConsult,
			whoBtn,
			buildingViewBtn,
			redundBtn,
			keyPlanBtn,
		)),
		sep(),
		container.NewPadded(container.NewVBox(
			sectionApp,
			configBtn,
			helpBtn,
			aboutBtn,
			quitBtn,
		)),
	)

	return container.NewVScroll(menuBox)
}

// setContent met à jour le contenu principal et enregistre la vue courante
func (a *App) setContent(content fyne.CanvasObject, viewName string) {
	a.currentView = viewName
	a.content = container.NewMax(content)
	menu := a.createMenu()
	a.window.SetContent(container.NewBorder(nil, nil, menu, nil, a.content))
}

// refreshCurrentView recharge la vue active
func (a *App) refreshCurrentView() {
	switch a.currentView {
	case "dashboard":
		a.showDashboard()
	case "activeLoans":
		a.showActiveLoans()
	case "history":
		a.showHistory()
	case "loansReport":
		a.showLoansReport()
	case "keys":
		a.showKeys()
	case "borrowers":
		a.showBorrowers()
	case "buildings":
		a.showBuildings()
	case "accesses":
		a.showAccesses()
	case "keyPlan":
		a.showKeyPlan()
	case "whoHasWhat":
		a.showWhoHasWhat()
	case "keyToAccess":
		a.showKeyToAccess()
	case "keysByBuilding":
		a.showKeysByBuilding()
	case "availableKeys":
		a.showAvailableKeys()
	case "redundancies":
		a.showRedundancies()
	case "newLoan":
		a.showNewLoan()
	default:
		a.showDashboard()
	}
}

// --- Méthodes de navigation ---

func (a *App) showDashboard() {
	a.setContent(createModernDashboard(a), "dashboard")
}

func (a *App) showKeys() {
	a.setContent(createKeysView(a), "keys")
}

func (a *App) showBorrowers() {
	a.setContent(createBorrowersView(a), "borrowers")
}

func (a *App) showBuildings() {
	a.setContent(createBuildingsView(a), "buildings")
}

func (a *App) showNewLoan() {
	a.setContent(createNewLoanView(a), "newLoan")
}

func (a *App) showAccesses() {
	a.setContent(createAccessesView(a), "accesses")
}

// showRooms redirige vers showAccesses (compatibilité interne)
func (a *App) showRooms() {
	a.showAccesses()
}

func (a *App) showActiveLoans() {
	a.setContent(createActiveLoansView(a), "activeLoans")
}

func (a *App) showLoansReport() {
	a.setContent(createLoansReportView(a), "loansReport")
}

func (a *App) showHistory() {
	a.setContent(createHistoryView(a), "history")
}

func (a *App) showKeyPlan() {
	a.setContent(createKeyPlanView(a), "keyPlan")
}

func (a *App) showConfig() {
	a.setContent(createConfigView(a), "config")
}

func (a *App) showBackups() {
	a.setContent(createBackupsView(a), "backups")
}

func (a *App) showAbout() {
	a.setContent(createAboutView(), "about")
}

func (a *App) showHelp() {
	a.setContent(createHelpView(), "help")
}

func (a *App) showWhoHasWhat() {
	a.setContent(createWhoHasWhatView(a), "whoHasWhat")
}

func (a *App) showKeyToAccess() {
	a.setContent(createKeyToAccessView(a), "keyToAccess")
}

func (a *App) showKeysByBuilding() {
	a.setContent(createKeysByBuildingView(a), "keysByBuilding")
}

func (a *App) showAvailableKeys() {
	a.setContent(createAvailableKeysView(a), "availableKeys")
}

func (a *App) showRedundancies() {
	a.setContent(createRedundancyView(a), "redundancies")
}

// --- Dialogues communs ---

func (a *App) showError(title, message string) {
	var p *widget.PopUp
	p = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message),
		widget.NewButton("OK", func() { a.window.Canvas().Overlays().Remove(p) }),
	), a.window.Canvas())
	p.Show()
}

func (a *App) showSuccess(message string) {
	var p *widget.PopUp
	p = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabel(message),
		widget.NewButton("OK", func() { a.window.Canvas().Overlays().Remove(p) }),
	), a.window.Canvas())
	p.Show()
}

func (a *App) showConfirm(title, message string, onConfirm func()) {
	var p *widget.PopUp
	p = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message),
		container.NewHBox(
			widget.NewButton("Annuler", func() { a.window.Canvas().Overlays().Remove(p) }),
			widget.NewButton("Confirmer", func() {
				a.window.Canvas().Overlays().Remove(p)
				onConfirm()
			}),
		),
	), a.window.Canvas())
	p.Show()
}

func (a *App) quit() {
	a.showConfirm("Quitter", "Êtes-vous sûr de vouloir quitter ?", func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Erreur fermeture DB: %v", err)
		}
		a.fyneApp.Quit()
	})
}

// Initialize initialise la DB et crée l'App
func Initialize(dbPath string) (*App, error) {
	if err := db.InitDB(dbPath); err != nil {
		return nil, fmt.Errorf("erreur lors de l'initialisation de la base de données: %w", err)
	}
	return NewApp(dbPath), nil
}
