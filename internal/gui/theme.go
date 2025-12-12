package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ModernTheme est un thème personnalisé moderne pour l'application
type ModernTheme struct{}

// Color palette simple et claire
var (
	// Couleurs principales
	primaryColor = color.NRGBA{R: 0, G: 123, B: 255, A: 255}  // Bleu vif
	primaryDark  = color.NRGBA{R: 0, G: 86, B: 179, A: 255}   // Bleu foncé
	primaryLight = color.NRGBA{R: 66, G: 165, B: 245, A: 255} // Bleu clair

	// Couleurs secondaires
	secondaryColor = color.NRGBA{R: 108, G: 117, B: 125, A: 255} // Gris
	accentColor    = color.NRGBA{R: 40, G: 167, B: 69, A: 255}   // Vert succès
	warningColor   = color.NRGBA{R: 255, G: 193, B: 7, A: 255}   // Jaune warning
	dangerColor    = color.NRGBA{R: 220, G: 53, B: 69, A: 255}   // Rouge danger

	// Couleurs de fond - Très claires
	backgroundColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc pur
	surfaceColor    = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc
	cardBackground  = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc pour les cards

	// Couleurs de texte - Maximum de contraste
	textPrimary   = color.NRGBA{R: 0, G: 0, B: 0, A: 255}       // Noir pur
	textSecondary = color.NRGBA{R: 73, G: 80, B: 87, A: 255}    // Gris foncé
	textOnPrimary = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc

	// Autres
	shadowColor = color.NRGBA{R: 0, G: 0, B: 0, A: 30}        // Ombre légère
	borderColor = color.NRGBA{R: 222, G: 226, B: 230, A: 255} // Bordure grise
	hoverColor  = color.NRGBA{R: 248, G: 249, B: 250, A: 255} // Gris très clair pour hover
)

// Color retourne les couleurs du thème
func (m ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc pur pour le fond

	case theme.ColorNameButton:
		return primaryColor

	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	case theme.ColorNameDisabled:
		return color.NRGBA{R: 150, G: 150, B: 150, A: 255}

	case theme.ColorNameError:
		return dangerColor

	case theme.ColorNameFocus:
		return primaryLight

	case theme.ColorNameForeground:
		return textPrimary

	case theme.ColorNameHover:
		return hoverColor

	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	case theme.ColorNameInputBorder:
		return borderColor

	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Blanc pour les popups

	case theme.ColorNamePlaceHolder:
		return textSecondary

	case theme.ColorNamePressed:
		return primaryDark

	case theme.ColorNamePrimary:
		return primaryColor

	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}

	case theme.ColorNameSelection:
		return color.NRGBA{R: 41, G: 128, B: 185, A: 50}

	case theme.ColorNameSeparator:
		return borderColor

	case theme.ColorNameShadow:
		return shadowColor

	case theme.ColorNameSuccess:
		return accentColor

	case theme.ColorNameWarning:
		return warningColor

	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font retourne la police du thème
func (m ModernTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Utiliser les polices par défaut de Fyne pour l'instant
	// On pourrait intégrer des polices personnalisées ici
	return theme.DefaultTheme().Font(style)
}

// Icon retourne les icônes du thème
func (m ModernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// Utiliser les icônes par défaut pour l'instant
	// On pourrait créer des icônes personnalisées ici
	return theme.DefaultTheme().Icon(name)
}

// Size retourne les tailles du thème
func (m ModernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 5 // Coins arrondis
	case theme.SizeNameSelectionRadius:
		return 3
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// CreateStyledButton crée un bouton avec un style personnalisé
func CreateStyledButton(label string, icon fyne.Resource, importance widget.Importance, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, tapped)
	btn.Importance = importance
	return btn
}

// CreateCard crée un conteneur style "card" avec ombre
func CreateCard(title string, content fyne.CanvasObject) *fyne.Container {
	// Titre de la card
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Séparateur
	separator := widget.NewSeparator()

	// Conteneur avec padding
	cardContent := container.NewVBox(
		titleLabel,
		separator,
		content,
	)

	// Card avec bordure et fond blanc
	card := container.NewBorder(
		container.NewPadded(cardContent),
		nil, nil, nil,
	)

	return card
}

// CreateModernTable crée un tableau avec style moderne
func CreateModernTable(data [][]string, headers []string) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(headers) // +1 pour les headers
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if i.Row == 0 {
				// Headers
				label.SetText(headers[i.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Alignment = fyne.TextAlignCenter
			} else {
				// Data
				if i.Row-1 < len(data) && i.Col < len(data[i.Row-1]) {
					label.SetText(data[i.Row-1][i.Col])
					label.TextStyle = fyne.TextStyle{}
					label.Alignment = fyne.TextAlignLeading
				}
			}
		})

	// Définir les largeurs de colonnes
	for i := range headers {
		table.SetColumnWidth(i, 150)
	}

	return table
}

// CreateInfoCard crée une card d'information avec icône
func CreateInfoCard(title string, value string, icon fyne.Resource, colorName fyne.ThemeColorName) *fyne.Container {
	// Valeur en gros
	valueLabel := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Titre
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{})

	// Conteneur vertical
	content := container.NewVBox(
		valueLabel,
		titleLabel,
	)

	// Card avec padding
	card := container.NewPadded(content)

	return card
}

// CreateSidebarButton crée un bouton pour la sidebar avec style moderne
func CreateSidebarButton(label string, icon string, tapped func()) *widget.Button {
	btn := widget.NewButton(icon+" "+label, tapped)
	return btn
}

// ApplyModernTheme applique le thème moderne à l'application
func ApplyModernTheme(app fyne.App) {
	app.Settings().SetTheme(&ModernTheme{})
}
