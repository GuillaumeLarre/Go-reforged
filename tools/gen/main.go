package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// ============= //
// == BIG TXT == //
// ============= //

// Fonction permettant de convertir l'entrée utilisateur en quantité //
func ConversionEntree(entree string) (int, error) {
	unite := ""
	multiplicateur := 1
	if strings.HasSuffix(entree, "KB") {
		unite = "KB"
		multiplicateur = 1024
	}
	if strings.HasSuffix(entree, "MB") {
		unite = "MB"
		multiplicateur = 1048576
	}
	if strings.HasSuffix(entree, "GB") {
		unite = "GB"
		multiplicateur = (1024 * 1024 * 1024)
	}
	suffixe := strings.TrimSuffix(entree, unite)
	numero, err := strconv.Atoi(suffixe)
	if err != nil {
		return 0, err
	}
	numero = numero * multiplicateur
	return numero, nil
}

// Lignes de test utilisées par le générateur //
var lignes = []string{
	"it was the best of times , it was the worst of times (up, 2) .",
	"Simply add 42 (hex) and 10 (bin) and you will see the result is 68 .",
	"There is no greater agony than bearing a untold story inside you .",
	"Punctuation tests are ... kinda boring ,what do you think ?",
	"[up]this is [low]VERY[/low] important[/up] , trust me .",
	`She said : " don't worry , it's ' fine ' " and left .`,
	"I ate an banana and a orange in a hour .",
	"The answer is FF (hex) (to-bin) and the year is MMXXVI (roman) .",
	"the lord of the rings (title, 5) is a long book about élan and noël (rev) .",
	"it was ( almost ) perfect , said 1984 (to-roman) people .",
	"Add ZZ (hex) and 3 (bin) please (up, -1) .",
	"Type \\(up) to shout (up) .",
}

// Fonction main permettant de générer un texte imposant afin de tester les performances //
func main() {
	if len(os.Args) != 3 {
		fmt.Println("Mauvais nombre de paramètre renseigné")
		os.Exit(1)
	}
	taille, err := ConversionEntree(os.Args[1])
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}
	fichier, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}
	tampon := bufio.NewWriter(fichier)
	octet := 0
	for octet <= taille {
		ligne := lignes[rand.Intn(len(lignes))]
		n, err := tampon.WriteString(ligne + "\n")
		octet += n
		if err != nil {
			fmt.Println("Erreur :", err)
			os.Exit(1)
		}
	}
	tampon.Flush()
	fichier.Close()
}
