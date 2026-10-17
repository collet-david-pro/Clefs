package app

import (
	"fmt"

	"clefs/internal/db"
	"clefs/internal/pdf"
)

// inventory.go : mode inventaire. On saisit, pour chaque clé comptée, le
// nombre d'exemplaires physiquement en stock (réserve comprise) ; le total
// devient « en stock + prêtés ». C'est la mise à jour rapide du stock après
// une clé retrouvée, cassée ou perdue, sans gestion de perte détaillée.

// InventoryItem est le comptage d'une clé.
type InventoryItem struct {
	KeyID   int `json:"keyId"`
	InStock int `json:"inStock"`
}

// InventoryResult est le compte rendu d'un inventaire validé.
type InventoryResult struct {
	Counted     int    `json:"counted"`     // clés comptées
	Changed     int    `json:"changed"`     // clés dont le stock a changé
	Delta       int    `json:"delta"`       // variation totale du nombre d'exemplaires
	ReportPath  string `json:"reportPath"`  // rapport d'inventaire PDF
	ReportError string `json:"reportError"` // erreur de génération du rapport, le cas échéant
}

// ApplyInventory enregistre un inventaire et génère son rapport PDF.
func (a *App) ApplyInventory(items []InventoryItem) (InventoryResult, error) {
	defer a.read()()
	if len(items) == 0 {
		return InventoryResult{}, fmt.Errorf("aucune clé comptée")
	}
	before, err := listKeys()
	if err != nil {
		return InventoryResult{}, err
	}
	seen := map[int]bool{}
	counts := make([]db.InventoryCount, 0, len(items))
	for _, it := range items {
		k, ok := keyByID(before, it.KeyID)
		if !ok {
			return InventoryResult{}, fmt.Errorf("clé introuvable (n° %d)", it.KeyID)
		}
		if it.InStock < 0 {
			return InventoryResult{}, fmt.Errorf("clé %s : le nombre en stock ne peut pas être négatif", k.Number)
		}
		if seen[it.KeyID] {
			return InventoryResult{}, fmt.Errorf("clé %s comptée deux fois", k.Number)
		}
		seen[it.KeyID] = true
		counts = append(counts, db.InventoryCount{KeyID: it.KeyID, InStock: it.InStock})
	}
	if err := db.ApplyInventory(counts); err != nil {
		return InventoryResult{}, fmt.Errorf("enregistrement de l'inventaire impossible : %w", err)
	}

	after, err := listKeys()
	if err != nil {
		return InventoryResult{}, err
	}
	res := InventoryResult{Counted: len(items)}
	report := pdf.InventoryReport{Agent: currentAgent(), Counted: len(items)}
	for _, it := range items {
		old, _ := keyByID(before, it.KeyID)
		cur, _ := keyByID(after, it.KeyID)
		if old.Total == cur.Total && old.Reserve == cur.Reserve {
			continue
		}
		res.Changed++
		res.Delta += cur.Total - old.Total
		report.Changes = append(report.Changes, pdf.InventoryRow{
			Number: cur.Number, Description: cur.Description, Loaned: cur.Loaned, Counted: it.InStock,
			OldTotal: old.Total, NewTotal: cur.Total, OldReserve: old.Reserve, NewReserve: cur.Reserve,
		})
	}
	data, err := pdf.GenerateInventoryReport(pdfSettings(), report)
	if path, err := a.savePDF("inventaire", data, err); err != nil {
		res.ReportError = err.Error()
	} else {
		res.ReportPath = path
	}
	return res, nil
}
