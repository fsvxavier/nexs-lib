# HTTP Server Middlewares

Este pacote fornece um sistema completo e extensível de middlewares para o HTTP server, implementando padrões Observer, Hook e Middleware com funcionalidades avançadas de observabilidade, autenticação, compressão, rate limiting e health checks.

## Arquitetura

O sistema de middlewares é baseado em três padrões principais:

### 1. Middleware Pattern
- **Chain of Responsibility**: Processamento sequencial com prioridades
- **Composição**: Middlewares podem ser combinados facilmente
- **Interceptação**: Processamento antes e depois da execução
- **Thread Safety**: Operações atômicas para métricas

### 2. Observer Pattern
- **Event Propagation**: Notificação de eventos do ciclo de vida
- **Desacoplamento**: Componentes independentes
- **Extensibilidade**: Fácil adição de novos observadores

### 3. Hook Pattern
- **Lifecycle Events**: Pontos específicos de interceptação
- **Configurabilidade**: Hooks ativados/desativados individualmente
- **Métricas**: Coleta automática de dados de performance

## Componentes Implementados

### Core Infrastructure (`middlewares.go`)
- **MiddlewareManager**: Gerenciamento central de middlewares
- **BaseMiddleware**: Classe base com funcionalidades comuns
- **ObserverManager**: Gerenciamento de observadores
- **Chain Processing**: Execução ordenada por prioridade

### Middlewares Disponíveis

#### 1. Logging Middleware (`logging.go`)
**Funcionalidades:**
- Logging de requests/responses/errors
- Múltiplos formatos (JSON, Text, Structured)
- Sanitização de dados sensíveis
- Filtragem por paths/métodos
- Rate limiting de logs

**Configuração:**
```go
config := LoggingConfig{
    LogRequests:     true,
    LogResponses:    true,
    LogHeaders:      true,
    LogBody:        false,
    Format:         LogFormatJSON,
    MaxBodySize:    1024,
    SensitiveHeaders: []string{"Authorization", "Cookie"},
}
middleware := NewLoggingMiddlewareWithConfig(1, config)
```

**Métricas:**
- Requests logados
- Responses logados
- Errors logados
- Logs falhados
- Tempo de atividade

#### 2. Authentication Middleware (`auth.go`)
**Funcionalidades:**
- Multiple auth methods (Basic, Bearer, API Key)
- Pluggable auth providers
- User context injection
- Rate limiting de autenticação
- Token validation

**Configuração:**
```go
config := AuthConfig{
    EnableBasicAuth:  true,
    EnableBearerAuth: true,
    BasicAuthUsers: map[string]string{
        "admin": "secret",
    },
    ValidTokens: map[string]AuthUser{
        "token123": {ID: "user1", Username: "john"},
    },
    RequireAuth: true,
}
middleware := NewAuthMiddlewareWithConfig(2, config)
```

**Providers:**
- **BasicAuthProvider**: HTTP Basic Authentication
- **APIKeyAuthProvider**: API Key validation
- **Custom Providers**: Interface extensível

**Métricas:**
- Tentativas de autenticação
- Sucessos/falhas
- Taxa de sucesso
- Validações de token

#### 3. CORS Middleware (`cors.go`)
**Funcionalidades:**
- Origin validation
- Preflight request handling
- Credential support
- Header configuration
- Wildcard support

**Configuração:**
```go
config := CORSConfig{
    AllowedOrigins:   []string{"https://example.com", "*.mydomain.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:          86400,
}
middleware := NewCORSMiddlewareWithConfig(3, config)
```

**Features:**
- Preflight caching
- Origin pattern matching
- Security headers
- Debug mode

**Métricas:**
- Requests CORS
- Preflight requests
- Origins permitidos/bloqueados
- Taxa de aprovação

#### 4. Compression Middleware (`compression.go`)
**Funcionalidades:**
- Multiple algorithms (gzip, deflate, brotli)
- Content-type filtering
- Size constraints
- Quality negotiation
- Performance optimization

**Configuração:**
```go
config := CompressionConfig{
    EnableGzip:    true,
    EnableDeflate: true,
    EnableBrotli:  false,
    GzipLevel:     6,
    MinSize:       1024,
    MaxSize:       1048576,
    CompressibleTypes: []string{
        "text/html", "text/css", "application/json",
    },
}
middleware := NewCompressionMiddlewareWithConfig(4, config)
```

**Algorithms:**
- **Gzip**: Padrão para web
- **Deflate**: Compatibilidade legada
- **Brotli**: Melhor compressão (opcional)

**Métricas:**
- Total de requests
- Responses comprimidas
- Taxa de compressão
- Bytes economizados
- Tempo de processamento

#### 5. Rate Limiting Middleware (`rate_limit.go`)
**Funcionalidades:**
- Multiple algorithms (Token Bucket, Fixed Window, Sliding Window)
- Flexible identification (IP, User, Header, Custom)
- Path-specific limits
- Distributed support
- Memory management

**Configuração:**
```go
config := RateLimitConfig{
    Strategy:          TokenBucket,
    RequestsPerSecond: 10.0,
    RequestsPerMinute: 600,
    BurstSize:         20,
    IdentifyByIP:      true,
    PathLimits: map[string]RateLimit{
        "/api/upload": {RequestsPerSecond: 1.0},
    },
}
middleware := NewRateLimitMiddlewareWithConfig(5, config)
```

**Strategies:**
- **Token Bucket**: Burst handling
- **Fixed Window**: Simple implementation
- **Sliding Window**: Precise rate limiting
- **Leaky Bucket**: Smooth rate limiting

**Métricas:**
- Total requests
- Requests permitidos/bloqueados
- Taxa de aprovação
- Limiters ativos

#### 6. Health Check Middleware (`health_check.go`)
**Funcionalidades:**
- Multiple endpoints (/health, /health/live, /health/ready)
- Pluggable health checkers
- Parallel/sequential execution
- Caching system
- Detailed reporting

**Configuração:**
```go
config := HealthCheckConfig{
    HealthPath:       "/health",
    LivenessPath:     "/health/live",
    ReadinessPath:    "/health/ready",
    DetailedResponse: true,
    IncludeMetrics:   true,
    ParallelChecks:   true,
    CacheTimeout:     time.Second * 5,
}
middleware := NewHealthCheckMiddlewareWithConfig(6, config)

// Add checkers
middleware.AddChecker(NewDatabaseHealthChecker("database", true))
middleware.AddChecker(NewServiceHealthChecker("api", "https://api.example.com", false))
middleware.AddChecker(NewMemoryHealthChecker("memory", 0.8, true))
```

**Checkers:**
- **DatabaseHealthChecker**: Database connectivity
- **ServiceHealthChecker**: External service health
- **MemoryHealthChecker**: Memory usage monitoring
- **Custom Checkers**: Interface extensível

**Métricas:**
- Total checks
- Successful/failed checks
- Taxa de sucesso
- Tempo desde último check

## Uso Básico

### 1. Configuração do Manager
```go
manager := NewMiddlewareManager()

// Configurar logger
logger := &CustomLogger{}
manager.SetLogger(logger)
```

### 2. Adicionar Middlewares
```go
// Logging (prioridade 1 - executa primeiro)
logging := NewLoggingMiddleware(1)
manager.AddMiddleware(logging)

// Authentication (prioridade 2)
auth := NewAuthMiddleware(2)
manager.AddMiddleware(auth)

// CORS (prioridade 3)
cors := NewCORSMiddleware(3)
manager.AddMiddleware(cors)

// Compression (prioridade 4)
compression := NewCompressionMiddleware(4)
manager.AddMiddleware(compression)

// Rate Limiting (prioridade 5)
rateLimit := NewRateLimitMiddleware(5)
manager.AddMiddleware(rateLimit)

// Health Check (prioridade 6 - executa último)
healthCheck := NewHealthCheckMiddleware(6)
manager.AddMiddleware(healthCheck)
```

### 3. Processar Requests
```go
ctx := context.Background()
req := map[string]interface{}{
    "method": "GET",
    "path":   "/api/users",
    "headers": map[string]string{
        "Authorization": "Bearer token123",
        "Accept":        "application/json",
    },
}

response, err := manager.ProcessRequest(ctx, req)
if err != nil {
    log.Printf("Error processing request: %v", err)
    return
}

log.Printf("Response: %+v", response)
```

## Observabilidade

### Métricas Disponíveis
Cada middleware expõe métricas específicas através do método `GetMetrics()`:

```go
// Métricas globais do manager
managerMetrics := manager.GetMetrics()

// Métricas específicas de cada middleware
loggingMetrics := logging.GetMetrics()
authMetrics := auth.GetMetrics()
corsMetrics := cors.GetMetrics()
compressionMetrics := compression.GetMetrics()
rateLimitMetrics := rateLimit.GetMetrics()
healthMetrics := healthCheck.GetMetrics()
```

### Logging Estruturado
Todos os middlewares utilizam logging estruturado com níveis configuráveis:

```go
type Logger interface {
    Info(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    Debug(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
}
```

### Event Propagation
O sistema suporta propagação de eventos através do Observer Pattern:

```go
type Observer interface {
    OnStart(ctx context.Context, addr string) error
    OnStop(ctx context.Context) error
    OnError(ctx context.Context, err error) error
    OnRequest(ctx context.Context, req interface{}) error
    OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error
    OnRouteEnter(ctx context.Context, method, path string, req interface{}) error
    OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error
}
```

## Extensibilidade

### Criando Middlewares Customizados
```go
type CustomMiddleware struct {
    *BaseMiddleware
    // Custom fields
}

func NewCustomMiddleware(priority int) *CustomMiddleware {
    return &CustomMiddleware{
        BaseMiddleware: NewBaseMiddleware("custom", priority),
    }
}

func (cm *CustomMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
    // Pre-processing
    cm.GetLogger().Debug("Custom middleware processing")
    
    // Call next middleware
    resp, err := next(ctx, req)
    
    // Post-processing
    cm.GetLogger().Debug("Custom middleware completed")
    
    return resp, err
}
```

### Criando Auth Providers Customizados
```go
type CustomAuthProvider struct {
    name    string
    enabled bool
}

func (cap *CustomAuthProvider) Authenticate(ctx context.Context, credentials interface{}) (*AuthUser, error) {
    // Custom authentication logic
    return &AuthUser{ID: "custom", Username: "user"}, nil
}

func (cap *CustomAuthProvider) ValidateToken(ctx context.Context, token string) (*AuthUser, error) {
    // Custom token validation logic
    return &AuthUser{ID: "custom", Username: "user"}, nil
}

func (cap *CustomAuthProvider) GetName() string {
    return cap.name
}

func (cap *CustomAuthProvider) IsEnabled() bool {
    return cap.enabled
}
```

### Criando Health Checkers Customizados
```go
type CustomHealthChecker struct {
    name     string
    critical bool
    timeout  time.Duration
}

func (chc *CustomHealthChecker) Check(ctx context.Context) HealthCheckResult {
    startTime := time.Now()
    
    // Custom health check logic
    
    return HealthCheckResult{
        Name:      chc.name,
        Status:    HealthStatusHealthy,
        Message:   "Custom check passed",
        Duration:  time.Since(startTime),
        Timestamp: startTime,
    }
}

func (chc *CustomHealthChecker) Name() string {
    return chc.name
}

func (chc *CustomHealthChecker) IsCritical() bool {
    return chc.critical
}

func (chc *CustomHealthChecker) GetTimeout() time.Duration {
    return chc.timeout
}
```

## Configurações Avançadas

### Thread Safety
Todos os middlewares são thread-safe e utilizam operações atômicas para métricas:

```go
// Operações atômicas para contadores
atomic.AddInt64(&middleware.requestCount, 1)
atomic.LoadInt64(&middleware.requestCount)
```

### Memory Management
O sistema inclui limpeza automática de recursos:

```go
// Rate limiting cleanup
middleware.Reset() // Reset metrics
middleware.Stop()  // Stop background routines
```

### Performance Optimization
- Caching de resultados (health checks)
- Pool de buffers (compression)
- Cleanup automático (rate limiting)
- Lazy initialization
- Memory limits

## Testes

O sistema inclui testes abrangentes para todos os componentes:

```bash
# Executar todos os testes
go test ./middlewares/...

# Executar com coverage
go test -cover ./middlewares/...

# Executar testes específicos
go test ./middlewares/ -run TestMiddlewareManager
go test ./middlewares/ -run TestLoggingMiddleware
go test ./middlewares/ -run TestAuthMiddleware
```

## Exemplo Completo

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/fsvxavier/nexs-lib/httpserver/middlewares"
)

func main() {
    // Create middleware manager
    manager := middlewares.NewMiddlewareManager()
    
    // Configure middlewares
    setupMiddlewares(manager)
    
    // Simulate request processing
    ctx := context.Background()
    req := map[string]interface{}{
        "method": "POST",
        "path":   "/api/users",
        "headers": map[string]string{
            "Authorization":  "Basic dXNlcjpwYXNz", // user:pass in base64
            "Content-Type":   "application/json",
            "Accept-Encoding": "gzip, deflate",
        },
        "body": `{"name": "John Doe", "email": "john@example.com"}`,
    }
    
    response, err := manager.ProcessRequest(ctx, req)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    
    log.Printf("Response: %+v", response)
    
    // Print metrics
    printMetrics(manager)
}

func setupMiddlewares(manager *middlewares.MiddlewareManager) {
    // 1. Logging
    loggingConfig := middlewares.DefaultLoggingConfig()
    loggingConfig.LogBody = true
    logging := middlewares.NewLoggingMiddlewareWithConfig(1, loggingConfig)
    manager.AddMiddleware(logging)
    
    // 2. Authentication
    authConfig := middlewares.DefaultAuthConfig()
    authConfig.BasicAuthUsers = map[string]string{"user": "pass"}
    auth := middlewares.NewAuthMiddlewareWithConfig(2, authConfig)
    manager.AddMiddleware(auth)
    
    // 3. CORS
    cors := middlewares.NewCORSMiddleware(3)
    manager.AddMiddleware(cors)
    
    // 4. Compression
    compression := middlewares.NewCompressionMiddleware(4)
    manager.AddMiddleware(compression)
    
    // 5. Rate Limiting
    rateLimit := middlewares.NewRateLimitMiddleware(5)
    manager.AddMiddleware(rateLimit)
    
    // 6. Health Check
    healthCheck := middlewares.NewHealthCheckMiddleware(6)
    healthCheck.AddChecker(middlewares.NewDatabaseHealthChecker("database", true))
    healthCheck.AddChecker(middlewares.NewMemoryHealthChecker("memory", 0.8, false))
    manager.AddMiddleware(healthCheck)
}

func printMetrics(manager *middlewares.MiddlewareManager) {
    log.Println("=== Middleware Metrics ===")
    
    middlewares := manager.ListMiddlewares()
    for _, name := range middlewares {
        middleware, _ := manager.GetMiddleware(name)
        if metricsProvider, ok := middleware.(interface{ GetMetrics() map[string]interface{} }); ok {
            metrics := metricsProvider.GetMetrics()
            log.Printf("%s: %+v", name, metrics)
        }
    }
}
```

## Conclusão

Este sistema de middlewares fornece uma base sólida e extensível para processamento de requests HTTP com:

- **Performance**: Operações otimizadas e thread-safe
- **Observabilidade**: Métricas detalhadas e logging estruturado
- **Segurança**: Autenticação, CORS e rate limiting
- **Extensibilidade**: Interfaces bem definidas para customização
- **Manutenibilidade**: Código limpo e bem documentado
- **Testabilidade**: Cobertura de testes abrangente

O sistema está pronto para uso em produção e pode ser facilmente integrado com qualquer framework HTTP em Go.
