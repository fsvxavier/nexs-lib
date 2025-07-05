package interfaces

import (
	"context"
	"time"
)

// Message representa uma mensagem no sistema de filas
type Message struct {
	// ID único da mensagem
	ID string

	// Corpo da mensagem
	Body []byte

	// Headers/Metadata da mensagem
	Headers map[string]interface{}

	// Timestamp de quando a mensagem foi criada
	Timestamp time.Time

	// Timestamp de quando a mensagem foi recebida (para consumers)
	ReceivedAt time.Time

	// Número de tentativas de processamento
	Attempts int

	// Queue/Topic de origem
	Source string

	// Queue/Topic de destino (para producers)
	Destination string

	// Propriedades específicas do provider
	ProviderMetadata map[string]interface{}

	// Trace ID para observabilidade distribuída
	TraceID string

	// Span ID para observabilidade distribuída
	SpanID string

	// Context para propagação de dados
	Context context.Context
}

// MessageOptions representa opções para envio de mensagens
type MessageOptions struct {
	// Headers personalizados
	Headers map[string]interface{}

	// Delay antes do processamento (TTL)
	Delay time.Duration

	// Prioridade da mensagem (se suportado pelo provider)
	Priority int

	// Se a mensagem deve ser persistente
	Persistent bool

	// Timeout para envio
	Timeout time.Duration

	// Queue/Topic de destino para DLQ
	DLQDestination string

	// ID único da mensagem (se não fornecido, será gerado automaticamente)
	MessageID string

	// Correlation ID para rastreamento
	CorrelationID string

	// Context para propagação de dados
	Context context.Context
}

// ConsumerOptions representa opções para configuração do consumer
type ConsumerOptions struct {
	// Nome do consumer group (para Kafka)
	ConsumerGroup string

	// Número de workers concorrentes
	Workers int

	// Tamanho do buffer de mensagens
	BufferSize int

	// Timeout para processamento de cada mensagem
	ProcessingTimeout time.Duration

	// Auto commit/ack das mensagens
	AutoAck bool

	// Batch size para processamento em lote
	BatchSize int

	// Intervalo para commit em lote
	BatchInterval time.Duration

	// Configurações de retry
	RetryPolicy *RetryPolicy

	// Configurações de DLQ
	DLQConfig *DLQConfig

	// Context para cancelamento
	Context context.Context
}

// RetryPolicy representa a política de retry para mensagens
type RetryPolicy struct {
	// Número máximo de tentativas
	MaxAttempts int

	// Delay inicial entre tentativas
	InitialDelay time.Duration

	// Multiplicador para backoff exponencial
	BackoffMultiplier float64

	// Delay máximo entre tentativas
	MaxDelay time.Duration

	// Jitter para aleatorizar delays
	Jitter bool
}

// DLQConfig representa configurações para Dead Letter Queue
type DLQConfig struct {
	// Se DLQ está habilitada
	Enabled bool

	// Nome da DLQ
	QueueName string

	// Número máximo de tentativas antes de enviar para DLQ
	MaxAttempts int

	// Headers adicionais para mensagens na DLQ
	Headers map[string]interface{}
}

// AckType representa o tipo de acknowledgment
type AckType int

const (
	// AckSuccess confirma o processamento bem-sucedido
	AckSuccess AckType = iota

	// AckReject rejeita a mensagem (pode ir para DLQ)
	AckReject

	// AckRequeue rejeita a mensagem e a recoloca na fila
	AckRequeue
)
