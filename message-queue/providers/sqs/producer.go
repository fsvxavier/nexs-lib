package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

// Producer implementa interfaces.Producer para Amazon SQS
type Producer struct {
	provider *SQSProvider
	client   *sqs.Client
	queueURL string
}

// NewProducer cria um novo producer para SQS
func NewProducer(provider *SQSProvider, queueURL string) (*Producer, error) {
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

	return &Producer{
		provider: provider,
		client:   provider.client,
		queueURL: queueURL,
	}, nil
}

// Send envia uma mensagem para o SQS
func (p *Producer) Send(ctx context.Context, message *interfaces.Message) error {
	if message == nil {
		return domainerrors.New(
			"MESSAGE_NIL",
			"message cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Converte body para string
	messageBody := string(message.Body)
	if messageBody == "" {
		return domainerrors.New(
			"MESSAGE_BODY_EMPTY",
			"message body cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	// Prepara atributos da mensagem
	messageAttributes := make(map[string]types.MessageAttributeValue)
	for k, v := range message.Headers {
		if str, ok := v.(string); ok {
			messageAttributes[k] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(str),
			}
		} else if num, ok := v.(int); ok {
			messageAttributes[k] = types.MessageAttributeValue{
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", num)),
			}
		} else {
			// Serializa outros tipos como JSON
			jsonBytes, err := json.Marshal(v)
			if err == nil {
				messageAttributes[k] = types.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(string(jsonBytes)),
				}
			}
		}
	}

	// Prepara input para SendMessage
	input := &sqs.SendMessageInput{
		QueueUrl:          aws.String(p.queueURL),
		MessageBody:       aws.String(messageBody),
		MessageAttributes: messageAttributes,
	}

	// Adiciona Message Group ID se especificado (para FIFO queues)
	if groupID, exists := message.Headers["MessageGroupId"]; exists {
		if groupIDStr, ok := groupID.(string); ok {
			input.MessageGroupId = aws.String(groupIDStr)
		}
	}

	// Adiciona Deduplication ID se especificado (para FIFO queues)
	if dedupID, exists := message.Headers["MessageDeduplicationId"]; exists {
		if dedupIDStr, ok := dedupID.(string); ok {
			input.MessageDeduplicationId = aws.String(dedupIDStr)
		}
	}

	// Envia a mensagem
	_, err := p.client.SendMessage(ctx, input)
	if err != nil {
		return domainerrors.New(
			"SQS_SEND_ERROR",
			fmt.Sprintf("failed to send message to SQS: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// SendBatch envia múltiplas mensagens em lote
func (p *Producer) SendBatch(ctx context.Context, messages []*interfaces.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// SQS suporta até 10 mensagens por lote
	const maxBatchSize = 10

	for i := 0; i < len(messages); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(messages) {
			end = len(messages)
		}

		batch := messages[i:end]
		if err := p.sendBatch(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

// sendBatch envia um lote de mensagens
func (p *Producer) sendBatch(ctx context.Context, messages []*interfaces.Message) error {
	entries := make([]types.SendMessageBatchRequestEntry, len(messages))

	for i, message := range messages {
		messageBody := string(message.Body)

		// Prepara atributos
		messageAttributes := make(map[string]types.MessageAttributeValue)
		for k, v := range message.Headers {
			if str, ok := v.(string); ok {
				messageAttributes[k] = types.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(str),
				}
			}
		}

		entry := types.SendMessageBatchRequestEntry{
			Id:                aws.String(fmt.Sprintf("msg-%d", i)),
			MessageBody:       aws.String(messageBody),
			MessageAttributes: messageAttributes,
		}

		// Adiciona campos FIFO se necessário
		if groupID, exists := message.Headers["MessageGroupId"]; exists {
			if groupIDStr, ok := groupID.(string); ok {
				entry.MessageGroupId = aws.String(groupIDStr)
			}
		}

		if dedupID, exists := message.Headers["MessageDeduplicationId"]; exists {
			if dedupIDStr, ok := dedupID.(string); ok {
				entry.MessageDeduplicationId = aws.String(dedupIDStr)
			}
		}

		entries[i] = entry
	}

	// Envia o lote
	input := &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(p.queueURL),
		Entries:  entries,
	}

	result, err := p.client.SendMessageBatch(ctx, input)
	if err != nil {
		return domainerrors.New(
			"SQS_BATCH_SEND_ERROR",
			fmt.Sprintf("failed to send batch to SQS: %v", err),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Verifica se houve falhas
	if len(result.Failed) > 0 {
		return domainerrors.New(
			"SQS_BATCH_PARTIAL_FAILURE",
			fmt.Sprintf("batch partially failed: %d out of %d messages failed", len(result.Failed), len(entries)),
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// Close fecha o producer
func (p *Producer) Close() error {
	// SQS client não precisa ser fechado explicitamente
	return nil
}

// GetMetrics retorna métricas do producer
func (p *Producer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	metrics["queue_url"] = p.queueURL
	metrics["provider"] = "sqs"
	return metrics
}

// IsHealthy verifica se o producer está saudável
func (p *Producer) IsHealthy() bool {
	return p.client != nil && p.queueURL != ""
}
