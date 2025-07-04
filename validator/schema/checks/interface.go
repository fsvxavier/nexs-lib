package checks

// FormatChecker defines the interface for format validation
type FormatChecker interface {
	IsFormat(input interface{}) bool
	FormatName() string
}
