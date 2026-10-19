package main

import "testing"

// =================== //
// == TESTS BIG TXT == //
// =================== //

// Fonction générale pour tester ConversionEntree //
func TestConversionEntree(t *testing.T) {
	test := []struct {
		nom           string
		entree        string
		attendu       int
		attenduErreur bool
	}{
		{"kilo-octets", "10KB", 10 * 1024, false},
		{"méga-octets", "500MB", 500 * 1024 * 1024, false},
		{"giga-octets", "2GB", 2 * 1024 * 1024 * 1024, false},
		{"sans unité", "100", 100, false},
		{"zéro", "0MB", 0, false},
		{"pas un nombre", "abcMB", 0, true},
		{"vide", "", 0, true},
		{"unité inconnue", "5TB", 0, true},
	}
	for _, te := range test {
		resultat, err := ConversionEntree(te.entree)
		if (err != nil) != te.attenduErreur {
			t.Errorf("%s : erreur %v, erreur attendue %v", te.nom, err, te.attenduErreur)
		}
		if resultat != te.attendu {
			t.Errorf("%s : resultat %d, attendu %d", te.nom, resultat, te.attendu)
		}
	}
}
