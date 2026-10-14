package transform

import (
	"fmt"
	"go-reforged/lexer"
	"strconv"
	"strings"
)

// ==================== //
// == TYPE INTERFACE == //
// ==================== //

const (
	Attention = "warning"
	Erreur    = "error"
)

// Creation du type Interface //
type Rule interface {
	Name() string
	Apply(tokens []lexer.Token) ([]lexer.Token, []Diagnostic)
}

// ================ //
// == DIAGNOSTIC == //
// ================ //

// Creation de la structure diagnostic //
type Diagnostic struct {
	Ligne   int
	Colonne int
	Gravite string
	Message string
}

// Fonction de creation de structure Diagnostic //
func CreerDiagnostic(ligne, colonne int, gravite, message string) Diagnostic {
	return Diagnostic{
		Ligne:   ligne,
		Colonne: colonne,
		Gravite: gravite,
		Message: message,
	}
}

// ======================= //
// == LES MODIFICATEURS == //
// ======================= //

// Strcture vide pour relier la règle à sa méthode //
type RuleModificateur struct{}

func (r RuleModificateur) Name() string {
	return "règle modificateur"
}

//--------------- REFACTORISATION --------------- //

// Fonction de refactorisation pour le traitement de marqueurs //
func TraitementModificateurs(nouvelleListe []lexer.Token, nombre int, fonction func(string) (string, error)) (int, []error) {
	compteur := 0
	listeErreur := []error{}
	for indice := 1; indice <= len(nouvelleListe); indice++ {
		if nouvelleListe[len(nouvelleListe)-indice].Sorte == lexer.Mot && compteur < nombre {
			majuscule, erreur := fonction(nouvelleListe[len(nouvelleListe)-indice].Contenu)
			if erreur == nil {
				compteur += 1
				nouvelleListe[len(nouvelleListe)-indice].Contenu = majuscule
			} else {
				listeErreur = append(listeErreur, erreur)
				compteur += 1
				continue
			}
		} else {
			continue
		}
	}
	return compteur, listeErreur
}

// Fonction de refactorisation pour traiter les modificateurs et ajouter les diagnostic //
func TraiterCas(nouvelleListe []lexer.Token, nombre int, fonction func(string) (string, error), token lexer.Token, diagnostic []Diagnostic) ([]lexer.Token, []Diagnostic) {
	compteur, listeErreur := TraitementModificateurs(nouvelleListe, nombre, fonction)
	if compteur == 0 {
		message := fmt.Sprintf("no word before modifier %q", token.Contenu)
		diag := CreerDiagnostic(token.Ligne, token.Colonne, Attention, message)
		diagnostic = append(diagnostic, diag)
		nouvelleListe = append(nouvelleListe, token)
	} else if compteur < nombre {
		message := fmt.Sprintf("not enough words for modifier %q: applied %d", token.Contenu, compteur)
		diag := CreerDiagnostic(token.Ligne, token.Colonne, Attention, message)
		diagnostic = append(diagnostic, diag)
	}
	for i := 0; i < len(listeErreur); i++ {
		message := listeErreur[i].Error()
		diag := CreerDiagnostic(token.Ligne, token.Colonne, Attention, message)
		diagnostic = append(diagnostic, diag)
	}
	if len(listeErreur) >= 1 {
		nouvelleListe = append(nouvelleListe, token)
	}
	return nouvelleListe, diagnostic
}

//--------------- APPLICATION MODIFICATION --------------- //

// Fonction pour (oct) //
func Oct(mot string) (string, error) {
	nombre, err := strconv.ParseInt(mot, 8, 64)
	if err != nil {
		return "", fmt.Errorf("invalid octal number %q", mot)
	}
	return strconv.FormatInt(nombre, 10), nil
}

// Fonction pour (hex) //
func Hexa(mot string) (string, error) {
	nombre, err := strconv.ParseInt(mot, 16, 64)
	if err != nil {
		return "", fmt.Errorf("invalid hexadecimal number %q", mot)
	}
	return strconv.FormatInt(nombre, 10), nil
}

// Fonction pour (to-hex) //
func ToHex(mot string) (string, error) {
	nombre, err := strconv.ParseInt(mot, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid decimal number %q", mot)
	}
	return strings.ToUpper(strconv.FormatInt(nombre, 16)), nil
}

// Fonction pour (bin) //
func Bin(mot string) (string, error) {
	nombre, err := strconv.ParseInt(mot, 2, 64)
	if err != nil {
		return "", fmt.Errorf("invalid binary number %q", mot)
	}
	return strconv.FormatInt(nombre, 10), nil
}

// Fonction pour (to-bin) //
func ToBin(mot string) (string, error) {
	nombre, err := strconv.ParseInt(mot, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid decimal number %q", mot)
	}
	return strconv.FormatInt(nombre, 2), nil
}

// Fonction pour (up) //
func Majuscule(mot string) (string, error) {
	return strings.ToUpper(mot), nil
}

// Fonction pour (low) //
func Minuscule(mot string) (string, error) {
	return strings.ToLower(mot), nil
}

// Fonction pour (cap) //
func ConversionCap(mot string) (string, error) {
	if mot == "" {
		return "", nil
	}
	motLower := strings.ToLower(mot)
	nouveauMot := ""
	for i, r := range motLower {
		if i == 0 {
			nouveauMot += strings.ToUpper(string(r))
		} else {
			nouveauMot += string(r)
		}
	}
	return nouveauMot, nil
}

// Fonction pour (rev) //
func Rev(mot string) (string, error) {
	listeRune := []rune(mot)
	nouvelleListe := []rune{}
	for i := len(listeRune) - 1; i >= 0; i-- {
		nouvelleListe = append(nouvelleListe, listeRune[i])
	}
	return string(nouvelleListe), nil
}

//--------------- APPLICATION DES REGLES SUR LES TOKENS --------------- //

// Fonction appliquant les règles des modificateurs //
func (r RuleModificateur) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostic := []Diagnostic{}
	for i := 0; i < len(liste); i++ {
		if liste[i].Sorte == lexer.Modif {
			modificateur, nombre, err := lexer.IdentifierModificateurs(liste[i].Contenu)
			if nombre <= 0 || err != nil {
				message := fmt.Sprintf("invalid count in modifier %q", liste[i].Contenu)
				diag := CreerDiagnostic(liste[i].Ligne, liste[i].Colonne, Attention, message)
				diagnostic = append(diagnostic, diag)
				nouvelleListe = append(nouvelleListe, liste[i])
				continue
			}
			switch modificateur {
			case "up":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Majuscule, liste[i], diagnostic)
			case "low":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Minuscule, liste[i], diagnostic)
			case "cap":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, ConversionCap, liste[i], diagnostic)
			case "hex":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Hexa, liste[i], diagnostic)
			case "bin":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Bin, liste[i], diagnostic)
			case "to-hex":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, ToHex, liste[i], diagnostic)
			case "to-bin":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, ToBin, liste[i], diagnostic)
			case "oct":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Oct, liste[i], diagnostic)
			case "rev":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Rev, liste[i], diagnostic)
			default:
				nouvelleListe = append(nouvelleListe, liste[i])
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	return nouvelleListe, diagnostic
}
