package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockKafkaConsumer é um mock específico do consumer Kafka
type MockKafkaConsumer struct {
	SubscribeFunc      func(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error
	SubscribeBatchFunc func(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error
	AckFunc            func(message *interfaces.Message) error
	NackFunc           func(message *interfaces.Message, requeue bool) error
	CloseFunc          func() error
	PauseFunc          func() error
	ResumeFunc         func() error
	IsConnectedFunc    func() bool
	IsRunningFunc      func() bool
	GetMetricsFunc     func() *interfaces.ConsumerMetrics

	// Estado interno
	Connected     bool
	Closed        bool
	Running       bool
	Paused        bool
	Subscriptions []string
	MessagesProc  int64
	MessagesErr   int64
	MessagesAcked int64
	MessagesNack  int64
	BytesProc     int64
}

// NewMockKafkaConsumer cria uma nova instância do mock
func NewMockKafkaConsumer() *MockKafkaConsumer {
	return &MockKafkaConsumer{
		Connected:     true,
		Closed:        false,
		Running:       false,
		Paused:        false,
		Subscriptions: make([]string, 0),
	}
}

func (m *MockKafkaConsumer) Subscribe(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, source, options, handler)
	}
	if m.Closed {
		return fmt.Errorf("consumer is closed")
	}
	if !m.Connected {
		return fmt.Errorf("consumer not connected")
	}

	// Adiciona à lista de subscrições
	found := false
	for _, sub := range m.Subscriptions {
		if sub == source {
			found = true
			break
		}
	}
	if !found {
		m.Subscriptions = append(m.Subscriptions, source)
	}

	m.Running = true
	return nil
}

func (m *MockKafkaConsumer) SubscribeBatch(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error {
	if m.SubscribeBatchFunc != nil {
		return m.SubscribeBatchFunc(ctx, source, options, handler)
	}
	return m.Subscribe(ctx, source, options, func(ctx context.Context, message *interfaces.Message) error {
		return handler(ctx, []*interfaces.Message{message})
	})
}

func (m *MockKafkaConsumer) Ack(message *interfaces.Message) error {
	if m.AckFunc != nil {
		return m.AckFunc(message)
	}
	if m.Closed {
		return fmt.Errorf("consumer is closed")
	}

	m.MessagesAcked++
	return nil
}

func (m *MockKafkaConsumer) Nack(message *interfaces.Message, requeue bool) error {
	if m.NackFunc != nil {
		return m.NackFunc(message, requeue)
	}
	if m.Closed {
		return fmt.Errorf("consumer is closed")
	}

	m.MessagesNack++
	return nil
}

func (m *MockKafkaConsumer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.Closed = true
	m.Connected = false
	m.Running = false
	m.Subscriptions = nil
	return nil
}

func (m *MockKafkaConsumer) Pause() error {
	if m.PauseFunc != nil {
		return m.PauseFunc()
	}
	if m.Closed {
		return fmt.Errorf("consumer is closed")
	}
	m.Paused = true
	return nil
}

func (m *MockKafkaConsumer) Resume() error {
	if m.ResumeFunc != nil {
		return m.ResumeFunc()
	}
	if m.Closed {
		return fmt.Errorf("consumer is closed")
	}
	m.Paused = false
	return nil
}

func (m *MockKafkaConsumer) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return m.Connected && !m.Closed
}

func (m *MockKafkaConsumer) IsRunning() bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc()
	}
	return m.Running && !m.Closed && !m.Paused
}

func (m *MockKafkaConsumer) GetMetrics() *interfaces.ConsumerMetrics {
	if m.GetMetricsFunc != nil {
		return m.GetMetricsFunc()
	}
	return &interfaces.ConsumerMetrics{
		MessagesProcessed:    m.MessagesProc,
		MessagesError:        m.MessagesErr,
		MessagesAcked:        m.MessagesAcked,
		MessagesNacked:       m.MessagesNack,
		BytesProcessed:       m.BytesProc,
		AvgProcessingLatency: time.Millisecond * 5,
		LastProcessedAt:      time.Now(),
		MessagesPerSecond:    50.0,
		ActiveWorkers:        1,
	}
}
