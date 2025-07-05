package main

import (
	"context"
	"fmt"
	"log"
	"time"

	messagequeue "github.com/fsvxavier/nexs-lib/message-queue"
	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

func main() {
	fmt.Println("=== Exemplo ActiveMQ - Message Queue ===")

	// Configuração específica para ActiveMQ
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider:   interfaces.ProviderActiveMQ,
			DefaultTimeout:    30 * time.Second,
			DefaultWorkers:    2,
			DefaultBufferSize: 100,
			MetricsEnabled:    true,
			TracingEnabled:    true,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderActiveMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers:          []string{"localhost:61613"}, // Porta STOMP padrão
					ConnectTimeout:   10 * time.Second,
					OperationTimeout: 30 * time.Second,
					Auth: &interfaces.AuthConfig{
						Username: "admin",
						Password: "admin",
					},
				},
			},
		},
		Observability: &config.ObservabilityConfig{
			LoggingEnabled: true,
			LogLevel:       "info",
			MetricsEnabled: true,
		},
		Idempotency: &config.IdempotencyConfig{
			Enabled:   true,
			CacheTTL:  1 * time.Hour,
			CacheSize: 10000,
		},
	}

	// Criar factory
	factory := messagequeue.NewFactory(cfg)
	defer factory.Close()

	ctx := context.Background()

	// Verificar se o ActiveMQ está disponível
	if !factory.IsProviderAvailable(interfaces.ProviderActiveMQ) {
		fmt.Println("❌ ActiveMQ não está disponível ou não configurado")
		fmt.Println("   Certifique-se de que o ActiveMQ está rodando em localhost:61613 (STOMP)")
		return
	}

	// Obter provider
	provider, err := factory.GetProvider(interfaces.ProviderActiveMQ)
	if err != nil {
		log.Printf("Erro ao obter provider ActiveMQ: %v\n", err)
		return
	}

	fmt.Printf("✓ Provider ActiveMQ obtido com sucesso\n")

	// Verificar health
	if err := provider.HealthCheck(ctx); err != nil {
		log.Printf("Health check falhou: %v\n", err)
		fmt.Println("   Isso é esperado se o ActiveMQ não estiver rodando")
		fmt.Println("   Para testar com ActiveMQ real, execute: docker run -d --name activemq -p 61613:61613 -p 8161:8161 rmohr/activemq:latest")
	} else {
		fmt.Printf("✓ Health check OK\n")
	}

	// Verificar se está conectado
	if provider.IsConnected() {
		fmt.Printf("✓ ActiveMQ está conectado\n")
	} else {
		fmt.Printf("⚠ ActiveMQ não está conectado\n")
	}

	// Obter métricas
	metrics := provider.GetMetrics()
	fmt.Printf("✓ Métricas do ActiveMQ:\n")
	fmt.Printf("  - Tempo de atividade: %v\n", metrics.Uptime)
	fmt.Printf("  - Conexões ativas: %d\n", metrics.ActiveConnections)
	fmt.Printf("  - Producers ativos: %d\n", metrics.ActiveProducers)
	fmt.Printf("  - Consumers ativos: %d\n", metrics.ActiveConsumers)
	fmt.Printf("  - Último health check: %v\n", metrics.LastHealthCheck)
	fmt.Printf("  - Status: %v\n", metrics.HealthCheckStatus)

	// Demonstrar producer (mesmo sem conexão real)
	demonstrateProducer(ctx, provider)

	// Demonstrar consumer (mesmo sem conexão real)
	demonstrateConsumer(ctx, provider)

	fmt.Printf("\n=== Exemplo ActiveMQ finalizado! ===\n")
}

func demonstrateProducer(ctx context.Context, provider interfaces.MessageQueueProvider) {
	fmt.Printf("\n📤 Demonstração Producer ActiveMQ:\n")

	// Configurar producer
	producerConfig := &interfaces.ProducerConfig{
		ID:            "activemq-producer",
		Transactional: false,
		SendTimeout:   30 * time.Second,
		BufferSize:    100,
		Compression:   "",
		ProviderConfig: map[string]interface{}{
			"destination":   "/queue/demo",
			"delivery_mode": "persistent",
			"message_type":  "text",
		},
	}

	// Tentar criar producer
	producer, err := provider.CreateProducer(producerConfig)
	if err != nil {
		log.Printf("Erro ao criar producer (esperado se ActiveMQ não estiver rodando): %v\n", err)
		fmt.Printf("  ⚠ Producer não criado devido à conexão\n")
		return
	}
	defer producer.Close()

	fmt.Printf("✓ Producer ActiveMQ criado com ID: %s\n", producerConfig.ID)

	// Criar mensagens de exemplo para ActiveMQ
	messages := []*interfaces.Message{
		{
			ID:   "activemq-msg-001",
			Body: []byte(`{"event": "order_created", "order_id": "12345", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
			Headers: map[string]interface{}{
				"event_type":   "order_created",
				"source":       "order-service",
				"destination":  "/queue/orders",
				"message_type": "application/json",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "activemq-msg-002",
			Body: []byte(`{"event": "payment_processed", "payment_id": "67890", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
			Headers: map[string]interface{}{
				"event_type":   "payment_processed",
				"source":       "payment-service",
				"destination":  "/queue/payments",
				"message_type": "application/json",
			},
			Timestamp: time.Now(),
		},
	}

	// Opções de mensagem para ActiveMQ
	options := &interfaces.MessageOptions{
		Headers: map[string]interface{}{
			"service":  "activemq-demo",
			"version":  "1.0.0",
			"protocol": "STOMP",
		},
		Priority:   5,
		Persistent: true,
		Timeout:    30 * time.Second,
	}

	// Publicar mensagens individualmente
	destination := "/queue/demo"
	for _, msg := range messages {
		err := producer.Publish(ctx, destination, msg, options)
		if err != nil {
			log.Printf("Erro ao publicar mensagem %s: %v\n", msg.ID, err)
		} else {
			fmt.Printf("  ✓ Mensagem %s publicada para %s\n", msg.ID, destination)
		}
	}

	// Demonstrar publicação em lote
	batchMessages := []*interfaces.Message{
		{
			ID:   "activemq-batch-001",
			Body: []byte(`{"event": "batch_message_1", "type": "bulk_operation"}`),
			Headers: map[string]interface{}{
				"batch":       "true",
				"batch_id":    "batch-001",
				"destination": "/queue/bulk",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "activemq-batch-002",
			Body: []byte(`{"event": "batch_message_2", "type": "bulk_operation"}`),
			Headers: map[string]interface{}{
				"batch":       "true",
				"batch_id":    "batch-001",
				"destination": "/queue/bulk",
			},
			Timestamp: time.Now(),
		},
	}

	err = producer.PublishBatch(ctx, "/queue/bulk", batchMessages, options)
	if err != nil {
		log.Printf("Erro ao publicar lote: %v\n", err)
	} else {
		fmt.Printf("  ✓ Lote de %d mensagens publicado para /queue/bulk\n", len(batchMessages))
	}
}

func demonstrateConsumer(ctx context.Context, provider interfaces.MessageQueueProvider) {
	fmt.Printf("\n📥 Demonstração Consumer ActiveMQ:\n")

	// Configurar consumer
	consumerConfig := &interfaces.ConsumerConfig{
		ID:             "activemq-consumer",
		ConsumerGroup:  "demo-consumer-group",
		AutoCommit:     true,
		CommitInterval: 5 * time.Second,
		ProviderConfig: map[string]interface{}{
			"destination":        "/queue/demo",
			"subscription_type":  "queue",
			"ack_mode":           "auto",
			"worker_count":       2,
			"buffer_size":        100,
			"processing_timeout": 30 * time.Second,
		},
	}

	// Tentar criar consumer
	consumer, err := provider.CreateConsumer(consumerConfig)
	if err != nil {
		log.Printf("Erro ao criar consumer (esperado se ActiveMQ não estiver rodando): %v\n", err)
		fmt.Printf("  ⚠ Consumer não criado devido à conexão\n")
		return
	}
	defer consumer.Close()

	fmt.Printf("✓ Consumer ActiveMQ criado com ID: %s\n", consumerConfig.ID)
	fmt.Printf("  - Grupo: %s\n", consumerConfig.ConsumerGroup)

	// Opções de consumo para ActiveMQ
	consumerOptions := &interfaces.ConsumerOptions{
		ConsumerGroup:     "demo-consumer-group",
		Workers:           2,
		BufferSize:        100,
		ProcessingTimeout: 30 * time.Second,
		AutoAck:           true,
		BatchSize:         10,
		BatchInterval:     5 * time.Second,
	}

	// Definir handler para processar mensagens (demonstração)
	handler := func(ctx context.Context, msg *interfaces.Message) error {
		fmt.Printf("  📨 Processando mensagem ActiveMQ %s: %s\n", msg.ID, string(msg.Body))

		// Simular processamento específico para ActiveMQ
		if destination, ok := msg.Headers["destination"]; ok {
			fmt.Printf("      Destino: %v\n", destination)
		}
		if msgType, ok := msg.Headers["message_type"]; ok {
			fmt.Printf("      Tipo: %v\n", msgType)
		}

		// Simular processamento
		time.Sleep(100 * time.Millisecond)

		return nil
	}

	fmt.Printf("✓ Handler ActiveMQ definido. Em uma aplicação real, você chamaria:\n")
	fmt.Printf("  consumer.Subscribe(ctx, \"/queue/demo\", consumerOptions, handler)\n")
	fmt.Printf("  (O consumer ficaria escutando mensagens continuamente)\n")

	// Demonstrar configuração de diferentes tipos de destination
	fmt.Printf("\n📋 Tipos de Destination ActiveMQ suportados:\n")
	fmt.Printf("  - Queue: /queue/nome-da-fila\n")
	fmt.Printf("  - Topic: /topic/nome-do-topico\n")
	fmt.Printf("  - Temporary Queue: /temp-queue/session-id\n")
	fmt.Printf("  - Temporary Topic: /temp-topic/session-id\n")

	// Não executamos o Subscribe real para não bloquear o exemplo
	_ = handler
	_ = consumerOptions
}
