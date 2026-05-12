# Plan technique détaillé — V3 Gestionnaire de Clés
## Collège Victor Hugo, Chauny

> Basé sur le TODO.md (cahier des charges v5) et l'analyse du code V2 (MAJ.md).  
> Objectif : V3 opérationnelle sans réécriture totale, évolution incrémentale de la V2.  
> Date : 2026-05-11

---

## Table des matières

1. [Architecture générale](#1-architecture-générale)
2. [Schéma de base de données](#2-schéma-de-base-de-données)
3. [Multi-postes simultanés](#3-multi-postes-simultanés)
4. [Couche données — refactoring + nouvelles queries](#4-couche-données)
5. [Logique métier — algorithme de prêt par besoin](#5-logique-métier)
6. [Interface graphique — vues nouvelles et modifiées](#6-interface-graphique)
7. [Export CSV](#7-export-csv)
8. [PDF — bon de remise enrichi](#8-pdf)
9. [Corrections critiques héritées de la V2](#9-corrections-critiques-héritées-de-la-v2)
10. [Plan de migration V2 → V3](#10-plan-de-migration-v2--v3)
11. [Ordre d'exécution des tâches](#11-ordre-dexécution-des-tâches)

---

## 1. Architecture générale

### Structure de fichiers cible

```
cmd/
  main.go                    Inchangé (ajout login au démarrage)

internal/
  db/
    store.go                 Interface Store (nouvelle)
    store_sqlite.go          Implémentation SQLiteStore (nouvelle)
    database.go              InitDB modifié (PRAGMA, WAL, connexion pool)
    models.go                Structs mis à jour (Access enrichi, Loan étendu)
    migrations.go            Gestion des migrations de schéma (nouvelle)
    backup.go                BackupDatabase → VACUUM INTO
    store_test.go            Tests sur :memory: (nouvelle)

  business/
    loan_advisor.go          Algorithme combinaison minimale de clés (nouvelle)
    redundancy.go            Détection redondances d'accès (nouvelle)

  gui/
    app.go                   Ajout : état utilisateur connecté, refresh correct
    dashboard_modern.go      Mis à jour : retards, redondances
    keys.go                  Mis à jour : filtres bâtiment/étage/catégorie
    accesses.go              Nouvelle vue : gestion des accès enrichis
    borrowers.go             Mis à jour : statut détenteur
    loans.go                 Mis à jour : prêt par besoin, historique
    loan_wizard.go           Nouveau : assistant de prêt en 3 étapes
    history.go               Nouvelle vue : historique filtrable
    views_transversal.go     Nouvelles vues transversales (qui a quoi, etc.)
    redundancy_view.go       Vue redondances d'accès
    login.go                 Identification simple par nom (nouvelle)
    config.go                Inchangé + export CSV
    [supprimé] dashboard.go
    [supprimé] dashboard_improved.go

  pdf/
    generator.go             Mis à jour : bon de remise enrichi
    exporter.go              Correction nom de variable filepath

  export/
    csv.go                   Export CSV/Excel-compatible (nouvelle)
```

### Dépendances à ajouter

```go
// go.mod — ajouts
go 1.23

// Pas de dépendances extérieures ajoutées : tout se fait avec la stdlib
// sauf si export .xlsx souhaité → "github.com/xuri/excelize/v2"
// Pour l'instant : CSV UTF-8 avec BOM (lisible par Excel directement)
```

---

## 2. Schéma de base de données

### 2.1 Tables existantes à modifier

**Table `rooms` → renommée `accesses`** (ou ajout de colonnes sur `rooms`)

Choix retenu : **garder le nom `rooms` et ajouter les colonnes** pour éviter une migration complexe et rester compatible avec la logique V2.

```sql
ALTER TABLE rooms ADD COLUMN floor TEXT;        -- étage/niveau (RDC, R+1, Sous-sol...)
ALTER TABLE rooms ADD COLUMN category TEXT;     -- salle de classe, local technique, bureau...
ALTER TABLE rooms ADD COLUMN notes TEXT;        -- accès sensible, PMR, alarme, etc.
```

**Table `borrowers` — ajout colonne statut**

```sql
ALTER TABLE borrowers ADD COLUMN status TEXT DEFAULT 'permanent';
-- Valeurs : 'permanent', 'contractuel', 'intervenant', 'entreprise'
ALTER TABLE borrowers ADD COLUMN phone TEXT;
```

**Table `loans` — ajout colonnes**

```sql
ALTER TABLE loans ADD COLUMN planned_return_date DATETIME;  -- date retour prévue
ALTER TABLE loans ADD COLUMN loan_type TEXT DEFAULT 'ponctuel';
-- Valeurs : 'ponctuel', 'permanent'
ALTER TABLE loans ADD COLUMN returned_condition TEXT;       -- état constaté au retour
ALTER TABLE loans ADD COLUMN created_by TEXT;               -- nom de l'agent qui a fait le prêt
```

**Table `keys` — ajout colonne catégorie**

```sql
ALTER TABLE keys ADD COLUMN category TEXT DEFAULT 'simple';
-- Valeurs : 'simple', 'trousseau', 'badge', 'passe'
ALTER TABLE keys ADD COLUMN notes TEXT;
```

### 2.2 Gestion des migrations

Fichier `internal/db/migrations.go` :

```go
// Principe : table schema_version stocke la version actuelle
// À chaque InitDB, on vérifie la version et on applique les migrations manquantes

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

// Migration 1 (V2 baseline) : création initiale — déjà en place
// Migration 2 (V3) : ajout des colonnes ci-dessus
```

Implémentation :

```go
type Migration struct {
    Version     int
    Description string
    SQL         string
}

var migrations = []Migration{
    {1, "baseline V2", ``}, // déjà appliquée si la DB existe
    {2, "V3 enrichissement accès, détenteurs, prêts", `
        ALTER TABLE rooms ADD COLUMN floor TEXT;
        ALTER TABLE rooms ADD COLUMN category TEXT;
        ALTER TABLE rooms ADD COLUMN notes TEXT;
        ALTER TABLE borrowers ADD COLUMN status TEXT DEFAULT 'permanent';
        ALTER TABLE borrowers ADD COLUMN phone TEXT;
        ALTER TABLE loans ADD COLUMN planned_return_date DATETIME;
        ALTER TABLE loans ADD COLUMN loan_type TEXT DEFAULT 'ponctuel';
        ALTER TABLE loans ADD COLUMN returned_condition TEXT;
        ALTER TABLE loans ADD COLUMN created_by TEXT;
        ALTER TABLE keys ADD COLUMN category TEXT DEFAULT 'simple';
        ALTER TABLE keys ADD COLUMN notes TEXT;
    `},
}

func applyMigrations(db *sql.DB) error {
    // 1. Créer schema_version si absent
    // 2. Lire version actuelle
    // 3. Appliquer les migrations manquantes dans une transaction
}
```

> **Important :** SQLite ne supporte pas `ADD COLUMN` avec contraintes complexes dans un `ALTER TABLE`. Les colonnes ajoutées doivent être nullable ou avoir une DEFAULT. C'est le cas ici.

### 2.3 PRAGMA à activer dans InitDB

```go
pragmas := []string{
    "PRAGMA foreign_keys = ON",
    "PRAGMA journal_mode = WAL",
    "PRAGMA synchronous = NORMAL",
    "PRAGMA busy_timeout = 5000",   // Clé pour le multi-postes : attend 5s avant erreur
    "PRAGMA cache_size = -4000",    // 4 Mo de cache
}
```

---

## 3. Multi-postes simultanés

### Analyse du problème

SQLite en mode WAL (`journal_mode = WAL`) supporte nativement :
- **Lectures simultanées** : illimitées, sans blocage
- **Écritures simultanées** : une seule écriture à la fois, les autres attendent

Avec `busy_timeout = 5000`, si un poste tente d'écrire pendant qu'un autre écrit, SQLite attend jusqu'à 5 secondes avant de retourner une erreur `SQLITE_BUSY`. C'est suffisant pour un collège (les opérations durent < 100ms).

### Solution retenue : WAL + busy_timeout (option légère)

C'est la solution recommandée pour ce contexte (3-4 postes max, opérations courtes). Pas besoin de client/serveur.

**Conditions pour que ça fonctionne :**
1. Le fichier `.db` est sur un partage réseau (SMB/CIFS) — **attention** : SQLite sur réseau avec WAL peut poser problème selon le serveur de fichiers.
2. Solution de contournement si problème réseau : utiliser `locking_mode = EXCLUSIVE` + retry applicatif.

### Implémentation dans `store_sqlite.go`

```go
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := sql.Open("sqlite", dbPath)
    // ...
    // Pool de connexions : 1 seule connexion active pour éviter les deadlocks internes
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    // ...
}
```

> `SetMaxOpenConns(1)` est crucial : avec SQLite, plusieurs connexions depuis le même process peuvent se bloquer mutuellement. Une seule connexion par process, WAL gère le reste entre les process.

### Gestion de l'erreur SQLITE_BUSY dans le code

```go
func retryOnBusy(fn func() error) error {
    const maxRetries = 3
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        if !isBusyError(err) {
            return err
        }
        time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
    }
    return fmt.Errorf("base de données occupée, réessayez dans quelques instants")
}
```

---

## 4. Couche données

### 4.1 Interface Store

Fichier `internal/db/store.go` :

```go
type Store interface {
    // Accès (ex-rooms)
    GetAllAccesses() ([]Room, error)
    GetAccessesByBuilding(buildingID int) ([]Room, error)
    GetAccessesByFloor(buildingID int, floor string) ([]Room, error)
    GetAccessesByCategory(category string) ([]Room, error)
    CreateAccess(r *Room) error
    UpdateAccess(r *Room) error
    DeleteAccess(id int) error

    // Clés
    GetAllKeys() ([]Key, error)
    GetKeysWithAvailability() ([]KeyWithAvailability, error)
    GetKeysByBuilding(buildingID int) ([]Key, error)
    GetKeysForAccess(accessID int) ([]Key, error)
    GetAvailableKeysForAccesses(accessIDs []int) ([]Key, error)  // NOUVEAU
    CreateKey(k *Key, accessIDs []int) error
    UpdateKey(k *Key, accessIDs []int) error
    DeleteKey(id int) error
    GetKeyHistory(keyID int) ([]LoanWithDetails, error)  // NOUVEAU

    // Détenteurs
    GetAllBorrowers() ([]Borrower, error)
    GetBorrowerWithCurrentKeys(id int) (*BorrowerWithKeys, error)  // NOUVEAU
    CreateBorrower(b *Borrower) error
    UpdateBorrower(b *Borrower) error
    DeleteBorrower(id int) error
    GetBorrowerHistory(borrowerID int) ([]LoanWithDetails, error)  // NOUVEAU

    // Prêts
    GetAllActiveLoans() ([]LoanWithDetails, error)
    GetOverdueLoans() ([]LoanWithDetails, error)  // NOUVEAU
    CreateLoan(keyIDs []int, borrowerID int, plannedReturn *time.Time, loanType string, createdBy string) error
    ReturnLoan(loanID int, condition string) error
    GetLoanHistory(filters LoanFilters) ([]LoanWithDetails, error)  // NOUVEAU

    // Bâtiments
    GetAllBuildings() ([]Building, error)
    CreateBuilding(b *Building) error
    UpdateBuilding(b *Building) error
    DeleteBuilding(id int) error

    // Redondances
    GetBorrowersWithRedundantAccesses() ([]RedundancyReport, error)  // NOUVEAU
}
```

### 4.2 Requête GetKeysWithAvailability (correction N+1)

```sql
SELECT
    k.id, k.number, k.description, k.quantity_total,
    k.quantity_reserve, k.storage_location, k.category, k.notes,
    COUNT(l.id) AS loaned_count,
    (k.quantity_total - k.quantity_reserve - COUNT(l.id)) AS available_count,
    GROUP_CONCAT(b.name, ', ') AS borrower_names
FROM keys k
LEFT JOIN loans l ON l.key_id = k.id AND l.return_date IS NULL
LEFT JOIN borrowers b ON l.borrower_id = b.id
GROUP BY k.id
ORDER BY k.number
```

Une seule requête remplace les 2N+1 requêtes actuelles.

### 4.3 GetAvailableKeysForAccesses (cœur du prêt par besoin)

```sql
-- Clés qui ouvrent AU MOINS UN des accès demandés ET qui sont disponibles
SELECT DISTINCT
    k.id, k.number, k.description, k.quantity_total, k.quantity_reserve,
    k.storage_location, k.category,
    (k.quantity_total - k.quantity_reserve - COUNT(l.id)) AS available_count,
    -- Accès couverts par cette clé (parmi ceux demandés)
    GROUP_CONCAT(kra.room_id) AS covered_access_ids
FROM keys k
INNER JOIN key_room_association kra ON kra.key_id = k.id
    AND kra.room_id IN (/* liste des accessIDs */)
LEFT JOIN loans l ON l.key_id = k.id AND l.return_date IS NULL
GROUP BY k.id
HAVING available_count > 0
ORDER BY k.number
```

### 4.4 Nouveaux modèles

```go
// models.go — ajouts

type LoanFilters struct {
    BorrowerID  *int
    KeyID       *int
    BuildingID  *int
    DateFrom    *time.Time
    DateTo      *time.Time
    Status      string  // "active", "returned", "overdue", ""
}

type RedundancyReport struct {
    Borrower        Borrower
    Keys            []Key
    RedundantAccesses []Room  // accès couverts par plusieurs clés du même détenteur
}

type BorrowerWithKeys struct {
    Borrower
    ActiveLoans     []LoanWithDetails
    CoveredAccesses []Room  // union de tous les accès couverts
    Redundancies    []Room  // accès en double
}
```

---

## 5. Logique métier

### 5.1 Algorithme de combinaison minimale de clés

Fichier `internal/business/loan_advisor.go`

**Problème :** étant donné une liste d'accès requis et un ensemble de clés disponibles, trouver la combinaison de clés **minimale** (en nombre) qui couvre tous les accès requis.

C'est le problème du **Set Cover** (NP-difficile en général). Avec ~50 clés et ~50 accès, une approche **greedy** est suffisante et rapide :

```go
// SuggestKeys retourne la combinaison minimale de clés disponibles
// pour couvrir tous les accès demandés.
// Algorithme greedy : à chaque étape, choisir la clé disponible
// qui couvre le plus d'accès non encore couverts.
func SuggestKeys(requiredAccessIDs []int, availableKeys []KeyWithCoverage) ([]Key, []int, error) {
    remaining := toSet(requiredAccessIDs)
    var selectedKeys []Key

    for len(remaining) > 0 {
        // Trouver la clé qui couvre le plus d'accès restants
        bestKey, bestCoverage := findBestKey(availableKeys, remaining)
        if bestKey == nil || len(bestCoverage) == 0 {
            // Certains accès ne peuvent pas être couverts
            return selectedKeys, setToSlice(remaining), ErrAccessesUncoverable
        }
        selectedKeys = append(selectedKeys, *bestKey)
        for id := range bestCoverage {
            delete(remaining, id)
        }
        // Retirer la clé sélectionnée des disponibles
        availableKeys = removeKey(availableKeys, bestKey.ID)
    }

    return selectedKeys, nil, nil
}
```

**Cas limites à gérer :**
- Accès non couvrable par aucune clé disponible → afficher un avertissement, pas une erreur bloquante
- Clé qui couvre plus d'accès que nécessaire (passe) → proposée en dernier recours si elle est la seule option
- Plusieurs clés ex-æquo → privilégier celle avec le moins d'accès "bonus" (minimiser la surface d'accès)

### 5.2 Détection des redondances

Fichier `internal/business/redundancy.go`

```go
// DetectRedundancies retourne les accès couverts par plus d'une clé
// chez le même détenteur
func DetectRedundancies(keys []Key, accessesByKey map[int][]Room) []Room {
    accessCount := make(map[int]int)  // accessID → nombre de clés qui l'ouvrent
    accessMap := make(map[int]Room)

    for _, key := range keys {
        for _, access := range accessesByKey[key.ID] {
            accessCount[access.ID]++
            accessMap[access.ID] = access
        }
    }

    var redundant []Room
    for id, count := range accessCount {
        if count > 1 {
            redundant = append(redundant, accessMap[id])
        }
    }
    return redundant
}
```

---

## 6. Interface graphique

### 6.1 Identification simple au démarrage

Fichier `internal/gui/login.go`

- Fenêtre modale au lancement : "Qui êtes-vous ?" + liste déroulante des agents (ou saisie libre)
- Pas de mot de passe
- Le nom sélectionné est stocké dans `App.currentUser string`
- Utilisé pour remplir `created_by` dans les prêts
- Conservé en mémoire pendant la session

```go
func showLoginDialog(a *App, onConfirm func(username string)) {
    // Widget Select avec les noms connus + option "Autre..."
    // Si "Autre..." → champ texte libre
}
```

### 6.2 Vue Accès enrichis

Fichier `internal/gui/accesses.go` (remplace `rooms.go`)

Colonnes affichées : Désignation | Bâtiment | Étage | Catégorie | Nb clés associées

Filtres en haut de vue :
- Select bâtiment (tous / bâtiment X)
- Select étage (tous / RDC / R+1 / ...)
- Select catégorie (tous / salle de classe / local technique / ...)

### 6.3 Assistant de prêt (loan_wizard.go)

3 étapes dans une fenêtre modale :

**Étape 1 — Détenteur**
- Recherche par nom dans la liste des détenteurs
- Bouton "Nouveau détenteur" si absent

**Étape 2 — Accès requis**
- Liste des accès avec cases à cocher
- Filtres bâtiment/étage/catégorie pour faciliter la sélection
- Date de retour prévue (optionnelle, champ date)
- Type de prêt : ponctuel / permanent

**Étape 3 — Proposition de clés**
- Affichage automatique de la combinaison minimale calculée par `loan_advisor.go`
- Possibilité de modifier manuellement (ajouter/retirer des clés)
- Si des accès ne peuvent pas être couverts → bandeau d'avertissement orange
- Si redondance détectée → bandeau d'avertissement jaune
- Bouton "Valider et imprimer le bon"

```go
type LoanWizardState struct {
    Step           int
    BorrowerID     int
    SelectedAccesses []int
    SuggestedKeys  []db.Key
    FinalKeys      []db.Key
    PlannedReturn  *time.Time
    LoanType       string
}
```

### 6.4 Vue Historique

Fichier `internal/gui/history.go`

- Tableau scrollable de tous les prêts (actifs + retournés)
- Filtres :
  - Période (date début / date fin) — widget DatePicker Fyne
  - Clé (select)
  - Détenteur (select)
  - Bâtiment (select)
  - Statut (tous / en cours / retourné / en retard)
- Colonnes : Clé | Détenteur | Date remise | Date retour prévue | Date retour réelle | Durée | État | Agent

> Note Fyne : il n'y a pas de widget DatePicker natif en Fyne 2.4. Utiliser 3 `widget.Select` (jour/mois/année) ou un champ texte avec format `DD/MM/YYYY` + parsing.

### 6.5 Vues transversales

Fichier `internal/gui/views_transversal.go`

Accessible depuis un nouveau bouton "Vues" dans le menu :

- **"Qui a quoi ?"** : tableau détenteur → clés actuelles → accès couverts + signal rouge si redondance
- **"Quelle clé pour quelle porte ?"** : sélectionner un accès → liste des clés qui l'ouvrent + disponibilité
- **"Quelles clés dans ce bâtiment ?"** : sélectionner bâtiment → toutes les clés concernées
- **"Clés disponibles"** : tableau stock / sorties / disponibles filtrable par bâtiment/catégorie

### 6.6 Tableau de bord — modifications

`dashboard_modern.go` — ajouts :

- Carte "Prêts en retard" (rouge, visible en permanence si > 0)
- Carte "Redondances détectées" (orange, cliquable → vue redondances)
- Correction : tri déterministe dans `createSimpleKeysTable`

### 6.7 Correction refresh de vue

Dans `app.go`, remplacer le mécanisme actuel par un système de vue courante :

```go
type App struct {
    // ...
    currentView string  // "dashboard", "keys", "loans", etc.
}

func (a *App) refreshCurrentView() {
    switch a.currentView {
    case "dashboard":   a.showDashboard()
    case "keys":        a.showKeys()
    case "loans":       a.showActiveLoans()
    // ...
    }
}

func (a *App) setContent(content fyne.CanvasObject, viewName string) {
    a.currentView = viewName
    // ...
}
```

---

## 7. Export CSV

Fichier `internal/export/csv.go`

CSV avec BOM UTF-8 pour compatibilité Excel sans configuration :

```go
func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
    // BOM UTF-8
    w.Write([]byte{0xEF, 0xBB, 0xBF})
    // Séparateur point-virgule (standard Excel France)
    writer := csv.NewWriter(w)
    writer.Comma = ';'
    writer.Write(headers)
    for _, row := range rows {
        writer.Write(row)
    }
    writer.Flush()
    return writer.Error()
}
```

Exports disponibles :
- Inventaire des clés (avec stock, dispo, localisation)
- Emprunts en cours
- Historique complet (avec filtres appliqués)
- Liste des détenteurs

Bouton "Exporter CSV" ajouté dans chaque vue concernée.

---

## 8. PDF

### 8.1 Bon de remise enrichi

`generator.go` — `GenerateLoanReceipt` mis à jour :

Ajouts au PDF :
- Date de retour prévue (si prêt ponctuel)
- **Tableau des accès couverts** : liste des portes/zones accessibles avec les clés remises
- Nom de l'agent ayant effectué la remise

Structure du document :
```
[En-tête : logo/nom établissement si configuré]
BON DE REMISE DE CLÉ(S)

Détenteur : ___________    Date remise : ___________
Type : Prêt ponctuel       Retour prévu : ___________
Agent : ___________

Clés remises :
┌──────────┬──────────────────────────┐
│ N° clé   │ Désignation              │
├──────────┼──────────────────────────┤
│ K001     │ Clé principale ...       │
└──────────┴──────────────────────────┘

Accès couverts :
┌──────────────────────────┬─────────────┬───────┐
│ Désignation              │ Bâtiment    │ Étage │
├──────────────────────────┼─────────────┼───────┤
│ Salle B12               │ Bâtiment B  │ R+1   │
└──────────────────────────┴─────────────┴───────┘

Je soussigné(e), __________, reconnais avoir reçu...

Signature : ___________________________
```

### 8.2 Factorisation des en-têtes de tableau PDF

```go
type PDFTableColumn struct {
    Header string
    Width  float64
    Align  string
}

func writePDFTableHeader(pdf *gofpdf.Fpdf, columns []PDFTableColumn) {
    pdf.SetFont("Arial", "B", 10)
    pdf.SetFillColor(200, 220, 255)
    for _, col := range columns {
        pdf.CellFormat(col.Width, 8, pdf.UnicodeTranslatorFromDescriptor("")(col.Header),
            "1", 0, col.Align, true, 0, "")
    }
    pdf.Ln(8)
    pdf.SetFont("Arial", "", 9)
}
```

---

## 9. Corrections critiques héritées de la V2

Ces corrections doivent être faites **en premier**, avant toute nouvelle fonctionnalité.

| # | Fichier | Correction |
|---|---------|-----------|
| C1 | `db/database.go` | Ajouter les 5 PRAGMA (FK, WAL, synchronous, busy_timeout, cache) |
| C2 | `db/backup.go` | Remplacer copie de fichier par `VACUUM INTO` |
| C3 | `db/backup.go` | Corriger `ImportFromPythonDB` : fermer chaque `rows` dans son propre sous-bloc |
| C4 | `db/queries.go` | Supprimer `GetActiveLoanCount` (doublon de `GetKeyActiveLoanCount`) |
| C5 | `db/queries.go` | Supprimer `GetActiveLoansForKey` (alias de `GetActiveLoansByKeyID`) |
| C6 | `pdf/exporter.go` | Renommer la variable locale `filepath` en `filePath` |
| C7 | `gui/app.go:289` | Supprimer le `log.Fatalf` dans `Initialize` (masque l'erreur retournée) |
| C8 | `gui/loans.go:41` | Trier les maps `loansByBorrower` et `loansByKey` avant affichage |
| C9 | `gui/keys.go:71-74` | Gérer les erreurs de `GetActiveLoansForKey` et `GetRoomsForKey` |
| C10 | Fichiers obsolètes | Supprimer `dashboard.go`, `dashboard_improved.go` |

### C2 — VACUUM INTO (remplacement backup)

```go
func BackupDatabase(dbPath string, backupPath string) error {
    // S'assurer que le répertoire destination existe
    if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
        return fmt.Errorf("impossible de créer le répertoire de sauvegarde: %w", err)
    }
    // VACUUM INTO copie proprement la DB pendant qu'elle est ouverte
    _, err := DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
    if err != nil {
        return fmt.Errorf("erreur lors de la sauvegarde: %w", err)
    }
    return nil
}
```

> **Note :** `VACUUM INTO` est disponible depuis SQLite 3.27.0 (2019). La version de `modernc.org/sqlite` utilisée le supporte.

---

## 10. Plan de migration V2 → V3

Le TODO.md stipule "aucune migration prévue, base vierge" comme point de départ. Cependant :
- La logique de migration de schéma doit exister pour les évolutions futures.
- Un import optionnel d'une sauvegarde V2 est prévu pour les utilisateurs qui souhaitent récupérer leurs données.

### 10.1 Démarrage sur base vierge (cas normal)

1. `InitDB` crée toutes les tables avec le schéma V3 complet (incluant les nouvelles colonnes)
2. `applyMigrations` enregistre la version 2 comme appliquée d'emblée

### 10.2 Migration automatique d'une base V2 existante

Si une base V2 est ouverte directement avec la V3 (ex : l'utilisateur pointe `clefs.db` de la V2) :
1. `applyMigrations` détecte que la table `schema_version` est absente → version 1
2. Applique la migration 2 (`ALTER TABLE`) dans une transaction
3. Toutes les données V2 sont conservées, les nouvelles colonnes ont leurs valeurs `DEFAULT`

### 10.3 Import d'une sauvegarde V2 vers une base V3 vierge

Cas d'usage : l'établissement a des données en V2 et veut démarrer la V3 proprement.

**Fichier :** `internal/db/import_v2.go`

**Fonction :** `ImportFromV2(v2DBPath string) error`

Différence avec l'`ImportFromPythonDB` existant : cette fonction gère les colonnes absentes en V2 (les nouvelles colonnes V3 sont remplies avec leurs valeurs `DEFAULT`).

```go
func ImportFromV2(v2DBPath string) error {
    // 1. Vérifier que le fichier source est bien une base V2
    //    (vérifier la présence des tables attendues : keys, loans, borrowers, rooms, buildings)
    if err := validateV2Schema(v2DBPath); err != nil {
        return fmt.Errorf("ce fichier ne semble pas être une base V2 valide: %w", err)
    }

    // 2. Sauvegarde de sécurité de la base V3 actuelle avant import
    backupPath := GetDefaultBackupPath(currentDBPath)
    if err := BackupDatabase(currentDBPath, backupPath); err != nil {
        return fmt.Errorf("impossible de sauvegarder la base actuelle: %w", err)
    }

    // 3. Ouvrir la base V2 en lecture seule
    v2DB, err := sql.Open("sqlite", "file:"+v2DBPath+"?mode=ro")
    // ...

    // 4. Importer dans une transaction sur la base V3
    tx, err := DB.Begin()
    defer tx.Rollback()

    // Ordre d'import : buildings → rooms → keys → key_room_association → borrowers → loans
    // Pour chaque table : INSERT OR IGNORE (les IDs V2 sont conservés)

    // buildings : colonnes identiques V2/V3
    importBuildings(tx, v2DB)

    // rooms : V2 n'a pas floor/category/notes → NULL accepté (colonnes nullable)
    importRooms(tx, v2DB)

    // keys : V2 n'a pas category/notes → DEFAULT 'simple' / NULL
    importKeys(tx, v2DB)

    // key_room_association : identique V2/V3
    importKeyRoomAssociations(tx, v2DB)

    // borrowers : V2 n'a pas status/phone → DEFAULT 'permanent' / NULL
    importBorrowers(tx, v2DB)

    // loans : V2 n'a pas planned_return_date/loan_type/returned_condition/created_by
    //         → NULL / DEFAULT 'ponctuel' / NULL / 'Import V2'
    importLoans(tx, v2DB)

    return tx.Commit()
}

// validateV2Schema vérifie que les tables V2 attendues sont présentes
func validateV2Schema(dbPath string) error {
    db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
    // ...
    expectedTables := []string{"keys", "loans", "borrowers", "rooms", "buildings", "key_room_association"}
    // vérifier chaque table avec : SELECT name FROM sqlite_master WHERE type='table' AND name=?
}
```

**Comportement en cas d'erreur partielle :**
- Si une table V2 est manquante ou corrompue → rollback complet + message explicite
- La sauvegarde de sécurité créée à l'étape 2 permet de revenir à l'état précédent

**Intégration dans l'UI :**

Dans `gui/config.go`, ajouter un bouton dédié dans la section import :

```
[📥 Importer depuis sauvegarde V2]
```

- Ouvre un sélecteur de fichier `.db`
- Affiche un résumé avant confirmation :
  ```
  Base V2 détectée :
  • 47 clés
  • 23 emprunteurs
  • 12 emprunts actifs
  • 156 emprunts historiques

  Les champs nouveaux en V3 (étage, catégorie, statut détenteur)
  seront vides et à compléter manuellement.

  Continuer ?
  ```
- Après import réussi : afficher le nombre d'enregistrements importés par table

---

## 11. Ordre d'exécution des tâches

### Phase 1 — Fondations (ne pas toucher à l'UI)

1. Corriger C1 à C10 (corrections critiques V2)
2. Créer `internal/db/store.go` (interface) + `store_sqlite.go` (implémentation)
3. Créer `internal/db/migrations.go` + schéma V3 complet
4. Mettre à jour `models.go` (nouveaux champs, nouveaux types)
5. Implémenter les nouvelles queries (GetKeysWithAvailability en 1 requête, GetOverdueLoans, GetKeyHistory, etc.)
6. Créer `internal/db/import_v2.go` (`ImportFromV2` + `validateV2Schema`)
7. Tester la couche DB avec `store_test.go` sur `:memory:` (inclure un test d'import V2)

### Phase 2 — Logique métier

7. Implémenter `business/loan_advisor.go` (algorithme greedy set cover)
8. Implémenter `business/redundancy.go`
9. Écrire les tests unitaires pour ces deux modules

### Phase 3 — Interface graphique

10. Mettre à jour `gui/app.go` (currentUser, refresh correct, login)
11. Créer `gui/login.go`
12. Mettre à jour `gui/accesses.go` (filtres bâtiment/étage/catégorie)
13. Créer `gui/loan_wizard.go` (assistant en 3 étapes)
14. Mettre à jour `gui/loans.go` (retour avec état constaté)
15. Créer `gui/history.go`
16. Créer `gui/views_transversal.go`
17. Mettre à jour `gui/dashboard_modern.go` (retards, redondances)
18. Supprimer les fichiers GUI obsolètes

### Phase 4 — Export & PDF

19. Créer `internal/export/csv.go`
20. Mettre à jour `pdf/generator.go` (bon de remise enrichi, factorisation)
21. Corriger `pdf/exporter.go` (variable `filepath`)

### Phase 5 — Multi-postes & finalisation

22. Valider le comportement WAL + busy_timeout sur partage réseau (test terrain)
23. Mettre à jour `go.mod` (Go 1.23, Fyne 2.5.x)
24. Mettre à jour le pipeline GitHub Actions pour la compilation Windows
25. Mettre à jour le README

---

## Résumé des fichiers à créer / modifier / supprimer

| Action | Fichier |
|--------|---------|
| **Créer** | `internal/db/store.go` |
| **Créer** | `internal/db/store_sqlite.go` |
| **Créer** | `internal/db/migrations.go` |
| **Créer** | `internal/db/import_v2.go` (ImportFromV2, validateV2Schema) |
| **Créer** | `internal/db/store_test.go` |
| **Créer** | `internal/business/loan_advisor.go` |
| **Créer** | `internal/business/redundancy.go` |
| **Créer** | `internal/export/csv.go` |
| **Créer** | `internal/gui/login.go` |
| **Créer** | `internal/gui/loan_wizard.go` |
| **Créer** | `internal/gui/history.go` |
| **Créer** | `internal/gui/views_transversal.go` |
| **Créer** | `internal/gui/redundancy_view.go` |
| **Modifier** | `internal/db/database.go` (PRAGMA) |
| **Modifier** | `internal/db/models.go` (nouveaux champs) |
| **Modifier** | `internal/db/queries.go` (N+1 fix, doublons supprimés) |
| **Modifier** | `internal/db/backup.go` (VACUUM INTO, fix rows) |
| **Modifier** | `internal/gui/config.go` (bouton "Importer depuis sauvegarde V2") |
| **Modifier** | `internal/gui/app.go` (currentUser, refresh, login) |
| **Modifier** | `internal/gui/accesses.go` (ex rooms.go — filtres) |
| **Modifier** | `internal/gui/borrowers.go` (statut détenteur) |
| **Modifier** | `internal/gui/loans.go` (état retour) |
| **Modifier** | `internal/gui/dashboard_modern.go` (retards, redondances) |
| **Modifier** | `internal/pdf/generator.go` (bon enrichi, factorisation) |
| **Modifier** | `internal/pdf/exporter.go` (fix variable filepath) |
| **Modifier** | `cmd/main.go` (login au démarrage) |
| **Supprimer** | `internal/gui/dashboard.go` |
| **Supprimer** | `internal/gui/dashboard_improved.go` |

---

*Ce plan sera mis à jour au fil de l'avancement. Les tâches concrètes sont à suivre dans `TODO.md`.*
