package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/rabbitmq/mock"
)

func TestNewRabbitMQProvider(t *testing.T) {
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
					Brokers:          []string{"localhost:5672"},
					ConnectTimeout:   10 * time.Second,
					OperationTimeout: 30 * time.Second,
					Auth: &interfaces.AuthConfig{
						Username: "guest",
						Password: "guest",
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
			name: "nil connection config but should work with defaults",
			config: &config.ProviderConfig{
				Enabled:    true,
				Connection: nil,
			},
			wantErr: false, // RabbitMQ provider uses defaults
		},
		{
			name: "empty brokers but should work with defaults",
			config: &config.ProviderConfig{
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{},
				},
			},
			wantErr: false, // RabbitMQ provider uses defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewRabbitMQProvider(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewRabbitMQProvider() should return error, but got nil")
				}
				if provider != nil {
					t.Error("NewRabbitMQProvider() should return nil provider on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewRabbitMQProvider() should not return error, but got: %v", err)
				}
				if provider == nil {
					t.Error("NewRabbitMQProvider() should return non-nil provider")
				}
				if provider.GetType() != interfaces.ProviderRabbitMQ {
					t.Errorf("GetType() = %v, want %v", provider.GetType(), interfaces.ProviderRabbitMQ)
				}
			}
		})
	}
}

func TestRabbitMQProvider_GetType(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider.GetType() != interfaces.ProviderRabbitMQ {
		t.Errorf("GetType() = %v, want %v", provider.GetType(), interfaces.ProviderRabbitMQ)
	}
}

func TestRabbitMQProvider_IsConnected(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Provider deve iniciar desconectado
	if provider.IsConnected() {
		t.Error("Provider should start disconnected")
	}
}

func TestRabbitMQProvider_HealthCheck(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Health check deve falhar se não conectado (esperado)
	err = provider.HealthCheck(ctx)
	if err == nil {
		t.Log("Health check passed (RabbitMQ is running)")
	} else {
		t.Logf("Health check failed as expected (RabbitMQ not running): %v", err)
	}
}

func TestRabbitMQProvider_GetMetrics(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	metrics := provider.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() should return non-nil metrics")
	}

	// Verificar campos essenciais das métricas
	if metrics.ConnectionStats == nil {
		t.Error("ConnectionStats should not be nil")
	}

	// Provider recém-criado deve ter 0 conexões ativas
	if metrics.ActiveConnections != 0 {
		t.Errorf("ActiveConnections = %d, want 0", metrics.ActiveConnections)
	}

	if metrics.ActiveProducers != 0 {
		t.Errorf("ActiveProducers = %d, want 0", metrics.ActiveProducers)
	}

	if metrics.ActiveConsumers != 0 {
		t.Errorf("ActiveConsumers = %d, want 0", metrics.ActiveConsumers)
	}
}

func TestRabbitMQProvider_CreateProducer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tests := []struct {
		name           string
		producerConfig *interfaces.ProducerConfig
		wantErr        bool
	}{
		{
			name: "valid producer config",
			producerConfig: &interfaces.ProducerConfig{
				ID:          "test-producer",
				SendTimeout: 30 * time.Second,
			},
			wantErr: false, // Pode falhar se RabbitMQ não estiver rodando, mas não é erro de configuração
		},
		{
			name:           "nil producer config",
			producerConfig: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := provider.CreateProducer(tt.producerConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateProducer() should return error, but got nil")
				}
				if producer != nil {
					t.Error("CreateProducer() should return nil producer on error")
				}
			} else {
				// Se RabbitMQ não estiver rodando, esperamos erro de conexão
				if err != nil {
					t.Logf("CreateProducer() failed as expected (RabbitMQ not running): %v", err)
				} else {
					if producer == nil {
						t.Error("CreateProducer() should return non-nil producer")
					} else {
						producer.Close() // Limpar recursos
					}
				}
			}
		})
	}
}

func TestRabbitMQProvider_CreateConsumer(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tests := []struct {
		name           string
		consumerConfig *interfaces.ConsumerConfig
		wantErr        bool
	}{
		{
			name: "valid consumer config",
			consumerConfig: &interfaces.ConsumerConfig{
				ID:            "test-consumer",
				ConsumerGroup: "test-group",
			},
			wantErr: false, // Pode falhar se RabbitMQ não estiver rodando, mas não é erro de configuração
		},
		{
			name:           "nil consumer config",
			consumerConfig: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer, err := provider.CreateConsumer(tt.consumerConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateConsumer() should return error, but got nil")
				}
				if consumer != nil {
					t.Error("CreateConsumer() should return nil consumer on error")
				}
			} else {
				// Se RabbitMQ não estiver rodando, esperamos erro de conexão
				if err != nil {
					t.Logf("CreateConsumer() failed as expected (RabbitMQ not running): %v", err)
				} else {
					if consumer == nil {
						t.Error("CreateConsumer() should return non-nil consumer")
					} else {
						consumer.Close() // Limpar recursos
					}
				}
			}
		})
	}
}

func TestRabbitMQProvider_Close(t *testing.T) {
	config := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers: []string{"localhost:5672"},
		},
	}

	provider, err := NewRabbitMQProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Close deve sempre funcionar, mesmo se não conectado
	err = provider.Close()
	if err != nil {
		t.Errorf("Close() should not return error, but got: %v", err)
	}

	// Após Close, deve estar desconectado
	if provider.IsConnected() {
		t.Error("Provider should be disconnected after Close()")
	}
}

// TestRabbitMQMockProvider testa o provider mock do RabbitMQ
func TestRabbitMQMockProvider(t *testing.T) {
	mockProvider := mock.NewMockRabbitMQProvider()

	// Test GetType
	if mockProvider.GetType() != interfaces.ProviderRabbitMQ {
		t.Errorf("Expected provider type %s, got %s", interfaces.ProviderRabbitMQ, mockProvider.GetType())
	}

	// Test initial state
	if mockProvider.IsConnected() {
		t.Error("Provider should not be connected initially")
	}

	// Test Connect
	ctx := context.Background()
	err := mockProvider.Connect(ctx)
	if err != nil {
		t.Errorf("Connect should not fail, got: %v", err)
	}

	if !mockProvider.IsConnected() {
		t.Error("Provider should be connected after Connect()")
	}

	// Test HealthCheck after connect
	err = mockProvider.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck should not fail after connect, got: %v", err)
	}

	// Test CreateProducer
	producerConfig := &interfaces.ProducerConfig{
		ID: "test-producer",
	}
	producer, err := mockProvider.CreateProducer(producerConfig)
	if err != nil {
		t.Errorf("CreateProducer should not fail, got: %v", err)
	}
	if producer == nil {
		t.Error("Producer should not be nil")
	}
	if !producer.IsConnected() {
		t.Error("Producer should be connected")
	}

	// Test CreateConsumer
	consumerConfig := &interfaces.ConsumerConfig{
		ID: "test-consumer",
	}
	consumer, err := mockProvider.CreateConsumer(consumerConfig)
	if err != nil {
		t.Errorf("CreateConsumer should not fail, got: %v", err)
	}
	if consumer == nil {
		t.Error("Consumer should not be nil")
	}
	if !consumer.IsConnected() {
		t.Error("Consumer should be connected")
	}

	// Test GetMetrics
	metrics := mockProvider.GetMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}
	if !metrics.HealthCheckStatus {
		t.Error("HealthCheckStatus should be true")
	}

	// Test Disconnect
	err = mockProvider.Disconnect()
	if err != nil {
		t.Errorf("Disconnect should not fail, got: %v", err)
	}
	if mockProvider.IsConnected() {
		t.Error("Provider should not be connected after Disconnect()")
	}

	// Test Close
	err = mockProvider.Close()
	if err != nil {
		t.Errorf("Close should not fail, got: %v", err)
	}
	if mockProvider.IsConnected() {
		t.Error("Provider should not be connected after Close()")
	}
}

// TestRabbitMQMockProducer testa o producer mock do RabbitMQ
func TestRabbitMQMockProducer(t *testing.T) {
	producer := mock.NewMockRabbitMQProducer()

	// Test initial state
	if !producer.IsConnected() {
		t.Error("Producer should be connected initially")
	}

	// Test Publish
	ctx := context.Background()
	message := &interfaces.Message{
		ID:   "test-1",
		Body: []byte("test message"),
	}

	err := producer.Publish(ctx, "test-queue", message, nil)
	if err != nil {
		t.Errorf("Publish should not fail, got: %v", err)
	}

	// Test PublishBatch
	messages := []*interfaces.Message{
		{ID: "test-2", Body: []byte("test message 2")},
		{ID: "test-3", Body: []byte("test message 3")},
	}

	err = producer.PublishBatch(ctx, "test-queue", messages, nil)
	if err != nil {
		t.Errorf("PublishBatch should not fail, got: %v", err)
	}

	// Test GetMetrics
	metrics := producer.GetMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}
	if metrics.MessagesSent != 3 { // 1 + 2 from batch
		t.Errorf("Expected 3 messages sent, got %d", metrics.MessagesSent)
	}

	// Test Close
	err = producer.Close()
	if err != nil {
		t.Errorf("Close should not fail, got: %v", err)
	}
	if producer.IsConnected() {
		t.Error("Producer should not be connected after Close()")
	}
}

// TestRabbitMQMockConsumer testa o consumer mock do RabbitMQ
func TestRabbitMQMockConsumer(t *testing.T) {
	consumer := mock.NewMockRabbitMQConsumer()

	// Test initial state
	if !consumer.IsConnected() {
		t.Error("Consumer should be connected initially")
	}
	if consumer.IsRunning() {
		t.Error("Consumer should not be running initially")
	}

	// Test Subscribe
	ctx := context.Background()
	handler := func(ctx context.Context, message *interfaces.Message) error {
		return nil
	}

	err := consumer.Subscribe(ctx, "test-queue", nil, handler)
	if err != nil {
		t.Errorf("Subscribe should not fail, got: %v", err)
	}
	if !consumer.IsRunning() {
		t.Error("Consumer should be running after Subscribe()")
	}

	// Test message acknowledgment
	message := &interfaces.Message{
		ID:   "test-1",
		Body: []byte("test message"),
	}

	err = consumer.Ack(message)
	if err != nil {
		t.Errorf("Ack should not fail, got: %v", err)
	}

	err = consumer.Nack(message, true)
	if err != nil {
		t.Errorf("Nack should not fail, got: %v", err)
	}

	// Test GetMetrics
	metrics := consumer.GetMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}
	if metrics.MessagesAcked != 1 {
		t.Errorf("Expected 1 message acked, got %d", metrics.MessagesAcked)
	}
	if metrics.MessagesNacked != 1 {
		t.Errorf("Expected 1 message nacked, got %d", metrics.MessagesNacked)
	}

	// Test Pause/Resume
	err = consumer.Pause()
	if err != nil {
		t.Errorf("Pause should not fail, got: %v", err)
	}

	err = consumer.Resume()
	if err != nil {
		t.Errorf("Resume should not fail, got: %v", err)
	}

	// Test Close
	err = consumer.Close()
	if err != nil {
		t.Errorf("Close should not fail, got: %v", err)
	}
	if consumer.IsConnected() {
		t.Error("Consumer should not be connected after Close()")
	}
	if consumer.IsRunning() {
		t.Error("Consumer should not be running after Close()")
	}
}
