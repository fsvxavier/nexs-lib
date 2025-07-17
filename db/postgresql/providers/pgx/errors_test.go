//go:build unit

package pgx

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPGXErrors(t *testing.T) {
	t.Run("WrapError with nil error", func(t *testing.T) {
		result := WrapError(nil, "test", "SELECT 1", []interface{}{})
		assert.Nil(t, result)
	})

	t.Run("WrapError with standard error", func(t *testing.T) {
		originalErr := errors.New("test error")
		wrappedErr := WrapError(originalErr, "query", "SELECT 1", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "test error")
		assert.Contains(t, wrappedErr.Error(), "query")
	})

	t.Run("WrapError with PGX ErrNoRows", func(t *testing.T) {
		wrappedErr := WrapError(pgx.ErrNoRows, "query", "SELECT 1", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "no rows in result set")
	})

	t.Run("WrapError with PostgreSQL connection error", func(t *testing.T) {
		// Simulate a connection error
		pgErr := &pgconn.PgError{
			Severity: "FATAL",
			Code:     "08001", // Connection error code
			Message:  "connection refused",
		}

		wrappedErr := WrapError(pgErr, "connect", "", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "connection refused")
	})

	t.Run("WrapError with PostgreSQL syntax error", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "42601", // Syntax error code
			Message:  "syntax error at or near",
		}

		wrappedErr := WrapError(pgErr, "query", "INVALID SQL", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "syntax error")
	})

	t.Run("WrapError with PostgreSQL unique violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "23505", // Unique violation code
			Message:  "duplicate key value violates unique constraint",
		}

		wrappedErr := WrapError(pgErr, "insert", "INSERT INTO users", []interface{}{1, "test"})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "duplicate key")
	})

	t.Run("WrapError with PostgreSQL foreign key violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "23503", // Foreign key violation code
			Message:  "violates foreign key constraint",
		}

		wrappedErr := WrapError(pgErr, "insert", "INSERT INTO orders", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "foreign key")
	})

	t.Run("WrapError with PostgreSQL check violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "23514", // Check violation code
			Message:  "violates check constraint",
		}

		wrappedErr := WrapError(pgErr, "update", "UPDATE users", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "check constraint")
	})

	t.Run("WrapError with PostgreSQL not null violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "23502", // Not null violation code
			Message:  "null value in column violates not-null constraint",
		}

		wrappedErr := WrapError(pgErr, "insert", "INSERT INTO users", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "not-null constraint")
	})

	t.Run("WrapError with PostgreSQL serialization failure", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "40001", // Serialization failure code
			Message:  "could not serialize access",
		}

		wrappedErr := WrapError(pgErr, "transaction", "UPDATE users", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "serialize access")
	})

	t.Run("WrapError with PostgreSQL deadlock", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "40P01", // Deadlock detected code
			Message:  "deadlock detected",
		}

		wrappedErr := WrapError(pgErr, "transaction", "UPDATE users", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "deadlock")
	})

	t.Run("WrapError with PostgreSQL permission denied", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "42501", // Insufficient privilege code
			Message:  "permission denied",
		}

		wrappedErr := WrapError(pgErr, "query", "SELECT * FROM sensitive_table", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "permission denied")
	})

	t.Run("WrapError with unknown PostgreSQL error", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity: "ERROR",
			Code:     "99999", // Unknown error code
			Message:  "unknown error",
		}

		wrappedErr := WrapError(pgErr, "query", "SELECT 1", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "unknown error")
	})

	t.Run("WrapError preserves original error", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := WrapError(originalErr, "test", "SELECT 1", []interface{}{})

		assert.NotNil(t, wrappedErr)
		assert.True(t, errors.Is(wrappedErr, originalErr))
	})

	t.Run("WrapError with context information", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity:         "ERROR",
			Code:             "23505",
			Message:          "duplicate key value",
			Detail:           "Key (id)=(1) already exists",
			Hint:             "Check for existing records",
			Position:         0,
			InternalPosition: 0,
			InternalQuery:    "",
			Where:            "",
			SchemaName:       "public",
			TableName:        "users",
			ColumnName:       "id",
			DataTypeName:     "",
			ConstraintName:   "users_pkey",
			File:             "",
			Line:             0,
			Routine:          "",
		}

		wrappedErr := WrapError(pgErr, "insert", "INSERT INTO users", []interface{}{1, "test"})

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "duplicate key")
	})
}

func TestIsRetryableError(t *testing.T) {
	t.Run("Nil error is not retryable", func(t *testing.T) {
		assert.False(t, IsRetryableError(nil))
	})

	t.Run("Standard error is not retryable", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsRetryableError(err))
	})

	t.Run("Connection errors are retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "08001",
			Message: "connection refused",
		}
		assert.True(t, IsRetryableError(pgErr))
	})

	t.Run("Serialization failures are retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "40001",
			Message: "serialization failure",
		}
		assert.True(t, IsRetryableError(pgErr))
	})

	t.Run("Deadlocks are retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "40P01",
			Message: "deadlock detected",
		}
		assert.True(t, IsRetryableError(pgErr))
	})

	t.Run("Constraint violations are not retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "23505",
			Message: "duplicate key value",
		}
		assert.False(t, IsRetryableError(pgErr))
	})

	t.Run("Syntax errors are not retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "42601",
			Message: "syntax error",
		}
		assert.False(t, IsRetryableError(pgErr))
	})

	t.Run("Permission errors are not retryable", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    "42501",
			Message: "permission denied",
		}
		assert.False(t, IsRetryableError(pgErr))
	})
}

func TestGetErrorDetails(t *testing.T) {
	t.Run("Nil error returns empty details", func(t *testing.T) {
		details := GetErrorDetails(nil)
		assert.Empty(t, details)
	})

	t.Run("Standard error returns basic details", func(t *testing.T) {
		err := errors.New("test error")
		details := GetErrorDetails(err)

		assert.Contains(t, details["error"], "test error")
		assert.Equal(t, "unknown", details["type"])
	})

	t.Run("PGX ErrNoRows returns specific details", func(t *testing.T) {
		details := GetErrorDetails(pgx.ErrNoRows)

		assert.Equal(t, "no_rows", details["type"])
		assert.Contains(t, details["error"], "no rows")
	})

	t.Run("PostgreSQL error returns detailed information", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Severity:       "ERROR",
			Code:           "23505",
			Message:        "duplicate key value",
			Detail:         "Key (id)=(1) already exists",
			Hint:           "Check for existing records",
			SchemaName:     "public",
			TableName:      "users",
			ColumnName:     "id",
			ConstraintName: "users_pkey",
		}

		details := GetErrorDetails(pgErr)

		assert.Equal(t, "constraint_violation", details["type"])
		assert.Equal(t, "23505", details["code"])
		assert.Equal(t, "ERROR", details["severity"])
		assert.Equal(t, "duplicate key value", details["message"])
		assert.Equal(t, "Key (id)=(1) already exists", details["detail"])
		assert.Equal(t, "Check for existing records", details["hint"])
		assert.Equal(t, "public", details["schema"])
		assert.Equal(t, "users", details["table"])
		assert.Equal(t, "id", details["column"])
		assert.Equal(t, "users_pkey", details["constraint"])
	})
}

// Benchmark tests
func BenchmarkWrapError_StandardError(b *testing.B) {
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(err, "test", "SELECT 1", []interface{}{})
	}
}

func BenchmarkWrapError_PGError(b *testing.B) {
	pgErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(pgErr, "test", "INSERT INTO users", []interface{}{})
	}
}

func BenchmarkIsRetryableError(b *testing.B) {
	pgErr := &pgconn.PgError{
		Code:    "40001",
		Message: "serialization failure",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsRetryableError(pgErr)
	}
}

func BenchmarkGetErrorDetails(b *testing.B) {
	pgErr := &pgconn.PgError{
		Severity:       "ERROR",
		Code:           "23505",
		Message:        "duplicate key value",
		Detail:         "Key (id)=(1) already exists",
		SchemaName:     "public",
		TableName:      "users",
		ConstraintName: "users_pkey",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetErrorDetails(pgErr)
	}
}
