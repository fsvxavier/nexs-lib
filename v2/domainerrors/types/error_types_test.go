package types

import (
	"net/http"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{"Repository", ErrorTypeRepository, "repository"},
		{"Database", ErrorTypeDatabase, "database"},
		{"Cache", ErrorTypeCache, "cache"},
		{"Migration", ErrorTypeMigration, "migration"},
		{"Serialization", ErrorTypeSerialization, "serialization"},
		{"Validation", ErrorTypeValidation, "validation"},
		{"BadRequest", ErrorTypeBadRequest, "bad_request"},
		{"NotFound", ErrorTypeNotFound, "not_found"},
		{"Conflict", ErrorTypeConflict, "conflict"},
		{"Unprocessable", ErrorTypeUnprocessable, "unprocessable"},
		{"Unsupported", ErrorTypeUnsupported, "unsupported"},
		{"BusinessRule", ErrorTypeBusinessRule, "business"},
		{"Authentication", ErrorTypeAuthentication, "authentication"},
		{"Authorization", ErrorTypeAuthorization, "authorization"},
		{"Internal", ErrorTypeInternal, "internal"},
		{"Configuration", ErrorTypeConfiguration, "configuration"},
		{"Dependency", ErrorTypeDependency, "dependency"},
		{"ExternalService", ErrorTypeExternalService, "external_service"},
		{"Timeout", ErrorTypeTimeout, "timeout"},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, "circuit_breaker"},
		{"RateLimit", ErrorTypeRateLimit, "rate_limit"},
		{"ResourceExhausted", ErrorTypeResourceExhausted, "resource_exhausted"},
		{"Network", ErrorTypeNetwork, "network"},
		{"Cloud", ErrorTypeCloud, "cloud"},
		{"HTTP", ErrorTypeHTTP, "http"},
		{"GRPC", ErrorTypeGRPC, "grpc"},
		{"GraphQL", ErrorTypeGraphQL, "graphql"},
		{"WebSocket", ErrorTypeWebSocket, "websocket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{"Valid Repository", ErrorTypeRepository, true},
		{"Valid Database", ErrorTypeDatabase, true},
		{"Valid Cache", ErrorTypeCache, true},
		{"Valid Migration", ErrorTypeMigration, true},
		{"Valid Serialization", ErrorTypeSerialization, true},
		{"Valid Validation", ErrorTypeValidation, true},
		{"Valid BadRequest", ErrorTypeBadRequest, true},
		{"Valid NotFound", ErrorTypeNotFound, true},
		{"Valid Conflict", ErrorTypeConflict, true},
		{"Valid Unprocessable", ErrorTypeUnprocessable, true},
		{"Valid Unsupported", ErrorTypeUnsupported, true},
		{"Valid BusinessRule", ErrorTypeBusinessRule, true},
		{"Valid Authentication", ErrorTypeAuthentication, true},
		{"Valid Authorization", ErrorTypeAuthorization, true},
		{"Valid Internal", ErrorTypeInternal, true},
		{"Valid Configuration", ErrorTypeConfiguration, true},
		{"Valid Dependency", ErrorTypeDependency, true},
		{"Valid ExternalService", ErrorTypeExternalService, true},
		{"Valid Timeout", ErrorTypeTimeout, true},
		{"Valid CircuitBreaker", ErrorTypeCircuitBreaker, true},
		{"Valid RateLimit", ErrorTypeRateLimit, true},
		{"Valid ResourceExhausted", ErrorTypeResourceExhausted, true},
		{"Valid Network", ErrorTypeNetwork, true},
		{"Valid Cloud", ErrorTypeCloud, true},
		{"Valid HTTP", ErrorTypeHTTP, true},
		{"Valid GRPC", ErrorTypeGRPC, true},
		{"Valid GraphQL", ErrorTypeGraphQL, true},
		{"Valid WebSocket", ErrorTypeWebSocket, true},
		{"Invalid Empty", ErrorType(""), false},
		{"Invalid Custom", ErrorType("invalid_type"), false},
		{"Invalid Random", ErrorType("xyz"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.IsValid()
			if result != tt.expected {
				t.Errorf("ErrorType.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorType_StatusCode(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  int
	}{
		{"Repository", ErrorTypeRepository, http.StatusInternalServerError},
		{"Database", ErrorTypeDatabase, http.StatusInternalServerError},
		{"Cache", ErrorTypeCache, http.StatusServiceUnavailable},
		{"Migration", ErrorTypeMigration, http.StatusInternalServerError},
		{"Serialization", ErrorTypeSerialization, http.StatusUnprocessableEntity},
		{"Validation", ErrorTypeValidation, http.StatusBadRequest},
		{"BadRequest", ErrorTypeBadRequest, http.StatusBadRequest},
		{"NotFound", ErrorTypeNotFound, http.StatusNotFound},
		{"Conflict", ErrorTypeConflict, http.StatusConflict},
		{"Unprocessable", ErrorTypeUnprocessable, http.StatusUnprocessableEntity},
		{"Unsupported", ErrorTypeUnsupported, http.StatusUnsupportedMediaType},
		{"BusinessRule", ErrorTypeBusinessRule, http.StatusUnprocessableEntity},
		{"Authentication", ErrorTypeAuthentication, http.StatusUnauthorized},
		{"Authorization", ErrorTypeAuthorization, http.StatusForbidden},
		{"Internal", ErrorTypeInternal, http.StatusInternalServerError},
		{"Configuration", ErrorTypeConfiguration, http.StatusInternalServerError},
		{"Dependency", ErrorTypeDependency, http.StatusFailedDependency},
		{"ExternalService", ErrorTypeExternalService, http.StatusBadGateway},
		{"Timeout", ErrorTypeTimeout, http.StatusGatewayTimeout},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, http.StatusServiceUnavailable},
		{"RateLimit", ErrorTypeRateLimit, http.StatusTooManyRequests},
		{"ResourceExhausted", ErrorTypeResourceExhausted, http.StatusInsufficientStorage},
		{"Network", ErrorTypeNetwork, http.StatusServiceUnavailable},
		{"Cloud", ErrorTypeCloud, http.StatusBadGateway},
		{"HTTP", ErrorTypeHTTP, http.StatusInternalServerError},
		{"GRPC", ErrorTypeGRPC, http.StatusInternalServerError},
		{"GraphQL", ErrorTypeGraphQL, http.StatusBadRequest},
		{"WebSocket", ErrorTypeWebSocket, http.StatusInternalServerError},
		{"Unknown", ErrorType("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.DefaultStatusCode()
			if result != tt.expected {
				t.Errorf("ErrorType.DefaultStatusCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorType_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		// Non-retryable errors
		{"Repository", ErrorTypeRepository, false},
		{"Database", ErrorTypeDatabase, false},
		{"Cache", ErrorTypeCache, false},
		{"Migration", ErrorTypeMigration, false},
		{"Serialization", ErrorTypeSerialization, false},
		{"Validation", ErrorTypeValidation, false},
		{"BadRequest", ErrorTypeBadRequest, false},
		{"NotFound", ErrorTypeNotFound, false},
		{"Conflict", ErrorTypeConflict, false},
		{"Unprocessable", ErrorTypeUnprocessable, false},
		{"Unsupported", ErrorTypeUnsupported, false},
		{"BusinessRule", ErrorTypeBusinessRule, false},
		{"Authentication", ErrorTypeAuthentication, false},
		{"Authorization", ErrorTypeAuthorization, false},
		{"Internal", ErrorTypeInternal, false},
		{"Configuration", ErrorTypeConfiguration, false},

		// Retryable errors
		{"Dependency", ErrorTypeDependency, false},
		{"ExternalService", ErrorTypeExternalService, true},
		{"Timeout", ErrorTypeTimeout, true},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, true},
		{"RateLimit", ErrorTypeRateLimit, true},
		{"ResourceExhausted", ErrorTypeResourceExhausted, true},
		{"Network", ErrorTypeNetwork, true},
		{"Cloud", ErrorTypeCloud, true},
		{"HTTP", ErrorTypeHTTP, false},
		{"GRPC", ErrorTypeGRPC, false},
		{"GraphQL", ErrorTypeGraphQL, false},
		{"WebSocket", ErrorTypeWebSocket, false},
		{"Unknown", ErrorType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.IsRetryable()
			if result != tt.expected {
				t.Errorf("ErrorType.IsRetryable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorType_IsTemporary(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		// Non-temporary errors
		{"Repository", ErrorTypeRepository, false},
		{"Database", ErrorTypeDatabase, false},
		{"Cache", ErrorTypeCache, false},
		{"Migration", ErrorTypeMigration, false},
		{"Serialization", ErrorTypeSerialization, false},
		{"Validation", ErrorTypeValidation, false},
		{"BadRequest", ErrorTypeBadRequest, false},
		{"NotFound", ErrorTypeNotFound, false},
		{"Conflict", ErrorTypeConflict, false},
		{"Unprocessable", ErrorTypeUnprocessable, false},
		{"Unsupported", ErrorTypeUnsupported, false},
		{"BusinessRule", ErrorTypeBusinessRule, false},
		{"Authentication", ErrorTypeAuthentication, false},
		{"Authorization", ErrorTypeAuthorization, false},
		{"Internal", ErrorTypeInternal, false},
		{"Configuration", ErrorTypeConfiguration, false},

		// Temporary errors
		{"Dependency", ErrorTypeDependency, false},
		{"ExternalService", ErrorTypeExternalService, true},
		{"Timeout", ErrorTypeTimeout, true},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, true},
		{"RateLimit", ErrorTypeRateLimit, true},
		{"ResourceExhausted", ErrorTypeResourceExhausted, true},
		{"Network", ErrorTypeNetwork, true},
		{"Cloud", ErrorTypeCloud, true},
		{"HTTP", ErrorTypeHTTP, false},
		{"GRPC", ErrorTypeGRPC, false},
		{"GraphQL", ErrorTypeGraphQL, false},
		{"WebSocket", ErrorTypeWebSocket, false},
		{"Unknown", ErrorType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.IsTemporary()
			if result != tt.expected {
				t.Errorf("ErrorType.IsTemporary() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorType_DefaultSeverity(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  ErrorSeverity
	}{
		{"Repository", ErrorTypeRepository, SeverityMedium},
		{"Database", ErrorTypeDatabase, SeverityCritical},
		{"Cache", ErrorTypeCache, SeverityMedium},
		{"Migration", ErrorTypeMigration, SeverityMedium},
		{"Serialization", ErrorTypeSerialization, SeverityMedium},
		{"Validation", ErrorTypeValidation, SeverityLow},
		{"BadRequest", ErrorTypeBadRequest, SeverityLow},
		{"NotFound", ErrorTypeNotFound, SeverityLow},
		{"Conflict", ErrorTypeConflict, SeverityLow},
		{"Unprocessable", ErrorTypeUnprocessable, SeverityMedium},
		{"Unsupported", ErrorTypeUnsupported, SeverityMedium},
		{"BusinessRule", ErrorTypeBusinessRule, SeverityMedium},
		{"Authentication", ErrorTypeAuthentication, SeverityMedium},
		{"Authorization", ErrorTypeAuthorization, SeverityMedium},
		{"Internal", ErrorTypeInternal, SeverityCritical},
		{"Configuration", ErrorTypeConfiguration, SeverityMedium},
		{"Dependency", ErrorTypeDependency, SeverityMedium},
		{"ExternalService", ErrorTypeExternalService, SeverityHigh},
		{"Timeout", ErrorTypeTimeout, SeverityHigh},
		{"CircuitBreaker", ErrorTypeCircuitBreaker, SeverityHigh},
		{"RateLimit", ErrorTypeRateLimit, SeverityHigh},
		{"ResourceExhausted", ErrorTypeResourceExhausted, SeverityMedium},
		{"Network", ErrorTypeNetwork, SeverityMedium},
		{"Cloud", ErrorTypeCloud, SeverityMedium},
		{"HTTP", ErrorTypeHTTP, SeverityMedium},
		{"GRPC", ErrorTypeGRPC, SeverityMedium},
		{"GraphQL", ErrorTypeGraphQL, SeverityMedium},
		{"WebSocket", ErrorTypeWebSocket, SeverityMedium},
		{"Unknown", ErrorType("unknown"), SeverityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.DefaultSeverity()
			if result != tt.expected {
				t.Errorf("ErrorType.DefaultSeverity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		expected string
	}{
		{"Low", SeverityLow, "low"},
		{"Medium", SeverityMedium, "medium"},
		{"High", SeverityHigh, "high"},
		{"Critical", SeverityCritical, "critical"},
		{"Unknown", ErrorSeverity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.severity.String()
			if result != tt.expected {
				t.Errorf("ErrorSeverity.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCommonErrorCode(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedType   ErrorType
		expectedMsg    string
		expectedStatus int
		expectedSev    ErrorSeverity
		expectedRetry  bool
		expectedFound  bool
	}{
		{
			name:           "E001 Validation",
			code:           "E001",
			expectedType:   ErrorTypeValidation,
			expectedMsg:    "Validation failed",
			expectedStatus: http.StatusBadRequest,
			expectedSev:    SeverityLow,
			expectedRetry:  false,
			expectedFound:  true,
		},
		{
			name:           "E002 NotFound",
			code:           "E002",
			expectedType:   ErrorTypeNotFound,
			expectedMsg:    "Resource not found",
			expectedStatus: http.StatusNotFound,
			expectedSev:    SeverityLow,
			expectedRetry:  false,
			expectedFound:  true,
		},
		{
			name:           "E003 Conflict",
			code:           "E003",
			expectedType:   ErrorTypeConflict,
			expectedMsg:    "Resource already exists",
			expectedStatus: http.StatusConflict,
			expectedSev:    SeverityLow,
			expectedRetry:  false,
			expectedFound:  true,
		},
		{
			name:           "E007 Internal",
			code:           "E007",
			expectedType:   ErrorTypeInternal,
			expectedMsg:    "Internal server error",
			expectedStatus: http.StatusInternalServerError,
			expectedSev:    SeverityCritical,
			expectedRetry:  false,
			expectedFound:  true,
		},
		{
			name:           "E010 RateLimit",
			code:           "E010",
			expectedType:   ErrorTypeRateLimit,
			expectedMsg:    "Rate limit exceeded",
			expectedStatus: http.StatusTooManyRequests,
			expectedSev:    SeverityHigh,
			expectedRetry:  true,
			expectedFound:  true,
		},
		{
			name:           "Unknown Code",
			code:           "UNKNOWN",
			expectedType:   ErrorTypeInternal,
			expectedMsg:    "Unknown error",
			expectedStatus: http.StatusInternalServerError,
			expectedSev:    SeverityMedium,
			expectedRetry:  false,
			expectedFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, msg, status, severity, retry, found := GetCommonErrorCode(tt.code)

			if errType != tt.expectedType {
				t.Errorf("GetCommonErrorCode() type = %v, want %v", errType, tt.expectedType)
			}
			if msg != tt.expectedMsg {
				t.Errorf("GetCommonErrorCode() message = %v, want %v", msg, tt.expectedMsg)
			}
			if status != tt.expectedStatus {
				t.Errorf("GetCommonErrorCode() status = %v, want %v", status, tt.expectedStatus)
			}
			if severity != tt.expectedSev {
				t.Errorf("GetCommonErrorCode() severity = %v, want %v", severity, tt.expectedSev)
			}
			if retry != tt.expectedRetry {
				t.Errorf("GetCommonErrorCode() retry = %v, want %v", retry, tt.expectedRetry)
			}
			if found != tt.expectedFound {
				t.Errorf("GetCommonErrorCode() found = %v, want %v", found, tt.expectedFound)
			}
		})
	}
}

// Testes de casos de borda
func TestErrorType_EdgeCases(t *testing.T) {
	t.Run("Empty ErrorType", func(t *testing.T) {
		et := ErrorType("")
		if et.IsValid() {
			t.Error("Empty ErrorType should not be valid")
		}
		if et.String() != "" {
			t.Errorf("Empty ErrorType.String() = %v, want empty string", et.String())
		}
	})

	t.Run("Very Long ErrorType", func(t *testing.T) {
		longType := ErrorType("very_long_error_type_that_should_not_be_valid_in_normal_circumstances")
		if longType.IsValid() {
			t.Error("Very long ErrorType should not be valid")
		}
	})
}

func TestErrorSeverity_EdgeCases(t *testing.T) {
	t.Run("Negative Severity", func(t *testing.T) {
		severity := ErrorSeverity(-1)
		result := severity.String()
		if result != "unknown" {
			t.Errorf("Negative ErrorSeverity.String() = %v, want 'unknown'", result)
		}
	})

	t.Run("Large Severity Value", func(t *testing.T) {
		severity := ErrorSeverity(1000)
		result := severity.String()
		if result != "unknown" {
			t.Errorf("Large ErrorSeverity.String() = %v, want 'unknown'", result)
		}
	})
}

// Benchmarks
func BenchmarkErrorType_String(b *testing.B) {
	et := ErrorTypeValidation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = et.String()
	}
}

func BenchmarkErrorType_IsValid(b *testing.B) {
	et := ErrorTypeValidation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = et.IsValid()
	}
}

func BenchmarkErrorType_StatusCode(b *testing.B) {
	et := ErrorTypeValidation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = et.DefaultStatusCode()
	}
}

func BenchmarkErrorSeverity_String(b *testing.B) {
	es := SeverityMedium
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = es.String()
	}
}

func BenchmarkGetCommonErrorCode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _, _, _ = GetCommonErrorCode("E001")
	}
}
