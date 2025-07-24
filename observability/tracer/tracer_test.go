package tracer

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

func TestNewTracerManager(t *testing.T) {
	tm := NewTracerManager()
	if tm == nil {
		t.Fatal("NewTracerManager() returned nil")
	}

	if tm.provider != nil {
		t.Error("NewTracerManager() should return uninitialized manager")
	}
}

func TestTracerManager_Init_ValidConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config interfaces.Config
	}{
		{
			name: "opentelemetry config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://localhost:4318/v1/traces",
				SamplingRatio: 1.0,
				Version:       "1.0.0",
				Propagators:   []string{"tracecontext"},
			},
		},
		{
			name: "datadog config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "datadog",
				APIKey:        "test-api-key",
				SamplingRatio: 0.5,
				Version:       "1.0.0",
			},
		},
		{
			name: "newrelic config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "newrelic",
				LicenseKey:    "test-license-key",
				SamplingRatio: 0.8,
				Version:       "1.0.0",
			},
		},
		{
			name: "grafana config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "grafana",
				Endpoint:      "http://tempo:3200",
				SamplingRatio: 1.0,
				Version:       "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTracerManager()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			tracerProvider, err := tm.Init(ctx, tt.config)
			if err != nil {
				t.Fatalf("Init() error = %v", err)
			}

			if tracerProvider == nil {
				t.Fatal("Init() returned nil tracer provider")
			}

			if tm.provider == nil {
				t.Error("TracerManager should have provider set after Init()")
			}

			if tm.config.ServiceName != tt.config.ServiceName {
				t.Errorf("Config not stored correctly, got %s, want %s", tm.config.ServiceName, tt.config.ServiceName)
			}

			// Test shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()

			if err := tm.Shutdown(shutdownCtx); err != nil {
				t.Errorf("Shutdown() error = %v", err)
			}
		})
	}
}

func TestTracerManager_Init_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  interfaces.Config
		wantErr bool
	}{
		{
			name: "missing service name",
			config: interfaces.Config{
				ExporterType:  "opentelemetry",
				SamplingRatio: 1.0,
			},
			wantErr: true,
		},
		{
			name: "unsupported exporter type",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "unsupported",
				SamplingRatio: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid sampling ratio",
			config: interfaces.Config{
				ServiceName:   "test-service",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://localhost:4318",
				SamplingRatio: 2.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTracerManager()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := tm.Init(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Always try to shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			tm.Shutdown(shutdownCtx)
		})
	}
}

func TestTracerManager_Shutdown_NoInit(t *testing.T) {
	tm := NewTracerManager()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should not error even if not initialized
	if err := tm.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestTracerManager_GetConfig(t *testing.T) {
	tm := NewTracerManager()
	config := interfaces.Config{
		ServiceName:   "test-service",
		Environment:   "test",
		ExporterType:  "opentelemetry",
		Endpoint:      "http://localhost:4318/v1/traces",
		SamplingRatio: 1.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := tm.Init(ctx, config)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	retrievedConfig := tm.GetConfig()
	if retrievedConfig.ServiceName != config.ServiceName {
		t.Errorf("GetConfig() service name = %s, want %s", retrievedConfig.ServiceName, config.ServiceName)
	}

	if retrievedConfig.ExporterType != config.ExporterType {
		t.Errorf("GetConfig() exporter type = %s, want %s", retrievedConfig.ExporterType, config.ExporterType)
	}

	// Cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	tm.Shutdown(shutdownCtx)
}

func TestTracerManager_GetProvider(t *testing.T) {
	tm := NewTracerManager()
	config := interfaces.Config{
		ServiceName:   "test-service",
		Environment:   "test",
		ExporterType:  "opentelemetry",
		Endpoint:      "http://localhost:4318/v1/traces",
		SamplingRatio: 1.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := tm.Init(ctx, config)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	provider := tm.GetProvider()
	if provider == nil {
		t.Error("GetProvider() returned nil")
	}

	// Cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	tm.Shutdown(shutdownCtx)
}

func TestNewTracerProvider(t *testing.T) {
	tests := []struct {
		name         string
		exporterType string
		wantErr      bool
	}{
		{"datadog", "datadog", false},
		{"grafana", "grafana", false},
		{"newrelic", "newrelic", false},
		{"opentelemetry", "opentelemetry", false},
		{"unsupported", "unsupported", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := interfaces.Config{
				ServiceName:  "test-service",
				ExporterType: tt.exporterType,
			}

			provider, err := NewTracerProvider(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTracerProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && provider == nil {
				t.Error("NewTracerProvider() returned nil provider for valid type")
			}
		})
	}
}

func TestFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Fatal("NewFactory() returned nil")
	}

	// Test SupportedTypes
	types := factory.SupportedTypes()
	expectedTypes := []string{"datadog", "grafana", "newrelic", "opentelemetry"}

	if len(types) != len(expectedTypes) {
		t.Errorf("SupportedTypes() returned %d types, want %d", len(types), len(expectedTypes))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, t := range types {
			if t == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedTypes() missing type: %s", expectedType)
		}
	}

	// Test CreateProvider
	config := interfaces.Config{
		ServiceName:  "test-service",
		ExporterType: "opentelemetry",
	}

	provider, err := factory.CreateProvider(config)
	if err != nil {
		t.Errorf("CreateProvider() error = %v", err)
	}

	if provider == nil {
		t.Error("CreateProvider() returned nil provider")
	}
}

func TestQuickStart(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		exporterType string
		wantErr      bool
	}{
		{"valid opentelemetry", "test-service", "opentelemetry", false},
		{"valid datadog", "test-service", "datadog", false},
		{"invalid exporter", "test-service", "invalid", true},
		{"empty service name", "", "opentelemetry", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, tm, err := QuickStart(tt.serviceName, tt.exporterType)
			if (err != nil) != tt.wantErr {
				t.Errorf("QuickStart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if provider == nil {
					t.Error("QuickStart() returned nil provider")
				}
				if tm == nil {
					t.Error("QuickStart() returned nil tracer manager")
				}

				// Cleanup
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				tm.Shutdown(ctx)
			}
		})
	}
}
