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
		{"accent", "élan go", 1, []Token{
			{Mot, "élan", 1, 1, false},
			{Mot, "go", 1, 6, false},
		}},
		{"plusieurs espaces", "élan  go", 1, []Token{
			{Mot, "élan", 1, 1, false},
			{Mot, "go", 1, 7, false},
		}},
		{"plusieurs pièges", " élan  go ", 1, []Token{
			{Mot, "élan", 1, 2, false},
			{Mot, "go", 1, 8, false},
		}},
		{"vide", "", 1, []Token{}},
		{"marqueur", "élan (up)", 1, []Token{
			{Mot, "élan", 1, 1, false},
			{Modif, "(up)", 1, 6, false},
		}},
		{"espace parenthèse", "it was ( almost ) perfect", 1, []Token{
			{Mot, "it", 1, 1, false},
			{Mot, "was", 1, 4, false},
			{Parenthese, "(", 1, 8, false},
			{Mot, "almost", 1, 10, false},
			{Parenthese, ")", 1, 17, false},
			{Mot, "perfect", 1, 19, false},
		}},
		{"mot ponctuation et mot", "BAMM!!yes", 1, []Token{
			{Mot, "BAMM", 1, 1, false},
			{Ponctuation, "!!", 1, 5, false},
			{Mot, "yes", 1, 7, false},
		}},
		{"mot parenthèse et indice", "(hello world) (up, 2)", 1, []Token{
			{Parenthese, "(", 1, 1, false},
			{Mot, "hello", 1, 2, false},
			{Mot, "world", 1, 8, false},
			{Parenthese, ")", 1, 13, false},
			{Modif, "(up, 2)", 1, 15, false},
		}},
		{"mot modificateur et ponctuation", "go(up)!", 1, []Token{
			{Mot, "go", 1, 1, false},
			{Modif, "(up)", 1, 3, false},
			{Ponctuation, "!", 1, 7, false},
		}},
		{"marqueur et ponctuation", "élan (up)!?", 1, []Token{
			{Mot, "élan", 1, 1, false},
			{Modif, "(up)", 1, 6, false},
			{Ponctuation, "!?", 1, 10, false},
		}},
		{"échappement", "Type \\(up) to shout (up)", 1, []Token{
			{Mot, "Type", 1, 1, false},
			{Mot, "(up)", 1, 6, false},
			{Mot, "to", 1, 12, false},
			{Mot, "shout", 1, 15, false},
			{Modif, "(up)", 1, 21, false},
		}},
		{"barre devant un non-modificateur", "\\(hello)", 1, []Token{
			{Mot, "\\", 1, 1, false},
			{Parenthese, "(", 1, 2, false},
			{Mot, "hello", 1, 3, false},
			{Parenthese, ")", 1, 8, false},
		}},
		{"barre isolée", "a \\ b", 1, []Token{
			{Mot, "a", 1, 1, false},
			{Mot, "\\", 1, 3, false},
			{Mot, "b", 1, 5, false},
		}},
		{"balises imbriquées", "[up]this is [low]VERY[/low] important[/up]", 1, []Token{
			{Balise, "[up]", 1, 1, false},
			{Mot, "this", 1, 5, false},
			{Mot, "is", 1, 10, false},
			{Balise, "[low]", 1, 13, false},
			{Mot, "VERY", 1, 18, false},
			{Balise, "[/low]", 1, 22, false},
			{Mot, "important", 1, 29, false},
			{Balise, "[/up]", 1, 38, false},
		}},
		{"crochets qui ne sont pas des balises", "see [some note] here", 1, []Token{
			{Mot, "see", 1, 1, false},
			{Mot, "[some", 1, 5, false},
			{Mot, "note]", 1, 11, false},
			{Mot, "here", 1, 17, false},
		}},
		{"guillemets et contractions", `She said : " don't worry , it's ' fine ' "`, 1, []Token{
			{Mot, "She", 1, 1, false},
			{Mot, "said", 1, 5, false},
			{Ponctuation, ":", 1, 10, false},
			{Guillemet, "\"", 1, 12, false},
			{Mot, "don't", 1, 14, false},
			{Mot, "worry", 1, 20, false},
			{Ponctuation, ",", 1, 26, false},
			{Mot, "it's", 1, 28, false},
			{Guillemet, "'", 1, 33, false},
			{Mot, "fine", 1, 35, false},
			{Guillemet, "'", 1, 40, false},
			{Guillemet, "\"", 1, 42, false},
		}},
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
