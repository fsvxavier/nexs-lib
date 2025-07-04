package strutl

import (
	"strings"
	"unicode"
)

// Summary corta a string para um novo comprimento e adiciona o "end" a ela.
// Quebra as palavras apenas por espaços. Veja "unicode.IsSpace" para quais caracteres são
// aceitos como espaços.
//
// Exemplos:
//
//	Summary("Lorem ipsum dolor sit amet", 12, "...") -> "Lorem ipsum..."
//	Summary("Lorem ipsum", 5, "...") -> "Lorem..."
func Summary(str string, length int, end string) string {
	if str == "" || length <= 0 {
		return str
	}
	var runeIndex int
	var i int
	var r rune
	var lastSpaceIndex int
	for i, r = range str {
		switch {
		case r == newLine:
			return str[:i] + end
		case unicode.IsSpace(r):
			lastSpaceIndex = i
		}

		if runeIndex+1 > length {
			break
		}
		runeIndex++
	}

	if runeIndex < length {
		return str
	}

	if lastSpaceIndex == 0 {
		// sem espaço até aqui, então quebra a palavra
		str = str[:i]
	} else {
		// quebra a partir do último espaço visto
		str = str[:lastSpaceIndex]
		str = strings.TrimRight(str, " ")
	}

	return str + end
}
