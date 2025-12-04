package main

import (
	"clefs/internal/gui"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Déterminer le chemin de la base de données
	dbPath := getDBPath()

	log.Printf("Démarrage de l'application Gestionnaire de Clés")
	log.Printf("Base de données: %s", dbPath)

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
	// Essayer d'utiliser le répertoire de l'exécutable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		dbPath := filepath.Join(exeDir, "clefs.db")
		return dbPath
	}

	// Sinon, utiliser le répertoire courant
	return "clefs.db"
}
