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
	"runtime/debug"
	"strings"
	"sync"
)

// Constante pour limiter le ramasse miette de GO //
const taillePaquet = 1000

// =================== //
// == ECRITURE SURE == //
// =================== //

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

// ======================== //
// == EXCEPTIONS ARTICLE == //
// ======================== //

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

// =================== //
// == OPTION STRICT == //
// =================== //

// Fonction appliquant l'option --strict : un avertissement devient une erreur //
func AppliquerStrict(diag transform.Diagnostic, strict bool) transform.Diagnostic {
	if strict {
		diag.Gravite = transform.Erreur
	}
	return diag
}

// ================ //
// == TRAITEMENT == //
// ================ //

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

// Fonction lisant un paquet de lignes, et signalant la fin du fichier //
func LirePaquet(lecteur *bufio.Reader, taille int) ([]string, bool) {
	paquet := []string{}
	for len(paquet) < taille {
		ligne, err := lecteur.ReadString('\n')
		if ligne != "" {
			paquet = append(paquet, ligne)
		}
		if err != nil {
			return paquet, true
		}
	}
	return paquet, false
}

// Fonction mettant un diagnostic au format du sujet //
func FormaterDiagnostic(nomFichier string, diag transform.Diagnostic) string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", nomFichier, diag.Ligne, diag.Colonne, diag.Gravite, diag.Message)
}

// ============= //
// == WORKERS == //
// ============= //

// Travail d'un worker : traite une ligne sur n du paquet, à partir de son numéro //
func Workers(numWorker, totalWorker int, paquetLigne []string, numeroPremiereLigne int, listeRegle []transform.Rule, listeResultat []string, diagnostics [][]transform.Diagnostic, attente *sync.WaitGroup) {
	for i := numWorker; i < len(paquetLigne); i += totalWorker {
		resultat, diags := TraitementLigne(paquetLigne[i], numeroPremiereLigne+i, listeRegle)
		diagnostics[i] = diags
		listeResultat[i] = resultat
	}
	attente.Done()
}

// Fonction calculant le nombre de workers réellement utilisés //
func NombreWorkers(demande int, sentenceCase bool) int {
	if demande < 1 || sentenceCase {
		return 1
	}
	return demande
}

// Fonction traitant un paquet de lignes en parallèle, en conservant l'ordre //
func TraiterPaquet(paquet []string, premiereLigne int, regles []transform.Rule, nombreWorkers int) ([]string, [][]transform.Diagnostic) {
	resultats := make([]string, len(paquet))
	diagsParLigne := make([][]transform.Diagnostic, len(paquet))
	var attente sync.WaitGroup
	attente.Add(nombreWorkers)
	for w := 0; w < nombreWorkers; w++ {
		go Workers(w, nombreWorkers, paquet, premiereLigne, regles, resultats, diagsParLigne, &attente)
	}
	attente.Wait()
	return resultats, diagsParLigne
}

// =================== //
// == MON PROGRAMME == //
// =================== //

func main() {
	// Règlage du ramasse-miette de go pour ne pas freiner les workers //
	debug.SetGCPercent(400)
	// Déclaration et lecture des options //
	sentence := flag.Bool("sentence-case", false, "majuscule en début de phrase")
	articles := flag.String("articles", "", "Chemin vers le fichier texte à lire")
	strict := flag.Bool("strict", false, "tout avertissement devient une erreur")
	check := flag.Bool("check", false, "affiche les avertissements et erreurs")
	workers := flag.Int("workers", runtime.NumCPU(), "Nombre de goroutines utilisées pour le traitement")
	flag.Parse()
	// Lecture du fichier d'exception si --article est activé //
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
			fmt.Fprintln(os.Stderr, FormaterDiagnostic(*articles, diag))
		}
		if len(diagsArticles) > 0 {
			os.Exit(2)
		}
	}
	// Vérification du nombre d’arguments : la sortie est facultative avec --check //
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
	// Ouverture de l’entrée : un fichier, ou l’entrée standard avec - //
	var fichier *os.File
	var sortie *os.File
	var err error
	nom := flag.Args()[0]
	var nomAffiche string
	var nomSortie string
	var sortieFichier bool
	var tampon *bufio.Writer
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
	// Préparation de la sortie : un fichier temporaire, ou la sortie standard avec - //
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
		tampon = bufio.NewWriter(sortie)
	}
	if !*check && nomSortie != "-" {
		sortieFichier = true
	}
	// Les règles, dans l’ordre imposé par le sujet //
	regles := []transform.Rule{transform.RuleBloc{}, transform.RuleModificateur{}, transform.RuleArticle{Exceptions: exceptions}, transform.RulePonctuation{}, transform.RuleGuillemet{}, transform.RuleParenthese{}}
	// Ajout de la règle de majuscule, si --sentence-case est donnée //
	if *sentence {
		regles = append(regles, &transform.RuleSentenceCase{DebutPhrase: true})
	}
	// Nombre de workers effectif //
	nombreWorkers := NombreWorkers(*workers, *sentence)
	// Traitement du contenu //
	lecteur := bufio.NewReader(fichier)
	numero := 1
	grave := false
	diagnosticTrouve := false
	finFichier := false
	for !finFichier {
		// Lecture d'un paquet //
		paquet, fin := LirePaquet(lecteur, taillePaquet)
		finFichier = fin
		// Traitement du paquet en parallèle //
		resultats, diagsParLigne := TraiterPaquet(paquet, numero, regles, nombreWorkers)
		// Écriture et diagnostics, dans l'ordre //
		for j := 0; j < len(paquet); j++ {
			if !*check {
				_, errEcriture := tampon.WriteString(resultats[j])
				if errEcriture != nil {
					if sortieFichier {
						DeciderClotureTemporaire(sortie, nomSortie, false)
					}
					fmt.Fprintln(os.Stderr, "Erreur :", errEcriture)
					os.Exit(1)
				}
			}
			for _, diag := range diagsParLigne[j] {
				diag = AppliquerStrict(diag, *strict)
				if diag.Gravite == transform.Erreur {
					grave = true
				}
				diagnosticTrouve = true
				fmt.Fprintln(os.Stderr, FormaterDiagnostic(nomAffiche, diag))
			}
		}
		numero += len(paquet)
	}
	fichier.Close()
	// Erreur de contenu : le temporaire est jeté, code 2 //
	if grave {
		if sortieFichier {
			DeciderClotureTemporaire(sortie, nomSortie, false)
		}
		os.Exit(2)
	}
	// Mode --check : rien n’est écrit, code 3 s’il y a des avertissements //
	if *check {
		if diagnosticTrouve {
			os.Exit(3)
		}
		os.Exit(0)
	}
	// Vider le tampon : envoyer les dernières lignes //
	err = tampon.Flush()
	if err != nil {
		if sortieFichier {
			DeciderClotureTemporaire(sortie, nomSortie, false)
		}
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
	// Le temporaire remplace le fichier final //
	if sortieFichier {
		err = DeciderClotureTemporaire(sortie, nomSortie, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erreur :", err)
			os.Exit(1)
		}
	}
}
