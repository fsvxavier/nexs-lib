package sqs

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/sqs/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQSProvider(t *testing.T) {
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
					Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
					Auth: &interfaces.AuthConfig{
						Username: "access-key",
						Password: "secret-key",
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
					Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing connection config but should still work with defaults",
			config: &config.ProviderConfig{
				Enabled: true,
			},
			wantErr: false, // SQS provider uses defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSQSProvider(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				// Verify provider type
				if provider != nil {
					assert.Equal(t, interfaces.ProviderSQS, provider.GetType())
				}
			}
		})
	}
}

func TestSQSProvider_Connect(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
			ConnectTimeout: 5 * time.Second,
		},
	}

	provider, err := NewSQSProvider(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	ctx := context.Background()

	// Test connect (SQS creates client locally, so this should succeed)
	err = provider.Connect(ctx)
	// SQS doesn't require network connection for client creation
	assert.NoError(t, err)

	// Test IsConnected after successful connect
	assert.True(t, provider.IsConnected())
}

func TestSQSProvider_HealthCheck(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
		},
	}

	provider, err := NewSQSProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Health check without connection should fail
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestSQSProvider_Close(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
		},
	}

	provider, err := NewSQSProvider(config)
	require.NoError(t, err)

	// Close should work even if not connected
	err = provider.Close()
	assert.NoError(t, err)

	// Multiple closes should be safe
	err = provider.Close()
	assert.NoError(t, err)
}

func TestSQSProvider_GetMetrics(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
		},
	}

	provider, err := NewSQSProvider(config)
	require.NoError(t, err)

	metrics := provider.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ActiveConnections)
	assert.NotZero(t, metrics.LastHealthCheck)
}

func TestSQSProvider_CreateProducer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
		},
	}

	provider, err := NewSQSProvider(config)
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

func TestSQSProvider_CreateConsumer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
		},
	}

	provider, err := NewSQSProvider(config)
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

func TestSQSProvider_Integration(t *testing.T) {
	// This test demonstrates the full provider workflow
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
			Auth: &interfaces.AuthConfig{
				Username: "access-key",
				Password: "secret-key",
			},
			ConnectTimeout:   5 * time.Second,
			OperationTimeout: 10 * time.Second,
		},
	}

	// Create provider
	provider, err := NewSQSProvider(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Get initial metrics
	metrics := provider.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ActiveConnections)

	// Try to connect (SQS should succeed locally)
	ctx := context.Background()
	err = provider.Connect(ctx)
	assert.NoError(t, err) // SQS client creation should succeed

	// Health check should fail
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)

	// Close should always work
	err = provider.Close()
	assert.NoError(t, err)
}

// TestSQSMockProvider testa o provider mock do SQS
func TestSQSMockProvider(t *testing.T) {
	mockProvider := mock.NewMockSQSProvider()

	// Test GetType
	assert.Equal(t, interfaces.ProviderSQS, mockProvider.GetType())

	// Test initial state
	assert.False(t, mockProvider.IsConnected())

	// Test Connect
	ctx := context.Background()
	err := mockProvider.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, mockProvider.IsConnected())

	// Test HealthCheck after connect
	err = mockProvider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test CreateProducer
	producerConfig := &interfaces.ProducerConfig{
		ID: "test-producer",
	}
	producer, err := mockProvider.CreateProducer(producerConfig)
	require.NoError(t, err)
	require.NotNil(t, producer)
	assert.True(t, producer.IsConnected())

	// Test CreateConsumer
	consumerConfig := &interfaces.ConsumerConfig{
		ID: "test-consumer",
	}
	consumer, err := mockProvider.CreateConsumer(consumerConfig)
	require.NoError(t, err)
	require.NotNil(t, consumer)
	assert.True(t, consumer.IsConnected())

	// Test GetMetrics
	metrics := mockProvider.GetMetrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.HealthCheckStatus)

	// Test Disconnect
	err = mockProvider.Disconnect()
	assert.NoError(t, err)
	assert.False(t, mockProvider.IsConnected())

	// Test Close
	err = mockProvider.Close()
	assert.NoError(t, err)
	assert.False(t, mockProvider.IsConnected())
}

// TestSQSMockProducer testa o producer mock do SQS
func TestSQSMockProducer(t *testing.T) {
	producer := mock.NewMockSQSProducer()

	// Test initial state
	assert.True(t, producer.IsConnected())

	// Test Publish
	ctx := context.Background()
	message := &interfaces.Message{
		ID:   "test-1",
		Body: []byte("test message"),
	}

	err := producer.Publish(ctx, "test-queue", message, nil)
	require.NoError(t, err)

	// Test PublishBatch
	messages := []*interfaces.Message{
		{ID: "test-2", Body: []byte("test message 2")},
		{ID: "test-3", Body: []byte("test message 3")},
	}

	err = producer.PublishBatch(ctx, "test-queue", messages, nil)
	require.NoError(t, err)

	// Test GetMetrics
	metrics := producer.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.MessagesSent) // 1 + 2 from batch

	// Test Close
	err = producer.Close()
	assert.NoError(t, err)
	assert.False(t, producer.IsConnected())
}

// TestSQSMockConsumer testa o consumer mock do SQS
func TestSQSMockConsumer(t *testing.T) {
	consumer := mock.NewMockSQSConsumer()

	// Test initial state
	assert.True(t, consumer.IsConnected())
	assert.False(t, consumer.IsRunning())

	// Test Subscribe
	ctx := context.Background()
	handler := func(ctx context.Context, message *interfaces.Message) error {
		return nil
	}

	err := consumer.Subscribe(ctx, "test-queue", nil, handler)
	require.NoError(t, err)
	assert.True(t, consumer.IsRunning())

	// Test message acknowledgment
	message := &interfaces.Message{
		ID:   "test-1",
		Body: []byte("test message"),
	}

	err = consumer.Ack(message)
	assert.NoError(t, err)

	err = consumer.Nack(message, true)
	assert.NoError(t, err)

	// Test GetMetrics
	metrics := consumer.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.MessagesAcked)
	assert.Equal(t, int64(1), metrics.MessagesNacked)

	// Test Pause/Resume
	err = consumer.Pause()
	assert.NoError(t, err)

	err = consumer.Resume()
	assert.NoError(t, err)

	// Test Close
	err = consumer.Close()
	assert.NoError(t, err)
	assert.False(t, consumer.IsConnected())
	assert.False(t, consumer.IsRunning())
}
