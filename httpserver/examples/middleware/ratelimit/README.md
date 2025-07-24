# Rate Limiting Middleware Example

Este exemplo demonstra o uso do middleware de rate limiting da nexs-lib com diferentes algoritmos e configurações.

## 🚦 Sobre Rate Limiting

O rate limiting é uma técnica essencial para:
- Proteger APIs contra abuso
- Garantir fair usage entre usuários
- Prevenir ataques DDoS
- Manter qualidade de serviço

## 🚀 Executando o Exemplo

```bash
go run main.go
```

O servidor iniciará na porta `:8080` com diferentes políticas de rate limiting.

## 📍 Endpoints

| Endpoint | Rate Limit | Algoritmo | Chave |
|----------|------------|-----------|-------|
| `GET /health` | Sem limite | - | - |
| `GET /api/test` | 10 req/min | Token Bucket | IP |
| `GET /strict` | 5 req/min | Sliding Window | User ID ou IP |

## 🔧 Algoritmos Implementados

### 1. Token Bucket (Endpoint `/api/test`)
```go
rateLimitConfig := ratelimit.Config{
    Enabled:   true,
    Limit:     10, // 10 requests per minute
    Window:    time.Minute,
    Algorithm: ratelimit.TokenBucket,
    KeyGenerator: func(r *http.Request) string {
        return r.RemoteAddr // Rate limit by IP
    },
}
```

**Características:**
- ✅ Permite rajadas de tráfego (burst)
- ✅ Reposição constante de tokens
- ✅ Ideal para APIs que podem lidar com picos

### 2. Sliding Window (Endpoint `/strict`)
```go
slidingWindowConfig := ratelimit.Config{
    Enabled:   true,
    Limit:     5, // 5 requests per minute
    Window:    time.Minute,
    Algorithm: ratelimit.SlidingWindow,
    KeyGenerator: func(r *http.Request) string {
        // Rate limit by user or IP
        userID := r.Header.Get("X-User-ID")
        if userID != "" {
            return "user:" + userID
        }
        return "ip:" + r.RemoteAddr
    },
}
```

**Características:**
- ✅ Distribuição uniforme no tempo
- ✅ Previne rajadas de tráfego
- ✅ Mais rigoroso que token bucket

## 🧪 Testando

### Teste Básico
```bash
# Endpoint sem rate limit
curl http://localhost:8080/health

# Endpoint com rate limiting normal
curl http://localhost:8080/api/test

# Endpoint strict
curl http://localhost:8080/strict
```

### Teste de Rate Limiting

#### Token Bucket (10 req/min)
```bash
# Teste multiple requests rapidamente
for i in {1..15}; do
    echo "Request $i:"
    curl -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -s http://localhost:8080/api/test
    echo ""
done
```

#### Sliding Window (5 req/min)
```bash
# Teste como usuário específico
for i in {1..8}; do
    echo "Request $i (with User-ID):"
    curl -H 'X-User-ID: user123' \
         -w "Status: %{http_code}\n" \
         -s http://localhost:8080/strict
done
```

### Headers de Resposta

Quando o rate limit é aplicado, você verá headers como:
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1642694460
```

### Resposta de Rate Limit Excedido
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests, please try again later",
  "retry_after": 60
}
```

## 🔍 Strategies de Chave

### Por IP Address
```go
KeyGenerator: func(r *http.Request) string {
    return r.RemoteAddr
}
```
- Simples e efetivo
- Funciona para usuários não autenticados

### Por User ID
```go
KeyGenerator: func(r *http.Request) string {
    return r.Header.Get("X-User-ID")
}
```
- Controle granular por usuário
- Requer autenticação

### Híbrido (User + IP)
```go
KeyGenerator: func(r *http.Request) string {
    userID := r.Header.Get("X-User-ID")
    if userID != "" {
        return "user:" + userID
    }
    return "ip:" + r.RemoteAddr
}
```
- Melhor experiência para usuários logados
- Fallback para IP quando não autenticado

### Por Endpoint
```go
KeyGenerator: func(r *http.Request) string {
    return r.RemoteAddr + ":" + r.URL.Path
}
```
- Rate limits diferentes por endpoint
- Mais granularidade

## ⚙️ Configurações Avançadas

### Custom Error Handler
```go
OnLimitExceeded: func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusTooManyRequests)
    
    response := map[string]interface{}{
        "error": "Custom rate limit message",
        "retry_after": 60,
        "contact": "support@company.com",
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### Skip Paths
```go
SkipPaths: []string{
    "/health",
    "/metrics",
    "/favicon.ico",
}
```

### Different Limits per User Tier
```go
KeyGenerator: func(r *http.Request) string {
    userTier := r.Header.Get("X-User-Tier")
    userID := r.Header.Get("X-User-ID")
    
    switch userTier {
    case "premium":
        return "premium:" + userID
    case "basic":
        return "basic:" + userID
    default:
        return "free:" + r.RemoteAddr
    }
}
```

## 📊 Monitoramento

### Métricas Importantes

- **Total Requests** - Volume total de requests
- **Rate Limited Requests** - Requests bloqueados
- **Rate Limit Ratio** - Percentual de requests bloqueados
- **Top Rate Limited IPs** - IPs mais bloqueados

### Logs de Rate Limiting
```go
OnLimitExceeded: func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Rate limit exceeded for %s on %s", 
        r.RemoteAddr, r.URL.Path)
    // Standard error response...
}
```

### Alertas Recomendados

- Rate limit ratio > 10%
- Spike súbito em requests bloqueados
- IPs específicos com muitos blocks

## 🛡️ Casos de Uso

### API Pública
```go
// Rate limiting agressivo para usuários não autenticados
publicConfig := ratelimit.Config{
    Limit:     60,  // 1 req/sec
    Window:    time.Minute,
    Algorithm: ratelimit.SlidingWindow,
}
```

### API de Login
```go
// Rate limiting rigoroso para tentativas de login
loginConfig := ratelimit.Config{
    Limit:     5,   // 5 tentativas por 15 min
    Window:    15 * time.Minute,
    Algorithm: ratelimit.FixedWindow,
}
```

### API Internal
```go
// Rate limiting relaxado para serviços internos
internalConfig := ratelimit.Config{
    Limit:     1000, // 1000 req/min
    Window:    time.Minute,
    Algorithm: ratelimit.TokenBucket,
}
```

## 🔧 Performance

### Otimizações

- Use cache em memória para contadores
- Implemente cleanup de chaves antigas
- Configure TTL apropriado
- Considere sharding para alta escala

### Métricas de Performance
```go
// Latência adicionada pelo middleware (deve ser < 1ms)
// Uso de memória por chave ativa
// Hit rate do cache de rate limiting
```

## 🚨 Troubleshooting

### Rate limit muito agressivo
- Usuários legítimos sendo bloqueados
- **Solução**: Aumentar limite ou janela de tempo

### Rate limit muito permissivo
- Ataques passando pelo filtro
- **Solução**: Diminuir limite ou usar algoritmo mais rigoroso

### Falsos positivos
- Usuários atrás de NAT sendo penalizados
- **Solução**: Usar autenticação para chaves únicas

### Performance degradada
- Middleware adicionando latência significativa
- **Solução**: Otimizar storage de contadores

## 📚 Referências

- [Rate Limiting Strategies](https://cloud.google.com/blog/products/api-management/rate-limiting-strategies-techniques)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [Sliding Window Rate Limiting](https://medium.com/@saisandeepmopuri/system-design-rate-limiter-and-data-modelling-9304b0d18250)
