package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
)

// KafkaProducer implementa MessageProducer para Apache Kafka
type KafkaProducer struct {
	provider      *KafkaProvider
	config        *interfaces.ProducerConfig
	producer      sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	metrics       *interfaces.ProducerMetrics
	mutex         sync.RWMutex
	closed        bool
	logger        logger.Logger
}

// NewKafkaProducer cria um novo producer Kafka
func NewKafkaProducer(provider *KafkaProvider, config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if provider == nil {
		return nil, domainerrors.New(
			"INVALID_PROVIDER",
			"kafka provider cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if config == nil {
		// Usa configuração padrão do provider
		config = provider.config.DefaultProducer
	}

	if config == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"producer config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	producer := &KafkaProducer{
		provider: provider,
		config:   config,
		metrics: &interfaces.ProducerMetrics{
			MessagesSent:      0,
			MessagesError:     0,
			BytesSent:         0,
			AvgLatency:        0,
			LastSentAt:        time.Time{},
			MessagesPerSecond: 0,
		},
		closed: false,
		logger: logger.GetCurrentProvider(),
	}

	// Cria o producer Sarama
	if err := producer.createSaramaProducer(); err != nil {
		return nil, err
	}

	return producer, nil
}

// Publish envia uma mensagem para o tópico especificado
func (p *KafkaProducer) Publish(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions) error {
	if p.closed {
		return domainerrors.New(
			"PRODUCER_CLOSED",
			"producer is closed",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	if destination == "" {
		return domainerrors.New(
			"INVALID_DESTINATION",
			"destination topic cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if message == nil {
		return domainerrors.New(
			"INVALID_MESSAGE",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	startTime := time.Now()

	// Converte para mensagem Kafka
	kafkaMessage, err := p.convertToKafkaMessage(destination, message, options)
	if err != nil {
		p.updateErrorMetrics()
		return err
	}

	// Envia usando producer síncrono
	partition, offset, err := p.producer.SendMessage(kafkaMessage)
	if err != nil {
		p.updateErrorMetrics()
		p.logger.Error(ctx, "Failed to send Kafka message",
			logger.Field{Key: "topic", Value: destination},
			logger.Field{Key: "error", Value: err.Error()},
		)

		return domainerrors.New(
			"KAFKA_SEND_FAILED",
			"failed to send message to Kafka",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("topic", destination).
			WithDetail("partition", partition).
			WithDetail("offset", offset).
			Wrap("kafka send error", err)
	}

	// Atualiza métricas de sucesso
	latency := time.Since(startTime)
	p.updateSuccessMetrics(len(message.Body), latency)

	p.logger.Debug(ctx, "Message sent to Kafka",
		logger.Field{Key: "topic", Value: destination},
		logger.Field{Key: "partition", Value: partition},
		logger.Field{Key: "offset", Value: offset},
		logger.Field{Key: "latency", Value: latency.String()},
	)

	return nil
}

// PublishBatch envia múltiplas mensagens em lote
func (p *KafkaProducer) PublishBatch(ctx context.Context, destination string, messages []*interfaces.Message, options *interfaces.MessageOptions) error {
	if p.closed {
		return domainerrors.New(
			"PRODUCER_CLOSED",
			"producer is closed",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	if len(messages) == 0 {
		return nil // Nada para enviar
	}

	// Para Kafka, enviamos uma por uma (Sarama não tem batch nativo simples)
	var errors []error
	successCount := 0

	for i, message := range messages {
		if err := p.Publish(ctx, destination, message, options); err != nil {
			errors = append(errors, fmt.Errorf("message %d failed: %w", i, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return domainerrors.New(
			"KAFKA_BATCH_SEND_PARTIAL_FAILURE",
			fmt.Sprintf("batch send partially failed: %d/%d messages sent", successCount, len(messages)),
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("success_count", successCount).
			WithDetail("total_count", len(messages)).
			WithDetail("errors", errors)
	}

	return nil
}

// PublishWithCallback envia uma mensagem com callback de confirmação
func (p *KafkaProducer) PublishWithCallback(ctx context.Context, destination string, message *interfaces.Message, options *interfaces.MessageOptions, callback func(error)) error {
	// Para simplicidade, executa síncronamente e chama o callback
	err := p.Publish(ctx, destination, message, options)
	if callback != nil {
		callback(err)
	}
	return err
}

// Close fecha o producer e libera recursos
func (p *KafkaProducer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil
	}

	var errors []error

	if p.asyncProducer != nil {
		if err := p.asyncProducer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close async producer: %w", err))
		}
	}

	if p.producer != nil {
		if err := p.producer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close producer: %w", err))
		}
	}

	p.closed = true

	// Atualiza métricas do provider
	p.provider.mutex.Lock()
	if p.provider.metrics.ActiveProducers > 0 {
		p.provider.metrics.ActiveProducers--
	}
	p.provider.mutex.Unlock()

	if len(errors) > 0 {
		return domainerrors.New(
			"KAFKA_PRODUCER_CLOSE_ERROR",
			"errors occurred while closing producer",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("errors", errors)
	}

	return nil
}

// IsConnected verifica se o producer está conectado
func (p *KafkaProducer) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return !p.closed && p.producer != nil && p.provider.IsConnected()
}

// GetMetrics retorna métricas do producer
func (p *KafkaProducer) GetMetrics() *interfaces.ProducerMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Cria uma cópia das métricas
	return &interfaces.ProducerMetrics{
		MessagesSent:      p.metrics.MessagesSent,
		MessagesError:     p.metrics.MessagesError,
		BytesSent:         p.metrics.BytesSent,
		AvgLatency:        p.metrics.AvgLatency,
		LastSentAt:        p.metrics.LastSentAt,
		MessagesPerSecond: p.metrics.MessagesPerSecond,
	}
}

// createSaramaProducer cria o producer Sarama interno
func (p *KafkaProducer) createSaramaProducer() error {
	client := p.provider.GetClient()
	if client == nil {
		return domainerrors.New(
			"KAFKA_CLIENT_NOT_AVAILABLE",
			"kafka client is not available",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	// Configurações específicas do producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// Aplica configurações do config
	if p.config.Transactional {
		config.Producer.Idempotent = true
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Net.MaxOpenRequests = 1
	}

	if p.config.Compression != "" {
		switch p.config.Compression {
		case "gzip":
			config.Producer.Compression = sarama.CompressionGZIP
		case "snappy":
			config.Producer.Compression = sarama.CompressionSnappy
		case "lz4":
			config.Producer.Compression = sarama.CompressionLZ4
		case "zstd":
			config.Producer.Compression = sarama.CompressionZSTD
		default:
			config.Producer.Compression = sarama.CompressionNone
		}
	}

	// Cria o producer síncrono
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return domainerrors.New(
			"KAFKA_PRODUCER_CREATION_FAILED",
			"failed to create Kafka producer",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("producer creation error", err)
	}

	p.producer = producer
	return nil
}

// convertToKafkaMessage converte uma Message para ProducerMessage do Sarama
func (p *KafkaProducer) convertToKafkaMessage(topic string, message *interfaces.Message, options *interfaces.MessageOptions) (*sarama.ProducerMessage, error) {
	kafkaMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message.Body),
	}

	// Adiciona headers
	if message.Headers != nil || (options != nil && options.Headers != nil) {
		headers := make([]sarama.RecordHeader, 0)

		// Headers da mensagem
		if message.Headers != nil {
			for k, v := range message.Headers {
				if strVal, ok := v.(string); ok {
					headers = append(headers, sarama.RecordHeader{
						Key:   []byte(k),
						Value: []byte(strVal),
					})
				}
			}
		}

		// Headers das opções
		if options != nil && options.Headers != nil {
			for k, v := range options.Headers {
				if strVal, ok := v.(string); ok {
					headers = append(headers, sarama.RecordHeader{
						Key:   []byte(k),
						Value: []byte(strVal),
					})
				}
			}
		}

		// Adiciona headers de observabilidade
		if message.TraceID != "" {
			headers = append(headers, sarama.RecordHeader{
				Key:   []byte("trace_id"),
				Value: []byte(message.TraceID),
			})
		}

		if message.SpanID != "" {
			headers = append(headers, sarama.RecordHeader{
				Key:   []byte("span_id"),
				Value: []byte(message.SpanID),
			})
		}

		kafkaMessage.Headers = headers
	}

	// Define key se disponível nos headers
	if options != nil && options.MessageID != "" {
		kafkaMessage.Key = sarama.StringEncoder(options.MessageID)
	} else if message.ID != "" {
		kafkaMessage.Key = sarama.StringEncoder(message.ID)
	}

	// Timestamp
	if !message.Timestamp.IsZero() {
		kafkaMessage.Timestamp = message.Timestamp
	} else {
		kafkaMessage.Timestamp = time.Now()
	}

	return kafkaMessage, nil
}

// updateSuccessMetrics atualiza métricas de sucesso
func (p *KafkaProducer) updateSuccessMetrics(messageSize int, latency time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics.MessagesSent++
	p.metrics.BytesSent += int64(messageSize)
	p.metrics.LastSentAt = time.Now()

	// Atualiza latência média (média móvel simples)
	if p.metrics.AvgLatency == 0 {
		p.metrics.AvgLatency = latency
	} else {
		p.metrics.AvgLatency = (p.metrics.AvgLatency + latency) / 2
	}

	// Calcula mensagens por segundo (últimos 60 segundos)
	if p.metrics.MessagesSent > 0 {
		duration := time.Since(p.metrics.LastSentAt)
		if duration > 0 {
			p.metrics.MessagesPerSecond = float64(p.metrics.MessagesSent) / duration.Seconds()
		}
	}
}

// updateErrorMetrics atualiza métricas de erro
func (p *KafkaProducer) updateErrorMetrics() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics.MessagesError++
}
