package metrics

import (
	"testing"
)

func TestMetricsError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := NewMetricsError(ErrTypeInvalidConfig, "test error")
		if err.Type != ErrTypeInvalidConfig {
			t.Errorf("expected type %v, got %v", ErrTypeInvalidConfig, err.Type)
		}
		if err.Message != "test error" {
			t.Errorf("expected message 'test error', got '%s'", err.Message)
		}
	})

	t.Run("error with provider", func(t *testing.T) {
		err := NewMetricsError(
			ErrTypeMetricCreation,
			"creation failed",
			WithProvider("prometheus"),
		)
		if err.Provider != "prometheus" {
			t.Errorf("expected provider 'prometheus', got '%s'", err.Provider)
		}
		if err.Error() != "prometheus: creation failed" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("error with metric", func(t *testing.T) {
		err := NewMetricsError(
			ErrTypeMetricOperation,
			"operation failed",
			WithMetric("test_counter"),
		)
		if err.Metric != "test_counter" {
			t.Errorf("expected metric 'test_counter', got '%s'", err.Metric)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := NewMetricsError(ErrTypeInvalidConfig, "root cause")
		err := NewMetricsError(
			ErrTypeMetricCreation,
			"wrapped error",
			WithCause(cause),
		)
		if err.Unwrap() != cause {
			t.Errorf("expected cause to be preserved")
		}
	})
}

func TestErrorTypes(t *testing.T) {
	errorTypes := []ErrorType{
		ErrTypeInvalidConfig,
		ErrTypeProviderNotFound,
		ErrTypeMetricCreation,
		ErrTypeMetricOperation,
		ErrTypeProviderShutdown,
	}

	for _, errType := range errorTypes {
		t.Run(string(errType), func(t *testing.T) {
			err := NewMetricsError(errType, "test message")
			if err.Type != errType {
				t.Errorf("expected type %v, got %v", errType, err.Type)
			}
		})
	}
}

func TestCounterOptions(t *testing.T) {
	opts := CounterOptions{
		MetricOptions: MetricOptions{
			Name:      "test_counter",
			Help:      "A test counter",
			Labels:    []string{"method", "status"},
			Namespace: "app",
			Subsystem: "http",
		},
	}

	if opts.Name != "test_counter" {
		t.Errorf("expected name 'test_counter', got '%s'", opts.Name)
	}
	if opts.Help != "A test counter" {
		t.Errorf("expected help 'A test counter', got '%s'", opts.Help)
	}
	if len(opts.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(opts.Labels))
	}
}

func TestHistogramOptions(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	opts := HistogramOptions{
		MetricOptions: MetricOptions{
			Name: "test_histogram",
			Help: "A test histogram",
		},
		Buckets: buckets,
	}

	if len(opts.Buckets) != 5 {
		t.Errorf("expected 5 buckets, got %d", len(opts.Buckets))
	}
	for i, bucket := range buckets {
		if opts.Buckets[i] != bucket {
			t.Errorf("expected bucket %f at index %d, got %f", bucket, i, opts.Buckets[i])
		}
	}
}

func TestGaugeOptions(t *testing.T) {
	opts := GaugeOptions{
		MetricOptions: MetricOptions{
			Name: "test_gauge",
			Help: "A test gauge",
		},
	}

	if opts.Name != "test_gauge" {
		t.Errorf("expected name 'test_gauge', got '%s'", opts.Name)
	}
}

func TestSummaryOptions(t *testing.T) {
	objectives := map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}

	opts := SummaryOptions{
		MetricOptions: MetricOptions{
			Name: "test_summary",
			Help: "A test summary",
		},
		Objectives: objectives,
	}

	if len(opts.Objectives) != 3 {
		t.Errorf("expected 3 objectives, got %d", len(opts.Objectives))
	}
	for quantile, tolerance := range objectives {
		if opts.Objectives[quantile] != tolerance {
			t.Errorf("expected tolerance %f for quantile %f, got %f", tolerance, quantile, opts.Objectives[quantile])
		}
	}
}

// Benchmark tests
func BenchmarkMetricsErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMetricsError(ErrTypeInvalidConfig, "benchmark error")
	}
}

func BenchmarkMetricsErrorWithOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMetricsError(
			ErrTypeMetricCreation,
			"benchmark error",
			WithProvider("test"),
			WithMetric("test_metric"),
		)
	}
}
