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

	version := widget.NewLabel("Version 2.1")
	version.Alignment = fyne.TextAlignCenter

	// Description
	description := widget.NewLabel("Application de gestion des clés et des emprunts avec génération de reçus PDF.")
	description.Wrapping = fyne.TextWrapWord
	description.Alignment = fyne.TextAlignCenter

	// Fonctionnalités principales
	featuresTitle := widget.NewLabelWithStyle("✨ Fonctionnalités", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	featuresList := container.NewVBox(
		widget.NewLabel("• Gestion des clés et emprunteurs"),
		widget.NewLabel("• Tableau de bord en temps réel"),
		widget.NewLabel("• Génération de reçus PDF"),
		widget.NewLabel("• Sauvegardes automatiques"),
		widget.NewLabel("• Compatible Windows"),
	)

	// Nouveautés Version 2.1
	newTitle := widget.NewLabelWithStyle("🆕 Version 2.1", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	newList := container.NewVBox(
		widget.NewLabel("• Affichage optimisé des emprunteurs multiples"),
		widget.NewLabel("• Amélioration de l'interface du tableau de bord"),
	)

	// Contact
	contactTitle := widget.NewLabelWithStyle("📧 Contact", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	contactInfo := widget.NewLabel("david.collet@ac-amiens.fr")
	contactInfo.Alignment = fyne.TextAlignCenter

	// Copyright
	copyright := widget.NewLabel("© 2025")
	copyright.Alignment = fyne.TextAlignCenter

	// Assembler le contenu avec scroll
	content := container.NewVBox(
		title,
		version,
		description,
		widget.NewSeparator(),
		featuresTitle,
		featuresList,
		widget.NewSeparator(),
		newTitle,
		newList,
		widget.NewSeparator(),
		contactTitle,
		contactInfo,
		widget.NewSeparator(),
		copyright,
	)

	// Retourner avec scroll pour gérer le contenu long
	return container.NewVScroll(
		container.NewPadded(content),
	)
}
