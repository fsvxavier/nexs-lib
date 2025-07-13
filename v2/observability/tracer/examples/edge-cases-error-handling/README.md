# Edge Cases & Error Handling Example

Este exemplo demonstra padrões avançados de tratamento de erro, resiliência e cenários extremos para a biblioteca de tracer. Ele mostra como construir soluções de rastreamento distribuído robustas que lidam graciosamente com falhas e condições extremas.

## 🎯 O Que Este Exemplo Demonstra

### Padrões de Resiliência Fundamentais
- **Circuit Breaker Pattern**: Previne falhas em cascata bloqueando temporariamente requisições quando taxas de erro excedem limites
- **Retry com Exponential Backoff**: Mecanismo inteligente de retry com jitter para evitar problemas de thundering herd
- **Gerenciamento de Recursos**: Monitoramento e limitação do uso de recursos para prevenir esgotamento do sistema
- **Degradação Graciosa**: Manutenção da disponibilidade do serviço mesmo quando alguns componentes falham

### Cenários de Edge Cases
- **Falhas de Rede**: Timeouts de conexão, falhas de DNS, erros de handshake SSL
- **Esgotamento de Recursos**: Memory leaks, goroutine leaks, esgotamento de file descriptors
- **Corrupção de Dados**: Contextos de trace inválidos, dados de span corrompidos, problemas de encoding
- **Problemas de Concorrência**: Race conditions, deadlocks, padrões de acesso concorrente

### Melhores Práticas de Tratamento de Erro
- **Classificação Estruturada de Erros**: Categorização de erros para tratamento apropriado
- **Propagação de Contexto**: Manutenção do contexto de trace através de cenários de erro
- **Telemetria Durante Falhas**: Captura de dados de observabilidade mesmo quando as coisas dão errado
- **Modos de Falha Seguros**: Garantia de que falhas não quebrem todo o sistema

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                 Simulador de Edge Cases                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────┐    ┌──────────────────┐              │
│  │ Circuit Breaker  │    │ Retry Manager    │              │
│  │                  │    │                  │              │
│  │ • Contagem Falhas│    │ • Lógica Backoff │              │
│  │ • Máquina Estado │    │ • Suporte Jitter │              │
│  │ • Timer Reset    │    │ • Cancel Context │              │
│  └──────────────────┘    └──────────────────┘              │
│                                                             │
│  ┌──────────────────┐    ┌──────────────────┐              │
│  │ Monitor Recursos │    │ Classificador    │              │
│  │                  │    │ de Erros         │              │
│  │ • Uso Memória    │    │ • Erros de Rede  │              │
│  │ • Cont Goroutine │    │ • Erros Recursos │              │
│  │ • File Handles   │    │ • Erros de Dados │              │
│  └──────────────────┘    └──────────────────┘              │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  Cenários de Simulação                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Problemas de Rede     Problemas de Recursos               │
│  ├─ Timeout Conexão    ├─ Esgotamento Memória               │
│  ├─ Falha DNS          ├─ Vazamentos Goroutine              │
│  ├─ Conexão Recusada   ├─ Esgotamento File Handle          │
│  ├─ Erros SSL          └─ Pool Conexão Cheio               │
│  └─ Falhas Intermitentes                                    │
│                                                             │
│  Corrupção de Dados    Problemas Concorrência              │
│  ├─ Contextos Inválidos├─ Race Conditions                  │
│  ├─ Spans Malformados  ├─ Detecção Deadlock                │
│  ├─ Erros Encoding     ├─ Operações Atômicas               │
│  └─ Payloads Muito Grandes └─ Criação Concorrente Spans    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Início Rápido

### Pré-requisitos
```bash
# Certifique-se que Go 1.19+ está instalado
go version

# Navegue para o diretório do exemplo
cd examples/edge-cases-error-handling

# Instale dependências
go mod tidy
```

### Executando o Exemplo
```bash
# Execute a simulação completa de edge cases
go run main.go

# Execute cenários específicos de teste
go test -v

# Execute benchmarks de performance
go test -bench=. -benchmem

# Execute com detecção de race
go test -race -v
```

## 📋 Detalhes dos Cenários

### 1. Simulação de Falhas de Rede

O exemplo simula vários padrões de falha de rede que ocorrem comumente em sistemas distribuídos:

```go
scenarios := []struct {
    name        string
    probability float64
    simulate    func(context.Context) error
}{
    {"connection_timeout", 0.3, simulateConnectionTimeout},
    {"dns_failure", 0.2, simulateDNSFailure},
    {"connection_refused", 0.25, simulateConnectionRefused},
    {"intermittent_failure", 0.15, simulateIntermittentFailure},
    {"ssl_handshake_failure", 0.1, simulateSSLHandshakeFailure},
}
```

**Características Principais:**
- Simulação realística de erros de rede
- Classificação e tratamento adequado de erros
- Suporte a cancelamento de contexto
- Correlação de trace através de falhas

### 2. Implementação do Circuit Breaker

```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    failures        int64
    lastFailureTime time.Time
    state           CircuitBreakerState
    mu              sync.RWMutex
}
```

**Estados:**
- **Fechado**: Operação normal, requisições passam
- **Aberto**: Limite de falhas excedido, requisições bloqueadas
- **Meio-Aberto**: Testando se o serviço se recuperou

**Benefícios:**
- Previne falhas em cascata
- Detecção automática de recuperação
- Limites de falha configuráveis
- Telemetria detalhada sobre transições de estado

### 3. Retry com Exponential Backoff

```go
type RetryConfig struct {
    MaxAttempts    int
    BaseDelay      time.Duration
    MaxDelay       time.Duration
    BackoffFactor  float64
    Jitter         bool
}
```

**Funcionalidades:**
- Cálculo de backoff exponencial
- Jitter para prevenir thundering herd
- Suporte a cancelamento de contexto
- Rastreamento por tentativa
- Políticas de retry configuráveis

## 🧪 Cenários de Teste

### Executando Testes Individuais

```bash
# Testar funcionalidade do circuit breaker
go test -run TestCircuitBreaker -v

# Testar mecanismo de retry
go test -run TestRetryWithBackoff -v

# Testar gerenciamento de recursos
go test -run TestResourceManager -v

# Testar cálculo de backoff
go test -run TestBackoffCalculation -v
```

### Benchmarks de Performance

```bash
# Benchmark da performance do circuit breaker
go test -bench=BenchmarkCircuitBreakerExecute -benchmem

# Benchmark do mecanismo de retry
go test -bench=BenchmarkRetryWithBackoff -benchmem
```

### Detecção de Race Conditions

```bash
# Executar com detector de race
go test -race -v

# Executar teste de stress de concorrência
go test -run TestConcurrencyIssues -v -count=10
```

## 📊 Métricas e Observabilidade

### Principais Métricas Capturadas

1. **Métricas do Circuit Breaker**
   - Transições de estado (fechado → aberto → meio-aberto)
   - Contagem de falhas e taxa de sucesso
   - Duração do timer de reset
   - Contagem de rejeições de requisição

2. **Métricas de Retry**
   - Contagem de tentativas por operação
   - Cálculos de delay de backoff
   - Taxa de sucesso por número de tentativa
   - Frequência de cancelamento de contexto

3. **Métricas de Recursos**
   - Tendências de uso de memória
   - Contagem de goroutines ao longo do tempo
   - Utilização de file handles
   - Estatísticas do pool de conexões

### Exemplo de Atributos de Trace

```json
{
  "circuit_breaker.state": "closed",
  "circuit_breaker.failures": 0,
  "retry.attempt": 2,
  "retry.successful_attempt": 2,
  "retry.max_attempts": 3,
  "resource.memory_mb": 45.2,
  "resource.goroutines": 127,
  "error.type": "network_failure",
  "error.recoverable": true,
  "simulation.type": "network_failures",
  "scenario.name": "connection_timeout"
}
```

## 🛠️ Opções de Configuração

### Configuração do Circuit Breaker

```go
circuitBreaker := NewCircuitBreaker(
    5,                    // maxFailures
    30*time.Second,       // resetTimeout
)
```

### Configuração de Retry

```go
retryConfig := RetryConfig{
    MaxAttempts:   3,                    // Máximo de tentativas
    BaseDelay:     100*time.Millisecond, // Delay inicial
    MaxDelay:      2*time.Second,        // Delay máximo
    BackoffFactor: 2.0,                  // Fator exponencial
    Jitter:        true,                 // Habilitar jitter
}
```

## 🎯 Melhores Práticas Demonstradas

### 1. Classificação de Erros
- Distinguir entre erros transitórios e permanentes
- Usar estratégias de retry apropriadas para cada tipo de erro
- Capturar contexto de erro para debugging

### 2. Gerenciamento de Recursos
- Monitorar uso de recursos continuamente
- Implementar degradação graciosa quando limites são aproximados
- Limpar recursos adequadamente para prevenir vazamentos

### 3. Observabilidade Durante Falhas
- Manter rastreamento mesmo durante condições de erro
- Capturar contexto detalhado de erro e métricas de recuperação
- Usar logging estruturado para análise de erros

## 📈 Resultados Esperados

Ao executar este exemplo, você deve observar:

1. **Comportamento Resiliente**: Sistema continua operando apesar de falhas simuladas
2. **Recuperação Inteligente**: Circuit breakers e retries permitem recuperação automática
3. **Proteção de Recursos**: Sistema previne esgotamento de recursos através de monitoramento
4. **Telemetria Detalhada**: Dados ricos de observabilidade capturados durante todos os cenários
5. **Características de Performance**: Overhead mínimo dos padrões de resiliência

### Exemplo de Saída

```
Iniciando Exemplo de Edge Cases & Tratamento de Erros
====================================================

1. Testando vários cenários de falha de rede
   Executando cenário: network_failures
   ✅ Cenário completado com sucesso
   Testando mecanismo de retry...
   ✅ Mecanismo de retry bem-sucedido

2. Testando cenários de esgotamento de recursos
   Executando cenário: resource_exhaustion
   ✅ Cenário completado com sucesso

3. Testando tratamento de corrupção de dados
   Executando cenário: data_corruption
   ✅ Cenário completado com sucesso

4. Testando acesso concorrente e race conditions
   Executando cenário: concurrency_issues
   ✅ Cenário completado com sucesso

====================================================
Resumo do Teste de Edge Cases:
Falhas de rede encontradas: 3
Estado do circuit breaker: fechado
Conexões ativas: 23
Goroutines atuais: 8
Uso de memória: 12.34 MB
Ciclos de GC: 5

Exemplo de Edge Cases & Tratamento de Erros completado!
Verifique seu backend de rastreamento para traces e métricas detalhados.
```

## 🚨 Solução de Problemas

### Problemas Comuns

1. **Circuit Breaker Travado Aberto**
   - Verificar configuração do limite de falhas
   - Verificar se timeout de reset é apropriado
   - Garantir recuperação do serviço subjacente

2. **Esgotamento de Retry**
   - Ajustar contagem de retry e parâmetros de backoff
   - Verificar lógica de classificação de erro
   - Verificar valores de timeout de contexto

3. **Esgotamento de Recursos**
   - Monitorar limites de recursos
   - Implementar limpeza adequada
   - Verificar vazamentos de goroutine

### Dicas de Debug

```bash
# Habilitar logging de debug
export NEXS_TRACER_DEBUG=true

# Executar com saída detalhada
go run main.go -v

# Profilear uso de memória
go test -memprofile=mem.prof -bench=.

# Profilear uso de CPU
go test -cpuprofile=cpu.prof -bench=.
```

## 🔗 Exemplos Relacionados

- [Uso Básico](../basic-usage/): Configuração simples de tracing
- [Servidores HTTP](../http-servers/): Tracing de serviços web
- [Microserviços](../microservices/): Tracing service-to-service
- [Benchmark de Performance](../performance-benchmark/): Padrões de teste de carga

---

Este exemplo demonstra padrões de resiliência e estratégias de tratamento de erro prontos para produção, essenciais para construir soluções robustas de rastreamento distribuído. Os padrões mostrados aqui devem ser adaptados ao seu caso de uso específico e requisitos operacionais.
- Intermittent connectivity
- DNS resolution failures
- SSL/TLS handshake errors
- Connection pool exhaustion

### Resource Exhaustion
- Memory limits
- File descriptor limits
- Thread pool exhaustion
- Disk space issues

### Data Corruption
- Invalid trace context
- Malformed span data
- Encoding errors
- Serialization failures
