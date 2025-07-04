package checks

import (
	"time"
)

// Constants for datetime formats
const (
	RFC3339TimeOnlyFormat = "15:04:05Z07:00"
	ISO8601DateTimeFormat = "2006-01-02T15:04:05.999Z07:00"
)

// DateTimeChecker validates various date/time formats
type DateTimeChecker struct{}

// IsFormat validates if input is a valid datetime string
func (DateTimeChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	formats := []string{
		time.TimeOnly,         // 15:04:05
		RFC3339TimeOnlyFormat, // 15:04:05Z07:00
		time.DateOnly,         // 2006-01-02
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
		ISO8601DateTimeFormat, // 2006-01-02T15:04:05.999Z07:00
	}

	for _, format := range formats {
		if _, err := time.Parse(format, asString); err == nil {
			return true
		}
	}

	return false
}

// FormatName returns the name of this format checker
func (DateTimeChecker) FormatName() string {
	return "date_time"
}
