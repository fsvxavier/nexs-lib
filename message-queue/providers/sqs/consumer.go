package sqs

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

// Consumer implementa interfaces.Consumer para Amazon SQS
type Consumer struct {
	provider          *SQSProvider
	client            *sqs.Client
	queueURL          string
	handler           interfaces.MessageHandler
	options           *interfaces.ConsumerOptions
	running           bool
	wg                sync.WaitGroup
	ctx               context.Context
	cancel            context.CancelFunc
	visibilityTimeout int32
}

// NewConsumer cria um novo consumer para SQS
func NewConsumer(provider *SQSProvider, queueURL string, handler interfaces.MessageHandler, options *interfaces.ConsumerOptions) (*Consumer, error) {
	if provider == nil || provider.client == nil {
		return nil, domainerrors.New(
			"SQS_INVALID_CONNECTION",
			"invalid SQS client",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	if queueURL == "" {
		return nil, domainerrors.New(
			"SQS_INVALID_QUEUE_URL",
			"queue URL cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
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
			BufferSize:        10,
			ProcessingTimeout: 30 * time.Second,
			AutoAck:           false,
			BatchSize:         1,
		}
	}

	// Calcula visibility timeout baseado no processing timeout
	visibilityTimeout := int32(options.ProcessingTimeout.Seconds()) + 10 // +10 segundos de buffer

	ctx, cancel := context.WithCancel(context.Background())
	if options.Context != nil {
		ctx, cancel = context.WithCancel(options.Context)
	}

	consumer := &Consumer{
		provider:          provider,
		client:            provider.client,
		queueURL:          queueURL,
		handler:           handler,
		options:           options,
		running:           false,
		ctx:               ctx,
		cancel:            cancel,
		visibilityTimeout: visibilityTimeout,
	}

	return consumer, nil
}

// Start inicia o consumo de mensagens
func (c *Consumer) Start(ctx context.Context) error {
	if c.running {
		return domainerrors.New(
			"CONSUMER_ALREADY_RUNNING",
			"consumer is already running",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	c.running = true

	// Inicia workers
	for i := 0; i < c.options.Workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	return nil
}

// Stop para o consumo de mensagens
func (c *Consumer) Stop() error {
	if !c.running {
		return nil
	}

	c.running = false
	c.cancel()

	// Aguarda workers terminarem
	c.wg.Wait()

	return nil
}

// worker processa mensagens continuamente
func (c *Consumer) worker(workerID int) {
	defer c.wg.Done()

	for c.running {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.pollMessages(workerID)
		}
	}
}

// pollMessages faz polling de mensagens do SQS
func (c *Consumer) pollMessages(workerID int) {
	// Configura parâmetros de receive
	maxMessages := int32(c.options.BatchSize)
	if maxMessages > 10 { // SQS máximo é 10
		maxMessages = 10
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   maxMessages,
		WaitTimeSeconds:       20, // Long polling
		VisibilityTimeout:     c.visibilityTimeout,
		MessageAttributeNames: []string{"All"},
	}

	// Recebe mensagens
	result, err := c.client.ReceiveMessage(c.ctx, input)
	if err != nil {
		// Log do erro seria aqui
		time.Sleep(1 * time.Second) // Evita loop muito rápido em caso de erro
		return
	}

	// Processa mensagens recebidas
	for _, sqsMessage := range result.Messages {
		if !c.running {
			break
		}
		c.processMessage(sqsMessage, workerID)
	}
}

// processMessage processa uma mensagem individual
func (c *Consumer) processMessage(sqsMessage types.Message, workerID int) {
	// Cria contexto com timeout
	ctx := c.ctx
	if c.options.ProcessingTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(c.ctx, c.options.ProcessingTimeout)
		defer cancel()
	}

	// Converte mensagem SQS para Message
	message := &interfaces.Message{
		ID:        aws.ToString(sqsMessage.MessageId),
		Body:      []byte(aws.ToString(sqsMessage.Body)),
		Headers:   make(map[string]interface{}),
		Timestamp: time.Now(), // SQS não fornece timestamp original
		Source:    c.queueURL,
		Context:   ctx,
	}

	// Converte atributos
	for k, v := range sqsMessage.MessageAttributes {
		switch aws.ToString(v.DataType) {
		case "String":
			message.Headers[k] = aws.ToString(v.StringValue)
		case "Number":
			if num, err := strconv.Atoi(aws.ToString(v.StringValue)); err == nil {
				message.Headers[k] = num
			} else {
				message.Headers[k] = aws.ToString(v.StringValue)
			}
		default:
			message.Headers[k] = aws.ToString(v.StringValue)
		}
	}

	// Adiciona metadados SQS
	if sqsMessage.Attributes != nil {
		message.Headers["SQS_ApproximateReceiveCount"] = sqsMessage.Attributes["ApproximateReceiveCount"]
		message.Headers["SQS_SentTimestamp"] = sqsMessage.Attributes["SentTimestamp"]
	}

	// Processa a mensagem
	ackType := interfaces.AckSuccess
	err := c.handler(ctx, message)
	if err != nil {
		ackType = interfaces.AckReject
		// Log do erro seria aqui
	}

	// Faz acknowledge da mensagem
	c.acknowledge(sqsMessage, ackType)
}

// acknowledge faz o acknowledgment da mensagem
func (c *Consumer) acknowledge(sqsMessage types.Message, ackType interfaces.AckType) {
	if c.options.AutoAck && ackType == interfaces.AckSuccess {
		return // Já foi processada automaticamente
	}

	switch ackType {
	case interfaces.AckSuccess:
		// Deleta a mensagem da queue
		_, err := c.client.DeleteMessage(c.ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(c.queueURL),
			ReceiptHandle: sqsMessage.ReceiptHandle,
		})
		if err != nil {
			// Log do erro seria aqui
		}

	case interfaces.AckReject:
		// Para SQS, não fazer nada permite que a mensagem retorne após visibility timeout
		// ou seja enviada para DLQ se configurada

	case interfaces.AckRequeue:
		// Muda visibility timeout para 0 para reprocessar imediatamente
		_, err := c.client.ChangeMessageVisibility(c.ctx, &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String(c.queueURL),
			ReceiptHandle:     sqsMessage.ReceiptHandle,
			VisibilityTimeout: 0,
		})
		if err != nil {
			// Log do erro seria aqui
		}
	}
}

// GetMetrics retorna métricas do consumer
func (c *Consumer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	metrics["queue_url"] = c.queueURL
	metrics["provider"] = "sqs"
	metrics["running"] = c.running
	metrics["workers"] = c.options.Workers
	metrics["visibility_timeout"] = c.visibilityTimeout
	return metrics
}

// IsHealthy verifica se o consumer está saudável
func (c *Consumer) IsHealthy() bool {
	return c.client != nil && c.queueURL != ""
}
