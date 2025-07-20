package config

import (
	"os"
	"testing"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

func TestWithServiceName(t *testing.T) {
	config := NewConfig(WithServiceName("custom-service"))
	if config.ServiceName != "custom-service" {
		t.Errorf("Expected ServiceName to be 'custom-service', got %s", config.ServiceName)
	}
}

func TestWithEnvironment(t *testing.T) {
	config := NewConfig(WithEnvironment("production"))
	if config.Environment != "production" {
		t.Errorf("Expected Environment to be 'production', got %s", config.Environment)
	}
}

func TestWithExporterType(t *testing.T) {
	config := NewConfig(WithExporterType("datadog"))
	if config.ExporterType != "datadog" {
		t.Errorf("Expected ExporterType to be 'datadog', got %s", config.ExporterType)
	}
}

func TestWithEndpoint(t *testing.T) {
	endpoint := "http://localhost:4318/v1/traces"
	config := NewConfig(WithEndpoint(endpoint))
	if config.Endpoint != endpoint {
		t.Errorf("Expected Endpoint to be '%s', got %s", endpoint, config.Endpoint)
	}
}

func TestWithAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	config := NewConfig(WithAPIKey(apiKey))
	if config.APIKey != apiKey {
		t.Errorf("Expected APIKey to be '%s', got %s", apiKey, config.APIKey)
	}
}

func TestWithLicenseKey(t *testing.T) {
	licenseKey := "test-license-key"
	config := NewConfig(WithLicenseKey(licenseKey))
	if config.LicenseKey != licenseKey {
		t.Errorf("Expected LicenseKey to be '%s', got %s", licenseKey, config.LicenseKey)
	}
}

func TestWithVersion(t *testing.T) {
	version := "2.0.0"
	config := NewConfig(WithVersion(version))
	if config.Version != version {
		t.Errorf("Expected Version to be '%s', got %s", version, config.Version)
	}
}

func TestWithSamplingRatio(t *testing.T) {
	ratio := 0.75
	config := NewConfig(WithSamplingRatio(ratio))
	if config.SamplingRatio != ratio {
		t.Errorf("Expected SamplingRatio to be %f, got %f", ratio, config.SamplingRatio)
	}
}

func TestWithPropagators(t *testing.T) {
	propagators := []string{"tracecontext", "b3", "jaeger"}
	config := NewConfig(WithPropagators(propagators...))

	if len(config.Propagators) != len(propagators) {
		t.Errorf("Expected %d propagators, got %d", len(propagators), len(config.Propagators))
	}

	for i, expected := range propagators {
		if config.Propagators[i] != expected {
			t.Errorf("Expected propagator %d to be '%s', got '%s'", i, expected, config.Propagators[i])
		}
	}
}

func TestWithHeaders(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom":      "custom-value",
	}
	config := NewConfig(WithHeaders(headers))

	for key, expectedValue := range headers {
		if config.Headers[key] != expectedValue {
			t.Errorf("Expected header '%s' to be '%s', got '%s'", key, expectedValue, config.Headers[key])
		}
	}
}

func TestWithHeader(t *testing.T) {
	config := NewConfig(
		WithHeader("Authorization", "Bearer token123"),
		WithHeader("X-Custom", "custom-value"),
	)

	if config.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header to be 'Bearer token123', got %s", config.Headers["Authorization"])
	}

	if config.Headers["X-Custom"] != "custom-value" {
		t.Errorf("Expected X-Custom header to be 'custom-value', got %s", config.Headers["X-Custom"])
	}
}

func TestWithAttributes(t *testing.T) {
	attributes := map[string]string{
		"team":   "platform",
		"region": "us-east-1",
	}
	config := NewConfig(WithAttributes(attributes))

	for key, expectedValue := range attributes {
		if config.Attributes[key] != expectedValue {
			t.Errorf("Expected attribute '%s' to be '%s', got '%s'", key, expectedValue, config.Attributes[key])
		}
	}
}

func TestWithAttribute(t *testing.T) {
	config := NewConfig(
		WithAttribute("team", "platform"),
		WithAttribute("region", "us-east-1"),
	)

	if config.Attributes["team"] != "platform" {
		t.Errorf("Expected team attribute to be 'platform', got %s", config.Attributes["team"])
	}

	if config.Attributes["region"] != "us-east-1" {
		t.Errorf("Expected region attribute to be 'us-east-1', got %s", config.Attributes["region"])
	}
}

func TestWithInsecure(t *testing.T) {
	config := NewConfig(WithInsecure(true))
	if !config.Insecure {
		t.Error("Expected Insecure to be true")
	}

	config = NewConfig(WithInsecure(false))
	if config.Insecure {
		t.Error("Expected Insecure to be false")
	}
}

func TestNewConfig_MultipleOptions(t *testing.T) {
	config := NewConfig(
		WithServiceName("test-service"),
		WithEnvironment("production"),
		WithExporterType("datadog"),
		WithAPIKey("test-api-key"),
		WithSamplingRatio(0.5),
		WithPropagators("tracecontext", "b3"),
		WithHeader("Authorization", "Bearer token"),
		WithAttribute("team", "platform"),
		WithInsecure(true),
	)

	if config.ServiceName != "test-service" {
		t.Errorf("Expected ServiceName to be 'test-service', got %s", config.ServiceName)
	}

	if config.Environment != "production" {
		t.Errorf("Expected Environment to be 'production', got %s", config.Environment)
	}

	if config.ExporterType != "datadog" {
		t.Errorf("Expected ExporterType to be 'datadog', got %s", config.ExporterType)
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got %s", config.APIKey)
	}

	if config.SamplingRatio != 0.5 {
		t.Errorf("Expected SamplingRatio to be 0.5, got %f", config.SamplingRatio)
	}

	if len(config.Propagators) != 2 || config.Propagators[0] != "tracecontext" || config.Propagators[1] != "b3" {
		t.Errorf("Expected propagators to be ['tracecontext', 'b3'], got %v", config.Propagators)
	}

	if config.Headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header to be 'Bearer token', got %s", config.Headers["Authorization"])
	}

	if config.Attributes["team"] != "platform" {
		t.Errorf("Expected team attribute to be 'platform', got %s", config.Attributes["team"])
	}

	if !config.Insecure {
		t.Error("Expected Insecure to be true")
	}
}

func TestNewConfigFromEnv(t *testing.T) {
	// Set test environment variables
	testEnvVars := map[string]string{
		"TRACER_SERVICE_NAME":   "env-service",
		"TRACER_ENVIRONMENT":    "staging",
		"TRACER_EXPORTER_TYPE":  "opentelemetry",
		"TRACER_ENDPOINT":       "http://localhost:4317",
		"TRACER_SAMPLING_RATIO": "0.8",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	// Test with additional options
	config := NewConfigFromEnv(
		WithServiceName("override-service"), // Should override env
		WithAPIKey("test-api-key"),          // Should add to env config
	)

	if config.ServiceName != "override-service" {
		t.Errorf("Expected ServiceName to be 'override-service', got %s", config.ServiceName)
	}

	if config.Environment != "staging" {
		t.Errorf("Expected Environment to be 'staging' from env, got %s", config.Environment)
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got %s", config.APIKey)
	}

	if config.SamplingRatio != 0.8 {
		t.Errorf("Expected SamplingRatio to be 0.8 from env, got %f", config.SamplingRatio)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "unknown-service" {
		t.Errorf("Expected ServiceName to be 'unknown-service', got %s", config.ServiceName)
	}

	if config.Environment != "development" {
		t.Errorf("Expected Environment to be 'development', got %s", config.Environment)
	}

	if config.ExporterType != "opentelemetry" {
		t.Errorf("Expected ExporterType to be 'opentelemetry', got %s", config.ExporterType)
	}

	if config.SamplingRatio != 1.0 {
		t.Errorf("Expected SamplingRatio to be 1.0, got %f", config.SamplingRatio)
	}

	if len(config.Propagators) != 2 {
		t.Errorf("Expected 2 propagators, got %d", len(config.Propagators))
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set test environment variables
	testEnvVars := map[string]string{
		"TRACER_SERVICE_NAME":    "test-service",
		"TRACER_ENVIRONMENT":     "production",
		"TRACER_EXPORTER_TYPE":   "datadog",
		"TRACER_ENDPOINT":        "http://localhost:8080",
		"TRACER_API_KEY":         "test-api-key",
		"TRACER_LICENSE_KEY":     "test-license-key",
		"TRACER_VERSION":         "2.0.0",
		"TRACER_SAMPLING_RATIO":  "0.5",
		"TRACER_INSECURE":        "true",
		"TRACER_PROPAGATORS":     "tracecontext,b3,jaeger",
		"TRACER_HEADER_AUTH":     "Bearer token123",
		"TRACER_HEADER_X_CUSTOM": "custom-value",
		"TRACER_ATTR_TEAM":       "platform",
		"TRACER_ATTR_REGION":     "us-east-1",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	config := LoadFromEnv()

	if config.ServiceName != "test-service" {
		t.Errorf("Expected ServiceName to be 'test-service', got %s", config.ServiceName)
	}

	if config.Environment != "production" {
		t.Errorf("Expected Environment to be 'production', got %s", config.Environment)
	}

	if config.ExporterType != "datadog" {
		t.Errorf("Expected ExporterType to be 'datadog', got %s", config.ExporterType)
	}

	if config.Endpoint != "http://localhost:8080" {
		t.Errorf("Expected Endpoint to be 'http://localhost:8080', got %s", config.Endpoint)
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got %s", config.APIKey)
	}

	if config.LicenseKey != "test-license-key" {
		t.Errorf("Expected LicenseKey to be 'test-license-key', got %s", config.LicenseKey)
	}

	if config.Version != "2.0.0" {
		t.Errorf("Expected Version to be '2.0.0', got %s", config.Version)
	}

	if config.SamplingRatio != 0.5 {
		t.Errorf("Expected SamplingRatio to be 0.5, got %f", config.SamplingRatio)
	}

	if !config.Insecure {
		t.Error("Expected Insecure to be true")
	}

	expectedPropagators := []string{"tracecontext", "b3", "jaeger"}
	if len(config.Propagators) != len(expectedPropagators) {
		t.Errorf("Expected %d propagators, got %d", len(expectedPropagators), len(config.Propagators))
	}

	for i, expected := range expectedPropagators {
		if config.Propagators[i] != expected {
			t.Errorf("Expected propagator %d to be '%s', got '%s'", i, expected, config.Propagators[i])
		}
	}

	if config.Headers["auth"] != "Bearer token123" {
		t.Errorf("Expected header 'auth' to be 'Bearer token123', got %s", config.Headers["auth"])
	}

	if config.Headers["x-custom"] != "custom-value" {
		t.Errorf("Expected header 'x-custom' to be 'custom-value', got %s", config.Headers["x-custom"])
	}

	if config.Attributes["team"] != "platform" {
		t.Errorf("Expected attribute 'team' to be 'platform', got %s", config.Attributes["team"])
	}

	if config.Attributes["region"] != "us-east-1" {
		t.Errorf("Expected attribute 'region' to be 'us-east-1', got %s", config.Attributes["region"])
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    interfaces.Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid opentelemetry config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://localhost:4317",
				SamplingRatio: 1.0,
			},
			wantError: false,
		},
		{
			name: "valid datadog config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "datadog",
				APIKey:        "test-key",
				SamplingRatio: 0.5,
			},
			wantError: false,
		},
		{
			name: "valid newrelic config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "newrelic",
				LicenseKey:    "test-license",
				SamplingRatio: 0.8,
			},
			wantError: false,
		},
		{
			name: "valid grafana config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "grafana",
				Endpoint:      "http://tempo:3200",
				SamplingRatio: 1.0,
			},
			wantError: false,
		},
		{
			name: "missing service name",
			config: interfaces.Config{
				ExporterType:  "opentelemetry",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "service name is required",
		},
		{
			name: "missing exporter type",
			config: interfaces.Config{
				ServiceName:   "test-service",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "exporter type is required",
		},
		{
			name: "unsupported exporter type",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "unsupported",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "unsupported exporter type",
		},
		{
			name: "invalid sampling ratio - negative",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://localhost:4317",
				SamplingRatio: -0.1,
			},
			wantError: true,
			errorMsg:  "sampling ratio must be between 0 and 1",
		},
		{
			name: "invalid sampling ratio - greater than 1",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://localhost:4317",
				SamplingRatio: 1.5,
			},
			wantError: true,
			errorMsg:  "sampling ratio must be between 0 and 1",
		},
		{
			name: "datadog missing API key",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "datadog",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "API key is required for Datadog exporter",
		},
		{
			name: "newrelic missing license key",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "newrelic",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "License key is required for New Relic exporter",
		},
		{
			name: "opentelemetry missing endpoint",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "opentelemetry",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "endpoint is required for opentelemetry exporter",
		},
		{
			name: "grafana missing endpoint",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "grafana",
				SamplingRatio: 1.0,
			},
			wantError: true,
			errorMsg:  "endpoint is required for grafana exporter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	base := interfaces.Config{
		ServiceName:   "base-service",
		Environment:   "development",
		ExporterType:  "opentelemetry",
		SamplingRatio: 1.0,
		Headers:       map[string]string{"base-header": "base-value"},
		Attributes:    map[string]string{"base-attr": "base-value"},
	}

	override := interfaces.Config{
		ServiceName: "override-service",
		Environment: "production",
		APIKey:      "override-key",
		Headers:     map[string]string{"override-header": "override-value"},
		Attributes:  map[string]string{"override-attr": "override-value"},
	}

	result := MergeConfigs(base, override)

	if result.ServiceName != "override-service" {
		t.Errorf("Expected ServiceName to be 'override-service', got %s", result.ServiceName)
	}

	if result.Environment != "production" {
		t.Errorf("Expected Environment to be 'production', got %s", result.Environment)
	}

	if result.ExporterType != "opentelemetry" {
		t.Errorf("Expected ExporterType to remain 'opentelemetry', got %s", result.ExporterType)
	}

	if result.APIKey != "override-key" {
		t.Errorf("Expected APIKey to be 'override-key', got %s", result.APIKey)
	}

	if result.Headers["base-header"] != "base-value" {
		t.Errorf("Expected base header to be preserved")
	}

	if result.Headers["override-header"] != "override-value" {
		t.Errorf("Expected override header to be applied")
	}

	if result.Attributes["base-attr"] != "base-value" {
		t.Errorf("Expected base attribute to be preserved")
	}

	if result.Attributes["override-attr"] != "override-value" {
		t.Errorf("Expected override attribute to be applied")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
