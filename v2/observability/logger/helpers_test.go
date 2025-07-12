package logger

import (
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// TestFieldHelpers testa as funções helper para criar fields
func TestFieldHelpers(t *testing.T) {
	// Testa String
	field := String("key", "value")
	if field.Key != "key" || field.Value != "value" {
		t.Errorf("String field incorrect: %+v", field)
	}

	// Testa Int
	field = Int("count", 42)
	if field.Key != "count" || field.Value != 42 {
		t.Errorf("Int field incorrect: %+v", field)
	}

	// Testa Int64
	field = Int64("large", int64(9223372036854775807))
	if field.Key != "large" || field.Value != int64(9223372036854775807) {
		t.Errorf("Int64 field incorrect: %+v", field)
	}

	// Testa Float64
	field = Float64("price", 19.99)
	if field.Key != "price" || field.Value != 19.99 {
		t.Errorf("Float64 field incorrect: %+v", field)
	}

	// Testa Bool
	field = Bool("active", true)
	if field.Key != "active" || field.Value != true {
		t.Errorf("Bool field incorrect: %+v", field)
	}

	// Testa Time
	now := time.Now()
	field = Time("timestamp", now)
	if field.Key != "timestamp" || field.Value != now {
		t.Errorf("Time field incorrect: %+v", field)
	}

	// Testa Duration
	duration := 5 * time.Second
	field = Duration("timeout", duration)
	if field.Key != "timeout" || field.Value != duration {
		t.Errorf("Duration field incorrect: %+v", field)
	}

	// Testa Err
	testErr := errors.New("test error")
	field = Err(testErr)
	if field.Key != "error" || field.Value != "test error" {
		t.Errorf("Err field incorrect: key=%s, value=%v", field.Key, field.Value)
	}

	// Testa ErrorNamed
	field = ErrorNamed("custom_error", testErr)
	if field.Key != "custom_error" || field.Value != "test error" {
		t.Errorf("ErrorNamed field incorrect: key=%s, value=%v", field.Key, field.Value)
	}

	// Testa Object
	obj := map[string]interface{}{"nested": "value"}
	field = Object("data", obj)
	if field.Key != "data" {
		t.Errorf("Object field key incorrect: expected 'data', got '%s'", field.Key)
	}

	// Testa Array
	arr := []interface{}{1, 2, 3}
	field = Array("items", arr)
	if field.Key != "items" {
		t.Errorf("Array field key incorrect: expected 'items', got '%s'", field.Key)
	}
}

// TestTraceFields testa as funções helper para tracing
func TestTraceFields(t *testing.T) {
	// Testa TraceID
	field := TraceID("trace-123")
	if field.Key != "trace_id" || field.Value != "trace-123" {
		t.Errorf("TraceID field incorrect: %+v", field)
	}

	// Testa SpanID
	field = SpanID("span-456")
	if field.Key != "span_id" || field.Value != "span-456" {
		t.Errorf("SpanID field incorrect: %+v", field)
	}

	// Testa UserID
	field = UserID("user-789")
	if field.Key != "user_id" || field.Value != "user-789" {
		t.Errorf("UserID field incorrect: %+v", field)
	}

	// Testa RequestID
	field = RequestID("req-abc")
	if field.Key != "request_id" || field.Value != "req-abc" {
		t.Errorf("RequestID field incorrect: %+v", field)
	}

	// Testa CorrelationID
	field = CorrelationID("corr-def")
	if field.Key != "correlation_id" || field.Value != "corr-def" {
		t.Errorf("CorrelationID field incorrect: %+v", field)
	}
}

// TestHTTPFields testa as funções helper para HTTP
func TestHTTPFields(t *testing.T) {
	// Testa Method
	field := Method("POST")
	if field.Key != "method" || field.Value != "POST" {
		t.Errorf("Method field incorrect: %+v", field)
	}

	// Testa Path
	field = Path("/api/users")
	if field.Key != "path" || field.Value != "/api/users" {
		t.Errorf("Path field incorrect: %+v", field)
	}

	// Testa StatusCode
	field = StatusCode(200)
	if field.Key != "status_code" || field.Value != 200 {
		t.Errorf("StatusCode field incorrect: %+v", field)
	}

	// Testa Latency
	latency := 150 * time.Millisecond
	field = Latency(latency)
	if field.Key != "latency" || field.Value != latency {
		t.Errorf("Latency field incorrect: %+v", field)
	}
}

// TestApplicationFields testa as funções helper para aplicação
func TestApplicationFields(t *testing.T) {
	// Testa Operation
	field := Operation("user.create")
	if field.Key != "operation" || field.Value != "user.create" {
		t.Errorf("Operation field incorrect: %+v", field)
	}

	// Testa Component
	field = Component("auth-service")
	if field.Key != "component" || field.Value != "auth-service" {
		t.Errorf("Component field incorrect: %+v", field)
	}

	// Testa Version
	field = Version("v1.2.3")
	if field.Key != "version" || field.Value != "v1.2.3" {
		t.Errorf("Version field incorrect: %+v", field)
	}

	// Testa Environment
	field = Environment("production")
	if field.Key != "environment" || field.Value != "production" {
		t.Errorf("Environment field incorrect: %+v", field)
	}
}

// TestConfigBuilder testa o builder de configuração
func TestConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()

	config := builder.
		Level(interfaces.InfoLevel).
		Format(interfaces.JSONFormat).
		ServiceName("test-service").
		ServiceVersion("1.0.0").
		Environment("test").
		AddSource(true).
		AddStacktrace(true).
		AddCaller(true).
		WithGlobalField("service", "test").
		WithAsync(1000, 2, 5*time.Second).
		WithSampling(1, 100, 1*time.Second).
		EnableMetrics("test_").
		Build()

	if config.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if config.Format != interfaces.JSONFormat {
		t.Errorf("Expected JSONFormat, got %v", config.Format)
	}

	if config.ServiceName != "test-service" {
		t.Errorf("Expected test-service, got %s", config.ServiceName)
	}

	if config.ServiceVersion != "1.0.0" {
		t.Errorf("Expected 1.0.0, got %s", config.ServiceVersion)
	}

	if config.Environment != "test" {
		t.Errorf("Expected test, got %s", config.Environment)
	}

	if !config.AddSource {
		t.Error("Expected AddSource to be true")
	}

	if !config.AddStacktrace {
		t.Error("Expected AddStacktrace to be true")
	}

	if !config.AddCaller {
		t.Error("Expected AddCaller to be true")
	}

	if config.GlobalFields["service"] != "test" {
		t.Errorf("Expected global field service=test, got %v", config.GlobalFields["service"])
	}

	if config.Async == nil {
		t.Fatal("Expected Async config to be set")
	}

	if config.Async.BufferSize != 1000 {
		t.Errorf("Expected buffer size 1000, got %d", config.Async.BufferSize)
	}

	if config.Async.Workers != 2 {
		t.Errorf("Expected 2 workers, got %d", config.Async.Workers)
	}

	if config.Async.FlushInterval != 5*time.Second {
		t.Errorf("Expected 5s flush interval, got %v", config.Async.FlushInterval)
	}

	if config.Sampling == nil {
		t.Fatal("Expected Sampling config to be set")
	}

	if config.Sampling.Initial != 1 {
		t.Errorf("Expected initial sampling 1, got %d", config.Sampling.Initial)
	}

	if config.Sampling.Thereafter != 100 {
		t.Errorf("Expected thereafter sampling 100, got %d", config.Sampling.Thereafter)
	}

	if config.MetricsPrefix != "test_" {
		t.Errorf("Expected metrics prefix test_, got %s", config.MetricsPrefix)
	}
}

// TestEnvironmentConfig testa a configuração por environment
func TestEnvironmentConfig(t *testing.T) {
	// Mock environment variables para o teste
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("SERVICE_NAME", "env-service")

	config := EnvironmentConfig()

	// Verifica se a função executa sem erros
	if config.ServiceName == "" {
		t.Error("Expected service name to be set from environment or defaults")
	}

	if config.Level == 0 {
		t.Error("Expected log level to be set from environment or defaults")
	}
}
