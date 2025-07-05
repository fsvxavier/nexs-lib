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
	fmt.Println("=== Exemplo Completo - Message Queue ===")

	// Configuração mínima para demonstração
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider:   interfaces.ProviderRabbitMQ,
			DefaultTimeout:    30 * time.Second,
			DefaultWorkers:    2,
			DefaultBufferSize: 100,
			MetricsEnabled:    true,
			TracingEnabled:    true,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderRabbitMQ: {
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

	// Verificar se o RabbitMQ está disponível
	if !factory.IsProviderAvailable(interfaces.ProviderRabbitMQ) {
		fmt.Println("❌ RabbitMQ não está disponível ou não configurado")
		fmt.Println("   Certifique-se de que o RabbitMQ está rodando em localhost:5672")
		return
	}

	// Obter provider
	provider, err := factory.GetProvider(interfaces.ProviderRabbitMQ)
	if err != nil {
		log.Printf("Erro ao obter provider RabbitMQ: %v\n", err)
		return
	}

	fmt.Printf("✓ Provider RabbitMQ obtido com sucesso\n")

	// Verificar health
	if err := provider.HealthCheck(ctx); err != nil {
		log.Printf("Health check falhou: %v\n", err)
		fmt.Println("   Isso é esperado se o RabbitMQ não estiver rodando")
		fmt.Println("   Para testar com RabbitMQ real, execute: docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management")
	} else {
		fmt.Printf("✓ Health check OK\n")
	}

	// Demonstrar criação de producer
	demonstrateProducer(ctx, provider)

	// Demonstrar criação de consumer
	demonstrateConsumer(ctx, provider)

	// Mostrar métricas
	demonstrateMetrics(provider)

	fmt.Printf("\n=== Exemplo completo finalizado! ===\n")
}

func demonstrateProducer(ctx context.Context, provider interfaces.MessageQueueProvider) {
	fmt.Printf("\n📤 Demonstração Producer:\n")

	// Configurar producer
	producerConfig := &interfaces.ProducerConfig{
		ID:            "demo-producer",
		Transactional: false,
		SendTimeout:   30 * time.Second,
		BufferSize:    100,
		Compression:   "gzip",
		ProviderConfig: map[string]interface{}{
			"topic":     "demo.topic",
			"partition": "auto",
		},
	}

	// Criar producer
	producer, err := provider.CreateProducer(producerConfig)
	if err != nil {
		log.Printf("Erro ao criar producer: %v\n", err)
		return
	}
	defer producer.Close()

	fmt.Printf("✓ Producer criado com ID: %s\n", producerConfig.ID)

	// Criar mensagens de exemplo
	messages := []*interfaces.Message{
		{
			ID:   "msg-001",
			Body: []byte(`{"event": "user_created", "user_id": "123", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
			Headers: map[string]interface{}{
				"event_type": "user_created",
				"source":     "user-service",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "msg-002",
			Body: []byte(`{"event": "order_placed", "order_id": "456", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
			Headers: map[string]interface{}{
				"event_type": "order_placed",
				"source":     "order-service",
			},
			Timestamp: time.Now(),
		},
	}

	// Opções de mensagem
	options := &interfaces.MessageOptions{
		Headers: map[string]interface{}{
			"service": "demo-app",
			"version": "1.0.0",
		},
		Priority:   5,
		Persistent: true,
		Timeout:    30 * time.Second,
	}

	// Publicar mensagens individualmente
	destination := "demo.topic"
	for _, msg := range messages {
		err := producer.Publish(ctx, destination, msg, options)
		if err != nil {
			log.Printf("Erro ao publicar mensagem %s: %v\n", msg.ID, err)
		} else {
			fmt.Printf("  ✓ Mensagem %s publicada com sucesso\n", msg.ID)
		}
	}

	// Demonstrar publicação em lote
	batchMessages := []*interfaces.Message{
		{
			ID:   "batch-001",
			Body: []byte(`{"event": "batch_message_1"}`),
			Headers: map[string]interface{}{
				"batch": "true",
			},
			Timestamp: time.Now(),
		},
		{
			ID:   "batch-002",
			Body: []byte(`{"event": "batch_message_2"}`),
			Headers: map[string]interface{}{
				"batch": "true",
			},
			Timestamp: time.Now(),
		},
	}

	err = producer.PublishBatch(ctx, destination, batchMessages, options)
	if err != nil {
		log.Printf("Erro ao publicar lote: %v\n", err)
	} else {
		fmt.Printf("  ✓ Lote de %d mensagens publicado com sucesso\n", len(batchMessages))
	}
}

func demonstrateConsumer(ctx context.Context, provider interfaces.MessageQueueProvider) {
	fmt.Printf("\n📥 Demonstração Consumer:\n")

	// Configurar consumer
	consumerConfig := &interfaces.ConsumerConfig{
		ID:             "demo-consumer",
		ConsumerGroup:  "demo-consumer-group",
		AutoCommit:     true,
		CommitInterval: 5 * time.Second,
		ProviderConfig: map[string]interface{}{
			"topic":              "demo.topic",
			"worker_count":       2,
			"buffer_size":        100,
			"processing_timeout": 30 * time.Second,
		},
	}

	// Criar consumer
	consumer, err := provider.CreateConsumer(consumerConfig)
	if err != nil {
		log.Printf("Erro ao criar consumer: %v\n", err)
		return
	}
	defer consumer.Close()

	fmt.Printf("✓ Consumer criado com ID: %s\n", consumerConfig.ID)
	fmt.Printf("  - Grupo: %s\n", consumerConfig.ConsumerGroup)

	// Opções de consumo (demonstração)
	_ = &interfaces.ConsumerOptions{
		ConsumerGroup:     "demo-consumer-group",
		Workers:           2,
		BufferSize:        100,
		ProcessingTimeout: 30 * time.Second,
		AutoAck:           true,
	}

	// Definir handler para processar mensagens (demonstração)
	_ = func(ctx context.Context, msg *interfaces.Message) error {
		fmt.Printf("  📨 Processando mensagem %s: %s\n", msg.ID, string(msg.Body))

		// Simular processamento
		time.Sleep(100 * time.Millisecond)

		// Retornar sucesso
		return nil
	}

	// Demonstração de subscribe (não executado realmente para não bloquear)
	fmt.Printf("✓ Handler definido. Em uma aplicação real, você chamaria:\n")
	fmt.Printf("  consumer.Subscribe(ctx, \"demo.topic\", consumerOptions, handler)\n")
	fmt.Printf("  (O consumer ficaria rodando continuamente)\n")
}

func demonstrateMetrics(provider interfaces.MessageQueueProvider) {
	fmt.Printf("\n📊 Métricas do Provider:\n")

	metrics := provider.GetMetrics()
	fmt.Printf("  - Tempo de atividade: %v\n", metrics.Uptime)
	fmt.Printf("  - Conexões ativas: %d\n", metrics.ActiveConnections)
	fmt.Printf("  - Producers ativos: %d\n", metrics.ActiveProducers)
	fmt.Printf("  - Consumers ativos: %d\n", metrics.ActiveConsumers)
	fmt.Printf("  - Último health check: %v\n", metrics.LastHealthCheck)
	fmt.Printf("  - Status health check: %v\n", metrics.HealthCheckStatus)

	if metrics.ConnectionStats != nil {
		fmt.Printf("  - Total de conexões: %d\n", metrics.ConnectionStats.TotalConnections)
		fmt.Printf("  - Conexões falharam: %d\n", metrics.ConnectionStats.FailedConnections)
		fmt.Printf("  - Reconexões: %d\n", metrics.ConnectionStats.Reconnections)
		fmt.Printf("  - Última conexão: %v\n", metrics.ConnectionStats.LastConnectedAt)
	}
}
