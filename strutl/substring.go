package strutl

import (
	"strings"
	"unicode/utf8"
)

// Substring retorna uma substring de uma string, baseado em índices de runas.
// Manipula caracteres Unicode corretamente, ao contrário de fatias de bytes.
//
// Parâmetros:
//   - s: A string de entrada
//   - start: O índice inicial (inclusivo)
//   - end: O índice final (exclusivo). Se negativo, conta a partir do final da string.
//     Se omitido (0), considera até o final da string.
//
// Exemplos:
//
//	Substring("Hello, 世界", 0, 5) -> "Hello"
//	Substring("Hello, 世界", 7, 9) -> "世界"
//	Substring("Hello, 世界", 7, 0) -> "世界" (end 0 significa até o final)
//	Substring("Hello, 世界", 7, -1) -> "世" (end -1 significa último caractere exclusivo)
func Substring(s string, start int, end int) string {
	if s == "" || start < 0 {
		return ""
	}

	// Conta o número total de runas
	count := utf8.RuneCountInString(s)

	// Ajusta índices
	if start > count {
		return ""
	}

	if end == 0 {
		end = count
	} else if end < 0 {
		end = count + end
	}

	if end < start {
		return ""
	}
	if end > count {
		end = count
	}

	// Caso especial: se estamos retornando a string completa
	if start == 0 && end == count {
		return s
	}

	// Para strings puramente ASCII, podemos usar uma abordagem mais eficiente
	if count == len(s) {
		return s[start:end]
	}

	// Para strings com caracteres Unicode, precisamos iterar
	var startPos, endPos, i int

	for pos := range s {
		if i == start {
			startPos = pos
		}
		if i == end {
			endPos = pos
			break
		}
		i++
	}

	// Se chegamos ao final da string sem encontrar endPos
	if i < end {
		endPos = len(s)
	}

	return s[startPos:endPos]
}

// SubstringAfter retorna a substring após a primeira ocorrência de uma string especificada.
// Se a string não for encontrada, retorna uma string vazia.
//
// Exemplos:
//
//	SubstringAfter("abc-def-ghi", "-") -> "def-ghi"
//	SubstringAfter("abc-def-ghi", ".") -> "" (não encontrado)
//	SubstringAfter("", "-") -> ""
//	SubstringAfter("abc", "") -> "abc"
func SubstringAfter(s, sep string) string {
	if s == "" {
		return s
	}
	if sep == "" {
		return s
	}

	pos := strings.Index(s, sep)
	if pos == -1 {
		return ""
	}

	return s[pos+len(sep):]
}

// SubstringAfterLast retorna a substring após a última ocorrência de uma string especificada.
// Se a string não for encontrada, retorna uma string vazia.
//
// Exemplos:
//
//	SubstringAfterLast("abc-def-ghi", "-") -> "ghi"
//	SubstringAfterLast("abc-def-ghi", ".") -> "" (não encontrado)
//	SubstringAfterLast("", "-") -> ""
//	SubstringAfterLast("abc", "") -> "abc"
func SubstringAfterLast(s, sep string) string {
	if s == "" {
		return s
	}
	if sep == "" {
		return s
	}

	pos := strings.LastIndex(s, sep)
	if pos == -1 {
		return ""
	}

	return s[pos+len(sep):]
}

// SubstringBefore retorna a substring antes da primeira ocorrência de uma string especificada.
// Se a string não for encontrada, retorna a string original.
//
// Exemplos:
//
//	SubstringBefore("abc-def-ghi", "-") -> "abc"
//	SubstringBefore("abc-def-ghi", ".") -> "abc-def-ghi" (não encontrado)
//	SubstringBefore("", "-") -> ""
//	SubstringBefore("abc", "") -> ""
func SubstringBefore(s, sep string) string {
	if s == "" {
		return s
	}
	if sep == "" {
		return ""
	}

	pos := strings.Index(s, sep)
	if pos == -1 {
		return s
	}

	return s[:pos]
}

// SubstringBeforeLast retorna a substring antes da última ocorrência de uma string especificada.
// Se a string não for encontrada, retorna a string original.
//
// Exemplos:
//
//	SubstringBeforeLast("abc-def-ghi", "-") -> "abc-def"
//	SubstringBeforeLast("abc-def-ghi", ".") -> "abc-def-ghi" (não encontrado)
//	SubstringBeforeLast("", "-") -> ""
//	SubstringBeforeLast("abc", "") -> ""
func SubstringBeforeLast(s, sep string) string {
	if s == "" {
		return s
	}
	if sep == "" {
		return ""
	}

	pos := strings.LastIndex(s, sep)
	if pos == -1 {
		return s
	}

	return s[:pos]
}
