package interfaces

import (
	"errors"
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{TraceLevel, "TRACE"},
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{PanicLevel, "PANIC"},
		{Level(99), "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if result := test.level.String(); result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{JSONFormat, "json"},
		{TextFormat, "text"},
		{ConsoleFormat, "console"},
		{Format(99), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if result := test.format.String(); result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestFieldCreation(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		field := String("key", "value")
		if field.Key != "key" || field.Value != "value" || field.Type != StringType {
			t.Errorf("String field creation failed: %+v", field)
		}
	})

	t.Run("Int field", func(t *testing.T) {
		field := Int("count", 42)
		if field.Key != "count" || field.Value != 42 || field.Type != IntType {
			t.Errorf("Int field creation failed: %+v", field)
		}
	})

	t.Run("Int64 field", func(t *testing.T) {
		field := Int64("id", 123456789)
		if field.Key != "id" || field.Value != int64(123456789) || field.Type != Int64Type {
			t.Errorf("Int64 field creation failed: %+v", field)
		}
	})

	t.Run("Float64 field", func(t *testing.T) {
		field := Float64("score", 95.5)
		if field.Key != "score" || field.Value != 95.5 || field.Type != Float64Type {
			t.Errorf("Float64 field creation failed: %+v", field)
		}
	})

	t.Run("Bool field", func(t *testing.T) {
		field := Bool("active", true)
		if field.Key != "active" || field.Value != true || field.Type != BoolType {
			t.Errorf("Bool field creation failed: %+v", field)
		}
	})

	t.Run("Time field", func(t *testing.T) {
		now := time.Now()
		field := Time("timestamp", now)
		if field.Key != "timestamp" || field.Value != now || field.Type != TimeType {
			t.Errorf("Time field creation failed: %+v", field)
		}
	})

	t.Run("Duration field", func(t *testing.T) {
		duration := 5 * time.Second
		field := Duration("latency", duration)
		if field.Key != "latency" || field.Value != duration || field.Type != DurationType {
			t.Errorf("Duration field creation failed: %+v", field)
		}
	})

	t.Run("Error field with error", func(t *testing.T) {
		err := errors.New("test error")
		field := Error(err)
		if field.Key != "error" || field.Value != "test error" || field.Type != ErrorType {
			t.Errorf("Error field creation failed: %+v", field)
		}
	})

	t.Run("Error field with nil", func(t *testing.T) {
		field := Error(nil)
		if field.Key != "error" || field.Value != nil || field.Type != ErrorType {
			t.Errorf("Error field creation with nil failed: %+v", field)
		}
	})

	t.Run("ErrorNamed field", func(t *testing.T) {
		err := errors.New("custom error")
		field := ErrorNamed("custom_error", err)
		if field.Key != "custom_error" || field.Value != "custom error" || field.Type != ErrorType {
			t.Errorf("ErrorNamed field creation failed: %+v", field)
		}
	})

	t.Run("Object field", func(t *testing.T) {
		obj := map[string]interface{}{"nested": "value"}
		field := Object("data", obj)
		if field.Key != "data" || field.Type != ObjectType {
			t.Errorf("Object field creation failed: %+v", field)
		}
	})

	t.Run("Array field", func(t *testing.T) {
		arr := []string{"item1", "item2"}
		field := Array("items", arr)
		if field.Key != "items" || field.Type != ArrayType {
			t.Errorf("Array field creation failed: %+v", field)
		}
	})
}

func TestPredefinedFields(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected string
	}{
		{"TraceID", TraceID("trace-123"), "trace_id"},
		{"SpanID", SpanID("span-456"), "span_id"},
		{"UserID", UserID("user-789"), "user_id"},
		{"RequestID", RequestID("req-abc"), "request_id"},
		{"CorrelationID", CorrelationID("corr-def"), "correlation_id"},
		{"Method", Method("POST"), "method"},
		{"Path", Path("/api/users"), "path"},
		{"StatusCode", StatusCode(200), "status_code"},
		{"Operation", Operation("create_user"), "operation"},
		{"Component", Component("user-service"), "component"},
		{"Version", Version("1.0.0"), "version"},
		{"Environment", Environment("production"), "environment"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.field.Key != test.expected {
				t.Errorf("Expected key %s, got %s", test.expected, test.field.Key)
			}
		})
	}

	t.Run("Latency", func(t *testing.T) {
		duration := 150 * time.Millisecond
		field := Latency(duration)
		if field.Key != "latency" || field.Value != duration || field.Type != DurationType {
			t.Errorf("Latency field creation failed: %+v", field)
		}
	})
}

func TestValidateLevel(t *testing.T) {
	tests := []struct {
		level     Level
		shouldErr bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, false},
		{WarnLevel, false},
		{ErrorLevel, false},
		{FatalLevel, false},
		{PanicLevel, false},
		{Level(-99), true},
		{Level(99), true},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			err := ValidateLevel(test.level)
			if test.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		format    Format
		shouldErr bool
	}{
		{JSONFormat, false},
		{TextFormat, false},
		{ConsoleFormat, false},
		{Format(-1), true},
		{Format(99), true},
	}

	for _, test := range tests {
		t.Run(test.format.String(), func(t *testing.T) {
			err := ValidateFormat(test.format)
			if test.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
		}
		if err := ValidateConfig(config); err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("Invalid level", func(t *testing.T) {
		config := Config{
			Level:       Level(99),
			Format:      JSONFormat,
			ServiceName: "test-service",
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid level")
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      Format(99),
			ServiceName: "test-service",
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("Empty service name", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "",
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for empty service name")
		}
	})

	t.Run("Invalid async config", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
			Async: &AsyncConfig{
				BufferSize: 0, // Invalid
				Workers:    1,
			},
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid async buffer size")
		}
	})

	t.Run("Invalid async workers", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
			Async: &AsyncConfig{
				BufferSize: 1024,
				Workers:    0, // Invalid
			},
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid async workers")
		}
	})

	t.Run("Invalid sampling initial", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
			Sampling: &SamplingConfig{
				Initial:    -1, // Invalid
				Thereafter: 100,
			},
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid sampling initial")
		}
	})

	t.Run("Invalid sampling thereafter", func(t *testing.T) {
		config := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
			Sampling: &SamplingConfig{
				Initial:    100,
				Thereafter: -1, // Invalid
			},
		}
		if err := ValidateConfig(config); err == nil {
			t.Error("Expected error for invalid sampling thereafter")
		}
	})
}
