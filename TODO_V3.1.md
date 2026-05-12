# TODO V3.1

## 1. Retirer l'identification au démarrage
- [ ] Supprimer `internal/gui/login.go`
- [ ] `app.go` — `Run()` : démarrer directement sur le dashboard, sans `showLoginDialog`
- [ ] `app.go` — supprimer `currentUser` de la struct `App` et `userLabel` du menu
- [ ] `loan_wizard.go` — sera supprimé entièrement (voir tâche 5), donc rien à faire ici

## 2. Simplifier le menu
- [ ] `app.go` — réécrire `createMenu()` avec 4 sections : Prêts / Référentiels / Consultation / Application
- [ ] Retirer "Rapport des Clés" du menu (accessible depuis "Emprunts en Cours")
- [ ] Retirer la Card "Vues rapides" — remplacer par 3 boutons directs : "Qui a quoi ?", "Par bâtiment", "Redondances"
- [ ] Retirer "Clé → Porte" et "Clés disponibles" du menu
- [ ] Retirer les Cards autour de Configuration et Aide

## 3. Corriger les émojis
- [ ] `app.go` — "Quitter" : retirer l'émoji 🚪 (conflit avec "Clé → Porte")
- [ ] `app.go` — "Historique" : retirer l'émoji 🕐 (rendu cassé Windows)
- [ ] `app.go` — "Plan de Clés" : retirer l'émoji 🗺️ (rendu cassé Windows)
- [ ] `app.go` — sections de menu : texte seul, sans émoji dans les titres de section
- [ ] `dashboard_modern.go` — vérifier et aligner les émojis des stats/alertes

## 4. Vue de prêt unifiée (remplace le wizard 3 étapes)
- [ ] Créer `internal/gui/new_loan_view.go` avec `createNewLoanView(a *App)`
  - [ ] Select détenteur + bouton "Nouveau détenteur"
  - [ ] Filtres bâtiment / étage sur les accès
  - [ ] Liste des accès avec cases à cocher (scrollable)
  - [ ] Section "Trousseau calculé" — se met à jour à chaque coche (appel `business.SuggestKeys`)
  - [ ] Section "Modifier manuellement" (expand/collapse)
  - [ ] Champ date de retour prévue + type ponctuel/permanent
  - [ ] Bouton "Valider + Imprimer le bon" (grisé si incomplet)
- [ ] `app.go` — ajouter `showNewLoan()` → `setContent(createNewLoanView(a), "newLoan")`
- [ ] `app.go` — ajouter `"newLoan"` dans `refreshCurrentView()`
- [ ] `app.go` — tous les boutons "Nouvel Emprunt" pointent vers `showNewLoan()`
- [ ] `dashboard_modern.go` — bouton "Nouvel Emprunt" → `a.showNewLoan()`

## 5. Nettoyage
- [ ] Supprimer `internal/gui/loan_wizard.go`
- [ ] `loan_dialogs.go` — supprimer `showNewLoanDialog` et `showNewLoanDialogWithKey`
- [ ] `loan_dialogs.go` — garder uniquement `showReturnDialog` et `showReturnWithConditionDialog`

## 6. Refonte visuelle — composants réutilisables

### Créer `gui/widgets.go`
- [ ] `accessCard(r, bName, keyCount, onEdit, onDelete)` — card plate : nom en gras / bâtiment·étage·catégorie en sous-ligne / nb clés / boutons droite
- [ ] `keyCard(k, onLoan, onReturn)` — card plate : numéro+description / catégorie·emplacement / indicateur stock coloré (vert/orange/rouge) / accès listés / boutons action
- [ ] `accessCheckRow(r, bName, checked, onChange)` — checkbox avec nom en gras + bâtiment·étage en sous-ligne indentée ; fond coloré si coché
- [ ] `loanKeyCard(k, coveredAccesses, suggested, onRemove)` — card trousseau : numéro+description / accès couverts indentés / badge suggérée ou manuelle / bouton retirer

### Refondre `gui/accesses.go`
- [ ] Remplacer la ligne label unique par `accessCard` pour chaque accès
- [ ] Supprimer tous les émojis parasites des libellés (`🚪 🏢 📶 🏷`)
- [ ] Garder les boutons ✏️ et 🗑️ uniquement sur les boutons d'action

### Refondre `gui/keys.go`
- [ ] Remplacer les accordéons par `keyCard` pour chaque clé
- [ ] Supprimer les émojis dans les libellés internes (`📝 📦 📍 🏢 ✅ ⚠️`)
- [ ] Indicateur de stock : widget coloré (pas un simple label)
- [ ] Accès associés : liste tronquée à 3, "+ N autres" si plus

### Améliorer `gui/views_transversal.go`
- [ ] `createAvailableKeysView` : remplacer la grille 5 colonnes par des cards
- [ ] `createWhoHasWhatView` : nom détenteur en gras, clés en liste indentée, badge rouge si redondance

## 7. Vérification finale
- [ ] `go build ./...` sans erreur
- [ ] `go test ./...` — 15/15 tests passent
- [ ] Tester le prêt de bout en bout : sélection accès → trousseau calculé → validation → bon PDF
- [ ] Vérifier visuellement : liste accès, liste clés, vue prêt, trousseau
