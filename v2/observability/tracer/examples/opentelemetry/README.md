# OpenTelemetry Integration Example

Este exemplo demonstra a integração nativa com OpenTelemetry no Nexs-Lib v2 Tracer, incluindo OTLP exporters, propagação de contexto W3C e configuração de recursos avançada.

## 🎯 Recursos Demonstrados

### 1. Configuração OpenTelemetry
- ✅ **OTLP Exporters**: Suporte para HTTP e gRPC
- ✅ **W3C Trace Context**: Propagação de contexto padrão
- ✅ **Resource Detection**: Identificação automática de serviços
- ✅ **Batch Processing**: Otimização de export em lotes
- ✅ **Sampling**: Configuração flexível de amostragem

### 2. Span Management
- Criação de spans básicos e aninhados
- Atributos estruturados e tipados
- Estados de span (OK, Error, Unset)
- Events com timestamp e metadados
- Links entre spans relacionados

### 3. Context Propagation
- Propagação automática via context.Context
- Extração e injeção de span context
- Suporte para múltiplos propagators
- Compatibilidade com padrões W3C

### 4. Error Handling Integration
- Record errors com contexto estruturado
- Classificação automática de erros
- Integration com circuit breaker patterns
- Observabilidade de falhas

## 🚀 Como Executar

### Pré-requisitos

1. **OpenTelemetry Collector** (opcional, para visualização)
```bash
# Docker Compose com OTLP Collector
docker run -p 4317:4317 -p 4318:4318 otel/opentelemetry-collector:latest
```

2. **Backend de Observabilidade** (escolha um):
   - Jaeger
   - Zipkin
   - New Relic
   - Datadog
   - Grafana + Tempo

### Execução

```bash
# Navegar para o diretório do exemplo
cd examples/opentelemetry

# Executar o exemplo
go run main.go
```

## 📖 Código Explicado

### Configuração Básica

```go
config := &tracer.OpenTelemetryConfig{
    ServiceName:      "example-service",      // Nome do serviço
    ServiceVersion:   "1.0.0",               // Versão do serviço
    ServiceNamespace: "examples",            // Namespace/ambiente
    Endpoint:         "localhost:4317",      // OTLP gRPC endpoint
    Insecure:         true,                  // SSL/TLS (false para produção)
    Timeout:          30 * time.Second,      // Timeout para exports
    BatchTimeout:     5 * time.Second,       // Timeout para batches
    MaxExportBatch:   512,                   // Máximo de spans por batch
    MaxQueueSize:     2048,                  // Tamanho da queue interna
    SamplingRatio:    1.0,                   // Taxa de amostragem (0.0-1.0)
    Propagators:      []string{"tracecontext", "baggage"}, // W3C standards
    ResourceAttrs: map[string]string{        // Atributos do recurso
        "environment": "development",
        "team":        "platform",
        "version":     "v1.0.0",
    },
}
```

### Configurações de Produção

```go
// Configuração para ambiente de produção
prodConfig := &tracer.OpenTelemetryConfig{
    ServiceName:      "payment-service",
    ServiceVersion:   "2.1.0",
    ServiceNamespace: "payments",
    Endpoint:         "collector.company.com:4317",
    Insecure:         false,                 // SSL/TLS habilitado
    SamplingRatio:    0.1,                   // 10% sampling para produção
    Headers: map[string]string{             // Headers customizados
        "x-api-key": "your-api-key",
    },
    ResourceAttrs: map[string]string{
        "environment":   "production",
        "datacenter":    "us-west-2",
        "cluster":       "payments-prod",
        "k8s.pod.name":  os.Getenv("POD_NAME"),
        "k8s.namespace": os.Getenv("POD_NAMESPACE"),
    },
}
```

## 🔍 Exemplos Detalhados

### 1. Span Básico com Atributos

```go
ctx, span := t.StartSpan(ctx, "user-registration",
    tracer.WithSpanKind(tracer.SpanKindServer),
    tracer.WithSpanAttributes(map[string]interface{}{
        "user.id":    "12345",
        "user.email": "user@example.com",
        "operation":  "register",
    }),
)
defer span.End()

// Adicionar atributos dinâmicos
span.SetAttribute("user.verified", true)
span.SetAttribute("registration.duration_ms", 50)

// Definir status de sucesso
span.SetStatus(tracer.StatusCodeOk, "User registered successfully")
```

### 2. Spans Aninhados (Parent-Child)

```go
// Parent span: HTTP request
ctx, parentSpan := t.StartSpan(ctx, "http-request",
    tracer.WithSpanKind(tracer.SpanKindServer),
    tracer.WithSpanAttributes(map[string]interface{}{
        "http.method": "POST",
        "http.url":    "/api/orders",
    }),
)
defer parentSpan.End()

// Child span: Database operation
ctx, dbSpan := t.StartSpan(ctx, "database-query",
    tracer.WithSpanKind(tracer.SpanKindClient),
    tracer.WithSpanAttributes(map[string]interface{}{
        "db.system":    "postgresql",
        "db.operation": "INSERT",
        "db.table":     "orders",
    }),
)
defer dbSpan.End()

// Child span: External API call
ctx, apiSpan := t.StartSpan(ctx, "payment-service",
    tracer.WithSpanKind(tracer.SpanKindClient),
)
defer apiSpan.End()
```

### 3. Error Recording

```go
ctx, span := t.StartSpan(ctx, "user-authentication")
defer span.End()

// Simular erro
err := fmt.Errorf("invalid credentials: password does not match")

// Registrar erro com contexto
span.RecordError(err, map[string]interface{}{
    "error.type":        "authentication_failed",
    "error.retry_count": 3,
    "error.user_agent":  "MyApp/1.0",
})

// Definir status de erro
span.SetStatus(tracer.StatusCodeError, "Authentication failed")

// Adicionar event para debugging
span.AddEvent("authentication_attempt", map[string]interface{}{
    "attempt_number": 3,
    "reason":         "invalid_password",
    "timestamp":      time.Now().Unix(),
})
```

### 4. Context Propagation

```go
// Extrair span do contexto
extractedSpan := t.SpanFromContext(ctx)
if extractedSpan != nil {
    spanCtx := extractedSpan.Context()
    fmt.Printf("Trace ID: %s\n", spanCtx.TraceID)
    fmt.Printf("Span ID: %s\n", spanCtx.SpanID)
}

// Criar novo contexto com span para downstream services
newCtx := t.ContextWithSpan(context.Background(), span)

// Propagar para funções downstream
processOrderStep(t, newCtx, "validate-inventory")
processOrderStep(t, newCtx, "reserve-items")
```

### 5. Events e Structured Logging

```go
span.AddEvent("cache_miss", map[string]interface{}{
    "cache.key":      "user:12345",
    "cache.ttl":      300,
    "cache.backend":  "redis",
})

span.AddEvent("processing_started", map[string]interface{}{
    "file.size":      1024000,
    "file.type":      "csv",
    "records.count":  5000,
})

span.AddEvent("processing_completed", map[string]interface{}{
    "records.processed": 4850,
    "records.failed":    150,
    "duration_ms":       2340,
})
```

## 🎯 Span Kinds Explicados

```go
// SERVER: Recebendo requests (HTTP handlers, gRPC servers)
tracer.WithSpanKind(tracer.SpanKindServer)

// CLIENT: Fazendo requests (HTTP calls, DB queries, gRPC calls)
tracer.WithSpanKind(tracer.SpanKindClient)

// PRODUCER: Produzindo mensagens (message queues, events)
tracer.WithSpanKind(tracer.SpanKindProducer)

// CONSUMER: Consumindo mensagens (message handlers, event listeners)
tracer.WithSpanKind(tracer.SpanKindConsumer)

// INTERNAL: Operações internas (business logic, calculations)
tracer.WithSpanKind(tracer.SpanKindInternal)
```

## 📊 Atributos Semânticos Recomendados

### HTTP Operations
```go
map[string]interface{}{
    "http.method":      "POST",
    "http.url":         "https://api.example.com/users",
    "http.status_code": 201,
    "http.user_agent":  "MyApp/1.0",
    "http.request.size": 1024,
    "http.response.size": 512,
}
```

### Database Operations
```go
map[string]interface{}{
    "db.system":        "postgresql",
    "db.connection_string": "postgresql://localhost:5432/mydb",
    "db.statement":     "SELECT * FROM users WHERE id = $1",
    "db.operation":     "SELECT",
    "db.table":         "users",
    "db.rows_affected": 1,
}
```

### Message Queue Operations
```go
map[string]interface{}{
    "messaging.system":      "rabbitmq",
    "messaging.destination": "order.events",
    "messaging.operation":   "publish",
    "messaging.message_id":  "msg_12345",
    "messaging.payload_size": 2048,
}
```

## 🔧 Configuração de Exporters

### OTLP via gRPC (Recomendado)
```go
config := &tracer.OpenTelemetryConfig{
    Endpoint: "https://api.honeycomb.io:443",
    Headers: map[string]string{
        "x-honeycomb-team": "your-api-key",
    },
    Insecure: false,
}
```

### OTLP via HTTP
```go
config := &tracer.OpenTelemetryConfig{
    Endpoint: "https://api.honeycomb.io:443/v1/traces",
    Headers: map[string]string{
        "x-honeycomb-team": "your-api-key",
    },
    Insecure: false,
}
```

### Self-hosted Jaeger
```go
config := &tracer.OpenTelemetryConfig{
    Endpoint: "http://jaeger-collector:14268/api/traces",
    Insecure: true,
}
```

## 🎛️ Sampling Strategies

### Production Sampling (10%)
```go
config.SamplingRatio = 0.1 // 10% das traces
```

### Development (100%)
```go
config.SamplingRatio = 1.0 // Todas as traces
```

### Error-based Sampling
```go
// Implementar custom sampler para sempre amostrar errors
// (Recurso avançado - consulte documentação OpenTelemetry)
```

## 📈 Resultados Esperados

Ao executar o exemplo, você verá:

```
=== OpenTelemetry Integration Example ===
✅ OpenTelemetry tracer created successfully

--- Basic Span Example ---
✅ Basic span created with attributes and status

--- Nested Spans Example ---
✅ Nested spans created: HTTP -> DB + Payment API

--- Error Handling Example ---
✅ Error recorded with context and debugging information

--- Context Propagation Example ---
✅ Span successfully extracted from context
📍 Trace ID: 1234567890abcdef1234567890abcdef
📍 Span ID: abcdef1234567890
✅ Context propagation verified across multiple steps

--- Events and Structured Logging Example ---
✅ Events and structured data recorded

🎉 All OpenTelemetry examples completed successfully!
📊 Check your OpenTelemetry collector/backend for traces
```

## 🚀 Integração com Backends

### Jaeger UI
```bash
# Acesse: http://localhost:16686
# Busque pelo service: "example-service"
```

### New Relic One
```bash
# Configure endpoint: https://otlp.nr-data.net:443
# Header: api-key: YOUR_LICENSE_KEY
```

### Datadog APM
```bash
# Configure endpoint: https://trace.agent.datadoghq.com:443
# Header: DD-API-KEY: YOUR_API_KEY
```

### Grafana Tempo
```bash
# Configure endpoint: http://tempo:3200/v1/traces
```

## 🔗 Recursos Relacionados

- [Complete Integration Example](../complete-integration/) - Todos os recursos trabalhando juntos
- [Performance Example](../performance/) - Otimizações de performance
- [Error Handling Example](../error_handling_example/) - Padrões avançados de erro

## 📚 Documentação Adicional

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OTLP Specification](https://opentelemetry.io/docs/reference/specification/protocol/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Semantic Conventions](https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/)

## 🚨 Melhores Práticas

### 1. Resource Attributes
```go
// Sempre configure atributos de recurso para identificação
ResourceAttrs: map[string]string{
    "service.name":    "my-service",
    "service.version": "1.2.3",
    "deployment.environment": "production",
    "k8s.pod.name":    os.Getenv("POD_NAME"),
}
```

### 2. Sampling Apropriado
```go
// Produção: sampling baixo para reduzir overhead
SamplingRatio: 0.05 // 5%

// Development: sampling alto para debugging
SamplingRatio: 1.0 // 100%
```

### 3. Error Handling
```go
// Sempre registre errors com contexto útil
span.RecordError(err, map[string]interface{}{
    "error.operation": "database_query",
    "error.table":     "users",
    "error.query_id":  queryID,
})
```

### 4. Performance
```go
// Use batch processing para alta throughput
BatchTimeout:   1 * time.Second,
MaxExportBatch: 512,
MaxQueueSize:   4096,
```
