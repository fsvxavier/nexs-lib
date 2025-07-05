package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/sqs"
)

func main() {
	// Configuração para Amazon SQS
	cfg := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{}, // SQS não usa brokers tradicionais
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			// Para SQS, as credenciais geralmente vêm de AWS CLI, IAM roles, ou environment variables
			ProviderConfig: map[string]interface{}{
				"region": "us-east-1",
				// "access_key_id": "your-access-key",     // Opcional se usar IAM roles
				// "secret_access_key": "your-secret-key", // Opcional se usar IAM roles
			},
		},
	}

	// Criar provider SQS
	provider, err := sqs.NewSQSProvider(cfg)
	if err != nil {
		log.Fatalf("Erro ao criar provider: %v", err)
	}

	// Conectar
	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer provider.Close()

	fmt.Println("Conectado ao Amazon SQS!")

	// Criar producer para uma fila específica
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	producer, err := sqs.NewProducer(provider.(*sqs.SQSProvider), queueURL)
	if err != nil {
		log.Fatalf("Erro ao criar producer: %v", err)
	}
	defer producer.Close()

	// Criar mensagem simples
	simpleMessage := &interfaces.Message{
		ID:   "sqs-msg-001",
		Body: []byte(`{"type": "order_created", "order_id": "ORD-12345", "amount": 99.99}`),
		Headers: map[string]interface{}{
			"MessageType": "order_event",
			"Source":      "order-service",
			"Version":     "1.0",
		},
		Timestamp: time.Now(),
	}

	// Enviar mensagem simples
	fmt.Println("Enviando mensagem simples...")
	if err := producer.Send(ctx, simpleMessage); err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}
	fmt.Printf("Mensagem %s enviada com sucesso!\n", simpleMessage.ID)

	// Criar mensagens FIFO (se usando fila FIFO)
	fifoMessages := []*interfaces.Message{
		{
			ID:   "fifo-msg-001",
			Body: []byte(`{"sequence": 1, "data": "primeira mensagem"}`),
			Headers: map[string]interface{}{
				"MessageGroupId":         "group-1",
				"MessageDeduplicationId": "dedup-001",
				"ContentType":            "application/json",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "fifo-msg-002",
			Body: []byte(`{"sequence": 2, "data": "segunda mensagem"}`),
			Headers: map[string]interface{}{
				"MessageGroupId":         "group-1",
				"MessageDeduplicationId": "dedup-002",
				"ContentType":            "application/json",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "fifo-msg-003",
			Body: []byte(`{"sequence": 3, "data": "terceira mensagem"}`),
			Headers: map[string]interface{}{
				"MessageGroupId":         "group-1",
				"MessageDeduplicationId": "dedup-003",
				"ContentType":            "application/json",
			},
			Timestamp: time.Now(),
		},
	}

	// Enviar mensagens em lote
	fmt.Println("\nEnviando mensagens em lote...")
	if err := producer.SendBatch(ctx, fifoMessages); err != nil {
		log.Fatalf("Erro ao enviar lote: %v", err)
	}
	fmt.Println("Lote FIFO enviado com sucesso!")

	// Criar consumer
	handler := func(ctx context.Context, msg *interfaces.Message) error {
		fmt.Printf("Mensagem recebida: %s\n", string(msg.Body))
		fmt.Printf("  ID: %s\n", msg.ID)
		fmt.Printf("  Source: %s\n", msg.Source)
		fmt.Printf("  Headers: %+v\n", msg.Headers)

		// Simular processamento
		time.Sleep(100 * time.Millisecond)

		return nil // Retorna nil para ACK, erro para NACK
	}

	consumerOptions := &interfaces.ConsumerOptions{
		Workers:           3,
		BufferSize:        10, // SQS máximo é 10 mensagens por request
		ProcessingTimeout: 30 * time.Second,
		AutoAck:           false, // Controle manual de ACK/NACK
		BatchSize:         5,     // Processar até 5 mensagens por vez
	}

	consumer, err := sqs.NewConsumer(provider.(*sqs.SQSProvider), queueURL, handler, consumerOptions)
	if err != nil {
		log.Fatalf("Erro ao criar consumer: %v", err)
	}

	// Iniciar consumo
	fmt.Println("\nIniciando consumer...")
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Erro ao iniciar consumer: %v", err)
	}

	fmt.Println("Consumer iniciado. Processando mensagens por 10 segundos...")

	// Aguardar processamento
	time.Sleep(10 * time.Second)

	// Parar consumer
	if err := consumer.Stop(); err != nil {
		log.Printf("Erro ao parar consumer: %v", err)
	}

	// Verificar métricas
	producerMetrics := producer.GetMetrics()
	consumerMetrics := consumer.GetMetrics()
	providerMetrics := provider.GetMetrics()

	fmt.Printf("\nMétricas finais:\n")
	fmt.Printf("Producer: %+v\n", producerMetrics)
	fmt.Printf("Consumer: %+v\n", consumerMetrics)
	fmt.Printf("Provider - Conexões ativas: %d\n", providerMetrics.ActiveConnections)
	fmt.Printf("Provider - Health status: %v\n", providerMetrics.HealthCheckStatus)

	fmt.Println("\nExemplo SQS concluído!")
}
