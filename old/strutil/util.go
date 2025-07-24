package strutil

import (
	"sync"
	"unicode/utf8"
)

// Len is an alias of utf8.RuneCountInString which returns the number of
// runes in s. Erroneous and short encodings are treated as single runes of
// width 1 byte.
var (
	Len              = utf8.RuneCountInString
	uppercaseAcronym = sync.Map{}
)

// "ID": "id",

// ConfigureAcronym allows you to add additional words which will be considered acronyms.
func ConfigureAcronym(key, val string) {
	uppercaseAcronym.Store(key, val)
}
