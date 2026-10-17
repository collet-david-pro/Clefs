package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"clefs/internal/business"
	"clefs/internal/db"
	"clefs/internal/natsort"
	"clefs/internal/pdf"
)

// loans.go : prêt par besoin, retours, bons de remise, historique et vues
// transversales (tableau de bord, qui a quoi, plan de clés, redondances).

// ============= PRÊT PAR BESOIN =============

// SuggestKeys calcule le trousseau minimal couvrant les accès demandés
// (Set Cover glouton de internal/business), en respectant les clés retirées
// ou ajoutées à la main, et signale les redondances avec les clés que le
// détenteur possède déjà.
func (a *App) SuggestKeys(req SuggestRequest) (Suggestion, error) {
	defer a.read()()
	out := Suggestion{Keys: []SuggestedKey{}, Uncoverable: []AccessDTO{}, Redundancies: []string{}}
	if len(req.AccessIDs) == 0 && len(req.ManualKeyIDs) == 0 {
		return out, nil
	}

	excluded := toIntSet(req.ExcludedKeyIDs)
	manual := toIntSet(req.ManualKeyIDs)
	requested := toIntSet(req.AccessIDs)

	candidates, err := business.BuildAvailableKeysForAccesses(req.AccessIDs)
	if err != nil {
		return out, err
	}
	var filtered []db.KeyWithCoverage
	for _, c := range candidates {
		if !excluded[c.ID] {
			filtered = append(filtered, c)
		}
	}
	result := business.SuggestKeys(req.AccessIDs, filtered)

	keys, err := listKeys()
	if err != nil {
		return out, err
	}
	accesses, err := listAccesses()
	if err != nil {
		return out, err
	}
	accessName := make(map[int]string, len(accesses))
	for _, acc := range accesses {
		accessName[acc.ID] = acc.Name
	}

	chosen := make(map[int]*SuggestedKey)
	add := func(id int, suggested bool) {
		if s, ok := chosen[id]; ok {
			s.Suggested = s.Suggested || suggested
			return
		}
		k, ok := keyByID(keys, id)
		if !ok {
			return
		}
		covers := []string{}
		for _, accID := range k.AccessIDs {
			if requested[accID] {
				covers = append(covers, accessName[accID])
			}
		}
		sort.Strings(covers)
		chosen[id] = &SuggestedKey{Key: k, Suggested: suggested, Manual: manual[id], Covers: covers}
	}
	for _, k := range result.SelectedKeys {
		add(k.ID, true)
	}
	for id := range manual {
		add(id, false)
	}
	for _, s := range chosen {
		out.Keys = append(out.Keys, *s)
	}
	sort.Slice(out.Keys, func(i, j int) bool { return out.Keys[i].Key.Number < out.Keys[j].Key.Number })

	// Les accès signalés « non couvrables » le restent seulement si aucune clé
	// ajoutée à la main ne les ouvre.
	for _, id := range result.UncoverableIDs {
		covered := false
		for _, s := range out.Keys {
			for _, accID := range s.Key.AccessIDs {
				if accID == id {
					covered = true
				}
			}
		}
		if covered {
			continue
		}
		for _, acc := range accesses {
			if acc.ID == id {
				out.Uncoverable = append(out.Uncoverable, acc)
			}
		}
	}

	// Redondances avec les clés déjà détenues par ce détenteur.
	if req.BorrowerID != 0 && len(out.Keys) > 0 {
		existing, err := db.GetActiveLoansByBorrowerID(req.BorrowerID)
		if err != nil {
			return out, err
		}
		if len(existing) > 0 {
			var all []db.Key
			accessesByKey := make(map[int][]db.Room)
			addKey := func(id int) {
				k, ok := keyByID(keys, id)
				if !ok {
					return
				}
				all = append(all, db.Key{ID: k.ID, Number: k.Number})
				rooms := make([]db.Room, 0, len(k.AccessIDs))
				for _, accID := range k.AccessIDs {
					rooms = append(rooms, db.Room{ID: accID, Name: accessName[accID]})
				}
				accessesByKey[k.ID] = rooms
			}
			for _, l := range existing {
				addKey(l.KeyID)
			}
			for _, s := range out.Keys {
				addKey(s.Key.ID)
			}
			for _, r := range business.DetectRedundancies(all, accessesByKey) {
				out.Redundancies = append(out.Redundancies, r.Name)
			}
		}
	}
	return out, nil
}

// toIntSet convertit une liste d'IDs en ensemble.
func toIntSet(ids []int) map[int]bool {
	s := make(map[int]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

// parseDay lit une date AAAA-MM-JJ (heure locale). Vide → nil.
func parseDay(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, fmt.Errorf("date invalide : %s", s)
	}
	return &t, nil
}

// CreateLoan enregistre la remise des clés, génère le bon de remise PDF et
// signale les clés passées en sur-prêt (sans jamais bloquer le prêt).
func (a *App) CreateLoan(req CreateLoanRequest) (LoanResult, error) {
	defer a.read()()
	if req.BorrowerID == 0 {
		return LoanResult{}, fmt.Errorf("choisissez le détenteur")
	}
	if len(req.KeyIDs) == 0 {
		return LoanResult{}, fmt.Errorf("le trousseau est vide")
	}
	borrower, err := db.GetBorrowerByID(req.BorrowerID)
	if err != nil {
		return LoanResult{}, fmt.Errorf("détenteur introuvable")
	}
	planned, err := parseDay(req.PlannedReturn)
	if err != nil {
		return LoanResult{}, err
	}
	if planned != nil {
		// Le prêt n'est en retard qu'à la fin du jour de retour prévu.
		end := planned.Add(24*time.Hour - time.Second)
		planned = &end
		if end.Before(time.Now()) {
			return LoanResult{}, fmt.Errorf("la date de retour prévue est déjà passée")
		}
	}
	loanType := req.LoanType
	if loanType != "permanent" {
		loanType = "ponctuel"
	}

	agent := currentAgent()
	loanIDs, err := db.CreateMultipleLoans(req.KeyIDs, req.BorrowerID, db.LoanOptions{
		PlannedReturnDate: planned, LoanType: loanType, CreatedBy: agent,
	})
	if err != nil {
		return LoanResult{}, fmt.Errorf("enregistrement du prêt impossible : %w", err)
	}

	res := LoanResult{Count: len(loanIDs), BorrowerName: borrower.Name, Anomalies: []AnomalyDTO{}}
	path, err := a.generateReceipt(req.BorrowerID, loanIDs)
	if err != nil {
		res.ReceiptError = err.Error()
	} else {
		res.ReceiptPath = path
	}

	if anomalies, err := db.CheckInventoryAnomalies(); err == nil {
		loaned := toIntSet(req.KeyIDs)
		for _, an := range toAnomalyDTOs(anomalies) {
			if loaned[an.KeyID] {
				res.Anomalies = append(res.Anomalies, an)
			}
		}
	}
	return res, nil
}

// generateReceipt génère et enregistre le bon de remise PDF d'un détenteur
// pour les prêts loanIDs, ou pour tous ses prêts actifs si loanIDs est vide
// (réédition d'un bon récapitulatif à jour). Retourne le chemin absolu du PDF.
func (a *App) generateReceipt(borrowerID int, loanIDs []int) (string, error) {
	borrower, err := db.GetBorrowerByID(borrowerID)
	if err != nil {
		return "", fmt.Errorf("détenteur introuvable : %w", err)
	}
	loans, err := db.GetActiveLoansByBorrowerID(borrowerID)
	if err != nil {
		return "", fmt.Errorf("lecture des prêts : %w", err)
	}
	reissue := len(loanIDs) == 0
	if !reissue {
		wanted := toIntSet(loanIDs)
		filtered := loans[:0]
		for _, l := range loans {
			if wanted[l.ID] {
				filtered = append(filtered, l)
			}
		}
		loans = filtered
	}
	if len(loans) == 0 {
		return "", fmt.Errorf("aucun prêt actif à documenter")
	}
	keys, err := listKeys()
	if err != nil {
		return "", err
	}
	data, err := pdf.GenerateReceipt(pdfSettings(), buildReceipt(*borrower, loans, keys, reissue))
	return a.savePDF("bon_remise_"+safeFileName(borrower.Name), data, err)
}

// buildReceipt prépare les données d'un bon de remise. Pour une réédition,
// la date et l'agent sont ceux de l'édition ; sinon ceux de la remise.
func buildReceipt(b db.Borrower, loans []db.LoanWithDetails, keys []KeyDTO, reissue bool) pdf.Receipt {
	r := pdf.Receipt{
		BorrowerName: b.Name, BorrowerStatus: borrowerStatusOrDefault(b.Status),
		Date: loans[0].LoanDate, Agent: loans[0].CreatedBy, Reissue: reissue,
	}
	if reissue {
		r.Date, r.Agent = time.Now(), currentAgent()
	}
	sorted := append([]db.LoanWithDetails(nil), loans...)
	sort.SliceStable(sorted, func(i, j int) bool { return natsort.Less(sorted[i].KeyNumber, sorted[j].KeyNumber) })
	types := map[string]bool{}
	for _, l := range sorted {
		k, _ := keyByID(keys, l.KeyID)
		r.Keys = append(r.Keys, pdf.ReceiptKey{Number: l.KeyNumber, Description: l.KeyDescription, Storage: k.StorageLocation, LoanDate: l.LoanDate})
		lt := l.LoanType
		if lt == "" {
			lt = "ponctuel"
		}
		types[lt] = true
		// Retour prévu le plus proche parmi les prêts du bon.
		if p := l.PlannedReturnDate; p != nil && (r.PlannedReturn == nil || p.Before(*r.PlannedReturn)) {
			r.PlannedReturn = p
		}
	}
	r.LoanType = "ponctuel et permanent"
	if len(types) == 1 {
		for t := range types {
			r.LoanType = t
		}
	}
	return r
}

// ExportAllReceipts édite en un seul PDF les bons récapitulatifs de tous les
// détenteurs qui ont des clés (une page par détenteur, sommaire en tête),
// pour l'impression ou l'archivage.
func (a *App) ExportAllReceipts() (string, error) {
	defer a.read()()
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return "", err
	}
	keys, err := listKeys()
	if err != nil {
		return "", err
	}
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return "", err
	}
	byBorrower := map[int][]db.LoanWithDetails{}
	for _, l := range loans {
		byBorrower[l.BorrowerID] = append(byBorrower[l.BorrowerID], l)
	}
	sort.SliceStable(borrowers, func(i, j int) bool { return natsort.Less(borrowers[i].Name, borrowers[j].Name) })
	var receipts []pdf.Receipt
	for _, b := range borrowers {
		if bl := byBorrower[b.ID]; len(bl) > 0 {
			receipts = append(receipts, buildReceipt(b, bl, keys, true))
		}
	}
	data, err := pdf.GenerateReceiptArchive(pdfSettings(), receipts)
	return a.savePDF("archive_bons_remise", data, err)
}

// ReissueReceipt régénère le bon de remise complet et à jour d'un détenteur
// (toutes ses clés actuelles) et retourne le chemin du PDF.
func (a *App) ReissueReceipt(borrowerID int) (string, error) {
	defer a.read()()
	return a.generateReceipt(borrowerID, nil)
}

// ============= RETOURS ET LISTES =============

// ListActiveLoans retourne les prêts en cours.
func (a *App) ListActiveLoans() ([]LoanDTO, error) {
	defer a.read()()
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return nil, err
	}
	return toLoanDTOs(loans), nil
}

// ReturnLoans enregistre le retour d'un ou plusieurs prêts, avec l'état
// constaté (facultatif, commun à tous).
func (a *App) ReturnLoans(loanIDs []int, condition string) error {
	defer a.read()()
	if len(loanIDs) == 0 {
		return fmt.Errorf("aucun prêt sélectionné")
	}
	for _, id := range loanIDs {
		if err := db.ReturnLoanWithCondition(id, strings.TrimSpace(condition)); err != nil {
			return fmt.Errorf("retour impossible : %w", err)
		}
	}
	return nil
}

// toLoanFilters convertit le filtre de l'interface en filtre de requête.
func toLoanFilters(f HistoryFilter) (db.LoanFilters, error) {
	var out db.LoanFilters
	if f.BorrowerID != 0 {
		id := f.BorrowerID
		out.BorrowerID = &id
	}
	if f.KeyID != 0 {
		id := f.KeyID
		out.KeyID = &id
	}
	from, err := parseDay(f.From)
	if err != nil {
		return out, err
	}
	out.DateFrom = from
	to, err := parseDay(f.To)
	if err != nil {
		return out, err
	}
	if to != nil {
		end := to.Add(24*time.Hour - time.Second) // jour de fin inclus
		out.DateTo = &end
	}
	switch f.Status {
	case "active", "returned", "overdue":
		out.Status = f.Status
	}
	return out, nil
}

// GetHistory retourne l'historique filtré des prêts (plus récents d'abord).
func (a *App) GetHistory(f HistoryFilter) ([]LoanDTO, error) {
	defer a.read()()
	filters, err := toLoanFilters(f)
	if err != nil {
		return nil, err
	}
	loans, err := db.GetLoanHistory(filters)
	if err != nil {
		return nil, err
	}
	return toLoanDTOs(loans), nil
}

// ============= VUES TRANSVERSALES =============

// GetDashboard retourne les indicateurs et alertes du tableau de bord.
func (a *App) GetDashboard() (Dashboard, error) {
	defer a.read()()
	d := Dashboard{}
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		return d, err
	}
	d.KeyTypes = len(keys)
	for _, k := range keys {
		d.KeyCopies += k.QuantityTotal
		d.LoanedCopies += k.LoanedCount
		if k.AvailableCount > 0 {
			d.AvailableCopies += k.AvailableCount
		}
	}
	borrowers, err := listBorrowers()
	if err != nil {
		return d, err
	}
	d.Borrowers = len(borrowers)
	for _, b := range borrowers {
		if b.ActiveLoans > 0 {
			d.BorrowersWithKeys++
		}
	}
	rooms, err := db.GetAllAccesses()
	if err != nil {
		return d, err
	}
	d.Accesses = len(rooms)

	overdue, err := db.GetOverdueLoans()
	if err != nil {
		return d, err
	}
	d.Overdue = toLoanDTOs(overdue)
	anomalies, err := db.CheckInventoryAnomalies()
	if err != nil {
		return d, err
	}
	d.Anomalies = toAnomalyDTOs(anomalies)
	if d.Redundancies, err = redundancies(); err != nil {
		return d, err
	}
	history, err := db.GetLoanHistory(db.LoanFilters{})
	if err != nil {
		return d, err
	}
	if len(history) > 8 {
		history = history[:8]
	}
	d.RecentLoans = toLoanDTOs(history)
	return d, nil
}

// GetRedundancies retourne les détenteurs dont plusieurs clés ouvrent les mêmes accès.
func (a *App) GetRedundancies() ([]RedundancyDTO, error) {
	defer a.read()()
	return redundancies()
}

// redundancies est la version sans verrou de GetRedundancies.
func redundancies() ([]RedundancyDTO, error) {
	reports, err := db.GetBorrowersWithRedundantAccesses()
	if err != nil {
		return nil, err
	}
	out := make([]RedundancyDTO, 0, len(reports))
	for _, r := range reports {
		dto := RedundancyDTO{BorrowerID: r.Borrower.ID, BorrowerName: r.Borrower.Name, KeyNumbers: []string{}, Accesses: []string{}}
		for _, k := range r.Keys {
			dto.KeyNumbers = append(dto.KeyNumbers, k.Number)
		}
		for _, acc := range r.RedundantAccesses {
			dto.Accesses = append(dto.Accesses, acc.Name)
		}
		sort.Strings(dto.Accesses)
		out = append(out, dto)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].BorrowerName < out[j].BorrowerName })
	return out, nil
}

// borrowerWithKeys construit le détail des clés et accès d'un détenteur.
func borrowerWithKeys(b db.Borrower) (db.BorrowerWithKeys, error) {
	loans, err := db.GetActiveLoansByBorrowerID(b.ID)
	if err != nil {
		return db.BorrowerWithKeys{}, err
	}
	return business.BuildBorrowerWithKeys(b, loans)
}

// GetWhoHasWhat retourne, pour chaque détenteur ayant des clés, ses clés et
// les accès qu'elles lui ouvrent.
func (a *App) GetWhoHasWhat() ([]WhoHasWhat, error) {
	defer a.read()()
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return nil, err
	}
	out := []WhoHasWhat{}
	for _, b := range borrowers {
		bwk, err := borrowerWithKeys(b)
		if err != nil {
			return nil, err
		}
		if len(bwk.ActiveLoans) == 0 {
			continue
		}
		row := WhoHasWhat{
			Borrower: toBorrowerDTO(b, len(bwk.ActiveLoans)), Loans: toLoanDTOs(bwk.ActiveLoans),
			Accesses: []string{}, Redundancies: []string{},
		}
		for _, r := range bwk.CoveredAccesses {
			row.Accesses = append(row.Accesses, r.Name)
		}
		for _, r := range bwk.Redundancies {
			row.Redundancies = append(row.Redundancies, r.Name)
		}
		out = append(out, row)
	}
	return out, nil
}

// GetKeyPlan retourne le plan de clés : bâtiments → accès → clés.
func (a *App) GetKeyPlan() ([]PlanBuilding, error) {
	defer a.read()()
	return keyPlan()
}

// keyPlan est la version sans verrou de GetKeyPlan.
func keyPlan() ([]PlanBuilding, error) {
	buildings, err := db.GetAllBuildings()
	if err != nil {
		return nil, err
	}
	accesses, err := listAccesses()
	if err != nil {
		return nil, err
	}
	keys, err := listKeys()
	if err != nil {
		return nil, err
	}
	out := make([]PlanBuilding, 0, len(buildings))
	for _, b := range buildings {
		pb := PlanBuilding{ID: b.ID, Name: b.Name, Accesses: []PlanAccess{}}
		for _, acc := range accesses {
			if acc.BuildingID != b.ID {
				continue
			}
			pa := PlanAccess{Access: acc, Keys: []KeyDTO{}}
			for _, id := range acc.KeyIDs {
				if k, ok := keyByID(keys, id); ok {
					pa.Keys = append(pa.Keys, k)
				}
			}
			sort.SliceStable(pa.Keys, func(i, j int) bool { return natsort.Less(pa.Keys[i].Number, pa.Keys[j].Number) })
			pb.Accesses = append(pb.Accesses, pa)
		}
		out = append(out, pb)
	}
	return out, nil
}
