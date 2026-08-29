package db

import (
	"database/sql"
	"time"
)

// ============= KEYS =============

const keySelectSQL = `SELECT id, number, description, quantity_total, quantity_reserve, storage_location, category, notes FROM keys`

// GetAllKeys récupère toutes les clés
func GetAllKeys() ([]Key, error) {
	rows, err := DB.Query(keySelectSQL + ` ORDER BY number`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKeys(rows)
}

// GetKeyByID récupère une clé par son ID
func GetKeyByID(id int) (*Key, error) {
	rows, err := DB.Query(keySelectSQL+` WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys, err := scanKeys(rows)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, sql.ErrNoRows
	}
	return &keys[0], nil
}

// scanKeys lit un *sql.Rows (issu de keySelectSQL) en slice de Key.
// Centralise le mapping colonnes->struct pour toutes les requêtes sur les clés.
func scanKeys(rows *sql.Rows) ([]Key, error) {
	var keys []Key
	for rows.Next() {
		var k Key
		var sl, cat, notes sql.NullString
		if err := rows.Scan(&k.ID, &k.Number, &k.Description, &k.QuantityTotal, &k.QuantityReserve, &sl, &cat, &notes); err != nil {
			return nil, err
		}
		k.StorageLocation = sl.String
		k.Category = cat.String
		k.Notes = notes.String
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// CreateKey crée une nouvelle clé
func CreateKey(k *Key, roomIDs []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`INSERT INTO keys (number, description, quantity_total, quantity_reserve, storage_location) VALUES (?, ?, ?, ?, ?)`,
		k.Number, k.Description, k.QuantityTotal, k.QuantityReserve, k.StorageLocation)
	if err != nil {
		return err
	}

	keyID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	k.ID = int(keyID)

	// Associer les salles
	for _, roomID := range roomIDs {
		_, err = tx.Exec(`INSERT INTO key_room_association (key_id, room_id) VALUES (?, ?)`, keyID, roomID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateKey met à jour une clé
func UpdateKey(k *Key, roomIDs []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE keys SET number = ?, description = ?, quantity_total = ?, quantity_reserve = ?, storage_location = ? WHERE id = ?`,
		k.Number, k.Description, k.QuantityTotal, k.QuantityReserve, k.StorageLocation, k.ID)
	if err != nil {
		return err
	}

	// Supprimer les anciennes associations
	_, err = tx.Exec(`DELETE FROM key_room_association WHERE key_id = ?`, k.ID)
	if err != nil {
		return err
	}

	// Créer les nouvelles associations
	for _, roomID := range roomIDs {
		_, err = tx.Exec(`INSERT INTO key_room_association (key_id, room_id) VALUES (?, ?)`, k.ID, roomID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateKeyQuantityTotal met à jour uniquement le stock total d'une clé, sans
// toucher aux associations clé↔salle (contrairement à UpdateKey).
func UpdateKeyQuantityTotal(keyID int, newTotal int) error {
	_, err := DB.Exec(`UPDATE keys SET quantity_total = ? WHERE id = ?`, newTotal, keyID)
	return err
}

// DeleteKey supprime une clé
func DeleteKey(id int) error {
	_, err := DB.Exec(`DELETE FROM keys WHERE id = ?`, id)
	return err
}

// GetRoomsForKey récupère les salles associées à une clé
func GetRoomsForKey(keyID int) ([]Room, error) {
	rows, err := DB.Query(`
		SELECT r.id, r.name, r.type, r.building_id 
		FROM rooms r
		INNER JOIN key_room_association kra ON r.id = kra.room_id
		WHERE kra.key_id = ?
		ORDER BY r.name`, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		var roomType sql.NullString
		err := rows.Scan(&r.ID, &r.Name, &roomType, &r.BuildingID)
		if err != nil {
			return nil, err
		}
		if roomType.Valid {
			r.Type = roomType.String
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

// ============= BORROWERS =============

const borrowerSelectSQL = `SELECT id, name, email, status, phone FROM borrowers`

// GetAllBorrowers récupère tous les emprunteurs
func GetAllBorrowers() ([]Borrower, error) {
	rows, err := DB.Query(borrowerSelectSQL + ` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBorrowers(rows)
}

// GetBorrowerByID récupère un emprunteur par son ID
func GetBorrowerByID(id int) (*Borrower, error) {
	rows, err := DB.Query(borrowerSelectSQL+` WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	borrowers, err := scanBorrowers(rows)
	if err != nil {
		return nil, err
	}
	if len(borrowers) == 0 {
		return nil, sql.ErrNoRows
	}
	return &borrowers[0], nil
}

// scanBorrowers lit un *sql.Rows (issu de borrowerSelectSQL) en slice de Borrower.
func scanBorrowers(rows *sql.Rows) ([]Borrower, error) {
	var borrowers []Borrower
	for rows.Next() {
		var b Borrower
		var email, status, phone sql.NullString
		if err := rows.Scan(&b.ID, &b.Name, &email, &status, &phone); err != nil {
			return nil, err
		}
		b.Email = email.String
		b.Status = status.String
		b.Phone = phone.String
		borrowers = append(borrowers, b)
	}
	return borrowers, rows.Err()
}

// CreateBorrower crée un nouvel emprunteur
func CreateBorrower(b *Borrower) error {
	result, err := DB.Exec(`INSERT INTO borrowers (name, email) VALUES (?, ?)`, b.Name, b.Email)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = int(id)
	return nil
}

// UpdateBorrower met à jour un emprunteur
func UpdateBorrower(b *Borrower) error {
	_, err := DB.Exec(`UPDATE borrowers SET name = ?, email = ? WHERE id = ?`, b.Name, b.Email, b.ID)
	return err
}

// DeleteBorrower supprime un emprunteur
func DeleteBorrower(id int) error {
	_, err := DB.Exec(`DELETE FROM borrowers WHERE id = ?`, id)
	return err
}

// ============= BUILDINGS =============

// GetAllBuildings récupère tous les bâtiments
func GetAllBuildings() ([]Building, error) {
	rows, err := DB.Query(`SELECT id, name FROM buildings ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []Building
	for rows.Next() {
		var b Building
		err := rows.Scan(&b.ID, &b.Name)
		if err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, rows.Err()
}

// GetBuildingByID récupère un bâtiment par son ID
func GetBuildingByID(id int) (*Building, error) {
	var b Building
	err := DB.QueryRow(`SELECT id, name FROM buildings WHERE id = ?`, id).Scan(&b.ID, &b.Name)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBuilding crée un nouveau bâtiment
func CreateBuilding(b *Building) error {
	result, err := DB.Exec(`INSERT INTO buildings (name) VALUES (?)`, b.Name)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = int(id)
	return nil
}

// UpdateBuilding met à jour un bâtiment
func UpdateBuilding(b *Building) error {
	_, err := DB.Exec(`UPDATE buildings SET name = ? WHERE id = ?`, b.Name, b.ID)
	return err
}

// DeleteBuilding supprime un bâtiment
func DeleteBuilding(id int) error {
	_, err := DB.Exec(`DELETE FROM buildings WHERE id = ?`, id)
	return err
}

// ============= ACCÈS (rooms) =============

const roomSelectSQL = `SELECT id, name, type, building_id, floor, category, notes FROM rooms`

// GetAllAccesses récupère tous les accès (alias V3 de GetAllRooms)
func GetAllAccesses() ([]Room, error) {
	rows, err := DB.Query(roomSelectSQL + ` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRooms(rows)
}

// GetRoomByID récupère un accès par son ID
func GetRoomByID(id int) (*Room, error) {
	rows, err := DB.Query(roomSelectSQL+` WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms, err := scanRooms(rows)
	if err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		return nil, sql.ErrNoRows
	}
	return &rooms[0], nil
}

// GetAccessesByBuilding récupère les accès d'un bâtiment
func GetAccessesByBuilding(buildingID int) ([]Room, error) {
	rows, err := DB.Query(roomSelectSQL+` WHERE building_id = ? ORDER BY name`, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRooms(rows)
}

// GetRoomsByBuildingID est un alias de GetAccessesByBuilding
func GetRoomsByBuildingID(buildingID int) ([]Room, error) { return GetAccessesByBuilding(buildingID) }

// GetKeysForAccess récupère les clés associées à un accès
func GetKeysForAccess(accessID int) ([]Key, error) {
	rows, err := DB.Query(`
		SELECT k.id, k.number, k.description, k.quantity_total, k.quantity_reserve, k.storage_location, k.category, k.notes
		FROM keys k
		INNER JOIN key_room_association kra ON k.id = kra.key_id
		WHERE kra.room_id = ?
		ORDER BY k.number`, accessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKeys(rows)
}

// CreateRoom crée un nouvel accès avec tous les champs V3
func CreateRoom(r *Room) error {
	result, err := DB.Exec(
		`INSERT INTO rooms (name, type, building_id, floor, category, notes) VALUES (?, ?, ?, ?, ?, ?)`,
		r.Name, r.Type, r.BuildingID, r.Floor, r.Category, r.Notes,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = int(id)
	return nil
}

// UpdateRoom met à jour un accès avec tous les champs V3
func UpdateRoom(r *Room) error {
	_, err := DB.Exec(
		`UPDATE rooms SET name = ?, type = ?, building_id = ?, floor = ?, category = ?, notes = ? WHERE id = ?`,
		r.Name, r.Type, r.BuildingID, r.Floor, r.Category, r.Notes, r.ID,
	)
	return err
}

// DeleteRoom supprime un accès
func DeleteRoom(id int) error {
	_, err := DB.Exec(`DELETE FROM rooms WHERE id = ?`, id)
	return err
}

// scanRooms lit un *sql.Rows (issu de roomSelectSQL) en slice de Room (accès).
func scanRooms(rows *sql.Rows) ([]Room, error) {
	var rooms []Room
	for rows.Next() {
		var r Room
		var roomType, floor, category, notes sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &roomType, &r.BuildingID, &floor, &category, &notes); err != nil {
			return nil, err
		}
		r.Type = roomType.String
		r.Floor = floor.String
		r.Category = category.String
		r.Notes = notes.String
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

// ============= LOANS =============

const loanSelectSQL = `
	SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
	       l.planned_return_date, l.loan_type, l.returned_condition, l.created_by,
	       k.number, k.description, b.name, b.email
	FROM loans l
	INNER JOIN keys k ON l.key_id = k.id
	INNER JOIN borrowers b ON l.borrower_id = b.id`

// GetAllActiveLoans récupère tous les emprunts actifs
func GetAllActiveLoans() ([]LoanWithDetails, error) {
	rows, err := DB.Query(loanSelectSQL + ` WHERE l.return_date IS NULL ORDER BY b.name, l.loan_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoansWithDetails(rows)
}

// GetActiveLoansByKeyID récupère les emprunts actifs pour une clé
func GetActiveLoansByKeyID(keyID int) ([]LoanWithDetails, error) {
	rows, err := DB.Query(loanSelectSQL+` WHERE l.key_id = ? AND l.return_date IS NULL ORDER BY l.loan_date`, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoansWithDetails(rows)
}

// GetActiveLoansByBorrowerID récupère les emprunts actifs pour un emprunteur
func GetActiveLoansByBorrowerID(borrowerID int) ([]LoanWithDetails, error) {
	rows, err := DB.Query(loanSelectSQL+` WHERE l.borrower_id = ? AND l.return_date IS NULL ORDER BY l.loan_date`, borrowerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoansWithDetails(rows)
}

// ReturnLoan marque un emprunt comme retourné, avec état constaté optionnel.
func ReturnLoan(loanID int) error {
	return ReturnLoanWithCondition(loanID, "")
}

// ReturnLoanWithCondition marque un emprunt comme retourné et enregistre l'état constaté.
func ReturnLoanWithCondition(loanID int, condition string) error {
	_, err := DB.Exec(
		`UPDATE loans SET return_date = ?, returned_condition = ? WHERE id = ?`,
		time.Now(), condition, loanID,
	)
	return err
}

// GetActiveLoanCount récupère le nombre d'emprunts actifs pour une clé
func GetActiveLoanCount(keyID int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM loans WHERE key_id = ? AND return_date IS NULL`, keyID).Scan(&count)
	return count, err
}

// GetKeysWithAvailability récupère toutes les clés avec disponibilité en une seule requête SQL.
func GetKeysWithAvailability() ([]KeyWithAvailability, error) {
	rows, err := DB.Query(`
		SELECT
			k.id, k.number, k.description, k.quantity_total, k.quantity_reserve,
			k.storage_location, k.category, k.notes,
			COUNT(l.id) AS loaned_count,
			(k.quantity_total - k.quantity_reserve - COUNT(l.id)) AS available_count,
			GROUP_CONCAT(b.name, ', ') AS borrower_names
		FROM keys k
		LEFT JOIN loans l ON l.key_id = k.id AND l.return_date IS NULL
		LEFT JOIN borrowers b ON l.borrower_id = b.id
		GROUP BY k.id
		ORDER BY k.number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []KeyWithAvailability
	for rows.Next() {
		var kwa KeyWithAvailability
		var storageLocation, category, notes, borrowerNames sql.NullString
		err := rows.Scan(
			&kwa.ID, &kwa.Number, &kwa.Description, &kwa.QuantityTotal, &kwa.QuantityReserve,
			&storageLocation, &category, &notes,
			&kwa.LoanedCount, &kwa.AvailableCount, &borrowerNames,
		)
		if err != nil {
			return nil, err
		}
		kwa.StorageLocation = storageLocation.String
		kwa.Category = category.String
		kwa.Notes = notes.String
		if borrowerNames.Valid && borrowerNames.String != "" {
			kwa.BorrowerNames = splitNames(borrowerNames.String)
		}
		result = append(result, kwa)
	}
	return result, rows.Err()
}

// GetAvailableKeys récupère les clés qui ont au moins un exemplaire disponible.
func GetAvailableKeys() ([]Key, error) {
	rows, err := DB.Query(`
		SELECT k.id, k.number, k.description, k.quantity_total, k.quantity_reserve, k.storage_location, k.category, k.notes
		FROM keys k
		WHERE (k.quantity_total - k.quantity_reserve -
			(SELECT COUNT(*) FROM loans l WHERE l.key_id = k.id AND l.return_date IS NULL)) > 0
		ORDER BY k.number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []Key
	for rows.Next() {
		var k Key
		var sl, cat, notes sql.NullString
		if err := rows.Scan(&k.ID, &k.Number, &k.Description, &k.QuantityTotal, &k.QuantityReserve, &sl, &cat, &notes); err != nil {
			return nil, err
		}
		k.StorageLocation = sl.String
		k.Category = cat.String
		k.Notes = notes.String
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetOverdueLoans récupère les prêts dont la date de retour prévue est dépassée.
func GetOverdueLoans() ([]LoanWithDetails, error) {
	rows, err := DB.Query(`
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       l.planned_return_date, l.loan_type, l.returned_condition, l.created_by,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE l.return_date IS NULL
		  AND l.planned_return_date IS NOT NULL
		  AND l.planned_return_date < datetime('now')
		ORDER BY l.planned_return_date
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoansWithDetails(rows)
}

// GetLoanHistory récupère l'historique filtré des prêts.
func GetLoanHistory(filters LoanFilters) ([]LoanWithDetails, error) {
	query := `
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       l.planned_return_date, l.loan_type, l.returned_condition, l.created_by,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE 1=1`
	var args []interface{}

	if filters.BorrowerID != nil {
		query += " AND l.borrower_id = ?"
		args = append(args, *filters.BorrowerID)
	}
	if filters.KeyID != nil {
		query += " AND l.key_id = ?"
		args = append(args, *filters.KeyID)
	}
	if filters.DateFrom != nil {
		query += " AND l.loan_date >= ?"
		args = append(args, *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query += " AND l.loan_date <= ?"
		args = append(args, *filters.DateTo)
	}
	switch filters.Status {
	case "active":
		query += " AND l.return_date IS NULL"
	case "returned":
		query += " AND l.return_date IS NOT NULL"
	case "overdue":
		query += " AND l.return_date IS NULL AND l.planned_return_date < datetime('now')"
	}
	query += " ORDER BY l.loan_date DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoansWithDetails(rows)
}

// GetBorrowersWithRedundantAccesses retourne les détenteurs ayant des accès couverts par plusieurs de leurs clés.
func GetBorrowersWithRedundantAccesses() ([]RedundancyReport, error) {
	borrowers, err := GetAllBorrowers()
	if err != nil {
		return nil, err
	}
	var reports []RedundancyReport
	for _, b := range borrowers {
		loans, err := GetActiveLoansByBorrowerID(b.ID)
		if err != nil {
			return nil, err
		}
		if len(loans) < 2 {
			continue
		}
		accessCount := make(map[int]int)
		accessMap := make(map[int]Room)
		var keys []Key
		for _, loan := range loans {
			key, err := GetKeyByID(loan.KeyID)
			if err != nil {
				continue
			}
			keys = append(keys, *key)
			rooms, err := GetRoomsForKey(loan.KeyID)
			if err != nil {
				continue
			}
			for _, r := range rooms {
				accessCount[r.ID]++
				accessMap[r.ID] = r
			}
		}
		var redundant []Room
		for id, count := range accessCount {
			if count > 1 {
				redundant = append(redundant, accessMap[id])
			}
		}
		if len(redundant) > 0 {
			reports = append(reports, RedundancyReport{
				Borrower:          b,
				Keys:              keys,
				RedundantAccesses: redundant,
			})
		}
	}
	return reports, nil
}

// scanLoansWithDetails est un helper commun pour scanner les lignes de prêts avec détails.
func scanLoansWithDetails(rows *sql.Rows) ([]LoanWithDetails, error) {
	var loans []LoanWithDetails
	for rows.Next() {
		var l LoanWithDetails
		var returnDate, plannedReturnDate sql.NullTime
		var email, loanType, condition, createdBy sql.NullString
		err := rows.Scan(
			&l.ID, &l.KeyID, &l.BorrowerID, &l.LoanDate, &returnDate,
			&plannedReturnDate, &loanType, &condition, &createdBy,
			&l.KeyNumber, &l.KeyDescription, &l.BorrowerName, &email,
		)
		if err != nil {
			return nil, err
		}
		if returnDate.Valid {
			l.ReturnDate = &returnDate.Time
		}
		if plannedReturnDate.Valid {
			l.PlannedReturnDate = &plannedReturnDate.Time
		}
		l.LoanType = loanType.String
		l.ReturnedCondition = condition.String
		l.CreatedBy = createdBy.String
		l.BorrowerEmail = email.String
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

// splitNames découpe une chaîne GROUP_CONCAT en slice de noms.
func splitNames(s string) []string {
	var names []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || (i+1 < len(s) && s[i] == ',' && s[i+1] == ' ') {
			names = append(names, s[start:i])
			i++ // sauter l'espace
			start = i + 1
		}
	}
	return names
}

// GetBorrowerActiveLoanCount récupère le nombre d'emprunts actifs pour un emprunteur
func GetBorrowerActiveLoanCount(borrowerID int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM loans WHERE borrower_id = ? AND return_date IS NULL`, borrowerID).Scan(&count)
	return count, err
}

// GetKeyPlanData récupère les données pour le plan de clés
func GetKeyPlanData() (map[int]Building, error) {
	buildings, err := GetAllBuildings()
	if err != nil {
		return nil, err
	}

	buildingMap := make(map[int]Building)
	for _, building := range buildings {
		// Récupérer les salles du bâtiment
		rooms, err := GetRoomsByBuildingID(building.ID)
		if err != nil {
			return nil, err
		}

		// Pour chaque salle, récupérer les clés
		for i := range rooms {
			keys, err := GetKeysForRoom(rooms[i].ID)
			if err != nil {
				return nil, err
			}
			rooms[i].Keys = keys
		}

		building.Rooms = rooms
		buildingMap[building.ID] = building
	}

	return buildingMap, nil
}

// GetKeysForRoom récupère les clés associées à une salle
func GetKeysForRoom(roomID int) ([]Key, error) {
	rows, err := DB.Query(`
		SELECT k.id, k.number, k.description, k.quantity_total, k.quantity_reserve, k.storage_location
		FROM keys k
		INNER JOIN key_room_association kra ON k.id = kra.key_id
		WHERE kra.room_id = ?
		ORDER BY k.number`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []Key
	for rows.Next() {
		var k Key
		var storageLocation sql.NullString
		err := rows.Scan(&k.ID, &k.Number, &k.Description, &k.QuantityTotal, &k.QuantityReserve, &storageLocation)
		if err != nil {
			return nil, err
		}
		if storageLocation.Valid {
			k.StorageLocation = storageLocation.String
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// CreateMultipleLoans crée plusieurs emprunts pour un emprunteur.
// La disponibilité est vérifiée en amont par l'interface — pas de requête imbriquée
// dans la transaction pour éviter le deadlock avec SetMaxOpenConns(1).
func CreateMultipleLoans(keyIDs []int, borrowerID int) ([]int, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	loanIDs := make([]int, 0, len(keyIDs))
	for _, keyID := range keyIDs {
		res, err := tx.Exec(`INSERT INTO loans (key_id, borrower_id, loan_date) VALUES (?, ?, ?)`,
			keyID, borrowerID, now)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		loanIDs = append(loanIDs, int(id))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return loanIDs, nil
}

// CheckInventoryAnomalies retourne les clés dont le nombre d'exemplaires
// prêtés dépasse le stock utilisable (total - réserve). Le résultat est vide
// quand l'inventaire est cohérent. Ce contrôle est purement informatif : il
// n'empêche jamais un prêt (le sur-prêt est un comportement métier assumé).
func CheckInventoryAnomalies() ([]InventoryAnomaly, error) {
	rows, err := DB.Query(`
		SELECT k.id, k.number, k.quantity_total, k.quantity_reserve, COUNT(l.id) AS loaned
		FROM keys k
		LEFT JOIN loans l ON l.key_id = k.id AND l.return_date IS NULL
		GROUP BY k.id
		HAVING (k.quantity_total - k.quantity_reserve - COUNT(l.id)) < 0
		ORDER BY k.number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []InventoryAnomaly
	for rows.Next() {
		var a InventoryAnomaly
		if err := rows.Scan(&a.KeyID, &a.KeyNumber, &a.Total, &a.Reserve, &a.Loaned); err != nil {
			return nil, err
		}
		a.Available = a.Total - a.Reserve - a.Loaned
		anomalies = append(anomalies, a)
	}
	return anomalies, rows.Err()
}

// GetLoanDuration calcule la durée d'un emprunt en jours
func GetLoanDuration(loanDate time.Time) float64 {
	return time.Since(loanDate).Hours() / 24
}
