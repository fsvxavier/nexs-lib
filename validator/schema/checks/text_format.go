package checks

import (
	"regexp"
)

// TextMatchChecker validates text with only letters, underscore and spaces
type TextMatchChecker struct{}

// IsFormat validates if input matches text pattern
func (TextMatchChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z_ ]*$")
	return r.MatchString(asString)
}

// FormatName returns the name of this format checker
func (TextMatchChecker) FormatName() string {
	return "text_match"
}

// TextMatchWithNumberChecker validates text with letters, numbers, underscore and spaces
type TextMatchWithNumberChecker struct{}

// IsFormat validates if input matches text with number pattern
func (TextMatchWithNumberChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z1-9_ ]*$")
	return r.MatchString(asString)
}

// FormatName returns the name of this format checker
func (TextMatchWithNumberChecker) FormatName() string {
	return "text_match_with_number"
}

// TextMatchCustomChecker validates text with custom regex
type TextMatchCustomChecker struct {
	regex   string
	pattern *regexp.Regexp
}

// NewTextMatchCustom creates a new custom text matcher
func NewTextMatchCustom(regex string) *TextMatchCustomChecker {
	pattern, err := regexp.Compile(regex)
	if err != nil {
		// Return a checker that always fails if pattern is invalid
		pattern = regexp.MustCompile("$.^") // This regex never matches
	}

	return &TextMatchCustomChecker{
		regex:   regex,
		pattern: pattern,
	}
}

// IsFormat validates if input matches custom pattern
func (match *TextMatchCustomChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	return match.pattern.MatchString(asString)
}

// FormatName returns the name of this format checker
func (match *TextMatchCustomChecker) FormatName() string {
	return "text_match_custom"
}
