package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockRabbitMQProducer é um mock específico do producer RabbitMQ
type MockRabbitMQProducer struct {
	PublishFunc             func(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error
	PublishBatchFunc        func(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error
	PublishWithCallbackFunc func(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error
	CloseFunc               func() error
	IsConnectedFunc         func() bool
	GetMetricsFunc          func() *interfaces.ProducerMetrics

	// Estado interno
	Connected    bool
	Closed       bool
	MessagesSent int64
	MessagesErr  int64
	BytesSent    int64
}

// NewMockRabbitMQProducer cria uma nova instância do mock
func NewMockRabbitMQProducer() *MockRabbitMQProducer {
	return &MockRabbitMQProducer{
		Connected: true,
		Closed:    false,
	}
}

func (m *MockRabbitMQProducer) Publish(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, destination, message, options)
	}
	if m.Closed {
		return fmt.Errorf("producer is closed")
	}
	if !m.Connected {
		return fmt.Errorf("producer not connected")
	}

	// Simula comportamento de publicação
	m.MessagesSent++
	if message != nil && message.Body != nil {
		m.BytesSent += int64(len(message.Body))
	}
	return nil
}

func (m *MockRabbitMQProducer) PublishBatch(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error {
	if m.PublishBatchFunc != nil {
		return m.PublishBatchFunc(ctx, destination, messages, options)
	}
	if m.Closed {
		return fmt.Errorf("producer is closed")
	}
	if !m.Connected {
		return fmt.Errorf("producer not connected")
	}

	// Simula comportamento de publicação em lote
	for _, msg := range messages {
		m.MessagesSent++
		if msg != nil && msg.Body != nil {
			m.BytesSent += int64(len(msg.Body))
		}
	}
	return nil
}

func (m *MockRabbitMQProducer) PublishWithCallback(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error {
	if m.PublishWithCallbackFunc != nil {
		return m.PublishWithCallbackFunc(ctx, destination, message, options, callback)
	}

	err := m.Publish(ctx, destination, message, options)
	if callback != nil {
		go callback(err)
	}
	return err
}

func (m *MockRabbitMQProducer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.Closed = true
	m.Connected = false
	return nil
}

func (m *MockRabbitMQProducer) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return m.Connected && !m.Closed
}

func (m *MockRabbitMQProducer) GetMetrics() *interfaces.ProducerMetrics {
	if m.GetMetricsFunc != nil {
		return m.GetMetricsFunc()
	}
	return &interfaces.ProducerMetrics{
		MessagesSent:      m.MessagesSent,
		MessagesError:     m.MessagesErr,
		BytesSent:         m.BytesSent,
		AvgLatency:        time.Millisecond * 15,
		LastSentAt:        time.Now(),
		MessagesPerSecond: 75.0,
	}
}
