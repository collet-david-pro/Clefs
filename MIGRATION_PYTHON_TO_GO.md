# 📦 Guide de Migration : Python vers Go

## Vue d'ensemble

Ce document explique comment migrer vos données de l'ancienne version Python du Gestionnaire de Clés vers la nouvelle version Go.

## 🎯 Pourquoi Migrer ?

La nouvelle version Go offre :
- ✅ **Performances améliorées** - Application plus rapide et réactive
- ✅ **Exécutable unique** - Pas besoin d'installer Python ou des dépendances
- ✅ **Multi-plateforme** - Windows, macOS Intel, macOS Apple Silicon
- ✅ **Interface modernisée** - Design amélioré avec Fyne
- ✅ **Nouvelles fonctionnalités** - Gestion des sauvegardes, releases automatiques, etc.

## 📋 Prérequis

- Avoir l'ancienne version Python installée avec des données
- Avoir téléchargé et installé la nouvelle version Go
- Connaître l'emplacement de votre fichier `clefs.db` de la version Python

## 🚀 Processus de Migration

### Étape 1 : Localiser votre Base de Données Python

Votre ancienne base de données se trouve généralement dans :
- **macOS/Linux** : `~/Documents/clefs.db` ou dans le dossier de l'application Python
- **Windows** : `C:\Users\VotreNom\Documents\clefs.db`

Le fichier peut avoir différents noms :
- `clefs.db`
- `clefs - old.db`
- `database.db`

### Étape 2 : Lancer la Nouvelle Application Go

1. Ouvrez l'application Gestionnaire de Clés (version Go)
2. L'application démarre avec une base de données vide

### Étape 3 : Importer vos Données

1. Dans le menu principal, cliquez sur **⚙️ Configuration**
2. Dans la section "Sauvegarde et Restauration", cliquez sur **📥 Importer depuis Version Python**
3. Une fenêtre de sélection de fichier s'ouvre
4. Naviguez jusqu'à votre fichier `clefs.db` de la version Python
5. Sélectionnez le fichier et cliquez sur **Ouvrir**
6. Lisez le message de confirmation qui explique :
   - Une sauvegarde automatique sera créée
   - Les données seront fusionnées
   - Les doublons seront ignorés
7. Cliquez sur **Confirmer** pour lancer l'importation

### Étape 4 : Vérification

Après l'importation :
1. L'application se rafraîchit automatiquement
2. Vous devriez voir toutes vos données dans le tableau de bord
3. Vérifiez que :
   - ✅ Toutes les clés sont présentes
   - ✅ Les emprunteurs sont listés
   - ✅ Les bâtiments et salles sont importés
   - ✅ Les emprunts actifs sont affichés

## 📊 Données Importées

L'importation inclut **TOUTES** vos données :

### ✅ Bâtiments
- Tous les bâtiments avec leurs noms

### ✅ Salles/Points d'Accès
- Toutes les salles avec :
  - Nom
  - Type
  - Association au bâtiment

### ✅ Clés
- Toutes les clés avec :
  - Numéro
  - Description
  - Quantité totale
  - Quantité en réserve
  - Lieu de stockage
  - Associations aux salles

### ✅ Emprunteurs
- Tous les emprunteurs avec :
  - Nom
  - Email

### ✅ Emprunts
- Tous les emprunts (actifs et historique) avec :
  - Date d'emprunt
  - Date de retour (si applicable)
  - Clé empruntée
  - Emprunteur

## 🔒 Sécurité

### Sauvegarde Automatique

Avant chaque importation, le système :
1. Crée automatiquement une sauvegarde de votre base actuelle
2. La stocke dans le dossier `backups/`
3. Nomme le fichier avec la date et l'heure : `clefs_backup_AAAAMMJJ_HHMMSS.db`

### Gestion des Doublons

- Les données sont importées avec `INSERT OR IGNORE`
- Si un ID existe déjà, il est ignoré
- Aucune donnée n'est écrasée
- Les nouvelles données sont ajoutées

## 🛠️ Dépannage

### Problème : "Le fichier de base de données Python n'existe pas"

**Solution** :
- Vérifiez que vous avez sélectionné le bon fichier
- Assurez-vous que le fichier a l'extension `.db`
- Vérifiez les permissions de lecture du fichier

### Problème : "Erreur lors de l'ouverture de la base Python"

**Solution** :
- Le fichier peut être corrompu
- Essayez d'ouvrir le fichier avec un outil SQLite pour vérifier son intégrité
- Utilisez une sauvegarde de votre base Python si disponible

### Problème : "Erreur lors de la lecture des [table]"

**Solution** :
- Le schéma de votre base Python peut être différent
- Contactez le support avec le message d'erreur complet
- Une mise à jour peut être nécessaire pour supporter votre version

### Problème : Données manquantes après l'importation

**Solution** :
1. Vérifiez que toutes les données étaient présentes dans la base Python
2. Consultez les logs de l'application pour voir le résumé de l'importation
3. Les doublons (même ID) sont automatiquement ignorés

## 📝 Après la Migration

### Recommandations

1. **Vérifiez vos données** - Parcourez toutes les sections pour confirmer l'importation
2. **Créez une sauvegarde** - Utilisez la fonction de sauvegarde pour sécuriser vos données
3. **Testez les fonctionnalités** - Créez un emprunt test pour vérifier le fonctionnement
4. **Conservez l'ancienne base** - Gardez votre fichier Python en backup pendant quelques semaines

### Nouvelles Fonctionnalités à Découvrir

Après la migration, explorez les nouvelles fonctionnalités :

1. **📋 Gestion des Sauvegardes**
   - Liste de toutes vos sauvegardes
   - Restauration en un clic
   - Suppression des anciennes sauvegardes

2. **📊 Dashboard Amélioré**
   - Tableau avec colonnes alignées
   - Affichage clair de la disponibilité
   - Actions rapides (Emprunter/Retourner)

3. **📖 Mode d'Emploi Intégré**
   - Guide complet accessible depuis le menu
   - Explications détaillées de toutes les fonctionnalités

4. **🎮 Mode Démonstration**
   - Testez l'application avec des données de démo
   - Parfait pour la formation

## 🆘 Support

Si vous rencontrez des problèmes lors de la migration :

1. **Consultez les logs** - L'application affiche des messages détaillés
2. **Vérifiez le fichier source** - Assurez-vous que votre base Python est valide
3. **Utilisez les sauvegardes** - Toutes les importations créent des sauvegardes automatiques
4. **Contactez le support** - Ouvrez une issue sur GitHub avec :
   - Le message d'erreur complet
   - La version de votre ancienne application Python
   - Le système d'exploitation utilisé

## ✅ Checklist de Migration

- [ ] Localiser le fichier `clefs.db` de la version Python
- [ ] Installer la nouvelle version Go
- [ ] Lancer l'application Go
- [ ] Aller dans Configuration
- [ ] Cliquer sur "Importer depuis Version Python"
- [ ] Sélectionner le fichier de la base Python
- [ ] Confirmer l'importation
- [ ] Vérifier que toutes les données sont présentes
- [ ] Créer une sauvegarde de sécurité
- [ ] Tester les fonctionnalités principales
- [ ] Conserver l'ancienne base en backup

## 🎉 Félicitations !

Vous avez réussi à migrer vos données vers la nouvelle version Go du Gestionnaire de Clés !

Profitez des nouvelles fonctionnalités et de l'amélioration des performances.
