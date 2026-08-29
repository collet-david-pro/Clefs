# NOTICE TECHNIQUE — Gestionnaire de Clés (V3)

> Document à jour pour la version 3.3.0.

> **Note importante : code intégralement généré par IA.**
> L'ensemble de cette application (architecture, code Go, requêtes SQL,
> interface graphique, tests et documentation) a été **produit par une
> intelligence artificielle** (Claude, Anthropic), sous la direction et la
> relecture de **COLLET David** (Secrétaire Général, Collège Victor Hugo,
> Chauny). Ce document décrit le fonctionnement technique pour permettre à un
> développeur de reprendre, maintenir ou faire évoluer le projet.

Cette notice s'adresse à un **développeur familier du web** (JavaScript/PHP/Python,
SQL, MVC) mais pas forcément de Go ni des applications de bureau. Les concepts
sont donc systématiquement reliés à leurs équivalents web.

---

## 1. Vue d'ensemble

Le **Gestionnaire de Clés** est une application de **bureau** (desktop), pas un
site web. Il n'y a ni serveur HTTP, ni navigateur, ni client/serveur : c'est un
**exécutable unique** qui ouvre une fenêtre et lit/écrit dans un fichier de base
de données local.

| Aspect | Ce projet (desktop Go) | Équivalent web courant |
|---|---|---|
| Langage | Go (compilé) | JS / PHP / Python |
| Interface | Fyne (widgets natifs) | HTML/CSS + framework JS |
| Base de données | SQLite (fichier local) | MySQL/PostgreSQL (serveur) |
| Déploiement | 1 fichier .exe à copier | déploiement serveur + build front |
| Authentification | aucune (admin unique) | login/session |

**Pourquoi ces choix ?** L'application doit tourner sur les postes d'un collège,
sans serveur, sans installation, et être copiable sur une clé USB ou un partage
réseau. D'où : un binaire autonome (Go compile tout, y compris ses dépendances)
et une base fichier (SQLite).

---

## 2. Stack technique

- **Go 1.21** — langage compilé, typé statiquement. Pas de runtime à installer :
  le binaire contient tout.
- **Fyne v2.4.5** ([fyne.io](https://fyne.io)) — boîte à outils d'interface
  graphique. Joue le rôle qu'aurait un framework front (React/Vue) mais pour des
  fenêtres natives. On compose des `widget` (boutons, listes, champs) dans des
  `container` (équivalents des `<div>` avec un système de layout).
- **modernc.org/sqlite v1.28.0** — pilote SQLite **en Go pur** (réécriture de
  SQLite sans C). Avantage clé : pas besoin de compilateur C, donc on compile un
  `.exe` Windows depuis un Mac/Linux sans rien installer.
- **github.com/phpdave11/gofpdf v1.4.2** — génération de PDF (bons de remise,
  rapports).

> Le fichier `go.mod` liste 2 dépendances directes utiles (Fyne, gofpdf) plus
> SQLite ; tout le reste (`// indirect`) sont des dépendances transitives de Fyne.

---

## 3. Architecture des dossiers

Go organise le code en **packages** (un package ≈ un dossier ≈ un module au sens
JS). Le dossier `internal/` a une signification spéciale en Go : son contenu ne
peut être importé que par ce projet (équivalent d'un `private`/non publié).

```
cmd/
  main.go              Point d'entrée (le "index.php" / "main.js" du projet)

internal/
  db/                  COUCHE DONNÉES (≈ models + repository + migrations)
    database.go          Connexion SQLite, PRAGMA, création des tables
    models.go            Structs = les "entités" (Key, Room, Borrower, Loan...)
    queries.go           Toutes les requêtes SQL (CRUD + agrégats)
    migrations.go        Migrations de schéma versionnées
    backup.go            Sauvegarde/restauration + import base Python (V1)
    import_v2.go         Import d'une base de la version précédente (V2)

  business/            LOGIQUE MÉTIER PURE (≈ services, sans I/O ni UI)
    loan_advisor.go      Algorithme "prêt par besoin" (Set Cover glouton)
    redundancy.go        Détection des clés redondantes
    *_test.go            Tests unitaires de cette logique

  export/
    csv.go               Export CSV compatible Excel France

  pdf/
    generator.go         Construction du contenu des PDF
    exporter.go          Écriture des PDF sur disque

  gui/                 INTERFACE (≈ la couche "vues"/"composants")
    app.go               App principale, fenêtre, menu, navigation
    *_view / *.go        Une vue par fichier (keys.go, borrowers.go...)
    new_loan_view.go     Vue centrale : création d'un emprunt
    widgets.go           Composants réutilisables (cards, lignes à cocher)
    theme_simple.go      Thème visuel
```

**Sens de dépendance** (comme en architecture en couches) :

```
gui  ──>  business  ──>  db  ──>  SQLite
 │                        ▲
 └────────────────────────┘   (la GUI appelle aussi db directement)
export / pdf  <──  gui
```

`db` ne connaît ni `business` ni `gui`. `business` ne connaît que `db`. La GUI
est tout en haut. C'est l'inverse exact des imports : les couches basses ne
remontent jamais.

---

## 4. La couche données (`internal/db`)

### 4.1 Connexion — un singleton global

```go
var DB *sql.DB   // database.go
```

`DB` est un **pool de connexions** partagé par tout le package (équivalent d'une
instance PDO globale en PHP, ou d'un pool `pg` en Node). Il est initialisé une
fois par `InitDB(path)` au démarrage. Toutes les fonctions de `queries.go`
l'utilisent directement.

### 4.2 Réglages SQLite critiques (à connaître absolument)

Dans `InitDB`, deux décisions conditionnent tout le comportement concurrent :

1. **Mode WAL** (`PRAGMA journal_mode = WAL`) + `busy_timeout = 5000` : permet
   à plusieurs postes du réseau d'ouvrir le même fichier `.db` simultanément.
   Les lecteurs ne bloquent pas l'écrivain. Si la base est verrouillée, on
   attend jusqu'à 5 secondes au lieu d'échouer.

2. **Une seule connexion** (`SetMaxOpenConns(1)`). SQLite n'a qu'un écrivain ;
   on sérialise donc au niveau du pool Go.

   > ⚠️ **Piège connu** : avec une seule connexion, **ne jamais lancer une
   > requête imbriquée pendant une transaction ouverte** — cela provoque un
   > auto-blocage (deadlock) qui gèle l'application. C'est pourquoi
   > `CreateMultipleLoans` n'appelle aucune fonction de lecture dans sa
   > transaction : la disponibilité est vérifiée *avant*, par la couche métier.

### 4.3 Modèles (`models.go`)

Les `struct` Go sont les entités. Exemples :

- `Key` — une clé (numéro, description, quantités total/réserve, stockage).
- `Room` — un **accès** (porte/zone). Historiquement nommé "room"/"salle" ;
  l'UI parle d'« accès ». Champs : nom, bâtiment, étage, catégorie.
- `Borrower` — un détenteur (nom, statut, email, téléphone).
- `Loan` — un emprunt (clé + détenteur + dates + état au retour).
- `Building` — un bâtiment.

Quelques structs sont des **vues enrichies** (résultats de jointures, pas des
tables) :

- `KeyWithAvailability` — une clé + son nombre disponible calculé + qui la détient.
- `KeyWithCoverage` — une clé + la liste des accès qu'elle ouvre (pour l'algo).
- `LoanWithDetails` — un emprunt + nom de clé/détenteur (jointure prête à afficher).

### 4.4 Schéma relationnel

```
buildings (1) ──< (N) rooms
keys (N) >──< (N) rooms      via la table de liaison key_room_association
keys (1) ──< (N) loans >── (1) borrowers
```

La relation **clés ↔ accès est N-N** : une clé peut ouvrir plusieurs portes, une
porte peut être ouverte par plusieurs clés (ex. un passe général). D'où la table
de jointure `key_room_association`.

### 4.5 Requêtes (`queries.go`)

Toutes les requêtes SQL sont ici, regroupées par entité. Conventions :

- Constantes `keySelectSQL`, `borrowerSelectSQL`... : le `SELECT ... FROM`
  réutilisé, auquel on ajoute des `WHERE`/`ORDER BY` selon le besoin (évite la
  duplication de la liste de colonnes).
- Helpers `scanKeys`, `scanBorrowers`, `scanRooms` : centralisent le mapping
  ligne SQL → struct (équivalent d'un `fetchAll` typé).
- **Requêtes paramétrées partout** (`?`) : aucune concaténation de variable dans
  le SQL → pas d'injection SQL possible.
- `CheckInventoryAnomalies` liste les clés en **sur-prêt** (disponible négatif).
  Règle métier importante : le sur-prêt est *autorisé* — aucune requête ne bloque
  un prêt au-delà du stock ; l'anomalie est seulement signalée à l'écran
  (« erreur d'inventaire, vérifier le stock »). Ne pas « corriger » cela en
  ajoutant une contrainte : c'est un choix délibéré, verrouillé par
  `TestOverloanAllowed`.
- `CreateMultipleLoans` retourne les IDs des prêts créés : c'est ce qui permet
  au bon de remise PDF de cibler exactement les prêts de la remise en cours.

### 4.6 Migrations (`migrations.go`)

Le schéma évolue via des migrations **versionnées**, enregistrées dans une table
`schema_version`. Au démarrage :
- base vierge → `createTables()` crée directement le schéma V3 complet ;
- base existante d'une ancienne version → les migrations manquantes sont
  appliquées (ALTER TABLE) et marquées comme faites.

C'est le même principe que les migrations Laravel/Rails/Knex.

### 4.7 Sauvegarde (`backup.go`)

`BackupDatabase` utilise `VACUUM INTO` plutôt qu'une copie de fichier : c'est
**atomique** et sûr même si quelqu'un écrit pendant la sauvegarde.

Piège connu : `VACUUM INTO` passe par la **connexion ouverte**. Dans
`RestoreDatabase`, la copie de sécurité doit donc être prise *avant* de fermer
la connexion — l'ordre inverse a rendu la restauration inopérante jusqu'en
3.2.1 (corrigé en 3.3.0, verrouillé par `TestBackupAndRestore`).

### 4.8 Garantie de compatibilité des sauvegardes (règle absolue)

Une sauvegarde d'une ancienne version doit TOUJOURS pouvoir être restaurée ou
importée dans la version courante. Concrètement, pour toute évolution du
schéma :

- uniquement des `ALTER TABLE ... ADD COLUMN` via `migrations.go` — jamais de
  suppression/renommage de colonne, jamais de contrainte `CHECK` ajoutée à une
  table existante (SQLite l'interdirait de toute façon sans reconstruire la
  table, ce qui casserait la restauration de bases contenant des données
  « hors norme », par ex. un stock en sur-prêt) ;
- les validations de saisie vivent dans l'interface (`internal/gui`), pas dans
  le schéma ;
- trois chemins d'entrée sont couverts par les tests et doivent le rester :
  `RestoreDatabase` (sauvegardes 3.x), `ImportFromV2` (bases V2) et
  `ImportFromPythonDB` (bases V1).

---

## 5. La logique métier (`internal/business`)

Package **pur** : il reçoit ses données en paramètre et ne touche ni à l'écran ni
(presque) à la base. C'est ce qui le rend testable unitairement (`go test`).

### 5.1 Prêt par besoin — `loan_advisor.go`

**Problème** : l'utilisateur coche des portes ; on veut lui proposer le **plus
petit jeu de clés** qui ouvre toutes ces portes. C'est le problème classique du
**Set Cover** (couverture d'ensembles), NP-difficile, donc résolu par une
**heuristique gloutonne** (*greedy*) :

```
Tant qu'il reste des portes à couvrir :
    choisir la clé qui couvre le PLUS de portes encore non couvertes
    (en cas d'égalité, préférer celle qui ouvre le moins de portes "en trop")
    ajouter cette clé au trousseau ; retirer les portes qu'elle couvre
Si une porte n'est ouvrable par aucune clé disponible → la signaler
```

`SuggestKeys` implémente cette boucle. `findBestKey` choisit la meilleure clé à
chaque tour. Les ensembles d'IDs sont représentés par des `map[int]struct{}`
(un `Set`, `struct{}` ne consommant aucune mémoire).

### 5.2 Redondances — `redundancy.go`

Détecte qu'un détenteur a une clé **inutile** car tous ses accès sont déjà
couverts par ses autres clés. Sert d'alerte (sécurité, optimisation du parc).

---

## 6. L'interface graphique (`internal/gui`)

### 6.1 Modèle mental Fyne pour un développeur web

| Fyne | Équivalent web |
|---|---|
| `widget.Button`, `widget.Label`, `widget.Entry` | `<button>`, `<span>`, `<input>` |
| `container.NewVBox(...)` | flexbox vertical |
| `container.NewHBox(...)` | flexbox horizontal |
| `container.NewBorder(top, bottom, left, right, center)` | layout "holy grail" |
| `container.NewVScroll(...)` | conteneur `overflow:auto` |
| callback `func() { ... }` sur un bouton | `onClick` |
| `widget.Refresh()` | re-render du composant |

### 6.2 Navigation

`App` (dans `app.go`) détient la fenêtre et un **menu latéral permanent**. La
zone centrale est remplacée à chaque navigation par `setContent`. Chaque vue est
une fonction `createXxxView(app) fyne.CanvasObject` dans son propre fichier —
c'est l'équivalent d'un composant de page. Les méthodes `showXxx()` font le lien
menu → vue.

Il n'y a **pas de routeur** ni d'URL : changer de vue = remplacer le contenu
central et reconstruire le menu.

### 6.3 Deux pièges Fyne récurrents (documentés dans le code)

1. **`VScroll` dans un `VBox` se réduit à zéro.** Pour qu'une liste scrollable
   prenne la hauteur restante, il faut l'utiliser comme *centre* d'un
   `container.NewBorder` (le centre absorbe l'espace libre). Ce motif est partout.

2. **Widget référencé dans sa propre closure.** En Go, on ne peut pas utiliser
   une variable avant sa déclaration. Pour qu'un bouton se ferme lui-même, ou
   qu'une fonction de recalcul s'appelle récursivement, on déclare d'abord
   `var x *widget.X` (ou `var f func()`) **puis** on l'assigne :

   ```go
   var recalculate func()
   recalculate = func() { /* ... peut s'appeler recalculate() ... */ }
   ```

### 6.4 La vue centrale : `new_loan_view.go`

C'est la pièce maîtresse. Deux panneaux :
- **gauche** : choix du détenteur + cases à cocher des portes ;
- **droite** : le « trousseau calculé » mis à jour en temps réel.

Logique notable :
- `recalculate()` est rappelée à chaque changement et relance l'algo `SuggestKeys`.
- Deux ensembles mémorisent les choix manuels de l'utilisateur :
  `excludedKeyIDs` (clés retirées à la main, jamais re-proposées) et
  `manualKeyIDs` (clés ajoutées à la main, toujours conservées). Changer les
  portes cochées remet `excludedKeyIDs` à zéro (nouvelle situation).
- À la validation : `CreateMultipleLoans` enregistre les prêts puis un bon PDF
  est généré **dans une goroutine** (thread léger Go, ≈ tâche asynchrone) pour ne
  pas bloquer l'interface. La navigation vers le tableau de bord se fait
  **après** clic sur OK (sinon Fyne détruit la vue pendant le traitement du clic).

---

## 7. Génération de documents (`pdf`, `export`)

- **PDF** (`pdf/generator.go`) : chaque `GenerateXxx` construit un document
  gofpdf et retourne des `[]byte` ; `exporter.go` écrit dans `documents/`.
  Toute chaîne accentuée passe par le traducteur Unicode
  (`p.UnicodeTranslatorFromDescriptor("")`) car les polices PDF de base sont en
  Latin-1.
- **Bon de remise** : généré de façon *synchrone* à la validation d'un emprunt
  (`generateHandoverReceipt`, `internal/gui/new_loan_view.go`) — une erreur de
  génération s'affiche dans la confirmation. Le même helper sert la
  **réédition** (bouton « Rééditer le bon », fiche détenteur et emprunts en
  cours) : `loanIDs` vide → bon complet couvrant tous les prêts actifs.
- **CSV** (`export/csv.go`) : UTF-8 **avec BOM** + séparateur `;`. C'est ce qui
  permet à Excel français d'ouvrir le fichier directement, accents corrects et
  colonnes séparées, sans assistant d'import.

---

## 8. Compiler et lancer

```bash
# Lancer en développement (compile + exécute, comme `npm run dev`)
go run ./cmd/main.go

# Compiler le binaire de la machine courante
go build -o clefs ./cmd/main.go

# Compiler le .exe Windows depuis n'importe quel OS (cross-compilation)
GOOS=windows GOARCH=amd64 go build -o clefs-windows-amd64.exe ./cmd/main.go

# Compiler le bundle .app macOS (machine courante)
./compilationmacos.sh

# Lancer les tests (logique métier + couche données sur base temporaire)
go test ./...

# Détecter le code mort (outil officiel)
go run golang.org/x/tools/cmd/deadcode@latest ./...
```

Au premier lancement, l'application crée à côté de l'exécutable :
`clefs.db` (la base), `backups/` et `documents/`.

Les releases GitHub sont produites par `.github/workflows/release.yml` au push
d'un tag `v*` : Windows x64 et macOS Apple Silicon (binaires non signés — le
contournement Gatekeeper est documenté dans `infos.txt`).

---

## 9. Pour étendre le projet — recettes

**Ajouter un champ à une entité** (ex. un champ sur les clés) :
1. Ajouter la colonne dans `createTables()` (database.go) **et** une migration
   dans `migrations.go` (pour les bases existantes).
2. Ajouter le champ dans la struct correspondante (`models.go`).
3. Mettre à jour la requête `SELECT` (constante `xxxSelectSQL`) et le `scanXxx`.
4. Ajouter le champ dans le formulaire de la vue concernée (`gui/keys.go`...).

**Ajouter une nouvelle vue/écran** :
1. Créer `internal/gui/ma_vue.go` avec `func createMaVue(a *App) fyne.CanvasObject`.
2. Ajouter une méthode `func (a *App) showMaVue()` dans `app.go`.
3. Ajouter un bouton dans `createMenu()` qui appelle `a.showMaVue()`.

**Ajouter une requête** : l'écrire dans `queries.go` en réutilisant les
constantes `SELECT` et les helpers `scan*`, avec des paramètres `?`.

---

## 10. Conventions et qualité

- `gofmt` est la référence de formatage (non négociable en Go) : lancez
  `gofmt -w internal cmd` avant de committer.
- `go vet ./...` doit rester silencieux.
- `go test ./...` doit rester vert. Les tests de `internal/db` créent chacun
  une base SQLite dans `t.TempDir()` via `setupTestDB` ; le package reposant
  sur le singleton `DB`, ils n'utilisent **pas** `t.Parallel()`.
- Le projet est garanti **sans code mort** (vérifié avec `deadcode`).
- Chaque package et fonction non triviale porte un commentaire godoc en français.
- Toute évolution de schéma respecte la garantie de compatibilité des
  sauvegardes (§ 4.8).

---

## 11. Licence et crédits

Distribué sous licence **MIT**.

- **Conception & direction :** COLLET David — Secrétaire Général,
  Collège Victor Hugo, Chauny (02300) — david.collet@ac-amiens.fr
- **Réalisation du code :** intégralement générée par IA (Claude, Anthropic),
  sous supervision humaine.
