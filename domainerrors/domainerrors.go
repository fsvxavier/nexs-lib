package domainerrors

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorType representa o tipo de erro de domínio
type ErrorType string

// Tipos de erro definidos
const (
	ErrorTypeValidation         ErrorType = "validation"
	ErrorTypeNotFound           ErrorType = "not_found"
	ErrorTypeBusinessRule       ErrorType = "business_rule"
	ErrorTypeDatabase           ErrorType = "database"
	ErrorTypeExternalService    ErrorType = "external_service"
	ErrorTypeinfraestructure     ErrorType = "infraestructure"
	ErrorTypeDependency         ErrorType = "dependency"
	ErrorTypeAuthentication     ErrorType = "authentication"
	ErrorTypeAuthorization      ErrorType = "authorization"
	ErrorTypeSecurity           ErrorType = "security"
	ErrorTypeTimeout            ErrorType = "timeout"
	ErrorTypeRateLimit          ErrorType = "rate_limit"
	ErrorTypeResourceExhausted  ErrorType = "resource_exhausted"
	ErrorTypeCircuitBreaker     ErrorType = "circuit_breaker"
	ErrorTypeSerialization      ErrorType = "serialization"
	ErrorTypeCache              ErrorType = "cache"
	ErrorTypeMigration          ErrorType = "migration"
	ErrorTypeConfiguration      ErrorType = "configuration"
	ErrorTypeUnsupported        ErrorType = "unsupported"
	ErrorTypeBadRequest         ErrorType = "bad_request"
	ErrorTypeConflict           ErrorType = "conflict"
	ErrorTypeInvalidSchema      ErrorType = "invalid_schema"
	ErrorTypeUnsupportedMedia   ErrorType = "unsupported_media"
	ErrorTypeServer             ErrorType = "server"
	ErrorTypeUnprocessable      ErrorType = "unprocessable"
	ErrorTypeServiceUnavailable ErrorType = "service_unavailable"
	ErrorTypeWorkflow           ErrorType = "workflow"
)

// StackFrame representa um frame do stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message,omitempty"`
	Time     string `json:"time"`
}

// DomainError é a estrutura principal para erros de domínio
type DomainError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Type      ErrorType              `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Cause     error                  `json:"-"`
	Stack     []StackFrame           `json:"stack,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Error implementa a interface error
func (e *DomainError) Error() string {
	var builder strings.Builder

	if e.Code != "" {
		builder.WriteString(fmt.Sprintf("[%s] ", e.Code))
	}

	builder.WriteString(e.Message)

	if e.Cause != nil {
		builder.WriteString(": ")
		builder.WriteString(e.Cause.Error())
	}

	return builder.String()
}

// Unwrap retorna o erro original (causa raiz)
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// HTTPStatus retorna o código HTTP correspondente ao tipo de erro
func (e *DomainError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeInvalidSchema:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization, ErrorTypeSecurity:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	case ErrorTypeBusinessRule, ErrorTypeUnprocessable:
		return http.StatusUnprocessableEntity
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeUnsupported:
		return http.StatusNotImplemented
	case ErrorTypeServiceUnavailable, ErrorTypeCircuitBreaker:
		return http.StatusServiceUnavailable
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeResourceExhausted:
		return http.StatusInsufficientStorage
	case ErrorTypeDependency:
		return http.StatusFailedDependency
	default:
		return http.StatusInternalServerError
	}
}

// StackTrace retorna o stack trace formatado
func (e *DomainError) StackTrace() string {
	if len(e.Stack) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Stack trace:\n")

	for i, frame := range e.Stack {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.Function))
		builder.WriteString(fmt.Sprintf("     %s:%d\n", frame.File, frame.Line))
		if frame.Message != "" {
			builder.WriteString(fmt.Sprintf("     Message: %s\n", frame.Message))
		}
		builder.WriteString(fmt.Sprintf("     Time: %s\n", frame.Time))
	}

	return builder.String()
}

// WithMetadata adiciona metadados ao erro
func (e *DomainError) WithMetadata(key string, value interface{}) *DomainError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithContext adiciona contexto ao erro (adiciona stack frame)
func (e *DomainError) WithContext(ctx context.Context, message string) *DomainError {
	e.captureStackFrame(message)
	return e
}

// Wrap empilha um erro com contexto
func (e *DomainError) Wrap(message string, cause error) *DomainError {
	e.Cause = cause
	e.captureStackFrame(message)
	return e
}

// captureStackFrame captura informações do stack trace atual
func (e *DomainError) captureStackFrame(message string) {
	if e.Stack == nil {
		e.Stack = make([]StackFrame, 0)
	}

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return
	}

	fn := runtime.FuncForPC(pc)
	fnName := "unknown"
	if fn != nil {
		fnName = fn.Name()
	}

	frame := StackFrame{
		Function: fnName,
		File:     file,
		Line:     line,
		Message:  message,
		Time:     time.Now().Format(time.RFC3339),
	}

	e.Stack = append(e.Stack, frame)
}

// New cria um novo erro de domínio
func New(code, message string) *DomainError {
	err := &DomainError{
		Code:      code,
		Message:   message,
		Type:      ErrorTypeServer,
		Timestamp: time.Now(),
	}
	err.captureStackFrame("error created")
	return err
}

// NewWithCause cria um erro de domínio com causa
func NewWithCause(code, message string, cause error) *DomainError {
	err := &DomainError{
		Code:      code,
		Message:   message,
		Type:      ErrorTypeServer,
		Cause:     cause,
		Timestamp: time.Now(),
	}
	err.captureStackFrame("error created with cause")
	return err
}

// NewWithType cria um erro de domínio com tipo específico
func NewWithType(code, message string, errorType ErrorType) *DomainError {
	err := &DomainError{
		Code:      code,
		Message:   message,
		Type:      errorType,
		Timestamp: time.Now(),
	}
	err.captureStackFrame("typed error created")
	return err
}

// IsType verifica se o erro é de um tipo específico
func IsType(err error, errorType ErrorType) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Type == errorType
	}
	return false
}

// MapHTTPStatus mapeia um erro para um código HTTP
func MapHTTPStatus(err error) int {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}
