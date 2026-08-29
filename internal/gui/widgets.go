package gui

import (
	"clefs/internal/db"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// widgets.go regroupe les composants d'affichage spécifiques au domaine
// (cartes de clé, d'accès, lignes à cocher, carte de trousseau). Ils s'appuient
// sur les briques génériques de ui_kit.go (card, badge, coloredDot) pour un
// rendu homogène et moderne.

// accessCard retourne une carte pour un accès (porte/zone) : nom en gras,
// méta-données discrètes, nombre de clés, et actions Modifier/Supprimer.
func accessCard(r db.Room, bName string, keyCount int, onEdit, onDelete func()) fyne.CanvasObject {
	nameText := canvas.NewText(r.Name, colorText)
	nameText.TextStyle = fyne.TextStyle{Bold: true}
	nameText.TextSize = 15

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
	metaText := canvas.NewText(strings.Join(meta, " · "), colorTextMuted)
	metaText.TextSize = 12

	keyBadge := badge(fmt.Sprintf("%d clé(s)", keyCount), colorPrimarySoft, colorPrimary)

	info := container.NewVBox(nameText, metaText, container.NewHBox(keyBadge))

	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), onEdit)
	editBtn.Importance = widget.LowImportance
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onDelete)
	deleteBtn.Importance = widget.LowImportance
	actions := container.NewHBox(editBtn, deleteBtn)

	row := container.NewBorder(nil, nil, nil, container.NewCenter(actions), info)
	return card(row)
}

// stockAdjuster retourne une rangée d'ajustement du stock total : un curseur et
// un champ numérique synchronisés, plus un bouton Enregistrer activé uniquement
// quand la valeur diffère du stock actuel.
func stockAdjuster(current int, onSave func(newTotal int)) fyne.CanvasObject {
	maxVal := 50
	if current*2 > maxVal {
		maxVal = current * 2
	}
	value := current

	entry := widget.NewEntry()
	entry.SetText(strconv.Itoa(current))
	// Retour visuel immédiat (champ rouge) sur saisie non numérique ou
	// négative ; la valeur invalide n'est de toute façon jamais enregistrée
	// (voir OnChanged ci-dessous, qui l'ignore).
	entry.Validator = func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return fmt.Errorf("nombre positif requis")
		}
		return nil
	}

	slider := widget.NewSlider(0, float64(maxVal))
	slider.Step = 1
	slider.Value = float64(current)

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.DocumentSaveIcon(), func() {
		onSave(value)
	})
	saveBtn.Importance = widget.HighImportance
	saveBtn.Disable()

	syncSaveBtn := func() {
		if value != current {
			saveBtn.Enable()
		} else {
			saveBtn.Disable()
		}
	}

	slider.OnChanged = func(v float64) {
		value = int(v)
		if entry.Text != strconv.Itoa(value) {
			entry.SetText(strconv.Itoa(value))
		}
		syncSaveBtn()
	}
	entry.OnChanged = func(s string) {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return
		}
		value = n
		if n <= maxVal && slider.Value != float64(n) {
			slider.Value = float64(n)
			slider.Refresh()
		}
		syncSaveBtn()
	}

	label := canvas.NewText("Ajuster le stock total", colorTextMuted)
	label.TextSize = 12

	return container.NewVBox(
		label,
		container.NewBorder(nil, nil, nil,
			container.NewHBox(fixedWidth(entry, 64), saveBtn),
			slider),
	)
}

// keyCard retourne une carte pour une clé avec une pastille de stock colorée
// (vert = dispo, orange = dernière unité, rouge = épuisée), une zone
// d'ajustement rapide du stock total et les actions.
func keyCard(k db.KeyWithAvailability, onLoan func(), onReturn func(), onSaveStock func(newTotal int)) fyne.CanvasObject {
	nameText := canvas.NewText(fmt.Sprintf("%s — %s", k.Number, k.Description), colorText)
	nameText.TextStyle = fyne.TextStyle{Bold: true}
	nameText.TextSize = 15

	meta := []string{}
	if k.Category != "" {
		meta = append(meta, k.Category)
	}
	if k.StorageLocation != "" {
		meta = append(meta, "rangement : "+k.StorageLocation)
	}
	metaText := canvas.NewText(strings.Join(meta, " · "), colorTextMuted)
	metaText.TextSize = 12

	// Pastille + badge de disponibilité selon le stock.
	var dotColor, badgeFill, badgeFg color.Color
	var availText string
	switch {
	case k.AvailableCount < 0:
		// Sur-prêt : plus d'exemplaires sortis que de stock utilisable.
		dotColor, badgeFill, badgeFg = colorDanger, colorDangerSoft, colorDanger
		availText = fmt.Sprintf("⚠ Erreur d'inventaire (%d)", k.AvailableCount)
	case k.AvailableCount == 0:
		dotColor, badgeFill, badgeFg = colorDanger, colorDangerSoft, colorDanger
		availText = "Indisponible"
	case k.AvailableCount == 1:
		dotColor, badgeFill, badgeFg = colorWarning, colorWarningSoft, colorWarning
		availText = "1 disponible"
	default:
		dotColor, badgeFill, badgeFg = colorSuccess, colorSuccessSoft, colorSuccess
		availText = fmt.Sprintf("%d disponibles", k.AvailableCount)
	}
	availBadge := badge(availText, badgeFill, badgeFg)

	stockText := canvas.NewText(
		fmt.Sprintf("Total %d  ·  Réserve %d  ·  Sorties %d",
			k.QuantityTotal, k.QuantityReserve, k.LoanedCount), colorTextMuted)
	stockText.TextSize = 12

	infoItems := []fyne.CanvasObject{
		nameText,
		metaText,
		container.NewHBox(availBadge),
		stockText,
	}
	if len(k.BorrowerNames) > 0 {
		holders := canvas.NewText("Détenu par : "+strings.Join(k.BorrowerNames, ", "), colorTextMuted)
		holders.TextSize = 12
		infoItems = append(infoItems, holders)
	}
	infoItems = append(infoItems, stockAdjuster(k.QuantityTotal, onSaveStock))
	info := container.NewVBox(infoItems...)

	// Boutons d'action.
	var actions []fyne.CanvasObject
	if k.AvailableCount > 0 {
		loanBtn := widget.NewButtonWithIcon("Emprunter", theme.ContentAddIcon(), onLoan)
		loanBtn.Importance = widget.HighImportance
		actions = append(actions, loanBtn)
	}
	if k.LoanedCount > 0 {
		returnBtn := widget.NewButtonWithIcon("Retour", theme.MailReplyIcon(), onReturn)
		actions = append(actions, returnBtn)
	}
	right := container.NewVBox(actions...)

	// Pastille de couleur à gauche comme repère visuel rapide.
	dot := container.NewVBox(coloredDot(dotColor, 12))

	body := container.NewBorder(nil, nil,
		container.NewVBox(dot),
		container.NewCenter(right),
		info)
	return card(body)
}

// accessCheckRow retourne une ligne à cocher pour la sélection d'accès lors d'un
// prêt. Le libellé est intégré au widget Check pour que toute la zone soit
// cliquable (la hitbox couvre nom + case).
func accessCheckRow(r db.Room, bName string, checked bool, onChange func(bool)) fyne.CanvasObject {
	sub := []string{}
	if bName != "" {
		sub = append(sub, bName)
	}
	if r.Floor != "" {
		sub = append(sub, r.Floor)
	}

	chk := widget.NewCheck(r.Name, onChange)
	chk.SetChecked(checked)

	var content fyne.CanvasObject
	if len(sub) > 0 {
		subText := canvas.NewText("      "+strings.Join(sub, " · "), colorTextMuted)
		subText.TextSize = 12
		content = container.NewVBox(chk, subText)
	} else {
		content = chk
	}
	return container.NewPadded(content)
}

// loanKeyCard représente une clé dans le trousseau calculé d'un prêt : nom,
// badge d'origine (suggérée / manuelle), accès couverts, et bouton Retirer.
func loanKeyCard(k db.Key, coveredAccesses []string, suggested bool, onRemove func()) fyne.CanvasObject {
	nameText := canvas.NewText(fmt.Sprintf("%s — %s", k.Number, k.Description), colorText)
	nameText.TextStyle = fyne.TextStyle{Bold: true}
	nameText.TextSize = 14

	var originBadge fyne.CanvasObject
	if suggested {
		originBadge = badge("Suggérée", colorInfoSoft, colorPrimary)
	} else {
		originBadge = badge("Manuelle", colorWarningSoft, colorWarning)
	}

	accessStr := "—"
	if len(coveredAccesses) > 0 {
		accessStr = strings.Join(coveredAccesses, ", ")
	}
	accessLabel := widget.NewLabel("Accès couverts : " + accessStr)
	accessLabel.Wrapping = fyne.TextWrapWord

	removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onRemove)
	removeBtn.Importance = widget.LowImportance

	info := container.NewVBox(
		container.NewHBox(nameText),
		container.NewHBox(originBadge),
		accessLabel,
	)
	body := container.NewBorder(nil, nil, nil, container.NewCenter(removeBtn), info)
	return card(body)
}

// inventoryAlertBanner construit le bandeau rouge « erreur d'inventaire »
// listant les clés dont plus d'exemplaires sont sortis que le stock utilisable.
// Affiché sur le tableau de bord et en tête de la vue clés ; purement
// informatif, il n'empêche aucune action.
func inventoryAlertBanner(anomalies []db.InventoryAnomaly) fyne.CanvasObject {
	title := canvas.NewText("⚠ Erreur d'inventaire, vérifier le stock", colorDanger)
	title.TextStyle = fyne.TextStyle{Bold: true}

	box := container.NewVBox(title)
	for _, a := range anomalies {
		line := canvas.NewText(fmt.Sprintf(
			"Clé n° %s : %d sortie(s) pour %d utilisable(s) (total %d, réserve %d)",
			a.KeyNumber, a.Loaned, a.Total-a.Reserve, a.Total, a.Reserve), colorDanger)
		line.TextSize = 12
		box.Add(line)
	}
	return coloredCard(box, colorDangerSoft, colorDanger)
}
