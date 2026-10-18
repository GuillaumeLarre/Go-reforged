package main

import (
	"go-reforged/transform"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifierEtCreerExceptions(t *testing.T) {
	test := []struct {
		nom                string
		contenu            string
		attendu            map[string]string
		attenduDiagnostics int
	}{
		{"sujet", "university a\nhouse a\n", map[string]string{"university": "a", "house": "a"}, 0},
		{"majuscules", "HOUSE A", map[string]string{"house": "a"}, 0},
		{"article an", "hour an", map[string]string{"hour": "an"}, 0},
		{"fichier vide", "", map[string]string{}, 0},
		{"lignes vides ignorées", "\n\nhouse a\n\n", map[string]string{"house": "a"}, 0},
		{"espaces et fin Windows", "  hour   an  \r\n", map[string]string{"hour": "an"}, 0},
		{"un seul mot", "house", map[string]string{}, 1},
		{"trois mots", "house a extra", map[string]string{}, 1},
		{"article invalide", "house the", map[string]string{}, 1},
		{"une ligne invalide parmi d'autres", "university a\nhouse the\nhour an", map[string]string{"university": "a", "hour": "an"}, 1},
	}
	for _, te := range test {
		resultat, diagnostics := VerifierEtCreerExceptions(te.contenu)
		if !maps.Equal(resultat, te.attendu) {
			t.Errorf("%s : resultat %v, attendu %v", te.nom, resultat, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
	}
}

func TestTraitementLigne(t *testing.T) {
	regles := []transform.Rule{transform.RuleModificateur{}, transform.RulePonctuation{}}
	test := []struct {
		nom                string
		ligne              string
		attendu            string
		attenduDiagnostics int
	}{
		{"fin Windows conservée", "hello (up)\r\n", "HELLO\r\n", 0},
		{"fin Linux conservée", "hello (up)\n", "HELLO\n", 0},
		{"dernière ligne sans fin", "hello (up)", "HELLO", 0},
		{"ligne vide Windows", "\r\n", "\r\n", 0},
		{"ligne vide Linux", "\n", "\n", 0},
		{"chaîne vide", "", "", 0},
		{"espaces normalisés, fin conservée", "  hello  ,  world  \n", "hello, world\n", 0},
		{"avertissement", "ZZ (hex)\n", "ZZ (hex)\n", 1},
	}
	for _, te := range test {
		resultat, diagnostics := TraitementLigne(te.ligne, 7, regles)
		if resultat != te.attendu {
			t.Errorf("%s : resultat %q, attendu %q", te.nom, resultat, te.attendu)
		}
		if len(diagnostics) != te.attenduDiagnostics {
			t.Errorf("%s : %d diagnostics, attendu %d", te.nom, len(diagnostics), te.attenduDiagnostics)
		}
		for _, d := range diagnostics {
			if d.Ligne != 7 {
				t.Errorf("%s : diagnostic en ligne %d, attendu 7", te.nom, d.Ligne)
			}
		}
	}
}

func TestGarderFichier(t *testing.T) {
	dossier := t.TempDir()
	final := filepath.Join(dossier, "result.txt")
	temp, err := OuvertureFichierTemporaire(final)
	if err != nil {
		t.Errorf("création du temporaire impossible : %v", err)
		return
	}
	temp.WriteString("bonjour\n")
	err = DeciderClotureTemporaire(temp, final, true)
	if err != nil {
		t.Errorf("clôture impossible : %v", err)
		return
	}
	contenu, err := os.ReadFile(final)
	if err != nil {
		t.Errorf("result.txt n'a pas été créé : %v", err)
		return
	}
	if string(contenu) != "bonjour\n" {
		t.Errorf("contenu %q, attendu %q", contenu, "bonjour\n")
	}
}

func TestNePasGarderFichier(t *testing.T) {
	dossier := t.TempDir()
	final := filepath.Join(dossier, "result.txt")
	temp, err := OuvertureFichierTemporaire(final)
	if err != nil {
		t.Errorf("création du temporaire impossible : %v", err)
		return
	}
	temp.WriteString("bonjour\n")
	err = DeciderClotureTemporaire(temp, final, false)
	if err != nil {
		t.Errorf("clôture impossible : %v", err)
		return
	}
	fichiers, _ := os.ReadDir(dossier)
	if len(fichiers) != 0 {
		t.Errorf("%d fichier(s) dans le dossier, attendu 0", len(fichiers))
	}
}

func FuzzTraitement(f *testing.F) {
	f.Add("hello (up)")
	f.Add("ZZ (hex) and 3 (bin) .")
	f.Add("[up]this is [low]VERY[/low] important[/up]")
	f.Add(`She said : " don't worry , it's ' fine ' "`)
	f.Add("I ate an banana and a orange")
	f.Fuzz(func(t *testing.T, texte string) {
		if strings.Contains(texte, "\\") {
			t.Skip()
		}
		regles := []transform.Rule{
			transform.RuleBloc{},
			transform.RuleModificateur{},
			transform.RuleArticle{},
			transform.RulePonctuation{},
			transform.RuleGuillemet{},
			transform.RuleParenthese{},
		}
		une, diagnostics := TraitementLigne(texte, 1, regles)
		for _, d := range diagnostics {
			if d.Gravite == transform.Erreur {
				t.Skip()
			}
		}
		deux, _ := TraitementLigne(une, 1, regles)
		if une != deux {
			t.Errorf("pas idempotent :\nentrée   %q\nune fois %q\ndeux fois %q", texte, une, deux)
		}
	})
}
