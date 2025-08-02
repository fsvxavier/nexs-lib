# 🎯 Exemplo Advanced - HTTP Server Multi-Provider

Este exemplo demonstra uso avançado da biblioteca `nexs-lib/httpserver` com múltiplos conceitos integrados e padrões de produção.

## 📋 Funcionalidades

- 🔄 Múltiplos providers (Fiber como principal)
- ✅ Sistema completo de hooks (7 tipos)
- ✅ Observer pattern avançado
- ✅ Graceful shutdown com signal handling
- ✅ Configuração modular e extensível
- ✅ Logs estruturados com níveis
- ✅ Health checks e metrics

## 🎯 Objetivo

Demonstrar padrões avançados de uso em produção, incluindo graceful shutdown, signal handling, e observabilidade completa.

## 🔧 Como Executar

### Pré-requisitos
```bash
go mod tidy
```

### Execução
```bash
cd advanced
go run main.go
```

### Teste de Graceful Shutdown
```bash
# Em outro terminal
curl http://localhost:8080/users & 
curl http://localhost:8080/health &
kill -TERM $(pgrep -f "go run main.go")
```

### Endpoints Disponíveis

| Endpoint | Método | Descrição | Features |
|----------|--------|-----------|----------|
| `/` | GET | Página inicial | Basic response |
| `/users` | GET | Lista usuários | JSON + hooks |
| `/users/:id` | GET | Buscar usuário | Path params |
| `/health` | GET | Health check | System status |
| `/metrics` | GET | Server metrics | Performance data |
| `/config` | GET | Server config | Runtime config |

### Exemplos de Teste
```bash
# Página inicial
curl http://localhost:8080/

# Listar usuários
curl http://localhost:8080/users

# Buscar usuário específico
curl http://localhost:8080/users/789

# Health check completo
curl http://localhost:8080/health

# Métricas do servidor
curl http://localhost:8080/metrics

# Configuração atual
curl http://localhost:8080/config
```

## 📊 Arquitetura Advanced Pattern

```
Signal Handler → Graceful Shutdown → Hook Cleanup
      ↓               ↓                    ↓
HTTP Request → Router → Middleware → Handler → Response
      ↓         ↓         ↓          ↓        ↓
  [OnRequest] [Route] [Logging] [Business] [OnResponse]
      ↓         ↓         ↓          ↓        ↓
  [Observer] [Metrics] [Audit] [Validation] [Cleanup]
```

## 🔍 Sistema de Hooks Advanced

### 1. **StartHook** - Inicialização Robusta
```go
OnStart(ctx, addr) // 🚀 Server started on localhost:8080
// + System health checks
// + Resource allocation
// + Service registration
```

### 2. **StopHook** - Shutdown Graceful
```go
OnStop(ctx) // 🛑 Server stopped
// + Connection draining
// + Resource cleanup  
// + Service deregistration
```

### 3. **RequestHook** - Request Lifecycle
```go
OnRequest(ctx, req) // 📨 Request received
// + Request ID generation
// + Rate limiting check
// + Security validation
```

### 4. **ResponseHook** - Response Analysis
```go
OnResponse(ctx, req, resp, duration) // 📤 Response sent in 1.5ms
// + Performance metrics
// + SLA tracking
// + Error rate calculation
```

### 5. **ErrorHook** - Error Management
```go
OnError(ctx, err) // ❌ Server error: database timeout
// + Error categorization
// + Alert triggering
// + Incident tracking
```

### 6. **RouteEnterHook** - Route Monitoring
```go
OnRouteEnter(ctx, method, path, req) // 🔄 Route: GET /users
// + Route-specific metrics
// + Access logging
// + Authorization checks
```

### 7. **RouteExitHook** - Route Completion
```go
OnRouteExit(ctx, method, path, req, duration) // 🔄 Route: GET /users (1.2ms)
// + Route performance
// + Business metrics
// + Audit logging
```

## 💡 Conceitos Demonstrados

1. **Signal Handling**: SIGTERM/SIGINT para graceful shutdown
2. **Context Propagation**: Context cancelation em toda stack
3. **Resource Management**: Cleanup automático de recursos
4. **Observer Pattern**: Logging estruturado e extensível
5. **Health Checks**: Verificação de dependências externas
6. **Metrics Collection**: Coleta de métricas de performance
7. **Configuration Management**: Configuração modular

## 🎓 Para Quem é Este Exemplo

- **Aplicações de produção** que precisam de robustez
- **Microserviços** com requirements enterprise
- **DevOps teams** implementando observabilidade
- **Architects** desenhando sistemas resilientes

## 🔗 Diferencial vs Outros Exemplos

| Característica | basic | gin | **advanced** |
|----------------|-------|-----|--------------|
| Graceful Shutdown | ❌ | ❌ | **✅** |
| Signal Handling | ❌ | ❌ | **✅** |
| Health Checks | ❌ | ❌ | **✅** |
| Metrics Collection | ❌ | ❌ | **✅** |
| Config Management | ❌ | ❌ | **✅** |
| Error Categorization | ❌ | ❌ | **✅** |
| Resource Cleanup | ❌ | ❌ | **✅** |

## 🏗️ Estrutura Advanced

```go
// Observer com contexto de produção
type LoggingObserver struct {
    logger *log.Logger
    metrics *MetricsCollector
    errorTracker *ErrorTracker
}

// Graceful shutdown handler
func setupGracefulShutdown(server Server, observer *LoggingObserver) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("📤 Graceful shutdown initiated...")
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.Stop(ctx); err != nil {
            log.Printf("❌ Shutdown error: %v", err)
        } else {
            log.Println("✅ Server shut down successfully")
        }
        
        os.Exit(0)
    }()
}

// Health check avançado
func healthHandler(c *fiber.Ctx) error {
    status := checkSystemHealth()
    if !status.Healthy {
        return c.Status(503).JSON(status)
    }
    return c.JSON(status)
}
```

## 📈 Logs Advanced

```
🚀 Server started on localhost:8080
📊 System resources initialized
🔍 Health checks passed
✅ Ready to accept connections

📨 Request received (ID: req-001)
🔄 Route: GET /users  
📤 Response sent in 1.2ms (ID: req-001)
📊 Metrics updated: requests=1247, avg_time=1.3ms

📤 Graceful shutdown initiated...
⏳ Draining connections (30s timeout)...
🔒 All connections closed
✅ Server shut down successfully
```

## 🚀 Funcionalidades Production-Ready

### Health Check Abrangente
```json
{
  "status": "healthy",
  "timestamp": "2025-08-02T15:30:45Z",
  "uptime": "2h15m32s",
  "checks": {
    "server": "healthy",
    "memory": "healthy",
    "database": "healthy", 
    "external_apis": "healthy"
  },
  "metrics": {
    "goroutines": 15,
    "memory_mb": 24.5,
    "requests_total": 15623
  }
}
```

### Métricas Detalhadas
```json
{
  "requests": {
    "total": 15623,
    "per_second": 143.2,
    "errors": 12,
    "error_rate": 0.077
  },
  "response_times": {
    "avg": "1.4ms",
    "p50": "1.1ms", 
    "p95": "3.2ms",
    "p99": "8.7ms"
  },
  "routes": {
    "/users": {"requests": 8934, "avg_time": "1.1ms"},
    "/health": {"requests": 6542, "avg_time": "0.3ms"}
  },
  "system": {
    "goroutines": 15,
    "memory_mb": 24.5,
    "gc_cycles": 89,
    "uptime": "2h15m"
  }
}
```

### Configuração Runtime
```json
{
  "server": {
    "provider": "fiber",
    "host": "localhost",
    "port": 8080,
    "timeout": "30s"
  },
  "hooks": {
    "enabled": true,
    "types": ["start", "stop", "request", "response", "error", "route_enter", "route_exit"]
  },
  "features": {
    "graceful_shutdown": true,
    "health_checks": true,
    "metrics": true,
    "signal_handling": true
  }
}
```

## 📊 Performance & Monitoring

### Sistema de Métricas
```go
type MetricsCollector struct {
    RequestsTotal    int64
    RequestsPerSec   float64
    AvgResponseTime  time.Duration
    ErrorRate        float64
    ActiveConnections int32
}

// Coletado automaticamente pelos hooks
func (m *MetricsCollector) RecordRequest(duration time.Duration, err error) {
    atomic.AddInt64(&m.RequestsTotal, 1)
    if err != nil {
        m.incrementErrorRate()
    }
    m.updateAvgResponseTime(duration)
}
```

### Error Tracking
```go
type ErrorTracker struct {
    Errors map[string]int
    LastErrors []string
}

// Categorização automática de erros
func (e *ErrorTracker) TrackError(err error) {
    category := categorizeError(err)
    e.Errors[category]++
    e.LastErrors = append(e.LastErrors, err.Error())
}
```

## 🐛 Troubleshooting Advanced

### Graceful Shutdown Issues
```bash
# Verificar se processo responde a SIGTERM
kill -TERM $(pgrep advanced)

# Verificar timeout de shutdown
grep "Graceful shutdown" logs.txt
```

### Memory Leaks
```bash
# Monitorar crescimento de memória
curl http://localhost:8080/metrics | jq .system.memory_mb

# Profile de memória
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Performance Degradation
```bash
# Analisar métricas por rota
curl http://localhost:8080/metrics | jq .routes

# Verificar goroutines leak
curl http://localhost:8080/metrics | jq .system.goroutines
```

## 🔧 Production Deployment

### 1. **Container Configuration**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o advanced main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/advanced .
EXPOSE 8080
CMD ["./advanced"]
```

### 2. **Kubernetes Deployment**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: advanced-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: server
        image: advanced-server:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 3. **Monitoring Stack**
```yaml
# Prometheus config
- job_name: 'advanced-server'
  static_configs:
  - targets: ['advanced-server:8080']
  metrics_path: '/metrics'
  scrape_interval: 15s
```

## 🔗 Próximos Passos

1. `complete/` - Exemplo com hooks + middlewares completos
2. Implementar distributed tracing
3. Adicionar circuit breaker
4. Implementar rate limiting avançado
5. Integração com service mesh
6. Adicionar A/B testing capability

---

*Exemplo avançado demonstrando padrões de produção com graceful shutdown, observabilidade completa e robustez empresarial*
