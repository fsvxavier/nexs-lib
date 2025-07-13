package activemq

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
)

// Producer implementa interfaces.Producer para Apache ActiveMQ
type Producer struct {
	provider    *ActiveMQProvider
	connection  *stomp.Conn
	destination string
}

// NewProducer cria um novo producer para ActiveMQ
func NewProducer(provider *ActiveMQProvider, destination string) (*Producer, error) {
	if provider == nil || provider.conn == nil {
		return nil, domainerrors.New(
			"ACTIVEMQ_INVALID_CONNECTION",
			"invalid ActiveMQ connection",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	if destination == "" {
		return nil, domainerrors.New(
			"ACTIVEMQ_INVALID_DESTINATION",
			"destination cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	return &Producer{
		provider:    provider,
		connection:  provider.conn,
		destination: destination,
	}, nil
}

// Publish envia uma mensagem para a fila/tópico especificado
func (p *Producer) Publish(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error {
	// Se destination foi especificado, usa ele; senão usa o padrão do producer
	dest := destination
	if dest == "" {
		dest = p.destination
	}

	// Atualiza o destination temporariamente
	originalDestination := p.destination
	p.destination = dest
	defer func() { p.destination = originalDestination }()

	return p.Send(ctx, message)
}

// PublishBatch envia múltiplas mensagens em lote
func (p *Producer) PublishBatch(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error {
	// Se destination foi especificado, usa ele; senão usa o padrão do producer
	dest := destination
	if dest == "" {
		dest = p.destination
	}

	// Atualiza o destination temporariamente
	originalDestination := p.destination
	p.destination = dest
	defer func() { p.destination = originalDestination }()

	return p.SendBatch(ctx, messages)
}

// PublishWithCallback envia uma mensagem com callback de confirmação
func (p *Producer) PublishWithCallback(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error {
	err := p.Publish(ctx, destination, message, options)
	if callback != nil {
		callback(err)
	}
	return err
}

// Send envia uma mensagem para o ActiveMQ
func (p *Producer) Send(ctx context.Context, message *interfaces.Message) error {
	if message == nil {
		return domainerrors.New(
			"MESSAGE_NIL",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if len(message.Body) == 0 {
		return domainerrors.New(
			"MESSAGE_BODY_EMPTY",
			"message body cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Prepara headers STOMP
	var headers []func(*frame.Frame) error

	// Adiciona headers customizados
	for k, v := range message.Headers {
		if str, ok := v.(string); ok {
			headers = append(headers, stomp.SendOpt.Header(k, str))
		} else {
			// Converte outros tipos para string
			headers = append(headers, stomp.SendOpt.Header(k, fmt.Sprintf("%v", v)))
		}
	}

	// Adiciona Message ID se fornecido
	if message.ID != "" {
		headers = append(headers, stomp.SendOpt.Header("message-id", message.ID))
	}

	// Adiciona timestamp
	headers = append(headers, stomp.SendOpt.Header("timestamp", fmt.Sprintf("%d", time.Now().Unix())))

	// Configura persistência (padrão: persistente)
	headers = append(headers, stomp.SendOpt.Header("persistent", "true"))

	// Determina se é queue ou topic
	dest := p.destination
	if dest[0] != '/' {
		dest = "/queue/" + dest // Assume queue se não especificado
	}

	// Envia a mensagem
	err := p.connection.Send(dest, "text/plain", message.Body, headers...)
	if err != nil {
		return domainerrors.New(
			"ACTIVEMQ_SEND_ERROR",
			fmt.Sprintf("failed to send message to ActiveMQ: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// SendBatch envia múltiplas mensagens em lote
func (p *Producer) SendBatch(ctx context.Context, messages []*interfaces.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// ActiveMQ via STOMP não tem suporte nativo a batching
	// Enviamos uma por vez
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}

	return nil
}

// Close fecha o producer
func (p *Producer) Close() error {
	// O connection é gerenciado pelo provider
	return nil
}

// GetMetrics retorna métricas do producer
func (p *Producer) GetMetrics() *interfaces.ProducerMetrics {
	return &interfaces.ProducerMetrics{
		MessagesSent:      0, // TODO: implementar contador
		MessagesError:     0, // TODO: implementar contador
		BytesSent:         0, // TODO: implementar contador
		AvgLatency:        0,
		LastSentAt:        time.Time{},
		MessagesPerSecond: 0,
	}
}

// IsConnected verifica se o producer está conectado
func (p *Producer) IsConnected() bool {
	return p.connection != nil
}

// IsHealthy verifica se o producer está saudável
func (p *Producer) IsHealthy() bool {
	return p.connection != nil
}
