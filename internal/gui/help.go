package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createHelpView construit la vue "Mode d.emploi" : une liste d.accordéons,
// un par thème (installation, navigation, prêt, sauvegardes...).
func createHelpView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Mode d'emploi", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel("Gestionnaire de Clés — Collège Victor Hugo, Chauny (02300).\nCliquez sur une section pour l'ouvrir.")
	intro.Wrapping = fyne.TextWrapWord

	sections := container.NewVBox()

	sections.Add(helpSection("Installation",
		"Placez l'application dans un dossier dédié (ex: C:\\Clefs\\ ou Documents/Clefs/).\n"+
			"Elle crée automatiquement ses sous-dossiers au premier lancement :\n\n"+
			"  MonDossierClefs/\n"+
			"  ├── clefs-windows-amd64.exe\n"+
			"  ├── clefs.db          ← base de données (ne pas supprimer)\n"+
			"  ├── backups/          ← sauvegardes\n"+
			"  └── documents/        ← PDF et CSV générés\n\n"+
			"Windows :\n"+
			"  • Double-cliquez sur le .exe pour lancer\n"+
			"  • Si Windows Defender bloque : 'Informations complémentaires' > 'Exécuter quand même'\n"+
			"  • Mise à jour : remplacez simplement l'ancien .exe par le nouveau\n\n"+
			"macOS (développement) :\n"+
			"  • Terminal : chmod +x clefs-macos-arm64 puis ./clefs-macos-arm64",
	))

	sections.Add(helpSection("Navigation — menu latéral",
		"Le menu à gauche est divisé en 4 sections.\n\n"+
			"Prêts :\n"+
			"  • Tableau de bord    — vue synthétique et actions rapides\n"+
			"  • Nouvel emprunt     — créer un prêt (voir section dédiée)\n"+
			"  • Emprunts en cours  — tous les prêts actifs, retour de clé\n"+
			"  • Historique         — prêts terminés, filtres, export CSV\n\n"+
			"Référentiels :\n"+
			"  • Clés        — inventaire, ajout/modification, bilan PDF, export CSV\n"+
			"  • Détenteurs  — personnes autorisées à emprunter\n"+
			"  • Accès       — portes et zones d'accès par bâtiment\n"+
			"  • Bâtiments   — structures physiques\n\n"+
			"Consultation :\n"+
			"  • Qui a quoi ?    — détenteur → ses clés et accès couverts\n"+
			"  • Par bâtiment    — bâtiment → clés associées\n"+
			"  • Redondances     — détenteurs avec des accès en doublon\n"+
			"  • Plan de clés    — vue hiérarchique bâtiment/salle/clé, export PDF\n\n"+
			"Application :\n"+
			"  • Configuration  — sauvegardes, import V2, réinitialisation\n"+
			"  • Aide           — ce guide\n"+
			"  • Quitter",
	))

	sections.Add(helpSection("Tableau de bord",
		"Vue d'ensemble au démarrage.\n\n"+
			"Statistiques (en haut) :\n"+
			"  • Total clés, emprunts actifs, clés disponibles, nombre de détenteurs\n\n"+
			"Alertes automatiques (si applicable) :\n"+
			"  • Prêts en retard (date de retour prévue dépassée)\n"+
			"  • Détenteurs avec accès redondants\n\n"+
			"Tableau des clés :\n"+
			"  • Disponibilité colorée : vert = dispo, rouge = indisponible\n"+
			"  • Bouton 'Emprunter' → ouvre directement la vue Nouvel emprunt\n"+
			"  • Bouton 'Retourner' → dialogue de retour rapide",
	))

	sections.Add(helpSection("Nouvel emprunt",
		"Accès : Prêts > Nouvel emprunt, ou bouton 'Nouvel emprunt' du tableau de bord.\n\n"+
			"La vue est divisée en deux panneaux côte à côte.\n\n"+
			"Panneau gauche — Sélection :\n"+
			"  1. Choisissez le détenteur dans la liste déroulante\n"+
			"     Bouton '+ Nouveau détenteur' pour en créer un à la volée\n"+
			"  2. Filtrez par bâtiment si la liste est longue\n"+
			"  3. Cochez les portes/zones d'accès à couvrir\n\n"+
			"Panneau droit — Trousseau calculé :\n"+
			"  • L'application propose automatiquement le jeu de clés minimal\n"+
			"    couvrant toutes les portes cochées\n"+
			"  • Bouton 'Retirer' sur une clé : la retire et l'exclut définitivement\n"+
			"    des suggestions pour cet emprunt (l'algo recalcule sans elle)\n"+
			"  • 'Ajouter une clé manuellement' : ajoutez une clé spécifique\n"+
			"    indépendamment des portes cochées — elle sera toujours conservée\n"+
			"  • Si on re-coche/décoche des portes, les exclusions sont remises à zéro\n\n"+
			"Options (bas du panneau droit) :\n"+
			"  • Retour prévu : date au format JJ/MM/AAAA (optionnel)\n"+
			"  • Type de prêt : ponctuel ou permanent\n\n"+
			"Validation :\n"+
			"  • Le bouton 'Valider et imprimer le bon' s'active quand un détenteur\n"+
			"    et au moins une clé sont présents dans le trousseau\n"+
			"  • Les prêts sont enregistrés et un bon de remise PDF est généré\n"+
			"    automatiquement dans documents/\n\n"+
			"Alertes affichées en temps réel :\n"+
			"  • Accès non couverts : aucune clé disponible pour cette porte\n"+
			"  • Redondances avec les clés déjà détenues par ce détenteur",
	))

	sections.Add(helpSection("Retour de clé",
		"Depuis le tableau de bord :\n"+
			"  • Colonne Actions > bouton 'Retourner' sur la ligne de la clé\n\n"+
			"Depuis Emprunts en cours :\n"+
			"  • Bouton 'Retour' sur la ligne de l'emprunt\n\n"+
			"Procédure :\n"+
			"  1. Si plusieurs personnes ont la même clé, sélectionnez qui la rend\n"+
			"  2. Confirmez le retour\n"+
			"  3. L'emprunt est archivé dans l'historique",
	))

	sections.Add(helpSection("Gestion des clés",
		"Accès : Référentiels > Clés\n\n"+
			"Ajouter une clé :\n"+
			"  1. Bouton 'Ajouter une Clé'\n"+
			"  2. Renseignez : numéro, description, quantité totale,\n"+
			"     quantité en réserve, emplacement de stockage\n"+
			"  3. Cochez les portes que cette clé ouvre (liste scrollable)\n"+
			"  4. Enregistrez\n\n"+
			"Formule disponibilité :\n"+
			"  Disponible = Total − Réserve − Emprunts en cours\n\n"+
			"Ajustement rapide du stock :\n"+
			"  • Dans la liste, chaque clé propose un curseur + champ pour modifier\n"+
			"    directement le stock total, sans ouvrir le formulaire complet\n"+
			"  • Le bouton 'Enregistrer' s'active dès que la valeur change\n\n"+
			"Vue compacte / détaillée :\n"+
			"  • Bouton 'Vue compacte' en haut : bascule vers une liste où chaque\n"+
			"    clé tient sur une ligne (numéro + stock dispo/total)\n"+
			"  • Cliquez une ligne pour la déplier et retrouver le détail complet\n"+
			"    (description, disponibilité, ajustement, actions) ; plusieurs\n"+
			"    lignes peuvent être ouvertes à la fois\n"+
			"  • Bouton 'Vue détaillée' pour revenir aux cards complètes\n\n"+
			"Exports disponibles :\n"+
			"  • Bilan des clés (PDF) — état du stock avec emprunts actifs\n"+
			"  • Export CSV — tableur complet compatible Excel",
	))

	sections.Add(helpSection("Gestion des accès",
		"Accès : Référentiels > Accès\n\n"+
			"Filtres (bâtiment / étage / catégorie) en haut de la vue.\n\n"+
			"Vue compacte / détaillée :\n"+
			"  • Même principe que pour les clés : le bouton 'Vue compacte'\n"+
			"    affiche un accès par ligne (nom · bâtiment — nombre de clés)\n"+
			"  • Cliquez une ligne pour la déplier et voir le détail (méta-données,\n"+
			"    clés associées, actions Modifier/Supprimer)\n"+
			"  • 'Vue détaillée' rétablit les cards complètes",
	))

	sections.Add(helpSection("Référentiels — ordre de configuration",
		"Pour un premier paramétrage, respectez cet ordre :\n\n"+
			"  1. Bâtiments (Référentiels > Bâtiments)\n"+
			"  2. Accès / Portes (Référentiels > Accès)\n"+
			"     Chaque accès appartient à un bâtiment\n"+
			"     Champs : nom, étage, catégorie, notes\n"+
			"  3. Clés (Référentiels > Clés)\n"+
			"     Associez à chaque clé les accès qu'elle ouvre\n"+
			"  4. Détenteurs (Référentiels > Détenteurs)\n"+
			"     Champs : nom, statut, email, téléphone",
	))

	sections.Add(helpSection("Sauvegardes",
		"Accès : Application > Configuration > Gérer les sauvegardes\n\n"+
			"Créer une sauvegarde :\n"+
			"  • Cliquez 'Créer une sauvegarde'\n"+
			"  • Fichier .db généré instantanément dans backups/\n"+
			"  • La sauvegarde est atomique (sûre même si l'app est utilisée)\n\n"+
			"Restaurer :\n"+
			"  1. Sélectionnez la sauvegarde dans la liste\n"+
			"  2. Cliquez 'Restaurer'\n"+
			"  3. Une sauvegarde de sécurité est créée automatiquement avant\n\n"+
			"Conseil : copiez régulièrement le dossier backups/ sur un support externe\n"+
			"ou un partage réseau.",
	))

	sections.Add(helpSection("Migration depuis V2",
		"Si vous disposez d'une base de données de la version Go précédente (V2) :\n\n"+
			"  1. Application > Configuration > Importer depuis V2\n"+
			"  2. Sélectionnez l'ancien fichier clefs.db\n"+
			"  3. Une sauvegarde de la base actuelle est créée avant l'import\n"+
			"  4. Les nouveaux champs V3 (étage, catégorie, statut détenteur)\n"+
			"     sont laissés vides et peuvent être complétés manuellement\n\n"+
			"Faites cette opération au démarrage, sur une base vide.",
	))

	sections.Add(helpSection("Utilisation en réseau",
		"Placez le dossier de l'application sur un partage réseau (SMB/NFS).\n\n"+
			"Plusieurs postes peuvent ouvrir l'application simultanément :\n"+
			"  • SQLite en mode WAL gère les accès concurrents automatiquement\n"+
			"  • Une seule écriture à la fois — les autres attendent jusqu'à 5 secondes\n"+
			"  • La lecture est toujours possible même pendant une écriture\n\n"+
			"Conseil : évitez les connexions réseau très lentes (Wi-Fi instable)\n"+
			"qui peuvent provoquer des timeouts sur les écritures.",
	))

	sections.Add(helpSection("Historique et exports CSV",
		"Accès : Prêts > Historique\n\n"+
			"  • Affiche tous les prêts terminés\n"+
			"  • Filtres : détenteur, clé, statut (en cours / retourné / en retard)\n"+
			"  • Export CSV : fichier dans documents/, encodage UTF-8 BOM,\n"+
			"    séparateur ';' — s'ouvre directement dans Excel sans conversion\n\n"+
			"Autres exports CSV disponibles :\n"+
			"  • Inventaire des clés (depuis Référentiels > Clés)\n"+
			"  • Liste des détenteurs (depuis Référentiels > Détenteurs)\n\n"+
			"Import en masse des détenteurs :\n"+
			"  1. Référentiels > Détenteurs > bouton 'Modèle CSV' :\n"+
			"     télécharge un fichier vide dans documents/ avec les bonnes\n"+
			"     colonnes (Nom, Email, Statut, Téléphone) et une ligne d'exemple\n"+
			"  2. Remplissez ce fichier (le nom est obligatoire ; statut parmi\n"+
			"     permanent, contractuel, intervenant, entreprise — défaut permanent)\n"+
			"  3. Bouton 'Importer CSV', choisissez votre fichier\n"+
			"  4. Un résumé indique le nombre de détenteurs créés et détaille\n"+
			"     les lignes invalides (une ligne fautive ne bloque pas les autres)",
	))

	sections.Add(helpSection("Bon de remise PDF",
		"Généré automatiquement à chaque validation d'emprunt.\n\n"+
			"Contenu du bon :\n"+
			"  • Nom du détenteur\n"+
			"  • Liste des clés remises\n"+
			"  • Accès couverts par le trousseau\n"+
			"  • Date de remise\n"+
			"  • Date de retour prévue (si renseignée)\n"+
			"  • Type de prêt (ponctuel / permanent)\n"+
			"  • Zone de signature\n\n"+
			"Fichier enregistré dans documents/ avec horodatage et nom du détenteur.\n\n"+
			"Plan de clés (Consultation > Plan de clés) :\n"+
			"  • Vue hiérarchique bâtiment > salle > clés\n"+
			"  • Export PDF du plan complet",
	))

	sections.Add(helpSection("Redondances",
		"Accès : Consultation > Redondances\n\n"+
			"Détecte les détenteurs qui ont plusieurs clés ouvrant les mêmes portes.\n\n"+
			"Exemple : clé A ouvre salles 1+2+3, clé B ouvre salles 2+3.\n"+
			"Si ce détenteur a déjà la clé A, la clé B est redondante.\n\n"+
			"Utilité : optimiser la distribution, réduire les risques de perte.\n\n"+
			"La vue Nouvel emprunt affiche aussi une alerte de redondance\n"+
			"en temps réel pendant la composition du trousseau.",
	))

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		intro,
		widget.NewSeparator(),
		sections,
	)

	return container.NewVScroll(container.NewPadded(content))
}

// helpSection fabrique un accordéon repliable (titre + corps de texte).
func helpSection(title string, body string) *widget.Accordion {
	label := widget.NewLabel(body)
	label.Wrapping = fyne.TextWrapWord
	return widget.NewAccordion(widget.NewAccordionItem(title, label))
}
