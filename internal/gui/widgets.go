package gui

import (
	"clefs/internal/db"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// accessCard retourne une card plate pour un accès (porte/zone).
// Nom en gras, bâtiment·étage·catégorie en sous-ligne, nb clés, boutons droite.
func accessCard(r db.Room, bName string, keyCount int, onEdit, onDelete func()) fyne.CanvasObject {
	nameLabel := widget.NewLabelWithStyle(r.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	meta := []string{}
	if bName != "" {
		meta = append(meta, bName)
	}
	if r.Floor != "" {
		meta = append(meta, r.Floor)
	}
	if r.Category != "" {
		meta = append(meta, r.Category)
	}
	metaLabel := widget.NewLabel(strings.Join(meta, " · "))
	metaLabel.TextStyle = fyne.TextStyle{Italic: true}

	keyLabel := widget.NewLabel(fmt.Sprintf("%d clé(s) associée(s)", keyCount))

	info := container.NewVBox(nameLabel, metaLabel, keyLabel)

	editBtn := widget.NewButton("Modifier", onEdit)
	deleteBtn := widget.NewButton("Supprimer", onDelete)
	deleteBtn.Importance = widget.DangerImportance
	actions := container.NewVBox(editBtn, deleteBtn)

	row := container.NewBorder(nil, nil, nil, actions, info)
	return container.NewVBox(row, widget.NewSeparator())
}

// keyCard retourne une card plate pour une clé avec indicateur de stock coloré.
func keyCard(k db.KeyWithAvailability, onLoan func(), onReturn func()) fyne.CanvasObject {
	nameLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("%s — %s", k.Number, k.Description),
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
	)

	meta := []string{}
	if k.Category != "" {
		meta = append(meta, k.Category)
	}
	if k.StorageLocation != "" {
		meta = append(meta, "rangement : "+k.StorageLocation)
	}
	metaLabel := widget.NewLabel(strings.Join(meta, " · "))
	metaLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Indicateur stock coloré
	stockText := fmt.Sprintf("Stock : %d  |  Réserve : %d  |  Sorties : %d  |  Dispo : %d",
		k.QuantityTotal, k.QuantityReserve, k.LoanedCount, k.AvailableCount)
	stockLabel := widget.NewLabel(stockText)

	// Couleur du rectangle indicateur
	var indicatorColor color.Color
	switch {
	case k.AvailableCount <= 0:
		indicatorColor = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
	case k.AvailableCount == 1:
		indicatorColor = color.NRGBA{R: 220, G: 150, B: 0, A: 255}
	default:
		indicatorColor = color.NRGBA{R: 50, G: 180, B: 50, A: 255}
	}
	indicator := canvas.NewRectangle(indicatorColor)
	indicator.SetMinSize(fyne.NewSize(6, 0))

	// Emprunteurs actuels
	var detailLines []fyne.CanvasObject
	if len(k.BorrowerNames) > 0 {
		detailLines = append(detailLines, widget.NewLabel("Détenu par : "+strings.Join(k.BorrowerNames, ", ")))
	}

	info := container.NewVBox(append([]fyne.CanvasObject{nameLabel, metaLabel, stockLabel}, detailLines...)...)

	// Boutons action
	var actions []fyne.CanvasObject
	if k.AvailableCount > 0 {
		loanBtn := widget.NewButton("Emprunter", onLoan)
		loanBtn.Importance = widget.HighImportance
		actions = append(actions, loanBtn)
	}
	if k.LoanedCount > 0 {
		returnBtn := widget.NewButton("Retour", onReturn)
		actions = append(actions, returnBtn)
	}

	right := container.NewVBox(actions...)
	body := container.NewBorder(nil, nil, indicator, right, container.NewPadded(info))
	return container.NewVBox(body, widget.NewSeparator())
}

// accessCheckRow retourne une ligne de checkbox stylisée pour la sélection d'accès dans le prêt.
// Le nom est intégré dans le widget Check pour que toute la zone soit cliquable.
func accessCheckRow(r db.Room, bName string, checked bool, onChange func(bool)) fyne.CanvasObject {
	sub := []string{}
	if bName != "" {
		sub = append(sub, bName)
	}
	if r.Floor != "" {
		sub = append(sub, r.Floor)
	}

	// Le label est mis directement dans le Check pour maximiser la hitbox
	chk := widget.NewCheck(r.Name, onChange)
	chk.SetChecked(checked)

	if len(sub) > 0 {
		subLabel := widget.NewLabel("    " + strings.Join(sub, " · "))
		subLabel.TextStyle = fyne.TextStyle{Italic: true}
		return container.NewVBox(container.NewVBox(chk, subLabel), widget.NewSeparator())
	}
	return container.NewVBox(chk, widget.NewSeparator())
}

// loanKeyCard retourne la représentation d'une clé dans le trousseau calculé.
func loanKeyCard(k db.Key, coveredAccesses []string, suggested bool, onRemove func()) fyne.CanvasObject {
	nameLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("%s — %s", k.Number, k.Description),
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
	)

	badge := "ajoutée manuellement"
	if suggested {
		badge = "suggestion automatique"
	}
	badgeLabel := widget.NewLabel(badge)
	badgeLabel.TextStyle = fyne.TextStyle{Italic: true}

	accessText := "Accès couverts : "
	if len(coveredAccesses) == 0 {
		accessText += "—"
	} else {
		accessText += strings.Join(coveredAccesses, ", ")
	}
	accessLabel := widget.NewLabel(accessText)
	accessLabel.Wrapping = fyne.TextWrapWord

	removeBtn := widget.NewButton("Retirer", onRemove)
	removeBtn.Importance = widget.DangerImportance

	info := container.NewVBox(nameLabel, badgeLabel, accessLabel)
	row := container.NewBorder(nil, nil, nil, removeBtn, info)
	return container.NewVBox(container.NewPadded(row), widget.NewSeparator())
}
