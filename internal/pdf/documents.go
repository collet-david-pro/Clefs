package pdf

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"clefs/internal/natsort"
)

// documents.go : les documents de l'application, tous au modèle de doc.go.

// ============= BON DE REMISE =============

// Receipt décrit un bon de remise de clés.
type Receipt struct {
	BorrowerName   string
	BorrowerStatus string
	Date           time.Time  // date de la remise (ou de l'édition pour une réédition)
	LoanType       string     // ponctuel, permanent
	PlannedReturn  *time.Time // nil si non précisé
	Agent          string
	Keys           []ReceiptKey
	Reissue        bool // bon récapitulatif de toutes les clés détenues
}

// ReceiptKey est une clé remise.
type ReceiptKey struct {
	Number, Description, Storage string
	LoanDate                     time.Time
}

// GenerateReceipt génère le bon de remise, à faire signer par le détenteur.
// Il tient toujours sur une seule page A4 (voir receiptKeys).
func GenerateReceipt(s Settings, r Receipt) ([]byte, error) {
	if len(r.Keys) == 0 {
		return nil, fmt.Errorf("aucune clé à faire figurer sur le bon")
	}
	d := newBlankDoc(s, s.withDefaults().ReceiptTitle)
	d.receipt(r, "")
	return d.output()
}

// GenerateReceiptArchive regroupe dans un seul PDF les bons récapitulatifs de
// tous les détenteurs (une page par détenteur), précédés d'un sommaire : à
// imprimer ou à conserver pour l'archivage.
func GenerateReceiptArchive(s Settings, receipts []Receipt) ([]byte, error) {
	d, err := receiptArchive(s, receipts)
	if err != nil {
		return nil, err
	}
	return d.output()
}

// receiptArchive construit le document de GenerateReceiptArchive.
func receiptArchive(s Settings, receipts []Receipt) (*doc, error) {
	if len(receipts) == 0 {
		return nil, fmt.Errorf("aucun prêt en cours : aucun bon à éditer")
	}
	keys := 0
	for _, r := range receipts {
		keys += len(r.Keys)
	}
	const title = "Archive des bons de remise"
	d := newBlankDoc(s, title)
	d.page(title, fmt.Sprintf("Bons récapitulatifs de %s, soit %s en cours, au %s.",
		plural(len(receipts), "détenteur", "détenteurs"), plural(keys, "clé", "clés"), d.generated.Format("02/01/2006")))
	rows := make([][]cell, 0, len(receipts))
	for i, r := range receipts {
		rows = append(rows, []cell{muted(fmt.Sprint(i + 1)), bold(r.BorrowerName), muted(capitalize(r.BorrowerStatus)), num(len(r.Keys))})
	}
	summary := tableStyle{x: marginX, w: contentW, size: 8.5, padY: 1.5, pages: true}
	d.tableWith(summary, []column{{"Bon n°", 0.8, "L"}, {"Détenteur", 4, "L"}, {"Statut", 1.6, "L"}, {"Clés", 0.8, "R"}}, rows)

	for i, r := range receipts {
		d.receipt(r, fmt.Sprintf("Bon %d / %d — ", i+1, len(receipts)))
	}
	return d, nil
}

// receipt ajoute une page de bon de remise. prefix précède le sous-titre
// (numérotation dans une archive).
func (d *doc) receipt(r Receipt, prefix string) {
	n := len(r.Keys)
	subtitle := fmt.Sprintf("Remise de %s à %s.", plural(n, "clé", "clés"), r.BorrowerName)
	dateLabel := "Date de remise"
	if r.Reissue {
		subtitle = fmt.Sprintf("Récapitulatif des %d clés détenues par %s à ce jour.", n, r.BorrowerName)
		if n == 1 {
			subtitle = fmt.Sprintf("Récapitulatif de la clé détenue par %s à ce jour.", r.BorrowerName)
		}
		dateLabel = "Édité le"
	}
	d.page(d.s.ReceiptTitle, prefix+subtitle)

	planned := "Non précisé"
	if r.PlannedReturn != nil {
		planned = r.PlannedReturn.Format("02/01/2006")
	}
	loanType := r.LoanType
	if loanType == "" {
		loanType = "ponctuel"
	}
	d.infoCard([]field{
		{"Détenteur", r.BorrowerName},
		{"Statut", orDash(capitalize(r.BorrowerStatus))},
		{dateLabel, r.Date.Format("02/01/2006 à 15:04")},
		{"Type de prêt", capitalize(loanType)},
		{"Retour prévu", planned},
		{"Agent", orDash(r.Agent)},
	}, 3)

	engagement := d.s.engagementText(r.BorrowerName, n)
	d.font(false, 9.5, cInk)
	engagementH := float64(len(d.wrap(engagement, contentW)))*(lineH(9.5)+0.8) + 4

	d.section(fmt.Sprintf("Clés remises (%d)", n))
	// Place restante pour les clés : tout le bas de page moins l'engagement
	// (titre de section compris) et les cadres de signature.
	d.receiptKeys(r.Keys, bottomY-d.GetY()-(10+engagementH)-(signaturesH+2)-4)

	d.section("Engagement")
	d.font(false, 9.5, cInk)
	d.SetX(marginX)
	d.multi(contentW, engagement, 9.5)
	d.SetY(d.GetY() + 4)
	d.signatures(r.BorrowerName)
}

// receiptLayout est une densité possible du tableau des clés d'un bon.
type receiptLayout struct {
	cols       int     // nombre de colonnes de clés côte à côte
	size, padY float64 // police et marge verticale des lignes
}

// receiptLayouts va du plus confortable au plus dense : le premier qui fait
// tenir toutes les clés dans la place disponible est retenu, pour que le bon
// tienne sur une seule page (jusqu'à environ 70 clés).
var receiptLayouts = []receiptLayout{
	{1, 9, 2.4}, {1, 8.5, 1.6}, {2, 8.5, 1.6}, {2, 7.5, 1.1}, {3, 7.5, 1.1}, {3, 7, 0.8},
}

// receiptKeys dessine le tableau des clés d'un bon dans la hauteur avail :
// textes complets (retour à la ligne) si la place le permet, sinon la
// première densité de receiptLayouts qui fait tenir toutes les clés.
func (d *doc) receiptKeys(keys []ReceiptKey, avail float64) {
	oneCol := []column{{"N°", 1.8, "L"}, {"Description", 3, "L"}, {"Rangement", 1.7, "L"}, {"Remise le", 1.3, "R"}}
	oneColRows := func() [][]cell {
		rows := make([][]cell, 0, len(keys))
		for _, k := range keys {
			rows = append(rows, []cell{key(k.Number), txt(orDash(k.Description)), muted(orDash(k.Storage)), txt(k.LoanDate.Format("02/01/2006"))})
		}
		return rows
	}

	wrapped := defaultTable
	if rows := oneColRows(); d.tableHeight(wrapped, oneCol, rows) <= avail {
		d.tableWith(wrapped, oneCol, rows)
		d.SetY(d.GetY() + 4)
		return
	}

	layout := receiptLayouts[len(receiptLayouts)-1]
	for _, l := range receiptLayouts {
		s := tableStyle{size: l.size, padY: l.padY}
		if l.cols*int((avail-headerH)/s.rowHeight()) >= len(keys) {
			layout = l
			break
		}
	}

	if layout.cols == 1 {
		s := tableStyle{x: marginX, w: contentW, size: layout.size, padY: layout.padY, pages: true}
		d.tableWith(s, oneCol, oneColRows())
		d.SetY(d.GetY() + 4)
		return
	}

	// Plusieurs colonnes côte à côte, remplies l'une après l'autre.
	const gap = 4.0
	w := (contentW - gap*float64(layout.cols-1)) / float64(layout.cols)
	perCol := (len(keys) + layout.cols - 1) / layout.cols
	y0, yEnd := d.GetY(), d.GetY()
	for c := 0; c < layout.cols; c++ {
		start := c * perCol
		if start >= len(keys) {
			break
		}
		end := min(start+perCol, len(keys))
		rows := make([][]cell, 0, end-start)
		for _, k := range keys[start:end] {
			rows = append(rows, []cell{key(k.Number), txt(orDash(k.Description)), muted(k.LoanDate.Format("02/01/06"))})
		}
		d.SetY(y0)
		s := tableStyle{x: marginX + float64(c)*(w+gap), w: w, size: layout.size, padY: layout.padY}
		d.tableWith(s, []column{{"N°", 1.4, "L"}, {"Description", 1.8, "L"}, {"Remise", 1.4, "R"}}, rows)
		yEnd = max(yEnd, d.GetY())
	}
	d.SetY(yEnd + 4)
}

// signaturesH est la hauteur des cadres de signature.
const signaturesH = 32.0

// signatures dessine les deux cadres de signature du bon de remise.
func (d *doc) signatures(borrower string) {
	h := signaturesH
	d.ensure(h + 2)
	y := d.GetY()
	w := (contentW - 6) / 2
	for i, label := range []string{d.s.SignatureBorrower, d.s.SignatureEstablishment} {
		x := marginX + float64(i)*(w+6)
		d.draw(cLine)
		d.SetLineWidth(0.3)
		d.RoundedRect(x, y, w, h, 2.5, "1234", "D")
		d.SetXY(x+4, y+3)
		d.font(true, 9, cInk)
		d.CellFormat(w-8, 4.5, label, "", 2, "L", false, 0, "")
		d.SetX(x + 4)
		d.font(false, 8, cMuted)
		sub := "Date et signature"
		if i == 1 && strings.TrimSpace(d.s.Signatory) != "" {
			sub = d.ellipsis(d.s.Signatory, w-8) + " — date et signature"
		}
		if i == 0 {
			sub = d.ellipsis(borrower, w-8) + " — date et signature"
		}
		d.CellFormat(w-8, 4, sub, "", 0, "L", false, 0, "")
	}
	d.SetY(y + h + 2)
}

// capitalize met une majuscule initiale (« permanent » → « Permanent »).
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if size == len(s) {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(string(r)) + s[size:]
}

// ============= PLAN DE CLÉS =============

// PlanBuilding est un bâtiment du plan de clés.
type PlanBuilding struct {
	Name     string
	Accesses []PlanAccess
}

// PlanAccess est un accès avec les numéros des clés qui l'ouvrent.
type PlanAccess struct {
	Name, Floor, Category string
	Keys                  []string
}

// GenerateKeyPlan génère le plan de clés : bâtiment → accès → clés.
func GenerateKeyPlan(s Settings, buildings []PlanBuilding) ([]byte, error) {
	sorted := make([]PlanBuilding, len(buildings))
	copy(sorted, buildings)
	sort.SliceStable(sorted, func(i, j int) bool { return natsort.Less(sorted[i].Name, sorted[j].Name) })

	accesses, withoutKey := 0, 0
	keys := map[string]bool{}
	for _, b := range sorted {
		accesses += len(b.Accesses)
		for _, a := range b.Accesses {
			if len(a.Keys) == 0 {
				withoutKey++
			}
			for _, k := range a.Keys {
				keys[k] = true
			}
		}
	}

	d := newDoc(s, "Plan de clés", "Pour chaque bâtiment, les accès et les clés qui les ouvrent.")
	noKeyTone := toneSuccess
	if withoutKey > 0 {
		noKeyTone = toneWarning
	}
	d.stats([]stat{
		{"bâtiments", fmt.Sprint(len(sorted)), toneNone},
		{"accès", fmt.Sprint(accesses), toneNone},
		{"clés utilisées", fmt.Sprint(len(keys)), toneNone},
		{"accès sans clé", fmt.Sprint(withoutKey), noKeyTone},
	})

	cols := []column{{"Accès", 3, "L"}, {"Étage", 1.1, "L"}, {"Catégorie", 1.8, "L"}, {"Clés", 3, "L"}}
	for _, b := range sorted {
		d.section(fmt.Sprintf("%s — %s", b.Name, plural(len(b.Accesses), "accès", "accès")))
		if len(b.Accesses) == 0 {
			d.font(false, 9, cMuted)
			d.SetX(marginX)
			d.CellFormat(contentW, 6, "Aucun accès dans ce bâtiment.", "", 1, "L", false, 0, "")
			d.SetY(d.GetY() + 2)
			continue
		}
		list := make([]PlanAccess, len(b.Accesses))
		copy(list, b.Accesses)
		sort.SliceStable(list, func(i, j int) bool { return natsort.Less(list[i].Name, list[j].Name) })
		rows := make([][]cell, 0, len(list))
		for _, a := range list {
			k := badge("aucune clé", toneWarning)
			if len(a.Keys) > 0 {
				nums := append([]string(nil), a.Keys...)
				sort.SliceStable(nums, func(i, j int) bool { return natsort.Less(nums[i], nums[j]) })
				k = bold(strings.Join(nums, "  ·  "))
			}
			rows = append(rows, []cell{txt(a.Name), muted(orDash(a.Floor)), muted(orDash(a.Category)), k})
		}
		d.table(cols, rows)
	}
	return d.output()
}

// ============= BILAN DU STOCK =============

// StockRow est une clé du bilan du stock.
type StockRow struct {
	Number, Description, Storage      string
	Total, Reserve, Loaned, Available int
}

// availabilityTone colore une disponibilité : négative (erreur d'inventaire),
// nulle (épuisée) ou positive.
func availabilityTone(n int) tone {
	switch {
	case n < 0:
		return toneDanger
	case n == 0:
		return toneWarning
	default:
		return toneSuccess
	}
}

// GenerateStockReport génère le bilan du stock de clés.
func GenerateStockReport(s Settings, rows []StockRow) ([]byte, error) {
	copies, loaned, available, errors, empty := 0, 0, 0, 0, 0
	var inError []string
	for _, r := range rows {
		copies += r.Total
		loaned += r.Loaned
		if r.Available > 0 {
			available += r.Available
		}
		if r.Available < 0 {
			errors++
			inError = append(inError, r.Number)
		}
		if r.Available == 0 {
			empty++
		}
	}

	d := newDoc(s, "Bilan du stock de clés", "Disponible = total − réserve − exemplaires sortis.")
	errTone := toneSuccess
	if errors > 0 {
		errTone = toneDanger
	}
	d.stats([]stat{
		{"clés", fmt.Sprint(len(rows)), toneNone},
		{"exemplaires", fmt.Sprint(copies), toneNone},
		{"sortis", fmt.Sprint(loaned), tonePrimary},
		{"disponibles", fmt.Sprint(available), toneSuccess},
		{"en erreur", fmt.Sprint(errors), errTone},
	})
	if errors > 0 {
		d.note("Erreur d'inventaire, vérifier le stock : plus d'exemplaires sont sortis que le stock ne le permet pour "+
			strings.Join(inError, ", ")+". Un inventaire (onglet Inventaire) remet le stock à jour.", toneDanger)
	}
	if empty > 0 {
		d.note(fmt.Sprintf("%s plus aucun exemplaire disponible.", plural(empty, "clé n'a", "clés n'ont")), toneWarning)
	}

	d.section("Détail par clé")
	table := make([][]cell, 0, len(rows))
	for _, r := range rows {
		table = append(table, []cell{
			key(r.Number), txt(orDash(r.Description)), muted(orDash(r.Storage)),
			num(r.Total), num(r.Reserve), num(r.Loaned), badge(fmt.Sprint(r.Available), availabilityTone(r.Available)),
		})
	}
	d.table([]column{
		{"N°", 1.2, "L"}, {"Description", 2.8, "L"}, {"Rangement", 1.7, "L"},
		{"Total", 0.9, "R"}, {"Réserve", 1.1, "R"}, {"Sorties", 1, "R"}, {"Dispo", 0.9, "R"},
	}, table)
	return d.output()
}

// ============= RAPPORT D'INVENTAIRE =============

// InventoryRow est une clé dont le stock a été modifié par l'inventaire.
type InventoryRow struct {
	Number, Description    string
	Loaned, Counted        int // exemplaires prêtés, exemplaires comptés en stock
	OldTotal, NewTotal     int
	OldReserve, NewReserve int
}

// InventoryReport est le compte rendu d'un inventaire.
type InventoryReport struct {
	Agent   string
	Counted int            // nombre de clés comptées
	Changes []InventoryRow // clés dont le total ou la réserve a changé
}

// GenerateInventoryReport génère le rapport d'inventaire (écarts constatés).
func GenerateInventoryReport(s Settings, r InventoryReport) ([]byte, error) {
	plus, minus := 0, 0
	for _, c := range r.Changes {
		if diff := c.NewTotal - c.OldTotal; diff > 0 {
			plus += diff
		} else {
			minus -= diff
		}
	}
	d := newDoc(s, "Rapport d'inventaire", "Nouveau total = exemplaires comptés en stock + exemplaires prêtés.")
	d.infoCard([]field{
		{"Date", d.generated.Format("02/01/2006 à 15:04")},
		{"Réalisé par", orDash(r.Agent)},
		{"Clés comptées", fmt.Sprint(r.Counted)},
	}, 3)
	minusTone, plusTone := toneSuccess, toneNone
	if minus > 0 {
		minusTone = toneDanger
	}
	if plus > 0 {
		plusTone = toneSuccess
	}
	d.stats([]stat{
		{"clés modifiées", fmt.Sprint(len(r.Changes)), tonePrimary},
		{"exemplaires en plus", signed(plus), plusTone},
		{"exemplaires en moins", fmt.Sprintf("−%d", minus), minusTone},
	})

	if len(r.Changes) == 0 {
		d.note("Aucun écart constaté : le stock compté correspond au stock enregistré.", toneSuccess)
		return d.output()
	}

	d.section("Écarts constatés")
	rows := make([][]cell, 0, len(r.Changes))
	var reserveNotes []string
	for _, c := range r.Changes {
		diff := c.NewTotal - c.OldTotal
		t := toneNeutral
		switch {
		case diff > 0:
			t = toneSuccess
		case diff < 0:
			t = toneDanger
		}
		rows = append(rows, []cell{
			key(c.Number), txt(orDash(c.Description)), num(c.Loaned), boldNum(c.Counted),
			muted(fmt.Sprint(c.OldTotal)), boldNum(c.NewTotal), badge(signed(diff), t),
		})
		if c.NewReserve != c.OldReserve {
			reserveNotes = append(reserveNotes, fmt.Sprintf("%s : %d → %d", c.Number, c.OldReserve, c.NewReserve))
		}
	}
	d.table([]column{
		{"N°", 1.2, "L"}, {"Description", 3, "L"}, {"Prêtées", 0.9, "R"}, {"En stock", 1, "R"},
		{"Avant", 0.9, "R"}, {"Après", 0.9, "R"}, {"Écart", 0.9, "R"},
	}, rows)
	if len(reserveNotes) > 0 {
		d.note("Réserve réduite faute d'exemplaires en stock : "+strings.Join(reserveNotes, " ; ")+".", toneWarning)
	}
	return d.output()
}

// signed formate un entier avec son signe (+2, −3, 0).
func signed(n int) string {
	switch {
	case n > 0:
		return fmt.Sprintf("+%d", n)
	case n < 0:
		return fmt.Sprintf("−%d", -n)
	default:
		return "0"
	}
}
