package db

import "testing"

// TestMigrationsOnFreshDatabase vérifie qu'une base vierge sort d'InitDB avec
// toutes les migrations marquées appliquées.
func TestMigrationsOnFreshDatabase(t *testing.T) {
	setupTestDB(t)

	var current int
	if err := DB.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current); err != nil {
		t.Fatalf("lecture schema_version: %v", err)
	}
	latest := migrations[len(migrations)-1].version
	if current != latest {
		t.Errorf("version de schéma = %d, attendu %d", current, latest)
	}
}

// TestMigrationsIdempotent vérifie que rejouer applyMigrations sur une base à
// jour ne change rien et ne produit aucune erreur — c'est ce qui se passe à
// chaque démarrage de l'application et après chaque restauration.
func TestMigrationsIdempotent(t *testing.T) {
	setupTestDB(t)
	if err := applyMigrations(DB); err != nil {
		t.Fatalf("applyMigrations rejoué: %v", err)
	}
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&count); err != nil {
		t.Fatalf("comptage schema_version: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("entrées schema_version = %d, attendu %d", count, len(migrations))
	}
}

// TestReopenExistingDatabase simule un redémarrage de l'application sur une
// base existante : fermeture puis nouvel InitDB sur le même fichier.
func TestReopenExistingDatabase(t *testing.T) {
	dbPath := setupTestDB(t)
	mustCreateKey(t, "K-01", 2, 0)

	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB: %v", err)
	}
	if err := InitDB(dbPath); err != nil {
		t.Fatalf("réouverture: %v", err)
	}
	keys, err := GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys après réouverture: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("clés après réouverture = %d, attendu 1", len(keys))
	}
}
