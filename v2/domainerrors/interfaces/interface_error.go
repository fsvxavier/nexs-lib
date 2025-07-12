// Package interfaces define as interfaces para o sistema de erros de domínio.
// Seguindo os princípios de Inversão de Dependências e Interface Segregation.
package interfaces

import (
	"context"
)

// DomainErrorInterface define o contrato básico para erros de domínio.
type DomainErrorInterface interface {
	error
	ErrorWrapper
	ErrorFormatter
	ErrorMetadata
	ErrorHTTP
}

// ErrorWrapper define capacidades de wrapping de erros.
type ErrorWrapper interface {
	// Unwrap retorna o erro subjacente
	Unwrap() error

	// Wrap adiciona um erro ao stack e retorna uma nova instância
	Wrap(message string, err error) DomainErrorInterface

	// Chain adiciona um erro ao final da cadeia
	Chain(err error) DomainErrorInterface

	// RootCause retorna o erro original da cadeia
	RootCause() error
}

// ErrorFormatter define capacidades de formatação de erros.
type ErrorFormatter interface {
	// String retorna uma representação string do erro
	String() string

	// JSON retorna uma representação JSON do erro
	JSON() ([]byte, error)

	// FormatStackTrace retorna o stack trace formatado
	FormatStackTrace() string

	// DetailedString retorna uma string detalhada incluindo metadados
	DetailedString() string
}

// ErrorMetadata define capacidades de metadados de erros.
type ErrorMetadata interface {
	// Code retorna o código do erro
	Code() string

	// Type retorna o tipo do erro
	Type() string

	// Message retorna a mensagem do erro
	Message() string

	// Details retorna os detalhes do erro
	Details() map[string]interface{}

	// Metadata retorna os metadados internos
	Metadata() map[string]interface{}

	// Severity retorna a severidade do erro
	Severity() Severity

	// Category retorna a categoria do erro
	Category() Category

	// Tags retorna as tags associadas ao erro
	Tags() []string

	// IsRetryable indica se o erro permite retry
	IsRetryable() bool

	// IsTemporary indica se o erro é temporário
	IsTemporary() bool
}

// ErrorHTTP define capacidades específicas para HTTP.
type ErrorHTTP interface {
	// StatusCode retorna o código de status HTTP
	StatusCode() int

	// Headers retorna headers HTTP adicionais
	Headers() map[string]string

	// ResponseBody retorna o corpo da resposta HTTP
	ResponseBody() interface{}

	// SetStatusCode define o código de status HTTP
	SetStatusCode(code int) DomainErrorInterface
}

// ErrorBuilder define o contrato para construção fluente de erros.
type ErrorBuilder interface {
	// WithCode define o código do erro
	WithCode(code string) ErrorBuilder

	// WithMessage define a mensagem do erro
	WithMessage(message string) ErrorBuilder

	// WithType define o tipo do erro
	WithType(errorType string) ErrorBuilder

	// WithSeverity define a severidade do erro
	WithSeverity(severity Severity) ErrorBuilder

	// WithCategory define a categoria do erro
	WithCategory(category Category) ErrorBuilder

	// WithDetail adiciona um detalhe específico
	WithDetail(key string, value interface{}) ErrorBuilder

	// WithDetails adiciona múltiplos detalhes
	WithDetails(details map[string]interface{}) ErrorBuilder

	// WithMetadata adiciona metadados
	WithMetadata(metadata map[string]interface{}) ErrorBuilder

	// WithTag adiciona uma tag
	WithTag(tag string) ErrorBuilder

	// WithTags adiciona múltiplas tags
	WithTags(tags []string) ErrorBuilder

	// WithCause define o erro causa
	WithCause(err error) ErrorBuilder

	// WithStatusCode define o código de status HTTP
	WithStatusCode(code int) ErrorBuilder

	// WithHeader adiciona um header HTTP
	WithHeader(key, value string) ErrorBuilder

	// WithHeaders adiciona múltiplos headers HTTP
	WithHeaders(headers map[string]string) ErrorBuilder

	// WithContext adiciona informações de contexto
	WithContext(ctx context.Context) ErrorBuilder

	// Build constrói o erro final
	Build() DomainErrorInterface
}

// ErrorFactory define o contrato para criação de erros.
type ErrorFactory interface {
	// New cria um novo erro com código e mensagem
	New(code, message string) DomainErrorInterface

	// NewWithCause cria um novo erro com causa
	NewWithCause(code, message string, cause error) DomainErrorInterface

	// NewValidation cria um erro de validação
	NewValidation(message string, fields map[string][]string) ValidationErrorInterface

	// NewNotFound cria um erro de não encontrado
	NewNotFound(entity, id string) DomainErrorInterface

	// NewUnauthorized cria um erro de não autorizado
	NewUnauthorized(message string) DomainErrorInterface

	// NewForbidden cria um erro de acesso negado
	NewForbidden(message string) DomainErrorInterface

	// NewInternal cria um erro interno
	NewInternal(message string, cause error) DomainErrorInterface

	// NewBadRequest cria um erro de requisição inválida
	NewBadRequest(message string) DomainErrorInterface

	// NewConflict cria um erro de conflito
	NewConflict(message string) DomainErrorInterface

	// NewTimeout cria um erro de timeout
	NewTimeout(message string) DomainErrorInterface

	// NewCircuitBreaker cria um erro de circuit breaker
	NewCircuitBreaker(service string) DomainErrorInterface

	// Builder retorna um novo builder
	Builder() ErrorBuilder
}

// ValidationErrorInterface define o contrato para erros de validação.
type ValidationErrorInterface interface {
	DomainErrorInterface

	// Fields retorna os campos com erro
	Fields() map[string][]string

	// AddField adiciona um erro de campo
	AddField(field, message string) ValidationErrorInterface

	// AddFields adiciona múltiplos erros de campo
	AddFields(fields map[string][]string) ValidationErrorInterface

	// HasField verifica se um campo tem erro
	HasField(field string) bool

	// FieldErrors retorna os erros de um campo específico
	FieldErrors(field string) []string
}

// ErrorParser define o contrato para parsers de erro.
type ErrorParser interface {
	// CanParse verifica se o parser pode processar o erro
	CanParse(err error) bool

	// Parse processa o erro e retorna informações estruturadas
	Parse(err error) ParsedError
}

// ErrorRegistry define o contrato para registro de códigos de erro.
type ErrorRegistry interface {
	// Register registra um novo código de erro
	Register(info ErrorCodeInfo) error

	// Get obtém informações de um código de erro
	Get(code string) (ErrorCodeInfo, bool)

	// Exists verifica se um código existe
	Exists(code string) bool

	// List retorna todos os códigos registrados
	List() []ErrorCodeInfo

	// CreateError cria um erro baseado em um código registrado
	CreateError(code string, args ...interface{}) (DomainErrorInterface, error)
}

// ErrorHandler define o contrato para tratamento de erros.
type ErrorHandler interface {
	// Handle processa um erro e retorna uma resposta apropriada
	Handle(ctx context.Context, err error) *ErrorResponse

	// CanHandle verifica se o handler pode processar o erro
	CanHandle(err error) bool

	// Priority retorna a prioridade do handler (menor = maior prioridade)
	Priority() int
}

// ErrorMiddleware define o contrato para middleware de erro.
type ErrorMiddleware interface {
	// Process processa o erro e retorna o erro modificado ou nil
	Process(ctx context.Context, err error) error

	// Order retorna a ordem de execução do middleware
	Order() int
}

// ErrorReporter define o contrato para relatório de erros.
type ErrorReporter interface {
	// Report envia o erro para o sistema de monitoramento
	Report(ctx context.Context, err error) error

	// ShouldReport verifica se o erro deve ser reportado
	ShouldReport(err error) bool
}

// Severity define os níveis de severidade de erro.
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String retorna a representação string da severidade.
func (s Severity) String() string {
	switch s {
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

// IsValid verifica se a severidade é válida.
func (s Severity) IsValid() bool {
	return s >= SeverityLow && s <= SeverityCritical
}

// Category define as categorias de erro.
type Category string

const (
	CategoryBusiness       Category = "business"
	CategoryTechnical      Category = "technical"
	CategoryInfrastructure Category = "infrastructure"
	CategorySecurity       Category = "security"
	CategoryPerformance    Category = "performance"
	CategoryIntegration    Category = "integration"
)

// String retorna a representação string da categoria.
func (c Category) String() string {
	return string(c)
}

// IsValid verifica se a categoria é válida.
func (c Category) IsValid() bool {
	switch c {
	case CategoryBusiness, CategoryTechnical, CategoryInfrastructure,
		CategorySecurity, CategoryPerformance, CategoryIntegration:
		return true
	default:
		return false
	}
}

// ParsedError representa um erro parseado.
type ParsedError struct {
	Code      string
	Message   string
	Type      string
	Details   map[string]interface{}
	Severity  Severity
	Category  Category
	Retryable bool
	Temporary bool
}

// ErrorCodeInfo representa informações sobre um código de erro.
type ErrorCodeInfo struct {
	Code        string
	Message     string
	Type        string
	StatusCode  int
	Severity    Severity
	Category    Category
	Retryable   bool
	Temporary   bool
	Tags        []string
	Description string
	Examples    []string
}

// ErrorResponse representa uma resposta de erro estruturada.
type ErrorResponse struct {
	Error      interface{}            `json:"error"`
	StatusCode int                    `json:"-"`
	Headers    map[string]string      `json:"-"`
	Metadata   map[string]interface{} `json:"-"`
}

// HTTPErrorResponse representa uma resposta HTTP de erro.
type HTTPErrorResponse struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Path      string                 `json:"path,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}
