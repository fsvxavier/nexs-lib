# Nexs Tracer Library

Uma biblioteca Go idiomática e extensível para gerenciamento de tracer providers distribuídos com suporte a múltiplos backends.

## 🎯 Funcionalidades

- ✅ **Múltiplos Providers**: Datadog APM, Grafana Tempo, New Relic, OpenTelemetry OTLP
- ✅ **Interface Unificada**: API consistente independente do provider
- ✅ **Configuração Flexível**: Suporte a variáveis de ambiente e funções `With*`
- ✅ **Propagadores**: TraceContext, B3, Jaeger
- ✅ **Context-Aware**: Suporte completo a `context.Context`
- ✅ **Thread-Safe**: Implementação segura para uso concorrente
- ✅ **Extensível**: Fácil adição de novos providers
- ✅ **Testável**: Mocks centralizados para testes isolados
- ✅ **Exemplos Completos**: Exemplos práticos para todos os cenários

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/observability/tracer
```

## 📁 Estrutura do Projeto

```
observability/tracer/
├── README.md              # Documentação principal
├── NEXT_STEPS.md          # Roadmap e próximos passos
├── tracer.go              # API principal e factory
├── config/                # Sistema de configuração
│   ├── config.go          # Configuração base
│   ├── options.go         # Opções funcionais
│   └── env.go             # Carregamento de variáveis de ambiente
├── interfaces/            # Interfaces e contratos
│   └── interfaces.go      # Interface TracerProvider
├── providers/             # Implementações dos providers
│   ├── datadog/           # Provider Datadog APM
│   ├── grafana/           # Provider Grafana Tempo
│   ├── newrelic/          # Provider New Relic
│   └── opentelemetry/     # Provider OpenTelemetry OTLP
├── mocks/                 # Mocks para testes
│   └── providers.go       # Mock providers centralizados
└── examples/              # Exemplos práticos
    ├── README.md          # Guia dos exemplos
    ├── datadog/           # Exemplo Datadog
    ├── grafana/           # Exemplo Grafana
    ├── newrelic/          # Exemplo New Relic
    ├── opentelemetry/     # Exemplo OpenTelemetry
    ├── global/            # Exemplo configuração global
    └── advanced/          # Exemplo traces + logs + métricas
```

## 🚀 Uso Rápido

### Configuração Básica com Opções

```go
package main

import (
    "context"
    "log"

    "github.com/fsvxavier/nexs-lib/observability/tracer"
    "github.com/fsvxavier/nexs-lib/observability/tracer/config"
)

func main() {
    ctx := context.Background()

    // Criar tracer provider com opções
    provider, err := tracer.NewTracerProviderWithOptions(ctx,
        config.WithServiceName("my-service"),
        config.WithExporterType("opentelemetry"),
        config.WithEndpoint("http://localhost:4318/v1/traces"),
        config.WithSamplingRatio(1.0),
        config.WithPropagators("tracecontext", "b3"),
        config.WithAttribute("team", "platform"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Usar o tracer
    tracer := provider.Tracer("my-component")
    _, span := tracer.Start(ctx, "my-operation")
    defer span.End()

    // Fazer alguma operação...

    log.Println("Tracing configured successfully!")
}
```

### Configuração via Variáveis de Ambiente

```bash
export TRACER_SERVICE_NAME="my-service"
export TRACER_EXPORTER_TYPE="datadog"
export TRACER_API_KEY="your-datadog-api-key"
export TRACER_SAMPLING_RATIO="0.1"
export TRACER_PROPAGATORS="tracecontext,b3"
export TRACER_ATTR_TEAM="platform"
export TRACER_HEADER_AUTHORIZATION="Bearer token123"
```

```go
package main

import (
    "context"
    "log"

    "github.com/fsvxavier/nexs-lib/observability/tracer"
)

func main() {
    ctx := context.Background()

    // Carregar configuração das variáveis de ambiente
    provider, err := tracer.NewTracerProviderFromEnv(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Tracing configured from environment!")
}
```

### Usando TracerManager

```go
package main

import (
    "context"
    "log"

    "github.com/fsvxavier/nexs-lib/observability/tracer"
    "github.com/fsvxavier/nexs-lib/observability/tracer/config"
)

func main() {
    ctx := context.Background()

    // Criar configuração
    cfg := config.NewConfig(
        config.WithServiceName("my-service"),
        config.WithExporterType("grafana"),
        config.WithEndpoint("http://tempo:3200"),
        config.WithSamplingRatio(0.5),
    )

    // Criar e inicializar manager
    manager := tracer.NewTracerManager()
    provider, err := manager.Init(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Shutdown(ctx)

    // Usar o tracer
    tracer := provider.Tracer("my-component")
    _, span := tracer.Start(ctx, "my-operation")
    defer span.End()

    log.Println("Tracing with manager configured!")
}
```

## 🔧 Configuração por Provider

### OpenTelemetry OTLP

```go
cfg := config.NewConfig(
    config.WithServiceName("my-service"),
    config.WithExporterType("opentelemetry"),
    config.WithEndpoint("http://localhost:4318/v1/traces"), // HTTP
    // config.WithEndpoint("localhost:4317"),              // gRPC
    config.WithSamplingRatio(1.0),
    config.WithInsecure(false),
    config.WithHeader("Authorization", "Bearer token"),
)
```

### Datadog APM

```go
cfg := config.NewConfig(
    config.WithServiceName("my-service"),
    config.WithExporterType("datadog"),
    config.WithAPIKey("your-datadog-api-key"),
    config.WithEnvironment("production"),
    config.WithSamplingRatio(0.1),
    config.WithAttribute("team", "platform"),
)
```

### Grafana Tempo

```go
cfg := config.NewConfig(
    config.WithServiceName("my-service"),
    config.WithExporterType("grafana"),
    config.WithEndpoint("http://tempo:3200"), // HTTP
    // config.WithEndpoint("tempo:9095"),     // gRPC
    config.WithSamplingRatio(0.5),
    config.WithHeader("X-Scope-OrgID", "tenant1"),
    config.WithAttribute("org-id", "tenant1"),
)
```

### New Relic

```go
cfg := config.NewConfig(
    config.WithServiceName("my-service"),
    config.WithExporterType("newrelic"),
    config.WithLicenseKey("your-40-character-license-key"),
    config.WithSamplingRatio(0.8),
    config.WithAttribute("region", "us-east-1"),
)
```

## 🌍 Variáveis de Ambiente

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `TRACER_SERVICE_NAME` | Nome do serviço | `my-service` |
| `TRACER_ENVIRONMENT` | Ambiente de execução | `production` |
| `TRACER_EXPORTER_TYPE` | Tipo do provider | `opentelemetry` |
| `TRACER_ENDPOINT` | Endpoint do collector | `http://localhost:4318/v1/traces` |
| `TRACER_API_KEY` | API Key (Datadog) | `your-api-key` |
| `TRACER_LICENSE_KEY` | License Key (New Relic) | `your-license-key` |
| `TRACER_SAMPLING_RATIO` | Taxa de amostragem | `0.1` |
| `TRACER_PROPAGATORS` | Propagadores (CSV) | `tracecontext,b3` |
| `TRACER_INSECURE` | Conexão insegura | `true` |
| `TRACER_VERSION` | Versão da aplicação | `1.0.0` |
| `TRACER_HEADER_*` | Cabeçalhos customizados | `TRACER_HEADER_AUTH=Bearer token` |
| `TRACER_ATTR_*` | Atributos customizados | `TRACER_ATTR_TEAM=platform` |

## 📚 Exemplos Completos

A biblioteca inclui **exemplos práticos e funcionais** na pasta [`examples/`](./examples/):

### 🔧 Exemplos por Provider
- **[`examples/datadog/`](./examples/datadog/)** - API de usuários com Datadog APM
- **[`examples/grafana/`](./examples/grafana/)** - API de usuários com Grafana Tempo  
- **[`examples/newrelic/`](./examples/newrelic/)** - API de usuários com New Relic
- **[`examples/opentelemetry/`](./examples/opentelemetry/)** - API de usuários com OTLP

### 🌍 Configuração Global
- **[`examples/global/`](./examples/global/)** - Demonstra uso de `otel.SetTracerProvider()` globalmente

### 🚀 Observabilidade Completa
- **[`examples/advanced/`](./examples/advanced/)** - **Integração traces + logs + métricas**

### 🏃‍♂️ Quick Start com Exemplos

```bash
# Escolher um exemplo
cd examples/opentelemetry/

# Configurar variáveis
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://localhost:4318/v1/traces"

# Executar
go run main.go

# Testar API
curl -X POST http://localhost:8080/users -d '{"name":"João","email":"joao@example.com"}'
```

Todos os exemplos incluem:
- ✅ Código funcional e compilável
- ✅ README.md detalhado
- ✅ Configuração específica do backend
- ✅ Instruções de execução e teste
- ✅ Estrutura de traces documentada

### Exemplo Avançado: Observabilidade Completa

O [`examples/advanced/`](./examples/advanced/) demonstra **integração completa** de:

#### 🔍 Traces Distribuídos
```go
tracer := otel.Tracer("order-service")
ctx, span := tracer.Start(r.Context(), "create-order")
defer span.End()

// Atributos semânticos
span.SetAttributes(
    attribute.String("order.id", order.ID),
    attribute.Float64("order.amount", order.Amount),
)
```

#### 📝 Logs Correlacionados
```go
logger.Info("Order created successfully",
    zap.String("order_id", order.ID),
    zap.String("trace_id", span.SpanContext().TraceID().String()),
    zap.String("span_id", span.SpanContext().SpanID().String()))
```

#### 📊 Métricas OpenTelemetry
```go
// Contador de pedidos
orderProcessed.Add(ctx, 1, 
    metric.WithAttributes(attribute.String("status", "success")))

// Histograma de valores
orderValue.Record(ctx, order.Amount, 
    metric.WithAttributes(attribute.String("payment_method", order.PaymentMethod)))
```

### Instrumentação HTTP

```go
import (
    "net/http"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Criar o tracer provider
provider, err := tracer.NewTracerProviderWithOptions(ctx,
    config.WithServiceName("http-service"),
    config.WithExporterType("opentelemetry"),
    config.WithEndpoint("http://localhost:4318/v1/traces"),
)

// Instrumentar HTTP handler
handler := otelhttp.NewHandler(http.HandlerFunc(myHandler), "my-endpoint")
http.Handle("/api", handler)
```

### Spans Customizados

```go
func myBusinessLogic(ctx context.Context) error {
    tracer := otel.Tracer("my-service")
    ctx, span := tracer.Start(ctx, "business-operation")
    defer span.End()

    // Adicionar atributos
    span.SetAttributes(
        attribute.String("user.id", "123"),
        attribute.Int("order.count", 5),
    )

    // Fazer alguma operação...

    return nil
}
```

## 🧪 Testes

```bash
# Executar todos os testes
go test -v -timeout 30s -race ./observability/tracer/...

# Executar testes com cobertura
go test -v -coverprofile=coverage.out ./observability/tracer/...
go tool cover -html=coverage.out

# Testar exemplo específico
cd examples/opentelemetry/
go run main.go
```

### Mocks para Testes

A biblioteca inclui **mocks centralizados** em [`mocks/providers.go`](./mocks/providers.go):

```go
import "github.com/fsvxavier/nexs-lib/observability/tracer/mocks"

func TestMyFunction(t *testing.T) {
    // Usar mock provider para testes isolados
    mockProvider := mocks.NewMockProviderForBackend("opentelemetry")
    
    // Configurar comportamento esperado
    mockProvider.SetShouldFailInit(false)
    
    // Testar sem dependências externas
    err := mockProvider.Init(ctx, config)
    assert.NoError(t, err)
}
```

## 🔄 Extensibilidade

Para adicionar um novo provider, implemente a interface `TracerProvider`:

```go
type TracerProvider interface {
    Init(ctx context.Context, config Config) (trace.TracerProvider, error)
    Shutdown(ctx context.Context) error
}
```

E registre na factory em `tracer.go`:

```go
func NewTracerProvider(config interfaces.Config) (interfaces.TracerProvider, error) {
    switch config.ExporterType {
    case "my-new-provider":
        return mynewprovider.NewProvider(), nil
    // ... outros providers
    }
}
```

## 📋 Requisitos

- Go 1.21+
- OpenTelemetry Go SDK v1.37.0+

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Add nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🔗 Links Úteis

- **[Exemplos Práticos](./examples/)** - Exemplos completos para todos os providers
- **[Configuração](./config/)** - Sistema de configuração flexível
- **[Interfaces](./interfaces/)** - Interfaces e contratos
- **[Providers](./providers/)** - Implementações dos providers
- **[Mocks](./mocks/)** - Mocks para testes isolados
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [Datadog APM Go](https://github.com/DataDog/dd-trace-go)
- [New Relic Go Agent](https://github.com/newrelic/go-agent)
- [Grafana Tempo](https://grafana.com/oss/tempo/)
