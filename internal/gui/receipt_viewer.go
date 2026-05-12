package gui

import (
	"clefs/internal/db"
	"clefs/internal/pdf"
	"fmt"
	"log"
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

// ReceiptViewer gère l'affichage et l'impression des reçus
type ReceiptViewer struct {
	app         *App
	loan        *db.LoanWithDetails
	pdfContent  []byte
	htmlContent string
}

// NewReceiptViewer crée un nouveau visualiseur de reçu
func NewReceiptViewer(app *App, loan *db.LoanWithDetails) *ReceiptViewer {
	return &ReceiptViewer{
		app:  app,
		loan: loan,
	}
}

// generateHTMLReceipt génère le contenu HTML du reçu
func (rv *ReceiptViewer) generateHTMLReceipt() string {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<style>
			body {
				font-family: Arial, sans-serif;
				max-width: 600px;
				margin: 20px auto;
				padding: 20px;
				background: white;
			}
			.header {
				text-align: center;
				border-bottom: 2px solid #007BFF;
				padding-bottom: 20px;
				margin-bottom: 30px;
			}
			.title {
				font-size: 24px;
				font-weight: bold;
				color: #333;
				margin-bottom: 10px;
			}
			.subtitle {
				font-size: 14px;
				color: #666;
			}
			.section {
				margin: 20px 0;
				padding: 15px;
				background: #f8f9fa;
				border-radius: 5px;
			}
			.section-title {
				font-weight: bold;
				color: #007BFF;
				margin-bottom: 10px;
				font-size: 16px;
			}
			.info-row {
				display: flex;
				justify-content: space-between;
				margin: 8px 0;
				padding: 5px 0;
				border-bottom: 1px dotted #ddd;
			}
			.info-label {
				font-weight: bold;
				color: #555;
			}
			.info-value {
				color: #333;
			}
			.footer {
				margin-top: 40px;
				padding-top: 20px;
				border-top: 1px solid #ddd;
				text-align: center;
				font-size: 12px;
				color: #999;
			}
			.signature-box {
				margin-top: 30px;
				padding: 20px;
				border: 1px dashed #999;
				background: white;
			}
			.signature-line {
				margin-top: 40px;
				border-bottom: 1px solid #333;
				width: 250px;
				margin-left: auto;
				margin-right: auto;
			}
			@media print {
				body {
					margin: 0;
					padding: 10px;
				}
				.no-print {
					display: none;
				}
			}
		</style>
	</head>
	<body>
		<div class="header">
			<div class="title">🔑 REÇU D'EMPRUNT DE CLÉ</div>
			<div class="subtitle">Gestionnaire de Clés - Système de Gestion</div>
		</div>

		<div class="section">
			<div class="section-title">📋 INFORMATIONS DE L'EMPRUNT</div>
			<div class="info-row">
				<span class="info-label">N° de Reçu:</span>
				<span class="info-value">REC-%06d</span>
			</div>
			<div class="info-row">
				<span class="info-label">Date d'emprunt:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Heure:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="section">
			<div class="section-title">👤 EMPRUNTEUR</div>
			<div class="info-row">
				<span class="info-label">Nom:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Email:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="section">
			<div class="section-title">🔑 CLÉ EMPRUNTÉE</div>
			<div class="info-row">
				<span class="info-label">Numéro de clé:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Description:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="signature-box">
			<div class="section-title">✍️ SIGNATURE</div>
			<p style="font-size: 12px; color: #666;">
				Je reconnais avoir emprunté la clé mentionnée ci-dessus et m'engage à la restituer en bon état.
			</p>
			<div class="signature-line"></div>
			<p style="text-align: center; font-size: 12px; margin-top: 10px;">Signature de l'emprunteur</p>
		</div>

		<div class="footer">
			<p>Document généré le %s à %s</p>
			<p>Gestionnaire de Clés v2.0 - Conservez ce reçu jusqu'au retour de la clé</p>
		</div>
	</body>
	</html>
	`,
		rv.loan.ID,
		rv.loan.LoanDate.Format("02/01/2006"),
		rv.loan.LoanDate.Format("15:04"),
		rv.loan.BorrowerName,
		rv.loan.BorrowerEmail,
		rv.loan.KeyNumber,
		rv.loan.KeyDescription,
		time.Now().Format("02/01/2006"),
		time.Now().Format("15:04"),
	)

	return html
}

// generatePDF génère le PDF du reçu
func (rv *ReceiptViewer) generatePDF() ([]byte, error) {
	// Utiliser le générateur PDF existant
	pdfBytes, err := pdf.GenerateLoanReceipt(rv.loan)
	if err != nil {
		return nil, fmt.Errorf("erreur génération PDF: %v", err)
	}
	return pdfBytes, nil
}

// Show affiche le visualiseur de reçu
func (rv *ReceiptViewer) Show() {
	// Générer le contenu HTML
	rv.htmlContent = rv.generateHTMLReceipt()

	// Générer le PDF en arrière-plan
	go func() {
		var err error
		rv.pdfContent, err = rv.generatePDF()
		if err != nil {
			log.Printf("Erreur génération PDF: %v", err)
		}
	}()

	// Créer un widget HTML custom
	htmlDisplay := widget.NewCard("", "",
		container.NewScroll(widget.NewLabel(rv.getSimplifiedHTML())),
	)
	htmlDisplay.Resize(fyne.NewSize(600, 500))

	// Boutons d'action
	printBtn := widget.NewButton("🖨️ Imprimer", func() {
		rv.print()
	})
	printBtn.Importance = widget.HighImportance

	exportBtn := widget.NewButton("💾 Exporter PDF", func() {
		rv.exportPDF()
	})

	previewBtn := widget.NewButton("👁️ Aperçu Navigateur", func() {
		rv.openInBrowser()
	})

	closeBtn := widget.NewButton("Fermer", func() {
		// Fermeture gérée par le dialog
	})

	// Layout
	buttons := container.NewHBox(
		printBtn,
		exportBtn,
		previewBtn,
		widget.NewSeparator(),
		closeBtn,
	)

	content := container.NewBorder(
		widget.NewLabelWithStyle("📄 Aperçu du Reçu", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		buttons,
		nil,
		nil,
		htmlDisplay,
	)

	// Créer et afficher le dialog
	dialog := dialog.NewCustom("Reçu d'Emprunt", "Fermer", content, rv.app.window)
	dialog.Resize(fyne.NewSize(700, 600))
	dialog.Show()
}

// getSimplifiedHTML retourne une version simplifiée pour l'affichage dans Fyne
func (rv *ReceiptViewer) getSimplifiedHTML() string {
	return fmt.Sprintf(`
REÇU D'EMPRUNT DE CLÉ
═══════════════════════════════════════

📋 INFORMATIONS DE L'EMPRUNT
────────────────────────────
N° de Reçu:        REC-%06d
Date d'emprunt:    %s
Heure:             %s

👤 EMPRUNTEUR
────────────────────────────
Nom:               %s
Email:             %s

🔑 CLÉ EMPRUNTÉE
────────────────────────────
Numéro de clé:     %s
Description:       %s

✍️ SIGNATURE
────────────────────────────
Je reconnais avoir emprunté la clé mentionnée
ci-dessus et m'engage à la restituer en bon état.


_______________________________
Signature de l'emprunteur


═══════════════════════════════════════
Document généré le %s à %s
Gestionnaire de Clés v2.0
Conservez ce reçu jusqu'au retour de la clé
`,
		rv.loan.ID,
		rv.loan.LoanDate.Format("02/01/2006"),
		rv.loan.LoanDate.Format("15:04"),
		rv.loan.BorrowerName,
		rv.loan.BorrowerEmail,
		rv.loan.KeyNumber,
		rv.loan.KeyDescription,
		time.Now().Format("02/01/2006"),
		time.Now().Format("15:04"),
	)
}

// print imprime le reçu
func (rv *ReceiptViewer) print() {
	if rv.pdfContent == nil {
		// Générer le PDF si pas encore fait
		var err error
		rv.pdfContent, err = rv.generatePDF()
		if err != nil {
			rv.app.showError("Erreur", fmt.Sprintf("Impossible de générer le PDF: %v", err))
			return
		}
	}

	// Créer un fichier temporaire
	tmpFile, err := os.CreateTemp("", "receipt_*.pdf")
	if err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Impossible de créer le fichier temporaire: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	// Écrire le PDF
	if _, err := tmpFile.Write(rv.pdfContent); err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le PDF: %v", err))
		return
	}
	tmpFile.Close()

	// Imprimer selon l'OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("lpr", tmpFile.Name())
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "/min", "notepad", "/p", tmpFile.Name())
	case "linux":
		cmd = exec.Command("lpr", tmpFile.Name())
	default:
		rv.app.showError("Erreur", "Impression non supportée sur cet OS")
		return
	}

	if err := cmd.Run(); err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Erreur lors de l'impression: %v", err))
		return
	}

	rv.app.showSuccess("Document envoyé à l'imprimante")
}

// exportPDF exporte le PDF
func (rv *ReceiptViewer) exportPDF() {
	if rv.pdfContent == nil {
		// Générer le PDF si pas encore fait
		var err error
		rv.pdfContent, err = rv.generatePDF()
		if err != nil {
			rv.app.showError("Erreur", fmt.Sprintf("Impossible de générer le PDF: %v", err))
			return
		}
	}

	// Créer un dialog de sauvegarde
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			rv.app.showError("Erreur", fmt.Sprintf("Erreur lors de la sauvegarde: %v", err))
			return
		}
		if writer == nil {
			return
		}

		// Écrire le PDF
		if _, err := writer.Write(rv.pdfContent); err != nil {
			rv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le fichier: %v", err))
			return
		}

		rv.app.showSuccess("PDF exporté avec succès")
	}, rv.app.window)

	saveDialog.SetFileName(fmt.Sprintf("recu_emprunt_%d.pdf", rv.loan.ID))
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}

// openInBrowser ouvre l'aperçu dans le navigateur
func (rv *ReceiptViewer) openInBrowser() {
	// Créer un fichier HTML temporaire
	tmpFile, err := os.CreateTemp("", "receipt_*.html")
	if err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Impossible de créer le fichier temporaire: %v", err))
		return
	}

	// Écrire le HTML
	if _, err := tmpFile.WriteString(rv.htmlContent); err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Impossible d'écrire le HTML: %v", err))
		return
	}
	tmpFile.Close()

	// Ouvrir dans le navigateur par défaut
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", tmpFile.Name())
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", tmpFile.Name())
	case "linux":
		cmd = exec.Command("xdg-open", tmpFile.Name())
	default:
		rv.app.showError("Erreur", "Ouverture du navigateur non supportée sur cet OS")
		return
	}

	if err := cmd.Start(); err != nil {
		rv.app.showError("Erreur", fmt.Sprintf("Impossible d'ouvrir le navigateur: %v", err))
		return
	}

	// Nettoyer le fichier après 10 secondes
	go func() {
		time.Sleep(10 * time.Second)
		os.Remove(tmpFile.Name())
	}()
}

// ShowReceiptForLoan affiche le reçu pour un emprunt donné
func ShowReceiptForLoan(app *App, loanID int) {
	// Récupérer les détails de l'emprunt
	loan, err := db.GetLoanByID(loanID)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Impossible de charger l'emprunt: %v", err))
		return
	}

	// Créer et afficher le visualiseur
	viewer := NewReceiptViewer(app, loan)
	viewer.Show()
}

// GenerateAndSaveReceipt génère et enregistre un reçu PDF
func GenerateAndSaveReceipt(app *App, loan *db.LoanWithDetails) {
	// Générer le PDF
	pdfData, err := pdf.GenerateLoanReceipt(loan)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de la génération du PDF: %v", err))
		return
	}

	// Enregistrer automatiquement
	filename := pdf.GenerateFilename("recu_emprunt", loan.ID)
	filepath, err := pdf.SavePDF(filename, pdfData)
	if err != nil {
		app.showError("Erreur", fmt.Sprintf("Erreur lors de l'enregistrement: %v", err))
		return
	}

	app.showSuccess(fmt.Sprintf("✅ Reçu enregistré : %s", filepath))
}

// GenerateReceiptHTML génère le HTML pour un reçu d'emprunt
func GenerateReceiptHTML(loan *db.LoanWithDetails) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Reçu d'Emprunt</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 15px;
			padding: 40px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			border-bottom: 3px solid #667eea;
			padding-bottom: 20px;
			margin-bottom: 30px;
		}
		.title {
			font-size: 28px;
			font-weight: bold;
			color: #333;
			margin-bottom: 10px;
		}
		.subtitle {
			font-size: 14px;
			color: #666;
		}
		.section {
			margin: 25px 0;
			padding: 20px;
			background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%);
			border-radius: 10px;
		}
		.section-title {
			font-weight: bold;
			color: #667eea;
			margin-bottom: 15px;
			font-size: 18px;
			display: flex;
			align-items: center;
		}
		.info-row {
			display: flex;
			justify-content: space-between;
			margin: 10px 0;
			padding: 8px 0;
			border-bottom: 1px dotted #ddd;
		}
		.info-label {
			font-weight: 600;
			color: #555;
		}
		.info-value {
			color: #333;
			font-weight: 500;
		}
		.signature-box {
			margin-top: 40px;
			padding: 25px;
			border: 2px dashed #667eea;
			background: #f8f9ff;
			border-radius: 10px;
		}
		.signature-line {
			margin-top: 50px;
			border-bottom: 2px solid #333;
			width: 300px;
			margin-left: auto;
			margin-right: auto;
		}
		.footer {
			margin-top: 40px;
			padding-top: 20px;
			border-top: 1px solid #ddd;
			text-align: center;
			font-size: 12px;
			color: #999;
		}
		.receipt-number {
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			color: white;
			padding: 5px 15px;
			border-radius: 20px;
			font-weight: bold;
		}
		@media print {
			body {
				background: white;
			}
			.container {
				box-shadow: none;
				padding: 20px;
			}
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<div class="title">🔑 REÇU D'EMPRUNT DE CLÉ</div>
			<div class="subtitle">Gestionnaire de Clés - Système de Gestion</div>
		</div>

		<div class="section">
			<div class="section-title">📋 INFORMATIONS DE L'EMPRUNT</div>
			<div class="info-row">
				<span class="info-label">N° de Reçu:</span>
				<span class="info-value"><span class="receipt-number">REC-%06d</span></span>
			</div>
			<div class="info-row">
				<span class="info-label">Date d'emprunt:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Heure:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="section">
			<div class="section-title">👤 EMPRUNTEUR</div>
			<div class="info-row">
				<span class="info-label">Nom:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Email:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="section">
			<div class="section-title">🔑 CLÉ EMPRUNTÉE</div>
			<div class="info-row">
				<span class="info-label">Numéro de clé:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Description:</span>
				<span class="info-value">%s</span>
			</div>
		</div>

		<div class="signature-box">
			<div class="section-title">✍️ SIGNATURE</div>
			<p style="font-size: 14px; color: #666; text-align: center;">
				Je reconnais avoir emprunté la clé mentionnée ci-dessus et m'engage à la restituer en bon état.
			</p>
			<div class="signature-line"></div>
			<p style="text-align: center; font-size: 12px; margin-top: 10px; color: #666;">Signature de l'emprunteur</p>
		</div>

		<div class="footer">
			<p>Document généré le %s à %s</p>
			<p>Gestionnaire de Clés v2.0 - Conservez ce reçu jusqu'au retour de la clé</p>
		</div>
	</div>
</body>
</html>`,
		loan.ID,
		loan.LoanDate.Format("02/01/2006"),
		loan.LoanDate.Format("15:04"),
		loan.BorrowerName,
		loan.BorrowerEmail,
		loan.KeyNumber,
		loan.KeyDescription,
		time.Now().Format("02/01/2006"),
		time.Now().Format("15:04"),
	)
}
