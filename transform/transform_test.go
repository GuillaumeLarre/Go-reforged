package transform

import (
	"go-reforged/lexer"
	"go-reforged/render"
	"testing"
)

func TestRuleModificateurApply(t *testing.T) {
	test := []struct {
		nom                string
		entree             string
		attendu            string
		attenduDiagnostics int
	}{
		{"up avec nombre et ponctuation ignorée", "hello , world (up, 2)", "HELLO , WORLD", 0},
		{"sujet up", "This is so exciting (up, 2)", "This is SO EXCITING", 0},
		{"sujet low", "I should stop SHOUTING (low)", "I should stop shouting", 0},
		{"sujet cap", "Welcome to the Brooklyn bridge (cap)", "Welcome to the Brooklyn Bridge", 0},
		{"cap accentué", "élan (cap)", "Élan", 0},
		{"up accentué sur plusieurs mots", "élan déjà vu (up, 3)", "ÉLAN DÉJÀ VU", 0},
		{"deux modificateurs sur le même mot", "hello (cap) (up)", "HELLO", 0},
		{"le modificateur précédent n'est pas compté", "one TWO (low) three (up, 2)", "one TWO THREE", 0},
		{"guillemets ignorés", `she said " hi there " (up, 2)`, `she said " HI THERE "`, 0},
		{"nombre négatif", "hello (up, -1)", "hello (up, -1)", 1},
		{"nombre illisible", "hello (up, two)", "hello (up, two)", 1},
		{"nombre nul", "hello (up, 0)", "hello (up, 0)", 1},
		{"aucun mot avant", "(up) hello", "(up) hello", 1},
		{"aucun mot avant, seulement de la ponctuation", ", (up) hello", ", (up) hello", 1},
		{"moins de mots que demandé", "hello (up, 3)", "HELLO", 1},
		{"sujet oct", "17 (oct)", "15", 0},
		{"sujet to-hex", "255 (to-hex)", "FF", 0},
		{"sujet to-bin", "5 (to-bin)", "101", 0},
		{"sujet hex négatif", "-1A (hex)", "-26", 0},
		{"hex insensible à la casse", "ff (hex)", "255", 0},
		{"hex avec un mot après", "1E (hex) files", "30 files", 0},
		{"bin avec nombre", "10 11 (bin, 2)", "2 3", 0},
		{"sujet chaînage", "FF (hex) (to-bin)", "11111111", 0},
		{"hex invalide", "ZZ (hex)", "ZZ (hex)", 1},
		{"bin invalide", "3 (bin)", "3 (bin)", 1},
		{"oct invalide", "18 (oct)", "18 (oct)", 1},
		{"décimal invalide", "abc (to-hex)", "abc (to-hex)", 1},
		{"dépassement int64", "FFFFFFFFFFFFFFFFFF (hex)", "FFFFFFFFFFFFFFFFFF (hex)", 1},
		{"un mot valide, un invalide", "FF ZZ (hex, 2)", "255 ZZ (hex, 2)", 1},
		{"exemple du sujet avec trois erreurs", "Add ZZ (hex) and 3 (bin) please (up, -1) .", "Add ZZ (hex) and 3 (bin) please (up, -1) .", 3},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, diagnostics := RuleModificateur{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
	}
}
