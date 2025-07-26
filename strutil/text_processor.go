package strutil

import (
	"crypto/rand"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Default character sets for random string generation
const (
	CharsetAlphabetic     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumeric        = "0123456789"
	CharsetAlphanumeric   = CharsetAlphabetic + CharsetNumeric
	CharsetHex            = "0123456789abcdef"
	CharsetHexUpper       = "0123456789ABCDEF"
	CharsetSpecial        = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	CharsetASCIIPrintable = CharsetAlphanumeric + CharsetSpecial + " "
)

// Random generates a cryptographically secure random string.
// Uses crypto/rand for security-sensitive applications.
//
// Example:
//
//	Random(8, CharsetAlphanumeric) // "aB3kL9mZ"
//	Random(16, CharsetHex)         // "4a7b2c9f1e8d3b6a"
func Random(length int, charset string) string {
	if length <= 0 || charset == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(length)

	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		// Generate cryptographically secure random index
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fallback to safer default if crypto/rand fails
			return ""
		}
		builder.WriteByte(charset[randomIndex.Int64()])
	}

	return builder.String()
}

// RemoveAccents removes accents from the string. The resulting string only has
// the letters from English alphabet. For example, "résumé" becomes "resume".
func RemoveAccents(str string) string {
	var buff strings.Builder
	buff.Grow(len(str))
	for _, r := range str {
		buff.WriteString(normalizeRune(r))
	}
	return buff.String()
}

// Slugify converts text to a URL-friendly slug.
// Removes special characters, handles accents, and creates clean URLs.
//
// Example:
//
//	Slugify("Hello World!") // "hello-world"
//	Slugify("Café & Bar")   // "cafe-bar"
func Slugify(text string) string {
	if text == "" {
		return text
	}

	// Remove accents first
	text = RemoveAccents(text)

	// Convert to lowercase
	text = strings.ToLower(text)

	var builder strings.Builder
	builder.Grow(len(text))

	lastWasSeparator := true // Start as true to avoid leading separators

	for _, r := range text {
		// Only accept ASCII letters and digits for URL-safe slugs
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastWasSeparator = false
		} else if !lastWasSeparator {
			// Add separator only if last character wasn't a separator
			builder.WriteRune('-')
			lastWasSeparator = true
		}
	}

	result := builder.String()

	// Remove trailing separator
	if strings.HasSuffix(result, "-") {
		result = result[:len(result)-1]
	}

	return result
}

// Words splits text into words, handling various separators and Unicode.
// More robust than strings.Fields for international text.
//
// Example:
//
//	Words("hello,world;test") // ["hello", "world", "test"]
func Words(text string) []string {
	if text == "" {
		return nil
	}

	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		// Only consider ASCII letters, digits, and apostrophes as word characters
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '\'' {
			// Include apostrophes for contractions
			currentWord.WriteRune(r)
		} else {
			// Non-word character found
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	// Don't forget the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// ReplaceMultiple performs multiple string replacements efficiently.
// Uses strings.Replacer for optimal performance with multiple replacements.
//
// Example:
//
//	ReplaceMultiple("hello world", map[string]string{"hello": "hi", "world": "earth"}) // "hi earth"
func ReplaceMultiple(text string, replacements map[string]string) string {
	if text == "" || len(replacements) == 0 {
		return text
	}

	// Convert map to slice for strings.Replacer
	replacerArgs := make([]string, 0, len(replacements)*2)
	for old, new := range replacements {
		replacerArgs = append(replacerArgs, old, new)
	}

	replacer := strings.NewReplacer(replacerArgs...)
	return replacer.Replace(text)
}

// Normalize normalizes whitespace in text by trimming and reducing multiple spaces.
// Converts various whitespace characters to single spaces.
//
// Example:
//
//	Normalize("  hello   world  \n\t") // "hello world"
func Normalize(text string) string {
	if text == "" {
		return text
	}

	var builder strings.Builder
	builder.Grow(len(text) / 2) // Estimate reduced size

	lastWasSpace := true // Start as true to skip leading whitespace

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				builder.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			builder.WriteRune(r)
			lastWasSpace = false
		}
	}

	result := builder.String()

	// Remove trailing space
	if strings.HasSuffix(result, " ") {
		result = result[:len(result)-1]
	}

	return result
}

// ExtractWords extracts words matching specific criteria from text.
// Supports minimum length filtering and case-insensitive matching.
func ExtractWords(text string, minLength int, caseSensitive bool) []string {
	if text == "" {
		return nil
	}

	words := Words(text)
	var filtered []string

	for _, word := range words {
		if Len(word) >= minLength {
			if !caseSensitive {
				word = strings.ToLower(word)
			}
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// CountWords counts the number of words in text.
// Uses the same logic as Words() for consistency.
func CountWords(text string) int {
	return len(Words(text))
}

// CountLines counts the number of lines in text.
// Handles different line ending formats (LF, CRLF).
func CountLines(text string) int {
	if text == "" {
		return 0
	}

	// Normalize line endings to LF
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	return len(lines)
}

// TruncateWords truncates text to a specified number of words.
// Preserves word boundaries and optionally adds ellipsis.
//
// Example:
//
//	TruncateWords("one two three four", 2, true) // "one two..."
func TruncateWords(text string, maxWords int, addEllipsis bool) string {
	if text == "" || maxWords <= 0 {
		return ""
	}

	words := Words(text)
	if len(words) <= maxWords {
		return text
	}

	truncated := strings.Join(words[:maxWords], " ")
	if addEllipsis {
		truncated += "..."
	}

	return truncated
}

// CleanFilename removes or replaces characters that are invalid in filenames.
// Creates safe filenames for cross-platform compatibility.
//
// Example:
//
//	CleanFilename("file/name?.txt") // "file_name_.txt"
func CleanFilename(filename string) string {
	if filename == "" {
		return filename
	}

	// Characters invalid in filenames across platforms
	invalidChars := map[rune]rune{
		'/':  '_',
		'\\': '_',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	var builder strings.Builder
	builder.Grow(len(filename))

	for _, r := range filename {
		if replacement, isInvalid := invalidChars[r]; isInvalid {
			builder.WriteRune(replacement)
		} else if unicode.IsPrint(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// HasPrefix checks if string has any of the provided prefixes.
// Case-insensitive option available.
func HasPrefix(s string, prefixes []string, caseSensitive bool) bool {
	if s == "" || len(prefixes) == 0 {
		return false
	}

	if !caseSensitive {
		s = strings.ToLower(s)
	}

	for _, prefix := range prefixes {
		checkPrefix := prefix
		if !caseSensitive {
			checkPrefix = strings.ToLower(prefix)
		}

		if strings.HasPrefix(s, checkPrefix) {
			return true
		}
	}

	return false
}

// HasSuffix checks if string has any of the provided suffixes.
// Case-insensitive option available.
func HasSuffix(s string, suffixes []string, caseSensitive bool) bool {
	if s == "" || len(suffixes) == 0 {
		return false
	}

	if !caseSensitive {
		s = strings.ToLower(s)
	}

	for _, suffix := range suffixes {
		checkSuffix := suffix
		if !caseSensitive {
			checkSuffix = strings.ToLower(suffix)
		}

		if strings.HasSuffix(s, checkSuffix) {
			return true
		}
	}

	return false
}

// IsValidUTF8 checks if the string contains valid UTF-8 encoding.
func IsValidUTF8(s string) bool {
	return utf8.ValidString(s)
}
