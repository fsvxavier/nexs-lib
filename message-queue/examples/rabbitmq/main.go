package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/rabbitmq"
)

func main() {
	// Configuração para RabbitMQ
	cfg := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{"localhost:5672"},
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			Auth: &interfaces.AuthConfig{
				Username: "guest",
				Password: "guest",
			},
		},
	}

	// Criar provider RabbitMQ
	provider, err := rabbitmq.NewRabbitMQProvider(cfg)
	if err != nil {
		log.Fatalf("Erro ao criar provider: %v", err)
	}

	// Conectar
	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer provider.Close()

	// Criar producer
	producer, err := rabbitmq.NewProducer(provider.(*rabbitmq.RabbitMQProvider), "test-exchange", "test-key")
	if err != nil {
		log.Fatalf("Erro ao criar producer: %v", err)
	}
	defer producer.Close()

	// Criar uma mensagem
	message := &interfaces.Message{
		ID:        "msg-001",
		Body:      []byte(`{"message": "Hello RabbitMQ!", "timestamp": "2025-01-01T12:00:00Z"}`),
		Headers:   map[string]interface{}{"content-type": "application/json"},
		Timestamp: time.Now(),
	}

	// Enviar mensagem
	if err := producer.Send(ctx, message); err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}

	fmt.Println("Mensagem enviada com sucesso!")

	// Criar consumer
	handler := func(ctx context.Context, msg *interfaces.Message) error {
		fmt.Printf("Mensagem recebida: %s\n", string(msg.Body))
		fmt.Printf("Headers: %+v\n", msg.Headers)
		return nil
	}

	consumerOptions := &interfaces.ConsumerOptions{
		Workers:           2,
		BufferSize:        100,
		ProcessingTimeout: 30 * time.Second,
		AutoAck:           false,
	}

	consumer, err := rabbitmq.NewConsumer(provider.(*rabbitmq.RabbitMQProvider), "test-queue", handler, consumerOptions)
	if err != nil {
		log.Fatalf("Erro ao criar consumer: %v", err)
	}
	defer consumer.Stop()

	// Iniciar consumo
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Erro ao iniciar consumer: %v", err)
	}

	fmt.Println("Consumer iniciado. Pressione Ctrl+C para parar...")

	// Aguardar um pouco para demonstração
	time.Sleep(5 * time.Second)

	// Verificar métricas
	producerMetrics := producer.GetMetrics()
	consumerMetrics := consumer.GetMetrics()

	fmt.Printf("Métricas do Producer: %+v\n", producerMetrics)
	fmt.Printf("Métricas do Consumer: %+v\n", consumerMetrics)
}
