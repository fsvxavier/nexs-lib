package strutl

import "unicode"

// CountWords conta o número de palavras no texto.
// Usa a mesma função base de 'Words' mas sem alocar um array,
// sendo mais rápida e eficiente em memória.
//
// Exemplos:
//
//	CountWords("hello world") -> 2
//	CountWords("one,two three") -> 3
//	CountWords("") -> 0
//	CountWords("hello-world") -> 1
func CountWords(str string) int {
	_, count := words(str, true)
	return count
}

// Words retorna as palavras dentro do texto.
// - Números são contados como palavras
// - Se estiverem dentro de uma palavra, estas pontuações não quebram a palavra: ', -, _.
//
// Exemplos:
//
//	Words("hello world") -> ["hello", "world"]
//	Words("O'Neil's micro-service") -> ["O'Neil's", "micro-service"]
//	Words("word1, word2; word3") -> ["word1", "word2", "word3"]
func Words(str string) []string {
	arr, _ := words(str, false)
	return arr
}

const (
	wordRune = iota
	wordPuncRune
	nonWordRune
)

// wordPuncRunes são pontuações que podem estar dentro de palavras: O'Neil, micro-service.
var wordPuncRunes = [...]rune{rune('\''), rune('-'), rune('_')}

func inWordPuncRune(r rune) bool {
	for _, p := range wordPuncRunes {
		if r == p {
			return true
		}
	}
	return false
}

// words é a função base para Words e CountWords. Retorna as palavras
// e a contagem das palavras. Se onlyCount for true, apenas a contagem é retornada,
// nenhum array é criado.
func words(str string, onlyCount bool) ([]string, int) {
	var arr []string
	if !onlyCount {
		arr = make([]string, 0, len(str)/5) // Estimativa inicial conservadora
	}
	prevCat := nonWordRune
	lastStart := -1
	count := 0

	for i, r := range str {
		var cat int
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			cat = wordRune
		case inWordPuncRune(r):
			cat = wordPuncRune
		default:
			cat = nonWordRune
		}

		switch {
		// Inicia palavra
		case cat == wordRune && prevCat != wordRune && lastStart == -1:
			lastStart = i
		// Termina palavra
		case cat == nonWordRune && (prevCat == wordRune || prevCat == wordPuncRune) && lastStart >= 0:
			if !onlyCount {
				arr = append(arr, str[lastStart:i])
			}
			lastStart = -1
			count++
		}

		prevCat = cat
	}

	// Captura a última palavra se o texto terminar com ela
	if lastStart >= 0 {
		if !onlyCount {
			arr = append(arr, str[lastStart:])
		}
		count++
	}
	return arr, count
}
