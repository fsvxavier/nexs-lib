package gorm

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"gorm.io/gorm"
)

// ErrorType represents the type of database error
type ErrorType string

const (
	// Connection errors
	ErrorTypeConnectionFailed   ErrorType = "connection_failed"
	ErrorTypeConnectionLost     ErrorType = "connection_lost"
	ErrorTypeConnectionTimeout  ErrorType = "connection_timeout"
	ErrorTypeConnectionRefused  ErrorType = "connection_refused"
	ErrorTypePoolExhausted      ErrorType = "pool_exhausted"
	ErrorTypeAuthenticationFail ErrorType = "authentication_failed"

	// Query errors
	ErrorTypeSyntaxError       ErrorType = "syntax_error"
	ErrorTypeUndefinedTable    ErrorType = "undefined_table"
	ErrorTypeUndefinedColumn   ErrorType = "undefined_column"
	ErrorTypeUndefinedFunction ErrorType = "undefined_function"
	ErrorTypeDataTypeMismatch  ErrorType = "data_type_mismatch"
	ErrorTypeDivisionByZero    ErrorType = "division_by_zero"

	// Constraint errors
	ErrorTypeUniqueViolation     ErrorType = "unique_violation"
	ErrorTypeForeignKeyViolation ErrorType = "foreign_key_violation"
	ErrorTypeNotNullViolation    ErrorType = "not_null_violation"
	ErrorTypeCheckViolation      ErrorType = "check_violation"

	// Transaction errors
	ErrorTypeTransactionRollback     ErrorType = "transaction_rollback"
	ErrorTypeSerializationFailure    ErrorType = "serialization_failure"
	ErrorTypeDeadlockDetected        ErrorType = "deadlock_detected"
	ErrorTypeTransactionAborted      ErrorType = "transaction_aborted"
	ErrorTypeInvalidTransactionState ErrorType = "invalid_transaction_state"

	// Data errors
	ErrorTypeStringDataRightTruncation ErrorType = "string_data_right_truncation"
	ErrorTypeNumericValueOutOfRange    ErrorType = "numeric_value_out_of_range"
	ErrorTypeInvalidTextRepresentation ErrorType = "invalid_text_representation"
	ErrorTypeInvalidDatetimeFormat     ErrorType = "invalid_datetime_format"

	// System errors
	ErrorTypeDiskFull           ErrorType = "disk_full"
	ErrorTypeInsufficientMemory ErrorType = "insufficient_memory"
	ErrorTypeSystemError        ErrorType = "system_error"

	// GORM specific errors
	ErrorTypeRecordNotFound     ErrorType = "record_not_found"
	ErrorTypeInvalidTransaction ErrorType = "invalid_transaction"
	ErrorTypeInvalidData        ErrorType = "invalid_data"
	ErrorTypeInvalidSQL         ErrorType = "invalid_sql"
	ErrorTypeInvalidValue       ErrorType = "invalid_value"

	// Unknown errors
	ErrorTypeUnknown ErrorType = "unknown"
)

// DatabaseError represents a wrapped database error with additional context
type DatabaseError struct {
	Type        ErrorType
	OriginalErr error
	Message     string
	Detail      string
	Hint        string
	Context     map[string]interface{}
}

// Error implements the error interface
func (e *DatabaseError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (Detail: %s)", e.Type, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the original error
func (e *DatabaseError) Unwrap() error {
	return e.OriginalErr
}

// Is implements error comparison
func (e *DatabaseError) Is(target error) bool {
	if target, ok := target.(*DatabaseError); ok {
		return e.Type == target.Type
	}
	return false
}

// WrapError wraps a GORM error with additional context and classification
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Handle GORM specific errors first
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &DatabaseError{
			Type:        ErrorTypeRecordNotFound,
			OriginalErr: err,
			Message:     "record not found",
		}
	}

	if errors.Is(err, gorm.ErrInvalidTransaction) {
		return &DatabaseError{
			Type:        ErrorTypeInvalidTransaction,
			OriginalErr: err,
			Message:     "invalid transaction",
		}
	}

	if errors.Is(err, gorm.ErrInvalidData) {
		return &DatabaseError{
			Type:        ErrorTypeInvalidData,
			OriginalErr: err,
			Message:     "invalid data",
		}
	}

	if errors.Is(err, gorm.ErrInvalidField) {
		return &DatabaseError{
			Type:        ErrorTypeUndefinedColumn,
			OriginalErr: err,
			Message:     "invalid field",
		}
	}

	if errors.Is(err, gorm.ErrEmptySlice) {
		return &DatabaseError{
			Type:        ErrorTypeInvalidValue,
			OriginalErr: err,
			Message:     "empty slice",
		}
	}

	if errors.Is(err, gorm.ErrInvalidValue) {
		return &DatabaseError{
			Type:        ErrorTypeInvalidValue,
			OriginalErr: err,
			Message:     "invalid value",
		}
	}

	if errors.Is(err, gorm.ErrInvalidValueOfLength) {
		return &DatabaseError{
			Type:        ErrorTypeStringDataRightTruncation,
			OriginalErr: err,
			Message:     "invalid value length",
		}
	}

	// Handle network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return wrapNetError(netErr)
	}

	// Handle syscall errors
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		return wrapSyscallError(syscallErr, err)
	}

	// Handle database errors by message content (GORM often wraps underlying driver errors)
	return wrapByMessage(err)
}

// wrapByMessage classifies errors based on their message content
func wrapByMessage(err error) *DatabaseError {
	message := strings.ToLower(err.Error())

	dbErr := &DatabaseError{
		OriginalErr: err,
		Message:     err.Error(),
		Type:        ErrorTypeUnknown,
	}

	// Connection errors
	switch {
	case strings.Contains(message, "connection refused"):
		dbErr.Type = ErrorTypeConnectionRefused
	case strings.Contains(message, "connection timeout"), strings.Contains(message, "timeout"):
		dbErr.Type = ErrorTypeConnectionTimeout
	case strings.Contains(message, "connection reset"), strings.Contains(message, "connection lost"):
		dbErr.Type = ErrorTypeConnectionLost
	case strings.Contains(message, "authentication failed"), strings.Contains(message, "password authentication failed"):
		dbErr.Type = ErrorTypeAuthenticationFail
	case strings.Contains(message, "too many connections"), strings.Contains(message, "pool exhausted"):
		dbErr.Type = ErrorTypePoolExhausted

	// Syntax and schema errors
	case strings.Contains(message, "syntax error"):
		dbErr.Type = ErrorTypeSyntaxError
	case strings.Contains(message, "relation") && strings.Contains(message, "does not exist"):
		dbErr.Type = ErrorTypeUndefinedTable
	case strings.Contains(message, "column") && strings.Contains(message, "does not exist"):
		dbErr.Type = ErrorTypeUndefinedColumn
	case strings.Contains(message, "function") && strings.Contains(message, "does not exist"):
		dbErr.Type = ErrorTypeUndefinedFunction

	// Data type errors
	case strings.Contains(message, "invalid input syntax"):
		dbErr.Type = ErrorTypeInvalidTextRepresentation
	case strings.Contains(message, "division by zero"):
		dbErr.Type = ErrorTypeDivisionByZero
	case strings.Contains(message, "value too long"):
		dbErr.Type = ErrorTypeStringDataRightTruncation
	case strings.Contains(message, "numeric value out of range"):
		dbErr.Type = ErrorTypeNumericValueOutOfRange
	case strings.Contains(message, "invalid date"), strings.Contains(message, "invalid time"):
		dbErr.Type = ErrorTypeInvalidDatetimeFormat

	// Constraint violations
	case strings.Contains(message, "unique constraint"), strings.Contains(message, "duplicate key"):
		dbErr.Type = ErrorTypeUniqueViolation
	case strings.Contains(message, "foreign key constraint"):
		dbErr.Type = ErrorTypeForeignKeyViolation
	case strings.Contains(message, "not null constraint"), strings.Contains(message, "null value"):
		dbErr.Type = ErrorTypeNotNullViolation
	case strings.Contains(message, "check constraint"):
		dbErr.Type = ErrorTypeCheckViolation

	// Transaction errors
	case strings.Contains(message, "deadlock"):
		dbErr.Type = ErrorTypeDeadlockDetected
	case strings.Contains(message, "serialization failure"), strings.Contains(message, "serialize access"):
		dbErr.Type = ErrorTypeSerializationFailure
	case strings.Contains(message, "transaction"):
		if strings.Contains(message, "aborted") || strings.Contains(message, "rolled back") {
			dbErr.Type = ErrorTypeTransactionAborted
		} else if strings.Contains(message, "invalid") {
			dbErr.Type = ErrorTypeInvalidTransactionState
		} else {
			dbErr.Type = ErrorTypeTransactionRollback
		}

	// System errors
	case strings.Contains(message, "disk full"), strings.Contains(message, "no space left"):
		dbErr.Type = ErrorTypeDiskFull
	case strings.Contains(message, "out of memory"):
		dbErr.Type = ErrorTypeInsufficientMemory
	}

	return dbErr
}

// wrapNetError wraps network-related errors
func wrapNetError(netErr net.Error) *DatabaseError {
	errorType := ErrorTypeConnectionFailed
	if netErr.Timeout() {
		errorType = ErrorTypeConnectionTimeout
	}

	return &DatabaseError{
		Type:        errorType,
		OriginalErr: netErr,
		Message:     netErr.Error(),
	}
}

// wrapSyscallError wraps system call errors
func wrapSyscallError(errno syscall.Errno, err error) *DatabaseError {
	var errorType ErrorType
	switch errno {
	case syscall.ECONNREFUSED:
		errorType = ErrorTypeConnectionRefused
	case syscall.ETIMEDOUT:
		errorType = ErrorTypeConnectionTimeout
	case syscall.ECONNRESET, syscall.EPIPE:
		errorType = ErrorTypeConnectionLost
	default:
		errorType = ErrorTypeSystemError
	}

	return &DatabaseError{
		Type:        errorType,
		OriginalErr: err,
		Message:     err.Error(),
	}
}

// IsConnectionError checks if the error is a connection-related error
func IsConnectionError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		switch dbErr.Type {
		case ErrorTypeConnectionFailed, ErrorTypeConnectionLost,
			ErrorTypeConnectionTimeout, ErrorTypeConnectionRefused,
			ErrorTypePoolExhausted, ErrorTypeAuthenticationFail:
			return true
		}
	}
	return false
}

// IsConstraintViolation checks if the error is a constraint violation
func IsConstraintViolation(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		switch dbErr.Type {
		case ErrorTypeUniqueViolation, ErrorTypeForeignKeyViolation,
			ErrorTypeNotNullViolation, ErrorTypeCheckViolation:
			return true
		}
	}
	return false
}

// IsTransactionError checks if the error is transaction-related
func IsTransactionError(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		switch dbErr.Type {
		case ErrorTypeTransactionRollback, ErrorTypeSerializationFailure,
			ErrorTypeDeadlockDetected, ErrorTypeTransactionAborted,
			ErrorTypeInvalidTransactionState, ErrorTypeInvalidTransaction:
			return true
		}
	}
	return false
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		switch dbErr.Type {
		case ErrorTypeConnectionTimeout, ErrorTypeConnectionLost,
			ErrorTypePoolExhausted, ErrorTypeSerializationFailure,
			ErrorTypeDeadlockDetected:
			return true
		}
	}
	return false
}

// IsNotFound checks if the error indicates a record was not found
func IsNotFound(err error) bool {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr.Type == ErrorTypeRecordNotFound
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}
