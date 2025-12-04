package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createAboutView crée la vue À propos améliorée
func createAboutView() fyne.CanvasObject {
	// En-tête avec icône et titre
	title := widget.NewLabelWithStyle("🔑 Gestionnaire de Clés", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle.Bold = true

	version := widget.NewLabel("Version 2.0.0 - Go Edition")
	version.Alignment = fyne.TextAlignCenter

	// Description
	descTitle := widget.NewLabelWithStyle("📋 Description", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("Application professionnelle de gestion des clés et des emprunts. " +
		"Suivez en temps réel la disponibilité de vos clés, gérez les emprunts, " +
		"et générez automatiquement des reçus PDF.")
	description.Wrapping = fyne.TextWrapWord

	// Fonctionnalités principales
	featuresTitle := widget.NewLabelWithStyle("✨ Fonctionnalités Principales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	featuresList := container.NewVBox(
		widget.NewLabel("• 🔑 Gestion complète des clés (quantités, réserves, stockage)"),
		widget.NewLabel("• 👥 Gestion des emprunteurs avec coordonnées"),
		widget.NewLabel("• 📊 Tableau de bord avec disponibilité en temps réel"),
		widget.NewLabel("• 🏢 Organisation par bâtiments et points d'accès"),
		widget.NewLabel("• 📝 Emprunts simples ou multiples"),
		widget.NewLabel("• 📄 Génération automatique de reçus PDF"),
		widget.NewLabel("• 🗺️ Plan de clés détaillé"),
		widget.NewLabel("• 📈 Rapports et statistiques d'emprunts"),
		widget.NewLabel("• 💾 Gestion complète des sauvegardes"),
	)

	// Nouveautés Version 2.0
	newTitle := widget.NewLabelWithStyle("🆕 Nouveautés Version 2.0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	newList := container.NewVBox(
		widget.NewLabel("• 🚀 Releases automatiques multi-plateformes (Windows, macOS Intel & Apple Silicon)"),
		widget.NewLabel("• 💾 Interface dédiée de gestion des sauvegardes"),
		widget.NewLabel("• 📊 Tableau du dashboard avec colonnes alignées"),
		widget.NewLabel("• 📖 Mode d'emploi intégré dans l'application"),
		widget.NewLabel("• ⚡ Performance et stabilité améliorées"),
	)

	// Technologies
	techTitle := widget.NewLabelWithStyle("🛠️ Technologies Utilisées", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	techList := container.NewVBox(
		widget.NewLabel("• Go (Golang) - Langage de programmation"),
		widget.NewLabel("• Fyne v2 - Interface graphique native cross-platform"),
		widget.NewLabel("• SQLite (Pure Go) - Base de données embarquée"),
		widget.NewLabel("• gofpdf - Génération de documents PDF avec UTF-8"),
	)

	// Avantages
	advantagesTitle := widget.NewLabelWithStyle("🎯 Avantages", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	advantagesList := container.NewVBox(
		widget.NewLabel("✅ Application native (pas de navigateur requis)"),
		widget.NewLabel("✅ Performance optimale"),
		widget.NewLabel("✅ Un seul fichier exécutable, aucune installation"),
		widget.NewLabel("✅ Compatible Windows, macOS et Linux"),
		widget.NewLabel("✅ Base de données locale sécurisée"),
		widget.NewLabel("✅ Support complet des caractères accentués"),
	)

	// Licence et informations
	licenseTitle := widget.NewLabelWithStyle("📜 Licence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	licenseInfo := widget.NewLabel("Cette application est distribuée sous licence MIT.")
	licenseInfo.Wrapping = fyne.TextWrapWord

	copyright := widget.NewLabel("© 2025 - Application développée en Go")
	copyright.Alignment = fyne.TextAlignCenter

	madeWith := widget.NewLabel("Fait avec ❤️ et Go")
	madeWith.Alignment = fyne.TextAlignCenter

	// Assembler le contenu avec scroll
	content := container.NewVBox(
		title,
		version,
		widget.NewSeparator(),
		descTitle,
		description,
		widget.NewSeparator(),
		newTitle,
		newList,
		widget.NewSeparator(),
		featuresTitle,
		featuresList,
		widget.NewSeparator(),
		techTitle,
		techList,
		widget.NewSeparator(),
		advantagesTitle,
		advantagesList,
		widget.NewSeparator(),
		licenseTitle,
		licenseInfo,
		widget.NewSeparator(),
		copyright,
		madeWith,
	)

	// Retourner avec scroll pour gérer le contenu long
	return container.NewVScroll(
		container.NewPadded(content),
	)
}
