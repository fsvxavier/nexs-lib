package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockRabbitMQProvider é um mock específico do provider RabbitMQ
type MockRabbitMQProvider struct {
	ConnectFunc        func(ctx context.Context) error
	DisconnectFunc     func() error
	CreateProducerFunc func(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error)
	CreateConsumerFunc func(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error)
	HealthCheckFunc    func(ctx context.Context) error
	CloseFunc          func() error
	GetTypeFunc        func() interfaces.ProviderType
	IsConnectedFunc    func() bool
	GetMetricsFunc     func() *interfaces.ProviderMetrics

	// Estado interno para simular comportamento
	Connected bool
	Closed    bool
}

// NewMockRabbitMQProvider cria uma nova instância do mock
func NewMockRabbitMQProvider() *MockRabbitMQProvider {
	return &MockRabbitMQProvider{
		Connected: false,
		Closed:    false,
	}
}

func (m *MockRabbitMQProvider) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx)
	}
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	m.Connected = true
	return nil
}

func (m *MockRabbitMQProvider) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	m.Connected = false
	return nil
}

func (m *MockRabbitMQProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if m.CreateProducerFunc != nil {
		return m.CreateProducerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}
	return NewMockRabbitMQProducer(), nil
}

func (m *MockRabbitMQProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if m.CreateConsumerFunc != nil {
		return m.CreateConsumerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}
	return NewMockRabbitMQConsumer(), nil
}

func (m *MockRabbitMQProvider) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return fmt.Errorf("not connected to RabbitMQ")
	}
	return nil
}

func (m *MockRabbitMQProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.Closed = true
	m.Connected = false
	return nil
}

func (m *MockRabbitMQProvider) GetType() interfaces.ProviderType {
	if m.GetTypeFunc != nil {
		return m.GetTypeFunc()
	}
	return interfaces.ProviderRabbitMQ
}

func (m *MockRabbitMQProvider) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return m.Connected && !m.Closed
}

func (m *MockRabbitMQProvider) GetMetrics() *interfaces.ProviderMetrics {
	if m.GetMetricsFunc != nil {
		return m.GetMetricsFunc()
	}
	return &interfaces.ProviderMetrics{
		Uptime:            time.Hour,
		ActiveConnections: 1,
		ActiveProducers:   0,
		ActiveConsumers:   0,
		LastHealthCheck:   time.Now(),
		HealthCheckStatus: m.Connected,
		ConnectionStats: &interfaces.ConnectionStats{
			TotalConnections:  1,
			FailedConnections: 0,
			AvgConnectionTime: time.Millisecond * 100,
			Reconnections:     0,
			LastConnectedAt:   time.Now(),
		},
	}
}
