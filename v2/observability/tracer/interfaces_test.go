package tracer

import (
	"context"
	"testing"
	"time"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		code     StatusCode
		expected string
	}{
		{"Unset", StatusCodeUnset, "UNSET"},
		{"Ok", StatusCodeOk, "OK"},
		{"Error", StatusCodeError, "ERROR"},
		{"Unknown", StatusCode(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.String(); got != tt.expected {
				t.Errorf("StatusCode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSpanKind(t *testing.T) {
	tests := []struct {
		name     string
		kind     SpanKind
		expected string
	}{
		{"Unspecified", SpanKindUnspecified, "UNSPECIFIED"},
		{"Internal", SpanKindInternal, "INTERNAL"},
		{"Server", SpanKindServer, "SERVER"},
		{"Client", SpanKindClient, "CLIENT"},
		{"Producer", SpanKindProducer, "PRODUCER"},
		{"Consumer", SpanKindConsumer, "CONSUMER"},
		{"Unknown", SpanKind(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("SpanKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTraceFlags(t *testing.T) {
	t.Run("IsSampled", func(t *testing.T) {
		sampled := TraceFlagsSampled
		if !sampled.IsSampled() {
			t.Error("Expected sampled flag to be true")
		}
		if sampled.IsDebug() {
			t.Error("Expected debug flag to be false")
		}

		unsampled := TraceFlags(0)
		if unsampled.IsSampled() {
			t.Error("Expected unsampled flag to be false")
		}
		if unsampled.IsDebug() {
			t.Error("Expected debug flag to be false")
		}
	})

	t.Run("IsDebug", func(t *testing.T) {
		debug := TraceFlagsDebug
		if debug.IsSampled() {
			t.Error("Expected sampled flag to be false")
		}
		if !debug.IsDebug() {
			t.Error("Expected debug flag to be true")
		}

		both := TraceFlagsSampled | TraceFlagsDebug
		if !both.IsSampled() {
			t.Error("Expected sampled flag to be true")
		}
		if !both.IsDebug() {
			t.Error("Expected debug flag to be true")
		}
	})
}

func TestSpanContext(t *testing.T) {
	ctx := SpanContext{
		TraceID:  "test-trace-id",
		SpanID:   "test-span-id",
		ParentID: "test-parent-id",
		Flags:    TraceFlagsSampled,
		TraceState: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Baggage: map[string]string{
			"user.id": "12345",
		},
	}

	if ctx.TraceID != "test-trace-id" {
		t.Errorf("Expected TraceID to be 'test-trace-id', got %s", ctx.TraceID)
	}
	if ctx.SpanID != "test-span-id" {
		t.Errorf("Expected SpanID to be 'test-span-id', got %s", ctx.SpanID)
	}
	if ctx.ParentID != "test-parent-id" {
		t.Errorf("Expected ParentID to be 'test-parent-id', got %s", ctx.ParentID)
	}
	if !ctx.Flags.IsSampled() {
		t.Error("Expected flags to indicate sampled")
	}
	if ctx.TraceState["key1"] != "value1" {
		t.Errorf("Expected TraceState['key1'] to be 'value1', got %s", ctx.TraceState["key1"])
	}
	if ctx.Baggage["user.id"] != "12345" {
		t.Errorf("Expected Baggage['user.id'] to be '12345', got %s", ctx.Baggage["user.id"])
	}
}

func TestProviderType(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		expected string
	}{
		{"Datadog", ProviderTypeDatadog, "datadog"},
		{"NewRelic", ProviderTypeNewRelic, "newrelic"},
		{"Prometheus", ProviderTypePrometheus, "prometheus"},
		{"Noop", ProviderTypeNoop, "noop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.String(); got != tt.expected {
				t.Errorf("ProviderType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTracerMetrics(t *testing.T) {
	now := time.Now()
	metrics := TracerMetrics{
		SpansCreated:    100,
		SpansFinished:   95,
		SpansDropped:    5,
		AvgSpanDuration: 100 * time.Millisecond,
		LastActivity:    now,
	}

	if metrics.SpansCreated != 100 {
		t.Errorf("Expected SpansCreated to be 100, got %d", metrics.SpansCreated)
	}
	if metrics.SpansFinished != 95 {
		t.Errorf("Expected SpansFinished to be 95, got %d", metrics.SpansFinished)
	}
	if metrics.SpansDropped != 5 {
		t.Errorf("Expected SpansDropped to be 5, got %d", metrics.SpansDropped)
	}
	if metrics.AvgSpanDuration != 100*time.Millisecond {
		t.Errorf("Expected AvgSpanDuration to be 100ms, got %v", metrics.AvgSpanDuration)
	}
	if metrics.LastActivity != now {
		t.Errorf("Expected LastActivity to be %v, got %v", now, metrics.LastActivity)
	}
}

func TestProviderMetrics(t *testing.T) {
	now := time.Now()
	metrics := ProviderMetrics{
		TracersActive:   3,
		ConnectionState: "connected",
		LastFlush:       now,
		ErrorCount:      2,
		BytesSent:       1024,
	}

	if metrics.TracersActive != 3 {
		t.Errorf("Expected TracersActive to be 3, got %d", metrics.TracersActive)
	}
	if metrics.ConnectionState != "connected" {
		t.Errorf("Expected ConnectionState to be 'connected', got %s", metrics.ConnectionState)
	}
	if metrics.LastFlush != now {
		t.Errorf("Expected LastFlush to be %v, got %v", now, metrics.LastFlush)
	}
	if metrics.ErrorCount != 2 {
		t.Errorf("Expected ErrorCount to be 2, got %d", metrics.ErrorCount)
	}
	if metrics.BytesSent != 1024 {
		t.Errorf("Expected BytesSent to be 1024, got %d", metrics.BytesSent)
	}
}

func TestSpanConfig(t *testing.T) {
	now := time.Now()
	parentCtx := SpanContext{
		TraceID: "parent-trace",
		SpanID:  "parent-span",
	}

	config := SpanConfig{
		Kind:      SpanKindServer,
		StartTime: now,
		Attributes: map[string]interface{}{
			"http.method": "GET",
			"http.url":    "/api/test",
		},
		Parent: &parentCtx,
		Links: []SpanLink{
			{
				SpanContext: SpanContext{
					TraceID: "linked-trace",
					SpanID:  "linked-span",
				},
				Attributes: map[string]interface{}{
					"link.type": "child-of",
				},
			},
		},
	}

	if config.Kind != SpanKindServer {
		t.Errorf("Expected Kind to be SpanKindServer, got %v", config.Kind)
	}
	if config.StartTime != now {
		t.Errorf("Expected StartTime to be %v, got %v", now, config.StartTime)
	}
	if config.Attributes["http.method"] != "GET" {
		t.Errorf("Expected http.method to be 'GET', got %v", config.Attributes["http.method"])
	}
	if config.Attributes["http.url"] != "/api/test" {
		t.Errorf("Expected http.url to be '/api/test', got %v", config.Attributes["http.url"])
	}
	if config.Parent == nil {
		t.Error("Expected Parent to not be nil")
	} else if config.Parent.TraceID != "parent-trace" {
		t.Errorf("Expected Parent.TraceID to be 'parent-trace', got %s", config.Parent.TraceID)
	}
	if len(config.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(config.Links))
	} else if config.Links[0].SpanContext.TraceID != "linked-trace" {
		t.Errorf("Expected link TraceID to be 'linked-trace', got %s", config.Links[0].SpanContext.TraceID)
	}
}

func TestTracerConfig(t *testing.T) {
	config := TracerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "v1.2.3",
		Environment:    "testing",
		Attributes: map[string]interface{}{
			"team":    "backend",
			"version": "1.0.0",
		},
		SamplingRate:    0.1,
		EnableProfiling: true,
		BatchSize:       500,
		FlushInterval:   10 * time.Second,
	}

	if config.ServiceName != "test-service" {
		t.Errorf("Expected ServiceName to be 'test-service', got %s", config.ServiceName)
	}
	if config.ServiceVersion != "v1.2.3" {
		t.Errorf("Expected ServiceVersion to be 'v1.2.3', got %s", config.ServiceVersion)
	}
	if config.Environment != "testing" {
		t.Errorf("Expected Environment to be 'testing', got %s", config.Environment)
	}
	if config.Attributes["team"] != "backend" {
		t.Errorf("Expected team to be 'backend', got %v", config.Attributes["team"])
	}
	if config.Attributes["version"] != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got %v", config.Attributes["version"])
	}
	if config.SamplingRate != 0.1 {
		t.Errorf("Expected SamplingRate to be 0.1, got %f", config.SamplingRate)
	}
	if !config.EnableProfiling {
		t.Error("Expected EnableProfiling to be true")
	}
	if config.BatchSize != 500 {
		t.Errorf("Expected BatchSize to be 500, got %d", config.BatchSize)
	}
	if config.FlushInterval != 10*time.Second {
		t.Errorf("Expected FlushInterval to be 10s, got %v", config.FlushInterval)
	}
}

// Benchmark tests
func BenchmarkStatusCodeString(b *testing.B) {
	code := StatusCodeError
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = code.String()
	}
}

func BenchmarkSpanKindString(b *testing.B) {
	kind := SpanKindServer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = kind.String()
	}
}

func BenchmarkTraceFlagsCheck(b *testing.B) {
	flags := TraceFlagsSampled | TraceFlagsDebug
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = flags.IsSampled()
		_ = flags.IsDebug()
	}
}

// Test timeout scenarios
func TestWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	select {
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got %v", ctx.Err())
		}
	default:
		t.Fatal("Expected context to be done")
	}
}

// Race condition tests - demonstrative
func TestConcurrentAccess(t *testing.T) {
	metrics := &TracerMetrics{}

	done := make(chan bool)

	// Concurrent writers (this would demonstrate race conditions)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				metrics.SpansCreated++
				metrics.SpansFinished++
				metrics.LastActivity = time.Now()
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Note: This test demonstrates the need for proper synchronization
	// In production code, metrics should be protected by mutex or use atomic operations
	if metrics.SpansCreated == 0 {
		t.Error("Expected SpansCreated to be greater than 0")
	}
	if metrics.SpansFinished == 0 {
		t.Error("Expected SpansFinished to be greater than 0")
	}
}
