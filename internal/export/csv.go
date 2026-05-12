package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const documentsDir = "documents"

// WriteCSV écrit un CSV UTF-8 avec BOM (compatible Excel France) dans w.
// Séparateur point-virgule, standard France.
func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	// BOM UTF-8 pour qu'Excel ouvre sans configuration
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// SaveCSV enregistre un CSV dans le dossier documents/ et retourne le chemin.
func SaveCSV(filename string, headers []string, rows [][]string) (string, error) {
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		return "", fmt.Errorf("impossible de créer le dossier documents: %w", err)
	}
	filePath := filepath.Join(documentsDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("impossible de créer le fichier: %w", err)
	}
	defer f.Close()
	if err := WriteCSV(f, headers, rows); err != nil {
		return "", err
	}
	return filePath, nil
}

// Filename génère un nom de fichier horodaté.
func Filename(prefix string) string {
	return fmt.Sprintf("%s_%s.csv", prefix, time.Now().Format("20060102_150405"))
}
