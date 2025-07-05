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
	fmt.Println("=== Exemplo Factory Pattern - Message Queue ===")

	// Configuração completa com múltiplos providers
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
			interfaces.ProviderKafka: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers:          []string{"localhost:9092"},
					ConnectTimeout:   10 * time.Second,
					OperationTimeout: 30 * time.Second,
				},
			},
			interfaces.ProviderActiveMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers:          []string{"localhost:61613"},
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
			LoggingEnabled:      true,
			LogLevel:            "info",
			TracingEnabled:      true,
			ServiceName:         "message-queue-example",
			MetricsEnabled:      true,
			TracingSamplingRate: 1.0,
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

	// Demonstrar uso com diferentes providers
	providers := []interfaces.ProviderType{
		interfaces.ProviderRabbitMQ,
		interfaces.ProviderKafka,
		interfaces.ProviderActiveMQ,
	}

	for _, providerType := range providers {
		fmt.Printf("=== Testando Provider: %s ===\n", providerType)

		// Verificar se o provider está disponível
		if !factory.IsProviderAvailable(providerType) {
			fmt.Printf("Provider %s não está disponível\n\n", providerType)
			continue
		}

		// Obter provider
		provider, err := factory.GetProvider(providerType)
		if err != nil {
			log.Printf("Erro ao obter provider %s: %v\n", providerType, err)
			continue
		}

		// Verificar health
		if err := provider.HealthCheck(ctx); err != nil {
			log.Printf("Health check falhou para %s: %v\n", providerType, err)
			continue
		}
		fmt.Printf("✓ Health check OK para %s\n", providerType)

		// Verificar se está conectado
		if provider.IsConnected() {
			fmt.Printf("✓ %s está conectado\n", providerType)
		} else {
			fmt.Printf("⚠ %s não está conectado\n", providerType)
			continue
		}

		// Obter métricas
		metrics := provider.GetMetrics()
		fmt.Printf("✓ Métricas do %s:\n", providerType)
		fmt.Printf("  - Conexões ativas: %d\n", metrics.ActiveConnections)
		fmt.Printf("  - Producers ativos: %d\n", metrics.ActiveProducers)
		fmt.Printf("  - Consumers ativos: %d\n", metrics.ActiveConsumers)
		fmt.Printf("  - Último health check: %v\n", metrics.LastHealthCheck)
		fmt.Printf("  - Status: %v\n", metrics.HealthCheckStatus)

		// Exemplo específico por provider
		switch providerType {
		case interfaces.ProviderRabbitMQ:
			demonstrateRabbitMQ(ctx, factory)
		case interfaces.ProviderKafka:
			demonstrateKafka(ctx, factory)
		case interfaces.ProviderActiveMQ:
			demonstrateActiveMQ(ctx, factory)
		}

		fmt.Println()
	}

	// Demonstrar alternância entre providers
	fmt.Printf("=== Demonstrando Alternância de Providers ===\n")
	demonstrateProviderSwitching(ctx, factory)

	fmt.Printf("=== Exemplo Factory Pattern concluído! ===\n")
}

func demonstrateRabbitMQ(ctx context.Context, factory messagequeue.Factory) {
	fmt.Println("📨 Demonstração RabbitMQ:")

	// Usar producer através do factory
	message := &interfaces.Message{
		ID:        "rabbitmq-demo-001",
		Body:      []byte(`{"demo": "rabbitmq", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
		Headers:   map[string]interface{}{"source": "factory-demo"},
		Timestamp: time.Now(),
	}

	// Simular envio (implementação específica dependeria da criação do producer)
	fmt.Printf("  ✓ Mensagem preparada: %s\n", message.ID)
}

func demonstrateKafka(ctx context.Context, factory messagequeue.Factory) {
	fmt.Println("📊 Demonstração Kafka:")

	// Criar mensagem para Kafka
	message := &interfaces.Message{
		ID:        "kafka-demo-001",
		Body:      []byte(`{"event": "demo", "provider": "kafka", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
		Headers:   map[string]interface{}{"partition": "0", "source": "factory-demo"},
		Timestamp: time.Now(),
	}

	fmt.Printf("  ✓ Evento preparado: %s\n", message.ID)
}

func demonstrateActiveMQ(ctx context.Context, factory messagequeue.Factory) {
	fmt.Println("🔗 Demonstração ActiveMQ:")

	// Criar mensagem para ActiveMQ
	message := &interfaces.Message{
		ID:        "activemq-demo-001",
		Body:      []byte(`{"message": "demo", "provider": "activemq", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`),
		Headers:   map[string]interface{}{"destination": "/queue/demo", "source": "factory-demo"},
		Timestamp: time.Now(),
	}

	fmt.Printf("  ✓ Mensagem preparada: %s\n", message.ID)
}

func demonstrateProviderSwitching(ctx context.Context, factory messagequeue.Factory) {
	// Simular cenário onde um provider falha e precisamos alternar
	providers := []interfaces.ProviderType{
		interfaces.ProviderRabbitMQ,
		interfaces.ProviderKafka,
		interfaces.ProviderActiveMQ,
	}

	var activeProvider interfaces.MessageQueueProvider
	var activeProviderType interfaces.ProviderType

	// Tentar encontrar um provider ativo
	for _, providerType := range providers {
		if factory.IsProviderAvailable(providerType) {
			provider, err := factory.GetProvider(providerType)
			if err == nil && provider.IsConnected() {
				activeProvider = provider
				activeProviderType = providerType
				break
			}
		}
	}

	if activeProvider != nil {
		fmt.Printf("✓ Provider ativo encontrado: %s\n", activeProviderType)

		// Simular failover
		fmt.Printf("🔄 Simulando failover...\n")

		// Buscar provider alternativo
		for _, providerType := range providers {
			if providerType != activeProviderType && factory.IsProviderAvailable(providerType) {
				provider, err := factory.GetProvider(providerType)
				if err == nil && provider.IsConnected() {
					fmt.Printf("✓ Failover para %s bem-sucedido\n", providerType)
					break
				}
			}
		}
	} else {
		fmt.Println("⚠ Nenhum provider ativo encontrado")
	}
}
