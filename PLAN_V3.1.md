# Plan V3.1 — Simplifications et améliorations UX

> Date : 2026-05-12  
> Base : V3.0.0 fonctionnelle  
> Mis à jour : 2026-05-12 (ajout section 5 — affichage listes portes/clés)

---

## 1. Retirer l'identification au démarrage

### Problème
Le dialogue de login au démarrage (`login.go`, `showLoginDialog`) est inutile dans le contexte réel : c'est toujours un administrateur qui ouvre l'application.

### Solution
- Supprimer `login.go` entièrement
- Dans `app.go`, `Run()` démarre directement sur le dashboard sans passer par `showLoginDialog`
- Supprimer `currentUser` de la struct `App` (plus utilisé en surface — le champ `created_by` des prêts sera rempli avec `"Administrateur"` fixe)
- Supprimer l'affichage `userLabel` dans le menu
- Mettre à jour `loan_wizard.go` : le champ `created_by` est fixé à `"Administrateur"` sans saisie

**Fichiers touchés :** `gui/app.go`, `gui/login.go` (supprimé), `gui/loan_wizard.go`

---

## 2. Simplification du menu — suppression des doublons

### Problème actuel
Le menu contient des redondances et une organisation peu claire :

| Bouton | Doublon / problème |
|--------|-------------------|
| "Nouvel Emprunt" (bouton principal) | Aussi accessible depuis le tableau de bord |
| "Rapport des Clés" | Doublon fonctionnel avec "Emprunts en Cours" |
| Card "Vues rapides" avec 5 sous-boutons | Trop chargé, certains se chevauchent avec d'autres vues |
| Card "Configuration" avec 1 seul bouton | Inutile d'encadrer un seul bouton dans une card |
| Card "Aide" | Inutile d'encadrer dans une card |
| "🚪 Quitter" avec emoji porte | Même emoji que "🚪 Clé → Porte" — confusion |

### Solution — menu réorganisé en 3 sections claires

```
MENU
────────────────────
[ Tableau de Bord ]      ← toujours en premier, HighImportance

── Prêts ─────────────
[ Nouvel Emprunt ]
[ Emprunts en Cours ]
[ Historique ]

── Référentiels ───────
[ Clés ]
[ Détenteurs ]
[ Accès ]
[ Bâtiments ]

── Consultation ───────
[ Qui a quoi ? ]
[ Par bâtiment ]
[ Redondances ]
[ Plan de Clés ]

── Application ────────
[ Configuration ]
[ Aide ]
[ Quitter ]
────────────────────
```

**Suppressions :**
- "Rapport des Clés" → fusionné dans "Emprunts en Cours" (un onglet ou un bouton dans la vue)
- Card "Vues rapides" → remplacée par 3 boutons directs dans la section "Consultation"
- "Clé → Porte" et "Clés disponibles" → accessibles depuis le tableau de bord et la vue Clés, pas besoin d'entrées de menu dédiées
- Card "Configuration" → bouton direct
- Card "Aide" → boutons directs

**Fichiers touchés :** `gui/app.go`

---

## 3. Correction des émojis

### Problèmes identifiés

| Emplacement | Émoji actuel | Problème | Correction |
|-------------|-------------|----------|------------|
| Menu "Quitter" | 🚪 | Même que "Clé → Porte" | ✕ ou X ou rien |
| Menu "Clé → Porte" | 🚪 | Ambigu | Supprimé du menu principal |
| Bouton "Rapport des Clés" | 📄 | Identique à d'autres PDF | Supprimé |
| Dashboard "Nouvel Emprunt" | ➕ | OK | Conservé |
| Tableau de bord alertes | 🔴 / ⚠️ | OK | Conservés |
| Config "Réinitialiser" | 🗑️ | OK | Conservé |
| Menu "Historique" | 🕐 | Peu lisible sur Windows | Remplacé par rien (texte seul) |
| Menu "Plan de Clés" | 🗺️ | Rendu parfois cassé Windows | Remplacé par texte seul |
| Menu "Accès" | pas d'émoji | Incohérent | Ajouter icône cohérente |

### Règle générale appliquée
- Garder les émojis **uniquement** pour les actions critiques et les alertes
- Sections de menu : **texte seul**, sans émoji
- Boutons d'action principaux : un seul émoji simple et universel

**Fichiers touchés :** `gui/app.go`, `gui/dashboard_modern.go`, `gui/config.go`

---

## 4. Simplification du prêt — vue unifiée en une étape

### Problème actuel
Le wizard en 3 étapes (popup modale → étape 1 → étape 2 → étape 3) est trop complexe :
- L'utilisateur doit naviguer 3 fois avec des boutons Suivant/Retour
- La logique de calcul automatique des clés est cachée à l'étape 3
- Deux systèmes parallèles coexistent : `loan_wizard.go` ET `loan_dialogs.go` (`showNewLoanDialog`)

### Solution — vue de prêt unifiée en une seule page

**Supprimer :** `loan_wizard.go` entièrement, et `showNewLoanDialog` / `showNewLoanDialogWithKey` dans `loan_dialogs.go`

**Créer :** une vue de prêt complète dans la zone de contenu principale (pas une popup), avec tout visible d'un coup :

```
┌─────────────────────────────────────────────────────┐
│  NOUVEAU PRÊT                                        │
│                                                      │
│  Détenteur :  [ Select ou recherche            ▼ ]  │
│               [ + Nouveau détenteur ]               │
│                                                      │
│  ─── Filtrer les accès ─────────────────────────    │
│  Bâtiment : [ Tous ▼ ]   Étage : [ Tous ▼ ]        │
│                                                      │
│  Cocher les portes auxquelles il faut avoir accès : │
│  ☐  Salle B12          [Bât. B — R+1]              │
│  ☐  Portail cour nord  [Extérieur — RDC]            │
│  ☑  Bureau direction   [Bât. A — R+2]              │
│  ...                                                 │
│                                                      │
│  ─── Trousseau calculé automatiquement ─────────    │
│  → K003 — Clé direction (couvre 3 accès)            │
│  → K007 — Passe bâtiment B (couvre 1 accès)         │
│  [ Modifier le trousseau manuellement ]              │
│                                                      │
│  Date de retour prévue : [            ] (optionnel) │
│  Type : ( ) Ponctuel   ( ) Permanent                │
│                                                      │
│        [ Annuler ]   [ Valider + Imprimer bon ]     │
└─────────────────────────────────────────────────────┘
```

**Comportement :**
- Le trousseau se recalcule automatiquement à chaque changement de case à cocher (sans bouton "Suivant")
- Si aucun détenteur sélectionné, le trousseau reste vide
- Le bouton "Valider" est grisé tant qu'aucun détenteur et aucune porte ne sont sélectionnés
- "Modifier le trousseau manuellement" : expand/collapse une section avec des checkboxes sur toutes les clés disponibles
- Le bon PDF est généré en arrière-plan après validation

**Fichiers touchés :**
- Supprimer : `gui/loan_wizard.go`
- Modifier : `gui/loan_dialogs.go` → ne garde que `showReturnDialog` et `showReturnWithConditionDialog`
- Créer : `gui/new_loan_view.go` — vue complète dans la zone de contenu (pas popup)
- Modifier : `gui/app.go` → `showNewLoan()` appelle `createNewLoanView(a)` dans `setContent`

---

## 5. Refonte visuelle des listes de portes et de clés

### Problème actuel

Partout dans l'application, les listes de portes (accès) et de clés souffrent des mêmes problèmes visuels :

| Vue | Problème |
|-----|---------|
| `accesses.go` — liste des accès | Tout sur une ligne `label` unique, texte tronqué, émojis parasites (`🚪 🏢 📶 🏷 🔑`), pas de hiérarchie visuelle |
| `keys.go` — liste des clés | Accordéon lourd, texte long accumulé en `"📝 Description: … \| 📦 Quantité: … \| 🏢 Salles: …"`, surcharge |
| `new_loan_view.go` — liste des portes à cocher | `container.NewVBox` de checkboxes brutes sans séparation visuelle, label sur une ligne pouvant dépasser la largeur |
| `new_loan_view.go` — trousseau calculé | Texte brut `"→ K003 — Clé direction"`, pas de distinction visuelle entre clés suggérées et manuelles |
| `views_transversal.go` — "Clés disponibles" | Grille de 5 colonnes avec `widget.NewLabel` — pas lisible sur petits écrans, alignement cassé |
| `borrowers.go` — liste des détenteurs | Info sur 2 lignes max mais statut/email/tel fusionnés sur une seule ligne sans séparation |

### Règles visuelles à appliquer partout

**Listes de portes (accès) :**
```
┌──────────────────────────────────────────────┐
│ Salle B12                          [✏️] [🗑️] │
│ Bâtiment B · R+1 · Salle de classe            │
│ 2 clé(s) associée(s)                          │
└──────────────────────────────────────────────┘
```
- Nom en gras sur la première ligne
- Bâtiment · étage · catégorie sur la deuxième ligne (grisé, petite taille)
- Nombre de clés sur la troisième ligne
- Boutons d'action alignés à droite
- Séparateur entre chaque entrée
- **Pas d'émojis dans les libellés** sauf pour les boutons

**Listes de clés :**
```
┌──────────────────────────────────────────────┐
│ K003 — Clé direction               [✏️] [🗑️] │
│ Trousseau · Rangement : Accueil               │
│ Stock : 3  |  Réserve : 1  |  Dispo : 1      │
│ Accès : Salle B12, Bureau direction           │
│                      [ Emprunter ] [ Retour ] │
└──────────────────────────────────────────────┘
```
- Plus d'accordéon : chaque clé est une card plate avec toutes les infos visibles
- Indicateur de stock coloré : vert si dispo > 1, orange si dispo = 1, rouge si 0
- Accès listés sur une ligne, tronqués si trop longs

**Checkboxes de portes dans le prêt :**
```
  ☐  Salle B12
     Bât. B · R+1
  ☐  Portail cour nord
     Extérieur · RDC
  ☑  Bureau direction          ← coché = fond légèrement coloré
     Bât. A · R+2
```
- Nom de la porte en gras
- Bâtiment et étage sur une sous-ligne indentée, en gris
- Si coché : fond légèrement mis en valeur (card colorée ou bordure)
- Largeur fixe, pas de troncature

**Trousseau calculé dans la vue prêt :**
```
┌─ Trousseau calculé (2 clés) ───────────────┐
│  K003 — Clé direction           suggérée ✓  │
│         couvre : Bureau direction, Salle B12 │
│                                              │
│  K007 — Passe bâtiment A        suggérée ✓  │
│         couvre : Portail cour nord           │
│                                              │
│  [ + Ajouter une clé manuellement ]          │
└─────────────────────────────────────────────┘
```
- Card distincte visuellement (fond légèrement différent)
- Pour chaque clé : numéro + description sur la ligne principale
- Accès couverts listés en dessous, indentés
- Badge "suggérée" ou "ajoutée manuellement"
- Bouton pour ajouter une clé non suggérée

### Composant réutilisable à créer

Fichier `gui/widgets.go` — fonctions helper pour éviter la duplication :

```go
// accessCard retourne une card visuelle pour un accès
func accessCard(r db.Room, bName string, keyCount int, onEdit, onDelete func()) fyne.CanvasObject

// keyCard retourne une card visuelle pour une clé
func keyCard(k db.KeyWithAvailability, onLoan, onReturn func()) fyne.CanvasObject

// accessCheckRow retourne une ligne de checkbox pour la sélection d'accès dans le prêt
func accessCheckRow(r db.Room, bName string, checked bool, onChange func(bool)) fyne.CanvasObject

// loanKeyCard retourne la représentation d'une clé dans le trousseau calculé
func loanKeyCard(k db.Key, coveredAccesses []string, suggested bool, onRemove func()) fyne.CanvasObject
```

**Fichiers touchés :**
- Créer : `gui/widgets.go`
- Modifier : `gui/accesses.go` → utiliser `accessCard`
- Modifier : `gui/keys.go` → supprimer accordéons, utiliser `keyCard`
- Modifier : `gui/new_loan_view.go` → utiliser `accessCheckRow` + `loanKeyCard`
- Modifier : `gui/views_transversal.go` → `createAvailableKeysView` utiliser `keyCard` simplifié

---

## 6. Résumé des fichiers à modifier / créer / supprimer

| Action | Fichier | Motif |
|--------|---------|-------|
| Supprimer | `gui/login.go` | Plus de sélection de compte |
| Supprimer | `gui/loan_wizard.go` | Remplacé par vue unifiée |
| Créer | `gui/new_loan_view.go` | Vue de prêt unifiée, une seule page |
| Créer | `gui/widgets.go` | Composants visuels réutilisables (cards, checkrows) |
| Modifier | `gui/app.go` | Retrait login, menu simplifié, showNewLoan |
| Modifier | `gui/loan_dialogs.go` | Retirer showNewLoanDialog et showNewLoanDialogWithKey |
| Modifier | `gui/dashboard_modern.go` | Correction émojis, bouton Nouvel Emprunt → showNewLoan |
| Modifier | `gui/config.go` | Correction émojis mineurs |
| Modifier | `gui/accesses.go` | Visuel cards sans émojis parasites |
| Modifier | `gui/keys.go` | Remplacer accordéons par cards plates |
| Modifier | `gui/views_transversal.go` | Améliorer `createAvailableKeysView` |

---

## 7. Ordre d'exécution

1. Retirer le login (`login.go`, `app.go`)
2. Simplifier le menu + corriger les émojis (`app.go`, `dashboard_modern.go`)
3. Créer `gui/widgets.go` (composants visuels réutilisables)
4. Refondre `gui/accesses.go` et `gui/keys.go` avec les nouveaux composants
5. Créer `gui/new_loan_view.go` et câbler dans `app.go`
6. Supprimer `gui/loan_wizard.go`, nettoyer `gui/loan_dialogs.go`
7. Améliorer `gui/views_transversal.go`
8. Compilation + vérification
