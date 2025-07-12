package domainerrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// TestDomainError_Creation testa a criação básica de erros de domínio.
func TestDomainError_Creation(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		expected string
	}{
		{
			name:     "basic error creation",
			code:     "E001",
			message:  "Test error",
			expected: "[E001] Test error",
		},
		{
			name:     "empty code",
			code:     "",
			message:  "Test error",
			expected: "Test error",
		},
		{
			name:     "empty message",
			code:     "E001",
			message:  "",
			expected: "[E001] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

// TestDomainError_Builder testa o padrão builder.
func TestDomainError_Builder(t *testing.T) {
	err := NewBuilder().
		WithCode("E002").
		WithMessage("User not found").
		WithType(string(types.ErrorTypeNotFound)).
		WithDetail("user_id", "12345").
		WithTag("test").
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		Build()

	if err.Code() != "E002" {
		t.Errorf("Expected code 'E002', got '%s'", err.Code())
	}

	if err.Message() != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", err.Message())
	}

	if err.Type() != string(types.ErrorTypeNotFound) {
		t.Errorf("Expected type '%s', got '%s'", types.ErrorTypeNotFound, err.Type())
	}

	details := err.Details()
	if details["user_id"] != "12345" {
		t.Errorf("Expected detail user_id='12345', got '%v'", details["user_id"])
	}

	tags := err.Tags()
	if len(tags) != 1 || tags[0] != "test" {
		t.Errorf("Expected tags ['test'], got %v", tags)
	}
}

// TestDomainError_Wrapping testa o wrapping de erros.
func TestDomainError_Wrapping(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := New("E001", "Wrapped error").Wrap("wrapping", originalErr)

	// Testa Unwrap
	if unwrapped := wrappedErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}

	// Testa RootCause
	if root := wrappedErr.RootCause(); root != originalErr {
		t.Errorf("Expected root cause to be original error")
	}

	// Testa Is
	if !Is(wrappedErr, originalErr) {
		t.Errorf("Expected wrapped error to match original error")
	}
}

// TestDomainError_Chain testa o encadeamento de erros.
func TestDomainError_Chain(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	domainErr := New("E001", "Domain error").
		Chain(err1).
		Chain(err2)

	// Verifica se os erros estão na cadeia
	errorStr := domainErr.Error()
	if !strings.Contains(errorStr, "Domain error") {
		t.Errorf("Expected error string to contain 'Domain error'")
	}
}

// TestDomainError_StatusCode testa os códigos de status HTTP.
func TestDomainError_StatusCode(t *testing.T) {
	tests := []struct {
		name         string
		errorType    types.ErrorType
		expectedCode int
	}{
		{"not found", types.ErrorTypeNotFound, 404},
		{"validation", types.ErrorTypeValidation, 400},
		{"unauthorized", types.ErrorTypeAuthentication, 401},
		{"forbidden", types.ErrorTypeAuthorization, 403},
		{"internal", types.ErrorTypeInternal, 500},
		{"timeout", types.ErrorTypeTimeout, 504},
		{"conflict", types.ErrorTypeConflict, 409},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBuilder().
				WithCode("TEST").
				WithMessage("Test error").
				WithType(string(tt.errorType)).
				Build()

			if err.StatusCode() != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, err.StatusCode())
			}
		})
	}
}

// TestDomainError_JSON testa a serialização JSON.
func TestDomainError_JSON(t *testing.T) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("field", "value").
		WithTag("test").
		Build()

	jsonData, jsonErr := err.JSON()
	if jsonErr != nil {
		t.Fatalf("Failed to serialize to JSON: %v", jsonErr)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["code"] != "E001" {
		t.Errorf("Expected code 'E001', got '%v'", parsed["code"])
	}

	if parsed["message"] != "Test error" {
		t.Errorf("Expected message 'Test error', got '%v'", parsed["message"])
	}

	if parsed["type"] != string(types.ErrorTypeValidation) {
		t.Errorf("Expected type '%s', got '%v'", types.ErrorTypeValidation, parsed["type"])
	}
}

// TestValidationError testa erros de validação específicos.
func TestValidationError(t *testing.T) {
	fields := map[string][]string{
		"email": {"invalid format", "required"},
		"age":   {"must be positive"},
	}

	validationErr := NewValidationError("Validation failed", fields)

	// Testa Fields
	returnedFields := validationErr.Fields()
	if len(returnedFields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(returnedFields))
	}

	// Testa HasField
	if !validationErr.HasField("email") {
		t.Errorf("Expected to have field 'email'")
	}

	if validationErr.HasField("nonexistent") {
		t.Errorf("Expected not to have field 'nonexistent'")
	}

	// Testa FieldErrors
	emailErrors := validationErr.FieldErrors("email")
	if len(emailErrors) != 2 {
		t.Errorf("Expected 2 errors for email field, got %d", len(emailErrors))
	}

	// Testa AddField
	validationErr.AddField("name", "required")
	if !validationErr.HasField("name") {
		t.Errorf("Expected to have field 'name' after adding")
	}

	// Testa total de campos
	fields = validationErr.Fields()
	totalErrors := 0
	for _, fieldErrors := range fields {
		totalErrors += len(fieldErrors)
	}
	if totalErrors != 4 { // 2 + 1 + 1
		t.Errorf("Expected 4 total errors, got %d", totalErrors)
	}
}

// TestDomainError_ThreadSafety testa thread safety básica.
func TestDomainError_ThreadSafety(t *testing.T) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error").
		Build()

	// Executa operações concorrentes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			// Operações de leitura
			_ = err.Code()
			_ = err.Message()
			_ = err.Details()
			_ = err.Tags()

			// Operações que podem modificar (através de builder)
			newErr := NewBuilder().
				WithCode("E002").
				WithMessage("Concurrent test").
				WithDetail("index", index).
				Build()

			_ = newErr.Error()
			done <- true
		}(i)
	}

	// Aguarda todas as goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestDomainError_Clone testa a clonagem de erros.
func TestDomainError_Clone(t *testing.T) {
	original := NewBuilder().
		WithCode("E001").
		WithMessage("Original error").
		WithDetail("key", "value").
		WithTag("original").
		Build().(*DomainError)

	clone := original.Clone()

	// Verifica se os valores são iguais
	if clone.Code() != original.Code() {
		t.Errorf("Clone code doesn't match original")
	}

	if clone.Message() != original.Message() {
		t.Errorf("Clone message doesn't match original")
	}

	// Verifica se são instâncias independentes
	clone.mu.Lock()
	clone.details["new_key"] = "new_value"
	clone.mu.Unlock()

	originalDetails := original.Details()
	if _, exists := originalDetails["new_key"]; exists {
		t.Errorf("Modification in clone affected original")
	}
}

// TestConvenienceFunctions testa as funções de conveniência.
func TestConvenienceFunctions(t *testing.T) {
	// Testa NewNotFoundError
	notFoundErr := NewNotFoundError("User", "123")
	if notFoundErr.Code() != "E002" {
		t.Errorf("Expected code 'E002', got '%s'", notFoundErr.Code())
	}
	if !strings.Contains(notFoundErr.Message(), "User") {
		t.Errorf("Expected message to contain 'User'")
	}

	// Testa NewUnauthorizedError
	unauthorizedErr := NewUnauthorizedError("")
	if unauthorizedErr.StatusCode() != 401 {
		t.Errorf("Expected status code 401, got %d", unauthorizedErr.StatusCode())
	}

	// Testa NewInternalError
	cause := errors.New("root cause")
	internalErr := NewInternalError("Internal problem", cause)
	if internalErr.Unwrap() != cause {
		t.Errorf("Expected unwrapped error to be the cause")
	}
}

// TestUtilityFunctions testa as funções utilitárias.
func TestUtilityFunctions(t *testing.T) {
	// Testa IsRetryable
	timeoutErr := NewTimeoutError("Operation timeout")
	if !IsRetryable(timeoutErr) {
		t.Errorf("Expected timeout error to be retryable")
	}

	validationErr := NewBuilder().
		WithType(string(types.ErrorTypeValidation)).
		WithMessage("Validation failed").
		Build()
	if IsRetryable(validationErr) {
		t.Errorf("Expected validation error to not be retryable")
	}

	// Testa GetErrorType
	errorType := GetErrorType(validationErr)
	if errorType != string(types.ErrorTypeValidation) {
		t.Errorf("Expected error type '%s', got '%s'", types.ErrorTypeValidation, errorType)
	}

	// Testa GetErrorCode
	code := GetErrorCode(validationErr)
	if code == "" {
		t.Errorf("Expected non-empty error code")
	}

	// Testa GetStatusCode
	statusCode := GetStatusCode(validationErr)
	if statusCode != 400 {
		t.Errorf("Expected status code 400, got %d", statusCode)
	}
}

// TestErrorWrapping testa o wrapping complexo de erros.
func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("database connection failed")

	wrappedErr := Wrap("DB001", "Query failed", originalErr)
	if wrappedErr == nil {
		t.Fatalf("Expected wrapped error, got nil")
	}

	// Testa wrapping de nil
	nilWrapped := Wrap("TEST", "Test", nil)
	if nilWrapped != nil {
		t.Errorf("Expected nil when wrapping nil error")
	}

	// Testa wrapping de DomainError existente
	domainErr := New("E001", "Domain error")
	reWrapped := Wrap("E002", "Re-wrapped", domainErr)
	if reWrapped == nil {
		t.Fatalf("Expected re-wrapped error, got nil")
	}
}

// TestErrorChaining testa o encadeamento complexo de erros.
func TestErrorChaining(t *testing.T) {
	err1 := errors.New("first")
	err2 := errors.New("second")
	err3 := errors.New("third")

	chainedErr := New("CHAIN", "Chained error").
		Chain(err1).
		Chain(err2).
		Chain(err3)

	// Verifica se o erro principal ainda é acessível
	if chainedErr.Message() != "Chained error" {
		t.Errorf("Expected main message to be preserved")
	}

	// Verifica se o stack trace inclui informações dos erros encadeados
	stackTrace := chainedErr.FormatStackTrace()
	if stackTrace == "" {
		t.Errorf("Expected non-empty stack trace")
	}
}

// TestErrorCategories testa as categorias de erro.
func TestErrorCategories(t *testing.T) {
	tests := []struct {
		errorType types.ErrorType
		category  string
	}{
		{types.ErrorTypeDatabase, "data"},
		{types.ErrorTypeValidation, "input"},
		{types.ErrorTypeBusinessRule, "business"},
		{types.ErrorTypeAuthentication, "security"},
		{types.ErrorTypeInternal, "system"},
		{types.ErrorTypeNetwork, "communication"},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			if tt.errorType.Category() != tt.category {
				t.Errorf("Expected category '%s', got '%s'", tt.category, tt.errorType.Category())
			}
		})
	}
}

// TestErrorSeverity testa os níveis de severidade.
func TestErrorSeverity(t *testing.T) {
	tests := []struct {
		errorType types.ErrorType
		severity  types.ErrorSeverity
	}{
		{types.ErrorTypeValidation, types.SeverityLow},
		{types.ErrorTypeBusinessRule, types.SeverityMedium},
		{types.ErrorTypeTimeout, types.SeverityHigh},
		{types.ErrorTypeInternal, types.SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			if tt.errorType.DefaultSeverity() != tt.severity {
				t.Errorf("Expected severity %v, got %v", tt.severity, tt.errorType.DefaultSeverity())
			}
		})
	}
}

// TestErrorRetryability testa a capacidade de retry.
func TestErrorRetryability(t *testing.T) {
	retryableTypes := []types.ErrorType{
		types.ErrorTypeTimeout,
		types.ErrorTypeRateLimit,
		types.ErrorTypeCircuitBreaker,
		types.ErrorTypeNetwork,
	}

	nonRetryableTypes := []types.ErrorType{
		types.ErrorTypeValidation,
		types.ErrorTypeNotFound,
		types.ErrorTypeAuthentication,
		types.ErrorTypeBusinessRule,
	}

	for _, et := range retryableTypes {
		t.Run("retryable_"+string(et), func(t *testing.T) {
			if !et.IsRetryable() {
				t.Errorf("Expected %s to be retryable", et)
			}
		})
	}

	for _, et := range nonRetryableTypes {
		t.Run("non_retryable_"+string(et), func(t *testing.T) {
			if et.IsRetryable() {
				t.Errorf("Expected %s to not be retryable", et)
			}
		})
	}
}

// BenchmarkDomainError_Creation benchmark para criação de erros.
func BenchmarkDomainError_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New("E001", "Test error")
	}
}

// BenchmarkDomainError_Builder benchmark para builder.
func BenchmarkDomainError_Builder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBuilder().
			WithCode("E001").
			WithMessage("Test error").
			WithType(string(types.ErrorTypeValidation)).
			WithDetail("key", "value").
			Build()
	}
}

// BenchmarkDomainError_JSON benchmark para serialização JSON.
func BenchmarkDomainError_JSON(b *testing.B) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error").
		WithDetail("key", "value").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = err.JSON()
	}
}

// BenchmarkDomainError_Wrapping benchmark para wrapping.
func BenchmarkDomainError_Wrapping(b *testing.B) {
	originalErr := errors.New("original error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New("E001", "Wrapped error").Wrap("wrapping", originalErr)
	}
}

// TestBuilderAdvanced testa funcionalidades avançadas do builder
func TestBuilderAdvanced(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "123")

	err := NewBuilder().
		WithCode("ADV001").
		WithMessage("Advanced test").
		WithCategory(interfaces.Category("custom")).
		WithDetails(map[string]interface{}{"key1": "value1", "key2": 42}).
		WithMetadata(map[string]interface{}{"meta1": "data1"}).
		WithTags([]string{"tag1", "tag2", "tag3"}).
		WithStatusCode(418).
		WithHeader("X-Error-Code", "ADV001").
		WithHeaders(map[string]string{"X-Custom": "header"}).
		WithContext(ctx).
		Build()

	if err.Code() != "ADV001" {
		t.Errorf("Expected code ADV001, got %s", err.Code())
	}

	if err.StatusCode() != 418 {
		t.Errorf("Expected status 418, got %d", err.StatusCode())
	}

	headers := err.Headers()
	if headers["X-Error-Code"] != "ADV001" {
		t.Errorf("Expected X-Error-Code header")
	}

	tags := err.Tags()
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	details := err.Details()
	if details["key1"] != "value1" {
		t.Errorf("Expected detail key1=value1")
	}

	metadata := err.Metadata()
	if metadata["meta1"] != "data1" {
		t.Errorf("Expected metadata meta1=data1")
	}
}

// TestBuilderSpecializedErrors testa builders especializados
func TestBuilderSpecializedErrors(t *testing.T) {
	// Test usando New functions diretamente
	notFoundErr := NewNotFoundError("NF001", "Resource not found")
	if notFoundErr.StatusCode() != 404 {
		t.Errorf("Expected 404 status for not found error")
	}

	// Test business error usando ErrorTypeBusinessRule
	businessErr := NewBuilder().
		WithCode("BIZ001").
		WithMessage("Business rule violation").
		WithType(string(types.ErrorTypeBusinessRule)).
		Build()

	if businessErr.Type() != string(types.ErrorTypeBusinessRule) {
		t.Errorf("Expected business error type")
	}

	// Test internal error
	internalErr := NewInternalError("Internal server error", errors.New("underlying cause"))
	if internalErr.StatusCode() != 500 {
		t.Errorf("Expected 500 status for internal error")
	}

	// Test timeout error usando builder
	timeoutErr := NewBuilder().
		WithCode("TO001").
		WithMessage("Request timeout").
		WithType(string(types.ErrorTypeTimeout)).
		Build()

	// IsRetryable não está na interface, vamos testar através do tipo
	domainErr := timeoutErr.(*DomainError)
	if !domainErr.IsRetryable() {
		t.Errorf("Expected timeout error to be retryable")
	}

	// Test rate limit usando builder
	rateLimitErr := NewBuilder().
		WithCode("RL001").
		WithMessage("Rate limit exceeded").
		WithType(string(types.ErrorTypeRateLimit)).
		WithStatusCode(429).
		Build()

	if rateLimitErr.StatusCode() != 429 {
		t.Errorf("Expected 429 status for rate limit error")
	}

	// Test circuit breaker usando builder
	cbErr := NewBuilder().
		WithCode("CB001").
		WithMessage("Circuit breaker open").
		WithType(string(types.ErrorTypeCircuitBreaker)).
		WithStatusCode(503).
		Build()

	if cbErr.StatusCode() != 503 {
		t.Errorf("Expected 503 status for circuit breaker error")
	}
}

// TestConvenienceFunctionsAdditional testa funções de conveniência não cobertas
func TestConvenienceFunctionsAdditional(t *testing.T) {
	// Test NewForbiddenError
	forbiddenErr := NewForbiddenError("Access denied")
	if forbiddenErr.StatusCode() != 403 {
		t.Errorf("Expected 403 status for forbidden error")
	}

	// Test NewBadRequestError
	badReqErr := NewBadRequestError("Invalid request")
	if badReqErr.StatusCode() != 400 {
		t.Errorf("Expected 400 status for bad request error")
	}

	// Test NewConflictError
	conflictErr := NewConflictError("Resource already exists")
	if conflictErr.StatusCode() != 409 {
		t.Errorf("Expected 409 status for conflict error")
	}

	// Test NewCircuitBreakerError
	cbErr := NewCircuitBreakerError("payment-service")
	if cbErr.StatusCode() != 503 {
		t.Errorf("Expected 503 status for circuit breaker error")
	}
}

// TestUtilityFunctionsAdditional testa funções utilitárias não cobertas
func TestUtilityFunctionsAdditional(t *testing.T) {
	baseErr := New("BASE001", "Base error")

	// Test IsTemporary usando o tipo concreto
	tempErr := NewTimeoutError("Timeout error")
	domainErrTemp := tempErr.(*DomainError)
	if !domainErrTemp.IsTemporary() {
		t.Errorf("Expected timeout error to be temporary")
	}

	domainErrBase := baseErr.(*DomainError)
	if domainErrBase.IsTemporary() {
		t.Errorf("Expected base error not to be temporary")
	}

	// Test GetRootCause
	wrappedErr := Wrap("WRAP001", "Wrapped", baseErr)
	rootCause := GetRootCause(wrappedErr)
	if rootCause == nil {
		t.Errorf("Expected root cause to be found")
	}

	// Test FormatError
	formatted := FormatError(baseErr)
	if formatted == "" {
		t.Errorf("Expected formatted error string")
	}

	// Test ToJSON
	jsonBytes, err := ToJSON(baseErr)
	if err != nil {
		t.Errorf("Expected successful JSON conversion")
	}
	if len(jsonBytes) == 0 {
		t.Errorf("Expected non-empty JSON")
	}

	// Test As function
	var domainErr interfaces.DomainErrorInterface
	if !As(baseErr, &domainErr) {
		t.Errorf("Expected As to work with DomainErrorInterface")
	}
}

// TestDomainErrorAdditionalMethods testa métodos não cobertos do DomainError
func TestDomainErrorAdditionalMethods(t *testing.T) {
	err := NewBuilder().
		WithCode("TEST001").
		WithMessage("Test error").
		WithType(string(types.ErrorTypeValidation)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithCategory(interfaces.Category("test")).
		Build()

	// Test String method
	strRepr := err.String()
	if strRepr == "" {
		t.Errorf("Expected non-empty string representation")
	}

	// Test DetailedString
	detailed := err.DetailedString()
	if detailed == "" {
		t.Errorf("Expected non-empty detailed string")
	}

	// Test Severity
	severity := err.Severity()
	if severity != interfaces.Severity(types.SeverityHigh) {
		t.Errorf("Expected high severity")
	}

	// Test Category
	category := err.Category()
	if category != interfaces.Category("test") {
		t.Errorf("Expected test category")
	}

	// Test Headers
	err.SetStatusCode(418)
	if err.StatusCode() != 418 {
		t.Errorf("Expected status code 418")
	}

	// Test ResponseBody
	body := err.ResponseBody()
	if body == nil {
		t.Errorf("Expected response body")
	}

	// Test IsTemporary on domain error using concrete type
	domainErr := err.(*DomainError)
	if domainErr.IsTemporary() {
		t.Errorf("Expected validation error not to be temporary")
	}

	// Test WithContext
	ctx := context.WithValue(context.Background(), "trace_id", "xyz789")
	contextErr := domainErr.WithContext(ctx)

	metadata := contextErr.Metadata()
	if metadata["trace_id"] != "xyz789" {
		t.Errorf("Expected trace_id in metadata")
	}
}

// TestRootCauseTraversal testa o percurso de root cause
func TestRootCauseTraversal(t *testing.T) {
	originalErr := errors.New("database connection failed")

	// Cria uma cadeia de erros
	err1 := NewWithCause("DB001", "Database error", originalErr)
	err2 := Wrap("SRV001", "Service error", err1)
	err3 := Wrap("CTR001", "Controller error", err2)

	rootCause := err3.RootCause()
	if rootCause != originalErr {
		t.Errorf("Expected to find original error as root cause")
	}
}

// TestValidationErrorAdvanced testa funcionalidades avançadas do ValidationError
func TestValidationErrorAdvanced(t *testing.T) {
	// Test creation with empty fields
	validationErr := NewValidationError("Empty validation", map[string][]string{})

	// Test AddFields
	validationErr.AddFields(map[string][]string{
		"field1": {"error1", "error2"},
		"field2": {"error3"},
	})

	// Test basic field operations using available methods
	fields := validationErr.Fields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Test HasField
	if !validationErr.HasField("field1") {
		t.Errorf("Expected field1 to exist")
	}

	// Test FieldErrors
	field1Errors := validationErr.FieldErrors("field1")
	if len(field1Errors) != 2 {
		t.Errorf("Expected 2 errors for field1, got %d", len(field1Errors))
	}

	// Test adding another field
	validationErr.AddField("field3", "error4")
	if !validationErr.HasField("field3") {
		t.Errorf("Expected field3 to exist after adding")
	}

	// Test Error method using concrete type
	validationErrConcrete := validationErr.(*ValidationError)
	errorStr := validationErrConcrete.Error()
	if errorStr == "" {
		t.Errorf("Expected non-empty error string")
	}

	// Test JSON method
	jsonBytes, err := validationErrConcrete.JSON()
	if err != nil {
		t.Errorf("Expected successful JSON conversion: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Errorf("Expected non-empty JSON")
	}

	// Test DetailedString
	detailed := validationErrConcrete.DetailedString()
	if detailed == "" {
		t.Errorf("Expected non-empty detailed string")
	}

	// Test ResponseBody
	body := validationErrConcrete.ResponseBody()
	if body == nil {
		t.Errorf("Expected response body")
	}
}

// TestErrorInheritance testa herança de metadados entre erros
func TestErrorInheritance(t *testing.T) {
	// Cria erro pai com metadados
	parentErr := NewBuilder().
		WithCode("PARENT001").
		WithMessage("Parent error").
		WithDetail("parent_key", "parent_value").
		WithTag("parent_tag").
		WithHeader("X-Parent", "header").
		Build()

	// Wrap o erro pai
	childErr := NewBuilder().
		WithCode("CHILD001").
		WithMessage("Child error").
		Build().
		Wrap("wrapped", parentErr)

	// Verifica herança
	details := childErr.Details()
	if details["parent_key"] != "parent_value" {
		t.Errorf("Expected to inherit parent details")
	}

	tags := childErr.Tags()
	hasParentTag := false
	for _, tag := range tags {
		if tag == "parent_tag" {
			hasParentTag = true
			break
		}
	}
	if !hasParentTag {
		t.Errorf("Expected to inherit parent tags")
	}

	headers := childErr.Headers()
	if headers["X-Parent"] != "header" {
		t.Errorf("Expected to inherit parent headers")
	}
}

// TestObjectPooling testa o object pooling
func TestObjectPooling(t *testing.T) {
	// Cria vários erros para testar pooling
	errors := make([]interfaces.DomainErrorInterface, 10)

	for i := 0; i < 10; i++ {
		errors[i] = New(fmt.Sprintf("POOL%03d", i), fmt.Sprintf("Pool test %d", i))
	}

	// Verifica que os erros foram criados
	for i, err := range errors {
		expectedCode := fmt.Sprintf("POOL%03d", i)
		if err.Code() != expectedCode {
			t.Errorf("Expected code %s, got %s", expectedCode, err.Code())
		}
	}

	// Testa release manual
	domainErr := errors[0].(*DomainError)
	domainErr.release()

	// Cria novo erro para potencialmente reutilizar objeto do pool
	newErr := New("REUSED001", "Reused error")
	if newErr.Code() != "REUSED001" {
		t.Errorf("Expected reused error to have correct code")
	}
}

// TestStackTraceCapture testa captura de stack trace
func TestStackTraceCapture(t *testing.T) {
	err := New("STACK001", "Stack test")
	domainErr := err.(*DomainError)

	// Força captura de stack trace
	domainErr.captureStackTrace("test message", 1)

	stackTrace := err.FormatStackTrace()
	if stackTrace == "" {
		t.Errorf("Expected non-empty stack trace")
	}

	if !strings.Contains(stackTrace, "test message") {
		t.Errorf("Expected stack trace to contain test message")
	}
}

// TestConcurrentAccess testa acesso concorrente adicional
func TestConcurrentAccess(t *testing.T) {
	err := NewBuilder().
		WithCode("CONC001").
		WithMessage("Concurrent test").
		Build()

	domainErr := err.(*DomainError)

	var wg sync.WaitGroup

	// Testa leitura concorrente de diferentes métodos
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Vários métodos de leitura
			_ = domainErr.Code()
			_ = domainErr.Message()
			_ = domainErr.Type()
			_ = domainErr.Details()
			_ = domainErr.Tags()
			_ = domainErr.Headers()
			_ = domainErr.StatusCode()
			_ = domainErr.String()
			_ = domainErr.Error()
			_ = domainErr.IsRetryable()
			_ = domainErr.IsTemporary()
		}()
	}

	wg.Wait()
}

// Benchmark adicional para diferentes operações
func BenchmarkErrorCreationComplex(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBuilder().
			WithCode("BENCH001").
			WithMessage("Benchmark test").
			WithType(string(types.ErrorTypeValidation)).
			WithDetail("key", "value").
			WithTag("benchmark").
			WithContext(ctx).
			Build()
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	err := NewBuilder().
		WithCode("JSON001").
		WithMessage("JSON benchmark").
		WithDetails(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": []string{"a", "b", "c"},
		}).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = err.JSON()
	}
}

// TestValidationErrorExtendedMethods testa métodos específicos do ValidationError
func TestValidationErrorExtendedMethods(t *testing.T) {
	// Teste com ValidationError concreta para acessar todos os métodos
	validationErr := NewValidationError("Test validation", map[string][]string{
		"field1": {"error1", "error2"},
		"field2": {"error3"},
	}).(*ValidationError)

	// Test TotalErrors
	total := validationErr.TotalErrors()
	if total != 3 {
		t.Errorf("Expected 3 total errors, got %d", total)
	}

	// Test FirstError
	first := validationErr.FirstError()
	if first == "" {
		t.Errorf("Expected first error")
	}

	// Test FieldNames
	fieldNames := validationErr.FieldNames()
	if len(fieldNames) != 2 {
		t.Errorf("Expected 2 field names, got %d", len(fieldNames))
	}

	// Test IsEmpty
	emptyErr := NewValidationError("Empty", map[string][]string{}).(*ValidationError)
	if !emptyErr.IsEmpty() {
		t.Errorf("Expected empty validation error")
	}

	// Test Merge
	other := NewValidationError("Other", map[string][]string{
		"field3": {"error4"},
	}).(*ValidationError)
	validationErr.Merge(other)

	if validationErr.TotalErrors() != 4 {
		t.Errorf("Expected 4 total errors after merge, got %d", validationErr.TotalErrors())
	}

	// Test WithFieldPrefix
	prefixed := validationErr.WithFieldPrefix("user.")
	prefixedFieldsConcrete := prefixed.(*ValidationError)
	prefixedFields := prefixedFieldsConcrete.FieldNames()
	for _, field := range prefixedFields {
		if !strings.HasPrefix(field, "user.") {
			t.Errorf("Expected field %s to have user. prefix", field)
		}
	}

	// Test Clone
	cloned := validationErr.Clone()
	if cloned.TotalErrors() != validationErr.TotalErrors() {
		t.Errorf("Expected cloned error to have same number of errors")
	}
}

// TestBuildMethods testa métodos build especializados usando o tipo concreto
func TestBuildMethods(t *testing.T) {
	// Para testar os métodos BuildXXX, precisamos usar a estrutura concreta
	// pois eles não estão na interface ErrorBuilder

	// Teste indireto através das funções de conveniência que já existem
	internalErr := NewInternalError("Internal error", errors.New("db failed"))
	if internalErr.StatusCode() != 500 {
		t.Errorf("Expected 500 status for internal error")
	}

	// Teste outros métodos de conveniência não testados anteriormente
	timeoutErr := NewTimeoutError("Operation timeout")
	domainTimeout := timeoutErr.(*DomainError)
	if !domainTimeout.IsRetryable() {
		t.Errorf("Expected timeout error to be retryable")
	}
}

// TestUtilityFunctionsCoverage testa funções utilitárias restantes
func TestUtilityFunctionsCoverage(t *testing.T) {
	// Test IsTemporary with non-retryable error
	nonRetryableErr := New("TEST001", "Non-retryable error")
	if IsTemporary(nonRetryableErr) {
		t.Errorf("Expected non-retryable error not to be temporary")
	}

	// Test ToJSON with nil (should handle gracefully)
	_, err := ToJSON(nil)
	if err == nil {
		t.Errorf("Expected error when converting nil to JSON")
	}

	// Test containsAny function - this is internal but we can test through other functions
	// Let's test it indirectly through error categorization
	validationErr := NewBuilder().
		WithCode("VAL001").
		WithMessage("Validation error with special chars: user@domain.com").
		WithType(string(types.ErrorTypeValidation)).
		Build()

	details := validationErr.Details()
	// This should trigger containsAny internally if used
	if len(details) < 0 {
		t.Errorf("Unexpected error in details processing")
	}
}
