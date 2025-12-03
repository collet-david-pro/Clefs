# Gestionnaire de Clés

Application de bureau simple et complète pour la gestion des clés, des stocks, des emprunts et des droits d'accès au sein d'un établissement.

## Fonctionnalités

- **Tableau de Bord :** Vue d'ensemble en temps réel du statut de toutes les clés (disponibilité, stock, qui a emprunté quoi).
- **Gestion des Clés (CRUD) :**
    - Créez, modifiez et supprimez des types de clés.
    - Définissez un **lieu de stockage** (ex: Accueil, Administration...).
    - Gérez un stock fin avec une **quantité totale** et une **quantité en réserve**. Seul le stock "utilisable" (`total - réserve`) est disponible à l'emprunt.
- **Gestion des Emprunteurs :** Maintenez une liste des personnes autorisées à emprunter des clés.
- **Gestion de la Configuration :**
    - Définissez les **Bâtiments** de votre établissement.
    - Créez tous les **Points d'Accès** (salles, portes, entrées...) et liez-les à un bâtiment.
- **Liaison Clés <-> Accès :** Lors de la création ou de la modification d'une clé, cochez simplement tous les points d'accès qu'elle peut ouvrir.
- **Plan de Clés :** Un outil puissant pour visualiser les relations entre clés et points d'accès.
    - **Vue par Clé :** Affichez tous les lieux qu'une clé spécifique peut ouvrir.
    - **Vue par Point d'Accès :** Affichez toutes les clés qui peuvent ouvrir un lieu spécifique.
- **Système d'Emprunt et de Retour :**
    - Empruntez une ou plusieurs clés pour une personne en une seule fois via une **liste à cocher** intuitive.
    - Le système vérifie le stock utilisable et empêche l'emprunt de clés non disponibles.
    - Lors du retour, si plusieurs personnes ont le même type de clé, une page de sélection vous permet de choisir précisément quel emprunt clôturer.
- **Génération de PDF :** Un bon de sortie en PDF est généré pour chaque emprunt individuel, prêt à être signé.
- **Liste des Emprunts en Cours :** Une page dédiée, **groupée par personne**, pour voir rapidement qui a quoi et pour réimprimer les bons de sortie.
- **Autonome et Multi-plateforme :** Fonctionne comme une application native sur Windows et macOS, sans nécessiter de navigateur externe ni de connexion internet.

## Installation (pour les utilisateurs)

L'application est disponible pour Windows et macOS.

1.  Allez sur la **page des Releases** de ce projet.
2.  Téléchargez le fichier `.zip` correspondant à votre système d'exploitation (`GestionnaireCles-Windows.zip` ou `GestionnaireCles-macOS.zip`).
3.  Décompressez le fichier.
4.  Lancez l'exécutable (`GestionnaireCles.exe` sur Windows, `GestionnaireCles.app` sur macOS).

## Fonctionnement

Lors du premier lancement de l'application, un fichier de base de données nommé `clefs.db` est automatiquement créé dans le même dossier que l'exécutable. **Ce fichier est essentiel** car il stocke toutes les informations : les clés, les emprunteurs, les prêts, etc.

- **Ne supprimez pas** ce fichier, sinon vous perdrez toutes vos données.
- Si vous déplacez l'application, déplacez également le fichier `clefs.db` avec elle.
- Pour faire une sauvegarde, il vous suffit de copier le fichier `clefs.db`.

## Développement (pour les contributeurs)

### Prérequis

- **Python 3** (version 3.7 ou supérieure).
- `pip` pour l'installation des dépendances.
- Un environnement virtuel est fortement recommandé.

### Instructions

1.  **Clonez le dépôt :**
    ```bash
    git clone https://github.com/votre-nom/votre-repo.git
    cd Clefs
    ```

2.  **Créez un environnement virtuel et installez les dépendances :**
    ```bash
    python3 -m venv venv
    source venv/bin/activate  # Sur macOS/Linux
    # venv\Scripts\activate    # Sur Windows
    pip install -r requirements.txt
    ```

3.  **Lancer l'application en mode développement :**
    ```bash
    python app/main.py
    ```
    Cela lancera le serveur avec le rechargement automatique et ouvrira la fenêtre de l'application.

### Remplir avec des données de test
Le script `seed.sh` permet de peupler la base de données avec des données de démonstration.
> **Attention :** Ce script supprime toutes les données existantes.
```bash
chmod +x seed.sh
./seed.sh
```
