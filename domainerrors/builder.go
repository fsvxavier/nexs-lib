package domainerrors

import (
	"fmt"
	"time"
)

// ErrorBuilder implementa o padrão Builder para construção fluente de erros
type ErrorBuilder struct {
	error      *DomainError
	metadata   map[string]interface{}
	details    map[string]interface{}
	stackTrace bool
}

// NewErrorBuilder cria um novo builder de erro
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		error:      &DomainError{},
		metadata:   make(map[string]interface{}),
		details:    make(map[string]interface{}),
		stackTrace: true,
	}
}

// Code define o código do erro
func (eb *ErrorBuilder) Code(code string) *ErrorBuilder {
	eb.error.Code = code
	return eb
}

// Message define a mensagem do erro
func (eb *ErrorBuilder) Message(message string) *ErrorBuilder {
	eb.error.Message = message
	return eb
}

// MessageF define a mensagem do erro com formatação
func (eb *ErrorBuilder) MessageF(format string, args ...interface{}) *ErrorBuilder {
	eb.error.Message = fmt.Sprintf(format, args...)
	return eb
}

// Type define o tipo do erro
func (eb *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
	eb.error.Type = errorType
	return eb
}

// Cause define o erro original
func (eb *ErrorBuilder) Cause(err error) *ErrorBuilder {
	eb.error.Err = err
	if eb.stackTrace && err != nil {
		eb.error.captureStackTrace(eb.error.Message, err)
	}
	return eb
}

// Entity define a entidade relacionada
func (eb *ErrorBuilder) Entity(entity interface{}) *ErrorBuilder {
	eb.error.WithEntity(entity)
	return eb
}

// Detail adiciona um detalhe específico
func (eb *ErrorBuilder) Detail(key string, value interface{}) *ErrorBuilder {
	eb.details[key] = value
	return eb
}

// Details adiciona múltiplos detalhes
func (eb *ErrorBuilder) Details(details map[string]interface{}) *ErrorBuilder {
	for k, v := range details {
		eb.details[k] = v
	}
	return eb
}

// Metadata adiciona um metadado específico
func (eb *ErrorBuilder) Metadata(key string, value interface{}) *ErrorBuilder {
	eb.metadata[key] = value
	return eb
}

// MetadataMap adiciona múltiplos metadados
func (eb *ErrorBuilder) MetadataMap(metadata map[string]interface{}) *ErrorBuilder {
	for k, v := range metadata {
		eb.metadata[k] = v
	}
	return eb
}

// DisableStackTrace desabilita a captura automática de stack trace
func (eb *ErrorBuilder) DisableStackTrace() *ErrorBuilder {
	eb.stackTrace = false
	return eb
}

// WithTimestamp adiciona timestamp ao erro
func (eb *ErrorBuilder) WithTimestamp() *ErrorBuilder {
	eb.metadata["timestamp"] = time.Now().UTC()
	return eb
}

// WithRequestID adiciona ID da requisição
func (eb *ErrorBuilder) WithRequestID(requestID string) *ErrorBuilder {
	eb.metadata["request_id"] = requestID
	return eb
}

// WithUserID adiciona ID do usuário
func (eb *ErrorBuilder) WithUserID(userID string) *ErrorBuilder {
	eb.metadata["user_id"] = userID
	return eb
}

// WithOperationID adiciona ID da operação
func (eb *ErrorBuilder) WithOperationID(operationID string) *ErrorBuilder {
	eb.metadata["operation_id"] = operationID
	return eb
}

// WithStatusCode define um status code customizado
func (eb *ErrorBuilder) WithStatusCode(statusCode int) *ErrorBuilder {
	eb.metadata["custom_status_code"] = statusCode
	return eb
}

// ValidationField adiciona um erro de validação para um campo específico
func (eb *ErrorBuilder) ValidationField(field, message string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeValidation
	}

	if _, exists := eb.details["validation_fields"]; !exists {
		eb.details["validation_fields"] = make(map[string][]string)
	}

	fields := eb.details["validation_fields"].(map[string][]string)
	fields[field] = append(fields[field], message)
	eb.details["validation_fields"] = fields

	return eb
}

// ValidationFields adiciona múltiplos erros de validação
func (eb *ErrorBuilder) ValidationFields(fields map[string][]string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeValidation
	}

	for field, messages := range fields {
		for _, message := range messages {
			eb.ValidationField(field, message)
		}
	}

	return eb
}

// ResourceInfo adiciona informações sobre o recurso (útil para NotFound)
func (eb *ErrorBuilder) ResourceInfo(resourceType, resourceID string) *ErrorBuilder {
	eb.details["resource_type"] = resourceType
	eb.details["resource_id"] = resourceID
	return eb
}

// BusinessRule adiciona informações sobre regra de negócio violada
func (eb *ErrorBuilder) BusinessRule(rule string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeBusinessRule
	}
	eb.details["business_rule"] = rule
	return eb
}

// ExternalService adiciona informações sobre serviço externo
func (eb *ErrorBuilder) ExternalService(service string, statusCode int) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeExternalService
	}
	eb.details["external_service"] = service
	eb.details["external_status_code"] = statusCode
	return eb
}

// Database adiciona informações sobre erro de banco de dados
func (eb *ErrorBuilder) Database(operation, table string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeRepository
	}
	eb.details["db_operation"] = operation
	eb.details["db_table"] = table
	return eb
}

// Security adiciona informações sobre erro de segurança
func (eb *ErrorBuilder) Security(securityContext string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeSecurity
	}
	eb.details["security_context"] = securityContext
	return eb
}

// Timeout adiciona informações sobre timeout
func (eb *ErrorBuilder) Timeout(threshold time.Duration) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeTimeout
	}
	eb.details["timeout_threshold"] = threshold.String()
	return eb
}

// Conflict adiciona informações sobre conflito
func (eb *ErrorBuilder) Conflict(resource string) *ErrorBuilder {
	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeConflict
	}
	eb.details["conflict_resource"] = resource
	return eb
}

// Build constrói e retorna o DomainError final
func (eb *ErrorBuilder) Build() *DomainError {
	// Aplica detalhes acumulados
	if len(eb.details) > 0 {
		if eb.error.Details == nil {
			eb.error.Details = make(map[string]interface{})
		}
		for k, v := range eb.details {
			eb.error.Details[k] = v
		}
	}

	// Aplica metadados acumulados
	if len(eb.metadata) > 0 {
		if eb.error.Metadata == nil {
			eb.error.Metadata = make(map[string]interface{})
		}
		for k, v := range eb.metadata {
			eb.error.Metadata[k] = v
		}
	}

	// Define valores padrão se não foram especificados
	if eb.error.Code == "" {
		eb.error.Code = DefaultErrorCode
	}

	if eb.error.Message == "" {
		eb.error.Message = "An error occurred"
	}

	if eb.error.Type == "" {
		eb.error.Type = ErrorTypeInternal
	}

	// Inicializa slices se necessário
	if eb.error.stack == nil {
		eb.error.stack = make([]stackTrace, 0)
	}

	return eb.error
}

// BuildAsType constrói e retorna o erro como um tipo específico
func (eb *ErrorBuilder) BuildAsType(errorType interface{}) error {
	domainErr := eb.Build()

	switch errorType.(type) {
	case *ValidationError:
		validationErr := &ValidationError{
			DomainError: domainErr,
		}
		if fields, ok := domainErr.Details["validation_fields"].(map[string][]string); ok {
			validationErr.ValidatedFields = fields
		}
		return validationErr

	case *NotFoundError:
		notFoundErr := &NotFoundError{
			DomainError: domainErr,
		}
		if resourceType, ok := domainErr.Details["resource_type"].(string); ok {
			notFoundErr.ResourceType = resourceType
		}
		if resourceID, ok := domainErr.Details["resource_id"].(string); ok {
			notFoundErr.ResourceID = resourceID
		}
		return notFoundErr

	case *BusinessError:
		businessErr := &BusinessError{
			DomainError: domainErr,
		}
		if rule, ok := domainErr.Details["business_rule"].(string); ok {
			businessErr.BusinessCode = rule
		}
		return businessErr

	default:
		return domainErr
	}
}

// Reset limpa o builder para reutilização
func (eb *ErrorBuilder) Reset() *ErrorBuilder {
	eb.error = &DomainError{}
	eb.metadata = make(map[string]interface{})
	eb.details = make(map[string]interface{})
	eb.stackTrace = true
	return eb
}

// Clone cria uma cópia do builder atual
func (eb *ErrorBuilder) Clone() *ErrorBuilder {
	newBuilder := &ErrorBuilder{
		error:      &DomainError{},
		metadata:   make(map[string]interface{}),
		details:    make(map[string]interface{}),
		stackTrace: eb.stackTrace,
	}

	// Copia campos da estrutura principal
	if eb.error != nil {
		newBuilder.error.Code = eb.error.Code
		newBuilder.error.Message = eb.error.Message
		newBuilder.error.Type = eb.error.Type
		newBuilder.error.Err = eb.error.Err
		newBuilder.error.EntityName = eb.error.EntityName
	}

	// Copia metadados
	for k, v := range eb.metadata {
		newBuilder.metadata[k] = v
	}

	// Copia detalhes
	for k, v := range eb.details {
		newBuilder.details[k] = v
	}

	return newBuilder
}
