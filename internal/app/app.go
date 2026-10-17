// Package app est la couche de services exposée à l'interface (V4, Wails).
//
// Chaque méthode publique de App devient une fonction JavaScript appelable
// depuis le frontend (frontend/wailsjs/go/app/App.js, généré par Wails).
// Cette couche est volontairement mince : elle valide les saisies, appelle
// internal/db et internal/business, et renvoie des structures simples (DTO,
// voir dto.go) sérialisées en JSON. Aucune règle métier dans le frontend.
//
// Concurrence : Wails exécute les appels du frontend en parallèle. Les
// opérations qui remplacent la base (restauration, import, réinitialisation)
// prennent le verrou en écriture ; toutes les autres le prennent en lecture,
// pour qu'aucune requête ne tombe sur une connexion en cours de remplacement.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"clefs/internal/db"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version est la version courante de l'application, affichée dans « À propos ».
// Centralisée ici : toute évolution de version se fait à cet unique endroit.
const Version = "4.0.0"

// App porte l'état de l'application : chemins de travail et contexte Wails.
type App struct {
	ctx     context.Context
	baseDir string // dossier de l'exécutable : base, documents/, backups/
	dbPath  string
	mu      sync.RWMutex
}

// New crée le service. baseDir est le dossier de travail (celui de
// l'exécutable) ; la base doit déjà être ouverte (db.InitDB).
func New(baseDir, dbPath string) *App {
	return &App{baseDir: baseDir, dbPath: dbPath}
}

// Startup est appelé par Wails au démarrage : il conserve le contexte,
// nécessaire aux boîtes de dialogue natives.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Nom de l'application choisi dans les Paramètres → titre de la fenêtre.
	if s, err := loadSettings(); err == nil {
		runtime.WindowSetTitle(ctx, s.AppName)
	}
}

// Shutdown est appelé par Wails à la fermeture de la fenêtre.
func (a *App) Shutdown(_ context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = db.CloseDB()
}

// Quit ferme l'application (bouton « Quitter » de l'interface).
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// read prend le verrou en lecture et retourne la fonction de libération,
// à utiliser sous la forme : defer a.read()()
func (a *App) read() func() {
	a.mu.RLock()
	return a.mu.RUnlock
}

// write prend le verrou exclusif (opérations qui remplacent la base).
func (a *App) write() func() {
	a.mu.Lock()
	return a.mu.Unlock
}

// GetAppInfo retourne la version, les chemins de travail et l'état de la base.
func (a *App) GetAppInfo() (AppInfo, error) {
	defer a.read()()
	empty, err := db.IsDatabaseEmpty()
	if err != nil {
		return AppInfo{}, err
	}
	return AppInfo{
		Version:      Version,
		DataDir:      a.baseDir,
		DBPath:       a.dbPath,
		DocumentsDir: a.documentsDir(),
		BackupsDir:   filepath.Join(a.baseDir, "backups"),
		Agent:        currentAgent(),
		IsEmpty:      empty,
	}, nil
}

// OpenPath ouvre un fichier (avec l'application par défaut) ou un dossier
// (dans l'Explorateur). Seuls les chemins situés dans le dossier de travail
// sont acceptés.
func (a *App) OpenPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(a.baseDir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("chemin hors du dossier de l'application")
	}
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("fichier introuvable : %s", abs)
	}
	// explorer.exe ouvre un dossier, ou un fichier avec son application par
	// défaut. Il retourne un code non nul même en cas de succès : on l'ignore.
	_ = exec.Command("explorer.exe", abs).Start()
	return nil
}

// OpenDocumentsFolder ouvre le dossier documents/ dans l'Explorateur.
func (a *App) OpenDocumentsFolder() error {
	dir := a.documentsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return a.OpenPath(dir)
}

// documentsDir retourne le dossier des PDF et CSV générés.
func (a *App) documentsDir() string {
	return filepath.Join(a.baseDir, "documents")
}

// absPath convertit un chemin relatif (documents/...) en chemin absolu.
func (a *App) absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(a.baseDir, p)
}

// chooseFile ouvre la boîte de dialogue native de sélection de fichier.
// Retourne "" si l'utilisateur annule.
func (a *App) chooseFile(title, filterName, pattern string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: []runtime.FileFilter{{DisplayName: filterName, Pattern: pattern}},
	})
}

// currentAgent retourne le nom de l'agent enregistré sur les prêts et les
// bons de remise : celui des Paramètres s'il est renseigné, sinon le nom de
// la session Windows (utile en multi-postes).
func currentAgent() string {
	if s, err := loadSettings(); err == nil && s.AgentName != "" {
		return s.AgentName
	}
	u, err := user.Current()
	if err != nil || u.Username == "" {
		return "Administrateur"
	}
	name := u.Username
	if i := strings.LastIndex(name, `\`); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// safeFileName remplace les caractères interdits dans un nom de fichier
// Windows (le nom d'un détenteur sert à nommer son bon de remise).
func safeFileName(s string) string {
	r := strings.NewReplacer(`\`, "-", "/", "-", ":", "-", "*", "", "?", "", `"`, "", "<", "", ">", "", "|", "-")
	s = strings.TrimSpace(r.Replace(s))
	s = strings.Join(strings.Fields(s), "_")
	if s == "" {
		return "sans_nom"
	}
	return s
}

// sameFile indique si deux chemins désignent le même fichier.
func sameFile(p1, p2 string) bool {
	s1, err1 := os.Stat(p1)
	s2, err2 := os.Stat(p2)
	return err1 == nil && err2 == nil && os.SameFile(s1, s2)
}

// friendlyDBError traduit les erreurs SQLite courantes en message lisible.
func friendlyDBError(err error, uniqueMsg string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return fmt.Errorf("%s", uniqueMsg)
	}
	return err
}
