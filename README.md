# 🔑 Gestionnaire de Clés - Version Go

Application de gestion de clés et d'emprunts, portée de Python vers Go avec interface graphique native Fyne.

## 📋 Vue d'Ensemble

Cette application permet de :
- ✅ Gérer un inventaire de clés avec quantités et réserves
- ✅ Suivre les emprunts et retours de clés
- ✅ Gérer les emprunteurs, bâtiments et salles
- ✅ Générer des reçus PDF
- ✅ Visualiser les rapports et le plan de clés


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

### 2. 🎨 Interface  
- **Emprunts en Cours** : Vue  par emprunteur avec déploiement/repliement
- **Clés** : Vue  par clé avec statut de disponibilité et emprunts actifs
- **Rapport des Clés Sorties** : Vue  groupée par clé avec liste des emprunteurs
- **Mode d'Emploi** :
- Interface compacte et intuitive
- Indicateurs visuels (nombre d'éléments, durées, alertes)

### 3. Gestion des Clés
- Création, modification, suppression
- Quantités totales et réservées
- Lieu de stockage
- Association avec des salles (many-to-many)
- **Vue accordéon** avec statut de disponibilité
- **Alertes visuelles** : ⚠️ STOCK ÉPUISÉ si disponibilité = 0
- Liste des emprunts actifs par clé

### 4. Gestion des Emprunteurs
- Nom et email
- Historique des emprunts
- Vue groupée par emprunteur

### 5. Gestion des Bâtiments et Salles
- Organisation hiérarchique
- Types de salles
- Associations avec les clés

### 6. Emprunts
- Création d'emprunts simples ou multiples
- Vérification automatique de disponibilité
- Retour de clés avec sélection si multiples emprunts
- Horodatage automatique
- Vue par emprunteur avec détails déployables

### 7. Rapports
- Emprunts actifs groupés par emprunteur
- Plan de clés (bâtiments → salles → clés)
- Rapport des clés sorties
- Vue pour tous les rapports

### 8. 📄 Génération de PDFs Automatique
- **Enregistrement automatique** dans `./documents/`
- **Pas de dialogue de sauvegarde** : génération instantanée
- **Notifications** avec chemin complet du fichier
- **Dossier créé au démarrage** : `./documents/` créé automatiquement

#### Types de PDFs Disponibles
- Reçus d'emprunt individuels
- Reçus groupés par emprunteur
- Rapport des clés sorties
- Rapport global par emprunteur
- Bilan des clés (stock)
- Plan de clés complet

#### Structure des Fichiers
```
Clefs/
├── clefs.exe (ou clefs)
├── clefs.db (créé automatiquement)
├── backups/ (sauvegardes automatiques)
└── documents/ (créé au démarrage)
    ├── recu_emprunt_123_20251204_215538.pdf
    ├── rapport_cles_sorties_20251204_220015.pdf
    ├── rapport_global_emprunts_20251204_220130.pdf
    └── ...
```

### 9. 💾 Gestion des Sauvegardes 
- **Liste complète** des sauvegardes avec date, heure et taille
- **Restauration** en un clic avec sauvegarde automatique de sécurité
- **Suppression** des anciennes sauvegardes
- **Création rapide** de nouvelles sauvegardes
- **Importation depuis Python** : Migrez facilement vos données de l'ancienne version
- Interface dédiée accessible depuis Configuration
- Sauvegardes exportables

### 10. 🚀 Releases Automatiques 
- Support actuel :
  - Windows x64 (compatible x86)
  - **macOS** : Support en cours de développement, disponible prochainement

### 11. 📖 Mode d'Emploi Intégré
- **Interface accordéon** avec 10 sections
- Guide d'utilisation complet dans l'application
- Instructions pas à pas pour chaque fonctionnalité
- Accessible depuis le menu principal
- Sections : Démarrage, Tableau de Bord, Emprunts, Clés, Sauvegardes, PDFs, Configuration, Astuces, Navigation, Support

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
- **Enregistrement automatique** dans `./documents/`
- **Notifications** avec chemin complet


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


---

**Version** : 2.1.0  
**Date** : Décembre 2024  
**Langage** : Go 1.21+  
**Plateformes** : Windows x64 (macOS disponible prochainement)

---

COLLET David, cette application aurait été impossible à créer pour moi sans IA.
