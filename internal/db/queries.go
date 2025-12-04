package db

import (
	"database/sql"
	"fmt"
	"time"
)

// ============= KEYS =============

// GetAllKeys récupère toutes les clés
func GetAllKeys() ([]Key, error) {
	rows, err := DB.Query(`SELECT id, number, description, quantity_total, quantity_reserve, storage_location FROM keys ORDER BY number`)
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

// GetKeyByID récupère une clé par son ID
func GetKeyByID(id int) (*Key, error) {
	var k Key
	var storageLocation sql.NullString
	err := DB.QueryRow(`SELECT id, number, description, quantity_total, quantity_reserve, storage_location FROM keys WHERE id = ?`, id).
		Scan(&k.ID, &k.Number, &k.Description, &k.QuantityTotal, &k.QuantityReserve, &storageLocation)
	if err != nil {
		return nil, err
	}
	if storageLocation.Valid {
		k.StorageLocation = storageLocation.String
	}
	return &k, nil
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

// GetAllBorrowers récupère tous les emprunteurs
func GetAllBorrowers() ([]Borrower, error) {
	rows, err := DB.Query(`SELECT id, name, email FROM borrowers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrowers []Borrower
	for rows.Next() {
		var b Borrower
		var email sql.NullString
		err := rows.Scan(&b.ID, &b.Name, &email)
		if err != nil {
			return nil, err
		}
		if email.Valid {
			b.Email = email.String
		}
		borrowers = append(borrowers, b)
	}
	return borrowers, rows.Err()
}

// GetBorrowerByID récupère un emprunteur par son ID
func GetBorrowerByID(id int) (*Borrower, error) {
	var b Borrower
	var email sql.NullString
	err := DB.QueryRow(`SELECT id, name, email FROM borrowers WHERE id = ?`, id).
		Scan(&b.ID, &b.Name, &email)
	if err != nil {
		return nil, err
	}
	if email.Valid {
		b.Email = email.String
	}
	return &b, nil
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

// ============= ROOMS =============

// GetAllRooms récupère toutes les salles
func GetAllRooms() ([]Room, error) {
	rows, err := DB.Query(`SELECT id, name, type, building_id FROM rooms ORDER BY name`)
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

// GetRoomsByBuildingID récupère les salles d'un bâtiment
func GetRoomsByBuildingID(buildingID int) ([]Room, error) {
	rows, err := DB.Query(`SELECT id, name, type, building_id FROM rooms WHERE building_id = ? ORDER BY name`, buildingID)
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

// CreateRoom crée une nouvelle salle
func CreateRoom(r *Room) error {
	result, err := DB.Exec(`INSERT INTO rooms (name, type, building_id) VALUES (?, ?, ?)`, r.Name, r.Type, r.BuildingID)
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

// UpdateRoom met à jour une salle
func UpdateRoom(r *Room) error {
	_, err := DB.Exec(`UPDATE rooms SET name = ?, type = ?, building_id = ? WHERE id = ?`, r.Name, r.Type, r.BuildingID, r.ID)
	return err
}

// DeleteRoom supprime une salle
func DeleteRoom(id int) error {
	_, err := DB.Exec(`DELETE FROM rooms WHERE id = ?`, id)
	return err
}

// ============= LOANS =============

// GetAllActiveLoans récupère tous les emprunts actifs
func GetAllActiveLoans() ([]LoanWithDetails, error) {
	rows, err := DB.Query(`
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE l.return_date IS NULL
		ORDER BY b.name, l.loan_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []LoanWithDetails
	for rows.Next() {
		var l LoanWithDetails
		var returnDate sql.NullTime
		var email sql.NullString
		err := rows.Scan(&l.ID, &l.KeyID, &l.BorrowerID, &l.LoanDate, &returnDate,
			&l.KeyNumber, &l.KeyDescription, &l.BorrowerName, &email)
		if err != nil {
			return nil, err
		}
		if returnDate.Valid {
			l.ReturnDate = &returnDate.Time
		}
		if email.Valid {
			l.BorrowerEmail = email.String
		}
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

// GetActiveLoansByKeyID récupère les emprunts actifs pour une clé
func GetActiveLoansByKeyID(keyID int) ([]LoanWithDetails, error) {
	rows, err := DB.Query(`
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE l.key_id = ? AND l.return_date IS NULL
		ORDER BY l.loan_date`, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []LoanWithDetails
	for rows.Next() {
		var l LoanWithDetails
		var returnDate sql.NullTime
		var email sql.NullString
		err := rows.Scan(&l.ID, &l.KeyID, &l.BorrowerID, &l.LoanDate, &returnDate,
			&l.KeyNumber, &l.KeyDescription, &l.BorrowerName, &email)
		if err != nil {
			return nil, err
		}
		if returnDate.Valid {
			l.ReturnDate = &returnDate.Time
		}
		if email.Valid {
			l.BorrowerEmail = email.String
		}
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

// GetActiveLoansByBorrowerID récupère les emprunts actifs pour un emprunteur
func GetActiveLoansByBorrowerID(borrowerID int) ([]LoanWithDetails, error) {
	rows, err := DB.Query(`
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE l.borrower_id = ? AND l.return_date IS NULL
		ORDER BY l.loan_date`, borrowerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []LoanWithDetails
	for rows.Next() {
		var l LoanWithDetails
		var returnDate sql.NullTime
		var email sql.NullString
		err := rows.Scan(&l.ID, &l.KeyID, &l.BorrowerID, &l.LoanDate, &returnDate,
			&l.KeyNumber, &l.KeyDescription, &l.BorrowerName, &email)
		if err != nil {
			return nil, err
		}
		if returnDate.Valid {
			l.ReturnDate = &returnDate.Time
		}
		if email.Valid {
			l.BorrowerEmail = email.String
		}
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

// GetLoanByID récupère un emprunt par son ID
func GetLoanByID(id int) (*LoanWithDetails, error) {
	var l LoanWithDetails
	var returnDate sql.NullTime
	var email sql.NullString
	err := DB.QueryRow(`
		SELECT l.id, l.key_id, l.borrower_id, l.loan_date, l.return_date,
		       k.number, k.description, b.name, b.email
		FROM loans l
		INNER JOIN keys k ON l.key_id = k.id
		INNER JOIN borrowers b ON l.borrower_id = b.id
		WHERE l.id = ?`, id).
		Scan(&l.ID, &l.KeyID, &l.BorrowerID, &l.LoanDate, &returnDate,
			&l.KeyNumber, &l.KeyDescription, &l.BorrowerName, &email)
	if err != nil {
		return nil, err
	}
	if returnDate.Valid {
		l.ReturnDate = &returnDate.Time
	}
	if email.Valid {
		l.BorrowerEmail = email.String
	}
	return &l, nil
}

// CreateLoan crée un nouvel emprunt
func CreateLoan(keyID, borrowerID int) error {
	_, err := DB.Exec(`INSERT INTO loans (key_id, borrower_id, loan_date) VALUES (?, ?, ?)`,
		keyID, borrowerID, time.Now())
	return err
}

// ReturnLoan marque un emprunt comme retourné
func ReturnLoan(loanID int) error {
	_, err := DB.Exec(`UPDATE loans SET return_date = ? WHERE id = ?`, time.Now(), loanID)
	return err
}

// GetActiveLoanCount récupère le nombre d'emprunts actifs pour une clé
func GetActiveLoanCount(keyID int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM loans WHERE key_id = ? AND return_date IS NULL`, keyID).Scan(&count)
	return count, err
}

// GetKeysWithAvailability récupère toutes les clés avec leurs informations de disponibilité
func GetKeysWithAvailability() ([]KeyWithAvailability, error) {
	keys, err := GetAllKeys()
	if err != nil {
		return nil, err
	}

	var result []KeyWithAvailability
	for _, key := range keys {
		kwa := KeyWithAvailability{Key: key}

		// Compter les emprunts actifs
		count, err := GetActiveLoanCount(key.ID)
		if err != nil {
			return nil, err
		}
		kwa.LoanedCount = count

		// Calculer la disponibilité
		usable := key.QuantityTotal - key.QuantityReserve
		kwa.AvailableCount = usable - count

		// Récupérer les noms des emprunteurs
		if count > 0 {
			loans, err := GetActiveLoansByKeyID(key.ID)
			if err != nil {
				return nil, err
			}
			for _, loan := range loans {
				kwa.BorrowerNames = append(kwa.BorrowerNames, loan.BorrowerName)
			}
		}

		result = append(result, kwa)
	}

	return result, nil
}

// GetAvailableKeys récupère les clés disponibles pour un emprunt
func GetAvailableKeys() ([]Key, error) {
	keys, err := GetAllKeys()
	if err != nil {
		return nil, err
	}

	var available []Key
	for _, key := range keys {
		count, err := GetActiveLoanCount(key.ID)
		if err != nil {
			return nil, err
		}
		usable := key.QuantityTotal - key.QuantityReserve
		if usable > count {
			available = append(available, key)
		}
	}

	return available, nil
}

// GetBorrowerActiveLoanCount récupère le nombre d'emprunts actifs pour un emprunteur
func GetBorrowerActiveLoanCount(borrowerID int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM loans WHERE borrower_id = ? AND return_date IS NULL`, borrowerID).Scan(&count)
	return count, err
}

// GetKeyActiveLoanCount récupère le nombre d'emprunts actifs pour une clé
func GetKeyActiveLoanCount(keyID int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM loans WHERE key_id = ? AND return_date IS NULL`, keyID).Scan(&count)
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

// CheckKeyAvailability vérifie si une clé est disponible pour un emprunt
func CheckKeyAvailability(keyID int) (bool, error) {
	key, err := GetKeyByID(keyID)
	if err != nil {
		return false, err
	}

	count, err := GetActiveLoanCount(keyID)
	if err != nil {
		return false, err
	}

	usable := key.QuantityTotal - key.QuantityReserve
	return usable > count, nil
}

// CreateMultipleLoans crée plusieurs emprunts pour un emprunteur
func CreateMultipleLoans(keyIDs []int, borrowerID int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, keyID := range keyIDs {
		// Vérifier la disponibilité
		available, err := CheckKeyAvailability(keyID)
		if err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("la clé %d n'est pas disponible", keyID)
		}

		// Créer l'emprunt
		_, err = tx.Exec(`INSERT INTO loans (key_id, borrower_id, loan_date) VALUES (?, ?, ?)`,
			keyID, borrowerID, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
