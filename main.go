// Commande principale du Gestionnaire de Clés (V4, Wails).
//
// Point d'entrée de l'application : détermine le dossier de travail (celui
// de l'exécutable), prépare documents/ et backups/, ouvre la base SQLite,
// puis lance la fenêtre Wails. L'interface (frontend/, Svelte) est compilée
// dans frontend/dist et embarquée dans l'exécutable : rien à installer.
package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"strings"

	"clefs/internal/app"
	"clefs/internal/db"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// bindingsOnly vaut vrai quand le binaire est compilé avec le tag « bindings »
// (génération des liaisons JavaScript, voir main_bindings.go).
var bindingsOnly bool

func main() {
	if bindingsOnly {
		_ = wails.Run(&options.App{Bind: []interface{}{app.New("", "")}})
		return
	}
	baseDir := resolveBaseDir()
	// Les PDF et CSV sont écrits dans documents/, relatif au répertoire
	// courant : on se place dans le dossier de l'application, quel que soit
	// le dossier depuis lequel elle a été lancée.
	if err := os.Chdir(baseDir); err != nil {
		log.Fatalf("Dossier de l'application inaccessible : %v", err)
	}
	dbPath := filepath.Join(baseDir, "clefs.db")

	log.Printf("Démarrage du Gestionnaire de Clés %s", app.Version)
	log.Printf("Base de données : %s", dbPath)

	for _, dir := range []string{"documents", "backups"} {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			log.Printf("Avertissement : impossible de créer %s/ : %v", dir, err)
		}
	}
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Ouverture de la base impossible : %v", err)
	}

	service := app.New(baseDir, dbPath)
	err := wails.Run(&options.App{
		Title:            "Gestionnaire de Clés",
		Width:            1360,
		Height:           860,
		MinWidth:         1024,
		MinHeight:        680,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 248, G: 249, B: 251, A: 1},
		OnStartup:        service.Startup,
		OnShutdown:       service.Shutdown,
		Bind:             []interface{}{service},
		// Une seule fenêtre par poste : relancer l'application ramène la
		// fenêtre existante au premier plan.
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "fr.college-victor-hugo.gestionnaire-de-cles",
		},
		Windows: &windows.Options{Theme: windows.SystemDefault},
	})
	if err != nil {
		log.Fatalf("Erreur de l'interface : %v", err)
	}
}

// resolveBaseDir retourne le dossier de travail de l'application.
//
// En production, c'est le dossier de l'exécutable (le dossier complet peut
// être copié sur une clé USB ou un partage réseau). Avec `go run`,
// l'exécutable est compilé dans un dossier temporaire : on retombe alors sur
// le répertoire courant pour ne pas éparpiller des bases de test.
func resolveBaseDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	if strings.Contains(exe, "go-build") || strings.Contains(exe, `AppData\Local\Temp`) {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	return filepath.Dir(exe)
}
