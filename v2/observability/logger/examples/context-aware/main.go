// Package main demonstra o uso de context-aware logging
// Este exemplo mostra como extrair automaticamente informações
// de contexto como trace_id, span_id, user_id e request_id.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// Simulação de contexto com informações de tracing
type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	SessionIDKey contextKey = "session_id"
)

func main() {
	fmt.Println("=== Logger v2 - Context-Aware Logging ===")

	// 1. Configuração para logging com contexto
	config := logger.DefaultConfig()
	config.ServiceName = "context-aware-example"
	config.ServiceVersion = "1.0.0"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.DebugLevel
	config.AddCaller = true

	// 2. Criação da factory e logger
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	baseLogger, err := factory.CreateLogger("context-aware", config)
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}

	// 3. Criação de contexto com informações de tracing
	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, generateTraceID())
	ctx = context.WithValue(ctx, SpanIDKey, generateSpanID())
	ctx = context.WithValue(ctx, UserIDKey, "user_12345")
	ctx = context.WithValue(ctx, RequestIDKey, "req_"+generateID())
	ctx = context.WithValue(ctx, SessionIDKey, "sess_"+generateID())

	// 4. Logger com extração automática de contexto
	fmt.Println("\n--- Logger com Context ---")
	contextLogger := baseLogger.
		WithTraceID(getTraceID(ctx)).
		WithSpanID(getSpanID(ctx)).
		WithFields(
			interfaces.String("user_id", getUserID(ctx)),
			interfaces.String("request_id", getRequestID(ctx)),
			interfaces.String("session_id", getSessionID(ctx)),
		)

	contextLogger.Info(ctx, "Operação iniciada")

	// 5. Simulação de operação com sub-spans
	fmt.Println("\n--- Sub-operações com Contexto ---")
	simulateAuthenticationFlow(ctx, contextLogger)
	simulateBusinessLogic(ctx, contextLogger)
	simulateDataAccess(ctx, contextLogger)

	// 6. Propagação de contexto através de middlewares
	fmt.Println("\n--- Middleware Pattern ---")
	simulateHTTPRequest(ctx, baseLogger)

	// 7. Logging com contexto em goroutines
	fmt.Println("\n--- Concurrent Operations ---")
	simulateConcurrentOperations(ctx, baseLogger)

	// 8. Cleanup
	if err := baseLogger.Flush(); err != nil {
		fmt.Printf("Erro ao fazer flush: %v\n", err)
	}

	if err := baseLogger.Close(); err != nil {
		fmt.Printf("Erro ao fechar logger: %v\n", err)
	}

	fmt.Println("\n=== Context-Aware Logging Concluído ===")
}

// simulateAuthenticationFlow simula um fluxo de autenticação
func simulateAuthenticationFlow(ctx context.Context, logger interfaces.Logger) {
	// Criação de sub-span para autenticação
	authCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
	authLogger := logger.WithSpanID(getSpanID(authCtx)).WithFields(
		interfaces.String("operation", "authentication"),
		interfaces.String("method", "jwt_validation"),
	)

	authLogger.Debug(authCtx, "Iniciando validação de token")

	// Simulação de tempo de processamento
	time.Sleep(10 * time.Millisecond)

	authLogger.Info(authCtx, "Token validado com sucesso",
		interfaces.String("token_type", "bearer"),
		interfaces.Int64("expires_in", 3600),
		interfaces.String("scope", "read write"),
	)
}

// simulateBusinessLogic simula lógica de negócio
func simulateBusinessLogic(ctx context.Context, logger interfaces.Logger) {
	businessCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
	businessLogger := logger.WithSpanID(getSpanID(businessCtx)).WithFields(
		interfaces.String("operation", "business_logic"),
		interfaces.String("service", "order_processor"),
	)

	businessLogger.Debug(businessCtx, "Processando regras de negócio")

	// Simulação de processamento
	time.Sleep(25 * time.Millisecond)

	businessLogger.Info(businessCtx, "Pedido processado",
		interfaces.String("order_id", "ord_"+generateID()),
		interfaces.Float64("amount", 299.99),
		interfaces.String("status", "approved"),
	)
}

// simulateDataAccess simula acesso a dados
func simulateDataAccess(ctx context.Context, logger interfaces.Logger) {
	dataCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
	dataLogger := logger.WithSpanID(getSpanID(dataCtx)).WithFields(
		interfaces.String("operation", "data_access"),
		interfaces.String("database", "postgresql"),
	)

	dataLogger.Debug(dataCtx, "Executando consulta no banco")

	start := time.Now()
	// Simulação de query
	time.Sleep(15 * time.Millisecond)
	duration := time.Since(start)

	dataLogger.Info(dataCtx, "Consulta executada",
		interfaces.String("query", "SELECT * FROM orders WHERE user_id = $1"),
		interfaces.Duration("duration", duration),
		interfaces.Int("rows_returned", 5),
	)
}

// simulateHTTPRequest simula o processamento de uma requisição HTTP
func simulateHTTPRequest(ctx context.Context, logger interfaces.Logger) {
	// Middleware 1: Request ID
	ctx = context.WithValue(ctx, RequestIDKey, "req_"+generateID())
	requestLogger := logger.WithFields(
		interfaces.String("request_id", getRequestID(ctx)),
		interfaces.String("method", "POST"),
		interfaces.String("path", "/api/v1/orders"),
	)

	requestLogger.Info(ctx, "Requisição recebida")

	// Middleware 2: Authentication
	ctx = context.WithValue(ctx, UserIDKey, "user_67890")
	authLogger := requestLogger.WithFields(
		interfaces.String("user_id", getUserID(ctx)),
	)

	authLogger.Debug(ctx, "Usuário autenticado")

	// Business Logic
	authLogger.Info(ctx, "Processamento concluído",
		interfaces.Int("status_code", 201),
		interfaces.Duration("response_time", 89*time.Millisecond),
	)
}

// simulateConcurrentOperations simula operações concorrentes
func simulateConcurrentOperations(ctx context.Context, logger interfaces.Logger) {
	done := make(chan bool, 3)

	// Goroutine 1: Worker A
	go func() {
		workerCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
		workerLogger := logger.WithSpanID(getSpanID(workerCtx)).WithFields(
			interfaces.String("worker", "worker_a"),
		)

		workerLogger.Info(workerCtx, "Worker A iniciado")
		time.Sleep(50 * time.Millisecond)
		workerLogger.Info(workerCtx, "Worker A concluído")
		done <- true
	}()

	// Goroutine 2: Worker B
	go func() {
		workerCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
		workerLogger := logger.WithSpanID(getSpanID(workerCtx)).WithFields(
			interfaces.String("worker", "worker_b"),
		)

		workerLogger.Info(workerCtx, "Worker B iniciado")
		time.Sleep(30 * time.Millisecond)
		workerLogger.Info(workerCtx, "Worker B concluído")
		done <- true
	}()

	// Goroutine 3: Worker C
	go func() {
		workerCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
		workerLogger := logger.WithSpanID(getSpanID(workerCtx)).WithFields(
			interfaces.String("worker", "worker_c"),
		)

		workerLogger.Info(workerCtx, "Worker C iniciado")
		time.Sleep(20 * time.Millisecond)
		workerLogger.Info(workerCtx, "Worker C concluído")
		done <- true
	}()

	// Aguarda todos os workers
	for i := 0; i < 3; i++ {
		<-done
	}

	logger.Info(ctx, "Todos os workers concluídos")
}

// Helper functions para extração de contexto
func getTraceID(ctx context.Context) string {
	if val := ctx.Value(TraceIDKey); val != nil {
		return val.(string)
	}
	return ""
}

func getSpanID(ctx context.Context) string {
	if val := ctx.Value(SpanIDKey); val != nil {
		return val.(string)
	}
	return ""
}

func getUserID(ctx context.Context) string {
	if val := ctx.Value(UserIDKey); val != nil {
		return val.(string)
	}
	return ""
}

func getRequestID(ctx context.Context) string {
	if val := ctx.Value(RequestIDKey); val != nil {
		return val.(string)
	}
	return ""
}

func getSessionID(ctx context.Context) string {
	if val := ctx.Value(SessionIDKey); val != nil {
		return val.(string)
	}
	return ""
}

// Utility functions para geração de IDs
func generateTraceID() string {
	return fmt.Sprintf("trace_%016x", rand.Uint64())
}

func generateSpanID() string {
	return fmt.Sprintf("span_%08x", rand.Uint32())
}

func generateID() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}
