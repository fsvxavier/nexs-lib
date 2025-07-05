package interfaces

import (
	"context"
	"time"
)

// MessageConsumer define a interface para consumidores de mensagens
type MessageConsumer interface {
	// Subscribe inicia o consumo de mensagens de uma fila/tópico
	Subscribe(ctx context.Context, source string, options *ConsumerOptions, handler MessageHandler) error

	// SubscribeBatch inicia o consumo de mensagens em lote
	SubscribeBatch(ctx context.Context, source string, options *ConsumerOptions, handler BatchMessageHandler) error

	// Ack confirma o processamento de uma mensagem
	Ack(message *Message) error

	// Nack rejeita uma mensagem (pode ser reprocessada ou enviada para DLQ)
	Nack(message *Message, requeue bool) error

	// Close para o consumer e libera recursos
	Close() error

	// Pause pausa temporariamente o consumo
	Pause() error

	// Resume retoma o consumo após uma pausa
	Resume() error

	// IsConnected verifica se o consumer está conectado
	IsConnected() bool

	// IsRunning verifica se o consumer está ativo
	IsRunning() bool

	// GetMetrics retorna métricas do consumer
	GetMetrics() *ConsumerMetrics
}

// MessageHandler define a função para processar mensagens individuais
type MessageHandler func(ctx context.Context, message *Message) error

// BatchMessageHandler define a função para processar mensagens em lote
type BatchMessageHandler func(ctx context.Context, messages []*Message) error

// ConsumerMetrics representa métricas do consumer
type ConsumerMetrics struct {
	// Número total de mensagens processadas
	MessagesProcessed int64

	// Número total de mensagens com erro
	MessagesError int64

	// Número total de mensagens confirmadas (ACK)
	MessagesAcked int64

	// Número total de mensagens rejeitadas (NACK)
	MessagesNacked int64

	// Número total de bytes processados
	BytesProcessed int64

	// Latência média de processamento
	AvgProcessingLatency time.Duration

	// Última vez que uma mensagem foi processada
	LastProcessedAt time.Time

	// Taxa de mensagens por segundo
	MessagesPerSecond float64

	// Workers ativos
	ActiveWorkers int
}
