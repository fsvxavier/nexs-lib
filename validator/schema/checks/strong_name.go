package checks

import (
	"regexp"
)

// StrongNameChecker validates strong name format
type StrongNameChecker struct{}

var strongNameRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")

// IsFormat validates if input is a valid strong name
func (StrongNameChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if len(asString) == 0 {
		return false
	}

	// Strong name: alphanumeric, underscore, hyphen, starts with letter
	return strongNameRegex.MatchString(asString)
}

// FormatName returns the name of this format checker
func (StrongNameChecker) FormatName() string {
	return "strong_name"
}
