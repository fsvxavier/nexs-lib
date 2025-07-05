# Exemplos de Uso dos Mocks

Este diretório contém exemplos práticos de como usar os mocks específicos de cada provider do sistema de message queue.

## 📁 Arquivos

- **`main.go`**: Exemplo completo demonstrando o uso de todos os providers mock

## 🚀 Como Executar

```bash
# A partir do diretório raiz do projeto
cd message-queue/examples/mocks
go run main.go
```

## 📖 O que o Exemplo Demonstra

### 1. **Uso Básico dos Mocks**
- Criação de providers mock para ActiveMQ, Kafka, RabbitMQ e SQS
- Conexão e desconexão
- Criação de producers e consumers
- Envio e recebimento de mensagens

### 2. **Funcionalidades Específicas**
- **ActiveMQ**: Envio de mensagem individual com headers
- **Kafka**: Envio em lote (batch publishing)
- **RabbitMQ**: Subscription e acknowledgment
- **SQS**: Health checks e métricas do provider

### 3. **Cenários de Falha**
- Simulação de falhas de conexão
- Simulação de falhas de health check
- Recuperação após falhas

### 4. **Métricas e Observabilidade**
- Coleta de métricas de producer
- Métricas de provider (conexões, status)
- Estatísticas de performance

## 🎯 Saída Esperada

Quando executado, o exemplo produzirá uma saída similar a:

```
🎯 Exemplos de Uso dos Mocks Específicos
=====================================

🔥 ActiveMQ Mock Example
========================
✅ Provider conectado com sucesso
✅ Producer criado com sucesso
✅ Mensagem enviada com sucesso
📊 Mensagens enviadas: 1
✅ Consumer criado com sucesso

📨 Kafka Mock Example
=====================
✅ Kafka provider conectado
✅ Batch de 3 mensagens enviado
📊 Total de mensagens: 3

🐰 RabbitMQ Mock Example
========================
✅ RabbitMQ provider conectado
✅ Subscription ativa
📩 Mensagem recebida: test message
✅ ACK enviado com sucesso

☁️ SQS Mock Example
==================
✅ SQS provider conectado
✅ Health check passou
📊 Uptime: 1h0m0s
📊 Conexões ativas: 1

❌ Teste de Cenários de Falha
=============================
✅ Falha simulada capturada: simulação de falha de rede
✅ Reconexão bem-sucedida após falha
✅ Falha de health check simulada: serviço indisponível

📈 Demonstração de Métricas
===========================
📊 Producer Metrics:
   - Mensagens enviadas: 5
   - Bytes enviados: 75
   - Latência média: 2ms
📊 Provider Metrics:
   - Status da conexão: true
   - Producers ativos: 1
   - Consumers ativos: 0
   - Última verificação: 2024-01-01T12:00:00Z
📊 Connection Stats:
   - Total de conexões: 1
   - Conexões falharam: 0
   - Tempo médio de conexão: 100ms
✅ Demonstração de métricas concluída
```

## 🧪 Uso em Testes

Os padrões demonstrados neste exemplo podem ser aplicados em testes unitários:

```go
func TestMessageProcessing(t *testing.T) {
    // Setup mock
    provider := activemqMock.NewMockActiveMQProvider()
    
    ctx := context.Background()
    err := provider.Connect(ctx)
    require.NoError(t, err)
    
    // Test your logic here
    producer, err := provider.CreateProducer(&interfaces.ProducerConfig{
        ID: "test-producer",
    })
    require.NoError(t, err)
    
    message := &interfaces.Message{
        ID:   "test-msg",
        Body: []byte("test"),
    }
    
    err = producer.Publish(ctx, "test.queue", message, nil)
    assert.NoError(t, err)
    
    // Verify metrics
    metrics := producer.GetMetrics()
    assert.Equal(t, int64(1), metrics.MessagesSent)
    
    // Cleanup
    producer.Close()
    provider.Close()
}
```

## 🔗 Documentação Relacionada

- [Mocks Específicos](../../MOCKS_ESPECIFICOS.md) - Documentação completa dos mocks
- [Mocks Gerais](../../mocks/README.md) - Documentação dos mocks genéricos
- [Message Queue](../../README.md) - Documentação principal do módulo

## 💡 Dicas de Uso

1. **Thread Safety**: Todos os mocks são thread-safe e podem ser usados em testes concorrentes
2. **Estado Limpo**: Sempre chame `Close()` para limpar o estado entre testes
3. **Configuração de Falhas**: Use as funções `Func` para simular cenários específicos
4. **Métricas Realistas**: Os mocks retornam métricas que simulam comportamento real
5. **Provider Específico**: Cada provider mock tem características únicas (latência, throughput)
