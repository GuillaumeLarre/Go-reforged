package transform

import (
	"fmt"
	"go-reforged/lexer"
	"go-reforged/render"
	"strconv"
	"strings"
	"unicode"
)

// ==================== //
// == TYPE INTERFACE == //
// ==================== //

const (
	Attention = "warning"
	Erreur    = "error"
)

var petitsmots = map[string]bool{
	"a":   true,
	"an":  true,
	"the": true,
	"of":  true,
	"and": true,
	"or":  true,
	"in":  true,
	"on":  true,
	"at":  true,
	"to":  true,
}

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

// ========================== //
// == REGLES MODIFICATEURS == //
// ========================== //

// Strcture vide pour relier la règle à sa méthode //
type RuleModificateur struct{}

func (r RuleModificateur) Name() string {
	return "règle modificateurs"
}

//--------------- REFACTORISATION --------------- //

// Fonction de refactorisation pour le traitement de marqueurs //
func TraitementModificateurs(nouvelleListe []lexer.Token, nombre int, fonction func(string) (string, error)) (int, []error) {
	compteur := 0
	listeErreur := []error{}
	for indice := 1; indice <= len(nouvelleListe); indice++ {
		if nouvelleListe[len(nouvelleListe)-indice].Sorte == lexer.Mot && compteur < nombre {
			resultat, erreur := fonction(nouvelleListe[len(nouvelleListe)-indice].Contenu)
			if erreur == nil {
				compteur += 1
				nouvelleListe[len(nouvelleListe)-indice].Contenu = resultat
				nouvelleListe[len(nouvelleListe)-indice].EstModifie = true
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

// Fonction pour (title) //
func Title(mot string) (string, error) {
	if petitsmots[strings.ToLower(mot)] {
		return strings.ToLower(mot), nil
	} else {
		return ConversionCap(mot)
	}
}

// Fonction pour mettre en majuscule le premier mot de la liste //
func MajusculePremierMotTitle(liste []lexer.Token, nombre int) {
	compteur := 0
	indicePremier := -1
	for i := len(liste) - 1; i >= 0 && compteur < nombre; i-- {
		if liste[i].Sorte == lexer.Mot {
			compteur += 1
			indicePremier = i
		}
	}
	if indicePremier != -1 {
		liste[indicePremier].Contenu, _ = ConversionCap(liste[indicePremier].Contenu)
	}
}

// Fonction pour convertir un nombre décimal en chiffre romain //
func ToRoman(mot string) (string, error) {
	listeRomain := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	listeNombre := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	nombre, err := strconv.ParseInt(mot, 10, 64)
	romain := ""
	if err != nil {
		return "", fmt.Errorf("invalid decimal number %q", mot)
	}
	if nombre < 1 || nombre > 3999 {
		return "", fmt.Errorf("out of range decimal number %q", mot)
	}
	for i := 0; i < len(listeNombre); i++ {
		for nombre-int64(listeNombre[i]) >= 0 {
			nombre -= int64(listeNombre[i])
			romain += listeRomain[i]
		}
	}
	return romain, nil
}

// Fonction pour convertir un nombre romain en décimal //
func Roman(mot string) (string, error) {
	mots := []rune(mot)
	var nombre int
	romain := map[rune]int{
		'M': 1000,
		'D': 500,
		'C': 100,
		'L': 50,
		'X': 10,
		'V': 5,
		'I': 1,
	}
	i := 0
	for i < len(mots) {
		if i+1 < len(mots) && romain[mots[i+1]] > romain[mots[i]] {
			nombre += (romain[mots[i+1]] - romain[mots[i]])
			i += 2
		} else {
			nombre += romain[mots[i]]
			i++
		}
	}
	nombrestr := strconv.Itoa(nombre)
	resultat, _ := ToRoman(nombrestr)
	if resultat != mot {
		return "", fmt.Errorf("invalid roman number %q", mot)
	} else {
		return nombrestr, nil
	}
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
			case "title":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Title, liste[i], diagnostic)
				MajusculePremierMotTitle(nouvelleListe, nombre)
			case "roman":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, Roman, liste[i], diagnostic)
			case "to-roman":
				nouvelleListe, diagnostic = TraiterCas(nouvelleListe, nombre, ToRoman, liste[i], diagnostic)
			default:
				nouvelleListe = append(nouvelleListe, liste[i])
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	return nouvelleListe, diagnostic
}

// ===================== //
// == REGLES ARTICLES == //
// ===================== //

type RuleArticle struct {
	Exceptions map[string]string
}

func (r RuleArticle) Name() string {
	return "Règle articles"
}

func IdentifierArticle(mot string) bool {
	switch mot {
	case "a", "A", "an", "An", "AN":
		return true
	default:
		return false
	}
}

func TrouverMotSuivant(liste []lexer.Token) int {
	i := 0
	position := -1
	for i < len(liste) {
		if liste[i].Sorte == lexer.Mot {
			position = i
			break
		} else {
			i++
		}
	}
	return position
}

func EstUneVoyelle(mot string) bool {
	voyelles := map[string]bool{
		"a": true,
		"e": true,
		"i": true,
		"o": true,
		"u": true,
		"h": true,
		"A": true,
		"E": true,
		"I": true,
		"O": true,
		"U": true,
		"H": true,
	}
	if voyelles[string(mot[0])] {
		return true
	} else {
		return false
	}
}

func AppliquerRegleArticle(liste []lexer.Token, indice int, exceptions map[string]string) lexer.Token {
	token := liste[indice]
	position := TrouverMotSuivant(liste[indice+1:])
	if position != -1 {
		listetokenEntre := liste[indice : position+(indice+1)]
		for i := 0; i < len(listetokenEntre); i++ {
			if listetokenEntre[i].Sorte == lexer.Modif {
				return token
			}
		}
		motSuivant := liste[position+(indice+1)].Contenu
		article, trouve := exceptions[strings.ToLower(motSuivant)]
		premiere := []rune((liste[position+(indice+1)].Contenu))[0]
		estUnNombre := unicode.IsDigit(premiere)
		if estUnNombre {
			return token
		}
		if trouve {
			if unicode.IsUpper(rune(token.Contenu[0])) {
				token.Contenu, _ = ConversionCap(article)
			} else {
				token.Contenu = article
			}
			return token
		}
		if EstUneVoyelle(liste[position+(indice+1)].Contenu) {
			switch token.Contenu {
			case "a":
				token.Contenu = "an"
			case "A":
				token.Contenu = "An"
			}
		} else {
			switch token.Contenu {
			case "an":
				token.Contenu = "a"
			case "An", "AN":
				token.Contenu = "A"
			}
		}
	}
	return token
}

func (r RuleArticle) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostic := []Diagnostic{}
	for i := 0; i < len(liste); i++ {
		if liste[i].Sorte == lexer.Mot {
			trouve := IdentifierArticle(liste[i].Contenu)
			if trouve {
				token := AppliquerRegleArticle(liste, i, r.Exceptions)
				nouvelleListe = append(nouvelleListe, token)
			} else {
				nouvelleListe = append(nouvelleListe, liste[i])
				continue
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	return nouvelleListe, diagnostic
}

// ================== //
// == REGLES BLOCS == //
// ================== //

type RuleBloc struct {
	Exceptions map[string]string
}

func (r RuleBloc) Name() string {
	return "Règle blocs"
}

func IdentifierType(mot string) (string, bool) {
	prefixe := strings.TrimPrefix(mot, "[")
	suffixe := strings.TrimSuffix(prefixe, "]")
	trouve := strings.HasPrefix(suffixe, "/")
	if trouve {
		sansFermant := strings.TrimPrefix(suffixe, "/")
		return sansFermant, true
	} else {
		return suffixe, false
	}
}

func (r RuleBloc) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostic := []Diagnostic{}
	pile := []lexer.Token{}
	premierMot := false
	for i := 0; i < len(liste); i++ {
		if liste[i].Sorte == lexer.Balise {
			balise, fermante := IdentifierType(liste[i].Contenu)
			if !fermante {
				if balise == "title" {
					premierMot = true
				}
				pile = append(pile, liste[i])
			} else if len(pile) == 0 {
				message := fmt.Sprintf("closing tag %q without opening tag", liste[i].Contenu)
				diag := CreerDiagnostic(liste[i].Ligne, liste[i].Colonne, Erreur, message)
				diagnostic = append(diagnostic, diag)
				continue
			} else {
				ancienneBalise, _ := IdentifierType(pile[len(pile)-1].Contenu)
				if balise == ancienneBalise {
					pile = pile[:len(pile)-1]
				} else {
					message := fmt.Sprintf("closing tag %q does not match %q", liste[i].Contenu, pile[len(pile)-1].Contenu)
					diag := CreerDiagnostic(liste[i].Ligne, liste[i].Colonne, Erreur, message)
					diagnostic = append(diagnostic, diag)
				}
			}
		} else if liste[i].Sorte == lexer.Mot {
			if len(pile) == 0 {
				nouvelleListe = append(nouvelleListe, liste[i])
				continue
			}
			ancienneBalise, _ := IdentifierType(pile[len(pile)-1].Contenu)
			switch ancienneBalise {
			case "up":
				token := liste[i]
				token.EstModifie = true
				majuscule, _ := Majuscule(token.Contenu)
				token.Contenu = majuscule
				nouvelleListe = append(nouvelleListe, token)
			case "low":
				token := liste[i]
				token.EstModifie = true
				minuscule, _ := Minuscule(token.Contenu)
				token.Contenu = minuscule
				nouvelleListe = append(nouvelleListe, token)
			case "title":
				token := liste[i]
				if premierMot {
					maj, _ := ConversionCap(token.Contenu)
					token.Contenu = maj
					token.EstModifie = true
					nouvelleListe = append(nouvelleListe, token)
					premierMot = false
				} else {
					title, _ := Title(token.Contenu)
					token.EstModifie = true
					token.Contenu = title
					nouvelleListe = append(nouvelleListe, token)
				}
			case "cap":
				token := liste[i]
				cap, _ := ConversionCap(token.Contenu)
				token.EstModifie = true
				token.Contenu = cap
				nouvelleListe = append(nouvelleListe, token)
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	for _, ouvrante := range pile {
		message := fmt.Sprintf("unclosed block %q", ouvrante.Contenu)
		diag := CreerDiagnostic(ouvrante.Ligne, ouvrante.Colonne, Erreur, message)
		diagnostic = append(diagnostic, diag)
	}
	return nouvelleListe, diagnostic
}

// ========================= //
// == REGLES PONCTUATIONS == //
// ========================= //

type RulePonctuation struct{}

func (r RulePonctuation) Name() string {
	return "Règle ponctuations"
}

func (r RulePonctuation) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostics := []Diagnostic{}
	for i := 0; i < len(liste); i++ {
		if liste[i].Sorte == lexer.Ponctuation {
			token := liste[i]
			token.Colle = true
			nouvelleListe = append(nouvelleListe, token)
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	return nouvelleListe, diagnostics
}

// ======================= //
// == REGLES GUILLEMETS == //
// ======================= //

type RuleGuillemet struct{}

func (r RuleGuillemet) Name() string {
	return "Règle guillemets"
}

func (r RuleGuillemet) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostics := []Diagnostic{}
	pile := []int{}
	for i := 0; i < len(liste); i++ {
		token := liste[i]
		if liste[i].Sorte == lexer.Guillemet {
			if len(pile) == 0 {
				pile = append(pile, i)
				nouvelleListe = append(nouvelleListe, token)
			} else if token.Contenu == nouvelleListe[pile[len(pile)-1]].Contenu {
				if pile[len(pile)-1]+1 < len(nouvelleListe) {
					nouvelleListe[pile[len(pile)-1]+1].Colle = true
				}
				token.Colle = true
				pile = pile[:len(pile)-1]
				nouvelleListe = append(nouvelleListe, token)
			} else {
				pile = append(pile, i)
				nouvelleListe = append(nouvelleListe, token)
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	for _, ouvrante := range pile {
		message := fmt.Sprintf("unmatched quote %q", nouvelleListe[ouvrante].Contenu)
		diag := CreerDiagnostic(nouvelleListe[ouvrante].Ligne, nouvelleListe[ouvrante].Colonne, Attention, message)
		diagnostics = append(diagnostics, diag)
	}
	return nouvelleListe, diagnostics
}

// ======================== //
// == REGLES PARENTHESES == //
// ======================== //

type RuleParenthese struct{}

func (r RuleParenthese) Name() string {
	return "Règle parenthèses"
}

func (r RuleParenthese) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostics := []Diagnostic{}
	pile := []int{}
	for i := 0; i < len(liste); i++ {
		token := liste[i]
		if liste[i].Sorte == lexer.Parenthese {
			if token.Contenu == "(" {
				pile = append(pile, i)
				nouvelleListe = append(nouvelleListe, token)
			} else if token.Contenu == ")" && len(pile) > 0 {
				coller := true
				sommet := pile[len(pile)-1]
				if sommet+1 < len(nouvelleListe) {
					verification := render.RemettreEntexte(nouvelleListe[sommet+1:])
					verification = "(" + verification + ")"
					estmarqueur, _, _ := lexer.IdentifierModificateurs(verification)
					if estmarqueur != "" {
						coller = false
					} else {
						nouvelleListe[sommet+1].Colle = true
					}
				}
				token.Colle = coller
				pile = pile[:len(pile)-1]
				nouvelleListe = append(nouvelleListe, token)
			} else {
				message := fmt.Sprintf("unmatched parenthesis %q", token.Contenu)
				diag := CreerDiagnostic(token.Ligne, token.Colonne, Attention, message)
				diagnostics = append(diagnostics, diag)
				nouvelleListe = append(nouvelleListe, token)
			}
		} else {
			nouvelleListe = append(nouvelleListe, liste[i])
		}
	}
	for _, ouvrante := range pile {
		message := fmt.Sprintf("unmatched parenthesis %q", nouvelleListe[ouvrante].Contenu)
		diag := CreerDiagnostic(nouvelleListe[ouvrante].Ligne, nouvelleListe[ouvrante].Colonne, Attention, message)
		diagnostics = append(diagnostics, diag)
	}
	return nouvelleListe, diagnostics
}

// ========================== //
// == REGLES SENTENCE CASE == //
// ========================== //

type RuleSentenceCase struct{ DebutPhrase bool }

func (r RuleSentenceCase) Name() string {
	return "Règle sentence-case"
}

func (r *RuleSentenceCase) Apply(liste []lexer.Token) ([]lexer.Token, []Diagnostic) {
	nouvelleListe := []lexer.Token{}
	diagnostics := []Diagnostic{}
	for i := 0; i < len(liste); i++ {
		token := liste[i]
		if token.Sorte == lexer.Mot {
			if r.DebutPhrase {
				if token.EstModifie {
					nouvelleListe = append(nouvelleListe, token)
					r.DebutPhrase = false
				} else {
					token.Contenu = MajPremiereLettre(token.Contenu)
					nouvelleListe = append(nouvelleListe, token)
					r.DebutPhrase = false
				}
			} else {
				nouvelleListe = append(nouvelleListe, token)
			}
		} else if token.Sorte == lexer.Ponctuation {
			if token.Contenu == "." || token.Contenu == "!" || token.Contenu == "?" {
				r.DebutPhrase = true
				nouvelleListe = append(nouvelleListe, token)
			} else {
				nouvelleListe = append(nouvelleListe, token)
			}
		} else {
			nouvelleListe = append(nouvelleListe, token)
		}
	}
	return nouvelleListe, diagnostics
}

func MajPremiereLettre(mot string) string {
	if mot == "" {
		return ""
	}
	nouveauMot := ""
	for i, r := range mot {
		if i == 0 {
			nouveauMot += strings.ToUpper(string(r))
		} else {
			nouveauMot += string(r)
		}
	}
	return nouveauMot
}
