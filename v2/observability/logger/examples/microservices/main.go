// Package main demonstra o uso do logger em arquitetura de microserviços
// Este exemplo mostra como implementar logging distribuído, correlação
// entre serviços e observabilidade em sistemas distribuídos.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// ServiceInfo representa informações de um microserviço
type ServiceInfo struct {
	Name        string
	Version     string
	Environment string
	Instance    string
	Port        int
}

// TraceContext representa contexto de tracing distribuído
type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	UserID    string
	RequestID string
}

// MicroserviceLogger encapsula logger específico para microserviços
type MicroserviceLogger struct {
	logger  interfaces.Logger
	service ServiceInfo
}

func NewMicroserviceLogger(service ServiceInfo, factory *logger.Factory) (*MicroserviceLogger, error) {
	config := logger.ProductionConfig()
	config.ServiceName = service.Name
	config.ServiceVersion = service.Version
	config.Environment = service.Environment

	// Campos globais específicos do serviço
	config.GlobalFields = map[string]interface{}{
		"service_instance": service.Instance,
		"service_port":     service.Port,
		"service_type":     "microservice",
	}

	// Configuração otimizada para produção distribuída
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    4096,
		FlushInterval: 100 * time.Millisecond,
		Workers:       2,
		DropOnFull:    false,
	}

	baseLogger, err := factory.CreateLogger(service.Name, config)
	if err != nil {
		return nil, err
	}

	return &MicroserviceLogger{
		logger:  baseLogger,
		service: service,
	}, nil
}

func (ml *MicroserviceLogger) WithTrace(trace TraceContext) interfaces.Logger {
	return ml.logger.
		WithTraceID(trace.TraceID).
		WithSpanID(trace.SpanID).
		WithFields(
			interfaces.String("parent_span_id", trace.ParentID),
			interfaces.String("user_id", trace.UserID),
			interfaces.String("request_id", trace.RequestID),
		)
}

func main() {
	fmt.Println("=== Logger v2 - Microservices Architecture ===")

	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	// 1. Simulação de múltiplos microserviços
	fmt.Println("\n--- Inicializando Microserviços ---")
	services := initializeMicroservices(factory)

	// 2. Simulação de requisição distribuída
	fmt.Println("\n--- Requisição Distribuída ---")
	simulateDistributedRequest(services)

	// 3. Simulação de comunicação inter-serviços
	fmt.Println("\n--- Comunicação Inter-Serviços ---")
	simulateServiceCommunication(services)

	// 4. Simulação de falhas e recuperação
	fmt.Println("\n--- Falhas e Recuperação ---")
	simulateFailureScenarios(services)

	// 5. Agregação de logs e métricas
	fmt.Println("\n--- Agregação de Logs ---")
	simulateLogAggregation(services)

	// 6. Cleanup de todos os serviços
	fmt.Println("\n--- Finalizando Serviços ---")
	for _, service := range services {
		service.logger.Flush()
		service.logger.Close()
	}

	fmt.Println("\n=== Microservices Architecture Concluído ===")
}

// initializeMicroservices inicializa múltiplos microserviços
func initializeMicroservices(factory *logger.Factory) map[string]*MicroserviceLogger {
	servicesConfig := []ServiceInfo{
		{
			Name:        "api-gateway",
			Version:     "v1.2.0",
			Environment: "production",
			Instance:    "api-gw-01",
			Port:        8080,
		},
		{
			Name:        "user-service",
			Version:     "v2.1.0",
			Environment: "production",
			Instance:    "user-svc-01",
			Port:        8081,
		},
		{
			Name:        "order-service",
			Version:     "v1.5.0",
			Environment: "production",
			Instance:    "order-svc-01",
			Port:        8082,
		},
		{
			Name:        "payment-service",
			Version:     "v3.0.0",
			Environment: "production",
			Instance:    "payment-svc-01",
			Port:        8083,
		},
		{
			Name:        "notification-service",
			Version:     "v1.1.0",
			Environment: "production",
			Instance:    "notif-svc-01",
			Port:        8084,
		},
	}

	services := make(map[string]*MicroserviceLogger)

	for _, serviceInfo := range servicesConfig {
		microLogger, err := NewMicroserviceLogger(serviceInfo, factory)
		if err != nil {
			log.Printf("Erro ao criar logger para %s: %v", serviceInfo.Name, err)
			continue
		}

		services[serviceInfo.Name] = microLogger

		// Log de inicialização do serviço
		ctx := context.Background()
		microLogger.logger.Info(ctx, "Microserviço inicializado",
			interfaces.String("service", serviceInfo.Name),
			interfaces.String("version", serviceInfo.Version),
			interfaces.String("instance", serviceInfo.Instance),
			interfaces.Int("port", serviceInfo.Port),
		)
	}

	return services
}

// simulateDistributedRequest simula uma requisição que passa por múltiplos serviços
func simulateDistributedRequest(services map[string]*MicroserviceLogger) {
	// Criação do contexto de trace distribuído
	trace := TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		UserID:    "user_12345",
		RequestID: "req_" + generateID(),
	}

	ctx := context.Background()

	// 1. API Gateway recebe a requisição
	gatewayLogger := services["api-gateway"].WithTrace(trace)
	gatewayLogger.Info(ctx, "Requisição recebida",
		interfaces.String("method", "POST"),
		interfaces.String("path", "/api/v1/orders"),
		interfaces.String("client_ip", "192.168.1.100"),
		interfaces.String("user_agent", "mobile-app/2.1.0"),
	)

	// 2. Gateway chama User Service para autenticação
	userTrace := trace
	userTrace.SpanID = generateSpanID()
	userTrace.ParentID = trace.SpanID

	userLogger := services["user-service"].WithTrace(userTrace)
	userLogger.Info(ctx, "Validando autenticação",
		interfaces.String("operation", "authenticate"),
		interfaces.String("auth_method", "jwt"),
	)

	// Simulação de processamento
	time.Sleep(50 * time.Millisecond)

	userLogger.Info(ctx, "Usuário autenticado",
		interfaces.String("user_role", "premium"),
		interfaces.Bool("email_verified", true),
		interfaces.Duration("auth_time", 45*time.Millisecond),
	)

	// 3. Gateway chama Order Service
	orderTrace := trace
	orderTrace.SpanID = generateSpanID()
	orderTrace.ParentID = trace.SpanID

	orderLogger := services["order-service"].WithTrace(orderTrace)
	orderLogger.Info(ctx, "Processando pedido",
		interfaces.String("operation", "create_order"),
		interfaces.Float64("order_value", 299.99),
		interfaces.Int("item_count", 3),
	)

	// 4. Order Service chama Payment Service
	paymentTrace := orderTrace
	paymentTrace.SpanID = generateSpanID()
	paymentTrace.ParentID = orderTrace.SpanID

	paymentLogger := services["payment-service"].WithTrace(paymentTrace)
	paymentLogger.Info(ctx, "Processando pagamento",
		interfaces.String("operation", "process_payment"),
		interfaces.String("payment_method", "credit_card"),
		interfaces.Float64("amount", 299.99),
		interfaces.String("currency", "BRL"),
	)

	time.Sleep(200 * time.Millisecond)

	paymentLogger.Info(ctx, "Pagamento aprovado",
		interfaces.String("payment_id", "pay_"+generateID()),
		interfaces.String("status", "approved"),
		interfaces.String("authorization_code", "AUTH123456"),
	)

	// 5. Notificação assíncrona
	notifTrace := trace
	notifTrace.SpanID = generateSpanID()
	notifTrace.ParentID = trace.SpanID

	notifLogger := services["notification-service"].WithTrace(notifTrace)
	notifLogger.Info(ctx, "Enviando notificação",
		interfaces.String("operation", "send_notification"),
		interfaces.String("type", "order_confirmation"),
		interfaces.String("channel", "email"),
	)

	// 6. Gateway responde ao cliente
	gatewayLogger.Info(ctx, "Requisição processada com sucesso",
		interfaces.String("order_id", "ord_"+generateID()),
		interfaces.Int("status_code", 201),
		interfaces.Duration("total_time", 350*time.Millisecond),
	)
}

// simulateServiceCommunication simula comunicação assíncrona entre serviços
func simulateServiceCommunication(services map[string]*MicroserviceLogger) {
	var wg sync.WaitGroup

	// Event: Usuário criado
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateUserCreatedEvent(services)
	}()

	// Event: Pedido cancelado
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateOrderCancelledEvent(services)
	}()

	// Event: Pagamento falhado
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulatePaymentFailedEvent(services)
	}()

	wg.Wait()
}

func simulateUserCreatedEvent(services map[string]*MicroserviceLogger) {
	trace := TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		UserID:    "user_67890",
		RequestID: "req_" + generateID(),
	}

	ctx := context.Background()

	// User Service publica evento
	userLogger := services["user-service"].WithTrace(trace)
	userLogger.Info(ctx, "Evento publicado",
		interfaces.String("event_type", "user.created"),
		interfaces.String("event_id", "evt_"+generateID()),
		interfaces.String("user_email", "newuser@example.com"),
	)

	// Notification Service processa evento
	notifTrace := trace
	notifTrace.SpanID = generateSpanID()
	notifTrace.ParentID = trace.SpanID

	notifLogger := services["notification-service"].WithTrace(notifTrace)
	notifLogger.Info(ctx, "Evento processado",
		interfaces.String("event_type", "user.created"),
		interfaces.String("action", "send_welcome_email"),
		interfaces.String("template", "welcome_template_v2"),
	)
}

func simulateOrderCancelledEvent(services map[string]*MicroserviceLogger) {
	trace := TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		UserID:    "user_11111",
		RequestID: "req_" + generateID(),
	}

	ctx := context.Background()

	// Order Service publica evento de cancelamento
	orderLogger := services["order-service"].WithTrace(trace)
	orderLogger.Warn(ctx, "Pedido cancelado",
		interfaces.String("event_type", "order.cancelled"),
		interfaces.String("order_id", "ord_"+generateID()),
		interfaces.String("reason", "user_request"),
		interfaces.Float64("refund_amount", 150.00),
	)

	// Payment Service processa reembolso
	paymentTrace := trace
	paymentTrace.SpanID = generateSpanID()
	paymentTrace.ParentID = trace.SpanID

	paymentLogger := services["payment-service"].WithTrace(paymentTrace)
	paymentLogger.Info(ctx, "Reembolso processado",
		interfaces.String("refund_id", "ref_"+generateID()),
		interfaces.Float64("amount", 150.00),
		interfaces.String("status", "completed"),
	)
}

func simulatePaymentFailedEvent(services map[string]*MicroserviceLogger) {
	trace := TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		UserID:    "user_22222",
		RequestID: "req_" + generateID(),
	}

	ctx := context.Background()

	// Payment Service falha
	paymentLogger := services["payment-service"].WithTrace(trace)
	paymentLogger.Error(ctx, "Falha no processamento do pagamento",
		interfaces.String("event_type", "payment.failed"),
		interfaces.String("payment_id", "pay_"+generateID()),
		interfaces.String("error_code", "INSUFFICIENT_FUNDS"),
		interfaces.Float64("attempted_amount", 599.99),
	)

	// Order Service atualiza status
	orderTrace := trace
	orderTrace.SpanID = generateSpanID()
	orderTrace.ParentID = trace.SpanID

	orderLogger := services["order-service"].WithTrace(orderTrace)
	orderLogger.Warn(ctx, "Status do pedido atualizado",
		interfaces.String("order_id", "ord_"+generateID()),
		interfaces.String("old_status", "processing"),
		interfaces.String("new_status", "payment_failed"),
	)

	// Notification Service envia notificação de falha
	notifTrace := trace
	notifTrace.SpanID = generateSpanID()
	notifTrace.ParentID = trace.SpanID

	notifLogger := services["notification-service"].WithTrace(notifTrace)
	notifLogger.Info(ctx, "Notificação de falha enviada",
		interfaces.String("notification_type", "payment_failed"),
		interfaces.String("channel", "email"),
		interfaces.String("template", "payment_failed_template"),
	)
}

// simulateFailureScenarios simula cenários de falha e recuperação
func simulateFailureScenarios(services map[string]*MicroserviceLogger) {
	ctx := context.Background()

	// Simulação de timeout de serviço
	trace := TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		UserID:    "user_33333",
		RequestID: "req_" + generateID(),
	}

	gatewayLogger := services["api-gateway"].WithTrace(trace)
	gatewayLogger.Error(ctx, "Timeout na comunicação com serviço",
		interfaces.String("target_service", "payment-service"),
		interfaces.Duration("timeout", 5*time.Second),
		interfaces.String("error", "connection_timeout"),
		interfaces.String("fallback_action", "retry_with_backoff"),
	)

	// Simulação de circuit breaker
	gatewayLogger.Warn(ctx, "Circuit breaker ativado",
		interfaces.String("service", "payment-service"),
		interfaces.Int("failure_count", 5),
		interfaces.Duration("cooldown_period", 30*time.Second),
	)

	// Simulação de recuperação
	time.Sleep(100 * time.Millisecond)
	gatewayLogger.Info(ctx, "Serviço recuperado",
		interfaces.String("service", "payment-service"),
		interfaces.String("health_status", "healthy"),
		interfaces.Duration("downtime", 45*time.Second),
	)
}

// simulateLogAggregation simula agregação de logs para análise
func simulateLogAggregation(services map[string]*MicroserviceLogger) {
	ctx := context.Background()

	// Simula agregação de métricas de cada serviço
	for serviceName, service := range services {
		service.logger.Info(ctx, "Métricas do serviço",
			interfaces.String("metric_type", "service_summary"),
			interfaces.Int("requests_processed", rand.Intn(1000)+500),
			interfaces.Int("errors_count", rand.Intn(10)),
			interfaces.Float64("avg_response_time", float64(rand.Intn(200)+50)),
			interfaces.Float64("cpu_usage", float64(rand.Intn(30)+20)),
			interfaces.Float64("memory_usage", float64(rand.Intn(40)+30)),
			interfaces.String("service_name", serviceName),
		)
	}

	// Log de correlação para análise distribuída
	correlationLogger := services["api-gateway"].logger.WithFields(
		interfaces.String("analysis_type", "distributed_trace_summary"),
		interfaces.String("aggregation_window", "1m"),
	)

	correlationLogger.Info(ctx, "Resumo de traces distribuídos",
		interfaces.Int("total_traces", 45),
		interfaces.Int("successful_traces", 42),
		interfaces.Int("failed_traces", 3),
		interfaces.Float64("avg_trace_duration", 285.5),
		interfaces.Float64("p95_trace_duration", 450.0),
		interfaces.Float64("error_rate", 6.67),
	)
}

// Utility functions
func generateTraceID() string {
	return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
}

func generateSpanID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

func generateID() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}
