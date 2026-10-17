// Package pdf génère les documents imprimables de l'application avec gofpdf.
//
// Tous les documents partagent le même modèle, aligné sur l'interface V4 :
// police Inter embarquée (UTF-8, accents et typographie française sans
// conversion), couleurs du thème clair (frontend/src/app.css), en-tête avec
// logo, cartes d'informations, tableaux à lignes fines et badges colorés,
// pied de page avec date de génération et pagination.
//
// doc.go contient ce modèle commun ; documents.go les quatre documents
// (bon de remise, plan de clés, bilan du stock, rapport d'inventaire) ;
// exporter.go l'écriture dans documents/. Les générateurs reçoivent des
// structures simples (pas d'accès à la base) : ils sont testables isolément.
package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/phpdave11/gofpdf"
)

// Police Inter (SIL Open Font License, voir fonts/OFL.txt).
var (
	//go:embed fonts/Inter-Regular.ttf
	fontRegular []byte
	//go:embed fonts/Inter-SemiBold.ttf
	fontSemiBold []byte
)

// rgb est une couleur du thème.
type rgb struct{ r, g, b int }

// Couleurs du thème clair de l'application (mêmes valeurs que app.css).
var (
	cPrimary     = rgb{79, 70, 229}
	cPrimarySoft = rgb{238, 240, 255}
	cInk         = rgb{20, 26, 38}
	cMuted       = rgb{102, 112, 133}
	cLine        = rgb{227, 230, 238}
	cSurface2    = rgb{240, 242, 247}
	cCard        = rgb{245, 246, 250}
	cSuccess     = rgb{14, 138, 95}
	cSuccessSoft = rgb{231, 247, 240}
	cWarning     = rgb{180, 83, 9}
	cWarningSoft = rgb{253, 243, 226}
	cDanger      = rgb{217, 45, 32}
	cDangerSoft  = rgb{253, 236, 234}
	cWhite       = rgb{255, 255, 255}
)

// Tons des badges et des valeurs mises en avant.
type tone int

const (
	toneNone tone = iota
	toneNeutral
	tonePrimary
	toneSuccess
	toneWarning
	toneDanger
)

// colors retourne (texte, fond) d'un ton.
func (t tone) colors() (rgb, rgb) {
	switch t {
	case tonePrimary:
		return cPrimary, cPrimarySoft
	case toneSuccess:
		return cSuccess, cSuccessSoft
	case toneWarning:
		return cWarning, cWarningSoft
	case toneDanger:
		return cDanger, cDangerSoft
	default:
		return cMuted, cSurface2
	}
}

// Mise en page A4 (mm).
const (
	pageW     = 210.0
	marginX   = 16.0
	contentW  = pageW - 2*marginX
	topY      = 30.0  // début du contenu sous l'en-tête
	bottomY   = 275.0 // limite basse du contenu (pied de page en dessous)
	fontFam   = "Inter"
	lineRatio = 0.42 // hauteur de ligne ≈ taille de police (pt) × ratio (mm)
)

// doc est un document au modèle de l'application.
type doc struct {
	*gofpdf.Fpdf
	generated time.Time
	s         Settings
	logoW     float64 // largeur du logo dans l'en-tête (0 = clé stylisée)
	logoHgt   float64 // hauteur du logo (≤ logoH)
}

// logoH est la hauteur du logo dans l'en-tête.
const logoH = 9.0

// newDoc crée un document A4 d'une page, avec en-tête, pied de page et bloc de titre.
func newDoc(s Settings, title, subtitle string) *doc {
	d := newBlankDoc(s, title)
	d.page(title, subtitle)
	return d
}

// newBlankDoc crée un document A4 sans page ; headerTitle est rappelé dans
// l'en-tête de chaque page.
func newBlankDoc(s Settings, headerTitle string) *doc {
	s = s.withDefaults()
	p := gofpdf.New("P", "mm", "A4", "")
	p.SetMargins(marginX, topY, marginX)
	p.SetAutoPageBreak(false, 0)
	p.AddUTF8FontFromBytes(fontFam, "", fontRegular)
	p.AddUTF8FontFromBytes(fontFam, "B", fontSemiBold)
	p.SetTitle(headerTitle, true)
	p.SetAuthor(s.AppName+" — "+s.Establishment, true)
	p.SetCreator(s.AppName, true)
	p.AliasNbPages("{nb}")

	d := &doc{Fpdf: p, generated: time.Now(), s: s}
	// Logo de l'établissement : ajusté à la hauteur de l'en-tête, largeur
	// limitée ; un logo illisible est ignoré (clé stylisée par défaut).
	if len(s.Logo) > 0 {
		if format, w, h, err := ValidateLogo(s.Logo); err == nil && w > 0 && h > 0 {
			p.RegisterImageOptionsReader("logo", gofpdf.ImageOptions{ImageType: format}, bytes.NewReader(s.Logo))
			if p.Ok() {
				d.logoW = min(logoH*float64(w)/float64(h), 30)
				d.logoHgt = d.logoW * float64(h) / float64(w)
			} else {
				p.ClearError()
			}
		}
	}
	p.SetHeaderFunc(func() { d.header(headerTitle) })
	p.SetFooterFunc(d.footer)
	return d
}

// page ouvre une nouvelle page avec un bloc de titre.
func (d *doc) page(title, subtitle string) {
	d.AddPage()
	d.font(true, 19, cInk)
	d.SetXY(marginX, topY)
	d.CellFormat(contentW, 9, title, "", 1, "L", false, 0, "")
	if subtitle != "" {
		d.font(false, 10, cMuted)
		d.multi(contentW, subtitle, 10)
	}
	d.SetY(d.GetY() + 5)
}

// font sélectionne la police (semi-gras ou normale), la taille et la couleur.
func (d *doc) font(bold bool, size float64, c rgb) {
	style := ""
	if bold {
		style = "B"
	}
	d.SetFont(fontFam, style, size)
	d.SetTextColor(c.r, c.g, c.b)
}

func (d *doc) fill(c rgb) { d.SetFillColor(c.r, c.g, c.b) }
func (d *doc) draw(c rgb) { d.SetDrawColor(c.r, c.g, c.b) }

// lineH retourne la hauteur de ligne pour une taille de police.
func lineH(size float64) float64 { return size * lineRatio }

// header dessine l'en-tête de chaque page : logo, nom de l'application et
// de l'établissement, titre du document à droite.
func (d *doc) header(title string) {
	x, y := marginX, 10.0
	textX := x + 12
	if d.logoW > 0 {
		// Logo centré dans sa zone, hauteur logoH au plus.
		d.ImageOptions("logo", x, y+(logoH-d.logoHgt)/2, d.logoW, d.logoHgt, false, gofpdf.ImageOptions{}, 0, "")
		textX = x + d.logoW + 3
	} else {
		d.fill(cPrimary)
		d.RoundedRect(x, y, logoH, logoH, 2.2, "1234", "F")
		d.drawKeyGlyph(x, y, logoH)
	}

	textW := contentW/2 + marginX - textX + 20
	d.SetXY(textX, y+0.6)
	d.font(true, 10.5, cInk)
	d.CellFormat(textW, 4.6, d.ellipsis(d.s.AppName, textW), "", 2, "L", false, 0, "")
	d.SetX(textX)
	d.font(false, 8.5, cMuted)
	d.CellFormat(textW, 4, d.ellipsis(d.s.establishmentLine(), textW), "", 0, "L", false, 0, "")

	d.SetXY(marginX, y+2.5)
	d.font(false, 8.5, cMuted)
	d.CellFormat(contentW, 4, title, "", 0, "R", false, 0, "")

	d.draw(cLine)
	d.SetLineWidth(0.3)
	d.Line(marginX, 23, pageW-marginX, 23)
	d.SetY(topY)
}

// drawKeyGlyph dessine une clé blanche stylisée dans le carré du logo.
func (d *doc) drawKeyGlyph(x, y, s float64) {
	d.draw(cWhite)
	d.SetLineCapStyle("round")
	d.SetLineWidth(s * 0.09)
	cx, cy := x+s*0.36, y+s*0.36
	d.Circle(cx, cy, s*0.14, "D")
	d.Line(cx+s*0.1, cy+s*0.1, x+s*0.76, y+s*0.76)
	d.Line(x+s*0.62, y+s*0.62, x+s*0.52, y+s*0.72)
	d.Line(x+s*0.71, y+s*0.71, x+s*0.63, y+s*0.79)
	d.SetLineCapStyle("butt")
}

// footer dessine le pied de page : date de génération, mention éventuelle
// de l'établissement et pagination.
func (d *doc) footer() {
	y := 283.0
	d.draw(cLine)
	d.SetLineWidth(0.3)
	d.Line(marginX, y, pageW-marginX, y)
	d.font(false, 8, cMuted)
	d.SetXY(marginX, y+2)
	left := "Généré le " + d.generated.Format("02/01/2006 à 15:04")
	if note := strings.TrimSpace(d.s.FooterNote); note != "" {
		left += "  ·  " + note
	}
	d.CellFormat(contentW-25, 4, d.ellipsis(left, contentW-25), "", 0, "L", false, 0, "")
	d.CellFormat(25, 4, fmt.Sprintf("Page %d / {nb}", d.PageNo()), "", 0, "R", false, 0, "")
}

// ensure passe à la page suivante s'il reste moins de h mm.
func (d *doc) ensure(h float64) bool {
	if d.GetY()+h > bottomY {
		d.AddPage()
		return true
	}
	return false
}

// multi écrit un paragraphe sur la largeur w, avec retour à la ligne.
func (d *doc) multi(w float64, text string, size float64) {
	x := d.GetX()
	for _, l := range d.wrap(text, w) {
		d.SetX(x)
		d.CellFormat(w, lineH(size)+0.8, l, "", 2, "L", false, 0, "")
	}
}

// wrap découpe un texte en lignes tenant dans la largeur w avec la police
// courante (mot par mot ; un mot trop long est coupé). Sûr en UTF-8.
func (d *doc) wrap(text string, w float64) []string {
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if d.GetStringWidth(candidate) <= w {
				line = candidate
				continue
			}
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			for d.GetStringWidth(word) > w {
				cut := fitRunes(d, word, w)
				lines = append(lines, word[:cut])
				word = word[cut:]
			}
			line = word
		}
		lines = append(lines, line)
	}
	return lines
}

// fitRunes retourne la longueur en octets du plus long préfixe de s (coupé
// sur une frontière de caractère) qui tient dans la largeur w.
func fitRunes(d *doc, s string, w float64) int {
	end := 0
	for i := range s {
		if i > 0 && d.GetStringWidth(s[:i]) > w {
			break
		}
		end = i
	}
	if end == 0 {
		_, size := utf8.DecodeRuneInString(s)
		return size
	}
	return end
}

// section écrit un titre de section (petites capitales grises).
func (d *doc) section(title string) {
	d.ensure(22)
	d.SetY(d.GetY() + 3)
	d.font(true, 8.5, cMuted)
	d.SetX(marginX)
	d.CellFormat(contentW, 6, strings.ToUpper(title), "", 1, "L", false, 0, "")
	d.SetY(d.GetY() + 1)
}

// field est un couple libellé / valeur d'une carte d'informations.
type field struct{ label, value string }

// infoCard affiche une carte grise de couples libellé / valeur en colonnes.
func (d *doc) infoCard(fields []field, cols int) {
	rows := (len(fields) + cols - 1) / cols
	rowH := 11.0
	h := float64(rows)*rowH + 5
	d.ensure(h + 4)
	y0 := d.GetY()
	d.fill(cCard)
	d.RoundedRect(marginX, y0, contentW, h, 2.5, "1234", "F")
	colW := (contentW - 10) / float64(cols)
	for i, f := range fields {
		x := marginX + 5 + float64(i%cols)*colW
		y := y0 + 3 + float64(i/cols)*rowH
		d.SetXY(x, y)
		d.font(false, 8, cMuted)
		d.CellFormat(colW-3, 4, f.label, "", 2, "L", false, 0, "")
		d.SetX(x)
		d.font(true, 10.5, cInk)
		d.CellFormat(colW-3, 5.5, d.ellipsis(f.value, colW-3), "", 0, "L", false, 0, "")
	}
	d.SetY(y0 + h + 4)
}

// stat est un indicateur chiffré.
type stat struct {
	label, value string
	tone         tone
}

// stats affiche une rangée d'indicateurs.
func (d *doc) stats(items []stat) {
	gap := 3.0
	w := (contentW - gap*float64(len(items)-1)) / float64(len(items))
	h := 17.0
	d.ensure(h + 4)
	y := d.GetY()
	for i, s := range items {
		x := marginX + float64(i)*(w+gap)
		bg := cCard
		fg := cInk
		if s.tone != toneNone && s.tone != toneNeutral {
			fg, bg = s.tone.colors()
		}
		d.fill(bg)
		d.RoundedRect(x, y, w, h, 2.5, "1234", "F")
		d.SetXY(x+4, y+2.8)
		d.font(true, 15, fg)
		d.CellFormat(w-8, 7, s.value, "", 2, "L", false, 0, "")
		d.SetX(x + 4)
		d.font(false, 8, cMuted)
		d.CellFormat(w-8, 4, d.ellipsis(s.label, w-8), "", 0, "L", false, 0, "")
	}
	d.SetY(y + h + 4)
}

// ellipsis tronque un texte (sur une frontière de caractère) pour qu'il tienne en w.
func (d *doc) ellipsis(s string, w float64) string {
	if d.GetStringWidth(s) <= w {
		return s
	}
	cut := fitRunes(d, s, w-d.GetStringWidth("…"))
	return strings.TrimSpace(s[:cut]) + "…"
}

// Types de cellule d'un tableau.
const (
	cellText  = iota // texte, retour à la ligne automatique
	cellKey          // numéro de clé dans une pastille
	cellBadge        // badge coloré
	cellBold         // texte semi-gras
	cellMuted        // texte gris
)

// cell est une cellule de tableau.
type cell struct {
	text string
	kind int
	tone tone
}

// Constructeurs de cellules.
func txt(s string) cell           { return cell{text: s} }
func bold(s string) cell          { return cell{text: s, kind: cellBold} }
func muted(s string) cell         { return cell{text: s, kind: cellMuted} }
func key(s string) cell           { return cell{text: s, kind: cellKey} }
func badge(s string, t tone) cell { return cell{text: s, kind: cellBadge, tone: t} }
func num(n int) cell              { return cell{text: fmt.Sprint(n)} }
func boldNum(n int) cell          { return cell{text: fmt.Sprint(n), kind: cellBold} }

// orDash retourne « — » pour une valeur vide.
func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// plural accorde un nom : plural(2, "clé", "clés") → « 2 clés ».
func plural(n int, one, many string) string {
	if n > 1 {
		return fmt.Sprintf("%d %s", n, many)
	}
	return fmt.Sprintf("%d %s", n, one)
}

// column décrit une colonne : en-tête, largeur relative et alignement (L, C, R).
type column struct {
	header string
	weight float64
	align  string
}

// Mesures des tableaux.
const (
	cellPadX   = 2.5
	headerSize = 7.5
	headerH    = 7.0
)

// tableStyle règle la position et la densité d'un tableau.
type tableStyle struct {
	x, w  float64 // position et largeur
	size  float64 // taille de police des cellules
	padY  float64 // marge verticale des cellules
	wrap  bool    // retour à la ligne (sinon une ligne par cellule, tronquée « … »)
	pages bool    // saut de page automatique avec en-tête répété
}

// defaultTable est le style courant : pleine largeur, retour à la ligne.
var defaultTable = tableStyle{x: marginX, w: contentW, size: 9, padY: 2.4, wrap: true, pages: true}

// rowHeight retourne la hauteur d'une ligne d'une seule ligne de texte.
func (s tableStyle) rowHeight() float64 { return lineH(s.size) + 2*s.padY }

// table affiche un tableau au style courant.
func (d *doc) table(cols []column, rows [][]cell) {
	d.tableWith(defaultTable, cols, rows)
	d.SetY(d.GetY() + 4)
}

// tableWith affiche un tableau : hauteur de ligne adaptée au contenu, lignes
// séparées par un filet, en-tête répété à chaque page si s.pages.
func (d *doc) tableWith(s tableStyle, cols []column, rows [][]cell) {
	widths := s.columnWidths(cols)
	if s.pages {
		d.ensure(headerH + 10)
	}
	d.tableHeader(s, cols, widths)
	for _, row := range rows {
		lines, h := d.measureRow(s, widths, row)
		if s.pages && d.ensure(h) {
			d.tableHeader(s, cols, widths)
		}
		y := d.GetY()
		x := s.x
		for i, c := range row {
			d.drawCell(s, c, lines[i], x, y, widths[i], cols[i].align)
			x += widths[i]
		}
		d.draw(cLine)
		d.SetLineWidth(0.2)
		d.Line(s.x, y+h, s.x+s.w, y+h)
		d.SetY(y + h)
	}
}

// columnWidths répartit la largeur du tableau selon le poids des colonnes.
func (s tableStyle) columnWidths(cols []column) []float64 {
	total := 0.0
	for _, c := range cols {
		total += c.weight
	}
	widths := make([]float64, len(cols))
	for i, c := range cols {
		widths[i] = s.w * c.weight / total
	}
	return widths
}

// measureRow découpe le texte des cellules d'une ligne et retourne la
// hauteur de la ligne.
func (d *doc) measureRow(s tableStyle, widths []float64, row []cell) ([][]string, float64) {
	lines := make([][]string, len(row))
	h := 0.0
	for i, c := range row {
		d.cellFont(c, s.size)
		inner := widths[i] - 2*cellPadX
		// Un numéro de clé trop long pour sa pastille passe à la ligne
		// en texte semi-gras plutôt que d'être tronqué.
		if s.wrap && c.kind == cellKey && d.GetStringWidth(c.text)+3.4 > inner {
			row[i].kind = cellBold
			c = row[i]
		}
		switch {
		case c.kind == cellKey || c.kind == cellBadge:
			lines[i] = []string{c.text}
		case s.wrap:
			lines[i] = d.wrap(c.text, inner)
		default:
			lines[i] = []string{d.ellipsis(c.text, inner)}
		}
		if ch := float64(len(lines[i]))*lineH(s.size) + 2*s.padY; ch > h {
			h = ch
		}
	}
	return lines, h
}

// tableHeight retourne la hauteur qu'occuperait un tableau (sans saut de page).
func (d *doc) tableHeight(s tableStyle, cols []column, rows [][]cell) float64 {
	widths := s.columnWidths(cols)
	h := headerH
	for _, row := range rows {
		_, rh := d.measureRow(s, widths, row)
		h += rh
	}
	return h
}

// tableHeader dessine la ligne d'en-tête d'un tableau.
func (d *doc) tableHeader(s tableStyle, cols []column, widths []float64) {
	y := d.GetY()
	d.fill(cSurface2)
	d.RoundedRect(s.x, y, s.w, headerH, 1.5, "1234", "F")
	d.font(true, headerSize, cMuted)
	x := s.x
	for i, c := range cols {
		d.SetXY(x+cellPadX, y)
		d.CellFormat(widths[i]-2*cellPadX, headerH, d.ellipsis(strings.ToUpper(c.header), widths[i]-2*cellPadX), "", 0, c.align, false, 0, "")
		x += widths[i]
	}
	d.SetY(y + headerH)
}

// cellFont sélectionne la police d'une cellule.
func (d *doc) cellFont(c cell, size float64) {
	switch c.kind {
	case cellBold, cellKey, cellBadge:
		d.font(true, size, cInk)
	case cellMuted:
		d.font(false, size, cMuted)
	default:
		d.font(false, size, cInk)
	}
}

// drawCell dessine une cellule à sa position.
func (d *doc) drawCell(s tableStyle, c cell, lines []string, x, y, w float64, align string) {
	d.cellFont(c, s.size)
	inner := w - 2*cellPadX
	lh := lineH(s.size)
	switch c.kind {
	case cellKey, cellBadge:
		label := d.ellipsis(c.text, inner-3)
		bw := d.GetStringWidth(label) + 3.4
		bx := x + cellPadX
		switch align {
		case "C":
			bx = x + (w-bw)/2
		case "R":
			bx = x + w - cellPadX - bw
		}
		fg, bg := c.tone.colors()
		if c.kind == cellKey {
			fg, bg = cInk, cSurface2
		}
		d.fill(bg)
		d.RoundedRect(bx, y+s.padY-0.6, bw, lh+1.2, 1.2, "1234", "F")
		d.SetTextColor(fg.r, fg.g, fg.b)
		d.SetXY(bx, y+s.padY)
		d.CellFormat(bw, lh, label, "", 0, "C", false, 0, "")
	default:
		for j, l := range lines {
			d.SetXY(x+cellPadX, y+s.padY+float64(j)*lh)
			d.CellFormat(inner, lh, l, "", 0, align, false, 0, "")
		}
	}
}

// note affiche un encadré coloré (avertissement, information).
func (d *doc) note(text string, t tone) {
	fg, bg := t.colors()
	d.font(false, 9, cInk)
	lines := d.wrap(text, contentW-10)
	h := float64(len(lines))*lineH(9) + 6
	d.ensure(h + 3)
	y := d.GetY()
	d.fill(bg)
	d.RoundedRect(marginX, y, contentW, h, 2, "1234", "F")
	d.fill(fg)
	d.Rect(marginX, y+1.5, 0.9, h-3, "F")
	for i, l := range lines {
		d.SetXY(marginX+5, y+3+float64(i)*lineH(9))
		d.CellFormat(contentW-10, lineH(9), l, "", 0, "L", false, 0, "")
	}
	d.SetY(y + h + 3)
}

// output finalise le document et retourne ses octets.
func (d *doc) output() ([]byte, error) {
	var buf bytes.Buffer
	if err := d.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
