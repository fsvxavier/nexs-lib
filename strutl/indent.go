package strutl

import "strings"

// Indent indenta cada linha da string str com o parâmetro left.
// Linhas vazias também são indentadas.
//
// Exemplos:
//
//	Indent("linha1\nlinha2", "  ") -> "  linha1\n  linha2"
//	Indent("texto", "> ") -> "> texto"
//	Indent("a\n\nb", "-") -> "-a\n-\n-b"
func Indent(str, left string) string {
	if str == "" {
		return str
	}
	if left == "" {
		return str
	}

	// Estratégia mais eficiente que strings.ReplaceAll para indentação
	lines := strings.Split(str, "\n")
	for i := range lines {
		lines[i] = left + lines[i]
	}
	return strings.Join(lines, "\n")
}
