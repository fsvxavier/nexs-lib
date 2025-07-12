// Package main demonstra o uso de middleware para transformação de logs
// Este exemplo mostra como implementar e usar middleware para enriquecer,
// filtrar e transformar logs antes de serem processados.
package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// CorrelationMiddleware adiciona ID de correlação automático
type CorrelationMiddleware struct {
	correlationID string
}

func NewCorrelationMiddleware(correlationID string) *CorrelationMiddleware {
	return &CorrelationMiddleware{correlationID: correlationID}
}

func (cm *CorrelationMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Adiciona correlation_id a todos os logs
	entry.Fields = append(entry.Fields, interfaces.String("correlation_id", cm.correlationID))
	return entry
}

// SensitiveDataMiddleware remove dados sensíveis dos logs
type SensitiveDataMiddleware struct {
	sensitiveFields []string
}

func NewSensitiveDataMiddleware(fields []string) *SensitiveDataMiddleware {
	return &SensitiveDataMiddleware{sensitiveFields: fields}
}

func (sdm *SensitiveDataMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Mascara dados sensíveis
	for i, field := range entry.Fields {
		for _, sensitive := range sdm.sensitiveFields {
			if strings.Contains(strings.ToLower(field.Key), sensitive) {
				entry.Fields[i].Value = "[MASKED]"
			}
		}
	}

	// Mascara dados sensíveis na mensagem
	message := entry.Message
	for _, sensitive := range sdm.sensitiveFields {
		if strings.Contains(strings.ToLower(message), sensitive) {
			message = strings.ReplaceAll(message, sensitive, "[MASKED]")
		}
	}
	entry.Message = message

	return entry
}

// PerformanceMiddleware adiciona métricas de performance
type PerformanceMiddleware struct {
	startTime time.Time
}

func NewPerformanceMiddleware() *PerformanceMiddleware {
	return &PerformanceMiddleware{startTime: time.Now()}
}

func (pm *PerformanceMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Adiciona informações de performance
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	entry.Fields = append(entry.Fields,
		interfaces.Duration("uptime", time.Since(pm.startTime)),
		interfaces.Int64("memory_alloc", int64(memStats.Alloc)),
		interfaces.Int64("memory_sys", int64(memStats.Sys)),
		interfaces.Int("goroutines", runtime.NumGoroutine()),
	)

	return entry
}

// EnrichmentMiddleware enriquece logs com informações do ambiente
type EnrichmentMiddleware struct {
	environment map[string]string
}

func NewEnrichmentMiddleware(env map[string]string) *EnrichmentMiddleware {
	return &EnrichmentMiddleware{environment: env}
}

func (em *EnrichmentMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Adiciona informações do ambiente
	for key, value := range em.environment {
		entry.Fields = append(entry.Fields, interfaces.String(key, value))
	}

	// Adiciona informações do caller se disponível
	if entry.Caller != nil {
		entry.Fields = append(entry.Fields,
			interfaces.String("source_file", entry.Caller.File),
			interfaces.Int("source_line", entry.Caller.Line),
			interfaces.String("source_function", entry.Caller.Function),
		)
	}

	return entry
}

// FilterMiddleware filtra logs baseado em critérios
type FilterMiddleware struct {
	excludePatterns []string
	minLevel        interfaces.Level
}

func NewFilterMiddleware(excludePatterns []string, minLevel interfaces.Level) *FilterMiddleware {
	return &FilterMiddleware{
		excludePatterns: excludePatterns,
		minLevel:        minLevel,
	}
}

func (fm *FilterMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Filtra por nível
	if entry.Level < fm.minLevel {
		return nil // Log será descartado
	}

	// Filtra por padrões na mensagem
	for _, pattern := range fm.excludePatterns {
		if strings.Contains(strings.ToLower(entry.Message), pattern) {
			return nil // Log será descartado
		}
	}

	return entry
}

// TransformationMiddleware transforma campos baseado em regras
type TransformationMiddleware struct {
	fieldTransforms map[string]func(interface{}) interface{}
}

func NewTransformationMiddleware() *TransformationMiddleware {
	return &TransformationMiddleware{
		fieldTransforms: make(map[string]func(interface{}) interface{}),
	}
}

func (tm *TransformationMiddleware) AddTransform(fieldName string, transform func(interface{}) interface{}) {
	tm.fieldTransforms[fieldName] = transform
}

func (tm *TransformationMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	// Aplica transformações nos campos
	for i, field := range entry.Fields {
		if transform, exists := tm.fieldTransforms[field.Key]; exists {
			entry.Fields[i].Value = transform(field.Value)
		}
	}

	return entry
}

func main() {
	fmt.Println("=== Logger v2 - Middleware System ===")

	// 1. Configuração básica
	config := logger.DefaultConfig()
	config.ServiceName = "middleware-example"
	config.ServiceVersion = "1.0.0"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.DebugLevel
	config.AddCaller = true

	// 2. Criação dos middlewares
	correlationMw := NewCorrelationMiddleware("corr_123456")

	sensitiveMw := NewSensitiveDataMiddleware([]string{
		"password", "token", "secret", "key", "credential",
	})

	performanceMw := NewPerformanceMiddleware()

	enrichmentMw := NewEnrichmentMiddleware(map[string]string{
		"environment": "production",
		"region":      "us-east-1",
		"version":     "v2.1.0",
		"hostname":    "app-server-01",
	})

	filterMw := NewFilterMiddleware([]string{
		"health check", "ping", "heartbeat",
	}, interfaces.InfoLevel)

	transformMw := NewTransformationMiddleware()
	// Transformação: converte duração para milissegundos
	transformMw.AddTransform("duration", func(v interface{}) interface{} {
		if duration, ok := v.(time.Duration); ok {
			return duration.Milliseconds()
		}
		return v
	})
	// Transformação: normaliza emails para lowercase
	transformMw.AddTransform("email", func(v interface{}) interface{} {
		if email, ok := v.(string); ok {
			return strings.ToLower(email)
		}
		return v
	})

	// 3. Configuração dos middlewares no config
	config.Middlewares = []interfaces.Middleware{
		correlationMw,
		sensitiveMw,
		performanceMw,
		enrichmentMw,
		filterMw,
		transformMw,
	}

	// 4. Criação da factory e logger
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	logger, err := factory.CreateLogger("middleware", config)
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}

	ctx := context.Background()

	// 5. Demonstração de funcionamento dos middlewares
	fmt.Println("\n--- Demonstração de Middlewares ---")

	// Log normal - passará por todos os middlewares
	logger.Info(ctx, "Usuário autenticado",
		interfaces.String("user_id", "user123"),
		interfaces.String("email", "USER@EXAMPLE.COM"),        // Será transformado para lowercase
		interfaces.Duration("duration", 150*time.Millisecond), // Será convertido para milissegundos
	)

	// Log com dados sensíveis - será mascarado
	logger.Warn(ctx, "Falha na autenticação",
		interfaces.String("user_id", "user456"),
		interfaces.String("password", "secret123"),  // Será mascarado
		interfaces.String("api_token", "abc123xyz"), // Será mascarado
	)

	// Log que será filtrado - não aparecerá
	logger.Debug(ctx, "Health check realizado com sucesso")

	// Log de erro com informações completas
	logger.Error(ctx, "Erro no processamento de pagamento",
		interfaces.String("payment_id", "pay_789"),
		interfaces.Float64("amount", 299.99),
		interfaces.String("error_code", "INSUFFICIENT_FUNDS"),
		interfaces.Duration("processing_time", 2*time.Second),
	)

	// 6. Demonstração de middleware condicional
	fmt.Println("\n--- Middleware Condicional ---")
	demonstrateConditionalMiddleware(factory)

	// 7. Demonstração de middleware em cadeia
	fmt.Println("\n--- Cadeia de Middlewares ---")
	demonstrateMiddlewareChain(factory)

	// 8. Performance com middlewares
	fmt.Println("\n--- Impacto de Performance ---")
	measureMiddlewarePerformance(factory)

	// 9. Cleanup
	if err := logger.Flush(); err != nil {
		fmt.Printf("Erro ao fazer flush: %v\n", err)
	}

	if err := logger.Close(); err != nil {
		fmt.Printf("Erro ao fechar logger: %v\n", err)
	}

	fmt.Println("\n=== Middleware System Concluído ===")
}

// demonstrateConditionalMiddleware mostra middleware que age condicionalmente
func demonstrateConditionalMiddleware(factory *logger.Factory) {
	// Middleware que só age em logs de erro
	errorEnrichmentMw := &ConditionalMiddleware{
		condition: func(entry *interfaces.Entry) bool {
			return entry.Level >= interfaces.ErrorLevel
		},
		middleware: NewEnrichmentMiddleware(map[string]string{
			"alert_level": "high",
			"notify_ops":  "true",
		}),
	}

	config := logger.DefaultConfig()
	config.ServiceName = "conditional-middleware"
	config.Format = interfaces.JSONFormat
	config.Middlewares = []interfaces.Middleware{errorEnrichmentMw}

	logger, _ := factory.CreateLogger("conditional", config)
	ctx := context.Background()

	// Log normal - não será enriquecido
	logger.Info(ctx, "Operação normal")

	// Log de erro - será enriquecido
	logger.Error(ctx, "Erro crítico detectado")
}

// ConditionalMiddleware executa middleware apenas se condição for atendida
type ConditionalMiddleware struct {
	condition  func(*interfaces.Entry) bool
	middleware interfaces.Middleware
}

func (cm *ConditionalMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	if cm.condition(entry) {
		return cm.middleware.Process(entry)
	}
	return entry
}

// demonstrateMiddlewareChain mostra como middlewares trabalham em cadeia
func demonstrateMiddlewareChain(factory *logger.Factory) {
	// Middleware que adiciona timestamp personalizado
	timestampMw := &TimestampMiddleware{}

	// Middleware que adiciona índice sequencial
	sequenceMw := &SequenceMiddleware{}

	config := logger.DefaultConfig()
	config.ServiceName = "chain-middleware"
	config.Format = interfaces.JSONFormat
	config.Middlewares = []interfaces.Middleware{
		timestampMw,
		sequenceMw,
		NewCorrelationMiddleware("chain_test"),
	}

	logger, _ := factory.CreateLogger("chain", config)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		logger.Info(ctx, fmt.Sprintf("Mensagem em cadeia %d", i))
	}
}

// TimestampMiddleware adiciona timestamp customizado
type TimestampMiddleware struct{}

func (tm *TimestampMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	entry.Fields = append(entry.Fields,
		interfaces.String("custom_timestamp", time.Now().Format("2006-01-02T15:04:05.000Z")),
	)
	return entry
}

// SequenceMiddleware adiciona número sequencial
type SequenceMiddleware struct {
	counter int64
}

func (sm *SequenceMiddleware) Process(entry *interfaces.Entry) *interfaces.Entry {
	sm.counter++
	entry.Fields = append(entry.Fields,
		interfaces.Int64("sequence_number", sm.counter),
	)
	return entry
}

// measureMiddlewarePerformance mede o impacto dos middlewares na performance
func measureMiddlewarePerformance(factory *logger.Factory) {
	const iterations = 1000

	// Logger sem middlewares
	configNoMw := logger.DefaultConfig()
	configNoMw.ServiceName = "no-middleware"
	loggerNoMw, _ := factory.CreateLogger("no-mw", configNoMw)

	// Logger com middlewares
	configWithMw := logger.DefaultConfig()
	configWithMw.ServiceName = "with-middleware"
	configWithMw.Middlewares = []interfaces.Middleware{
		NewCorrelationMiddleware("perf_test"),
		NewPerformanceMiddleware(),
		NewEnrichmentMiddleware(map[string]string{"test": "performance"}),
	}
	loggerWithMw, _ := factory.CreateLogger("with-mw", configWithMw)

	ctx := context.Background()

	// Teste sem middlewares
	start := time.Now()
	for i := 0; i < iterations; i++ {
		loggerNoMw.Info(ctx, "Test message",
			interfaces.Int("iteration", i),
		)
	}
	loggerNoMw.Flush()
	durationNoMw := time.Since(start)

	// Teste com middlewares
	start = time.Now()
	for i := 0; i < iterations; i++ {
		loggerWithMw.Info(ctx, "Test message",
			interfaces.Int("iteration", i),
		)
	}
	loggerWithMw.Flush()
	durationWithMw := time.Since(start)

	fmt.Printf("Performance Impact:\n")
	fmt.Printf("  Sem middlewares: %v\n", durationNoMw)
	fmt.Printf("  Com middlewares: %v\n", durationWithMw)
	fmt.Printf("  Overhead: %v (%.2f%%)\n",
		durationWithMw-durationNoMw,
		float64(durationWithMw-durationNoMw)/float64(durationNoMw)*100,
	)
}
