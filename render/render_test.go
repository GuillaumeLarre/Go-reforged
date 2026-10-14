package render

import (
	"go-reforged/lexer"
	"testing"
)

func TestRemettreEnTexte(t *testing.T) {
	test := []struct {
		nom     string
		token   []lexer.Token
		attendu string
	}{
		{"Vide", []lexer.Token{}, ""},
		{"un seul mot", []lexer.Token{
			{Contenu: "hello"},
		}, "hello"},
		{"plusieurs mots", []lexer.Token{
			{Contenu: "hello"},
			{Contenu: "world"},
		}, "hello world"},
		{"ponctuation collée", []lexer.Token{
			{Contenu: "there"},
			{Contenu: ",", Colle: true},
			{Contenu: "and"},
		}, "there, and"},
		{"guillemets", []lexer.Token{
			{Contenu: "me"},
			{Contenu: ":", Colle: true},
			{Contenu: "'"},
			{Contenu: "awesome", Colle: true},
			{Contenu: "'", Colle: true},
		}, "me: 'awesome'"},
		{"premier token collé", []lexer.Token{
			{Contenu: "hello", Colle: true},
			{Contenu: "world"},
		}, "hello world"},
	}
	for _, te := range test {
		resultat := RemettreEntexte(te.token)
		if resultat != te.attendu {
			t.Errorf("%s : resultat : %s, Attendu : %s", te.nom, resultat, te.attendu)
		}
	}
}
