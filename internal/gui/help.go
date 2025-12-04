package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createHelpView crée la vue du mode d'emploi
func createHelpView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Mode d'Emploi", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Introduction
	introTitle := widget.NewLabelWithStyle("Bienvenue", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	intro := widget.NewLabel("Ce guide vous aidera à utiliser toutes les fonctionnalités du Gestionnaire de Clés. " +
		"Suivez les instructions pas à pas pour une prise en main rapide.")
	intro.Wrapping = fyne.TextWrapWord

	// Section 1: Démarrage
	section1Title := widget.NewLabelWithStyle("1. Demarrage Rapide", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section1 := container.NewVBox(
		widget.NewLabel("Pour commencer a utiliser l'application :"),
		widget.NewLabel(""),
		widget.NewLabel("1. Creez vos batiments (Configuration > Batiments)"),
		widget.NewLabel("2. Ajoutez des salles/points d'acces (Configuration > Salles)"),
		widget.NewLabel("3. Enregistrez vos cles (Configuration > Cles)"),
		widget.NewLabel("4. Ajoutez des emprunteurs (Configuration > Emprunteurs)"),
		widget.NewLabel("5. Commencez a gerer les emprunts depuis le Tableau de Bord"),
	)

	// Section 2: Tableau de Bord
	section2Title := widget.NewLabelWithStyle("2. Tableau de Bord", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section2 := container.NewVBox(
		widget.NewLabel("Le tableau de bord affiche toutes vos cles avec leur disponibilite en temps reel."),
		widget.NewLabel(""),
		widget.NewLabel("Colonnes du tableau :"),
		widget.NewLabel("  - Numero : Identifiant de la cle"),
		widget.NewLabel("  - Description : Description detaillee"),
		widget.NewLabel("  - Disponibilite : Nombre disponible / Total utilisable"),
		widget.NewLabel("  - Emprunte Par : Liste des emprunteurs actuels"),
		widget.NewLabel("  - Actions : Boutons Emprunter/Retourner"),
		widget.NewLabel(""),
		widget.NewLabel("Astuce : Les cles disponibles sont en vert, les indisponibles en rouge."),
	)

	// Section 3: Emprunts
	section3Title := widget.NewLabelWithStyle("3. Gerer les Emprunts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section3 := container.NewVBox(
		widget.NewLabel("Creer un emprunt :"),
		widget.NewLabel("  1. Cliquez sur 'Nouvel Emprunt' ou 'Emprunter' sur une cle"),
		widget.NewLabel("  2. Selectionnez la/les cle(s) a emprunter"),
		widget.NewLabel("  3. Choisissez l'emprunteur"),
		widget.NewLabel("  4. Confirmez l'emprunt"),
		widget.NewLabel(""),
		widget.NewLabel("Retourner une cle :"),
		widget.NewLabel("  1. Cliquez sur 'Retourner' sur la cle concernee"),
		widget.NewLabel("  2. Si plusieurs emprunts, selectionnez celui a retourner"),
		widget.NewLabel("  3. Confirmez le retour"),
		widget.NewLabel(""),
		widget.NewLabel("Astuce : Vous pouvez emprunter plusieurs cles en meme temps !"),
	)

	// Section 4: Gestion des Clés
	section4Title := widget.NewLabelWithStyle("4. Gestion des Cles", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section4 := container.NewVBox(
		widget.NewLabel("Acces : Configuration > Cles"),
		widget.NewLabel(""),
		widget.NewLabel("Ajouter une cle :"),
		widget.NewLabel("  1. Cliquez sur 'Ajouter une Cle'"),
		widget.NewLabel("  2. Remplissez les informations :"),
		widget.NewLabel("     - Numero (ex: K001)"),
		widget.NewLabel("     - Description"),
		widget.NewLabel("     - Quantite totale"),
		widget.NewLabel("     - Quantite en reserve (non empruntable)"),
		widget.NewLabel("     - Lieu de stockage"),
		widget.NewLabel("  3. Associez les salles accessibles avec cette cle"),
		widget.NewLabel("  4. Enregistrez"),
		widget.NewLabel(""),
		widget.NewLabel("Quantite disponible = Total - Reserve - Emprunts en cours"),
	)

	// Section 5: Sauvegardes
	section5Title := widget.NewLabelWithStyle("5. Gestion des Sauvegardes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section5 := container.NewVBox(
		widget.NewLabel("Acces : Configuration > Gerer les Sauvegardes"),
		widget.NewLabel(""),
		widget.NewLabel("Creer une sauvegarde :"),
		widget.NewLabel("  - Cliquez sur 'Creer une Nouvelle Sauvegarde'"),
		widget.NewLabel("  - La sauvegarde est creee instantanement"),
		widget.NewLabel(""),
		widget.NewLabel("Restaurer une sauvegarde :"),
		widget.NewLabel("  1. Selectionnez la sauvegarde dans la liste"),
		widget.NewLabel("  2. Cliquez sur 'Restaurer'"),
		widget.NewLabel("  3. Confirmez (une sauvegarde de securite est creee automatiquement)"),
		widget.NewLabel(""),
		widget.NewLabel("Supprimer une sauvegarde :"),
		widget.NewLabel("  1. Cliquez sur 'Supprimer' a cote de la sauvegarde"),
		widget.NewLabel("  2. Confirmez la suppression"),
		widget.NewLabel(""),
		widget.NewLabel("Emplacement : Les sauvegardes sont dans le dossier 'backups/'"),
		widget.NewLabel("Pensez a sauvegarder regulierement vos donnees !"),
	)

	// Section 6: Rapports et PDFs
	section6Title := widget.NewLabelWithStyle("6. Rapports et PDFs", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section6 := container.NewVBox(
		widget.NewLabel("Emprunts en Cours :"),
		widget.NewLabel("  - Vue de tous les emprunts actifs"),
		widget.NewLabel("  - Generation de recus individuels ou groupes"),
		widget.NewLabel("  - Export PDF avec toutes les informations"),
		widget.NewLabel(""),
		widget.NewLabel("Rapport des Cles :"),
		widget.NewLabel("  - Liste complete des cles sorties"),
		widget.NewLabel("  - Export PDF du rapport"),
		widget.NewLabel(""),
		widget.NewLabel("Plan de Cles :"),
		widget.NewLabel("  - Vue hierarchique : Batiments > Salles > Cles"),
		widget.NewLabel("  - Export PDF du plan complet"),
		widget.NewLabel(""),
		widget.NewLabel("Tous les PDFs supportent les caracteres accentues !"),
	)

	// Section 7: Configuration
	section7Title := widget.NewLabelWithStyle("7. Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section7 := container.NewVBox(
		widget.NewLabel("Le menu Configuration vous permet de gerer :"),
		widget.NewLabel(""),
		widget.NewLabel("Batiments : Creez et organisez vos batiments"),
		widget.NewLabel("Salles : Ajoutez des salles/points d'acces par batiment"),
		widget.NewLabel("Cles : Gerez votre inventaire de cles"),
		widget.NewLabel("Emprunteurs : Enregistrez les personnes autorisees"),
		widget.NewLabel("Sauvegardes : Gerez vos sauvegardes"),
		widget.NewLabel("Mode Demo : Chargez des donnees de test"),
		widget.NewLabel("Reinitialisation : Remettez a zero la base de donnees"),
	)

	// Section 8: Astuces
	section8Title := widget.NewLabelWithStyle("8. Astuces et Bonnes Pratiques", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section8 := container.NewVBox(
		widget.NewLabel("Sauvegardez regulierement votre base de donnees"),
		widget.NewLabel("Utilisez des numeros de cles coherents (ex: K001, K002...)"),
		widget.NewLabel("Definissez une reserve pour les cles critiques"),
		widget.NewLabel("Associez correctement les cles aux salles"),
		widget.NewLabel("Verifiez les emprunts en cours regulierement"),
		widget.NewLabel("Generez des recus PDF pour garder une trace"),
		widget.NewLabel("Utilisez le mode demo pour vous familiariser"),
		widget.NewLabel(""),
		widget.NewLabel("Attention :"),
		widget.NewLabel("  - La reinitialisation supprime TOUTES les donnees"),
		widget.NewLabel("  - Toujours confirmer avant de supprimer"),
		widget.NewLabel("  - Les sauvegardes ne sont pas synchronisees avec Git"),
	)

	// Section 9: Raccourcis
	section9Title := widget.NewLabelWithStyle("9. Navigation Rapide", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section9 := container.NewVBox(
		widget.NewLabel("Utilisez le menu de gauche pour naviguer rapidement :"),
		widget.NewLabel(""),
		widget.NewLabel("Tableau de Bord : Vue d'ensemble et actions rapides"),
		widget.NewLabel("Emprunts en Cours : Gestion des emprunts actifs"),
		widget.NewLabel("Rapport des Cles : Export et statistiques"),
		widget.NewLabel("Plan de Cles : Vue hierarchique complete"),
		widget.NewLabel("Configuration : Parametres et gestion des donnees"),
		widget.NewLabel("A Propos : Informations sur l'application"),
		widget.NewLabel("Mode d'Emploi : Ce guide (vous y etes !)"),
	)

	// Section 10: Support
	section10Title := widget.NewLabelWithStyle("10. Besoin d'Aide ?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	section10 := container.NewVBox(
		widget.NewLabel("Si vous rencontrez un probleme :"),
		widget.NewLabel(""),
		widget.NewLabel("1. Consultez ce mode d'emploi"),
		widget.NewLabel("2. Verifiez la page 'A Propos' pour les informations"),
		widget.NewLabel("3. Consultez le fichier README.md dans le dossier de l'application"),
		widget.NewLabel("4. Verifiez CHANGELOG_NOUVELLES_FONCTIONNALITES.md pour les nouveautes"),
		widget.NewLabel(""),
		widget.NewLabel("Documentation complete disponible dans les fichiers :"),
		widget.NewLabel("  - README.md - Guide complet"),
		widget.NewLabel("  - INSTALLATION.md - Installation detaillee"),
		widget.NewLabel("  - QUICK_START.md - Demarrage rapide"),
	)

	// Assembler tout le contenu
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		introTitle,
		intro,
		widget.NewSeparator(),
		section1Title,
		section1,
		widget.NewSeparator(),
		section2Title,
		section2,
		widget.NewSeparator(),
		section3Title,
		section3,
		widget.NewSeparator(),
		section4Title,
		section4,
		widget.NewSeparator(),
		section5Title,
		section5,
		widget.NewSeparator(),
		section6Title,
		section6,
		widget.NewSeparator(),
		section7Title,
		section7,
		widget.NewSeparator(),
		section8Title,
		section8,
		widget.NewSeparator(),
		section9Title,
		section9,
		widget.NewSeparator(),
		section10Title,
		section10,
	)

	return container.NewVScroll(
		container.NewPadded(content),
	)
}
