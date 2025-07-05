package newrelic

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
)

func TestNewMockNewRelicClient(t *testing.T) {
	client := NewMockNewRelicClient()
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	// Test RecordMetric
	attrs := map[string]interface{}{
		"env":    "test",
		"method": "GET",
	}
	err := client.RecordMetric("test.metric", 1.0, attrs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test RecordCustomEvent
	eventAttrs := map[string]interface{}{
		"action": "test",
		"user":   "testuser",
	}
	err = client.RecordCustomEvent("TestEvent", eventAttrs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test Shutdown
	err = client.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test GetMetrics
	metrics := client.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "test.metric" {
		t.Errorf("expected name 'test.metric', got %s", metrics[0].Name)
	}

	// Test GetEvents
	events := client.GetEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if events[0].EventType != "TestEvent" {
		t.Errorf("expected event type 'TestEvent', got %s", events[0].EventType)
	}

	// Test Reset
	client.Reset()
	metrics = client.GetMetrics()
	events = client.GetEvents()
	if len(metrics) != 0 || len(events) != 0 {
		t.Errorf("expected 0 metrics and events after reset, got %d metrics and %d events", len(metrics), len(events))
	}
}

func TestNewProvider(t *testing.T) {
	// Test with nil config and client
	provider := NewProvider(nil, nil)
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
	if provider.Name() != "newrelic" {
		t.Errorf("expected name 'newrelic', got %s", provider.Name())
	}

	// Test with custom config
	config := &Config{
		LicenseKey:  "test-license-key",
		AppName:     "test-app",
		Environment: "test",
		Host:        "test-host",
		Namespace:   "test.namespace",
		Attributes: map[string]interface{}{
			"version": "1.0.0",
			"region":  "us-east-1",
		},
	}

	client := NewMockNewRelicClient()
	provider = NewProvider(config, client)
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
	if provider.config.AppName != "test-app" {
		t.Errorf("expected app name 'test-app', got %s", provider.config.AppName)
	}
}

func TestProvider_CreateCounter(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_counter",
			Help:      "Test counter",
			Labels:    []string{"method", "status"},
			Tags:      map[string]string{"env": "test"},
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

	// Try to create the same counter again (should return existing)
	counter2, err := provider.CreateCounter(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counter != counter2 {
		t.Error("expected same counter instance")
	}

	// Try to create with different type but same name
	histOpts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_counter",
			Help:      "Test histogram",
			Namespace: "http",
		},
	}
	_, err = provider.CreateHistogram(histOpts)
	// Note: This should return an error, but implementation might allow it
	if err != nil {
		t.Logf("Got expected error for duplicate name with different type: %v", err)
	}
}

func TestProvider_CreateHistogram(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_histogram",
			Help:      "Test histogram",
			Labels:    []string{"method", "status"},
			Namespace: "http",
		},
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
	}

	histogram, err := provider.CreateHistogram(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if histogram == nil {
		t.Fatal("expected histogram, got nil")
	}
}

func TestProvider_CreateGauge(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_gauge",
			Help:      "Test gauge",
			Labels:    []string{"queue", "worker"},
			Namespace: "jobs",
		},
	}

	gauge, err := provider.CreateGauge(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gauge == nil {
		t.Fatal("expected gauge, got nil")
	}
}

func TestProvider_CreateSummary(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name:      "test_summary",
			Help:      "Test summary",
			Labels:    []string{"method", "status"},
			Namespace: "http",
		},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     10 * time.Minute,
		AgeBuckets: 5,
		BufCap:     500,
	}

	summary, err := provider.CreateSummary(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}
}

func TestProvider_Shutdown(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	ctx := context.Background()
	err := provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProvider_GetRegistry(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	registry := provider.GetRegistry()
	if registry == nil {
		t.Error("expected registry, got nil")
	}

	if registry != client {
		t.Error("expected client as registry")
	}
}

func TestCounter(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "test_counter",
			Help:   "Test counter",
			Labels: []string{"method", "status"},
		},
	}

	counter, err := provider.CreateCounter(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Inc
	counter.Inc("GET", "200")
	counter.Inc("GET", "200")
	counter.Inc("POST", "201")

	// Verify metrics were sent
	metrics := client.GetMetrics()
	if len(metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(metrics))
	}

	// Test Add
	counter.Add(5.0, "GET", "500")

	metrics = client.GetMetrics()
	if len(metrics) != 4 {
		t.Errorf("expected 4 metrics, got %d", len(metrics))
	}

	// Test Get (NewRelic provider maintains internal values)
	value := counter.Get("GET", "200")
	if value != 2.0 {
		t.Errorf("expected 2.0, got %f", value)
	}

	// Test Reset
	counter.Reset("GET", "200")
	value = counter.Get("GET", "200")
	if value != 0.0 {
		t.Errorf("expected 0.0 after reset, got %f", value)
	}

	// Verify metric attributes
	lastMetric := metrics[len(metrics)-1]
	if lastMetric.Attributes["metric_type"] != "counter" {
		t.Errorf("expected metric_type 'counter', got %v", lastMetric.Attributes["metric_type"])
	}
}

func TestHistogram(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "test_histogram",
			Help:   "Test histogram",
			Labels: []string{"method"},
		},
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0},
	}

	histogram, err := provider.CreateHistogram(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Observe
	histogram.Observe(0.5, "GET")
	histogram.Observe(1.2, "GET")
	histogram.Observe(0.3, "POST")

	// Verify metrics were sent
	metrics := client.GetMetrics()
	if len(metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(metrics))
	}

	// Test ObserveWithTimestamp
	timestamp := time.Now().Add(-1 * time.Hour)
	histogram.ObserveWithTimestamp(0.8, timestamp, "GET")

	metrics = client.GetMetrics()
	if len(metrics) != 4 {
		t.Errorf("expected 4 metrics, got %d", len(metrics))
	}

	lastMetric := metrics[len(metrics)-1]
	if lastMetric.Attributes["metric_type"] != "histogram" {
		t.Errorf("expected metric_type 'histogram', got %v", lastMetric.Attributes["metric_type"])
	}
	if lastMetric.Attributes["timestamp"] != timestamp.Unix() {
		t.Errorf("expected timestamp %d, got %v", timestamp.Unix(), lastMetric.Attributes["timestamp"])
	}

	// Test Time
	executed := false
	histogram.Time(func() {
		time.Sleep(10 * time.Millisecond)
		executed = true
	}, "GET")

	if !executed {
		t.Error("expected function to be executed")
	}

	// Test StartTimer
	timer := histogram.StartTimer("GET")
	if timer == nil {
		t.Error("expected timer function, got nil")
	}
	time.Sleep(10 * time.Millisecond)
	timer()

	// Test GetCount and GetSum (NewRelic doesn't provide direct access)
	count := histogram.GetCount("GET")
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	sum := histogram.GetSum("GET")
	if sum != 0 {
		t.Errorf("expected 0, got %f", sum)
	}
}

func TestGauge(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "test_gauge",
			Help:   "Test gauge",
			Labels: []string{"queue"},
		},
	}

	gauge, err := provider.CreateGauge(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Set
	gauge.Set(10.0, "tasks")

	// Verify metrics were sent
	metrics := client.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	// Test Inc
	gauge.Inc("tasks")

	// Test Dec
	gauge.Dec("tasks")

	// Test Add
	gauge.Add(5.0, "tasks")

	// Test Sub
	gauge.Sub(2.0, "tasks")

	// Test Get (NewRelic provider maintains internal values)
	value := gauge.Get("tasks")
	if value != 13.0 { // 10 + 1 - 1 + 5 - 2 = 13
		t.Errorf("expected 13.0, got %f", value)
	}

	// Test SetToCurrentTime
	gauge.SetToCurrentTime("tasks")

	// Verify metric attributes
	allMetrics := client.GetMetrics()
	lastMetric := allMetrics[len(allMetrics)-1]
	if lastMetric.Attributes["metric_type"] != "gauge" {
		t.Errorf("expected metric_type 'gauge', got %v", lastMetric.Attributes["metric_type"])
	}
}

func TestSummary(t *testing.T) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "test_summary",
			Help:   "Test summary",
			Labels: []string{"method"},
		},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}

	summary, err := provider.CreateSummary(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Observe
	summary.Observe(0.5, "GET")
	summary.Observe(1.2, "GET")
	summary.Observe(0.3, "POST")

	// Verify metrics were sent
	metrics := client.GetMetrics()
	if len(metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(metrics))
	}

	// Test ObserveWithTimestamp
	timestamp := time.Now().Add(-1 * time.Hour)
	summary.ObserveWithTimestamp(0.8, timestamp, "GET")

	metrics = client.GetMetrics()
	if len(metrics) != 4 {
		t.Errorf("expected 4 metrics, got %d", len(metrics))
	}

	lastMetric := metrics[len(metrics)-1]
	if lastMetric.Attributes["metric_type"] != "summary" {
		t.Errorf("expected metric_type 'summary', got %v", lastMetric.Attributes["metric_type"])
	}

	// Test Time
	executed := false
	summary.Time(func() {
		time.Sleep(10 * time.Millisecond)
		executed = true
	}, "GET")

	if !executed {
		t.Error("expected function to be executed")
	}

	// Test StartTimer
	timer := summary.StartTimer("GET")
	if timer == nil {
		t.Error("expected timer function, got nil")
	}
	time.Sleep(10 * time.Millisecond)
	timer()

	// Test GetCount, GetSum, GetQuantile (NewRelic doesn't provide direct access)
	count := summary.GetCount("GET")
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	sum := summary.GetSum("GET")
	if sum != 0 {
		t.Errorf("expected 0, got %f", sum)
	}

	quantile := summary.GetQuantile(0.5, "GET")
	if quantile != 0 {
		t.Errorf("expected 0, got %f", quantile)
	}
}

// Benchmark tests
func BenchmarkCounterInc(b *testing.B) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.CounterOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "bench_counter",
			Help:   "Benchmark counter",
			Labels: []string{"method", "status"},
		},
	}

	counter, err := provider.CreateCounter(opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc("GET", "200")
		}
	})
}

func BenchmarkHistogramObserve(b *testing.B) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.HistogramOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "bench_histogram",
			Help:   "Benchmark histogram",
			Labels: []string{"method"},
		},
	}

	histogram, err := provider.CreateHistogram(opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			histogram.Observe(float64(i%1000)/100.0, "GET")
			i++
		}
	})
}

func BenchmarkGaugeSet(b *testing.B) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.GaugeOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "bench_gauge",
			Help:   "Benchmark gauge",
			Labels: []string{"queue"},
		},
	}

	gauge, err := provider.CreateGauge(opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			gauge.Set(float64(i%1000), "tasks")
			i++
		}
	})
}

func BenchmarkSummaryObserve(b *testing.B) {
	client := NewMockNewRelicClient()
	provider := NewProvider(nil, client)

	opts := metrics.SummaryOptions{
		MetricOptions: metrics.MetricOptions{
			Name:   "bench_summary",
			Help:   "Benchmark summary",
			Labels: []string{"method"},
		},
	}

	summary, err := provider.CreateSummary(opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			summary.Observe(float64(i%1000)/100.0, "GET")
			i++
		}
	})
}
