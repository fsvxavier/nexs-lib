package domainerrors

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDomainErrorFactory_CreateValidationError(t *testing.T) {
	factory := NewDomainErrorFactory()
	fields := map[string][]string{
		"email": {"is required", "must be valid email"},
		"name":  {"is required"},
	}

	err := factory.CreateValidationError("Validation failed", fields)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, domainErr.Type)
	}

	if domainErr.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", domainErr.Message)
	}

	// Verificar se os campos de validação foram armazenados nos detalhes
	if validationFields, ok := domainErr.Details["validation_fields"].(map[string][]string); ok {
		if len(validationFields) != 2 {
			t.Errorf("Expected 2 validation fields, got %d", len(validationFields))
		}

		if len(validationFields["email"]) != 2 {
			t.Errorf("Expected 2 errors for email field, got %d", len(validationFields["email"]))
		}
	} else {
		t.Error("Expected validation fields to be stored in Details")
	}
}

func TestDomainErrorFactory_CreateNotFoundError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateNotFoundError("User not found", "User", "123")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeNotFound {
		t.Errorf("Expected type %s, got %s", ErrorTypeNotFound, domainErr.Type)
	}

	if domainErr.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", domainErr.Message)
	}

	// Verificar se os detalhes do recurso foram armazenados
	if resourceType, ok := domainErr.Details["resource_type"].(string); ok && resourceType != "User" {
		t.Errorf("Expected resource type 'User', got '%s'", resourceType)
	}

	if resourceID, ok := domainErr.Details["resource_id"].(string); ok && resourceID != "123" {
		t.Errorf("Expected resource ID '123', got '%s'", resourceID)
	}
}

func TestDomainErrorFactory_CreateBusinessError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateBusinessError("INSUFFICIENT_BALANCE", "Insufficient balance for transaction")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeBusinessRule {
		t.Errorf("Expected type %s, got %s", ErrorTypeBusinessRule, domainErr.Type)
	}

	if domainErr.Message != "Insufficient balance for transaction" {
		t.Errorf("Expected message 'Insufficient balance for transaction', got '%s'", domainErr.Message)
	}

	if domainErr.Code != "INSUFFICIENT_BALANCE" {
		t.Errorf("Expected code 'INSUFFICIENT_BALANCE', got '%s'", domainErr.Code)
	}

	// Verificar se a regra de negócio foi armazenada
	if businessRule, ok := domainErr.Details["business_rule"].(string); ok && businessRule != "INSUFFICIENT_BALANCE" {
		t.Errorf("Expected business rule 'INSUFFICIENT_BALANCE', got '%s'", businessRule)
	}
}

func TestDomainErrorFactory_CreateInfrastructureError(t *testing.T) {
	factory := NewDomainErrorFactory()
	originalErr := fmt.Errorf("connection timeout")

	err := factory.CreateInfrastructureError("database", "Database connection failed", originalErr)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeInfrastructure {
		t.Errorf("Expected type %s, got %s", ErrorTypeInfrastructure, domainErr.Type)
	}

	if domainErr.Message != "Database connection failed" {
		t.Errorf("Expected message 'Database connection failed', got '%s'", domainErr.Message)
	}

	if domainErr.Err != originalErr {
		t.Errorf("Expected wrapped error to be preserved")
	}

	// Verificar se o componente foi armazenado
	if component, ok := domainErr.Details["infrastructure_component"].(string); ok && component != "database" {
		t.Errorf("Expected component 'database', got '%s'", component)
	}
}

func TestDomainErrorFactory_CreateExternalServiceError(t *testing.T) {
	factory := NewDomainErrorFactory()
	originalErr := fmt.Errorf("HTTP 500")

	err := factory.CreateExternalServiceError("payment-api", "Payment service unavailable", 500, originalErr)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeExternalService {
		t.Errorf("Expected type %s, got %s", ErrorTypeExternalService, domainErr.Type)
	}

	if domainErr.Message != "Payment service unavailable" {
		t.Errorf("Expected message 'Payment service unavailable', got '%s'", domainErr.Message)
	}

	if domainErr.Err != originalErr {
		t.Errorf("Expected wrapped error to be preserved")
	}

	// Verificar se o serviço foi armazenado
	if service, ok := domainErr.Details["external_service"].(string); ok && service != "payment-api" {
		t.Errorf("Expected service 'payment-api', got '%s'", service)
	}

	if statusCode, ok := domainErr.Details["external_status_code"].(int); ok && statusCode != 500 {
		t.Errorf("Expected status code 500, got %d", statusCode)
	}
}

func TestDomainErrorFactory_CreateAuthenticationError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateAuthenticationError("Invalid credentials", "expired_token")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeAuthentication {
		t.Errorf("Expected type %s, got %s", ErrorTypeAuthentication, domainErr.Type)
	}

	if domainErr.Message != "Invalid credentials" {
		t.Errorf("Expected message 'Invalid credentials', got '%s'", domainErr.Message)
	}

	// Verificar se a razão foi armazenada
	if reason, ok := domainErr.Details["authentication_reason"].(string); ok && reason != "expired_token" {
		t.Errorf("Expected reason 'expired_token', got '%s'", reason)
	}
}

func TestDomainErrorFactory_CreateAuthorizationError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateAuthorizationError("Access denied", "admin:write")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeAuthorization {
		t.Errorf("Expected type %s, got %s", ErrorTypeAuthorization, domainErr.Type)
	}

	if domainErr.Message != "Access denied" {
		t.Errorf("Expected message 'Access denied', got '%s'", domainErr.Message)
	}

	// Verificar se a permissão foi armazenada
	if permission, ok := domainErr.Details["authorization_permission"].(string); ok && permission != "admin:write" {
		t.Errorf("Expected permission 'admin:write', got '%s'", permission)
	}
}

func TestDomainErrorFactory_CreateTimeoutError(t *testing.T) {
	factory := NewDomainErrorFactory()
	threshold := 30 * time.Second

	err := factory.CreateTimeoutError("Operation timed out", threshold)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeTimeout {
		t.Errorf("Expected type %s, got %s", ErrorTypeTimeout, domainErr.Type)
	}

	if domainErr.Message != "Operation timed out" {
		t.Errorf("Expected message 'Operation timed out', got '%s'", domainErr.Message)
	}

	// Verificar se o threshold foi armazenado
	if timeoutThreshold, ok := domainErr.Details["timeout_threshold"].(string); ok && timeoutThreshold != threshold.String() {
		t.Errorf("Expected timeout threshold '%s', got '%s'", threshold.String(), timeoutThreshold)
	}
}

func TestDomainErrorFactory_CreateConflictError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateConflictError("Resource already exists", "User")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeConflict {
		t.Errorf("Expected type %s, got %s", ErrorTypeConflict, domainErr.Type)
	}

	if domainErr.Message != "Resource already exists" {
		t.Errorf("Expected message 'Resource already exists', got '%s'", domainErr.Message)
	}

	// Verificar se o recurso foi armazenado
	if resource, ok := domainErr.Details["conflict_resource"].(string); ok && resource != "User" {
		t.Errorf("Expected resource 'User', got '%s'", resource)
	}
}

func TestDomainErrorFactory_CreateRateLimitError(t *testing.T) {
	factory := NewDomainErrorFactory()

	err := factory.CreateRateLimitError("Rate limit exceeded", 100, time.Minute)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeRateLimit {
		t.Errorf("Expected type %s, got %s", ErrorTypeRateLimit, domainErr.Type)
	}

	if domainErr.Message != "Rate limit exceeded" {
		t.Errorf("Expected message 'Rate limit exceeded', got '%s'", domainErr.Message)
	}

	// Verificar se o limite foi armazenado
	if limit, ok := domainErr.Details["rate_limit"].(int); ok && limit != 100 {
		t.Errorf("Expected limit 100, got %d", limit)
	}

	if window, ok := domainErr.Details["rate_window"].(string); ok && window != time.Minute.String() {
		t.Errorf("Expected window '%s', got '%s'", time.Minute.String(), window)
	}
}

// Mock observer para testes
type MockErrorObserver struct {
	LastError   error
	LastContext map[string]interface{}
	CallCount   int
}

func (m *MockErrorObserver) OnErrorCreated(err error, ctx map[string]interface{}) {
	m.LastError = err
	m.LastContext = ctx
	m.CallCount++
}

func TestDomainErrorFactory_WithObserver(t *testing.T) {
	observer := &MockErrorObserver{}
	factory := NewDomainErrorFactory(WithErrorObserver(observer))

	err := factory.CreateBusinessError("TEST_ERROR", "Test error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if observer.CallCount != 1 {
		t.Errorf("Expected observer to be called 1 time, got %d", observer.CallCount)
	}

	if observer.LastError != err {
		t.Error("Expected observer to receive the created error")
	}
}

// Mock context enricher para testes
type MockContextEnricher struct {
	EnrichCalled bool
}

func (m *MockContextEnricher) EnrichError(ctx context.Context, builder *ErrorBuilder) *ErrorBuilder {
	m.EnrichCalled = true
	if ctx != nil {
		if requestID := ctx.Value("requestID"); requestID != nil {
			builder.WithRequestID(requestID.(string))
		}
	}
	return builder
}

func TestDomainErrorFactory_WithContextEnricher(t *testing.T) {
	enricher := &MockContextEnricher{}
	factory := NewDomainErrorFactory(WithContextEnricher(enricher))

	ctx := context.WithValue(context.Background(), "requestID", "req-123")
	err := factory.createError(ctx, NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR"))

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if !enricher.EnrichCalled {
		t.Error("Expected context enricher to be called")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	// Verificar se o request ID foi adicionado
	if requestID, ok := domainErr.Metadata["request_id"].(string); !ok || requestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%v'", domainErr.Metadata["request_id"])
	}
}

func TestDomainErrorFactory_GetDefaultFactory(t *testing.T) {
	factory1 := GetDefaultFactory()
	factory2 := GetDefaultFactory()

	if factory1 != factory2 {
		t.Error("Expected GetDefaultFactory to return the same instance")
	}
}

func TestDomainErrorFactory_SetDefaultFactory(t *testing.T) {
	// Guardar a factory original
	originalFactory := GetDefaultFactory()

	// Criar uma nova factory customizada
	observer := &MockErrorObserver{}
	customFactory := NewDomainErrorFactory(WithErrorObserver(observer))

	// Definir como padrão
	SetDefaultFactory(customFactory)

	// Verificar se a nova factory é retornada
	if GetDefaultFactory() != customFactory {
		t.Error("Expected GetDefaultFactory to return the custom factory")
	}

	// Restaurar a factory original
	SetDefaultFactory(originalFactory)

	// Verificar se foi restaurada
	if GetDefaultFactory() != originalFactory {
		t.Error("Expected GetDefaultFactory to return the original factory")
	}
}

func TestDomainErrorFactory_GetDefaultFactoryWithOptions(t *testing.T) {
	observer := &MockErrorObserver{}
	factory := GetDefaultFactoryWithOptions(WithErrorObserver(observer))

	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	// Testar se o observer foi configurado
	err := factory.CreateBusinessError("TEST_ERROR", "Test error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if observer.CallCount != 1 {
		t.Errorf("Expected observer to be called 1 time, got %d", observer.CallCount)
	}
}

func TestDefaultErrorFactory_CreatesValidErrors(t *testing.T) {
	factory := NewDomainErrorFactory()

	tests := []struct {
		name         string
		createError  func() error
		expectedType ErrorType
	}{
		{
			name: "ValidationError",
			createError: func() error {
				return factory.CreateValidationError("Validation failed", nil)
			},
			expectedType: ErrorTypeValidation,
		},
		{
			name: "NotFoundError",
			createError: func() error {
				return factory.CreateNotFoundError("Not found", "User", "123")
			},
			expectedType: ErrorTypeNotFound,
		},
		{
			name: "InfrastructureError",
			createError: func() error {
				return factory.CreateInfrastructureError("database", "DB error", nil)
			},
			expectedType: ErrorTypeInfrastructure,
		},
		{
			name: "BusinessError",
			createError: func() error {
				return factory.CreateBusinessError("BUSINESS_ERROR", "Business error")
			},
			expectedType: ErrorTypeBusinessRule,
		},
		{
			name: "ExternalServiceError",
			createError: func() error {
				return factory.CreateExternalServiceError("api", "API error", 500, nil)
			},
			expectedType: ErrorTypeExternalService,
		},
		{
			name: "AuthenticationError",
			createError: func() error {
				return factory.CreateAuthenticationError("Auth failed", "invalid_token")
			},
			expectedType: ErrorTypeAuthentication,
		},
		{
			name: "AuthorizationError",
			createError: func() error {
				return factory.CreateAuthorizationError("Access denied", "admin:write")
			},
			expectedType: ErrorTypeAuthorization,
		},
		{
			name: "TimeoutError",
			createError: func() error {
				return factory.CreateTimeoutError("Timeout", 30*time.Second)
			},
			expectedType: ErrorTypeTimeout,
		},
		{
			name: "ConflictError",
			createError: func() error {
				return factory.CreateConflictError("Conflict", "resource")
			},
			expectedType: ErrorTypeConflict,
		},
		{
			name: "RateLimitError",
			createError: func() error {
				return factory.CreateRateLimitError("Rate limit", 100, time.Minute)
			},
			expectedType: ErrorTypeRateLimit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.createError()
			if err == nil {
				t.Fatal("Expected error to be created")
			}

			domainErr, ok := err.(*DomainError)
			if !ok {
				t.Fatalf("Expected DomainError, got %T", err)
			}

			if domainErr.Type != test.expectedType {
				t.Errorf("Expected type %s, got %s", test.expectedType, domainErr.Type)
			}
		})
	}
}

// Teste adicional para logging observer
func TestLoggingErrorObserver(t *testing.T) {
	var loggedLevel string
	var loggedMessage string
	var loggedMetadata map[string]interface{}

	logger := func(level, message string, metadata map[string]interface{}) {
		loggedLevel = level
		loggedMessage = message
		loggedMetadata = metadata
	}

	observer := NewLoggingErrorObserver(logger)
	factory := NewDomainErrorFactory(WithErrorObserver(observer))

	err := factory.CreateBusinessError("TEST_ERROR", "Test error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if loggedLevel != "warn" {
		t.Errorf("Expected log level 'warn', got '%s'", loggedLevel)
	}

	expectedMessage := "Domain error occurred: [TEST_ERROR] Test error message"
	if loggedMessage != expectedMessage {
		t.Errorf("Expected log message '%s', got '%s'", expectedMessage, loggedMessage)
	}

	if loggedMetadata == nil {
		t.Error("Expected metadata to be logged")
	}

	// Verificar se o tipo de erro foi registrado nos metadados
	if errorType, ok := loggedMetadata["error_type"].(string); !ok || errorType != string(ErrorTypeBusinessRule) {
		t.Errorf("Expected error_type '%s', got '%v'", ErrorTypeBusinessRule, loggedMetadata["error_type"])
	}

	if errorCode, ok := loggedMetadata["error_code"].(string); !ok || errorCode != "TEST_ERROR" {
		t.Errorf("Expected error_code 'TEST_ERROR', got '%v'", loggedMetadata["error_code"])
	}
}

// Teste adicional para metrics observer
func TestMetricsErrorObserver(t *testing.T) {
	var counterName string
	var counterTags map[string]string

	incrementCounter := func(name string, tags map[string]string) {
		counterName = name
		counterTags = tags
	}

	observer := NewMetricsErrorObserver(incrementCounter)
	factory := NewDomainErrorFactory(WithErrorObserver(observer))

	err := factory.CreateValidationError("Validation failed", nil)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if counterName != "domain_errors_total" {
		t.Errorf("Expected counter name 'domain_errors_total', got '%s'", counterName)
	}

	if counterTags == nil {
		t.Fatal("Expected counter tags to be set")
	}

	if counterTags["error_type"] != string(ErrorTypeValidation) {
		t.Errorf("Expected type tag '%s', got '%s'", ErrorTypeValidation, counterTags["error_type"])
	}
}

// Teste de factory com builder customizado
func TestDomainErrorFactory_WithCustomBuilder(t *testing.T) {
	customBuilderFactory := func() *ErrorBuilder {
		builder := NewErrorBuilder()
		builder.WithTimestamp() // Sempre adiciona timestamp
		return builder
	}

	factory := NewDomainErrorFactory(WithBuilderFactory(customBuilderFactory))

	err := factory.CreateBusinessError("TEST_ERROR", "Test error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	// Verificar se o timestamp foi adicionado automaticamente
	if _, ok := domainErr.Metadata["timestamp"]; !ok {
		t.Error("Expected timestamp to be added automatically by custom builder")
	}
}

// Teste do método createError interno
func TestDomainErrorFactory_createError(t *testing.T) {
	enricher := &MockContextEnricher{}
	observer := &MockErrorObserver{}
	factory := NewDomainErrorFactory(WithContextEnricher(enricher), WithErrorObserver(observer))

	ctx := context.WithValue(context.Background(), "requestID", "req-456")
	builder := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR")

	err := factory.createError(ctx, builder)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	// Verificar se o enricher foi chamado
	if !enricher.EnrichCalled {
		t.Error("Expected context enricher to be called")
	}

	// Verificar se o observer foi chamado
	if observer.CallCount != 1 {
		t.Errorf("Expected observer to be called 1 time, got %d", observer.CallCount)
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	// Verificar se o request ID foi adicionado pelo enricher
	if requestID, ok := domainErr.Metadata["request_id"].(string); !ok || requestID != "req-456" {
		t.Errorf("Expected request ID 'req-456', got '%v'", domainErr.Metadata["request_id"])
	}
}

// Teste de múltiplos observers
func TestDomainErrorFactory_MultipleObservers(t *testing.T) {
	observer1 := &MockErrorObserver{}
	observer2 := &MockErrorObserver{}

	factory := NewDomainErrorFactory(
		WithErrorObserver(observer1),
		WithErrorObserver(observer2),
	)

	err := factory.CreateBusinessError("TEST_ERROR", "Test error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if observer1.CallCount != 1 {
		t.Errorf("Expected observer1 to be called 1 time, got %d", observer1.CallCount)
	}

	if observer2.CallCount != 1 {
		t.Errorf("Expected observer2 to be called 1 time, got %d", observer2.CallCount)
	}
}

// Teste de múltiplas opções de factory
func TestDomainErrorFactory_MultipleOptions(t *testing.T) {
	enricher := &MockContextEnricher{}
	observer := &MockErrorObserver{}
	customBuilderFactory := func() *ErrorBuilder {
		return NewErrorBuilder().WithTimestamp()
	}

	factory := NewDomainErrorFactory(
		WithContextEnricher(enricher),
		WithErrorObserver(observer),
		WithBuilderFactory(customBuilderFactory),
	)

	ctx := context.WithValue(context.Background(), "requestID", "req-789")
	err := factory.createError(ctx, NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR"))

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	// Verificar se todas as opções foram aplicadas
	if !enricher.EnrichCalled {
		t.Error("Expected context enricher to be called")
	}

	if observer.CallCount != 1 {
		t.Errorf("Expected observer to be called 1 time, got %d", observer.CallCount)
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	// Verificar se o request ID foi adicionado pelo enricher
	if requestID, ok := domainErr.Metadata["request_id"].(string); !ok || requestID != "req-789" {
		t.Errorf("Expected request ID 'req-789', got '%v'", domainErr.Metadata["request_id"])
	}
}

// Teste sem context enricher e sem observer
func TestDomainErrorFactory_MinimalConfiguration(t *testing.T) {
	factory := NewDomainErrorFactory() // Configuração mínima

	err := factory.CreateBusinessError("MINIMAL_ERROR", "Minimal error message")

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err)
	}

	if domainErr.Type != ErrorTypeBusinessRule {
		t.Errorf("Expected type %s, got %s", ErrorTypeBusinessRule, domainErr.Type)
	}

	if domainErr.Message != "Minimal error message" {
		t.Errorf("Expected message 'Minimal error message', got '%s'", domainErr.Message)
	}

	if domainErr.Code != "MINIMAL_ERROR" {
		t.Errorf("Expected code 'MINIMAL_ERROR', got '%s'", domainErr.Code)
	}
}

// Teste de factory tipos específicos
func TestDomainErrorFactory_SpecificErrorTypes(t *testing.T) {
	factory := NewDomainErrorFactory()

	// Teste validação com campos específicos
	fields := map[string][]string{
		"email":    {"is required", "must be valid format"},
		"password": {"is required", "must be at least 8 characters"},
	}

	validationErr := factory.CreateValidationError("Validation failed", fields)
	domainErr := validationErr.(*DomainError)

	if domainErr.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, domainErr.Type)
	}

	if validationFields, ok := domainErr.Details["validation_fields"].(map[string][]string); ok {
		if len(validationFields["email"]) != 2 {
			t.Errorf("Expected 2 email errors, got %d", len(validationFields["email"]))
		}
		if len(validationFields["password"]) != 2 {
			t.Errorf("Expected 2 password errors, got %d", len(validationFields["password"]))
		}
	} else {
		t.Error("Expected validation fields to be stored")
	}

	// Teste not found com detalhes de recurso
	notFoundErr := factory.CreateNotFoundError("User not found", "User", "user-123")
	domainErr2 := notFoundErr.(*DomainError)

	if domainErr2.Type != ErrorTypeNotFound {
		t.Errorf("Expected type %s, got %s", ErrorTypeNotFound, domainErr2.Type)
	}

	if resourceType, ok := domainErr2.Details["resource_type"].(string); !ok || resourceType != "User" {
		t.Errorf("Expected resource type 'User', got '%v'", domainErr2.Details["resource_type"])
	}

	if resourceID, ok := domainErr2.Details["resource_id"].(string); !ok || resourceID != "user-123" {
		t.Errorf("Expected resource ID 'user-123', got '%v'", domainErr2.Details["resource_id"])
	}

	// Teste external service com status code
	originalErr := fmt.Errorf("HTTP 503")
	extErr := factory.CreateExternalServiceError("payment-api", "Service unavailable", 503, originalErr)
	domainErr3 := extErr.(*DomainError)

	if domainErr3.Type != ErrorTypeExternalService {
		t.Errorf("Expected type %s, got %s", ErrorTypeExternalService, domainErr3.Type)
	}

	if service, ok := domainErr3.Details["external_service"].(string); !ok || service != "payment-api" {
		t.Errorf("Expected service 'payment-api', got '%v'", domainErr3.Details["external_service"])
	}

	if statusCode, ok := domainErr3.Details["external_status_code"].(int); !ok || statusCode != 503 {
		t.Errorf("Expected status code 503, got '%v'", domainErr3.Details["external_status_code"])
	}

	if domainErr3.Err != originalErr {
		t.Error("Expected original error to be wrapped")
	}
}
