# Historique des versions

## 4.0.0 — Refonte complète

Une nouvelle application, plus moderne et plus simple à utiliser, qui reprend sans perte les données des versions 1, 2 et 3.

### Nouvelle interface
- Interface entièrement redessinée (Go + Wails + Svelte) : menu latéral, tableau de bord orienté actions, mode clair et sombre.
- **Recherche globale** (`Ctrl + K`) sur les clés, les portes et les personnes.
- **Fiches détaillées** en panneau latéral pour chaque clé, accès et détenteur : stock, accès, détenteurs, historique et actions.
- Formulaires guidés avec messages d'erreur clairs, confirmations avant les actions sensibles, notifications discrètes.
- **Aide intégrée** pas à pas, avec accès direct à chaque écran, et bouton **Quitter**.

### Prêts
- Écran de prêt en 3 étapes : la personne, les portes à ouvrir, le trousseau calculé en direct.
- Type de prêt, date de retour prévue et agent désormais **enregistrés** (les retards sont détectés).
- Retours groupés (« Tout rendre ») avec état constaté.

### Inventaire (nouveau)
- Onglet **Inventaire** : on saisit les exemplaires présents ; nouveau total = comptés + prêtés.
- Saisie rapide au clavier, écarts en direct, **rapport d'inventaire PDF**.
- Un total de 0 est accepté (clés toutes perdues ou cassées).

### Documents
- **Tous les PDF refaits** au modèle de l'application : police Inter, couleurs du thème, tableaux lisibles, pagination.
- **Bon de remise sur une seule page A4**, quel que soit le nombre de clés.
- **Tous les bons en un seul PDF** (sommaire + un bon par détenteur) pour l'archivage.
- Nouveau rapport d'inventaire ; bilan du stock et plan de clés redessinés.
- Tri naturel des numéros de clés (1, 2, … 10).

### Paramètres (nouveau)
- Nom et adresse de l'établissement, nom de l'application, **logo**.
- Titre et texte d'engagement des bons, libellés des signatures, signataire, mention de pied de page.
- Type de prêt et durée de retour proposés par défaut, nom de l'agent.
- Communs à tous les postes et inclus dans les sauvegardes.

### Données
- **Assistant d'import** des bases V1, V2 et V3 avec détection de la version et aperçu.
- Corrections de la V3 : statut et téléphone des détenteurs, catégorie et notes des clés n'étaient pas enregistrés.
- Suppression impossible d'une clé prêtée ou d'une personne qui détient des clés.
- Sauvegarde fonctionnelle même avec une apostrophe dans le chemin du dossier.

### Déploiement
- Windows x64 uniquement, un seul exécutable, **rien à installer** (WebView2 fourni avec Windows 10 et 11).
- macOS n'est plus pris en charge.

## 3.3.0
- Bon de remise fiabilisé et réédition en un clic, alerte « erreur d'inventaire », correction critique de la restauration de sauvegarde, build macOS Apple Silicon.

## 3.2.x et antérieures
- Voir l'historique Git.
