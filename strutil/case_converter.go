package strutil

import (
	"strings"
	"unicode"
)

// toCamelInitCase is the core implementation for camel case conversion.
// Optimized with strings.Builder pre-allocation and efficient byte operations.
func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Check for acronym replacement
	replacement, hasAcronym := GetAcronym(s)
	if hasAcronym {
		s = replacement
	}

	// First convert to snake_case to properly identify word boundaries
	snakeCase := ToDelimited(s, '_')

	// Then convert snake_case to camelCase
	words := strings.Split(snakeCase, "_")
	if len(words) == 0 {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s))

	for i, word := range words {
		if word == "" {
			continue
		}

		if i == 0 && !initCase {
			// First word should be lowercase for lowerCamelCase
			builder.WriteString(strings.ToLower(word))
		} else {
			// Capitalize first letter, lowercase the rest
			if len(word) > 0 {
				builder.WriteString(strings.ToUpper(word[:1]))
				if len(word) > 1 {
					builder.WriteString(strings.ToLower(word[1:]))
				}
			}
		}
	}

	return builder.String()
}

// isWordBoundary determines if a character represents a word boundary.
// Optimized for common delimiters used in string conversion.
func isWordBoundary(b byte) bool {
	return b == '_' || b == ' ' || b == '-' || b == '.' || b == ':' || b == '/'
}

// ToCamel converts a string to CamelCase.
// Optimized for performance with pre-allocated string builder.
//
// Example:
//
//	ToCamel("hello_world") // "HelloWorld"
//	ToCamel("API_key")     // "APIKey" (respects acronyms)
func ToCamel(s string) string {
	return toCamelInitCase(s, true)
}

// ToLowerCamel converts a string to lowerCamelCase.
// First character is always lowercase, subsequent words are capitalized.
//
// Example:
//
//	ToLowerCamel("hello_world") // "helloWorld"
//	ToLowerCamel("API_key")     // "apiKey"
func ToLowerCamel(s string) string {
	return toCamelInitCase(s, false)
}

// ToCamelCase is an alternative implementation focusing on space-separated words.
// This version preserves existing case except for word boundaries marked by spaces.
//
// Example:
//
//	ToCamelCase("camel case")           // "camelCase"
//	ToCamelCase("inside dynaMIC-HTML")  // "insideDynaMIC-HTML"
func ToCamelCase(str string) string {
	str = strings.TrimSpace(str)
	if Len(str) < 2 {
		return str
	}

	var builder strings.Builder
	builder.Grow(len(str)) // Pre-allocate capacity

	var prev rune
	for _, r := range str {
		if r != ' ' {
			if prev == ' ' {
				// Capitalize first letter of each word (except first)
				builder.WriteString(strings.ToUpper(string(r)))
			} else {
				builder.WriteRune(r)
			}
		}
		prev = r
	}

	return builder.String()
}

// ToSnakeCase converts string to snake_case with space-to-underscore conversion.
// Optimized for simple space replacement after trimming and lowercasing.
//
// Example:
//
//	ToSnakeCase("Snake Case") // "snake_case"
func ToSnakeCase(str string) string {
	if str == "" {
		return str
	}

	str = strings.TrimSpace(strings.ToLower(str))
	return strings.ReplaceAll(str, " ", "_")
}

// ToSnake converts a string to snake_case using delimiter-based conversion.
// More sophisticated than ToSnakeCase, handles various word boundaries.
//
// Example:
//
//	ToSnake("HelloWorld")    // "hello_world"
//	ToSnake("XMLParser")     // "xml_parser"
func ToSnake(s string) string {
	return ToDelimited(s, '_')
}

// ToSnakeWithIgnore converts to snake_case while ignoring specified characters.
func ToSnakeWithIgnore(s, ignore string) string {
	return ToScreamingDelimited(s, ignore, '_', false)
}

// ToScreamingSnake converts a string to SCREAMING_SNAKE_CASE.
//
// Example:
//
//	ToScreamingSnake("hello_world") // "HELLO_WORLD"
func ToScreamingSnake(s string) string {
	return ToScreamingDelimited(s, "", '_', true)
}

// ToKebab converts a string to kebab-case.
//
// Example:
//
//	ToKebab("HelloWorld") // "hello-world"
func ToKebab(s string) string {
	return ToDelimited(s, '-')
}

// ToScreamingKebab converts a string to SCREAMING-KEBAB-CASE.
//
// Example:
//
//	ToScreamingKebab("hello-world") // "HELLO-WORLD"
func ToScreamingKebab(s string) string {
	return ToScreamingDelimited(s, "", '-', true)
}

// ToDelimited converts a string to delimited format with custom delimiter.
// This is the base function for snake_case, kebab-case, and other formats.
//
// Example:
//
//	ToDelimited("HelloWorld", '.') // "hello.world"
func ToDelimited(s string, delimiter uint8) string {
	return ToScreamingDelimited(s, "", delimiter, false)
}

// ToScreamingDelimited is the core conversion function supporting various formats.
// Highly optimized with strings.Builder and efficient character processing.
func ToScreamingDelimited(s, ignore string, delimiter uint8, screaming bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	var builder strings.Builder
	// Pre-allocate with extra space for potential delimiters
	builder.Grow(len(s) + len(s)/4)

	ignoreMap := make(map[rune]bool)
	for _, r := range ignore {
		ignoreMap[r] = true
	}

	runes := []rune(s)

	for i, r := range runes {
		// Check if this character should be ignored
		if ignoreMap[r] {
			builder.WriteRune(r)
			continue
		}

		isUpper := unicode.IsUpper(r)
		isDigit := unicode.IsDigit(r)
		isLetter := unicode.IsLetter(r)

		// Check if we need to add delimiter before this character
		needDelimiter := false

		if i > 0 && (isLetter || isDigit) {
			prevRune := runes[i-1]
			prevIsUpper := unicode.IsUpper(prevRune)
			prevIsLower := unicode.IsLower(prevRune)
			prevIsDigit := unicode.IsDigit(prevRune)

			// Add delimiter in these cases:
			// 1. Transition from lowercase to uppercase: helloWorld -> hello_World
			// 2. Transition from digit to letter: hello2World -> hello2_World
			// 3. Transition from letter to digit: hello2 -> hello_2
			// 4. In consecutive uppercase, before the last one if followed by lowercase:
			//    XMLParser -> XML_Parser (before P)

			if prevIsLower && isUpper {
				needDelimiter = true
			} else if prevIsDigit && isLetter {
				needDelimiter = true
			} else if prevIsUpper && isUpper && i+1 < len(runes) {
				// Look ahead for lowercase to detect end of acronym
				nextRune := runes[i+1]
				if unicode.IsLower(nextRune) {
					needDelimiter = true
				}
			}
		}

		// Add delimiter if needed
		if needDelimiter && builder.Len() > 0 {
			builder.WriteByte(delimiter)
		}

		// Process separators (spaces, underscores, hyphens, dots)
		if unicode.IsSpace(r) || r == '_' || r == '-' || r == '.' {
			if builder.Len() > 0 && builder.String()[builder.Len()-1] != delimiter {
				builder.WriteByte(delimiter)
			}
			continue
		}

		// Write the character
		if isLetter {
			if screaming {
				builder.WriteRune(unicode.ToUpper(r))
			} else {
				builder.WriteRune(unicode.ToLower(r))
			}
		} else if isDigit {
			builder.WriteRune(r)
		}
	}

	return builder.String()
} // GetDelimiter returns the byte value for common delimiter names.
// Provides a convenient way to get delimiter values for conversion functions.
func GetDelimiter(name string) (uint8, bool) {
	if delimiter, exists := commonDelimiters[name]; exists {
		return delimiter, true
	}
	return 0, false
}

// isLetter checks if a rune is a letter (optimized for ASCII).
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || unicode.IsLetter(r)
}

// isDigit checks if a rune is a digit (optimized for ASCII).
func isDigit(r rune) bool {
	return (r >= '0' && r <= '9') || unicode.IsDigit(r)
}
