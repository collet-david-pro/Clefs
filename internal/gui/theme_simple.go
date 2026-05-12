package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// SimpleTheme est un thème simple et lisible basé sur le thème par défaut
type SimpleTheme struct{}

// Color retourne les couleurs du thème
func (s SimpleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Utiliser les couleurs par défaut de Fyne pour la plupart des éléments
	// et personnaliser seulement quelques couleurs clés
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 123, B: 255, A: 255} // Bleu moderne
	case theme.ColorNameButton:
		return color.NRGBA{R: 0, G: 123, B: 255, A: 255} // Bleu pour les boutons
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc
		}
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0, G: 0, B: 0, A: 255} // Noir
		}
		return theme.DefaultTheme().Color(name, variant)
	default:
		// Pour tout le reste, utiliser le thème par défaut
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font retourne la police du thème
func (s SimpleTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon retourne l'icône du thème
func (s SimpleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size retourne les tailles du thème
func (s SimpleTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ApplySimpleTheme applique le thème simple à l'application
func ApplySimpleTheme(a fyne.App) {
	a.Settings().SetTheme(&SimpleTheme{})
}
