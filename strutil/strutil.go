// Package strutil provides high-performance string manipulation utilities.
// This package follows Go best practices with focus on performance, memory efficiency,
// and comprehensive testing coverage.
package strutil

import (
	"strings"
	"sync"
	"unicode/utf8"
)

// Global configuration for acronym handling
var (
	// Len is an optimized alias for UTF-8 rune counting
	Len = utf8.RuneCountInString

	// acronymCache provides thread-safe acronym storage with sync.Map for high-performance concurrent access
	acronymCache = sync.Map{}

	// Common delimiters used in string conversion operations
	commonDelimiters = map[string]uint8{
		"snake": '_',
		"kebab": '-',
		"dot":   '.',
		"space": ' ',
		"colon": ':',
		"pipe":  '|',
	}
)

// ConfigureAcronym adds or updates an acronym mapping for case conversion.
// This function is thread-safe and optimized for concurrent access.
//
// Example:
//
//	ConfigureAcronym("ID", "id")
//	ConfigureAcronym("URL", "url")
func ConfigureAcronym(acronym, replacement string) {
	if acronym == "" {
		return
	}
	acronymCache.Store(acronym, replacement)
}

// GetAcronym retrieves the configured replacement for an acronym.
// Returns the replacement string and true if found, empty string and false otherwise.
func GetAcronym(acronym string) (string, bool) {
	if acronym == "" {
		return "", false
	}

	if replacement, exists := acronymCache.Load(acronym); exists {
		return replacement.(string), true
	}
	return "", false
}

// RemoveAcronym removes an acronym mapping from the cache.
func RemoveAcronym(acronym string) {
	if acronym == "" {
		return
	}
	acronymCache.Delete(acronym)
}

// ClearAcronyms removes all acronym mappings from the cache.
func ClearAcronyms() {
	acronymCache.Range(func(key, value interface{}) bool {
		acronymCache.Delete(key)
		return true
	})
}

// IsEmpty checks if a string is empty or contains only whitespace characters.
// This function is optimized for common use cases and handles Unicode correctly.
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsASCII checks if all characters in the string are ASCII (0-127).
// Uses byte-level checking for maximum performance.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		// ASCII range is 0-127, but exclude control characters except common whitespace
		if b > 127 || (b < 32 && b != '\n' && b != '\r' && b != '\t') {
			return false
		}
	}
	return true
}

// SafeSubstring extracts a substring with bounds checking to prevent panics.
// Handles negative indices and out-of-bounds gracefully.
func SafeSubstring(s string, start, end int) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	length := len(runes)

	// Normalize negative indices
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = length
	}

	// Ensure bounds are within string length
	if start >= length {
		return ""
	}
	if end > length {
		end = length
	}
	if start >= end {
		return ""
	}

	return string(runes[start:end])
}

// Reverse reverses a string while correctly handling Unicode characters.
// Uses rune-based reversal for proper Unicode support.
func Reverse(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	length := len(runes)

	// In-place reversal for memory efficiency
	for i := 0; i < length/2; i++ {
		runes[i], runes[length-1-i] = runes[length-1-i], runes[i]
	}

	return string(runes)
}
