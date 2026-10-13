package lexer

import (
	"slices"
	"testing"
)

// Fonction de test pour CreationToken //
func TestCreationToken(t *testing.T) {
	test := []struct {
		nom     string
		sorte   string
		contenu string
		ligne   int
		colonne int
		attendu Token
	}{
		{"cas normal", Mot, "hello", 1, 1, Token{Sorte: Mot,
			Contenu: "hello",
			Ligne:   1,
			Colonne: 1}},
	}
	for _, te := range test {
		resultat := CreationToken(te.sorte, te.contenu, te.ligne, te.colonne)
		if resultat != te.attendu {
			t.Errorf("Resultat de %s : %+v, Attendu: %+v", te.nom, resultat, te.attendu)
		}
	}
}

// Fonction de test pour ListerToken //
func TestListerToken(t *testing.T) {
	test := []struct {
		nom      string
		ligne    string
		numligne int
		attendu  []Token
	}{
		{"accent", "élan go", 1, []Token{{
			Sorte:   Mot,
			Contenu: "élan",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Mot,
			Contenu: "go",
			Ligne:   1,
			Colonne: 6,
		}}},
		{"plusieurs espaces", "élan  go", 1, []Token{{
			Sorte:   Mot,
			Contenu: "élan",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Mot,
			Contenu: "go",
			Ligne:   1,
			Colonne: 7,
		}}},
		{"plusieurs pièges", " élan  go ", 1, []Token{{
			Sorte:   Mot,
			Contenu: "élan",
			Ligne:   1,
			Colonne: 2,
		}, {
			Sorte:   Mot,
			Contenu: "go",
			Ligne:   1,
			Colonne: 8,
		}}},
		{"vide", "", 1, []Token{}},
		{"marqueur", "élan (up)", 1, []Token{{
			Sorte:   Mot,
			Contenu: "élan",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Modif,
			Contenu: "(up)",
			Ligne:   1,
			Colonne: 6,
		}}},
		{"espace parenthèse", "it was ( almost ) perfect", 1, []Token{{
			Sorte:   Mot,
			Contenu: "it",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Mot,
			Contenu: "was",
			Ligne:   1,
			Colonne: 4,
		}, {
			Sorte:   Parenthese,
			Contenu: "(",
			Ligne:   1,
			Colonne: 8,
		}, {
			Sorte:   Mot,
			Contenu: "almost",
			Ligne:   1,
			Colonne: 10,
		}, {
			Sorte:   Parenthese,
			Contenu: ")",
			Ligne:   1,
			Colonne: 17,
		}, {
			Sorte:   Mot,
			Contenu: "perfect",
			Ligne:   1,
			Colonne: 19,
		}}},
		{"mot ponctuation et mot", "BAMM!!yes", 1, []Token{{
			Sorte:   Mot,
			Contenu: "BAMM",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Ponctuation,
			Contenu: "!!",
			Ligne:   1,
			Colonne: 5,
		}, {
			Sorte:   Mot,
			Contenu: "yes",
			Ligne:   1,
			Colonne: 7,
		}}},
		{"mot parenthèse et indice", "(hello world) (up, 2)", 1, []Token{{
			Sorte:   Parenthese,
			Contenu: "(",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Mot,
			Contenu: "hello",
			Ligne:   1,
			Colonne: 2,
		}, {
			Sorte:   Mot,
			Contenu: "world",
			Ligne:   1,
			Colonne: 8,
		}, {
			Sorte:   Parenthese,
			Contenu: ")",
			Ligne:   1,
			Colonne: 13,
		}, {
			Sorte:   Modif,
			Contenu: "(up, 2)",
			Ligne:   1,
			Colonne: 15,
		}}},
		{"mot modificateur et ponctuation", "go(up)!", 1, []Token{{
			Sorte:   Mot,
			Contenu: "go",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Modif,
			Contenu: "(up)",
			Ligne:   1,
			Colonne: 3,
		}, {
			Sorte:   Ponctuation,
			Contenu: "!",
			Ligne:   1,
			Colonne: 7,
		}}},
		{"marqueur et ponctuation", "élan (up)!?", 1, []Token{{
			Sorte:   Mot,
			Contenu: "élan",
			Ligne:   1,
			Colonne: 1,
		}, {
			Sorte:   Modif,
			Contenu: "(up)",
			Ligne:   1,
			Colonne: 6,
		}, {
			Sorte:   Ponctuation,
			Contenu: "!?",
			Ligne:   1,
			Colonne: 10,
		}}},
	}
	for _, te := range test {
		resultat := ListerToken(te.ligne, te.numligne)
		if !slices.Equal(resultat, te.attendu) {
			t.Errorf("Resultat de %s : %+v, Attendu: %+v", te.nom, resultat, te.attendu)
		}
	}
}

// Fonction de test pour IdentifierModificateurs //
func TestIdentifierModificateurs(t *testing.T) {
	test := []struct {
		nom                 string
		mot                 string
		attenduModificateur string
		attenduNombre       int
		attenduErreur       bool
	}{
		{"Modificateur seul", "(up)", "up", 1, false},
		{"Modificateur et indice", "(low, 2)", "low", 2, false},
		{"Modificateur avec tiret", "(to-hex)", "to-hex", 1, false},
		{"Modificateur problème indice", "(cap, two)", "cap", 0, true},
		{"Mot", "hello", "", 0, false},
		{"Mot entre parenthèse", "(hello)", "", 0, false},
		{"Modificateur indice négatif", "(up, -2)", "up", -2, false},
		{"Modificateur sans parenthèse de fermeture", "(low", "", 0, false},
		{"Mot avec indice valide", "(hello, 3)", "", 0, false},
	}
	for _, te := range test {
		modificateur, nombre, err := IdentifierModificateurs(te.mot)
		if modificateur != te.attenduModificateur {
			t.Errorf("%s : Resultat modificateur : %s, Attendu : %s", te.nom, modificateur, te.attenduModificateur)
		}
		if nombre != te.attenduNombre {
			t.Errorf("%s : Resultat nombre : %d, Attendu : %d", te.nom, nombre, te.attenduNombre)
		}
		if (err != nil) != te.attenduErreur {
			t.Errorf("%s : Resultat erreur : %v, Attendu : %v", te.nom, err, te.attenduErreur)
		}
	}
}

// Fonctionn de test pour tester AccumulerPonctuation //
func TestAccumulerPonctuation(t *testing.T) {
	test := []struct {
		nom             string
		entree          []rune
		attenduString   string
		attenduPosition int
	}{
		{"groupe ponctuation", []rune("!!!"), "!!!", 3},
		{"groupe espace", []rune("!!! "), "!!!", 3},
		{"ponctuation et lettre", []rune("!he"), "!", 1},
		{"ponctuation mixte", []rune("!?s"), "!?", 2},
	}
	for _, te := range test {
		groupe, position := AccumulerPonctuation(te.entree)
		if groupe != te.attenduString {
			t.Errorf("%s : Resultat groupe : %s, Attendu : %s", te.nom, groupe, te.attenduString)
		}
		if position != te.attenduPosition {
			t.Errorf("%s : Resultat position : %d, Attendu : %d", te.nom, position, te.attenduPosition)
		}
	}
}
