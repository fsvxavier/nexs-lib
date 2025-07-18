package interfaces

import (
	"context"
	"encoding/json"
)

// ErrorType define os tipos de erro disponíveis
type ErrorType string

// Definição dos tipos de erro
const (
	ErrorTypeValidation         ErrorType = "validation"
	ErrorTypeNotFound           ErrorType = "not_found"
	ErrorTypeBusiness           ErrorType = "business"
	ErrorTypeDatabase           ErrorType = "database"
	ErrorTypeExternalService    ErrorType = "external_service"
	ErrorTypeInfrastructure     ErrorType = "infrastructure"
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

// DomainError é a interface base para todos os erros de domínio
type DomainError interface {
	error

	// Error retorna a mensagem de erro
	Error() string

	// Unwrap retorna o erro encapsulado
	Unwrap() error

	// Type retorna o tipo do erro
	Type() ErrorType

	// Metadata retorna metadados adicionais
	Metadata() map[string]interface{}

	// HTTPStatus retorna o código de status HTTP apropriado
	HTTPStatus() int

	// StackTrace retorna o stack trace capturado
	StackTrace() string

	// WithContext adiciona contexto ao erro
	WithContext(ctx context.Context) DomainError

	// Wrap encapsula outro erro com contexto opcional
	Wrap(message string, err error) DomainError

	// JSON serializa o erro para JSON
	JSON() ([]byte, error)
}

// ErrorValidator define interface para validação de erros
type ErrorValidator interface {
	// IsType verifica se o erro é do tipo especificado
	IsType(err error, errorType ErrorType) bool

	// ExtractType extrai o tipo do erro
	ExtractType(err error) (ErrorType, bool)
}

// ErrorHandler define interface para manipulação de erros
type ErrorHandler interface {
	// Handle processa o erro e retorna uma resposta apropriada
	Handle(ctx context.Context, err error) interface{}

	// ShouldRetry determina se a operação deve ser repetida
	ShouldRetry(err error) bool

	// GetRetryDelay retorna o delay para retry
	GetRetryDelay(err error, attempt int) int
}

// ErrorRegistry define interface para registro de códigos de erro
type ErrorRegistry interface {
	// RegisterCode registra um código de erro
	RegisterCode(code string, description string, httpStatus int)

	// GetDescription retorna a descrição do código
	GetDescription(code string) (string, bool)

	// GetHTTPStatus retorna o status HTTP do código
	GetHTTPStatus(code string) (int, bool)

	// ListCodes lista todos os códigos registrados
	ListCodes() map[string]string
}

// ErrorMetrics define interface para métricas de erro
type ErrorMetrics interface {
	// RecordError registra um erro nas métricas
	RecordError(errorType ErrorType, code string)

	// GetErrorCount retorna o contador de erros
	GetErrorCount(errorType ErrorType) int64

	// GetErrorRate retorna a taxa de erros
	GetErrorRate(errorType ErrorType) float64
}

// StackTraceProvider define interface para captura de stack trace
type StackTraceProvider interface {
	// CaptureStackTrace captura o stack trace atual
	CaptureStackTrace(skip int) string

	// FormatStackTrace formata o stack trace
	FormatStackTrace(trace string) string
}

// ErrorContextProvider define interface para contexto de erro
type ErrorContextProvider interface {
	// AddContext adiciona contexto ao erro
	AddContext(key string, value interface{})

	// GetContext retorna o contexto do erro
	GetContext(key string) (interface{}, bool)

	// GetAllContext retorna todo o contexto
	GetAllContext() map[string]interface{}
}

// HTTPStatusProvider define interface para códigos de status HTTP
type HTTPStatusProvider interface {
	// StatusCode retorna o código de status HTTP
	StatusCode() int
}

// HasCode define interface para erros que têm código
type HasCode interface {
	// Code retorna o código do erro
	Code() string
}

// ErrorSerializer define interface para serialização de erros
type ErrorSerializer interface {
	// Serialize serializa o erro
	Serialize(err error) ([]byte, error)

	// Deserialize deserializa o erro
	Deserialize(data []byte) (error, error)

	// Format formata o erro para apresentação
	Format(err error) string
}

// ErrorAggregator define interface para agregação de erros
type ErrorAggregator interface {
	// AddError adiciona um erro à agregação
	AddError(err error)

	// HasErrors retorna se há erros
	HasErrors() bool

	// GetErrors retorna todos os erros
	GetErrors() []error

	// GetFirstError retorna o primeiro erro
	GetFirstError() error

	// Clear limpa todos os erros
	Clear()
}

// ErrorChainWalker define interface para percorrer cadeia de erros
type ErrorChainWalker interface {
	// Walk percorre a cadeia de erros
	Walk(err error, fn func(error) bool)

	// FindFirst encontra o primeiro erro que satisfaz a condição
	FindFirst(err error, fn func(error) bool) error

	// GetChain retorna toda a cadeia de erros
	GetChain(err error) []error
}

// ErrorFactory define interface para criação de erros
type ErrorFactory interface {
	// NewError cria um novo erro
	NewError(code string, message string) DomainError

	// NewErrorWithType cria um novo erro com tipo específico
	NewErrorWithType(code string, message string, errorType ErrorType) DomainError

	// NewErrorFromTemplate cria um erro a partir de template
	NewErrorFromTemplate(template string, args ...interface{}) DomainError

	// WrapError encapsula um erro existente
	WrapError(message string, err error) DomainError
}

// ErrorRecovery define interface para recuperação de erros
type ErrorRecovery interface {
	// CanRecover verifica se pode recuperar do erro
	CanRecover(err error) bool

	// Recover tenta recuperar do erro
	Recover(ctx context.Context, err error) error

	// GetRecoveryStrategy retorna a estratégia de recuperação
	GetRecoveryStrategy(err error) string
}

// ErrorNotifier define interface para notificação de erros
type ErrorNotifier interface {
	// Notify notifica sobre um erro
	Notify(ctx context.Context, err error) error

	// ShouldNotify verifica se deve notificar
	ShouldNotify(err error) bool

	// GetNotificationLevel retorna o nível de notificação
	GetNotificationLevel(err error) string
}

// ErrorLocalizer define interface para localização de erros
type ErrorLocalizer interface {
	// Localize localiza a mensagem do erro
	Localize(ctx context.Context, err error) string

	// GetSupportedLocales retorna os locales suportados
	GetSupportedLocales() []string

	// SetLocale define o locale
	SetLocale(locale string) error
}

// ErrorTransformer define interface para transformação de erros
type ErrorTransformer interface {
	// Transform transforma um erro
	Transform(err error) error

	// CanTransform verifica se pode transformar
	CanTransform(err error) bool

	// GetTransformationRules retorna as regras de transformação
	GetTransformationRules() map[string]string
}

// ErrorFilter define interface para filtro de erros
type ErrorFilter interface {
	// ShouldFilter verifica se deve filtrar o erro
	ShouldFilter(err error) bool

	// Filter filtra o erro
	Filter(err error) error

	// GetFilterRules retorna as regras de filtro
	GetFilterRules() []string
}

// ErrorEnricher define interface para enriquecimento de erros
type ErrorEnricher interface {
	// Enrich enriquece o erro com informações adicionais
	Enrich(ctx context.Context, err error) error

	// GetEnrichmentData retorna dados de enriquecimento
	GetEnrichmentData(err error) map[string]interface{}

	// SetEnrichmentProvider define o provedor de enriquecimento
	SetEnrichmentProvider(provider interface{}) error
}

// ErrorMarshalJSON define interface para serialização JSON customizada
type ErrorMarshalJSON interface {
	json.Marshaler
	json.Unmarshaler
}

// ErrorConfig define interface para configuração de erros
type ErrorConfig interface {
	// SetStackTraceEnabled habilita/desabilita stack trace
	SetStackTraceEnabled(enabled bool)

	// IsStackTraceEnabled retorna se stack trace está habilitado
	IsStackTraceEnabled() bool

	// SetMaxStackDepth define a profundidade máxima do stack
	SetMaxStackDepth(depth int)

	// GetMaxStackDepth retorna a profundidade máxima do stack
	GetMaxStackDepth() int

	// SetDefaultHTTPStatus define o status HTTP padrão
	SetDefaultHTTPStatus(status int)

	// GetDefaultHTTPStatus retorna o status HTTP padrão
	GetDefaultHTTPStatus() int
}
