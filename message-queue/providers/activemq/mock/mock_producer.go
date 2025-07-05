package mock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockActiveMQProducer implementa um mock do ActiveMQ producer
type MockActiveMQProducer struct {
	id               string
	closed           bool
	sentMessages     []*interfaces.Message
	sentBatches      [][]*interfaces.Message
	failPublish      bool
	failPublishBatch bool
	failHealthCheck  bool
	publishDelay     time.Duration
	mutex            sync.RWMutex

	// Callbacks para testes
	OnPublish      func(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error
	OnPublishBatch func(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error
	OnClose        func() error
	OnHealthCheck  func(ctx context.Context) error
}

// NewMockActiveMQProducer cria um novo mock producer
func NewMockActiveMQProducer() *MockActiveMQProducer {
	return &MockActiveMQProducer{
		id:           "mock-producer",
		sentMessages: make([]*interfaces.Message, 0),
		sentBatches:  make([][]*interfaces.Message, 0),
	}
}

// GetID retorna o ID do producer
func (m *MockActiveMQProducer) GetID() string {
	return m.id
}

// Publish simula publicação de mensagem
func (m *MockActiveMQProducer) Publish(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error {
	if m.OnPublish != nil {
		return m.OnPublish(ctx, destination, message, options)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return errors.New("producer is closed")
	}

	if m.failPublish {
		return errors.New("mock publish failed")
	}

	if message == nil {
		return errors.New("message cannot be nil")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	// Simular delay se configurado
	if m.publishDelay > 0 {
		time.Sleep(m.publishDelay)
	}

	// Adicionar timestamp se não estiver presente
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Armazenar mensagem
	msgCopy := *message
	m.sentMessages = append(m.sentMessages, &msgCopy)

	return nil
}

// PublishBatch simula publicação em lote
func (m *MockActiveMQProducer) PublishBatch(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error {
	if m.OnPublishBatch != nil {
		return m.OnPublishBatch(ctx, destination, messages, options)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return errors.New("producer is closed")
	}

	if m.failPublishBatch {
		return errors.New("mock publish batch failed")
	}

	if len(messages) == 0 {
		return errors.New("messages batch cannot be empty")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	// Validar todas as mensagens
	for _, msg := range messages {
		if msg == nil {
			return errors.New("message in batch cannot be nil")
		}
	}

	// Simular delay proporcional ao tamanho do lote
	if m.publishDelay > 0 {
		time.Sleep(time.Duration(len(messages)) * m.publishDelay)
	}

	// Armazenar lote
	batchCopy := make([]*interfaces.Message, len(messages))
	for i, msg := range messages {
		if msg.Timestamp.IsZero() {
			msg.Timestamp = time.Now()
		}
		msgCopy := *msg
		batchCopy[i] = &msgCopy
		m.sentMessages = append(m.sentMessages, &msgCopy)
	}
	m.sentBatches = append(m.sentBatches, batchCopy)

	return nil
}

// HealthCheck simula health check do producer
func (m *MockActiveMQProducer) HealthCheck(ctx context.Context) error {
	if m.OnHealthCheck != nil {
		return m.OnHealthCheck(ctx)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.closed {
		return errors.New("producer is closed")
	}

	if m.failHealthCheck {
		return errors.New("producer health check failed")
	}

	return nil
}

// Close simula fechamento do producer
func (m *MockActiveMQProducer) Close() error {
	if m.OnClose != nil {
		return m.OnClose()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.closed = true
	return nil
}

// GetSentMessages retorna todas as mensagens enviadas
func (m *MockActiveMQProducer) GetSentMessages() []*interfaces.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages := make([]*interfaces.Message, len(m.sentMessages))
	copy(messages, m.sentMessages)
	return messages
}

// GetSentBatches retorna todos os lotes enviados
func (m *MockActiveMQProducer) GetSentBatches() [][]*interfaces.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	batches := make([][]*interfaces.Message, len(m.sentBatches))
	copy(batches, m.sentBatches)
	return batches
}

// GetSentMessageCount retorna o número total de mensagens enviadas
func (m *MockActiveMQProducer) GetSentMessageCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sentMessages)
}

// GetSentBatchCount retorna o número total de lotes enviados
func (m *MockActiveMQProducer) GetSentBatchCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sentBatches)
}

// SetFailPublish configura se Publish deve falhar
func (m *MockActiveMQProducer) SetFailPublish(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failPublish = fail
}

// SetFailPublishBatch configura se PublishBatch deve falhar
func (m *MockActiveMQProducer) SetFailPublishBatch(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failPublishBatch = fail
}

// SetFailHealthCheck configura se HealthCheck deve falhar
func (m *MockActiveMQProducer) SetFailHealthCheck(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failHealthCheck = fail
}

// SetPublishDelay configura delay para simulação
func (m *MockActiveMQProducer) SetPublishDelay(delay time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.publishDelay = delay
}

// IsClosed retorna se o producer está fechado
func (m *MockActiveMQProducer) IsClosed() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.closed
}

// Reset limpa todas as mensagens e lotes enviados
func (m *MockActiveMQProducer) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sentMessages = m.sentMessages[:0]
	m.sentBatches = m.sentBatches[:0]
	m.closed = false
	m.failPublish = false
	m.failPublishBatch = false
	m.failHealthCheck = false
	m.publishDelay = 0
}

// GetMetrics retorna métricas do producer
func (m *MockActiveMQProducer) GetMetrics() *interfaces.ProducerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &interfaces.ProducerMetrics{
		MessagesSent:      int64(len(m.sentMessages)),
		MessagesError:     0,
		BytesSent:         int64(len(m.sentMessages) * 100), // Mock: assume 100 bytes per message
		AvgLatency:        time.Millisecond * 5,
		LastSentAt:        time.Now(),
		MessagesPerSecond: 100.0,
	}
}

// IsConnected verifica se o producer está conectado
func (m *MockActiveMQProducer) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return !m.closed
}

// PublishWithCallback envia uma mensagem com callback de confirmação
func (m *MockActiveMQProducer) PublishWithCallback(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error {
	err := m.Publish(ctx, destination, message, options)
	if callback != nil {
		go callback(err)
	}
	return err
}
