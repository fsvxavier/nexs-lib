package errors

import "errors"

// Common errors
var (
	// ErrUnsupportedOperation is returned when a provider doesn't support a particular operation
	ErrUnsupportedOperation = errors.New("operation not supported by this json provider")
)
