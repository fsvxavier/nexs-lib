package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	activemqMock "github.com/fsvxavier/nexs-lib/message-queue/providers/activemq/mock"
	kafkaMock "github.com/fsvxavier/nexs-lib/message-queue/providers/kafka/mock"
	rabbitmqMock "github.com/fsvxavier/nexs-lib/message-queue/providers/rabbitmq/mock"
	sqsMock "github.com/fsvxavier/nexs-lib/message-queue/providers/sqs/mock"
)

func main() {
	fmt.Println("🎯 Exemplos de Uso dos Mocks Específicos")
	fmt.Println("=====================================")

	// Demonstrar uso de cada provider mock
	demonstrateActiveMQMock()
	demonstrateKafkaMock()
	demonstrateRabbitMQMock()
	demonstrateSQSMock()

	// Demonstrar cenários de falha
	demonstrateFailureScenarios()

	// Demonstrar métricas
	demonstrateMetrics()
}

func demonstrateActiveMQMock() {
	fmt.Println("\n🔥 ActiveMQ Mock Example")
	fmt.Println("========================")

	// Criar provider mock
	provider := activemqMock.NewMockActiveMQProvider()

	// Conectar
	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	fmt.Println("✅ Provider conectado com sucesso")

	// Criar producer
	producer, err := provider.CreateProducer(&interfaces.ProducerConfig{
		ID: "activemq-producer-1",
	})
	if err != nil {
		log.Printf("Erro ao criar producer: %v", err)
		return
	}
	fmt.Println("✅ Producer criado com sucesso")

	// Enviar mensagem
	message := &interfaces.Message{
		ID:   "msg-001",
		Body: []byte("Hello from ActiveMQ Mock!"),
		Headers: map[string]interface{}{
			"source": "mock-example",
		},
	}

	if err := producer.Publish(ctx, "test.queue", message, nil); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
		return
	}
	fmt.Println("✅ Mensagem enviada com sucesso")

	// Verificar métricas
	metrics := producer.GetMetrics()
	fmt.Printf("📊 Mensagens enviadas: %d\n", metrics.MessagesSent)

	// Criar consumer
	consumer, err := provider.CreateConsumer(&interfaces.ConsumerConfig{
		ID: "activemq-consumer-1",
	})
	if err != nil {
		log.Printf("Erro ao criar consumer: %v", err)
		return
	}
	fmt.Println("✅ Consumer criado com sucesso")

	// Cleanup
	producer.Close()
	consumer.Close()
	provider.Close()
}

func demonstrateKafkaMock() {
	fmt.Println("\n📨 Kafka Mock Example")
	fmt.Println("=====================")

	provider := kafkaMock.NewMockKafkaProvider()

	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	fmt.Println("✅ Kafka provider conectado")

	// Testar batch publishing
	producer, err := provider.CreateProducer(&interfaces.ProducerConfig{
		ID: "kafka-producer-1",
	})
	if err != nil {
		log.Printf("Erro ao criar producer: %v", err)
		return
	}

	// Enviar batch de mensagens
	messages := []*interfaces.Message{
		{ID: "batch-1", Body: []byte("Message 1")},
		{ID: "batch-2", Body: []byte("Message 2")},
		{ID: "batch-3", Body: []byte("Message 3")},
	}

	if err := producer.PublishBatch(ctx, "test.topic", messages, nil); err != nil {
		log.Printf("Erro ao enviar batch: %v", err)
		return
	}
	fmt.Printf("✅ Batch de %d mensagens enviado\n", len(messages))

	metrics := producer.GetMetrics()
	fmt.Printf("📊 Total de mensagens: %d\n", metrics.MessagesSent)

	producer.Close()
	provider.Close()
}

func demonstrateRabbitMQMock() {
	fmt.Println("\n🐰 RabbitMQ Mock Example")
	fmt.Println("========================")

	provider := rabbitmqMock.NewMockRabbitMQProvider()

	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	fmt.Println("✅ RabbitMQ provider conectado")

	// Criar consumer e testar subscription
	consumer, err := provider.CreateConsumer(&interfaces.ConsumerConfig{
		ID: "rabbitmq-consumer-1",
	})
	if err != nil {
		log.Printf("Erro ao criar consumer: %v", err)
		return
	}

	// Handler de mensagem
	messageHandler := func(ctx context.Context, message *interfaces.Message) error {
		fmt.Printf("📩 Mensagem recebida: %s\n", string(message.Body))
		return nil
	}

	// Fazer subscription
	if err := consumer.Subscribe(ctx, "test.queue", nil, messageHandler); err != nil {
		log.Printf("Erro na subscription: %v", err)
		return
	}
	fmt.Println("✅ Subscription ativa")

	// Simular ACK
	testMessage := &interfaces.Message{ID: "test-ack"}
	if err := consumer.Ack(testMessage); err != nil {
		log.Printf("Erro no ACK: %v", err)
		return
	}
	fmt.Println("✅ ACK enviado com sucesso")

	consumer.Close()
	provider.Close()
}

func demonstrateSQSMock() {
	fmt.Println("\n☁️ SQS Mock Example")
	fmt.Println("==================")

	provider := sqsMock.NewMockSQSProvider()

	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	fmt.Println("✅ SQS provider conectado")

	// Verificar health
	if err := provider.HealthCheck(ctx); err != nil {
		log.Printf("Health check falhou: %v", err)
		return
	}
	fmt.Println("✅ Health check passou")

	// Obter métricas do provider
	metrics := provider.GetMetrics()
	fmt.Printf("📊 Uptime: %v\n", metrics.Uptime)
	fmt.Printf("📊 Conexões ativas: %d\n", metrics.ActiveConnections)

	provider.Close()
}

func demonstrateFailureScenarios() {
	fmt.Println("\n❌ Teste de Cenários de Falha")
	fmt.Println("=============================")

	provider := activemqMock.NewMockActiveMQProvider()

	// Configurar falha de conexão
	provider.ConnectFunc = func(ctx context.Context) error {
		return fmt.Errorf("simulação de falha de rede")
	}

	ctx := context.Background()
	if err := provider.Connect(ctx); err != nil {
		fmt.Printf("✅ Falha simulada capturada: %v\n", err)
	}

	// Resetar para funcionamento normal
	provider.ConnectFunc = nil
	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro inesperado: %v", err)
		return
	}
	fmt.Println("✅ Reconexão bem-sucedida após falha")

	// Testar falha de health check
	provider.HealthCheckFunc = func(ctx context.Context) error {
		return fmt.Errorf("serviço indisponível")
	}

	if err := provider.HealthCheck(ctx); err != nil {
		fmt.Printf("✅ Falha de health check simulada: %v\n", err)
	}

	provider.Close()
}

func demonstrateMetrics() {
	fmt.Println("\n📈 Demonstração de Métricas")
	fmt.Println("===========================")

	provider := kafkaMock.NewMockKafkaProvider()
	ctx := context.Background()

	if err := provider.Connect(ctx); err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}

	producer, err := provider.CreateProducer(&interfaces.ProducerConfig{
		ID: "metrics-producer",
	})
	if err != nil {
		log.Printf("Erro ao criar producer: %v", err)
		return
	}

	// Enviar várias mensagens para gerar métricas
	for i := 0; i < 5; i++ {
		message := &interfaces.Message{
			ID:   fmt.Sprintf("metric-msg-%d", i+1),
			Body: []byte(fmt.Sprintf("Mensagem de teste %d", i+1)),
		}

		if err := producer.Publish(ctx, "metrics.topic", message, nil); err != nil {
			log.Printf("Erro ao enviar mensagem %d: %v", i+1, err)
			continue
		}

		// Pequeno delay para simular throughput realista
		time.Sleep(10 * time.Millisecond)
	}

	// Obter métricas detalhadas
	producerMetrics := producer.GetMetrics()
	providerMetrics := provider.GetMetrics()

	fmt.Printf("📊 Producer Metrics:\n")
	fmt.Printf("   - Mensagens enviadas: %d\n", producerMetrics.MessagesSent)
	fmt.Printf("   - Bytes enviados: %d\n", producerMetrics.BytesSent)
	fmt.Printf("   - Latência média: %v\n", producerMetrics.AvgLatency)

	fmt.Printf("📊 Provider Metrics:\n")
	fmt.Printf("   - Status da conexão: %t\n", providerMetrics.HealthCheckStatus)
	fmt.Printf("   - Producers ativos: %d\n", providerMetrics.ActiveProducers)
	fmt.Printf("   - Consumers ativos: %d\n", providerMetrics.ActiveConsumers)
	fmt.Printf("   - Última verificação: %v\n", providerMetrics.LastHealthCheck.Format(time.RFC3339))

	if providerMetrics.ConnectionStats != nil {
		fmt.Printf("📊 Connection Stats:\n")
		fmt.Printf("   - Total de conexões: %d\n", providerMetrics.ConnectionStats.TotalConnections)
		fmt.Printf("   - Conexões falharam: %d\n", providerMetrics.ConnectionStats.FailedConnections)
		fmt.Printf("   - Tempo médio de conexão: %v\n", providerMetrics.ConnectionStats.AvgConnectionTime)
	}

	producer.Close()
	provider.Close()
	fmt.Println("✅ Demonstração de métricas concluída")
}
