package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockMessageProducer implementa um mock do MessageProducer
type MockMessageProducer struct {
	connected bool
	metrics   *interfaces.ProducerMetrics
	mutex     sync.RWMutex

	// Callbacks para testes
	OnPublish             func(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error
	OnPublishBatch        func(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error
	OnPublishWithCallback func(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error
	OnClose               func() error
}

// NewMockMessageProducer cria um novo mock producer
func NewMockMessageProducer() *MockMessageProducer {
	return &MockMessageProducer{
		connected: true,
		metrics: &interfaces.ProducerMetrics{
			MessagesSent:      0,
			MessagesError:     0,
			BytesSent:         0,
			AvgLatency:        0,
			LastSentAt:        time.Time{},
			MessagesPerSecond: 0,
		},
	}
}

// Publish simula envio de mensagem
func (m *MockMessageProducer) Publish(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error {
	if m.OnPublish != nil {
		return m.OnPublish(ctx, destination, message, options)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics.MessagesSent++
	m.metrics.BytesSent += int64(len(message.Body))
	m.metrics.LastSentAt = time.Now()

	return nil
}

// PublishBatch simula envio de mensagens em lote
func (m *MockMessageProducer) PublishBatch(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error {
	if m.OnPublishBatch != nil {
		return m.OnPublishBatch(ctx, destination, messages, options)
	}

	for _, message := range messages {
		if err := m.Publish(ctx, destination, message, options); err != nil {
			return err
		}
	}

	return nil
}

// PublishWithCallback simula envio com callback
func (m *MockMessageProducer) PublishWithCallback(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error {
	if m.OnPublishWithCallback != nil {
		return m.OnPublishWithCallback(ctx, destination, message, options, callback)
	}

	err := m.Publish(ctx, destination, message, options)
	if callback != nil {
		callback(err)
	}

	return err
}

// Close simula fechamento
func (m *MockMessageProducer) Close() error {
	if m.OnClose != nil {
		return m.OnClose()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connected = false
	return nil
}

// IsConnected retorna status de conexão
func (m *MockMessageProducer) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

// GetMetrics retorna métricas mockadas
func (m *MockMessageProducer) GetMetrics() *interfaces.ProducerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &interfaces.ProducerMetrics{
		MessagesSent:      m.metrics.MessagesSent,
		MessagesError:     m.metrics.MessagesError,
		BytesSent:         m.metrics.BytesSent,
		AvgLatency:        m.metrics.AvgLatency,
		LastSentAt:        m.metrics.LastSentAt,
		MessagesPerSecond: m.metrics.MessagesPerSecond,
	}
}

// SetConnected define o status de conexão
func (m *MockMessageProducer) SetConnected(connected bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.connected = connected
}
