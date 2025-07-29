package domainerrors

import (
	"errors"
	"testing"
	"time"
)

func TestNewValidationError(t *testing.T) {
	cause := errors.New("validation failed")

	err := NewValidationError("VALIDATION_FAILED", "validation error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "VALIDATION_FAILED" {
		t.Errorf("expected code 'VALIDATION_FAILED', got %q", err.Code)
	}

	if err.Message != "validation error" {
		t.Errorf("expected message 'validation error', got %q", err.Message)
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("expected type %q, got %q", ErrorTypeValidation, err.Type)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}

	if err.Fields == nil {
		t.Error("expected fields to be initialized")
	}
}

func TestValidationError_WithField(t *testing.T) {
	err := NewValidationError("VALIDATION_FAILED", "validation error", nil)

	result := err.WithField("email", "invalid format")
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.Fields["email"]) != 1 {
		t.Errorf("expected 1 error for email, got %d", len(err.Fields["email"]))
	}

	if err.Fields["email"][0] != "invalid format" {
		t.Errorf("expected 'invalid format', got %q", err.Fields["email"][0])
	}

	// Test adding multiple errors for same field
	err.WithField("email", "required")
	if len(err.Fields["email"]) != 2 {
		t.Errorf("expected 2 errors for email, got %d", len(err.Fields["email"]))
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("NOT_FOUND", "user not found", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got %q", err.Code)
	}

	if err.Message != "user not found" {
		t.Errorf("expected message 'user not found', got %q", err.Message)
	}

	if err.Type != ErrorTypeNotFound {
		t.Errorf("expected type %q, got %q", ErrorTypeNotFound, err.Type)
	}
}

func TestNotFoundError_WithResource(t *testing.T) {
	err := NewNotFoundError("NOT_FOUND", "resource not found", nil)

	result := err.WithResource("user", "123")
	if result != err {
		t.Error("expected same instance")
	}

	if err.ResourceType != "user" {
		t.Errorf("expected resource type 'user', got %q", err.ResourceType)
	}

	if err.ResourceID != "123" {
		t.Errorf("expected resource ID '123', got %q", err.ResourceID)
	}
}

func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError("INSUFFICIENT_FUNDS", "insufficient funds")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("expected code 'INSUFFICIENT_FUNDS', got %q", err.Code)
	}

	if err.Message != "insufficient funds" {
		t.Errorf("expected message 'insufficient funds', got %q", err.Message)
	}

	if err.Type != ErrorTypeBusinessRule {
		t.Errorf("expected type %q, got %q", ErrorTypeBusinessRule, err.Type)
	}

	if err.BusinessCode != "INSUFFICIENT_FUNDS" {
		t.Errorf("expected business code 'INSUFFICIENT_FUNDS', got %q", err.BusinessCode)
	}

	if err.Rules == nil {
		t.Error("expected rules to be initialized")
	}
}

func TestBusinessError_WithRule(t *testing.T) {
	err := NewBusinessError("BIZ_001", "business error")

	result := err.WithRule("minimum balance required")
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(err.Rules))
	}

	if err.Rules[0] != "minimum balance required" {
		t.Errorf("expected 'minimum balance required', got %q", err.Rules[0])
	}
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection failed")

	err := NewDatabaseError("DATABASE_ERROR", "database error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "DATABASE_ERROR" {
		t.Errorf("expected code 'DATABASE_ERROR', got %q", err.Code)
	}

	if err.Message != "database error" {
		t.Errorf("expected message 'database error', got %q", err.Message)
	}

	if err.Type != ErrorTypeDatabase {
		t.Errorf("expected type %q, got %q", ErrorTypeDatabase, err.Type)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestDatabaseError_WithOperation(t *testing.T) {
	err := NewDatabaseError("DATABASE_ERROR", "database error", nil)

	result := err.WithOperation("SELECT", "users")
	if result != err {
		t.Error("expected same instance")
	}

	if err.Operation != "SELECT" {
		t.Errorf("expected operation 'SELECT', got %q", err.Operation)
	}

	if err.Table != "users" {
		t.Errorf("expected table 'users', got %q", err.Table)
	}
}

func TestDatabaseError_WithQuery(t *testing.T) {
	err := NewDatabaseError("DATABASE_ERROR", "database error", nil)

	result := err.WithQuery("SELECT * FROM users")
	if result != err {
		t.Error("expected same instance")
	}

	if err.Query != "SELECT * FROM users" {
		t.Errorf("expected query 'SELECT * FROM users', got %q", err.Query)
	}
}

func TestNewExternalServiceError(t *testing.T) {
	cause := errors.New("timeout")

	err := NewExternalServiceError("EXTERNAL_SERVICE_ERROR", "payment-service", "service error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "EXTERNAL_SERVICE_ERROR" {
		t.Errorf("expected code 'EXTERNAL_SERVICE_ERROR', got %q", err.Code)
	}

	if err.Message != "service error" {
		t.Errorf("expected message 'service error', got %q", err.Message)
	}

	if err.Type != ErrorTypeExternalService {
		t.Errorf("expected type %q, got %q", ErrorTypeExternalService, err.Type)
	}

	if err.Service != "payment-service" {
		t.Errorf("expected service 'payment-service', got %q", err.Service)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestExternalServiceError_WithEndpoint(t *testing.T) {
	err := NewExternalServiceError("EXTERNAL_SERVICE_ERROR", "service", "error", nil)

	result := err.WithEndpoint("/api/v1/users")
	if result != err {
		t.Error("expected same instance")
	}

	if err.Endpoint != "/api/v1/users" {
		t.Errorf("expected endpoint '/api/v1/users', got %q", err.Endpoint)
	}
}

func TestExternalServiceError_WithResponse(t *testing.T) {
	err := NewExternalServiceError("EXTERNAL_SERVICE_ERROR", "service", "error", nil)

	result := err.WithResponse(404, "Not Found")
	if result != err {
		t.Error("expected same instance")
	}

	if err.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", err.StatusCode)
	}

	if err.Response != "Not Found" {
		t.Errorf("expected response 'Not Found', got %q", err.Response)
	}
}

func TestNewTimeoutError(t *testing.T) {
	cause := errors.New("context deadline exceeded")

	err := NewTimeoutError("TIMEOUT_ERROR", "database-query", "timeout error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "TIMEOUT_ERROR" {
		t.Errorf("expected code 'TIMEOUT_ERROR', got %q", err.Code)
	}

	if err.Message != "timeout error" {
		t.Errorf("expected message 'timeout error', got %q", err.Message)
	}

	if err.Type != ErrorTypeTimeout {
		t.Errorf("expected type %q, got %q", ErrorTypeTimeout, err.Type)
	}

	if err.Operation != "database-query" {
		t.Errorf("expected operation 'database-query', got %q", err.Operation)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestTimeoutError_WithDuration(t *testing.T) {
	err := NewTimeoutError("TIMEOUT_ERROR", "operation", "timeout", nil)

	duration := 5 * time.Second
	timeout := 10 * time.Second

	result := err.WithDuration(duration, timeout)
	if result != err {
		t.Error("expected same instance")
	}

	if err.Duration != duration {
		t.Errorf("expected duration %v, got %v", duration, err.Duration)
	}

	if err.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, err.Timeout)
	}
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("RATE_LIMIT_ERROR", "rate limit exceeded")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "RATE_LIMIT_ERROR" {
		t.Errorf("expected code 'RATE_LIMIT_ERROR', got %q", err.Code)
	}

	if err.Message != "rate limit exceeded" {
		t.Errorf("expected message 'rate limit exceeded', got %q", err.Message)
	}

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("expected type %q, got %q", ErrorTypeRateLimit, err.Type)
	}
}

func TestRateLimitError_WithRateLimit(t *testing.T) {
	err := NewRateLimitError("RATE_LIMIT_ERROR", "rate limit exceeded")

	result := err.WithRateLimit(100, 50, "2025-01-01T00:00:00Z", "3600s")
	if result != err {
		t.Error("expected same instance")
	}

	if err.Limit != 100 {
		t.Errorf("expected limit 100, got %d", err.Limit)
	}

	if err.Remaining != 50 {
		t.Errorf("expected remaining 50, got %d", err.Remaining)
	}

	if err.ResetTime != "2025-01-01T00:00:00Z" {
		t.Errorf("expected reset time '2025-01-01T00:00:00Z', got %q", err.ResetTime)
	}

	if err.Window != "3600s" {
		t.Errorf("expected window '3600s', got %q", err.Window)
	}
}

func TestNewCircuitBreakerError(t *testing.T) {
	err := NewCircuitBreakerError("CIRCUIT_BREAKER_ERROR", "payment-service", "circuit breaker open")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "CIRCUIT_BREAKER_ERROR" {
		t.Errorf("expected code 'CIRCUIT_BREAKER_ERROR', got %q", err.Code)
	}

	if err.Message != "circuit breaker open" {
		t.Errorf("expected message 'circuit breaker open', got %q", err.Message)
	}

	if err.Type != ErrorTypeCircuitBreaker {
		t.Errorf("expected type %q, got %q", ErrorTypeCircuitBreaker, err.Type)
	}

	if err.CircuitName != "payment-service" {
		t.Errorf("expected circuit name 'payment-service', got %q", err.CircuitName)
	}
}

func TestCircuitBreakerError_WithCircuitState(t *testing.T) {
	err := NewCircuitBreakerError("CIRCUIT_BREAKER_ERROR", "service", "circuit breaker open")

	result := err.WithCircuitState("OPEN", 5)
	if result != err {
		t.Error("expected same instance")
	}

	if err.State != "OPEN" {
		t.Errorf("expected state 'OPEN', got %q", err.State)
	}

	if err.Failures != 5 {
		t.Errorf("expected failures 5, got %d", err.Failures)
	}
}

func TestNewInvalidSchemaError(t *testing.T) {
	err := NewInvalidSchemaError("INVALID_SCHEMA", "invalid schema")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "INVALID_SCHEMA" {
		t.Errorf("expected code 'INVALID_SCHEMA', got %q", err.Code)
	}

	if err.Message != "invalid schema" {
		t.Errorf("expected message 'invalid schema', got %q", err.Message)
	}

	if err.Type != ErrorTypeInvalidSchema {
		t.Errorf("expected type %q, got %q", ErrorTypeInvalidSchema, err.Type)
	}

	if err.Details == nil {
		t.Error("expected details to be initialized")
	}
}

func TestInvalidSchemaError_WithSchemaInfo(t *testing.T) {
	err := NewInvalidSchemaError("INVALID_SCHEMA", "invalid schema")

	result := err.WithSchemaInfo("user-schema", "v1.0")
	if result != err {
		t.Error("expected same instance")
	}

	if err.SchemaName != "user-schema" {
		t.Errorf("expected schema name 'user-schema', got %q", err.SchemaName)
	}

	if err.Version != "v1.0" {
		t.Errorf("expected version 'v1.0', got %q", err.Version)
	}
}

func TestInvalidSchemaError_WithSchemaDetails(t *testing.T) {
	err := NewInvalidSchemaError("INVALID_SCHEMA", "invalid schema")

	details := map[string][]string{
		"name": {"required"},
		"age":  {"must be positive"},
	}

	result := err.WithSchemaDetails(details)
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(err.Details))
	}

	if err.Details["name"][0] != "required" {
		t.Errorf("expected 'required', got %q", err.Details["name"][0])
	}
}

func TestNewServerError(t *testing.T) {
	cause := errors.New("internal error")

	err := NewServerError("SERVER_ERROR", "server error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "SERVER_ERROR" {
		t.Errorf("expected code 'SERVER_ERROR', got %q", err.Code)
	}

	if err.Message != "server error" {
		t.Errorf("expected message 'server error', got %q", err.Message)
	}

	if err.Type != ErrorTypeServer {
		t.Errorf("expected type %q, got %q", ErrorTypeServer, err.Type)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestServerError_WithRequestInfo(t *testing.T) {
	err := NewServerError("SERVER_ERROR", "server error", nil)

	result := err.WithRequestInfo("req-123", "corr-456")
	if result != err {
		t.Error("expected same instance")
	}

	if err.RequestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got %q", err.RequestID)
	}

	if err.CorrelationID != "corr-456" {
		t.Errorf("expected correlation ID 'corr-456', got %q", err.CorrelationID)
	}
}

func TestNewUnprocessableEntityError(t *testing.T) {
	err := NewUnprocessableEntityError("UNPROCESSABLE_ENTITY", "entity cannot be processed")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "UNPROCESSABLE_ENTITY" {
		t.Errorf("expected code 'UNPROCESSABLE_ENTITY', got %q", err.Code)
	}

	if err.Message != "entity cannot be processed" {
		t.Errorf("expected message 'entity cannot be processed', got %q", err.Message)
	}

	if err.Type != ErrorTypeUnprocessable {
		t.Errorf("expected type %q, got %q", ErrorTypeUnprocessable, err.Type)
	}

	if err.ValidationErrors == nil {
		t.Error("expected validation errors to be initialized")
	}

	if err.BusinessRules == nil {
		t.Error("expected business rules to be initialized")
	}
}

func TestUnprocessableEntityError_WithEntityInfo(t *testing.T) {
	err := NewUnprocessableEntityError("UNPROCESSABLE_ENTITY", "entity error")

	result := err.WithEntityInfo("User", "user-123")
	if result != err {
		t.Error("expected same instance")
	}

	if err.EntityType != "User" {
		t.Errorf("expected entity type 'User', got %q", err.EntityType)
	}

	if err.EntityID != "user-123" {
		t.Errorf("expected entity ID 'user-123', got %q", err.EntityID)
	}
}

func TestUnprocessableEntityError_WithValidationErrors(t *testing.T) {
	err := NewUnprocessableEntityError("UNPROCESSABLE_ENTITY", "entity error")

	validationErrors := map[string][]string{
		"email": {"invalid format"},
		"age":   {"must be positive"},
	}

	result := err.WithValidationErrors(validationErrors)
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.ValidationErrors) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(err.ValidationErrors))
	}

	if err.ValidationErrors["email"][0] != "invalid format" {
		t.Errorf("expected 'invalid format', got %q", err.ValidationErrors["email"][0])
	}
}

func TestUnprocessableEntityError_WithBusinessRuleViolation(t *testing.T) {
	err := NewUnprocessableEntityError("UNPROCESSABLE_ENTITY", "entity error")

	result := err.WithBusinessRuleViolation("minimum age required")
	if result != err {
		t.Error("expected same instance")
	}

	if len(err.BusinessRules) != 1 {
		t.Errorf("expected 1 business rule, got %d", len(err.BusinessRules))
	}

	if err.BusinessRules[0] != "minimum age required" {
		t.Errorf("expected 'minimum age required', got %q", err.BusinessRules[0])
	}
}

func TestNewServiceUnavailableError(t *testing.T) {
	cause := errors.New("service down")

	err := NewServiceUnavailableError("SERVICE_UNAVAILABLE", "payment-service", "service unavailable", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "SERVICE_UNAVAILABLE" {
		t.Errorf("expected code 'SERVICE_UNAVAILABLE', got %q", err.Code)
	}

	if err.Message != "service unavailable" {
		t.Errorf("expected message 'service unavailable', got %q", err.Message)
	}

	if err.Type != ErrorTypeServiceUnavailable {
		t.Errorf("expected type %q, got %q", ErrorTypeServiceUnavailable, err.Type)
	}

	if err.ServiceName != "payment-service" {
		t.Errorf("expected service name 'payment-service', got %q", err.ServiceName)
	}

	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

func TestServiceUnavailableError_WithRetryInfo(t *testing.T) {
	err := NewServiceUnavailableError("SERVICE_UNAVAILABLE", "service", "unavailable", nil)

	result := err.WithRetryInfo("30s")
	if result != err {
		t.Error("expected same instance")
	}

	if err.RetryAfter != "30s" {
		t.Errorf("expected retry after '30s', got %q", err.RetryAfter)
	}
}

func TestNewWorkflowError(t *testing.T) {
	err := NewWorkflowError("WORKFLOW_ERROR", "order-process", "payment", "payment failed")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "WORKFLOW_ERROR" {
		t.Errorf("expected code 'WORKFLOW_ERROR', got %q", err.Code)
	}

	if err.Message != "payment failed" {
		t.Errorf("expected message 'payment failed', got %q", err.Message)
	}

	if err.Type != ErrorTypeWorkflow {
		t.Errorf("expected type %q, got %q", ErrorTypeWorkflow, err.Type)
	}

	if err.WorkflowName != "order-process" {
		t.Errorf("expected workflow name 'order-process', got %q", err.WorkflowName)
	}

	if err.StepName != "payment" {
		t.Errorf("expected step name 'payment', got %q", err.StepName)
	}
}

func TestWorkflowError_WithStateInfo(t *testing.T) {
	err := NewWorkflowError("WORKFLOW_ERROR", "workflow", "step", "error")

	result := err.WithStateInfo("pending", "completed")
	if result != err {
		t.Error("expected same instance")
	}

	if err.CurrentState != "pending" {
		t.Errorf("expected current state 'pending', got %q", err.CurrentState)
	}

	if err.ExpectedState != "completed" {
		t.Errorf("expected expected state 'completed', got %q", err.ExpectedState)
	}
}

// Additional error types tests
func TestNewinfraestructureError(t *testing.T) {
	cause := errors.New("connection failed")

	err := NewinfraestructureError("infraestructure_ERROR", "database", "infraestructure error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "infraestructure_ERROR" {
		t.Errorf("expected code 'infraestructure_ERROR', got %q", err.Code)
	}

	if err.Component != "database" {
		t.Errorf("expected component 'database', got %q", err.Component)
	}
}

func TestNewAuthenticationError(t *testing.T) {
	cause := errors.New("invalid token")

	err := NewAuthenticationError("AUTHENTICATION_ERROR", "authentication failed", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "AUTHENTICATION_ERROR" {
		t.Errorf("expected code 'AUTHENTICATION_ERROR', got %q", err.Code)
	}

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("expected type %q, got %q", ErrorTypeAuthentication, err.Type)
	}
}

func TestNewAuthorizationError(t *testing.T) {
	cause := errors.New("insufficient permissions")

	err := NewAuthorizationError("AUTHORIZATION_ERROR", "authorization failed", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "AUTHORIZATION_ERROR" {
		t.Errorf("expected code 'AUTHORIZATION_ERROR', got %q", err.Code)
	}

	if err.Type != ErrorTypeAuthorization {
		t.Errorf("expected type %q, got %q", ErrorTypeAuthorization, err.Type)
	}
}

func TestNewSecurityError(t *testing.T) {
	err := NewSecurityError("SECURITY_ERROR", "security threat detected")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "SECURITY_ERROR" {
		t.Errorf("expected code 'SECURITY_ERROR', got %q", err.Code)
	}

	if err.Type != ErrorTypeSecurity {
		t.Errorf("expected type %q, got %q", ErrorTypeSecurity, err.Type)
	}
}

func TestNewResourceExhaustedError(t *testing.T) {
	err := NewResourceExhaustedError("RESOURCE_EXHAUSTED", "memory", "memory exhausted")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "RESOURCE_EXHAUSTED" {
		t.Errorf("expected code 'RESOURCE_EXHAUSTED', got %q", err.Code)
	}

	if err.Resource != "memory" {
		t.Errorf("expected resource 'memory', got %q", err.Resource)
	}

	if err.Type != ErrorTypeResourceExhausted {
		t.Errorf("expected type %q, got %q", ErrorTypeResourceExhausted, err.Type)
	}
}

func TestNewDependencyError(t *testing.T) {
	cause := errors.New("dependency unavailable")

	err := NewDependencyError("DEPENDENCY_ERROR", "elasticsearch", "dependency error", cause)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "DEPENDENCY_ERROR" {
		t.Errorf("expected code 'DEPENDENCY_ERROR', got %q", err.Code)
	}

	if err.Dependency != "elasticsearch" {
		t.Errorf("expected dependency 'elasticsearch', got %q", err.Dependency)
	}

	if err.Type != ErrorTypeDependency {
		t.Errorf("expected type %q, got %q", ErrorTypeDependency, err.Type)
	}
}

// Benchmark tests for error types
func BenchmarkNewValidationError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewValidationError("VALIDATION_ERROR", "validation error", nil)
	}
}

func BenchmarkNewBusinessError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewBusinessError("BIZ_001", "business error")
	}
}

func BenchmarkNewDatabaseError(b *testing.B) {
	cause := errors.New("connection failed")
	for i := 0; i < b.N; i++ {
		_ = NewDatabaseError("DATABASE_ERROR", "database error", cause)
	}
}

func BenchmarkNewExternalServiceError(b *testing.B) {
	cause := errors.New("timeout")
	for i := 0; i < b.N; i++ {
		_ = NewExternalServiceError("EXTERNAL_SERVICE_ERROR", "service", "error", cause)
	}
}
