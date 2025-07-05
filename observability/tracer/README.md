# Tracer - Distributed Tracing Abstraction

Uma abstração moderna e extensível para tracing distribuído em aplicações Go, seguindo os padrões OpenTelemetry e suportando múltiplos backends de tracing.

## 🚀 Quick Start

```bash
# Clonar o repositório
git clone https://github.com/fsvxavier/nexs-lib
cd isis-golang-lib/tracer

# Instalar dependências
go mod tidy

# Executar testes
make test

# Executar exemplos
make examples

# Configurar ambiente completo (Docker)
make setup-env
```

## Características

- **Interface Moderna**: Baseada nos padrões OpenTelemetry
- **Múltiplos Providers**: Suporte para Datadog, New Relic e Prometheus/Grafana
- **Extensível**: Fácil adição de novos providers
- **Type-Safe**: APIs totalmente tipadas
- **Performance**: Implementações otimizadas para produção
- **Observabilidade**: Métricas e logs integrados

## Providers Suportados

### 1. Datadog APM
- **Biblioteca**: `github.com/DataDog/dd-trace-go/v2`
- **Características**: 
  - APM completo com traces distribuídos
  - Profiling integrado
  - Suporte a tags customizadas
  - Configuração flexível de sampling

### 2. New Relic APM
- **Biblioteca**: `github.com/newrelic/go-agent/v3`
- **Características**:
  - APM e monitoramento de infraestrutura
  - Distributed tracing nativo
  - Custom attributes e eventos
  - Integração com logs

### 3. Prometheus + Grafana
- **Biblioteca**: `github.com/prometheus/client_golang`
- **Características**:
  - Métricas customizadas de tracing
  - Histogramas de duração
  - Contadores de spans e erros
  - Integração com Grafana dashboards

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/tracer
```

## Uso Básico

### Datadog Provider

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/tracer"
    "github.com/fsvxavier/nexs-lib/tracer/providers/datadog"
)

func main() {
    // Configurar provider
    config := &datadog.Config{
        ServiceName:     "my-service",
        ServiceVersion:  "1.0.0",
        Environment:     "production",
        AgentHost:       "localhost",
        AgentPort:       8126,
        EnableProfiling: true,
        SampleRate:      1.0,
    }

    provider := datadog.NewProvider(config)
    defer provider.Shutdown(context.Background())

    // Criar tracer
    tr := provider.CreateTracer("my-tracer",
        tracer.WithServiceName("my-service"),
        tracer.WithEnvironment("production"),
    )

    // Criar span
    ctx, span := tr.StartSpan(context.Background(), "operation-name",
        tracer.WithSpanKind(tracer.SpanKindServer),
        tracer.WithAttributes(map[string]interface{}{
            "user_id": 12345,
            "action": "create_user",
        }),
    )
    defer span.End()

    // Adicionar atributos e eventos
    span.SetAttribute("processing_time", 100)
    span.AddEvent("validation_completed", map[string]interface{}{
        "result": "success",
    })

    // Trabalho da aplicação...
    
    span.SetStatus(tracer.StatusCodeOk, "Operation completed")
}
```

### New Relic Provider

```go
config := &newrelic.Config{
    AppName:           "my-app",
    LicenseKey:        "your-license-key",
    Environment:       "production",
    DistributedTracer: true,
}

provider, err := newrelic.NewProvider(config)
if err != nil {
    log.Fatal(err)
}
defer provider.Shutdown(context.Background())

tr := provider.CreateTracer("newrelic-tracer")
ctx, span := tr.StartSpan(context.Background(), "database-query")
defer span.End()

// Trabalho da aplicação...
span.SetStatus(tracer.StatusCodeOk, "Query completed")
```

### Prometheus Provider

```go
config := &prometheus.Config{
    ServiceName:     "my-service",
    Namespace:       "myapp",
    Subsystem:       "traces",
    EnableDuration:  true,
    EnableErrors:    true,
}

provider := prometheus.NewProvider(config)
tr := provider.CreateTracer("prometheus-tracer")

ctx, span := tr.StartSpan(context.Background(), "business-logic")
defer span.End()

// Trabalho da aplicação...
// Métricas são automaticamente coletadas

// Expor métricas via HTTP
import "github.com/prometheus/client_golang/prometheus/promhttp"
http.Handle("/metrics", promhttp.HandlerFor(provider.GetRegistry(), promhttp.HandlerOpts{}))
```

## Configuração Avançada

### Múltiplos Providers

Você pode usar múltiplos providers simultaneamente:

```go
// APM detalhado com Datadog
ddProvider := datadog.NewProvider(&datadog.Config{
    ServiceName: "my-service",
    Environment: "production",
})

// Métricas customizadas com Prometheus
promProvider := prometheus.NewProvider(&prometheus.Config{
    ServiceName: "my-service",
    Namespace:   "myapp",
})

// Usar ambos os tracers
ddTracer := ddProvider.CreateTracer("apm-tracer")
promTracer := promProvider.CreateTracer("metrics-tracer")

ctx, ddSpan := ddTracer.StartSpan(context.Background(), "operation")
_, promSpan := promTracer.StartSpan(ctx, "operation")

// Trabalho da aplicação...

promSpan.End()
ddSpan.End()
```

### Spans Aninhados

```go
ctx, parentSpan := tr.StartSpan(context.Background(), "parent-operation")
defer parentSpan.End()

// Span filho
ctx, childSpan := tr.StartSpan(ctx, "child-operation",
    tracer.WithSpanKind(tracer.SpanKindInternal),
)
defer childSpan.End()

// Trabalho da aplicação...
childSpan.SetStatus(tracer.StatusCodeOk, "Child completed")
parentSpan.SetStatus(tracer.StatusCodeOk, "Parent completed")
```

### Tratamento de Erros

```go
ctx, span := tr.StartSpan(context.Background(), "risky-operation")
defer span.End()

if err := riskyFunction(); err != nil {
    span.RecordError(err, map[string]interface{}{
        "function": "riskyFunction",
        "retry_count": 3,
    })
    span.SetStatus(tracer.StatusCodeError, "Operation failed")
    return err
}

span.SetStatus(tracer.StatusCodeOk, "Operation successful")
```

## Configurações por Provider

### Datadog Config

```go
type Config struct {
    ServiceName     string            // Nome do serviço
    ServiceVersion  string            // Versão do serviço
    Environment     string            // Ambiente (dev, staging, prod)
    AgentHost       string            // Host do agente Datadog
    AgentPort       int               // Porta do agente Datadog
    EnableProfiling bool              // Habilitar profiling
    SampleRate      float64           // Taxa de sampling (0.0-1.0)
    Tags            map[string]string // Tags globais
    Debug           bool              // Modo debug
}
```

### New Relic Config

```go
type Config struct {
    AppName           string                     // Nome da aplicação
    LicenseKey        string                     // Chave de licença
    Environment       string                     // Ambiente
    ServiceVersion    string                     // Versão do serviço
    DistributedTracer bool                       // Habilitar distributed tracing
    Enabled           bool                       // Habilitar agente
    LogLevel          string                     // Nível de log
    Attributes        map[string]interface{}     // Atributos customizados
    Labels            map[string]string          // Labels
}
```

### Prometheus Config

```go
type Config struct {
    ServiceName     string                 // Nome do serviço
    ServiceVersion  string                 // Versão do serviço
    Environment     string                 // Ambiente
    Namespace       string                 // Namespace das métricas
    Subsystem       string                 // Subsistema das métricas 
    Registry        *prometheus.Registry   // Registry customizado
    Labels          map[string]string      // Labels globais
    EnableDuration  bool                   // Habilitar métricas de duração
    EnableErrors    bool                   // Habilitar métricas de erro
    EnableActive    bool                   // Habilitar métricas de spans ativos
    DurationBuckets []float64             // Buckets para histograma
}
```

## Métricas Prometheus

O provider Prometheus coleta automaticamente as seguintes métricas:

- `{namespace}_{subsystem}_total`: Contador total de spans
- `{namespace}_{subsystem}_duration_seconds`: Histograma de duração dos spans
- `{namespace}_{subsystem}_errors_total`: Contador de erros nos spans
- `{namespace}_{subsystem}_active`: Gauge de spans ativos
- `{namespace}_operations_errors_total`: Contador de erros por operação

## Integrações

### Fiber (HTTP Server)

```go
import "github.com/gofiber/fiber/v2"

func tracingMiddleware(tr tracer.Tracer) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx, span := tr.StartSpan(c.Context(), c.Route().Path,
            tracer.WithSpanKind(tracer.SpanKindServer),
            tracer.WithAttributes(map[string]interface{}{
                "http.method": c.Method(),
                "http.url":    c.OriginalURL(),
            }),
        )
        defer span.End()

        c.SetUserContext(ctx)
        
        err := c.Next()
        
        span.SetAttribute("http.status_code", c.Response().StatusCode())
        if err != nil {
            span.RecordError(err, nil)
            span.SetStatus(tracer.StatusCodeError, "HTTP request failed")
        } else {
            span.SetStatus(tracer.StatusCodeOk, "HTTP request completed")
        }
        
        return err
    }
}
```

### Database (PostgreSQL)

```go
func queryWithTracing(ctx context.Context, tr tracer.Tracer, db *sql.DB, query string, args ...interface{}) error {
    ctx, span := tr.StartSpan(ctx, "database-query",
        tracer.WithSpanKind(tracer.SpanKindClient),
        tracer.WithAttributes(map[string]interface{}{
            "db.system":    "postgresql",
            "db.statement": query,
        }),
    )
    defer span.End()

    _, err := db.ExecContext(ctx, query, args...)
    if err != nil {
        span.RecordError(err, map[string]interface{}{
            "db.operation": "exec",
        })
        span.SetStatus(tracer.StatusCodeError, "Database query failed")
        return err
    }

    span.SetStatus(tracer.StatusCodeOk, "Database query completed")
    return nil
}
```

## Melhores Práticas

### 1. Nomenclatura de Spans
- Use nomes descritivos e consistentes
- Inclua o tipo de operação (ex: `http-get-users`, `db-select-users`)
- Evite cardinalidade alta (não inclua IDs únicos no nome)

### 2. Atributos
- Use atributos para dados variáveis (IDs, valores)
- Prefira tipos primitivos (string, int, bool)
- Use prefixos padronizados (ex: `http.`, `db.`, `custom.`)

### 3. Tratamento de Erros
- Sempre registre erros nos spans
- Inclua contexto relevante nos atributos de erro
- Use status codes apropriados

### 4. Performance
- Configure sampling adequadamente em produção
- Evite spans muito granulares em hot paths
- Use spans assíncronos quando apropriado

### 5. Configuração
- Use variáveis de ambiente para configuração
- Tenha configurações diferentes por ambiente
- Monitore overhead de performance

## Exemplos Completos

Veja os exemplos completos em [`examples/main.go`](./examples/main.go).

## Contribuição

Para adicionar um novo provider:

1. Crie uma pasta em `providers/`
2. Implemente as interfaces `Provider` e `Tracer`
3. Adicione testes unitários
4. Atualize a documentação
5. Adicione exemplo de uso

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo LICENSE para detalhes.
