# 🚀 Message Queue - Nexs Lib

[![Go Version](https://img.shields.io/badge/go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/badge/tests-70%20passing-brightgreen.svg)](./tests)
[![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen.svg)](./coverage)

Um módulo **completo** e **performático** para gerenciamento de filas de mensagem em Go, com suporte a múltiplos providers, sistema de retry avançado, idempotência e observabilidade integrada.

## ✨ Características Principais

- 🏗️ **Arquitetura Modular** - Factory Pattern com providers intercambiáveis
- 📡 **4 Providers Suportados** - RabbitMQ, Kafka, SQS, ActiveMQ
- 🔄 **Sistema de Retry Avançado** - Políticas exponencial, linear e customizáveis
- 🛡️ **Idempotência Built-in** - Prevenção automática contra processamento duplicado
- 📊 **Observabilidade Completa** - Métricas, health checks e tracing
- ⚡ **Alta Performance** - Otimizado para throughput e baixa latência
- 🧪 **100% Testável** - Mocks específicos e cobertura completa
- 📚 **Documentação Rica** - Exemplos práticos e guias detalhados

## 🚀 Instalação Rápida

```bash
go get github.com/fsvxavier/nexs-lib/message-queue
```

### Dependências dos Providers

```bash
# Apache Kafka
go get github.com/Shopify/sarama

# RabbitMQ
go get github.com/rabbitmq/amqp091-go

# Apache ActiveMQ (STOMP)
go get github.com/go-stomp/stomp/v3
```

## 🎯 Uso Básico

### Producer - Enviar Mensagens

```go
package main

import (
    "context"
    "time"
    
    mq "github.com/fsvxavier/nexs-lib/message-queue"
    "github.com/fsvxavier/nexs-lib/message-queue/config"
    "github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

func main() {
    // Configuração do RabbitMQ
    cfg := &config.Config{
        Global: &config.GlobalConfig{
            DefaultProvider: interfaces.ProviderRabbitMQ,
            DefaultTimeout:  30 * time.Second,
        },
        Providers: map[interfaces.ProviderType]*config.ProviderConfig{
            interfaces.ProviderRabbitMQ: {
                Enabled: true,
                Connection: &interfaces.ConnectionConfig{
                    Brokers: []string{"localhost:5672"},
                    Auth: &interfaces.AuthConfig{
                        Username: "guest",
                        Password: "guest",
                    },
                },
            },
        },
    }

    // Criar factory e provider
    factory := mq.NewFactory(cfg)
    defer factory.Close()

    provider, _ := factory.GetProvider(interfaces.ProviderRabbitMQ)
    
    // Criar producer
    producer, _ := provider.CreateProducer(&interfaces.ProducerConfig{
        ID: "order-producer",
    })
    defer producer.Close()

    // Enviar mensagem
    message := &interfaces.Message{
        ID:   "order-12345",
        Body: []byte(`{"order_id": 12345, "amount": 99.90}`),
        Headers: map[string]interface{}{
            "content-type": "application/json",
            "version":      "v1",
        },
    }

    err := producer.Publish(context.Background(), "orders.new", message, nil)
    if err != nil {
        panic(err)
    }
}
```

### Consumer - Processar Mensagens

```go
// Criar consumer
consumer, _ := provider.CreateConsumer(&interfaces.ConsumerConfig{
    ID:            "order-processor",
    ConsumerGroup: "order-processing-group",
})
defer consumer.Close()

// Handler para processar mensagens
handler := func(ctx context.Context, msg *interfaces.Message) error {
    // Processar pedido
    fmt.Printf("Processando pedido: %s\n", string(msg.Body))
    
    // Simular processamento
    time.Sleep(100 * time.Millisecond)
    
    return nil // Sucesso - ACK automático
}

// Configurar opções de consumo
options := &interfaces.ConsumerOptions{
    Workers:           5,               // 5 workers paralelos
    BufferSize:        100,             // Buffer de 100 mensagens
    ProcessingTimeout: 30 * time.Second, // Timeout por mensagem
    AutoAck:          true,             // ACK automático em caso de sucesso
}

// Iniciar consumo
err := consumer.Subscribe(context.Background(), "orders.new", options, handler)
```

## 🏗️ Arquitetura

```
message-queue/
├── 📁 config/              # Configurações centralizadas
├── 📁 interfaces/          # Contratos e tipos (Message, Producer, Consumer, Provider)
├── 📁 providers/           # Implementações específicas
│   ├── 📁 activemq/       # Apache ActiveMQ (STOMP)
│   ├── 📁 kafka/          # Apache Kafka (Sarama)
│   ├── 📁 rabbitmq/       # RabbitMQ (AMQP)
│   └── 📁 sqs/            # Amazon SQS
├── 📁 commons/            # Utilitários (idempotência, métricas)
├── 📁 internal/retry/     # Sistema de retry avançado
├── 📁 mocks/              # Mocks genéricos
├── 📁 examples/           # Exemplos práticos executáveis
└── 📄 factory.go          # Factory Pattern principal
```

## 🔧 Funcionalidades Avançadas

### 🔄 Sistema de Retry

```go
import "github.com/fsvxavier/nexs-lib/message-queue/internal/retry"

// Política exponencial: 100ms → 200ms → 400ms → 800ms → 1.6s
retryPolicy := retry.ExponentialBackoffPolicy(5, 100*time.Millisecond, 2.0)

retryer := retry.NewRetryer(retryPolicy)

err := retryer.Execute(ctx, func() error {
    return producer.Publish(ctx, "orders", message, nil)
})
```

### 🛡️ Idempotência

```go
import "github.com/fsvxavier/nexs-lib/message-queue/commons"

// Cache em memória com TTL de 1 hora
idempotency := commons.NewMemoryIdempotencyManager(1 * time.Hour)

handler := func(ctx context.Context, msg *interfaces.Message) error {
    // Verificar se já foi processado
    if idempotency.IsProcessed(ctx, msg.ID) {
        return nil // Já processado - skip
    }
    
    // Processar mensagem
    if err := processOrder(msg); err != nil {
        return err
    }
    
    // Marcar como processado
    idempotency.MarkAsProcessed(ctx, msg.ID)
    return nil
}
```

### 📊 Observabilidade e Métricas

```go
// Métricas do Producer
metrics := producer.GetMetrics()
fmt.Printf("Mensagens enviadas: %d\n", metrics.MessagesSent)
fmt.Printf("Latência média: %v\n", metrics.AvgLatency)
fmt.Printf("Taxa: %.2f msg/s\n", metrics.MessagesPerSecond)

// Métricas do Provider
providerMetrics := provider.GetMetrics()
fmt.Printf("Conexões ativas: %d\n", providerMetrics.ActiveConnections)
fmt.Printf("Status: %t\n", providerMetrics.HealthCheckStatus)

// Health Check
if err := provider.HealthCheck(ctx); err != nil {
    log.Printf("Provider não saudável: %v", err)
}
```

## � Providers Suportados

| Provider | Status | Latência Típica | Throughput | Recursos |
|----------|--------|-----------------|------------|----------|
| **RabbitMQ** | ✅ Completo | ~100ms | 75 msg/s | Queues, Exchanges, Routing Keys |
| **Apache Kafka** | ✅ Completo | ~100ms | 100 msg/s | Topics, Partitions, Consumer Groups |
| **Amazon SQS** | ✅ Completo | ~50ms | 60 msg/s | Standard/FIFO Queues, DLQ |
| **Apache ActiveMQ** | ✅ Completo | ~80ms | 70 msg/s | STOMP Protocol, Queues, Topics |

## 🧪 Testes e Mocks

### Executar Testes

```bash
# Todos os testes (70 testes, 100% passando)
go test -v ./...

# Testes com cobertura
go test -cover ./...

# Benchmarks de performance
go test -bench=. -benchmem ./commons ./internal/retry
```

### Usando Mocks em Testes

```go
import "github.com/fsvxavier/nexs-lib/message-queue/providers/kafka/mock"

func TestOrderProcessing(t *testing.T) {
    // Mock específico do Kafka
    mockProvider := mock.NewMockKafkaProvider()
    
    // Configurar comportamento
    mockProvider.ConnectFunc = func(ctx context.Context) error {
        return nil // Simular sucesso
    }
    
    // Testar sua lógica
    err := mockProvider.Connect(context.Background())
    assert.NoError(t, err)
    assert.True(t, mockProvider.IsConnected())
}
```

## 📖 Exemplos Executáveis

```bash
# Exemplo completo com todos os providers
cd examples/complete && go run main.go

# Exemplos específicos por provider
cd examples/rabbitmq && go run main.go
cd examples/kafka && go run main.go
cd examples/sqs && go run main.go
cd examples/activemq && go run main.go

# Demonstração dos mocks
cd examples/mocks && go run main.go
```

## 🐳 Configuração com Docker

```bash
# RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management

# Apache Kafka
docker run -d --name kafka -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  confluentinc/cp-kafka:latest

# Apache ActiveMQ
docker run -d --name activemq -p 61613:61613 -p 8161:8161 \
  rmohr/activemq:latest
```

## � Performance Benchmarks

**Ambiente**: Intel i7-10750H, 16GB RAM, Go 1.21

| Componente | Operação | Tempo | Memória | Throughput |
|------------|----------|-------|---------|------------|
| **Idempotência** | IsProcessed | ~160 ns/op | 142 B/op | - |
| **Idempotência** | MarkAsProcessed | ~673 ns/op | 207 B/op | - |
| **Retry** | Política Creation | ~0.3 ns/op | 0 allocs | - |
| **Retry** | Execute Success | ~6.4 ns/op | 0 allocs | - |
| **Kafka Mock** | Producer | ~100ms | - | 100 msg/s |
| **RabbitMQ Mock** | Producer | ~100ms | - | 75 msg/s |

## � Documentação Adicional

- 📄 [**Mocks Específicos**](./MOCKS_ESPECIFICOS.md) - Guia completo dos mocks por provider
- 📄 [**Mocks Gerais**](./mocks/README.md) - Sistema de mocks genéricos
- 📄 [**Changelog**](./CHANGELOG.md) - Histórico de versões e mudanças
- 📄 [**Status Final**](./STATUS_FINAL.md) - Resumo das implementações
- 📁 [**Exemplos**](./examples/) - Códigos de exemplo executáveis

## 🤝 Contribuindo

1. **Fork** o projeto
2. **Crie** uma branch (`git checkout -b feature/amazing-feature`)
3. **Commit** suas mudanças (`git commit -m 'Add amazing feature'`)
4. **Push** para a branch (`git push origin feature/amazing-feature`)
5. **Abra** um Pull Request

## � Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](../LICENSE) para detalhes.

## 🔗 Links Úteis

- 🏠 [**nexs-lib**](../) - Biblioteca principal
- 📖 [**Documentação Completa**](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/message-queue)
- 🐛 [**Issues**](https://github.com/fsvxavier/nexs-lib/issues)
- 💡 [**Discussões**](https://github.com/fsvxavier/nexs-lib/discussions)

---

<div align="center">

**Feito com ❤️ pela equipe nexs-lib**

[![Go](https://img.shields.io/badge/Made%20with-Go-blue.svg)](https://golang.org/)
[![Contributors](https://img.shields.io/badge/Contributors-Welcome-brightgreen.svg)](./CONTRIBUTING.md)

</div>
