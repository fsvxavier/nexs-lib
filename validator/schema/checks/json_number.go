package checks

import "encoding/json"

// JSONNumberChecker validates JSON number format
type JSONNumberChecker struct{}

// IsFormat validates if input is a JSON number
func (JSONNumberChecker) IsFormat(input interface{}) bool {
	_, ok := input.(json.Number)
	return ok
}

// FormatName returns the name of this format checker
func (JSONNumberChecker) FormatName() string {
	return "json_number"
}
