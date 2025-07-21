# Middleware System

Este módulo fornece um sistema abrangente de middleware para servidores HTTP, seguindo a arquitetura de Design Patterns e boas práticas do projeto nexs-lib.

## Arquitetura

O sistema de middleware segue os seguintes padrões:

- **Chain of Responsibility Pattern**: Middlewares são organizados em uma cadeia de execução
- **Decorator Pattern**: Cada middleware envolve o handler HTTP com funcionalidade adicional  
- **Strategy Pattern**: Diferentes algoritmos para rate limiting, compressão, etc.
- **Observer Pattern**: Sistema de logging e métricas

## Estrutura

```
middleware/
├── chain.go                   # Implementação da cadeia de middleware
├── health/                    # Health checks avançados
│   └── health.go
├── ratelimit/                 # Rate limiting com múltiplos algoritmos
│   └── ratelimit.go
├── cors/                      # CORS (Cross-Origin Resource Sharing)
│   └── cors.go  
├── logging/                   # Request/Response logging
│   └── logging.go
├── compression/               # Compressão de resposta
│   └── compression.go
├── timeout/                   # Timeout de requisições
│   └── timeout.go
├── bulkhead/                  # Bulkhead pattern para isolamento
│   └── bulkhead.go
├── retry/                     # Políticas de retry
│   └── retry.go
└── README.md                  # Esta documentação

examples/
└── middleware/                # Exemplos de uso
    ├── main.go                # Exemplo avançado completo
    └── README.md              # Documentação do exemplo
```

**Nota**: As interfaces base estão definidas em `interfaces/interfaces.go` seguindo a estrutura do projeto.

## Interfaces Principais

### Middleware Interface

```go
type Middleware interface {
    Wrap(next http.Handler) http.Handler
    Name() string
    Priority() int  // Lower numbers execute first
}
```

### MiddlewareChain Interface

```go
type MiddlewareChain interface {
    Add(middleware Middleware) MiddlewareChain
    Then(h http.Handler) http.Handler
}
```

## Middlewares Disponíveis

### 1. Health Checks (Priority: N/A - HTTP Handlers)

Sistema avançado de health checks com suporte a:
- **Liveness Probes**: Verifica se a aplicação está viva
- **Readiness Probes**: Verifica se está pronta para receber tráfego  
- **Startup Probes**: Verifica se a aplicação terminou a inicialização

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/health"

// Criar registry de health checks
registry := health.NewRegistry()

// Registrar checks
registry.Register("database", health.DatabaseCheck(db.Ping), 
    health.WithType(health.CheckTypeReadiness))
registry.Register("external-api", health.URLCheck("https://api.example.com/health"))

// Criar handlers HTTP
handler := health.NewHandler(registry)
http.Handle("/health/live", handler.LivenessHandler())
http.Handle("/health/ready", handler.ReadinessHandler())
http.Handle("/health", handler.HealthHandler())
```

### 2. Rate Limiting (Priority: 200)

Rate limiting com múltiplos algoritmos:
- **Token Bucket**: Permite rajadas controladas
- **Sliding Window**: Janela deslizante precisa
- **Fixed Window**: Janela fixa simples

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/ratelimit"

config := ratelimit.Config{
    Config: middleware.Config{Enabled: true},
    Limit: 100,  // 100 requests
    Window: time.Minute,  // per minute
    Algorithm: ratelimit.TokenBucket,
}

rateLimitMiddleware := ratelimit.NewMiddleware(config)
```

### 3. CORS (Priority: 100)

CORS completo com suporte a:
- Origens múltiplas e wildcards
- Preflight requests
- Credentials
- Headers customizados

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/cors"

config := cors.DefaultConfig()
config.AllowedOrigins = []string{"https://yourdomain.com", "*.yourdomain.com"}
config.AllowCredentials = true

corsMiddleware := cors.NewMiddleware(config)
```

### 4. Request/Response Logging (Priority: 50)

Logging estruturado com:
- Correlation IDs automáticos
- Captura de headers
- Métricas de performance
- IP real do cliente

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/logging"

config := logging.DefaultConfig()
config.Logger = func(entry logging.LogEntry) {
    // Seu logger customizado aqui
    log.Info("HTTP Request", 
        "method", entry.Method,
        "path", entry.Path,
        "status", entry.StatusCode,
        "duration", entry.Duration,
        "correlation_id", entry.CorrelationID)
}

loggingMiddleware := logging.NewMiddleware(config)
```

### 5. Compression (Priority: 800)

Compressão automática de resposta:
- Gzip e Deflate
- Compressão condicional por tipo MIME
- Tamanho mínimo configurável

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/compression"

config := compression.DefaultConfig()
config.Level = gzip.BestCompression
config.MinSize = 2048  // 2KB minimum

compressionMiddleware := compression.NewMiddleware(config)
```

### 6. Timeout Management (Priority: 150)

Timeout configurável por requisição:
- Context-based timeouts
- Handlers customizados para timeout
- Integração com outros middlewares

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/timeout"

config := timeout.DefaultConfig()
config.Timeout = 30 * time.Second
config.Message = "Request took too long"

timeoutMiddleware := timeout.NewMiddleware(config)
```

### 7. Bulkhead Pattern (Priority: 300)

Isolamento de recursos:
- Limitação de concorrência por recurso
- Filas configuráveis
- Métricas de uso

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/bulkhead"

config := bulkhead.DefaultConfig()
config.MaxConcurrent = 50
config.QueueSize = 100
config.ResourceKey = func(r *http.Request) string {
    return r.Header.Get("X-Service-Type") // Isolamento por serviço
}

bulkheadMiddleware := bulkhead.NewMiddleware(config)
```

### 8. Retry Policies (Priority: 400)

Retry automático para falhas transitórias:
- Exponential backoff
- Status codes configuráveis
- Métodos idempotentes apenas

```go
import "github.com/fsvxavier/nexs-lib/httpserver/middleware/retry"

config := retry.DefaultConfig()
config.MaxRetries = 3
config.InitialDelay = 100 * time.Millisecond
config.BackoffMultiplier = 2.0

retryMiddleware := retry.NewMiddleware(config)
```

## Uso Básico

### Exemplo Simples

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/fsvxavier/nexs-lib/httpserver/middleware"
    "github.com/fsvxavier/nexs-lib/httpserver/middleware/cors"
    "github.com/fsvxavier/nexs-lib/httpserver/middleware/logging"
    "github.com/fsvxavier/nexs-lib/httpserver/middleware/ratelimit"
)

func main() {
    // Criar cadeia de middleware
    chain := middleware.NewChain()
    
    // Adicionar middlewares (ordem automática por prioridade)
    chain.Add(cors.NewMiddleware(cors.DefaultConfig()))
    chain.Add(logging.NewMiddleware(logging.DefaultConfig()))
    
    rateLimitConfig := ratelimit.Config{
        Config: middleware.Config{Enabled: true},
        Limit: 1000,
        Window: time.Minute,
        Algorithm: ratelimit.TokenBucket,
    }
    chain.Add(ratelimit.NewMiddleware(rateLimitConfig))
    
    // Handler da aplicação
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello World"))
    })
    
    // Aplicar middleware
    server := &http.Server{
        Addr:    ":8080",
        Handler: chain.Then(handler),
    }
    
    server.ListenAndServe()
}
```

### Exemplo com Health Checks

```go
package main

import (
    "database/sql"
    "net/http"
    
    "github.com/fsvxavier/nexs-lib/httpserver/middleware/health"
)

func main() {
    // Setup health checks
    registry := health.NewRegistry()
    
    // Database health check
    registry.Register("database", health.DatabaseCheck(func(ctx context.Context) error {
        return db.PingContext(ctx)
    }), health.WithType(health.CheckTypeReadiness))
    
    // External service health check  
    registry.Register("payment-api", health.URLCheck("https://payments.api.com/health"))
    
    // Setup handlers
    healthHandler := health.NewHandler(registry)
    
    http.Handle("/health", healthHandler.HealthHandler())
    http.Handle("/health/live", healthHandler.LivenessHandler())
    http.Handle("/health/ready", healthHandler.ReadinessHandler())
    
    // Main application routes
    http.HandleFunc("/api/users", usersHandler)
    
    http.ListenAndServe(":8080", nil)
}
```

## Configuração de Produção

### Exemplo Completo para Produção

```go
func setupProductionMiddleware() http.Handler {
    chain := middleware.NewChain()
    
    // 1. CORS restritivo
    corsConfig := cors.DefaultConfig()
    corsConfig.AllowedOrigins = []string{"https://yourdomain.com"}
    corsConfig.AllowCredentials = true
    chain.Add(cors.NewMiddleware(corsConfig))
    
    // 2. Rate limiting agressivo
    rateLimitConfig := ratelimit.Config{
        Config: middleware.Config{Enabled: true},
        Limit: 1000,
        Window: time.Minute,
        Algorithm: ratelimit.TokenBucket,
    }
    chain.Add(ratelimit.NewMiddleware(rateLimitConfig))
    
    // 3. Logging detalhado
    loggingConfig := logging.DefaultConfig()
    loggingConfig.Logger = prodLogger
    chain.Add(logging.NewMiddleware(loggingConfig))
    
    // 4. Timeout conservador
    timeoutConfig := timeout.DefaultConfig()
    timeoutConfig.Timeout = 30 * time.Second
    chain.Add(timeout.NewMiddleware(timeoutConfig))
    
    // 5. Bulkhead por serviço
    bulkheadConfig := bulkhead.DefaultConfig()
    bulkheadConfig.MaxConcurrent = 100
    bulkheadConfig.ResourceKey = serviceTypeExtractor
    chain.Add(bulkhead.NewMiddleware(bulkheadConfig))
    
    // 6. Retry para resiliência
    retryConfig := retry.DefaultConfig()
    retryConfig.MaxRetries = 3
    chain.Add(retry.NewMiddleware(retryConfig))
    
    // 7. Compressão para eficiência
    compressionConfig := compression.DefaultConfig()
    chain.Add(compression.NewMiddleware(compressionConfig))
    
    return chain.Then(yourAppHandler)
}
```

## Context Keys

O sistema utiliza context keys padronizados:

```go
const (
    CorrelationIDKey    = "correlation_id"
    RequestStartTimeKey = "request_start_time" 
    BulkheadResourceKey = "bulkhead_resource"
)
```

## Métricas e Monitoramento

Cada middleware fornece métricas através de:
- Headers HTTP apropriados
- Context values
- Callbacks configuráveis
- Integration com sistemas de observabilidade

## Ordem de Execução

Os middlewares são executados em ordem de prioridade (menor número = executa primeiro):

1. **CORS** (100) - Headers CORS devem ser definidos primeiro
2. **Timeout** (150) - Timeout deve envolver toda a cadeia
3. **Rate Limiting** (200) - Bloquear requests não autorizados cedo
4. **Bulkhead** (300) - Controle de recursos
5. **Retry** (400) - Tentativas de retry
6. **Logging** (50) - Logging deve capturar tudo
7. **Compression** (800) - Compressão deve ser a última transformação

## Extensibilidade

Para criar middleware customizado:

```go
type CustomMiddleware struct {
    config CustomConfig
}

func (m *CustomMiddleware) Wrap(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Sua lógica aqui
        next.ServeHTTP(w, r)
    })
}

func (m *CustomMiddleware) Name() string { return "custom" }
func (m *CustomMiddleware) Priority() int { return 500 }
```

## Testes

Todos os middlewares incluem:
- Testes unitários completos
- Testes de integração  
- Benchmarks de performance
- Cobertura mínima de 98%

Execute os testes:

```bash
go test -race -timeout 30s -coverprofile=coverage.out ./...
go test -bench=. -benchmem ./...
```

## Próximos Passos

- [ ] Integração com métricas Prometheus
- [ ] Support para middleware assíncrono
- [ ] Circuit breaker pattern
- [ ] Cache middleware
- [ ] Authentication/Authorization middleware

## Exemplo Completo

Para exemplos completos demonstrando o sistema de middleware, consulte:

### Exemplo Simples
`examples/middleware/simple/main.go` - Demonstra uso básico com CORS, logging e rate limiting.

### Exemplo Avançado  
`examples/middleware/advanced/main.go` - Demonstra todos os middlewares em ação com health checks.

Estes exemplos incluem:
- Health checks configurados
- Cadeia completa de middleware
- Múltiplos endpoints de demonstração
- Configurações de produção e desenvolvimento
