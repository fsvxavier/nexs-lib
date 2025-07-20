package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/config"
	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

func main() {
	fmt.Println("🟢 Exemplo New Relic Distributed Tracing")
	fmt.Println("=========================================")

	// Configuração para New Relic
	cfg := interfaces.Config{
		ServiceName:   "newrelic-example-service",
		Environment:   "development",
		ExporterType:  "newrelic",
		LicenseKey:    "your-newrelic-license-key", // 40 caracteres
		SamplingRatio: 1.0,                         // 100% sampling para desenvolvimento
		Version:       "1.0.0",
		Propagators:   []string{"tracecontext", "b3"},
		Endpoint:      "https://trace-api.newrelic.com/trace/v1", // US datacenter
		// Para EU: "https://trace-api.eu.newrelic.com/trace/v1"
		Attributes: map[string]string{
			"team":        "platform",
			"environment": "development",
			"datacenter":  "us-east-1",
			"application": "microservices-demo",
		},
	}

	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("❌ Erro na configuração: %v", err)
	}

	// Inicializar TracerManager
	tracerManager := tracer.NewTracerManager()
	ctx := context.Background()

	fmt.Println("📡 Inicializando New Relic tracer...")
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar tracer: %v", err)
	}

	// Configurar como tracer global (opcional)
	otel.SetTracerProvider(tracerProvider)
	fmt.Println("✅ New Relic tracer configurado globalmente")

	// Obter tracer para este serviço
	tr := tracerProvider.Tracer("newrelic-example")

	// Exemplo de operação com tracing
	runMicroservicesWorkflow(ctx, tr)

	// Aguardar um pouco para envio dos dados
	fmt.Println("⏳ Aguardando envio de traces...")
	time.Sleep(3 * time.Second)

	// Shutdown graceful
	fmt.Println("🔄 Fazendo shutdown do tracer...")
	if err := tracerManager.Shutdown(ctx); err != nil {
		log.Printf("⚠️ Erro no shutdown: %v", err)
	}

	fmt.Println("✅ Exemplo concluído!")
	fmt.Println("\n📊 Verifique os traces em:")
	fmt.Println("   https://one.newrelic.com/distributed-tracing")
}

func runMicroservicesWorkflow(ctx context.Context, tracer trace.Tracer) {
	// Criar span principal para workflow de microserviços
	ctx, span := tracer.Start(ctx, "microservices-api-workflow")
	defer span.End()

	requestID := "req-abc123456"
	userID := "user-789012"

	span.SetAttributes(
		attribute.String("workflow.type", "api-request"),
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/orders"),
		attribute.String("request.id", requestID),
		attribute.String("user.id", userID),
		attribute.Int("http.status_code", 200),
	)

	fmt.Println("🌐 Iniciando workflow de microserviços...")

	// Simular request HTTP recebido
	if !authenticateUser(ctx, tracer, userID) {
		span.SetStatus(codes.Error, "Falha na autenticação")
		return
	}

	// Chamar User Service
	userData := callUserService(ctx, tracer, userID)
	if userData == nil {
		span.SetStatus(codes.Error, "Falha ao obter dados do usuário")
		return
	}

	// Chamar Product Service
	if !callProductService(ctx, tracer, "product-123") {
		span.SetStatus(codes.Error, "Produto não encontrado")
		return
	}

	// Chamar Inventory Service
	if !callInventoryService(ctx, tracer, "product-123") {
		span.SetStatus(codes.Error, "Produto indisponível")
		return
	}

	// Chamar Payment Service
	if !callPaymentService(ctx, tracer, userID, 99.99) {
		span.SetStatus(codes.Error, "Falha no pagamento")
		return
	}

	// Chamar Order Service
	orderID := callOrderService(ctx, tracer, userID, "product-123")

	// Chamar Notification Service
	callNotificationService(ctx, tracer, userID, orderID)

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("response.status", "success"),
	)

	span.SetStatus(codes.Ok, "Workflow concluído com sucesso")
	fmt.Println("✅ Workflow de microserviços concluído")
}

func authenticateUser(ctx context.Context, tracer trace.Tracer, userID string) bool {
	ctx, span := tracer.Start(ctx, "auth-service.authenticate")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "auth-service"),
		attribute.String("user.id", userID),
		attribute.String("auth.method", "jwt"),
	)

	// Simular autenticação
	time.Sleep(30 * time.Millisecond)

	fmt.Printf("🔐 Usuário %s autenticado\n", userID)
	return true
}

func callUserService(ctx context.Context, tracer trace.Tracer, userID string) map[string]interface{} {
	ctx, span := tracer.Start(ctx, "user-service.get-user")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "user-service"),
		attribute.String("user.id", userID),
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "/users/"+userID),
	)

	// Simular chamada para banco de dados
	queryUserDatabase(ctx, tracer, userID)

	// Simular resposta
	time.Sleep(50 * time.Millisecond)

	userData := map[string]interface{}{
		"id":    userID,
		"name":  "João Silva",
		"email": "joao@example.com",
	}

	span.SetAttributes(
		attribute.String("user.name", userData["name"].(string)),
		attribute.String("user.email", userData["email"].(string)),
	)

	fmt.Printf("👤 Dados do usuário %s obtidos\n", userID)
	return userData
}

func queryUserDatabase(ctx context.Context, tracer trace.Tracer, userID string) {
	ctx, span := tracer.Start(ctx, "database.query-user")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "users_db"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
		attribute.String("db.query", "SELECT * FROM users WHERE id = $1"),
	)

	time.Sleep(25 * time.Millisecond)
	fmt.Printf("🗃️ Consulta ao banco para usuário %s\n", userID)
}

func callProductService(ctx context.Context, tracer trace.Tracer, productID string) bool {
	ctx, span := tracer.Start(ctx, "product-service.get-product")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "product-service"),
		attribute.String("product.id", productID),
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "/products/"+productID),
	)

	// Simular cache lookup
	if checkProductCache(ctx, tracer, productID) {
		span.SetAttributes(attribute.Bool("cache.hit", true))
	} else {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		queryProductDatabase(ctx, tracer, productID)
	}

	time.Sleep(40 * time.Millisecond)

	fmt.Printf("🛍️ Produto %s encontrado\n", productID)
	return true
}

func checkProductCache(ctx context.Context, tracer trace.Tracer, productID string) bool {
	ctx, span := tracer.Start(ctx, "cache.lookup-product")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.system", "redis"),
		attribute.String("cache.key", "product:"+productID),
	)

	time.Sleep(5 * time.Millisecond)
	return false // Simular cache miss
}

func queryProductDatabase(ctx context.Context, tracer trace.Tracer, productID string) {
	ctx, span := tracer.Start(ctx, "database.query-product")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "mongodb"),
		attribute.String("db.name", "products_db"),
		attribute.String("db.collection", "products"),
		attribute.String("db.operation", "findOne"),
	)

	time.Sleep(30 * time.Millisecond)
	fmt.Printf("🗃️ Consulta ao banco para produto %s\n", productID)
}

func callInventoryService(ctx context.Context, tracer trace.Tracer, productID string) bool {
	ctx, span := tracer.Start(ctx, "inventory-service.check-stock")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "inventory-service"),
		attribute.String("product.id", productID),
		attribute.Int("stock.quantity", 15),
		attribute.String("warehouse.location", "warehouse-east"),
	)

	time.Sleep(60 * time.Millisecond)

	fmt.Printf("📦 Estoque verificado para produto %s\n", productID)
	return true
}

func callPaymentService(ctx context.Context, tracer trace.Tracer, userID string, amount float64) bool {
	ctx, span := tracer.Start(ctx, "payment-service.process-payment")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "payment-service"),
		attribute.String("user.id", userID),
		attribute.Float64("payment.amount", amount),
		attribute.String("payment.currency", "USD"),
		attribute.String("payment.provider", "stripe"),
	)

	// Simular chamada para gateway externo
	callExternalPaymentGateway(ctx, tracer, amount)

	time.Sleep(80 * time.Millisecond)

	fmt.Printf("💳 Pagamento de $%.2f processado para usuário %s\n", amount, userID)
	return true
}

func callExternalPaymentGateway(ctx context.Context, tracer trace.Tracer, amount float64) {
	ctx, span := tracer.Start(ctx, "external.stripe-api")
	defer span.End()

	span.SetAttributes(
		attribute.String("external.service", "stripe"),
		attribute.String("http.method", "POST"),
		attribute.String("http.url", "https://api.stripe.com/v1/charges"),
		attribute.Float64("charge.amount", amount),
	)

	time.Sleep(100 * time.Millisecond)
	fmt.Println("🔗 Chamada para gateway externo (Stripe)")
}

func callOrderService(ctx context.Context, tracer trace.Tracer, userID, productID string) string {
	ctx, span := tracer.Start(ctx, "order-service.create-order")
	defer span.End()

	orderID := "order-" + fmt.Sprintf("%d", time.Now().Unix())

	span.SetAttributes(
		attribute.String("service.name", "order-service"),
		attribute.String("user.id", userID),
		attribute.String("product.id", productID),
		attribute.String("order.id", orderID),
		attribute.String("order.status", "created"),
	)

	time.Sleep(45 * time.Millisecond)

	fmt.Printf("📝 Pedido %s criado para usuário %s\n", orderID, userID)
	return orderID
}

func callNotificationService(ctx context.Context, tracer trace.Tracer, userID, orderID string) {
	ctx, span := tracer.Start(ctx, "notification-service.send-confirmation")
	defer span.End()

	span.SetAttributes(
		attribute.String("service.name", "notification-service"),
		attribute.String("user.id", userID),
		attribute.String("order.id", orderID),
		attribute.String("notification.type", "order-confirmation"),
		attribute.String("notification.channel", "email"),
	)

	time.Sleep(35 * time.Millisecond)

	fmt.Printf("📧 Notificação enviada para usuário %s sobre pedido %s\n", userID, orderID)
}
