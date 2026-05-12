# Analyse & Recommandations — Mise à Jour Majeure

> Analyse du code source Go de l'application **Gestionnaire de Clés** (v2.x).  
> Base : Fyne v2.4.5 · SQLite (modernc) · Go 1.21  
> Date d'analyse : 2026-05-11

---

## 1. Vue d'ensemble de l'architecture actuelle

```
cmd/main.go                  Point d'entrée
internal/
  db/
    database.go              Connexion SQLite, création du schéma
    models.go                Structs Go (Key, Room, Building, Borrower, Loan…)
    queries.go               Toutes les fonctions d'accès aux données
    backup.go                Sauvegarde / restauration / import Python / démo
  gui/
    app.go                   App struct, menu, navigation, dialogues communs
    dashboard_modern.go      Tableau de bord principal
    keys.go / borrowers.go   Vues CRUD clés & emprunteurs
    buildings.go / rooms.go  Vues CRUD bâtiments & salles
    loans.go                 Vues emprunts actifs & rapport
    config.go                Sauvegarde, import Python, démo, reset
    backups.go / keyplan.go  Gestion sauvegardes & plan de clés
    theme.go / theme_simple.go  Thème Fyne personnalisé
  pdf/
    generator.go             Génération des PDF (reçus, rapports)
    exporter.go              Sauvegarde fichier & génération noms
```

**Ce que fait l'application :** gestion d'un parc de clés physiques (bâtiments → salles → clés), prêts aux emprunteurs, rapports PDF, sauvegardes SQLite.

---

## 2. Points forts existants

- Transactions SQL correctement utilisées (CreateKey, UpdateKey, CreateMultipleLoans).
- `defer tx.Rollback()` systématiquement placé après `Begin()`.
- Schéma SQLite propre avec clés étrangères et index.
- Sauvegarde de sécurité automatique avant toute opération destructrice.
- Séparation db / gui / pdf assez claire.

---

## 3. Problèmes identifiés

### 3.1 Architecture & couplage

| # | Fichier | Problème |
|---|---------|----------|
| A1 | `db/database.go` | Variable globale `var DB *sql.DB` — impossible à tester, pas thread-safe pour multi-instance |
| A2 | `db/queries.go` | Toutes les fonctions accèdent à `DB` directement ; aucune interface, aucune injection de dépendance |
| A3 | `cmd/main.go:35` | `log.Fatalf` dans `Initialize()` (package gui) masque l'erreur retournée ; la fonction retourne toujours `nil` en cas d'erreur |
| A4 | `gui/app.go:278` | `refreshCurrentView()` redirige systématiquement vers le dashboard plutôt que de recharger la vue courante |
| A5 | `db/backup.go` | `BackupDatabase` fait une copie de fichier brute ; si SQLite écrit au même moment, la copie peut être corrompue (pas d'utilisation de l'API SQLite `VACUUM INTO` ou WAL checkpoint) |

### 3.2 Gestion des erreurs

| # | Fichier | Problème |
|---|---------|----------|
| E1 | `gui/keys.go:71-74` | Erreurs de `GetActiveLoansForKey` et `GetRoomsForKey` silencieusement ignorées (`_`) |
| E2 | `db/backup.go:454` | `fmt.Printf` utilisé pour logger le résumé d'import — devrait utiliser `log` ou être retourné |
| E3 | `db/queries.go` | `GetKeysWithAvailability` fait N+1 requêtes (1 par clé pour compter les emprunts + 1 par clé pour les emprunteurs) — pas d'erreur mais problème de performance |
| E4 | `pdf/exporter.go:28` | Nom de variable `filepath` écrase l'import standard `path/filepath` dans la même portée |

### 3.3 Sécurité & données

| # | Fichier | Problème |
|---|---------|----------|
| S1 | `db/database.go` | Pas de `PRAGMA foreign_keys = ON` — les contraintes FK ne sont pas activées par défaut dans SQLite |
| S2 | `db/database.go` | Pas de `PRAGMA journal_mode = WAL` — performances et résistance aux crashs sous-optimales |
| S3 | `db/queries.go:665` | `CreateMultipleLoans` vérifie la dispo via `CheckKeyAvailability` hors transaction — race condition possible si deux emprunts arrivent simultanément |
| S4 | `db/backup.go:272` | `ImportFromPythonDB` : plusieurs `rows` sont créés et `defer rows.Close()` est appelé en boucle dans la même fonction — tous les curseurs s'empilent et ne se ferment qu'à la fin de la fonction |

### 3.4 Qualité du code

| # | Fichier | Problème |
|---|---------|----------|
| Q1 | `db/queries.go:572-577` | `GetActiveLoanCount` et `GetKeyActiveLoanCount` sont identiques — doublon |
| Q2 | `db/queries.go:684` | `GetActiveLoansForKey` est un alias redondant de `GetActiveLoansByKeyID` |
| Q3 | `db/queries.go:503-541` | `GetKeysWithAvailability` — N+1 requêtes SQL, devrait être une seule requête avec `LEFT JOIN` et `COUNT` |
| Q4 | `pdf/generator.go` | Duplication des en-têtes de tableau en cas de changement de page (blocs répétés à 3 endroits) |
| Q5 | `go.mod` | Go 1.21, Fyne 2.4.5 — versions obsolètes, Go 1.23+ et Fyne 2.5+ disponibles |
| Q6 | Aucun fichier `*_test.go` | Zéro test dans tout le projet |

### 3.5 UX / Comportement

| # | Fichier | Problème |
|---|---------|----------|
| U1 | `gui/app.go:127-134` | `setContent` recrée le menu entier à chaque changement de vue — coûteux, provoque des flashs |
| U2 | `gui/loans.go:41` | L'itération sur une `map` pour afficher les emprunteurs est non-déterministe (ordre aléatoire à chaque affichage) |
| U3 | `pdf/generator.go:347` | `GenerateGlobalBorrowerReport` — même problème, la map n'est pas triée avant génération du PDF |
| U4 | `gui/config.go:258-293` | 3 confirmations imbriquées pour le reset — la logique est difficile à lire et à maintenir |

---

## 4. Recommandations pour la mise à jour majeure

### 4.1 Refactoring prioritaire (obligatoire)

**R1 — Supprimer la variable globale `DB`**  
Créer une struct `Repository` ou `Store` qui encapsule `*sql.DB` et implémenter une interface. Permet les tests unitaires et l'injection de dépendances.

```go
type Store interface {
    GetAllKeys() ([]Key, error)
    CreateKey(k *Key, roomIDs []int) error
    // ...
}

type SQLiteStore struct {
    db *sql.DB
}
```

**R2 — Activer les PRAGMA SQLite essentiels**  
Dans `InitDB`, ajouter immédiatement après `Ping()` :
```go
DB.Exec("PRAGMA foreign_keys = ON")
DB.Exec("PRAGMA journal_mode = WAL")
DB.Exec("PRAGMA synchronous = NORMAL")
```

**R3 — Remplacer GetKeysWithAvailability par une requête SQL unique**  
```sql
SELECT k.*, 
       COUNT(l.id) as loaned_count,
       (k.quantity_total - k.quantity_reserve - COUNT(l.id)) as available_count
FROM keys k
LEFT JOIN loans l ON l.key_id = k.id AND l.return_date IS NULL
GROUP BY k.id
```

**R4 — Corriger la sauvegarde SQLite**  
Utiliser `VACUUM INTO 'chemin/backup.db'` au lieu de la copie de fichier brute pour garantir l'intégrité.

**R5 — Corriger `ImportFromPythonDB`**  
Fermer chaque `rows` explicitement dans sa propre fonction ou sous-bloc, pas avec `defer` en cascade dans la même fonction.

### 4.2 Améliorations importantes

**R6 — Ajouter un logger structuré**  
Remplacer les `log.Printf` / `fmt.Printf` épars par `log/slog` (standard depuis Go 1.21) avec niveaux DEBUG/INFO/WARN/ERROR.

**R7 — Corriger l'ordre d'affichage des maps**  
Toutes les maps `loansByBorrower` et `loansByKey` doivent être triées par clé avant affichage (ui et pdf).

**R8 — Mettre à jour les dépendances**  
- `go 1.21` → `go 1.23`  
- `fyne.io/fyne/v2 v2.4.5` → `v2.5.x`  
- `modernc.org/sqlite v1.28.0` → dernière version stable  
- Vérifier et mettre à jour `golang.org/x/*`

**R9 — Supprimer les doublons de fonctions**  
- Fusionner `GetActiveLoanCount` et `GetKeyActiveLoanCount`
- Supprimer l'alias `GetActiveLoansForKey`

**R10 — Factoriser les en-têtes de tableau PDF**  
Extraire une fonction `writePDFTableHeader(pdf, columns)` réutilisable.

### 4.3 Nouveautés à envisager

**R11 — Historique complet des emprunts**  
Ajouter une vue "Historique" qui affiche tous les emprunts (actifs + retournés) avec filtres par date, emprunteur, clé. La table `loans` le supporte déjà.

**R12 — Recherche / filtrage**  
Ajouter une barre de recherche dans les vues Clés, Emprunteurs et Emprunts.

**R13 — Notifications d'emprunts longs**  
Mettre en évidence (couleur rouge) les emprunts actifs depuis plus de X jours (configurable).

**R14 — Tests unitaires**  
Ajouter au minimum des tests sur la couche `db` avec une base SQLite en mémoire (`:memory:`).

```go
func TestCreateKey(t *testing.T) {
    store := newTestStore(t) // ouvre :memory:
    // ...
}
```

**R15 — Export CSV**  
Ajouter un export CSV en complément du PDF pour les rapports (utile pour Excel/Google Sheets).

---

## 5. Ordre de priorité suggéré

| Priorité | Items | Impact |
|----------|-------|--------|
| 🔴 Critique | R2 (PRAGMA FK), R4 (backup SQLite), R5 (import rows) | Intégrité des données |
| 🟠 Haute | R1 (suppr. global DB), R3 (N+1 queries), R7 (tri maps) | Fiabilité & perf |
| 🟡 Moyenne | R6 (slog), R8 (deps), R9 (doublons), R10 (PDF) | Qualité code |
| 🟢 Basse | R11 à R15 (nouvelles fonctionnalités) | UX |

---

## 6. Fichiers à créer / renommer

| Action | Fichier | Raison |
|--------|---------|--------|
| Créer | `internal/db/store.go` | Interface `Store` pour l'injection de dépendances |
| Créer | `internal/db/store_sqlite.go` | Implémentation SQLite de l'interface |
| Créer | `internal/db/store_test.go` | Tests sur base `:memory:` |
| Supprimer | `internal/gui/dashboard.go` | Fichier dashboard obsolète (remplacé par `dashboard_modern.go`) |
| Supprimer | `internal/gui/dashboard_improved.go` | Idem, version intermédiaire abandonnée |
| Supprimer | `internal/gui/theme.go` | Si `theme_simple.go` est la version finale |

---

*Ce fichier sera complété par un `TODO.md` listant les tâches concrètes de mise en œuvre.*
