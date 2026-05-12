# Cahier des charges - Outil de gestion des clés
## Collège Victor Hugo - Chauny (02300)

**Rédacteur :** David COLLET, Secrétaire Général  
**Date :** Mai 2026  
**Statut :** Version 5 - base de travail

---

## 1. Contexte

Le collège Victor Hugo gère un parc de plus de 50 clés (clés simples, trousseaux, passes) couvrant l'ensemble des locaux : salles de classe, locaux techniques, réserves, gymnase, demi-pension, bureaux administratifs, accès extérieurs.

Aujourd'hui, aucune procédure formalisée n'existe pour le suivi des remises et retours de clés. Les attributions permanentes sont connues de mémoire ou par usage, sans traçabilité écrite. Les prêts ponctuels (intervenants extérieurs, entreprises, agents remplaçants) ne sont pas enregistrés. Ce manque de traçabilité génère des risques en cas de perte ou de litige.

Une version précédente de l'outil existe (V2, Go/Fyne, repo GitHub `collet-david-pro/Clefs`). Elle couvre déjà un socle fonctionnel solide : tableau de bord, gestion des clés avec points d'accès liés, emprunt multi-clés, génération de PDF, rapport des emprunts en cours avec code couleur de durée, sauvegarde/restauration intégrée. La V3 est une évolution directe de cette base, pas une réécriture.

---

## 2. Utilisateurs

| Profil | Rôle dans l'outil |
|---|---|
| Secrétaire Général | Administration complète, référent de l'outil |
| Agent d'accueil désigné | Saisie des remises et retours au quotidien |
| Autres agents | Saisie possible (accès simultané) |
| Chef d'établissement (Principale) | Consultation en lecture seule |

L'outil doit pouvoir être utilisé par plusieurs personnes en même temps depuis des postes différents connectés au réseau local.

---

## 3. Périmètre fonctionnel

### 3.1 Référentiel des accès (portes et zones)

Avant les clés, l'outil référence les accès physiques de l'établissement : chaque porte, portail ou zone verrouillée est enregistré comme un "accès" indépendant.

- Identifiant unique de l'accès
- Désignation (ex. : "Porte salle B12", "Portail cour nord", "Local chaufferie")
- Bâtiment (ex. : bâtiment A, gymnase, demi-pension, externat...)
- Étage ou niveau (RDC, R+1, sous-sol...)
- Catégorie (salle de classe, local technique, bureau, accès extérieur, sanitaires...)
- Observations (accès sensible, accès PMR, alarme associée, etc.)

Ce référentiel permet ensuite de filtrer les accès par bâtiment, par étage ou par catégorie, et de raisonner en termes de "quels accès faut-il ouvrir" avant de choisir les clés.

### 3.2 Référentiel des clés (inventaire)

Chaque clé physique est enregistrée en lien avec la liste des accès qu'elle permet d'ouvrir.

- Identifiant unique par clé (numéro gravé ou code interne)
- Désignation (ex. : "Clé salle B12", "Trousseau cuisine", "Passe direction")
- Catégorie (clé simple, trousseau, badge, passe)
- **Liste des accès ouverts par cette clé** (relation multiple : une clé peut ouvrir plusieurs portes)
- Nombre d'exemplaires existants (stock total)
- Nombre d'exemplaires disponibles (calculé automatiquement : stock - attributions en cours)
- Localisation de référence (où elle doit être rangée quand elle n'est pas en circulation)
- Observations éventuelles (clé fragile, accès sensible, etc.)

### 3.3 Référentiel des détenteurs

- Nom, prénom
- Statut : personnel permanent, personnel contractuel, intervenant extérieur, entreprise
- Coordonnées (optionnel)

### 3.4 Attributions permanentes

Possibilité d'affecter une ou plusieurs clés à un détenteur de manière permanente (pour toute l'année scolaire ou jusqu'à révocation), avec date d'attribution. Un même détenteur peut détenir plusieurs clés simultanément.

### 3.5 Prêts ponctuels

Un prêt est défini par un **besoin d'accès**, pas par une clé. Le processus est le suivant :

1. On saisit le détenteur et la liste des portes/zones auxquelles il doit avoir accès
2. L'outil calcule automatiquement la combinaison de clés la plus réduite possible permettant d'ouvrir exactement ces accès, en tenant compte du stock disponible au moment du prêt
3. L'utilisateur valide ou ajuste la proposition
4. Le prêt est enregistré avec la liste des clés effectivement remises, la date de remise et la date de retour prévue

Lors du retour :
- Enregistrement de la date réelle de retour
- Enregistrement de l'état constaté
- Statut automatique mis à jour : en cours / retourné / en retard

Un détenteur peut avoir plusieurs prêts actifs simultanément (clés différentes).

### 3.6 Contrôle des doublons et redondances

L'outil doit alerter et permettre de détecter les situations suivantes :

- Un détenteur possède plusieurs clés qui ouvrent les mêmes portes (redondance inutile)
- Un détenteur dispose d'accès non justifiés par rapport à son profil ou à sa demande initiale
- Vue "par détenteur" : liste de toutes les clés en sa possession avec la liste des accès cumulés, et signal visuel si des accès se recoupent entre plusieurs de ses clés

### 3.7 Historique des mouvements

- Consultation de l'historique complet d'une clé (qui l'a eue, quand, combien de temps)
- Consultation de l'historique d'un détenteur (quelles clés il a eues ou a encore)
- Filtres : par période, par statut, par clé, par détenteur, par bâtiment, par étage, par accès

### 3.8 Vues et filtres transversaux

L'outil doit proposer les vues suivantes, accessibles rapidement :

- "Quelle clé ouvre quelle(s) porte(s) ?" : liste des accès pour une clé donnée
- "Quelles clés ouvrent cette porte ?" : liste des clés donnant accès à un accès donné
- "Toutes les clés utiles dans ce bâtiment / à cet étage" : filtrage par localisation
- "Qui a quoi en ce moment ?" : liste des détenteurs avec leurs clés actuelles et les accès cumulés
- "Quelle clé est en possession de qui ?" : recherche par clé pour connaître son détenteur actuel
- "Clés disponibles" : stock disponible par clé, filtrable par bâtiment ou type d'accès

### 3.9 Tableau de bord

- Liste des clés actuellement sorties (prêts en cours et attributions permanentes)
- Liste des prêts en retard (date de retour dépassée sans retour enregistré)
- Détenteurs avec redondances d'accès détectées
- Nombre de clés disponibles par catégorie

### 3.10 Edition d'un bon de remise

Génération d'un document imprimable (ou PDF) à faire signer au détenteur lors de la remise, comportant au minimum :
- L'identification de chaque clé remise (il peut y en avoir plusieurs)
- La liste des accès couverts
- Le nom du détenteur
- La date de remise
- La date de retour prévue (pour un prêt ponctuel)
- Un espace signature

---

## 4. Contraintes techniques

- Application de bureau Windows, distribuée sous forme d'un **unique exécutable (.exe)** sans installation ni dépendance à installer sur le poste (identique à la V2)
- Langage Go, framework Fyne (continuité de la V2)
- Base de données SQLite, fichier stocké sur le partage réseau de l'établissement
- Pas d'abonnement SaaS ni de dépendance à Internet pour le fonctionnement courant
- Données exportables au format CSV ou Excel
- Pas besoin d'authentification complexe (identification simple par nom à la connexion suffit)
- **Multi-postes simultanés à résoudre :** la V2 interdit explicitement l'usage simultané sous peine de corruption de données. La V3 doit traiter ce point. Deux approches à évaluer lors du développement : gestion des verrous SQLite avec retry automatique (solution légère), ou passage à un mode client/serveur léger embarqué (solution plus robuste mais plus complexe à déployer).

---

## 5. Contraintes d'usage

- Prise en main rapide pour un agent non technique
- Interface en français
- Pas de formation spécifique requise
- La saisie d'un prêt ou d'un retour ne doit pas prendre plus de 2 minutes

---

## 6. Ce qui n'est pas dans le périmètre (exclusions)

- Contrôle d'accès électronique (badges, lecteurs, portes connectées) : hors périmètre
- Gestion des clés USB ou clés informatiques : hors périmètre
- Comptabilité ou valorisation des clés : hors périmètre
- Notifications automatiques par mail ou SMS : non prioritaire (à envisager en version ultérieure)

---

## 7. Orientation technique retenue

**V3 : évolution de la V2 existante, Go + Fyne + SQLite.**

Le code de base est celui du repo `collet-david-pro/Clefs` (branche main, version v2.1.1). La V3 s'appuie sur la même stack sans la remettre en cause.

| Composant | Choix | Statut |
|---|---|---|
| Langage | Go 1.21+ | Inchangé depuis V2 |
| Interface graphique | Fyne | Inchangé depuis V2 |
| Base de données | SQLite (`clefs.db`) | Inchangé depuis V2 |
| Génération PDF | Bibliothèque existante dans le repo | Inchangé depuis V2 |
| Compilation | Pipeline GitHub Actions existant | Inchangé depuis V2 |

**Fonctionnalités déjà présentes en V2 (à conserver sans régression) :**
- Tableau de bord temps réel
- Gestion des clés avec liaison clé ↔ points d'accès
- Gestion des emprunteurs
- Emprunt multi-clés en une opération
- Génération de PDF individuel et groupé
- Rapport des emprunts en cours avec code couleur de durée
- Sauvegarde / restauration intégrée
- Outil de migration depuis la V1 Python

**Fonctionnalités nouvelles à développer en V3 :**
- Référentiel des accès enrichi : ajout du champ étage/niveau et de la catégorie sur les points d'accès existants
- Filtrage des clés et des accès par bâtiment, étage, catégorie
- Logique de prêt par besoin : saisir les accès requis → proposition automatique de la combinaison minimale de clés disponibles
- Détection et signalement des redondances d'accès chez un même détenteur
- Vues transversales : "qui a quoi", "quelle clé pour quelle porte", "clés disponibles par zone"
- Résolution du problème multi-postes simultanés (voir section 4)
- Bon de remise enrichi : liste des accès couverts en plus des clés remises

**Migration des données V2 → V3 :** aucune migration prévue. La V3 démarre avec une base vierge. Les données de la V2 ne sont pas reprises.

---

## 8. Critères d'acceptation

L'outil sera considéré opérationnel si :

1. On peut enregistrer une remise de clé(s) en moins de 2 minutes, en partant du besoin d'accès
2. L'outil propose automatiquement la combinaison de clés minimale pour couvrir un besoin d'accès donné
3. On peut savoir en moins de 30 secondes qui détient actuellement une clé donnée
4. On peut savoir en moins de 30 secondes quelles clés ouvrent une porte donnée
5. On peut filtrer les clés par bâtiment et par étage
6. Les redondances d'accès chez un même détenteur sont signalées visuellement
7. On peut éditer et imprimer un bon de remise depuis l'outil
8. L'historique complet d'une clé est consultable
9. Les prêts en retard sont visibles sans recherche manuelle
10. L'outil fonctionne depuis au moins deux postes simultanément sur le réseau local

---

*Fin du document - version 5*
