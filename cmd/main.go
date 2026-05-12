package main

import (
	"clefs/internal/db"
	"clefs/internal/gui"
	"clefs/internal/pdf"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Déterminer le chemin de la base de données
	dbPath := getDBPath()

	log.Printf("Démarrage de l'application Gestionnaire de Clés")
	log.Printf("Base de données: %s", dbPath)

	// Créer le dossier documents au démarrage
	if err := pdf.EnsureDocumentsDir(); err != nil {
		log.Printf("Avertissement: Impossible de créer le dossier documents: %v", err)
	} else {
		log.Printf("Dossier documents prêt")
	}

	// Créer le dossier backups au démarrage
	if err := db.CreateBackupDirectory(dbPath); err != nil {
		log.Printf("Avertissement: Impossible de créer le dossier backups: %v", err)
	} else {
		log.Printf("Dossier backups prêt")
	}

	// Initialiser l'application
	app, err := gui.Initialize(dbPath)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation: %v", err)
	}

	// Lancer l'application
	app.Run()
}

// getDBPath retourne le chemin de la base de données
func getDBPath() string {
	// Vérifier si on est en mode développement (go run)
	exePath, err := os.Executable()
	if err == nil {
		// Si le chemin contient "go-build" ou est dans un dossier temporaire, on est probablement en mode go run
		if strings.Contains(exePath, "go-build") || strings.Contains(exePath, "/var/folders/") || strings.Contains(exePath, "AppData\\Local\\Temp") {
			// Utiliser le répertoire courant
			cwd, err := os.Getwd()
			if err == nil {
				return filepath.Join(cwd, "clefs.db")
			}
		}
		// Sinon, utiliser le répertoire de l'exécutable (mode production)
		return filepath.Join(filepath.Dir(exePath), "clefs.db")
	}

	// Fallback: utiliser le répertoire courant
	return "clefs.db"
}
