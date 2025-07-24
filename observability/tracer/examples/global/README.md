# Exemplo Global OpenTelemetry

Este exemplo demonstra como configurar um **TracerProvider global** que pode ser usado em toda a aplicação sem precisar passar o provider explicitamente.

## Conceito

Ao usar `otel.SetTracerProvider()`, você configura um TracerProvider globalmente que permite:

1. **Simplicidade**: Qualquer código pode obter um tracer com `otel.Tracer("name")`
2. **Desacoplamento**: Código de negócio não precisa conhecer configuração de tracing
3. **Flexibilidade**: Configuração centralizada com múltiplos backends
4. **Interceptação**: Middlewares e bibliotecas podem usar tracing automaticamente

## Configuração

### 1. Variáveis de Ambiente

```bash
# Configuração automática baseada em ambiente
export TRACER_SERVICE_NAME="global-web-app"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="datadog"     # ou grafana, newrelic, opentelemetry
export DATADOG_API_KEY="your-api-key"
export TRACER_SAMPLING_RATIO="0.1"        # 10% para produção
```

### 2. Configuração por Backend

#### Datadog
```bash
export TRACER_EXPORTER_TYPE="datadog"
export DATADOG_API_KEY="your-datadog-api-key"
```

#### Grafana Tempo
```bash
export TRACER_EXPORTER_TYPE="grafana"
export TRACER_ENDPOINT="http://tempo:3200"
```

#### New Relic
```bash
export TRACER_EXPORTER_TYPE="newrelic"
export NEW_RELIC_LICENSE_KEY="your-40-char-license-key"
```

#### OpenTelemetry Genérico
```bash
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"
```

## Executar o Exemplo

```bash
# Configurar backend desejado
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"

# Executar
go run main.go
```

## Como Funciona

### 1. Inicialização Global

```go
func initGlobalTracing() func() {
    // Carrega configuração do ambiente
    cfg := config.NewConfigFromEnv()
    
    // Inicializa TracerManager
    tracerManager := tracer.NewTracerManager()
    tracerProvider, err := tracerManager.Init(ctx, cfg)
    
    // ⭐ CONFIGURA GLOBALMENTE ⭐
    otel.SetTracerProvider(tracerProvider)
    
    return shutdownFunc
}
```

### 2. Uso em Qualquer Lugar

```go
func anyFunction(ctx context.Context) {
    // Obtém tracer global - NÃO precisa passar provider
    tracer := otel.Tracer("my-component")
    
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()
    
    // ... lógica de negócio
}
```

### 3. Estrutura da Aplicação Exemplo

```
http-request (root span)
├── authentication
│   ├── validate-jwt
│   └── check-permissions
├── process-business-logic
│   ├── fetch-user-data
│   │   ├── query-user
│   │   └── cache-user-data
│   ├── enrich-user-profile
│   │   ├── fetch-preferences
│   │   ├── fetch-history
│   │   └── fetch-recommendations
│   └── audit-user-access
```

## Benefícios do Padrão Global

### 1. **Código Limpo**
```go
// ❌ Sem global - precisa passar provider
func BusinessLogic(ctx context.Context, provider trace.TracerProvider) {
    tracer := provider.Tracer("business")
    // ...
}

// ✅ Com global - código mais limpo
func BusinessLogic(ctx context.Context) {
    tracer := otel.Tracer("business")
    // ...
}
```

### 2. **Bibliotecas de Terceiros**
Muitas bibliotecas já suportam OpenTelemetry automaticamente:

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/contrib/instrumentation/database/sql/otelsql"
)

// HTTP client automaticamente instrumentado
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

// Database automaticamente instrumentado
db, err := otelsql.Open("postgres", dsn)
```

### 3. **Middlewares HTTP**
```go
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tracer := otel.Tracer("http-middleware")
        ctx, span := tracer.Start(r.Context(), "http-request")
        defer span.End()
        
        // Propaga context automaticamente
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Componentes Demonstrados

### 1. **HTTP Handler**
- Request processing
- Response tracking
- Status codes
- Error handling

### 2. **Authentication Middleware**
- JWT validation
- Permission checking
- Security attributes

### 3. **Business Logic**
- Multi-step operations
- Data fetching
- Profile enrichment
- Audit logging

### 4. **Data Layer**
- Database queries
- Cache operations
- External API calls

### 5. **Background Services**
- User preferences
- History tracking
- Recommendations
- Audit trails

## Atributos Semânticos

### HTTP
- `http.method`: GET, POST, etc.
- `http.route`: Route pattern
- `http.status_code`: Response status
- `http.user_agent`: Client user agent

### Authentication
- `auth.method`: Authentication method
- `auth.provider`: Provider (JWT, OAuth)
- `token.type`: Token type
- `permission.resource`: Resource being accessed

### Database
- `db.system`: Database system
- `db.name`: Database name
- `db.operation`: SQL operation
- `db.table`: Table name
- `db.rows_affected`: Affected rows

### Cache
- `cache.system`: Cache system (Redis, Memcached)
- `cache.key`: Cache key
- `cache.ttl`: Time to live
- `cache.hit`: Cache hit/miss

### Business
- `user.id`: User identifier
- `request.id`: Request identifier
- `operation.type`: Operation type
- `audit.action`: Audit action

## Configuração para Produção

### 1. **Sampling**
```bash
# Reduzir sampling para produção
export TRACER_SAMPLING_RATIO="0.1"  # 10%
```

### 2. **Batching**
Configure batching no collector para otimizar performance:

```yaml
processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
```

### 3. **Resource Attributes**
```bash
export TRACER_ATTRIBUTES='{"service.version":"1.2.3","deployment.environment":"production","k8s.cluster":"prod-cluster"}'
```

### 4. **Security**
```bash
export TRACER_HEADERS='{"Authorization":"Bearer production-token"}'
export TRACER_INSECURE="false"
```

## Troubleshooting

### TracerProvider não configurado
```go
// Verificar se foi configurado
provider := otel.GetTracerProvider()
if provider == nil {
    log.Fatal("TracerProvider not configured")
}
```

### Context não propagado
```go
// Sempre propagar context
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

// ❌ Criar novo context
badCtx := context.Background()

// ✅ Usar context propagado
result := callService(ctx)  // Propaga tracing
```

### Performance em Produção
- Use sampling adequado (1-10%)
- Configure batching no exporter
- Monitore overhead de CPU/memória
- Use tail-based sampling se disponível

### Debugging
```bash
# Logs detalhados
export OTEL_LOG_LEVEL=debug

# Verificar propagação
export OTEL_PROPAGATORS=tracecontext,b3,baggage
```

## Integração com Frameworks

### Gin (HTTP)
```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

r := gin.New()
r.Use(otelgin.Middleware("web-service"))
```

### gRPC
```go
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

conn, err := grpc.Dial(target, 
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
)
```

### SQL
```go
import "go.opentelemetry.io/contrib/instrumentation/database/sql/otelsql"

db, err := otelsql.Open("postgres", dsn)
```

Isso permite instrumentação automática sem modificar código existente!
