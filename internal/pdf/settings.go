package pdf

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // décodage des logos JPEG
	_ "image/png"  // décodage des logos PNG
	"strings"
)

// Settings regroupe les réglages d'impression choisis dans les Paramètres de
// l'application : identité de l'établissement et textes des bons de remise.
// Les champs vides prennent la valeur de DefaultSettings.
type Settings struct {
	AppName       string // nom de l'application dans l'en-tête
	Establishment string // nom de l'établissement
	Address       string // adresse ou ville, à la suite du nom
	Logo          []byte // logo PNG ou JPEG ; nil = clé stylisée
	FooterNote    string // mention en pied de page de tous les documents

	ReceiptTitle           string // titre du bon de remise
	Engagement             string // texte d'engagement, avec {detenteur}, {cles}, {la_les}
	SignatureBorrower      string // titre du cadre de signature du détenteur
	SignatureEstablishment string // titre du cadre de signature de l'établissement
	Signatory              string // nom et fonction du signataire (facultatif)
}

// DefaultEngagement est le texte d'engagement par défaut du bon de remise.
const DefaultEngagement = "Je soussigné(e), {detenteur}, reconnais avoir reçu {cles}. Je m'engage à en prendre soin " +
	"et à {la_les} restituer à la fin de leur utilisation ou à première demande, et à signaler sans délai toute perte. " +
	"En cas de perte ou de dégradation, ma responsabilité pourra être engagée."

// DefaultSettings retourne les réglages d'origine.
func DefaultSettings() Settings {
	return Settings{
		AppName:                "Gestionnaire de Clés",
		Establishment:          "Collège Victor Hugo",
		Address:                "Chauny",
		ReceiptTitle:           "Bon de remise de clés",
		Engagement:             DefaultEngagement,
		SignatureBorrower:      "Le détenteur",
		SignatureEstablishment: "Pour l'établissement",
	}
}

// withDefaults complète les champs vides par les valeurs d'origine.
func (s Settings) withDefaults() Settings {
	d := DefaultSettings()
	fill := func(v *string, def string) {
		if strings.TrimSpace(*v) == "" {
			*v = def
		}
	}
	fill(&s.AppName, d.AppName)
	fill(&s.Establishment, d.Establishment)
	fill(&s.ReceiptTitle, d.ReceiptTitle)
	fill(&s.Engagement, d.Engagement)
	fill(&s.SignatureBorrower, d.SignatureBorrower)
	fill(&s.SignatureEstablishment, d.SignatureEstablishment)
	return s
}

// establishmentLine retourne « Établissement — Adresse ».
func (s Settings) establishmentLine() string {
	if strings.TrimSpace(s.Address) == "" {
		return s.Establishment
	}
	return s.Establishment + " — " + s.Address
}

// engagementText remplit le modèle de texte d'engagement pour un détenteur
// et un nombre de clés.
func (s Settings) engagementText(borrower string, n int) string {
	cles, laLes := "la clé mentionnée ci-dessus", "la"
	if n > 1 {
		cles, laLes = fmt.Sprintf("les %d clés mentionnées ci-dessus", n), "les"
	}
	return strings.NewReplacer("{detenteur}", borrower, "{cles}", cles, "{la_les}", laLes).Replace(s.Engagement)
}

// ValidateLogo vérifie qu'un logo est une image PNG ou JPEG lisible et
// retourne son format (« PNG » ou « JPG ») et ses dimensions.
func ValidateLogo(data []byte) (format string, w, h int, err error) {
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", 0, 0, fmt.Errorf("le logo doit être une image PNG ou JPEG")
	}
	switch kind {
	case "png":
		return "PNG", cfg.Width, cfg.Height, nil
	case "jpeg":
		return "JPG", cfg.Width, cfg.Height, nil
	}
	return "", 0, 0, fmt.Errorf("format d'image non pris en charge : %s", kind)
}
