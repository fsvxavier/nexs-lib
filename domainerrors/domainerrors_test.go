package domainerrors

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{
			name:    "valid error creation",
			code:    "TEST_001",
			message: "test error message",
		},
		{
			name:    "empty code",
			code:    "",
			message: "test error message",
		},
		{
			name:    "empty message",
			code:    "TEST_002",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.message)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, err.Message)
			}

			if err.Type != ErrorTypeServer {
				t.Errorf("expected type %q, got %q", ErrorTypeServer, err.Type)
			}

			if err.Cause != nil {
				t.Errorf("expected no cause, got %v", err.Cause)
			}

			if len(err.Stack) == 0 {
				t.Error("expected stack trace, got empty")
			}
		})
	}
}

func TestNewWithCause(t *testing.T) {
	cause := errors.New("original error")

	tests := []struct {
		name    string
		code    string
		message string
		cause   error
	}{
		{
			name:    "valid error with cause",
			code:    "TEST_001",
			message: "test error message",
			cause:   cause,
		},
		{
			name:    "error with nil cause",
			code:    "TEST_002",
			message: "test error message",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewWithCause(tt.code, tt.message, tt.cause)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, err.Message)
			}

			if err.Cause != tt.cause {
				t.Errorf("expected cause %v, got %v", tt.cause, err.Cause)
			}

			if len(err.Stack) == 0 {
				t.Error("expected stack trace, got empty")
			}
		})
	}
}

func TestNewWithType(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		message   string
		errorType ErrorType
	}{
		{
			name:      "validation error type",
			code:      "VAL_001",
			message:   "validation failed",
			errorType: ErrorTypeValidation,
		},
		{
			name:      "not found error type",
			code:      "NF_001",
			message:   "resource not found",
			errorType: ErrorTypeNotFound,
		},
		{
			name:      "business error type",
			code:      "BIZ_001",
			message:   "business rule violated",
			errorType: ErrorTypeBusinessRule,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewWithType(tt.code, tt.message, tt.errorType)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, err.Message)
			}

			if err.Type != tt.errorType {
				t.Errorf("expected type %q, got %q", tt.errorType, err.Type)
			}
		})
	}
}

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		cause    error
		expected string
	}{
		{
			name:     "error with code and message",
			code:     "TEST_001",
			message:  "test error",
			cause:    nil,
			expected: "[TEST_001] test error",
		},
		{
			name:     "error with code, message and cause",
			code:     "TEST_002",
			message:  "test error",
			cause:    errors.New("original error"),
			expected: "[TEST_002] test error: original error",
		},
		{
			name:     "error without code",
			code:     "",
			message:  "test error",
			cause:    nil,
			expected: "test error",
		},
		{
			name:     "error without code but with cause",
			code:     "",
			message:  "test error",
			cause:    errors.New("original error"),
			expected: "test error: original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &DomainError{
				Code:    tt.code,
				Message: tt.message,
				Cause:   tt.cause,
			}

			result := err.Error()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	cause := errors.New("original error")

	tests := []struct {
		name     string
		cause    error
		expected error
	}{
		{
			name:     "unwrap with cause",
			cause:    cause,
			expected: cause,
		},
		{
			name:     "unwrap without cause",
			cause:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &DomainError{
				Code:    "TEST_001",
				Message: "test error",
				Cause:   tt.cause,
			}

			result := err.Unwrap()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDomainError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		errorType      ErrorType
		expectedStatus int
	}{
		{
			name:           "validation error",
			errorType:      ErrorTypeValidation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request error",
			errorType:      ErrorTypeBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid schema error",
			errorType:      ErrorTypeInvalidSchema,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "authentication error",
			errorType:      ErrorTypeAuthentication,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "authorization error",
			errorType:      ErrorTypeAuthorization,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "security error",
			errorType:      ErrorTypeSecurity,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "not found error",
			errorType:      ErrorTypeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "conflict error",
			errorType:      ErrorTypeConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "timeout error",
			errorType:      ErrorTypeTimeout,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name:           "unsupported media error",
			errorType:      ErrorTypeUnsupportedMedia,
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "business rule error",
			errorType:      ErrorTypeBusinessRule,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "unprocessable error",
			errorType:      ErrorTypeUnprocessable,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "rate limit error",
			errorType:      ErrorTypeRateLimit,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "unsupported error",
			errorType:      ErrorTypeUnsupported,
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "service unavailable error",
			errorType:      ErrorTypeServiceUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "circuit breaker error",
			errorType:      ErrorTypeCircuitBreaker,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "resource exhausted error",
			errorType:      ErrorTypeResourceExhausted,
			expectedStatus: http.StatusInsufficientStorage,
		},
		{
			name:           "dependency error",
			errorType:      ErrorTypeDependency,
			expectedStatus: http.StatusFailedDependency,
		},
		{
			name:           "server error (default)",
			errorType:      ErrorTypeServer,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "unknown error type",
			errorType:      ErrorType("unknown"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &DomainError{
				Code:    "TEST_001",
				Message: "test error",
				Type:    tt.errorType,
			}

			status := err.HTTPStatus()
			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestDomainError_WithMetadata(t *testing.T) {
	err := New("TEST_001", "test error")

	// Test adding metadata
	result := err.WithMetadata("key1", "value1")
	if result != err {
		t.Error("expected same instance")
	}

	if err.Metadata["key1"] != "value1" {
		t.Errorf("expected metadata key1 to be 'value1', got %v", err.Metadata["key1"])
	}

	// Test adding multiple metadata
	err.WithMetadata("key2", 42)
	err.WithMetadata("key3", true)

	if err.Metadata["key2"] != 42 {
		t.Errorf("expected metadata key2 to be 42, got %v", err.Metadata["key2"])
	}

	if err.Metadata["key3"] != true {
		t.Errorf("expected metadata key3 to be true, got %v", err.Metadata["key3"])
	}
}

func TestDomainError_WithContext(t *testing.T) {
	ctx := context.Background()
	err := New("TEST_001", "test error")

	initialStackSize := len(err.Stack)

	result := err.WithContext(ctx, "additional context")
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.Stack) != initialStackSize+1 {
		t.Errorf("expected stack size %d, got %d", initialStackSize+1, len(err.Stack))
	}

	lastFrame := err.Stack[len(err.Stack)-1]
	if lastFrame.Message != "additional context" {
		t.Errorf("expected last frame message 'additional context', got %q", lastFrame.Message)
	}
}

func TestDomainError_Wrap(t *testing.T) {
	cause := errors.New("original error")
	err := New("TEST_001", "test error")

	initialStackSize := len(err.Stack)

	result := err.Wrap("wrapping context", cause)
	if result != err {
		t.Error("expected same instance")
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}

	if len(err.Stack) != initialStackSize+1 {
		t.Errorf("expected stack size %d, got %d", initialStackSize+1, len(err.Stack))
	}

	lastFrame := err.Stack[len(err.Stack)-1]
	if lastFrame.Message != "wrapping context" {
		t.Errorf("expected last frame message 'wrapping context', got %q", lastFrame.Message)
	}
}

func TestDomainError_StackTrace(t *testing.T) {
	err := New("TEST_001", "test error")

	stackTrace := err.StackTrace()
	if stackTrace == "" {
		t.Error("expected non-empty stack trace")
	}

	if !strings.Contains(stackTrace, "Stack trace:") {
		t.Error("expected stack trace to contain 'Stack trace:'")
	}

	// Test with empty stack
	err.Stack = nil
	stackTrace = err.StackTrace()
	if stackTrace != "" {
		t.Error("expected empty stack trace")
	}
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "matching type",
			err:       NewWithType("TEST_001", "test error", ErrorTypeValidation),
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "non-matching type",
			err:       NewWithType("TEST_002", "test error", ErrorTypeValidation),
			errorType: ErrorTypeNotFound,
			expected:  false,
		},
		{
			name:      "non-domain error",
			err:       errors.New("regular error"),
			errorType: ErrorTypeValidation,
			expected:  false,
		},
		{
			name:      "nil error",
			err:       nil,
			errorType: ErrorTypeValidation,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsType(tt.err, tt.errorType)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMapHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "domain error",
			err:            NewWithType("TEST_001", "test error", ErrorTypeValidation),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-domain error",
			err:            errors.New("regular error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := MapHTTPStatus(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")

	err := Wrap("WRAP_001", "wrapped error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "WRAP_001" {
		t.Errorf("expected code 'WRAP_001', got %q", err.Code)
	}

	if err.Message != "wrapped error" {
		t.Errorf("expected message 'wrapped error', got %q", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}

	if err.Type != ErrorTypeServer {
		t.Errorf("expected type %q, got %q", ErrorTypeServer, err.Type)
	}

	if len(err.Stack) == 0 {
		t.Error("expected stack trace, got empty")
	}
}

func TestWrapWithContext(t *testing.T) {
	ctx := context.Background()
	cause := errors.New("original error")

	err := WrapWithContext(ctx, "WRAP_001", "wrapped error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "WRAP_001" {
		t.Errorf("expected code 'WRAP_001', got %q", err.Code)
	}

	if err.Message != "wrapped error" {
		t.Errorf("expected message 'wrapped error', got %q", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}

	if err.Type != ErrorTypeServer {
		t.Errorf("expected type %q, got %q", ErrorTypeServer, err.Type)
	}

	if len(err.Stack) == 0 {
		t.Error("expected stack trace, got empty")
	}
}

func TestDomainError_Timestamp(t *testing.T) {
	before := time.Now()
	err := New("TEST_001", "test error")
	after := time.Now()

	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Error("timestamp should be between before and after")
	}
}

func TestDomainError_captureStackFrame(t *testing.T) {
	err := &DomainError{
		Code:    "TEST_001",
		Message: "test error",
		Type:    ErrorTypeServer,
	}

	// Test initial empty stack
	if len(err.Stack) != 0 {
		t.Error("expected empty stack initially")
	}

	// Test adding stack frame
	err.captureStackFrame("test message")

	if len(err.Stack) != 1 {
		t.Errorf("expected stack size 1, got %d", len(err.Stack))
	}

	frame := err.Stack[0]
	if frame.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", frame.Message)
	}

	if frame.Function == "" {
		t.Error("expected function name to be set")
	}

	if frame.File == "" {
		t.Error("expected file name to be set")
	}

	if frame.Line == 0 {
		t.Error("expected line number to be set")
	}

	if frame.Time == "" {
		t.Error("expected time to be set")
	}
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("BENCH_001", "benchmark error")
	}
}

func BenchmarkNewWithCause(b *testing.B) {
	cause := errors.New("original error")
	for i := 0; i < b.N; i++ {
		_ = NewWithCause("BENCH_001", "benchmark error", cause)
	}
}

func BenchmarkError(b *testing.B) {
	err := New("BENCH_001", "benchmark error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkWithMetadata(b *testing.B) {
	err := New("BENCH_001", "benchmark error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err.WithMetadata("key", "value")
	}
}

func BenchmarkHTTPStatus(b *testing.B) {
	err := NewWithType("BENCH_001", "benchmark error", ErrorTypeValidation)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.HTTPStatus()
	}
}

func BenchmarkStackTrace(b *testing.B) {
	err := New("BENCH_001", "benchmark error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.StackTrace()
	}
}
