package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockMessageConsumer implementa um mock do MessageConsumer
type MockMessageConsumer struct {
	connected bool
	running   bool
	metrics   *interfaces.ConsumerMetrics
	mutex     sync.RWMutex

	// Callbacks para testes
	OnSubscribe      func(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error
	OnSubscribeBatch func(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error
	OnAck            func(message *interfaces.Message) error
	OnNack           func(message *interfaces.Message, requeue bool) error
	OnClose          func() error
	OnPause          func() error
	OnResume         func() error
}

// NewMockMessageConsumer cria um novo mock consumer
func NewMockMessageConsumer() *MockMessageConsumer {
	return &MockMessageConsumer{
		connected: true,
		running:   false,
		metrics: &interfaces.ConsumerMetrics{
			MessagesProcessed:    0,
			MessagesError:        0,
			MessagesAcked:        0,
			MessagesNacked:       0,
			BytesProcessed:       0,
			AvgProcessingLatency: 0,
			LastProcessedAt:      time.Time{},
			MessagesPerSecond:    0,
			ActiveWorkers:        0,
		},
	}
}

// Subscribe simula início de consumo
func (m *MockMessageConsumer) Subscribe(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
	if m.OnSubscribe != nil {
		return m.OnSubscribe(ctx, source, options, handler)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running = true
	return nil
}

// SubscribeBatch simula início de consumo em lote
func (m *MockMessageConsumer) SubscribeBatch(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error {
	if m.OnSubscribeBatch != nil {
		return m.OnSubscribeBatch(ctx, source, options, handler)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running = true
	return nil
}

// Ack simula confirmação de mensagem
func (m *MockMessageConsumer) Ack(message *interfaces.Message) error {
	if m.OnAck != nil {
		return m.OnAck(message)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.MessagesAcked++
	return nil
}

// Nack simula rejeição de mensagem
func (m *MockMessageConsumer) Nack(message *interfaces.Message, requeue bool) error {
	if m.OnNack != nil {
		return m.OnNack(message, requeue)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.MessagesNacked++
	return nil
}

// Close simula fechamento
func (m *MockMessageConsumer) Close() error {
	if m.OnClose != nil {
		return m.OnClose()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connected = false
	m.running = false
	return nil
}

// Pause simula pausa
func (m *MockMessageConsumer) Pause() error {
	if m.OnPause != nil {
		return m.OnPause()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running = false
	return nil
}

// Resume simula retomada
func (m *MockMessageConsumer) Resume() error {
	if m.OnResume != nil {
		return m.OnResume()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running = true
	return nil
}

// IsConnected retorna status de conexão
func (m *MockMessageConsumer) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

// IsRunning retorna se está executando
func (m *MockMessageConsumer) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// GetMetrics retorna métricas mockadas
func (m *MockMessageConsumer) GetMetrics() *interfaces.ConsumerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &interfaces.ConsumerMetrics{
		MessagesProcessed:    m.metrics.MessagesProcessed,
		MessagesError:        m.metrics.MessagesError,
		MessagesAcked:        m.metrics.MessagesAcked,
		MessagesNacked:       m.metrics.MessagesNacked,
		BytesProcessed:       m.metrics.BytesProcessed,
		AvgProcessingLatency: m.metrics.AvgProcessingLatency,
		LastProcessedAt:      m.metrics.LastProcessedAt,
		MessagesPerSecond:    m.metrics.MessagesPerSecond,
		ActiveWorkers:        m.metrics.ActiveWorkers,
	}
}

// SetConnected define o status de conexão
func (m *MockMessageConsumer) SetConnected(connected bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.connected = connected
}

// SetRunning define o status de execução
func (m *MockMessageConsumer) SetRunning(running bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.running = running
}

// SimulateMessageProcessed simula processamento de mensagem
func (m *MockMessageConsumer) SimulateMessageProcessed(messageSize int, latency time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.MessagesProcessed++
	m.metrics.BytesProcessed += int64(messageSize)
	m.metrics.LastProcessedAt = time.Now()

	if m.metrics.AvgProcessingLatency == 0 {
		m.metrics.AvgProcessingLatency = latency
	} else {
		m.metrics.AvgProcessingLatency = (m.metrics.AvgProcessingLatency + latency) / 2
	}
}
