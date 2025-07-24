package newrelic

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
			name: "basic config",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "newrelic",
				APIKey:        "test-api-key",
				SamplingRatio: 1.0,
				Version:       "1.0.0",
				Propagators:   []string{"tracecontext"},
			},
		},
		{
			name: "config with custom endpoint",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "production",
				ExporterType:  "newrelic",
				APIKey:        "test-api-key",
				Endpoint:      "https://trace-api.eu.newrelic.com/trace/v1",
				SamplingRatio: 0.5,
				Version:       "2.0.0",
				Propagators:   []string{"tracecontext", "b3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Usar mock em vez do provider real para testes unitários isolados
			mockProvider := mocks.NewMockProviderForBackend("newrelic")
			ctx := context.Background()

			tracerProvider, err := mockProvider.Init(ctx, tt.config)

			assert.NoError(t, err)
			assert.NotNil(t, tracerProvider)
			assert.True(t, mockProvider.WasInitCalled())
			assert.True(t, mockProvider.IsInitialized())
			assert.Equal(t, tt.config, mockProvider.GetLastConfig())

			// Test shutdown
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
			name: "missing api key",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "newrelic",
				SamplingRatio: 1.0,
				Propagators:   []string{"tracecontext"},
			},
			wantErr: true,
		},
		{
			name: "invalid propagator",
			config: interfaces.Config{
				ServiceName:   "test-service",
				Environment:   "test",
				ExporterType:  "newrelic",
				APIKey:        "test-api-key",
				SamplingRatio: 1.0,
				Propagators:   []string{"invalid-propagator"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Usar mock configurado para simular erro
			mockProvider := mocks.NewMockProviderForBackend("newrelic")
			if tt.wantErr {
				mockProvider.SetInitError(errors.New("mock error: " + tt.name))
			}
			ctx := context.Background()

			_, err := mockProvider.Init(ctx, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Sempre tentar shutdown
			mockProvider.Shutdown(ctx)
		})
	}
}

func TestProvider_Shutdown_NoInit(t *testing.T) {
	// Usar mock para testar comportamento de shutdown sem init
	mockProvider := mocks.NewMockProviderForBackend("newrelic")
	ctx := context.Background()

	// Shutdown sem init não deve dar erro
	err := mockProvider.Shutdown(ctx)
	assert.NoError(t, err)
	assert.True(t, mockProvider.WasShutdownCalled())
	assert.False(t, mockProvider.IsInitialized())
}

func TestProvider_MultipleInit(t *testing.T) {
	// Usar mock para testar múltiplas inicializações
	mockProvider := mocks.NewMockProviderForBackend("newrelic")
	config := interfaces.Config{
		ServiceName:   "test-service",
		Environment:   "test",
		ExporterType:  "newrelic",
		APIKey:        "test-api-key",
		SamplingRatio: 1.0,
		Propagators:   []string{"tracecontext"},
	}

	ctx := context.Background()

	// Primeira inicialização
	_, err := mockProvider.Init(ctx, config)
	assert.NoError(t, err)
	assert.True(t, mockProvider.IsInitialized())

	// Segunda inicialização
	_, err = mockProvider.Init(ctx, config)
	assert.NoError(t, err)
	assert.True(t, mockProvider.IsInitialized())
	assert.Equal(t, 2, mockProvider.GetInitCallCount())

	// Cleanup
	err = mockProvider.Shutdown(ctx)
	assert.NoError(t, err)
}
