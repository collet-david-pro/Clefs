package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// SimpleTheme est le thème visuel de l'application.
//
// Il dérive du thème clair par défaut de Fyne et ne surcharge que les éléments
// qui donnent à l'application son identité : palette de couleurs (bleu indigo
// moderne), arrondis plus marqués, espacements aérés et typographie légèrement
// agrandie. Les couleurs exposées via le UI kit (ui_kit.go) reprennent la même
// palette pour rester cohérentes dans tout le projet.
type SimpleTheme struct{}

// Color retourne la couleur associée à un nom de rôle de l'interface.
func (s SimpleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameHover:
		return colorPrimaryHover
	case theme.ColorNameFocus:
		return colorPrimary
	case theme.ColorNameSelection:
		return color.NRGBA{R: colorPrimary.R, G: colorPrimary.G, B: colorPrimary.B, A: 40}
	case theme.ColorNameButton:
		// Boutons "neutres" : gris très clair, pour que seuls les boutons
		// Importance=High (qui utilisent Primary) ressortent en couleur.
		return colorSurfaceAlt
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return colorText
	case theme.ColorNameInputBackground:
		return colorSurface
	case theme.ColorNamePlaceHolder:
		return colorTextMuted
	case theme.ColorNameSeparator:
		return colorBorder
	case theme.ColorNameError:
		return colorDanger
	case theme.ColorNameSuccess:
		return colorSuccess
	case theme.ColorNameWarning:
		return colorWarning
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font retourne la police (police par défaut de Fyne, très lisible).
func (s SimpleTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon retourne l'icône associée à un nom (jeu d'icônes par défaut de Fyne).
func (s SimpleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size retourne la dimension associée à un nom (espacements, rayons, polices).
// On agrandit légèrement les paddings et les rayons d'arrondi pour un rendu
// plus aéré et plus moderne que le thème par défaut.
func (s SimpleTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameInlineIcon:
		return 22
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameInputRadius:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ApplySimpleTheme applique le thème à l'application.
func ApplySimpleTheme(a fyne.App) {
	a.Settings().SetTheme(&SimpleTheme{})
}
