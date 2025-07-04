package strutl

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RemoveAccents remove acentos e sinais diacríticos de uma string.
//
// Exemplos:
//
//	RemoveAccents("olá") -> "ola"
//	RemoveAccents("àéêöhello") -> "aeeohello"
//	RemoveAccents("ação") -> "acao"
//	RemoveAccents("über") -> "uber"
func RemoveAccents(s string) string {
	// Primeiro normaliza para a forma de decomposição
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// Slugify converte uma string para um formato seguro para uso em URLs.
// Remove acentos, converte para minúsculas e substitui espaços e outros
// caracteres não alfanuméricos por hífens.
//
// Exemplos:
//
//	Slugify("Hello World!") -> "hello-world"
//	Slugify("São Paulo") -> "sao-paulo"
//	Slugify("  a b c  ") -> "a-b-c"
//	Slugify("100%") -> "100"
func Slugify(s string) string {
	// Remove acentos e sinais diacríticos
	s = RemoveAccents(s)

	// Converte para minúsculas
	s = strings.ToLower(s)

	// Substitui caracteres não alfanuméricos por hífens
	var result strings.Builder
	result.Grow(len(s))

	var lastCharWasDash bool

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Se é letra ou dígito, mantém no slug
			result.WriteRune(r)
			lastCharWasDash = false
		} else if !lastCharWasDash {
			// Se não é alfanumérico e não tivemos um hífen anterior, adiciona hífen
			result.WriteRune('-')
			lastCharWasDash = true
		}
		// Ignora outros caracteres e evita hífens consecutivos
	}

	// Remove hífens no início e fim
	slug := result.String()
	slug = strings.Trim(slug, "-")

	return slug
}
