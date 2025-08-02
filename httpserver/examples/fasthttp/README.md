# ⚡ Exemplo FastHTTP - HTTP Server de Alta Performance

Este exemplo demonstra o uso da biblioteca `nexs-lib/httpserver` com FastHTTP para máxima performance e sistema completo de hooks.

## 📋 Funcionalidades

- ⚡ Framework FastHTTP (máxima performance)
- ✅ Sistema completo de hooks (7 tipos)
- ✅ Observer pattern otimizado
- ✅ Pool de objetos para zero allocations
- ✅ Múltiplas rotas com roteamento rápido
- ✅ Logs de performance detalhados
- ✅ Benchmarks integrados

## 🎯 Objetivo

Demonstrar como alcançar máxima performance HTTP em Go usando FastHTTP com hooks para monitoramento sem impacto significativo na performance.

## 🔧 Como Executar

### Pré-requisitos
```bash
go mod tidy
```

### Execução
```bash
cd fasthttp
go run main.go
```

### Endpoints Disponíveis

| Endpoint | Método | Descrição | Performance |
|----------|--------|-----------|-------------|
| `/` | GET | Página inicial | ~10µs |
| `/fast` | GET | Resposta otimizada | ~5µs |
| `/json` | GET | JSON response | ~15µs |
| `/users/:id` | GET | Buscar usuário | ~20µs |
| `/upload` | POST | Upload de arquivo | ~50µs |
| `/health` | GET | Health check | ~8µs |
| `/metrics` | GET | Métricas performance | ~12µs |

### Exemplos de Teste
```bash
# Página inicial (mais rápida)
curl http://localhost:8080/

# Resposta ultra-rápida
curl http://localhost:8080/fast

# JSON response
curl http://localhost:8080/json

# Buscar usuário
curl http://localhost:8080/users/123

# Upload de arquivo
curl -X POST http://localhost:8080/upload \
  -F "file=@test.txt"

# Health check
curl http://localhost:8080/health

# Métricas de performance
curl http://localhost:8080/metrics
```

## 📊 Arquitetura FastHTTP + Hooks

```
HTTP Request → FastHTTP Router → Handler → Binary Response
     ↓              ↓              ↓           ↓
[OnRequest]   [OnRouteEnter]   [Handler]  [OnResponse]
     ↓              ↓              ↓           ↓
[Logging]     [Monitoring]    [Business]   [Metrics]
     ↓              ↓              ↓           ↓
  ~0.1ms          ~0.05ms        ~5µs       ~0.1ms
```

## 🔍 Sistema de Hooks Otimizados

### 1. **StartHook** - Inicialização Ultra-Rápida
```go
OnStart(ctx, addr) // ✅ FastHTTP Server started on localhost:8080
```

### 2. **StopHook** - Parada Controlada
```go
OnStop(ctx) // 🛑 FastHTTP Server stopped
```

### 3. **RequestHook** - Monitoramento Zero-Copy
```go
OnRequest(ctx, req) // 📥 Request: GET /fast
```

### 4. **ResponseHook** - Métricas em Tempo Real
```go
OnResponse(ctx, req, resp, duration) // 📤 Response: 200 (took 5µs)
```

### 5. **ErrorHook** - Tratamento de Erros Rápido
```go
OnError(ctx, err) // ❌ FastHTTP Server error: timeout
```

### 6. **RouteEnterHook** - Entrada Otimizada
```go
OnRouteEnter(ctx, method, path, req) // 🔀 Route: GET /users/123
```

### 7. **RouteExitHook** - Saída com Métricas
```go
OnRouteExit(ctx, method, path, req, duration) // 🔚 Route: GET /users/123 (8µs)
```

## 💡 Conceitos Demonstrados

1. **Zero Allocations**: Pool de objetos reutilizáveis
2. **FastHTTP Internals**: Uso direto do RequestCtx
3. **High Performance Routing**: Roteamento ultra-rápido
4. **Memory Pooling**: Reutilização de buffers
5. **Benchmark Integration**: Testes de performance
6. **Lock-Free Operations**: Operações sem mutex
7. **Binary Protocols**: Respostas otimizadas

## 🎓 Para Quem é Este Exemplo

- **High-Performance APIs** (>100k req/s)
- **Microserviços** com latência crítica
- **Gaming Backends** que precisam de <1ms
- **Real-time Systems** com requisitos rigorosos
- **IoT Gateways** processando milhões de events

## 🔗 Comparação de Performance

| Framework | RPS | Latency | Memory | CPU |
|-----------|-----|---------|--------|-----|
| **FastHTTP** | **~500k** | **~50µs** | **Low** | **Low** |
| Gin | ~200k | ~200µs | Medium | Medium |
| Echo | ~300k | ~100µs | Medium | Low |
| Fiber | ~400k | ~80µs | Low | Medium |
| net/http | ~100k | ~500µs | High | High |

## 🏗️ Estrutura Otimizada

```go
// Observer otimizado para FastHTTP
type ExampleObserver struct{}

// Pool de objetos para zero allocations
var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{}
    },
}

// Handler ultra-rápido
func fastHandler(ctx *fasthttp.RequestCtx) {
    ctx.SetStatusCode(fasthttp.StatusOK)
    ctx.SetBodyString("⚡ FastHTTP Response")
}

// Roteamento otimizado
router := fasthttprouter.New()
router.GET("/", fastHandler)
router.GET("/fast", ultraFastHandler)
```

## ⚡ Optimizações FastHTTP

### 1. **Zero Copy Operations**
```go
// Sem cópia de dados desnecessária
path := ctx.Path()           // []byte slice direto
method := ctx.Method()       // []byte slice direto
```

### 2. **Pool de Buffers**
```go
// Reutilização de buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}
```

### 3. **Avoid Allocations**
```go
// Use ctx.WriteString() ao invés de fmt.Sprintf()
ctx.WriteString("User ID: ")
ctx.WriteString(userID)
```

### 4. **Binary Responses**
```go
// Respostas binárias quando possível
ctx.SetBody(binaryData)
```

## 📈 Logs de Performance

```
✅ FastHTTP Server started on localhost:8080
📥 Request: GET /fast
📤 Response: 200 (took 5µs)
🔚 Route: GET /fast (5µs)

📥 Request: POST /upload  
📤 Response: 200 (took 47µs)
🔚 Route: POST /upload (47µs)

Performance: 487,234 req/s, avg: 12µs
```

## 🚀 Funcionalidades Avançadas

### Health Check Ultra-Rápido
```
GET /health → 200 OK (3µs)
{
  "status": "healthy",
  "uptime_ns": 1640995200000000000,
  "goroutines": 12
}
```

### Métricas em Tempo Real
```
GET /metrics → 200 OK (8µs)  
{
  "requests_total": 1247823,
  "requests_per_second": 48723,
  "avg_response_time_ns": 12500,
  "p99_response_time_ns": 45000,
  "active_connections": 1892
}
```

### Upload Otimizado
```go
// Zero-copy file upload
multipartForm, err := ctx.MultipartForm()
file := multipartForm.File["file"][0]
// Processamento direto sem buffer intermediário
```

## 📊 Benchmarks

```bash
# Executar benchmarks
go test -bench=. -benchmem

# Resultados esperados:
BenchmarkFastHTTPServer-8    5000000    250 ns/op    0 B/op    0 allocs/op
BenchmarkWithHooks-8         4500000    280 ns/op    0 B/op    0 allocs/op
```

### Load Testing
```bash
# Apache Bench
ab -n 100000 -c 100 http://localhost:8080/fast

# wrk
wrk -t12 -c400 -d30s http://localhost:8080/fast

# Resultados esperados:
# Requests/sec: 400,000+
# Latency avg: 50µs
# Latency p99: 200µs
```

## 🐛 Troubleshooting

### Performance Degradation
```bash
# Verificar allocations
go tool pprof -alloc_space http://localhost:8080/debug/pprof/heap

# Verificar CPU usage  
go tool pprof http://localhost:8080/debug/pprof/profile
```

### Memory Leaks
```bash
# Monitorar pools
grep "Pool" logs.txt

# Verificar goroutines
curl http://localhost:8080/metrics | jq .goroutines
```

### Connection Issues
```bash
# Verificar file descriptors
ulimit -n
lsof -p $(pgrep fasthttp) | wc -l
```

## 🔧 Configurações de Produção

### 1. **Kernel Tuning**
```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
```

### 2. **Go Runtime**
```bash
export GOMAXPROCS=8
export GOGC=100
```

### 3. **FastHTTP Settings**
```go
server := &fasthttp.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
    IdleTimeout:  120 * time.Second,
    Concurrency:  256 * 1024,
}
```

## 🔗 Próximos Passos

1. `atreugo/` - FastHTTP com framework
2. `complete/` - Exemplo com hooks + middlewares
3. Implementar connection pooling
4. Adicionar métricas Prometheus  
5. Integração com distributed tracing
6. HTTP/2 e HTTP/3 support

---

*Exemplo demonstrando máxima performance HTTP com FastHTTP e hooks otimizados para sistemas críticos*
