package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Métricas OpenTelemetry
var (
	meter           metric.Meter
	requestsTotal   metric.Int64Counter
	requestDuration metric.Float64Histogram
	activeOps       metric.Int64UpDownCounter
	orderProcessed  metric.Int64Counter
	orderValue      metric.Float64Histogram
	tracesGenerated metric.Int64Counter
	spanDuration    metric.Float64Histogram
)

// Logger estruturado
var logger *zap.Logger

func init() {
	// Configurar logger estruturado
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	var err error
	logger, err = config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

// Estrutura de dados para pedidos
type Order struct {
	ID            string  `json:"id"`
	CustomerID    string  `json:"customer_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Items         []Item  `json:"items"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func main() {
	ctx := context.Background()

	// Inicializar observabilidade
	shutdown := initObservability(ctx)
	defer shutdown()

	// Configurar HTTP server
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/orders", createOrderHandler)
	http.HandleFunc("/orders/", getOrderHandler)

	// Iniciar servidor
	server := &http.Server{Addr: ":8080"}
	
	go func() {
		logger.Info("Starting advanced observability example server",
			zap.String("port", "8080"))
		
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Simular carga de trabalho
	go simulateWorkload(ctx)

	// Aguardar sinal de parada
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}
}

func initObservability(ctx context.Context) func() {
	// Configurar tracing
	cfg := config.NewConfigFromEnv()
	tracerManager := tracer.NewTracerManager()
	
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}

	// Configurar TracerProvider global
	otel.SetTracerProvider(tracerProvider)

	// Configurar métricas OpenTelemetry
	meter = otel.Meter("advanced-example")
	
	// Inicializar métricas
	initMetrics()

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := tracerManager.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown tracer", zap.Error(err))
		}
		logger.Sync()
	}
}

func initMetrics() {
	var err error

	requestsTotal, err = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		logger.Error("Failed to create requests counter", zap.Error(err))
	}

	requestDuration, err = meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"))
	if err != nil {
		logger.Error("Failed to create request duration histogram", zap.Error(err))
	}

	activeOps, err = meter.Int64UpDownCounter("active_operations",
		metric.WithDescription("Number of active operations"))
	if err != nil {
		logger.Error("Failed to create active operations counter", zap.Error(err))
	}

	orderProcessed, err = meter.Int64Counter("orders_processed_total",
		metric.WithDescription("Total number of processed orders"))
	if err != nil {
		logger.Error("Failed to create orders counter", zap.Error(err))
	}

	orderValue, err = meter.Float64Histogram("order_value_dollars",
		metric.WithDescription("Order value in dollars"))
	if err != nil {
		logger.Error("Failed to create order value histogram", zap.Error(err))
	}

	tracesGenerated, err = meter.Int64Counter("traces_generated_total",
		metric.WithDescription("Total number of traces generated"))
	if err != nil {
		logger.Error("Failed to create traces counter", zap.Error(err))
	}

	spanDuration, err = meter.Float64Histogram("span_duration_seconds",
		metric.WithDescription("Span duration in seconds"))
	if err != nil {
		logger.Error("Failed to create span duration histogram", zap.Error(err))
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Métricas
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		requestDuration.Record(r.Context(), duration, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/health")))
		requestsTotal.Add(r.Context(), 1, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/health"), 
				attribute.String("status", "200")))
	}()

	// Tracing
	tracer := otel.Tracer("http-handler")
	_, span := tracer.Start(r.Context(), "health-check")
	defer span.End()

	// Logging estruturado
	logger.Info("Health check requested",
		zap.String("method", r.Method),
		zap.String("user_agent", r.UserAgent()),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()))

	// Resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Métricas de request
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		requestDuration.Record(r.Context(), duration, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/orders")))
	}()

	// Increment active operations
	activeOps.Add(r.Context(), 1, 
		metric.WithAttributes(attribute.String("operation_type", "order_creation")))
	defer activeOps.Add(r.Context(), -1, 
		metric.WithAttributes(attribute.String("operation_type", "order_creation")))

	// Tracing
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(r.Context(), "create-order")
	defer span.End()

	// Logging com trace context
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	logger.Info("Creating new order",
		zap.String("method", r.Method),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID))

	// Processar pedido
	order, err := processOrder(ctx)
	if err != nil {
		// Métricas de erro
		requestsTotal.Add(r.Context(), 1, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/orders"), 
				attribute.String("status", "500")))
		
		// Tracing de erro
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		
		// Log de erro
		logger.Error("Failed to process order",
			zap.Error(err),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID))
		
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Métricas de sucesso
	requestsTotal.Add(r.Context(), 1, 
		metric.WithAttributes(attribute.String("method", r.Method), 
			attribute.String("endpoint", "/orders"), 
			attribute.String("status", "201")))
	orderProcessed.Add(r.Context(), 1, 
		metric.WithAttributes(attribute.String("status", "success"), 
			attribute.String("payment_method", order.PaymentMethod)))
	orderValue.Record(r.Context(), order.Amount, 
		metric.WithAttributes(attribute.String("payment_method", order.PaymentMethod)))

	// Atributos de span
	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("order.customer_id", order.CustomerID),
		attribute.Float64("order.amount", order.Amount),
		attribute.String("order.payment_method", order.PaymentMethod),
		attribute.Int("order.items_count", len(order.Items)),
	)

	// Log de sucesso
	logger.Info("Order created successfully",
		zap.String("order_id", order.ID),
		zap.String("customer_id", order.CustomerID),
		zap.Float64("amount", order.Amount),
		zap.String("payment_method", order.PaymentMethod),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID))

	// Resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"order_id":"%s","status":"created","amount":%.2f}`, 
		order.ID, order.Amount)
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		requestDuration.Record(r.Context(), duration, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/orders/:id")))
		requestsTotal.Add(r.Context(), 1, 
			metric.WithAttributes(attribute.String("method", r.Method), 
				attribute.String("endpoint", "/orders/:id"), 
				attribute.String("status", "200")))
	}()

	tracer := otel.Tracer("order-service")
	_, span := tracer.Start(r.Context(), "get-order")
	defer span.End()

	orderID := r.URL.Path[len("/orders/"):]
	
	span.SetAttributes(attribute.String("order.id", orderID))
	
	logger.Info("Fetching order",
		zap.String("order_id", orderID),
		zap.String("trace_id", span.SpanContext().TraceID().String()))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"order_id":"%s","status":"completed","amount":150.00}`, orderID)
}

func processOrder(ctx context.Context) (*Order, error) {
	tracer := otel.Tracer("order-processor")
	ctx, span := tracer.Start(ctx, "process-order")
	defer span.End()

	// Simular processamento com múltiplas etapas
	order := &Order{
		ID:            generateOrderID(),
		CustomerID:    generateCustomerID(),
		Amount:        generateOrderAmount(),
		PaymentMethod: generatePaymentMethod(),
		Items:         generateItems(),
	}

	// Validar pedido
	if err := validateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Processar pagamento
	if err := processPayment(ctx, order); err != nil {
		return nil, err
	}

	// Verificar estoque
	if err := checkInventory(ctx, order); err != nil {
		return nil, err
	}

	// Calcular frete
	if err := calculateShipping(ctx, order); err != nil {
		return nil, err
	}

	// Enviar notificações
	if err := sendNotifications(ctx, order); err != nil {
		logger.Warn("Failed to send notifications, but order was processed",
			zap.String("order_id", order.ID),
			zap.Error(err))
		// Não falha o pedido por falha de notificação
	}

	// Métricas OpenTelemetry
	tracesGenerated.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "order_processing"),
	))

	return order, nil
}

func validateOrder(ctx context.Context, order *Order) error {
	tracer := otel.Tracer("order-validator")
	_, span := tracer.Start(ctx, "validate-order")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		spanDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "validate_order"),
		))
	}()

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.Float64("order.amount", order.Amount),
	)

	logger.Debug("Validating order",
		zap.String("order_id", order.ID),
		zap.String("trace_id", span.SpanContext().TraceID().String()))

	// Simular validação
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	if order.Amount <= 0 {
		err := fmt.Errorf("invalid order amount: %.2f", order.Amount)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func processPayment(ctx context.Context, order *Order) error {
	tracer := otel.Tracer("payment-service")
	_, span := tracer.Start(ctx, "process-payment")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		spanDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "process_payment"),
		))
	}()

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("payment.method", order.PaymentMethod),
		attribute.Float64("payment.amount", order.Amount),
	)

	logger.Info("Processing payment",
		zap.String("order_id", order.ID),
		zap.String("payment_method", order.PaymentMethod),
		zap.Float64("amount", order.Amount),
		zap.String("trace_id", span.SpanContext().TraceID().String()))

	// Simular processamento de pagamento
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	// Simular falha de pagamento ocasional
	if rand.Float32() < 0.05 { // 5% de falha
		err := fmt.Errorf("payment failed for order %s", order.ID)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Payment processing failed",
			zap.String("order_id", order.ID),
			zap.Error(err))
		return err
	}

	logger.Info("Payment processed successfully",
		zap.String("order_id", order.ID),
		zap.String("payment_method", order.PaymentMethod))

	return nil
}

func checkInventory(ctx context.Context, order *Order) error {
	tracer := otel.Tracer("inventory-service")
	_, span := tracer.Start(ctx, "check-inventory")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		spanDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "check_inventory"),
		))
	}()

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.Int("items.count", len(order.Items)),
	)

	logger.Debug("Checking inventory",
		zap.String("order_id", order.ID),
		zap.Int("items_count", len(order.Items)))

	// Verificar cada item
	for i, item := range order.Items {
		itemSpan := trace.SpanFromContext(ctx)
		itemSpan.AddEvent("checking_item", trace.WithAttributes(
			attribute.String("item.product_id", item.ProductID),
			attribute.Int("item.quantity", item.Quantity),
		))

		// Simular verificação de estoque
		time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)

		logger.Debug("Item inventory checked",
			zap.String("product_id", item.ProductID),
			zap.Int("quantity", item.Quantity),
			zap.Int("item_index", i))
	}

	return nil
}

func calculateShipping(ctx context.Context, order *Order) error {
	tracer := otel.Tracer("shipping-service")
	_, span := tracer.Start(ctx, "calculate-shipping")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		spanDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "calculate_shipping"),
		))
	}()

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("customer.id", order.CustomerID),
	)

	logger.Debug("Calculating shipping",
		zap.String("order_id", order.ID),
		zap.String("customer_id", order.CustomerID))

	// Simular cálculo de frete
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	shippingCost := order.Amount * 0.1 // 10% do valor do pedido
	span.SetAttributes(attribute.Float64("shipping.cost", shippingCost))

	logger.Info("Shipping calculated",
		zap.String("order_id", order.ID),
		zap.Float64("shipping_cost", shippingCost))

	return nil
}

func sendNotifications(ctx context.Context, order *Order) error {
	tracer := otel.Tracer("notification-service")
	_, span := tracer.Start(ctx, "send-notifications")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		spanDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", "send_notifications"),
		))
	}()

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("customer.id", order.CustomerID),
	)

	// Enviar email
	_, emailSpan := tracer.Start(ctx, "send-email")
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	logger.Debug("Email notification sent", zap.String("order_id", order.ID))
	emailSpan.End()

	// Enviar SMS
	_, smsSpan := tracer.Start(ctx, "send-sms")
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	logger.Debug("SMS notification sent", zap.String("order_id", order.ID))
	smsSpan.End()

	// Notificação push
	_, pushSpan := tracer.Start(ctx, "send-push-notification")
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
	logger.Debug("Push notification sent", zap.String("order_id", order.ID))
	pushSpan.End()

	return nil
}

func simulateWorkload(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simular processamento em background
			go func() {
				tracer := otel.Tracer("background-worker")
				_, span := tracer.Start(context.Background(), "background-task")
				defer span.End()

				taskType := []string{"cleanup", "analytics", "maintenance"}[rand.Intn(3)]
				span.SetAttributes(attribute.String("task.type", taskType))

				logger.Debug("Background task started", zap.String("task_type", taskType))

				// Simular trabalho
				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

				logger.Debug("Background task completed", zap.String("task_type", taskType))
			}()
		}
	}
}

// Funções auxiliares de geração de dados
func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano()%1000000)
}

func generateCustomerID() string {
	return fmt.Sprintf("CUST-%d", rand.Intn(10000))
}

func generateOrderAmount() float64 {
	return float64(rand.Intn(500)+10) + rand.Float64()
}

func generatePaymentMethod() string {
	methods := []string{"credit_card", "debit_card", "paypal", "pix", "boleto"}
	return methods[rand.Intn(len(methods))]
}

func generateItems() []Item {
	itemCount := rand.Intn(5) + 1
	items := make([]Item, itemCount)
	
	for i := 0; i < itemCount; i++ {
		items[i] = Item{
			ProductID: fmt.Sprintf("PROD-%d", rand.Intn(1000)),
			Quantity:  rand.Intn(3) + 1,
			Price:     float64(rand.Intn(100)+5) + rand.Float64(),
		}
	}
	
	return items
}
