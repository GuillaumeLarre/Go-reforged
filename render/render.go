package render

import (
	"go-reforged/lexer"
	"strings"
)

// ============ //
// == RENDER == //
// ============ //

// Fonction permettant tranformer les Tokens en texte //
func RemettreEntexte(liste []lexer.Token) string {
	var ligne strings.Builder
	for i := 0; i < len(liste); i++ {
		if i == 0 || liste[i].Colle {
			ligne.WriteString(liste[i].Contenu)
		} else {
			ligne.WriteString(" ")
			ligne.WriteString(liste[i].Contenu)
		}
	}
	return ligne.String()
}
