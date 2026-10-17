package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"clefs/internal/db"
	"clefs/internal/pdf"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// settings.go : paramètres de l'application (Paramètres → Établissement,
// Bons de remise, Prêts, Documents). Ils sont stockés dans la base (table
// settings) : communs à tous les postes et inclus dans les sauvegardes.

// Settings sont les paramètres modifiables dans l'interface.
type Settings struct {
	AppName       string `json:"appName"`
	Establishment string `json:"establishment"`
	Address       string `json:"address"`
	LogoDataURL   string `json:"logoDataUrl"` // data:image/png;base64,… ou vide
	FooterNote    string `json:"footerNote"`

	ReceiptTitle           string `json:"receiptTitle"`
	Engagement             string `json:"engagement"`
	SignatureBorrower      string `json:"signatureBorrower"`
	SignatureEstablishment string `json:"signatureEstablishment"`
	Signatory              string `json:"signatory"`

	DefaultLoanType   string `json:"defaultLoanType"`   // ponctuel ou permanent
	DefaultReturnDays int    `json:"defaultReturnDays"` // 0 = pas de retour prévu proposé
	AgentName         string `json:"agentName"`         // vide = nom de la session Windows
}

// Limites de saisie.
const (
	maxLogoBytes  = 1 << 20 // 1 Mo
	maxShortText  = 150
	maxEngagement = 1500
)

// defaultSettings retourne les paramètres d'origine.
func defaultSettings() Settings {
	p := pdf.DefaultSettings()
	return Settings{
		AppName: p.AppName, Establishment: p.Establishment, Address: p.Address,
		ReceiptTitle: p.ReceiptTitle, Engagement: p.Engagement,
		SignatureBorrower: p.SignatureBorrower, SignatureEstablishment: p.SignatureEstablishment,
		DefaultLoanType: "ponctuel",
	}
}

// settingFields associe chaque paramètre à sa clé dans la table settings.
func settingFields(s *Settings) map[string]*string {
	return map[string]*string{
		"app_name": &s.AppName, "establishment": &s.Establishment, "address": &s.Address,
		"logo": &s.LogoDataURL, "footer_note": &s.FooterNote,
		"receipt_title": &s.ReceiptTitle, "engagement": &s.Engagement,
		"signature_borrower": &s.SignatureBorrower, "signature_establishment": &s.SignatureEstablishment,
		"signatory": &s.Signatory, "default_loan_type": &s.DefaultLoanType, "agent_name": &s.AgentName,
	}
}

// loadSettings lit les paramètres enregistrés, complétés par les valeurs d'origine.
func loadSettings() (Settings, error) {
	stored, err := db.GetSettings()
	if err != nil {
		return Settings{}, err
	}
	s := defaultSettings()
	for key, field := range settingFields(&s) {
		if v, ok := stored[key]; ok {
			*field = v
		}
	}
	if v, err := strconv.Atoi(stored["default_return_days"]); err == nil {
		s.DefaultReturnDays = v
	}
	return s, nil
}

// GetSettings retourne les paramètres de l'application.
func (a *App) GetSettings() (Settings, error) {
	defer a.read()()
	return loadSettings()
}

// GetDefaultSettings retourne les paramètres d'origine (bouton « Rétablir »).
func (a *App) GetDefaultSettings() Settings { return defaultSettings() }

// validateSettings nettoie et contrôle les paramètres saisis.
func validateSettings(in Settings) (Settings, error) {
	def := defaultSettings()
	for key, field := range settingFields(&in) {
		*field = strings.TrimSpace(*field)
		limit := maxShortText
		switch key {
		case "logo":
			continue
		case "engagement":
			limit = maxEngagement
		case "address":
			limit = 200
		}
		if len([]rune(*field)) > limit {
			return in, fmt.Errorf("texte trop long (%d caractères au plus) : %.40s…", limit, *field)
		}
	}
	// Champs obligatoires : vides → valeur d'origine.
	for _, f := range []struct {
		v   *string
		def string
	}{
		{&in.AppName, def.AppName}, {&in.Establishment, def.Establishment},
		{&in.ReceiptTitle, def.ReceiptTitle}, {&in.Engagement, def.Engagement},
		{&in.SignatureBorrower, def.SignatureBorrower}, {&in.SignatureEstablishment, def.SignatureEstablishment},
	} {
		if *f.v == "" {
			*f.v = f.def
		}
	}
	if in.DefaultLoanType != "permanent" {
		in.DefaultLoanType = "ponctuel"
	}
	if in.DefaultReturnDays < 0 || in.DefaultReturnDays > 365 {
		return in, fmt.Errorf("la durée de retour proposée doit être comprise entre 0 et 365 jours")
	}
	if in.LogoDataURL != "" {
		if _, err := decodeLogo(in.LogoDataURL); err != nil {
			return in, err
		}
	}
	return in, nil
}

// SaveSettings enregistre les paramètres et retourne leur version enregistrée.
// Une valeur égale à celle d'origine n'est pas stockée (elle suivra les
// éventuelles évolutions des valeurs par défaut).
func (a *App) SaveSettings(in Settings) (Settings, error) {
	defer a.read()()
	s, err := validateSettings(in)
	if err != nil {
		return Settings{}, err
	}
	def := defaultSettings()
	values := map[string]string{}
	defFields := settingFields(&def)
	for key, field := range settingFields(&s) {
		values[key] = *field
		if *field == *defFields[key] {
			values[key] = ""
		}
	}
	values["default_return_days"] = ""
	if s.DefaultReturnDays > 0 {
		values["default_return_days"] = strconv.Itoa(s.DefaultReturnDays)
	}
	if err := db.SaveSettings(values); err != nil {
		return Settings{}, fmt.Errorf("enregistrement des paramètres impossible : %w", err)
	}
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, s.AppName)
	}
	return s, nil
}

// decodeLogo extrait l'image d'une URL data:… et la contrôle.
func decodeLogo(dataURL string) ([]byte, error) {
	i := strings.Index(dataURL, ";base64,")
	if !strings.HasPrefix(dataURL, "data:image/") || i < 0 {
		return nil, fmt.Errorf("logo illisible")
	}
	data, err := base64.StdEncoding.DecodeString(dataURL[i+len(";base64,"):])
	if err != nil {
		return nil, fmt.Errorf("logo illisible")
	}
	if len(data) > maxLogoBytes {
		return nil, fmt.Errorf("le logo ne doit pas dépasser 1 Mo")
	}
	if _, _, _, err := pdf.ValidateLogo(data); err != nil {
		return nil, err
	}
	return data, nil
}

// ChooseLogo ouvre la sélection d'une image et la retourne en URL data:…
// (non enregistrée : elle le sera avec les autres paramètres). "" si annulé.
func (a *App) ChooseLogo() (string, error) {
	path, err := a.chooseFile("Choisir le logo de l'établissement", "Image (*.png, *.jpg)", "*.png;*.jpg;*.jpeg")
	if err != nil || path == "" {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("lecture de l'image impossible : %w", err)
	}
	if len(data) > maxLogoBytes {
		return "", fmt.Errorf("le logo ne doit pas dépasser 1 Mo")
	}
	format, _, _, err := pdf.ValidateLogo(data)
	if err != nil {
		return "", err
	}
	mime := "image/png"
	if format == "JPG" {
		mime = "image/jpeg"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

// toPDFSettings convertit les paramètres pour le générateur de PDF.
func toPDFSettings(s Settings) pdf.Settings {
	logo, _ := decodeLogo(s.LogoDataURL) // logo invalide : clé stylisée
	return pdf.Settings{
		AppName: s.AppName, Establishment: s.Establishment, Address: s.Address, Logo: logo,
		FooterNote: s.FooterNote, ReceiptTitle: s.ReceiptTitle, Engagement: s.Engagement,
		SignatureBorrower: s.SignatureBorrower, SignatureEstablishment: s.SignatureEstablishment,
		Signatory: s.Signatory,
	}
}

// pdfSettings retourne les paramètres d'impression enregistrés.
func pdfSettings() pdf.Settings {
	s, err := loadSettings()
	if err != nil {
		s = defaultSettings()
	}
	return toPDFSettings(s)
}

// PreviewReceipt génère un bon de remise d'exemple avec les paramètres
// donnés (même non enregistrés) et retourne son chemin, pour vérifier le rendu.
func (a *App) PreviewReceipt(in Settings) (string, error) {
	defer a.read()()
	s, err := validateSettings(in)
	if err != nil {
		return "", err
	}
	planned := time.Now().AddDate(0, 0, 30)
	r := pdf.Receipt{
		BorrowerName: "Dupont Jean (exemple)", BorrowerStatus: "permanent", Date: time.Now(),
		LoanType: "ponctuel", PlannedReturn: &planned, Agent: currentAgent(),
		Keys: []pdf.ReceiptKey{
			{Number: "K-001", Description: "Salle 101", Storage: "Armoire à clés, crochet 1", LoanDate: time.Now()},
			{Number: "PG", Description: "Passe général du bâtiment A", Storage: "Coffre", LoanDate: time.Now()},
		},
	}
	data, err := pdf.GenerateReceipt(toPDFSettings(s), r)
	return a.savePDF("apercu_bon_remise", data, err)
}

// preserveSettings exécute une opération qui remplace la base (import,
// restauration) en conservant les paramètres actuels si la base obtenue
// n'en contient aucun (base d'une version antérieure).
func preserveSettings(op func() error) error {
	before, err := db.GetSettings()
	if err != nil {
		return op()
	}
	if err := op(); err != nil {
		return err
	}
	after, err := db.GetSettings()
	if err == nil && len(after) == 0 && len(before) > 0 {
		return db.SaveSettings(before)
	}
	return nil
}
