package strutil

import "strings"

// Indent indents every line of string str with the left parameter
// Empty lines are indented too.
func Indent(str, left string) string {
	return left + strings.ReplaceAll(str, "\n", "\n"+left)
}
