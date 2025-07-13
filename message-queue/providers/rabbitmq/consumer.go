package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer implementa interfaces.Consumer para RabbitMQ
type Consumer struct {
	provider    *RabbitMQProvider
	channel     *amqp.Channel
	queue       string
	consumerTag string
	handler     interfaces.MessageHandler
	options     *interfaces.ConsumerOptions
	deliveries  <-chan amqp.Delivery
	done        chan bool
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewConsumer cria um novo consumer para RabbitMQ
func NewConsumer(provider *RabbitMQProvider, queue string, handler interfaces.MessageHandler, options *interfaces.ConsumerOptions) (*Consumer, error) {
	if provider == nil || provider.connection == nil {
		return nil, domainerrors.New(
			"RABBITMQ_INVALID_CONNECTION",
			"invalid connection",
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

	channel, err := provider.connection.Channel()
	if err != nil {
		return nil, domainerrors.New(
			"RABBITMQ_CHANNEL_ERROR",
			fmt.Sprintf("failed to open channel: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Declara a queue se necessário
	_, err = channel.QueueDeclare(
		queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		return nil, domainerrors.New(
			"RABBITMQ_QUEUE_ERROR",
			fmt.Sprintf("failed to declare queue: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Configura QoS
	err = channel.Qos(
		options.BufferSize, // prefetch count
		0,                  // prefetch size
		false,              // global
	)
	if err != nil {
		channel.Close()
		return nil, domainerrors.New(
			"RABBITMQ_QOS_ERROR",
			fmt.Sprintf("failed to set QoS: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if options.Context != nil {
		ctx, cancel = context.WithCancel(options.Context)
	}

	consumer := &Consumer{
		provider:    provider,
		channel:     channel,
		queue:       queue,
		consumerTag: fmt.Sprintf("consumer-%d", time.Now().UnixNano()),
		handler:     handler,
		options:     options,
		done:        make(chan bool),
		ctx:         ctx,
		cancel:      cancel,
	}

	return consumer, nil
}

// Start inicia o consumo de mensagens
func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.channel.Consume(
		c.queue,
		c.consumerTag,
		c.options.AutoAck,
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return domainerrors.New(
			"RABBITMQ_CONSUME_ERROR",
			fmt.Sprintf("failed to start consuming: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	c.deliveries = deliveries

	// Inicia workers
	for i := 0; i < c.options.Workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	return nil
}

// Stop para o consumo de mensagens
func (c *Consumer) Stop() error {
	// Cancela o consumer
	if err := c.channel.Cancel(c.consumerTag, false); err != nil {
		return domainerrors.New(
			"RABBITMQ_CANCEL_ERROR",
			fmt.Sprintf("failed to cancel consumer: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Sinaliza para parar
	c.cancel()
	close(c.done)

	// Aguarda workers terminarem
	c.wg.Wait()

	// Fecha o canal
	if err := c.channel.Close(); err != nil {
		return domainerrors.New(
			"RABBITMQ_CLOSE_ERROR",
			fmt.Sprintf("failed to close channel: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// worker processa mensagens
func (c *Consumer) worker(workerID int) {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			return
		case <-c.ctx.Done():
			return
		case delivery, ok := <-c.deliveries:
			if !ok {
				return
			}
			c.processMessage(delivery, workerID)
		}
	}
}

// processMessage processa uma mensagem individual
func (c *Consumer) processMessage(delivery amqp.Delivery, workerID int) {
	// Cria contexto com timeout
	ctx := c.ctx
	if c.options.ProcessingTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(c.ctx, c.options.ProcessingTimeout)
		defer cancel()
	}

	// Converte delivery para Message
	message := &interfaces.Message{
		ID:        delivery.MessageId,
		Body:      delivery.Body,
		Headers:   make(map[string]interface{}),
		Timestamp: delivery.Timestamp,
		Source:    c.queue,
		Context:   ctx,
	}

	// Converte headers
	for k, v := range delivery.Headers {
		message.Headers[k] = v
	}

	// Processa a mensagem
	ackType := interfaces.AckSuccess
	err := c.handler(ctx, message)
	if err != nil {
		ackType = interfaces.AckReject
		// Log do erro seria aqui
	}

	// Faz acknowledge da mensagem
	c.acknowledge(delivery, ackType)
}

// acknowledge faz o acknowledgment da mensagem
func (c *Consumer) acknowledge(delivery amqp.Delivery, ackType interfaces.AckType) {
	if c.options.AutoAck {
		return // Já foi feito automaticamente
	}

	switch ackType {
	case interfaces.AckSuccess:
		delivery.Ack(false)
	case interfaces.AckReject:
		delivery.Nack(false, false) // não requeue
	case interfaces.AckRequeue:
		delivery.Nack(false, true) // requeue
	}
}

// GetMetrics retorna métricas do consumer
func (c *Consumer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if c.channel != nil && !c.channel.IsClosed() {
		metrics["channel_open"] = true
		metrics["queue"] = c.queue
		metrics["consumer_tag"] = c.consumerTag
		metrics["workers"] = c.options.Workers
		metrics["auto_ack"] = c.options.AutoAck
	} else {
		metrics["channel_open"] = false
	}

	return metrics
}

// IsHealthy verifica se o consumer está saudável
func (c *Consumer) IsHealthy() bool {
	return c.channel != nil && !c.channel.IsClosed()
}
