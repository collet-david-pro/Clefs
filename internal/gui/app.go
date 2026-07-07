// Package gui contient toute l'interface graphique, construite avec Fyne.
//
// Organisation : le type App (app.go) détient la fenêtre et le menu latéral
// permanent ; chaque "vue" est une fonction createXxxView(...) fyne.CanvasObject
// définie dans un fichier dédié (keys.go, borrowers.go, new_loan_view.go, ...).
// La navigation se fait en remplaçant le panneau central via App.setContent.
//
// Conventions Fyne récurrentes dans ce package :
//   - container.NewBorder(top, bottom, left, right, center) : le centre prend
//     l'espace restant. C'est le seul moyen fiable de contraindre la hauteur
//     d'un VScroll (un VScroll dans un VBox se réduit à zéro).
//   - widget capturé par sa propre closure : on déclare `var w *widget.X`
//     AVANT de l'instancier pour pouvoir y faire référence dans son callback.
//
// L'application n'a pas d'authentification : l'utilisateur est toujours
// l'administrateur (cf. README / NOTICE).
package gui

import (
	"clefs/internal/db"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// App est l'état global de l'interface : application et fenêtre Fyne,
// conteneur central courant, et chemin de la base (passé aux dialogues
// d'import/sauvegarde).
type App struct {
	fyneApp     fyne.App        // instance Fyne (boucle d'événements, thème)
	window      fyne.Window     // fenêtre principale unique
	content     *fyne.Container // panneau central remplacé à chaque navigation
	dbPath      string          // chemin du fichier clefs.db (pour import/backup)
	currentView string          // nom de la vue active (diagnostic uniquement)
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

// navButton fabrique une entrée de menu : icône + libellé, alignée à gauche.
// Quand viewID correspond à la vue active (a.currentView), l'entrée est mise en
// surbrillance (importance haute = fond coloré) pour indiquer où l'on se trouve.
func (a *App) navButton(label string, icon fyne.Resource, viewID string, action func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, action)
	btn.Alignment = widget.ButtonAlignLeading
	if a.currentView == viewID {
		btn.Importance = widget.HighImportance
	} else {
		btn.Importance = widget.LowImportance
	}
	return btn
}

// menuSection retourne un titre de section discret (gris, petites majuscules).
func menuSection(title string) fyne.CanvasObject {
	t := canvas.NewText(title, colorTextMuted)
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = 11
	return container.NewPadded(t)
}

// createMenu crée la barre de navigation latérale.
//
// Structure : un en-tête (logo + nom), puis quatre sections (Prêts,
// Référentiels, Consultation, Application), chacune faite d'entrées navButton
// avec icône. L'entrée correspondant à la vue courante est surlignée.
func (a *App) createMenu() fyne.CanvasObject {
	// En-tête : pastille colorée façon logo + nom de l'application.
	logo := canvas.NewRectangle(colorPrimary)
	logo.CornerRadius = 8
	logo.SetMinSize(fyne.NewSize(34, 34))
	logoIcon := widget.NewIcon(theme.NewThemedResource(theme.HomeIcon()))
	appName := canvas.NewText("Gestionnaire de Clés", colorText)
	appName.TextStyle = fyne.TextStyle{Bold: true}
	appName.TextSize = 15
	header := container.NewHBox(
		container.NewStack(logo, container.NewCenter(logoIcon)),
		container.NewVBox(layout.NewSpacer(), appName, layout.NewSpacer()),
	)

	menuBox := container.NewVBox(
		container.NewPadded(header),

		menuSection("PRÊTS"),
		a.navButton("Tableau de bord", theme.HomeIcon(), "dashboard", func() { a.showDashboard() }),
		a.navButton("Nouvel emprunt", theme.ContentAddIcon(), "newLoan", func() { a.showNewLoan() }),
		a.navButton("Emprunts en cours", theme.MailSendIcon(), "activeLoans", func() { a.showActiveLoans() }),
		a.navButton("Historique", theme.HistoryIcon(), "history", func() { a.showHistory() }),

		menuSection("RÉFÉRENTIELS"),
		a.navButton("Clés", theme.LoginIcon(), "keys", func() { a.showKeys() }),
		a.navButton("Détenteurs", theme.AccountIcon(), "borrowers", func() { a.showBorrowers() }),
		a.navButton("Accès", theme.FolderOpenIcon(), "accesses", func() { a.showAccesses() }),
		a.navButton("Bâtiments", theme.StorageIcon(), "buildings", func() { a.showBuildings() }),

		menuSection("CONSULTATION"),
		a.navButton("Qui a quoi ?", theme.SearchIcon(), "whoHasWhat", func() { a.showWhoHasWhat() }),
		a.navButton("Par bâtiment", theme.GridIcon(), "keysByBuilding", func() { a.showKeysByBuilding() }),
		a.navButton("Redondances", theme.WarningIcon(), "redundancies", func() { a.showRedundancies() }),
		a.navButton("Plan de clés", theme.DocumentIcon(), "keyPlan", func() { a.showKeyPlan() }),

		menuSection("APPLICATION"),
		a.navButton("Configuration", theme.SettingsIcon(), "config", func() { a.showConfig() }),
		a.navButton("Aide", theme.HelpIcon(), "help", func() { a.showHelp() }),
		a.navButton("À propos", theme.InfoIcon(), "about", func() { a.showAbout() }),
	)

	// Bouton Quitter épinglé en bas de la barre.
	quitBtn := widget.NewButtonWithIcon("Quitter", theme.LogoutIcon(), func() { a.quit() })
	quitBtn.Alignment = widget.ButtonAlignLeading
	quitBtn.Importance = widget.LowImportance

	// Fond légèrement distinct pour la barre latérale.
	sidebarBg := canvas.NewRectangle(colorSurface)

	inner := container.NewBorder(
		nil,
		container.NewPadded(quitBtn),
		nil, nil,
		container.NewVScroll(menuBox),
	)
	return container.NewStack(sidebarBg, inner)
}

// setContent met à jour le contenu principal et reconstruit le menu latéral.
// viewName identifie la vue active (utile pour le diagnostic ; non lu par le runtime).
func (a *App) setContent(content fyne.CanvasObject, viewName string) {
	a.currentView = viewName
	a.content = container.NewMax(content)
	menu := fixedWidth(a.createMenu(), 240)
	// Le contenu central est posé sur le fond gris clair de l'application.
	pageBg := canvas.NewRectangle(colorBackground)
	page := container.NewStack(pageBg, container.NewPadded(a.content))
	a.window.SetContent(container.NewBorder(nil, nil, menu, nil, page))
}

// --- Méthodes de navigation ---
//
// Chaque showXxx construit la vue correspondante via sa fonction createXxxView
// et l'installe au centre avec setContent. Elles sont déclenchées par les
// boutons du menu latéral (createMenu) ou par des actions dans les vues
// (ex. un bouton "Emprunter" appelle showNewLoan). Le second argument de
// setContent est le nom de la vue, conservé dans App.currentView.

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

func (a *App) showKeysByBuilding() {
	a.setContent(createKeysByBuildingView(a), "keysByBuilding")
}

func (a *App) showRedundancies() {
	a.setContent(createRedundancyView(a), "redundancies")
}

// --- Dialogues communs ---
//
// Ces trois helpers affichent une popup modale réutilisable. Le motif Fyne
// `var p *widget.PopUp; p = ...` est nécessaire pour que le bouton OK puisse
// se référer à la popup (la fermer) depuis sa propre closure.

// showError affiche un message d'erreur bloquant avec un titre en gras.
func (a *App) showError(title, message string) {
	var p *widget.PopUp
	p = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message),
		widget.NewButton("OK", func() { a.window.Canvas().Overlays().Remove(p) }),
	), a.window.Canvas())
	p.Show()
}

// showSuccess affiche un message de confirmation simple.
func (a *App) showSuccess(message string) {
	var p *widget.PopUp
	p = widget.NewModalPopUp(container.NewVBox(
		widget.NewLabel(message),
		widget.NewButton("OK", func() { a.window.Canvas().Overlays().Remove(p) }),
	), a.window.Canvas())
	p.Show()
}

// showConfirm demande une confirmation ; onConfirm n'est exécuté que si
// l'utilisateur clique sur "Confirmer". Utilisé pour les suppressions et le quit.
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

// quit demande confirmation puis ferme proprement la base avant de quitter.
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
