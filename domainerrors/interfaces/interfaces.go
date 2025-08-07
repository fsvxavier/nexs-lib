package interfaces

import (
	"context"
	"time"
)

// ErrorType representa o tipo de erro de domínio
type ErrorType string

// Tipos de erro disponíveis
const (
	ValidationError           ErrorType = "validation_error"
	NotFoundError             ErrorType = "not_found_error"
	BusinessError             ErrorType = "business_error"
	DatabaseError             ErrorType = "database_error"
	ExternalServiceError      ErrorType = "external_service_error"
	InfrastructureError       ErrorType = "infrastructure_error"
	DependencyError           ErrorType = "dependency_error"
	AuthenticationError       ErrorType = "authentication_error"
	AuthorizationError        ErrorType = "authorization_error"
	SecurityError             ErrorType = "security_error"
	TimeoutError              ErrorType = "timeout_error"
	RateLimitError            ErrorType = "rate_limit_error"
	ResourceExhaustedError    ErrorType = "resource_exhausted_error"
	CircuitBreakerError       ErrorType = "circuit_breaker_error"
	SerializationError        ErrorType = "serialization_error"
	CacheError                ErrorType = "cache_error"
	MigrationError            ErrorType = "migration_error"
	ConfigurationError        ErrorType = "configuration_error"
	UnsupportedOperationError ErrorType = "unsupported_operation_error"
	BadRequestError           ErrorType = "bad_request_error"
	ConflictError             ErrorType = "conflict_error"
	InvalidSchemaError        ErrorType = "invalid_schema_error"
	UnsupportedMediaTypeError ErrorType = "unsupported_media_type_error"
	ServerError               ErrorType = "server_error"
	UnprocessableEntityError  ErrorType = "unprocessable_entity_error"
	ServiceUnavailableError   ErrorType = "service_unavailable_error"
	WorkflowError             ErrorType = "workflow_error"
)

// StackFrame representa um frame do stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message,omitempty"`
	Time     string `json:"time"`
}

// DomainErrorInterface define a interface principal para erros de domínio
type DomainErrorInterface interface {
	// Error retorna a mensagem de erro (implementa error interface)
	Error() string

	// Unwrap retorna a causa raiz do erro
	Unwrap() error

	// Type retorna o tipo do erro
	Type() ErrorType

	// Metadata retorna os metadados do erro
	Metadata() map[string]interface{}

	// HTTPStatus retorna o código HTTP correspondente
	HTTPStatus() int

	// StackTrace retorna o stack trace formatado
	StackTrace() string

	// WithContext adiciona contexto ao erro
	WithContext(ctx context.Context) DomainErrorInterface

	// Wrap encapsula outro erro mantendo o contexto
	Wrap(err error) DomainErrorInterface

	// WithMetadata adiciona metadados ao erro
	WithMetadata(key string, value interface{}) DomainErrorInterface

	// Code retorna o código único do erro
	Code() string

	// Timestamp retorna o momento da criação do erro
	Timestamp() time.Time

	// ToJSON serializa o erro para JSON
	ToJSON() ([]byte, error)
}

// ErrorTypeChecker define interface para verificação de tipos
type ErrorTypeChecker interface {
	IsType(err error, errorType ErrorType) bool
}

// HTTPStatusProvider define interface para provedores de status HTTP
type HTTPStatusProvider interface {
	HTTPStatus() int
}

// ErrorFactory define interface para criação de erros
type ErrorFactory interface {
	New(errorType ErrorType, code, message string) DomainErrorInterface
	NewWithMetadata(errorType ErrorType, code, message string, metadata map[string]interface{}) DomainErrorInterface
	Wrap(err error, errorType ErrorType, code, message string) DomainErrorInterface
}

// HookManager define interface para gerenciamento de hooks
type HookManager interface {
	RegisterStartHook(hook StartHookFunc)
	RegisterStopHook(hook StopHookFunc)
	RegisterErrorHook(hook ErrorHookFunc)
	RegisterI18nHook(hook I18nHookFunc)
	ExecuteStartHooks(ctx context.Context) error
	ExecuteStopHooks(ctx context.Context) error
	ExecuteErrorHooks(ctx context.Context, err DomainErrorInterface) error
	ExecuteI18nHooks(ctx context.Context, err DomainErrorInterface, locale string) error
}

// MiddlewareManager define interface para gerenciamento de middlewares
type MiddlewareManager interface {
	RegisterMiddleware(middleware MiddlewareFunc)
	RegisterI18nMiddleware(middleware I18nMiddlewareFunc)
	ExecuteMiddlewares(ctx context.Context, err DomainErrorInterface) DomainErrorInterface
	ExecuteI18nMiddlewares(ctx context.Context, err DomainErrorInterface, locale string) DomainErrorInterface
}

// Hook function types
type StartHookFunc func(ctx context.Context) error
type StopHookFunc func(ctx context.Context) error
type ErrorHookFunc func(ctx context.Context, err DomainErrorInterface) error
type I18nHookFunc func(ctx context.Context, err DomainErrorInterface, locale string) error

// Middleware function types
type MiddlewareFunc func(ctx context.Context, err DomainErrorInterface, next func(DomainErrorInterface) DomainErrorInterface) DomainErrorInterface
type I18nMiddlewareFunc func(ctx context.Context, err DomainErrorInterface, locale string, next func(DomainErrorInterface) DomainErrorInterface) DomainErrorInterface

// Observer define interface para observadores de erros (Observer Pattern)
type Observer interface {
	OnError(ctx context.Context, err DomainErrorInterface) error
}

// Subject define interface para sujeitos observáveis
type Subject interface {
	RegisterObserver(observer Observer)
	UnregisterObserver(observer Observer)
	NotifyObservers(ctx context.Context, err DomainErrorInterface) error
}

// StackTraceCapture define interface para captura de stack trace
type StackTraceCapture interface {
	CaptureStackTrace(skip int) []StackFrame
	FormatStackTrace(frames []StackFrame) string
}

// ErrorAggregator define interface para agregação de erros
type ErrorAggregator interface {
	Add(err error)
	HasErrors() bool
	Errors() []error
	Error() string
	Count() int
}
