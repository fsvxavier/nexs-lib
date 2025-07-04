package strutl

import (
	"strings"
	"unicode"
)

// ReplaceAll substitui todas as ocorrências de uma substring por outra.
// É um wrapper para strings.ReplaceAll com validações adicionais.
//
// Exemplos:
//
//	ReplaceAll("hello world", "world", "golang") -> "hello golang"
//	ReplaceAll("", "x", "y") -> ""
//	ReplaceAll("hello", "", "x") -> "hello"
func ReplaceAll(s, old, new string) string {
	if s == "" {
		return s
	}
	if old == "" {
		return s
	}
	return strings.ReplaceAll(s, old, new)
}

// ReplaceFirst substitui apenas a primeira ocorrência de uma substring por outra.
//
// Exemplos:
//
//	ReplaceFirst("hello hello", "hello", "hi") -> "hi hello"
//	ReplaceFirst("", "x", "y") -> ""
//	ReplaceFirst("hello", "", "x") -> "hello"
func ReplaceFirst(s, old, new string) string {
	if s == "" {
		return s
	}
	if old == "" {
		return s
	}

	pos := strings.Index(s, old)
	if pos == -1 {
		return s
	}

	return s[:pos] + new + s[pos+len(old):]
}

// ReplaceLast substitui apenas a última ocorrência de uma substring por outra.
//
// Exemplos:
//
//	ReplaceLast("hello hello", "hello", "hi") -> "hello hi"
//	ReplaceLast("", "x", "y") -> ""
//	ReplaceLast("hello", "", "x") -> "hello"
func ReplaceLast(s, old, new string) string {
	if s == "" {
		return s
	}
	if old == "" {
		return s
	}

	pos := strings.LastIndex(s, old)
	if pos == -1 {
		return s
	}

	return s[:pos] + new + s[pos+len(old):]
}

// ReplaceNonAlphanumeric substitui todos os caracteres não alfanuméricos por uma string especificada.
//
// Exemplos:
//
//	ReplaceNonAlphanumeric("hello, world!", "-") -> "hello--world-"
//	ReplaceNonAlphanumeric("test@example.com", "_") -> "test_example_com"
func ReplaceNonAlphanumeric(s, replacement string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s))

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteString(replacement)
		}
	}

	return result.String()
}
