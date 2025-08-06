package domainerrors

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// HookFunc define uma função de hook
type HookFunc func(ctx context.Context, err *DomainError, operation string) error

// MiddlewareFunc define uma função de middleware
type MiddlewareFunc func(ctx context.Context, err *DomainError, next func(*DomainError) *DomainError) *DomainError

// HookRegistry mantém hooks registrados
var hooksRegistry = make(map[string][]HookFunc)

// MiddlewareChain mantém middlewares registrados
var middlewareChain []MiddlewareFunc

// RegisterHook registra um hook para um tipo específico de operação
func RegisterHook(operation string, hook HookFunc) {
	if hooksRegistry[operation] == nil {
		hooksRegistry[operation] = make([]HookFunc, 0)
	}
	hooksRegistry[operation] = append(hooksRegistry[operation], hook)
}

// RegisterMiddleware registra um middleware na cadeia
func RegisterMiddleware(middleware MiddlewareFunc) {
	middlewareChain = append(middlewareChain, middleware)
}

// executeHooks executa todos os hooks para uma operação específica
func executeHooks(ctx context.Context, err *DomainError, operation string) error {
	hooks := hooksRegistry[operation]
	for _, hook := range hooks {
		if err := hook(ctx, err, operation); err != nil {
			return err
		}
	}
	return nil
}

// executeMiddleware executa a cadeia de middlewares
func executeMiddleware(ctx context.Context, err *DomainError) *DomainError {
	if len(middlewareChain) == 0 {
		return err
	}

	index := 0
	var next func(*DomainError) *DomainError
	next = func(e *DomainError) *DomainError {
		if index >= len(middlewareChain) {
			return e
		}
		middleware := middlewareChain[index]
		index++
		return middleware(ctx, e, next)
	}

	return next(err)
}

// ErrorType representa o tipo de erro de domínio
type ErrorType string

// Tipos de erro definidos
const (
	ErrorTypeValidation         ErrorType = "validation"
	ErrorTypeNotFound           ErrorType = "not_found"
	ErrorTypeBusinessRule       ErrorType = "business_rule"
	ErrorTypeDatabase           ErrorType = "database"
	ErrorTypeExternalService    ErrorType = "external_service"
	ErrorTypeinfraestructure    ErrorType = "infraestructure"
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
	context   context.Context        `json:"-"`
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
	// Executa hooks antes de adicionar metadados
	if e.context != nil {
		executeHooks(e.context, e, "before_metadata")
	}

	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value

	// Executa hooks após adicionar metadados
	if e.context != nil {
		executeHooks(e.context, e, "after_metadata")
	}

	return e
}

// WithContext adiciona contexto ao erro (adiciona stack frame)
func (e *DomainError) WithContext(ctx context.Context, message string) *DomainError {
	e.context = ctx
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
	// Executa hooks antes de capturar stack trace
	if e.context != nil {
		executeHooks(e.context, e, "before_stack_trace")
	}

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

	// Executa hooks após capturar stack trace
	if e.context != nil {
		executeHooks(e.context, e, "after_stack_trace")
	}
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

	// Executa middlewares
	ctx := context.Background()
	err.context = ctx
	processedErr := executeMiddleware(ctx, err)

	// Executa hooks após criação do erro
	executeHooks(ctx, processedErr, "after_error")

	return processedErr
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

	// Executa middlewares
	ctx := context.Background()
	err.context = ctx
	processedErr := executeMiddleware(ctx, err)

	// Executa hooks após criação do erro
	executeHooks(ctx, processedErr, "after_error")

	return processedErr
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

	// Executa middlewares
	ctx := context.Background()
	err.context = ctx
	processedErr := executeMiddleware(ctx, err)

	// Executa hooks após criação do erro
	executeHooks(ctx, processedErr, "after_error")

	return processedErr
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
