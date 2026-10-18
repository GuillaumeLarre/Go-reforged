package main

import (
	"bufio"
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

// Fonction permettant d'ouvrir un fichier temporaire qui sert de garde fou //
func OuvertureFichierTemporaire(chemin string) (*os.File, error) {
	dossier := filepath.Dir(chemin)
	temp, err := os.CreateTemp(dossier, "modele-*")
	if err != nil {
		return nil, err
	}
	return temp, nil
}

// Fonction permettant de décider ce qu'il faut faire du fichier temporaire //
func DeciderClotureTemporaire(temp *os.File, final string, garder bool) error {
	err := temp.Close()
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	if !garder {
		os.Remove(temp.Name())
		return nil
	}
	err = os.Rename(temp.Name(), final)
	if err != nil {
		os.Remove(temp.Name())
		return err
	}
	return nil
}

// Fonction de vérification et de création des exceptions (comme pour article) //
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

// Fonction de refactorisation de la boucle de traitement des lignes dans le main //
func TraitementLigne(texte string, numero int, regles []transform.Rule) (string, []transform.Diagnostic) {
	contenu := texte
	fin := ""
	if strings.HasSuffix(texte, "\n") {
		contenu = strings.TrimSuffix(contenu, "\n")
		fin = "\n"
		if strings.HasSuffix(contenu, "\r") {
			contenu = strings.TrimSuffix(contenu, "\r")
			fin = "\r\n"
		}
	}
	listeDiagnostic := []transform.Diagnostic{}
	listeToken := lexer.ListerToken(contenu, numero)
	for j := 0; j < len(regles); j++ {
		t, diag := regles[j].Apply(listeToken)
		listeDiagnostic = append(listeDiagnostic, diag...)
		listeToken = t
	}
	ligneCorrigee := render.RemettreEntexte(listeToken)
	ligneCorrigee += fin
	return ligneCorrigee, listeDiagnostic
}

// =================== //
// == MON PROGRAMME == //
// =================== //

func main() {
	// Les options //
	sentence := flag.Bool("sentence-case", false, "majuscule en début de phrase")
	articles := flag.String("articles", "", "Chemin vers le fichier texte à lire")
	strict := flag.Bool("strict", false, "tout avertissement devient une erreur")
	check := flag.Bool("check", false, "affiche les avertissements et erreurs")
	workers := flag.Int("workers", runtime.NumCPU(), "Nombre de goroutines utilisées pour le traitement")
	_ = workers
	flag.Parse()
	// Gestion option article //
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
	// Gestion check et nombre d'arguments //
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
	// Gestion des entrées '-' et initialisation lecteur //
	var fichier *os.File
	var sortie *os.File
	var err error
	nom := flag.Args()[0]
	var nomAffiche string
	var nomSortie string
	var sortieFichier bool
	if nom == "-" {
		nomAffiche = "<stdin>"
		fichier = os.Stdin
	} else {
		nomAffiche = nom
		fichier, err = os.Open(nom)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
	if !*check {
		nomSortie = flag.Args()[1]
		if nomSortie == "-" {
			sortie = os.Stdout
		} else {
			sortie, err = OuvertureFichierTemporaire(nomSortie)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erreur :", err)
			os.Exit(1)
		}
	}
	if !*check && nomSortie != "-" {
		sortieFichier = true
	}
	regles := []transform.Rule{transform.RuleBloc{}, transform.RuleModificateur{}, transform.RuleArticle{Exceptions: exceptions}, transform.RulePonctuation{}, transform.RuleGuillemet{}, transform.RuleParenthese{}}
	listeDiagnostic := []transform.Diagnostic{}
	// Gestion de setence-case //
	if *sentence {
		regles = append(regles, &transform.RuleSentenceCase{DebutPhrase: true})
	}
	// Traitement du contenu //
	lecteur := bufio.NewReader(fichier)
	numero := 1
	for {
		ligne, err := lecteur.ReadString('\n')
		if ligne != "" {
			ligneCorrigee, diag := TraitementLigne(ligne, numero, regles)
			if !*check {
				_, errEcriture := sortie.WriteString(ligneCorrigee)
				if errEcriture != nil {
					if sortieFichier {
						DeciderClotureTemporaire(sortie, nomSortie, false)
					}
					fmt.Fprintln(os.Stderr, "Erreur :", errEcriture)
					os.Exit(1)
				}
			}
			listeDiagnostic = append(listeDiagnostic, diag...)
			numero++
		}
		if err != nil {
			break
		}
	}
	fichier.Close()
	// Gestion des diagnostics //
	grave := false
	for k := 0; k < len(listeDiagnostic); k++ {
		diag := listeDiagnostic[k]
		if *strict {
			diag.Gravite = transform.Erreur
		}
		if diag.Gravite == transform.Erreur {
			grave = true
		}
		fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", nomAffiche, diag.Ligne, diag.Colonne, diag.Gravite, diag.Message)
	}
	if grave {
		if sortieFichier {
			err = DeciderClotureTemporaire(sortie, flag.Args()[1], false)
		}
		os.Exit(2)
	}
	// Gestion option check //
	if *check {
		if len(listeDiagnostic) > 0 {
			os.Exit(3)
		} else {
			os.Exit(0)
		}
	}
	// Gestion de la sortie du fichier //
	if sortieFichier {
		err = DeciderClotureTemporaire(sortie, flag.Args()[1], true)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erreur :", err)
			os.Exit(1)
		}
	}
}
