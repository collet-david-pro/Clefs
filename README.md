# 🔑 Gestionnaire de Clés - V2 (Version Go)

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