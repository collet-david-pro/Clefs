package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// createFakeV2Database fabrique un fichier SQLite au schéma V2 (sans les
// colonnes ajoutées en V3) avec un petit jeu de données représentatif.
func createFakeV2Database(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "clefs_v2.db")
	src, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("ouverture base V2 factice: %v", err)
	}
	defer src.Close()

	stmts := []string{
		`CREATE TABLE buildings (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL)`,
		`CREATE TABLE rooms (id INTEGER PRIMARY KEY, name TEXT NOT NULL, type TEXT, building_id INTEGER)`,
		`CREATE TABLE keys (id INTEGER PRIMARY KEY, number TEXT UNIQUE NOT NULL, description TEXT,
			quantity_total INTEGER DEFAULT 1, quantity_reserve INTEGER DEFAULT 0, storage_location TEXT)`,
		`CREATE TABLE borrowers (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT)`,
		`CREATE TABLE loans (id INTEGER PRIMARY KEY, key_id INTEGER NOT NULL, borrower_id INTEGER NOT NULL,
			loan_date DATETIME DEFAULT CURRENT_TIMESTAMP, return_date DATETIME)`,
		`CREATE TABLE key_room_association (key_id INTEGER NOT NULL, room_id INTEGER NOT NULL,
			PRIMARY KEY (key_id, room_id))`,
		`INSERT INTO buildings (id, name) VALUES (1, 'Bâtiment A')`,
		`INSERT INTO rooms (id, name, type, building_id) VALUES (1, 'Salle 101', 'salle', 1)`,
		`INSERT INTO keys (id, number, description, quantity_total, quantity_reserve, storage_location)
			VALUES (1, 'K-01', 'Passe général', 3, 1, 'Armoire 1')`,
		`INSERT INTO borrowers (id, name, email) VALUES (1, 'Alice', 'alice@test.local')`,
		`INSERT INTO loans (id, key_id, borrower_id, loan_date, return_date)
			VALUES (1, 1, 1, '2025-01-15 10:00:00', NULL)`,
		`INSERT INTO loans (id, key_id, borrower_id, loan_date, return_date)
			VALUES (2, 1, 1, '2024-06-01 09:00:00', '2024-06-02 17:00:00')`,
		`INSERT INTO key_room_association (key_id, room_id) VALUES (1, 1)`,
	}
	for _, stmt := range stmts {
		if _, err := src.Exec(stmt); err != nil {
			t.Fatalf("préparation base V2 (%.40s...): %v", stmt, err)
		}
	}
	return path
}

// TestValidateV2Schema vérifie la détection d'une base V2 valide et le rejet
// d'un fichier quelconque.
func TestValidateV2Schema(t *testing.T) {
	setupTestDB(t)
	v2Path := createFakeV2Database(t)
	if err := ValidateV2Schema(v2Path); err != nil {
		t.Errorf("base V2 valide rejetée: %v", err)
	}

	empty := filepath.Join(t.TempDir(), "vide.db")
	if src, err := sql.Open("sqlite", empty); err == nil {
		src.Exec(`CREATE TABLE autre (id INTEGER)`)
		src.Close()
	}
	if err := ValidateV2Schema(empty); err == nil {
		t.Error("fichier sans schéma V2 accepté à tort")
	}
}

// TestImportFromV2 verrouille la garantie de compatibilité ascendante :
// une base V2 doit s'importer intégralement, les colonnes V3 absentes prenant
// leur valeur par défaut (category 'simple', status 'permanent'...).
func TestImportFromV2(t *testing.T) {
	dbPath := setupTestDB(t)
	v2Path := createFakeV2Database(t)

	summary, err := ImportFromV2(v2Path, dbPath)
	if err != nil {
		t.Fatalf("ImportFromV2: %v", err)
	}
	want := ImportSummary{Buildings: 1, Rooms: 1, Keys: 1, Associations: 1, Borrowers: 1, Loans: 2}
	if *summary != want {
		t.Errorf("résumé d'import = %+v, attendu %+v", *summary, want)
	}

	// Valeurs par défaut V3 appliquées aux données importées
	k, err := GetKeyByID(1)
	if err != nil {
		t.Fatalf("GetKeyByID: %v", err)
	}
	if k.Category != "simple" {
		t.Errorf("catégorie de clé importée = %q, attendu 'simple'", k.Category)
	}
	b, err := GetBorrowerByID(1)
	if err != nil {
		t.Fatalf("GetBorrowerByID: %v", err)
	}
	if b.Status != "permanent" {
		t.Errorf("statut du détenteur importé = %q, attendu 'permanent'", b.Status)
	}

	// Le prêt actif V2 doit rester actif, le prêt rendu rester rendu
	if n, _ := GetActiveLoanCount(1); n != 1 {
		t.Errorf("prêts actifs importés = %d, attendu 1", n)
	}
	// Les associations clé↔accès doivent suivre
	rooms, err := GetRoomsForKey(1)
	if err != nil {
		t.Fatalf("GetRoomsForKey: %v", err)
	}
	if len(rooms) != 1 || rooms[0].Name != "Salle 101" {
		t.Errorf("accès associés importés inattendus : %+v", rooms)
	}
}
