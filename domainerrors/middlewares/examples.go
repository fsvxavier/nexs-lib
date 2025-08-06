package middlewares

import (
	"fmt"
	"log"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// LoggingMiddleware middleware básico de logging
type LoggingMiddleware struct {
	name     string
	priority int
	enabled  bool
}

// NewLoggingMiddleware cria um novo middleware de logging
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		name:     "logging_middleware",
		priority: 100,
		enabled:  true,
	}
}

// Name retorna o nome do middleware
func (m *LoggingMiddleware) Name() string {
	return m.name
}

// Handle processa o middleware
func (m *LoggingMiddleware) Handle(ctx *interfaces.MiddlewareContext, next interfaces.NextFunction) error {
	start := time.Now()

	log.Printf("[MIDDLEWARE] Processing error: %s (Type: %s)",
		ctx.Error.Message, ctx.Error.Type)

	// Chama o próximo middleware
	err := next(ctx)

	duration := time.Since(start)
	log.Printf("[MIDDLEWARE] Error processed in %v", duration)

	return err
}

// Type retorna o tipo do middleware
func (m *LoggingMiddleware) Type() interfaces.MiddlewareType {
	return interfaces.MiddlewareTypeLogging
}

// Priority retorna a prioridade de execução
func (m *LoggingMiddleware) Priority() int {
	return m.priority
}

// Enabled indica se o middleware está habilitado
func (m *LoggingMiddleware) Enabled() bool {
	return m.enabled
}

// SetEnabled habilita/desabilita o middleware
func (m *LoggingMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// MetricsMiddleware middleware básico de métricas
type MetricsMiddleware struct {
	name        string
	priority    int
	enabled     bool
	errorCounts map[interfaces.ErrorType]int
}

// NewMetricsMiddleware cria um novo middleware de métricas
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		name:        "metrics_middleware",
		priority:    200,
		enabled:     true,
		errorCounts: make(map[interfaces.ErrorType]int),
	}
}

// Name retorna o nome do middleware
func (m *MetricsMiddleware) Name() string {
	return m.name
}

// Handle processa o middleware
func (m *MetricsMiddleware) Handle(ctx *interfaces.MiddlewareContext, next interfaces.NextFunction) error {
	// Incrementa contador de erros
	m.errorCounts[ctx.Error.Type]++

	// Chama o próximo middleware
	return next(ctx)
}

// Type retorna o tipo do middleware
func (m *MetricsMiddleware) Type() interfaces.MiddlewareType {
	return interfaces.MiddlewareTypeMetrics
}

// Priority retorna a prioridade de execução
func (m *MetricsMiddleware) Priority() int {
	return m.priority
}

// Enabled indica se o middleware está habilitado
func (m *MetricsMiddleware) Enabled() bool {
	return m.enabled
}

// SetEnabled habilita/desabilita o middleware
func (m *MetricsMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// GetErrorCounts retorna os contadores de erro
func (m *MetricsMiddleware) GetErrorCounts() map[interfaces.ErrorType]int {
	counts := make(map[interfaces.ErrorType]int)
	for k, v := range m.errorCounts {
		counts[k] = v
	}
	return counts
}

// PrintStats imprime estatísticas dos erros
func (m *MetricsMiddleware) PrintStats() {
	fmt.Println("=== Error Statistics ===")
	for errorType, count := range m.errorCounts {
		fmt.Printf("%s: %d\n", errorType, count)
	}
	fmt.Println("========================")
}

// EnrichmentMiddleware middleware para enriquecimento de erros
type EnrichmentMiddleware struct {
	name     string
	priority int
	enabled  bool
}

// NewEnrichmentMiddleware cria um novo middleware de enriquecimento
func NewEnrichmentMiddleware() *EnrichmentMiddleware {
	return &EnrichmentMiddleware{
		name:     "enrichment_middleware",
		priority: 50,
		enabled:  true,
	}
}

// Name retorna o nome do middleware
func (m *EnrichmentMiddleware) Name() string {
	return m.name
}

// Handle processa o middleware
func (m *EnrichmentMiddleware) Handle(ctx *interfaces.MiddlewareContext, next interfaces.NextFunction) error {
	// Enriquece o erro com dados contextuais
	if ctx.Error.Metadata == nil {
		ctx.Error.Metadata = make(map[string]interface{})
	}

	ctx.Error.Metadata["processed_at"] = time.Now()
	ctx.Error.Metadata["middleware_chain"] = "enrichment"

	if ctx.TraceID != "" {
		ctx.Error.Metadata["trace_id"] = ctx.TraceID
	}

	if ctx.UserID != "" {
		ctx.Error.Metadata["user_id"] = ctx.UserID
	}

	// Chama o próximo middleware
	return next(ctx)
}

// Type retorna o tipo do middleware
func (m *EnrichmentMiddleware) Type() interfaces.MiddlewareType {
	return interfaces.MiddlewareTypeEnrichment
}

// Priority retorna a prioridade de execução
func (m *EnrichmentMiddleware) Priority() int {
	return m.priority
}

// Enabled indica se o middleware está habilitado
func (m *EnrichmentMiddleware) Enabled() bool {
	return m.enabled
}

// SetEnabled habilita/desabilita o middleware
func (m *EnrichmentMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}
