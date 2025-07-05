package activemq

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActiveMQProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.ProviderConfig{
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:61613"},
					Auth: &interfaces.AuthConfig{
						Username: "admin",
						Password: "admin",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "disabled provider",
			config: &config.ProviderConfig{
				Enabled: false,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:61613"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing connection config but should still work with defaults",
			config: &config.ProviderConfig{
				Enabled: true,
			},
			wantErr: false, // ActiveMQ provider uses defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewActiveMQProvider(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				// Verify provider type
				if provider != nil {
					assert.Equal(t, interfaces.ProviderActiveMQ, provider.GetType())
				}
			}
		})
	}
}

func TestActiveMQProvider_Connect(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
			ConnectTimeout: 5 * time.Second,
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	ctx := context.Background()

	// Test connect (will likely fail due to no real ActiveMQ server)
	err = provider.Connect(ctx)
	// We expect this to fail in test environment
	assert.Error(t, err)

	// Test IsConnected when not connected
	assert.False(t, provider.IsConnected())
}

func TestActiveMQProvider_HealthCheck(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Health check without connection should fail
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestActiveMQProvider_Close(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)

	// Close should work even if not connected
	err = provider.Close()
	assert.NoError(t, err)

	// Multiple closes should be safe
	err = provider.Close()
	assert.NoError(t, err)
}

func TestActiveMQProvider_GetMetrics(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)

	metrics := provider.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ActiveConnections)
	assert.NotZero(t, metrics.LastHealthCheck)
}

func TestActiveMQProvider_CreateProducer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)

	producerConfig := &interfaces.ProducerConfig{
		ID:         "test-producer",
		BufferSize: 100,
	}

	producer, err := provider.CreateProducer(producerConfig)
	// Expected to fail due to no real connection
	assert.Error(t, err)
	assert.Nil(t, producer)
}

func TestActiveMQProvider_CreateConsumer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
		},
	}

	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)

	consumerConfig := &interfaces.ConsumerConfig{
		ID:            "test-consumer",
		ConsumerGroup: "test-group",
		AutoCommit:    true,
	}

	consumer, err := provider.CreateConsumer(consumerConfig)
	// Expected to fail due to no real connection
	assert.Error(t, err)
	assert.Nil(t, consumer)
}

func TestActiveMQProvider_Integration(t *testing.T) {
	// This test demonstrates the full provider workflow
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:61613"},
			Auth: &interfaces.AuthConfig{
				Username: "admin",
				Password: "admin",
			},
			ConnectTimeout:   5 * time.Second,
			OperationTimeout: 10 * time.Second,
		},
	}

	// Create provider
	provider, err := NewActiveMQProvider(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Get initial metrics
	metrics := provider.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ActiveConnections)

	// Try to connect (will fail in test environment)
	ctx := context.Background()
	err = provider.Connect(ctx)
	assert.Error(t, err) // Expected to fail without real ActiveMQ

	// Health check should fail
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)

	// Close should always work
	err = provider.Close()
	assert.NoError(t, err)
}
