package opentelemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
	"github.com/fsvxavier/nexs-lib/observability/tracer/mocks"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	assert.NotNil(t, provider)
}

func TestProvider_Init_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config interfaces.Config
	}{
		{
			name: "jaeger http endpoint",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://jaeger:14268/api/traces",
				SamplingRatio: 1.0,
				Version:       "1.0.0",
				Propagators:   []string{"tracecontext"},
			},
		},
		{
			name: "jaeger grpc endpoint",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "production",
				ExporterType:  "opentelemetry",
				Endpoint:      "jaeger:14250",
				SamplingRatio: 0.5,
				Version:       "2.0.0",
				Propagators:   []string{"tracecontext", "b3"},
			},
		},
		{
			name: "config with headers",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://otel-collector:4318/v1/traces",
				SamplingRatio: 1.0,
				Headers:       map[string]string{"authorization": "Bearer token"},
				Propagators:   []string{"tracecontext"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := mocks.NewMockProviderForBackend("opentelemetry")
			ctx := context.Background()

			tracerProvider, err := mockProvider.Init(ctx, tt.config)

			assert.NoError(t, err)
			assert.NotNil(t, tracerProvider)
			assert.True(t, mockProvider.WasInitCalled())
			assert.True(t, mockProvider.IsInitialized())
			assert.Equal(t, tt.config, mockProvider.GetLastConfig())

			err = mockProvider.Shutdown(ctx)
			assert.NoError(t, err)
			assert.True(t, mockProvider.WasShutdownCalled())
			assert.False(t, mockProvider.IsInitialized())
		})
	}
}

func TestProvider_Init_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  interfaces.Config
		wantErr bool
	}{
		{
			name: "invalid propagator",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "opentelemetry",
				Endpoint:      "http://jaeger:14268/api/traces",
				SamplingRatio: 1.0,
				Propagators:   []string{"invalid-propagator"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := mocks.NewMockProviderForBackend("opentelemetry")
			if tt.wantErr {
				mockProvider.SetInitError(errors.New("mock error: invalid propagator"))
			}
			ctx := context.Background()

			_, err := mockProvider.Init(ctx, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockProvider.Shutdown(ctx)
		})
	}
}

func TestProvider_Shutdown_NoInit(t *testing.T) {
	mockProvider := mocks.NewMockProviderForBackend("opentelemetry")
	ctx := context.Background()

	err := mockProvider.Shutdown(ctx)
	assert.NoError(t, err)
	assert.True(t, mockProvider.WasShutdownCalled())
	assert.False(t, mockProvider.IsInitialized())
}

func TestProvider_MultipleInit(t *testing.T) {
	mockProvider := mocks.NewMockProviderForBackend("opentelemetry")
	config := interfaces.Config{
		ServiceName:   "test-service",
		Environment:   "test",
		ExporterType:  "opentelemetry",
		Endpoint:      "http://jaeger:14268/api/traces",
		SamplingRatio: 1.0,
		Propagators:   []string{"tracecontext"},
	}

	ctx := context.Background()

	_, err := mockProvider.Init(ctx, config)
	assert.NoError(t, err)
	assert.True(t, mockProvider.IsInitialized())

	_, err = mockProvider.Init(ctx, config)
	assert.NoError(t, err)
	assert.True(t, mockProvider.IsInitialized())
	assert.Equal(t, 2, mockProvider.GetInitCallCount())

	err = mockProvider.Shutdown(ctx)
	assert.NoError(t, err)
}
