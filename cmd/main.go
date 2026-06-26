// Commande principale du Gestionnaire de Clés.
//
// Point d'entrée de l'application : résout le chemin de la base SQLite,
// prépare les dossiers de travail (documents/, backups/), initialise la
// couche base de données + l'interface Fyne, puis lance la boucle d'événements.
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

// main orchestre le démarrage : chemin DB → dossiers → init GUI → boucle d'événements.
// Toute erreur d'initialisation est fatale (log.Fatalf) car l'application
// ne peut pas fonctionner sans base de données.
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

// getDBPath retourne le chemin du fichier clefs.db.
//
// En production, la base est placée à côté de l'exécutable (permet de copier
// tout le dossier sur une clé USB ou un partage réseau). En développement
// (go run), l'exécutable est compilé dans un dossier temporaire système ;
// on détecte ce cas et on retombe sur le répertoire courant pour ne pas
// éparpiller des bases de test dans /var/folders.
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
