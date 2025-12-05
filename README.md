# Information importante

La version actuelle est la V1x disponible dans les Releases, les versions 2.X sont en test.


# Gestionnaire de Clés

Application de bureau simple et complète pour la gestion des clés, des stocks, des emprunts et des droits d'accès au sein d'un établissement.

## Fonctionnalités

- **Tableau de Bord :** Vue d'ensemble en temps réel du statut de toutes les clés (disponibilité, stock, qui a emprunté quoi).
- **Gestion des Clés :**
    - Créez, modifiez et supprimez des types de clés.
    - Définissez un **lieu de stockage** (ex: Accueil, Administration...).
    - Gérez un stock fin avec :
        - **Nombre total de clés** : Le nombre total de clés de ce type en votre possession
        - **Nombre de clés en réserve** : Les clés placées en réserve (non disponibles au prêt)
        - Le système calcule automatiquement les clés disponibles au prêt : `Disponibles = Total - Réserve`
    - Interface claire avec labels explicites et textes d'aide pour éviter toute confusion
- **Gestion des Emprunteurs :** Maintenez une liste des personnes autorisées à emprunter des clés.
- **Gestion de la Configuration :**
    - Définissez les **Bâtiments** de votre établissement.
    - Créez tous les **Points d'Accès** (salles, portes, entrées, armoires...) et liez-les à un bâtiment.
- **Liaison Clés <-> Accès :** Lors de la création ou de la modification d'une clé, cochez simplement tous les points d'accès qu'elle peut ouvrir.
- **Plan de Clés :** Un outil puissant pour visualiser les relations entre clés et points d'accès.
    - **Vue par Clé :** Affichez tous les lieux qu'une clé spécifique peut ouvrir.
    - **Vue par Point d'Accès :** Affichez toutes les clés qui peuvent ouvrir un lieu spécifique.
- **Système d'Emprunt et de Retour :**
    - Empruntez une ou plusieurs clés pour une personne en une seule fois via une **liste à cocher** intuitive.
    - Le système vérifie le stock utilisable et empêche l'emprunt de clés non disponibles.
    - Lors du retour, si plusieurs personnes ont le même type de clé, une page de sélection vous permet de choisir précisément quel emprunt clôturer.
- **Génération de PDF :**
    - **PDF individuel** : Un bon de sortie en PDF est généré pour chaque emprunt individuel, prêt à être signé. En effet, un utilisateur peut simplement avoir besoin d'une clé en plus pour uen période donnée.
    - **PDF groupé** : Générez un document unique avec toutes les clés empruntées par une personne, idéal pour une signature groupée.
- **Liste des Emprunts en Cours :** Une page dédiée, **groupée par personne**, pour voir rapidement qui a quoi et pour réimprimer les bons de sortie (individuels ou groupés).
- **Rapport Complet des Clés Sorties :**
    - Vue d'ensemble de toutes les clés actuellement empruntées et donc en circulation.
    - Indicateurs de durée d'emprunt avec code couleur (vert=aujourd'hui, bleu=1-6j, jaune=7-29j, rouge=30+j).
    - Résumé groupé par emprunteur.
    - Fonction d'impression/export PDF pour archivage ou présentation.
- **Autonome:** Fonctionne comme une application native sur Windows, sans nécessiter de navigateur externe ni de connexion internet. L'application peut se trouver sur le réseau, mais attention, vous ne pouvez pas ouvrir l'application à plusieurs sous risque de corruption de données. 

    **ATTENTION, LE FICHIER EST AUTO-SIGNÉ, WINDOWS OU VOTRE ANTIVIRUS VOUS DONNERA UNE ALERTE PROBABLEMENT**

## Installation (pour les utilisateurs)

L'application est disponible pour Windows.

1.  Allez sur la **page des Releases** de ce projet.
2.  Téléchargez le fichier `.zip`.
3.  Décompressez le fichier.
4.  Mettez le dans un dossier dédié.
5.  Lancez l'exécutable.

## Fonctionnement

Lors du premier lancement de l'application, un fichier de base de données nommé `clefs.db` est automatiquement créé dans le même dossier que l'exécutable. **Ce fichier est essentiel** car il stocke toutes les informations : les clés, les emprunteurs, les prêts, etc.

- **Ne supprimez pas** ce fichier, sinon vous perdrez toutes vos données.
- Si vous déplacez l'application, déplacez également le fichier `clefs.db` avec elle.
- Pour faire une sauvegarde, il vous suffit de copier le fichier `clefs.db`.

## Développement (pour les ceux qui veulent regarder le code)

### Prérequis

- **Python 3** (version 3.7 ou supérieure).
- `pip` pour l'installation des dépendances.
- Un environnement virtuel est fortement recommandé.

### Instructions

1.  **Clonez le dépôt :**
    ```bash
    git clone https://github.com/collet-david-pro/Clefs.git
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


## Licence

Ce projet est sous licence MIT.

## TODO 

- Version MacOS (ARM)
- Tester l'application en reseau en ouvrant 2 instances en même temps
- Changer la licence dans l'application pour MIT
