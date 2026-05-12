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

	// Activer les options SQLite essentielles
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA cache_size = -4000",
	}
	for _, pragma := range pragmas {
		if _, err = DB.Exec(pragma); err != nil {
			return fmt.Errorf("erreur pragma (%s): %w", pragma, err)
		}
	}

	// Une seule connexion par process pour SQLite
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	// Créer les tables si elles n'existent pas
	if err = createTables(); err != nil {
		return fmt.Errorf("erreur lors de la création des tables: %w", err)
	}

	// Appliquer les migrations de schéma
	if err = applyMigrations(DB); err != nil {
		return fmt.Errorf("erreur lors des migrations: %w", err)
	}

	log.Println("Base de données initialisée avec succès")
	return nil
}

// createTables crée toutes les tables V3 pour une base vierge.
func createTables() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS buildings (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL,
			type        TEXT,
			building_id INTEGER,
			floor       TEXT,
			category    TEXT,
			notes       TEXT,
			FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS keys (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			number           TEXT UNIQUE NOT NULL,
			description      TEXT,
			quantity_total   INTEGER DEFAULT 1,
			quantity_reserve INTEGER DEFAULT 0,
			storage_location TEXT,
			category         TEXT DEFAULT 'simple',
			notes            TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS borrowers (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			name   TEXT NOT NULL,
			email  TEXT,
			status TEXT DEFAULT 'permanent',
			phone  TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS loans (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			key_id               INTEGER NOT NULL,
			borrower_id          INTEGER NOT NULL,
			loan_date            DATETIME DEFAULT CURRENT_TIMESTAMP,
			return_date          DATETIME,
			planned_return_date  DATETIME,
			loan_type            TEXT DEFAULT 'ponctuel',
			returned_condition   TEXT,
			created_by           TEXT,
			FOREIGN KEY (key_id)      REFERENCES keys(id)      ON DELETE CASCADE,
			FOREIGN KEY (borrower_id) REFERENCES borrowers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS key_room_association (
			key_id  INTEGER NOT NULL,
			room_id INTEGER NOT NULL,
			PRIMARY KEY (key_id, room_id),
			FOREIGN KEY (key_id)  REFERENCES keys(id)  ON DELETE CASCADE,
			FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS schema_version (
			version    INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_keys_number       ON keys(number)`,
		`CREATE INDEX IF NOT EXISTS idx_borrowers_name    ON borrowers(name)`,
		`CREATE INDEX IF NOT EXISTS idx_loans_key_id      ON loans(key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_loans_borrower_id ON loans(borrower_id)`,
		`CREATE INDEX IF NOT EXISTS idx_loans_return_date ON loans(return_date)`,
		// Marquer toutes les migrations comme appliquées sur base vierge
		`INSERT OR IGNORE INTO schema_version (version, description) VALUES (1, 'baseline V2')`,
		`INSERT OR IGNORE INTO schema_version (version, description) VALUES (2, 'V3 — enrichissement accès, détenteurs, prêts, clés')`,
	}
	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("%q: %w", stmt[:min(40, len(stmt))], err)
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CloseDB ferme la connexion à la base de données
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
