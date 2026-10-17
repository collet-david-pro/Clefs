package db

// settings.go : paramètres de l'application (table settings, clé → valeur),
// communs à tous les postes et inclus dans les sauvegardes.

// GetSettings retourne tous les paramètres enregistrés. Une clé absente
// signifie « valeur par défaut » (gérée par la couche app).
func GetSettings() (map[string]string, error) {
	rows, err := DB.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}

// SaveSettings enregistre les paramètres en une transaction (remplace les
// valeurs existantes ; une valeur vide supprime la clé).
func SaveSettings(values map[string]string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for k, v := range values {
		if v == "" {
			_, err = tx.Exec(`DELETE FROM settings WHERE key = ?`, k)
		} else {
			_, err = tx.Exec(`INSERT INTO settings (key, value) VALUES (?, ?)
				ON CONFLICT(key) DO UPDATE SET value = excluded.value`, k, v)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
