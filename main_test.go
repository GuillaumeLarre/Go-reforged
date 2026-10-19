package main

import (
	"bufio"
	"go-reforged/transform"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ================ //
// == TESTS MAIN == //
// ================ //

// Tests pour VerifierEtCreerExceptions //
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

// Tests pour TraitementLigne //
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

// Tests pour vérifier que l'on garde bien le fichier temporaire //
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

// Tests pour vérifier que le fichier temporaire n'est pas gardé //
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

// Test d'ouverture d'un dossier inexistant //
func TestOuvertureDossierInexistant(t *testing.T) {
	dossier := t.TempDir()
	final := filepath.Join(dossier, "inexistant", "result.txt")
	temp, err := OuvertureFichierTemporaire(final)
	if err == nil {
		t.Errorf("une erreur était attendue pour un dossier inexistant")
		temp.Close()
	}
}

// Fuzzing de mon programme //
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

// Benchmark de mon TraitementLigne //
func BenchmarkTraitementLigne(b *testing.B) {
	regles := []transform.Rule{
		transform.RuleBloc{},
		transform.RuleModificateur{},
		transform.RuleArticle{},
		transform.RulePonctuation{},
		transform.RuleGuillemet{},
		transform.RuleParenthese{},
	}
	ligne := `[up]this is[/up] a orange , FF (hex) (to-bin) and " don't ' worry ' " ( really ) .` + "\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TraitementLigne(ligne, 1, regles)
	}
}

// Tests pour LirePaquet //
func TestLirePaquet(t *testing.T) {
	lecteur := bufio.NewReader(strings.NewReader("a\nb\nc"))
	paquet, fin := LirePaquet(lecteur, 2)
	if !slices.Equal(paquet, []string{"a\n", "b\n"}) || fin {
		t.Errorf("premier paquet %q, fin %v", paquet, fin)
	}
	paquet, fin = LirePaquet(lecteur, 2)
	if !slices.Equal(paquet, []string{"c"}) || !fin {
		t.Errorf("second paquet %q, fin %v", paquet, fin)
	}
}

// Vérifie que le résultat d'un paquet est identique avec 1 et 4 workers //
func TestTraiterPaquetIdentique(t *testing.T) {
	regles := []transform.Rule{transform.RuleModificateur{}, transform.RulePonctuation{}}
	paquet := []string{"hello (up)\n", "a , b\n", "FF (hex)\n", "c d (cap, 2)\n", "ZZ (hex)\n"}
	un, _ := TraiterPaquet(paquet, 1, regles, 1)
	quatre, _ := TraiterPaquet(paquet, 1, regles, 4)
	if !slices.Equal(un, quatre) {
		t.Errorf("résultats différents :\n1 worker  %q\n4 workers %q", un, quatre)
	}
}

// Vérifie le nombre de workers réellement utilisé //
func TestNombreWorkers(t *testing.T) {
	test := []struct {
		nom          string
		demande      int
		sentenceCase bool
		attendu      int
	}{
		{"valeur normale", 4, false, 4},
		{"zéro ramené à un", 0, false, 1},
		{"négatif ramené à un", -3, false, 1},
		{"sentence-case force un", 4, true, 1},
	}
	for _, te := range test {
		resultat := NombreWorkers(te.demande, te.sentenceCase)
		if resultat != te.attendu {
			t.Errorf("%s : resultat %d, attendu %d", te.nom, resultat, te.attendu)
		}
	}
}

// Vérifie le format exact des diagnostics, sur l'exemple de la partie 5.3 du sujet //
func TestFormaterDiagnostic(t *testing.T) {
	diag := transform.CreerDiagnostic(1, 8, transform.Attention, `invalid hexadecimal number "ZZ"`)
	resultat := FormaterDiagnostic("sample.txt", diag)
	attendu := `sample.txt:1:8: warning: invalid hexadecimal number "ZZ"`
	if resultat != attendu {
		t.Errorf("resultat %q, attendu %q", resultat, attendu)
	}
}

// Vérifie l'option --strict //
func TestAppliquerStrict(t *testing.T) {
	avertissement := transform.CreerDiagnostic(1, 1, transform.Attention, "message")
	if AppliquerStrict(avertissement, true).Gravite != transform.Erreur {
		t.Errorf("avec --strict, un avertissement doit devenir une erreur")
	}
	if AppliquerStrict(avertissement, false).Gravite != transform.Attention {
		t.Errorf("sans --strict, un avertissement doit rester un avertissement")
	}
}
