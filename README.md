# Gestionnaire de Clés — V3

Application de bureau native pour la gestion du parc de clés du **Collège Victor Hugo — Chauny (02300)**.

Développée en Go + Fyne, elle fonctionne comme un exécutable unique Windows, sans installation ni connexion internet.

---

## Fonctionnalités

### Référentiel des accès
- Enregistrement de chaque porte, portail ou zone verrouillée comme un **accès** indépendant
- Champs : désignation, bâtiment, étage/niveau, catégorie, observations
- Filtrage par bâtiment

### Référentiel des clés
- Numéro, désignation, catégorie, stock total, réserve, emplacement de rangement, observations
- **Liaison clé ↔ accès** : une clé peut ouvrir plusieurs portes
- Disponibilité calculée automatiquement : `Dispo = Total − Réserve − Sorties`

### Référentiel des détenteurs
- Nom, statut (permanent / contractuel / intervenant / entreprise), email, téléphone

### Prêt par besoin — vue unifiée en une page
1. Sélection du détenteur
2. Sélection des portes à couvrir par cases à cocher (filtrable par bâtiment)
3. **Proposition automatique du jeu de clés minimal** couvrant les accès cochés (algorithme greedy Set Cover)
4. Possibilité de retirer une clé suggérée ou d'en ajouter une manuellement
5. Date de retour prévue + type de prêt (ponctuel / permanent)
6. Validation → enregistrement en base + bon de remise PDF automatique

### Retour de clé
- Depuis le tableau de bord ou les emprunts en cours

### Détection des redondances
- Signal visuel si un détenteur possède plusieurs clés ouvrant les mêmes portes
- Vue dédiée accessible depuis le menu Consultation

### Historique complet
- Tous les prêts (actifs + retournés) filtrables par clé, détenteur, statut, période
- Statuts : En cours / Retourné / En retard

### Tableau de bord
- Statistiques en temps réel : clés totales, emprunts actifs, disponibles, détenteurs
- Alertes prêts en retard et redondances d'accès visibles dès l'ouverture
- Bouton Nouvel emprunt en accès direct

### Vues rapides (menu Consultation)
- **Qui a quoi ?** — détenteurs avec leurs clés actuelles et accès couverts
- **Par bâtiment** — toutes les clés d'un bâtiment donné
- **Redondances** — détenteurs avec accès en doublon
- **Plan de clés** — vue bâtiment → salles → clés, exportable en PDF

### Génération de PDF
- Bon de remise : clés remises, accès couverts, agent, date de retour prévue, zone signature
- Rapport des clés sorties
- Bilan du stock de clés
- Plan de clés

### Export CSV
- Inventaire des clés, liste des détenteurs, historique filtré
- Format UTF-8 avec BOM, séparateur `;` — compatible Excel directement

### Sauvegarde / Restauration
- Sauvegarde atomique via `VACUUM INTO` (sûre même en cours d'utilisation)
- Restauration depuis une sauvegarde avec sauvegarde de sécurité automatique

### Migration
- Import depuis une base **V2** Go (`.db`) : validation du schéma, sauvegarde automatique avant import
- Import depuis une base **V1 Python** (`.db`)

### Multi-postes simultanés
- Mode WAL SQLite + `busy_timeout` : plusieurs postes peuvent consulter et saisir en même temps depuis le réseau local

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
├── clefs.db          ← base de données (NE PAS SUPPRIMER)
├── backups/          ← sauvegardes
└── documents/        ← PDF et CSV générés
```

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
    app.go            Application principale, navigation
    dashboard_modern.go
    new_loan_view.go  Vue prêt unifiée (portes + trousseau)
    loan_dialogs.go   Dialogue retour de clé
    accesses.go       Gestion des accès (portes/zones)
    keys.go           Gestion des clés
    borrowers.go      Gestion des détenteurs
    loans.go          Emprunts actifs
    history.go        Historique filtrable
    views_transversal.go  Vues rapides (consultation)
    redundancy_view.go    Vue redondances
    widgets.go        Composants réutilisables (cards, checkrows)
    config.go         Configuration, sauvegarde, import
    keyplan.go        Plan de clés
    help.go           Mode d'emploi
    theme_simple.go   Thème visuel

  pdf/
    generator.go      Génération PDF (bons, rapports)
    exporter.go       Sauvegarde fichier PDF
```

### Lancer en développement

```bash
go run ./cmd/main.go
```

### Créer une release Windows

```bash
./create-release.sh 3.1.0
```

Le script committe les modifications, crée le tag et pousse vers GitHub. Le workflow Actions compile l'exécutable Windows et publie la release automatiquement.

---

## Licence

Distribué sous licence **MIT**.

Développé par **COLLET David** — Secrétaire Général, Collège Victor Hugo, Chauny (02300).
