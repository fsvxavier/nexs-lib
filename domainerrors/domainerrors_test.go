//go:build unit
// +build unit

package domainerrors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test DomainError basic functionality
func TestDomainError_New(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		want    *DomainError
	}{
		{
			name:    "basic error creation",
			code:    "TEST_001",
			message: "test error message",
			want: &DomainError{
				CodeField: "TEST_001",
				Message:   "test error message",
				ErrorType: ErrorTypeBusiness,
			},
		},
		{
			name:    "empty code",
			code:    "",
			message: "test error message",
			want: &DomainError{
				CodeField: "",
				Message:   "test error message",
				ErrorType: ErrorTypeBusiness,
			},
		},
		{
			name:    "empty message",
			code:    "TEST_001",
			message: "",
			want: &DomainError{
				CodeField: "TEST_001",
				Message:   "",
				ErrorType: ErrorTypeBusiness,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.code, tt.message)
			assert.Equal(t, tt.code, got.Code())
			assert.Equal(t, tt.want.Message, got.Message)
			assert.Equal(t, tt.want.ErrorType, got.ErrorType)
			assert.NotNil(t, got.MetadataMap)
			assert.NotEmpty(t, got.ID)
			assert.NotZero(t, got.Timestamp)
		})
	}
}

func TestDomainError_NewWithError(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		code    string
		message string
		cause   error
	}{
		{
			name:    "with cause error",
			code:    "TEST_001",
			message: "test error message",
			cause:   baseErr,
		},
		{
			name:    "with nil cause",
			code:    "TEST_001",
			message: "test error message",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithError(tt.code, tt.message, tt.cause)
			assert.Equal(t, tt.code, got.Code())
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.Equal(t, ErrorTypeBusiness, got.ErrorType)
		})
	}
}

func TestDomainError_NewWithType(t *testing.T) {
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
			name:      "timeout error type",
			code:      "TO_001",
			message:   "request timeout",
			errorType: ErrorTypeTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithType(tt.code, tt.message, tt.errorType)
			assert.Equal(t, tt.code, got.Code())
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.errorType, got.ErrorType)
		})
	}
}

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *DomainError
		expected string
	}{
		{
			name: "with code and message",
			error: &DomainError{
				CodeField: "TEST_001",
				Message:   "test error",
			},
			expected: "[TEST_001] test error",
		},
		{
			name: "with code, message and cause",
			error: &DomainError{
				CodeField: "TEST_001",
				Message:   "test error",
				Cause:     errors.New("cause error"),
			},
			expected: "[TEST_001] test error: cause error",
		},
		{
			name: "without code",
			error: &DomainError{
				Message: "test error",
			},
			expected: "test error",
		},
		{
			name: "message only with cause",
			error: &DomainError{
				Message: "test error",
				Cause:   errors.New("cause error"),
			},
			expected: "test error: cause error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.error.Error()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		error    *DomainError
		expected error
	}{
		{
			name: "with cause",
			error: &DomainError{
				Cause: baseErr,
			},
			expected: baseErr,
		},
		{
			name: "without cause",
			error: &DomainError{
				Cause: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.error.Unwrap()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDomainError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  int
	}{
		{
			name:      "validation error",
			errorType: ErrorTypeValidation,
			expected:  http.StatusBadRequest,
		},
		{
			name:      "not found error",
			errorType: ErrorTypeNotFound,
			expected:  http.StatusNotFound,
		},
		{
			name:      "authentication error",
			errorType: ErrorTypeAuthentication,
			expected:  http.StatusUnauthorized,
		},
		{
			name:      "authorization error",
			errorType: ErrorTypeAuthorization,
			expected:  http.StatusForbidden,
		},
		{
			name:      "timeout error",
			errorType: ErrorTypeTimeout,
			expected:  http.StatusRequestTimeout,
		},
		{
			name:      "business error",
			errorType: ErrorTypeBusiness,
			expected:  http.StatusUnprocessableEntity,
		},
		{
			name:      "server error",
			errorType: ErrorTypeServer,
			expected:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewWithType("TEST", "test", tt.errorType)
			got := err.HTTPStatus()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDomainError_WithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test", "value")
	err := New("TEST", "test error")

	result := err.WithContext(ctx)
	domainErr, ok := result.(*DomainError)
	require.True(t, ok)

	assert.Equal(t, ctx, domainErr.Context)
}

func TestDomainError_WithMetadata(t *testing.T) {
	err := New("TEST", "test error")

	result := err.WithMetadata("key1", "value1")
	assert.Equal(t, "value1", result.MetadataMap["key1"])

	result = result.WithMetadata("key2", 123)
	assert.Equal(t, 123, result.MetadataMap["key2"])
}

func TestDomainError_WithMetadataMap(t *testing.T) {
	err := New("TEST", "test error")
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	result := err.WithMetadataMap(metadata)
	assert.Equal(t, metadata["key1"], result.MetadataMap["key1"])
	assert.Equal(t, metadata["key2"], result.MetadataMap["key2"])
	assert.Equal(t, metadata["key3"], result.MetadataMap["key3"])
}

func TestDomainError_Wrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := New("TEST", "test error")

	wrapped := err.Wrap("wrapped message", baseErr)
	domainErr, ok := wrapped.(*DomainError)
	require.True(t, ok)

	assert.Equal(t, "wrapped message", domainErr.Message)
	assert.Equal(t, baseErr, domainErr.Cause)
	assert.Equal(t, err.ErrorType, domainErr.ErrorType)
}

func TestDomainError_JSON(t *testing.T) {
	err := New("TEST", "test error").
		WithMetadata("key1", "value1").
		WithType(ErrorTypeValidation)

	data, jsonErr := err.JSON()
	assert.NoError(t, jsonErr)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "TEST")
	assert.Contains(t, string(data), "test error")
	assert.Contains(t, string(data), "validation")
}

// Test ValidationError
func TestValidationError_New(t *testing.T) {
	fields := map[string][]string{
		"email": {"invalid format"},
		"age":   {"must be positive"},
	}

	err := NewValidationError("validation failed", fields)

	assert.Equal(t, "VALIDATION_ERROR", err.Code())
	assert.Equal(t, "validation failed", err.Message)
	assert.Equal(t, ErrorTypeValidation, err.ErrorType)
	assert.Equal(t, fields, err.Fields)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode())
}

func TestValidationError_WithField(t *testing.T) {
	err := NewValidationError("validation failed", nil)

	result := err.WithField("email", "invalid format")
	assert.Contains(t, result.Fields["email"], "invalid format")

	result = result.WithField("email", "required")
	assert.Contains(t, result.Fields["email"], "invalid format")
	assert.Contains(t, result.Fields["email"], "required")
}

// Test NotFoundError
func TestNotFoundError_New(t *testing.T) {
	err := NewNotFoundError("resource not found")

	assert.Equal(t, "NOT_FOUND", err.Code())
	assert.Equal(t, "resource not found", err.Message)
	assert.Equal(t, ErrorTypeNotFound, err.ErrorType)
	assert.Equal(t, http.StatusNotFound, err.StatusCode())
}

func TestNotFoundError_WithResource(t *testing.T) {
	err := NewNotFoundError("resource not found")

	result := err.WithResource("user", "123")
	assert.Equal(t, "user", result.Resource)
	assert.Equal(t, "123", result.ResourceID)
}

// Test BusinessError
func TestBusinessError_New(t *testing.T) {
	err := NewBusinessError("INSUFFICIENT_FUNDS", "insufficient funds for transfer")

	assert.Equal(t, "INSUFFICIENT_FUNDS", err.Code())
	assert.Equal(t, "insufficient funds for transfer", err.Message)
	assert.Equal(t, ErrorTypeBusiness, err.ErrorType)
	assert.Equal(t, "INSUFFICIENT_FUNDS", err.BusinessCode)
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
}

func TestBusinessError_WithRule(t *testing.T) {
	err := NewBusinessError("RULE_VIOLATION", "rule violated")

	result := err.WithRule("minimum_balance")
	assert.Equal(t, "minimum_balance", result.RuleName)
}

// Test DatabaseError
func TestDatabaseError_New(t *testing.T) {
	baseErr := errors.New("connection failed")
	err := NewDatabaseError("database error", baseErr)

	assert.Equal(t, "DATABASE_ERROR", err.Code())
	assert.Equal(t, "database error", err.Message)
	assert.Equal(t, ErrorTypeDatabase, err.ErrorType)
	assert.Equal(t, baseErr, err.Cause)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
}

func TestDatabaseError_WithOperation(t *testing.T) {
	err := NewDatabaseError("database error", nil)

	result := err.WithOperation("SELECT", "users")
	assert.Equal(t, "SELECT", result.Operation)
	assert.Equal(t, "users", result.Table)
}

func TestDatabaseError_WithQuery(t *testing.T) {
	err := NewDatabaseError("database error", nil)

	result := err.WithQuery("SELECT * FROM users")
	assert.Equal(t, "SELECT * FROM users", result.Query)
}

// Test ExternalServiceError
func TestExternalServiceError_New(t *testing.T) {
	baseErr := errors.New("service unavailable")
	err := NewExternalServiceError("payment-service", "external service error", baseErr)

	assert.Equal(t, "EXTERNAL_SERVICE_ERROR", err.Code())
	assert.Equal(t, "external service error", err.Message)
	assert.Equal(t, ErrorTypeExternalService, err.ErrorType)
	assert.Equal(t, "payment-service", err.Service)
	assert.Equal(t, baseErr, err.Cause)
	assert.Equal(t, http.StatusBadGateway, err.StatusCode())
}

func TestExternalServiceError_WithEndpoint(t *testing.T) {
	err := NewExternalServiceError("payment-service", "error", nil)

	result := err.WithEndpoint("/api/v1/payments")
	assert.Equal(t, "/api/v1/payments", result.Endpoint)
}

func TestExternalServiceError_WithStatusCode(t *testing.T) {
	err := NewExternalServiceError("payment-service", "error", nil)

	result := err.WithStatusCode(503)
	assert.Equal(t, 503, result.HTTPStatusCode)
	assert.Equal(t, 503, result.StatusCode())
}

func TestExternalServiceError_WithResponse(t *testing.T) {
	err := NewExternalServiceError("payment-service", "error", nil)

	result := err.WithResponse("service temporarily unavailable")
	assert.Equal(t, "service temporarily unavailable", result.Response)
}

// Test InfrastructureError
func TestInfrastructureError_New(t *testing.T) {
	baseErr := errors.New("network error")
	err := NewInfrastructureError("network", "infrastructure error", baseErr)

	assert.Equal(t, "INFRASTRUCTURE_ERROR", err.Code())
	assert.Equal(t, "infrastructure error", err.Message)
	assert.Equal(t, ErrorTypeInfrastructure, err.ErrorType)
	assert.Equal(t, "network", err.Component)
	assert.Equal(t, baseErr, err.Cause)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
}

func TestInfrastructureError_WithDetails(t *testing.T) {
	err := NewInfrastructureError("network", "error", nil)

	result := err.WithDetails("connection timeout after 30s")
	assert.Equal(t, "connection timeout after 30s", result.Details)
}

// Test AuthenticationError
func TestAuthenticationError_New(t *testing.T) {
	err := NewAuthenticationError("invalid credentials")

	assert.Equal(t, "AUTHENTICATION_ERROR", err.Code())
	assert.Equal(t, "invalid credentials", err.Message)
	assert.Equal(t, ErrorTypeAuthentication, err.ErrorType)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode())
}

func TestAuthenticationError_WithReason(t *testing.T) {
	err := NewAuthenticationError("authentication failed")

	result := err.WithReason("password mismatch")
	assert.Equal(t, "password mismatch", result.Reason)
}

// Test AuthorizationError
func TestAuthorizationError_New(t *testing.T) {
	err := NewAuthorizationError("access denied")

	assert.Equal(t, "AUTHORIZATION_ERROR", err.Code())
	assert.Equal(t, "access denied", err.Message)
	assert.Equal(t, ErrorTypeAuthorization, err.ErrorType)
	assert.Equal(t, http.StatusForbidden, err.StatusCode())
}

func TestAuthorizationError_WithResource(t *testing.T) {
	err := NewAuthorizationError("access denied")

	result := err.WithResource("user", "delete")
	assert.Equal(t, "user", result.Resource)
	assert.Equal(t, "delete", result.Action)
}

func TestAuthorizationError_WithPermission(t *testing.T) {
	err := NewAuthorizationError("access denied")

	result := err.WithPermission("admin")
	assert.Equal(t, "admin", result.Permission)
}

// Test TimeoutError
func TestTimeoutError_New(t *testing.T) {
	duration := 5 * time.Second
	timeout := 10 * time.Second
	err := NewTimeoutError("request timeout", duration, timeout)

	assert.Equal(t, "TIMEOUT_ERROR", err.Code())
	assert.Equal(t, "request timeout", err.Message)
	assert.Equal(t, ErrorTypeTimeout, err.ErrorType)
	assert.Equal(t, duration, err.Duration)
	assert.Equal(t, timeout, err.Timeout)
	assert.Equal(t, http.StatusRequestTimeout, err.StatusCode())
}

// Test ServerError
func TestServerError_New(t *testing.T) {
	baseErr := errors.New("internal error")
	err := NewServerError("server error", baseErr)

	assert.Equal(t, "SERVER_ERROR", err.Code())
	assert.Equal(t, "server error", err.Message)
	assert.Equal(t, ErrorTypeServer, err.ErrorType)
	assert.Equal(t, baseErr, err.Cause)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
}

func TestServerError_WithRequestInfo(t *testing.T) {
	err := NewServerError("server error", nil)

	result := err.WithRequestInfo("req-123", "corr-456")
	assert.Equal(t, "req-123", result.RequestID)
	assert.Equal(t, "corr-456", result.CorrelationID)
}

func TestServerError_WithComponent(t *testing.T) {
	err := NewServerError("server error", nil)

	result := err.WithComponent("payment-processor")
	assert.Equal(t, "payment-processor", result.Component)
}

// Test UnprocessableEntityError
func TestUnprocessableEntityError_New(t *testing.T) {
	err := NewUnprocessableEntityError("entity cannot be processed")

	assert.Equal(t, "UNPROCESSABLE_ENTITY", err.Code())
	assert.Equal(t, "entity cannot be processed", err.Message)
	assert.Equal(t, ErrorTypeUnprocessable, err.ErrorType)
	assert.NotNil(t, err.Fields)
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
}

func TestUnprocessableEntityError_WithEntity(t *testing.T) {
	err := NewUnprocessableEntityError("entity error")

	result := err.WithEntity("User", "123")
	assert.Equal(t, "User", result.EntityType)
	assert.Equal(t, "123", result.EntityID)
}

func TestUnprocessableEntityError_WithViolation(t *testing.T) {
	err := NewUnprocessableEntityError("entity error")

	result := err.WithViolation("business rule violated")
	assert.Contains(t, result.Violations, "business rule violated")
}

func TestUnprocessableEntityError_WithFieldError(t *testing.T) {
	err := NewUnprocessableEntityError("entity error")

	result := err.WithFieldError("email", "invalid format")
	assert.Contains(t, result.Fields["email"], "invalid format")
}

// Test utility functions
func TestIsType(t *testing.T) {
	tests := []struct {
		name      string
		error     error
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "validation error matches",
			error:     NewValidationError("validation failed", nil),
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "validation error doesn't match business",
			error:     NewValidationError("validation failed", nil),
			errorType: ErrorTypeBusiness,
			expected:  false,
		},
		{
			name:      "not found error matches",
			error:     NewNotFoundError("not found"),
			errorType: ErrorTypeNotFound,
			expected:  true,
		},
		{
			name:      "business error matches",
			error:     NewBusinessError("RULE", "rule violated"),
			errorType: ErrorTypeBusiness,
			expected:  true,
		},
		{
			name:      "database error matches",
			error:     NewDatabaseError("db error", nil),
			errorType: ErrorTypeDatabase,
			expected:  true,
		},
		{
			name:      "nil error",
			error:     nil,
			errorType: ErrorTypeValidation,
			expected:  false,
		},
		{
			name:      "regular error",
			error:     errors.New("regular error"),
			errorType: ErrorTypeValidation,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsType(tt.error, tt.errorType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMapHTTPStatus(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  int
	}{
		{
			name:      "validation error",
			errorType: ErrorTypeValidation,
			expected:  http.StatusBadRequest,
		},
		{
			name:      "not found error",
			errorType: ErrorTypeNotFound,
			expected:  http.StatusNotFound,
		},
		{
			name:      "authentication error",
			errorType: ErrorTypeAuthentication,
			expected:  http.StatusUnauthorized,
		},
		{
			name:      "authorization error",
			errorType: ErrorTypeAuthorization,
			expected:  http.StatusForbidden,
		},
		{
			name:      "timeout error",
			errorType: ErrorTypeTimeout,
			expected:  http.StatusRequestTimeout,
		},
		{
			name:      "business error",
			errorType: ErrorTypeBusiness,
			expected:  http.StatusUnprocessableEntity,
		},
		{
			name:      "unknown error type",
			errorType: ErrorType("unknown"),
			expected:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapHTTPStatus(tt.errorType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		expected int
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: http.StatusOK,
		},
		{
			name:     "validation error",
			error:    NewValidationError("validation failed", nil),
			expected: http.StatusBadRequest,
		},
		{
			name:     "not found error",
			error:    NewNotFoundError("not found"),
			expected: http.StatusNotFound,
		},
		{
			name:     "business error",
			error:    NewBusinessError("RULE", "rule violated"),
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "domain error",
			error:    NewWithType("TEST", "test", ErrorTypeTimeout),
			expected: http.StatusRequestTimeout,
		},
		{
			name:     "regular error",
			error:    errors.New("regular error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHTTPStatus(tt.error)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := Wrap("wrapped message", baseErr)

	assert.Equal(t, "WRAPPED_ERROR", wrapped.Code())
	assert.Equal(t, "wrapped message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ErrorTypeBusiness, wrapped.ErrorType)
}

func TestWrapWithType(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := WrapWithType("wrapped message", baseErr, ErrorTypeTimeout)

	assert.Equal(t, "WRAPPED_ERROR", wrapped.Code())
	assert.Equal(t, "wrapped message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ErrorTypeTimeout, wrapped.ErrorType)
}

// Test ErrorGroup
func TestErrorGroup_New(t *testing.T) {
	eg := NewErrorGroup()

	assert.NotNil(t, eg)
	assert.Empty(t, eg.Errors)
	assert.False(t, eg.HasErrors())
}

func TestErrorGroup_Add(t *testing.T) {
	eg := NewErrorGroup()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	eg.Add(err1)
	assert.True(t, eg.HasErrors())
	assert.Equal(t, 1, eg.Count())

	eg.Add(err2)
	assert.Equal(t, 2, eg.Count())

	// Adding nil should not increase count
	eg.Add(nil)
	assert.Equal(t, 2, eg.Count())
}

func TestErrorGroup_Error(t *testing.T) {
	eg := NewErrorGroup()

	// Empty group
	assert.Equal(t, "", eg.Error())

	// Single error
	err1 := errors.New("error 1")
	eg.Add(err1)
	assert.Equal(t, "error 1", eg.Error())

	// Multiple errors
	err2 := errors.New("error 2")
	eg.Add(err2)
	expected := "multiple errors (2):\n  1. error 1\n  2. error 2"
	assert.Equal(t, expected, eg.Error())
}

func TestErrorGroup_First(t *testing.T) {
	eg := NewErrorGroup()

	// Empty group
	assert.Nil(t, eg.First())

	// With errors
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	eg.Add(err1)
	eg.Add(err2)

	assert.Equal(t, err1, eg.First())
}

func TestErrorGroup_Last(t *testing.T) {
	eg := NewErrorGroup()

	// Empty group
	assert.Nil(t, eg.Last())

	// With errors
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	eg.Add(err1)
	eg.Add(err2)

	assert.Equal(t, err2, eg.Last())
}

func TestErrorGroup_Clear(t *testing.T) {
	eg := NewErrorGroup()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	eg.Add(err1)
	eg.Add(err2)
	assert.Equal(t, 2, eg.Count())

	eg.Clear()
	assert.Equal(t, 0, eg.Count())
	assert.False(t, eg.HasErrors())
}

func TestErrorGroup_ToSlice(t *testing.T) {
	eg := NewErrorGroup()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	eg.Add(err1)
	eg.Add(err2)

	slice := eg.ToSlice()
	assert.Equal(t, 2, len(slice))
	assert.Equal(t, err1, slice[0])
	assert.Equal(t, err2, slice[1])

	// Ensure it's a copy
	slice[0] = nil
	assert.Equal(t, err1, eg.Errors[0])
}

func TestErrorGroup_FilterByType(t *testing.T) {
	eg := NewErrorGroup()

	validationErr := NewValidationError("validation failed", nil)
	businessErr := NewBusinessError("RULE", "rule violated")
	notFoundErr := NewNotFoundError("not found")

	eg.Add(validationErr)
	eg.Add(businessErr)
	eg.Add(notFoundErr)

	// Filter validation errors
	validationErrors := eg.FilterByType(ErrorTypeValidation)
	assert.Equal(t, 1, len(validationErrors))
	assert.Equal(t, validationErr, validationErrors[0])

	// Filter business errors
	businessErrors := eg.FilterByType(ErrorTypeBusiness)
	assert.Equal(t, 1, len(businessErrors))
	assert.Equal(t, businessErr, businessErrors[0])

	// Filter non-existent type
	timeoutErrors := eg.FilterByType(ErrorTypeTimeout)
	assert.Equal(t, 0, len(timeoutErrors))
}

// Test utility functions
func TestGetRootCause(t *testing.T) {
	baseErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	domainErr := NewWithError("DOMAIN", "domain error", wrappedErr)

	rootCause := GetRootCause(domainErr)
	assert.Equal(t, baseErr, rootCause)

	// Test with nil
	assert.Nil(t, GetRootCause(nil))

	// Test with single error
	singleErr := errors.New("single error")
	assert.Equal(t, singleErr, GetRootCause(singleErr))
}

func TestGetErrorChain(t *testing.T) {
	baseErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	domainErr := NewWithError("DOMAIN", "domain error", wrappedErr)

	chain := GetErrorChain(domainErr)
	assert.Equal(t, 3, len(chain))
	assert.Equal(t, domainErr, chain[0])
	assert.Equal(t, wrappedErr, chain[1])
	assert.Equal(t, baseErr, chain[2])
}

func TestFormatErrorChain(t *testing.T) {
	baseErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	domainErr := NewWithError("DOMAIN", "domain error", wrappedErr)

	formatted := FormatErrorChain(domainErr)
	assert.Contains(t, formatted, "domain error")
	assert.Contains(t, formatted, "wrapped")
	assert.Contains(t, formatted, "root cause")
	assert.Contains(t, formatted, " -> ")
}

func TestGetTypeName(t *testing.T) {
	validationErr := NewValidationError("validation failed", nil)
	typeName := GetTypeName(validationErr)
	assert.Contains(t, typeName, "ValidationError")

	// Test with nil
	assert.Equal(t, "", GetTypeName(nil))
}

func TestIsRecoverable(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		expected bool
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: false,
		},
		{
			name:     "authentication error - not recoverable",
			error:    NewAuthenticationError("invalid credentials"),
			expected: false,
		},
		{
			name:     "validation error - not recoverable",
			error:    NewValidationError("validation failed", nil),
			expected: false,
		},
		{
			name:     "business error - not recoverable",
			error:    NewBusinessError("RULE", "rule violated"),
			expected: false,
		},
		{
			name:     "timeout error - recoverable",
			error:    NewTimeoutError("timeout", 5*time.Second, 10*time.Second),
			expected: true,
		},
		{
			name:     "database error - recoverable",
			error:    NewDatabaseError("db error", nil),
			expected: true,
		},
		{
			name:     "external service error - recoverable",
			error:    NewExternalServiceError("service", "error", nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRecoverable(tt.error)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		expected bool
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: false,
		},
		{
			name:     "validation error - should not retry",
			error:    NewValidationError("validation failed", nil),
			expected: false,
		},
		{
			name:     "timeout error - should retry",
			error:    NewTimeoutError("timeout", 5*time.Second, 10*time.Second),
			expected: true,
		},
		{
			name:     "database error - should retry",
			error:    NewDatabaseError("db error", nil),
			expected: true,
		},
		{
			name:     "external service error - should retry",
			error:    NewExternalServiceError("service", "error", nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRetry(tt.error)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		expected string
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: "none",
		},
		{
			name:     "security error - high severity",
			error:    NewWithType("SEC", "security error", ErrorTypeSecurity),
			expected: "high",
		},
		{
			name:     "authentication error - high severity",
			error:    NewAuthenticationError("invalid credentials"),
			expected: "high",
		},
		{
			name:     "database error - high severity",
			error:    NewDatabaseError("db error", nil),
			expected: "high",
		},
		{
			name:     "external service error - medium severity",
			error:    NewExternalServiceError("service", "error", nil),
			expected: "medium",
		},
		{
			name:     "timeout error - medium severity",
			error:    NewTimeoutError("timeout", 5*time.Second, 10*time.Second),
			expected: "medium",
		},
		{
			name:     "validation error - low severity",
			error:    NewValidationError("validation failed", nil),
			expected: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSeverity(tt.error)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMust(t *testing.T) {
	// Test with nil error (should not panic)
	assert.NotPanics(t, func() {
		Must(nil)
	})

	// Test with error (should panic)
	err := errors.New("test error")
	assert.Panics(t, func() {
		Must(err)
	})
}

func TestMustReturn(t *testing.T) {
	// Test with nil error
	result := MustReturn("success", nil)
	assert.Equal(t, "success", result)

	// Test with error (should panic)
	err := errors.New("test error")
	assert.Panics(t, func() {
		MustReturn("value", err)
	})
}

func TestRecover(t *testing.T) {
	// Test without panic
	err := Recover()
	assert.NoError(t, err)
}

func TestRecoverWithStackTrace(t *testing.T) {
	// Test without panic
	err := RecoverWithStackTrace()
	assert.NoError(t, err)
}

// Test configuration functions
func TestSetGlobalStackTraceEnabled(t *testing.T) {
	// Test enabling
	SetGlobalStackTraceEnabled(true)
	assert.True(t, GlobalStackTraceEnabled)

	// Test disabling
	SetGlobalStackTraceEnabled(false)
	assert.False(t, GlobalStackTraceEnabled)

	// Reset to default
	SetGlobalStackTraceEnabled(true)
}

func TestSetGlobalMaxStackDepth(t *testing.T) {
	original := GlobalMaxStackDepth

	SetGlobalMaxStackDepth(20)
	assert.Equal(t, 20, GlobalMaxStackDepth)

	// Reset to original
	SetGlobalMaxStackDepth(original)
}

func TestSetGlobalSkipFrames(t *testing.T) {
	original := GlobalSkipFrames

	SetGlobalSkipFrames(5)
	assert.Equal(t, 5, GlobalSkipFrames)

	// Reset to original
	SetGlobalSkipFrames(original)
}

// Test WithCallerInfo
func TestWithCallerInfo(t *testing.T) {
	err := New("TEST", "test error")
	result := WithCallerInfo(err)

	assert.Contains(t, result.MetadataMap, "caller")
	callerInfo, ok := result.MetadataMap["caller"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, callerInfo, "function")
	assert.Contains(t, callerInfo, "file")
	assert.Contains(t, callerInfo, "line")
}

// Test edge cases and error conditions
func TestDomainError_EdgeCases(t *testing.T) {
	// Test with nil metadata map
	err := &DomainError{
		CodeField: "TEST",
		Message:   "test",
	}

	metadata := err.Metadata()
	assert.NotNil(t, metadata)

	// Test error with empty values
	emptyErr := New("", "")
	assert.Equal(t, "", emptyErr.Code())
	assert.Equal(t, "", emptyErr.Message)
	assert.Equal(t, "", emptyErr.Error())
}

func TestStackTraceCapture(t *testing.T) {
	// Test with stack trace enabled
	SetGlobalStackTraceEnabled(true)
	err := New("TEST", "test error")
	assert.NotEmpty(t, err.StackTraceString)

	// Test with stack trace disabled
	SetGlobalStackTraceEnabled(false)
	err2 := New("TEST", "test error")
	assert.Empty(t, err2.StackTraceString)

	// Reset to enabled
	SetGlobalStackTraceEnabled(true)
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New("TEST", "test error")
	}
}

func BenchmarkNewWithError(b *testing.B) {
	baseErr := errors.New("base error")
	for i := 0; i < b.N; i++ {
		NewWithError("TEST", "test error", baseErr)
	}
}

func BenchmarkValidationError(b *testing.B) {
	fields := map[string][]string{
		"email": {"invalid format"},
		"age":   {"must be positive"},
	}

	for i := 0; i < b.N; i++ {
		NewValidationError("validation failed", fields)
	}
}

func BenchmarkErrorChain(b *testing.B) {
	baseErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	domainErr := NewWithError("DOMAIN", "domain error", wrappedErr)

	for i := 0; i < b.N; i++ {
		GetErrorChain(domainErr)
	}
}

func BenchmarkIsType(b *testing.B) {
	err := NewValidationError("validation failed", nil)

	for i := 0; i < b.N; i++ {
		IsType(err, ErrorTypeValidation)
	}
}

func BenchmarkHTTPStatus(b *testing.B) {
	err := NewValidationError("validation failed", nil)

	for i := 0; i < b.N; i++ {
		GetHTTPStatus(err)
	}
}

func BenchmarkJSON(b *testing.B) {
	err := New("TEST", "test error").
		WithMetadata("key1", "value1").
		WithType(ErrorTypeValidation)

	for i := 0; i < b.N; i++ {
		err.JSON()
	}
}

func BenchmarkStackTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		captureStackTrace()
	}
}

// Test with different error types to ensure comprehensive coverage
func TestAllErrorTypes(t *testing.T) {
	errorTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeNotFound,
		ErrorTypeBusiness,
		ErrorTypeDatabase,
		ErrorTypeExternalService,
		ErrorTypeInfrastructure,
		ErrorTypeDependency,
		ErrorTypeAuthentication,
		ErrorTypeAuthorization,
		ErrorTypeSecurity,
		ErrorTypeTimeout,
		ErrorTypeRateLimit,
		ErrorTypeResourceExhausted,
		ErrorTypeCircuitBreaker,
		ErrorTypeSerialization,
		ErrorTypeCache,
		ErrorTypeMigration,
		ErrorTypeConfiguration,
		ErrorTypeUnsupported,
		ErrorTypeBadRequest,
		ErrorTypeConflict,
		ErrorTypeInvalidSchema,
		ErrorTypeUnsupportedMedia,
		ErrorTypeServer,
		ErrorTypeUnprocessable,
		ErrorTypeServiceUnavailable,
		ErrorTypeWorkflow,
	}

	for _, errorType := range errorTypes {
		t.Run(string(errorType), func(t *testing.T) {
			err := NewWithType("TEST", "test error", errorType)
			assert.Equal(t, errorType, err.Type())

			// Test HTTP status mapping
			status := MapHTTPStatus(errorType)
			assert.Greater(t, status, 0)
			assert.Less(t, status, 600)

			// Test IsType function
			assert.True(t, IsType(err, errorType))

			// Test with a different error type to ensure it returns false
			differentType := ErrorTypeBusiness
			if errorType == ErrorTypeBusiness {
				differentType = ErrorTypeValidation
			}
			assert.False(t, IsType(err, differentType))
		})
	}
} // Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	var _ error = (*DomainError)(nil)
	var _ error = (*ValidationError)(nil)
	var _ error = (*NotFoundError)(nil)
	var _ error = (*BusinessError)(nil)
	var _ error = (*DatabaseError)(nil)
	var _ error = (*ExternalServiceError)(nil)
	var _ error = (*InfrastructureError)(nil)
	var _ error = (*AuthenticationError)(nil)
	var _ error = (*AuthorizationError)(nil)
	var _ error = (*TimeoutError)(nil)
	var _ error = (*ServerError)(nil)
	var _ error = (*UnprocessableEntityError)(nil)
	var _ error = (*ErrorGroup)(nil)
	var _ error = (*ErrorStack)(nil)
	var _ error = (*ErrorWrapper)(nil)
}

// Test advanced wrapping functions
func TestWrapWithCode(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := WrapWithCode("CUSTOM_001", "custom message", baseErr)

	assert.Equal(t, "CUSTOM_001", wrapped.Code())
	assert.Equal(t, "custom message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ErrorTypeBusiness, wrapped.ErrorType)
}

func TestWrapWithTypeAndCode(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := WrapWithTypeAndCode("CUSTOM_001", "custom message", baseErr, ErrorTypeValidation)

	assert.Equal(t, "CUSTOM_001", wrapped.Code())
	assert.Equal(t, "custom message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ErrorTypeValidation, wrapped.ErrorType)
}

func TestWrapWithContext(t *testing.T) {
	baseErr := errors.New("base error")
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	wrapped := WrapWithContext(ctx, "context message", baseErr)

	assert.Equal(t, "context message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ctx, wrapped.Context)
}

func TestWrapWithMetadata(t *testing.T) {
	baseErr := errors.New("base error")
	metadata := map[string]interface{}{
		"user_id": "123",
		"action":  "create",
	}

	wrapped := WrapWithMetadata("metadata message", baseErr, metadata)

	assert.Equal(t, "metadata message", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, "123", wrapped.Metadata()["user_id"])
	assert.Equal(t, "create", wrapped.Metadata()["action"])
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := Wrapf("user %s failed to %s", baseErr, "john", "login")

	assert.Equal(t, "user john failed to login", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
}

func TestWrapWithCodef(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := WrapWithCodef("USER_001", "user %s failed to %s", baseErr, "john", "login")

	assert.Equal(t, "USER_001", wrapped.Code())
	assert.Equal(t, "user john failed to login", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
}

func TestWrapWithTypef(t *testing.T) {
	baseErr := errors.New("base error")

	wrapped := WrapWithTypef("validation failed for %s", baseErr, ErrorTypeValidation, "email")

	assert.Equal(t, "validation failed for email", wrapped.Message)
	assert.Equal(t, baseErr, wrapped.Cause)
	assert.Equal(t, ErrorTypeValidation, wrapped.ErrorType)
}

func TestWrapMultiple(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	tests := []struct {
		name     string
		errors   []error
		expected bool
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: false,
		},
		{
			name:     "single error",
			errors:   []error{err1},
			expected: true,
		},
		{
			name:     "multiple errors",
			errors:   []error{err1, err2, err3},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapMultiple("operation failed", tt.errors...)
			if tt.expected {
				assert.NotNil(t, result)
				assert.Contains(t, result.Message, "operation failed")
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// Test ErrorStack
func TestErrorStack_New(t *testing.T) {
	rootErr := errors.New("root error")
	stack := NewErrorStack(rootErr)

	assert.NotNil(t, stack)
	assert.Equal(t, rootErr, stack.Root())
	assert.Equal(t, 1, stack.Size())
	assert.False(t, stack.IsEmpty())
}

func TestErrorStack_Push(t *testing.T) {
	rootErr := errors.New("root error")
	stack := NewErrorStack(rootErr)

	err2 := errors.New("error 2")
	stack.Push(err2)

	assert.Equal(t, 2, stack.Size())
	assert.Equal(t, err2, stack.Peek())
}

func TestErrorStack_Pop(t *testing.T) {
	rootErr := errors.New("root error")
	stack := NewErrorStack(rootErr)

	err2 := errors.New("error 2")
	stack.Push(err2)

	popped := stack.Pop()
	assert.Equal(t, err2, popped)
	assert.Equal(t, 1, stack.Size())
	assert.Equal(t, rootErr, stack.Peek())
}

func TestErrorStack_Empty(t *testing.T) {
	rootErr := errors.New("root error")
	stack := NewErrorStack(rootErr)

	// Pop the only error
	popped := stack.Pop()
	assert.Equal(t, rootErr, popped)
	assert.True(t, stack.IsEmpty())
	assert.Equal(t, 0, stack.Size())

	// Pop from empty stack
	nilErr := stack.Pop()
	assert.Nil(t, nilErr)
}

func TestErrorStack_Error(t *testing.T) {
	rootErr := errors.New("root error")
	stack := NewErrorStack(rootErr)

	err2 := errors.New("error 2")
	stack.Push(err2)

	errorStr := stack.Error()
	assert.Contains(t, errorStr, "error stack")
	assert.Contains(t, errorStr, "root error")
	assert.Contains(t, errorStr, "error 2")
}

// Test ErrorWrapper
func TestErrorWrapper_New(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	assert.NotNil(t, wrapper)
	assert.Equal(t, originalErr, wrapper.Root())
	assert.Equal(t, originalErr, wrapper.Current())
	assert.Equal(t, 1, wrapper.Depth())
}

func TestErrorWrapper_Wrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	wrapper.Wrap("wrapped message")

	assert.Equal(t, originalErr, wrapper.Root())
	assert.Equal(t, 2, wrapper.Depth())
	assert.Contains(t, wrapper.Current().Error(), "wrapped message")
}

func TestErrorWrapper_WrapWithCode(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	wrapper.WrapWithCode("WRAP_001", "wrapped with code")

	assert.Equal(t, originalErr, wrapper.Root())
	assert.Equal(t, 2, wrapper.Depth())

	domainErr, ok := wrapper.Current().(*DomainError)
	assert.True(t, ok)
	assert.Equal(t, "WRAP_001", domainErr.Code())
	assert.Equal(t, "wrapped with code", domainErr.Message)
}

func TestErrorWrapper_WrapWithType(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	wrapper.WrapWithType("wrapped with type", ErrorTypeValidation)

	assert.Equal(t, originalErr, wrapper.Root())
	assert.Equal(t, 2, wrapper.Depth())

	domainErr, ok := wrapper.Current().(*DomainError)
	assert.True(t, ok)
	assert.Equal(t, ErrorTypeValidation, domainErr.ErrorType)
}

func TestErrorWrapper_WithMetadata(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	wrapper.WithMetadata("key", "value")

	assert.Equal(t, "value", wrapper.metadata["key"])
}

func TestErrorWrapper_Chain(t *testing.T) {
	originalErr := errors.New("original error")
	wrapper := NewErrorWrapper(originalErr)

	wrapper.Wrap("first wrap")
	wrapper.Wrap("second wrap")

	chain := wrapper.Chain()
	assert.Equal(t, 3, len(chain))
	assert.Equal(t, originalErr, chain[0])
	assert.Contains(t, chain[1].Error(), "first wrap")
	assert.Contains(t, chain[2].Error(), "second wrap")
}

// Test ErrorChainNavigator
func TestErrorChainNavigator_New(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	navigator := NewErrorChainNavigator(err3)

	assert.NotNil(t, navigator)
	assert.Equal(t, 3, navigator.Size())
	assert.Equal(t, err3, navigator.Current())
	assert.Equal(t, err3, navigator.Top())
	assert.Equal(t, err1, navigator.Root())
}

func TestErrorChainNavigator_Navigation(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	navigator := NewErrorChainNavigator(err3)

	// Test Next
	assert.True(t, navigator.HasNext())
	next := navigator.Next()
	assert.Equal(t, err2, next)

	// Test Next again
	assert.True(t, navigator.HasNext())
	next = navigator.Next()
	assert.Equal(t, err1, next)

	// Test no more next
	assert.False(t, navigator.HasNext())
	next = navigator.Next()
	assert.Nil(t, next)

	// Test Previous
	assert.True(t, navigator.HasPrevious())
	prev := navigator.Previous()
	assert.Equal(t, err2, prev)

	// Test Previous again
	assert.True(t, navigator.HasPrevious())
	prev = navigator.Previous()
	assert.Equal(t, err3, prev)

	// Test no more previous
	assert.False(t, navigator.HasPrevious())
	prev = navigator.Previous()
	assert.Nil(t, prev)
}

func TestErrorChainNavigator_GoToRoot(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	navigator := NewErrorChainNavigator(err3)

	root := navigator.GoToRoot()
	assert.Equal(t, err1, root)
	assert.Equal(t, err1, navigator.Current())
	assert.Equal(t, 2, navigator.Position())
}

func TestErrorChainNavigator_GoToTop(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	navigator := NewErrorChainNavigator(err3)
	navigator.GoToRoot() // Move to root first

	top := navigator.GoToTop()
	assert.Equal(t, err3, top)
	assert.Equal(t, err3, navigator.Current())
	assert.Equal(t, 0, navigator.Position())
}

func TestErrorChainNavigator_FindByType(t *testing.T) {
	validationErr := NewValidationError("validation error", nil)
	wrappedErr := Wrap("wrapped", validationErr)

	navigator := NewErrorChainNavigator(wrappedErr)

	found := navigator.FindByType(ErrorTypeValidation)
	assert.NotNil(t, found)
	assert.IsType(t, &ValidationError{}, found)

	notFound := navigator.FindByType(ErrorTypeDatabase)
	assert.Nil(t, notFound)
}

func TestErrorChainNavigator_FindByCode(t *testing.T) {
	businessErr := NewBusinessError("BUS_001", "business error")
	wrappedErr := Wrap("wrapped", businessErr)

	navigator := NewErrorChainNavigator(wrappedErr)

	found := navigator.FindByCode("BUS_001")
	assert.NotNil(t, found)

	notFound := navigator.FindByCode("NOT_FOUND")
	assert.Nil(t, notFound)
}

func TestErrorChainNavigator_FilterByType(t *testing.T) {
	businessErr := NewBusinessError("BUS_001", "business error")
	wrapped3 := Wrap("wrapped 3", businessErr)

	navigator := NewErrorChainNavigator(wrapped3)

	validationErrs := navigator.FilterByType(ErrorTypeValidation)
	assert.Equal(t, 0, len(validationErrs)) // No validation errors in this chain

	businessErrs := navigator.FilterByType(ErrorTypeBusiness)
	assert.Equal(t, 2, len(businessErrs)) // Both wrapped3 and businessErr are ErrorTypeBusiness
} // Test advanced unwrapping functions
func TestUnwrapAll(t *testing.T) {
	err1 := errors.New("root error")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	root := UnwrapAll(err3)
	assert.Equal(t, err1, root)

	// Test with single error
	single := UnwrapAll(err1)
	assert.Equal(t, err1, single)
}

func TestUnwrapToType(t *testing.T) {
	validationErr := NewValidationError("validation error", nil)
	wrappedErr := Wrap("wrapped", validationErr)

	found := UnwrapToType(wrappedErr, ErrorTypeValidation)
	assert.NotNil(t, found)
	assert.IsType(t, &ValidationError{}, found)

	notFound := UnwrapToType(wrappedErr, ErrorTypeDatabase)
	assert.Nil(t, notFound)
}

func TestUnwrapToCode(t *testing.T) {
	businessErr := NewBusinessError("BUS_001", "business error")
	wrappedErr := Wrap("wrapped", businessErr)

	found := UnwrapToCode(wrappedErr, "BUS_001")
	assert.NotNil(t, found)

	notFound := UnwrapToCode(wrappedErr, "NOT_FOUND")
	assert.Nil(t, notFound)
}

func TestGetErrorAtDepth(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	// Test valid depths
	assert.Equal(t, err3, GetErrorAtDepth(err3, 0))
	assert.Equal(t, err2, GetErrorAtDepth(err3, 1))
	assert.Equal(t, err1, GetErrorAtDepth(err3, 2))

	// Test invalid depths
	assert.Nil(t, GetErrorAtDepth(err3, -1))
	assert.Nil(t, GetErrorAtDepth(err3, 3))
}

func TestGetErrorDepth(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	assert.Equal(t, 0, GetErrorDepth(err3, err3))
	assert.Equal(t, 1, GetErrorDepth(err3, err2))
	assert.Equal(t, 2, GetErrorDepth(err3, err1))

	// Test error not in chain
	otherErr := errors.New("other error")
	assert.Equal(t, -1, GetErrorDepth(err3, otherErr))
}

func TestHasErrorInChain(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("error 2: %w", err1)
	err3 := fmt.Errorf("error 3: %w", err2)

	assert.True(t, HasErrorInChain(err3, err1))
	assert.True(t, HasErrorInChain(err3, err2))
	assert.True(t, HasErrorInChain(err3, err3))

	otherErr := errors.New("other error")
	assert.False(t, HasErrorInChain(err3, otherErr))
}

func TestHasErrorTypeInChain(t *testing.T) {
	validationErr := NewValidationError("validation error", nil)
	wrappedErr := Wrap("wrapped", validationErr)

	assert.True(t, HasErrorTypeInChain(wrappedErr, ErrorTypeValidation))
	assert.False(t, HasErrorTypeInChain(wrappedErr, ErrorTypeDatabase))
}

func TestHasErrorCodeInChain(t *testing.T) {
	businessErr := NewBusinessError("BUS_001", "business error")
	wrappedErr := Wrap("wrapped", businessErr)

	assert.True(t, HasErrorCodeInChain(wrappedErr, "BUS_001"))
	assert.False(t, HasErrorCodeInChain(wrappedErr, "NOT_FOUND"))
}
