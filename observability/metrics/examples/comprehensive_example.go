// Package examples demonstrates usage of the metrics library with different providers
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
	"github.com/fsvxavier/nexs-lib/observability/metrics/providers/datadog"
	"github.com/fsvxavier/nexs-lib/observability/metrics/providers/newrelic"
	"github.com/fsvxavier/nexs-lib/observability/metrics/providers/prometheus"
)

// MetricsCollector holds all metrics for the application
type MetricsCollector struct {
	// HTTP metrics
	httpRequests      metrics.Counter
	httpDuration      metrics.Histogram
	httpResponseTime  metrics.Summary
	activeConnections metrics.Gauge

	// Business metrics
	ordersProcessed    metrics.Counter
	queueSize          metrics.Gauge
	processingDuration metrics.Histogram
}

// setupPrometheusMetrics configures Prometheus metrics provider
func setupPrometheusMetrics() (metrics.Provider, *MetricsCollector, error) {
	config := metrics.PrometheusConfig{
		Prefix: "myapp",
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create prometheus provider: %w", err)
	}
	collector := &MetricsCollector{}

	// Create HTTP metrics
	collector.httpRequests, err = provider.CreateCounter(metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
			Labels:    []string{"method", "status", "endpoint"},
			Namespace: "http",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_requests counter: %w", err)
	}

	collector.httpDuration, err = provider.CreateHistogram(metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Labels:    []string{"method", "endpoint"},
			Namespace: "http",
		},
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_duration histogram: %w", err)
	}

	collector.activeConnections, err = provider.CreateGauge(metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "active_connections",
			Help:      "Number of active connections",
			Namespace: "http",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create active_connections gauge: %w", err)
	}

	collector.httpResponseTime, err = provider.CreateSummary(metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "http_response_time_seconds",
			Help:      "HTTP response time summary",
			Labels:    []string{"method", "endpoint"},
			Namespace: "http",
		},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_response_time summary: %w", err)
	}

	// Create business metrics
	collector.ordersProcessed, err = provider.CreateCounter(metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "orders_processed_total",
			Help:      "Total number of orders processed",
			Labels:    []string{"status", "payment_method"},
			Namespace: "business",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create orders_processed counter: %w", err)
	}

	collector.queueSize, err = provider.CreateGauge(metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "queue_size",
			Help:      "Current queue size",
			Labels:    []string{"queue_name"},
			Namespace: "queue",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create queue_size gauge: %w", err)
	}

	collector.processingDuration, err = provider.CreateHistogram(metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "processing_duration_seconds",
			Help:      "Order processing duration in seconds",
			Labels:    []string{"order_type"},
			Namespace: "business",
		},
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create processing_duration histogram: %w", err)
	}

	return provider, collector, nil
}

// setupDataDogMetrics configures DataDog metrics provider
func setupDataDogMetrics() (metrics.Provider, *MetricsCollector, error) {
	config := &datadog.Config{
		APIKey: "test-api-key",
		Tags: []string{
			"service:my-service",
			"environment:development",
			"version:1.0.0",
		},
		Namespace: "myapp",
	}

	// Use mock client for demo
	mockClient := datadog.NewMockDataDogClient()
	provider := datadog.NewProvider(config, mockClient)
	collector := &MetricsCollector{}

	var err error

	// Create HTTP metrics
	collector.httpRequests, err = provider.CreateCounter(metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "http.requests.total",
			Help:   "Total number of HTTP requests",
			Labels: []string{"method", "status", "endpoint"},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_requests counter: %w", err)
	}

	collector.httpDuration, err = provider.CreateHistogram(metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "http.request.duration",
			Help:   "HTTP request duration",
			Labels: []string{"method", "endpoint"},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_duration histogram: %w", err)
	}

	collector.activeConnections, err = provider.CreateGauge(metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "http.active_connections",
			Help: "Number of active connections",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create active_connections gauge: %w", err)
	}

	return provider, collector, nil
}

// setupNewRelicMetrics configures NewRelic metrics provider
func setupNewRelicMetrics() (metrics.Provider, *MetricsCollector, error) {
	config := &newrelic.Config{
		LicenseKey:  "test-license-key",
		AppName:     "my-app",
		Environment: "development",
		Attributes: map[string]interface{}{
			"version": "1.0.0",
			"team":    "backend",
		},
	}

	// Use mock client for demo
	mockClient := newrelic.NewMockNewRelicClient()
	provider := newrelic.NewProvider(config, mockClient)
	collector := &MetricsCollector{}

	var err error

	// Create HTTP metrics
	collector.httpRequests, err = provider.CreateCounter(metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "http_requests_total",
			Help:   "Total number of HTTP requests",
			Labels: []string{"method", "status", "endpoint"},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_requests counter: %w", err)
	}

	collector.httpDuration, err = provider.CreateHistogram(metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "http_request_duration",
			Help:   "HTTP request duration",
			Labels: []string{"method", "endpoint"},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_duration histogram: %w", err)
	}

	collector.activeConnections, err = provider.CreateGauge(metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create active_connections gauge: %w", err)
	}

	collector.httpResponseTime, err = provider.CreateSummary(metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "http_response_time",
			Help:   "HTTP response time summary",
			Labels: []string{"method", "endpoint"},
		},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http_response_time summary: %w", err)
	}

	return provider, collector, nil
}

// simulateHTTPRequests demonstrates metric collection for HTTP requests
func simulateHTTPRequests(collector *MetricsCollector) {
	endpoints := []string{"/api/users", "/api/orders", "/health"}
	methods := []string{"GET", "POST", "PUT"}
	statuses := []string{"200", "201", "400", "500"}

	for i := 0; i < 100; i++ {
		endpoint := endpoints[i%len(endpoints)]
		method := methods[i%len(methods)]
		status := statuses[i%len(statuses)]

		// Simulate request processing time
		start := time.Now()
		time.Sleep(time.Duration(i%50) * time.Millisecond)
		duration := time.Since(start).Seconds()

		// Record metrics
		collector.httpRequests.Inc(method, status, endpoint)
		collector.httpDuration.Observe(duration, method, endpoint)

		if collector.httpResponseTime != nil {
			collector.httpResponseTime.Observe(duration, method, endpoint)
		}

		// Update active connections (simulate fluctuation)
		if i%10 == 0 {
			collector.activeConnections.Set(float64(10+i%5), endpoint)
		}
	}
}

// simulateBusinessMetrics demonstrates business metric collection
func simulateBusinessMetrics(collector *MetricsCollector) {
	paymentMethods := []string{"credit_card", "paypal", "bank_transfer"}
	orderTypes := []string{"standard", "express", "priority"}
	queueNames := []string{"orders", "notifications", "exports"}

	for i := 0; i < 50; i++ {
		paymentMethod := paymentMethods[i%len(paymentMethods)]
		orderType := orderTypes[i%len(orderTypes)]
		queueName := queueNames[i%len(queueNames)]

		// Simulate order processing
		status := "completed"
		if i%10 == 0 {
			status = "failed"
		}

		if collector.ordersProcessed != nil {
			collector.ordersProcessed.Inc(status, paymentMethod)
		}

		// Simulate processing duration
		processingTime := float64(1+i%10) * 0.5
		if collector.processingDuration != nil {
			collector.processingDuration.Observe(processingTime, orderType)
		}

		// Update queue sizes
		if collector.queueSize != nil {
			queueLen := float64(i % 20)
			collector.queueSize.Set(queueLen, queueName)
		}
	}
}

// demonstrateProvider shows metrics usage with a specific provider
func demonstrateProvider(name string, provider metrics.Provider, collector *MetricsCollector) {
	fmt.Printf("\n=== Demonstrating %s Provider ===\n", name)

	// Record some metrics
	fmt.Println("Recording HTTP metrics...")
	simulateHTTPRequests(collector)

	fmt.Println("Recording business metrics...")
	simulateBusinessMetrics(collector)

	// Show some values (note: some providers may return 0 for Get methods)
	fmt.Printf("HTTP requests (GET, 200, /api/users): %.0f\n", collector.httpRequests.Get("GET", "200", "/api/users"))
	fmt.Printf("Active connections (/health): %.0f\n", collector.activeConnections.Get("/health"))

	// Demonstrate timer usage
	fmt.Println("Demonstrating timer functionality...")
	timer := collector.httpDuration.StartTimer("GET", "/api/timer-test")
	time.Sleep(100 * time.Millisecond)
	timer() // Stop timer and record observation

	// Demonstrate function timing
	collector.httpDuration.Time(func() {
		fmt.Println("Processing simulated work...")
		time.Sleep(50 * time.Millisecond)
	}, "POST", "/api/work")

	fmt.Printf("Provider: %s completed successfully\n", provider.Name())
}

func main() {
	fmt.Println("Comprehensive Metrics Library Demo")
	fmt.Println("===================================")

	// Setup and demonstrate Prometheus
	prometheusProvider, prometheusCollector, err := setupPrometheusMetrics()
	if err != nil {
		log.Fatalf("Failed to setup Prometheus: %v", err)
	}
	demonstrateProvider("Prometheus", prometheusProvider, prometheusCollector)

	// Setup and demonstrate DataDog
	datadogProvider, datadogCollector, err := setupDataDogMetrics()
	if err != nil {
		log.Fatalf("Failed to setup DataDog: %v", err)
	}
	demonstrateProvider("DataDog", datadogProvider, datadogCollector)

	// Setup and demonstrate NewRelic
	newrelicProvider, newrelicCollector, err := setupNewRelicMetrics()
	if err != nil {
		log.Fatalf("Failed to setup NewRelic: %v", err)
	}
	demonstrateProvider("NewRelic", newrelicProvider, newrelicCollector)

	// Demonstrate HTTP handler integration
	fmt.Println("\n=== HTTP Handler Integration Demo ===")
	httpHandler := createHTTPHandler(prometheusCollector)

	// Simulate some HTTP requests
	fmt.Println("Simulating HTTP server requests...")
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/api/test", nil)
		httpHandler.ServeHTTP(nil, req)
	}

	// Graceful shutdown demonstration
	fmt.Println("\n=== Graceful Shutdown Demo ===")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := prometheusProvider.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down Prometheus: %v\n", err)
	} else {
		fmt.Println("Prometheus provider shut down successfully")
	}

	if err := datadogProvider.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down DataDog: %v\n", err)
	} else {
		fmt.Println("DataDog provider shut down successfully")
	}

	if err := newrelicProvider.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down NewRelic: %v\n", err)
	} else {
		fmt.Println("NewRelic provider shut down successfully")
	}

	fmt.Println("\nDemo completed successfully!")
}

// createHTTPHandler demonstrates metrics integration with HTTP handlers
func createHTTPHandler(collector *MetricsCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Simulate processing
		time.Sleep(time.Duration(10+len(r.URL.Path)) * time.Millisecond)

		// Record metrics
		duration := time.Since(start).Seconds()
		collector.httpRequests.Inc(r.Method, "200", r.URL.Path)
		collector.httpDuration.Observe(duration, r.Method, r.URL.Path)

		if collector.httpResponseTime != nil {
			collector.httpResponseTime.Observe(duration, r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
