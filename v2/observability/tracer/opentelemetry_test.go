package tracer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOpenTelemetryConfig(t *testing.T) {
	config := DefaultOpenTelemetryConfig()

	assert.Equal(t, "unknown-service", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "localhost:4317", config.Endpoint)
	assert.True(t, config.Insecure)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 5*time.Second, config.BatchTimeout)
	assert.Equal(t, 512, config.MaxExportBatch)
	assert.Equal(t, 2048, config.MaxQueueSize)
	assert.Equal(t, 1.0, config.SamplingRatio)
	assert.Equal(t, []string{"tracecontext", "baggage"}, config.Propagators)
	assert.NotNil(t, config.ResourceAttrs)
}

func TestValidateOpenTelemetryConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *OpenTelemetryConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &OpenTelemetryConfig{
				ServiceName:   "test",
				Endpoint:      "localhost:4317",
				SamplingRatio: 0.5,
				Timeout:       30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "empty service name",
			config: &OpenTelemetryConfig{
				ServiceName: "",
				Endpoint:    "localhost:4317",
			},
			expectError: true,
		},
		{
			name: "empty endpoint",
			config: &OpenTelemetryConfig{
				ServiceName: "test",
				Endpoint:    "",
			},
			expectError: true,
		},
		{
			name: "negative sampling ratio",
			config: &OpenTelemetryConfig{
				ServiceName:   "test",
				Endpoint:      "localhost:4317",
				SamplingRatio: -0.1,
			},
			expectError: true,
		},
		{
			name: "sampling ratio > 1",
			config: &OpenTelemetryConfig{
				ServiceName:   "test",
				Endpoint:      "localhost:4317",
				SamplingRatio: 1.1,
			},
			expectError: true,
		},
		{
			name: "zero timeout",
			config: &OpenTelemetryConfig{
				ServiceName: "test",
				Endpoint:    "localhost:4317",
				Timeout:     0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOpenTelemetryConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertAttribute(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "key", "value"},
		{"int", "key", 42},
		{"int64", "key", int64(42)},
		{"float64", "key", 3.14},
		{"bool", "key", true},
		{"string slice", "key", []string{"a", "b"}},
		{"int slice", "key", []int{1, 2}},
		{"int64 slice", "key", []int64{1, 2}},
		{"float64 slice", "key", []float64{1.1, 2.2}},
		{"bool slice", "key", []bool{true, false}},
		{"unknown type", "key", struct{ Name string }{"test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := convertAttribute(tt.key, tt.value)
			assert.Equal(t, tt.key, string(attr.Key))
			assert.NotNil(t, attr.Value)
		})
	}
}
