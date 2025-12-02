# Gestionnaire de Clés

Application web simple et complète pour la gestion des clés, des stocks, des emprunts et des droits d'accès au sein d'un établissement.

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

## Installation et Lancement

Cette application est conçue pour fonctionner localement sur votre machine (macOS, Linux, Windows avec un interpréteur bash).

### Prérequis

- **Python 3** (version 3.7 ou supérieure).
- `pip` pour l'installation des dépendances.

### Instructions

1.  **Rendre les scripts exécutables :**
    Ouvrez un terminal dans le dossier du projet et lancez cette commande une seule fois :
    ```bash
    chmod +x start.sh seed.sh
    ```

2.  **Lancer l'application :**
    Pour démarrer l'application, exécutez simplement :
    ```bash
    ./start.sh
    ```
    Ce script s'occupe d'installer les dépendances nécessaires, de lancer le serveur web et d'ouvrir automatiquement l'application dans votre navigateur.

3.  **Arrêter l'application :**
    Retournez dans le terminal où vous avez lancé le script et appuyez sur `Ctrl+C`.

## Utilisation

### Remplir avec des données de test (Optionnel)

> **ATTENTION :** Le script `seed.sh` est uniquement destiné à des fins de **test** ou de **démonstration**. Il **supprime toutes les données existantes** dans la base de données avant de la remplir avec un jeu de données aléatoires. Ne l'utilisez pas si vous avez déjà entré des données réelles que vous souhaitez conserver.

Pour peupler l'application avec des données de test :
1.  Assurez-vous que l'application est arrêtée.
2.  Lancez le script `./seed.sh` et confirmez avec `o`.
3.  Redémarrez l'application avec `./start.sh`.

### Démarrage et Configuration Manuelle

Pour une utilisation normale, vous pouvez configurer l'application manuellement :
1.  Lancez l'application avec `./start.sh`.
2.  Allez dans l'onglet **Configuration**.
3.  Suivez l'ordre logique : créez d'abord vos **Emprunteurs**, vos **Bâtiments** et vos **Points d'Accès**.
4.  Enfin, créez vos **Clés** en les liant aux points d'accès.

### Gestion des Données
Toute la gestion des données de base (Clés, Emprunteurs, Bâtiments, Points d'accès) se fait depuis la page **Configuration**.

### Emprunts et Retours
- **Pour emprunter :** Depuis le tableau de bord, cliquez sur "Nouvel Emprunt". Le formulaire vous présente une **liste de cases à cocher** pour sélectionner facilement une ou plusieurs clés pour un emprunteur.
- **Pour retourner :** Depuis le tableau de bord ou la page "Emprunts en Cours", cliquez sur "Retourner". Si plusieurs personnes ont ce type de clé, sélectionnez l'emprunt exact à clôturer.
