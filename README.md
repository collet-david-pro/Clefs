# 🔑 Gestionnaire de Clés - Version Go

Application de gestion de clés et d'emprunts, portée de Python vers Go avec interface graphique native Fyne.

## 📋 Vue d'Ensemble

Cette application permet de :
- ✅ Gérer un inventaire de clés avec quantités et réserves
- ✅ Suivre les emprunts et retours de clés
- ✅ Gérer les emprunteurs, bâtiments et salles
- ✅ Générer des reçus PDF avec support UTF-8 complet
- ✅ Visualiser les rapports et le plan de clés
- ✅ Compiler pour Windows, macOS et Linux sans dépendances CGO


## 📁 Structure du Projet

```
go_app/
├── cmd/
│   └── main.go              # Point d'entrée
├── internal/
│   ├── db/                  # Couche base de données
│   │   ├── database.go      # Connexion SQLite
│   │   ├── models.go        # Modèles de données
│   │   └── queries.go       # Requêtes SQL
│   ├── gui/                 # Interface Fyne
│   │   ├── app.go           # Application principale
│   │   ├── dashboard.go     # Tableau de bord
│   │   ├── keys.go          # Gestion des clés
│   │   ├── borrowers.go     # Gestion des emprunteurs
│   │   ├── buildings.go     # Gestion des bâtiments
│   │   ├── rooms.go         # Gestion des salles
│   │   ├── loans.go         # Gestion des emprunts
│   │   ├── keyplan.go       # Plan de clés
│   │   ├── reports.go       # Rapports
│   │   └── utils.go         # Utilitaires GUI
│   └── pdf/
│       └── generator.go     # Génération de PDFs
├── clefs.db                 # Base de données SQLite
├── go.mod                   # Dépendances Go
└── README.md               # Ce fichier
```

## 🎯 Fonctionnalités

### 1. Tableau de Bord
- Vue d'ensemble de toutes les clés avec tableau
- Calcul automatique de la disponibilité
- Actions rapides (Emprunter/Retourner)
- Affichage des emprunteurs actuels
- Interface optimisée avec colonnes fixes

### 2. Gestion des Clés
- Création, modification, suppression
- Quantités totales et réservées
- Lieu de stockage
- Association avec des salles (many-to-many)

### 3. Gestion des Emprunteurs
- Nom et email
- Historique des emprunts

### 4. Gestion des Bâtiments et Salles
- Organisation hiérarchique
- Types de salles
- Associations avec les clés

### 5. Emprunts
- Création d'emprunts simples ou multiples
- Vérification automatique de disponibilité
- Retour de clés avec sélection si multiples emprunts
- Horodatage automatique

### 6. Rapports
- Emprunts actifs groupés par emprunteur
- Plan de clés (bâtiments → salles → clés)
- Rapport des clés sorties

### 7. Génération de PDFs
- Reçus d'emprunt individuels
- Reçus groupés par emprunteur
- Plan de clés exportable
- Rapport des emprunts
- **Support complet UTF-8** (caractères accentués)

### 8. 💾 Gestion des Sauvegardes 
- **Liste complète** des sauvegardes avec date, heure et taille
- **Restauration** en un clic avec sauvegarde automatique de sécurité
- **Suppression** des anciennes sauvegardes
- **Création rapide** de nouvelles sauvegardes
- **Importation depuis Python** - Migrez facilement vos données de l'ancienne version
- Interface dédiée accessible depuis Configuration
- Sauvegardes exportables

### 9. 🚀 Releases Automatiques Multi-Plateformes 
- Support de **3 plateformes** :
  - Windows x64 (compatible x86)
  - macOS Intel (amd64)
  - macOS Apple Silicon (arm64)


### 10. 📖 Mode d'Emploi Intégré 
- Guide d'utilisation complet dans l'application
- Instructions pas à pas pour chaque fonctionnalité
- Accessible depuis le menu principal

## 🛠️ Technologies Utilisées

### Backend
- **Go 1.21+** : Langage principal
- **modernc.org/sqlite** : Driver SQLite pure Go (sans CGO)
- **Database/sql** : Interface standard Go pour SQL

### Frontend
- **Fyne v2.4.5** : Framework GUI cross-platform
- Interface native sur chaque OS
- Responsive et moderne

### PDF
- **github.com/phpdave11/gofpdf** : Génération de PDFs
- Support UTF-8 avec UnicodeTranslator
- Mise en page professionnelle


## 🗄️ Base de Données

### Schéma

**Tables** :
- `keys` : Clés avec quantités et stockage
- `borrowers` : Emprunteurs
- `buildings` : Bâtiments
- `rooms` : Salles/Pièces
- `loans` : Emprunts avec dates
- `key_room_association` : Table de liaison many-to-many

### Localisation
La base de données `clefs.db` est créée automatiquement dans le répertoire de l'application.

### 💾 Gestion des Sauvegardes 

L'application intègre maintenant un système complet de gestion des sauvegardes :

**Via l'interface graphique** :
1. Aller dans **Configuration**
2. Cliquer sur **📋 Gérer les Sauvegardes**
3. Utiliser l'interface pour :
   - Lister toutes les sauvegardes
   - Créer une nouvelle sauvegarde
   - Restaurer une sauvegarde
   - Supprimer d'anciennes sauvegardes


**Emplacement** : Les sauvegardes sont stockées dans `backups/`

**Format des noms** : `clefs_backup_AAAAMMJJ_HHMMSS.db`

### 📥 Importation depuis la Version Python 

Si vous utilisez l'ancienne version Python de l'application, vous pouvez facilement importer toutes vos données :

**Via l'interface graphique** :
1. Aller dans **Configuration**
2. Cliquer sur **📥 Importer depuis Version Python**
3. Sélectionner votre fichier `clefs.db` issue de la version python.
4. Confirmer l'importation

**Ce qui est importé** :
- ✅ Tous les bâtiments
- ✅ Toutes les salles/points d'accès
- ✅ Toutes les clés avec quantités et associations
- ✅ Tous les emprunteurs
- ✅ Tous les emprunts (actifs et historique)

**Sécurité** : Une sauvegarde automatique de votre base actuelle est créée avant l'importation.

**Note** : Les doublons sont automatiquement ignorés (basé sur les IDs).



## 🔄 Migration depuis Python

### Différences Principales

| Aspect | Python (Original) | Go (Nouveau) |
|--------|------------------|--------------|
| Framework Web | FastAPI | Fyne (GUI native) |
| Base de données | SQLAlchemy | database/sql |
| Driver SQLite | sqlite3 (CGO) | modernc.org/sqlite (Pure Go) |
| Templates | Jinja2 | Widgets Fyne |
| PDF | ReportLab | gofpdf |
| Packaging | PyInstaller | Go build natif |

### Avantages de la Version Go

✅ **Performance** : Exécution native, pas d'interpréteur
✅ **Taille** : ~20 MB vs ~50+ MB avec PyInstaller
✅ **Déploiement** : Un seul exécutable, pas de dépendances
✅ **Cross-compilation** : Build pour toutes les plateformes depuis un seul OS
✅ **Maintenance** : Typage statique, moins de bugs runtime
✅ **Interface** : GUI native au lieu de navigateur web


---

**Version** : 2.0.0  
**Date** : Décembre 2024  
**Langage** : Go 1.21+  
**Plateformes** : Windows x64, macOS (Intel & Apple Silicon)

---

COLLET David, cette application aurait été impossible à créer pour moi sans IA. 
