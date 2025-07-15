package pgx

import (
	"errors"
	"net"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError_Nil(t *testing.T) {
	result := WrapError(nil)
	assert.Nil(t, result)
}

func TestWrapError_PgError(t *testing.T) {
	tests := []struct {
		name         string
		pgErr        *pgconn.PgError
		expectedType ErrorType
	}{
		{
			name: "unique violation",
			pgErr: &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint",
				Detail:  "Key (email)=(test@example.com) already exists.",
			},
			expectedType: ErrorTypeUniqueViolation,
		},
		{
			name: "foreign key violation",
			pgErr: &pgconn.PgError{
				Code:    "23503",
				Message: "insert or update on table violates foreign key constraint",
				Detail:  "Key (user_id)=(999) is not present in table \"users\".",
			},
			expectedType: ErrorTypeForeignKeyViolation,
		},
		{
			name: "not null violation",
			pgErr: &pgconn.PgError{
				Code:    "23502",
				Message: "null value in column violates not-null constraint",
				Detail:  "Failing column: email",
			},
			expectedType: ErrorTypeNotNullViolation,
		},
		{
			name: "check violation",
			pgErr: &pgconn.PgError{
				Code:    "23514",
				Message: "new row for relation violates check constraint",
				Detail:  "Failing row contains (age = -5).",
			},
			expectedType: ErrorTypeCheckViolation,
		},
		{
			name: "syntax error",
			pgErr: &pgconn.PgError{
				Code:    "42601",
				Message: "syntax error at or near \"SELCT\"",
			},
			expectedType: ErrorTypeSyntaxError,
		},
		{
			name: "undefined table",
			pgErr: &pgconn.PgError{
				Code:    "42P01",
				Message: "relation \"nonexistent_table\" does not exist",
			},
			expectedType: ErrorTypeUndefinedTable,
		},
		{
			name: "undefined column",
			pgErr: &pgconn.PgError{
				Code:    "42703",
				Message: "column \"nonexistent_column\" does not exist",
			},
			expectedType: ErrorTypeUndefinedColumn,
		},
		{
			name: "undefined function",
			pgErr: &pgconn.PgError{
				Code:    "42883",
				Message: "function nonexistent_function() does not exist",
			},
			expectedType: ErrorTypeUndefinedFunction,
		},
		{
			name: "division by zero",
			pgErr: &pgconn.PgError{
				Code:    "22012",
				Message: "division by zero",
			},
			expectedType: ErrorTypeDivisionByZero,
		},
		{
			name: "connection failure",
			pgErr: &pgconn.PgError{
				Code:    "08000",
				Message: "connection exception",
			},
			expectedType: ErrorTypeConnectionFailed,
		},
		{
			name: "authentication failure",
			pgErr: &pgconn.PgError{
				Code:    "08004",
				Message: "password authentication failed",
			},
			expectedType: ErrorTypeAuthenticationFail,
		},
		{
			name: "serialization failure",
			pgErr: &pgconn.PgError{
				Code:    "40001",
				Message: "could not serialize access due to concurrent update",
			},
			expectedType: ErrorTypeSerializationFailure,
		},
		{
			name: "deadlock detected",
			pgErr: &pgconn.PgError{
				Code:    "40P01",
				Message: "deadlock detected",
			},
			expectedType: ErrorTypeDeadlockDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.pgErr)
			require.NotNil(t, wrapped)

			var dbErr *DatabaseError
			require.True(t, errors.As(wrapped, &dbErr))
			assert.Equal(t, tt.expectedType, dbErr.Type)
			assert.Equal(t, tt.pgErr.Code, dbErr.SQLState)
			assert.Equal(t, tt.pgErr.Message, dbErr.Message)
			assert.Equal(t, tt.pgErr.Detail, dbErr.Detail)
			assert.Equal(t, tt.pgErr, dbErr.OriginalErr)
		})
	}
}

func TestWrapError_PgxSpecificErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
	}{
		{
			name:         "no rows",
			err:          pgx.ErrNoRows,
			expectedType: ErrorTypeUnknown,
		},
		{
			name:         "transaction closed",
			err:          pgx.ErrTxClosed,
			expectedType: ErrorTypeInvalidTransactionState,
		},
		{
			name:         "transaction commit rollback",
			err:          pgx.ErrTxCommitRollback,
			expectedType: ErrorTypeTransactionRollback,
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
				Message: "syntax error at or near \"SELCT\"",
			},
			expected: "syntax_error: syntax error at or near \"SELCT\"",
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

func TestWrapError_UnknownError(t *testing.T) {
	err := errors.New("unknown error")
	wrapped := WrapError(err)
	require.NotNil(t, wrapped)

	var dbErr *DatabaseError
	require.True(t, errors.As(wrapped, &dbErr))
	assert.Equal(t, ErrorTypeUnknown, dbErr.Type)
	assert.Equal(t, err, dbErr.OriginalErr)
	assert.Equal(t, "unknown error", dbErr.Message)
}

func TestWrapPgError_AllSQLStates(t *testing.T) {
	tests := []struct {
		code         string
		expectedType ErrorType
	}{
		// Test all SQL state codes for comprehensive coverage
		{"08003", ErrorTypeConnectionFailed},
		{"08006", ErrorTypeConnectionFailed},
		{"08001", ErrorTypeConnectionRefused},
		{"42804", ErrorTypeDataTypeMismatch},
		{"22001", ErrorTypeStringDataRightTruncation},
		{"22003", ErrorTypeNumericValueOutOfRange},
		{"22P02", ErrorTypeInvalidTextRepresentation},
		{"22008", ErrorTypeInvalidDatetimeFormat},
		{"25P02", ErrorTypeTransactionAborted},
		{"25000", ErrorTypeInvalidTransactionState},
		{"25001", ErrorTypeInvalidTransactionState},
		{"25002", ErrorTypeInvalidTransactionState},
		{"25008", ErrorTypeInvalidTransactionState},
		{"53100", ErrorTypeDiskFull},
		{"53200", ErrorTypeInsufficientMemory},
		{"99999", ErrorTypeUnknown}, // Unknown code
	}

	for _, tt := range tests {
		t.Run("code_"+tt.code, func(t *testing.T) {
			pgErr := &pgconn.PgError{
				Code:    tt.code,
				Message: "test message",
			}
			wrapped := wrapPgError(pgErr)
			assert.Equal(t, tt.expectedType, wrapped.Type)
		})
	}
}

func TestWrapNetError_Coverage(t *testing.T) {
	timeoutErr := &timeoutError{}
	nonTimeoutErr := &nonTimeoutError{}

	// Test timeout error
	wrapped1 := wrapNetError(timeoutErr)
	assert.Equal(t, ErrorTypeConnectionTimeout, wrapped1.Type)

	// Test non-timeout error
	wrapped2 := wrapNetError(nonTimeoutErr)
	assert.Equal(t, ErrorTypeConnectionFailed, wrapped2.Type)
}

func TestWrapSyscallError_Coverage(t *testing.T) {
	tests := []struct {
		errno        syscall.Errno
		expectedType ErrorType
	}{
		{syscall.ECONNREFUSED, ErrorTypeConnectionRefused},
		{syscall.ETIMEDOUT, ErrorTypeConnectionTimeout},
		{syscall.ECONNRESET, ErrorTypeConnectionLost},
		{syscall.EPIPE, ErrorTypeConnectionLost},
		{syscall.EACCES, ErrorTypeSystemError},
	}

	for _, tt := range tests {
		t.Run(tt.errno.Error(), func(t *testing.T) {
			err := errors.New("syscall error")
			wrapped := wrapSyscallError(tt.errno, err)
			assert.Equal(t, tt.expectedType, wrapped.Type)
			assert.Equal(t, err, wrapped.OriginalErr)
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
func BenchmarkWrapError_PgError(b *testing.B) {
	pgErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value violates unique constraint",
		Detail:  "Key (email)=(test@example.com) already exists.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(pgErr)
	}
}

func BenchmarkWrapError_NetworkError(b *testing.B) {
	netErr := &timeoutError{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(netErr)
	}
}

func BenchmarkIsConnectionError(b *testing.B) {
	err := &DatabaseError{Type: ErrorTypeConnectionFailed}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsConnectionError(err)
	}
}
