package mock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// MockActiveMQConsumer implementa um mock do ActiveMQ consumer
type MockActiveMQConsumer struct {
	id                string
	closed            bool
	subscribed        bool
	subscribedTo      string
	receivedMessages  []*interfaces.Message
	processedMessages []*interfaces.Message
	failSubscribe     bool
	failUnsubscribe   bool
	failHealthCheck   bool
	currentHandler    interfaces.MessageHandler
	mutex             sync.RWMutex

	// Callbacks para testes
	OnSubscribe   func(ctx context.Context, destination string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error
	OnUnsubscribe func() error
	OnClose       func() error
	OnHealthCheck func(ctx context.Context) error
}

// NewMockActiveMQConsumer cria um novo mock consumer
func NewMockActiveMQConsumer() *MockActiveMQConsumer {
	return &MockActiveMQConsumer{
		id:                "mock-consumer",
		receivedMessages:  make([]*interfaces.Message, 0),
		processedMessages: make([]*interfaces.Message, 0),
	}
}

// GetID retorna o ID do consumer
func (m *MockActiveMQConsumer) GetID() string {
	return m.id
}

// Subscribe simula subscrição a um destination
func (m *MockActiveMQConsumer) Subscribe(ctx context.Context, destination string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
	if m.OnSubscribe != nil {
		return m.OnSubscribe(ctx, destination, options, handler)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return errors.New("consumer is closed")
	}

	if m.failSubscribe {
		return errors.New("mock subscribe failed")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	m.subscribed = true
	m.subscribedTo = destination
	m.currentHandler = handler

	return nil
}

// Unsubscribe simula cancelamento da subscrição
func (m *MockActiveMQConsumer) Unsubscribe() error {
	if m.OnUnsubscribe != nil {
		return m.OnUnsubscribe()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return errors.New("consumer is closed")
	}

	if m.failUnsubscribe {
		return errors.New("mock unsubscribe failed")
	}

	m.subscribed = false
	m.subscribedTo = ""
	m.currentHandler = nil

	return nil
}

// HealthCheck simula health check do consumer
func (m *MockActiveMQConsumer) HealthCheck(ctx context.Context) error {
	if m.OnHealthCheck != nil {
		return m.OnHealthCheck(ctx)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.closed {
		return errors.New("consumer is closed")
	}

	if m.failHealthCheck {
		return errors.New("consumer health check failed")
	}

	return nil
}

// Close simula fechamento do consumer
func (m *MockActiveMQConsumer) Close() error {
	if m.OnClose != nil {
		return m.OnClose()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.subscribed {
		m.subscribed = false
		m.subscribedTo = ""
		m.currentHandler = nil
	}

	m.closed = true
	return nil
}

// SimulateMessage simula o recebimento de uma mensagem (para testes)
func (m *MockActiveMQConsumer) SimulateMessage(ctx context.Context, message *interfaces.Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed || !m.subscribed || m.currentHandler == nil {
		return errors.New("consumer not ready to receive messages")
	}

	if message == nil {
		return errors.New("message cannot be nil")
	}

	// Adicionar timestamp se não estiver presente
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Armazenar mensagem recebida
	msgCopy := *message
	m.receivedMessages = append(m.receivedMessages, &msgCopy)

	// Processar mensagem com o handler
	err := m.currentHandler(ctx, &msgCopy)
	if err == nil {
		m.processedMessages = append(m.processedMessages, &msgCopy)
	}

	return err
}

// SimulateMessages simula o recebimento de múltiplas mensagens
func (m *MockActiveMQConsumer) SimulateMessages(ctx context.Context, messages []*interfaces.Message) []error {
	var errors []error

	for _, msg := range messages {
		if err := m.SimulateMessage(ctx, msg); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// GetReceivedMessages retorna todas as mensagens recebidas
func (m *MockActiveMQConsumer) GetReceivedMessages() []*interfaces.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages := make([]*interfaces.Message, len(m.receivedMessages))
	copy(messages, m.receivedMessages)
	return messages
}

// GetProcessedMessages retorna todas as mensagens processadas com sucesso
func (m *MockActiveMQConsumer) GetProcessedMessages() []*interfaces.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages := make([]*interfaces.Message, len(m.processedMessages))
	copy(messages, m.processedMessages)
	return messages
}

// GetReceivedMessageCount retorna o número de mensagens recebidas
func (m *MockActiveMQConsumer) GetReceivedMessageCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.receivedMessages)
}

// GetProcessedMessageCount retorna o número de mensagens processadas
func (m *MockActiveMQConsumer) GetProcessedMessageCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.processedMessages)
}

// IsSubscribed retorna se está subscrito
func (m *MockActiveMQConsumer) IsSubscribed() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.subscribed
}

// GetSubscribedDestination retorna o destination subscrito
func (m *MockActiveMQConsumer) GetSubscribedDestination() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.subscribedTo
}

// SetFailSubscribe configura se Subscribe deve falhar
func (m *MockActiveMQConsumer) SetFailSubscribe(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failSubscribe = fail
}

// SetFailUnsubscribe configura se Unsubscribe deve falhar
func (m *MockActiveMQConsumer) SetFailUnsubscribe(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failUnsubscribe = fail
}

// SetFailHealthCheck configura se HealthCheck deve falhar
func (m *MockActiveMQConsumer) SetFailHealthCheck(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failHealthCheck = fail
}

// IsClosed retorna se o consumer está fechado
func (m *MockActiveMQConsumer) IsClosed() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.closed
}

// Reset limpa todas as mensagens recebidas/processadas
func (m *MockActiveMQConsumer) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.receivedMessages = m.receivedMessages[:0]
	m.processedMessages = m.processedMessages[:0]
	m.subscribed = false
	m.subscribedTo = ""
	m.currentHandler = nil
	m.closed = false
	m.failSubscribe = false
	m.failUnsubscribe = false
	m.failHealthCheck = false
}

// SubscribeBatch inicia o consumo de mensagens em lote
func (m *MockActiveMQConsumer) SubscribeBatch(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error {
	// Para simplicidade, convertemos para Subscribe individual
	return m.Subscribe(ctx, source, options, func(ctx context.Context, message *interfaces.Message) error {
		return handler(ctx, []*interfaces.Message{message})
	})
}

// Ack confirma o processamento de uma mensagem
func (m *MockActiveMQConsumer) Ack(message *interfaces.Message) error {
	if m.closed {
		return fmt.Errorf("consumer is closed")
	}
	// Simula ack bem-sucedido
	return nil
}

// Nack rejeita uma mensagem
func (m *MockActiveMQConsumer) Nack(message *interfaces.Message, requeue bool) error {
	if m.closed {
		return fmt.Errorf("consumer is closed")
	}
	// Simula nack bem-sucedido
	return nil
}

// Pause pausa temporariamente o consumo
func (m *MockActiveMQConsumer) Pause() error {
	if m.closed {
		return fmt.Errorf("consumer is closed")
	}
	// Simula pause bem-sucedido
	return nil
}

// Resume retoma o consumo após uma pausa
func (m *MockActiveMQConsumer) Resume() error {
	if m.closed {
		return fmt.Errorf("consumer is closed")
	}
	// Simula resume bem-sucedido
	return nil
}

// IsConnected verifica se o consumer está conectado
func (m *MockActiveMQConsumer) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return !m.closed
}

// IsRunning verifica se o consumer está ativo
func (m *MockActiveMQConsumer) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.subscribed && !m.closed
}

// GetMetrics retorna métricas do consumer
func (m *MockActiveMQConsumer) GetMetrics() *interfaces.ConsumerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &interfaces.ConsumerMetrics{
		MessagesProcessed:    int64(len(m.processedMessages)),
		MessagesError:        0,
		MessagesAcked:        int64(len(m.processedMessages)),
		MessagesNacked:       0,
		BytesProcessed:       int64(len(m.processedMessages) * 100), // Mock: assume 100 bytes per message
		AvgProcessingLatency: time.Millisecond * 10,
		LastProcessedAt:      time.Now(),
		MessagesPerSecond:    50.0,
		ActiveWorkers:        1,
	}
}
