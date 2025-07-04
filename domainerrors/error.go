package domainerrors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

var (
	// ErrNoRows é um erro comum usado para indicar que não há resultados em uma consulta
	ErrNoRows = errors.New("no rows in result set")

	// DefaultErrorCode é o código de erro padrão quando não especificado
	DefaultErrorCode = "E999"

	// DefaultStatusCode é o código de status HTTP padrão quando não especificado
	DefaultStatusCode = http.StatusInternalServerError
)

// ErrorType representa o tipo de erro, usado para categorização e tratamento específico
type ErrorType string

// Constantes para os tipos de erro mais comuns
const (
	ErrorTypeRepository        ErrorType = "repository"         // Erros relacionados a acesso a dados
	ErrorTypeValidation        ErrorType = "validation"         // Erros de validação
	ErrorTypeBusinessRule      ErrorType = "business"           // Erros de regras de negócio
	ErrorTypeNotFound          ErrorType = "not_found"          // Erros de recurso não encontrado
	ErrorTypeExternalService   ErrorType = "external_service"   // Erros de serviços externos
	ErrorTypeAuthentication    ErrorType = "authentication"     // Erros de autenticação
	ErrorTypeAuthorization     ErrorType = "authorization"      // Erros de autorização
	ErrorTypeBadRequest        ErrorType = "bad_request"        // Erros de requisição inválida
	ErrorTypeUnprocessable     ErrorType = "unprocessable"      // Erros de entidade não processável
	ErrorTypeUnsupported       ErrorType = "unsupported"        // Operação não suportada
	ErrorTypeTimeout           ErrorType = "timeout"            // Erros de timeout
	ErrorTypeInternal          ErrorType = "internal"           // Erros internos gerais
	ErrorTypeInfrastructure    ErrorType = "infrastructure"     // Erros de infraestrutura
	ErrorTypeConflict          ErrorType = "conflict"           // Erros de conflito/duplicação
	ErrorTypeRateLimit         ErrorType = "rate_limit"         // Erros de limite de taxa
	ErrorTypeCircuitBreaker    ErrorType = "circuit_breaker"    // Erros de circuit breaker
	ErrorTypeConfiguration     ErrorType = "configuration"      // Erros de configuração
	ErrorTypeSecurity          ErrorType = "security"           // Erros de segurança
	ErrorTypeResourceExhausted ErrorType = "resource_exhausted" // Erros de recursos esgotados
	ErrorTypeDependency        ErrorType = "dependency"         // Erros de dependências
	ErrorTypeSerialization     ErrorType = "serialization"      // Erros de serialização
	ErrorTypeCache             ErrorType = "cache"              // Erros de cache
	ErrorTypeWorkflow          ErrorType = "workflow"           // Erros de workflow
	ErrorTypeMigration         ErrorType = "migration"          // Erros de migração
)

// DomainError é a estrutura base para todos os erros de domínio
type DomainError struct {
	// Código do erro
	Code string `json:"code,omitempty"`

	// Mensagem descritiva
	Message string `json:"message,omitempty"`

	// Erro original
	Err error `json:"-"`

	// Tipo do erro
	Type ErrorType `json:"type,omitempty"`

	// Detalhes adicionais
	Details map[string]interface{} `json:"details,omitempty"`

	// Metadados para uso interno
	Metadata map[string]interface{} `json:"-"`

	// Stack de erros
	stack []stackTrace

	// Entidade relacionada
	EntityName string `json:"entity,omitempty"`
}

// stackTrace armazena informações sobre o stack de chamadas
type stackTrace struct {
	Function string
	File     string
	Line     int
	Message  string
	Error    error
}

// New cria um novo DomainError com código e mensagem
func New(code string, message string) *DomainError {
	return &DomainError{
		Code:     code,
		Message:  message,
		Details:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
		stack:    make([]stackTrace, 0),
	}
}

// NewWithError cria um novo DomainError com código, mensagem e um erro subjacente
func NewWithError(code string, message string, err error) *DomainError {
	de := New(code, message)
	if err != nil {
		de.Err = err
		// Captura o stack trace atual
		de.captureStackTrace(message, err)
	}
	return de
}

// Wrap adiciona um erro ao DomainError e retorna o mesmo DomainError
func (e *DomainError) Wrap(message string, err error) *DomainError {
	if err != nil {
		e.captureStackTrace(message, err)

		// Se o erro for um DomainError, copie os detalhes
		if de, ok := err.(*DomainError); ok {
			for k, v := range de.Details {
				if _, exists := e.Details[k]; !exists {
					e.Details[k] = v
				}
			}
		}

		// O último erro adicionado se torna o principal para o Unwrap
		e.Err = err
	}
	return e
}

// WithType define o tipo do erro e retorna o mesmo DomainError
func (e *DomainError) WithType(errType ErrorType) *DomainError {
	e.Type = errType
	return e
}

// WithDetails adiciona detalhes ao erro e retorna o mesmo DomainError
func (e *DomainError) WithDetails(details map[string]interface{}) *DomainError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithDetail adiciona um único detalhe ao erro e retorna o mesmo DomainError
func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	e.Details[key] = value
	return e
}

// WithMetadata adiciona metadados ao erro e retorna o mesmo DomainError
func (e *DomainError) WithMetadata(metadata map[string]interface{}) *DomainError {
	for k, v := range metadata {
		e.Metadata[k] = v
	}
	return e
}

// WithEntity define a entidade relacionada ao erro
func (e *DomainError) WithEntity(entity interface{}) *DomainError {
	e.EntityName = reflect.TypeOf(entity).Name()
	return e
}

// Error implementa a interface error
func (e *DomainError) Error() string {
	var b strings.Builder

	if e.Code != "" {
		b.WriteString(fmt.Sprintf("[%s] ", e.Code))
	}

	b.WriteString(e.Message)

	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}

	return b.String()
}

// Unwrap implementa a interface errors.Wrapper
func (e *DomainError) Unwrap() error {
	return e.Err
}

// StatusCode retorna o código de status HTTP associado ao tipo de erro
func (e *DomainError) StatusCode() int {
	// Mapeia tipos de erro para códigos HTTP
	switch e.Type {
	case ErrorTypeRepository:
		return http.StatusInternalServerError
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeBusinessRule:
		return http.StatusUnprocessableEntity
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeExternalService:
		return http.StatusBadGateway
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeUnprocessable:
		return http.StatusUnprocessableEntity
	case ErrorTypeUnsupported:
		return http.StatusUnsupportedMediaType
	case ErrorTypeTimeout:
		return http.StatusGatewayTimeout
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeInfrastructure:
		return http.StatusServiceUnavailable
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeCircuitBreaker:
		return http.StatusServiceUnavailable
	case ErrorTypeConfiguration:
		return http.StatusInternalServerError
	case ErrorTypeSecurity:
		return http.StatusForbidden
	case ErrorTypeResourceExhausted:
		return http.StatusInsufficientStorage
	case ErrorTypeDependency:
		return http.StatusFailedDependency
	case ErrorTypeSerialization:
		return http.StatusUnprocessableEntity
	case ErrorTypeCache:
		return http.StatusInternalServerError
	case ErrorTypeWorkflow:
		return http.StatusUnprocessableEntity
	case ErrorTypeMigration:
		return http.StatusInternalServerError
	default:
		return DefaultStatusCode
	}
}

// StackTrace retorna todo o stack de erros
func (e *DomainError) StackTrace() []stackTrace {
	return e.stack
}

// FormatStackTrace retorna uma string formatada do stack de erros
func (e *DomainError) FormatStackTrace() string {
	if len(e.stack) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Error Stack Trace:\n")

	for i, st := range e.stack {
		b.WriteString(fmt.Sprintf("%d: [%s] in %s (%s:%d)\n",
			i+1, st.Message, st.Function, st.File, st.Line))
		if st.Error != nil {
			b.WriteString(fmt.Sprintf("   Error: %s\n", st.Error.Error()))
		}
	}

	return b.String()
}

// captureStackTrace captura informações sobre a pilha de chamadas atual
func (e *DomainError) captureStackTrace(message string, err error) {
	pc, file, line, ok := runtime.Caller(2) // Pega o chamador do chamador
	if !ok {
		file = "unknown"
		line = 0
	}

	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
	} else {
		funcName = "unknown"
	}

	st := stackTrace{
		Function: funcName,
		File:     file,
		Line:     line,
		Message:  message,
		Error:    err,
	}

	e.stack = append(e.stack, st)
}

// WrapError é uma função helper para criar ou atualizar um DomainError
func WrapError(code string, message string, err error) error {
	if err == nil {
		return nil
	}

	// Se já é um DomainError, apenas adicione ao stack
	if de, ok := err.(*DomainError); ok {
		return de.Wrap(message, err)
	}

	// Caso contrário, crie um novo
	return NewWithError(code, message, err)
}

// Is verifica se err é do tipo target
// Implementa errors.Is
func (e *DomainError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Verifica se o alvo é do tipo DomainError e compara os códigos
	if t, ok := target.(*DomainError); ok {
		return e.Code == t.Code
	}

	// Compara com o erro interno
	return errors.Is(e.Err, target)
}
