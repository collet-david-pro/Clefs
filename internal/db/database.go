package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initialise la connexion à la base de données SQLite
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture de la base de données: %w", err)
	}

	// Tester la connexion
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("erreur lors du ping de la base de données: %w", err)
	}

	// Créer les tables si elles n'existent pas
	if err = createTables(); err != nil {
		return fmt.Errorf("erreur lors de la création des tables: %w", err)
	}

	log.Println("Base de données initialisée avec succès")
	return nil
}

// createTables crée toutes les tables nécessaires
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS buildings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT,
		building_id INTEGER,
		FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		number TEXT UNIQUE NOT NULL,
		description TEXT,
		quantity_total INTEGER DEFAULT 1,
		quantity_reserve INTEGER DEFAULT 0,
		storage_location TEXT
	);

	CREATE TABLE IF NOT EXISTS borrowers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT
	);

	CREATE TABLE IF NOT EXISTS loans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_id INTEGER NOT NULL,
		borrower_id INTEGER NOT NULL,
		loan_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		return_date DATETIME,
		FOREIGN KEY (key_id) REFERENCES keys(id) ON DELETE CASCADE,
		FOREIGN KEY (borrower_id) REFERENCES borrowers(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS key_room_association (
		key_id INTEGER NOT NULL,
		room_id INTEGER NOT NULL,
		PRIMARY KEY (key_id, room_id),
		FOREIGN KEY (key_id) REFERENCES keys(id) ON DELETE CASCADE,
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_keys_number ON keys(number);
	CREATE INDEX IF NOT EXISTS idx_borrowers_name ON borrowers(name);
	CREATE INDEX IF NOT EXISTS idx_loans_key_id ON loans(key_id);
	CREATE INDEX IF NOT EXISTS idx_loans_borrower_id ON loans(borrower_id);
	CREATE INDEX IF NOT EXISTS idx_loans_return_date ON loans(return_date);
	`

	_, err := DB.Exec(schema)
	return err
}

// CloseDB ferme la connexion à la base de données
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
