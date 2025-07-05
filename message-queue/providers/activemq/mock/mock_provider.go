package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockActiveMQProvider é um mock específico do provider ActiveMQ
type MockActiveMQProvider struct {
	ConnectFunc       func(ctx context.Context) error
	DisconnectFunc    func() error
	CreateProducerFunc func(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error)
	CreateConsumerFunc func(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error)
	HealthCheckFunc   func(ctx context.Context) error
	CloseFunc         func() error
	GetTypeFunc       func() interfaces.ProviderType
	IsConnectedFunc   func() bool
	GetMetricsFunc    func() *interfaces.ProviderMetrics

	// Estado interno para simular comportamento
	Connected bool
	Closed    bool
	mutex     sync.RWMutex
}

// NewMockActiveMQProvider cria uma nova instância do mock
func NewMockActiveMQProvider() *MockActiveMQProvider {
	return &MockActiveMQProvider{
		Connected: false,
		Closed:    false,
	}
}

func (m *MockActiveMQProvider) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx)
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	m.Connected = true
	return nil
}

func (m *MockActiveMQProvider) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.Connected = false
	return nil
}

func (m *MockActiveMQProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if m.CreateProducerFunc != nil {
		return m.CreateProducerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to ActiveMQ")
	}
	return NewMockActiveMQProducer(), nil
}

func (m *MockActiveMQProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if m.CreateConsumerFunc != nil {
		return m.CreateConsumerFunc(config)
	}
	if m.Closed {
		return nil, fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return nil, fmt.Errorf("not connected to ActiveMQ")
	}
	return NewMockActiveMQConsumer(), nil
}

func (m *MockActiveMQProvider) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	if m.Closed {
		return fmt.Errorf("provider is closed")
	}
	if !m.Connected {
		return fmt.Errorf("not connected to ActiveMQ")
	}
	return nil
}

func (m *MockActiveMQProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.Closed = true
	m.Connected = false
	return nil
}

func (m *MockActiveMQProvider) GetType() interfaces.ProviderType {
	if m.GetTypeFunc != nil {
		return m.GetTypeFunc()
	}
	return interfaces.ProviderActiveMQ
}

func (m *MockActiveMQProvider) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Connected && !m.Closed
}

func (m *MockActiveMQProvider) GetMetrics() *interfaces.ProviderMetrics {
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
			AvgConnectionTime: time.Millisecond * 80,
			Reconnections:     0,
			LastConnectedAt:   time.Now(),
		},
	}
}
