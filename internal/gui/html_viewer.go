package gui

import (
	"clefs/internal/db"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// HTMLViewer gère l'affichage HTML universel pour tous les documents
type HTMLViewer struct {
	app          *App
	title        string
	htmlContent  string
	pdfContent   []byte
	pdfGenerator func() ([]byte, error)
}

// NewHTMLViewer crée un nouveau visualiseur HTML
func NewHTMLViewer(app *App, title string) *HTMLViewer {
	return &HTMLViewer{
		app:   app,
		title: title,
	}
}

// SetHTMLContent définit le contenu HTML
func (hv *HTMLViewer) SetHTMLContent(html string) {
	hv.htmlContent = html
}

// SetPDFGenerator définit la fonction de génération PDF
func (hv *HTMLViewer) SetPDFGenerator(generator func() ([]byte, error)) {
	hv.pdfGenerator = generator
}

// Show affiche le visualiseur
func (hv *HTMLViewer) Show() {
	// Générer le PDF en arrière-plan si un générateur est fourni
	if hv.pdfGenerator != nil {
		go func() {
			var err error
			hv.pdfContent, err = hv.pdfGenerator()
			if err != nil {
				fmt.Printf("Erreur génération PDF: %v\n", err)
			}
		}()
	}

	// Créer un conteneur avec iframe simulé
	htmlDisplay := hv.createHTMLDisplay()

	// Boutons d'action
	buttons := hv.createActionButtons()

	content := container.NewBorder(
		widget.NewLabelWithStyle(hv.title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		buttons,
		nil,
		nil,
		htmlDisplay,
	)

	// Créer et afficher le dialog
	dialog := dialog.NewCustom(hv.title, "Fermer", content, hv.app.window)
	dialog.Resize(fyne.NewSize(900, 700))
	dialog.Show()
}

// createHTMLDisplay crée l'affichage HTML interne
func (hv *HTMLViewer) createHTMLDisplay() fyne.CanvasObject {
	// Créer un fichier HTML temporaire pour l'affichage
	tmpFile, err := os.CreateTemp("", "preview_*.html")
	if err != nil {
		return widget.NewLabel("Erreur lors de la création de l'aperçu")
	}

	// Écrire le contenu HTML
	tmpFile.WriteString(hv.htmlContent)
	tmpFile.Close()

	// Créer un conteneur avec un message et un bouton pour ouvrir dans le navigateur
	// Note: Fyne n'a pas de WebView natif, donc on simule avec un aperçu texte et option navigateur

	// Extraire le texte du HTML pour l'aperçu (version simplifiée)
	previewText := hv.extractTextFromHTML()

	textWidget := widget.NewMultiLineEntry()
	textWidget.SetText(previewText)
	textWidget.Disable() // Read-only

	// Bouton pour ouvrir dans le navigateur interne (simulé)
	openInternalBtn := widget.NewButton("🌐 Ouvrir l'aperçu complet", func() {
		hv.openInBrowser(tmpFile.Name())
	})
	openInternalBtn.Importance = widget.HighImportance

	// Nettoyer après 30 secondes
	go func() {
		time.Sleep(30 * time.Second)
		os.Remove(tmpFile.Name())
	}()

	return container.NewBorder(
		openInternalBtn,
		nil,
		nil,
		nil,
		container.NewScroll(textWidget),
	)
}

// extractTextFromHTML extrait le texte du HTML pour l'aperçu
func (hv *HTMLViewer) extractTextFromHTML() string {
	// Version simplifiée - dans un cas réel, on utiliserait un parser HTML
	// Pour l'instant, on retourne un aperçu basique
	return `
📄 APERÇU DU DOCUMENT
═══════════════════════════════════════

Cliquez sur "Ouvrir l'aperçu complet" ci-dessus pour voir le document formaté avec :
• Mise en page professionnelle
• Couleurs et styles
• Tableaux et sections organisées
• Format optimisé pour l'impression

Le document complet s'ouvrira dans votre navigateur par défaut avec toutes les fonctionnalités de mise en page.

Vous pouvez également :
• Imprimer directement depuis cette fenêtre
• Exporter en PDF haute qualité
• Rafraîchir l'aperçu si nécessaire

═══════════════════════════════════════
`
}

// createActionButtons crée les boutons d'action
func (hv *HTMLViewer) createActionButtons() fyne.CanvasObject {
	printBtn := widget.NewButton("🖨️ Imprimer", func() {
		hv.print()
	})
	printBtn.Importance = widget.HighImportance

	exportBtn := widget.NewButton("💾 Exporter PDF", func() {
		hv.exportPDF()
	})

	refreshBtn := widget.NewButton("🔄 Rafraîchir", func() {
		// Recréer l'affichage
		hv.Show()
	})

	return container.NewHBox(
		printBtn,
		exportBtn,
		refreshBtn,
	)
}

// openInBrowser ouvre dans le navigateur
func (hv *HTMLViewer) openInBrowser(filepath string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filepath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filepath)
	case "linux":
		cmd = exec.Command("xdg-open", filepath)
	default:
		hv.app.showError("Erreur", "Ouverture du navigateur non supportée sur cet OS")
		return
	}

	if err := cmd.Start(); err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Impossible d'ouvrir le navigateur: %v", err))
	}
}

// print imprime le document
func (hv *HTMLViewer) print() {
	if hv.pdfContent == nil && hv.pdfGenerator != nil {
		// Générer le PDF si pas encore fait
		var err error
		hv.pdfContent, err = hv.pdfGenerator()
		if err != nil {
			hv.app.showError("Erreur", fmt.Sprintf("Impossible de générer le PDF: %v", err))
			return
		}
	}

	if hv.pdfContent == nil {
		// Si pas de PDF, imprimer le HTML
		hv.printHTML()
		return
	}

	// Créer un fichier temporaire pour le PDF
	tmpFile, err := os.CreateTemp("", "print_*.pdf")
	if err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Impossible de créer le fichier temporaire: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	// Écrire le PDF
	if _, err := tmpFile.Write(hv.pdfContent); err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le PDF: %v", err))
		return
	}
	tmpFile.Close()

	// Imprimer selon l'OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("lpr", tmpFile.Name())
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "/min", "notepad", "/p", tmpFile.Name())
	case "linux":
		cmd = exec.Command("lpr", tmpFile.Name())
	default:
		hv.app.showError("Erreur", "Impression non supportée sur cet OS")
		return
	}

	if err := cmd.Run(); err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Erreur lors de l'impression: %v", err))
		return
	}

	hv.app.showSuccess("Document envoyé à l'imprimante")
}

// printHTML imprime le HTML directement
func (hv *HTMLViewer) printHTML() {
	// Créer un fichier HTML temporaire
	tmpFile, err := os.CreateTemp("", "print_*.html")
	if err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Impossible de créer le fichier temporaire: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	// Écrire le HTML
	if _, err := tmpFile.WriteString(hv.htmlContent); err != nil {
		hv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le HTML: %v", err))
		return
	}
	tmpFile.Close()

	// Ouvrir dans le navigateur pour impression
	hv.openInBrowser(tmpFile.Name())
	dialog.ShowInformation("Impression", "Le document s'est ouvert dans votre navigateur. Utilisez Ctrl+P (ou Cmd+P sur Mac) pour imprimer.", hv.app.window)
}

// exportPDF exporte le PDF
func (hv *HTMLViewer) exportPDF() {
	if hv.pdfContent == nil && hv.pdfGenerator != nil {
		// Générer le PDF si pas encore fait
		var err error
		hv.pdfContent, err = hv.pdfGenerator()
		if err != nil {
			hv.app.showError("Erreur", fmt.Sprintf("Impossible de générer le PDF: %v", err))
			return
		}
	}

	if hv.pdfContent == nil {
		hv.app.showError("Erreur", "Aucun PDF à exporter")
		return
	}

	// Créer un dialog de sauvegarde
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			hv.app.showError("Erreur", fmt.Sprintf("Erreur lors de la sauvegarde: %v", err))
			return
		}
		if writer == nil {
			return
		}

		// Écrire le PDF
		if _, err := writer.Write(hv.pdfContent); err != nil {
			hv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le fichier: %v", err))
			return
		}

		hv.app.showSuccess("PDF exporté avec succès")
	}, hv.app.window)

	saveDialog.SetFileName(fmt.Sprintf("document_%s.pdf", time.Now().Format("20060102_150405")))
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}

// Fonctions helper pour générer le HTML des différents rapports

// GenerateKeyPlanHTML génère le HTML pour le plan de clés
func GenerateKeyPlanHTML(buildings map[int]db.Building) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Plan de Clés</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 15px;
			padding: 30px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			margin-bottom: 40px;
			padding-bottom: 20px;
			border-bottom: 3px solid #667eea;
		}
		h1 {
			color: #333;
			font-size: 2.5em;
			margin: 0;
		}
		.subtitle {
			color: #666;
			margin-top: 10px;
		}
		.building {
			margin: 30px 0;
			background: #f8f9fa;
			border-radius: 10px;
			padding: 20px;
			border-left: 5px solid #667eea;
		}
		.building-name {
			font-size: 1.5em;
			color: #667eea;
			font-weight: bold;
			margin-bottom: 15px;
		}
		.room {
			margin: 15px 0;
			padding: 15px;
			background: white;
			border-radius: 8px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
		}
		.room-name {
			font-weight: bold;
			color: #333;
			font-size: 1.1em;
			margin-bottom: 10px;
		}
		.room-type {
			color: #888;
			font-size: 0.9em;
			font-style: italic;
		}
		.keys-list {
			margin-top: 10px;
			padding-left: 20px;
		}
		.key-item {
			margin: 5px 0;
			padding: 8px;
			background: #f0f4ff;
			border-radius: 5px;
			border-left: 3px solid #764ba2;
		}
		.key-number {
			font-weight: bold;
			color: #764ba2;
		}
		.no-keys {
			color: #999;
			font-style: italic;
		}
		.footer {
			margin-top: 40px;
			text-align: center;
			color: #666;
			font-size: 0.9em;
			padding-top: 20px;
			border-top: 1px solid #ddd;
		}
		@media print {
			body {
				background: white;
			}
			.container {
				box-shadow: none;
			}
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🏢 Plan de Clés</h1>
			<div class="subtitle">Généré le %s</div>
		</div>
		
		<div class="content">`

	html += fmt.Sprintf(html, time.Now().Format("02/01/2006 à 15:04"))

	// Ajouter les bâtiments
	for _, building := range buildings {
		html += fmt.Sprintf(`
		<div class="building">
			<div class="building-name">%s</div>`, building.Name)

		// Ajouter les salles
		for _, room := range building.Rooms {
			html += fmt.Sprintf(`
			<div class="room">
				<div class="room-name">%s`, room.Name)

			if room.Type != "" {
				html += fmt.Sprintf(` <span class="room-type">(%s)</span>`, room.Type)
			}
			html += `</div>`

			// Ajouter les clés
			if len(room.Keys) > 0 {
				html += `<div class="keys-list">`
				for _, key := range room.Keys {
					html += fmt.Sprintf(`
					<div class="key-item">
						<span class="key-number">Clé %s</span> - %s
					</div>`, key.Number, key.Description)
				}
				html += `</div>`
			} else {
				html += `<div class="no-keys">Aucune clé associée</div>`
			}

			html += `</div>`
		}

		html += `</div>`
	}

	html += `
		</div>
		<div class="footer">
			<p>Gestionnaire de Clés v2.0 - Document généré automatiquement</p>
		</div>
	</div>
</body>
</html>`

	return html
}

// GenerateLoansReportHTML génère le HTML pour le rapport des emprunts
func GenerateLoansReportHTML(loans []db.LoanWithDetails) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Rapport des Clés Sorties</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 15px;
			padding: 30px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			margin-bottom: 40px;
			padding-bottom: 20px;
			border-bottom: 3px solid #f5576c;
		}
		h1 {
			color: #333;
			font-size: 2.5em;
			margin: 0;
		}
		.stats {
			display: flex;
			justify-content: center;
			gap: 30px;
			margin: 20px 0;
		}
		.stat-box {
			background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
			color: white;
			padding: 15px 30px;
			border-radius: 10px;
			text-align: center;
		}
		.stat-number {
			font-size: 2em;
			font-weight: bold;
		}
		.stat-label {
			font-size: 0.9em;
			opacity: 0.9;
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 30px;
		}
		th {
			background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
			color: white;
			padding: 15px;
			text-align: left;
			font-weight: bold;
		}
		td {
			padding: 12px 15px;
			border-bottom: 1px solid #eee;
		}
		tr:hover {
			background: #f8f9fa;
		}
		.key-number {
			font-weight: bold;
			color: #f5576c;
		}
		.borrower-name {
			color: #333;
			font-weight: 500;
		}
		.date {
			color: #666;
		}
		.footer {
			margin-top: 40px;
			text-align: center;
			color: #666;
			font-size: 0.9em;
			padding-top: 20px;
			border-top: 1px solid #ddd;
		}
		@media print {
			body {
				background: white;
			}
			.container {
				box-shadow: none;
			}
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📊 Rapport des Clés Sorties</h1>
			<div class="stats">
				<div class="stat-box">
					<div class="stat-number">%d</div>
					<div class="stat-label">Emprunts Actifs</div>
				</div>
			</div>
			<div style="color: #666; margin-top: 15px;">Généré le %s</div>
		</div>
		
		<table>
			<thead>
				<tr>
					<th>Clé</th>
					<th>Description</th>
					<th>Emprunteur</th>
					<th>Date d'emprunt</th>
				</tr>
			</thead>
			<tbody>`

	html = fmt.Sprintf(html, len(loans), time.Now().Format("02/01/2006 à 15:04"))

	// Ajouter les lignes du tableau
	for _, loan := range loans {
		html += fmt.Sprintf(`
				<tr>
					<td><span class="key-number">%s</span></td>
					<td>%s</td>
					<td><span class="borrower-name">%s</span></td>
					<td><span class="date">%s</span></td>
				</tr>`,
			loan.KeyNumber,
			loan.KeyDescription,
			loan.BorrowerName,
			loan.LoanDate.Format("02/01/2006"),
		)
	}

	html += `
			</tbody>
		</table>
		
		<div class="footer">
			<p>Gestionnaire de Clés v2.0 - Document généré automatiquement</p>
		</div>
	</div>
</body>
</html>`

	return html
}

// GenerateGlobalBorrowerReportHTML génère le HTML pour le rapport global des emprunts
func GenerateGlobalBorrowerReportHTML(loansByBorrower map[string][]db.LoanWithDetails) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Rapport Global des Emprunts</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 15px;
			padding: 30px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			margin-bottom: 40px;
			padding-bottom: 20px;
			border-bottom: 3px solid #8ec5fc;
		}
		h1 {
			color: #333;
			font-size: 2.5em;
			margin: 0;
		}
		.summary {
			background: #f0f7ff;
			padding: 15px;
			border-radius: 8px;
			text-align: center;
			margin-bottom: 30px;
			border: 1px solid #cce5ff;
		}
		.borrower-section {
			margin-bottom: 30px;
			border: 1px solid #eee;
			border-radius: 8px;
			overflow: hidden;
		}
		.borrower-header {
			background: #f8f9fa;
			padding: 15px;
			border-bottom: 1px solid #eee;
			font-weight: bold;
			color: #333;
			font-size: 1.2em;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		.badge {
			background: #8ec5fc;
			color: white;
			padding: 5px 10px;
			border-radius: 15px;
			font-size: 0.8em;
		}
		table {
			width: 100%;
			border-collapse: collapse;
		}
		th {
			background: #f1f1f1;
			color: #666;
			padding: 10px 15px;
			text-align: left;
			font-size: 0.9em;
			text-transform: uppercase;
		}
		td {
			padding: 12px 15px;
			border-bottom: 1px solid #eee;
		}
		tr:last-child td {
			border-bottom: none;
		}
		.key-number {
			font-weight: bold;
			color: #667eea;
		}
		.duration {
			color: #888;
			font-style: italic;
		}
		.footer {
			margin-top: 40px;
			text-align: center;
			color: #666;
			font-size: 0.9em;
			padding-top: 20px;
			border-top: 1px solid #ddd;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📋 Rapport Global des Emprunts</h1>
			<div style="color: #666; margin-top: 10px;">Généré le %s</div>
		</div>`

	html = fmt.Sprintf(html, time.Now().Format("02/01/2006 à 15:04"))

	// Calculer le total
	totalLoans := 0
	for _, loans := range loansByBorrower {
		totalLoans += len(loans)
	}

	html += fmt.Sprintf(`
		<div class="summary">
			<strong>Total :</strong> %d emprunteurs actifs | %d clés sorties
		</div>`, len(loansByBorrower), totalLoans)

	// Pour chaque emprunteur
	for borrower, loans := range loansByBorrower {
		html += fmt.Sprintf(`
		<div class="borrower-section">
			<div class="borrower-header">
				<span>👤 %s</span>
				<span class="badge">%d clés</span>
			</div>
			<table>
				<thead>
					<tr>
						<th>Clé</th>
						<th>Description</th>
						<th>Date d'emprunt</th>
						<th>Durée</th>
					</tr>
				</thead>
				<tbody>`, borrower, len(loans))

		for _, loan := range loans {
			days := int(time.Since(loan.LoanDate).Hours() / 24)
			duration := fmt.Sprintf("%d jours", days)
			if days == 0 {
				duration = "Aujourd'hui"
			}

			html += fmt.Sprintf(`
					<tr>
						<td><span class="key-number">%s</span></td>
						<td>%s</td>
						<td>%s</td>
						<td><span class="duration">%s</span></td>
					</tr>`,
				loan.KeyNumber,
				loan.KeyDescription,
				loan.LoanDate.Format("02/01/2006"),
				duration)
		}

		html += `
				</tbody>
			</table>
		</div>`
	}

	html += `
		<div class="footer">
			<p>Gestionnaire de Clés v2.0 - Document généré automatiquement</p>
		</div>
	</div>
</body>
</html>`

	return html
}

// GenerateKeyStockReportHTML génère le HTML pour le bilan du stock
func GenerateKeyStockReportHTML(keys []db.Key, loanCounts map[int]int) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Bilan du Stock de Clés</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			max-width: 1200px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 15px;
			padding: 30px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			margin-bottom: 40px;
			padding-bottom: 20px;
			border-bottom: 3px solid #a1c4fd;
		}
		h1 {
			color: #333;
			font-size: 2.5em;
			margin: 0;
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 20px;
		}
		th {
			background: #e3f2fd;
			color: #1565c0;
			padding: 15px;
			text-align: left;
			font-weight: bold;
			border-bottom: 2px solid #bbdefb;
		}
		td {
			padding: 12px 15px;
			border-bottom: 1px solid #eee;
		}
		tr:hover {
			background: #f5f5f5;
		}
		.key-number {
			font-weight: bold;
			color: #1565c0;
		}
		.stock-ok {
			color: #2e7d32;
			font-weight: bold;
		}
		.stock-low {
			color: #ef6c00;
			font-weight: bold;
			background: #fff3e0;
			padding: 2px 8px;
			border-radius: 10px;
		}
		.stock-critical {
			color: #c62828;
			font-weight: bold;
			background: #ffebee;
			padding: 2px 8px;
			border-radius: 10px;
		}
		.footer {
			margin-top: 40px;
			text-align: center;
			color: #666;
			font-size: 0.9em;
			padding-top: 20px;
			border-top: 1px solid #ddd;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📦 Bilan du Stock de Clés</h1>
			<div style="color: #666; margin-top: 10px;">Généré le %s</div>
		</div>
		
		<table>
			<thead>
				<tr>
					<th>Numéro</th>
					<th>Description</th>
					<th style="text-align: center;">Total</th>
					<th style="text-align: center;">Réserve</th>
					<th style="text-align: center;">Sorties</th>
					<th style="text-align: center;">Disponibles</th>
				</tr>
			</thead>
			<tbody>`

	html = fmt.Sprintf(html, time.Now().Format("02/01/2006 à 15:04"))

	for _, key := range keys {
		borrowed := loanCounts[key.ID]
		available := key.QuantityTotal - key.QuantityReserve - borrowed

		availClass := "stock-ok"
		if available <= 0 {
			availClass = "stock-critical"
		} else if available == 1 {
			availClass = "stock-low"
		}

		html += fmt.Sprintf(`
				<tr>
					<td><span class="key-number">%s</span></td>
					<td>%s</td>
					<td style="text-align: center;">%d</td>
					<td style="text-align: center;">%d</td>
					<td style="text-align: center;">%d</td>
					<td style="text-align: center;"><span class="%s">%d</span></td>
				</tr>`,
			key.Number,
			key.Description,
			key.QuantityTotal,
			key.QuantityReserve,
			borrowed,
			availClass,
			available)
	}

	html += `
			</tbody>
		</table>
		
		<div class="footer">
			<p>Gestionnaire de Clés v2.0 - Document généré automatiquement</p>
		</div>
	</div>
</body>
</html>`

	return html
}
