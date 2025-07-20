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
	fmt.Println("📊 Exemplo Grafana Tempo")
	fmt.Println("========================")

	// Configuração para Grafana Tempo
	cfg := interfaces.Config{
		ServiceName:   "grafana-example-service",
		Environment:   "development",
		ExporterType:  "grafana",
		Endpoint:      "http://tempo:3200", // ou "tempo:9095" para gRPC
		SamplingRatio: 1.0,                 // 100% sampling para desenvolvimento
		Version:       "1.0.0",
		Propagators:   []string{"tracecontext", "b3"},
		Headers: map[string]string{
			"X-Scope-OrgID": "tenant-1", // Para multi-tenancy
		},
		Attributes: map[string]string{
			"team":        "platform",
			"environment": "development",
			"cluster":     "dev-cluster",
		},
	}

	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("❌ Erro na configuração: %v", err)
	}

	// Inicializar TracerManager
	tracerManager := tracer.NewTracerManager()
	ctx := context.Background()

	fmt.Println("📡 Inicializando Grafana Tempo tracer...")
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar tracer: %v", err)
	}

	// Configurar como tracer global (opcional)
	otel.SetTracerProvider(tracerProvider)
	fmt.Println("✅ Grafana Tempo tracer configurado globalmente")

	// Obter tracer para este serviço
	tr := tracerProvider.Tracer("grafana-example")

	// Exemplo de operação com tracing
	runECommerceWorkflow(ctx, tr)

	// Aguardar um pouco para envio dos dados
	fmt.Println("⏳ Aguardando envio de traces...")
	time.Sleep(2 * time.Second)

	// Shutdown graceful
	fmt.Println("🔄 Fazendo shutdown do tracer...")
	if err := tracerManager.Shutdown(ctx); err != nil {
		log.Printf("⚠️ Erro no shutdown: %v", err)
	}

	fmt.Println("✅ Exemplo concluído!")
	fmt.Println("\n📊 Verifique os traces em:")
	fmt.Println("   http://grafana:3000/explore (Tempo datasource)")
	fmt.Println("   ou busque por trace ID no Grafana")
}

func runECommerceWorkflow(ctx context.Context, tracer trace.Tracer) {
	// Criar span principal para o workflow de e-commerce
	ctx, span := tracer.Start(ctx, "ecommerce-order-workflow")
	defer span.End()

	orderID := "order-123456"
	customerID := "customer-789"

	span.SetAttributes(
		attribute.String("workflow.type", "order-processing"),
		attribute.String("order.id", orderID),
		attribute.String("customer.id", customerID),
		attribute.Int("order.items_count", 3),
		attribute.Float64("order.total", 299.99),
	)

	fmt.Println("🛒 Iniciando workflow de pedido...")

	// Validação do pedido
	if !validateOrder(ctx, tracer, orderID) {
		span.SetStatus(codes.Error, "Validação do pedido falhou")
		return
	}

	// Verificação de inventário
	if !checkInventory(ctx, tracer, orderID) {
		span.SetStatus(codes.Error, "Itens indisponíveis")
		return
	}

	// Processamento de pagamento
	if !processPayment(ctx, tracer, customerID, 299.99) {
		span.SetStatus(codes.Error, "Falha no pagamento")
		return
	}

	// Reserva de inventário
	reserveInventory(ctx, tracer, orderID)

	// Criar shipping
	createShipping(ctx, tracer, orderID)

	// Notificar cliente
	notifyCustomer(ctx, tracer, customerID, orderID)

	span.SetStatus(codes.Ok, "Pedido processado com sucesso")
	fmt.Println("✅ Workflow de pedido concluído")
}

func validateOrder(ctx context.Context, tracer trace.Tracer, orderID string) bool {
	ctx, span := tracer.Start(ctx, "validate-order")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("validation.type", "business-rules"),
	)

	// Simular validação
	time.Sleep(30 * time.Millisecond)

	fmt.Printf("✅ Pedido %s validado\n", orderID)
	return true
}

func checkInventory(ctx context.Context, tracer trace.Tracer, orderID string) bool {
	ctx, span := tracer.Start(ctx, "check-inventory")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("inventory.system", "warehouse-api"),
	)

	// Simular verificação de inventário
	time.Sleep(50 * time.Millisecond)

	// Simulação de chamada para múltiplos warehouses
	warehouses := []string{"warehouse-1", "warehouse-2", "warehouse-3"}
	for _, warehouse := range warehouses {
		checkWarehouse(ctx, tracer, warehouse, orderID)
	}

	fmt.Printf("📦 Inventário verificado para pedido %s\n", orderID)
	return true
}

func checkWarehouse(ctx context.Context, tracer trace.Tracer, warehouseID, orderID string) {
	ctx, span := tracer.Start(ctx, "check-warehouse")
	defer span.End()

	span.SetAttributes(
		attribute.String("warehouse.id", warehouseID),
		attribute.String("order.id", orderID),
		attribute.Int("items.available", 10),
	)

	time.Sleep(20 * time.Millisecond)
	fmt.Printf("🏪 Warehouse %s verificado\n", warehouseID)
}

func processPayment(ctx context.Context, tracer trace.Tracer, customerID string, amount float64) bool {
	ctx, span := tracer.Start(ctx, "process-payment")
	defer span.End()

	span.SetAttributes(
		attribute.String("customer.id", customerID),
		attribute.Float64("payment.amount", amount),
		attribute.String("payment.method", "credit-card"),
		attribute.String("payment.gateway", "stripe"),
	)

	// Simular processamento de pagamento
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("💳 Pagamento de $%.2f processado para cliente %s\n", amount, customerID)
	return true
}

func reserveInventory(ctx context.Context, tracer trace.Tracer, orderID string) {
	ctx, span := tracer.Start(ctx, "reserve-inventory")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("reservation.status", "confirmed"),
	)

	time.Sleep(40 * time.Millisecond)
	fmt.Printf("📦 Inventário reservado para pedido %s\n", orderID)
}

func createShipping(ctx context.Context, tracer trace.Tracer, orderID string) {
	ctx, span := tracer.Start(ctx, "create-shipping")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("shipping.carrier", "fedex"),
		attribute.String("shipping.service", "ground"),
		attribute.String("tracking.number", "1Z999AA1234567890"),
	)

	time.Sleep(60 * time.Millisecond)
	fmt.Printf("🚚 Shipping criado para pedido %s\n", orderID)
}

func notifyCustomer(ctx context.Context, tracer trace.Tracer, customerID, orderID string) {
	ctx, span := tracer.Start(ctx, "notify-customer")
	defer span.End()

	span.SetAttributes(
		attribute.String("customer.id", customerID),
		attribute.String("order.id", orderID),
		attribute.String("notification.type", "order-confirmation"),
		attribute.String("notification.channel", "email"),
	)

	time.Sleep(25 * time.Millisecond)
	fmt.Printf("📧 Cliente %s notificado sobre pedido %s\n", customerID, orderID)
}
