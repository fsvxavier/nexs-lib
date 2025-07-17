package pgx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Custom error types for PGX provider
var (
	// ErrPoolClosed is returned when trying to use a closed pool
	ErrPoolClosed = errors.New("pool is closed")

	// ErrConnectionClosed is returned when trying to use a closed connection
	ErrConnectionClosed = errors.New("connection is closed")

	// ErrTransactionClosed is returned when trying to use a closed transaction
	ErrTransactionClosed = errors.New("transaction is closed")

	// ErrBatchClosed is returned when trying to use a closed batch
	ErrBatchClosed = errors.New("batch is closed")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrUnsupportedFeature is returned when a feature is not supported
	ErrUnsupportedFeature = errors.New("feature not supported")
)

// PGXError wraps PostgreSQL errors with additional context
type PGXError struct {
	Err       error
	Operation string
	Query     string
	Args      []interface{}
	Context   map[string]interface{}
}

func (e *PGXError) Error() string {
	if e.Query != "" {
		return fmt.Sprintf("pgx error in %s: %v (query: %s)", e.Operation, e.Err, e.Query)
	}
	return fmt.Sprintf("pgx error in %s: %v", e.Operation, e.Err)
}

func (e *PGXError) Unwrap() error {
	return e.Err
}

// IsPGError checks if an error is a PostgreSQL error
func IsPGError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

// GetPGError extracts PostgreSQL error details
func GetPGError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

// IsNoRowsError checks if an error indicates no rows were found
func IsNoRowsError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// IsConnError checks if an error is related to connection issues
func IsConnError(err error) bool {
	// Check for our custom connection error
	if errors.Is(err, ErrConnectionClosed) {
		return true
	}

	// Check for PostgreSQL connection errors
	pgErr := GetPGError(err)
	if pgErr != nil {
		// PostgreSQL connection error codes (Class 08)
		return strings.HasPrefix(pgErr.Code, "08")
	}

	return false
}

// IsTxError checks if an error is related to transaction issues
func IsTxError(err error) bool {
	return errors.Is(err, pgx.ErrTxClosed) ||
		errors.Is(err, pgx.ErrTxCommitRollback) ||
		errors.Is(err, ErrTransactionClosed)
}

// WrapError wraps an error with PGX context
func WrapError(err error, operation, query string, args []interface{}) error {
	if err == nil {
		return nil
	}

	return &PGXError{
		Err:       err,
		Operation: operation,
		Query:     query,
		Args:      args,
		Context:   make(map[string]interface{}),
	}
}

// AddErrorContext adds context information to a PGXError
func AddErrorContext(err error, key string, value interface{}) error {
	if pgxErr, ok := err.(*PGXError); ok {
		if pgxErr.Context == nil {
			pgxErr.Context = make(map[string]interface{})
		}
		pgxErr.Context[key] = value
		return pgxErr
	}
	return err
}

// Error type constants for common PostgreSQL errors
const (
	// Connection errors
	PGErrorClassConnection = "08"

	// Data type errors
	PGErrorClassDataException = "22"

	// Integrity constraint violations
	PGErrorClassIntegrityConstraintViolation = "23"

	// Syntax errors
	PGErrorClassSyntaxError = "42"

	// Insufficient resources
	PGErrorClassInsufficientResources = "53"

	// System errors
	PGErrorClassSystemError = "54"

	// Specific error codes
	PGErrorCodeUniqueViolation      = "23505"
	PGErrorCodeForeignKeyViolation  = "23503"
	PGErrorCodeNotNullViolation     = "23502"
	PGErrorCodeCheckViolation       = "23514"
	PGErrorCodeSerializationFailure = "40001"
	PGErrorCodeDeadlockDetected     = "40P01"
)

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeUniqueViolation
}

// IsForeignKeyViolation checks if the error is a foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeForeignKeyViolation
}

// IsNotNullViolation checks if the error is a not-null constraint violation
func IsNotNullViolation(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeNotNullViolation
}

// IsCheckViolation checks if the error is a check constraint violation
func IsCheckViolation(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeCheckViolation
}

// IsSerializationFailure checks if the error is a serialization failure
func IsSerializationFailure(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeSerializationFailure
}

// IsDeadlock checks if the error is a deadlock
func IsDeadlock(err error) bool {
	pgErr := GetPGError(err)
	return pgErr != nil && pgErr.Code == PGErrorCodeDeadlockDetected
}

// GetErrorClass returns the PostgreSQL error class
func GetErrorClass(err error) string {
	pgErr := GetPGError(err)
	if pgErr != nil && len(pgErr.Code) >= 2 {
		return pgErr.Code[:2]
	}
	return ""
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if IsConnError(err) {
		return true
	}

	pgErr := GetPGError(err)
	if pgErr == nil {
		return false
	}

	// Serialization failures and deadlocks are typically retryable
	return pgErr.Code == PGErrorCodeSerializationFailure ||
		pgErr.Code == PGErrorCodeDeadlockDetected
}

// GetErrorDetails extracts detailed information from an error
func GetErrorDetails(err error) map[string]interface{} {
	details := make(map[string]interface{})

	if err == nil {
		return details
	}

	details["error"] = err.Error()

	// Check for PGX specific errors
	if errors.Is(err, pgx.ErrNoRows) {
		details["type"] = "no_rows"
		return details
	}

	// Check for PostgreSQL errors
	pgErr := GetPGError(err)
	if pgErr != nil {
		details["type"] = getErrorType(pgErr.Code)
		details["code"] = pgErr.Code
		details["severity"] = pgErr.Severity
		details["message"] = pgErr.Message

		if pgErr.Detail != "" {
			details["detail"] = pgErr.Detail
		}
		if pgErr.Hint != "" {
			details["hint"] = pgErr.Hint
		}
		if pgErr.SchemaName != "" {
			details["schema"] = pgErr.SchemaName
		}
		if pgErr.TableName != "" {
			details["table"] = pgErr.TableName
		}
		if pgErr.ColumnName != "" {
			details["column"] = pgErr.ColumnName
		}
		if pgErr.ConstraintName != "" {
			details["constraint"] = pgErr.ConstraintName
		}

		return details
	}

	details["type"] = "unknown"
	return details
}

// getErrorType returns a human-readable error type based on PostgreSQL error code
func getErrorType(code string) string {
	switch code {
	case "23505":
		return "constraint_violation"
	case "23503":
		return "constraint_violation"
	case "23502":
		return "constraint_violation"
	case "23514":
		return "constraint_violation"
	case "42601":
		return "syntax_error"
	case "42501":
		return "permission_error"
	case "08001", "08003", "08004", "08006":
		return "connection_error"
	case "40001":
		return "transaction_error"
	case "40P01":
		return "transaction_error"
	default:
		return "database_error"
	}
}
