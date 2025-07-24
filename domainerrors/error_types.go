package domainerrors

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ValidationError representa erros de validação
type ValidationError struct {
	*DomainError
	Fields map[string][]string `json:"fields,omitempty"`
}

// NewValidationError cria um erro de validação
func NewValidationError(code, message string, cause error) *ValidationError {
	return &ValidationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeValidation,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Fields: make(map[string][]string),
	}
}

// WithField adiciona um campo com erro de validação
func (e *ValidationError) WithField(field, message string) *ValidationError {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
	return e
}

// NotFoundError representa erros de recurso não encontrado
type NotFoundError struct {
	*DomainError
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
}

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(code, message string, cause error) *NotFoundError {
	return &NotFoundError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeNotFound,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithResource adiciona informações do recurso não encontrado
func (e *NotFoundError) WithResource(resourceType, resourceID string) *NotFoundError {
	e.ResourceType = resourceType
	e.ResourceID = resourceID
	return e
}

// BusinessError representa erros de regra de negócio
type BusinessError struct {
	*DomainError
	BusinessCode string   `json:"business_code,omitempty"`
	Rules        []string `json:"rules,omitempty"`
}

// NewBusinessError cria um erro de regra de negócio
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeBusinessRule,
			Timestamp: time.Now(),
		},
		BusinessCode: code,
		Rules:        make([]string, 0),
	}
}

// WithRule adiciona uma regra de negócio violada
func (e *BusinessError) WithRule(rule string) *BusinessError {
	e.Rules = append(e.Rules, rule)
	return e
}

// DatabaseError representa erros de banco de dados
type DatabaseError struct {
	*DomainError
	Operation string `json:"operation,omitempty"`
	Table     string `json:"table,omitempty"`
	Query     string `json:"query,omitempty"`
}

// NewDatabaseError cria um erro de banco de dados
func NewDatabaseError(code, message string, cause error) *DatabaseError {
	return &DatabaseError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeDatabase,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithOperation adiciona informações da operação
func (e *DatabaseError) WithOperation(operation, table string) *DatabaseError {
	e.Operation = operation
	e.Table = table
	return e
}

// WithQuery adiciona a query que falhou
func (e *DatabaseError) WithQuery(query string) *DatabaseError {
	e.Query = query
	return e
}

// ExternalServiceError representa erros de serviços externos
type ExternalServiceError struct {
	*DomainError
	Service    string `json:"service,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Response   string `json:"response,omitempty"`
}

// NewExternalServiceError cria um erro de serviço externo
func NewExternalServiceError(code, service, message string, cause error) *ExternalServiceError {
	return &ExternalServiceError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeExternalService,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Service: service,
	}
}

// WithEndpoint adiciona informações do endpoint
func (e *ExternalServiceError) WithEndpoint(endpoint string) *ExternalServiceError {
	e.Endpoint = endpoint
	return e
}

// WithResponse adiciona informações da resposta
func (e *ExternalServiceError) WithResponse(statusCode int, response string) *ExternalServiceError {
	e.StatusCode = statusCode
	e.Response = response
	return e
}

// InfrastructureError representa erros de infraestrutura
type InfrastructureError struct {
	*DomainError
	Component string `json:"component,omitempty"`
	Action    string `json:"action,omitempty"`
}

// NewInfrastructureError cria um erro de infraestrutura
func NewInfrastructureError(code, component, message string, cause error) *InfrastructureError {
	return &InfrastructureError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeInfrastructure,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Component: component,
	}
}

// WithAction adiciona informações da ação que falhou
func (e *InfrastructureError) WithAction(action string) *InfrastructureError {
	e.Action = action
	return e
}

// DependencyError representa erros de dependência
type DependencyError struct {
	*DomainError
	Dependency string `json:"dependency,omitempty"`
	Version    string `json:"version,omitempty"`
	Status     string `json:"status,omitempty"`
}

// NewDependencyError cria um erro de dependência
func NewDependencyError(code, dependency, message string, cause error) *DependencyError {
	return &DependencyError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeDependency,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Dependency: dependency,
	}
}

// WithDependencyInfo adiciona informações da dependência
func (e *DependencyError) WithDependencyInfo(version, status string) *DependencyError {
	e.Version = version
	e.Status = status
	return e
}

// AuthenticationError representa erros de autenticação
type AuthenticationError struct {
	*DomainError
	Scheme string `json:"scheme,omitempty"`
	Token  string `json:"token,omitempty"`
}

// NewAuthenticationError cria um erro de autenticação
func NewAuthenticationError(code, message string, cause error) *AuthenticationError {
	return &AuthenticationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeAuthentication,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithScheme adiciona o esquema de autenticação
func (e *AuthenticationError) WithScheme(scheme string) *AuthenticationError {
	e.Scheme = scheme
	return e
}

// AuthorizationError representa erros de autorização
type AuthorizationError struct {
	*DomainError
	Permission string `json:"permission,omitempty"`
	Resource   string `json:"resource,omitempty"`
}

// NewAuthorizationError cria um erro de autorização
func NewAuthorizationError(code, message string, cause error) *AuthorizationError {
	return &AuthorizationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeAuthorization,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithPermission adiciona informações da permissão
func (e *AuthorizationError) WithPermission(permission, resource string) *AuthorizationError {
	e.Permission = permission
	e.Resource = resource
	return e
}

// SecurityError representa erros de segurança
type SecurityError struct {
	*DomainError
	ThreatType string `json:"threat_type,omitempty"`
	Severity   string `json:"severity,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
}

// NewSecurityError cria um erro de segurança
func NewSecurityError(code, message string) *SecurityError {
	return &SecurityError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeSecurity,
			Timestamp: time.Now(),
		},
	}
}

// WithThreat adiciona informações da ameaça
func (e *SecurityError) WithThreat(threatType, severity string) *SecurityError {
	e.ThreatType = threatType
	e.Severity = severity
	return e
}

// WithClientIP adiciona o IP do cliente
func (e *SecurityError) WithClientIP(clientIP string) *SecurityError {
	e.ClientIP = clientIP
	return e
}

// TimeoutError representa erros de timeout
type TimeoutError struct {
	*DomainError
	Operation string        `json:"operation,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"`
}

// NewTimeoutError cria um erro de timeout
func NewTimeoutError(code, operation, message string, cause error) *TimeoutError {
	return &TimeoutError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeTimeout,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Operation: operation,
	}
}

// WithDuration adiciona informações de duração
func (e *TimeoutError) WithDuration(duration, timeout time.Duration) *TimeoutError {
	e.Duration = duration
	e.Timeout = timeout
	return e
}

// RateLimitError representa erros de limite de taxa
type RateLimitError struct {
	*DomainError
	Limit     int    `json:"limit,omitempty"`
	Remaining int    `json:"remaining,omitempty"`
	ResetTime string `json:"reset_time,omitempty"`
	Window    string `json:"window,omitempty"`
}

// NewRateLimitError cria um erro de limite de taxa
func NewRateLimitError(code, message string) *RateLimitError {
	return &RateLimitError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeRateLimit,
			Timestamp: time.Now(),
		},
	}
}

// WithRateLimit adiciona informações do limite de taxa
func (e *RateLimitError) WithRateLimit(limit, remaining int, resetTime, window string) *RateLimitError {
	e.Limit = limit
	e.Remaining = remaining
	e.ResetTime = resetTime
	e.Window = window
	return e
}

// ResourceExhaustedError representa erros de recurso esgotado
type ResourceExhaustedError struct {
	*DomainError
	Resource string `json:"resource,omitempty"`
	Limit    int64  `json:"limit,omitempty"`
	Used     int64  `json:"used,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

// NewResourceExhaustedError cria um erro de recurso esgotado
func NewResourceExhaustedError(code, resource, message string) *ResourceExhaustedError {
	return &ResourceExhaustedError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeResourceExhausted,
			Timestamp: time.Now(),
		},
		Resource: resource,
	}
}

// WithLimits adiciona informações dos limites do recurso
func (e *ResourceExhaustedError) WithLimits(limit, used int64, unit string) *ResourceExhaustedError {
	e.Limit = limit
	e.Used = used
	e.Unit = unit
	return e
}

// CircuitBreakerError representa erros de circuit breaker
type CircuitBreakerError struct {
	*DomainError
	CircuitName string `json:"circuit_name,omitempty"`
	State       string `json:"state,omitempty"`
	Failures    int    `json:"failures,omitempty"`
	Timeout     string `json:"timeout,omitempty"`
}

// NewCircuitBreakerError cria um erro de circuit breaker
func NewCircuitBreakerError(code, circuitName, message string) *CircuitBreakerError {
	return &CircuitBreakerError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeCircuitBreaker,
			Timestamp: time.Now(),
		},
		CircuitName: circuitName,
	}
}

// WithCircuitState adiciona informações do estado do circuit breaker
func (e *CircuitBreakerError) WithCircuitState(state string, failures int) *CircuitBreakerError {
	e.State = state
	e.Failures = failures
	return e
}

// SerializationError representa erros de serialização
type SerializationError struct {
	*DomainError
	Format   string `json:"format,omitempty"`
	Field    string `json:"field,omitempty"`
	Expected string `json:"expected,omitempty"`
	Received string `json:"received,omitempty"`
}

// NewSerializationError cria um erro de serialização
func NewSerializationError(code, format, message string, cause error) *SerializationError {
	return &SerializationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeSerialization,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Format: format,
	}
}

// WithTypeInfo adiciona informações do tipo
func (e *SerializationError) WithTypeInfo(field, expected, received string) *SerializationError {
	e.Field = field
	e.Expected = expected
	e.Received = received
	return e
}

// CacheError representa erros de cache
type CacheError struct {
	*DomainError
	CacheType string `json:"cache_type,omitempty"`
	Operation string `json:"operation,omitempty"`
	Key       string `json:"key,omitempty"`
	TTL       string `json:"ttl,omitempty"`
}

// NewCacheError cria um erro de cache
func NewCacheError(code, cacheType, operation, message string, cause error) *CacheError {
	return &CacheError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeCache,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		CacheType: cacheType,
		Operation: operation,
	}
}

// WithCacheDetails adiciona detalhes do cache
func (e *CacheError) WithCacheDetails(key, ttl string) *CacheError {
	e.Key = key
	e.TTL = ttl
	return e
}

// MigrationError representa erros de migração
type MigrationError struct {
	*DomainError
	Version string `json:"version,omitempty"`
	Script  string `json:"script,omitempty"`
	Stage   string `json:"stage,omitempty"`
}

// NewMigrationError cria um erro de migração
func NewMigrationError(code, version, message string, cause error) *MigrationError {
	return &MigrationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeMigration,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		Version: version,
	}
}

// WithMigrationDetails adiciona detalhes da migração
func (e *MigrationError) WithMigrationDetails(script, stage string) *MigrationError {
	e.Script = script
	e.Stage = stage
	return e
}

// ConfigurationError representa erros de configuração
type ConfigurationError struct {
	*DomainError
	ConfigKey string `json:"config_key,omitempty"`
	Expected  string `json:"expected,omitempty"`
	Received  string `json:"received,omitempty"`
}

// NewConfigurationError cria um erro de configuração
func NewConfigurationError(code, configKey, message string, cause error) *ConfigurationError {
	return &ConfigurationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeConfiguration,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		ConfigKey: configKey,
	}
}

// WithConfigDetails adiciona detalhes da configuração
func (e *ConfigurationError) WithConfigDetails(expected, received string) *ConfigurationError {
	e.Expected = expected
	e.Received = received
	return e
}

// UnsupportedOperationError representa operações não suportadas
type UnsupportedOperationError struct {
	*DomainError
	Operation string `json:"operation,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewUnsupportedOperationError cria um erro de operação não suportada
func NewUnsupportedOperationError(code, operation, message string) *UnsupportedOperationError {
	return &UnsupportedOperationError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeUnsupported,
			Timestamp: time.Now(),
		},
		Operation: operation,
	}
}

// WithReason adiciona o motivo da operação não suportada
func (e *UnsupportedOperationError) WithReason(reason string) *UnsupportedOperationError {
	e.Reason = reason
	return e
}

// BadRequestError representa erros de requisição inválida
type BadRequestError struct {
	*DomainError
	Parameter string `json:"parameter,omitempty"`
	Expected  string `json:"expected,omitempty"`
	Received  string `json:"received,omitempty"`
}

// NewBadRequestError cria um erro de requisição inválida
func NewBadRequestError(code, message string, cause error) *BadRequestError {
	return &BadRequestError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeBadRequest,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithParameter adiciona informações do parâmetro inválido
func (e *BadRequestError) WithParameter(parameter, expected, received string) *BadRequestError {
	e.Parameter = parameter
	e.Expected = expected
	e.Received = received
	return e
}

// ConflictError representa erros de conflito
type ConflictError struct {
	*DomainError
	Resource       string `json:"resource,omitempty"`
	ConflictReason string `json:"conflict_reason,omitempty"`
}

// NewConflictError cria um erro de conflito
func NewConflictError(code, message string) *ConflictError {
	return &ConflictError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeConflict,
			Timestamp: time.Now(),
		},
	}
}

// WithConflictingResource adiciona informações do recurso em conflito
func (e *ConflictError) WithConflictingResource(resource, reason string) *ConflictError {
	e.Resource = resource
	e.ConflictReason = reason
	return e
}

// InvalidSchemaError representa erros de schema inválido
type InvalidSchemaError struct {
	*DomainError
	SchemaName string              `json:"schema_name,omitempty"`
	Version    string              `json:"version,omitempty"`
	Details    map[string][]string `json:"details,omitempty"`
}

// NewInvalidSchemaError cria um erro de schema inválido
func NewInvalidSchemaError(code, message string) *InvalidSchemaError {
	return &InvalidSchemaError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeInvalidSchema,
			Timestamp: time.Now(),
		},
		Details: make(map[string][]string),
	}
}

// WithSchemaInfo adiciona informações do schema
func (e *InvalidSchemaError) WithSchemaInfo(schemaName, version string) *InvalidSchemaError {
	e.SchemaName = schemaName
	e.Version = version
	return e
}

// WithSchemaDetails adiciona detalhes de validação do schema
func (e *InvalidSchemaError) WithSchemaDetails(details map[string][]string) *InvalidSchemaError {
	e.Details = details
	return e
}

// UnsupportedMediaTypeError representa erros de tipo de mídia não suportado
type UnsupportedMediaTypeError struct {
	*DomainError
	MediaType string   `json:"media_type,omitempty"`
	Supported []string `json:"supported,omitempty"`
}

// NewUnsupportedMediaTypeError cria um erro de tipo de mídia não suportado
func NewUnsupportedMediaTypeError(code, mediaType, message string) *UnsupportedMediaTypeError {
	return &UnsupportedMediaTypeError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeUnsupportedMedia,
			Timestamp: time.Now(),
		},
		MediaType: mediaType,
	}
}

// WithSupportedTypes adiciona tipos suportados
func (e *UnsupportedMediaTypeError) WithSupportedTypes(supported []string) *UnsupportedMediaTypeError {
	e.Supported = supported
	return e
}

// ServerError representa erros internos do servidor
type ServerError struct {
	*DomainError
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Component     string `json:"component,omitempty"`
}

// NewServerError cria um erro interno do servidor
func NewServerError(code, message string, cause error) *ServerError {
	return &ServerError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeServer,
			Cause:     cause,
			Timestamp: time.Now(),
		},
	}
}

// WithRequestInfo adiciona informações da requisição
func (e *ServerError) WithRequestInfo(requestID, correlationID string) *ServerError {
	e.RequestID = requestID
	e.CorrelationID = correlationID
	return e
}

// WithComponent adiciona informações do componente
func (e *ServerError) WithComponent(component string) *ServerError {
	e.Component = component
	return e
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
func NewUnprocessableEntityError(code, message string) *UnprocessableEntityError {
	return &UnprocessableEntityError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeUnprocessable,
			Timestamp: time.Now(),
		},
		ValidationErrors: make(map[string][]string),
		BusinessRules:    make([]string, 0),
	}
}

// WithEntityInfo adiciona informações da entidade
func (e *UnprocessableEntityError) WithEntityInfo(entityType, entityID string) *UnprocessableEntityError {
	e.EntityType = entityType
	e.EntityID = entityID
	return e
}

// WithValidationErrors adiciona erros de validação
func (e *UnprocessableEntityError) WithValidationErrors(validationErrors map[string][]string) *UnprocessableEntityError {
	e.ValidationErrors = validationErrors
	return e
}

// WithBusinessRuleViolation adiciona violação de regra de negócio
func (e *UnprocessableEntityError) WithBusinessRuleViolation(rule string) *UnprocessableEntityError {
	e.BusinessRules = append(e.BusinessRules, rule)
	return e
}

// ServiceUnavailableError representa erros de serviço indisponível
type ServiceUnavailableError struct {
	*DomainError
	ServiceName string `json:"service_name,omitempty"`
	RetryAfter  string `json:"retry_after,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
}

// NewServiceUnavailableError cria um erro de serviço indisponível
func NewServiceUnavailableError(code, serviceName, message string, cause error) *ServiceUnavailableError {
	return &ServiceUnavailableError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeServiceUnavailable,
			Cause:     cause,
			Timestamp: time.Now(),
		},
		ServiceName: serviceName,
	}
}

// WithRetryInfo adiciona informações de retry
func (e *ServiceUnavailableError) WithRetryInfo(retryAfter string) *ServiceUnavailableError {
	e.RetryAfter = retryAfter
	return e
}

// WithEndpoint adiciona informações do endpoint
func (e *ServiceUnavailableError) WithEndpoint(endpoint string) *ServiceUnavailableError {
	e.Endpoint = endpoint
	return e
}

// WorkflowError representa erros de workflow
type WorkflowError struct {
	*DomainError
	WorkflowName  string `json:"workflow_name,omitempty"`
	StepName      string `json:"step_name,omitempty"`
	CurrentState  string `json:"current_state,omitempty"`
	ExpectedState string `json:"expected_state,omitempty"`
}

// NewWorkflowError cria um erro de workflow
func NewWorkflowError(code, workflowName, stepName, message string) *WorkflowError {
	return &WorkflowError{
		DomainError: &DomainError{
			Code:      code,
			Message:   message,
			Type:      ErrorTypeWorkflow,
			Timestamp: time.Now(),
		},
		WorkflowName: workflowName,
		StepName:     stepName,
	}
}

// WithStateInfo adiciona informações do estado
func (e *WorkflowError) WithStateInfo(currentState, expectedState string) *WorkflowError {
	e.CurrentState = currentState
	e.ExpectedState = expectedState
	return e
}

// Wrap é uma função utilitária para empilhar erros com código
// Cria um novo DomainError encapsulando o erro original mantendo a cadeia de erros
func Wrap(code, message string, cause error) *DomainError {
	if cause == nil {
		return nil
	}

	// Function to extract DomainError from various error types
	extractDomainError := func(err error) *DomainError {
		switch e := err.(type) {
		case *DomainError:
			return e
		case *ValidationError:
			return e.DomainError
		case *NotFoundError:
			return e.DomainError
		case *BusinessError:
			return e.DomainError
		case *DatabaseError:
			return e.DomainError
		case *ExternalServiceError:
			return e.DomainError
		case *TimeoutError:
			return e.DomainError
		case *RateLimitError:
			return e.DomainError
		case *CircuitBreakerError:
			return e.DomainError
		case *InvalidSchemaError:
			return e.DomainError
		case *ServerError:
			return e.DomainError
		case *UnprocessableEntityError:
			return e.DomainError
		case *ServiceUnavailableError:
			return e.DomainError
		case *WorkflowError:
			return e.DomainError
		case *InfrastructureError:
			return e.DomainError
		case *AuthenticationError:
			return e.DomainError
		case *AuthorizationError:
			return e.DomainError
		case *SecurityError:
			return e.DomainError
		case *ResourceExhaustedError:
			return e.DomainError
		case *DependencyError:
			return e.DomainError
		case *BadRequestError:
			return e.DomainError
		case *ConflictError:
			return e.DomainError
		case *UnsupportedMediaTypeError:
			return e.DomainError
		case *MigrationError:
			return e.DomainError
		case *ConfigurationError:
			return e.DomainError
		case *UnsupportedOperationError:
			return e.DomainError
		case *SerializationError:
			return e.DomainError
		case *CacheError:
			return e.DomainError
		}
		return nil
	}

	// Se o erro original já é um DomainError (ou tem DomainError embutido), preserve suas informações
	if domainErr := extractDomainError(cause); domainErr != nil {
		// Cria novo erro mantendo o contexto original
		wrapped := &DomainError{
			Code:      code,
			Message:   message,
			Type:      domainErr.Type, // Preserva o tipo original
			Cause:     cause,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		// Copia metadados importantes do erro original
		if domainErr.Metadata != nil {
			for k, v := range domainErr.Metadata {
				wrapped.Metadata[k] = v
			}
		}

		wrapped.captureStackFrame("error wrapped")
		return wrapped
	}

	// Para erros não-domain, cria um novo erro server
	err := &DomainError{
		Code:      code,
		Message:   message,
		Type:      ErrorTypeServer,
		Cause:     cause,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	err.captureStackFrame("error wrapped")
	return err
}

// WrapWithContext é uma função utilitária para empilhar erros com contexto
// Adiciona informações do contexto ao erro encapsulado
func WrapWithContext(ctx context.Context, code, message string, cause error) *DomainError {
	if cause == nil {
		return nil
	}

	wrapped := Wrap(code, message, cause)

	// Adiciona informações do contexto aos metadados
	if ctx != nil {
		// Extrai informações úteis do contexto
		if deadline, ok := ctx.Deadline(); ok {
			wrapped.WithMetadata("context_deadline", deadline.Format(time.RFC3339))
		}

		// Adiciona timeout se disponível
		if ctx.Err() == context.DeadlineExceeded {
			wrapped.WithMetadata("context_timeout", true)
		}

		// Adiciona cancelamento se disponível
		if ctx.Err() == context.Canceled {
			wrapped.WithMetadata("context_canceled", true)
		}

		// Adiciona valores do contexto se existirem
		if traceID := ctx.Value("trace_id"); traceID != nil {
			wrapped.WithMetadata("trace_id", traceID)
		}

		if requestID := ctx.Value("request_id"); requestID != nil {
			wrapped.WithMetadata("request_id", requestID)
		}
	}

	wrapped.captureStackFrame("error wrapped with context")
	return wrapped
}

// WrapWithType é uma função utilitária para empilhar erros com tipo específico
// Permite especificar o tipo de erro ao fazer wrap
func WrapWithType(code, message string, errorType ErrorType, cause error) *DomainError {
	if cause == nil {
		return nil
	}

	// Function to extract DomainError from various error types
	extractDomainError := func(err error) *DomainError {
		switch e := err.(type) {
		case *DomainError:
			return e
		case *ValidationError:
			return e.DomainError
		case *NotFoundError:
			return e.DomainError
		case *BusinessError:
			return e.DomainError
		case *DatabaseError:
			return e.DomainError
		case *ExternalServiceError:
			return e.DomainError
		case *TimeoutError:
			return e.DomainError
		case *RateLimitError:
			return e.DomainError
		case *CircuitBreakerError:
			return e.DomainError
		case *InvalidSchemaError:
			return e.DomainError
		case *ServerError:
			return e.DomainError
		case *UnprocessableEntityError:
			return e.DomainError
		case *ServiceUnavailableError:
			return e.DomainError
		case *WorkflowError:
			return e.DomainError
		case *InfrastructureError:
			return e.DomainError
		case *AuthenticationError:
			return e.DomainError
		case *AuthorizationError:
			return e.DomainError
		case *SecurityError:
			return e.DomainError
		case *ResourceExhaustedError:
			return e.DomainError
		case *DependencyError:
			return e.DomainError
		case *BadRequestError:
			return e.DomainError
		case *ConflictError:
			return e.DomainError
		case *UnsupportedMediaTypeError:
			return e.DomainError
		case *MigrationError:
			return e.DomainError
		case *ConfigurationError:
			return e.DomainError
		case *UnsupportedOperationError:
			return e.DomainError
		case *SerializationError:
			return e.DomainError
		case *CacheError:
			return e.DomainError
		}
		return nil
	}

	err := &DomainError{
		Code:      code,
		Message:   message,
		Type:      errorType,
		Cause:     cause,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Se o erro original é um DomainError, copia metadados relevantes
	if domainErr := extractDomainError(cause); domainErr != nil && domainErr.Metadata != nil {
		for k, v := range domainErr.Metadata {
			err.Metadata[k] = v
		}
	}

	err.captureStackFrame("error wrapped with type")
	return err
}

// WrapChain é uma função utilitária para criar uma cadeia de erros
// Permite adicionar múltiplas camadas de contexto ao erro
func WrapChain(code, message string, cause error, layers ...string) *DomainError {
	if cause == nil {
		return nil
	}

	wrapped := Wrap(code, message, cause)

	// Adiciona camadas de contexto
	for i, layer := range layers {
		wrapped.WithMetadata(fmt.Sprintf("layer_%d", i), layer)
	}

	wrapped.captureStackFrame("error wrapped in chain")
	return wrapped
}

// FormatStackTrace retorna uma string formatada do stack de erros
func (e *DomainError) FormatStackTrace() string {
	if len(e.Stack) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Error Stack Trace:\n")

	for i, st := range e.Stack {
		b.WriteString(fmt.Sprintf("%d: [%s] in %s (%s:%d)\n",
			i+1, st.Message, st.Function, st.File, st.Line))
	}

	return b.String()
}
