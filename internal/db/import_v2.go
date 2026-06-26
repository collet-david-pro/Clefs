package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// ImportSummary résume le résultat d'un import V2.
type ImportSummary struct {
	Buildings    int
	Rooms        int
	Keys         int
	Associations int
	Borrowers    int
	Loans        int
}

// ValidateV2Schema ouvre la base source et vérifie que les tables V2 attendues sont présentes.
func ValidateV2Schema(dbPath string) error {
	src, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier: %w", err)
	}
	defer src.Close()

	expected := []string{"buildings", "rooms", "keys", "borrowers", "loans", "key_room_association"}
	for _, table := range expected {
		var name string
		err := src.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			return fmt.Errorf("table manquante dans la base V2: %q", table)
		}
	}
	return nil
}

// ImportFromV2 importe les données d'une base V2 vers la base V3 courante.
// Une sauvegarde de sécurité est créée avant l'import.
// Les nouvelles colonnes V3 (floor, category, status, etc.) restent à leur valeur DEFAULT.
func ImportFromV2(v2DBPath string, currentDBPath string) (*ImportSummary, error) {
	if err := ValidateV2Schema(v2DBPath); err != nil {
		return nil, fmt.Errorf("ce fichier ne semble pas être une base V2 valide: %w", err)
	}

	// Sauvegarde de sécurité avant toute modification
	if err := CreateBackupDirectory(currentDBPath); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire de sauvegarde: %w", err)
	}
	backupPath := GetDefaultBackupPath(currentDBPath)
	if err := BackupDatabase(currentDBPath, backupPath); err != nil {
		return nil, fmt.Errorf("impossible de sauvegarder la base actuelle: %w", err)
	}
	log.Printf("Sauvegarde de sécurité créée: %s", backupPath)

	// Ouvrir la base V2 en lecture seule
	src, err := sql.Open("sqlite", "file:"+v2DBPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir la base V2: %w", err)
	}
	defer src.Close()

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("impossible de démarrer la transaction: %w", err)
	}
	defer tx.Rollback()

	summary := &ImportSummary{}

	// buildings — identique V2/V3
	if summary.Buildings, err = importV2Buildings(tx, src); err != nil {
		return nil, err
	}
	// rooms — V2 n'a pas floor/category/notes → NULL
	if summary.Rooms, err = importV2Rooms(tx, src); err != nil {
		return nil, err
	}
	// keys — V2 n'a pas category/notes → DEFAULT 'simple' / NULL
	if summary.Keys, err = importV2Keys(tx, src); err != nil {
		return nil, err
	}
	// key_room_association — identique V2/V3
	if summary.Associations, err = importV2Associations(tx, src); err != nil {
		return nil, err
	}
	// borrowers — V2 n'a pas status/phone → DEFAULT 'permanent' / NULL
	if summary.Borrowers, err = importV2Borrowers(tx, src); err != nil {
		return nil, err
	}
	// loans — V2 n'a pas planned_return_date/loan_type/returned_condition/created_by
	if summary.Loans, err = importV2Loans(tx, src); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erreur lors de la validation: %w", err)
	}

	log.Printf("Import V2 réussi: %d bâtiments, %d salles, %d clés, %d associations, %d emprunteurs, %d prêts",
		summary.Buildings, summary.Rooms, summary.Keys, summary.Associations, summary.Borrowers, summary.Loans)
	return summary, nil
}

// Les fonctions importV2* lisent une table de la base V2 (src) et la recopient
// dans la transaction d.import (tx). Elles retournent le nombre de lignes importées.

// importV2Buildings importe la table buildings.
func importV2Buildings(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT id, name FROM buildings ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("lecture bâtiments V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return 0, fmt.Errorf("scan bâtiment: %w", err)
		}
		if _, err := tx.Exec(`INSERT OR IGNORE INTO buildings (id, name) VALUES (?, ?)`, id, name); err != nil {
			return 0, fmt.Errorf("insertion bâtiment %q: %w", name, err)
		}
		count++
	}
	return count, rows.Err()
}

// importV2Rooms importe la table rooms (salles/accès).
func importV2Rooms(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT id, name, type, building_id FROM rooms ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("lecture salles V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int
		var name, roomType sql.NullString
		var buildingID sql.NullInt64
		if err := rows.Scan(&id, &name, &roomType, &buildingID); err != nil {
			return 0, fmt.Errorf("scan salle: %w", err)
		}
		// floor, category, notes restent NULL (colonnes V3 absentes en V2)
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO rooms (id, name, type, building_id) VALUES (?, ?, ?, ?)`,
			id, name.String, roomType.String, buildingID.Int64,
		); err != nil {
			return 0, fmt.Errorf("insertion salle %q: %w", name.String, err)
		}
		count++
	}
	return count, rows.Err()
}

// importV2Keys importe la table keys.
func importV2Keys(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT id, number, description, quantity_total, quantity_reserve, storage_location FROM keys ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("lecture clés V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int
		var number string
		var description, storageLocation sql.NullString
		var total, reserve sql.NullInt64
		if err := rows.Scan(&id, &number, &description, &total, &reserve, &storageLocation); err != nil {
			return 0, fmt.Errorf("scan clé: %w", err)
		}
		// category DEFAULT 'simple', notes NULL
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO keys (id, number, description, quantity_total, quantity_reserve, storage_location) VALUES (?, ?, ?, ?, ?, ?)`,
			id, number, description.String, total.Int64, reserve.Int64, storageLocation.String,
		); err != nil {
			return 0, fmt.Errorf("insertion clé %q: %w", number, err)
		}
		count++
	}
	return count, rows.Err()
}

// importV2Associations importe la table de liaison clé<->salle.
func importV2Associations(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT key_id, room_id FROM key_room_association`)
	if err != nil {
		return 0, fmt.Errorf("lecture associations V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var keyID, roomID int
		if err := rows.Scan(&keyID, &roomID); err != nil {
			return 0, fmt.Errorf("scan association: %w", err)
		}
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO key_room_association (key_id, room_id) VALUES (?, ?)`, keyID, roomID,
		); err != nil {
			return 0, fmt.Errorf("insertion association %d-%d: %w", keyID, roomID, err)
		}
		count++
	}
	return count, rows.Err()
}

// importV2Borrowers importe la table borrowers (détenteurs).
func importV2Borrowers(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT id, name, email FROM borrowers ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("lecture emprunteurs V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int
		var name string
		var email sql.NullString
		if err := rows.Scan(&id, &name, &email); err != nil {
			return 0, fmt.Errorf("scan emprunteur: %w", err)
		}
		// status DEFAULT 'permanent', phone NULL
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO borrowers (id, name, email) VALUES (?, ?, ?)`,
			id, name, email.String,
		); err != nil {
			return 0, fmt.Errorf("insertion emprunteur %q: %w", name, err)
		}
		count++
	}
	return count, rows.Err()
}

// importV2Loans importe la table loans (emprunts).
func importV2Loans(tx *sql.Tx, src *sql.DB) (int, error) {
	rows, err := src.Query(`SELECT id, key_id, borrower_id, loan_date, return_date FROM loans ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("lecture prêts V2: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id, keyID, borrowerID int
		var loanDate, returnDate sql.NullString
		if err := rows.Scan(&id, &keyID, &borrowerID, &loanDate, &returnDate); err != nil {
			return 0, fmt.Errorf("scan prêt: %w", err)
		}
		var returnDateVal interface{}
		if returnDate.Valid && returnDate.String != "" {
			returnDateVal = returnDate.String
		}
		// planned_return_date, returned_condition NULL ; created_by = 'Import V2'
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO loans (id, key_id, borrower_id, loan_date, return_date, created_by) VALUES (?, ?, ?, ?, ?, 'Import V2')`,
			id, keyID, borrowerID, loanDate.String, returnDateVal,
		); err != nil {
			return 0, fmt.Errorf("insertion prêt %d: %w", id, err)
		}
		count++
	}
	return count, rows.Err()
}
