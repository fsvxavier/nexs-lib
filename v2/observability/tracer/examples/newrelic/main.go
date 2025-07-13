package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/newrelic"
)

func main() {
	// Configure New Relic provider
	config := &newrelic.Config{
		AppName:               "example-newrelic-service",
		LicenseKey:            "your-40-character-license-key", // Replace with your actual license key
		Environment:           "development",
		ServiceVersion:        "1.0.0",
		DistributedTracer:     true,
		Enabled:               true,
		HighSecurity:          false,
		CodeLevelMetrics:      true,
		LogLevel:              "info",
		MaxSamplesStored:      10000,
		DatastoreTracer:       true,
		CrossApplicationTrace: true,
		AttributesEnabled:     true,
		AttributesExclude: []string{
			"request.headers.authorization",
			"request.headers.cookie",
		},
		CustomInsightsEvents: true,
		Labels: map[string]string{
			"team":        "backend",
			"environment": "development",
			"version":     "1.0.0",
			"component":   "api-server",
		},
	}

	// Create provider
	provider, err := newrelic.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create New Relic provider: %v", err)
	}

	// Ensure proper shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := provider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down provider: %v", err)
		}
	}()

	// Create tracer
	tr, err := provider.CreateTracer("api-server",
		tracer.WithServiceName("example-newrelic-service"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Setup HTTP server with tracing
	http.HandleFunc("/", handleRoot(tr))
	http.HandleFunc("/orders", handleOrders(tr))
	http.HandleFunc("/metrics", handleMetrics(tr))
	http.HandleFunc("/error", handleError(tr))

	fmt.Println("Starting New Relic instrumented server on :8080...")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/orders")
	fmt.Println("  curl http://localhost:8080/metrics")
	fmt.Println("  curl http://localhost:8080/error")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleRoot demonstrates basic transaction tracking
func handleRoot(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start root span (transaction in New Relic terms)
		ctx, span := tr.StartSpan(r.Context(), "web.transaction",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method":      r.Method,
				"http.url":         r.URL.String(),
				"http.user_agent":  r.UserAgent(),
				"http.remote_addr": r.RemoteAddr,
				"request.id":       generateRequestID(),
			}),
		)
		defer span.End()

		// Add custom attributes for New Relic insights
		span.SetAttribute("transaction.type", "web")
		span.SetAttribute("endpoint.name", "home")
		span.SetAttribute("user.session", "sess_"+generateRequestID())

		// Record custom event
		span.AddEvent("page.view", map[string]interface{}{
			"page":      "home",
			"timestamp": time.Now().Unix(),
			"browser":   r.UserAgent(),
		})

		// Simulate some business logic
		processHomeRequest(ctx, tr)

		// Record business metrics
		span.SetAttribute("business.conversion_value", 0.0)
		span.SetAttribute("business.feature_flags", "home_v2:true,analytics:true")

		// Set response attributes
		span.SetAttribute("http.status_code", 200)
		span.SetAttribute("response.content_type", "text/html")
		span.SetStatus(tracer.StatusCodeOk, "Home page rendered successfully")

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>New Relic Example</title></head>
<body>
	<h1>Hello from New Relic Instrumented Service!</h1>
	<p>This request is being tracked by New Relic APM.</p>
	<p>Try other endpoints:</p>
	<ul>
		<li><a href="/orders">Orders</a></li>
		<li><a href="/metrics">Metrics</a></li>
		<li><a href="/error">Error Example</a></li>
	</ul>
</body>
</html>`)
	}
}

// handleOrders demonstrates business transaction tracking
func handleOrders(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start transaction span
		ctx, span := tr.StartSpan(r.Context(), "orders.list",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method": r.Method,
				"http.route":  "/orders",
			}),
		)
		defer span.End()

		// Add business context
		span.SetAttribute("transaction.type", "business")
		span.SetAttribute("operation.type", "read")
		span.SetAttribute("user.id", "user_12345")
		span.SetAttribute("user.tier", "premium")

		// Simulate authentication and authorization
		userID, authenticated := authenticateUserNR(ctx, tr, r)
		if !authenticated {
			span.SetStatus(tracer.StatusCodeError, "Authentication failed")
			span.SetAttribute("error", true)
			span.SetAttribute("error.class", "AuthenticationError")
			span.SetAttribute("error.message", "Invalid or missing authentication token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		span.SetAttribute("user.authenticated", true)
		span.SetAttribute("user.id", userID)

		// Fetch orders from database
		orders, totalValue, err := fetchOrdersFromDB(ctx, tr, userID)
		if err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Database error: %v", err))
			span.SetAttribute("error", true)
			span.SetAttribute("error.class", "DatabaseError")
			span.SetAttribute("error.message", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Record business metrics
		span.SetAttribute("orders.count", len(orders))
		span.SetAttribute("orders.total_value", totalValue)
		span.SetAttribute("business.revenue_impact", totalValue)

		// Add custom event for business intelligence
		span.AddEvent("orders.retrieved", map[string]interface{}{
			"user_id":     userID,
			"order_count": len(orders),
			"total_value": totalValue,
			"timestamp":   time.Now().Unix(),
		})

		// Call external payment service
		paymentStatus := checkPaymentService(ctx, tr)
		span.SetAttribute("payment.service_status", paymentStatus)

		span.SetAttribute("http.status_code", 200)
		span.SetStatus(tracer.StatusCodeOk, "Orders retrieved successfully")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"user_id": "%s",
	"orders": %d,
	"total_value": %.2f,
	"payment_status": "%s",
	"data": %v
}`, userID, len(orders), totalValue, paymentStatus, orders)
	}
}

// handleMetrics demonstrates custom metrics and performance tracking
func handleMetrics(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.StartSpan(r.Context(), "metrics.dashboard",
			tracer.WithSpanKind(tracer.SpanKindServer),
		)
		defer span.End()

		// Record performance metrics
		span.SetAttribute("operation.type", "metrics_collection")
		span.SetAttribute("dashboard.type", "business")

		// Collect various metrics
		metrics := collectBusinessMetrics(ctx, tr)

		// Add custom New Relic attributes
		for key, value := range metrics {
			span.SetAttribute(fmt.Sprintf("metric.%s", key), value)
		}

		// Record custom events for each metric
		for key, value := range metrics {
			span.AddEvent("metric.collected", map[string]interface{}{
				"metric_name":  key,
				"metric_value": value,
				"timestamp":    time.Now().Unix(),
			})
		}

		span.SetAttribute("metrics.count", len(metrics))
		span.SetStatus(tracer.StatusCodeOk, "Metrics collected successfully")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"timestamp": %d,
	"metrics": %d,
	"data": {`, time.Now().Unix(), len(metrics))

		first := true
		for key, value := range metrics {
			if !first {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `"%s": %v`, key, value)
			first = false
		}

		fmt.Fprint(w, `}
}`)
	}
}

// handleError demonstrates error tracking and reporting
func handleError(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.StartSpan(r.Context(), "error.simulation",
			tracer.WithSpanKind(tracer.SpanKindServer),
		)
		defer span.End()

		// Simulate different types of errors
		errorType := r.URL.Query().Get("type")
		if errorType == "" {
			errorType = "generic"
		}

		span.SetAttribute("error.simulation", true)
		span.SetAttribute("error.type_requested", errorType)

		var err error
		var statusCode int

		switch errorType {
		case "validation":
			err = fmt.Errorf("validation failed: invalid email format")
			statusCode = http.StatusBadRequest
			span.SetAttribute("error.class", "ValidationError")
			span.SetAttribute("validation.field", "email")
		case "database":
			err = simulateDatabaseError(ctx, tr)
			statusCode = http.StatusInternalServerError
			span.SetAttribute("error.class", "DatabaseError")
		case "external":
			err = simulateExternalServiceError(ctx, tr)
			statusCode = http.StatusBadGateway
			span.SetAttribute("error.class", "ExternalServiceError")
		case "timeout":
			err = fmt.Errorf("operation timed out after 30 seconds")
			statusCode = http.StatusRequestTimeout
			span.SetAttribute("error.class", "TimeoutError")
			span.SetAttribute("timeout.duration", 30)
		default:
			err = fmt.Errorf("unexpected error occurred")
			statusCode = http.StatusInternalServerError
			span.SetAttribute("error.class", "GenericError")
		}

		// Record error details
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.SetAttribute("error", true)
		span.SetAttribute("error.message", err.Error())
		span.SetAttribute("http.status_code", statusCode)

		// Add error event for detailed tracking
		span.AddEvent("error.occurred", map[string]interface{}{
			"error_type":    errorType,
			"error_message": err.Error(),
			"status_code":   statusCode,
			"timestamp":     time.Now().Unix(),
		})

		http.Error(w, fmt.Sprintf(`{
	"error": true,
	"type": "%s",
	"message": "%s",
	"status_code": %d,
	"timestamp": %d
}`, errorType, err.Error(), statusCode, time.Now().Unix()), statusCode)
	}
}

// Helper functions

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func processHomeRequest(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "home.render",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate template rendering
	time.Sleep(20 * time.Millisecond)

	span.SetAttribute("template.name", "home.html")
	span.SetAttribute("template.size", 1024)
	span.SetAttribute("render.duration_ms", 20)
	span.SetStatus(tracer.StatusCodeOk, "Template rendered")
}

func authenticateUserNR(ctx context.Context, tr tracer.Tracer, r *http.Request) (string, bool) {
	_, span := tr.StartSpan(ctx, "auth.validate",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate auth validation
	time.Sleep(15 * time.Millisecond)

	token := r.Header.Get("Authorization")
	userID := "user_12345"

	span.SetAttribute("auth.method", "bearer_token")
	span.SetAttribute("auth.token_present", token != "")

	if token != "" {
		span.SetAttribute("auth.success", true)
		span.SetAttribute("user.id", userID)
		span.SetStatus(tracer.StatusCodeOk, "Authentication successful")
		return userID, true
	}

	span.SetAttribute("auth.success", false)
	span.SetStatus(tracer.StatusCodeError, "Authentication failed")
	return "", false
}

func fetchOrdersFromDB(ctx context.Context, tr tracer.Tracer, userID string) ([]string, float64, error) {
	_, span := tr.StartSpan(ctx, "db.orders.select",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db.system":    "postgresql",
			"db.name":      "orders",
			"db.operation": "SELECT",
			"db.table":     "orders",
			"user.id":      userID,
		}),
	)
	defer span.End()

	// Simulate database query
	time.Sleep(80 * time.Millisecond)

	// Simulate occasional database errors
	if time.Now().UnixNano()%15 == 0 {
		err := fmt.Errorf("database connection pool exhausted")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.SetAttribute("error", true)
		span.SetAttribute("error.class", "ConnectionError")
		return nil, 0, err
	}

	orders := []string{"order_001", "order_002", "order_003"}
	totalValue := 299.97

	span.SetAttribute("db.rows_returned", len(orders))
	span.SetAttribute("orders.total_value", totalValue)
	span.SetAttribute("db.query_duration_ms", 80)
	span.SetStatus(tracer.StatusCodeOk, "Orders fetched successfully")

	return orders, totalValue, nil
}

func checkPaymentService(ctx context.Context, tr tracer.Tracer) string {
	_, span := tr.StartSpan(ctx, "external.payment.status",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"external.service": "payment_gateway",
			"external.vendor":  "stripe",
		}),
	)
	defer span.End()

	// Simulate external API call
	time.Sleep(45 * time.Millisecond)

	status := "operational"
	span.SetAttribute("payment.status", status)
	span.SetAttribute("external.response_time_ms", 45)
	span.SetStatus(tracer.StatusCodeOk, "Payment service checked")

	return status
}

func collectBusinessMetrics(ctx context.Context, tr tracer.Tracer) map[string]interface{} {
	_, span := tr.StartSpan(ctx, "metrics.collect",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate metrics collection
	time.Sleep(60 * time.Millisecond)

	metrics := map[string]interface{}{
		"active_users":      1250,
		"orders_today":      89,
		"revenue_today":     4567.89,
		"avg_response_time": 125.5,
		"error_rate":        0.02,
		"cpu_usage":         0.45,
		"memory_usage":      0.67,
		"db_connections":    12,
		"cache_hit_rate":    0.94,
		"queue_depth":       3,
	}

	span.SetAttribute("metrics.collection_time_ms", 60)
	span.SetAttribute("metrics.count", len(metrics))
	span.SetStatus(tracer.StatusCodeOk, "Metrics collected")

	return metrics
}

func simulateDatabaseError(ctx context.Context, tr tracer.Tracer) error {
	_, span := tr.StartSpan(ctx, "db.error_simulation",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(30 * time.Millisecond)

	err := fmt.Errorf("database connection timeout after 30s")
	span.SetStatus(tracer.StatusCodeError, err.Error())
	span.SetAttribute("error", true)
	span.SetAttribute("error.type", "timeout")
	span.SetAttribute("timeout.duration", 30)

	return err
}

func simulateExternalServiceError(ctx context.Context, tr tracer.Tracer) error {
	_, span := tr.StartSpan(ctx, "external.error_simulation",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(25 * time.Millisecond)

	err := fmt.Errorf("external service returned 503 Service Unavailable")
	span.SetStatus(tracer.StatusCodeError, err.Error())
	span.SetAttribute("error", true)
	span.SetAttribute("http.status_code", 503)
	span.SetAttribute("external.service", "payment_api")

	return err
}
