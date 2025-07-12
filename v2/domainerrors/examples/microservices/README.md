# Microservices Examples

Este exemplo demonstra padrões avançados de tratamento de erros em arquiteturas de microserviços, incluindo service registry, circuit breaker, distributed tracing, service mesh, e padrões de resiliência.

## 🎯 Objetivos

- **Service Registry**: Descoberta e gerenciamento de serviços com tratamento de erros
- **Circuit Breaker**: Proteção contra falhas em cascata
- **Distributed Tracing**: Rastreamento distribuído com contexto de erro
- **Error Propagation**: Propagação inteligente de erros entre serviços
- **Service Mesh**: Gerenciamento de comunicação entre serviços
- **Timeout Handling**: Tratamento robusto de timeouts
- **Bulkhead Pattern**: Isolamento de recursos para evitar falhas
- **Correlation Tracking**: Rastreamento de correlação de requests

## 🏗️ Arquitetura

### Componentes Principais

1. **ServiceRegistry**: Registro e descoberta de serviços
2. **CircuitBreaker**: Proteção contra falhas em cascata
3. **DistributedTracer**: Sistema de tracing distribuído
4. **ErrorPropagator**: Propagação de erros entre serviços
5. **ServiceMesh**: Malha de serviços com políticas
6. **TimeoutHandler**: Gerenciamento de timeouts
7. **Bulkhead**: Isolamento de recursos
8. **CorrelationManager**: Gerenciamento de correlation IDs

### Padrões Implementados

- **Service Registry Pattern**: Descoberta dinâmica de serviços
- **Circuit Breaker Pattern**: Proteção contra falhas em cascata
- **Bulkhead Pattern**: Isolamento de falhas
- **Retry Pattern**: Tentativas automáticas com backoff
- **Timeout Pattern**: Controle de tempo limite
- **Correlation Pattern**: Rastreamento de requests

## 📚 Exemplos

### 1. Service Registry
```go
registry := NewServiceRegistry()

service := &ServiceInfo{
    Name:     "user-service",
    Version:  "v1.2.3",
    Host:     "user-service.default.svc.cluster.local",
    Port:     8080,
    Tags:     []string{"user", "authentication"},
}

registry.Register(service)
discoveredService, err := registry.Discover("user-service")
```

### 2. Circuit Breaker
```go
cb := NewCircuitBreaker("payment-service", CircuitBreakerConfig{
    FailureThreshold: 3,
    TimeoutDuration:  5 * time.Second,
    RecoveryTimeout:  10 * time.Second,
})

result, err := cb.Call(func() (interface{}, interfaces.DomainErrorInterface) {
    return callPaymentService()
})

// States: CLOSED -> OPEN -> HALF_OPEN -> CLOSED
fmt.Printf("Circuit state: %s\n", cb.GetState())
```

### 3. Distributed Tracing
```go
tracer := NewDistributedTracer()
traceID := tracer.StartTrace(ctx, "order_processing")

ctx, span := tracer.StartSpan(ctx, "user-service", "validate_user")
result, err := userService.ValidateUser(userID)
span.Finish(err)

if err != nil {
    tracer.RecordError(ctx, err)
    // Error includes trace_id, span_id, parent_spans
}
```

### 4. Error Propagation
```go
propagator := NewErrorPropagator()

// Original error from database
dbErr := factory.GetDefaultFactory().NewInternal("DB connection failed", nil)

// Propagate through services
userServiceErr := propagator.PropagateError(dbErr, PropagationContext{
    Service:   "user-service",
    Operation: "get_user_profile",
    Version:   "v1.2.3",
})

// Chain includes: root_cause_code, services_involved, chain_length
```

### 5. Service Mesh
```go
mesh := NewServiceMesh()

mesh.SetRetryPolicy("payment-service", RetryPolicy{
    MaxRetries:    3,
    BackoffFactor: 2,
    InitialDelay:  100 * time.Millisecond,
})

result, err := mesh.RouteRequest("order-service", "payment-service", "charge_card", false)
// Includes automatic retries, timeout handling, load balancing
```

### 6. Timeout Handling
```go
timeoutHandler := NewTimeoutHandler()

result, err := timeoutHandler.CallWithTimeout(
    "slow-service", 
    "heavy_operation", 
    2*time.Second,  // actual duration
    1*time.Second,  // timeout
)

// Returns SERVICE_TIMEOUT error if exceeded
```

### 7. Bulkhead Pattern
```go
bulkhead := NewBulkhead()

bulkhead.CreatePool("critical", BulkheadConfig{
    MaxConcurrency: 5,
    QueueSize:     10,
    Timeout:       2 * time.Second,
})

result, err := bulkhead.Execute("critical", "task-1", func() (interface{}, interfaces.DomainErrorInterface) {
    return performCriticalTask()
})
```

### 8. Correlation Tracking
```go
correlator := NewCorrelationManager()
correlationID := correlator.GenerateCorrelationID()

// Propagate through service chain
result := correlator.ProcessRequest(correlationID, "user-service", "authenticate")

// Create correlated error
err := correlator.CreateCorrelatedError(correlationID, "payment-service", "PAYMENT_FAILED", "Gateway timeout")

// Error includes: correlation_id, service, timestamp, request_path
```

## 🔧 Funcionalidades

### Service Discovery
- **Dynamic Registration**: Registro dinâmico de serviços
- **Health Checking**: Verificação de saúde automática
- **Load Balancing**: Distribuição de carga
- **Service Metadata**: Metadados e tags de serviços

### Resilience Patterns
- **Circuit Breaker**: Estados CLOSED/OPEN/HALF_OPEN
- **Retry Logic**: Backoff exponencial configurável
- **Timeout Management**: Timeouts por operação
- **Bulkhead Isolation**: Isolamento de recursos

### Error Handling
- **Error Propagation**: Propagação inteligente entre serviços
- **Context Preservation**: Preservação de contexto distribuído
- **Error Aggregation**: Agregação de erros relacionados
- **Root Cause Analysis**: Análise de causa raiz

### Observability
- **Distributed Tracing**: Tracing completo de requests
- **Correlation IDs**: IDs de correlação automáticos
- **Service Metrics**: Métricas por serviço
- **Error Analytics**: Análise de padrões de erro

## 🎨 Patterns Demonstrados

### 1. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    state       CircuitBreakerState  // CLOSED, OPEN, HALF_OPEN
    failureCount int
    lastFailure  time.Time
    config      CircuitBreakerConfig
}

func (cb *CircuitBreaker) Call(fn func() (interface{}, error)) (interface{}, error) {
    if cb.state == CircuitBreakerOpen {
        if time.Since(cb.lastFailure) > cb.config.RecoveryTimeout {
            cb.state = CircuitBreakerHalfOpen
        } else {
            return nil, NewCircuitBreakerOpenError()
        }
    }
    
    result, err := fn()
    cb.recordResult(err)
    return result, err
}
```

### 2. Bulkhead Pattern
```go
type Bulkhead struct {
    pools map[string]*ResourcePool
}

type ResourcePool struct {
    semaphore chan struct{}  // Limits concurrency
    queue     chan Task       // Queues overflow
    timeout   time.Duration   // Task timeout
}

func (b *Bulkhead) Execute(pool string, task Task) (interface{}, error) {
    select {
    case b.pools[pool].semaphore <- struct{}{}:
        defer func() { <-b.pools[pool].semaphore }()
        return task.Execute()
    case <-time.After(b.pools[pool].timeout):
        return nil, NewBulkheadTimeoutError()
    }
}
```

### 3. Service Mesh Pattern
```go
type ServiceMesh struct {
    services       map[string]*ServiceInfo
    retryPolicies  map[string]RetryPolicy
    timeoutPolicies map[string]TimeoutPolicy
}

func (sm *ServiceMesh) RouteRequest(from, to, operation string) (interface{}, error) {
    policy := sm.retryPolicies[to]
    
    for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
        result, err := sm.callService(to, operation)
        if err == nil || !isRetryableError(err) {
            return result, err
        }
        
        time.Sleep(policy.BackoffDelay(attempt))
    }
    
    return nil, NewMaxRetriesExceededError()
}
```

## 🚀 Execução

```bash
# Executar exemplo completo
go run main.go

# Executar com tracing habilitado
go run main.go -tracing

# Simular falhas específicas
go run main.go -simulate-failures

# Benchmarks de resiliência
go test -bench=BenchmarkResilience -benchmem
```

## 📊 Métricas de Performance

- **Service Discovery**: ~1ms/lookup
- **Circuit Breaker**: ~100ns/call overhead
- **Distributed Tracing**: ~5μs/span
- **Error Propagation**: ~2μs/hop
- **Service Mesh Routing**: ~500μs/request
- **Timeout Handling**: ~50ns overhead
- **Bulkhead Execution**: ~1μs/task
- **Correlation Tracking**: ~200ns/operation

## 🎯 Casos de Uso

### E-commerce Platform
- **Order Processing**: Circuit breakers para payment gateways
- **Inventory Management**: Bulkhead para critical/non-critical operations
- **User Authentication**: Service mesh com retry policies
- **Notification System**: Timeout handling para external providers

### Financial Services
- **Transaction Processing**: Multi-level circuit breakers
- **Risk Assessment**: Bulkhead isolation por risk level
- **Compliance Reporting**: Distributed tracing para auditoria
- **External APIs**: Comprehensive error propagation

### IoT Platform
- **Device Management**: Service registry para device services
- **Data Processing**: Bulkhead por device type
- **Alert System**: Circuit breaker para notification services
- **Analytics Pipeline**: Distributed tracing para data flow

### Media Streaming
- **Content Delivery**: Service mesh para CDN routing
- **User Preferences**: Circuit breaker para recommendation engine
- **Live Streaming**: Timeout handling para real-time services
- **Analytics**: Correlation tracking para user sessions

## 🔍 Observabilidade

### Service Metrics
- **Request Rate**: Requests por segundo por serviço
- **Error Rate**: Taxa de erro por serviço
- **Response Time**: Latência média e P99
- **Circuit Breaker States**: Estado atual e transições

### Distributed Tracing
- **Request Flow**: Fluxo completo de requests
- **Error Correlation**: Correlação de erros distribuídos
- **Performance Bottlenecks**: Identificação de gargalos
- **Service Dependencies**: Mapa de dependências

### Error Analytics
- **Error Patterns**: Padrões de erro mais comuns
- **Service Health**: Saúde geral dos serviços
- **Recovery Metrics**: Métricas de recuperação
- **Propagation Analysis**: Análise de propagação

## 📋 Checklist de Implementação

- ✅ Service registry com health checks
- ✅ Circuit breaker com estados corretos
- ✅ Distributed tracing funcional
- ✅ Error propagation com contexto
- ✅ Service mesh com políticas
- ✅ Timeout handling robusto
- ✅ Bulkhead pattern implementado
- ✅ Correlation tracking completo
- ✅ Retry policies configuráveis
- ✅ Performance monitoring integrado
- ✅ Error analytics implementado
- ✅ Fault injection para testes

## 🔮 Próximos Passos

1. **Service Mesh Integration**: Integração com Istio/Linkerd
2. **Chaos Engineering**: Fault injection automático
3. **Auto-scaling**: Scaling baseado em error rates
4. **ML-based Prediction**: Predição de falhas com ML
5. **Multi-region Support**: Suporte a múltiplas regiões
6. **Event-driven Architecture**: Integração com event sourcing
7. **Security Integration**: Security context em errors
8. **Cost Optimization**: Otimização baseada em métricas
