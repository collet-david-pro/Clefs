package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createHelpView crée la vue du mode d'emploi avec accordéons
func createHelpView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("📖 Mode d'Emploi", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Introduction
	intro := widget.NewLabel("Ce guide vous aidera à utiliser toutes les fonctionnalités du Gestionnaire de Clés. " +
		"Cliquez sur chaque section pour afficher les détails.")
	intro.Wrapping = fyne.TextWrapWord

	// Créer les accordéons pour chaque section
	accordions := container.NewVBox()

	// Section 1: Installation & Mise à jour
	section1 := createHelpSection(
		"📥 Installation & Mise à jour",
		"IMPORTANT : Placez toujours l'application dans un dossier dédié (ex: Documents/Clefs) car elle crée ses propres fichiers (base de données, documents, sauvegardes).\n\n"+
			"Windows :\n"+
			"  • Lancement : Double-cliquez simplement sur le fichier .exe\n"+
			"  • Mise à jour : Remplacez l'ancien .exe par le nouveau\n\n"+
			"macOS & Linux :\n"+
			"  • Installation : Ouvrez un terminal dans le dossier et lancez 'chmod +x nom_du_fichier'\n"+
			"  • Lancement : Via le terminal avec './nom_du_fichier'\n"+
			"  • Mise à jour : Remplacez le fichier et refaites le 'chmod +x'",
	)
	accordions.Add(section1)

	// Section 2: Migration depuis V1 (Python)
	section2 := createHelpSection(
		"🔄 Migration depuis V1 (Python)",
		"Si vous venez de l'ancienne version Python, vous pouvez récupérer toutes vos données :\n\n"+
			"1. Localisez votre ancien fichier 'clefs.db'\n"+
			"2. Dans cette application, allez dans 'Configuration' > 'Importer depuis V1'\n"+
			"3. Sélectionnez votre ancien fichier 'clefs.db'\n"+
			"4. Validez l'importation\n\n"+
			"⚠️ Attention : Faites cette opération au tout début, car elle fusionne les données.",
	)
	accordions.Add(section2)

	// Section 3: Démarrage Rapide
	section3 := createHelpSection(
		"🚀 Démarrage Rapide",
		"Pour configurer votre inventaire :\n\n"+
			"1. Créez vos bâtiments (Configuration > Bâtiments)\n"+
			"2. Ajoutez des salles/points d'accès (Configuration > Salles)\n"+
			"3. Enregistrez vos clés (Configuration > Clés)\n"+
			"4. Ajoutez des emprunteurs (Configuration > Emprunteurs)\n"+
			"5. Commencez à gérer les emprunts depuis le Tableau de Bord",
	)
	accordions.Add(section3)

	// Section 4: Tableau de Bord Moderne
	section4 := createHelpSection(
		"📊 Tableau de Bord",
		"Le nouveau tableau de bord vous offre une vue synthétique :\n\n"+
			"Statistiques (en haut) :\n"+
			"  • Total des clés gérées\n"+
			"  • Nombre d'emprunts actifs\n"+
			"  • Clés disponibles immédiatement\n"+
			"  • Nombre d'emprunteurs enregistrés\n\n"+
			"Tableau de gestion :\n"+
			"  • Numéro & Description : Identification de la clé\n"+
			"  • Disponibilité : Code couleur (Vert = Dispo, Rouge = Indispo)\n"+
			"  • Emprunteurs : Liste compacte des personnes ayant la clé\n"+
			"  • Actions : Boutons rapides pour Emprunter ou Retourner",
	)
	accordions.Add(section4)

	// Section 5: Gestion des Emprunts
	section5 := createHelpSection(
		"🔄 Gérer les Emprunts",
		"Créer un emprunt :\n"+
			"  1. Cliquez sur '➕ Nouvel Emprunt' (en haut) ou 'Emprunter' (dans la liste)\n"+
			"  2. Sélectionnez la/les clé(s) à emprunter\n"+
			"  3. Choisissez l'emprunteur\n"+
			"  4. Confirmez l'emprunt\n\n"+
			"Retourner une clé :\n"+
			"  1. Cliquez sur 'Retourner' sur la ligne de la clé\n"+
			"  2. Si plusieurs personnes ont cette clé, choisissez qui la rend\n"+
			"  3. Confirmez le retour\n\n"+
			"💡 Astuce : Vous pouvez sélectionner plusieurs clés d'un coup lors d'un nouvel emprunt !",
	)
	accordions.Add(section5)

	// Section 6: Gestion des Clés
	section6 := createHelpSection(
		"🔑 Gestion des Clés",
		"Accès : Configuration > Clés\n\n"+
			"Ajouter une clé :\n"+
			"  1. Cliquez sur 'Ajouter une Clé'\n"+
			"  2. Remplissez les informations :\n"+
			"     • Numéro (ex: K001)\n"+
			"     • Description\n"+
			"     • Quantité totale\n"+
			"     • Quantité en réserve (stock de sécurité non empruntable)\n"+
			"     • Lieu de stockage\n"+
			"  3. Associez les salles que cette clé ouvre\n"+
			"  4. Enregistrez\n\n"+
			"📐 Formule : Disponible = Total - Réserve - Emprunts en cours",
	)
	accordions.Add(section6)

	// Section 7: Sauvegardes
	section7 := createHelpSection(
		"💾 Gestion des Sauvegardes",
		"Accès : Configuration > Gérer les Sauvegardes\n\n"+
			"Créer une sauvegarde :\n"+
			"  • Cliquez sur 'Créer une Nouvelle Sauvegarde'\n"+
			"  • La sauvegarde est créée instantanément dans le dossier 'backups/'\n\n"+
			"Restaurer une sauvegarde :\n"+
			"  1. Sélectionnez la sauvegarde dans la liste\n"+
			"  2. Cliquez sur 'Restaurer'\n"+
			"  3. Confirmez (une sauvegarde de sécurité est créée automatiquement avant)\n\n"+
			"⚠️ Conseil : Copiez régulièrement le dossier 'backups/' sur un support externe.",
	)
	accordions.Add(section7)

	// Section 8: Rapports et PDFs
	section8 := createHelpSection(
		"📄 Rapports et PDFs",
		"Emprunts en Cours :\n"+
			"  • Vue par emprunteur\n"+
			"  • Génération de reçus de prêt (PDF)\n\n"+
			"Rapport des Clés Sorties :\n"+
			"  • Vue par clé\n"+
			"  • Liste de qui a quoi\n\n"+
			"Plan de Clés :\n"+
			"  • Vue hiérarchique : Bâtiments > Salles > Clés\n"+
			"  • Export PDF du plan complet\n\n"+
			"📂 Tous les documents sont générés automatiquement dans le dossier 'documents/'.",
	)
	accordions.Add(section8)

	// Section 9: Configuration
	section9 := createHelpSection(
		"⚙️ Configuration",
		"Le menu Configuration vous permet de gérer :\n\n"+
			"🏢 Bâtiments : Créez et organisez vos bâtiments\n"+
			"🚪 Salles : Ajoutez des salles/points d'accès par bâtiment\n"+
			"🔑 Clés : Gérez votre inventaire de clés\n"+
			"👤 Emprunteurs : Enregistrez les personnes autorisées\n"+
			"💾 Sauvegardes : Gérez vos sauvegardes\n"+
			"📥 Import V1 : Migrez vos données depuis l'ancienne version\n"+
			"🎭 Mode Démo : Chargez des données de test\n"+
			"🔄 Réinitialisation : Remettez à zéro la base de données",
	)
	accordions.Add(section9)

	// Section 10: Astuces
	section10 := createHelpSection(
		"💡 Astuces et Bonnes Pratiques",
		"✅ Sauvegardez régulièrement votre base de données\n"+
			"✅ Utilisez des numéros de clés cohérents (ex: K001, K002...)\n"+
			"✅ Définissez une réserve pour les clés critiques\n"+
			"✅ Vérifiez les emprunts en cours régulièrement\n"+
			"✅ Générez des reçus PDF pour garder une trace signée\n"+
			"✅ Utilisez le mode démo pour vous familiariser sans risque\n\n"+
			"⚠️ Attention : La réinitialisation est irréversible !",
	)
	accordions.Add(section10)

	// Section 11: Navigation
	section11 := createHelpSection(
		"🧭 Navigation Rapide",
		"Menu de gauche :\n\n"+
			"📊 Tableau de Bord : Vue d'ensemble et actions rapides\n"+
			"📋 Emprunts en Cours : Gestion des emprunts actifs\n"+
			"📄 Rapport des Clés : État des lieux des clés sorties\n"+
			"🗺️ Plan de Clés : Vue structurelle (Bâtiment/Salle)\n"+
			"⚙️ Configuration : Paramètres et données\n"+
			"À Propos : Version et crédits\n"+
			"📖 Mode d'Emploi : Ce guide",
	)
	accordions.Add(section11)

	// Section 12: Support
	section12 := createHelpSection(
		"❓ Besoin d'Aide ?",
		"En cas de problème :\n\n"+
			"1. Consultez ce mode d'emploi\n"+
			"2. Vérifiez le fichier 'infos.txt' inclus\n"+
			"3. Consultez le README.md pour les détails techniques\n"+
			"4. Vérifiez que vous avez bien les droits d'écriture dans le dossier",
	)
	accordions.Add(section12)

	// Assembler le contenu
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		intro,
		widget.NewSeparator(),
		accordions,
	)

	return container.NewVScroll(
		container.NewPadded(content),
	)
}

// createHelpSection crée une section d'aide avec accordéon
func createHelpSection(title string, content string) *widget.Accordion {
	label := widget.NewLabel(content)
	label.Wrapping = fyne.TextWrapWord

	item := widget.NewAccordionItem(title, label)
	accordion := widget.NewAccordion(item)

	return accordion
}
