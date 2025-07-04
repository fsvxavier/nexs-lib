package domainerrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Custom error type that implements HttpStatusProvider
type customStatusError struct{}

func (e customStatusError) Error() string {
	return "custom status error"
}

func (e customStatusError) StatusCode() int {
	return http.StatusTeapot
}

func TestValidationError2(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		fields := map[string][]string{"email": {"invalid format"}}
		err := NewValidationError("validation failed", fields)

		assert.Equal(t, "VALIDATION_ERROR", err.Code)
		assert.Equal(t, "validation failed", err.Message)
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, fields, err.ValidatedFields)
	})

	t.Run("WithField", func(t *testing.T) {
		err := NewValidationError("validation failed", nil)
		err.WithField("email", "invalid format")
		err.WithField("email", "required field")

		assert.Equal(t, 2, len(err.ValidatedFields["email"]))
		assert.Contains(t, err.ValidatedFields["email"], "invalid format")
		assert.Contains(t, err.ValidatedFields["email"], "required field")
	})

	t.Run("WithFields", func(t *testing.T) {
		err := NewValidationError("validation failed", nil)
		fields := map[string][]string{
			"email":    {"invalid format"},
			"password": {"too short", "needs special character"},
		}
		err.WithFields(fields)

		assert.Equal(t, 1, len(err.ValidatedFields["email"]))
		assert.Equal(t, 2, len(err.ValidatedFields["password"]))
	})
}

func TestNotFoundError2(t *testing.T) {
	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("user not found")

		assert.Equal(t, "NOT_FOUND", err.Code)
		assert.Equal(t, "user not found", err.Message)
		assert.Equal(t, ErrorTypeNotFound, err.Type)
	})

	t.Run("WithResource", func(t *testing.T) {
		err := NewNotFoundError("user not found").WithResource("user", "123")

		assert.Equal(t, "user", err.ResourceType)
		assert.Equal(t, "123", err.ResourceID)
	})
}

func TestBusinessError2(t *testing.T) {
	t.Run("NewBusinessError", func(t *testing.T) {
		err := NewBusinessError("CREDIT_LIMIT", "credit limit exceeded")

		assert.Equal(t, "CREDIT_LIMIT", err.Code)
		assert.Equal(t, "credit limit exceeded", err.Message)
		assert.Equal(t, ErrorTypeBusinessRule, err.Type)
		assert.Equal(t, "CREDIT_LIMIT", err.BusinessCode)
	})
}

func TestInfrastructureError(t *testing.T) {
	t.Run("NewInfrastructureError", func(t *testing.T) {
		underlying := errors.New("connection refused")
		err := NewInfrastructureError("database", "failed to connect", underlying)

		assert.Equal(t, "INFRA_ERROR", err.Code)
		assert.Equal(t, "failed to connect", err.Message)
		assert.Equal(t, ErrorTypeInfrastructure, err.Type)
		assert.Equal(t, "database", err.Component)
		assert.ErrorIs(t, err.Unwrap(), underlying)
	})
}

func TestDatabaseError2(t *testing.T) {
	t.Run("NewDatabaseError", func(t *testing.T) {
		underlying := errors.New("duplicate key value")
		err := NewDatabaseError("insert failed", underlying)

		assert.Equal(t, "INFRA_ERROR", err.Code)
		assert.Equal(t, "insert failed", err.Message)
		assert.Equal(t, ErrorTypeInfrastructure, err.Type)
		assert.Equal(t, "database", err.Component)
	})

	t.Run("WithOperation", func(t *testing.T) {
		err := NewDatabaseError("insert failed", nil).WithOperation("INSERT", "users")

		assert.Equal(t, "INSERT", err.Operation)
		assert.Equal(t, "users", err.Table)
	})

	t.Run("WithSQLState", func(t *testing.T) {
		err := NewDatabaseError("insert failed", nil).WithSQLState("23505")

		assert.Equal(t, "23505", err.SQLState)
	})
}

func TestExternalServiceError2(t *testing.T) {
	t.Run("NewExternalServiceError", func(t *testing.T) {
		underlying := errors.New("timeout")
		err := NewExternalServiceError("payment-api", "payment failed", underlying)

		assert.Equal(t, "EXTERNAL_ERROR", err.Code)
		assert.Equal(t, "payment failed", err.Message)
		assert.Equal(t, ErrorTypeExternalService, err.Type)
		assert.Equal(t, "payment-api", err.ServiceName)
	})

	t.Run("WithStatusCode", func(t *testing.T) {
		err := NewExternalServiceError("payment-api", "payment failed", nil).WithStatusCode(502)

		assert.Equal(t, 502, err.HTTPStatus)
	})

	t.Run("WithResponse", func(t *testing.T) {
		response := map[string]interface{}{"error": "insufficient funds"}
		err := NewExternalServiceError("payment-api", "payment failed", nil).WithResponse(response)

		assert.Equal(t, response, err.Response)
	})
}

func TestAuthenticationError(t *testing.T) {
	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError("invalid credentials")

		assert.Equal(t, "AUTH_ERROR", err.Code)
		assert.Equal(t, "invalid credentials", err.Message)
		assert.Equal(t, ErrorTypeAuthentication, err.Type)
	})

	t.Run("WithReason", func(t *testing.T) {
		err := NewAuthenticationError("invalid credentials").WithReason("password mismatch")

		assert.Equal(t, "password mismatch", err.Reason)
	})
}

func TestAuthorizationError(t *testing.T) {
	t.Run("NewAuthorizationError", func(t *testing.T) {
		err := NewAuthorizationError("insufficient permissions")

		assert.Equal(t, "FORBIDDEN", err.Code)
		assert.Equal(t, "insufficient permissions", err.Message)
		assert.Equal(t, ErrorTypeAuthorization, err.Type)
	})

	t.Run("WithRequiredPermission", func(t *testing.T) {
		err := NewAuthorizationError("insufficient permissions").WithRequiredPermission("admin:write", "user123")

		assert.Equal(t, "admin:write", err.RequiredPermission)
		assert.Equal(t, "user123", err.UserID)
	})
}

func TestGetStatusCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		code := GetStatusCode(nil)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("HttpStatusProvider implementation", func(t *testing.T) {
		err := customStatusError{}
		code := GetStatusCode(err)
		assert.Equal(t, http.StatusTeapot, code)
	})

	t.Run("ErrNoRows", func(t *testing.T) {
		code := GetStatusCode(ErrNoRows)
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("generic error", func(t *testing.T) {
		err := errors.New("some error")
		code := GetStatusCode(err)
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

func TestTimeoutError(t *testing.T) {
	t.Run("NewTimeoutError", func(t *testing.T) {
		err := NewTimeoutError("database query", "operation timed out")

		assert.Equal(t, "TIMEOUT", err.Code)
		assert.Equal(t, "operation timed out", err.Message)
		assert.Equal(t, "database query", err.OperationName)
		assert.Equal(t, ErrorTypeTimeout, err.Type)
	})

	t.Run("WithThreshold", func(t *testing.T) {
		err := NewTimeoutError("database query", "operation timed out").WithThreshold("30s")

		assert.Equal(t, "30s", err.Threshold)
	})
}

func TestBadRequestError(t *testing.T) {
	t.Run("NewBadRequestError", func(t *testing.T) {
		err := NewBadRequestError("invalid request parameters")

		assert.Equal(t, "BAD_REQUEST", err.Code)
		assert.Equal(t, "invalid request parameters", err.Message)
		assert.Equal(t, ErrorTypeBadRequest, err.Type)
		assert.NotNil(t, err.InvalidParams)
	})

	t.Run("WithInvalidParam", func(t *testing.T) {
		err := NewBadRequestError("invalid request parameters")
		err.WithInvalidParam("age", "must be positive number")

		assert.Equal(t, "must be positive number", err.InvalidParams["age"])
	})
}
func TestUnsupportedOperationError(t *testing.T) {
	t.Run("NewUnsupportedOperationError", func(t *testing.T) {
		operation := "deleteAllData"
		message := "operation not allowed"
		err := NewUnsupportedOperationError(operation, message)

		assert.Equal(t, "UNSUPPORTED", err.Code)
		assert.Equal(t, message, err.Message)
		assert.Equal(t, ErrorTypeUnsupported, err.Type)
		assert.Equal(t, operation, err.Operation)
	})
}

func TestInvalidSchemaError(t *testing.T) {
	t.Run("NewInvalidSchemaError", func(t *testing.T) {
		details := map[string][]string{
			"name": {"required field missing"},
			"age":  {"must be a positive number"},
		}

		err := NewInvalidSchemaError("Schema validation failed").
			WithSchemaInfo("user-schema", "v1.0").
			WithSchemaDetails(details)

		assert.Equal(t, "INVALID_SCHEMA", err.Code)
		assert.Equal(t, "Schema validation failed", err.Message)
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "user-schema", err.SchemaName)
		assert.Equal(t, "v1.0", err.SchemaVersion)
		assert.Equal(t, http.StatusBadRequest, err.StatusCode())
		assert.Len(t, err.Details["name"], 1)
		assert.Contains(t, err.Details["name"], "required field missing")
		assert.Len(t, err.Details["age"], 1)
		assert.Contains(t, err.Details["age"], "must be a positive number")
	})

	t.Run("WithSchemaInfo", func(t *testing.T) {
		err := NewInvalidSchemaError("test").WithSchemaInfo("product-schema", "v2.1")

		assert.Equal(t, "product-schema", err.SchemaName)
		assert.Equal(t, "v2.1", err.SchemaVersion)
	})

	t.Run("WithSchemaDetails", func(t *testing.T) {
		err := NewInvalidSchemaError("test")
		details := map[string][]string{
			"field1": {"error1", "error2"},
			"field2": {"error3"},
		}
		err.WithSchemaDetails(details)

		assert.Len(t, err.Details["field1"], 2)
		assert.Len(t, err.Details["field2"], 1)
		assert.Contains(t, err.Details["field1"], "error1")
		assert.Contains(t, err.Details["field1"], "error2")
	})
}

func TestUnsupportedMediaTypeError(t *testing.T) {
	t.Run("NewUnsupportedMediaTypeError", func(t *testing.T) {
		supportedTypes := []string{"application/json", "application/xml"}
		err := NewUnsupportedMediaTypeError("Media type not supported").
			WithMediaTypeInfo("text/plain", supportedTypes)

		assert.Equal(t, "UNSUPPORTED_MEDIA_TYPE", err.Code)
		assert.Equal(t, "Media type not supported", err.Message)
		assert.Equal(t, ErrorTypeUnsupported, err.Type)
		assert.Equal(t, "text/plain", err.ProvidedType)
		assert.Equal(t, http.StatusUnsupportedMediaType, err.StatusCode())
		assert.Len(t, err.SupportedTypes, 2)
		assert.Contains(t, err.SupportedTypes, "application/json")
		assert.Contains(t, err.SupportedTypes, "application/xml")
	})

	t.Run("WithMediaTypeInfo", func(t *testing.T) {
		supportedTypes := []string{"application/json"}
		err := NewUnsupportedMediaTypeError("test").WithMediaTypeInfo("text/csv", supportedTypes)

		assert.Equal(t, "text/csv", err.ProvidedType)
		assert.Equal(t, supportedTypes, err.SupportedTypes)
	})
}

func TestServerError(t *testing.T) {
	t.Run("NewServerError", func(t *testing.T) {
		originalErr := errors.New("database connection failed")
		metadata := map[string]any{
			"db_host": "localhost",
			"db_port": 5432,
		}

		err := NewServerError("Internal server error", originalErr).
			WithErrorCode("DB_CONN_001").
			WithRequestInfo("req-123", "corr-456").
			WithMetadata(metadata)

		assert.Equal(t, "SERVER_ERROR", err.Code)
		assert.Equal(t, "Internal server error", err.Message)
		assert.Equal(t, ErrorTypeInternal, err.Type)
		assert.Equal(t, "DB_CONN_001", err.ErrorCode)
		assert.Equal(t, "req-123", err.RequestID)
		assert.Equal(t, "corr-456", err.CorrelationID)
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
		assert.Equal(t, "localhost", err.Metadata["db_host"])
		assert.Equal(t, 5432, err.Metadata["db_port"])
		assert.ErrorIs(t, err.Unwrap(), originalErr)
	})

	t.Run("WithErrorCode", func(t *testing.T) {
		err := NewServerError("test", nil).WithErrorCode("ERR_001")

		assert.Equal(t, "ERR_001", err.ErrorCode)
	})

	t.Run("WithRequestInfo", func(t *testing.T) {
		err := NewServerError("test", nil).WithRequestInfo("req-456", "corr-789")

		assert.Equal(t, "req-456", err.RequestID)
		assert.Equal(t, "corr-789", err.CorrelationID)
	})

	t.Run("WithMetadata", func(t *testing.T) {
		metadata := map[string]any{
			"key1": "value1",
			"key2": 42,
		}
		err := NewServerError("test", nil).WithMetadata(metadata)

		assert.Equal(t, "value1", err.Metadata["key1"])
		assert.Equal(t, 42, err.Metadata["key2"])
	})
}

func TestUnprocessableEntityError(t *testing.T) {
	t.Run("NewUnprocessableEntityError", func(t *testing.T) {
		validationErrors := map[string][]string{
			"email": {"invalid format", "already exists"},
			"age":   {"must be 18 or older"},
		}

		err := NewUnprocessableEntityError("Entity validation failed").
			WithEntityInfo("User", "user-123").
			WithValidationErrors(validationErrors).
			WithBusinessRuleViolation("User must be verified before activation")

		assert.Equal(t, "UNPROCESSABLE_ENTITY", err.Code)
		assert.Equal(t, "Entity validation failed", err.Message)
		assert.Equal(t, ErrorTypeUnprocessable, err.Type)
		assert.Equal(t, "User", err.EntityType)
		assert.Equal(t, "user-123", err.EntityID)
		assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
		assert.Len(t, err.ValidationErrors["email"], 2)
		assert.Contains(t, err.ValidationErrors["email"], "invalid format")
		assert.Contains(t, err.ValidationErrors["email"], "already exists")
		assert.Len(t, err.BusinessRules, 1)
		assert.Contains(t, err.BusinessRules, "User must be verified before activation")
	})

	t.Run("WithEntityInfo", func(t *testing.T) {
		err := NewUnprocessableEntityError("test").WithEntityInfo("Product", "prod-456")

		assert.Equal(t, "Product", err.EntityType)
		assert.Equal(t, "prod-456", err.EntityID)
	})

	t.Run("WithValidationErrors", func(t *testing.T) {
		err := NewUnprocessableEntityError("test")
		errors := map[string][]string{
			"name":  {"required"},
			"price": {"must be positive", "required"},
		}
		err.WithValidationErrors(errors)

		assert.Len(t, err.ValidationErrors["name"], 1)
		assert.Len(t, err.ValidationErrors["price"], 2)
	})

	t.Run("WithBusinessRuleViolation", func(t *testing.T) {
		err := NewUnprocessableEntityError("test").
			WithBusinessRuleViolation("Rule 1").
			WithBusinessRuleViolation("Rule 2")

		assert.Len(t, err.BusinessRules, 2)
		assert.Contains(t, err.BusinessRules, "Rule 1")
		assert.Contains(t, err.BusinessRules, "Rule 2")
	})
}

func TestServiceUnavailableError(t *testing.T) {
	t.Run("NewServiceUnavailableError", func(t *testing.T) {
		originalErr := errors.New("connection timeout")
		err := NewServiceUnavailableError("payment-service", "Service temporarily unavailable", originalErr).
			WithServiceInfo("payment", "/health").
			WithRetryInfo("30s", "5 minutes")

		assert.Equal(t, "SERVICE_UNAVAILABLE", err.Code)
		assert.Equal(t, "Service temporarily unavailable", err.Message)
		assert.Equal(t, ErrorTypeExternalService, err.Type)
		assert.Equal(t, "payment-service", err.ServiceName)
		assert.Equal(t, "payment", err.ServiceType)
		assert.Equal(t, "/health", err.HealthEndpoint)
		assert.Equal(t, "30s", err.RetryAfter)
		assert.Equal(t, "5 minutes", err.EstimatedUptime)
		assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode())
		assert.ErrorIs(t, err.Unwrap(), originalErr)
	})

	t.Run("WithServiceInfo", func(t *testing.T) {
		err := NewServiceUnavailableError("test", "test", nil).
			WithServiceInfo("auth", "/status")

		assert.Equal(t, "auth", err.ServiceType)
		assert.Equal(t, "/status", err.HealthEndpoint)
	})

	t.Run("WithRetryInfo", func(t *testing.T) {
		err := NewServiceUnavailableError("test", "test", nil).
			WithRetryInfo("60s", "10 minutes")

		assert.Equal(t, "60s", err.RetryAfter)
		assert.Equal(t, "10 minutes", err.EstimatedUptime)
	})
}

func TestExtendedGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "InvalidSchemaError",
			err:      NewInvalidSchemaError("test"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "UnsupportedMediaTypeError",
			err:      NewUnsupportedMediaTypeError("test"),
			expected: http.StatusUnsupportedMediaType,
		},
		{
			name:     "ServerError",
			err:      NewServerError("test", nil),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "UnprocessableEntityError",
			err:      NewUnprocessableEntityError("test"),
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "ServiceUnavailableError",
			err:      NewServiceUnavailableError("test", "test", nil),
			expected: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode := GetStatusCode(tt.err)
			assert.Equal(t, tt.expected, statusCode, "Expected status code %d for %s, got %d", tt.expected, tt.name, statusCode)
		})
	}
}

func TestNewErrorTypesHelperFunctions(t *testing.T) {
	t.Run("IsInvalidSchemaError", func(t *testing.T) {
		err := NewInvalidSchemaError("test")
		assert.True(t, IsInvalidSchemaError(err))
		assert.False(t, IsInvalidSchemaError(NewValidationError("test", nil)))
		assert.False(t, IsInvalidSchemaError(errors.New("regular error")))
	})

	t.Run("IsUnsupportedMediaTypeError", func(t *testing.T) {
		err := NewUnsupportedMediaTypeError("test")
		assert.True(t, IsUnsupportedMediaTypeError(err))
		assert.False(t, IsUnsupportedMediaTypeError(NewValidationError("test", nil)))
	})

	t.Run("IsServerError", func(t *testing.T) {
		err := NewServerError("test", nil)
		assert.True(t, IsServerError(err))
		assert.False(t, IsServerError(NewValidationError("test", nil)))
	})

	t.Run("IsUnprocessableEntityError", func(t *testing.T) {
		err := NewUnprocessableEntityError("test")
		assert.True(t, IsUnprocessableEntityError(err))
		assert.False(t, IsUnprocessableEntityError(NewValidationError("test", nil)))
	})

	t.Run("IsServiceUnavailableError", func(t *testing.T) {
		err := NewServiceUnavailableError("test", "test", nil)
		assert.True(t, IsServiceUnavailableError(err))
		assert.False(t, IsServiceUnavailableError(NewValidationError("test", nil)))
	})

	t.Run("NewInvalidSchemaErrorFromDetails", func(t *testing.T) {
		details := map[string][]string{
			"field1": {"error1", "error2"},
		}
		err := NewInvalidSchemaErrorFromDetails("test-schema", details)

		assert.Equal(t, "test-schema", err.SchemaName)
		assert.Len(t, err.Details["field1"], 2)
		assert.Contains(t, err.Details["field1"], "error1")
		assert.Contains(t, err.Details["field1"], "error2")
		assert.Contains(t, err.Message, "test-schema")
	})

	t.Run("NewUnsupportedMediaTypeErrorFromTypes", func(t *testing.T) {
		supportedTypes := []string{"application/json", "application/xml"}
		err := NewUnsupportedMediaTypeErrorFromTypes("text/plain", supportedTypes)

		assert.Equal(t, "text/plain", err.ProvidedType)
		assert.Len(t, err.SupportedTypes, 2)
		assert.Contains(t, err.SupportedTypes, "application/json")
		assert.Contains(t, err.SupportedTypes, "application/xml")
		assert.Contains(t, err.Message, "text/plain")
	})

	t.Run("NewServerErrorWithCode", func(t *testing.T) {
		err := NewServerErrorWithCode("DB_001", "Database error", nil)

		assert.Equal(t, "SERVER_ERROR", err.Code)
		assert.Equal(t, "DB_001", err.ErrorCode)
		assert.Equal(t, "Database error", err.Message)
	})

	t.Run("NewUnprocessableEntityErrorFromValidation", func(t *testing.T) {
		validationErrors := map[string][]string{
			"email": {"invalid"},
			"age":   {"required"},
		}
		err := NewUnprocessableEntityErrorFromValidation("User", "user-123", validationErrors)

		assert.Equal(t, "User", err.EntityType)
		assert.Equal(t, "user-123", err.EntityID)
		assert.Equal(t, validationErrors, err.ValidationErrors)
		assert.Contains(t, err.Message, "User")
	})

	t.Run("NewServiceUnavailableErrorWithRetry", func(t *testing.T) {
		err := NewServiceUnavailableErrorWithRetry("payment-service", "30s", nil)

		assert.Equal(t, "payment-service", err.ServiceName)
		assert.Equal(t, "30s", err.RetryAfter)
		assert.Contains(t, err.Message, "payment-service")
	})
}
