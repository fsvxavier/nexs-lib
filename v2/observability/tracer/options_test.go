package tracer

import (
	"testing"
	"time"
)

func TestSpanOptions(t *testing.T) {
	t.Run("WithSpanKind", func(t *testing.T) {
		config := defaultSpanConfig()
		opt := WithSpanKind(SpanKindServer)
		opt.apply(config)

		if config.Kind != SpanKindServer {
			t.Errorf("Expected Kind to be SpanKindServer, got %v", config.Kind)
		}
	})

	t.Run("WithStartTime", func(t *testing.T) {
		config := defaultSpanConfig()
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		opt := WithStartTime(testTime)
		opt.apply(config)

		if !config.StartTime.Equal(testTime) {
			t.Errorf("Expected StartTime to be %v, got %v", testTime, config.StartTime)
		}
	})

	t.Run("WithSpanAttributes", func(t *testing.T) {
		config := defaultSpanConfig()
		attrs := map[string]interface{}{
			"http.method": "GET",
			"http.url":    "/api/test",
			"user.id":     12345,
		}
		opt := WithSpanAttributes(attrs)
		opt.apply(config)

		if config.Attributes["http.method"] != "GET" {
			t.Errorf("Expected http.method to be 'GET', got %v", config.Attributes["http.method"])
		}
		if config.Attributes["http.url"] != "/api/test" {
			t.Errorf("Expected http.url to be '/api/test', got %v", config.Attributes["http.url"])
		}
		if config.Attributes["user.id"] != 12345 {
			t.Errorf("Expected user.id to be 12345, got %v", config.Attributes["user.id"])
		}
	})

	t.Run("WithSpanAttributesAppend", func(t *testing.T) {
		config := defaultSpanConfig()
		config.Attributes["existing"] = "value"

		attrs := map[string]interface{}{
			"new.key": "new.value",
		}
		opt := WithSpanAttributes(attrs)
		opt.apply(config)

		if config.Attributes["existing"] != "value" {
			t.Error("Expected existing attribute to be preserved")
		}
		if config.Attributes["new.key"] != "new.value" {
			t.Error("Expected new attribute to be added")
		}
	})

	t.Run("WithParentSpan", func(t *testing.T) {
		config := defaultSpanConfig()
		parent := SpanContext{
			TraceID: "parent-trace-id",
			SpanID:  "parent-span-id",
		}
		opt := WithParentSpan(parent)
		opt.apply(config)

		if config.Parent == nil {
			t.Fatal("Expected Parent to not be nil")
		}
		if config.Parent.TraceID != "parent-trace-id" {
			t.Errorf("Expected Parent.TraceID to be 'parent-trace-id', got %s", config.Parent.TraceID)
		}
		if config.Parent.SpanID != "parent-span-id" {
			t.Errorf("Expected Parent.SpanID to be 'parent-span-id', got %s", config.Parent.SpanID)
		}
	})

	t.Run("WithSpanLinks", func(t *testing.T) {
		config := defaultSpanConfig()
		links := []SpanLink{
			{
				SpanContext: SpanContext{
					TraceID: "linked-trace-1",
					SpanID:  "linked-span-1",
				},
				Attributes: map[string]interface{}{
					"link.type": "child-of",
				},
			},
			{
				SpanContext: SpanContext{
					TraceID: "linked-trace-2",
					SpanID:  "linked-span-2",
				},
				Attributes: map[string]interface{}{
					"link.type": "follows-from",
				},
			},
		}
		opt := WithSpanLinks(links...)
		opt.apply(config)

		if len(config.Links) != 2 {
			t.Errorf("Expected 2 links, got %d", len(config.Links))
		}
		if config.Links[0].SpanContext.TraceID != "linked-trace-1" {
			t.Errorf("Expected first link TraceID to be 'linked-trace-1', got %s", config.Links[0].SpanContext.TraceID)
		}
		if config.Links[1].SpanContext.TraceID != "linked-trace-2" {
			t.Errorf("Expected second link TraceID to be 'linked-trace-2', got %s", config.Links[1].SpanContext.TraceID)
		}
	})
}

func TestTracerOptions(t *testing.T) {
	t.Run("WithServiceName", func(t *testing.T) {
		config := defaultTracerConfig()
		opt := WithServiceName("my-service")
		opt.apply(config)

		if config.ServiceName != "my-service" {
			t.Errorf("Expected ServiceName to be 'my-service', got %s", config.ServiceName)
		}
	})

	t.Run("WithServiceVersion", func(t *testing.T) {
		config := defaultTracerConfig()
		opt := WithServiceVersion("v2.0.0")
		opt.apply(config)

		if config.ServiceVersion != "v2.0.0" {
			t.Errorf("Expected ServiceVersion to be 'v2.0.0', got %s", config.ServiceVersion)
		}
	})

	t.Run("WithEnvironment", func(t *testing.T) {
		config := defaultTracerConfig()
		opt := WithEnvironment("production")
		opt.apply(config)

		if config.Environment != "production" {
			t.Errorf("Expected Environment to be 'production', got %s", config.Environment)
		}
	})

	t.Run("WithTracerAttributes", func(t *testing.T) {
		config := defaultTracerConfig()
		attrs := map[string]interface{}{
			"team":       "backend",
			"region":     "us-east-1",
			"datacenter": "aws",
		}
		opt := WithTracerAttributes(attrs)
		opt.apply(config)

		if config.Attributes["team"] != "backend" {
			t.Errorf("Expected team to be 'backend', got %v", config.Attributes["team"])
		}
		if config.Attributes["region"] != "us-east-1" {
			t.Errorf("Expected region to be 'us-east-1', got %v", config.Attributes["region"])
		}
		if config.Attributes["datacenter"] != "aws" {
			t.Errorf("Expected datacenter to be 'aws', got %v", config.Attributes["datacenter"])
		}
	})

	t.Run("WithTracerAttributesAppend", func(t *testing.T) {
		config := defaultTracerConfig()
		config.Attributes["existing"] = "value"

		attrs := map[string]interface{}{
			"new.key": "new.value",
		}
		opt := WithTracerAttributes(attrs)
		opt.apply(config)

		if config.Attributes["existing"] != "value" {
			t.Error("Expected existing attribute to be preserved")
		}
		if config.Attributes["new.key"] != "new.value" {
			t.Error("Expected new attribute to be added")
		}
	})

	t.Run("WithSamplingRate", func(t *testing.T) {
		config := defaultTracerConfig()

		// Test valid sampling rate
		opt := WithSamplingRate(0.5)
		opt.apply(config)
		if config.SamplingRate != 0.5 {
			t.Errorf("Expected SamplingRate to be 0.5, got %f", config.SamplingRate)
		}

		// Test invalid sampling rate (too low)
		opt = WithSamplingRate(-0.1)
		opt.apply(config)
		if config.SamplingRate != 0.5 { // Should remain unchanged
			t.Errorf("Expected SamplingRate to remain 0.5, got %f", config.SamplingRate)
		}

		// Test invalid sampling rate (too high)
		opt = WithSamplingRate(1.5)
		opt.apply(config)
		if config.SamplingRate != 0.5 { // Should remain unchanged
			t.Errorf("Expected SamplingRate to remain 0.5, got %f", config.SamplingRate)
		}

		// Test edge values
		opt = WithSamplingRate(0.0)
		opt.apply(config)
		if config.SamplingRate != 0.0 {
			t.Errorf("Expected SamplingRate to be 0.0, got %f", config.SamplingRate)
		}

		opt = WithSamplingRate(1.0)
		opt.apply(config)
		if config.SamplingRate != 1.0 {
			t.Errorf("Expected SamplingRate to be 1.0, got %f", config.SamplingRate)
		}
	})

	t.Run("WithProfiling", func(t *testing.T) {
		config := defaultTracerConfig()

		// Test enabling profiling
		opt := WithProfiling(true)
		opt.apply(config)
		if !config.EnableProfiling {
			t.Error("Expected EnableProfiling to be true")
		}

		// Test disabling profiling
		opt = WithProfiling(false)
		opt.apply(config)
		if config.EnableProfiling {
			t.Error("Expected EnableProfiling to be false")
		}
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		config := defaultTracerConfig()

		// Test valid batch size
		opt := WithBatchSize(200)
		opt.apply(config)
		if config.BatchSize != 200 {
			t.Errorf("Expected BatchSize to be 200, got %d", config.BatchSize)
		}

		// Test invalid batch size (zero)
		opt = WithBatchSize(0)
		opt.apply(config)
		if config.BatchSize != 200 { // Should remain unchanged
			t.Errorf("Expected BatchSize to remain 200, got %d", config.BatchSize)
		}

		// Test invalid batch size (negative)
		opt = WithBatchSize(-10)
		opt.apply(config)
		if config.BatchSize != 200 { // Should remain unchanged
			t.Errorf("Expected BatchSize to remain 200, got %d", config.BatchSize)
		}
	})

	t.Run("WithFlushInterval", func(t *testing.T) {
		config := defaultTracerConfig()

		// Test valid flush interval
		opt := WithFlushInterval(10 * time.Second)
		opt.apply(config)
		if config.FlushInterval != 10*time.Second {
			t.Errorf("Expected FlushInterval to be 10s, got %v", config.FlushInterval)
		}

		// Test invalid flush interval (zero)
		opt = WithFlushInterval(0)
		opt.apply(config)
		if config.FlushInterval != 10*time.Second { // Should remain unchanged
			t.Errorf("Expected FlushInterval to remain 10s, got %v", config.FlushInterval)
		}

		// Test invalid flush interval (negative)
		opt = WithFlushInterval(-5 * time.Second)
		opt.apply(config)
		if config.FlushInterval != 10*time.Second { // Should remain unchanged
			t.Errorf("Expected FlushInterval to remain 10s, got %v", config.FlushInterval)
		}
	})
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("defaultSpanConfig", func(t *testing.T) {
		config := defaultSpanConfig()

		if config.Kind != SpanKindInternal {
			t.Errorf("Expected default Kind to be SpanKindInternal, got %v", config.Kind)
		}
		if config.StartTime.IsZero() {
			t.Error("Expected default StartTime to be set")
		}
		if config.Attributes == nil {
			t.Error("Expected default Attributes to be initialized")
		}
		if config.Links == nil {
			t.Error("Expected default Links to be initialized")
		}
		if len(config.Links) != 0 {
			t.Errorf("Expected default Links to be empty, got %d", len(config.Links))
		}
	})

	t.Run("defaultTracerConfig", func(t *testing.T) {
		config := defaultTracerConfig()

		if config.ServiceName != "unknown-service" {
			t.Errorf("Expected default ServiceName to be 'unknown-service', got %s", config.ServiceName)
		}
		if config.ServiceVersion != "1.0.0" {
			t.Errorf("Expected default ServiceVersion to be '1.0.0', got %s", config.ServiceVersion)
		}
		if config.Environment != "development" {
			t.Errorf("Expected default Environment to be 'development', got %s", config.Environment)
		}
		if config.Attributes == nil {
			t.Error("Expected default Attributes to be initialized")
		}
		if config.SamplingRate != 1.0 {
			t.Errorf("Expected default SamplingRate to be 1.0, got %f", config.SamplingRate)
		}
		if config.EnableProfiling {
			t.Error("Expected default EnableProfiling to be false")
		}
		if config.BatchSize != 100 {
			t.Errorf("Expected default BatchSize to be 100, got %d", config.BatchSize)
		}
		if config.FlushInterval != 5*time.Second {
			t.Errorf("Expected default FlushInterval to be 5s, got %v", config.FlushInterval)
		}
	})
}

func TestApplyOptions(t *testing.T) {
	t.Run("ApplySpanOptions", func(t *testing.T) {
		opts := []SpanOption{
			WithSpanKind(SpanKindClient),
			WithSpanAttributes(map[string]interface{}{
				"test.key": "test.value",
			}),
		}

		config := ApplySpanOptions(opts)

		if config.Kind != SpanKindClient {
			t.Errorf("Expected Kind to be SpanKindClient, got %v", config.Kind)
		}
		if config.Attributes["test.key"] != "test.value" {
			t.Error("Expected test.key attribute to be set")
		}
	})

	t.Run("ApplySpanOptionsEmpty", func(t *testing.T) {
		config := ApplySpanOptions(nil)

		// Should return default config
		if config.Kind != SpanKindInternal {
			t.Errorf("Expected default Kind to be SpanKindInternal, got %v", config.Kind)
		}
	})

	t.Run("applySpanOptions", func(t *testing.T) {
		opts := []SpanOption{
			WithSpanKind(SpanKindServer),
		}

		config := applySpanOptions(opts)

		if config.Kind != SpanKindServer {
			t.Errorf("Expected Kind to be SpanKindServer, got %v", config.Kind)
		}
	})

	t.Run("ApplyTracerOptions", func(t *testing.T) {
		opts := []TracerOption{
			WithServiceName("test-service"),
			WithEnvironment("testing"),
			WithSamplingRate(0.1),
		}

		config := ApplyTracerOptions(opts)

		if config.ServiceName != "test-service" {
			t.Errorf("Expected ServiceName to be 'test-service', got %s", config.ServiceName)
		}
		if config.Environment != "testing" {
			t.Errorf("Expected Environment to be 'testing', got %s", config.Environment)
		}
		if config.SamplingRate != 0.1 {
			t.Errorf("Expected SamplingRate to be 0.1, got %f", config.SamplingRate)
		}
	})

	t.Run("ApplyTracerOptionsEmpty", func(t *testing.T) {
		config := ApplyTracerOptions(nil)

		// Should return default config
		if config.ServiceName != "unknown-service" {
			t.Errorf("Expected default ServiceName to be 'unknown-service', got %s", config.ServiceName)
		}
	})

	t.Run("applyTracerOptions", func(t *testing.T) {
		opts := []TracerOption{
			WithServiceName("internal-service"),
		}

		config := applyTracerOptions(opts)

		if config.ServiceName != "internal-service" {
			t.Errorf("Expected ServiceName to be 'internal-service', got %s", config.ServiceName)
		}
	})
}

func TestOptionsChaining(t *testing.T) {
	t.Run("ChainedSpanOptions", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		parent := SpanContext{
			TraceID: "parent-trace",
			SpanID:  "parent-span",
		}

		opts := []SpanOption{
			WithSpanKind(SpanKindServer),
			WithStartTime(testTime),
			WithSpanAttributes(map[string]interface{}{
				"http.method": "POST",
			}),
			WithSpanAttributes(map[string]interface{}{
				"http.url": "/api/users",
			}),
			WithParentSpan(parent),
		}

		config := ApplySpanOptions(opts)

		if config.Kind != SpanKindServer {
			t.Error("Expected span kind to be server")
		}
		if !config.StartTime.Equal(testTime) {
			t.Error("Expected start time to be set")
		}
		if config.Attributes["http.method"] != "POST" {
			t.Error("Expected http.method to be POST")
		}
		if config.Attributes["http.url"] != "/api/users" {
			t.Error("Expected http.url to be /api/users")
		}
		if config.Parent == nil || config.Parent.TraceID != "parent-trace" {
			t.Error("Expected parent span to be set")
		}
	})

	t.Run("ChainedTracerOptions", func(t *testing.T) {
		opts := []TracerOption{
			WithServiceName("chained-service"),
			WithServiceVersion("v2.1.0"),
			WithEnvironment("staging"),
			WithTracerAttributes(map[string]interface{}{
				"team": "platform",
			}),
			WithTracerAttributes(map[string]interface{}{
				"region": "us-west-2",
			}),
			WithSamplingRate(0.25),
			WithProfiling(true),
			WithBatchSize(250),
			WithFlushInterval(15 * time.Second),
		}

		config := ApplyTracerOptions(opts)

		if config.ServiceName != "chained-service" {
			t.Error("Expected service name to be chained-service")
		}
		if config.ServiceVersion != "v2.1.0" {
			t.Error("Expected service version to be v2.1.0")
		}
		if config.Environment != "staging" {
			t.Error("Expected environment to be staging")
		}
		if config.Attributes["team"] != "platform" {
			t.Error("Expected team attribute to be platform")
		}
		if config.Attributes["region"] != "us-west-2" {
			t.Error("Expected region attribute to be us-west-2")
		}
		if config.SamplingRate != 0.25 {
			t.Error("Expected sampling rate to be 0.25")
		}
		if !config.EnableProfiling {
			t.Error("Expected profiling to be enabled")
		}
		if config.BatchSize != 250 {
			t.Error("Expected batch size to be 250")
		}
		if config.FlushInterval != 15*time.Second {
			t.Error("Expected flush interval to be 15s")
		}
	})
}

// Benchmark tests
func BenchmarkApplySpanOptions(b *testing.B) {
	opts := []SpanOption{
		WithSpanKind(SpanKindServer),
		WithSpanAttributes(map[string]interface{}{
			"http.method": "GET",
			"http.url":    "/api/test",
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplySpanOptions(opts)
	}
}

func BenchmarkApplyTracerOptions(b *testing.B) {
	opts := []TracerOption{
		WithServiceName("bench-service"),
		WithEnvironment("production"),
		WithSamplingRate(0.1),
		WithBatchSize(100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyTracerOptions(opts)
	}
}

func BenchmarkOptionApplication(b *testing.B) {
	b.Run("SingleSpanOption", func(b *testing.B) {
		opt := WithSpanKind(SpanKindServer)
		config := defaultSpanConfig()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opt.apply(config)
		}
	})

	b.Run("SingleTracerOption", func(b *testing.B) {
		opt := WithServiceName("test-service")
		config := defaultTracerConfig()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opt.apply(config)
		}
	})
}
