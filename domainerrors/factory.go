package domainerrors

import (
	"context"
	"fmt"
	"time"
)

// ErrorFactory define a interface para criação de erros
type ErrorFactory interface {
	CreateValidationError(message string, fields map[string][]string) error
	CreateNotFoundError(message string, resourceType, resourceID string) error
	CreateBusinessError(code, message string) error
	CreateInfrastructureError(component, message string, err error) error
	CreateExternalServiceError(service, message string, statusCode int, err error) error
	CreateAuthenticationError(message, reason string) error
	CreateAuthorizationError(message, permission string) error
	CreateTimeoutError(message string, threshold time.Duration) error
	CreateConflictError(message, resource string) error
	CreateRateLimitError(message string, limit int, window time.Duration) error
}

// DomainErrorFactory implementa ErrorFactory usando Builder Pattern
type DomainErrorFactory struct {
	builderFactory  func() *ErrorBuilder
	contextEnricher ContextEnricher
	observers       []ErrorObserver
}

// ContextEnricher enriquece erros com informações contextuais
type ContextEnricher interface {
	EnrichError(ctx context.Context, builder *ErrorBuilder) *ErrorBuilder
}

// ErrorObserver observa a criação de erros para logging/monitoramento
type ErrorObserver interface {
	OnErrorCreated(error error, context map[string]interface{})
}

// NewDomainErrorFactory cria uma nova factory de erros
func NewDomainErrorFactory(options ...FactoryOption) *DomainErrorFactory {
	factory := &DomainErrorFactory{
		builderFactory: NewErrorBuilder,
		observers:      make([]ErrorObserver, 0),
	}

	for _, option := range options {
		option(factory)
	}

	return factory
}

// FactoryOption define opções para configurar a factory
type FactoryOption func(*DomainErrorFactory)

// WithBuilderFactory define uma factory customizada para builders
func WithBuilderFactory(builderFactory func() *ErrorBuilder) FactoryOption {
	return func(f *DomainErrorFactory) {
		f.builderFactory = builderFactory
	}
}

// WithContextEnricher define um enricher de contexto
func WithContextEnricher(enricher ContextEnricher) FactoryOption {
	return func(f *DomainErrorFactory) {
		f.contextEnricher = enricher
	}
}

// WithErrorObserver adiciona um observer de erros
func WithErrorObserver(observer ErrorObserver) FactoryOption {
	return func(f *DomainErrorFactory) {
		f.observers = append(f.observers, observer)
	}
}

// createError é um método helper para criar erros com enrichment e observação
func (f *DomainErrorFactory) createError(ctx context.Context, builder *ErrorBuilder) error {
	// Enriquece com contexto se disponível
	if f.contextEnricher != nil && ctx != nil {
		builder = f.contextEnricher.EnrichError(ctx, builder)
	}

	// Constrói o erro
	err := builder.Build()

	// Notifica observers
	for _, observer := range f.observers {
		observerCtx := make(map[string]interface{})
		if ctx != nil {
			observerCtx["context"] = ctx
		}
		observer.OnErrorCreated(err, observerCtx)
	}

	return err
}

// CreateValidationError implementa ErrorFactory
func (f *DomainErrorFactory) CreateValidationError(message string, fields map[string][]string) error {
	return f.createError(nil, f.builderFactory().
		Code("VALIDATION_ERROR").
		Message(message).
		Type(ErrorTypeValidation).
		ValidationFields(fields).
		WithTimestamp())
}

// CreateNotFoundError implementa ErrorFactory
func (f *DomainErrorFactory) CreateNotFoundError(message string, resourceType, resourceID string) error {
	return f.createError(nil, f.builderFactory().
		Code("NOT_FOUND").
		Message(message).
		Type(ErrorTypeNotFound).
		ResourceInfo(resourceType, resourceID).
		WithTimestamp())
}

// CreateBusinessError implementa ErrorFactory
func (f *DomainErrorFactory) CreateBusinessError(code, message string) error {
	return f.createError(nil, f.builderFactory().
		Code(code).
		Message(message).
		Type(ErrorTypeBusinessRule).
		WithTimestamp())
}

// CreateInfrastructureError implementa ErrorFactory
func (f *DomainErrorFactory) CreateInfrastructureError(component, message string, err error) error {
	return f.createError(nil, f.builderFactory().
		Code("INFRA_ERROR").
		Message(message).
		Type(ErrorTypeInfrastructure).
		Cause(err).
		Detail("component", component).
		WithTimestamp())
}

// CreateExternalServiceError implementa ErrorFactory
func (f *DomainErrorFactory) CreateExternalServiceError(service, message string, statusCode int, err error) error {
	return f.createError(nil, f.builderFactory().
		Code("EXTERNAL_SERVICE_ERROR").
		Message(message).
		Type(ErrorTypeExternalService).
		Cause(err).
		ExternalService(service, statusCode).
		WithTimestamp())
}

// CreateAuthenticationError implementa ErrorFactory
func (f *DomainErrorFactory) CreateAuthenticationError(message, reason string) error {
	return f.createError(nil, f.builderFactory().
		Code("AUTH_ERROR").
		Message(message).
		Type(ErrorTypeAuthentication).
		Detail("reason", reason).
		WithTimestamp())
}

// CreateAuthorizationError implementa ErrorFactory
func (f *DomainErrorFactory) CreateAuthorizationError(message, permission string) error {
	return f.createError(nil, f.builderFactory().
		Code("AUTHZ_ERROR").
		Message(message).
		Type(ErrorTypeAuthorization).
		Detail("required_permission", permission).
		WithTimestamp())
}

// CreateTimeoutError implementa ErrorFactory
func (f *DomainErrorFactory) CreateTimeoutError(message string, threshold time.Duration) error {
	return f.createError(nil, f.builderFactory().
		Code("TIMEOUT_ERROR").
		Message(message).
		Type(ErrorTypeTimeout).
		Timeout(threshold).
		WithTimestamp())
}

// CreateConflictError implementa ErrorFactory
func (f *DomainErrorFactory) CreateConflictError(message, resource string) error {
	return f.createError(nil, f.builderFactory().
		Code("CONFLICT_ERROR").
		Message(message).
		Type(ErrorTypeConflict).
		Conflict(resource).
		WithTimestamp())
}

// CreateRateLimitError implementa ErrorFactory
func (f *DomainErrorFactory) CreateRateLimitError(message string, limit int, window time.Duration) error {
	return f.createError(nil, f.builderFactory().
		Code("RATE_LIMIT_ERROR").
		Message(message).
		Type(ErrorTypeRateLimit).
		Detail("rate_limit", limit).
		Detail("time_window", window.String()).
		WithTimestamp())
}

// CreateWithContext cria um erro com contexto específico
func (f *DomainErrorFactory) CreateWithContext(ctx context.Context, builder *ErrorBuilder) error {
	return f.createError(ctx, builder)
}

// DefaultContextEnricher implementa ContextEnricher com informações básicas
type DefaultContextEnricher struct{}

// NewDefaultContextEnricher cria um enricher padrão
func NewDefaultContextEnricher() *DefaultContextEnricher {
	return &DefaultContextEnricher{}
}

// EnrichError implementa ContextEnricher
func (e *DefaultContextEnricher) EnrichError(ctx context.Context, builder *ErrorBuilder) *ErrorBuilder {
	if ctx == nil {
		return builder
	}

	// Enriquece com valores do contexto se disponíveis
	if requestID := ctx.Value("request_id"); requestID != nil {
		if requestIDStr, ok := requestID.(string); ok {
			builder.WithRequestID(requestIDStr)
		}
	}

	if userID := ctx.Value("user_id"); userID != nil {
		if userIDStr, ok := userID.(string); ok {
			builder.WithUserID(userIDStr)
		}
	}

	if operationID := ctx.Value("operation_id"); operationID != nil {
		if operationIDStr, ok := operationID.(string); ok {
			builder.WithOperationID(operationIDStr)
		}
	}

	// Adiciona deadline se configurado
	if deadline, ok := ctx.Deadline(); ok {
		builder.Metadata("deadline", deadline)
	}

	return builder
}

// LoggingErrorObserver implementa ErrorObserver para logging
type LoggingErrorObserver struct {
	Logger func(level string, message string, fields map[string]interface{})
}

// NewLoggingErrorObserver cria um observer de logging
func NewLoggingErrorObserver(logger func(string, string, map[string]interface{})) *LoggingErrorObserver {
	return &LoggingErrorObserver{
		Logger: logger,
	}
}

// OnErrorCreated implementa ErrorObserver
func (o *LoggingErrorObserver) OnErrorCreated(err error, context map[string]interface{}) {
	if o.Logger == nil {
		return
	}

	fields := make(map[string]interface{})

	// Adiciona informações do erro
	if domainErr, ok := err.(*DomainError); ok {
		fields["error_code"] = domainErr.Code
		fields["error_type"] = string(domainErr.Type)
		fields["entity"] = domainErr.EntityName

		// Adiciona metadados do erro
		for k, v := range domainErr.Metadata {
			fields[k] = v
		}
	}

	// Adiciona contexto
	for k, v := range context {
		fields[k] = v
	}

	level := "error"
	if domainErr, ok := err.(*DomainError); ok {
		switch domainErr.Type {
		case ErrorTypeValidation, ErrorTypeBadRequest:
			level = "warn"
		case ErrorTypeNotFound:
			level = "info"
		case ErrorTypeInternal, ErrorTypeInfrastructure:
			level = "error"
		default:
			level = "warn"
		}
	}

	o.Logger(level, fmt.Sprintf("Domain error occurred: %s", err.Error()), fields)
}

// MetricsErrorObserver implementa ErrorObserver para métricas
type MetricsErrorObserver struct {
	IncrementCounter func(name string, tags map[string]string)
}

// NewMetricsErrorObserver cria um observer de métricas
func NewMetricsErrorObserver(incrementCounter func(string, map[string]string)) *MetricsErrorObserver {
	return &MetricsErrorObserver{
		IncrementCounter: incrementCounter,
	}
}

// OnErrorCreated implementa ErrorObserver
func (o *MetricsErrorObserver) OnErrorCreated(err error, context map[string]interface{}) {
	if o.IncrementCounter == nil {
		return
	}

	tags := make(map[string]string)

	if domainErr, ok := err.(*DomainError); ok {
		tags["error_type"] = string(domainErr.Type)
		tags["error_code"] = domainErr.Code

		// Adiciona status code como tag
		statusCode := domainErr.StatusCode()
		tags["status_code"] = fmt.Sprintf("%d", statusCode)

		// Adiciona entity se disponível
		if domainErr.EntityName != "" {
			tags["entity"] = domainErr.EntityName
		}
	}

	o.IncrementCounter("domain_errors_total", tags)
}

// Instância global da factory
var defaultFactory = NewDomainErrorFactory()

// GetDefaultFactory retorna a factory padrão
func GetDefaultFactory() *DomainErrorFactory {
	return defaultFactory
}

// SetDefaultFactory define uma nova factory padrão
func SetDefaultFactory(factory *DomainErrorFactory) {
	defaultFactory = factory
}

// GetDefaultFactoryWithOptions retorna uma instância da factory com opções customizadas
func GetDefaultFactoryWithOptions(options ...FactoryOption) *DomainErrorFactory {
	return NewDomainErrorFactory(options...)
}
