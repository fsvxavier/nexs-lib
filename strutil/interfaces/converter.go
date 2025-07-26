// Package interfaces provides contract definitions for string utility operations.
// These interfaces follow SOLID principles and enable dependency injection for better testability.
package interfaces

// CaseConverter defines the contract for string case conversion operations.
// This interface follows the Interface Segregation Principle by grouping related case conversion methods.
type CaseConverter interface {
	// ToCamel converts a string to CamelCase
	ToCamel(s string) string

	// ToLowerCamel converts a string to lowerCamelCase
	ToLowerCamel(s string) string

	// ToSnake converts a string to snake_case
	ToSnake(s string) string

	// ToScreamingSnake converts a string to SCREAMING_SNAKE_CASE
	ToScreamingSnake(s string) string

	// ToKebab converts a string to kebab-case
	ToKebab(s string) string

	// ToScreamingKebab converts a string to SCREAMING-KEBAB-CASE
	ToScreamingKebab(s string) string

	// ToDelimited converts a string to delimited format with custom delimiter
	ToDelimited(s string, delimiter uint8) string
}

// StringFormatter defines the contract for string formatting operations.
// This interface groups formatting-related functionality following SRP.
type StringFormatter interface {
	// Align aligns text within specified width
	Align(text string, align int, width int) string

	// Center centers text within specified width
	Center(text string, width int) string

	// PadLeft pads string on the left with specified character
	PadLeft(str string, pad string, length int) string

	// PadRight pads string on the right with specified character
	PadRight(str string, pad string, length int) string

	// PadBoth pads string on both sides with specified character
	PadBoth(str string, pad string, length int) string
}

// StringValidator defines the contract for string validation operations.
type StringValidator interface {
	// IsASCII checks if string contains only ASCII characters
	IsASCII(s string) bool

	// IsEmpty checks if string is empty or contains only whitespace
	IsEmpty(s string) bool
}

// TextProcessor defines the contract for text processing operations.
type TextProcessor interface {
	// WordWrap wraps text to specified line length
	WordWrap(text string, lineWidth int) string

	// Indent indents text with specified prefix
	Indent(text string, prefix string) string

	// ExpandTabs expands tabs to spaces
	ExpandTabs(text string, tabSize int) string

	// RemoveAccents removes accents from text
	RemoveAccents(text string) string

	// Slugify converts text to URL-friendly slug
	Slugify(text string) string
}

// StringManipulator defines the contract for string manipulation operations.
type StringManipulator interface {
	// Reverse reverses a string
	Reverse(s string) string

	// Substring extracts substring with bounds checking
	Substring(s string, start, end int) string

	// Replace performs string replacement with various options
	Replace(s, old, new string) string

	// Random generates random string with specified length and character set
	Random(length int, charset string) string
}

// AcronymManager defines the contract for managing acronyms in string conversions.
type AcronymManager interface {
	// ConfigureAcronym adds or updates an acronym mapping
	ConfigureAcronym(acronym, replacement string)

	// GetAcronym retrieves the replacement for an acronym
	GetAcronym(acronym string) (string, bool)

	// RemoveAcronym removes an acronym mapping
	RemoveAcronym(acronym string)

	// ClearAcronyms removes all acronym mappings
	ClearAcronyms()
}
