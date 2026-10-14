package main

import (
	"fmt"
	"go-reforged/lexer"
	"go-reforged/render"
	"go-reforged/transform"
	"os"
	"strings"
)

// =================== //
// == MON PROGRAMME == //
// =================== //

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Mauvais nombre de paramètre renseigné. Le programme doit être exécuté avec 'go run . sample.txt result.txt'")
		return
	}
	contenu, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}
	lignes := strings.Split(string(contenu), "\n")
	ligneReformee := []string{}
	regles := []transform.Rule{transform.RuleModificateur{}}
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
	texteTraite := strings.Join(ligneReformee, "\n")
	err = os.WriteFile(os.Args[2], []byte(texteTraite), 0644)
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}
	fmt.Printf("Le fichier %s à bien été crée/modifié", os.Args[2])
}
