package lexer

import (
	"strconv"
	"strings"
	"unicode"
)

type Token struct {
	Sorte   string
	Contenu string
	Ligne   int
	Colonne int
}

// ================ //
// == CONSTANTES == //
// ================ //

// Constantes pour éviter les erreurs d'orthographes //
const (
	Ponctuation = "ponctuation"
	Guillemet   = "guillemet"
	Modif       = "modificateur"
	Balise      = "balise"
	Mot         = "mot"
	Parenthese  = "parenthèse"
)

// =============== //
// == VARIABLES == //
// =============== //

// Variables pour éviter la répétition dans les fonctions //
var ponctuations = map[rune]bool{
	',': true,
	';': true,
	':': true,
	'!': true,
	'?': true,
	'.': true,
	'…': true,
}
var modificateurs = map[string]bool{
	"cap":      true,
	"low":      true,
	"up":       true,
	"hex":      true,
	"bin":      true,
	"rev":      true,
	"title":    true,
	"oct":      true,
	"to-hex":   true,
	"to-bin":   true,
	"to-roman": true,
	"roman":    true,
}

// =================== //
// == MODIFICATEURS == //
// =================== //

// Fonction permettant d'identifier les modificateurs //
func IdentifierModificateurs(mot string) (modificateur string, nombre int, erreur error) {
	i := 0
	if string(mot[i]) == "(" {
		if string(mot[len(mot)-1]) == ")" {
			prefixe := strings.TrimPrefix(mot, "(")
			suffixe := strings.TrimSuffix(prefixe, ")")
			avant, apres, trouve := strings.Cut(suffixe, ", ")
			if !trouve && modificateurs[avant] {
				return avant, 1, nil
			} else if trouve && modificateurs[avant] {
				num, err := strconv.Atoi(apres)
				if err != nil {
					return avant, 0, err
				} else {
					return avant, num, nil
				}
			} else {
				return "", 0, nil
			}
		} else {
			return "", 0, nil
		}
	} else {
		return "", 0, nil
	}
}

// ================= //
// == PONCTUATION == //
// ================= //

// Fonction permettant de vérifier qu'une rune est une ponctuation //
func EstPonctuation(r rune) bool {
	return ponctuations[r]
}

// Fonction permettant d'accumuler les groupes de ponctuation //
func AccumulerPonctuation(liste []rune) (string, int) {
	groupe := ""
	var position int
	for i := 0; i < len(liste); i++ {
		if EstPonctuation(liste[i]) {
			groupe += string(liste[i])
			position += 1
		} else {
			return groupe, position
		}
	}
	return groupe, position
}

// ============ //
// == TOKENS == //
// ============ //

// Fonction permettant de créer une structure Token //
func CreationToken(sorte, contenu string, numligne, colonne int) Token {
	return Token{
		Sorte:   sorte,
		Contenu: contenu,
		Ligne:   numligne,
		Colonne: colonne,
	}
}

// Fonction permettant de transformer une ligne en liste de Token //
func ListerToken(ligne string, numLigne int) []Token {
	caracteres := []rune(ligne)
	liste := []Token{}
	contenu := ""
	var colonne int
	for i := 0; i < len(caracteres); i++ {
		if !unicode.IsSpace(caracteres[i]) {
			if contenu == "" {
				colonne = i + 1
			}
			if EstPonctuation(caracteres[i]) {
				if contenu != "" {
					liste = append(liste, CreationToken(Mot, contenu, numLigne, colonne))
					contenu = ""
				}
				colonne = i + 1
				groupe, position := AccumulerPonctuation(caracteres[i:])
				liste = append(liste, CreationToken(Ponctuation, groupe, numLigne, colonne))
				i += position - 1
			} else if caracteres[i] == ')' {
				if contenu != "" {
					liste = append(liste, CreationToken(Mot, contenu, numLigne, colonne))
					contenu = ""
				}
				liste = append(liste, CreationToken(Parenthese, string(caracteres[i]), numLigne, i+1))
			} else if caracteres[i] != '(' {
				contenu += string(caracteres[i])
			} else {
				if contenu != "" {
					liste = append(liste, CreationToken(Mot, contenu, numLigne, colonne))
					contenu = ""
				}
				colonne = i + 1
				fin := -1
				for j := i; j < len(caracteres); j++ {
					if caracteres[j] == ')' {
						fin = j
						break
					}
				}
				if fin == -1 {
					liste = append(liste, CreationToken(Parenthese, string(caracteres[i]), numLigne, colonne))
				} else {
					modif, _, _ := IdentifierModificateurs(string(caracteres[i : fin+1]))
					if modif != "" {
						liste = append(liste, CreationToken(Modif, string(caracteres[i:fin+1]), numLigne, colonne))
						i = fin
					} else {
						liste = append(liste, CreationToken(Parenthese, string(caracteres[i]), numLigne, colonne))
					}
				}
			}
		} else {
			if contenu != "" {
				liste = append(liste, CreationToken(Mot, contenu, numLigne, colonne))
				contenu = ""
			}
		}
	}
	if contenu != "" {
		liste = append(liste, CreationToken(Mot, contenu, numLigne, colonne))
	}
	return liste
}
