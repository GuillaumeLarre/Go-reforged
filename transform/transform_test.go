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
		{"sujet rev avec accent", "noël (rev)", "lëon", 0},
		{"rev avec nombre", "abc def (rev, 2)", "cba fed", 0},
		{"sujet title", "the lord of the rings (title, 5)", "The Lord of the Rings", 0},
		{"title précédé d'autres mots", "I read the lord of the rings (title, 5)", "I read The Lord of the Rings", 0},
		{"title sur un texte en majuscules", "THE LORD OF THE RINGS (title, 5)", "The Lord of the Rings", 0},
		{"title premier mot petit", "of mice and men (title, 4)", "Of Mice and Men", 0},
		{"title mots courts hors liste", "this is for you (title, 4)", "This Is For You", 0},
		{"title accentué", "élan (title)", "Élan", 0},
		{"title moins de mots que demandé", "hello (title, 3)", "Hello", 1},
		{"title sans mot avant", "(title) hello", "(title) hello", 1},
		{"sujet roman", "MMXXVI (roman)", "2026", 0},
		{"sujet to-roman", "1984 (to-roman)", "MCMLXXXIV", 0},
		{"aller-retour en chaîne", "XIV (roman) (to-roman)", "XIV", 0},
		{"sujet roman mal formé répétition", "IIII (roman)", "IIII (roman)", 1},
		{"sujet roman mal formé soustraction", "VX (roman)", "VX (roman)", 1},
		{"roman avec accent", "élan (roman)", "élan (roman)", 1},
		{"to-roman trop grand", "4000 (to-roman)", "4000 (to-roman)", 1},
		{"to-roman zéro", "0 (to-roman)", "0 (to-roman)", 1},
		{"to-roman pas un nombre", "abc (to-roman)", "abc (to-roman)", 1},
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

func TestRuleArticleApply(t *testing.T) {
	test := []struct {
		nom                string
		entree             string
		attendu            string
		attenduDiagnostics int
	}{
		{"sujet", "I ate an banana and a orange in a hour", "I ate a banana and an orange in an hour", 0},
		{"sujet go-reloaded", "bearing a untold story", "bearing an untold story", 0},
		{"majuscule conservée", "A amazing rock", "An amazing rock", 0},
		{"an devient a", "An banana", "A banana", 0},
		{"AN devient A", "AN banana", "A banana", 0},
		{"déjà correct", "an apple and a dog", "an apple and a dog", 0},
		{"h", "a hour and an house", "an hour and an house", 0},
		{"chiffre", "a 8 apples and an 8", "a 8 apples and an 8", 0},
		{"article en fin de ligne", "I want a", "I want a", 0},
		{"ponctuation ignorée", "a , apple", "an , apple", 0},
		{"guillemet ignoré", "a ' apple '", "an ' apple '", 0},
		{"mot qui commence par a", "about a apple", "about an apple", 0},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, diagnostics := RuleArticle{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
	}
}

func TestRuleArticleExceptions(t *testing.T) {
	regle := RuleArticle{Exceptions: map[string]string{"university": "a", "house": "a"}}
	test := []struct {
		nom     string
		entree  string
		attendu string
	}{
		{"sujet", "It is a university next to an house", "It is a university next to a house"},
		{"exception insensible à la casse", "It is an UNIVERSITY", "It is a UNIVERSITY"},
		{"casse de l'article conservée", "An university", "A university"},
		{"mot hors exceptions", "an apple and a orange", "an apple and an orange"},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, _ := regle.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
	}
}

func TestRuleBlocApply(t *testing.T) {
	test := []struct {
		nom                string
		entree             string
		attendu            string
		attenduDiagnostics int
	}{
		{"sujet imbrication", "[up]this is [low]VERY[/low] important[/up]", "THIS IS very IMPORTANT", 0},
		{"sans bloc", "hello world", "hello world", 0},
		{"cap", "[cap]big world[/cap]", "Big World", 0},
		{"le modificateur n'est pas touché", "[low]Hello World (up)[/low]", "hello world (up)", 0},
		{"crochets qui ne sont pas des balises", "see [some note] here", "see [some note] here", 0},
		{"bloc non fermé", "[up]hello", "HELLO", 1},
		{"fermante sans ouvrante", "hello[/up]", "hello", 1},
		{"mauvaise balise", "[up]a [low]b[/up][/low]", "A b", 2},
		{"sujet title", "[title]the lord of the rings[/title]", "The Lord of the Rings", 0},
		{"title précédé d'autres mots", "I read [title]the lord of the rings[/title]", "I read The Lord of the Rings", 0},
		{"title sur un texte en majuscules", "[title]THE LORD OF THE RINGS[/title]", "The Lord of the Rings", 0},
		{"title premier mot petit", "[title]of mice and men[/title]", "Of Mice and Men", 0},
		{"deux blocs title", "[title]a tale[/title] and [title]the end[/title]", "A Tale and The End", 0},
		{"bloc dans un title", "[title]the [up]big[/up] house[/title]", "The BIG House", 0},
		{"title dans un autre bloc", "[up]the [title]of mice[/title][/up]", "THE Of Mice", 0},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, diagnostics := RuleBloc{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
		for _, d := range diagnostics {
			if d.Gravite != Erreur {
				t.Errorf("%s : gravité %q, attendu %q", te.nom, d.Gravite, Erreur)
			}
		}
	}
}

func TestRulePonctuationApply(t *testing.T) {
	test := []struct {
		nom     string
		entree  string
		attendu string
	}{
		{"sujet virgule et groupe", "I was sitting over there ,and then BAMM !!", "I was sitting over there, and then BAMM!!"},
		{"sujet usage", "Punctuation tests are ... kinda boring ,what do you think ?", "Punctuation tests are... kinda boring, what do you think?"},
		{"sujet groupe et majuscule", "I was thinking ... You were right", "I was thinking... You were right"},
		{"déjà correct", "hello, world.", "hello, world."},
		{"deux-points et point-virgule", "first : second ; third", "first: second; third"},
		{"ponctuation en début de ligne", ", hello", ", hello"},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, _ := RulePonctuation{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
	}
}

func TestRuleGuillemetApply(t *testing.T) {
	test := []struct {
		nom                string
		entree             string
		attendu            string
		attenduDiagnostics int
	}{
		{"sujet go-reloaded", "describe me : ' awesome '", "describe me : 'awesome'", 0},
		{"sujet imbrication", `He said : " she told me ' no ' "`, `He said : "she told me 'no'"`, 0},
		{"sujet contraction", "' don't do it '", "'don't do it'", 0},
		{"deux citations", "' hello ' and ' world '", "'hello' and 'world'", 0},
		{"guillemets doubles", `say " hi there " now`, `say "hi there" now`, 0},
		{"guillemets déjà collés", "'hello' and 'world'", "'hello' and 'world'", 0},
		{"paire vide", `say " " now`, `say "" now`, 0},
		{"orphelin", "I said ' hello and left", "I said ' hello and left", 1},
		{"deux orphelins différents", `a " b ' c`, `a " b ' c`, 2},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, diagnostics := RuleGuillemet{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
		for _, d := range diagnostics {
			if d.Gravite != Attention {
				t.Errorf("%s : gravité %q, attendu %q", te.nom, d.Gravite, Attention)
			}
		}
	}
}

func TestRuleParentheseApply(t *testing.T) {
	test := []struct {
		nom                string
		entree             string
		attendu            string
		attenduDiagnostics int
	}{
		{"sujet", "it was ( almost ) perfect", "it was (almost) perfect", 0},
		{"déjà collées", "it was (almost) perfect", "it was (almost) perfect", 0},
		{"imbrication", "( a ( b ) c )", "(a (b) c)", 0},
		{"paire vide", "say ( ) now", "say () now", 0},
		{"ouvrante orpheline", "a ( b", "a ( b", 1},
		{"fermantes orphelines", "a ) b )", "a ) b )", 2},
		{"un modificateur n'est pas une parenthèse", "hello (up) world", "hello (up) world", 0},
	}
	for _, te := range test {
		tokens := lexer.ListerToken(te.entree, 1)
		resultat, diagnostics := RuleParenthese{}.Apply(tokens)
		texte := render.RemettreEntexte(resultat)
		if texte != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, texte, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
	}
}
