package app

import (
	"fmt"
	"sort"
	"strings"

	"clefs/internal/db"
	"clefs/internal/natsort"
)

// catalog.go : référentiels — clés, accès, bâtiments, détenteurs.

// Catégories et statuts proposés à la saisie (mêmes valeurs qu'en V3).
var (
	keyCategories    = []string{"simple", "trousseau", "badge", "passe"}
	borrowerStatuses = []string{"permanent", "contractuel", "intervenant", "entreprise"}
)

// ============= CLÉS =============

// ListKeys retourne toutes les clés avec leur disponibilité et leurs accès.
func (a *App) ListKeys() ([]KeyDTO, error) {
	defer a.read()()
	return listKeys()
}

// listKeys est la version sans verrou de ListKeys, réutilisée en interne.
func listKeys() ([]KeyDTO, error) {
	keys, err := db.GetKeysWithAvailability()
	if err != nil {
		return nil, err
	}
	accessMap, err := db.GetKeyAccessMap()
	if err != nil {
		return nil, err
	}
	out := make([]KeyDTO, 0, len(keys))
	for _, k := range keys {
		out = append(out, toKeyDTO(k, accessMap[k.ID]))
	}
	sort.SliceStable(out, func(i, j int) bool { return natsort.Less(out[i].Number, out[j].Number) })
	return out, nil
}

// keyByID retrouve une clé (avec disponibilité) dans une liste.
func keyByID(keys []KeyDTO, id int) (KeyDTO, bool) {
	for _, k := range keys {
		if k.ID == id {
			return k, true
		}
	}
	return KeyDTO{}, false
}

// GetKeyDetail retourne la fiche complète d'une clé.
func (a *App) GetKeyDetail(id int) (KeyDetail, error) {
	defer a.read()()
	keys, err := listKeys()
	if err != nil {
		return KeyDetail{}, err
	}
	k, ok := keyByID(keys, id)
	if !ok {
		return KeyDetail{}, fmt.Errorf("clé introuvable")
	}
	accesses, err := accessesByIDs(k.AccessIDs)
	if err != nil {
		return KeyDetail{}, err
	}
	active, err := db.GetActiveLoansByKeyID(id)
	if err != nil {
		return KeyDetail{}, err
	}
	history, err := db.GetLoanHistory(db.LoanFilters{KeyID: &id})
	if err != nil {
		return KeyDetail{}, err
	}
	return KeyDetail{Key: k, Accesses: accesses, ActiveLoans: toLoanDTOs(active), History: toLoanDTOs(history)}, nil
}

// validateKeyInput applique les règles de saisie de la V3.
func validateKeyInput(in KeyInput) error {
	if strings.TrimSpace(in.Number) == "" {
		return fmt.Errorf("le numéro de la clé est obligatoire")
	}
	// Un total de 0 est permis : toutes les clés perdues ou cassées, la fiche
	// est conservée pour l'historique.
	if in.Total < 0 {
		return fmt.Errorf("la quantité totale ne peut pas être négative")
	}
	if in.Reserve < 0 {
		return fmt.Errorf("la réserve ne peut pas être négative")
	}
	if in.Reserve > in.Total {
		// Sinon le stock utilisable serait négatif dès la saisie, sans aucun
		// prêt : c'est une erreur de frappe, pas un sur-prêt volontaire.
		return fmt.Errorf("la réserve (%d) ne peut pas dépasser la quantité totale (%d)", in.Reserve, in.Total)
	}
	return nil
}

// SaveKey crée ou modifie une clé et retourne son ID.
func (a *App) SaveKey(in KeyInput) (int, error) {
	defer a.read()()
	if err := validateKeyInput(in); err != nil {
		return 0, err
	}
	k := &db.Key{
		ID: in.ID, Number: strings.TrimSpace(in.Number), Description: strings.TrimSpace(in.Description),
		Category: in.Category, StorageLocation: strings.TrimSpace(in.StorageLocation), Notes: strings.TrimSpace(in.Notes),
		QuantityTotal: in.Total, QuantityReserve: in.Reserve,
	}
	unique := fmt.Sprintf("une clé porte déjà le numéro « %s »", k.Number)
	if in.ID == 0 {
		if err := db.CreateKey(k, in.AccessIDs); err != nil {
			return 0, friendlyDBError(err, unique)
		}
		return k.ID, nil
	}
	return k.ID, friendlyDBError(db.UpdateKey(k, in.AccessIDs), unique)
}

// SetKeyTotal ajuste rapidement le stock total d'une clé (règle V3 : pas en
// dessous des exemplaires sortis ni de la réserve).
func (a *App) SetKeyTotal(id int, total int) error {
	defer a.read()()
	keys, err := listKeys()
	if err != nil {
		return err
	}
	k, ok := keyByID(keys, id)
	if !ok {
		return fmt.Errorf("clé introuvable")
	}
	if total < 0 {
		return fmt.Errorf("la quantité totale ne peut pas être négative")
	}
	if total < k.Loaned || total < k.Reserve {
		return fmt.Errorf("le stock total ne peut pas être inférieur aux clés sorties (%d) ni à la réserve (%d)", k.Loaned, k.Reserve)
	}
	return db.UpdateKeyQuantityTotal(id, total)
}

// DeleteKey supprime une clé. Refusé tant qu'un exemplaire est prêté ;
// l'historique des prêts rendus de cette clé est supprimé avec elle.
func (a *App) DeleteKey(id int) error {
	defer a.read()()
	n, err := db.GetActiveLoanCount(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("cette clé est actuellement prêtée (%d exemplaire(s)) : enregistrez d'abord les retours", n)
	}
	return db.DeleteKey(id)
}

// GetKeyCategories retourne les catégories de clé proposées.
func (a *App) GetKeyCategories() []string { return keyCategories }

// ============= ACCÈS =============

// ListAccesses retourne tous les accès avec leur bâtiment et leurs clés.
func (a *App) ListAccesses() ([]AccessDTO, error) {
	defer a.read()()
	return listAccesses()
}

// listAccesses est la version sans verrou de ListAccesses.
func listAccesses() ([]AccessDTO, error) {
	rooms, err := db.GetAllAccesses()
	if err != nil {
		return nil, err
	}
	names, err := buildingNames()
	if err != nil {
		return nil, err
	}
	accessMap, err := db.GetKeyAccessMap()
	if err != nil {
		return nil, err
	}
	keysByAccess := make(map[int][]int)
	for keyID, roomIDs := range accessMap {
		for _, roomID := range roomIDs {
			keysByAccess[roomID] = append(keysByAccess[roomID], keyID)
		}
	}
	out := make([]AccessDTO, 0, len(rooms))
	for _, r := range rooms {
		ids := keysByAccess[r.ID]
		sort.Ints(ids)
		out = append(out, toAccessDTO(r, names[r.BuildingID], ids))
	}
	sort.SliceStable(out, func(i, j int) bool { return natsort.Less(out[i].Name, out[j].Name) })
	return out, nil
}

// accessesByIDs retourne les accès demandés, triés par nom.
func accessesByIDs(ids []int) ([]AccessDTO, error) {
	all, err := listAccesses()
	if err != nil {
		return nil, err
	}
	wanted := make(map[int]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	out := []AccessDTO{}
	for _, acc := range all {
		if wanted[acc.ID] {
			out = append(out, acc)
		}
	}
	return out, nil
}

// buildingNames retourne la table ID de bâtiment → nom.
func buildingNames() (map[int]string, error) {
	buildings, err := db.GetAllBuildings()
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(buildings))
	for _, b := range buildings {
		m[b.ID] = b.Name
	}
	return m, nil
}

// GetAccessDetail retourne la fiche complète d'un accès : clés qui l'ouvrent
// et détenteurs qui peuvent l'ouvrir actuellement.
func (a *App) GetAccessDetail(id int) (AccessDetail, error) {
	defer a.read()()
	all, err := listAccesses()
	if err != nil {
		return AccessDetail{}, err
	}
	var acc *AccessDTO
	for i := range all {
		if all[i].ID == id {
			acc = &all[i]
			break
		}
	}
	if acc == nil {
		return AccessDetail{}, fmt.Errorf("accès introuvable")
	}
	keys, err := listKeys()
	if err != nil {
		return AccessDetail{}, err
	}
	detail := AccessDetail{Access: *acc, Keys: []KeyDTO{}, Holders: []HolderDTO{}}
	for _, keyID := range acc.KeyIDs {
		k, ok := keyByID(keys, keyID)
		if !ok {
			continue
		}
		detail.Keys = append(detail.Keys, k)
		loans, err := db.GetActiveLoansByKeyID(keyID)
		if err != nil {
			return AccessDetail{}, err
		}
		for _, l := range loans {
			detail.Holders = append(detail.Holders, HolderDTO{BorrowerID: l.BorrowerID, BorrowerName: l.BorrowerName, KeyNumber: l.KeyNumber})
		}
	}
	sort.Slice(detail.Holders, func(i, j int) bool { return detail.Holders[i].BorrowerName < detail.Holders[j].BorrowerName })
	return detail, nil
}

// SaveAccess crée ou modifie un accès et retourne son ID.
func (a *App) SaveAccess(in AccessInput) (int, error) {
	defer a.read()()
	if strings.TrimSpace(in.Name) == "" {
		return 0, fmt.Errorf("la désignation de l'accès est obligatoire")
	}
	if in.BuildingID == 0 {
		return 0, fmt.Errorf("choisissez le bâtiment de l'accès")
	}
	r := &db.Room{
		ID: in.ID, Name: strings.TrimSpace(in.Name), Type: strings.TrimSpace(in.Type), BuildingID: in.BuildingID,
		Floor: strings.TrimSpace(in.Floor), Category: strings.TrimSpace(in.Category), Notes: strings.TrimSpace(in.Notes),
	}
	if in.ID == 0 {
		err := db.CreateRoom(r)
		return r.ID, err
	}
	return r.ID, db.UpdateRoom(r)
}

// DeleteAccess supprime un accès (les clés restent, elles n'ouvrent simplement plus cet accès).
func (a *App) DeleteAccess(id int) error {
	defer a.read()()
	return db.DeleteRoom(id)
}

// ============= BÂTIMENTS =============

// ListBuildings retourne les bâtiments avec leur nombre d'accès.
func (a *App) ListBuildings() ([]BuildingDTO, error) {
	defer a.read()()
	buildings, err := db.GetAllBuildings()
	if err != nil {
		return nil, err
	}
	rooms, err := db.GetAllAccesses()
	if err != nil {
		return nil, err
	}
	count := make(map[int]int)
	for _, r := range rooms {
		count[r.BuildingID]++
	}
	out := make([]BuildingDTO, 0, len(buildings))
	for _, b := range buildings {
		out = append(out, BuildingDTO{ID: b.ID, Name: b.Name, AccessCount: count[b.ID]})
	}
	return out, nil
}

// SaveBuilding crée ou renomme un bâtiment et retourne son ID.
func (a *App) SaveBuilding(id int, name string) (int, error) {
	defer a.read()()
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("le nom du bâtiment est obligatoire")
	}
	b := &db.Building{ID: id, Name: name}
	unique := fmt.Sprintf("un bâtiment s'appelle déjà « %s »", name)
	if id == 0 {
		return b.ID, friendlyDBError(db.CreateBuilding(b), unique)
	}
	return id, friendlyDBError(db.UpdateBuilding(b), unique)
}

// DeleteBuilding supprime un bâtiment et ses accès (cascade du schéma).
func (a *App) DeleteBuilding(id int) error {
	defer a.read()()
	return db.DeleteBuilding(id)
}

// ============= DÉTENTEURS =============

// ListBorrowers retourne les détenteurs avec leur nombre de clés détenues.
func (a *App) ListBorrowers() ([]BorrowerDTO, error) {
	defer a.read()()
	return listBorrowers()
}

// listBorrowers est la version sans verrou de ListBorrowers.
func listBorrowers() ([]BorrowerDTO, error) {
	borrowers, err := db.GetAllBorrowers()
	if err != nil {
		return nil, err
	}
	loans, err := db.GetAllActiveLoans()
	if err != nil {
		return nil, err
	}
	count := make(map[int]int)
	for _, l := range loans {
		count[l.BorrowerID]++
	}
	out := make([]BorrowerDTO, 0, len(borrowers))
	for _, b := range borrowers {
		out = append(out, toBorrowerDTO(b, count[b.ID]))
	}
	return out, nil
}

// GetBorrowerDetail retourne la fiche complète d'un détenteur.
func (a *App) GetBorrowerDetail(id int) (BorrowerDetail, error) {
	defer a.read()()
	return borrowerDetail(id)
}

// borrowerDetail est la version sans verrou de GetBorrowerDetail.
func borrowerDetail(id int) (BorrowerDetail, error) {
	b, err := db.GetBorrowerByID(id)
	if err != nil {
		return BorrowerDetail{}, fmt.Errorf("détenteur introuvable")
	}
	bwk, err := borrowerWithKeys(*b)
	if err != nil {
		return BorrowerDetail{}, err
	}
	history, err := db.GetLoanHistory(db.LoanFilters{BorrowerID: &id})
	if err != nil {
		return BorrowerDetail{}, err
	}
	names, err := buildingNames()
	if err != nil {
		return BorrowerDetail{}, err
	}
	conv := func(rooms []db.Room) []AccessDTO {
		out := make([]AccessDTO, 0, len(rooms))
		for _, r := range rooms {
			out = append(out, toAccessDTO(r, names[r.BuildingID], nil))
		}
		return out
	}
	return BorrowerDetail{
		Borrower:        toBorrowerDTO(*b, len(bwk.ActiveLoans)),
		ActiveLoans:     toLoanDTOs(bwk.ActiveLoans),
		CoveredAccesses: conv(bwk.CoveredAccesses),
		Redundancies:    conv(bwk.Redundancies),
		History:         toLoanDTOs(history),
	}, nil
}

// validateBorrowerInput applique les règles de saisie d'un détenteur.
func validateBorrowerInput(in BorrowerInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("le nom du détenteur est obligatoire")
	}
	if in.Status != "" && !contains(borrowerStatuses, in.Status) {
		return fmt.Errorf("statut inconnu : %s", in.Status)
	}
	return nil
}

// SaveBorrower crée ou modifie un détenteur et retourne son ID.
func (a *App) SaveBorrower(in BorrowerInput) (int, error) {
	defer a.read()()
	if err := validateBorrowerInput(in); err != nil {
		return 0, err
	}
	b := &db.Borrower{
		ID: in.ID, Name: strings.TrimSpace(in.Name), Email: strings.TrimSpace(in.Email),
		Status: in.Status, Phone: strings.TrimSpace(in.Phone),
	}
	if in.ID == 0 {
		err := db.CreateBorrower(b)
		return b.ID, err
	}
	return b.ID, db.UpdateBorrower(b)
}

// DeleteBorrower supprime un détenteur. Refusé tant qu'il détient des clés ;
// son historique de prêts est supprimé avec lui.
func (a *App) DeleteBorrower(id int) error {
	defer a.read()()
	n, err := db.GetBorrowerActiveLoanCount(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("ce détenteur détient encore %d clé(s) : enregistrez d'abord les retours", n)
	}
	return db.DeleteBorrower(id)
}

// GetBorrowerStatuses retourne les statuts de détenteur proposés.
func (a *App) GetBorrowerStatuses() []string { return borrowerStatuses }

// contains indique si s figure dans la liste.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
