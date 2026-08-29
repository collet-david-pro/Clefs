package db

import (
	"path/filepath"
	"testing"
)

// TestBackupAndRestore vérifie le cycle sauvegarde → modification → restauration :
// les données présentes au moment de la sauvegarde doivent réapparaître à
// l'identique, et la connexion doit être rouverte automatiquement.
func TestBackupAndRestore(t *testing.T) {
	dbPath := setupTestDB(t)
	k := mustCreateKey(t, "K-01", 3, 0)
	mustCreateBorrower(t, "Alice")

	backupPath := filepath.Join(t.TempDir(), "backup.db")
	if err := BackupDatabase(dbPath, backupPath); err != nil {
		t.Fatalf("BackupDatabase: %v", err)
	}

	// Modifier la base après la sauvegarde
	if err := DeleteKey(k.ID); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}
	if keys, _ := GetAllKeys(); len(keys) != 0 {
		t.Fatalf("la clé aurait dû être supprimée avant restauration")
	}

	// Restaurer : RestoreDatabase referme puis rouvre la connexion (InitDB),
	// ce qui rejoue aussi les migrations sur la base restaurée.
	if err := RestoreDatabase(backupPath, dbPath); err != nil {
		t.Fatalf("RestoreDatabase: %v", err)
	}

	keys, err := GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys après restauration: %v", err)
	}
	if len(keys) != 1 || keys[0].Number != "K-01" {
		t.Errorf("données restaurées inattendues : %+v", keys)
	}
	borrowers, _ := GetAllBorrowers()
	if len(borrowers) != 1 || borrowers[0].Name != "Alice" {
		t.Errorf("détenteurs restaurés inattendus : %+v", borrowers)
	}
}

// TestRestoreBackupWithNegativeStock verrouille la garantie de compatibilité :
// une sauvegarde contenant un sur-prêt (stock disponible négatif) doit se
// restaurer sans erreur — aucune contrainte de schéma ne doit la rejeter.
func TestRestoreBackupWithNegativeStock(t *testing.T) {
	dbPath := setupTestDB(t)
	k := mustCreateKey(t, "K-01", 1, 0)
	alice := mustCreateBorrower(t, "Alice")
	bob := mustCreateBorrower(t, "Bob")
	if err := CreateMultipleLoans([]int{k.ID}, alice.ID); err != nil {
		t.Fatalf("prêt 1: %v", err)
	}
	if err := CreateMultipleLoans([]int{k.ID}, bob.ID); err != nil {
		t.Fatalf("prêt 2: %v", err)
	}

	backupPath := filepath.Join(t.TempDir(), "backup_negatif.db")
	if err := BackupDatabase(dbPath, backupPath); err != nil {
		t.Fatalf("BackupDatabase: %v", err)
	}
	if err := RestoreDatabase(backupPath, dbPath); err != nil {
		t.Fatalf("restauration d'une sauvegarde à stock négatif refusée: %v", err)
	}

	kwa := availabilityOf(t, k.ID)
	if kwa.AvailableCount != -1 || kwa.LoanedCount != 2 {
		t.Errorf("après restauration : disponible=%d prêtées=%d, attendu -1/2",
			kwa.AvailableCount, kwa.LoanedCount)
	}
}

// TestListBackups vérifie que les sauvegardes déposées dans le dossier backups
// sont bien listées.
func TestListBackups(t *testing.T) {
	dbPath := setupTestDB(t)
	mustCreateKey(t, "K-01", 1, 0)

	if err := CreateBackupDirectory(dbPath); err != nil {
		t.Fatalf("CreateBackupDirectory: %v", err)
	}
	backupPath := GetDefaultBackupPath(dbPath)
	if err := BackupDatabase(dbPath, backupPath); err != nil {
		t.Fatalf("BackupDatabase: %v", err)
	}

	backups, err := ListBackups(dbPath)
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("sauvegardes listées = %d, attendu 1", len(backups))
	}
}
