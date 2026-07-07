package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createAboutView construit la page "À propos" : version, fonctionnalités,
// stack technique, auteur et licence.
func createAboutView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gestionnaire de Clés", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	version := widget.NewLabel("Version " + AppVersion)
	version.Alignment = fyne.TextAlignCenter

	sep := widget.NewSeparator

	// Présentation
	desc := widget.NewLabel(
		"Application de bureau native pour la gestion du parc de clés\n" +
			"du Collège Victor Hugo — Chauny (02300).",
	)
	desc.Wrapping = fyne.TextWrapWord
	desc.Alignment = fyne.TextAlignCenter

	// Fonctionnalités
	featTitle := widget.NewLabelWithStyle("Fonctionnalités principales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	featList := container.NewVBox(
		widget.NewLabel("  • Inventaire des clés avec liaison portes/accès"),
		widget.NewLabel("  • Vues compactes dépliables pour les clés et les accès"),
		widget.NewLabel("  • Ajustement rapide du stock d'une clé depuis la liste"),
		widget.NewLabel("  • Prêt par besoin : sélection des portes → trousseau minimal calculé automatiquement"),
		widget.NewLabel("  • Bon de remise PDF généré à chaque emprunt"),
		widget.NewLabel("  • Détection des redondances d'accès"),
		widget.NewLabel("  • Import/export CSV des détenteurs (avec modèle téléchargeable)"),
		widget.NewLabel("  • Historique complet filtrable et exportable CSV"),
		widget.NewLabel("  • Sauvegardes atomiques, restauration en un clic"),
		widget.NewLabel("  • Utilisation en réseau local (multi-postes simultanés)"),
	)

	// Stack technique
	techTitle := widget.NewLabelWithStyle("Technologies", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	techList := container.NewVBox(
		widget.NewLabel("  • Langage : Go 1.21"),
		widget.NewLabel("  • Interface graphique : Fyne v2.4.5"),
		widget.NewLabel("  • Base de données : SQLite (mode WAL, pur Go sans CGO)"),
		widget.NewLabel("  • PDF : gofpdf"),
		widget.NewLabel("  • Exécutable unique, sans installation, sans connexion internet"),
	)

	// Auteur
	authorTitle := widget.NewLabelWithStyle("Développement", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	authorInfo := container.NewVBox(
		widget.NewLabel("  COLLET David"),
		widget.NewLabel("  Secrétaire Général"),
		widget.NewLabel("  Collège Victor Hugo — Chauny (02300)"),
		widget.NewLabel("  david.collet@ac-amiens.fr"),
	)

	// Licence
	licTitle := widget.NewLabelWithStyle("Licence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	licText := widget.NewLabel(
		"Distribué sous licence MIT.\n\n" +
			"Permission est accordée, gratuitement, à toute personne obtenant\n" +
			"une copie de ce logiciel, de l'utiliser, le copier, le modifier,\n" +
			"le fusionner, le publier, le distribuer et/ou le vendre, sous réserve\n" +
			"de conserver la présente notice dans toutes les copies.",
	)
	licText.Wrapping = fyne.TextWrapWord

	copyright := widget.NewLabel("© 2026 COLLET David — MIT License")
	copyright.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		title,
		version,
		sep(),
		container.NewPadded(desc),
		sep(),
		container.NewPadded(container.NewVBox(featTitle, featList)),
		sep(),
		container.NewPadded(container.NewVBox(techTitle, techList)),
		sep(),
		container.NewPadded(container.NewVBox(authorTitle, authorInfo)),
		sep(),
		container.NewPadded(container.NewVBox(licTitle, licText)),
		sep(),
		container.NewPadded(copyright),
	)

	return container.NewVScroll(container.NewPadded(content))
}
