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

// KafkaConsumer implementa MessageConsumer para Apache Kafka
type KafkaConsumer struct {
	provider      *KafkaProvider
	config        *interfaces.ConsumerConfig
	consumerGroup sarama.ConsumerGroup
	metrics       *interfaces.ConsumerMetrics
	mutex         sync.RWMutex
	running       bool
	closed        bool
	cancel        context.CancelFunc
	handlers      map[string]interfaces.MessageHandler
	batchHandlers map[string]interfaces.BatchMessageHandler
	logger        logger.Logger
}

// NewKafkaConsumer cria um novo consumer Kafka
func NewKafkaConsumer(provider *KafkaProvider, config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if provider == nil {
		return nil, domainerrors.New(
			"INVALID_PROVIDER",
			"kafka provider cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if config == nil {
		// Usa configuração padrão do provider
		config = provider.config.DefaultConsumer
	}

	if config == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"consumer config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	consumer := &KafkaConsumer{
		provider: provider,
		config:   config,
		metrics: &interfaces.ConsumerMetrics{
			MessagesProcessed:    0,
			MessagesError:        0,
			MessagesAcked:        0,
			MessagesNacked:       0,
			BytesProcessed:       0,
			AvgProcessingLatency: 0,
			LastProcessedAt:      time.Time{},
			MessagesPerSecond:    0,
			ActiveWorkers:        0,
		},
		running:       false,
		closed:        false,
		handlers:      make(map[string]interfaces.MessageHandler),
		batchHandlers: make(map[string]interfaces.BatchMessageHandler),
		logger:        logger.GetCurrentProvider(),
	}

	// Cria o consumer group
	if err := consumer.createConsumerGroup(); err != nil {
		return nil, err
	}

	return consumer, nil
}

// Subscribe inicia o consumo de mensagens de um tópico
func (c *KafkaConsumer) Subscribe(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
	if c.closed {
		return domainerrors.New(
			"CONSUMER_CLOSED",
			"consumer is closed",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	if source == "" {
		return domainerrors.New(
			"INVALID_SOURCE",
			"source topic cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if handler == nil {
		return domainerrors.New(
			"INVALID_HANDLER",
			"message handler cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	c.mutex.Lock()
	c.handlers[source] = handler
	c.mutex.Unlock()

	// Cria contexto cancelável
	consumeCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Inicia o consumo em goroutine
	go c.consume(consumeCtx, []string{source}, options)

	c.mutex.Lock()
	c.running = true
	c.mutex.Unlock()

	return nil
}

// SubscribeBatch inicia o consumo de mensagens em lote
func (c *KafkaConsumer) SubscribeBatch(ctx context.Context, source string, options *interfaces.ConsumerOptions, handler interfaces.BatchMessageHandler) error {
	if c.closed {
		return domainerrors.New(
			"CONSUMER_CLOSED",
			"consumer is closed",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	if source == "" {
		return domainerrors.New(
			"INVALID_SOURCE",
			"source topic cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if handler == nil {
		return domainerrors.New(
			"INVALID_HANDLER",
			"batch message handler cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	c.mutex.Lock()
	c.batchHandlers[source] = handler
	c.mutex.Unlock()

	// Cria contexto cancelável
	consumeCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Inicia o consumo em goroutine
	go c.consumeBatch(consumeCtx, []string{source}, options)

	c.mutex.Lock()
	c.running = true
	c.mutex.Unlock()

	return nil
}

// Ack confirma o processamento de uma mensagem
func (c *KafkaConsumer) Ack(message *interfaces.Message) error {
	if message == nil {
		return domainerrors.New(
			"INVALID_MESSAGE",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Para Kafka com consumer group, o commit é automático ou manual
	// Aqui implementamos um commit manual se necessário

	c.mutex.Lock()
	c.metrics.MessagesAcked++
	c.mutex.Unlock()

	c.logger.Debug(context.Background(), "Message acknowledged",
		logger.Field{Key: "message_id", Value: message.ID},
		logger.Field{Key: "source", Value: message.Source},
	)

	return nil
}

// Nack rejeita uma mensagem
func (c *KafkaConsumer) Nack(message *interfaces.Message, requeue bool) error {
	if message == nil {
		return domainerrors.New(
			"INVALID_MESSAGE",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	c.mutex.Lock()
	c.metrics.MessagesNacked++
	c.mutex.Unlock()

	c.logger.Warn(context.Background(), "Message rejected",
		logger.Field{Key: "message_id", Value: message.ID},
		logger.Field{Key: "source", Value: message.Source},
		logger.Field{Key: "requeue", Value: requeue},
	)

	// Para Kafka, não há nack explícito, apenas não commitamos
	// Se requeue for true, a mensagem será reprocessada

	return nil
}

// Close para o consumer e libera recursos
func (c *KafkaConsumer) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	// Cancela o contexto de consumo
	if c.cancel != nil {
		c.cancel()
	}

	// Fecha o consumer group
	var err error
	if c.consumerGroup != nil {
		err = c.consumerGroup.Close()
	}

	c.running = false
	c.closed = true

	// Atualiza métricas do provider
	c.provider.mutex.Lock()
	if c.provider.metrics.ActiveConsumers > 0 {
		c.provider.metrics.ActiveConsumers--
	}
	c.provider.mutex.Unlock()

	if err != nil {
		return domainerrors.New(
			"KAFKA_CONSUMER_CLOSE_ERROR",
			"failed to close consumer group",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("consumer close error", err)
	}

	return nil
}

// Pause pausa temporariamente o consumo
func (c *KafkaConsumer) Pause() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running {
		return domainerrors.New(
			"CONSUMER_NOT_RUNNING",
			"consumer is not running",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Para Kafka, pausamos através do contexto
	if c.cancel != nil {
		c.cancel()
	}

	c.running = false
	return nil
}

// Resume retoma o consumo após uma pausa
func (c *KafkaConsumer) Resume() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return domainerrors.New(
			"CONSUMER_ALREADY_RUNNING",
			"consumer is already running",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Implementar lógica de resume se necessário
	c.running = true
	return nil
}

// IsConnected verifica se o consumer está conectado
func (c *KafkaConsumer) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return !c.closed && c.consumerGroup != nil && c.provider.IsConnected()
}

// IsRunning verifica se o consumer está ativo
func (c *KafkaConsumer) IsRunning() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.running && !c.closed
}

// GetMetrics retorna métricas do consumer
func (c *KafkaConsumer) GetMetrics() *interfaces.ConsumerMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Cria uma cópia das métricas
	return &interfaces.ConsumerMetrics{
		MessagesProcessed:    c.metrics.MessagesProcessed,
		MessagesError:        c.metrics.MessagesError,
		MessagesAcked:        c.metrics.MessagesAcked,
		MessagesNacked:       c.metrics.MessagesNacked,
		BytesProcessed:       c.metrics.BytesProcessed,
		AvgProcessingLatency: c.metrics.AvgProcessingLatency,
		LastProcessedAt:      c.metrics.LastProcessedAt,
		MessagesPerSecond:    c.metrics.MessagesPerSecond,
		ActiveWorkers:        c.metrics.ActiveWorkers,
	}
}

// createConsumerGroup cria o consumer group do Sarama
func (c *KafkaConsumer) createConsumerGroup() error {
	client := c.provider.GetClient()
	if client == nil {
		return domainerrors.New(
			"KAFKA_CLIENT_NOT_AVAILABLE",
			"kafka client is not available",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	groupID := c.config.ConsumerGroup
	if groupID == "" {
		groupID = "default-group"
	}

	// Configurações específicas do consumer
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Configuração de offset
	switch c.config.InitialOffset {
	case "earliest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	// Commit automático
	if c.config.AutoCommit {
		config.Consumer.Offsets.AutoCommit.Enable = true
		if c.config.CommitInterval > 0 {
			config.Consumer.Offsets.AutoCommit.Interval = c.config.CommitInterval
		}
	}

	// Cria o consumer group
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return domainerrors.New(
			"KAFKA_CONSUMER_GROUP_CREATION_FAILED",
			"failed to create Kafka consumer group",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("group_id", groupID).
			Wrap("consumer group creation error", err)
	}

	c.consumerGroup = consumerGroup
	return nil
}

// consume inicia o consumo de mensagens individuais
func (c *KafkaConsumer) consume(ctx context.Context, topics []string, options *interfaces.ConsumerOptions) {
	handler := &kafkaConsumerGroupHandler{
		consumer: c,
		options:  options,
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
				c.logger.Error(ctx, "Consumer group error",
					logger.Field{Key: "error", Value: err.Error()},
					logger.Field{Key: "topics", Value: topics},
				)

				// Pausa antes de tentar novamente
				time.Sleep(time.Second)
			}
		}
	}
}

// consumeBatch inicia o consumo de mensagens em lote
func (c *KafkaConsumer) consumeBatch(ctx context.Context, topics []string, options *interfaces.ConsumerOptions) {
	handler := &kafkaBatchConsumerGroupHandler{
		consumer: c,
		options:  options,
		messages: make([]*interfaces.Message, 0, options.BatchSize),
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
				c.logger.Error(ctx, "Batch consumer group error",
					logger.Field{Key: "error", Value: err.Error()},
					logger.Field{Key: "topics", Value: topics},
				)

				// Pausa antes de tentar novamente
				time.Sleep(time.Second)
			}
		}
	}
}

// updateSuccessMetrics atualiza métricas de sucesso
func (c *KafkaConsumer) updateSuccessMetrics(messageSize int, latency time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.MessagesProcessed++
	c.metrics.BytesProcessed += int64(messageSize)
	c.metrics.LastProcessedAt = time.Now()

	// Atualiza latência média (média móvel simples)
	if c.metrics.AvgProcessingLatency == 0 {
		c.metrics.AvgProcessingLatency = latency
	} else {
		c.metrics.AvgProcessingLatency = (c.metrics.AvgProcessingLatency + latency) / 2
	}

	// Calcula mensagens por segundo
	if c.metrics.MessagesProcessed > 0 {
		duration := time.Since(c.metrics.LastProcessedAt)
		if duration > 0 {
			c.metrics.MessagesPerSecond = float64(c.metrics.MessagesProcessed) / duration.Seconds()
		}
	}
}

// updateErrorMetrics atualiza métricas de erro
func (c *KafkaConsumer) updateErrorMetrics() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics.MessagesError++
}

// kafkaConsumerGroupHandler implementa sarama.ConsumerGroupHandler para mensagens individuais
type kafkaConsumerGroupHandler struct {
	consumer *KafkaConsumer
	options  *interfaces.ConsumerOptions
}

func (h *kafkaConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Converte para nossa interface
			msg := h.convertFromKafkaMessage(message)

			// Busca o handler para este tópico
			h.consumer.mutex.RLock()
			handler, exists := h.consumer.handlers[message.Topic]
			h.consumer.mutex.RUnlock()

			if !exists {
				h.consumer.logger.Warn(context.Background(), "No handler found for topic",
					logger.Field{Key: "topic", Value: message.Topic},
				)
				continue
			}

			// Processa a mensagem
			startTime := time.Now()
			ctx := context.Background()

			if err := handler(ctx, msg); err != nil {
				h.consumer.updateErrorMetrics()
				h.consumer.logger.Error(ctx, "Error processing message",
					logger.Field{Key: "error", Value: err.Error()},
					logger.Field{Key: "topic", Value: message.Topic},
					logger.Field{Key: "offset", Value: message.Offset},
				)
				continue
			}

			// Atualiza métricas de sucesso
			latency := time.Since(startTime)
			h.consumer.updateSuccessMetrics(len(message.Value), latency)

			// Marca mensagem como processada
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// kafkaBatchConsumerGroupHandler implementa sarama.ConsumerGroupHandler para mensagens em lote
type kafkaBatchConsumerGroupHandler struct {
	consumer *KafkaConsumer
	options  *interfaces.ConsumerOptions
	messages []*interfaces.Message
	mutex    sync.Mutex
}

func (h *kafkaBatchConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaBatchConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaBatchConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ticker := time.NewTicker(h.options.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Converte e adiciona ao batch
			msg := h.convertFromKafkaMessage(message)
			h.mutex.Lock()
			h.messages = append(h.messages, msg)
			batchReady := len(h.messages) >= h.options.BatchSize
			h.mutex.Unlock()

			// Processa o batch se estiver pronto
			if batchReady {
				h.processBatch(session, message.Topic)
			}

		case <-ticker.C:
			// Processa batch por timeout
			h.processBatch(session, "")

		case <-session.Context().Done():
			// Processa batch final antes de sair
			h.processBatch(session, "")
			return nil
		}
	}
}

func (h *kafkaBatchConsumerGroupHandler) processBatch(session sarama.ConsumerGroupSession, topic string) {
	h.mutex.Lock()
	if len(h.messages) == 0 {
		h.mutex.Unlock()
		return
	}

	batch := make([]*interfaces.Message, len(h.messages))
	copy(batch, h.messages)
	h.messages = h.messages[:0] // Reset slice
	h.mutex.Unlock()

	// Busca o handler
	if topic == "" && len(batch) > 0 {
		topic = batch[0].Source
	}

	h.consumer.mutex.RLock()
	handler, exists := h.consumer.batchHandlers[topic]
	h.consumer.mutex.RUnlock()

	if !exists {
		h.consumer.logger.Warn(context.Background(), "No batch handler found for topic",
			logger.Field{Key: "topic", Value: topic},
		)
		return
	}

	// Processa o batch
	startTime := time.Now()
	ctx := context.Background()

	if err := handler(ctx, batch); err != nil {
		h.consumer.updateErrorMetrics()
		h.consumer.logger.Error(ctx, "Error processing batch",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "topic", Value: topic},
			logger.Field{Key: "batch_size", Value: len(batch)},
		)
		return
	}

	// Atualiza métricas
	latency := time.Since(startTime)
	totalSize := 0
	for _, msg := range batch {
		totalSize += len(msg.Body)
	}
	h.consumer.updateSuccessMetrics(totalSize, latency)
}

func (h *kafkaConsumerGroupHandler) convertFromKafkaMessage(msg *sarama.ConsumerMessage) *interfaces.Message {
	message := &interfaces.Message{
		ID:         fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset),
		Body:       msg.Value,
		Headers:    make(map[string]interface{}),
		Timestamp:  msg.Timestamp,
		ReceivedAt: time.Now(),
		Attempts:   0,
		Source:     msg.Topic,
		Context:    context.Background(),
	}

	// Converte headers
	for _, header := range msg.Headers {
		message.Headers[string(header.Key)] = string(header.Value)
	}

	// Extrai trace ID e span ID dos headers
	if traceID, exists := message.Headers["trace_id"]; exists {
		if traceStr, ok := traceID.(string); ok {
			message.TraceID = traceStr
		}
	}

	if spanID, exists := message.Headers["span_id"]; exists {
		if spanStr, ok := spanID.(string); ok {
			message.SpanID = spanStr
		}
	}

	return message
}

func (h *kafkaBatchConsumerGroupHandler) convertFromKafkaMessage(msg *sarama.ConsumerMessage) *interfaces.Message {
	return (&kafkaConsumerGroupHandler{}).convertFromKafkaMessage(msg)
}
