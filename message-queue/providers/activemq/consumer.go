package activemq

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/go-stomp/stomp/v3"
)

// Consumer implementa interfaces.Consumer para Apache ActiveMQ
type Consumer struct {
	provider     *ActiveMQProvider
	connection   *stomp.Conn
	subscription *stomp.Subscription
	destination  string
	handler      interfaces.MessageHandler
	options      *interfaces.ConsumerOptions
	running      bool
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewConsumer cria um novo consumer para ActiveMQ
func NewConsumer(provider *ActiveMQProvider, destination string, handler interfaces.MessageHandler, options *interfaces.ConsumerOptions) (*Consumer, error) {
	if provider == nil || provider.conn == nil {
		return nil, domainerrors.New(
			"ACTIVEMQ_INVALID_CONNECTION",
			"invalid ActiveMQ connection",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	if handler == nil {
		return nil, domainerrors.New(
			"HANDLER_NIL",
			"message handler cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if options == nil {
		options = &interfaces.ConsumerOptions{
			Workers:           1,
			BufferSize:        100,
			ProcessingTimeout: 30 * time.Second,
			AutoAck:           false,
			BatchSize:         1,
		}
	}

	// Determina se é queue ou topic
	dest := destination
	if dest[0] != '/' {
		dest = "/queue/" + dest // Assume queue se não especificado
	}

	ctx, cancel := context.WithCancel(context.Background())
	if options.Context != nil {
		ctx, cancel = context.WithCancel(options.Context)
	}

	consumer := &Consumer{
		provider:    provider,
		connection:  provider.conn,
		destination: dest,
		handler:     handler,
		options:     options,
		running:     false,
		ctx:         ctx,
		cancel:      cancel,
	}

	return consumer, nil
}

// Subscribe inicia o consumo de mensagens
func (c *Consumer) Subscribe(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
	if c.running {
		return domainerrors.New(
			"CONSUMER_ALREADY_RUNNING",
			"consumer is already running",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Atualiza configurações se fornecidas
	if options != nil {
		c.options = options
	}
	if handler != nil {
		c.handler = handler
	}
	if source != "" {
		c.destination = source
		if source[0] != '/' {
			c.destination = "/queue/" + source
		}
	}

	// Subscreve ao destino
	subscription, err := c.connection.Subscribe(c.destination, stomp.AckClient)
	if err != nil {
		return domainerrors.New(
			"ACTIVEMQ_SUBSCRIBE_ERROR",
			"failed to subscribe to destination",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	c.subscription = subscription
	c.running = true

	// Inicia workers
	for i := 0; i < c.options.Workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	return nil
}

// SubscribeBatch inicia o consumo de mensagens em lote
func (c *Consumer) SubscribeBatch(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error {
	// ActiveMQ via STOMP não tem suporte nativo a batching
	// Implementação simples que processa uma por vez
	singleHandler := func(ctx context.Context, message *interfaces.Message) error {
		return handler(ctx, []*interfaces.Message{message})
	}

	return c.Subscribe(ctx, source, options, singleHandler)
}

// Ack confirma o processamento de uma mensagem
func (c *Consumer) Ack(message *interfaces.Message) error {
	// Implementação básica - em uma implementação real,
	// precisaríamos manter referência ao frame STOMP original
	return nil
}

// Nack rejeita uma mensagem
func (c *Consumer) Nack(message *interfaces.Message, requeue bool) error {
	// Implementação básica - em uma implementação real,
	// precisaríamos manter referência ao frame STOMP original
	return nil
}

// Close para o consumer e libera recursos
func (c *Consumer) Close() error {
	if !c.running {
		return nil
	}

	c.running = false
	c.cancel()

	// Cancela a subscription
	if c.subscription != nil {
		c.subscription.Unsubscribe()
	}

	// Aguarda workers terminarem
	c.wg.Wait()

	return nil
}

// Pause pausa temporariamente o consumo
func (c *Consumer) Pause() error {
	// Implementação simples - para o consumo
	return c.Close()
}

// Resume retoma o consumo após uma pausa
func (c *Consumer) Resume() error {
	// Implementação simples - reinicia o consumo
	return c.Subscribe(c.ctx, c.destination, c.options, c.handler)
}

// IsConnected verifica se o consumer está conectado
func (c *Consumer) IsConnected() bool {
	return c.connection != nil && c.running
}

// IsRunning verifica se o consumer está ativo
func (c *Consumer) IsRunning() bool {
	return c.running
}

// GetMetrics retorna métricas do consumer
func (c *Consumer) GetMetrics() *interfaces.ConsumerMetrics {
	return &interfaces.ConsumerMetrics{
		MessagesProcessed:    0, // TODO: implementar contador
		MessagesError:        0, // TODO: implementar contador
		MessagesAcked:        0, // TODO: implementar contador
		MessagesNacked:       0, // TODO: implementar contador
		BytesProcessed:       0, // TODO: implementar contador
		AvgProcessingLatency: 0,
		LastProcessedAt:      time.Time{},
		MessagesPerSecond:    0,
		ActiveWorkers:        c.options.Workers,
	}
}

// worker processa mensagens continuamente
func (c *Consumer) worker(workerID int) {
	defer c.wg.Done()

	for c.running {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.subscription.C:
			if msg.Err != nil {
				// Log do erro seria aqui
				continue
			}
			c.processMessage(msg, workerID)
		}
	}
}

// processMessage processa uma mensagem individual
func (c *Consumer) processMessage(stompMsg *stomp.Message, workerID int) {
	// Cria contexto com timeout
	ctx := c.ctx
	if c.options.ProcessingTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(c.ctx, c.options.ProcessingTimeout)
		defer cancel()
	}

	// Converte mensagem STOMP para Message
	message := &interfaces.Message{
		ID:        stompMsg.Header.Get("message-id"),
		Body:      stompMsg.Body,
		Headers:   make(map[string]interface{}),
		Timestamp: time.Now(), // STOMP pode não fornecer timestamp original
		Source:    c.destination,
		Context:   ctx,
	}

	// Converte headers STOMP
	if stompMsg.Header != nil {
		// Adiciona alguns headers importantes
		if msgID := stompMsg.Header.Get("message-id"); msgID != "" {
			message.Headers["message-id"] = msgID
		}
		if destination := stompMsg.Header.Get("destination"); destination != "" {
			message.Headers["destination"] = destination
		}
		if timestamp := stompMsg.Header.Get("timestamp"); timestamp != "" {
			message.Headers["timestamp"] = timestamp
		}
	}

	// Processa a mensagem
	ackType := interfaces.AckSuccess
	err := c.handler(ctx, message)
	if err != nil {
		ackType = interfaces.AckReject
		// Log do erro seria aqui
	}

	// Faz acknowledge da mensagem
	c.acknowledge(stompMsg, ackType)
}

// acknowledge faz o acknowledgment da mensagem
func (c *Consumer) acknowledge(stompMsg *stomp.Message, ackType interfaces.AckType) {
	if c.options.AutoAck {
		return // Já foi processada automaticamente
	}

	switch ackType {
	case interfaces.AckSuccess:
		stompMsg.Conn.Ack(stompMsg)
	case interfaces.AckReject:
		stompMsg.Conn.Nack(stompMsg)
	case interfaces.AckRequeue:
		// STOMP não tem conceito nativo de requeue
		// Podemos usar NACK para rejeitar e deixar o broker decidir
		stompMsg.Conn.Nack(stompMsg)
	}
}
