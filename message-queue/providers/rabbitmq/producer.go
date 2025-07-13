package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer implementa interfaces.Producer para RabbitMQ
type Producer struct {
	provider   *RabbitMQProvider
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

// NewProducer cria um novo producer para RabbitMQ
func NewProducer(provider *RabbitMQProvider, exchange, routingKey string) (*Producer, error) {
	if provider == nil || provider.connection == nil {
		return nil, domainerrors.New(
			"RABBITMQ_INVALID_CONNECTION",
			"invalid connection",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	channel, err := provider.connection.Channel()
	if err != nil {
		return nil, domainerrors.New(
			"RABBITMQ_CHANNEL_ERROR",
			fmt.Sprintf("failed to open channel: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Declara o exchange se necessário
	if exchange != "" {
		err = channel.ExchangeDeclare(
			exchange,
			"direct", // tipo
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			channel.Close()
			return nil, domainerrors.New(
				"RABBITMQ_EXCHANGE_ERROR",
				fmt.Sprintf("failed to declare exchange: %v", err),
			).WithType(domainerrors.ErrorTypeInternal)
		}
	}

	return &Producer{
		provider:   provider,
		channel:    channel,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

// Send envia uma mensagem
func (p *Producer) Send(ctx context.Context, message *interfaces.Message) error {
	if message == nil {
		return domainerrors.New(
			"MESSAGE_NIL",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Usa o Body da mensagem diretamente
	body := message.Body
	if len(body) == 0 {
		return domainerrors.New(
			"MESSAGE_BODY_EMPTY",
			"message body cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Prepara headers AMQP
	headers := amqp.Table{}
	for k, v := range message.Headers {
		headers[k] = v
	}

	// Configura delivery mode baseado na persistência (padrão: persistente)
	deliveryMode := uint8(2) // persistent

	// Cria a mensagem AMQP
	publishing := amqp.Publishing{
		Headers:         headers,
		ContentType:     "application/json",
		ContentEncoding: "",
		Body:            body,
		DeliveryMode:    deliveryMode,
		Priority:        0,
		Timestamp:       time.Now(),
		MessageId:       message.ID,
	}

	// Publica a mensagem
	err := p.channel.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)
	if err != nil {
		return domainerrors.New(
			"PUBLISH_ERROR",
			fmt.Sprintf("failed to publish message: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// SendBatch envia múltiplas mensagens em lote
func (p *Producer) SendBatch(ctx context.Context, messages []*interfaces.Message) error {
	if len(messages) == 0 {
		return nil
	}

	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}

	return nil
}

// Close fecha o producer
func (p *Producer) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}

// GetMetrics retorna métricas do producer
func (p *Producer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if p.channel != nil && !p.channel.IsClosed() {
		metrics["channel_open"] = true
		metrics["exchange"] = p.exchange
		metrics["routing_key"] = p.routingKey
	} else {
		metrics["channel_open"] = false
	}

	return metrics
}

// IsHealthy verifica se o producer está saudável
func (p *Producer) IsHealthy() bool {
	return p.channel != nil && !p.channel.IsClosed()
}
