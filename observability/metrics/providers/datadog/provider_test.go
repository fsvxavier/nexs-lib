package datadog

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
)

func TestNewProvider(t *testing.T) {
	config := &Config{
		APIKey:      "test-api-key",
		AppKey:      "test-app-key",
		Environment: "test",
		Service:     "test-service",
	}
	client := NewMockDataDogClient()

	provider := NewProvider(config, client)
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
	if provider.Name() != "datadog" {
		t.Errorf("expected name 'datadog', got %s", provider.Name())
	}
}

func TestProvider_CreateCounter(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_counter",
			Help:      "Test counter",
			Labels:    []string{"method", "status"},
			Namespace: "http",
		},
	}

	// Create counter
	counter, err := provider.CreateCounter(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counter == nil {
		t.Fatal("expected counter, got nil")
	}

	// Test counter operations
	counter.Inc("GET", "200")
	counter.Add(5, "POST", "201")

	// Check that metrics were sent to the client
	metrics := client.GetMetrics()
	if len(metrics) < 2 {
		t.Errorf("expected at least 2 metrics, got %d", len(metrics))
	}
}

func TestProvider_CreateHistogram(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "test_histogram",
			Help: "Test histogram",
		},
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0},
	}

	// Create histogram
	histogram, err := provider.CreateHistogram(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if histogram == nil {
		t.Fatal("expected histogram, got nil")
	}

	// Test histogram operations
	histogram.Observe(0.3)
	histogram.Observe(1.2)
	histogram.Observe(3.1)

	// Test timer functionality
	timer := histogram.StartTimer()
	timer()

	// Test Time function
	histogram.Time(func() {
		// Simulate some work
	})
}

func TestProvider_CreateGauge(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "test_gauge",
			Help: "Test gauge",
		},
	}

	// Create gauge
	gauge, err := provider.CreateGauge(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gauge == nil {
		t.Fatal("expected gauge, got nil")
	}

	// Test gauge operations
	gauge.Set(42.5)
	gauge.Inc()
	gauge.Dec()
	gauge.Add(10)
	gauge.Sub(5)
	gauge.SetToCurrentTime()
}

func TestProvider_CreateSummary(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "test_summary",
			Help: "Test summary",
		},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}

	// Create summary
	summary, err := provider.CreateSummary(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}

	// Test summary operations
	values := []float64{1.2, 3.4, 0.8, 2.1, 4.5}
	for _, v := range values {
		summary.Observe(v)
	}

	// Test timer functionality
	timer := summary.StartTimer()
	timer()
}

func TestProvider_Shutdown(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	ctx := context.Background()
	err := provider.Shutdown(ctx)
	if err != nil {
		t.Fatalf("unexpected error during shutdown: %v", err)
	}
}

func TestMockDataDogClient(t *testing.T) {
	client := NewMockDataDogClient()

	// Test sending metrics
	err := client.SendMetric("test.counter", 1.0, "counter", []string{"env:test"}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test getting metrics
	metrics := client.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "test.counter" {
		t.Errorf("expected metric name 'test.counter', got '%s'", metrics[0].Name)
	}

	// Test reset
	client.Reset()
	metrics = client.GetMetrics()
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics after reset, got %d", len(metrics))
	}
}

// Benchmark tests
func BenchmarkProvider_CreateCounter(b *testing.B) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := metrics.CounterOptions{
			MetricOptions: metrics.MetricOptions{
				Name: "benchmark_counter",
				Help: "Benchmark counter",
			},
		}
		_, _ = provider.CreateCounter(opts)
	}
}

func BenchmarkCounter_Inc(b *testing.B) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "benchmark_counter_inc",
			Help: "Benchmark counter inc",
		},
	}

	counter, _ := provider.CreateCounter(opts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}

func BenchmarkHistogram_Observe(b *testing.B) {
	config := &Config{
		APIKey: "test-api-key",
	}
	client := NewMockDataDogClient()
	provider := NewProvider(config, client)

	opts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name: "benchmark_histogram",
			Help: "Benchmark histogram",
		},
	}

	histogram, _ := provider.CreateHistogram(opts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		histogram.Observe(float64(i % 100))
	}
}
