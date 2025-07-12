// Package types define os tipos de erro do sistema de domínio.
// Implementa uma hierarquia bem definida de tipos para facilitar o tratamento específico.
package types

import (
	"net/http"
)

// ErrorType representa o tipo de erro para categorização e tratamento específico.
type ErrorType string

// Tipos de erro padrão do sistema
const (
	// Erros de dados e persistência
	ErrorTypeRepository    ErrorType = "repository"    // Erros relacionados a acesso a dados
	ErrorTypeDatabase      ErrorType = "database"      // Erros específicos de banco de dados
	ErrorTypeCache         ErrorType = "cache"         // Erros de cache
	ErrorTypeMigration     ErrorType = "migration"     // Erros de migração de dados
	ErrorTypeSerialization ErrorType = "serialization" // Erros de serialização/deserialização

	// Erros de validação e entrada
	ErrorTypeValidation    ErrorType = "validation"    // Erros de validação de dados
	ErrorTypeBadRequest    ErrorType = "bad_request"   // Erros de requisição mal formada
	ErrorTypeUnprocessable ErrorType = "unprocessable" // Entidade não processável
	ErrorTypeUnsupported   ErrorType = "unsupported"   // Operação não suportada

	// Erros de negócio
	ErrorTypeBusinessRule ErrorType = "business"  // Erros de regras de negócio
	ErrorTypeWorkflow     ErrorType = "workflow"  // Erros de fluxo de trabalho
	ErrorTypeConflict     ErrorType = "conflict"  // Erros de conflito/duplicação
	ErrorTypeNotFound     ErrorType = "not_found" // Recurso não encontrado

	// Erros de segurança e autorização
	ErrorTypeAuthentication ErrorType = "authentication" // Erros de autenticação
	ErrorTypeAuthorization  ErrorType = "authorization"  // Erros de autorização
	ErrorTypeSecurity       ErrorType = "security"       // Erros gerais de segurança
	ErrorTypeForbidden      ErrorType = "forbidden"      // Acesso negado

	// Erros de sistema e infraestrutura
	ErrorTypeInternal       ErrorType = "internal"        // Erros internos gerais
	ErrorTypeInfrastructure ErrorType = "infrastructure"  // Erros de infraestrutura
	ErrorTypeConfiguration  ErrorType = "configuration"   // Erros de configuração
	ErrorTypeDependency     ErrorType = "dependency"      // Erros de dependências
	ErrorTypeCircuitBreaker ErrorType = "circuit_breaker" // Erros de circuit breaker

	// Erros de comunicação e rede
	ErrorTypeExternalService   ErrorType = "external_service"   // Erros de serviços externos
	ErrorTypeTimeout           ErrorType = "timeout"            // Erros de timeout
	ErrorTypeRateLimit         ErrorType = "rate_limit"         // Erros de limite de taxa
	ErrorTypeResourceExhausted ErrorType = "resource_exhausted" // Recursos esgotados
	ErrorTypeNetwork           ErrorType = "network"            // Erros de rede
	ErrorTypeCloud             ErrorType = "cloud"              // Erros de serviços cloud (AWS, GCP, Azure)

	// Erros de protocolo
	ErrorTypeHTTP      ErrorType = "http"      // Erros HTTP específicos
	ErrorTypeGRPC      ErrorType = "grpc"      // Erros gRPC específicos
	ErrorTypeGraphQL   ErrorType = "graphql"   // Erros GraphQL específicos
	ErrorTypeWebSocket ErrorType = "websocket" // Erros WebSocket específicos
)

// String retorna a representação string do tipo de erro.
func (et ErrorType) String() string {
	return string(et)
}

// IsValid verifica se o tipo de erro é válido.
func (et ErrorType) IsValid() bool {
	switch et {
	case ErrorTypeRepository, ErrorTypeDatabase, ErrorTypeCache, ErrorTypeMigration, ErrorTypeSerialization,
		ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeUnprocessable, ErrorTypeUnsupported,
		ErrorTypeBusinessRule, ErrorTypeWorkflow, ErrorTypeConflict, ErrorTypeNotFound,
		ErrorTypeAuthentication, ErrorTypeAuthorization, ErrorTypeSecurity, ErrorTypeForbidden,
		ErrorTypeInternal, ErrorTypeInfrastructure, ErrorTypeConfiguration, ErrorTypeDependency, ErrorTypeCircuitBreaker,
		ErrorTypeExternalService, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeResourceExhausted, ErrorTypeNetwork, ErrorTypeCloud,
		ErrorTypeHTTP, ErrorTypeGRPC, ErrorTypeGraphQL, ErrorTypeWebSocket:
		return true
	default:
		return false
	}
}

// DefaultStatusCode retorna o código de status HTTP padrão para o tipo de erro.
func (et ErrorType) DefaultStatusCode() int {
	switch et {
	case ErrorTypeRepository, ErrorTypeDatabase, ErrorTypeMigration:
		return http.StatusInternalServerError
	case ErrorTypeCache:
		return http.StatusServiceUnavailable
	case ErrorTypeSerialization:
		return http.StatusUnprocessableEntity
	case ErrorTypeValidation, ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeUnprocessable:
		return http.StatusUnprocessableEntity
	case ErrorTypeUnsupported:
		return http.StatusUnsupportedMediaType
	case ErrorTypeBusinessRule, ErrorTypeWorkflow:
		return http.StatusUnprocessableEntity
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization, ErrorTypeForbidden, ErrorTypeSecurity:
		return http.StatusForbidden
	case ErrorTypeInternal, ErrorTypeConfiguration:
		return http.StatusInternalServerError
	case ErrorTypeInfrastructure, ErrorTypeCircuitBreaker:
		return http.StatusServiceUnavailable
	case ErrorTypeDependency:
		return http.StatusFailedDependency
	case ErrorTypeExternalService:
		return http.StatusBadGateway
	case ErrorTypeTimeout:
		return http.StatusGatewayTimeout
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeResourceExhausted:
		return http.StatusInsufficientStorage
	case ErrorTypeNetwork:
		return http.StatusServiceUnavailable
	case ErrorTypeCloud:
		return http.StatusBadGateway
	case ErrorTypeHTTP:
		return http.StatusInternalServerError
	case ErrorTypeGRPC:
		return http.StatusInternalServerError
	case ErrorTypeGraphQL:
		return http.StatusBadRequest
	case ErrorTypeWebSocket:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsRetryable indica se o erro é potencialmente recuperável com retry.
func (et ErrorType) IsRetryable() bool {
	switch et {
	case ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeResourceExhausted,
		ErrorTypeCircuitBreaker, ErrorTypeNetwork, ErrorTypeCloud, ErrorTypeExternalService:
		return true
	default:
		return false
	}
}

// IsTemporary indica se o erro é temporário.
func (et ErrorType) IsTemporary() bool {
	return et.IsRetryable()
}

// Category retorna a categoria geral do tipo de erro.
func (et ErrorType) Category() string {
	switch et {
	case ErrorTypeRepository, ErrorTypeDatabase, ErrorTypeCache, ErrorTypeMigration, ErrorTypeSerialization:
		return "data"
	case ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeUnprocessable, ErrorTypeUnsupported:
		return "input"
	case ErrorTypeBusinessRule, ErrorTypeWorkflow, ErrorTypeConflict, ErrorTypeNotFound:
		return "business"
	case ErrorTypeAuthentication, ErrorTypeAuthorization, ErrorTypeSecurity, ErrorTypeForbidden:
		return "security"
	case ErrorTypeInternal, ErrorTypeInfrastructure, ErrorTypeConfiguration, ErrorTypeDependency, ErrorTypeCircuitBreaker:
		return "system"
	case ErrorTypeExternalService, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeResourceExhausted, ErrorTypeNetwork:
		return "communication"
	case ErrorTypeHTTP, ErrorTypeGRPC, ErrorTypeGraphQL, ErrorTypeWebSocket:
		return "protocol"
	default:
		return "unknown"
	}
}

// ErrorSeverity representa o nível de severidade de um erro.
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String retorna a representação string da severidade.
func (es ErrorSeverity) String() string {
	switch es {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// DefaultSeverity retorna a severidade padrão para um tipo de erro.
func (et ErrorType) DefaultSeverity() ErrorSeverity {
	switch et {
	case ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeNotFound, ErrorTypeConflict:
		return SeverityLow
	case ErrorTypeBusinessRule, ErrorTypeUnprocessable, ErrorTypeAuthentication, ErrorTypeAuthorization:
		return SeverityMedium
	case ErrorTypeExternalService, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeCircuitBreaker:
		return SeverityHigh
	case ErrorTypeInternal, ErrorTypeInfrastructure, ErrorTypeDatabase, ErrorTypeSecurity:
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

// CommonErrorCodes define códigos de erro comuns do sistema.
var CommonErrorCodes = map[string]struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Severity   ErrorSeverity
	Retryable  bool
}{
	"E001": {ErrorTypeValidation, "Validation failed", http.StatusBadRequest, SeverityLow, false},
	"E002": {ErrorTypeNotFound, "Resource not found", http.StatusNotFound, SeverityLow, false},
	"E003": {ErrorTypeConflict, "Resource already exists", http.StatusConflict, SeverityLow, false},
	"E004": {ErrorTypeBusinessRule, "Business rule violation", http.StatusUnprocessableEntity, SeverityMedium, false},
	"E005": {ErrorTypeAuthentication, "Authentication failed", http.StatusUnauthorized, SeverityMedium, false},
	"E006": {ErrorTypeAuthorization, "Access denied", http.StatusForbidden, SeverityMedium, false},
	"E007": {ErrorTypeInternal, "Internal server error", http.StatusInternalServerError, SeverityCritical, false},
	"E008": {ErrorTypeExternalService, "External service unavailable", http.StatusBadGateway, SeverityHigh, true},
	"E009": {ErrorTypeTimeout, "Request timeout", http.StatusGatewayTimeout, SeverityHigh, true},
	"E010": {ErrorTypeRateLimit, "Rate limit exceeded", http.StatusTooManyRequests, SeverityHigh, true},
	"E011": {ErrorTypeDatabase, "Database error", http.StatusInternalServerError, SeverityCritical, false},
	"E012": {ErrorTypeConfiguration, "Configuration error", http.StatusInternalServerError, SeverityCritical, false},
	"E013": {ErrorTypeCircuitBreaker, "Circuit breaker open", http.StatusServiceUnavailable, SeverityHigh, true},
	"E014": {ErrorTypeResourceExhausted, "Resource exhausted", http.StatusInsufficientStorage, SeverityHigh, true},
	"E015": {ErrorTypeUnsupported, "Operation not supported", http.StatusUnsupportedMediaType, SeverityLow, false},
}

// GetCommonErrorCode retorna informações sobre um código de erro comum.
func GetCommonErrorCode(code string) (ErrorType, string, int, ErrorSeverity, bool, bool) {
	if info, exists := CommonErrorCodes[code]; exists {
		return info.Type, info.Message, info.StatusCode, info.Severity, info.Retryable, true
	}
	return ErrorTypeInternal, "Unknown error", http.StatusInternalServerError, SeverityMedium, false, false
}

// IsCommonErrorCode verifica se um código é um código de erro comum.
func IsCommonErrorCode(code string) bool {
	_, exists := CommonErrorCodes[code]
	return exists
}
