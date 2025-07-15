package gorm

import (
	"errors"
	"net"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWrapError_Nil(t *testing.T) {
	result := WrapError(nil)
	assert.Nil(t, result)
}

func TestWrapError_GormSpecificErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
	}{
		{
			name:         "record not found",
			err:          gorm.ErrRecordNotFound,
			expectedType: ErrorTypeRecordNotFound,
		},
		{
			name:         "invalid transaction",
			err:          gorm.ErrInvalidTransaction,
			expectedType: ErrorTypeInvalidTransaction,
		},
		{
			name:         "invalid data",
			err:          gorm.ErrInvalidData,
			expectedType: ErrorTypeInvalidData,
		},
		{
			name:         "invalid field",
			err:          gorm.ErrInvalidField,
			expectedType: ErrorTypeUndefinedColumn,
		},
		{
			name:         "empty slice",
			err:          gorm.ErrEmptySlice,
			expectedType: ErrorTypeInvalidValue,
		},
		{
			name:         "invalid value",
			err:          gorm.ErrInvalidValue,
			expectedType: ErrorTypeInvalidValue,
		},
		{
			name:         "invalid value length",
			err:          gorm.ErrInvalidValueOfLength,
			expectedType: ErrorTypeStringDataRightTruncation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.err)
			require.NotNil(t, wrapped)

			var dbErr *DatabaseError
			require.True(t, errors.As(wrapped, &dbErr))
			assert.Equal(t, tt.expectedType, dbErr.Type)
			assert.Equal(t, tt.err, dbErr.OriginalErr)
		})
	}
}

func TestWrapError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          net.Error
		expectedType ErrorType
	}{
		{
			name:         "timeout error",
			err:          &timeoutError{},
			expectedType: ErrorTypeConnectionTimeout,
		},
		{
			name:         "non-timeout network error",
			err:          &nonTimeoutError{},
			expectedType: ErrorTypeConnectionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.err)
			require.NotNil(t, wrapped)

			var dbErr *DatabaseError
			require.True(t, errors.As(wrapped, &dbErr))
			assert.Equal(t, tt.expectedType, dbErr.Type)
			assert.Equal(t, tt.err, dbErr.OriginalErr)
		})
	}
}

func TestWrapError_SyscallErrors(t *testing.T) {
	// Test specific cases by calling wrapSyscallError directly
	// since syscall.Errno doesn't implement error interface properly in tests
	t.Run("connection_refused", func(t *testing.T) {
		err := errors.New("connection refused")
		wrapped := wrapSyscallError(syscall.ECONNREFUSED, err)
		assert.Equal(t, ErrorTypeConnectionRefused, wrapped.Type)
	})

	t.Run("timeout", func(t *testing.T) {
		err := errors.New("timeout")
		wrapped := wrapSyscallError(syscall.ETIMEDOUT, err)
		assert.Equal(t, ErrorTypeConnectionTimeout, wrapped.Type)
	})

	t.Run("connection_reset", func(t *testing.T) {
		err := errors.New("connection reset")
		wrapped := wrapSyscallError(syscall.ECONNRESET, err)
		assert.Equal(t, ErrorTypeConnectionLost, wrapped.Type)
	})

	t.Run("broken_pipe", func(t *testing.T) {
		err := errors.New("broken pipe")
		wrapped := wrapSyscallError(syscall.EPIPE, err)
		assert.Equal(t, ErrorTypeConnectionLost, wrapped.Type)
	})

	t.Run("other_syscall_error", func(t *testing.T) {
		err := errors.New("permission denied")
		wrapped := wrapSyscallError(syscall.EACCES, err)
		assert.Equal(t, ErrorTypeSystemError, wrapped.Type)
	})
}

func TestWrapByMessage(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectedType ErrorType
	}{
		// Connection errors
		{
			name:         "connection refused",
			message:      "connection refused",
			expectedType: ErrorTypeConnectionRefused,
		},
		{
			name:         "connection timeout",
			message:      "connection timeout exceeded",
			expectedType: ErrorTypeConnectionTimeout,
		},
		{
			name:         "timeout",
			message:      "timeout occurred",
			expectedType: ErrorTypeConnectionTimeout,
		},
		{
			name:         "connection reset",
			message:      "connection reset by peer",
			expectedType: ErrorTypeConnectionLost,
		},
		{
			name:         "connection lost",
			message:      "connection lost unexpectedly",
			expectedType: ErrorTypeConnectionLost,
		},
		{
			name:         "authentication failed",
			message:      "password authentication failed for user",
			expectedType: ErrorTypeAuthenticationFail,
		},
		{
			name:         "too many connections",
			message:      "too many connections for role",
			expectedType: ErrorTypePoolExhausted,
		},
		{
			name:         "pool exhausted",
			message:      "connection pool exhausted",
			expectedType: ErrorTypePoolExhausted,
		},

		// Syntax and schema errors
		{
			name:         "syntax error",
			message:      "syntax error at or near 'SELECT'",
			expectedType: ErrorTypeSyntaxError,
		},
		{
			name:         "relation does not exist",
			message:      "relation \"users\" does not exist",
			expectedType: ErrorTypeUndefinedTable,
		},
		{
			name:         "column does not exist",
			message:      "column \"name\" does not exist",
			expectedType: ErrorTypeUndefinedColumn,
		},
		{
			name:         "function does not exist",
			message:      "function upper() does not exist",
			expectedType: ErrorTypeUndefinedFunction,
		},

		// Data type errors
		{
			name:         "invalid input syntax",
			message:      "invalid input syntax for integer: \"abc\"",
			expectedType: ErrorTypeInvalidTextRepresentation,
		},
		{
			name:         "division by zero",
			message:      "division by zero",
			expectedType: ErrorTypeDivisionByZero,
		},
		{
			name:         "value too long",
			message:      "value too long for type character varying(10)",
			expectedType: ErrorTypeStringDataRightTruncation,
		},
		{
			name:         "numeric value out of range",
			message:      "numeric value out of range",
			expectedType: ErrorTypeNumericValueOutOfRange,
		},
		{
			name:         "invalid date",
			message:      "invalid date value: \"2023-13-01\"",
			expectedType: ErrorTypeInvalidDatetimeFormat,
		},
		{
			name:         "invalid time",
			message:      "invalid time value: \"25:00:00\"",
			expectedType: ErrorTypeInvalidDatetimeFormat,
		},

		// Constraint violations
		{
			name:         "unique constraint",
			message:      "duplicate key value violates unique constraint",
			expectedType: ErrorTypeUniqueViolation,
		},
		{
			name:         "duplicate key",
			message:      "duplicate key value",
			expectedType: ErrorTypeUniqueViolation,
		},
		{
			name:         "foreign key constraint",
			message:      "insert or update violates foreign key constraint",
			expectedType: ErrorTypeForeignKeyViolation,
		},
		{
			name:         "not null constraint",
			message:      "null value violates not-null constraint",
			expectedType: ErrorTypeNotNullViolation,
		},
		{
			name:         "null value",
			message:      "null value in column \"email\"",
			expectedType: ErrorTypeNotNullViolation,
		},
		{
			name:         "check constraint",
			message:      "new row violates check constraint",
			expectedType: ErrorTypeCheckViolation,
		},

		// Transaction errors
		{
			name:         "deadlock",
			message:      "deadlock detected",
			expectedType: ErrorTypeDeadlockDetected,
		},
		{
			name:         "serialization failure",
			message:      "could not serialize access due to concurrent update",
			expectedType: ErrorTypeSerializationFailure,
		},
		{
			name:         "transaction aborted",
			message:      "current transaction is aborted",
			expectedType: ErrorTypeTransactionAborted,
		},
		{
			name:         "transaction rolled back",
			message:      "transaction has been rolled back",
			expectedType: ErrorTypeTransactionAborted,
		},
		{
			name:         "invalid transaction",
			message:      "invalid transaction state",
			expectedType: ErrorTypeInvalidTransactionState,
		},
		{
			name:         "transaction other",
			message:      "transaction error occurred",
			expectedType: ErrorTypeTransactionRollback,
		},

		// System errors
		{
			name:         "disk full",
			message:      "disk full error",
			expectedType: ErrorTypeDiskFull,
		},
		{
			name:         "no space left",
			message:      "no space left on device",
			expectedType: ErrorTypeDiskFull,
		},
		{
			name:         "out of memory",
			message:      "out of memory",
			expectedType: ErrorTypeInsufficientMemory,
		},

		// Unknown
		{
			name:         "unknown error",
			message:      "some random error message",
			expectedType: ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.message)
			wrapped := wrapByMessage(err)
			require.NotNil(t, wrapped)
			assert.Equal(t, tt.expectedType, wrapped.Type)
			assert.Equal(t, err, wrapped.OriginalErr)
		})
	}
}

func TestDatabaseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		dbErr    *DatabaseError
		expected string
	}{
		{
			name: "with detail",
			dbErr: &DatabaseError{
				Type:    ErrorTypeUniqueViolation,
				Message: "duplicate key value",
				Detail:  "Key (email)=(test@example.com) already exists.",
			},
			expected: "unique_violation: duplicate key value (Detail: Key (email)=(test@example.com) already exists.)",
		},
		{
			name: "without detail",
			dbErr: &DatabaseError{
				Type:    ErrorTypeSyntaxError,
				Message: "syntax error at or near \"SELECT\"",
			},
			expected: "syntax_error: syntax error at or near \"SELECT\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dbErr.Error())
		})
	}
}

func TestDatabaseError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	dbErr := &DatabaseError{
		Type:        ErrorTypeUnknown,
		OriginalErr: originalErr,
		Message:     "wrapped error",
	}

	assert.Equal(t, originalErr, dbErr.Unwrap())
}

func TestDatabaseError_Is(t *testing.T) {
	dbErr1 := &DatabaseError{Type: ErrorTypeUniqueViolation}
	dbErr2 := &DatabaseError{Type: ErrorTypeUniqueViolation}
	dbErr3 := &DatabaseError{Type: ErrorTypeSyntaxError}
	otherErr := errors.New("other error")

	assert.True(t, dbErr1.Is(dbErr2))
	assert.False(t, dbErr1.Is(dbErr3))
	assert.False(t, dbErr1.Is(otherErr))
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection failed",
			err:      &DatabaseError{Type: ErrorTypeConnectionFailed},
			expected: true,
		},
		{
			name:     "connection lost",
			err:      &DatabaseError{Type: ErrorTypeConnectionLost},
			expected: true,
		},
		{
			name:     "connection timeout",
			err:      &DatabaseError{Type: ErrorTypeConnectionTimeout},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      &DatabaseError{Type: ErrorTypeConnectionRefused},
			expected: true,
		},
		{
			name:     "pool exhausted",
			err:      &DatabaseError{Type: ErrorTypePoolExhausted},
			expected: true,
		},
		{
			name:     "authentication fail",
			err:      &DatabaseError{Type: ErrorTypeAuthenticationFail},
			expected: true,
		},
		{
			name:     "syntax error",
			err:      &DatabaseError{Type: ErrorTypeSyntaxError},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsConnectionError(tt.err))
		})
	}
}

func TestIsConstraintViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unique violation",
			err:      &DatabaseError{Type: ErrorTypeUniqueViolation},
			expected: true,
		},
		{
			name:     "foreign key violation",
			err:      &DatabaseError{Type: ErrorTypeForeignKeyViolation},
			expected: true,
		},
		{
			name:     "not null violation",
			err:      &DatabaseError{Type: ErrorTypeNotNullViolation},
			expected: true,
		},
		{
			name:     "check violation",
			err:      &DatabaseError{Type: ErrorTypeCheckViolation},
			expected: true,
		},
		{
			name:     "syntax error",
			err:      &DatabaseError{Type: ErrorTypeSyntaxError},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsConstraintViolation(tt.err))
		})
	}
}

func TestIsTransactionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "transaction rollback",
			err:      &DatabaseError{Type: ErrorTypeTransactionRollback},
			expected: true,
		},
		{
			name:     "serialization failure",
			err:      &DatabaseError{Type: ErrorTypeSerializationFailure},
			expected: true,
		},
		{
			name:     "deadlock detected",
			err:      &DatabaseError{Type: ErrorTypeDeadlockDetected},
			expected: true,
		},
		{
			name:     "transaction aborted",
			err:      &DatabaseError{Type: ErrorTypeTransactionAborted},
			expected: true,
		},
		{
			name:     "invalid transaction state",
			err:      &DatabaseError{Type: ErrorTypeInvalidTransactionState},
			expected: true,
		},
		{
			name:     "invalid transaction",
			err:      &DatabaseError{Type: ErrorTypeInvalidTransaction},
			expected: true,
		},
		{
			name:     "syntax error",
			err:      &DatabaseError{Type: ErrorTypeSyntaxError},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsTransactionError(tt.err))
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection timeout",
			err:      &DatabaseError{Type: ErrorTypeConnectionTimeout},
			expected: true,
		},
		{
			name:     "connection lost",
			err:      &DatabaseError{Type: ErrorTypeConnectionLost},
			expected: true,
		},
		{
			name:     "pool exhausted",
			err:      &DatabaseError{Type: ErrorTypePoolExhausted},
			expected: true,
		},
		{
			name:     "serialization failure",
			err:      &DatabaseError{Type: ErrorTypeSerializationFailure},
			expected: true,
		},
		{
			name:     "deadlock detected",
			err:      &DatabaseError{Type: ErrorTypeDeadlockDetected},
			expected: true,
		},
		{
			name:     "syntax error",
			err:      &DatabaseError{Type: ErrorTypeSyntaxError},
			expected: false,
		},
		{
			name:     "unique violation",
			err:      &DatabaseError{Type: ErrorTypeUniqueViolation},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRetryable(tt.err))
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "record not found",
			err:      &DatabaseError{Type: ErrorTypeRecordNotFound},
			expected: true,
		},
		{
			name:     "gorm.ErrRecordNotFound",
			err:      gorm.ErrRecordNotFound,
			expected: true,
		},
		{
			name:     "syntax error",
			err:      &DatabaseError{Type: ErrorTypeSyntaxError},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

// Helper types for testing network errors
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout error" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

type nonTimeoutError struct{}

func (e *nonTimeoutError) Error() string   { return "network error" }
func (e *nonTimeoutError) Timeout() bool   { return false }
func (e *nonTimeoutError) Temporary() bool { return true }

// Benchmark tests
func BenchmarkWrapError_GormError(b *testing.B) {
	err := gorm.ErrRecordNotFound

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(err)
	}
}

func BenchmarkWrapByMessage(b *testing.B) {
	err := errors.New("duplicate key value violates unique constraint")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapByMessage(err)
	}
}

func BenchmarkIsConnectionError(b *testing.B) {
	err := &DatabaseError{Type: ErrorTypeConnectionFailed}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsConnectionError(err)
	}
}
