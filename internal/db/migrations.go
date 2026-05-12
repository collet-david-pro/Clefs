package db

import (
	"database/sql"
	"fmt"
	"log"
)

type migration struct {
	version     int
	description string
	sql         string
}

var migrations = []migration{
	{
		version:     1,
		description: "baseline V2",
		sql:         "", // schéma déjà créé par createTables()
	},
	{
		version:     2,
		description: "V3 — enrichissement accès, détenteurs, prêts, clés",
		sql: `
			ALTER TABLE rooms ADD COLUMN floor TEXT;
			ALTER TABLE rooms ADD COLUMN category TEXT;
			ALTER TABLE rooms ADD COLUMN notes TEXT;
			ALTER TABLE borrowers ADD COLUMN status TEXT DEFAULT 'permanent';
			ALTER TABLE borrowers ADD COLUMN phone TEXT;
			ALTER TABLE loans ADD COLUMN planned_return_date DATETIME;
			ALTER TABLE loans ADD COLUMN loan_type TEXT DEFAULT 'ponctuel';
			ALTER TABLE loans ADD COLUMN returned_condition TEXT;
			ALTER TABLE loans ADD COLUMN created_by TEXT;
			ALTER TABLE keys ADD COLUMN category TEXT DEFAULT 'simple';
			ALTER TABLE keys ADD COLUMN notes TEXT;
		`,
	},
}

// applyMigrations crée la table schema_version si absente et applique les migrations manquantes.
func applyMigrations(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version     INTEGER PRIMARY KEY,
			applied_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)
	`); err != nil {
		return fmt.Errorf("impossible de créer schema_version: %w", err)
	}

	var current int
	_ = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current)

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		log.Printf("Application migration v%d: %s", m.version, m.description)
		if m.sql != "" {
			// SQLite ne supporte pas plusieurs instructions dans un seul Exec
			// On exécute chaque ALTER TABLE séparément
			stmts := splitSQL(m.sql)
			for _, stmt := range stmts {
				if stmt == "" {
					continue
				}
				if _, err := db.Exec(stmt); err != nil {
					return fmt.Errorf("migration v%d — %q: %w", m.version, stmt, err)
				}
			}
		}
		if _, err := db.Exec(
			`INSERT INTO schema_version (version, description) VALUES (?, ?)`,
			m.version, m.description,
		); err != nil {
			return fmt.Errorf("impossible d'enregistrer migration v%d: %w", m.version, err)
		}
		log.Printf("Migration v%d appliquée", m.version)
	}
	return nil
}

// splitSQL découpe un bloc SQL multi-instructions en instructions individuelles.
func splitSQL(block string) []string {
	var stmts []string
	var current []byte
	for i := 0; i < len(block); i++ {
		c := block[i]
		current = append(current, c)
		if c == ';' {
			stmt := trimSQL(string(current))
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			current = current[:0]
		}
	}
	return stmts
}

func trimSQL(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if c != '\n' && c != '\t' || len(out) > 0 {
			out = append(out, c)
		}
	}
	// trim leading/trailing whitespace
	start, end := 0, len(out)
	for start < end && (out[start] == ' ' || out[start] == '\n' || out[start] == '\t' || out[start] == '\r') {
		start++
	}
	for end > start && (out[end-1] == ' ' || out[end-1] == '\n' || out[end-1] == '\t' || out[end-1] == '\r' || out[end-1] == ';') {
		end--
	}
	return string(out[start:end])
}
