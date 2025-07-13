# Enhanced Error Handling Example

Este exemplo demonstra as capacidades avançadas de tratamento de erro implementadas no Nexs-Lib v2 Tracer, incluindo circuit breaker, retry com exponential backoff e classificação inteligente de erros.

## 🎯 Recursos Demonstrados

### 1. Circuit Breaker Pattern
- Proteção contra falhas em cascata
- Estados: CLOSED → OPEN → HALF_OPEN
- Configuração de threshold de falhas
- Timeout para reset automático

### 2. Retry com Exponential Backoff
- Retry automático para erros transitórios
- Exponential backoff com jitter opcional
- Classificação inteligente de erros
- Respect por context cancellation

### 3. Classificação de Erros
- **NETWORK**: Erros de conectividade (retryable)
- **TIMEOUT**: Erros de timeout (retryable)
- **AUTH**: Erros de autenticação (non-retryable)
- **RATE_LIMIT**: Rate limiting (retryable com delay)
- **VALIDATION**: Erros de validação (non-retryable)
- **RESOURCE**: Recursos não encontrados (context-dependent)
- **INTERNAL**: Erros internos do servidor (retryable)
- **UNKNOWN**: Erros não classificados (non-retryable)

## 🚀 Como Executar

```bash
# Navegar para o diretório do exemplo
cd examples/error_handling_example

# Executar o exemplo
go run main.go
```

## 📖 Código Explicado

### Configuração dos Componentes

```go
// Setup error handling components
retryConfig := tracer.DefaultRetryConfig()
circuitBreakerConfig := tracer.DefaultCircuitBreakerConfig()
errorHandler := tracer.NewDefaultErrorHandler(retryConfig, circuitBreakerConfig)
```

### Configurações Padrão

#### RetryConfig
```go
type RetryConfig struct {
    MaxRetries    int           // Máximo de tentativas (padrão: 3)
    BaseDelay     time.Duration // Delay base (padrão: 100ms)
    MaxDelay      time.Duration // Delay máximo (padrão: 30s)
    BackoffFactor float64       // Fator de multiplicação (padrão: 2.0)
    Jitter        bool          // Adicionar jitter (padrão: true)
}
```

#### CircuitBreakerConfig
```go
type CircuitBreakerConfig struct {
    FailureThreshold int           // Threshold de falhas (padrão: 5)
    Timeout          time.Duration // Timeout para reset (padrão: 60s)
    HalfOpenTimeout  time.Duration // Timeout no estado half-open (padrão: 30s)
}
```

## 🔍 Exemplos Detalhados

### 1. Network Error (Retryable)
```go
networkOperation := func() error {
    return errors.New("connection timeout")
}

err := tracer.RetryWithBackoff(ctx, networkOperation, errorHandler, "network_op")
// Classificação: NETWORK
// Comportamento: Retry automático com exponential backoff
```

### 2. Auth Error (Non-retryable)
```go
authOperation := func() error {
    return errors.New("unauthorized")
}

err := tracer.RetryWithBackoff(ctx, authOperation, errorHandler, "auth_op")
// Classificação: AUTH
// Comportamento: Falha imediata, sem retry
```

### 3. Circuit Breaker
```go
cb := tracer.NewDefaultCircuitBreaker(circuitBreakerConfig)

for i := 0; i < 10; i++ {
    operation := func() error {
        if i < 7 {
            return errors.New("service failure") // Primeiras 7 tentativas falham
        }
        return nil // Últimas 3 tentativas succedem
    }

    err := cb.Execute(ctx, operation)
    metrics := cb.GetMetrics()
    // Monitora: FailureCount, SuccessCount, State
}
```

## 📊 Métricas Disponíveis

### Circuit Breaker Metrics
```go
type CircuitBreakerMetrics struct {
    State           CircuitBreakerState // CLOSED, OPEN, HALF_OPEN
    FailureCount    int64               // Total de falhas
    SuccessCount    int64               // Total de sucessos
    RequestCount    int64               // Total de requests
    LastFailureTime time.Time           // Timestamp da última falha
    LastSuccessTime time.Time           // Timestamp do último sucesso
}
```

### Error Classification Metrics
```go
errorCounts := errorHandler.GetErrorCounts()
// Retorna map[ErrorClassification]int64 com contadores por tipo de erro
```

## 🎯 Casos de Uso

### 1. Microserviços com Dependências Externas
```go
// Protege chamadas para APIs externas
externalAPI := func() error {
    // Chamada para serviço externo
    return callExternalService()
}

err := circuitBreaker.Execute(ctx, externalAPI)
```

### 2. Operações de Database com Retry
```go
// Retry automático para operações de banco
dbOperation := func() error {
    return database.ExecuteQuery(query)
}

err := tracer.RetryWithBackoff(ctx, dbOperation, errorHandler, "db_query")
```

### 3. Processamento Batch com Tolerância a Falhas
```go
// Processa items com tolerância a falhas individuais
for _, item := range batch {
    err := processItemWithRetry(item)
    if err != nil {
        log.Printf("Item %v failed permanently: %v", item.ID, err)
        continue // Continua processando outros items
    }
}
```

## ⚙️ Configuração Personalizada

### Circuit Breaker Customizado
```go
config := tracer.CircuitBreakerConfig{
    FailureThreshold: 10,              // 10 falhas para abrir
    Timeout:         120 * time.Second, // 2 minutos para reset
    HalfOpenTimeout: 60 * time.Second,  // 1 minuto no half-open
}
cb := tracer.NewDefaultCircuitBreaker(config)
```

### Retry Customizado
```go
config := tracer.RetryConfig{
    MaxRetries:    5,                    // 5 tentativas máximas
    BaseDelay:     500 * time.Millisecond, // Delay inicial de 500ms
    MaxDelay:     60 * time.Second,      // Delay máximo de 1 minuto
    BackoffFactor: 1.5,                  // Crescimento de 50% por tentativa
    Jitter:       true,                  // Adiciona randomização
}
errorHandler := tracer.NewDefaultErrorHandler(config, circuitBreakerConfig)
```

## 🚨 Melhores Práticas

### 1. Context Awareness
```go
// Sempre respeite context cancellation
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()

err := tracer.RetryWithBackoff(ctx, operation, errorHandler, "operation")
```

### 2. Monitoring e Alertas
```go
// Monitore métricas para alertas
metrics := circuitBreaker.GetMetrics()
if metrics.FailureCount > threshold {
    alerting.SendAlert("Circuit breaker failure threshold exceeded")
}
```

### 3. Graceful Degradation
```go
// Implemente fallbacks para quando circuit breaker estiver aberto
err := circuitBreaker.Execute(ctx, primaryOperation)
if err != nil {
    // Fallback para operação degradada
    return fallbackOperation(ctx)
}
```

## 📈 Resultados Esperados

Ao executar o exemplo, você verá:

```
=== Enhanced Error Handling Example ===

--- Network Error Example ---
Network operation failed: connection timeout (Classification: NETWORK)

--- Auth Error Example ---
Auth operation failed: unauthorized (Classification: AUTH)

--- Circuit Breaker Example ---
Attempt 1: err=service failure, failures=1, successes=0
Attempt 2: err=service failure, failures=2, successes=0
...
Attempt 6: err=circuit breaker is open, failures=5, successes=0
...
🎉 Error handling examples completed!
```

## 🔗 Recursos Relacionados

- [Complete Integration Example](../complete-integration/) - Todos os recursos trabalhando juntos
- [Performance Example](../performance/) - Otimizações de performance
- [Edge Cases Example](../edge-cases-error-handling/) - Casos extremos e testes avançados

## 📚 Documentação Adicional

- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Exponential Backoff](https://en.wikipedia.org/wiki/Exponential_backoff)
- [Error Handling Best Practices](../../docs/error-handling.md)
