package checks

// StringChecker validates if input is a string
type StringChecker struct{}

// IsFormat validates if input is a string
func (StringChecker) IsFormat(input interface{}) bool {
	_, ok := input.(string)
	return ok
}

// FormatName returns the name of this format checker
func (StringChecker) FormatName() string {
	return "string"
}

// IsString is a helper function to check if value is a string
func IsString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}
