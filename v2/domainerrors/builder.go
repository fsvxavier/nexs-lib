package domainerrors

import (
	"context"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// ErrorBuilder implementa interfaces.ErrorBuilder para construção fluente de erros.
type ErrorBuilder struct {
	error *DomainError
}

// NewBuilder cria um novo builder de erros.
func NewBuilder() interfaces.ErrorBuilder {
	return &ErrorBuilder{
		error: newDomainError(),
	}
}

// WithCode define o código do erro.
func (b *ErrorBuilder) WithCode(code string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.code = code
	return b
}

// WithMessage define a mensagem do erro.
func (b *ErrorBuilder) WithMessage(message string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.message = message
	return b
}

// WithType define o tipo do erro.
func (b *ErrorBuilder) WithType(errorType string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.errorType = types.ErrorType(errorType)

	// Configura valores padrão baseados no tipo
	if b.error.errorType.IsValid() {
		if b.error.severity == types.SeverityMedium { // Se ainda é o padrão
			b.error.severity = b.error.errorType.DefaultSeverity()
		}
		if b.error.category == "" {
			b.error.category = b.error.errorType.Category()
		}
		b.error.retryable = b.error.errorType.IsRetryable()
		b.error.temporary = b.error.errorType.IsTemporary()
	}

	return b
}

// WithSeverity define a severidade do erro.
func (b *ErrorBuilder) WithSeverity(severity interfaces.Severity) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.severity = types.ErrorSeverity(severity)
	return b
}

// WithCategory define a categoria do erro.
func (b *ErrorBuilder) WithCategory(category interfaces.Category) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.category = string(category)
	return b
}

// WithDetail adiciona um detalhe específico.
func (b *ErrorBuilder) WithDetail(key string, value interface{}) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.details[key] = value
	return b
}

// WithDetails adiciona múltiplos detalhes.
func (b *ErrorBuilder) WithDetails(details map[string]interface{}) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	for k, v := range details {
		b.error.details[k] = v
	}
	return b
}

// WithMetadata adiciona metadados.
func (b *ErrorBuilder) WithMetadata(metadata map[string]interface{}) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	for k, v := range metadata {
		b.error.metadata[k] = v
	}
	return b
}

// WithTag adiciona uma tag.
func (b *ErrorBuilder) WithTag(tag string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	if !b.error.hasTag(tag) {
		b.error.tags = append(b.error.tags, tag)
	}
	return b
}

// WithTags adiciona múltiplas tags.
func (b *ErrorBuilder) WithTags(tags []string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	for _, tag := range tags {
		if !b.error.hasTag(tag) {
			b.error.tags = append(b.error.tags, tag)
		}
	}
	return b
}

// WithCause define o erro causa.
func (b *ErrorBuilder) WithCause(err error) interfaces.ErrorBuilder {
	if err != nil {
		b.error.mu.Lock()
		defer b.error.mu.Unlock()
		b.error.cause = err
		b.error.captureStackTrace("wrapped error", 2)

		// Herda metadados se for um DomainError
		if de, ok := err.(*DomainError); ok {
			b.error.inheritFromDomainError(de)
		}
	}
	return b
}

// WithStatusCode define o código de status HTTP.
func (b *ErrorBuilder) WithStatusCode(code int) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.statusCode = code
	return b
}

// WithHeader adiciona um header HTTP.
func (b *ErrorBuilder) WithHeader(key, value string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	b.error.headers[key] = value
	return b
}

// WithHeaders adiciona múltiplos headers HTTP.
func (b *ErrorBuilder) WithHeaders(headers map[string]string) interfaces.ErrorBuilder {
	b.error.mu.Lock()
	defer b.error.mu.Unlock()
	for k, v := range headers {
		b.error.headers[k] = v
	}
	return b
}

// WithContext adiciona informações de contexto.
func (b *ErrorBuilder) WithContext(ctx context.Context) interfaces.ErrorBuilder {
	b.error.WithContext(ctx)
	return b
}

// Build constrói o erro final.
func (b *ErrorBuilder) Build() interfaces.DomainErrorInterface {
	// Aplica valores padrão apenas se o tipo foi definido mas outros campos estão vazios
	if b.error.errorType != "" {
		// Se o tipo foi definido explicitamente, aplica padrões apenas se necessário
		if b.error.message == "" {
			b.error.message = "Unknown error"
		}
		if b.error.code == "" {
			b.error.code = "E999"
		}
	} else if b.error.code != "" || b.error.message != "" {
		// Se código ou mensagem foram definidos mas não o tipo, aplica tipo padrão
		if b.error.errorType == "" {
			b.error.errorType = types.ErrorTypeInternal
		}
	}

	return b.error
}

// Métodos de conveniência para construção rápida

// BuildValidationError constrói um erro de validação rapidamente.
func (b *ErrorBuilder) BuildValidationError(fields map[string][]string) interfaces.ValidationErrorInterface {
	b.WithType(string(types.ErrorTypeValidation))
	domainErr := b.Build()

	return &ValidationError{
		DomainError:     domainErr.(*DomainError),
		ValidatedFields: fields,
	}
}

// BuildNotFoundError constrói um erro de não encontrado rapidamente.
func (b *ErrorBuilder) BuildNotFoundError(entity, id string) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeNotFound)).
		WithCode("E002").
		WithMessage("Resource not found").
		WithDetail("entity", entity).
		WithDetail("id", id).
		Build()
}

// BuildBusinessError constrói um erro de negócio rapidamente.
func (b *ErrorBuilder) BuildBusinessError(rule string) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeBusinessRule)).
		WithCode("E004").
		WithMessage("Business rule violation").
		WithDetail("rule", rule).
		Build()
}

// BuildInternalError constrói um erro interno rapidamente.
func (b *ErrorBuilder) BuildInternalError(cause error) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeInternal)).
		WithCode("E007").
		WithMessage("Internal server error").
		WithCause(cause).
		Build()
}

// BuildTimeoutError constrói um erro de timeout rapidamente.
func (b *ErrorBuilder) BuildTimeoutError(operation string) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeTimeout)).
		WithCode("E009").
		WithMessage("Operation timeout").
		WithDetail("operation", operation).
		Build()
}

// BuildRateLimitError constrói um erro de rate limit rapidamente.
func (b *ErrorBuilder) BuildRateLimitError(limit int, window string) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeRateLimit)).
		WithCode("E010").
		WithMessage("Rate limit exceeded").
		WithDetail("limit", limit).
		WithDetail("window", window).
		Build()
}

// BuildCircuitBreakerError constrói um erro de circuit breaker rapidamente.
func (b *ErrorBuilder) BuildCircuitBreakerError(service string) interfaces.DomainErrorInterface {
	return b.WithType(string(types.ErrorTypeCircuitBreaker)).
		WithCode("E013").
		WithMessage("Circuit breaker is open").
		WithDetail("service", service).
		Build()
}
