// Package natsort compare des libellés dans l'ordre « naturel » : les
// nombres sont comparés par valeur (« A2 » avant « A10 », « 9 » avant
// « 10 »), le reste sans tenir compte de la casse. Sert à trier les numéros
// de clés et les noms d'accès.
package natsort

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Less indique si a précède b dans l'ordre naturel.
func Less(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	for a != "" && b != "" {
		ra, _ := utf8.DecodeRuneInString(a)
		rb, _ := utf8.DecodeRuneInString(b)
		if unicode.IsDigit(ra) && unicode.IsDigit(rb) {
			na, restA := digits(a)
			nb, restB := digits(b)
			if c := compareNumbers(na, nb); c != 0 {
				return c < 0
			}
			a, b = restA, restB
			continue
		}
		if ra != rb {
			return ra < rb
		}
		_, sa := utf8.DecodeRuneInString(a)
		_, sb := utf8.DecodeRuneInString(b)
		a, b = a[sa:], b[sb:]
	}
	return len(a) < len(b)
}

// digits sépare la suite de chiffres en tête de s du reste.
func digits(s string) (string, string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i], s[i:]
}

// compareNumbers compare deux suites de chiffres par valeur (sans limite de
// taille), puis par longueur (« 07 » après « 7 ») pour un ordre stable.
func compareNumbers(a, b string) int {
	ta, tb := strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	if len(ta) != len(tb) {
		return len(ta) - len(tb)
	}
	if c := strings.Compare(ta, tb); c != 0 {
		return c
	}
	return len(a) - len(b)
}
