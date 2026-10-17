package main

import (
	"flag"
	"fmt"
	"go-reforged/lexer"
	"go-reforged/render"
	"go-reforged/transform"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// =================== //
// == MON PROGRAMME == //
// =================== //

func EcrireFichier(chemin string, texte string) error {
	dossier := filepath.Dir(chemin)
	temp, err := os.CreateTemp(dossier, "modele-*")
	if err != nil {
		return err
	}
	_, err = temp.WriteString(texte)
	if err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}
	err = temp.Close()
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	err = os.Rename(temp.Name(), chemin)
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	return nil
}

func VerifierEtCreerExceptions(fichier string) (map[string]string, []transform.Diagnostic) {
	exceptions := map[string]string{}
	listeDiagnostic := []transform.Diagnostic{}
	lignes := strings.Split(fichier, "\n")
	for i := 0; i < len(lignes); i++ {
		if strings.TrimSpace(lignes[i]) == "" {
			continue
		}
		mot := strings.Fields(lignes[i])
		if len(mot) != 2 {
			diag := transform.CreerDiagnostic(i+1, 1, transform.Erreur, "malformed line, expected 'word article'")
			listeDiagnostic = append(listeDiagnostic, diag)
			continue
		}
		article := strings.ToLower(mot[1])
		if article != "a" && article != "an" {
			message := fmt.Sprintf("invalid article %q", mot[1])
			diag := transform.CreerDiagnostic(i+1, 1, transform.Erreur, message)
			listeDiagnostic = append(listeDiagnostic, diag)
			continue
		}
		exceptions[strings.ToLower(mot[0])] = article
	}
	return exceptions, listeDiagnostic
}

func main() {
	sentence := flag.Bool("sentence-case", false, "majuscule en début de phrase")
	_ = sentence
	articles := flag.String("articles", "", "Chemin vers le fichier texte à lire")
	strict := flag.Bool("strict", false, "tout avertissement devient une erreur")
	check := flag.Bool("check", false, "affiche les avertissements et erreurs")
	workers := flag.Int("workers", runtime.NumCPU(), "Nombre de goroutines utilisées pour le traitement")
	_ = workers
	flag.Parse()
	exceptions := map[string]string{}
	if *articles != "" {
		contenuExceptions, err := os.ReadFile(*articles)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erreur :", err)
			os.Exit(1)
		}
		diagsArticles := []transform.Diagnostic{}
		exceptions, diagsArticles = VerifierEtCreerExceptions(string(contenuExceptions))
		for k := 0; k < len(diagsArticles); k++ {
			diag := diagsArticles[k]
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", *articles, diag.Ligne, diag.Colonne, diag.Gravite, diag.Message)
		}
		if len(diagsArticles) > 0 {
			os.Exit(2)
		}
	}
	if !*check {
		if len(flag.Args()) != 2 {
			fmt.Fprintln(os.Stderr, "Mauvais nombre de paramètre renseigné. Le programme doit être exécuté avec 'go run . sample.txt result.txt'")
			os.Exit(1)
		}
	} else {
		if len(flag.Args()) < 1 {
			fmt.Fprintln(os.Stderr, "Mauvais nombre de paramètre renseigné. Le programme doit être exécuté avec une entrée au minimum")
			os.Exit(1)
		} else if len(flag.Args()) > 2 {
			fmt.Fprintln(os.Stderr, "Mauvais nombre de paramètre renseigné. Le programme doit être exécuté avec deux entrées au maximum")
			os.Exit(1)
		}
	}
	contenu, err := os.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
	lignes := strings.Split(string(contenu), "\n")
	ligneReformee := []string{}
	regles := []transform.Rule{transform.RuleBloc{}, transform.RuleModificateur{}, transform.RuleArticle{Exceptions: exceptions}, transform.RulePonctuation{}, transform.RuleGuillemet{}, transform.RuleParenthese{}}
	listeDiagnostic := []transform.Diagnostic{}
	for i := 0; i < len(lignes); i++ {
		listeToken := lexer.ListerToken(lignes[i], i+1)
		for j := 0; j < len(regles); j++ {
			t, diag := regles[j].Apply(listeToken)
			listeDiagnostic = append(listeDiagnostic, diag...)
			listeToken = t
		}
		ligneReformee = append(ligneReformee, render.RemettreEntexte(listeToken))
	}
	grave := false
	for k := 0; k < len(listeDiagnostic); k++ {
		diag := listeDiagnostic[k]
		if *strict {
			diag.Gravite = transform.Erreur
		}
		if diag.Gravite == transform.Erreur {
			grave = true
		}
		fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", flag.Args()[0], diag.Ligne, diag.Colonne, diag.Gravite, diag.Message)
	}
	if grave {
		os.Exit(2)
	}
	if *check {
		if len(listeDiagnostic) > 0 {
			os.Exit(3)
		} else {
			os.Exit(0)
		}
	}
	texteTraite := strings.Join(ligneReformee, "\n")
	err = EcrireFichier(flag.Args()[1], texteTraite)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
}
