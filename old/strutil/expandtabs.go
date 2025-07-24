package strutil

import "strings"

// ExpandTabs converts tabs to the spaces. count specifies the number of spaces.
func ExpandTabs(str string, count int) string {
	return strings.ReplaceAll(str, "\t", strings.Repeat(" ", count))
}
