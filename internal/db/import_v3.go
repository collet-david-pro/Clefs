package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Types de base reconnus par InspectDatabase.
const (
	DatabaseKindV3      = "v3"      // base 3.x (ou 4.x) : schéma courant, restaurée telle quelle
	DatabaseKindLegacy  = "legacy"  // base V1 (Python) ou V2 (Go) : importée table par table
	DatabaseKindInvalid = "invalid" // fichier qui n'est pas une base du Gestionnaire de Clés
)

// DatabaseInfo décrit une base à importer : son type et un aperçu de son
// contenu, affiché par l'assistant d'import avant toute modification.
type DatabaseInfo struct {
	Kind        string
	Buildings   int
	Rooms       int
	Keys        int
	Borrowers   int
	Loans       int
	ActiveLoans int
}

// requiredTables sont les tables présentes dans toutes les versions (V1 à V4).
var requiredTables = []string{"buildings", "rooms", "keys", "borrowers", "loans", "key_room_association"}

// InspectDatabase ouvre le fichier en lecture seule, détermine sa version et
// compte son contenu. Le fichier source n'est jamais modifié. Un fichier qui
// n'est pas une base du Gestionnaire de Clés donne Kind = DatabaseKindInvalid
// sans erreur ; une erreur n'est retournée que si le fichier est illisible.
func InspectDatabase(path string) (*DatabaseInfo, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("fichier introuvable : %w", err)
	}
	src, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("ouverture impossible : %w", err)
	}
	defer src.Close()

	info := &DatabaseInfo{Kind: DatabaseKindInvalid}
	for _, table := range requiredTables {
		ok, err := tableExists(src, table)
		if err != nil {
			// Fichier qui n'est pas du SQLite (document, image...) : pas une base.
			return info, nil
		}
		if !ok {
			return info, nil
		}
	}

	// Marqueurs V3 : la table schema_version (créée depuis la 3.0) ou la
	// colonne rooms.floor (ajoutée par la migration V3).
	hasVersion, err := tableExists(src, "schema_version")
	if err != nil {
		return nil, err
	}
	hasFloor, err := columnExists(src, "rooms", "floor")
	if err != nil {
		return nil, err
	}
	if hasVersion || hasFloor {
		info.Kind = DatabaseKindV3
	} else {
		info.Kind = DatabaseKindLegacy
	}

	counts := []struct {
		dest  *int
		query string
	}{
		{&info.Buildings, `SELECT COUNT(*) FROM buildings`},
		{&info.Rooms, `SELECT COUNT(*) FROM rooms`},
		{&info.Keys, `SELECT COUNT(*) FROM keys`},
		{&info.Borrowers, `SELECT COUNT(*) FROM borrowers`},
		{&info.Loans, `SELECT COUNT(*) FROM loans`},
		{&info.ActiveLoans, `SELECT COUNT(*) FROM loans WHERE return_date IS NULL OR return_date = ''`},
	}
	for _, c := range counts {
		if err := src.QueryRow(c.query).Scan(c.dest); err != nil {
			return nil, fmt.Errorf("lecture du contenu : %w", err)
		}
	}
	return info, nil
}

// tableExists indique si la table existe dans la base src.
func tableExists(src *sql.DB, table string) (bool, error) {
	var n int
	err := src.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&n)
	return n > 0, err
}

// columnExists indique si la colonne existe dans la table de la base src.
func columnExists(src *sql.DB, table, column string) (bool, error) {
	var n int
	err := src.QueryRow(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, column).Scan(&n)
	return n > 0, err
}

// ImportFromV3 remplace la base courante par le contenu d'une base 3.x.
//
// La source est d'abord copiée par VACUUM INTO depuis sa propre connexion :
// la copie intègre ainsi les écritures encore présentes dans son fichier
// journal (-wal), ce qu'une simple copie du .db perdrait. La copie est ensuite
// restaurée par RestoreDatabase (sauvegarde de sécurité de la base courante,
// remplacement, réouverture et migrations éventuelles).
func ImportFromV3(srcPath string, currentDBPath string) (*ImportSummary, error) {
	info, err := InspectDatabase(srcPath)
	if err != nil {
		return nil, err
	}
	if info.Kind != DatabaseKindV3 {
		return nil, fmt.Errorf("ce fichier n'est pas une base 3.x")
	}

	if err := CreateBackupDirectory(currentDBPath); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire de sauvegarde: %w", err)
	}
	snapshot := filepath.Join(filepath.Dir(currentDBPath), "backups",
		"import_source_"+time.Now().Format("20060102_150405")+".db")

	src, err := sql.Open("sqlite", srcPath)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base source : %w", err)
	}
	if _, err := src.Exec("VACUUM INTO ?", snapshot); err != nil {
		src.Close()
		return nil, fmt.Errorf("copie de la base source : %w", err)
	}
	if err := src.Close(); err != nil {
		return nil, fmt.Errorf("fermeture de la base source : %w", err)
	}
	defer os.Remove(snapshot)

	if err := RestoreDatabase(snapshot, currentDBPath); err != nil {
		return nil, err
	}

	summary := &ImportSummary{}
	counts := []struct {
		dest  *int
		query string
	}{
		{&summary.Buildings, `SELECT COUNT(*) FROM buildings`},
		{&summary.Rooms, `SELECT COUNT(*) FROM rooms`},
		{&summary.Keys, `SELECT COUNT(*) FROM keys`},
		{&summary.Associations, `SELECT COUNT(*) FROM key_room_association`},
		{&summary.Borrowers, `SELECT COUNT(*) FROM borrowers`},
		{&summary.Loans, `SELECT COUNT(*) FROM loans`},
	}
	for _, c := range counts {
		if err := DB.QueryRow(c.query).Scan(c.dest); err != nil {
			return nil, fmt.Errorf("vérification après import : %w", err)
		}
	}
	log.Printf("Import 3.x réussi: %d bâtiments, %d accès, %d clés, %d détenteurs, %d prêts",
		summary.Buildings, summary.Rooms, summary.Keys, summary.Borrowers, summary.Loans)
	return summary, nil
}

// IsDatabaseEmpty indique si la base courante ne contient encore aucune
// donnée (ni bâtiment, ni accès, ni clé, ni détenteur) : c'est le cas d'une
// nouvelle installation, où l'assistant d'import est proposé.
func IsDatabaseEmpty() (bool, error) {
	var n int
	err := DB.QueryRow(`SELECT (SELECT COUNT(*) FROM buildings) + (SELECT COUNT(*) FROM rooms)
		+ (SELECT COUNT(*) FROM keys) + (SELECT COUNT(*) FROM borrowers)`).Scan(&n)
	return n == 0, err
}
