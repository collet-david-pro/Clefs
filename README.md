# Gestionnaire de Clés — V3

Application de bureau native pour la gestion du parc de clés du **Collège Victor Hugo — Chauny (02300)**.

Développée en Go + Fyne, elle fonctionne comme un exécutable unique Windows, sans installation ni connexion internet.

---

## Fonctionnalités

### Référentiel des accès
- Enregistrement de chaque porte, portail ou zone verrouillée comme un **accès** indépendant
- Champs : désignation, bâtiment, étage/niveau, catégorie, observations
- Filtrage par bâtiment, étage et catégorie

### Référentiel des clés
- Numéro, désignation, catégorie (simple / trousseau / badge / passe)
- Stock total, réserve, emplacement de rangement, observations
- **Liaison clé ↔ accès** : une clé peut ouvrir plusieurs portes
- Disponibilité calculée automatiquement : `Dispo = Total − Réserve − Sorties`

### Référentiel des détenteurs
- Nom, statut (permanent / contractuel / intervenant / entreprise), email, téléphone

### Prêt par besoin — assistant en 3 étapes
1. Sélection du détenteur
2. Sélection des accès requis (filtrable par bâtiment / étage / catégorie) + date de retour prévue + type de prêt
3. **Proposition automatique de la combinaison minimale de clés** couvrant les accès demandés (algorithme greedy Set Cover), modifiable avant validation

### Retour de clé
- Enregistrement de l'état constaté au retour (bon état, rayé, etc.)

### Détection des redondances
- Signal visuel si un détenteur possède plusieurs clés ouvrant les mêmes portes
- Vue dédiée ⚠️ Redondances accessible depuis le menu

### Historique complet
- Tous les prêts (actifs + retournés) filtrables par clé, détenteur, statut, période
- Statuts : En cours / Retourné / En retard

### Tableau de bord
- Statistiques en temps réel : clés totales, emprunts actifs, disponibles, détenteurs
- Alerte 🔴 prêts en retard et ⚠️ redondances d'accès visibles dès l'ouverture
- Bouton ➕ Nouvel Emprunt en accès direct

### Vues rapides
- **Qui a quoi ?** — détenteurs avec leurs clés actuelles et accès couverts
- **Quelle clé pour quelle porte ?** — sélectionner un accès → clés associées + disponibilité
- **Clés par bâtiment** — toutes les clés d'un bâtiment donné
- **Clés disponibles** — stock / sorties / dispo filtrable

### Plan de clés
- Vue bâtiment → salles → clés associées, exportable en PDF

### Génération de PDF
- Bon de remise enrichi : clés remises, accès couverts, agent, date de retour prévue, zone signature
- Rapport des clés sorties
- Rapport global par détenteur
- Bilan du stock de clés

### Export CSV
- Inventaire des clés, liste des détenteurs, emprunts en cours, historique filtré
- Format UTF-8 avec BOM, séparateur `;` — compatible Excel directement

### Sauvegarde / Restauration
- Sauvegarde atomique via `VACUUM INTO` (sûre même en cours d'utilisation)
- Restauration depuis une sauvegarde
- Sauvegarde rapide en un clic

### Migration
- Import depuis une base **V2** (`.db`) : validation du schéma, sauvegarde automatique avant import, résumé par table
- Import depuis une base **V1 Python** (`.db`)

### Multi-postes simultanés
- Mode WAL SQLite + `busy_timeout` : plusieurs postes peuvent consulter et saisir en même temps depuis le réseau local
- Une seule écriture à la fois, les autres attendent jusqu'à 5 secondes automatiquement

---

## Installation

1. Aller sur la page [**Releases**](https://github.com/collet-david-pro/Clefs/releases)
2. Télécharger `clefs-windows-amd64.zip`
3. Décompresser dans un **dossier dédié** (ex. `C:\Clefs\`)
4. Double-cliquer sur `clefs-windows-amd64.exe`

> **Windows Defender peut afficher une alerte** car l'exécutable n'est pas signé. Cliquer sur "Informations complémentaires" → "Exécuter quand même".

---

## Premier lancement

Au démarrage, l'application crée automatiquement dans son dossier :

```
MonDossierClefs/
├── clefs-windows-amd64.exe
├── infos.txt
├── clefs.db          ← base de données (NE PAS SUPPRIMER)
├── backups/          ← sauvegardes automatiques
└── documents/        ← PDF et CSV générés
```

Une identification simple par nom est demandée à chaque ouverture (aucun mot de passe).

---

## Utilisation en réseau

Placer le dossier sur un partage réseau (SMB). Plusieurs postes peuvent ouvrir l'application simultanément — SQLite en mode WAL gère les accès concurrents automatiquement.

---

## Migration depuis V2

1. Configuration → **Importer depuis sauvegarde V2**
2. Sélectionner l'ancien fichier `clefs.db`
3. Une sauvegarde de la base actuelle est créée automatiquement avant l'import
4. Les nouveaux champs V3 (étage, catégorie, statut détenteur) sont laissés vides et à compléter manuellement

## Migration depuis V1 Python

1. Configuration → **Importer depuis Version Python (V1)**
2. Sélectionner l'ancien fichier `clefs.db`

---

## Pour les développeurs

### Stack technique
- **Langage :** Go 1.21
- **Interface graphique :** Fyne v2.4.5
- **Base de données :** SQLite (modernc.org/sqlite — pur Go, sans CGO)
- **PDF :** gofpdf
- **Tests :** package `testing` standard

### Structure du projet

```
cmd/
  main.go

internal/
  db/
    database.go       Connexion SQLite, PRAGMA, migrations
    models.go         Structs (Key, Room, Borrower, Loan...)
    queries.go        Toutes les requêtes SQL
    store.go          Interface Store
    store_sqlite.go   Implémentation SQLiteStore
    migrations.go     Migrations de schéma versionnées
    backup.go         Sauvegarde VACUUM INTO, import Python
    import_v2.go      Import base V2

  business/
    loan_advisor.go   Algorithme Set Cover (prêt par besoin)
    redundancy.go     Détection redondances d'accès

  export/
    csv.go            Export CSV UTF-8 + BOM

  gui/
    app.go            Application principale, navigation, login
    dashboard_modern.go
    accesses.go       Gestion des accès (portes/zones)
    keys.go           Gestion des clés
    borrowers.go      Gestion des détenteurs
    loans.go          Emprunts actifs, rapport
    loan_wizard.go    Assistant de prêt en 3 étapes
    loan_dialogs.go   Dialogues de prêt/retour rapides
    history.go        Historique filtrable
    views_transversal.go  Vues rapides
    redundancy_view.go    Vue redondances
    login.go          Identification au démarrage
    config.go         Configuration, sauvegarde, import
    keyplan.go        Plan de clés

  pdf/
    generator.go      Génération PDF (bons, rapports)
    exporter.go       Sauvegarde fichier PDF
```

### Lancer en développement

```bash
go run ./cmd/main.go
```

### Lancer les tests

```bash
go test ./...
```

### Créer une release

```bash
./create-release.sh 3.0.0
```

Le script committe les modifications, crée le tag et pousse vers GitHub. Le workflow Actions compile l'exécutable Windows et publie la release automatiquement.

---

## Licence

Distribué sous licence **MIT**.
