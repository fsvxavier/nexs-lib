package checks

// EmptyStringChecker validates empty string
type EmptyStringChecker struct{}

// IsFormat validates if input is an empty string
func (EmptyStringChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	return len(asString) == 0
}

// FormatName returns the name of this format checker
func (EmptyStringChecker) FormatName() string {
	return "empty_string"
}
