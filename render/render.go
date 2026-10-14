package render

import "go-reforged/lexer"

func RemettreEntexte(liste []lexer.Token) string {
	ligne := ""
	for i := 0; i < len(liste); i++ {
		if i == 0 || liste[i].Colle {
			ligne += liste[i].Contenu
		} else {
			ligne += " "
			ligne += liste[i].Contenu
		}
	}
	return ligne
}
