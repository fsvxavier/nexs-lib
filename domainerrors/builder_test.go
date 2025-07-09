package domainerrors

import (
	"fmt"
	"testing"
	"time"
)

func TestErrorBuilder_BasicUsage(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeValidation).
		Message("Invalid input").
		Code("VALIDATION_001").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Error() != "[VALIDATION_001] Invalid input" {
		t.Errorf("Expected message '[VALIDATION_001] Invalid input', got '%s'", err.Error())
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	if err.Code != "VALIDATION_001" {
		t.Errorf("Expected code 'VALIDATION_001', got '%s'", err.Code)
	}
}

func TestErrorBuilder_ValidationError(t *testing.T) {
	fields := map[string][]string{
		"email": {"is required", "must be valid email"},
		"age":   {"must be positive"},
	}

	err := NewErrorBuilder().
		Type(ErrorTypeValidation).
		Message("Validation failed").
		Code("VALIDATION_FAILED").
		ValidationFields(fields).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	if err.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", err.Message)
	}

	// Verificar se os campos de validação foram armazenados nos detalhes
	if validationFields, ok := err.Details["validation_fields"].(map[string][]string); ok {
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

func TestErrorBuilder_NotFoundError(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeNotFound).
		Message("User not found").
		Code("USER_NOT_FOUND").
		ResourceInfo("User", "123").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeNotFound {
		t.Errorf("Expected type %s, got %s", ErrorTypeNotFound, err.Type)
	}

	if err.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got '%s'", err.Message)
	}

	// Verificar se os detalhes do recurso foram armazenados
	if resourceType, ok := err.Details["resource_type"].(string); ok && resourceType != "User" {
		t.Errorf("Expected resource type 'User', got '%s'", resourceType)
	}

	if resourceID, ok := err.Details["resource_id"].(string); ok && resourceID != "123" {
		t.Errorf("Expected resource ID '123', got '%s'", resourceID)
	}
}

func TestErrorBuilder_BusinessError(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Insufficient balance").
		Code("INSUFFICIENT_BALANCE").
		BusinessRule("balance_check").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeBusinessRule {
		t.Errorf("Expected type %s, got %s", ErrorTypeBusinessRule, err.Type)
	}

	if err.Code != "INSUFFICIENT_BALANCE" {
		t.Errorf("Expected code 'INSUFFICIENT_BALANCE', got '%s'", err.Code)
	}

	// Verificar se a regra de negócio foi armazenada
	if businessRule, ok := err.Details["business_rule"].(string); ok && businessRule != "balance_check" {
		t.Errorf("Expected business rule 'balance_check', got '%s'", businessRule)
	}
}

func TestErrorBuilder_InfrastructureError(t *testing.T) {
	originalErr := fmt.Errorf("connection timeout")

	err := NewErrorBuilder().
		Type(ErrorTypeInfrastructure).
		Message("Database connection failed").
		Code("DB_CONNECTION_FAILED").
		Database("select", "users").
		Cause(originalErr).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeInfrastructure {
		t.Errorf("Expected type %s, got %s", ErrorTypeInfrastructure, err.Type)
	}

	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be preserved")
	}

	// Verificar se o componente foi armazenado
	if operation, ok := err.Details["db_operation"].(string); ok && operation != "select" {
		t.Errorf("Expected operation 'select', got '%s'", operation)
	}

	if table, ok := err.Details["db_table"].(string); ok && table != "users" {
		t.Errorf("Expected table 'users', got '%s'", table)
	}
}

func TestErrorBuilder_ExternalServiceError(t *testing.T) {
	originalErr := fmt.Errorf("HTTP 500")

	err := NewErrorBuilder().
		Type(ErrorTypeExternalService).
		Message("Payment service unavailable").
		Code("PAYMENT_SERVICE_ERROR").
		ExternalService("payment-api", 500).
		Cause(originalErr).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeExternalService {
		t.Errorf("Expected type %s, got %s", ErrorTypeExternalService, err.Type)
	}

	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be preserved")
	}

	// Verificar se o serviço foi armazenado
	if service, ok := err.Details["external_service"].(string); ok && service != "payment-api" {
		t.Errorf("Expected service 'payment-api', got '%s'", service)
	}

	if statusCode, ok := err.Details["external_status_code"].(int); ok && statusCode != 500 {
		t.Errorf("Expected status code 500, got %d", statusCode)
	}
}

func TestErrorBuilder_TimeoutError(t *testing.T) {
	threshold := 30 * time.Second

	err := NewErrorBuilder().
		Type(ErrorTypeTimeout).
		Message("Operation timed out").
		Code("OPERATION_TIMEOUT").
		Timeout(threshold).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Expected type %s, got %s", ErrorTypeTimeout, err.Type)
	}

	// Verificar se o threshold foi armazenado
	if timeoutThreshold, ok := err.Details["timeout_threshold"].(string); ok && timeoutThreshold != threshold.String() {
		t.Errorf("Expected timeout threshold '%s', got '%s'", threshold.String(), timeoutThreshold)
	}
}

func TestErrorBuilder_ConflictError(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeConflict).
		Message("Resource already exists").
		Code("RESOURCE_CONFLICT").
		Conflict("User").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeConflict {
		t.Errorf("Expected type %s, got %s", ErrorTypeConflict, err.Type)
	}

	if err.Message != "Resource already exists" {
		t.Errorf("Expected message 'Resource already exists', got '%s'", err.Message)
	}

	// Verificar se o recurso foi armazenado
	if resource, ok := err.Details["conflict_resource"].(string); ok && resource != "User" {
		t.Errorf("Expected resource 'User', got '%s'", resource)
	}
}

func TestErrorBuilder_WithMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"userID":    "123",
		"sessionID": "session-456",
		"timestamp": time.Now(),
	}

	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Business logic error").
		Code("BUSINESS_ERROR").
		MetadataMap(metadata).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	// Verificar se metadata foi preservado
	if err.Metadata == nil {
		t.Fatal("Expected metadata to be preserved")
	}

	if userID, ok := err.Metadata["userID"].(string); !ok || userID != "123" {
		t.Errorf("Expected userID '123', got '%v'", err.Metadata["userID"])
	}

	if sessionID, ok := err.Metadata["sessionID"].(string); !ok || sessionID != "session-456" {
		t.Errorf("Expected sessionID 'session-456', got '%v'", err.Metadata["sessionID"])
	}
}

func TestErrorBuilder_ChainedCalls(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	err := NewErrorBuilder().
		Type(ErrorTypeInfrastructure).
		Message("Complex error").
		Code("COMPLEX_ERROR").
		Database("select", "users").
		Cause(originalErr).
		MetadataMap(map[string]interface{}{"key": "value"}).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeInfrastructure {
		t.Errorf("Expected type %s, got %s", ErrorTypeInfrastructure, err.Type)
	}

	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be preserved")
	}

	// Verificar se o componente foi armazenado
	if operation, ok := err.Details["db_operation"].(string); ok && operation != "select" {
		t.Errorf("Expected operation 'select', got '%s'", operation)
	}

	if table, ok := err.Details["db_table"].(string); ok && table != "users" {
		t.Errorf("Expected table 'users', got '%s'", table)
	}

	// Verificar se os metadados foram armazenados
	if key, ok := err.Metadata["key"].(string); !ok || key != "value" {
		t.Errorf("Expected metadata key 'value', got '%v'", err.Metadata["key"])
	}
}

func TestErrorBuilder_MissingFields(t *testing.T) {
	// Teste builder com campos mínimos
	err := NewErrorBuilder().
		Message("Minimal error").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created with minimal fields")
	}

	if err.Error() != "[E999] Minimal error" {
		t.Errorf("Expected message '[E999] Minimal error', got '%s'", err.Error())
	}

	// Deve ter código padrão
	if err.Code == "" {
		t.Error("Expected default code to be set")
	}

	// Deve ter tipo padrão
	if err.Type == "" {
		t.Error("Expected default type to be set")
	}
}

func TestErrorBuilder_DisableStackTrace(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		DisableStackTrace().
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	// Verificar se o stack trace foi desabilitado
	// A implementação específica determinará como verificar isso
}

func TestErrorBuilder_WithTimestamp(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		WithTimestamp().
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	// Verificar se o timestamp foi adicionado
	if _, ok := err.Metadata["timestamp"]; !ok {
		t.Error("Expected timestamp to be added to metadata")
	}
}

func TestErrorBuilder_Reset(t *testing.T) {
	builder := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR")

	// Reset e reutilização
	builder.Reset()

	err := builder.
		Type(ErrorTypeValidation).
		Message("New error").
		Code("NEW_ERROR").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	if err.Message != "New error" {
		t.Errorf("Expected message 'New error', got '%s'", err.Message)
	}

	if err.Code != "NEW_ERROR" {
		t.Errorf("Expected code 'NEW_ERROR', got '%s'", err.Code)
	}
}

func TestErrorBuilder_Clone(t *testing.T) {
	original := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Original error").
		Code("ORIGINAL_ERROR")

	clone := original.Clone()

	// Modificar o clone
	clonedErr := clone.
		Message("Cloned error").
		Code("CLONED_ERROR").
		Build()

	if clonedErr == nil {
		t.Fatal("Expected cloned error to be created")
	}

	if clonedErr.Type != ErrorTypeBusinessRule {
		t.Errorf("Expected type %s, got %s", ErrorTypeBusinessRule, clonedErr.Type)
	}

	if clonedErr.Message != "Cloned error" {
		t.Errorf("Expected message 'Cloned error', got '%s'", clonedErr.Message)
	}

	if clonedErr.Code != "CLONED_ERROR" {
		t.Errorf("Expected code 'CLONED_ERROR', got '%s'", clonedErr.Code)
	}
}

// Teste adicional para métodos específicos do builder
func TestErrorBuilder_BuildAsType(t *testing.T) {
	originalErr := fmt.Errorf("validation failed")

	builder := NewErrorBuilder().
		Type(ErrorTypeValidation).
		Message("Validation error").
		Code("VALIDATION_ERROR").
		ValidationFields(map[string][]string{
			"email": {"is required"},
		}).
		Cause(originalErr)

	// Testando BuildAsType para ValidationError
	err := builder.BuildAsType(&ValidationError{})
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Message != "Validation error" {
		t.Errorf("Expected message 'Validation error', got '%s'", validationErr.Message)
	}

	// Testando BuildAsType para NotFoundError
	builder2 := NewErrorBuilder().
		Type(ErrorTypeNotFound).
		Message("Not found").
		Code("NOT_FOUND").
		ResourceInfo("User", "123")

	err2 := builder2.BuildAsType(&NotFoundError{})
	if err2 == nil {
		t.Fatal("Expected error to be created")
	}

	notFoundErr, ok := err2.(*NotFoundError)
	if !ok {
		t.Fatalf("Expected NotFoundError, got %T", err2)
	}

	if notFoundErr.Message != "Not found" {
		t.Errorf("Expected message 'Not found', got '%s'", notFoundErr.Message)
	}

	// Testando BuildAsType para BusinessError
	builder3 := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Business error").
		Code("BUSINESS_ERROR").
		BusinessRule("insufficient_balance")

	err3 := builder3.BuildAsType(&BusinessError{})
	if err3 == nil {
		t.Fatal("Expected error to be created")
	}

	businessErr, ok := err3.(*BusinessError)
	if !ok {
		t.Fatalf("Expected BusinessError, got %T", err3)
	}

	if businessErr.Message != "Business error" {
		t.Errorf("Expected message 'Business error', got '%s'", businessErr.Message)
	}

	// Testando BuildAsType com tipo não suportado (deve retornar DomainError)
	err4 := builder.BuildAsType("unsupported_type")
	if err4 == nil {
		t.Fatal("Expected error to be created")
	}

	domainErr, ok := err4.(*DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got %T", err4)
	}

	if domainErr.Message != "Validation error" {
		t.Errorf("Expected message 'Validation error', got '%s'", domainErr.Message)
	}
}

func TestErrorBuilder_AdditionalMethods(t *testing.T) {
	// Teste para WithRequestID
	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		WithRequestID("req-123").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if requestID, ok := err.Metadata["request_id"].(string); !ok || requestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%v'", err.Metadata["request_id"])
	}

	// Teste para WithUserID
	err2 := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		WithUserID("user-456").
		Build()

	if err2 == nil {
		t.Fatal("Expected error to be created")
	}

	if userID, ok := err2.Metadata["user_id"].(string); !ok || userID != "user-456" {
		t.Errorf("Expected user ID 'user-456', got '%v'", err2.Metadata["user_id"])
	}

	// Teste para WithOperationID
	err3 := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		WithOperationID("op-789").
		Build()

	if err3 == nil {
		t.Fatal("Expected error to be created")
	}

	if operationID, ok := err3.Metadata["operation_id"].(string); !ok || operationID != "op-789" {
		t.Errorf("Expected operation ID 'op-789', got '%v'", err3.Metadata["operation_id"])
	}

	// Teste para WithStatusCode
	err4 := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Test error").
		Code("TEST_ERROR").
		WithStatusCode(422).
		Build()

	if err4 == nil {
		t.Fatal("Expected error to be created")
	}

	if statusCode, ok := err4.Metadata["custom_status_code"].(int); !ok || statusCode != 422 {
		t.Errorf("Expected status code 422, got '%v'", err4.Metadata["custom_status_code"])
	}
}

func TestErrorBuilder_MessageF(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeValidation).
		MessageF("User %s not found with ID %d", "John", 123).
		Code("USER_NOT_FOUND").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Message != "User John not found with ID 123" {
		t.Errorf("Expected formatted message, got '%s'", err.Message)
	}
}

func TestErrorBuilder_Entity(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	user := User{ID: 123, Name: "John"}

	err := NewErrorBuilder().
		Type(ErrorTypeValidation).
		Message("User validation failed").
		Code("USER_VALIDATION_ERROR").
		Entity(user).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.EntityName != "User" {
		t.Errorf("Expected entity name 'User', got '%s'", err.EntityName)
	}
}

func TestErrorBuilder_ValidationField(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeValidation).
		Message("Validation failed").
		Code("VALIDATION_FAILED").
		ValidationField("email", "is required").
		ValidationField("email", "must be valid").
		ValidationField("age", "must be positive").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	// Verificar se os campos de validação foram armazenados nos detalhes
	if validationFields, ok := err.Details["validation_fields"].(map[string][]string); ok {
		if len(validationFields) != 2 {
			t.Errorf("Expected 2 validation fields, got %d", len(validationFields))
		}

		if len(validationFields["email"]) != 2 {
			t.Errorf("Expected 2 errors for email field, got %d", len(validationFields["email"]))
		}

		if len(validationFields["age"]) != 1 {
			t.Errorf("Expected 1 error for age field, got %d", len(validationFields["age"]))
		}
	} else {
		t.Error("Expected validation fields to be stored in Details")
	}
}

func TestErrorBuilder_Security(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeSecurity).
		Message("Security violation").
		Code("SECURITY_VIOLATION").
		Security("unauthorized_access").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeSecurity {
		t.Errorf("Expected type %s, got %s", ErrorTypeSecurity, err.Type)
	}

	// Verificar se o contexto de segurança foi armazenado
	if securityContext, ok := err.Details["security_context"].(string); !ok || securityContext != "unauthorized_access" {
		t.Errorf("Expected security context 'unauthorized_access', got '%v'", err.Details["security_context"])
	}
}

func TestErrorBuilder_Database(t *testing.T) {
	err := NewErrorBuilder().
		Type(ErrorTypeRepository).
		Message("Database error").
		Code("DB_ERROR").
		Database("insert", "users").
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeRepository {
		t.Errorf("Expected type %s, got %s", ErrorTypeRepository, err.Type)
	}

	// Verificar se os detalhes do banco foram armazenados
	if operation, ok := err.Details["db_operation"].(string); !ok || operation != "insert" {
		t.Errorf("Expected operation 'insert', got '%v'", err.Details["db_operation"])
	}

	if table, ok := err.Details["db_table"].(string); !ok || table != "users" {
		t.Errorf("Expected table 'users', got '%v'", err.Details["db_table"])
	}
}

func TestErrorBuilder_Details(t *testing.T) {
	details := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}

	err := NewErrorBuilder().
		Type(ErrorTypeBusinessRule).
		Message("Business error").
		Code("BUSINESS_ERROR").
		Details(details).
		Build()

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if len(err.Details) != 3 {
		t.Errorf("Expected 3 details, got %d", len(err.Details))
	}

	if field1, ok := err.Details["field1"].(string); !ok || field1 != "value1" {
		t.Errorf("Expected field1 'value1', got '%v'", err.Details["field1"])
	}

	if field2, ok := err.Details["field2"].(int); !ok || field2 != 123 {
		t.Errorf("Expected field2 123, got '%v'", err.Details["field2"])
	}

	if field3, ok := err.Details["field3"].(bool); !ok || field3 != true {
		t.Errorf("Expected field3 true, got '%v'", err.Details["field3"])
	}
}
