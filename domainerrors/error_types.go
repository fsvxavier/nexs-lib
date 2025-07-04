package domainerrors

import (
	"errors"
	"net/http"
)

// ValidationError é um tipo específico para erros de validação
type ValidationError struct {
	*DomainError
	ValidatedFields map[string][]string `json:"validated_fields,omitempty"`
}

// NewValidationError cria um erro de validação
func NewValidationError(message string, fields map[string][]string) *ValidationError {
	err := &ValidationError{
		DomainError:     New("VALIDATION_ERROR", message).WithType(ErrorTypeValidation),
		ValidatedFields: fields,
	}

	if err.ValidatedFields == nil {
		err.ValidatedFields = make(map[string][]string)
	}

	return err
}

// WithField adiciona um erro para um campo específico
func (e *ValidationError) WithField(field, message string) *ValidationError {
	if e.ValidatedFields == nil {
		e.ValidatedFields = make(map[string][]string)
	}

	e.ValidatedFields[field] = append(e.ValidatedFields[field], message)
	return e
}

// WithFields adiciona vários erros de campo de uma vez
func (e *ValidationError) WithFields(fields map[string][]string) *ValidationError {
	if e.ValidatedFields == nil {
		e.ValidatedFields = make(map[string][]string)
	}

	for field, messages := range fields {
		e.ValidatedFields[field] = append(e.ValidatedFields[field], messages...)
	}

	return e
}

// NotFoundError é usado quando um recurso não é encontrado
type NotFoundError struct {
	*DomainError
	ResourceID   string `json:"resource_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		DomainError: New("NOT_FOUND", message).WithType(ErrorTypeNotFound),
	}
}

// WithResource adiciona informações sobre o recurso não encontrado
func (e *NotFoundError) WithResource(resourceType, resourceID string) *NotFoundError {
	e.ResourceType = resourceType
	e.ResourceID = resourceID
	return e
}

// BusinessError representa um erro de regra de negócio
type BusinessError struct {
	*DomainError
	BusinessCode string `json:"business_code,omitempty"`
}

// NewBusinessError cria um erro de regra de negócio
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		DomainError:  New(code, message).WithType(ErrorTypeBusinessRule),
		BusinessCode: code,
	}
}

// InfrastructureError representa um erro de infraestrutura (banco de dados, rede, etc)
type InfrastructureError struct {
	*DomainError
	Component string `json:"component,omitempty"`
}

// NewInfrastructureError cria um erro de infraestrutura
func NewInfrastructureError(component, message string, err error) *InfrastructureError {
	return &InfrastructureError{
		DomainError: NewWithError("INFRA_ERROR", message, err).WithType(ErrorTypeInfrastructure),
		Component:   component,
	}
}

// DatabaseError é um tipo específico de erro de infraestrutura para problemas de banco de dados
type DatabaseError struct {
	*InfrastructureError
	Operation string `json:"operation,omitempty"`
	Table     string `json:"table,omitempty"`
	SQLState  string `json:"sqlstate,omitempty"`
}

// NewDatabaseError cria um erro específico de banco de dados
func NewDatabaseError(message string, err error) *DatabaseError {
	return &DatabaseError{
		InfrastructureError: NewInfrastructureError("database", message, err),
	}
}

// WithOperation adiciona informações sobre a operação que falhou
func (e *DatabaseError) WithOperation(operation, table string) *DatabaseError {
	e.Operation = operation
	e.Table = table
	return e
}

// WithSQLState adiciona informações sobre o código de erro SQL
func (e *DatabaseError) WithSQLState(sqlstate string) *DatabaseError {
	e.SQLState = sqlstate
	return e
}

// ExternalServiceError representa erros de integração com serviços externos
type ExternalServiceError struct {
	*DomainError
	ServiceName string      `json:"service_name,omitempty"`
	HTTPStatus  int         `json:"http_status,omitempty"`
	Response    interface{} `json:"response,omitempty"`
}

// NewExternalServiceError cria um erro de serviço externo
func NewExternalServiceError(serviceName, message string, err error) *ExternalServiceError {
	return &ExternalServiceError{
		DomainError: NewWithError("EXTERNAL_ERROR", message, err).WithType(ErrorTypeExternalService),
		ServiceName: serviceName,
	}
}

// WithStatusCode adiciona o código de status HTTP ao erro
func (e *ExternalServiceError) WithStatusCode(statusCode int) *ExternalServiceError {
	e.HTTPStatus = statusCode
	return e
}

// WithResponse adiciona detalhes da resposta ao erro
func (e *ExternalServiceError) WithResponse(response interface{}) *ExternalServiceError {
	e.Response = response
	return e
}

// AuthenticationError representa erros de autenticação
type AuthenticationError struct {
	*DomainError
	Reason string `json:"reason,omitempty"`
}

// NewAuthenticationError cria um erro de autenticação
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		DomainError: New("AUTH_ERROR", message).WithType(ErrorTypeAuthentication),
	}
}

// WithReason adiciona um motivo para a falha de autenticação
func (e *AuthenticationError) WithReason(reason string) *AuthenticationError {
	e.Reason = reason
	return e
}

// AuthorizationError representa erros de autorização
type AuthorizationError struct {
	*DomainError
	RequiredPermission string `json:"required_permission,omitempty"`
	UserID             string `json:"user_id,omitempty"`
}

// NewAuthorizationError cria um erro de autorização
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		DomainError: New("FORBIDDEN", message).WithType(ErrorTypeAuthorization),
	}
}

// WithRequiredPermission adiciona informações sobre a permissão necessária
func (e *AuthorizationError) WithRequiredPermission(permission, userID string) *AuthorizationError {
	e.RequiredPermission = permission
	e.UserID = userID
	return e
}

// TimeoutError representa erros de timeout
type TimeoutError struct {
	*DomainError
	OperationName string `json:"operation_name,omitempty"`
	Threshold     string `json:"threshold,omitempty"`
}

// NewTimeoutError cria um erro de timeout
func NewTimeoutError(operation, message string) *TimeoutError {
	return &TimeoutError{
		DomainError:   New("TIMEOUT", message).WithType(ErrorTypeTimeout),
		OperationName: operation,
	}
}

// WithThreshold adiciona informações sobre o limite de tempo excedido
func (e *TimeoutError) WithThreshold(threshold string) *TimeoutError {
	e.Threshold = threshold
	return e
}

// UnsupportedOperationError representa erros de operações não suportadas
type UnsupportedOperationError struct {
	*DomainError
	Operation string `json:"operation,omitempty"`
}

// NewUnsupportedOperationError cria um erro de operação não suportada
func NewUnsupportedOperationError(operation, message string) *UnsupportedOperationError {
	return &UnsupportedOperationError{
		DomainError: New("UNSUPPORTED", message).WithType(ErrorTypeUnsupported),
		Operation:   operation,
	}
}

// BadRequestError representa erros de requisição inválida
type BadRequestError struct {
	*DomainError
	InvalidParams map[string]string `json:"invalid_params,omitempty"`
}

// NewBadRequestError cria um erro de requisição inválida
func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		DomainError:   New("BAD_REQUEST", message).WithType(ErrorTypeBadRequest),
		InvalidParams: make(map[string]string),
	}
}

// WithInvalidParam adiciona informações sobre um parâmetro inválido
func (e *BadRequestError) WithInvalidParam(param, reason string) *BadRequestError {
	if e.InvalidParams == nil {
		e.InvalidParams = make(map[string]string)
	}
	e.InvalidParams[param] = reason
	return e
}

// ConflictError representa erros de conflito (quando um recurso já existe)
type ConflictError struct {
	*DomainError
	ConflictingResource string `json:"conflicting_resource,omitempty"`
	ConflictReason      string `json:"conflict_reason,omitempty"`
}

// NewConflictError cria um erro de conflito
func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		DomainError: New("CONFLICT", message).WithType(ErrorTypeConflict),
	}
}

// WithConflictingResource adiciona informações sobre o recurso em conflito
func (e *ConflictError) WithConflictingResource(resource, reason string) *ConflictError {
	e.ConflictingResource = resource
	e.ConflictReason = reason
	return e
}

// RateLimitError representa erros de limite de taxa
type RateLimitError struct {
	*DomainError
	Limit      int    `json:"limit,omitempty"`
	Remaining  int    `json:"remaining,omitempty"`
	ResetTime  string `json:"reset_time,omitempty"`
	RetryAfter string `json:"retry_after,omitempty"`
}

// NewRateLimitError cria um erro de limite de taxa
func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{
		DomainError: New("RATE_LIMIT", message).WithType(ErrorTypeRateLimit),
	}
}

// WithRateLimit adiciona informações sobre o limite de taxa
func (e *RateLimitError) WithRateLimit(limit, remaining int, resetTime, retryAfter string) *RateLimitError {
	e.Limit = limit
	e.Remaining = remaining
	e.ResetTime = resetTime
	e.RetryAfter = retryAfter
	return e
}

// CircuitBreakerError representa erros quando um circuit breaker está aberto
type CircuitBreakerError struct {
	*DomainError
	ServiceName string `json:"service_name,omitempty"`
	State       string `json:"state,omitempty"`
	Failures    int    `json:"failures,omitempty"`
}

// NewCircuitBreakerError cria um erro de circuit breaker
func NewCircuitBreakerError(serviceName, message string) *CircuitBreakerError {
	return &CircuitBreakerError{
		DomainError: New("CIRCUIT_BREAKER", message).WithType(ErrorTypeCircuitBreaker),
		ServiceName: serviceName,
	}
}

// WithCircuitState adiciona informações sobre o estado do circuit breaker
func (e *CircuitBreakerError) WithCircuitState(state string, failures int) *CircuitBreakerError {
	e.State = state
	e.Failures = failures
	return e
}

// ConfigurationError representa erros de configuração
type ConfigurationError struct {
	*DomainError
	ConfigKey   string `json:"config_key,omitempty"`
	ConfigValue string `json:"config_value,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

// NewConfigurationError cria um erro de configuração
func NewConfigurationError(message string) *ConfigurationError {
	return &ConfigurationError{
		DomainError: New("CONFIG_ERROR", message).WithType(ErrorTypeConfiguration),
	}
}

// WithConfigDetails adiciona informações sobre a configuração problemática
func (e *ConfigurationError) WithConfigDetails(key, value, expected string) *ConfigurationError {
	e.ConfigKey = key
	e.ConfigValue = value
	e.Expected = expected
	return e
}

// SecurityError representa erros de segurança
type SecurityError struct {
	*DomainError
	SecurityContext string `json:"security_context,omitempty"`
	ThreatLevel     string `json:"threat_level,omitempty"`
	UserAgent       string `json:"user_agent,omitempty"`
	IPAddress       string `json:"ip_address,omitempty"`
}

// NewSecurityError cria um erro de segurança
func NewSecurityError(message string) *SecurityError {
	return &SecurityError{
		DomainError: New("SECURITY_ERROR", message).WithType(ErrorTypeSecurity),
	}
}

// WithSecurityContext adiciona contexto de segurança
func (e *SecurityError) WithSecurityContext(context, threatLevel string) *SecurityError {
	e.SecurityContext = context
	e.ThreatLevel = threatLevel
	return e
}

// WithClientInfo adiciona informações do cliente
func (e *SecurityError) WithClientInfo(userAgent, ipAddress string) *SecurityError {
	e.UserAgent = userAgent
	e.IPAddress = ipAddress
	return e
}

// ResourceExhaustedError representa erros quando recursos estão esgotados
type ResourceExhaustedError struct {
	*DomainError
	ResourceType string `json:"resource_type,omitempty"`
	Limit        int64  `json:"limit,omitempty"`
	Current      int64  `json:"current,omitempty"`
	Unit         string `json:"unit,omitempty"`
}

// NewResourceExhaustedError cria um erro de recurso esgotado
func NewResourceExhaustedError(resourceType, message string) *ResourceExhaustedError {
	return &ResourceExhaustedError{
		DomainError:  New("RESOURCE_EXHAUSTED", message).WithType(ErrorTypeResourceExhausted),
		ResourceType: resourceType,
	}
}

// WithResourceLimits adiciona informações sobre os limites de recurso
func (e *ResourceExhaustedError) WithResourceLimits(limit, current int64, unit string) *ResourceExhaustedError {
	e.Limit = limit
	e.Current = current
	e.Unit = unit
	return e
}

// DependencyError representa erros de dependências externas
type DependencyError struct {
	*DomainError
	DependencyName string `json:"dependency_name,omitempty"`
	DependencyType string `json:"dependency_type,omitempty"`
	Version        string `json:"version,omitempty"`
	HealthStatus   string `json:"health_status,omitempty"`
}

// NewDependencyError cria um erro de dependência
func NewDependencyError(dependencyName, message string, err error) *DependencyError {
	return &DependencyError{
		DomainError:    NewWithError("DEPENDENCY_ERROR", message, err).WithType(ErrorTypeDependency),
		DependencyName: dependencyName,
	}
}

// WithDependencyInfo adiciona informações sobre a dependência
func (e *DependencyError) WithDependencyInfo(depType, version, healthStatus string) *DependencyError {
	e.DependencyType = depType
	e.Version = version
	e.HealthStatus = healthStatus
	return e
}

// SerializationError representa erros de serialização/deserialização
type SerializationError struct {
	*DomainError
	Format       string `json:"format,omitempty"`
	FieldName    string `json:"field_name,omitempty"`
	ExpectedType string `json:"expected_type,omitempty"`
	ActualType   string `json:"actual_type,omitempty"`
}

// NewSerializationError cria um erro de serialização
func NewSerializationError(format, message string, err error) *SerializationError {
	return &SerializationError{
		DomainError: NewWithError("SERIALIZATION_ERROR", message, err).WithType(ErrorTypeSerialization),
		Format:      format,
	}
}

// WithTypeInfo adiciona informações sobre tipos esperados e atuais
func (e *SerializationError) WithTypeInfo(fieldName, expectedType, actualType string) *SerializationError {
	e.FieldName = fieldName
	e.ExpectedType = expectedType
	e.ActualType = actualType
	return e
}

// CacheError representa erros relacionados a cache
type CacheError struct {
	*DomainError
	CacheType string `json:"cache_type,omitempty"`
	Operation string `json:"operation,omitempty"`
	Key       string `json:"key,omitempty"`
	TTL       string `json:"ttl,omitempty"`
}

// NewCacheError cria um erro de cache
func NewCacheError(cacheType, operation, message string, err error) *CacheError {
	return &CacheError{
		DomainError: NewWithError("CACHE_ERROR", message, err).WithType(ErrorTypeCache),
		CacheType:   cacheType,
		Operation:   operation,
	}
}

// WithCacheDetails adiciona detalhes sobre a operação de cache
func (e *CacheError) WithCacheDetails(key, ttl string) *CacheError {
	e.Key = key
	e.TTL = ttl
	return e
}

// WorkflowError representa erros em workflows ou processos de negócio
type WorkflowError struct {
	*DomainError
	WorkflowID    string `json:"workflow_id,omitempty"`
	StepName      string `json:"step_name,omitempty"`
	CurrentState  string `json:"current_state,omitempty"`
	ExpectedState string `json:"expected_state,omitempty"`
}

// NewWorkflowError cria um erro de workflow
func NewWorkflowError(workflowID, stepName, message string) *WorkflowError {
	return &WorkflowError{
		DomainError: New("WORKFLOW_ERROR", message).WithType(ErrorTypeWorkflow),
		WorkflowID:  workflowID,
		StepName:    stepName,
	}
}

// WithStateInfo adiciona informações sobre o estado do workflow
func (e *WorkflowError) WithStateInfo(currentState, expectedState string) *WorkflowError {
	e.CurrentState = currentState
	e.ExpectedState = expectedState
	return e
}

// MigrationError representa erros durante migrações de dados
type MigrationError struct {
	*DomainError
	MigrationVersion string `json:"migration_version,omitempty"`
	MigrationName    string `json:"migration_name,omitempty"`
	Direction        string `json:"direction,omitempty"`
	AffectedRecords  int64  `json:"affected_records,omitempty"`
}

// NewMigrationError cria um erro de migração
func NewMigrationError(version, name, message string, err error) *MigrationError {
	return &MigrationError{
		DomainError:      NewWithError("MIGRATION_ERROR", message, err).WithType(ErrorTypeMigration),
		MigrationVersion: version,
		MigrationName:    name,
	}
}

// WithMigrationDetails adiciona detalhes sobre a migração
func (e *MigrationError) WithMigrationDetails(direction string, affectedRecords int64) *MigrationError {
	e.Direction = direction
	e.AffectedRecords = affectedRecords
	return e
}

// InvalidSchemaError representa erros de schema inválido
type InvalidSchemaError struct {
	*DomainError
	SchemaName    string              `json:"schema_name,omitempty"`
	SchemaVersion string              `json:"schema_version,omitempty"`
	Details       map[string][]string `json:"details,omitempty"`
}

// NewInvalidSchemaError cria um erro de schema inválido
func NewInvalidSchemaError(message string) *InvalidSchemaError {
	return &InvalidSchemaError{
		DomainError: New("INVALID_SCHEMA", message).WithType(ErrorTypeValidation),
		Details:     make(map[string][]string),
	}
}

// WithSchemaInfo adiciona informações sobre o schema
func (e *InvalidSchemaError) WithSchemaInfo(name, version string) *InvalidSchemaError {
	e.SchemaName = name
	e.SchemaVersion = version
	return e
}

// WithSchemaDetails adiciona detalhes específicos sobre os erros do schema
func (e *InvalidSchemaError) WithSchemaDetails(details map[string][]string) *InvalidSchemaError {
	if e.Details == nil {
		e.Details = make(map[string][]string)
	}

	for field, messages := range details {
		e.Details[field] = append(e.Details[field], messages...)
	}
	return e
}

// UnsupportedMediaTypeError representa erros de tipo de mídia não suportado
type UnsupportedMediaTypeError struct {
	*DomainError
	ProvidedType   string   `json:"provided_type,omitempty"`
	SupportedTypes []string `json:"supported_types,omitempty"`
}

// NewUnsupportedMediaTypeError cria um erro de tipo de mídia não suportado
func NewUnsupportedMediaTypeError(message string) *UnsupportedMediaTypeError {
	return &UnsupportedMediaTypeError{
		DomainError: New("UNSUPPORTED_MEDIA_TYPE", message).WithType(ErrorTypeUnsupported),
	}
}

// WithMediaTypeInfo adiciona informações sobre os tipos de mídia
func (e *UnsupportedMediaTypeError) WithMediaTypeInfo(providedType string, supportedTypes []string) *UnsupportedMediaTypeError {
	e.ProvidedType = providedType
	e.SupportedTypes = supportedTypes
	return e
}

// StatusCode implementa HttpStatusProvider
func (e *UnsupportedMediaTypeError) StatusCode() int {
	return http.StatusUnsupportedMediaType
}

// ServerError representa erros internos do servidor
type ServerError struct {
	*DomainError
	ErrorCode     string         `json:"error_code,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
}

// NewServerError cria um erro interno do servidor
func NewServerError(message string, err error) *ServerError {
	return &ServerError{
		DomainError: NewWithError("SERVER_ERROR", message, err).WithType(ErrorTypeInternal),
		Metadata:    make(map[string]any),
	}
}

// WithErrorCode adiciona um código de erro específico
func (e *ServerError) WithErrorCode(code string) *ServerError {
	e.ErrorCode = code
	return e
}

// WithRequestInfo adiciona informações sobre a requisição
func (e *ServerError) WithRequestInfo(requestID, correlationID string) *ServerError {
	e.RequestID = requestID
	e.CorrelationID = correlationID
	return e
}

// WithMetadata adiciona metadados ao erro
func (e *ServerError) WithMetadata(metadata map[string]any) *ServerError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}

	for key, value := range metadata {
		e.Metadata[key] = value
	}
	return e
}

// StatusCode implementa HttpStatusProvider
func (e *ServerError) StatusCode() int {
	return http.StatusInternalServerError
}

// UnprocessableEntityError representa erros de entidade não processável
type UnprocessableEntityError struct {
	*DomainError
	EntityType       string              `json:"entity_type,omitempty"`
	EntityID         string              `json:"entity_id,omitempty"`
	ValidationErrors map[string][]string `json:"validation_errors,omitempty"`
	BusinessRules    []string            `json:"business_rules,omitempty"`
}

// NewUnprocessableEntityError cria um erro de entidade não processável
func NewUnprocessableEntityError(message string) *UnprocessableEntityError {
	return &UnprocessableEntityError{
		DomainError:      New("UNPROCESSABLE_ENTITY", message).WithType(ErrorTypeUnprocessable),
		ValidationErrors: make(map[string][]string),
		BusinessRules:    make([]string, 0),
	}
}

// WithEntityInfo adiciona informações sobre a entidade
func (e *UnprocessableEntityError) WithEntityInfo(entityType, entityID string) *UnprocessableEntityError {
	e.EntityType = entityType
	e.EntityID = entityID
	return e
}

// WithValidationErrors adiciona erros de validação
func (e *UnprocessableEntityError) WithValidationErrors(errors map[string][]string) *UnprocessableEntityError {
	if e.ValidationErrors == nil {
		e.ValidationErrors = make(map[string][]string)
	}

	for field, messages := range errors {
		e.ValidationErrors[field] = append(e.ValidationErrors[field], messages...)
	}
	return e
}

// WithBusinessRuleViolation adiciona uma violação de regra de negócio
func (e *UnprocessableEntityError) WithBusinessRuleViolation(rule string) *UnprocessableEntityError {
	e.BusinessRules = append(e.BusinessRules, rule)
	return e
}

// StatusCode implementa HttpStatusProvider
func (e *UnprocessableEntityError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// ServiceUnavailableError representa erros quando um serviço está indisponível
type ServiceUnavailableError struct {
	*DomainError
	ServiceName     string `json:"service_name,omitempty"`
	ServiceType     string `json:"service_type,omitempty"`
	RetryAfter      string `json:"retry_after,omitempty"`
	EstimatedUptime string `json:"estimated_uptime,omitempty"`
	HealthEndpoint  string `json:"health_endpoint,omitempty"`
}

// NewServiceUnavailableError cria um erro de serviço indisponível
func NewServiceUnavailableError(serviceName, message string, err error) *ServiceUnavailableError {
	return &ServiceUnavailableError{
		DomainError: NewWithError("SERVICE_UNAVAILABLE", message, err).WithType(ErrorTypeExternalService),
		ServiceName: serviceName,
	}
}

// WithServiceInfo adiciona informações sobre o serviço
func (e *ServiceUnavailableError) WithServiceInfo(serviceType, healthEndpoint string) *ServiceUnavailableError {
	e.ServiceType = serviceType
	e.HealthEndpoint = healthEndpoint
	return e
}

// WithRetryInfo adiciona informações sobre quando tentar novamente
func (e *ServiceUnavailableError) WithRetryInfo(retryAfter, estimatedUptime string) *ServiceUnavailableError {
	e.RetryAfter = retryAfter
	e.EstimatedUptime = estimatedUptime
	return e
}

// StatusCode implementa HttpStatusProvider
func (e *ServiceUnavailableError) StatusCode() int {
	return http.StatusServiceUnavailable
}

// HttpStatusProvider é uma interface para tipos que podem fornecer um código de status HTTP
type HttpStatusProvider interface {
	StatusCode() int
}

// GetStatusCode retorna o código de status HTTP para um erro
// Se o erro implementa HttpStatusProvider, usa esse método
// Caso contrário, tenta mapear com base no tipo do erro
func GetStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Se o erro implementa HttpStatusProvider, use o método
	if provider, ok := err.(HttpStatusProvider); ok {
		return provider.StatusCode()
	}

	// Tente verificar o tipo usando errors.Is e errors.As
	switch {
	case errors.Is(err, ErrNoRows):
		return http.StatusNotFound
	}

	// Verificar tipos específicos usando errors.As
	var (
		validationErr           *ValidationError
		notFoundErr             *NotFoundError
		businessErr             *BusinessError
		infraErr                *InfrastructureError
		dbErr                   *DatabaseError
		extServiceErr           *ExternalServiceError
		authErr                 *AuthenticationError
		authzErr                *AuthorizationError
		timeoutErr              *TimeoutError
		unsupportedOpErr        *UnsupportedOperationError
		badReqErr               *BadRequestError
		conflictErr             *ConflictError
		rateLimitErr            *RateLimitError
		circuitBreakerErr       *CircuitBreakerError
		configErr               *ConfigurationError
		securityErr             *SecurityError
		resourceExhaustedErr    *ResourceExhaustedError
		dependencyErr           *DependencyError
		serializationErr        *SerializationError
		cacheErr                *CacheError
		workflowErr             *WorkflowError
		migrationErr            *MigrationError
		invalidSchemaErr        *InvalidSchemaError
		unsupportedMediaTypeErr *UnsupportedMediaTypeError
		serverErr               *ServerError
		unprocessableEntityErr  *UnprocessableEntityError
		serviceUnavailableErr   *ServiceUnavailableError
	)

	switch {
	case errors.As(err, &validationErr):
		return validationErr.StatusCode()
	case errors.As(err, &notFoundErr):
		return notFoundErr.StatusCode()
	case errors.As(err, &businessErr):
		return businessErr.StatusCode()
	case errors.As(err, &infraErr):
		return infraErr.StatusCode()
	case errors.As(err, &dbErr):
		return dbErr.StatusCode()
	case errors.As(err, &extServiceErr):
		return extServiceErr.StatusCode()
	case errors.As(err, &authErr):
		return authErr.StatusCode()
	case errors.As(err, &authzErr):
		return authzErr.StatusCode()
	case errors.As(err, &timeoutErr):
		return timeoutErr.StatusCode()
	case errors.As(err, &unsupportedOpErr):
		return unsupportedOpErr.StatusCode()
	case errors.As(err, &badReqErr):
		return badReqErr.StatusCode()
	case errors.As(err, &conflictErr):
		return conflictErr.StatusCode()
	case errors.As(err, &rateLimitErr):
		return rateLimitErr.StatusCode()
	case errors.As(err, &circuitBreakerErr):
		return circuitBreakerErr.StatusCode()
	case errors.As(err, &configErr):
		return configErr.StatusCode()
	case errors.As(err, &securityErr):
		return securityErr.StatusCode()
	case errors.As(err, &resourceExhaustedErr):
		return resourceExhaustedErr.StatusCode()
	case errors.As(err, &dependencyErr):
		return dependencyErr.StatusCode()
	case errors.As(err, &serializationErr):
		return serializationErr.StatusCode()
	case errors.As(err, &cacheErr):
		return cacheErr.StatusCode()
	case errors.As(err, &workflowErr):
		return workflowErr.StatusCode()
	case errors.As(err, &migrationErr):
		return migrationErr.StatusCode()
	case errors.As(err, &invalidSchemaErr):
		return invalidSchemaErr.StatusCode()
	case errors.As(err, &unsupportedMediaTypeErr):
		return unsupportedMediaTypeErr.StatusCode()
	case errors.As(err, &serverErr):
		return serverErr.StatusCode()
	case errors.As(err, &unprocessableEntityErr):
		return unprocessableEntityErr.StatusCode()
	case errors.As(err, &serviceUnavailableErr):
		return serviceUnavailableErr.StatusCode()
	default:
		return http.StatusInternalServerError
	}
}

// StatusCode implementa HttpStatusProvider para ValidationError
func (e *ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

// StatusCode implementa HttpStatusProvider para NotFoundError
func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

// StatusCode implementa HttpStatusProvider para BusinessError
func (e *BusinessError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// StatusCode implementa HttpStatusProvider para InfrastructureError
func (e *InfrastructureError) StatusCode() int {
	return http.StatusInternalServerError
}

// StatusCode implementa HttpStatusProvider para DatabaseError
func (e *DatabaseError) StatusCode() int {
	return http.StatusInternalServerError
}

// StatusCode implementa HttpStatusProvider para ExternalServiceError
func (e *ExternalServiceError) StatusCode() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return http.StatusBadGateway
}

// StatusCode implementa HttpStatusProvider para AuthenticationError
func (e *AuthenticationError) StatusCode() int {
	return http.StatusUnauthorized
}

// StatusCode implementa HttpStatusProvider para AuthorizationError
func (e *AuthorizationError) StatusCode() int {
	return http.StatusForbidden
}

// StatusCode implementa HttpStatusProvider para TimeoutError
func (e *TimeoutError) StatusCode() int {
	return http.StatusRequestTimeout
}

// StatusCode implementa HttpStatusProvider para UnsupportedOperationError
func (e *UnsupportedOperationError) StatusCode() int {
	return http.StatusNotImplemented
}

// StatusCode implementa HttpStatusProvider para BadRequestError
func (e *BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

// StatusCode implementa HttpStatusProvider para ConflictError
func (e *ConflictError) StatusCode() int {
	return http.StatusConflict
}

// StatusCode implementa HttpStatusProvider para RateLimitError
func (e *RateLimitError) StatusCode() int {
	return http.StatusTooManyRequests
}

// StatusCode implementa HttpStatusProvider para CircuitBreakerError
func (e *CircuitBreakerError) StatusCode() int {
	return http.StatusServiceUnavailable
}

// StatusCode implementa HttpStatusProvider para ConfigurationError
func (e *ConfigurationError) StatusCode() int {
	return http.StatusInternalServerError
}

// StatusCode implementa HttpStatusProvider para SecurityError
func (e *SecurityError) StatusCode() int {
	return http.StatusForbidden
}

// StatusCode implementa HttpStatusProvider para ResourceExhaustedError
func (e *ResourceExhaustedError) StatusCode() int {
	return http.StatusInsufficientStorage
}

// StatusCode implementa HttpStatusProvider para DependencyError
func (e *DependencyError) StatusCode() int {
	return http.StatusFailedDependency
}

// StatusCode implementa HttpStatusProvider para SerializationError
func (e *SerializationError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// StatusCode implementa HttpStatusProvider para CacheError
func (e *CacheError) StatusCode() int {
	return http.StatusInternalServerError
}

// StatusCode implementa HttpStatusProvider para WorkflowError
func (e *WorkflowError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// StatusCode implementa HttpStatusProvider para MigrationError
func (e *MigrationError) StatusCode() int {
	return http.StatusInternalServerError
}

// StatusCode implementa HttpStatusProvider para InvalidSchemaError
func (e *InvalidSchemaError) StatusCode() int {
	return http.StatusBadRequest
}
