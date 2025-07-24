package domainerrors

// ErrorType representa o tipo de erro de domínio
type ErrorType string

// StackFrame representa um frame do stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message,omitempty"`
	Time     string `json:"time"`
}

// ErrorDomainInterface define a interface principal para erros de domínio
type ErrorDomainInterface interface {
	Error() string
	Unwrap() error
	Type() ErrorType
	HTTPStatus() int
	StackTrace() string
	WithMetadata(key string, value interface{}) ErrorDomainInterface
}

// ErrorTypeChecker define interface para verificação de tipos
type ErrorTypeChecker interface {
	IsType(err error, errorType ErrorType) bool
}

// HTTPStatusProvider define interface para provedores de status HTTP
type HTTPStatusProvider interface {
	HTTPStatus() int
}

// MetadataProvider define interface para provedores de metadados
type MetadataProvider interface {
	GetMetadata() map[string]interface{}
	SetMetadata(key string, value interface{})
}

// StackTraceProvider define interface para provedores de stack trace
type StackTraceProvider interface {
	StackTrace() string
	GetStackFrames() []StackFrame
}

// ContextualError define interface para erros com contexto
type ContextualError interface {
	WithContext(message string) ErrorDomainInterface
	Wrap(message string, cause error) ErrorDomainInterface
}

// ValidationErrorInterface define interface específica para erros de validação
type ValidationErrorInterface interface {
	ErrorDomainInterface
	WithField(field, message string) ValidationErrorInterface
	GetFields() map[string][]string
}

// NotFoundErrorInterface define interface específica para erros de não encontrado
type NotFoundErrorInterface interface {
	ErrorDomainInterface
	WithResource(resourceType, resourceID string) NotFoundErrorInterface
	GetResourceInfo() (resourceType, resourceID string)
}

// BusinessErrorInterface define interface específica para erros de negócio
type BusinessErrorInterface interface {
	ErrorDomainInterface
	WithRule(rule string) BusinessErrorInterface
	GetRules() []string
}

// DatabaseErrorInterface define interface específica para erros de banco
type DatabaseErrorInterface interface {
	ErrorDomainInterface
	WithOperation(operation, table string) DatabaseErrorInterface
	WithQuery(query string) DatabaseErrorInterface
	GetOperationInfo() (operation, table, query string)
}

// ExternalServiceErrorInterface define interface específica para erros de serviço externo
type ExternalServiceErrorInterface interface {
	ErrorDomainInterface
	WithEndpoint(endpoint string) ExternalServiceErrorInterface
	WithResponse(statusCode int, response string) ExternalServiceErrorInterface
	GetServiceInfo() (service, endpoint string, statusCode int)
}

// TimeoutErrorInterface define interface específica para erros de timeout
type TimeoutErrorInterface interface {
	ErrorDomainInterface
	WithDuration(duration, timeout interface{}) TimeoutErrorInterface
	GetTimeoutInfo() (operation string, duration, timeout interface{})
}

// RateLimitErrorInterface define interface específica para erros de limite de taxa
type RateLimitErrorInterface interface {
	ErrorDomainInterface
	WithRateLimit(limit, remaining int, resetTime, window string) RateLimitErrorInterface
	GetRateLimitInfo() (limit, remaining int, resetTime, window string)
}

// CircuitBreakerErrorInterface define interface específica para erros de circuit breaker
type CircuitBreakerErrorInterface interface {
	ErrorDomainInterface
	WithCircuitState(state string, failures int) CircuitBreakerErrorInterface
	GetCircuitInfo() (circuitName, state string, failures int)
}

// ErrorFactory define interface para factory de erros
type ErrorFactory interface {
	CreateValidationError(message string, cause error) ValidationErrorInterface
	CreateNotFoundError(message string) NotFoundErrorInterface
	CreateBusinessError(code, message string) BusinessErrorInterface
	CreateDatabaseError(message string, cause error) DatabaseErrorInterface
	CreateExternalServiceError(service, message string, cause error) ExternalServiceErrorInterface
	CreateTimeoutError(operation, message string, cause error) TimeoutErrorInterface
	CreateRateLimitError(message string) RateLimitErrorInterface
	CreateCircuitBreakerError(circuitName, message string) CircuitBreakerErrorInterface
}

// ErrorRegistry define interface para registro de erros
type ErrorRegistry interface {
	Register(code, description string, httpStatus int)
	Get(code string) (ErrorCodeInfo, bool)
	WrapWithCode(code string, err error) ErrorDomainInterface
}

// ErrorCodeInfo representa informações sobre um código de erro
type ErrorCodeInfo struct {
	Code        string
	Description string
	HTTPStatus  int
}

// ErrorAnalyzer define interface para análise de erros
type ErrorAnalyzer interface {
	IsTemporary(err error) bool
	IsRetryable(err error) bool
	GetErrorChain(err error) []error
	GetRootCause(err error) error
}

// ErrorHandler define interface para manipulação de erros
type ErrorHandler interface {
	Handle(err error) error
	CanHandle(err error) bool
	GetPriority() int
}

// ErrorMiddleware define interface para middleware de erro
type ErrorMiddleware interface {
	ProcessError(err error, next func(error) error) error
}

// ErrorLogger define interface para logging de erros
type ErrorLogger interface {
	LogError(err error, context map[string]interface{})
	LogErrorWithLevel(err error, level string, context map[string]interface{})
}

// ErrorSerializer define interface para serialização de erros
type ErrorSerializer interface {
	SerializeError(err error) ([]byte, error)
	DeserializeError(data []byte) (error, error)
}

// ErrorTransformer define interface para transformação de erros
type ErrorTransformer interface {
	TransformError(err error) error
	CanTransform(err error) bool
}

// ErrorPolicyProvider define interface para políticas de erro
type ErrorPolicyProvider interface {
	GetRetryPolicy(err error) RetryPolicy
	GetCircuitBreakerPolicy(err error) CircuitBreakerPolicy
	GetTimeoutPolicy(err error) TimeoutPolicy
}

// RetryPolicy define configuração de retry
type RetryPolicy struct {
	MaxRetries int
	Delay      interface{} // time.Duration
	Backoff    string      // "linear", "exponential", "constant"
}

// CircuitBreakerPolicy define configuração de circuit breaker
type CircuitBreakerPolicy struct {
	FailureThreshold int
	RecoveryTimeout  interface{} // time.Duration
	HalfOpenRequests int
}

// TimeoutPolicy define configuração de timeout
type TimeoutPolicy struct {
	RequestTimeout interface{} // time.Duration
	ReadTimeout    interface{} // time.Duration
	WriteTimeout   interface{} // time.Duration
}
