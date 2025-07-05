package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/kafka"
)

func main() {
	// Configuração para Kafka
	cfg := &config.ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{"localhost:9092"},
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
		},
	}

	// Criar provider Kafka
	provider, err := kafka.NewKafkaProvider(cfg)
	if err != nil {
		log.Fatalf("Erro ao criar provider: %v", err)
	}

	// Conectar
	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer provider.Close()

	// Verificar health
	if err := provider.HealthCheck(ctx); err != nil {
		log.Printf("Health check falhou: %v", err)
	} else {
		fmt.Println("Health check OK!")
	}

	// Criar producer usando a interface provider
	producerConfig := &interfaces.ProducerConfig{
		ID:         "kafka-producer-example",
		BufferSize: 1000,
		ProviderConfig: map[string]interface{}{
			"topic": "test-topic",
		},
	}

	producer, err := provider.CreateProducer(producerConfig)
	if err != nil {
		log.Fatalf("Erro ao criar producer: %v", err)
	}
	defer producer.Close()

	// Criar mensagens em lote
	messages := []*interfaces.Message{
		{
			ID:        "msg-001",
			Body:      []byte(`{"event": "user_created", "user_id": 123, "timestamp": "2025-01-01T12:00:00Z"}`),
			Headers:   map[string]interface{}{"event-type": "user_created", "source": "user-service"},
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-002",
			Body:      []byte(`{"event": "user_updated", "user_id": 123, "timestamp": "2025-01-01T12:01:00Z"}`),
			Headers:   map[string]interface{}{"event-type": "user_updated", "source": "user-service"},
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-003",
			Body:      []byte(`{"event": "user_deleted", "user_id": 123, "timestamp": "2025-01-01T12:02:00Z"}`),
			Headers:   map[string]interface{}{"event-type": "user_deleted", "source": "user-service"},
			Timestamp: time.Now(),
		},
	}

	// Enviar mensagens individuais
	fmt.Println("Enviando mensagens individuais...")
	for _, msg := range messages {
		if err := producer.Publish(ctx, "test-topic", msg, nil); err != nil {
			log.Printf("Erro ao enviar mensagem %s: %v", msg.ID, err)
		} else {
			fmt.Printf("Mensagem %s enviada com sucesso!\n", msg.ID)
		}
	}

	// Enviar mensagens em lote
	fmt.Println("\nEnviando mensagens em lote...")
	batchMessages := []*interfaces.Message{
		{
			ID:        "batch-001",
			Body:      []byte(`{"batch": true, "message": 1}`),
			Headers:   map[string]interface{}{"batch": "true"},
			Timestamp: time.Now(),
		},
		{
			ID:        "batch-002",
			Body:      []byte(`{"batch": true, "message": 2}`),
			Headers:   map[string]interface{}{"batch": "true"},
			Timestamp: time.Now(),
		},
		{
			ID:        "batch-003",
			Body:      []byte(`{"batch": true, "message": 3}`),
			Headers:   map[string]interface{}{"batch": "true"},
			Timestamp: time.Now(),
		},
	}

	if err := producer.PublishBatch(ctx, "test-topic", batchMessages, nil); err != nil {
		log.Fatalf("Erro ao enviar lote: %v", err)
	}
	fmt.Println("Lote enviado com sucesso!")

	// Aguardar um pouco para processamento
	time.Sleep(2 * time.Second)

	// Verificar métricas do producer
	producerMetrics := producer.GetMetrics()
	fmt.Printf("\nMétricas do Producer:\n")
	fmt.Printf("  Mensagens enviadas: %d\n", producerMetrics.MessagesSent)
	fmt.Printf("  Bytes enviados: %d\n", producerMetrics.BytesSent)
	fmt.Printf("  Latência média: %v\n", producerMetrics.AvgLatency)
	fmt.Printf("  Última mensagem enviada: %v\n", producerMetrics.LastSentAt)

	// Verificar métricas do provider
	providerMetrics := provider.GetMetrics()
	fmt.Printf("\nMétricas do Provider:\n")
	fmt.Printf("  Conexões ativas: %d\n", providerMetrics.ActiveConnections)
	fmt.Printf("  Producers ativos: %d\n", providerMetrics.ActiveProducers)
	fmt.Printf("  Consumers ativos: %d\n", providerMetrics.ActiveConsumers)
	fmt.Printf("  Último health check: %v\n", providerMetrics.LastHealthCheck)
	fmt.Printf("  Status do health check: %v\n", providerMetrics.HealthCheckStatus)

	fmt.Println("\nExemplo concluído!")
}
