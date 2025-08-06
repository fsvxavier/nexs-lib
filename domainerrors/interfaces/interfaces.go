package domainerrors

import (
	"context"
	"time"
)

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

// DomainError representa a estrutura principal para erros de domínio
type DomainError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Type      ErrorType              `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Cause     error                  `json:"-"`
	Stack     []StackFrame           `json:"stack,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
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

// ===============================
// HOOKS INTERFACES
// ===============================

// HookType representa o tipo de hook
type HookType string

const (
	// HookTypeBeforeError é executado antes da criação do erro
	HookTypeBeforeError HookType = "before_error"
	// HookTypeAfterError é executado após a criação do erro
	HookTypeAfterError HookType = "after_error"
	// HookTypeBeforeMetadata é executado antes de adicionar metadados
	HookTypeBeforeMetadata HookType = "before_metadata"
	// HookTypeAfterMetadata é executado após adicionar metadados
	HookTypeAfterMetadata HookType = "after_metadata"
	// HookTypeBeforeStackTrace é executado antes de capturar stack trace
	HookTypeBeforeStackTrace HookType = "before_stack_trace"
	// HookTypeAfterStackTrace é executado após capturar stack trace
	HookTypeAfterStackTrace HookType = "after_stack_trace"
	// HookTypeBeforeHTTPStatus é executado antes de calcular status HTTP
	HookTypeBeforeHTTPStatus HookType = "before_http_status"
	// HookTypeAfterHTTPStatus é executado após calcular status HTTP
	HookTypeAfterHTTPStatus HookType = "after_http_status"
)

// HookContext contém informações contextuais para hooks
type HookContext struct {
	Context   context.Context        `json:"-"`
	Error     *DomainError           `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Operation string                 `json:"operation"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// Hook define a interface para hooks
type Hook interface {
	// Name retorna o nome do hook
	Name() string
	// Execute executa o hook com o contexto fornecido
	Execute(ctx *HookContext) error
	// Type retorna o tipo do hook
	Type() HookType
	// Priority retorna a prioridade de execução (0 = maior prioridade)
	Priority() int
	// Enabled indica se o hook está habilitado
	Enabled() bool
}

// HookRegistry mantém registro de hooks
type HookRegistry interface {
	// Register registra um hook
	Register(hook Hook) error
	// Unregister remove um hook pelo nome
	Unregister(name string) error
	// GetHooks retorna hooks por tipo ordenados por prioridade
	GetHooks(hookType HookType) []Hook
	// ExecuteHooks executa todos os hooks de um tipo
	ExecuteHooks(ctx *HookContext, hookType HookType) error
	// Clear remove todos os hooks
	Clear()
	// Count retorna o número de hooks registrados
	Count() int
	// ListAll retorna todos os hooks registrados
	ListAll() map[HookType][]Hook
}

// ErrorHook define hooks específicos para erros
type ErrorHook interface {
	Hook
	// ShouldExecute verifica se o hook deve ser executado para este erro
	ShouldExecute(err *DomainError) bool
}

// MetadataHook define hooks específicos para metadados
type MetadataHook interface {
	Hook
	// OnMetadataAdd é chamado quando metadados são adicionados
	OnMetadataAdd(key string, value interface{}, err *DomainError) error
	// OnMetadataRemove é chamado quando metadados são removidos
	OnMetadataRemove(key string, err *DomainError) error
}

// StackTraceHook define hooks específicos para stack trace
type StackTraceHook interface {
	Hook
	// OnStackCapture é chamado quando um frame do stack é capturado
	OnStackCapture(frame *StackFrame, err *DomainError) error
}

// HTTPStatusHook define hooks específicos para status HTTP
type HTTPStatusHook interface {
	Hook
	// OnStatusCalculation é chamado quando o status HTTP é calculado
	OnStatusCalculation(status int, err *DomainError) (int, error)
}

// AsyncHook define interface para hooks assíncronos
type AsyncHook interface {
	Hook
	// ExecuteAsync executa o hook de forma assíncrona
	ExecuteAsync(ctx *HookContext) <-chan error
	// Timeout retorna o timeout para execução assíncrona
	Timeout() time.Duration
}

// ConditionalHook define interface para hooks condicionais
type ConditionalHook interface {
	Hook
	// Condition verifica se o hook deve ser executado
	Condition(ctx *HookContext) bool
}

// HookChain define uma cadeia de hooks
type HookChain interface {
	// Add adiciona um hook à cadeia
	Add(hook Hook) HookChain
	// Execute executa todos os hooks na cadeia
	Execute(ctx *HookContext) error
	// Size retorna o número de hooks na cadeia
	Size() int
	// Clear limpa a cadeia
	Clear()
}

// ===============================
// MIDDLEWARE INTERFACES
// ===============================

// MiddlewareType representa o tipo de middleware
type MiddlewareType string

const (
	// MiddlewareTypeError middleware para processamento de erros
	MiddlewareTypeError MiddlewareType = "error"
	// MiddlewareTypeLogging middleware para logging
	MiddlewareTypeLogging MiddlewareType = "logging"
	// MiddlewareTypeMetrics middleware para métricas
	MiddlewareTypeMetrics MiddlewareType = "metrics"
	// MiddlewareTypeValidation middleware para validação
	MiddlewareTypeValidation MiddlewareType = "validation"
	// MiddlewareTypeTransformation middleware para transformação
	MiddlewareTypeTransformation MiddlewareType = "transformation"
	// MiddlewareTypeEnrichment middleware para enriquecimento de dados
	MiddlewareTypeEnrichment MiddlewareType = "enrichment"
	// MiddlewareTypeFiltering middleware para filtragem
	MiddlewareTypeFiltering MiddlewareType = "filtering"
	// MiddlewareTypeSecurity middleware para segurança
	MiddlewareTypeSecurity MiddlewareType = "security"
	// MiddlewareTypeRetry middleware para retry
	MiddlewareTypeRetry MiddlewareType = "retry"
	// MiddlewareTypeCircuitBreaker middleware para circuit breaker
	MiddlewareTypeCircuitBreaker MiddlewareType = "circuit_breaker"
)

// MiddlewareContext contém informações contextuais para middlewares
type MiddlewareContext struct {
	Context     context.Context        `json:"-"`
	Error       *DomainError           `json:"error,omitempty"`
	Request     interface{}            `json:"request,omitempty"`
	Response    interface{}            `json:"response,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Operation   string                 `json:"operation"`
	Timestamp   time.Time              `json:"timestamp"`
	TraceID     string                 `json:"trace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RequestPath string                 `json:"request_path,omitempty"`
}

// NextFunction define a função para chamar o próximo middleware
type NextFunction func(ctx *MiddlewareContext) error

// Middleware define a interface principal para middlewares
type Middleware interface {
	// Name retorna o nome do middleware
	Name() string
	// Handle processa o middleware
	Handle(ctx *MiddlewareContext, next NextFunction) error
	// Type retorna o tipo do middleware
	Type() MiddlewareType
	// Priority retorna a prioridade de execução (0 = maior prioridade)
	Priority() int
	// Enabled indica se o middleware está habilitado
	Enabled() bool
}

// MiddlewareChain gerencia uma cadeia de middlewares
type MiddlewareChain interface {
	// Use adiciona um middleware à cadeia
	Use(middleware Middleware) MiddlewareChain
	// Execute executa toda a cadeia de middlewares
	Execute(ctx *MiddlewareContext) error
	// Size retorna o número de middlewares na cadeia
	Size() int
	// Clear limpa a cadeia
	Clear()
	// Remove remove um middleware pelo nome
	Remove(name string) bool
	// GetMiddlewares retorna todos os middlewares ordenados por prioridade
	GetMiddlewares() []Middleware
}

// ErrorMiddleware define middleware específico para erros
type ErrorMiddleware interface {
	Middleware
	// ProcessError processa o erro
	ProcessError(ctx *MiddlewareContext, err *DomainError) (*DomainError, error)
	// ShouldProcess verifica se deve processar este erro
	ShouldProcess(err *DomainError) bool
}

// LoggingMiddleware define middleware para logging
type LoggingMiddleware interface {
	Middleware
	// LogError registra o erro
	LogError(ctx *MiddlewareContext, err *DomainError) error
	// LogLevel retorna o nível de log para este tipo de erro
	LogLevel(err *DomainError) string
}

// MetricsMiddleware define middleware para métricas
type MetricsMiddleware interface {
	Middleware
	// RecordError registra métricas do erro
	RecordError(ctx *MiddlewareContext, err *DomainError) error
	// IncrementCounter incrementa contador de erros
	IncrementCounter(errorType ErrorType) error
}

// ValidationMiddleware define middleware para validação
type ValidationMiddleware interface {
	Middleware
	// ValidateError valida o erro
	ValidateError(ctx *MiddlewareContext, err *DomainError) error
	// GetValidationRules retorna regras de validação
	GetValidationRules() map[ErrorType][]string
}

// TransformationMiddleware define middleware para transformação
type TransformationMiddleware interface {
	Middleware
	// Transform transforma o erro
	Transform(ctx *MiddlewareContext, err *DomainError) (*DomainError, error)
	// CanTransform verifica se pode transformar este erro
	CanTransform(err *DomainError) bool
}

// EnrichmentMiddleware define middleware para enriquecimento
type EnrichmentMiddleware interface {
	Middleware
	// Enrich enriquece o erro com dados adicionais
	Enrich(ctx *MiddlewareContext, err *DomainError) error
	// GetEnrichmentData retorna dados para enriquecimento
	GetEnrichmentData(ctx *MiddlewareContext) map[string]interface{}
}

// FilteringMiddleware define middleware para filtragem
type FilteringMiddleware interface {
	Middleware
	// Filter filtra o erro
	Filter(ctx *MiddlewareContext, err *DomainError) (*DomainError, bool, error)
	// ShouldFilter verifica se deve filtrar este erro
	ShouldFilter(err *DomainError) bool
}

// SecurityMiddleware define middleware para segurança
type SecurityMiddleware interface {
	Middleware
	// SanitizeError sanitiza informações sensíveis do erro
	SanitizeError(ctx *MiddlewareContext, err *DomainError) (*DomainError, error)
	// IsSensitive verifica se o erro contém informações sensíveis
	IsSensitive(err *DomainError) bool
}

// RetryMiddleware define middleware para retry
type RetryMiddleware interface {
	Middleware
	// ShouldRetry verifica se deve tentar novamente
	ShouldRetry(ctx *MiddlewareContext, err *DomainError, attempt int) bool
	// GetRetryDelay retorna o delay para retry
	GetRetryDelay(attempt int) time.Duration
	// MaxRetries retorna o número máximo de tentativas
	MaxRetries() int
}

// CircuitBreakerMiddleware define middleware para circuit breaker
type CircuitBreakerMiddleware interface {
	Middleware
	// ShouldBreak verifica se deve quebrar o circuito
	ShouldBreak(ctx *MiddlewareContext, err *DomainError) bool
	// RecordFailure registra uma falha
	RecordFailure(ctx *MiddlewareContext, err *DomainError) error
	// RecordSuccess registra um sucesso
	RecordSuccess(ctx *MiddlewareContext) error
	// IsCircuitOpen verifica se o circuito está aberto
	IsCircuitOpen() bool
}

// ConditionalMiddleware define middleware condicional
type ConditionalMiddleware interface {
	Middleware
	// Condition verifica se o middleware deve ser executado
	Condition(ctx *MiddlewareContext) bool
}

// AsyncMiddleware define middleware assíncrono
type AsyncMiddleware interface {
	Middleware
	// HandleAsync processa o middleware de forma assíncrona
	HandleAsync(ctx *MiddlewareContext, next NextFunction) <-chan error
	// Timeout retorna o timeout para execução assíncrona
	Timeout() time.Duration
}
