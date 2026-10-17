# NOTICE TECHNIQUE — Gestionnaire de Clés (V4)

> Document à jour pour la version 4.0.0.

> **Note importante : code intégralement généré par IA.**
> L'ensemble de cette application (architecture, code Go, requêtes SQL,
> interface, tests et documentation) a été **produit par une intelligence
> artificielle** (Claude, Anthropic), sous la direction et la relecture de
> **COLLET David** (Secrétaire Général, Collège Victor Hugo, Chauny). Ce
> document décrit le fonctionnement technique pour permettre à un développeur
> de reprendre, maintenir ou faire évoluer le projet.

Cette notice s'adresse à un **développeur familier du web** (JavaScript/PHP/Python,
SQL, MVC) mais pas forcément de Go ni des applications de bureau. Bonne
nouvelle : depuis la V4, l'interface **est** une application web (Svelte),
affichée dans une fenêtre native.

---

## 1. Vue d'ensemble

Le **Gestionnaire de Clés** est une application de **bureau** : un **exécutable
unique** Windows qui ouvre une fenêtre et lit/écrit dans un fichier de base de
données local. Il n'y a ni serveur HTTP à déployer, ni navigateur à lancer.

| Aspect | Ce projet (V4) | Équivalent web courant |
|---|---|---|
| Moteur | Go (compilé) | API PHP / Node / Python |
| Interface | Svelte + Tailwind dans une fenêtre WebView2 | Front SPA |
| Lien interface ↔ moteur | Wails : appel direct de méthodes Go depuis JS | Appels `fetch` vers une API REST |
| Base de données | SQLite (fichier local) | MySQL/PostgreSQL (serveur) |
| Déploiement | 1 fichier .exe à copier | serveur + build front |
| Authentification | aucune (administrateur unique, agent = session Windows) | login/session |

**Pourquoi ces choix ?** L'application doit tourner sur les postes d'un collège,
sans serveur, sans droits administrateur, **sans rien installer**, et être
copiable sur une clé USB ou un partage réseau. Wails compile l'interface web
*dans* l'exécutable Go et l'affiche avec le moteur Edge WebView2, fourni avec
Windows 10 et 11.

---

## 2. Stack technique

- **Go 1.25** — langage compilé, typé statiquement. Le binaire contient tout.
- **Wails v2.16** ([wails.io](https://wails.io)) — fenêtre native + pont JS ↔ Go.
  Chaque méthode publique de `internal/app.App` devient une fonction JavaScript
  qui retourne une `Promise`.
- **modernc.org/sqlite v1.28** — SQLite réécrit en Go pur (pas de compilateur C).
- **github.com/phpdave11/gofpdf v1.4.2** — génération de PDF.
- **Frontend** (`frontend/`, compilé par Vite) : Svelte 5 (runes), TypeScript,
  Tailwind CSS 4, icônes `@lucide/svelte`, police Inter (`@fontsource-variable`).
  Tout est embarqué : **aucune ressource chargée depuis internet**.

Les versions npm sont **figées exactement** dans `frontend/package.json`
(choisies publiées depuis plus de deux semaines au moment de la mise à jour).

---

## 3. Architecture des dossiers

```
main.go               Point d'entrée : dossier de travail, ouverture de la base, fenêtre Wails
main_bindings.go      Compilé seulement avec le tag « bindings » (génération des liaisons JS)
wails.json            Configuration Wails (commandes npm, dossier des liaisons)
build/                Icône (appicon.png) et ressources Windows ; build/bin/ = exécutable

internal/
  db/                 COUCHE DONNÉES (≈ models + repository + migrations)
    database.go         Connexion SQLite, PRAGMA, création des tables
    models.go           Structs = les « entités » (Key, Room, Borrower, Loan…)
    queries.go          Toutes les requêtes SQL
    migrations.go       Migrations de schéma versionnées
    backup.go           Sauvegarde / restauration / réinitialisation / démo
    import_v2.go        Import d'une base V1 (Python) ou V2 (Go)
    import_v3.go        Détection de version, aperçu, import d'une base 3.x

  business/           LOGIQUE MÉTIER PURE (≈ services, sans I/O ni UI)
    loan_advisor.go     « Prêt par besoin » (Set Cover glouton)
    redundancy.go       Détection des clés redondantes

  app/                COUCHE DE SERVICES (≈ contrôleurs d'une API)
    app.go              Struct App, chemins, verrou, ouverture de fichiers
    dto.go              Structures JSON échangées avec l'interface
    catalog.go          Clés, accès, bâtiments, détenteurs (+ validations)
    loans.go            Prêt, retours, bons de remise, historique, vues
    documents.go        PDF, exports CSV, import CSV des détenteurs
    maintenance.go      Sauvegardes, assistant d'import, réinitialisation
    inventory.go        Inventaire (stock compté → nouveau total) + rapport

  pdf/                PDF au modèle de l'application (doc.go), documents.go, police Inter (fonts/)
  export/             CSV compatible Excel France
  natsort/            Tri naturel des numéros de clés et noms d'accès

frontend/             INTERFACE (application Svelte)
  src/app.css           Thème : jetons de couleur et composants de base
  src/main.ts           Démarrage
  src/App.svelte        Structure : menu latéral, barre de recherche, page courante
  src/pages/            Une page par écran (≈ routes)
  src/lib/api.ts        Accès aux services Go + gestion des erreurs
  src/lib/state.svelte.ts  État global : navigation, notifications, confirmations, thème
  src/lib/format.ts     Dates, pluriels, libellés français
  src/lib/components/   Composants réutilisables (formulaires, fiches, fenêtres…)
  src/lib/wailsjs/      Liaisons générées par Wails — NE PAS MODIFIER
```

**Sens de dépendance** :

```
frontend ──(Wails)──> app ──> business ──> db ──> SQLite
                       │                   ▲
                       └───────────────────┘
                       └──> pdf / export
```

Les couches basses ne connaissent jamais les couches hautes.

---

## 4. La couche données (`internal/db`)

### 4.1 Connexion — un singleton global

`var DB *sql.DB` (database.go) est un **pool de connexions** partagé par tout
le package, initialisé par `InitDB(path)` au démarrage.

### 4.2 Réglages SQLite critiques

1. **Mode WAL** + `busy_timeout = 5000` : plusieurs postes du réseau peuvent
   ouvrir le même fichier `.db` simultanément.
2. **Une seule connexion** (`SetMaxOpenConns(1)`).

   > ⚠️ **Piège connu** : ne **jamais** lancer une requête imbriquée pendant une
   > transaction ouverte — auto-blocage (deadlock) qui gèle l'application.
   > `CreateMultipleLoans` n'appelle aucune lecture dans sa transaction.
   > Wails exécutant les appels du frontend **en parallèle**, cette règle est
   > d'autant plus importante (les appels concurrents attendent la connexion,
   > ce qui est sans danger tant qu'aucun ne s'attend lui-même).

### 4.3 Modèles et schéma

- `Key` (clé), `Room` (un **accès** : historiquement « salle », l'interface dit
  « accès »), `Borrower` (détenteur), `Loan` (prêt), `Building` (bâtiment).
- Vues enrichies : `KeyWithAvailability`, `KeyWithCoverage`, `LoanWithDetails`.

```
buildings (1) ──< (N) rooms
keys (N) >──< (N) rooms      via key_room_association
keys (1) ──< (N) loans >── (1) borrowers
```

### 4.4 Requêtes (`queries.go`)

- Constantes `keySelectSQL`, `borrowerSelectSQL`… + helpers `scanKeys`,
  `scanBorrowers`, `scanRooms`.
- **Requêtes paramétrées partout** (`?`), y compris `VACUUM INTO ?`.
- `CheckInventoryAnomalies` liste les clés en **sur-prêt**. Le sur-prêt est
  *autorisé* et seulement signalé (verrouillé par `TestOverloanAllowed`).
- `CreateMultipleLoans(keyIDs, borrowerID, LoanOptions)` enregistre retour
  prévu, type de prêt et agent, et retourne les IDs créés (le bon PDF cible
  exactement ces prêts).
- `GetKeyAccessMap` : toutes les associations clé → accès en une requête.
- `ApplyInventory` : inventaire en une transaction. Pour chaque clé comptée,
  `total = en stock + prêts actifs` (sous-requête dans le même `UPDATE`, donc
  aucune lecture imbriquée) et réserve ramenée au nombre compté si elle le
  dépasse. Verrouillé par `TestApplyInventory`. Un total de 0 est permis
  (toutes les clés perdues ou cassées, fiche conservée pour l'historique).

> **Corrigé en V4** : la V3 affichait mais n'enregistrait pas le statut et le
> téléphone des détenteurs, la catégorie et les notes des clés, ni la date de
> retour prévue, le type de prêt et l'agent. Les colonnes existaient : aucune
> évolution de schéma n'a été nécessaire. Verrouillé par `TestV4FieldsPersisted`.

### 4.5 Migrations (`migrations.go`)

Table `schema_version` ; base vierge → schéma complet ; base ancienne →
migrations manquantes. Migration 3 (V4) : table `settings` (paramètres),
créée à l'ouverture d'une base 3.x (`TestSettingsTableAddedToV3Base`).
Prochaine migration : `version: 4`.

### 4.6 Sauvegarde (`backup.go`)

`BackupDatabase` utilise `VACUUM INTO` (atomique, sûr pendant les écritures).
Dans `RestoreDatabase`, la copie de sécurité est prise **avant** de fermer la
connexion (bug corrigé en 3.3.0, verrouillé par `TestBackupAndRestore`).

### 4.7 Imports des anciennes versions

| Source | Détection (`InspectDatabase`) | Import |
|---|---|---|
| **3.x** | table `schema_version` ou colonne `rooms.floor` | `ImportFromV3` : copie de la source par `VACUUM INTO` depuis *sa* connexion (intègre son journal `-wal`), puis `RestoreDatabase` |
| **V1 / V2** | les 6 tables historiques, sans marqueur V3 | `ImportFromV2` : recopie table par table (`INSERT OR IGNORE`, IDs conservés), **base vide uniquement** |
| autre | — | refusé (« fichier non reconnu ») |

V1 (Python) et V2 (Go) ont exactement les mêmes colonnes : un seul chemin
d'import les couvre (l'ancien `ImportFromPythonDB`, identique, a été supprimé).
La source n'est **jamais** modifiée (vérifié par `TestImportFromV3`).

### 4.8 Garantie de compatibilité des données (règle absolue)

Une base ou une sauvegarde d'une version précédente doit **toujours** pouvoir
être importée dans la version courante :

- uniquement des **ajouts** via `migrations.go` (`ALTER TABLE … ADD COLUMN`,
  `CREATE TABLE IF NOT EXISTS`) — jamais de suppression ni de renommage,
  jamais de contrainte `CHECK` ajoutée à une table existante ;
- les validations de saisie vivent dans `internal/app` et l'interface ;
- trois chemins d'entrée testés : `RestoreDatabase` (sauvegardes),
  `ImportFromV3` (bases 3.x) et `ImportFromV2` (bases V1/V2).

---

## 5. La logique métier (`internal/business`)

Package **pur**, testable unitairement.

- **Prêt par besoin** (`loan_advisor.go`) : problème du *Set Cover*, résolu par
  une heuristique gloutonne : tant qu'il reste des portes, prendre la clé
  disponible qui en couvre le plus (à égalité, celle qui ouvre le moins de
  portes en trop).
- **Redondances** (`redundancy.go`) : accès couverts par plusieurs clés d'un
  même détenteur.

---

## 6. La couche de services (`internal/app`)

C'est l'équivalent des **contrôleurs d'une API**, sans HTTP : Wails expose
chaque méthode publique de `App` au JavaScript.

```go
// Go (internal/app/catalog.go)
func (a *App) SaveKey(in KeyInput) (int, error)
```
```ts
// TypeScript (généré dans frontend/src/lib/wailsjs/go/app/App.d.ts)
export function SaveKey(arg1: app.KeyInput): Promise<number>
```

- Une `error` Go rejette la `Promise` avec son message (en français, affiché
  tel quel à l'utilisateur par `run()` dans `lib/api.ts`).
- Les structures échangées sont dans `dto.go` (balises `json` en camelCase,
  dates en ISO 8601). Wails en génère les classes TypeScript (`models.ts`).
- **Validations** : règles de saisie (numéro obligatoire, quantités ≥ 0,
  réserve ≤ total, date de retour non passée, stock compté ≥ 0…), refus de
  supprimer une clé prêtée ou un détenteur qui a des clés.
- `Quit()` ferme l'application (bouton « Quitter » du menu ; une saisie
  d'inventaire non validée est signalée avant).
- **Verrou** `sync.RWMutex` : les opérations qui remplacent la base
  (restauration, import, réinitialisation) prennent le verrou exclusif ; toutes
  les autres le verrou partagé (`defer a.read()()`).
- `OpenPath` n'ouvre que des fichiers situés dans le dossier de l'application.
- L'agent enregistré sur les prêts est le nom de la session Windows.

---

## 7. L'interface (`frontend/`)

### 7.1 Svelte 5 pour un développeur web

| Svelte 5 | Équivalent |
|---|---|
| `let x = $state(0)` | `useState` (React), `ref` (Vue) |
| `const y = $derived(x * 2)` | `useMemo`, `computed` |
| `$effect(() => …)` | `useEffect`, `watchEffect` |
| `let { a } = $props()` | props |
| `{#if}`, `{#each}`, `{@render children()}` | directives de template, slots |

### 7.2 Navigation et état global (`lib/state.svelte.ts`)

Pas de routeur ni d'URL : `route.page` désigne la page affichée, `navigate()`
en change. Le même fichier gère :
- `data.version` + `refresh()` : après chaque modification, les pages et fiches
  ouvertes rechargent leurs données (`$effect` qui lit `data.version`) ;
- les **notifications** (`notify`), les **confirmations** (`await confirm({…})`),
  les **fiches détaillées** (`openDetail('key', id)`) et le **thème**.

### 7.3 Appels au moteur (`lib/api.ts`)

```ts
const id = await run(Api.SaveKey(input), 'Clé créée.')
// erreur → notification rouge, retourne undefined
// succès + message → notification verte et refresh()
```

### 7.4 Thème (`src/app.css`)

**Fichier unique de personnalisation visuelle** : couleurs (clair et sombre)
sous forme de variables CSS, typographie, rayons, ombres, et classes de base
(`.btn`, `.input`, `.card`, `.badge`, `.table`…). Les pages n'utilisent que ces
jetons (`bg-surface`, `text-muted`, `btn-primary`) : changer une couleur ici la
change partout.

### 7.5 Écrans

Tableau de bord, Nouveau prêt, Prêts en cours, Historique, Clés, Inventaire, Accès,
Détenteurs, Bâtiments, Qui a quoi ?, Plan de clés, Doublons, Sauvegardes,
Importer une base (assistant), Paramètres, Aide, À propos. Les fiches
détaillées (clé, accès, détenteur) s'ouvrent dans un panneau latéral ; la
recherche globale (`Ctrl + K`) est dans `CommandPalette.svelte`.

---

## 8. Documents (`pdf`, `export`)

- **PDF** (`internal/pdf`) : tous les documents partagent le modèle de
  `doc.go`, aligné sur l'interface : police **Inter embarquée en UTF-8**
  (`fonts/`, licence OFL — accents et typographie française sans conversion),
  couleurs du thème clair, en-tête avec logo, cartes d'informations,
  indicateurs, tableaux (en-tête répété à chaque page, retour à la ligne
  automatique, découpage sûr en UTF-8), badges colorés, encadrés, pied de page
  « Généré le … · Page X / N ». Les générateurs de `documents.go` reçoivent
  des structures simples (`Receipt`, `PlanBuilding`, `StockRow`,
  `InventoryReport`), sans accès à la base : ils sont testés isolément
  (`PDF_PREVIEW_DIR=… go test ./internal/pdf` enregistre des exemples).
  `SavePDF` écrit dans `documents/`.
- **Documents** : bon de remise (à la validation d'un prêt, ou récapitulatif
  réédité de toutes les clés détenues), archive de tous les bons en un seul PDF
  (`GenerateReceiptArchive` : sommaire + un bon par page), bilan du stock, plan
  de clés, rapport d'inventaire (écarts).
- **Bon de remise sur une page** : le bon ne liste plus les portes ouvertes.
  `receiptKeys` mesure la place restante (engagement et signatures compris) et
  choisit la présentation : textes complets si tout tient, sinon une des
  densités de `receiptLayouts` (police réduite, puis 2 ou 3 colonnes de clés,
  jusqu'à environ 70 clés). Verrouillé par `TestReceiptFitsOnePage`. Les données sont lues en quelques requêtes groupées
  (`listKeys`, `listAccesses`), plus une requête par clé ou par porte.
- **Paramètres** (`internal/app/settings.go`, table `settings` clé → valeur,
  migration 3) : établissement, logo, textes des bons, valeurs par défaut des
  prêts, agent. Une valeur égale à celle d'origine n'est pas stockée. Ils sont
  transmis aux PDF via `pdf.Settings` (champs vides → `DefaultSettings`) et
  conservés lors d'un import 3.x, d'une restauration d'une base sans
  paramètres et d'une réinitialisation (`preserveSettings`).
- **Tri naturel** (`internal/natsort`) : numéros de clés et noms d'accès sont
  triés « A2, A10 » et « 9, 10 », dans l'interface comme dans les PDF.
- **CSV** : UTF-8 **avec BOM** + séparateur `;` (ouverture directe dans Excel).
- `documents/` et `backups/` sont relatifs au dossier de l'application :
  `main.go` s'y place (`os.Chdir`) au démarrage.

---

## 9. Compiler et lancer

Prérequis **sur le poste de développement seulement** : Go, Node.js + npm, CLI
Wails (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`, puis
`wails doctor`). Les postes utilisateurs n'ont besoin de rien.

```bash
make install-deps   # go mod download + npm install
make dev            # wails dev : fenêtre avec rechargement à chaud
make build          # wails build → build/bin/clefs.exe
make bindings       # régénère frontend/src/lib/wailsjs après ajout d'une méthode Go
make test           # go test ./...
make check          # svelte-check (types de l'interface)
make lint           # gofmt, go vet, deadcode -test
make package        # build + build/release/clefs-windows-amd64.zip (comme la release)
make captures       # captures d'écran du README (docs/images)
```

Sous Windows sans `make`, `build.ps1` fait l'essentiel :

```powershell
.\build.ps1                    # liaisons + build\bin\clefs.exe
.\build.ps1 -Test -Package     # + tests Go, svelte-check et archive de distribution
.\build.ps1 -Run               # lance l'application après compilation
```

`build/bin/` peut contenir une base de test : ni `wails build -clean` ni
suppression du dossier (`make clean` n'efface que l'exécutable).

> **Smart App Control** (Windows 11) peut bloquer l'exécutable temporaire que
> Wails crée pour générer les liaisons (`wailsbindings.exe`), et parfois même
> `go run` (les binaires de `go test` passent). D'où `make bindings` /
> `build.ps1`, qui génèrent les liaisons par `go run -tags bindings .` et
> compilent avec `wails build -skipbindings` ; les liaisons générées sont
> commitées. Sous Windows sans `make` : `.\build.ps1 -Test -Run`. La CI GitHub
> n'est pas concernée.

Au premier lancement, l'application crée à côté de l'exécutable : `clefs.db`,
`backups/` et `documents/`. Sur une base vide, elle ouvre l'assistant d'import.

Les releases GitHub sont produites par `.github/workflows/release.yml` au push
d'un tag `v*` : tests, puis `wails build -webview2 error` (Windows x64) — si
WebView2 manque, message clair, jamais d'installation proposée. Les notes de
release renvoient au guide du README et à `CHANGELOG.md`. Avant un tag,
`create-release.sh` vérifie que la version est identique dans
`internal/app/app.go`, `wails.json` et `frontend/package.json`.

### Documentation

- `README.md` : présentation et guide de prise en main illustré (7 étapes).
- `CHANGELOG.md` : nouveautés par version.
- Aide intégrée : `frontend/src/pages/Help.svelte` (même parcours, avec un
  bouton vers chaque écran) — à tenir à jour avec le README.
- `docs/images/` : captures d'écran et aperçus PDF du README, produits avec
  des **données fictives** : `docs/captures/capture.ps1` compile l'interface,
  la sert avec une maquette des services Go (`mock.js`, scènes en fin de
  fichier) et la photographie avec Edge sans fenêtre. Les aperçus PDF
  (`pdf-*.png`) se régénèrent depuis les générateurs de `internal/pdf`
  (`PDF_PREVIEW_DIR`) puis `pdftoppm`.

---

## 10. Pour étendre le projet — recettes

**Ajouter un champ à une entité** (ex. sur les clés) :
1. Colonne dans `createTables()` **et** migration dans `migrations.go`.
2. Champ dans la struct (`models.go`), la constante `xxxSelectSQL` et `scanXxx`,
   les `INSERT`/`UPDATE`.
3. Champ dans le DTO et l'input (`internal/app/dto.go`) et leurs conversions.
4. `make bindings`, puis champ dans le formulaire (`KeyForm.svelte`).

**Ajouter une action côté moteur** : méthode publique sur `App`
(`defer a.read()()` en tête), `make bindings`, puis appel
`run(Api.MaMethode(...))` dans l'interface.

**Ajouter un écran** :
1. `frontend/src/pages/MaPage.svelte`.
2. L'ajouter au type `Page` (`state.svelte.ts`), à la table `pages` et au menu
   `nav` de `App.svelte`.

---

## 11. Conventions et qualité

- `gofmt`, `go vet ./...` silencieux, `go test ./...` vert, **sans code mort**
  (`deadcode -test ./...`), `svelte-check` sans erreur ni avertissement.
- Tests `internal/db` et `internal/app` : base dans `t.TempDir()`, **pas de
  `t.Parallel()`** (singleton `db.DB`).
- Commentaires et textes en français.
- Toute évolution de schéma respecte la garantie de compatibilité (§ 4.8).

---

## 12. Licence et crédits

Distribué sous licence **MIT**.

- **Conception & direction :** COLLET David — Secrétaire Général,
  Collège Victor Hugo, Chauny (02300) — david.collet@ac-amiens.fr
- **Réalisation du code :** intégralement générée par IA (Claude, Anthropic),
  sous supervision humaine.
