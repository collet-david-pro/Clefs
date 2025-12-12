package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const documentsDir = "documents"

// EnsureDocumentsDir crée le dossier documents s'il n'existe pas
func EnsureDocumentsDir() error {
	if _, err := os.Stat(documentsDir); os.IsNotExist(err) {
		return os.MkdirAll(documentsDir, 0755)
	}
	return nil
}

// SavePDF enregistre un PDF dans le dossier documents et retourne le chemin complet
func SavePDF(filename string, data []byte) (string, error) {
	// Créer le dossier si nécessaire
	if err := EnsureDocumentsDir(); err != nil {
		return "", fmt.Errorf("impossible de créer le dossier documents: %v", err)
	}

	// Chemin complet du fichier
	filepath := filepath.Join(documentsDir, filename)

	// Écrire le fichier
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("impossible d'écrire le fichier: %v", err)
	}

	return filepath, nil
}

// GenerateFilename génère un nom de fichier avec la date
func GenerateFilename(prefix string, id int) string {
	timestamp := time.Now().Format("20060102_150405")
	if id > 0 {
		return fmt.Sprintf("%s_%d_%s.pdf", prefix, id, timestamp)
	}
	return fmt.Sprintf("%s_%s.pdf", prefix, timestamp)
}
