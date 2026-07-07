package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// ui_kit.go centralise l'identité visuelle de l'application : la palette de
// couleurs et un petit ensemble de composants réutilisables (cartes, tuiles de
// statistique, badges, en-têtes de page). Tout le reste de la GUI s'appuie sur
// ces helpers pour rester cohérent et donner un rendu moderne et homogène.
//
// Fyne 2.4 ne propose pas de "carte avec fond coloré" prête à l'emploi : on les
// fabrique en superposant un canvas.Rectangle (le fond arrondi) et le contenu
// via container.NewStack.

// --- Palette ---
//
// Inspirée des palettes "indigo/slate" modernes. Les teintes sont volontairement
// douces (fonds très clairs, texte gris foncé plutôt que noir pur) pour un rendu
// reposant et professionnel.
var (
	colorPrimary      = color.NRGBA{R: 79, G: 70, B: 229, A: 255}   // indigo 600
	colorPrimaryHover = color.NRGBA{R: 67, G: 56, B: 202, A: 255}   // indigo 700
	colorPrimarySoft  = color.NRGBA{R: 238, G: 242, B: 255, A: 255} // indigo 50

	colorBackground = color.NRGBA{R: 247, G: 248, B: 251, A: 255} // gris très clair
	colorSurface    = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // blanc (cartes)
	colorSurfaceAlt = color.NRGBA{R: 241, G: 243, B: 247, A: 255} // gris clair (boutons neutres)
	colorBorder     = color.NRGBA{R: 226, G: 230, B: 236, A: 255} // bordures discrètes

	colorText      = color.NRGBA{R: 30, G: 35, B: 45, A: 255}    // gris ardoise foncé
	colorTextMuted = color.NRGBA{R: 120, G: 128, B: 140, A: 255} // gris moyen

	colorSuccess     = color.NRGBA{R: 22, G: 163, B: 74, A: 255}   // vert
	colorSuccessSoft = color.NRGBA{R: 220, G: 252, B: 231, A: 255} // vert pâle
	colorWarning     = color.NRGBA{R: 217, G: 119, B: 6, A: 255}   // orange
	colorWarningSoft = color.NRGBA{R: 254, G: 243, B: 199, A: 255} // orange pâle
	colorDanger      = color.NRGBA{R: 220, G: 38, B: 38, A: 255}   // rouge
	colorDangerSoft  = color.NRGBA{R: 254, G: 226, B: 226, A: 255} // rouge pâle
	colorInfoSoft    = color.NRGBA{R: 224, G: 242, B: 254, A: 255} // bleu pâle
)

// --- Composants réutilisables ---

// card enveloppe un contenu dans une carte blanche à coins arrondis avec une
// fine bordure, posée sur le fond gris de la fenêtre. C'est la brique de base
// de la mise en page moderne (remplace les VBox + Separator plats).
func card(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorSurface)
	bg.CornerRadius = 10
	bg.StrokeColor = colorBorder
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.NewPadded(content))
}

// coloredCard se comporte comme card mais avec un fond et une bordure teintés
// (utilisé pour les bandeaux d'alerte : retard, redondance...).
func coloredCard(content fyne.CanvasObject, fill, stroke color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(fill)
	bg.CornerRadius = 10
	bg.StrokeColor = stroke
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.NewPadded(content))
}

// statTile fabrique une tuile de statistique : une grande valeur, un libellé,
// et une pastille de couleur. Utilisée en rangée sur le tableau de bord.
func statTile(value, label string, accent color.Color) fyne.CanvasObject {
	dot := canvas.NewRectangle(accent)
	dot.CornerRadius = 4
	dot.SetMinSize(fyne.NewSize(8, 8))

	valueText := canvas.NewText(value, colorText)
	valueText.TextStyle = fyne.TextStyle{Bold: true}
	valueText.TextSize = 30

	labelText := canvas.NewText(label, colorTextMuted)
	labelText.TextSize = 13

	inner := container.NewVBox(
		container.NewHBox(container.NewCenter(dot), valueText),
		labelText,
	)
	return card(inner)
}

// badge retourne une petite étiquette colorée arrondie (texte sur fond pâle).
// Sert à matérialiser un statut (suggestion, manuel, en retard...).
func badge(text string, fill, fg color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(fill)
	bg.CornerRadius = 6
	label := canvas.NewText(text, fg)
	label.TextSize = 12
	label.TextStyle = fyne.TextStyle{Bold: true}
	padded := container.NewPadded(label)
	// Réduire le padding vertical en l'enveloppant simplement (Fyne gère le min).
	return container.NewStack(bg, padded)
}

// pageHeader retourne un en-tête de page : titre en gros, sous-titre discret,
// et zone d'actions à droite (boutons). actions peut être nil.
func pageHeader(title, subtitle string, actions fyne.CanvasObject) fyne.CanvasObject {
	titleText := canvas.NewText(title, colorText)
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.TextSize = 24

	left := container.NewVBox(titleText)
	if subtitle != "" {
		sub := canvas.NewText(subtitle, colorTextMuted)
		sub.TextSize = 13
		left.Add(sub)
	}

	if actions != nil {
		return container.NewBorder(nil, nil, left, container.NewCenter(actions))
	}
	return left
}

// coloredDot retourne un petit rond/carré arrondi de couleur (indicateur).
func coloredDot(c color.Color, size float32) fyne.CanvasObject {
	r := canvas.NewRectangle(c)
	r.CornerRadius = size / 2
	r.SetMinSize(fyne.NewSize(size, size))
	return r
}

// fixedWidth force une largeur minimale sur un objet (ex. la barre latérale),
// en superposant un rectangle transparent de la largeur voulue.
func fixedWidth(obj fyne.CanvasObject, width float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(width, 0))
	return container.NewStack(obj, container.NewVBox(spacer))
}
