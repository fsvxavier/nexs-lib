package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/prometheus"
)

func main() {
	// Configure Prometheus provider
	config := &prometheus.Config{
		ServiceName:           "example-prometheus-service",
		ServiceVersion:        "1.0.0",
		Environment:           "development",
		Namespace:             "myapp",
		Subsystem:             "api",
		EnableDetailedMetrics: true,
		CustomLabels: map[string]string{
			"team":      "backend",
			"region":    "us-east-1",
			"component": "api-server",
		},
		BucketBoundaries: []float64{
			0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
		},
		MaxCardinality:     1000,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		BatchSize:          100,
		UseGlobalRegistry:  false,
	}

	// Create provider
	provider, err := prometheus.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create Prometheus provider: %v", err)
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
		tracer.WithServiceName("example-prometheus-service"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Setup HTTP server with tracing
	http.HandleFunc("/", handleRoot(tr))
	http.HandleFunc("/api/products", handleProducts(tr))
	http.HandleFunc("/api/inventory", handleInventory(tr))
	http.HandleFunc("/api/slow", handleSlowEndpoint(tr))
	http.HandleFunc("/metrics", handlePrometheusMetrics()) // Prometheus metrics endpoint

	fmt.Println("Starting Prometheus instrumented server on :8080...")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/api/products")
	fmt.Println("  curl http://localhost:8080/api/inventory")
	fmt.Println("  curl http://localhost:8080/api/slow")
	fmt.Println("  curl http://localhost:8080/metrics  # Prometheus metrics")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleRoot demonstrates basic metrics collection
func handleRoot(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Start root span for metrics collection
		ctx, span := tr.StartSpan(r.Context(), "http_request_home",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http_method":     r.Method,
				"http_path":       r.URL.Path,
				"http_user_agent": r.UserAgent(),
				"endpoint":        "home",
				"request_size":    r.ContentLength,
			}),
		)
		defer func() {
			duration := time.Since(startTime)
			span.SetAttribute("duration_seconds", duration.Seconds())
			span.SetAttribute("response_size", 500) // Approximate HTML size
			span.End()
		}()

		// Simulate some processing
		processHomePageRequest(ctx, tr)

		// Set success metrics
		span.SetAttribute("http_status_code", 200)
		span.SetAttribute("success", true)
		span.SetStatus(tracer.StatusCodeOk, "Home page served successfully")

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Prometheus Metrics Example</title></head>
<body>
	<h1>Hello from Prometheus Instrumented Service!</h1>
	<p>This service is collecting metrics for Prometheus.</p>
	<p>Check the metrics at <a href="/metrics">/metrics</a></p>
	<p>Try other endpoints:</p>
	<ul>
		<li><a href="/api/products">Products API</a></li>
		<li><a href="/api/inventory">Inventory API</a></li>
		<li><a href="/api/slow">Slow Endpoint</a></li>
	</ul>
</body>
</html>`)
	}
}

// handleProducts demonstrates business metrics collection
func handleProducts(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		ctx, span := tr.StartSpan(r.Context(), "api_products_list",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http_method":    r.Method,
				"http_path":      r.URL.Path,
				"api_version":    "v1",
				"operation_type": "read",
			}),
		)
		defer func() {
			duration := time.Since(startTime)
			span.SetAttribute("duration_seconds", duration.Seconds())
			span.End()
		}()

		// Simulate authentication
		authenticated := authenticateAPIRequest(ctx, tr, r)
		span.SetAttribute("authenticated", authenticated)

		if !authenticated {
			span.SetAttribute("http_status_code", 401)
			span.SetAttribute("error", true)
			span.SetAttribute("error_type", "authentication_failed")
			span.SetStatus(tracer.StatusCodeError, "Authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch products from database
		products, err := fetchProductsFromDatabase(ctx, tr)
		if err != nil {
			span.SetAttribute("http_status_code", 500)
			span.SetAttribute("error", true)
			span.SetAttribute("error_type", "database_error")
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Database error: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Record business metrics
		span.SetAttribute("products_count", len(products))
		span.SetAttribute("cache_hit", false) // Simulated cache miss
		span.SetAttribute("database_queries", 1)
		span.SetAttribute("http_status_code", 200)
		span.SetAttribute("response_size", len(products)*50) // Estimated response size

		span.SetStatus(tracer.StatusCodeOk, "Products retrieved successfully")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"products": %d,
	"data": %v,
	"timestamp": %d,
	"cache_hit": false
}`, len(products), products, time.Now().Unix())
	}
}

// handleInventory demonstrates complex business operations with multiple metrics
func handleInventory(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		ctx, span := tr.StartSpan(r.Context(), "api_inventory_check",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http_method":    r.Method,
				"http_path":      r.URL.Path,
				"operation_type": "inventory_check",
			}),
		)
		defer func() {
			duration := time.Since(startTime)
			span.SetAttribute("duration_seconds", duration.Seconds())
			span.End()
		}()

		// Get product ID from query params
		productID := r.URL.Query().Get("product_id")
		if productID == "" {
			productID = "default_product"
		}
		span.SetAttribute("product_id", productID)

		// Check inventory levels
		inventory, inStock := checkInventoryLevels(ctx, tr, productID)
		span.SetAttribute("inventory_level", inventory)
		span.SetAttribute("in_stock", inStock)
		span.SetAttribute("low_stock_threshold", 10)

		// Check warehouse locations
		warehouses := checkWarehouseAvailability(ctx, tr, productID)
		span.SetAttribute("warehouses_checked", len(warehouses))
		span.SetAttribute("warehouses_available", countAvailableWarehouses(warehouses))

		// Calculate shipping estimates
		shippingDays := calculateShippingEstimate(ctx, tr, warehouses)
		span.SetAttribute("shipping_estimate_days", shippingDays)

		// Record business metrics
		if inventory < 10 {
			span.SetAttribute("low_stock_alert", true)
		}

		span.SetAttribute("http_status_code", 200)
		span.SetStatus(tracer.StatusCodeOk, "Inventory check completed")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"product_id": "%s",
	"inventory_level": %d,
	"in_stock": %t,
	"warehouses": %v,
	"shipping_estimate_days": %d,
	"low_stock_alert": %t,
	"timestamp": %d
}`, productID, inventory, inStock, warehouses, shippingDays, inventory < 10, time.Now().Unix())
	}
}

// handleSlowEndpoint demonstrates performance metrics for slow operations
func handleSlowEndpoint(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		ctx, span := tr.StartSpan(r.Context(), "api_slow_operation",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http_method":    r.Method,
				"http_path":      r.URL.Path,
				"operation_type": "heavy_computation",
			}),
		)
		defer func() {
			duration := time.Since(startTime)
			span.SetAttribute("duration_seconds", duration.Seconds())
			span.End()
		}()

		// Simulate slow processing phases
		performHeavyComputation(ctx, tr)
		processLargeDataset(ctx, tr)
		generateComplexReport(ctx, tr)

		// Check if operation exceeded SLA
		duration := time.Since(startTime)
		slaThreshold := 3 * time.Second
		slaViolation := duration > slaThreshold

		span.SetAttribute("sla_threshold_seconds", slaThreshold.Seconds())
		span.SetAttribute("sla_violation", slaViolation)
		span.SetAttribute("performance_tier", "slow")

		if slaViolation {
			span.SetAttribute("alert_triggered", true)
		}

		span.SetAttribute("http_status_code", 200)
		span.SetStatus(tracer.StatusCodeOk, "Slow operation completed")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"operation": "heavy_computation",
	"duration_seconds": %.3f,
	"sla_threshold_seconds": %.1f,
	"sla_violation": %t,
	"timestamp": %d
}`, duration.Seconds(), slaThreshold.Seconds(), slaViolation, time.Now().Unix())
	}
}

// handlePrometheusMetrics serves the Prometheus metrics endpoint
func handlePrometheusMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would be handled by prometheus/promhttp
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, `# This would normally be handled by prometheus/promhttp.Handler()
# Example metrics that would be generated:

# HELP myapp_api_requests_total Total number of API requests
# TYPE myapp_api_requests_total counter
myapp_api_requests_total{method="GET",endpoint="home",status="200"} 42

# HELP myapp_api_request_duration_seconds Request duration in seconds
# TYPE myapp_api_request_duration_seconds histogram
myapp_api_request_duration_seconds_bucket{endpoint="products",le="0.1"} 10
myapp_api_request_duration_seconds_bucket{endpoint="products",le="0.5"} 15
myapp_api_request_duration_seconds_bucket{endpoint="products",le="1.0"} 18
myapp_api_request_duration_seconds_bucket{endpoint="products",le="+Inf"} 20

# HELP myapp_api_errors_total Total number of API errors
# TYPE myapp_api_errors_total counter
myapp_api_errors_total{endpoint="products",error_type="database_error"} 2

# HELP myapp_api_inventory_level Current inventory levels
# TYPE myapp_api_inventory_level gauge
myapp_api_inventory_level{product_id="default_product"} 25
`)
	}
}

// Helper functions for business logic simulation

func processHomePageRequest(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "home_page_render",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(30 * time.Millisecond)
	span.SetAttribute("template_name", "home.html")
	span.SetAttribute("render_time_ms", 30)
	span.SetStatus(tracer.StatusCodeOk, "Home page rendered")
}

func authenticateAPIRequest(ctx context.Context, tr tracer.Tracer, r *http.Request) bool {
	_, span := tr.StartSpan(ctx, "api_authentication",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(20 * time.Millisecond)

	apiKey := r.Header.Get("X-API-Key")
	authenticated := apiKey != ""

	span.SetAttribute("auth_method", "api_key")
	span.SetAttribute("auth_duration_ms", 20)
	span.SetAttribute("authenticated", authenticated)

	if authenticated {
		span.SetStatus(tracer.StatusCodeOk, "Authentication successful")
	} else {
		span.SetStatus(tracer.StatusCodeError, "Authentication failed")
	}

	return authenticated
}

func fetchProductsFromDatabase(ctx context.Context, tr tracer.Tracer) ([]string, error) {
	_, span := tr.StartSpan(ctx, "database_query_products",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db_system":    "postgresql",
			"db_operation": "SELECT",
			"db_table":     "products",
		}),
	)
	defer span.End()

	time.Sleep(75 * time.Millisecond)

	// Simulate occasional database issues
	if time.Now().UnixNano()%20 == 0 {
		err := fmt.Errorf("database query timeout")
		span.SetAttribute("error", true)
		span.SetAttribute("error_type", "timeout")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return nil, err
	}

	products := []string{"laptop", "mouse", "keyboard", "monitor", "headphones"}
	span.SetAttribute("db_rows_returned", len(products))
	span.SetAttribute("db_query_duration_ms", 75)
	span.SetAttribute("cache_hit", false)
	span.SetStatus(tracer.StatusCodeOk, "Products fetched successfully")

	return products, nil
}

func checkInventoryLevels(ctx context.Context, tr tracer.Tracer, productID string) (int, bool) {
	_, span := tr.StartSpan(ctx, "inventory_level_check",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"product_id": productID,
		}),
	)
	defer span.End()

	time.Sleep(40 * time.Millisecond)

	// Simulate inventory levels
	inventory := 25
	inStock := inventory > 0

	span.SetAttribute("inventory_count", inventory)
	span.SetAttribute("in_stock", inStock)
	span.SetAttribute("check_duration_ms", 40)
	span.SetStatus(tracer.StatusCodeOk, "Inventory checked")

	return inventory, inStock
}

func checkWarehouseAvailability(ctx context.Context, tr tracer.Tracer, productID string) []map[string]interface{} {
	_, span := tr.StartSpan(ctx, "warehouse_availability_check",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"product_id": productID,
		}),
	)
	defer span.End()

	time.Sleep(60 * time.Millisecond)

	warehouses := []map[string]interface{}{
		{"location": "us-east", "available": true, "stock": 15},
		{"location": "us-west", "available": true, "stock": 8},
		{"location": "eu-central", "available": false, "stock": 0},
	}

	span.SetAttribute("warehouses_checked", len(warehouses))
	span.SetAttribute("check_duration_ms", 60)
	span.SetStatus(tracer.StatusCodeOk, "Warehouse availability checked")

	return warehouses
}

func countAvailableWarehouses(warehouses []map[string]interface{}) int {
	count := 0
	for _, wh := range warehouses {
		if wh["available"].(bool) {
			count++
		}
	}
	return count
}

func calculateShippingEstimate(ctx context.Context, tr tracer.Tracer, warehouses []map[string]interface{}) int {
	_, span := tr.StartSpan(ctx, "shipping_estimate_calculation",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(25 * time.Millisecond)

	// Simple shipping calculation
	shippingDays := 3 // Default shipping time
	if countAvailableWarehouses(warehouses) > 1 {
		shippingDays = 2 // Faster if multiple warehouses available
	}

	span.SetAttribute("shipping_days", shippingDays)
	span.SetAttribute("calculation_duration_ms", 25)
	span.SetStatus(tracer.StatusCodeOk, "Shipping estimate calculated")

	return shippingDays
}

func performHeavyComputation(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "heavy_computation_phase1",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(800 * time.Millisecond)
	span.SetAttribute("computation_type", "matrix_multiplication")
	span.SetAttribute("computation_complexity", "O(n^3)")
	span.SetAttribute("computation_duration_ms", 800)
	span.SetStatus(tracer.StatusCodeOk, "Heavy computation completed")
}

func processLargeDataset(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "large_dataset_processing",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(1200 * time.Millisecond)
	span.SetAttribute("dataset_size_mb", 500)
	span.SetAttribute("processing_type", "aggregation")
	span.SetAttribute("memory_usage_mb", 150)
	span.SetAttribute("processing_duration_ms", 1200)
	span.SetStatus(tracer.StatusCodeOk, "Dataset processing completed")
}

func generateComplexReport(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "complex_report_generation",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	time.Sleep(600 * time.Millisecond)
	span.SetAttribute("report_type", "financial_summary")
	span.SetAttribute("report_pages", 25)
	span.SetAttribute("charts_generated", 8)
	span.SetAttribute("generation_duration_ms", 600)
	span.SetStatus(tracer.StatusCodeOk, "Report generation completed")
}
