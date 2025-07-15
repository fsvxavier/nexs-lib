package pgx

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	// Unknown errors
	ErrorTypeUnknown ErrorType = "unknown"
)

// DatabaseError represents a wrapped database error with additional context
type DatabaseError struct {
	Type           ErrorType
	OriginalErr    error
	SQLState       string
	Message        string
	Detail         string
	Hint           string
	Position       int32
	InternalPos    int32
	InternalQuery  string
	Where          string
	SchemaName     string
	TableName      string
	ColumnName     string
	DataTypeName   string
	ConstraintName string
	FileName       string
	LineNumber     int32
	RoutineName    string
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

// WrapError wraps a PGX error with additional context and classification
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Handle PGX specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return wrapPgError(pgErr)
	}

	// Handle connection errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return wrapNetError(netErr)
	}

	// Handle syscall errors
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		return wrapSyscallError(syscallErr, err)
	}

	// Handle PGX specific errors
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return &DatabaseError{
			Type:        ErrorTypeUnknown,
			OriginalErr: err,
			Message:     "no rows found",
		}
	case errors.Is(err, pgx.ErrTxClosed):
		return &DatabaseError{
			Type:        ErrorTypeInvalidTransactionState,
			OriginalErr: err,
			Message:     "transaction is closed",
		}
	case errors.Is(err, pgx.ErrTxCommitRollback):
		return &DatabaseError{
			Type:        ErrorTypeTransactionRollback,
			OriginalErr: err,
			Message:     "transaction has been rolled back",
		}
	}

	// Default unknown error
	return &DatabaseError{
		Type:        ErrorTypeUnknown,
		OriginalErr: err,
		Message:     err.Error(),
	}
}

// wrapPgError wraps PostgreSQL specific errors
func wrapPgError(pgErr *pgconn.PgError) *DatabaseError {
	dbErr := &DatabaseError{
		OriginalErr:    pgErr,
		SQLState:       pgErr.Code,
		Message:        pgErr.Message,
		Detail:         pgErr.Detail,
		Hint:           pgErr.Hint,
		Position:       pgErr.Position,
		InternalPos:    pgErr.InternalPosition,
		InternalQuery:  pgErr.InternalQuery,
		Where:          pgErr.Where,
		SchemaName:     pgErr.SchemaName,
		TableName:      pgErr.TableName,
		ColumnName:     pgErr.ColumnName,
		DataTypeName:   pgErr.DataTypeName,
		ConstraintName: pgErr.ConstraintName,
		FileName:       pgErr.File,
		LineNumber:     pgErr.Line,
		RoutineName:    pgErr.Routine,
	}

	// Map SQL state codes to error types
	switch pgErr.Code {
	// Connection errors
	case "08000", "08003", "08006":
		dbErr.Type = ErrorTypeConnectionFailed
	case "08001":
		dbErr.Type = ErrorTypeConnectionRefused
	case "08004":
		dbErr.Type = ErrorTypeAuthenticationFail

	// Syntax errors
	case "42601":
		dbErr.Type = ErrorTypeSyntaxError
	case "42P01":
		dbErr.Type = ErrorTypeUndefinedTable
	case "42703":
		dbErr.Type = ErrorTypeUndefinedColumn
	case "42883":
		dbErr.Type = ErrorTypeUndefinedFunction

	// Data type errors
	case "42804":
		dbErr.Type = ErrorTypeDataTypeMismatch
	case "22012":
		dbErr.Type = ErrorTypeDivisionByZero
	case "22001":
		dbErr.Type = ErrorTypeStringDataRightTruncation
	case "22003":
		dbErr.Type = ErrorTypeNumericValueOutOfRange
	case "22P02":
		dbErr.Type = ErrorTypeInvalidTextRepresentation
	case "22008":
		dbErr.Type = ErrorTypeInvalidDatetimeFormat

	// Constraint violations
	case "23505":
		dbErr.Type = ErrorTypeUniqueViolation
	case "23503":
		dbErr.Type = ErrorTypeForeignKeyViolation
	case "23502":
		dbErr.Type = ErrorTypeNotNullViolation
	case "23514":
		dbErr.Type = ErrorTypeCheckViolation

	// Transaction errors
	case "25P02":
		dbErr.Type = ErrorTypeTransactionAborted
	case "40001":
		dbErr.Type = ErrorTypeSerializationFailure
	case "40P01":
		dbErr.Type = ErrorTypeDeadlockDetected
	case "25000", "25001", "25002", "25008":
		dbErr.Type = ErrorTypeInvalidTransactionState

	// System errors
	case "53100":
		dbErr.Type = ErrorTypeDiskFull
	case "53200":
		dbErr.Type = ErrorTypeInsufficientMemory

	default:
		dbErr.Type = ErrorTypeUnknown
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
			ErrorTypeInvalidTransactionState:
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
