package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockKafkaProvider é um mock específico do provider Kafka
type MockKafkaProvider struct {
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

// NewMockKafkaProvider cria uma nova instância do mock
func NewMockKafkaProvider() *MockKafkaProvider {
	return &MockKafkaProvider{
		Connected: false,
		Closed:    false,
	}
}

func (m *MockKafkaProvider) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx)
	}
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	m.Connected = true
	return nil
}

func (m *MockKafkaProvider) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	m.Connected = false
	return nil
}

func (m *MockKafkaProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if m.CreateProducerFunc != nil {
		return m.CreateProducerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to Kafka")
	}
	return NewMockKafkaProducer(), nil
}

func (m *MockKafkaProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if m.CreateConsumerFunc != nil {
		return m.CreateConsumerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to Kafka")
	}
	return NewMockKafkaConsumer(), nil
}

func (m *MockKafkaProvider) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return fmt.Errorf("not connected to Kafka")
	}
	return nil
}

func (m *MockKafkaProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.Closed = true
	m.Connected = false
	return nil
}

func (m *MockKafkaProvider) GetType() interfaces.ProviderType {
	if m.GetTypeFunc != nil {
		return m.GetTypeFunc()
	}
	return interfaces.ProviderKafka
}

func (m *MockKafkaProvider) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return m.Connected && !m.Closed
}

func (m *MockKafkaProvider) GetMetrics() *interfaces.ProviderMetrics {
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
