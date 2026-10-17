package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// TestSettingsRoundTrip vérifie l'enregistrement, la mise à jour et la
// suppression (valeur vide) des paramètres.
func TestSettingsRoundTrip(t *testing.T) {
	setupTestDB(t)
	if m, err := GetSettings(); err != nil || len(m) != 0 {
		t.Fatalf("base neuve: %v, %v", m, err)
	}
	if err := SaveSettings(map[string]string{"establishment": "Collège A", "address": "Ville"}); err != nil {
		t.Fatal(err)
	}
	if err := SaveSettings(map[string]string{"establishment": "Lycée B", "address": ""}); err != nil {
		t.Fatal(err)
	}
	m, err := GetSettings()
	if err != nil {
		t.Fatal(err)
	}
	if m["establishment"] != "Lycée B" || len(m) != 1 {
		t.Errorf("paramètres: %v", m)
	}
}

// TestSettingsTableAddedToV3Base vérifie qu'une base 3.x (sans table
// settings) reçoit la table à son ouverture, sans perte de données.
func TestSettingsTableAddedToV3Base(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v3.db")
	if err := InitDB(path); err != nil {
		t.Fatal(err)
	}
	mustCreateKey(t, "K-01", 1, 0)
	// Ramener la base à l'état 3.x : sans table settings, schéma en version 2.
	for _, stmt := range []string{`DROP TABLE settings`, `DELETE FROM schema_version WHERE version > 2`} {
		if _, err := DB.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	if err := CloseDB(); err != nil {
		t.Fatal(err)
	}
	src, _ := sql.Open("sqlite", path)
	var n int
	src.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name = 'settings'`).Scan(&n)
	src.Close()
	if n != 0 {
		t.Fatal("la base 3.x simulée ne doit pas avoir de table settings")
	}

	if err := InitDB(path); err != nil {
		t.Fatalf("réouverture: %v", err)
	}
	t.Cleanup(func() { CloseDB() })
	if err := SaveSettings(map[string]string{"establishment": "X"}); err != nil {
		t.Errorf("table settings absente après migration: %v", err)
	}
	if keys, _ := GetAllKeys(); len(keys) != 1 {
		t.Errorf("données perdues: %v", keys)
	}
}
