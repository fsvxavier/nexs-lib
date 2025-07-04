package domainerrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainError_New(t *testing.T) {
	err := New("TEST_ERR", "This is a test error")

	assert.Equal(t, "TEST_ERR", err.Code)
	assert.Equal(t, "This is a test error", err.Message)
	assert.Nil(t, err.Err)
	assert.Equal(t, "[TEST_ERR] This is a test error", err.Error())
}

func TestDomainError_NewWithError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewWithError("TEST_ERR", "This is a test error", baseErr)

	assert.Equal(t, "TEST_ERR", err.Code)
	assert.Equal(t, "This is a test error", err.Message)
	assert.Equal(t, baseErr, err.Err)
	assert.Equal(t, "[TEST_ERR] This is a test error: base error", err.Error())
	assert.NotEmpty(t, err.stack)
}

func TestDomainError_Wrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := New("TEST_ERR", "This is a test error")
	err = err.Wrap("Wrapped error", baseErr)

	assert.Equal(t, "TEST_ERR", err.Code)
	assert.Equal(t, "This is a test error", err.Message)
	assert.Equal(t, baseErr, err.Err)
	assert.Len(t, err.stack, 1)

	// Wrap another error
	err2 := errors.New("another error")
	err = err.Wrap("Another wrapped", err2)
	assert.Equal(t, err2, err.Err)
	assert.Len(t, err.stack, 2)
}

func TestDomainError_WithType(t *testing.T) {
	err := New("TEST_ERR", "This is a test error").WithType(ErrorTypeValidation)

	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode())
}

func TestDomainError_WithDetails(t *testing.T) {
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	err := New("TEST_ERR", "This is a test error").WithDetails(details)

	assert.Equal(t, "value1", err.Details["key1"])
	assert.Equal(t, 42, err.Details["key2"])

	// Add another detail
	err.WithDetail("key3", true)
	assert.Equal(t, true, err.Details["key3"])
}

func TestDomainError_StatusCode(t *testing.T) {
	testCases := []struct {
		name     string
		errType  ErrorType
		expected int
	}{
		{"Repository", ErrorTypeRepository, http.StatusInternalServerError},
		{"Validation", ErrorTypeValidation, http.StatusBadRequest},
		{"BusinessRule", ErrorTypeBusinessRule, http.StatusUnprocessableEntity},
		{"NotFound", ErrorTypeNotFound, http.StatusNotFound},
		{"ExternalService", ErrorTypeExternalService, http.StatusBadGateway},
		{"Authentication", ErrorTypeAuthentication, http.StatusUnauthorized},
		{"Authorization", ErrorTypeAuthorization, http.StatusForbidden},
		{"BadRequest", ErrorTypeBadRequest, http.StatusBadRequest},
		{"Unprocessable", ErrorTypeUnprocessable, http.StatusUnprocessableEntity},
		{"Unsupported", ErrorTypeUnsupported, http.StatusUnsupportedMediaType},
		{"Timeout", ErrorTypeTimeout, http.StatusGatewayTimeout},
		{"Internal", ErrorTypeInternal, http.StatusInternalServerError},
		{"Infrastructure", ErrorTypeInfrastructure, http.StatusServiceUnavailable},
		{"Default", "unknown", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New("TEST_ERR", "Test").WithType(tc.errType)
			assert.Equal(t, tc.expected, err.StatusCode())
		})
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("Validation failed", nil)

	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode())

	// Add fields
	err.WithField("email", "Invalid email")
	err.WithField("password", "Too short")

	assert.Len(t, err.ValidatedFields, 2)
	assert.Equal(t, []string{"Invalid email"}, err.ValidatedFields["email"])
	assert.Equal(t, []string{"Too short"}, err.ValidatedFields["password"])

	// Add another message to same field
	err.WithField("email", "Email already taken")
	assert.Len(t, err.ValidatedFields["email"], 2)
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("User not found").WithResource("user", "123")

	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, http.StatusNotFound, err.StatusCode())
	assert.Equal(t, "user", err.ResourceType)
	assert.Equal(t, "123", err.ResourceID)
}

func TestBusinessError(t *testing.T) {
	err := NewBusinessError("INSUF_FUNDS", "Insufficient funds")

	assert.Equal(t, "INSUF_FUNDS", err.Code)
	assert.Equal(t, ErrorTypeBusinessRule, err.Type)
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
	assert.Equal(t, "INSUF_FUNDS", err.BusinessCode)
}

func TestDatabaseError(t *testing.T) {
	baseErr := errors.New("db connection error")
	err := NewDatabaseError("Failed to connect", baseErr).
		WithOperation("SELECT", "users")

	assert.Equal(t, "INFRA_ERROR", err.Code)
	assert.Equal(t, ErrorTypeInfrastructure, err.Type)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "database", err.Component)
	assert.Equal(t, "SELECT", err.Operation)
	assert.Equal(t, "users", err.Table)
}

func TestExternalServiceError(t *testing.T) {
	baseErr := errors.New("connection timeout")
	err := NewExternalServiceError("payment-api", "Payment processing failed", baseErr).
		WithStatusCode(http.StatusGatewayTimeout)

	assert.Equal(t, "EXTERNAL_ERROR", err.Code)
	assert.Equal(t, ErrorTypeExternalService, err.Type)
	// O método StatusCode() deve retornar o valor HTTPStatus se definido
	assert.Equal(t, http.StatusGatewayTimeout, err.StatusCode())
	assert.Equal(t, "payment-api", err.ServiceName)
	// O campo HTTPStatus específico do ExternalServiceError contém o valor explicitamente definido
	assert.Equal(t, http.StatusGatewayTimeout, err.HTTPStatus)
}

func TestWrapError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := WrapError("TEST", "Test message", nil)
		assert.Nil(t, err)
	})

	t.Run("normal error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := WrapError("TEST", "Test message", baseErr)

		domainErr, ok := err.(*DomainError)
		assert.True(t, ok)
		assert.Equal(t, "TEST", domainErr.Code)
		assert.Equal(t, "Test message", domainErr.Message)
		assert.Equal(t, baseErr, domainErr.Err)
	})

	t.Run("domain error", func(t *testing.T) {
		baseErr := New("BASE", "Base error")
		err := WrapError("TEST", "Test message", baseErr)

		domainErr, ok := err.(*DomainError)
		assert.True(t, ok)
		assert.Equal(t, "BASE", domainErr.Code)          // Should keep original code
		assert.Equal(t, "Base error", domainErr.Message) // Should keep original message
		assert.NotNil(t, domainErr.Err)
		assert.Len(t, domainErr.stack, 1)
	})
}

func TestErrorTypeCheckers(t *testing.T) {
	notFoundErr := NewNotFoundError("Resource not found")
	validationErr := NewValidationError("Invalid input", nil)
	businessErr := NewBusinessError("RULE_ERR", "Business rule violation")

	assert.True(t, IsNotFoundError(notFoundErr))
	assert.False(t, IsNotFoundError(validationErr))

	assert.True(t, IsValidationError(validationErr))
	assert.False(t, IsValidationError(businessErr))

	assert.True(t, IsBusinessError(businessErr))
	assert.False(t, IsBusinessError(notFoundErr))
}

func TestSQLErrorParser(t *testing.T) {
	parser := NewSQLErrorParser()

	t.Run("valid SQL error", func(t *testing.T) {
		err := errors.New("error at row 1 (SQLSTATE 23505)")
		code, ok := parser.Parse(err)

		assert.True(t, ok)
		assert.Equal(t, "23505", code)
	})

	t.Run("invalid SQL error", func(t *testing.T) {
		err := errors.New("some other error")
		code, ok := parser.Parse(err)

		assert.False(t, ok)
		assert.Equal(t, "", code)
	})

	t.Run("nil error", func(t *testing.T) {
		code, ok := parser.Parse(nil)

		assert.False(t, ok)
		assert.Equal(t, "", code)
	})
}

func TestErrorCodeRegistry(t *testing.T) {
	registry := NewErrorCodeRegistry()

	// Register some error codes
	registry.Register("E001", "Resource not found", http.StatusNotFound)
	registry.Register("E002", "Invalid input", http.StatusBadRequest)

	t.Run("get existing code", func(t *testing.T) {
		code, ok := registry.Get("E001")

		assert.True(t, ok)
		assert.Equal(t, "E001", code.Code)
		assert.Equal(t, "Resource not found", code.Description)
		assert.Equal(t, http.StatusNotFound, code.StatusCode)
	})

	t.Run("get non-existing code", func(t *testing.T) {
		_, ok := registry.Get("E999")
		assert.False(t, ok)
	})

	t.Run("wrap with registered code", func(t *testing.T) {
		baseErr := errors.New("item 123 not found")
		err := registry.WrapWithCode("E001", baseErr)

		domainErr, ok := err.(*DomainError)
		assert.True(t, ok)
		assert.Equal(t, "E001", domainErr.Code)
		assert.Equal(t, "Resource not found", domainErr.Message)
		assert.Equal(t, baseErr, domainErr.Err)
	})

	t.Run("wrap with unknown code", func(t *testing.T) {
		baseErr := errors.New("some error")
		err := registry.WrapWithCode("E999", baseErr)

		domainErr, ok := err.(*DomainError)
		assert.True(t, ok)
		assert.Equal(t, "E999", domainErr.Code)
		assert.Equal(t, "Unknown error", domainErr.Message)
		assert.Equal(t, baseErr, domainErr.Err)
	})
}

func TestErrorStack(t *testing.T) {
	stack := NewErrorStack()

	t.Run("empty stack", func(t *testing.T) {
		assert.True(t, stack.IsEmpty())
		assert.Equal(t, "", stack.Error())
		assert.Nil(t, stack.Unwrap())
	})

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	stack.Push(err1)
	stack.Push(err2)

	t.Run("non-empty stack", func(t *testing.T) {
		assert.False(t, stack.IsEmpty())
		assert.Equal(t, "error 1\nerror 2", stack.Error())
		assert.Equal(t, err1, stack.Unwrap())

		slice := stack.ToSlice()
		assert.Len(t, slice, 2)
		assert.Equal(t, err1, slice[0])
		assert.Equal(t, err2, slice[1])
	})
}

func TestNewErrorTypes(t *testing.T) {
	t.Run("ConflictError", func(t *testing.T) {
		err := NewConflictError("Resource already exists")
		err.WithConflictingResource("user", "email already taken")

		assert.Equal(t, "CONFLICT", err.Code)
		assert.Equal(t, ErrorTypeConflict, err.Type)
		assert.Equal(t, http.StatusConflict, err.StatusCode())
		assert.Equal(t, "user", err.ConflictingResource)
		assert.Equal(t, "email already taken", err.ConflictReason)
		assert.True(t, IsConflictError(err))
	})

	t.Run("RateLimitError", func(t *testing.T) {
		err := NewRateLimitError("Rate limit exceeded")
		err.WithRateLimit(100, 0, "2024-01-01T00:00:00Z", "60s")

		assert.Equal(t, "RATE_LIMIT", err.Code)
		assert.Equal(t, ErrorTypeRateLimit, err.Type)
		assert.Equal(t, http.StatusTooManyRequests, err.StatusCode())
		assert.Equal(t, 100, err.Limit)
		assert.Equal(t, 0, err.Remaining)
		assert.Equal(t, "60s", err.RetryAfter)
		assert.True(t, IsRateLimitError(err))
	})

	t.Run("CircuitBreakerError", func(t *testing.T) {
		err := NewCircuitBreakerError("payment-service", "Circuit breaker is open")
		err.WithCircuitState("OPEN", 5)

		assert.Equal(t, "CIRCUIT_BREAKER", err.Code)
		assert.Equal(t, ErrorTypeCircuitBreaker, err.Type)
		assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode())
		assert.Equal(t, "payment-service", err.ServiceName)
		assert.Equal(t, "OPEN", err.State)
		assert.Equal(t, 5, err.Failures)
		assert.True(t, IsCircuitBreakerError(err))
	})

	t.Run("ConfigurationError", func(t *testing.T) {
		err := NewConfigurationError("Invalid configuration")
		err.WithConfigDetails("database.host", "localhost", "valid hostname")

		assert.Equal(t, "CONFIG_ERROR", err.Code)
		assert.Equal(t, ErrorTypeConfiguration, err.Type)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
		assert.Equal(t, "database.host", err.ConfigKey)
		assert.Equal(t, "localhost", err.ConfigValue)
		assert.Equal(t, "valid hostname", err.Expected)
		assert.True(t, IsConfigurationError(err))
	})

	t.Run("SecurityError", func(t *testing.T) {
		err := NewSecurityError("Suspicious activity detected")
		err.WithSecurityContext("login_attempt", "HIGH")
		err.WithClientInfo("Mozilla/5.0", "192.168.1.1")

		assert.Equal(t, "SECURITY_ERROR", err.Code)
		assert.Equal(t, ErrorTypeSecurity, err.Type)
		assert.Equal(t, http.StatusForbidden, err.StatusCode())
		assert.Equal(t, "login_attempt", err.SecurityContext)
		assert.Equal(t, "HIGH", err.ThreatLevel)
		assert.Equal(t, "Mozilla/5.0", err.UserAgent)
		assert.Equal(t, "192.168.1.1", err.IPAddress)
		assert.True(t, IsSecurityError(err))
	})

	t.Run("ResourceExhaustedError", func(t *testing.T) {
		err := NewResourceExhaustedError("memory", "Memory limit exceeded")
		err.WithResourceLimits(1024, 1024, "MB")

		assert.Equal(t, "RESOURCE_EXHAUSTED", err.Code)
		assert.Equal(t, ErrorTypeResourceExhausted, err.Type)
		assert.Equal(t, http.StatusInsufficientStorage, err.StatusCode())
		assert.Equal(t, "memory", err.ResourceType)
		assert.Equal(t, int64(1024), err.Limit)
		assert.Equal(t, int64(1024), err.Current)
		assert.Equal(t, "MB", err.Unit)
		assert.True(t, IsResourceExhaustedError(err))
	})

	t.Run("DependencyError", func(t *testing.T) {
		baseErr := errors.New("connection refused")
		err := NewDependencyError("redis", "Redis connection failed", baseErr)
		err.WithDependencyInfo("cache", "6.2.0", "DOWN")

		assert.Equal(t, "DEPENDENCY_ERROR", err.Code)
		assert.Equal(t, ErrorTypeDependency, err.Type)
		assert.Equal(t, http.StatusFailedDependency, err.StatusCode())
		assert.Equal(t, "redis", err.DependencyName)
		assert.Equal(t, "cache", err.DependencyType)
		assert.Equal(t, "6.2.0", err.Version)
		assert.Equal(t, "DOWN", err.HealthStatus)
		assert.True(t, IsDependencyError(err))
	})

	t.Run("SerializationError", func(t *testing.T) {
		baseErr := errors.New("invalid json")
		err := NewSerializationError("JSON", "Failed to serialize data", baseErr)
		err.WithTypeInfo("user.age", "int", "string")

		assert.Equal(t, "SERIALIZATION_ERROR", err.Code)
		assert.Equal(t, ErrorTypeSerialization, err.Type)
		assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
		assert.Equal(t, "JSON", err.Format)
		assert.Equal(t, "user.age", err.FieldName)
		assert.Equal(t, "int", err.ExpectedType)
		assert.Equal(t, "string", err.ActualType)
		assert.True(t, IsSerializationError(err))
	})

	t.Run("CacheError", func(t *testing.T) {
		baseErr := errors.New("cache miss")
		err := NewCacheError("redis", "GET", "Failed to retrieve from cache", baseErr)
		err.WithCacheDetails("user:123", "300s")

		assert.Equal(t, "CACHE_ERROR", err.Code)
		assert.Equal(t, ErrorTypeCache, err.Type)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
		assert.Equal(t, "redis", err.CacheType)
		assert.Equal(t, "GET", err.Operation)
		assert.Equal(t, "user:123", err.Key)
		assert.Equal(t, "300s", err.TTL)
		assert.True(t, IsCacheError(err))
	})

	t.Run("WorkflowError", func(t *testing.T) {
		err := NewWorkflowError("order-process", "payment", "Payment step failed")
		err.WithStateInfo("pending_payment", "completed_payment")

		assert.Equal(t, "WORKFLOW_ERROR", err.Code)
		assert.Equal(t, ErrorTypeWorkflow, err.Type)
		assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
		assert.Equal(t, "order-process", err.WorkflowID)
		assert.Equal(t, "payment", err.StepName)
		assert.Equal(t, "pending_payment", err.CurrentState)
		assert.Equal(t, "completed_payment", err.ExpectedState)
		assert.True(t, IsWorkflowError(err))
	})

	t.Run("MigrationError", func(t *testing.T) {
		baseErr := errors.New("column already exists")
		err := NewMigrationError("v1.2.0", "add_user_table", "Migration failed", baseErr)
		err.WithMigrationDetails("up", 1500)

		assert.Equal(t, "MIGRATION_ERROR", err.Code)
		assert.Equal(t, ErrorTypeMigration, err.Type)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
		assert.Equal(t, "v1.2.0", err.MigrationVersion)
		assert.Equal(t, "add_user_table", err.MigrationName)
		assert.Equal(t, "up", err.Direction)
		assert.Equal(t, int64(1500), err.AffectedRecords)
		assert.True(t, IsMigrationError(err))
	})
}

func TestDomainError_StatusCodeNewTypes(t *testing.T) {
	testCases := []struct {
		name     string
		errType  ErrorType
		expected int
	}{
		{"Conflict", ErrorTypeConflict, http.StatusConflict},
		{"RateLimit", ErrorTypeRateLimit, http.StatusTooManyRequests},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, http.StatusServiceUnavailable},
		{"Configuration", ErrorTypeConfiguration, http.StatusInternalServerError},
		{"Security", ErrorTypeSecurity, http.StatusForbidden},
		{"ResourceExhausted", ErrorTypeResourceExhausted, http.StatusInsufficientStorage},
		{"Dependency", ErrorTypeDependency, http.StatusFailedDependency},
		{"Serialization", ErrorTypeSerialization, http.StatusUnprocessableEntity},
		{"Cache", ErrorTypeCache, http.StatusInternalServerError},
		{"Workflow", ErrorTypeWorkflow, http.StatusUnprocessableEntity},
		{"Migration", ErrorTypeMigration, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New("TEST_ERR", "Test").WithType(tc.errType)
			assert.Equal(t, tc.expected, err.StatusCode())
		})
	}
}
func TestDomainError_Is(t *testing.T) {
	// Test case: comparing with nil
	t.Run("comparing with nil", func(t *testing.T) {
		err := New("TEST_ERR", "This is a test error")
		assert.False(t, err.Is(nil))
	})

	// Test case: comparing with another DomainError with the same code
	t.Run("comparing with same code DomainError", func(t *testing.T) {
		err1 := New("TEST_ERR", "This is a test error")
		err2 := New("TEST_ERR", "Different message, same code")
		assert.True(t, err1.Is(err2))
	})

	// Test case: comparing with another DomainError with different code
	t.Run("comparing with different code DomainError", func(t *testing.T) {
		err1 := New("TEST_ERR", "This is a test error")
		err2 := New("OTHER_ERR", "Different error")
		assert.False(t, err1.Is(err2))
	})

	// Test case: comparing with wrapped error
	t.Run("comparing with wrapped error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := NewWithError("TEST_ERR", "This is a test error", baseErr)
		assert.True(t, err.Is(baseErr))
	})

	// Test case: comparing with different error
	t.Run("comparing with different error", func(t *testing.T) {
		baseErr := errors.New("base error")
		otherErr := errors.New("other error")
		err := NewWithError("TEST_ERR", "This is a test error", baseErr)
		assert.False(t, err.Is(otherErr))
	})

	// Test case: comparing with deeply wrapped error
	t.Run("comparing with deeply wrapped error", func(t *testing.T) {
		innerErr := errors.New("inner error")
		middleErr := NewWithError("MIDDLE_ERR", "Middle error", innerErr)
		outerErr := NewWithError("OUTER_ERR", "Outer error", middleErr)

		assert.True(t, outerErr.Is(innerErr))
	})
}
func TestDomainError_FormatStackTrace(t *testing.T) {
	// Test case: empty stack
	t.Run("empty stack", func(t *testing.T) {
		err := New("TEST_ERR", "This is a test error")
		assert.Equal(t, "", err.FormatStackTrace())
	})

	// Test case: stack with one error
	t.Run("stack with one error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := NewWithError("TEST_ERR", "This is a test error", baseErr)

		trace := err.FormatStackTrace()
		assert.Contains(t, trace, "Error Stack Trace:")
		assert.Contains(t, trace, "[This is a test error]")
		assert.Contains(t, trace, "Error: base error")
		assert.Contains(t, trace, "1:")
	})

	// Test case: stack with multiple errors
	t.Run("stack with multiple errors", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := NewWithError("TEST_ERR", "First error", baseErr)

		err2 := errors.New("second error")
		err.Wrap("Second wrap", err2)

		err3 := errors.New("third error")
		err.Wrap("Third wrap", err3)

		trace := err.FormatStackTrace()
		assert.Contains(t, trace, "Error Stack Trace:")
		assert.Contains(t, trace, "1: [First error]")
		assert.Contains(t, trace, "2: [Second wrap]")
		assert.Contains(t, trace, "3: [Third wrap]")
		assert.Contains(t, trace, "Error: base error")
		assert.Contains(t, trace, "Error: second error")
		assert.Contains(t, trace, "Error: third error")
	})
}
func TestDomainError_WithMetadata(t *testing.T) {
	err := New("TEST_ERR", "This is a test error")

	metadata := map[string]interface{}{
		"request_id": "abc123",
		"user_id":    42,
		"timestamp":  "2023-01-01T12:00:00Z",
	}

	err.WithMetadata(metadata)

	assert.Equal(t, "abc123", err.Metadata["request_id"])
	assert.Equal(t, 42, err.Metadata["user_id"])
	assert.Equal(t, "2023-01-01T12:00:00Z", err.Metadata["timestamp"])

	// Add more metadata
	err.WithMetadata(map[string]interface{}{"trace_id": "xyz789"})
	assert.Equal(t, "xyz789", err.Metadata["trace_id"])
	assert.Equal(t, "abc123", err.Metadata["request_id"]) // Should retain existing metadata
}

func TestDomainError_WithEntity(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	type Product struct{}

	t.Run("with struct instance", func(t *testing.T) {
		user := User{ID: 1, Name: "John"}
		err := New("TEST_ERR", "This is a test error").WithEntity(user)

		assert.Equal(t, "User", err.EntityName)
	})

	t.Run("with empty struct", func(t *testing.T) {
		product := Product{}
		err := New("TEST_ERR", "This is a test error").WithEntity(product)

		assert.Equal(t, "Product", err.EntityName)
	})

}
