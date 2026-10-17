package app

import (
	"fmt"
	"time"

	"clefs/internal/db"
)

// maintenance.go : sauvegardes, assistant d'import des anciennes bases,
// réinitialisation et données de démonstration.

// ============= SAUVEGARDES =============

// ListBackups retourne les sauvegardes du dossier backups/ (plus récentes d'abord).
func (a *App) ListBackups() ([]BackupDTO, error) {
	defer a.read()()
	backups, err := db.ListBackups(a.dbPath)
	if err != nil {
		return nil, err
	}
	out := make([]BackupDTO, 0, len(backups))
	for _, b := range backups {
		out = append(out, BackupDTO{Path: b.Path, Name: b.Name, Size: b.SizeStr, Date: b.ModTime.Format(time.RFC3339)})
	}
	return out, nil
}

// CreateBackup crée une sauvegarde horodatée et retourne son chemin.
func (a *App) CreateBackup() (string, error) {
	defer a.read()()
	if err := db.CreateBackupDirectory(a.dbPath); err != nil {
		return "", err
	}
	path := db.GetDefaultBackupPath(a.dbPath)
	if err := db.BackupDatabase(a.dbPath, path); err != nil {
		return "", err
	}
	return path, nil
}

// RestoreBackup remplace la base par une sauvegarde du dossier backups/.
// Une copie de sécurité de la base actuelle est prise avant (RestoreDatabase).
func (a *App) RestoreBackup(path string) error {
	defer a.write()()
	info, err := db.InspectDatabase(path)
	if err != nil {
		return err
	}
	if info.Kind != db.DatabaseKindV3 {
		return fmt.Errorf("ce fichier n'est pas une sauvegarde valide")
	}
	return preserveSettings(func() error { return db.RestoreDatabase(path, a.dbPath) })
}

// DeleteBackup supprime une sauvegarde du dossier backups/.
func (a *App) DeleteBackup(path string) error {
	defer a.read()()
	return db.DeleteBackup(path)
}

// ============= ASSISTANT D'IMPORT =============

// ChooseDatabaseFile ouvre la sélection de l'ancienne base. "" si annulé.
func (a *App) ChooseDatabaseFile() (string, error) {
	return a.chooseFile("Choisir l'ancienne base de données (clefs.db)", "Base de données (*.db)", "*.db")
}

// kindLabel retourne le libellé lisible d'un type de base.
func kindLabel(kind string) string {
	switch kind {
	case db.DatabaseKindV3:
		return "Version 3 (base ou sauvegarde 3.x)"
	case db.DatabaseKindLegacy:
		return "Ancienne version (V1 Python ou V2)"
	default:
		return "Fichier non reconnu"
	}
}

// InspectDatabaseFile détecte la version d'une base et en donne un aperçu,
// sans rien modifier.
func (a *App) InspectDatabaseFile(path string) (DatabaseFileInfo, error) {
	defer a.read()()
	if sameFile(path, a.dbPath) {
		return DatabaseFileInfo{}, fmt.Errorf("ce fichier est la base actuellement ouverte par l'application")
	}
	info, err := db.InspectDatabase(path)
	if err != nil {
		return DatabaseFileInfo{}, err
	}
	return DatabaseFileInfo{
		Path: path, Kind: info.Kind, Label: kindLabel(info.Kind),
		Buildings: info.Buildings, Accesses: info.Rooms, Keys: info.Keys,
		Borrowers: info.Borrowers, Loans: info.Loans, ActiveLoans: info.ActiveLoans,
	}, nil
}

// ImportDatabaseFile importe une ancienne base :
//   - base 3.x : remplace la base courante (copie de sécurité prise avant) ;
//   - base V1/V2 : recopiée table par table, uniquement dans une base vide
//     (les identifiants d'origine sont conservés, un mélange avec des données
//     existantes créerait des conflits silencieux).
func (a *App) ImportDatabaseFile(path string) (ImportResult, error) {
	defer a.write()()
	if sameFile(path, a.dbPath) {
		return ImportResult{}, fmt.Errorf("ce fichier est la base actuellement ouverte par l'application")
	}
	info, err := db.InspectDatabase(path)
	if err != nil {
		return ImportResult{}, err
	}
	var summary *db.ImportSummary
	switch info.Kind {
	case db.DatabaseKindV3:
		err = preserveSettings(func() (e error) {
			summary, e = db.ImportFromV3(path, a.dbPath)
			return e
		})
	case db.DatabaseKindLegacy:
		empty, e := db.IsDatabaseEmpty()
		if e != nil {
			return ImportResult{}, e
		}
		if !empty {
			return ImportResult{}, fmt.Errorf("l'import d'une base V1/V2 se fait dans une base vide : réinitialisez d'abord la base (Paramètres)")
		}
		summary, err = db.ImportFromV2(path, a.dbPath)
	default:
		return ImportResult{}, fmt.Errorf("ce fichier n'est pas une base du Gestionnaire de Clés")
	}
	if err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		Kind: info.Kind, Buildings: summary.Buildings, Accesses: summary.Rooms,
		Keys: summary.Keys, Borrowers: summary.Borrowers, Loans: summary.Loans,
	}, nil
}

// ============= RÉINITIALISATION ET DÉMO =============

// ResetDatabase vide complètement la base, après une sauvegarde de sécurité.
// Les paramètres (établissement, textes des bons…) sont conservés.
func (a *App) ResetDatabase() error {
	defer a.write()()
	return preserveSettings(func() error { return db.ResetDatabase(a.dbPath) })
}

// LoadDemoData remplit une base vide avec des données de démonstration.
func (a *App) LoadDemoData() error {
	defer a.read()()
	empty, err := db.IsDatabaseEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("les données de démonstration se chargent dans une base vide")
	}
	return db.GenerateDemoData()
}
