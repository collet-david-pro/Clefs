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

	// Section 1: Démarrage Rapide
	section1 := createHelpSection(
		"🚀 Démarrage Rapide",
		"Pour commencer à utiliser l'application :\n\n"+
			"1. Créez vos bâtiments (Configuration > Bâtiments)\n"+
			"2. Ajoutez des salles/points d'accès (Configuration > Salles)\n"+
			"3. Enregistrez vos clés (Configuration > Clés)\n"+
			"4. Ajoutez des emprunteurs (Configuration > Emprunteurs)\n"+
			"5. Commencez à gérer les emprunts depuis le Tableau de Bord",
	)
	accordions.Add(section1)

	// Section 2: Tableau de Bord
	section2 := createHelpSection(
		"📊 Tableau de Bord",
		"Le tableau de bord affiche toutes vos clés avec leur disponibilité en temps réel.\n\n"+
			"Colonnes du tableau :\n"+
			"  • Numéro : Identifiant de la clé\n"+
			"  • Description : Description détaillée\n"+
			"  • Disponibilité : Nombre disponible / Total utilisable\n"+
			"  • Emprunté Par : Liste des emprunteurs actuels\n"+
			"  • Actions : Boutons Emprunter/Retourner\n\n"+
			"💡 Astuce : Les clés disponibles sont en vert, les indisponibles en rouge.",
	)
	accordions.Add(section2)

	// Section 3: Gestion des Emprunts
	section3 := createHelpSection(
		"🔄 Gérer les Emprunts",
		"Créer un emprunt :\n"+
			"  1. Cliquez sur 'Nouvel Emprunt' ou 'Emprunter' sur une clé\n"+
			"  2. Sélectionnez la/les clé(s) à emprunter\n"+
			"  3. Choisissez l'emprunteur\n"+
			"  4. Confirmez l'emprunt\n\n"+
			"Retourner une clé :\n"+
			"  1. Cliquez sur 'Retourner' sur la clé concernée\n"+
			"  2. Si plusieurs emprunts, sélectionnez celui à retourner\n"+
			"  3. Confirmez le retour\n\n"+
			"💡 Astuce : Vous pouvez emprunter plusieurs clés en même temps !",
	)
	accordions.Add(section3)

	// Section 4: Gestion des Clés
	section4 := createHelpSection(
		"🔑 Gestion des Clés",
		"Accès : Configuration > Clés\n\n"+
			"Ajouter une clé :\n"+
			"  1. Cliquez sur 'Ajouter une Clé'\n"+
			"  2. Remplissez les informations :\n"+
			"     • Numéro (ex: K001)\n"+
			"     • Description\n"+
			"     • Quantité totale\n"+
			"     • Quantité en réserve (non empruntable)\n"+
			"     • Lieu de stockage\n"+
			"  3. Associez les salles accessibles avec cette clé\n"+
			"  4. Enregistrez\n\n"+
			"📐 Formule : Quantité disponible = Total - Réserve - Emprunts en cours",
	)
	accordions.Add(section4)

	// Section 5: Sauvegardes
	section5 := createHelpSection(
		"💾 Gestion des Sauvegardes",
		"Accès : Configuration > Gérer les Sauvegardes\n\n"+
			"Créer une sauvegarde :\n"+
			"  • Cliquez sur 'Créer une Nouvelle Sauvegarde'\n"+
			"  • La sauvegarde est créée instantanément\n\n"+
			"Restaurer une sauvegarde :\n"+
			"  1. Sélectionnez la sauvegarde dans la liste\n"+
			"  2. Cliquez sur 'Restaurer'\n"+
			"  3. Confirmez (une sauvegarde de sécurité est créée automatiquement)\n\n"+
			"Supprimer une sauvegarde :\n"+
			"  1. Cliquez sur 'Supprimer' à côté de la sauvegarde\n"+
			"  2. Confirmez la suppression\n\n"+
			"📁 Emplacement : Les sauvegardes sont dans le dossier 'backups/'\n"+
			"⚠️ Pensez à sauvegarder régulièrement vos données !",
	)
	accordions.Add(section5)

	// Section 6: Rapports et PDFs
	section6 := createHelpSection(
		"📄 Rapports et PDFs",
		"Emprunts en Cours :\n"+
			"  • Vue accordéon par emprunteur\n"+
			"  • Génération de reçus individuels ou groupés\n"+
			"  • Export PDF automatique dans ./documents/\n\n"+
			"Rapport des Clés Sorties :\n"+
			"  • Vue accordéon groupée par clé\n"+
			"  • Liste des emprunteurs par clé\n"+
			"  • Export PDF du rapport\n\n"+
			"Plan de Clés :\n"+
			"  • Vue hiérarchique : Bâtiments > Salles > Clés\n"+
			"  • Export PDF du plan complet\n\n"+
			"Bilan des Clés :\n"+
			"  • Vue accordéon de toutes les clés\n"+
			"  • Statut de disponibilité\n"+
			"  • Liste des emprunts actifs par clé\n\n"+
			"✅ Tous les PDFs supportent les caractères accentués !\n"+
			"📂 Tous les PDFs sont enregistrés dans ./documents/",
	)
	accordions.Add(section6)

	// Section 7: Configuration
	section7 := createHelpSection(
		"⚙️ Configuration",
		"Le menu Configuration vous permet de gérer :\n\n"+
			"🏢 Bâtiments : Créez et organisez vos bâtiments\n"+
			"🚪 Salles : Ajoutez des salles/points d'accès par bâtiment\n"+
			"🔑 Clés : Gérez votre inventaire de clés\n"+
			"👤 Emprunteurs : Enregistrez les personnes autorisées\n"+
			"💾 Sauvegardes : Gérez vos sauvegardes\n"+
			"🎭 Mode Démo : Chargez des données de test\n"+
			"🔄 Réinitialisation : Remettez à zéro la base de données",
	)
	accordions.Add(section7)

	// Section 8: Astuces
	section8 := createHelpSection(
		"💡 Astuces et Bonnes Pratiques",
		"✅ Sauvegardez régulièrement votre base de données\n"+
			"✅ Utilisez des numéros de clés cohérents (ex: K001, K002...)\n"+
			"✅ Définissez une réserve pour les clés critiques\n"+
			"✅ Associez correctement les clés aux salles\n"+
			"✅ Vérifiez les emprunts en cours régulièrement\n"+
			"✅ Générez des reçus PDF pour garder une trace\n"+
			"✅ Utilisez le mode démo pour vous familiariser\n\n"+
			"⚠️ Attention :\n"+
			"  • La réinitialisation supprime TOUTES les données\n"+
			"  • Toujours confirmer avant de supprimer\n"+
			"  • Les sauvegardes ne sont pas synchronisées avec Git",
	)
	accordions.Add(section8)

	// Section 9: Navigation
	section9 := createHelpSection(
		"🧭 Navigation Rapide",
		"Utilisez le menu de gauche pour naviguer rapidement :\n\n"+
			"📊 Tableau de Bord : Vue d'ensemble et actions rapides\n"+
			"📋 Emprunts en Cours : Gestion des emprunts actifs (accordéon par emprunteur)\n"+
			"📄 Rapport des Clés : Export et statistiques (accordéon par clé)\n"+
			"🗺️ Plan de Clés : Vue hiérarchique complète\n"+
			"⚙️ Configuration : Paramètres et gestion des données\n"+
			"À Propos : Informations sur l'application\n"+
			"📖 Mode d'Emploi : Ce guide (vous y êtes !)",
	)
	accordions.Add(section9)

	// Section 10: Support
	section10 := createHelpSection(
		"❓ Besoin d'Aide ?",
		"Si vous rencontrez un problème :\n\n"+
			"1. Consultez ce mode d'emploi\n"+
			"2. Vérifiez la page 'À Propos' pour les informations\n"+
			"3. Consultez le fichier README.md dans le dossier de l'application\n"+
			"4. Vérifiez CHANGELOG_NOUVELLES_FONCTIONNALITES.md pour les nouveautés\n\n"+
			"📚 Documentation complète disponible dans les fichiers :\n"+
			"  • README.md - Guide complet\n"+
			"  • INSTALLATION.md - Installation détaillée\n"+
			"  • QUICK_START.md - Démarrage rapide",
	)
	accordions.Add(section10)

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
