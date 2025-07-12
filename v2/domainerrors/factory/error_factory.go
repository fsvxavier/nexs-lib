// Package factory implementa o padrão Factory para criação de erros de domínio.
// Segue os princípios de Injeção de Dependências e Factory Pattern.
package factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// DefaultErrorFactory implementa interfaces.ErrorFactory.
type DefaultErrorFactory struct {
	defaultCode      string
	defaultSeverity  types.ErrorSeverity
	enableStackTrace bool
	timestampFormat  string
}

// NewDefaultFactory cria uma nova instância da factory padrão.
func NewDefaultFactory() interfaces.ErrorFactory {
	return &DefaultErrorFactory{
		defaultCode:      "E999",
		defaultSeverity:  types.SeverityMedium,
		enableStackTrace: true,
		timestampFormat:  time.RFC3339,
	}
}

// NewCustomFactory cria uma factory com configurações personalizadas.
func NewCustomFactory(defaultCode string, defaultSeverity types.ErrorSeverity, enableStackTrace bool) interfaces.ErrorFactory {
	return &DefaultErrorFactory{
		defaultCode:      defaultCode,
		defaultSeverity:  defaultSeverity,
		enableStackTrace: enableStackTrace,
		timestampFormat:  time.RFC3339,
	}
}

// New cria um novo erro com código e mensagem.
func (f *DefaultErrorFactory) New(code, message string) interfaces.DomainErrorInterface {
	if code == "" {
		code = f.defaultCode
	}

	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode(code).
		WithMessage(message).
		Build()

	return err
}

// NewWithCause cria um novo erro com causa.
func (f *DefaultErrorFactory) NewWithCause(code, message string, cause error) interfaces.DomainErrorInterface {
	if code == "" {
		code = f.defaultCode
	}

	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode(code).
		WithMessage(message).
		WithCause(cause).
		Build()

	return err
}

// NewValidation cria um erro de validação.
func (f *DefaultErrorFactory) NewValidation(message string, fields map[string][]string) interfaces.ValidationErrorInterface {
	return domainerrors.NewValidationError(message, fields)
}

// NewNotFound cria um erro de não encontrado.
func (f *DefaultErrorFactory) NewNotFound(entity, id string) interfaces.DomainErrorInterface {
	message := fmt.Sprintf("%s not found", entity)
	if id != "" {
		message = fmt.Sprintf("%s with ID '%s' not found", entity, id)
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E002").
		WithMessage(message).
		WithType(string(types.ErrorTypeNotFound)).
		WithDetail("entity", entity).
		WithDetail("id", id).
		WithTag("not_found").
		Build()
}

// NewUnauthorized cria um erro de não autorizado.
func (f *DefaultErrorFactory) NewUnauthorized(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Authentication required"
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E005").
		WithMessage(message).
		WithType(string(types.ErrorTypeAuthentication)).
		WithTag("authentication").
		Build()
}

// NewForbidden cria um erro de acesso negado.
func (f *DefaultErrorFactory) NewForbidden(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Access denied"
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E006").
		WithMessage(message).
		WithType(string(types.ErrorTypeAuthorization)).
		WithTag("authorization").
		Build()
}

// NewInternal cria um erro interno.
func (f *DefaultErrorFactory) NewInternal(message string, cause error) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Internal server error"
	}

	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode("E007").
		WithMessage(message).
		WithType(string(types.ErrorTypeInternal)).
		WithSeverity(interfaces.Severity(types.SeverityCritical)).
		WithTag("internal").
		Build()

	if cause != nil {
		err = err.Wrap("internal error", cause)
	}

	return err
}

// NewBadRequest cria um erro de requisição inválida.
func (f *DefaultErrorFactory) NewBadRequest(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Bad request"
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E001").
		WithMessage(message).
		WithType(string(types.ErrorTypeBadRequest)).
		WithTag("bad_request").
		Build()
}

// NewConflict cria um erro de conflito.
func (f *DefaultErrorFactory) NewConflict(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Resource conflict"
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E003").
		WithMessage(message).
		WithType(string(types.ErrorTypeConflict)).
		WithTag("conflict").
		Build()
}

// NewTimeout cria um erro de timeout.
func (f *DefaultErrorFactory) NewTimeout(message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = "Operation timeout"
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E009").
		WithMessage(message).
		WithType(string(types.ErrorTypeTimeout)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithTag("timeout").
		Build()
}

// NewCircuitBreaker cria um erro de circuit breaker.
func (f *DefaultErrorFactory) NewCircuitBreaker(service string) interfaces.DomainErrorInterface {
	message := "Circuit breaker is open"
	if service != "" {
		message = fmt.Sprintf("Circuit breaker is open for service: %s", service)
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("E013").
		WithMessage(message).
		WithType(string(types.ErrorTypeCircuitBreaker)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("service", service).
		WithTag("circuit_breaker").
		Build()
}

// Builder retorna um novo builder.
func (f *DefaultErrorFactory) Builder() interfaces.ErrorBuilder {
	return domainerrors.NewBuilder()
}

// SpecializedFactory factory especializada para tipos específicos de erro.
type SpecializedFactory struct {
	*DefaultErrorFactory
	prefix   string
	tags     []string
	metadata map[string]interface{}
}

// NewSpecializedFactory cria uma factory especializada com configurações específicas.
func NewSpecializedFactory(prefix string, tags []string, metadata map[string]interface{}) interfaces.ErrorFactory {
	return &SpecializedFactory{
		DefaultErrorFactory: NewDefaultFactory().(*DefaultErrorFactory),
		prefix:              prefix,
		tags:                tags,
		metadata:            metadata,
	}
}

// New cria um erro com configurações especializadas.
func (f *SpecializedFactory) New(code, message string) interfaces.DomainErrorInterface {
	if f.prefix != "" && code != "" {
		code = f.prefix + "_" + code
	}

	builder := domainerrors.NewBuilder().
		WithCode(code).
		WithMessage(message)

	// Adiciona tags padrão
	if len(f.tags) > 0 {
		builder.WithTags(f.tags)
	}

	// Adiciona metadados padrão
	if len(f.metadata) > 0 {
		builder.WithMetadata(f.metadata)
	}

	return builder.Build()
}

// DatabaseErrorFactory factory especializada para erros de banco de dados.
type DatabaseErrorFactory struct {
	*DefaultErrorFactory
}

// NewDatabaseErrorFactory cria uma factory para erros de banco de dados.
func NewDatabaseErrorFactory() *DatabaseErrorFactory {
	return &DatabaseErrorFactory{
		DefaultErrorFactory: NewDefaultFactory().(*DefaultErrorFactory),
	}
}

// NewConnectionError cria um erro de conexão com banco.
func (f *DatabaseErrorFactory) NewConnectionError(database string, cause error) interfaces.DomainErrorInterface {
	message := fmt.Sprintf("Failed to connect to database: %s", database)

	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode("DB001").
		WithMessage(message).
		WithType(string(types.ErrorTypeDatabase)).
		WithSeverity(interfaces.Severity(types.SeverityCritical)).
		WithDetail("database", database).
		WithTag("database").
		WithTag("connection").
		Build()

	if cause != nil {
		err = err.Wrap("database connection failed", cause)
	}

	return err
}

// NewQueryError cria um erro de consulta SQL.
func (f *DatabaseErrorFactory) NewQueryError(query string, cause error) interfaces.DomainErrorInterface {
	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode("DB002").
		WithMessage("Database query failed").
		WithType(string(types.ErrorTypeDatabase)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("query", query).
		WithTag("database").
		WithTag("query").
		Build()

	if cause != nil {
		err = err.Wrap("query execution failed", cause)
	}

	return err
}

// NewTransactionError cria um erro de transação.
func (f *DatabaseErrorFactory) NewTransactionError(operation string, cause error) interfaces.DomainErrorInterface {
	message := fmt.Sprintf("Transaction %s failed", operation)

	builder := domainerrors.NewBuilder()
	err := builder.
		WithCode("DB003").
		WithMessage(message).
		WithType(string(types.ErrorTypeDatabase)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("operation", operation).
		WithTag("database").
		WithTag("transaction").
		Build()

	if cause != nil {
		err = err.Wrap("transaction failed", cause)
	}

	return err
}

// HTTPErrorFactory factory especializada para erros HTTP.
type HTTPErrorFactory struct {
	*DefaultErrorFactory
}

// NewHTTPErrorFactory cria uma factory para erros HTTP.
func NewHTTPErrorFactory() *HTTPErrorFactory {
	return &HTTPErrorFactory{
		DefaultErrorFactory: NewDefaultFactory().(*DefaultErrorFactory),
	}
}

// NewHTTPError cria um erro HTTP com status code.
func (f *HTTPErrorFactory) NewHTTPError(statusCode int, message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = fmt.Sprintf("HTTP %d error", statusCode)
	}

	code := fmt.Sprintf("HTTP%d", statusCode)

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode(code).
		WithMessage(message).
		WithType(string(types.ErrorTypeHTTP)).
		WithStatusCode(statusCode).
		WithTag("http").
		Build()
}

// NewServiceUnavailableError cria um erro de serviço indisponível.
func (f *HTTPErrorFactory) NewServiceUnavailableError(service string) interfaces.DomainErrorInterface {
	message := "Service unavailable"
	if service != "" {
		message = fmt.Sprintf("Service unavailable: %s", service)
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode("HTTP503").
		WithMessage(message).
		WithType(string(types.ErrorTypeExternalService)).
		WithStatusCode(503).
		WithDetail("service", service).
		WithTag("http").
		WithTag("service_unavailable").
		Build()
}

// BusinessErrorFactory factory especializada para erros de negócio.
type BusinessErrorFactory struct {
	*DefaultErrorFactory
	domain string
}

// NewBusinessErrorFactory cria uma factory para erros de negócio.
func NewBusinessErrorFactory(domain string) *BusinessErrorFactory {
	return &BusinessErrorFactory{
		DefaultErrorFactory: NewDefaultFactory().(*DefaultErrorFactory),
		domain:              domain,
	}
}

// NewBusinessRuleError cria um erro de regra de negócio.
func (f *BusinessErrorFactory) NewBusinessRuleError(rule, message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = fmt.Sprintf("Business rule violation: %s", rule)
	}

	code := "BUS001"
	if f.domain != "" {
		code = fmt.Sprintf("%s_BUS001", strings.ToUpper(f.domain))
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode(code).
		WithMessage(message).
		WithType(string(types.ErrorTypeBusinessRule)).
		WithDetail("rule", rule).
		WithDetail("domain", f.domain).
		WithTag("business").
		WithTag("rule_violation").
		Build()
}

// NewInvariantViolationError cria um erro de violação de invariante.
func (f *BusinessErrorFactory) NewInvariantViolationError(invariant, message string) interfaces.DomainErrorInterface {
	if message == "" {
		message = fmt.Sprintf("Domain invariant violated: %s", invariant)
	}

	code := "BUS002"
	if f.domain != "" {
		code = fmt.Sprintf("%s_BUS002", strings.ToUpper(f.domain))
	}

	builder := domainerrors.NewBuilder()
	return builder.
		WithCode(code).
		WithMessage(message).
		WithType(string(types.ErrorTypeBusinessRule)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("invariant", invariant).
		WithDetail("domain", f.domain).
		WithTag("business").
		WithTag("invariant_violation").
		Build()
}

// Variáveis globais para factories comuns (singleton pattern)
var (
	defaultFactory  interfaces.ErrorFactory
	databaseFactory *DatabaseErrorFactory
	httpFactory     *HTTPErrorFactory
)

// GetDefaultFactory retorna a factory padrão (singleton).
func GetDefaultFactory() interfaces.ErrorFactory {
	if defaultFactory == nil {
		defaultFactory = NewDefaultFactory()
	}
	return defaultFactory
}

// GetDatabaseFactory retorna a factory de erros de banco (singleton).
func GetDatabaseFactory() *DatabaseErrorFactory {
	if databaseFactory == nil {
		databaseFactory = NewDatabaseErrorFactory()
	}
	return databaseFactory
}

// GetHTTPFactory retorna a factory de erros HTTP (singleton).
func GetHTTPFactory() *HTTPErrorFactory {
	if httpFactory == nil {
		httpFactory = NewHTTPErrorFactory()
	}
	return httpFactory
}
