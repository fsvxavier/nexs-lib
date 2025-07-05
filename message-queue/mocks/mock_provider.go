package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockMessageQueueProvider implementa um mock do MessageQueueProvider
type MockMessageQueueProvider struct {
	providerType      interfaces.ProviderType
	connected         bool
	healthCheckResult error
	metrics           *interfaces.ProviderMetrics
	mutex             sync.RWMutex

	// Callbacks para testes
	OnConnect        func(ctx context.Context) error
	OnDisconnect     func() error
	OnCreateProducer func(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error)
	OnCreateConsumer func(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error)
	OnHealthCheck    func(ctx context.Context) error
}

// NewMockMessageQueueProvider cria um novo mock provider
func NewMockMessageQueueProvider(providerType interfaces.ProviderType) *MockMessageQueueProvider {
	return &MockMessageQueueProvider{
		providerType: providerType,
		connected:    false,
		metrics: &interfaces.ProviderMetrics{
			Uptime:            0,
			ActiveConnections: 0,
			ActiveProducers:   0,
			ActiveConsumers:   0,
			LastHealthCheck:   time.Now(),
			HealthCheckStatus: true,
			ConnectionStats: &interfaces.ConnectionStats{
				TotalConnections:  0,
				FailedConnections: 0,
				AvgConnectionTime: 0,
				Reconnections:     0,
				LastConnectedAt:   time.Time{},
			},
		},
	}
}

// GetType retorna o tipo do provider
func (m *MockMessageQueueProvider) GetType() interfaces.ProviderType {
	return m.providerType
}

// Connect simula conexão
func (m *MockMessageQueueProvider) Connect(ctx context.Context) error {
	if m.OnConnect != nil {
		return m.OnConnect(ctx)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connected = true
	m.metrics.ActiveConnections = 1
	m.metrics.ConnectionStats.TotalConnections++
	m.metrics.ConnectionStats.LastConnectedAt = time.Now()

	return nil
}

// Disconnect simula desconexão
func (m *MockMessageQueueProvider) Disconnect() error {
	if m.OnDisconnect != nil {
		return m.OnDisconnect()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connected = false
	m.metrics.ActiveConnections = 0

	return nil
}

// CreateProducer cria um mock producer
func (m *MockMessageQueueProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if m.OnCreateProducer != nil {
		return m.OnCreateProducer(config)
	}

	m.mutex.Lock()
	m.metrics.ActiveProducers++
	m.mutex.Unlock()

	return NewMockMessageProducer(), nil
}

// CreateConsumer cria um mock consumer
func (m *MockMessageQueueProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if m.OnCreateConsumer != nil {
		return m.OnCreateConsumer(config)
	}

	m.mutex.Lock()
	m.metrics.ActiveConsumers++
	m.mutex.Unlock()

	return NewMockMessageConsumer(), nil
}

// IsConnected retorna o status de conexão
func (m *MockMessageQueueProvider) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

// HealthCheck simula verificação de saúde
func (m *MockMessageQueueProvider) HealthCheck(ctx context.Context) error {
	if m.OnHealthCheck != nil {
		return m.OnHealthCheck(ctx)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.LastHealthCheck = time.Now()
	m.metrics.HealthCheckStatus = m.healthCheckResult == nil

	return m.healthCheckResult
}

// GetMetrics retorna métricas mockadas
func (m *MockMessageQueueProvider) GetMetrics() *interfaces.ProviderMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Cria uma cópia das métricas
	return &interfaces.ProviderMetrics{
		Uptime:            m.metrics.Uptime,
		ActiveConnections: m.metrics.ActiveConnections,
		ActiveProducers:   m.metrics.ActiveProducers,
		ActiveConsumers:   m.metrics.ActiveConsumers,
		LastHealthCheck:   m.metrics.LastHealthCheck,
		HealthCheckStatus: m.metrics.HealthCheckStatus,
		ConnectionStats: &interfaces.ConnectionStats{
			TotalConnections:  m.metrics.ConnectionStats.TotalConnections,
			FailedConnections: m.metrics.ConnectionStats.FailedConnections,
			AvgConnectionTime: m.metrics.ConnectionStats.AvgConnectionTime,
			Reconnections:     m.metrics.ConnectionStats.Reconnections,
			LastConnectedAt:   m.metrics.ConnectionStats.LastConnectedAt,
		},
	}
}

// Close simula fechamento
func (m *MockMessageQueueProvider) Close() error {
	return m.Disconnect()
}

// SetHealthCheckResult define o resultado do health check
func (m *MockMessageQueueProvider) SetHealthCheckResult(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.healthCheckResult = err
}
