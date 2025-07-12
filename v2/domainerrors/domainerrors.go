// Package domainerrors implementa um sistema completo de tratamento de erros para aplicações Go.
//
// Este pacote segue os princípios de Clean Architecture, SOLID e oferece:
// - Hierarquia bem definida de tipos de erro
// - Construção fluente através de Builders
// - Registro centralizado de códigos de erro
// - Parsing automático de diferentes tipos de erro
// - Thread safety em operações concorrentes
// - Otimizações de performance com object pooling
// - Compatibilidade com interfaces padrão do Go (error, fmt.Stringer, json.Marshaler)
//
// Exemplo de uso básico:
//
//	import "github.com/fsvxavier/nexs-lib/v2/domainerrors"
//
//	// Criação simples
//	err := New("E001", "Validation failed")
//
//	// Criação com builder
//	err := NewBuilder().
//		WithCode("E002").
//		WithMessage("User not found").
//		WithType(string(types.ErrorTypeNotFound)).
//		WithDetail("user_id", "12345").
//		Build()
//
// Exemplo de erro de validação:
//
//	fields := map[string][]string{
//		"email": {"invalid format", "required"},
//		"age":   {"must be positive"},
//	}
//	err := NewValidationError("Validation failed", fields)
//
// Exemplo com wrapping:
//
//	originalErr := errors.New("database connection failed")
//	err := New("DB001", "Query failed").Wrap("database error", originalErr)
package domainerrors

import (
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// Variáveis globais comuns para compatibilidade
var (
	// ErrNoRows é um erro comum usado para indicar que não há resultados em uma consulta
	ErrNoRows = errors.New("no rows in result set")

	// DefaultErrorCode é o código de erro padrão quando não especificado
	DefaultErrorCode = "E999"

	// DefaultStatusCode é o código de status HTTP padrão quando não especificado
	DefaultStatusCode = 500
)

// New cria um novo DomainError com código e mensagem.
func New(code, message string) interfaces.DomainErrorInterface {
	return NewBuilder().
		WithCode(code).
		WithMessage(message).
		Build()
}

// NewWithCause cria um novo DomainError com código, mensagem e erro subjacente.
func NewWithCause(code, message string, cause error) interfaces.DomainErrorInterface {
	return NewBuilder().
		WithCode(code).
		WithMessage(message).
		WithCause(cause).
		Build()
}

// Wrap cria ou envolve um erro existente.
func Wrap(code, message string, err error) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}

	// Se já é um DomainError, cria um novo erro wrappando o existente
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return NewBuilder().
			WithCode(code).
			WithMessage(message).
			WithCause(domainErr).
			Build()
	}

	// Caso contrário, cria um novo
	return NewWithCause(code, message, err)
}

// Is verifica se err é do tipo target usando errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As encontra o primeiro erro na cadeia que corresponde ao tipo target usando errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Funções de conveniência para tipos específicos de erro

// NewNotFoundError cria um erro de recurso não encontrado.
func NewNotFoundError(entity, id string) interfaces.DomainErrorInterface {
	message := fmt.Sprintf("%s not found", entity)
	if id != "" {
		message = fmt.Sprintf("%s with ID '%s' not found", entity, id)
	}

	return NewBuilder().
		WithCode("E002").
		WithMessage(message).
		WithType(string(types.ErrorTypeNotFound)).
		WithDetail("entity", entity).
		WithDetail("id", id).
		WithTag("not_found").
		Build()
}

// NewUnauthorizedError cria um erro de não autorizado.
func NewUnauthorizedError(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Authentication required"
	}

	return NewBuilder().
		WithCode("E005").
		WithMessage(message).
		WithType(string(types.ErrorTypeAuthentication)).
		WithTag("authentication").
		Build()
}

// NewForbiddenError cria um erro de acesso negado.
func NewForbiddenError(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Access denied"
	}

	return NewBuilder().
		WithCode("E006").
		WithMessage(message).
		WithType(string(types.ErrorTypeAuthorization)).
		WithTag("authorization").
		Build()
}

// NewInternalError cria um erro interno.
func NewInternalError(message string, cause error) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Internal server error"
	}

	builder := NewBuilder().
		WithCode("E007").
		WithMessage(message).
		WithType(string(types.ErrorTypeInternal)).
		WithSeverity(interfaces.Severity(types.SeverityCritical)).
		WithTag("internal")

	if cause != nil {
		builder = builder.WithCause(cause)
	}

	return builder.Build()
}

// NewBadRequestError cria um erro de requisição inválida.
func NewBadRequestError(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Bad request"
	}

	return NewBuilder().
		WithCode("E001").
		WithMessage(message).
		WithType(string(types.ErrorTypeBadRequest)).
		WithTag("bad_request").
		Build()
}

// NewConflictError cria um erro de conflito.
func NewConflictError(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Resource conflict"
	}

	return NewBuilder().
		WithCode("E003").
		WithMessage(message).
		WithType(string(types.ErrorTypeConflict)).
		WithTag("conflict").
		Build()
}

// NewTimeoutError cria um erro de timeout.
func NewTimeoutError(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Operation timeout"
	}

	return NewBuilder().
		WithCode("E009").
		WithMessage(message).
		WithType(string(types.ErrorTypeTimeout)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithTag("timeout").
		Build()
}

// NewCircuitBreakerError cria um erro de circuit breaker.
func NewCircuitBreakerError(service string) interfaces.DomainErrorInterface {
	message := "Circuit breaker is open"
	if service != "" {
		message = fmt.Sprintf("Circuit breaker is open for service: %s", service)
	}

	return NewBuilder().
		WithCode("E013").
		WithMessage(message).
		WithType(string(types.ErrorTypeCircuitBreaker)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("service", service).
		WithTag("circuit_breaker").
		Build()
}

// Funções utilitárias para análise de erros

// IsRetryable verifica se um erro permite retry.
func IsRetryable(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.IsRetryable()
	}

	// Fallback para análise de string
	errStr := err.Error()
	return containsAny(errStr, []string{
		"timeout", "connection refused", "temporary failure",
		"rate limit", "circuit breaker", "service unavailable",
	})
}

// IsTemporary verifica se um erro é temporário.
func IsTemporary(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.IsTemporary()
	}

	// Verifica interface net.Error
	if netErr, ok := err.(interface{ Temporary() bool }); ok {
		return netErr.Temporary()
	}

	// Fallback para análise de string
	return IsRetryable(err)
}

// GetErrorType retorna o tipo de um erro.
func GetErrorType(err error) string {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.Type()
	}

	return string(types.ErrorTypeInternal)
}

// GetErrorCode retorna o código de um erro.
func GetErrorCode(err error) string {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.Code()
	}

	return DefaultErrorCode
}

// GetStatusCode retorna o código de status HTTP de um erro.
func GetStatusCode(err error) int {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.StatusCode()
	}

	return DefaultStatusCode
}

// GetRootCause retorna o erro original de uma cadeia.
func GetRootCause(err error) error {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		if root := domainErr.RootCause(); root != nil {
			return root
		}
	}

	// Fallback usando errors.Unwrap
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			break
		}
		err = unwrapped
	}

	return err
}

// FormatError formata um erro para exibição.
func FormatError(err error) string {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.DetailedString()
	}

	return err.Error()
}

// ToJSON converte um erro para JSON.
func ToJSON(err error) ([]byte, error) {
	if err == nil {
		return nil, errors.New("cannot convert nil error to JSON")
	}

	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.JSON()
	}

	// Fallback para estrutura simples
	simple := map[string]interface{}{
		"message": err.Error(),
		"type":    "unknown",
	}

	return fmt.Appendf(nil, "%+v", simple), nil
}

// containsAny verifica se uma string contém alguma das substrings.
func containsAny(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}

// Aliases para compatibilidade com versão anterior

// ErrorType é um alias para types.ErrorType.
type ErrorType = types.ErrorType

// Constantes de tipos de erro para compatibilidade
const (
	ErrorTypeRepository        = types.ErrorTypeRepository
	ErrorTypeValidation        = types.ErrorTypeValidation
	ErrorTypeBusinessRule      = types.ErrorTypeBusinessRule
	ErrorTypeNotFound          = types.ErrorTypeNotFound
	ErrorTypeExternalService   = types.ErrorTypeExternalService
	ErrorTypeAuthentication    = types.ErrorTypeAuthentication
	ErrorTypeAuthorization     = types.ErrorTypeAuthorization
	ErrorTypeBadRequest        = types.ErrorTypeBadRequest
	ErrorTypeUnprocessable     = types.ErrorTypeUnprocessable
	ErrorTypeUnsupported       = types.ErrorTypeUnsupported
	ErrorTypeTimeout           = types.ErrorTypeTimeout
	ErrorTypeInternal          = types.ErrorTypeInternal
	ErrorTypeInfrastructure    = types.ErrorTypeInfrastructure
	ErrorTypeConflict          = types.ErrorTypeConflict
	ErrorTypeRateLimit         = types.ErrorTypeRateLimit
	ErrorTypeCircuitBreaker    = types.ErrorTypeCircuitBreaker
	ErrorTypeConfiguration     = types.ErrorTypeConfiguration
	ErrorTypeSecurity          = types.ErrorTypeSecurity
	ErrorTypeResourceExhausted = types.ErrorTypeResourceExhausted
	ErrorTypeDependency        = types.ErrorTypeDependency
	ErrorTypeSerialization     = types.ErrorTypeSerialization
	ErrorTypeCache             = types.ErrorTypeCache
	ErrorTypeWorkflow          = types.ErrorTypeWorkflow
	ErrorTypeMigration         = types.ErrorTypeMigration
)
