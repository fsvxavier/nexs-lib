# Mocks para Message Queue

Esta pasta contém implementações mock (simuladas) para todos os componentes do sistema de message queue. Os mocks são essenciais para testes unitários e desenvolvimento sem dependências externas.

## Componentes Mockados

### 1. **MockMessageQueueProvider** (`mock_provider.go`)
Mock completo do `MessageQueueProvider` interface.

**Funcionalidades:**
- Simulação de conexão/desconexão
- Health checks configuráveis
- Métricas simuladas
- Criação de producers e consumers mock
- Callbacks para testes customizados

**Exemplo de uso:**
```go
// Criar mock provider
mockProvider := mocks.NewMockMessageQueueProvider(interfaces.ProviderRabbitMQ)

// Configurar callback para simular falha de conexão
mockProvider.OnConnect = func(ctx context.Context) error {
    return errors.New("connection failed")
}

// Usar em testes
err := mockProvider.Connect(context.Background())
assert.Error(t, err)
```

### 2. **MockProducer** (`mock_producer.go`)
Mock do `MessageProducer` interface.

**Funcionalidades:**
- Rastreamento de mensagens enviadas
- Simulação de falhas de envio
- Suporte a envio em lote
- Métricas de performance
- Delays configuráveis

**Exemplo de uso:**
```go
mockProducer := mocks.NewMockProducer()

// Configurar para falhar no envio
mockProducer.SetFailSendMessage(true)

// Rastrear mensagens enviadas
err := mockProducer.SendMessage(ctx, message)
sentMessages := mockProducer.GetSentMessages()
```

### 3. **MockConsumer** (`mock_consumer.go`)
Mock do `MessageConsumer` interface.

**Funcionalidades:**
- Simulação de consumo de mensagens
- Controle de subscrições
- Callbacks configuráveis
- Simulação de falhas de processamento

**Exemplo de uso:**
```go
mockConsumer := mocks.NewMockConsumer()

// Configurar callback personalizado
mockConsumer.OnSubscribe = func(ctx context.Context, destination string, options *interfaces.ConsumerOptions, handler interfaces.MessageHandler) error {
    // Simular processamento de mensagem
    return handler(ctx, testMessage)
}
```

### 4. **Componentes Auxiliares**

#### MockConnectionManager
- Gerenciamento de pool de conexões simulado
- Controle de falhas de conexão
- Métricas de conexão

#### MockLogger
- Sistema de logging para testes
- Captura de logs por nível
- Filtros de mensagens

#### MockMetricsCollector
- Coleta de métricas simulada
- Contadores e temporizadores
- Estatísticas de performance

## Padrões de Uso

### 1. **Testes de Sucesso**
```go
func TestSuccessfulMessage(t *testing.T) {
    mockProvider := mocks.NewMockMessageQueueProvider(interfaces.ProviderRabbitMQ)
    mockProducer := mocks.NewMockProducer()
    
    // Configurar para sucesso
    mockProvider.OnCreateProducer = func(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
        return mockProducer, nil
    }
    
    // Testar funcionalidade
    producer, err := mockProvider.CreateProducer(&interfaces.ProducerConfig{ID: "test"})
    assert.NoError(t, err)
    
    err = producer.SendMessage(context.Background(), testMessage)
    assert.NoError(t, err)
    assert.Equal(t, 1, mockProducer.GetSentMessageCount())
}
```

### 2. **Testes de Falha**
```go
func TestConnectionFailure(t *testing.T) {
    mockProvider := mocks.NewMockMessageQueueProvider(interfaces.ProviderKafka)
    
    // Configurar falha de conexão
    mockProvider.OnConnect = func(ctx context.Context) error {
        return errors.New("network timeout")
    }
    
    err := mockProvider.Connect(context.Background())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "network timeout")
}
```

### 3. **Testes de Performance**
```go
func TestPerformanceMetrics(t *testing.T) {
    mockMetrics := mocks.NewMockMetricsCollector()
    mockProducer := mocks.NewMockProducer()
    
    // Configurar delay
    mockProducer.SetSendDelay(100 * time.Millisecond)
    
    start := time.Now()
    err := mockProducer.SendMessage(context.Background(), testMessage)
    duration := time.Since(start)
    
    assert.NoError(t, err)
    assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}
```

### 4. **Testes de Cenários Complexos**
```go
func TestReconnectionScenario(t *testing.T) {
    mockProvider := mocks.NewMockMessageQueueProvider(interfaces.ProviderActiveMQ)
    
    // Primeira tentativa falha
    connectAttempts := 0
    mockProvider.OnConnect = func(ctx context.Context) error {
        connectAttempts++
        if connectAttempts == 1 {
            return errors.New("temporary failure")
        }
        return nil // Sucesso na segunda tentativa
    }
    
    // Primeira tentativa
    err := mockProvider.Connect(context.Background())
    assert.Error(t, err)
    
    // Segunda tentativa
    err = mockProvider.Connect(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, 2, connectAttempts)
}
```

## Configurações Avançadas

### Callbacks Disponíveis

#### Provider Callbacks
- `OnConnect`: Simular comportamento de conexão
- `OnDisconnect`: Simular desconexão
- `OnCreateProducer`: Controlar criação de producers
- `OnCreateConsumer`: Controlar criação de consumers
- `OnHealthCheck`: Simular health checks

#### Producer Callbacks
- `OnSendMessage`: Interceptar envio de mensagens
- `OnSendBatch`: Interceptar envio em lote
- `OnClose`: Simular fechamento
- `OnHealthCheck`: Simular health check

#### Consumer Callbacks
- `OnSubscribe`: Simular subscrição
- `OnUnsubscribe`: Simular cancelamento
- `OnClose`: Simular fechamento

### Configurações de Estado
```go
// Configurar falhas
mockProvider.SetFailHealthCheck(true)
mockProducer.SetFailSendMessage(true)
mockConsumer.SetFailSubscribe(true)

// Configurar delays
mockProducer.SetSendDelay(time.Second)

// Configurar limites
mockProducer.SetMaxBatchSize(50)
```

## Exemplo Completo

```go
func TestCompleteWorkflow(t *testing.T) {
    // Setup
    mockProvider := mocks.NewMockMessageQueueProvider(interfaces.ProviderRabbitMQ)
    mockProducer := mocks.NewMockProducer()
    mockConsumer := mocks.NewMockConsumer()
    mockMetrics := mocks.NewMockMetricsCollector()
    
    // Configurar comportamentos
    mockProvider.OnCreateProducer = func(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
        return mockProducer, nil
    }
    
    mockProvider.OnCreateConsumer = func(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
        return mockConsumer, nil
    }
    
    // Workflow de teste
    err := mockProvider.Connect(context.Background())
    assert.NoError(t, err)
    
    producer, err := mockProvider.CreateProducer(&interfaces.ProducerConfig{ID: "test"})
    assert.NoError(t, err)
    
    consumer, err := mockProvider.CreateConsumer(&interfaces.ConsumerConfig{ID: "test"})
    assert.NoError(t, err)
    
    // Enviar mensagem
    testMessage := &interfaces.Message{
        ID:   "test-123",
        Body: []byte("test payload"),
    }
    
    err = producer.SendMessage(context.Background(), testMessage)
    assert.NoError(t, err)
    
    // Verificar resultados
    assert.Equal(t, 1, mockProducer.GetSentMessageCount())
    sentMessages := mockProducer.GetSentMessages()
    assert.Equal(t, "test-123", sentMessages[0].ID)
    
    // Cleanup
    producer.Close()
    consumer.Close()
    mockProvider.Close()
}
```

## Benefícios dos Mocks

1. **Testes Rápidos**: Sem dependências externas
2. **Testes Determinísticos**: Comportamento controlado
3. **Cobertura Completa**: Teste de cenários de falha
4. **Desenvolvimento Isolado**: Trabalhar sem infraestrutura
5. **CI/CD Friendly**: Executar em qualquer ambiente

Os mocks seguem exatamente as mesmas interfaces dos componentes reais, garantindo que os testes sejam representativos do comportamento real do sistema.
