package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/newrelic"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/prometheus"
)

func main() {
	// Setup multiple providers
	multiTracer, cleanup := setupMultiProviderTracing()
	defer cleanup()

	// Setup HTTP server with multi-provider tracing
	http.HandleFunc("/", handleRoot(multiTracer))
	http.HandleFunc("/api/orders", handleOrders(multiTracer))
	http.HandleFunc("/api/health", handleHealth(multiTracer))
	http.HandleFunc("/metrics", handlePrometheusMetrics())

	fmt.Println("Starting multi-provider instrumented server on :8080...")
	fmt.Println("This service sends traces to:")
	fmt.Println("  - Datadog APM (if configured)")
	fmt.Println("  - New Relic APM (if configured)")
	fmt.Println("  - Prometheus metrics at /metrics")
	fmt.Println("")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/api/orders")
	fmt.Println("  curl http://localhost:8080/api/health")
	fmt.Println("  curl http://localhost:8080/metrics")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupMultiProviderTracing() (tracer.Tracer, func()) {
	var tracers []tracer.Tracer
	var cleanupFuncs []func()

	// Setup Datadog provider
	if ddTracer, cleanup := setupDatadogTracer(); ddTracer != nil {
		tracers = append(tracers, ddTracer)
		cleanupFuncs = append(cleanupFuncs, cleanup)
		fmt.Println("✅ Datadog provider initialized")
	} else {
		fmt.Println("⚠️  Datadog provider skipped (check configuration)")
	}

	// Setup New Relic provider
	if nrTracer, cleanup := setupNewRelicTracer(); nrTracer != nil {
		tracers = append(tracers, nrTracer)
		cleanupFuncs = append(cleanupFuncs, cleanup)
		fmt.Println("✅ New Relic provider initialized")
	} else {
		fmt.Println("⚠️  New Relic provider skipped (check license key)")
	}

	// Setup Prometheus provider
	if promTracer, cleanup := setupPrometheusTracer(); promTracer != nil {
		tracers = append(tracers, promTracer)
		cleanupFuncs = append(cleanupFuncs, cleanup)
		fmt.Println("✅ Prometheus provider initialized")
	} else {
		fmt.Println("❌ Prometheus provider failed to initialize")
	}

	if len(tracers) == 0 {
		log.Fatal("No providers could be initialized")
	}

	// Create multi-provider tracer
	var multiTracer tracer.Tracer
	if len(tracers) == 1 {
		multiTracer = tracers[0]
	} else {
		multiTracer = tracer.NewMultiProviderTracer(tracers[0], tracers[1:]...)
	}

	// Combined cleanup function
	cleanup := func() {
		for i, cleanupFunc := range cleanupFuncs {
			fmt.Printf("Shutting down provider %d...\n", i+1)
			cleanupFunc()
		}
		fmt.Println("All providers shut down")
	}

	return multiTracer, cleanup
}

func setupDatadogTracer() (tracer.Tracer, func()) {
	config := &datadog.Config{
		ServiceName:        "multi-provider-example",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		AgentHost:          "localhost",
		AgentPort:          8126,
		SampleRate:         1.0,
		EnableProfiling:    false, // Disable for demo
		RuntimeMetrics:     true,
		AnalyticsEnabled:   true,
		Debug:              false,
		MaxTracesPerSecond: 1000,
		Tags: map[string]string{
			"provider":  "datadog",
			"example":   "multi-provider",
			"component": "api",
		},
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		fmt.Printf("Failed to create Datadog provider: %v\n", err)
		return nil, nil
	}

	tr, err := provider.CreateTracer("multi-api",
		tracer.WithServiceName("multi-provider-example"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		fmt.Printf("Failed to create Datadog tracer: %v\n", err)
		return nil, nil
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}

	return tr, cleanup
}

func setupNewRelicTracer() (tracer.Tracer, func()) {
	// Use a placeholder license key for demo - replace with real key
	licenseKey := "your-40-character-license-key-placeholder"

	// Skip if no real license key provided
	if licenseKey == "your-40-character-license-key-placeholder" {
		return nil, nil
	}

	config := &newrelic.Config{
		AppName:              "multi-provider-example",
		LicenseKey:           licenseKey,
		Environment:          "development",
		ServiceVersion:       "1.0.0",
		DistributedTracer:    true,
		Enabled:              true,
		CodeLevelMetrics:     true,
		LogLevel:             "info",
		DatastoreTracer:      true,
		CustomInsightsEvents: true,
		Labels: map[string]string{
			"provider":  "newrelic",
			"example":   "multi-provider",
			"component": "api",
		},
	}

	provider, err := newrelic.NewProvider(config)
	if err != nil {
		fmt.Printf("Failed to create New Relic provider: %v\n", err)
		return nil, nil
	}

	tr, err := provider.CreateTracer("multi-api",
		tracer.WithServiceName("multi-provider-example"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		fmt.Printf("Failed to create New Relic tracer: %v\n", err)
		return nil, nil
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}

	return tr, cleanup
}

func setupPrometheusTracer() (tracer.Tracer, func()) {
	config := &prometheus.Config{
		ServiceName:           "multi-provider-example",
		ServiceVersion:        "1.0.0",
		Environment:           "development",
		Namespace:             "multiapp",
		Subsystem:             "api",
		EnableDetailedMetrics: true,
		CustomLabels: map[string]string{
			"provider":  "prometheus",
			"example":   "multi-provider",
			"component": "api",
		},
		BucketBoundaries: []float64{
			0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
		},
		MaxCardinality:     1000,
		CollectionInterval: 30 * time.Second,
		UseGlobalRegistry:  false,
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		fmt.Printf("Failed to create Prometheus provider: %v\n", err)
		return nil, nil
	}

	tr, err := provider.CreateTracer("multi-api",
		tracer.WithServiceName("multi-provider-example"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		fmt.Printf("Failed to create Prometheus tracer: %v\n", err)
		return nil, nil
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		provider.Shutdown(ctx)
	}

	return tr, cleanup
}

// handleRoot demonstrates multi-provider tracing for a simple endpoint
func handleRoot(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start span that will be sent to all configured providers
		ctx, span := tr.StartSpan(r.Context(), "http.request.home",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method":     r.Method,
				"http.url":        r.URL.String(),
				"http.user_agent": r.UserAgent(),
				"endpoint":        "home",
				"multi_provider":  true,
				"trace_id":        generateTraceID(),
			}),
		)
		defer span.End()

		// Add custom attributes for different provider purposes
		span.SetAttribute("datadog.service", "web-frontend") // Datadog service mapping
		span.SetAttribute("newrelic.transaction", "web")     // New Relic transaction type
		span.SetAttribute("prometheus.endpoint", "home")     // Prometheus endpoint label

		// Simulate some processing
		processMultiProviderRequest(ctx, tr)

		// Set final attributes
		span.SetAttribute("http.status_code", 200)
		span.SetAttribute("response.size", 1024)
		span.SetAttribute("processing.success", true)
		span.SetStatus(tracer.StatusCodeOk, "Multi-provider request completed")

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Multi-Provider Tracing Example</title></head>
<body>
	<h1>Hello from Multi-Provider Instrumented Service!</h1>
	<p>This request is being tracked by multiple observability providers:</p>
	<ul>
		<li><strong>Datadog APM</strong> - Distributed tracing and performance monitoring</li>
		<li><strong>New Relic APM</strong> - Application performance monitoring and insights</li>
		<li><strong>Prometheus</strong> - Metrics collection and alerting</li>
	</ul>
	<p>Try other endpoints:</p>
	<ul>
		<li><a href="/api/orders">Orders API</a> - Complex business transaction</li>
		<li><a href="/api/health">Health Check</a> - Service health monitoring</li>
		<li><a href="/metrics">Prometheus Metrics</a> - Raw metrics data</li>
	</ul>
	<p><strong>Trace ID:</strong> `+generateTraceID()+`</p>
</body>
</html>`)
	}
}

// handleOrders demonstrates complex business logic tracing across multiple providers
func handleOrders(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.StartSpan(r.Context(), "api.orders.list",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method":    r.Method,
				"http.route":     "/api/orders",
				"api.version":    "v1",
				"operation.type": "list",
				"multi_provider": true,
			}),
		)
		defer span.End()

		// Provider-specific attributes
		span.SetAttribute("datadog.resource", "GET /api/orders")
		span.SetAttribute("newrelic.transaction.type", "business")
		span.SetAttribute("prometheus.operation", "orders_list")

		// Simulate user authentication
		userID, authenticated := authenticateUser(ctx, tr, r)
		span.SetAttribute("auth.user_id", userID)
		span.SetAttribute("auth.authenticated", authenticated)

		if !authenticated {
			span.SetStatus(tracer.StatusCodeError, "Authentication failed")
			span.SetAttribute("error", true)
			span.SetAttribute("error.type", "authentication_failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch orders from database
		orders, totalValue, err := fetchOrdersMultiProvider(ctx, tr, userID)
		if err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Database error: %v", err))
			span.SetAttribute("error", true)
			span.SetAttribute("error.type", "database_error")
			span.SetAttribute("error.message", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Record business metrics for all providers
		span.SetAttribute("orders.count", len(orders))
		span.SetAttribute("orders.total_value", totalValue)
		span.SetAttribute("business.revenue_impact", totalValue)
		span.SetAttribute("user.tier", "premium")

		// Provider-specific business attributes
		span.SetAttribute("datadog.analytics", true)
		span.SetAttribute("newrelic.custom_attribute", "high_value_transaction")
		span.SetAttribute("prometheus.business_metric", totalValue)

		// Call external services
		paymentStatus := checkPaymentServiceMultiProvider(ctx, tr)
		shippingEstimate := calculateShippingMultiProvider(ctx, tr)

		span.SetAttribute("payment.status", paymentStatus)
		span.SetAttribute("shipping.estimate_days", shippingEstimate)

		span.SetAttribute("http.status_code", 200)
		span.SetStatus(tracer.StatusCodeOk, "Orders retrieved successfully")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"user_id": "%s",
	"orders": %d,
	"total_value": %.2f,
	"payment_status": "%s",
	"shipping_estimate_days": %d,
	"trace_providers": ["datadog", "newrelic", "prometheus"],
	"data": %v
}`, userID, len(orders), totalValue, paymentStatus, shippingEstimate, orders)
	}
}

// handleHealth demonstrates health check monitoring across all providers
func handleHealth(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.StartSpan(r.Context(), "api.health.check",
			tracer.WithSpanKind(tracer.SpanKindServer),
		)
		defer span.End()

		// Health check components
		healthStatus := checkSystemHealth(ctx, tr)

		// Set health metrics for all providers
		span.SetAttribute("health.overall", healthStatus.Overall)
		span.SetAttribute("health.database", healthStatus.Database)
		span.SetAttribute("health.cache", healthStatus.Cache)
		span.SetAttribute("health.external_services", healthStatus.ExternalServices)

		// Provider-specific health attributes
		span.SetAttribute("datadog.health_score", calculateHealthScore(healthStatus))
		span.SetAttribute("newrelic.availability", healthStatus.Overall)
		span.SetAttribute("prometheus.health_status", boolToInt(healthStatus.Overall))

		if healthStatus.Overall {
			span.SetStatus(tracer.StatusCodeOk, "All systems healthy")
			w.WriteHeader(http.StatusOK)
		} else {
			span.SetStatus(tracer.StatusCodeError, "Some systems unhealthy")
			span.SetAttribute("error", true)
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"status": "%s",
	"timestamp": %d,
	"components": {
		"database": %t,
		"cache": %t,
		"external_services": %t
	},
	"health_score": %.2f,
	"trace_providers": ["datadog", "newrelic", "prometheus"]
}`, map[bool]string{true: "healthy", false: "unhealthy"}[healthStatus.Overall],
			time.Now().Unix(),
			healthStatus.Database,
			healthStatus.Cache,
			healthStatus.ExternalServices,
			calculateHealthScore(healthStatus))
	}
}

// handlePrometheusMetrics serves Prometheus metrics
func handlePrometheusMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, `# Multi-provider metrics example
# In a real implementation, this would be handled by prometheus/promhttp

# HELP multiapp_api_requests_total Total API requests across all providers
# TYPE multiapp_api_requests_total counter
multiapp_api_requests_total{endpoint="home",status="200"} 42
multiapp_api_requests_total{endpoint="orders",status="200"} 23
multiapp_api_requests_total{endpoint="health",status="200"} 156

# HELP multiapp_api_request_duration_seconds Request duration
# TYPE multiapp_api_request_duration_seconds histogram
multiapp_api_request_duration_seconds_bucket{endpoint="orders",le="0.1"} 5
multiapp_api_request_duration_seconds_bucket{endpoint="orders",le="0.5"} 18
multiapp_api_request_duration_seconds_bucket{endpoint="orders",le="1.0"} 22
multiapp_api_request_duration_seconds_bucket{endpoint="orders",le="+Inf"} 23

# HELP multiapp_api_health_status System health status
# TYPE multiapp_api_health_status gauge
multiapp_api_health_status{component="database"} 1
multiapp_api_health_status{component="cache"} 1
multiapp_api_health_status{component="external_services"} 1
`)
	}
}

// Helper functions and types

type HealthStatus struct {
	Overall          bool
	Database         bool
	Cache            bool
	ExternalServices bool
}

func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func processMultiProviderRequest(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "request.processing",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(25 * time.Millisecond)

	span.SetAttribute("processing.duration_ms", 25)
	span.SetAttribute("processing.complexity", "low")
	span.SetAttribute("multi_provider", true)
	span.SetStatus(tracer.StatusCodeOk, "Processing completed")
}

func authenticateUser(ctx context.Context, tr tracer.Tracer, r *http.Request) (string, bool) {
	_, span := tr.StartSpan(ctx, "auth.user_validation",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(15 * time.Millisecond)

	authHeader := r.Header.Get("Authorization")
	userID := "user_12345"
	authenticated := authHeader != "" || r.URL.Query().Get("auth") == "true"

	span.SetAttribute("auth.method", "header_token")
	span.SetAttribute("auth.duration_ms", 15)
	span.SetAttribute("user.id", userID)
	span.SetAttribute("authenticated", authenticated)

	if authenticated {
		span.SetStatus(tracer.StatusCodeOk, "User authenticated")
	} else {
		span.SetStatus(tracer.StatusCodeError, "Authentication failed")
	}

	return userID, authenticated
}

func fetchOrdersMultiProvider(ctx context.Context, tr tracer.Tracer, userID string) ([]string, float64, error) {
	_, span := tr.StartSpan(ctx, "database.orders.fetch",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db.system":    "postgresql",
			"db.operation": "SELECT",
			"db.table":     "orders",
			"user.id":      userID,
		}),
	)
	defer span.End()

	time.Sleep(60 * time.Millisecond)

	// Simulate occasional errors
	if time.Now().UnixNano()%25 == 0 {
		err := fmt.Errorf("database connection timeout")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.SetAttribute("error", true)
		span.SetAttribute("error.type", "timeout")
		return nil, 0, err
	}

	orders := []string{"order_001", "order_002", "order_003", "order_004"}
	totalValue := 459.97

	span.SetAttribute("db.rows_returned", len(orders))
	span.SetAttribute("orders.total_value", totalValue)
	span.SetAttribute("db.query_duration_ms", 60)
	span.SetAttribute("cache.hit", false)
	span.SetStatus(tracer.StatusCodeOk, "Orders fetched successfully")

	return orders, totalValue, nil
}

func checkPaymentServiceMultiProvider(ctx context.Context, tr tracer.Tracer) string {
	_, span := tr.StartSpan(ctx, "external.payment.health",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"external.service": "payment_gateway",
			"external.vendor":  "stripe",
		}),
	)
	defer span.End()

	time.Sleep(40 * time.Millisecond)

	status := "operational"
	span.SetAttribute("payment.status", status)
	span.SetAttribute("external.response_time_ms", 40)
	span.SetAttribute("external.endpoint", "/health")
	span.SetStatus(tracer.StatusCodeOk, "Payment service healthy")

	return status
}

func calculateShippingMultiProvider(ctx context.Context, tr tracer.Tracer) int {
	_, span := tr.StartSpan(ctx, "logistics.shipping.estimate",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(20 * time.Millisecond)

	shippingDays := 3
	span.SetAttribute("shipping.estimate_days", shippingDays)
	span.SetAttribute("shipping.calculation_duration_ms", 20)
	span.SetAttribute("logistics.priority", "standard")
	span.SetStatus(tracer.StatusCodeOk, "Shipping estimate calculated")

	return shippingDays
}

func checkSystemHealth(ctx context.Context, tr tracer.Tracer) HealthStatus {
	_, span := tr.StartSpan(ctx, "health.system_check",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Check individual components
	dbHealth := checkDatabaseHealth(ctx, tr)
	cacheHealth := checkCacheHealth(ctx, tr)
	externalHealth := checkExternalServicesHealth(ctx, tr)

	overall := dbHealth && cacheHealth && externalHealth

	healthStatus := HealthStatus{
		Overall:          overall,
		Database:         dbHealth,
		Cache:            cacheHealth,
		ExternalServices: externalHealth,
	}

	span.SetAttribute("health.database", dbHealth)
	span.SetAttribute("health.cache", cacheHealth)
	span.SetAttribute("health.external_services", externalHealth)
	span.SetAttribute("health.overall", overall)
	span.SetAttribute("health.check_duration_ms", 60)

	if overall {
		span.SetStatus(tracer.StatusCodeOk, "System health check passed")
	} else {
		span.SetStatus(tracer.StatusCodeError, "System health check failed")
	}

	return healthStatus
}

func checkDatabaseHealth(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.database",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(20 * time.Millisecond)
	healthy := true

	span.SetAttribute("db.ping_duration_ms", 20)
	span.SetAttribute("db.connections_active", 5)
	span.SetAttribute("db.connections_max", 100)
	span.SetStatus(tracer.StatusCodeOk, "Database healthy")

	return healthy
}

func checkCacheHealth(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.cache",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(10 * time.Millisecond)
	healthy := true

	span.SetAttribute("cache.ping_duration_ms", 10)
	span.SetAttribute("cache.hit_rate", 0.95)
	span.SetAttribute("cache.memory_usage_percent", 0.67)
	span.SetStatus(tracer.StatusCodeOk, "Cache healthy")

	return healthy
}

func checkExternalServicesHealth(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.external_services",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(30 * time.Millisecond)
	healthy := true

	span.SetAttribute("external.services_checked", 3)
	span.SetAttribute("external.services_healthy", 3)
	span.SetAttribute("external.check_duration_ms", 30)
	span.SetStatus(tracer.StatusCodeOk, "External services healthy")

	return healthy
}

func calculateHealthScore(health HealthStatus) float64 {
	score := 0.0
	if health.Database {
		score += 0.4
	}
	if health.Cache {
		score += 0.3
	}
	if health.ExternalServices {
		score += 0.3
	}
	return score
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
