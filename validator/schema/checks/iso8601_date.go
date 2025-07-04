package checks

import "time"

// ISO8601DateChecker validates ISO 8601 date format
type ISO8601DateChecker struct{}

// IsFormat validates if input is a valid ISO 8601 date string
func (ISO8601DateChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	// ISO 8601 date format: YYYY-MM-DD
	_, err := time.Parse("2006-01-02", asString)
	return err == nil
}

// FormatName returns the name of this format checker
func (ISO8601DateChecker) FormatName() string {
	return "iso_8601_date"
}
