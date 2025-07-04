package domainerrors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSQLErrorParser2(t *testing.T) {
	parser := NewSQLErrorParser()

	tests := []struct {
		name     string
		err      error
		wantCode string
		wantOk   bool
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "non-matching error",
			err:      errors.New("some random error"),
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "matching SQL error",
			err:      errors.New("error description(SQLSTATE 23505) details"),
			wantCode: "23505",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := parser.Parse(tt.err)
			if code != tt.wantCode {
				t.Errorf("Parse() code = %v, wantCode %v", code, tt.wantCode)
			}
			if ok != tt.wantOk {
				t.Errorf("Parse() ok = %v, wantOk %v", ok, tt.wantOk)
			}
		})
	}
}

func TestErrorCodeRegistry2(t *testing.T) {
	registry := NewErrorCodeRegistry()
	registry.Register("CODE1", "Error 1", 400)
	registry.Register("CODE2", "Error 2", 404)

	t.Run("Get existing code", func(t *testing.T) {
		info, ok := registry.Get("CODE1")
		if !ok {
			t.Errorf("Get() ok = false, want true")
		}
		if info.Code != "CODE1" || info.Description != "Error 1" || info.StatusCode != 400 {
			t.Errorf("Get() info = %v, want {CODE1, Error 1, 400}", info)
		}
	})

	t.Run("Get non-existing code", func(t *testing.T) {
		_, ok := registry.Get("NONEXISTENT")
		if ok {
			t.Errorf("Get() ok = true, want false")
		}
	})

	t.Run("WrapWithCode with existing code", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := registry.WrapWithCode("CODE2", originalErr)

		// Check if the error message contains the code and description
		errMsg := wrappedErr.Error()
		if !strings.Contains(errMsg, "CODE2") || !strings.Contains(errMsg, "Error 2") {
			t.Errorf("WrapWithCode() error = %v, should contain code and description", errMsg)
		}
	})

	t.Run("WrapWithCode with non-existing code", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := registry.WrapWithCode("NONEXISTENT", originalErr)

		// Check if the error message contains the code and "Unknown error"
		errMsg := wrappedErr.Error()
		if !strings.Contains(errMsg, "NONEXISTENT") || !strings.Contains(errMsg, "Unknown error") {
			t.Errorf("WrapWithCode() error = %v, should contain code and 'Unknown error'", errMsg)
		}
	})

	t.Run("WrapWithCode with nil error", func(t *testing.T) {
		wrappedErr := registry.WrapWithCode("CODE1", nil)
		if wrappedErr != nil {
			t.Errorf("WrapWithCode() error = %v, want nil", wrappedErr)
		}
	})
}

func TestErrorStack2(t *testing.T) {
	stack := NewErrorStack()

	t.Run("Initial state", func(t *testing.T) {
		if !stack.IsEmpty() {
			t.Error("NewErrorStack() should create empty stack")
		}
		if stack.Error() != "" {
			t.Errorf("Empty stack Error() = %q, want \"\"", stack.Error())
		}
		if stack.Unwrap() != nil {
			t.Error("Empty stack Unwrap() should return nil")
		}
		if len(stack.ToSlice()) != 0 {
			t.Errorf("Empty stack ToSlice() length = %d, want 0", len(stack.ToSlice()))
		}
		if stack.Format() != "No errors" {
			t.Errorf("Empty stack Format() = %q, want \"No errors\"", stack.Format())
		}
	})

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	t.Run("Push errors", func(t *testing.T) {
		stack.Push(err1)
		stack.Push(nil) // Should be ignored
		stack.Push(err2)

		if stack.IsEmpty() {
			t.Error("Stack should not be empty after Push()")
		}
		if len(stack.Errors) != 2 {
			t.Errorf("Stack should have 2 errors, got %d", len(stack.Errors))
		}
	})

	t.Run("Error method", func(t *testing.T) {
		expected := "error 1\nerror 2"
		if stack.Error() != expected {
			t.Errorf("Error() = %q, want %q", stack.Error(), expected)
		}
	})

	t.Run("Unwrap method", func(t *testing.T) {
		if stack.Unwrap() != err1 {
			t.Error("Unwrap() should return the first error")
		}
	})

	t.Run("ToSlice method", func(t *testing.T) {
		slice := stack.ToSlice()
		if len(slice) != 2 || slice[0] != err1 || slice[1] != err2 {
			t.Error("ToSlice() returned incorrect slice")
		}
	})

	t.Run("Format method", func(t *testing.T) {
		formatted := stack.Format()
		if !strings.Contains(formatted, "Error stack:") ||
			!strings.Contains(formatted, "[1] error 1") ||
			!strings.Contains(formatted, "[2] error 2") {
			t.Errorf("Format() = %q, missing expected content", formatted)
		}
	})
}

func TestRecoverHandler(t *testing.T) {
	t.Run("Recover from string panic", func(t *testing.T) {
		err := RecoverHandler("panic message")
		if err == nil {
			t.Fatal("RecoverHandler() returned nil error")
		}
		if !strings.Contains(err.Error(), "panic message") {
			t.Errorf("RecoverHandler() error = %v, should contain panic message", err.Error())
		}
	})

	t.Run("Recover from error panic", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := RecoverHandler(originalErr)
		if err == nil {
			t.Fatal("RecoverHandler() returned nil error")
		}
		if !strings.Contains(err.Error(), "Recovered from panic") {
			t.Errorf("RecoverHandler() error = %v, should contain recovery message", err.Error())
		}
	})

	t.Run("Recover from other panic", func(t *testing.T) {
		err := RecoverHandler(123)
		if err == nil {
			t.Fatal("RecoverHandler() returned nil error")
		}
		if !strings.Contains(err.Error(), "Recovered from panic: 123") {
			t.Errorf("RecoverHandler() error = %v, should contain recovery message with value", err.Error())
		}
	})
}

func TestRecoverMiddleware(t *testing.T) {
	t.Run("No panic", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		err := RecoverMiddleware(func() error {
			return expectedErr
		})
		if err != expectedErr {
			t.Errorf("RecoverMiddleware() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("With panic", func(t *testing.T) {
		err := RecoverMiddleware(func() error {
			panic("middleware panic")
		})
		if err == nil {
			t.Fatal("RecoverMiddleware() returned nil error after panic")
		}
		if !strings.Contains(err.Error(), "middleware panic") {
			t.Errorf("RecoverMiddleware() error = %v, should contain panic message", err.Error())
		}
	})
}

func TestFormatErrorChain(t *testing.T) {
	t.Run("Nil error", func(t *testing.T) {
		result := FormatErrorChain(nil)
		if result != "no error" {
			t.Errorf("FormatErrorChain(nil) = %q, want \"no error\"", result)
		}
	})

	t.Run("Single error", func(t *testing.T) {
		err := errors.New("single error")
		result := FormatErrorChain(err)
		if result != "single error" {
			t.Errorf("FormatErrorChain() = %q, want %q", result, "single error")
		}
	})

	t.Run("Wrapped errors", func(t *testing.T) {
		err1 := errors.New("base error")
		err2 := fmt.Errorf("intermediate error: %w", err1)
		err3 := fmt.Errorf("top error: %w", err2)

		result := FormatErrorChain(err3)
		expected := "top error: intermediate error: base error\n  └─ intermediate error: base error\n  └─ base error"
		if result != expected {
			t.Errorf("FormatErrorChain() = %q, want %q", result, expected)
		}
	})
}

func TestGetErrorCode(t *testing.T) {
	t.Run("Regular error", func(t *testing.T) {
		err := errors.New("not a domain error")
		code := GetErrorCode(err)
		if code != "" {
			t.Errorf("GetErrorCode() = %q, want \"\"", code)
		}
	})

	t.Run("Domain error", func(t *testing.T) {
		err := New("TEST_CODE", "test error")
		code := GetErrorCode(err)
		if code != "TEST_CODE" {
			t.Errorf("GetErrorCode() = %q, want %q", code, "TEST_CODE")
		}
	})

	t.Run("Wrapped domain error", func(t *testing.T) {
		domainErr := New("WRAPPED_CODE", "wrapped error")
		err := fmt.Errorf("outer error: %w", domainErr)
		code := GetErrorCode(err)
		if code != "WRAPPED_CODE" {
			t.Errorf("GetErrorCode() = %q, want %q", code, "WRAPPED_CODE")
		}
	})
}

func TestErrorTypeChecks(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFoundError with not found error",
			err:      NewNotFoundError("NOT_FOUND"),
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "IsNotFoundError with other error",
			err:      New("OTHER", "Other error").WithType(ErrorTypeBusinessRule),
			checkFn:  IsNotFoundError,
			expected: false,
		},
		{
			name:     "IsValidationError with validation error",
			err:      NewValidationError("INVALID", map[string][]string{"field": {"Validation failed"}}),
			checkFn:  IsValidationError,
			expected: true,
		},
		{
			name:     "IsBusinessError with business error",
			err:      NewBusinessError("BIZ", "Business rule violation"),
			checkFn:  IsBusinessError,
			expected: true,
		},
		{
			name:     "IsAuthenticationError with auth error",
			err:      NewAuthenticationError("AUTH"),
			checkFn:  IsAuthenticationError,
			expected: true,
		},
		{
			name:     "IsAuthorizationError with authz error",
			err:      NewAuthorizationError("AUTHZ"),
			checkFn:  IsAuthorizationError,
			expected: true,
		},
		{
			name:     "IsExternalServiceError with external service error",
			err:      NewExternalServiceError("EXT", "External service failed", errors.New("underlying error")),
			checkFn:  IsExternalServiceError,
			expected: true,
		},
		{
			name:     "IsConflictError with conflict error",
			err:      NewConflictError("CONFLICT"),
			checkFn:  IsConflictError,
			expected: true,
		},
		{
			name:     "Wrapped error type check",
			err:      fmt.Errorf("wrapped: %w", NewValidationError("VAL", map[string][]string{"field": {"Validation error"}})),
			checkFn:  IsValidationError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			if result != tt.expected {
				t.Errorf("Error type check returned %v, want %v", result, tt.expected)
			}
		})
	}
}
func TestIsErrorType(t *testing.T) {
	// Test direct DomainError
	t.Run("Direct DomainError", func(t *testing.T) {
		domainErr := New("TEST", "Test error").WithType(ErrorTypeBusinessRule)

		// Should match correct type
		if !IsErrorType(domainErr, ErrorTypeBusinessRule) {
			t.Error("IsErrorType should return true for matching error type")
		}

		// Should not match incorrect type
		if IsErrorType(domainErr, ErrorTypeValidation) {
			t.Error("IsErrorType should return false for non-matching error type")
		}
	})

	// Test specific error types
	testCases := []struct {
		name      string
		err       error
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "ValidationError with correct type",
			err:       NewValidationError("VALIDATION", map[string][]string{"field": {"error"}}),
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "ValidationError with incorrect type",
			err:       NewValidationError("VALIDATION", map[string][]string{"field": {"error"}}),
			errorType: ErrorTypeNotFound,
			expected:  false,
		},
		{
			name:      "NotFoundError with correct type",
			err:       NewNotFoundError("NOT_FOUND"),
			errorType: ErrorTypeNotFound,
			expected:  true,
		},
		{
			name:      "BusinessError with correct type",
			err:       NewBusinessError("BUSINESS", "Business rule violation"),
			errorType: ErrorTypeBusinessRule,
			expected:  true,
		},
		{
			name:      "InfrastructureError with correct type",
			err:       NewInfrastructureError("INFRA", "Infrastructure error", nil),
			errorType: ErrorTypeInfrastructure,
			expected:  true,
		},
		{
			name:      "ExternalServiceError with correct type",
			err:       NewExternalServiceError("EXT_SVC", "External service error", nil),
			errorType: ErrorTypeExternalService,
			expected:  true,
		},
		{
			name:      "AuthenticationError with correct type",
			err:       NewAuthenticationError("AUTH"),
			errorType: ErrorTypeAuthentication,
			expected:  true,
		},
		{
			name:      "AuthorizationError with correct type",
			err:       NewAuthorizationError("AUTHZ"),
			errorType: ErrorTypeAuthorization,
			expected:  true,
		},
		{
			name:      "TimeoutError with correct type",
			err:       NewTimeoutError("TIMEOUT", "Operation timed out"),
			errorType: ErrorTypeTimeout,
			expected:  true,
		},
		{
			name:      "UnsupportedOperationError with correct type",
			err:       NewUnsupportedOperationError("UNSUPPORTED", "Operation not supported"),
			errorType: ErrorTypeUnsupported,
			expected:  true,
		},
		{
			name:      "BadRequestError with correct type",
			err:       NewBadRequestError("BAD_REQUEST"),
			errorType: ErrorTypeBadRequest,
			expected:  true,
		},
		{
			name:      "ConflictError with correct type",
			err:       NewConflictError("CONFLICT"),
			errorType: ErrorTypeConflict,
			expected:  true,
		},
		{
			name:      "Standard error",
			err:       errors.New("standard error"),
			errorType: ErrorTypeBusinessRule,
			expected:  false,
		},
		{
			name:      "Wrapped error",
			err:       fmt.Errorf("wrapped: %w", NewValidationError("VAL", map[string][]string{"field": {"error"}})),
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "Nil error",
			err:       nil,
			errorType: ErrorTypeBusinessRule,
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsErrorType(tc.err, tc.errorType)
			if result != tc.expected {
				t.Errorf("IsErrorType() = %v, want %v", result, tc.expected)
			}
		})
	}

	// Test remaining error types
	t.Run("Additional error types", func(t *testing.T) {
		// Test RateLimitError
		rateLimitErr := NewRateLimitError("RATE_LIMIT")
		if !IsErrorType(rateLimitErr, ErrorTypeRateLimit) {
			t.Error("Failed to detect RateLimitError")
		}

		// Test CircuitBreakerError
		circuitErr := NewCircuitBreakerError("CIRCUIT", "Circuit breaker is open")
		if !IsErrorType(circuitErr, ErrorTypeCircuitBreaker) {
			t.Error("Failed to detect CircuitBreakerError")
		}

		// Test ConfigurationError
		configErr := NewConfigurationError("CONFIG")
		if !IsErrorType(configErr, ErrorTypeConfiguration) {
			t.Error("Failed to detect ConfigurationError")
		}

		// Test SecurityError
		securityErr := NewSecurityError("SECURITY")
		if !IsErrorType(securityErr, ErrorTypeSecurity) {
			t.Error("Failed to detect SecurityError")
		}

		// Test ResourceExhaustedError
		resourceErr := NewResourceExhaustedError("RESOURCE", "Resource limit exceeded")
		if !IsErrorType(resourceErr, ErrorTypeResourceExhausted) {
			t.Error("Failed to detect ResourceExhaustedError")
		}

		// Test DependencyError
		depErr := NewDependencyError("DEP", "Dependency failed", nil)
		if !IsErrorType(depErr, ErrorTypeDependency) {
			t.Error("Failed to detect DependencyError")
		}

		// Test SerializationError
		serErr := NewSerializationError("SER", "Serialization failed", nil)
		if !IsErrorType(serErr, ErrorTypeSerialization) {
			t.Error("Failed to detect SerializationError")
		}

		// Test CacheError
		cacheErr := NewCacheError("CACHE", "Cache error", "cache_key", nil)
		if !IsErrorType(cacheErr, ErrorTypeCache) {
			t.Error("Failed to detect CacheError")
		}

		// Test WorkflowError
		workflowErr := NewWorkflowError("WORKFLOW", "Workflow failed", "workflow_id")
		if !IsErrorType(workflowErr, ErrorTypeWorkflow) {
			t.Error("Failed to detect WorkflowError")
		}

		// Test MigrationError
		migrationErr := NewMigrationError("MIGRATION", "Migration failed", "version_1.0", nil)
		if !IsErrorType(migrationErr, ErrorTypeMigration) {
			t.Error("Failed to detect MigrationError")
		}
	})
}
