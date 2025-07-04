package strutl

import "strings"

// ExpandTabs converte tabs para espaços. O parâmetro count especifica o número de espaços.
//
// Exemplos:
//
//	ExpandTabs("\tlorem\n\tipsum", 2) -> "  lorem\n  ipsum"
//	ExpandTabs("\t\t", 2) -> "    "
func ExpandTabs(str string, count int) string {
	if count <= 0 {
		return strings.ReplaceAll(str, "\t", "")
	}
	return strings.ReplaceAll(str, "\t", strings.Repeat(" ", count))
}
