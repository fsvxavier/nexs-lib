package interfaces

import (
	"context"
	"time"
)

// MessageProducer define a interface para produtores de mensagens
type MessageProducer interface {
	// Publish envia uma mensagem para a fila/tópico especificado
	Publish(ctx context.Context, destination string, message *Message, options *MessageOptions) error

	// PublishBatch envia múltiplas mensagens em lote
	PublishBatch(ctx context.Context, destination string, messages []*Message, options *MessageOptions) error

	// PublishWithCallback envia uma mensagem com callback de confirmação
	PublishWithCallback(ctx context.Context, destination string, message *Message, options *MessageOptions, callback func(error)) error

	// Close fecha o producer e libera recursos
	Close() error

	// IsConnected verifica se o producer está conectado
	IsConnected() bool

	// GetMetrics retorna métricas do producer
	GetMetrics() *ProducerMetrics
}

// ProducerMetrics representa métricas do producer
type ProducerMetrics struct {
	// Número total de mensagens enviadas
	MessagesSent int64

	// Número total de mensagens com erro
	MessagesError int64

	// Número total de bytes enviados
	BytesSent int64

	// Latência média de envio
	AvgLatency time.Duration

	// Última vez que uma mensagem foi enviada
	LastSentAt time.Time

	// Taxa de mensagens por segundo
	MessagesPerSecond float64
}
