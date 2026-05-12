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

![Version](https://img.shields.io/badge/version-2.1.0-blue.svg)
![Plateformes](https://img.shields.io/badge/plateformes-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)
![Licence](https://img.shields.io/badge/Licence-MIT-green.svg)

Cette nouvelle version (V2) est une **refonte complète** de l'application "Gestionnaire de Clés". L'application a été réécrite en **Go** avec le framework **Fyne** pour offrir une expérience **100% native, rapide et multi-plateforme**.

---

## 🌟 Nouveautés de la Version 2

Par rapport à l'ancienne version V1 (Python), cette version apporte des améliorations majeures :

-   **Application Native Multi-plateforme** : Un seul exécutable pour Windows, macOS et Linux, sans dépendre d'un navigateur web.
-   **Interface Moderne et Rapide** : Interface entièrement repensée, plus intuitive et réactive grâce à Fyne.
-   **Gestion des Données Intégrée** :
    -   **Sauvegarde & Restauration** : Créez, listez, restaurez et supprimez des sauvegardes directement depuis l'application.
    -   **Importation Facile** : Un outil dédié permet de migrer toutes vos données de l'ancienne base de données V1 (Python) en quelques clics.
-   **Automatisation Poussée** :
    -   Les dossiers `documents/` (pour les PDF) et `backups/` sont créés automatiquement.
    -   La génération de PDF se fait instantanément dans le dossier `documents`, sans boîte de dialogue.
-   **Mode d'Emploi Intégré** : Un guide complet est disponible directement dans l'application pour vous aider à maîtriser toutes les fonctionnalités.
-   **Aucune Installation Requise** : L'application est portable. Il suffit de la télécharger et de la lancer.

---

## 🚀 Installation

L'application ne nécessite aucune installation. Il suffit de la télécharger et de la placer dans un dossier dédié.

1.  Rendez-vous sur la page [**Releases**](https://github.com/votre-nom/votre-repo/releases) de ce projet.
2.  Téléchargez l'archive (`.zip` ou `.tar.gz`) correspondant à votre système.
3.  **Très important** : Extrayez l'archive et placez l'exécutable et le fichier `infos.txt` dans un **dossier qui lui sera dédié** (par exemple, `C:\Apps\Clefs` ou `~/Documents/Clefs`).

### Windows
-   Double-cliquez simplement sur le fichier `clefs-windows-amd64.exe` pour lancer l'application. Windows Defender ou votre antivirus peut afficher une alerte car l'exécutable n'est pas signé par une autorité reconnue. Vous pouvez l'ignorer en toute sécurité.

### macOS & Linux
1.  Ouvrez un terminal dans le dossier où se trouve l'application.
2.  Rendez l'exécutable exécutable avec la commande `chmod +x`.
    -   *Exemple sur macOS* : `chmod +x clefs-macos-amd64`
    -   *Exemple sur Linux* : `chmod +x clefs-linux-amd64`
3.  Lancez l'application depuis le terminal.
    -   *Exemple* : `./clefs-macos-amd64`

---

## 🔄 Migration depuis la V1 (Python)

Vous utilisiez l'ancienne version ? Vous pouvez récupérer **toutes** vos données en quelques secondes.

1.  **Sauvegardez votre ancienne base de données** : Localisez le fichier `clefs.db` de votre ancienne installation (version Python) et copiez-le dans un endroit sûr.
2.  **Lancez la nouvelle application (V2)** : Installez et ouvrez la nouvelle version en Go.
3.  **Allez dans l'outil d'importation** : Dans le menu, allez dans `Configuration` -> `Importer depuis V1 (Python)`.
4.  **Sélectionnez votre ancien fichier** : Cliquez sur le bouton pour choisir un fichier et sélectionnez la copie de votre ancien `clefs.db` que vous aviez sauvegardé.
5.  **Validez** : L'application importera tous vos bâtiments, salles, clés, emprunteurs et historiques d'emprunts. Un résumé de l'importation s'affichera.

---

## 💡 Guide d'Utilisation

### Premier Lancement
Au premier démarrage, l'application crée automatiquement les éléments suivants dans son dossier :
-   `clefs.db` : Le nouveau fichier de base de données.
-   `documents/` : Le dossier où tous les PDF générés seront stockés.
-   `backups/` : Le dossier pour les sauvegardes manuelles ou automatiques.

### ⚠️ Utilisation en Réseau et Multi-utilisateurs
-   **Réseau** : Vous pouvez placer le dossier de l'application sur un partage réseau pour y accéder depuis différents postes.
-   **Multi-accès (IMPORTANT)** : L'application **n'est pas conçue pour être ouverte par plusieurs utilisateurs en même temps**. Si deux personnes ou plus utilisent l'application simultanément sur la même base de données, cela **entraînera une corruption irréversible des données**. Assurez-vous qu'une seule instance est active à la fois.

---

## 👨‍💻 Pour les Développeurs

### Prérequis
-   Go 1.21+
-   Les dépendances du framework Fyne. Consultez [la documentation de Fyne](https://developer.fyne.io/started/) pour les installer sur votre système (ex: `xorg-dev` sur Linux, `xcode` sur macOS).

---

## 📜 Licence

Ce projet est distribué sous la **Licence MIT**.


--- 

## Ajout de fonctionnalités envisagées

- Import d'une base de donnée excel ou csv pour la liste des utilisateurs (avec un fichier modèle founi dans l'application)