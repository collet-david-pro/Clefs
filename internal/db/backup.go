package db

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BackupInfo contient les informations sur une sauvegarde
type BackupInfo struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
	SizeStr string
}

// BackupDatabase crée une sauvegarde atomique via VACUUM INTO.
// Contrairement à une copie de fichier, VACUUM INTO est sûr même si SQLite écrit simultanément.
func BackupDatabase(dbPath string, backupPath string) error {
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("erreur lors de la création du répertoire de sauvegarde: %w", err)
	}
	// Supprimer le fichier destination s'il existe déjà (VACUUM INTO échoue sinon)
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("erreur lors de la suppression de l'ancienne sauvegarde: %w", err)
	}
	// Chemin passé en paramètre : une apostrophe dans le chemin (ex. dossier
	// « d'Artagnan ») cassait la requête quand il était concaténé.
	if _, err := DB.Exec("VACUUM INTO ?", backupPath); err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde: %w", err)
	}
	return nil
}

// RestoreDatabase restaure une base de données depuis une sauvegarde
func RestoreDatabase(backupPath string, dbPath string) error {
	// Vérifier que le fichier de sauvegarde existe
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("le fichier de sauvegarde n'existe pas: %s", backupPath)
	}

	// Créer une sauvegarde de la base actuelle avant de la remplacer.
	// IMPORTANT : cette copie doit être prise AVANT de fermer la connexion,
	// car BackupDatabase s'appuie sur VACUUM INTO qui passe par la connexion
	// ouverte. L'ordre inverse (fermer puis sauvegarder) rendait toute
	// restauration impossible (« database is closed ») — corrigé en 3.3.0,
	// verrouillé par TestBackupAndRestore.
	if _, err := os.Stat(dbPath); err == nil && DB != nil {
		backupCurrent := dbPath + ".before_restore." + time.Now().Format("20060102_150405")
		if err := BackupDatabase(dbPath, backupCurrent); err != nil {
			return fmt.Errorf("erreur lors de la sauvegarde de la base actuelle: %w", err)
		}
	}

	// Fermer la connexion actuelle si elle existe
	if DB != nil {
		if err := DB.Close(); err != nil {
			return fmt.Errorf("erreur lors de la fermeture de la base de données: %w", err)
		}
	}

	// Ouvrir le fichier de sauvegarde
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier de sauvegarde: %w", err)
	}
	defer sourceFile.Close()

	// Créer/écraser le fichier de base de données
	destFile, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la base de données: %w", err)
	}
	defer destFile.Close()

	// Copier le contenu
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la copie de la sauvegarde: %w", err)
	}

	// Synchroniser
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("erreur lors de la synchronisation: %w", err)
	}

	// Rouvrir la connexion à la base de données
	err = InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("erreur lors de la réouverture de la base de données: %w", err)
	}

	return nil
}

// GetDefaultBackupPath retourne le chemin par défaut pour une sauvegarde
func GetDefaultBackupPath(dbPath string) string {
	timestamp := time.Now().Format("20060102_150405")
	dir := filepath.Dir(dbPath)
	filename := fmt.Sprintf("clefs_backup_%s.db", timestamp)
	return filepath.Join(dir, "backups", filename)
}

// CreateBackupDirectory crée le répertoire de sauvegarde s'il n'existe pas
func CreateBackupDirectory(dbPath string) error {
	dir := filepath.Dir(dbPath)
	backupDir := filepath.Join(dir, "backups")
	return os.MkdirAll(backupDir, 0755)
}

// ResetDatabase réinitialise complètement la base de données
// ATTENTION: Cette fonction supprime TOUTES les données !
func ResetDatabase(dbPath string) error {
	// Créer une sauvegarde de sécurité avant la réinitialisation
	if err := CreateBackupDirectory(dbPath); err != nil {
		return fmt.Errorf("erreur lors de la création du répertoire de sauvegarde: %w", err)
	}

	backupPath := GetDefaultBackupPath(dbPath)
	if err := BackupDatabase(dbPath, backupPath); err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde de sécurité: %w", err)
	}

	// Fermer la connexion actuelle
	if err := CloseDB(); err != nil {
		return fmt.Errorf("erreur lors de la fermeture de la base de données: %w", err)
	}

	// Supprimer le fichier de base de données
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("erreur lors de la suppression de la base de données: %w", err)
	}

	// Réinitialiser la base de données
	if err := InitDB(dbPath); err != nil {
		return fmt.Errorf("erreur lors de la réinitialisation de la base de données: %w", err)
	}

	return nil
}

// ListBackups liste toutes les sauvegardes disponibles dans le dossier backups
func ListBackups(dbPath string) ([]BackupInfo, error) {
	dir := filepath.Dir(dbPath)
	backupDir := filepath.Join(dir, "backups")

	// Vérifier si le répertoire existe
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	// Lire le contenu du répertoire
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du répertoire de sauvegarde: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Ne garder que les fichiers .db
		if filepath.Ext(entry.Name()) != ".db" {
			continue
		}

		fullPath := filepath.Join(backupDir, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		backup := BackupInfo{
			Path:    fullPath,
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			SizeStr: formatFileSize(info.Size()),
		}
		backups = append(backups, backup)
	}

	// Trier par date de modification (plus récent en premier)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// DeleteBackup supprime une sauvegarde spécifique
func DeleteBackup(backupPath string) error {
	// Vérifier que le fichier existe
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("le fichier de sauvegarde n'existe pas: %s", backupPath)
	}

	// Vérifier que c'est bien un fichier dans le dossier backups
	if !filepath.IsAbs(backupPath) {
		return fmt.Errorf("le chemin doit être absolu")
	}

	dir := filepath.Dir(backupPath)
	if filepath.Base(dir) != "backups" {
		return fmt.Errorf("le fichier doit être dans le dossier 'backups'")
	}

	// Supprimer le fichier
	err := os.Remove(backupPath)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de la sauvegarde: %w", err)
	}

	return nil
}

// formatFileSize formate la taille d'un fichier en une chaîne lisible
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GenerateDemoData remplit la base de données avec des données de test
func GenerateDemoData() error {
	// Créer des bâtiments
	buildings := []string{
		"Bâtiment Principal",
		"Annexe A",
		"Annexe B",
		"Laboratoire",
		"Bibliothèque",
	}

	buildingIDs := make(map[string]int)
	for _, name := range buildings {
		result, err := DB.Exec("INSERT INTO buildings (name) VALUES (?)", name)
		if err != nil {
			return fmt.Errorf("erreur lors de la création du bâtiment %s: %w", name, err)
		}
		id, _ := result.LastInsertId()
		buildingIDs[name] = int(id)
	}

	// Créer des salles
	rooms := []struct {
		name     string
		roomType string
		building string
	}{
		{"Salle 101", "Bureau", "Bâtiment Principal"},
		{"Salle 102", "Bureau", "Bâtiment Principal"},
		{"Salle 201", "Salle de réunion", "Bâtiment Principal"},
		{"Amphithéâtre A", "Amphithéâtre", "Bâtiment Principal"},
		{"Laboratoire 1", "Laboratoire", "Laboratoire"},
		{"Laboratoire 2", "Laboratoire", "Laboratoire"},
		{"Salle de lecture", "Bibliothèque", "Bibliothèque"},
		{"Archives", "Stockage", "Bibliothèque"},
		{"Bureau A1", "Bureau", "Annexe A"},
		{"Bureau A2", "Bureau", "Annexe A"},
		{"Salle B1", "Salle de cours", "Annexe B"},
		{"Cafétéria", "Restauration", "Annexe B"},
	}

	roomIDs := make(map[string]int)
	for _, room := range rooms {
		buildingID := buildingIDs[room.building]
		result, err := DB.Exec("INSERT INTO rooms (name, type, building_id) VALUES (?, ?, ?)",
			room.name, room.roomType, buildingID)
		if err != nil {
			return fmt.Errorf("erreur lors de la création de la salle %s: %w", room.name, err)
		}
		id, _ := result.LastInsertId()
		roomIDs[room.name] = int(id)
	}

	// Créer des clés
	keys := []struct {
		number      string
		description string
		total       int
		reserve     int
		storage     string
		rooms       []string
	}{
		{"K001", "Clé principale du bâtiment", 3, 1, "Bureau d'accueil", []string{"Salle 101", "Salle 102", "Salle 201"}},
		{"K002", "Clé de l'amphithéâtre", 2, 0, "Bureau d'accueil", []string{"Amphithéâtre A"}},
		{"K003", "Clé des laboratoires", 5, 2, "Laboratoire 1", []string{"Laboratoire 1", "Laboratoire 2"}},
		{"K004", "Clé de la bibliothèque", 4, 1, "Bureau d'accueil", []string{"Salle de lecture", "Archives"}},
		{"K005", "Clé Annexe A", 2, 0, "Bureau A1", []string{"Bureau A1", "Bureau A2"}},
		{"K006", "Clé Annexe B", 3, 1, "Salle B1", []string{"Salle B1", "Cafétéria"}},
		{"K007", "Passe-partout", 1, 0, "Direction", []string{"Salle 101", "Salle 102", "Salle 201", "Bureau A1", "Bureau A2"}},
		{"K008", "Clé salle de réunion", 2, 0, "Bureau d'accueil", []string{"Salle 201"}},
		{"K009", "Clé cafétéria", 3, 0, "Cafétéria", []string{"Cafétéria"}},
		{"K010", "Clé archives", 1, 0, "Archives", []string{"Archives"}},
	}

	keyIDs := make(map[string]int)
	for _, key := range keys {
		result, err := DB.Exec("INSERT INTO keys (number, description, quantity_total, quantity_reserve, storage_location) VALUES (?, ?, ?, ?, ?)",
			key.number, key.description, key.total, key.reserve, key.storage)
		if err != nil {
			return fmt.Errorf("erreur lors de la création de la clé %s: %w", key.number, err)
		}
		id, _ := result.LastInsertId()
		keyIDs[key.number] = int(id)

		// Associer les salles
		for _, roomName := range key.rooms {
			if roomID, ok := roomIDs[roomName]; ok {
				_, err = DB.Exec("INSERT INTO key_room_association (key_id, room_id) VALUES (?, ?)", id, roomID)
				if err != nil {
					return fmt.Errorf("erreur lors de l'association clé-salle: %w", err)
				}
			}
		}
	}

	// Créer des emprunteurs
	borrowers := []struct {
		name  string
		email string
	}{
		{"Jean Dupont", "jean.dupont@example.com"},
		{"Marie Martin", "marie.martin@example.com"},
		{"Pierre Durand", "pierre.durand@example.com"},
		{"Sophie Bernard", "sophie.bernard@example.com"},
		{"Luc Petit", "luc.petit@example.com"},
		{"Emma Roux", "emma.roux@example.com"},
		{"Thomas Moreau", "thomas.moreau@example.com"},
		{"Julie Simon", "julie.simon@example.com"},
	}

	borrowerIDs := make([]int, 0)
	for _, borrower := range borrowers {
		result, err := DB.Exec("INSERT INTO borrowers (name, email) VALUES (?, ?)",
			borrower.name, borrower.email)
		if err != nil {
			return fmt.Errorf("erreur lors de la création de l'emprunteur %s: %w", borrower.name, err)
		}
		id, _ := result.LastInsertId()
		borrowerIDs = append(borrowerIDs, int(id))
	}

	// Créer quelques emprunts actifs
	loans := []struct {
		keyNumber  string
		borrowerID int
	}{
		{"K001", borrowerIDs[0]},
		{"K003", borrowerIDs[1]},
		{"K003", borrowerIDs[2]},
		{"K004", borrowerIDs[3]},
		{"K006", borrowerIDs[4]},
		{"K009", borrowerIDs[5]},
	}

	for _, loan := range loans {
		keyID := keyIDs[loan.keyNumber]
		_, err := DB.Exec("INSERT INTO loans (key_id, borrower_id, loan_date) VALUES (?, ?, datetime('now', '-' || abs(random() % 10) || ' days'))",
			keyID, loan.borrowerID)
		if err != nil {
			return fmt.Errorf("erreur lors de la création de l'emprunt: %w", err)
		}
	}

	return nil
}
