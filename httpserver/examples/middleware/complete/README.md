# Complete Middleware Example

Este exemplo demonstra **TODOS** os middlewares da nexs-lib funcionando juntos em uma aplicação real.

## 🚀 Sobre Este Exemplo

Este é o exemplo mais completo que mostra:
- ✅ **8 Middlewares** trabalhando em conjunto
- ✅ **Configurações realistas** para produção
- ✅ **Monitoramento** com logs e métricas
- ✅ **Endpoints especializados** para diferentes cenários
- ✅ **Interface rica** com feedback visual

## 🎯 Middlewares Implementados

| Ordem | Middleware | Função | Configuração |
|-------|------------|--------|--------------|
| 1 | **CORS** | Cross-origin requests | Todas as origens |
| 2 | **Logging** | Request/response logging | Console colorido |
| 3 | **Timeout** | Request timeout | 30 segundos |
| 4 | **Rate Limiting** | Traffic control | 20 req/min |
| 5 | **Bulkhead** | Resource isolation | Por endpoint |
| 6 | **Retry** | Failure recovery | 2 tentativas |
| 7 | **Compression** | Response compression | Gzip/deflate |
| 8 | **Health Checks** | Service monitoring | Multi-probe |

## 🚀 Executando o Exemplo

```bash
go run main.go
```

Você verá uma saída rica com emojis e informações detalhadas:

```
🚀 Starting Complete Middleware Demo Server
==================================================

📡 Server starting on :8080

🌐 Available Endpoints:
──────────────────────

🏥 Health Checks:
  GET /health         - Overall health status
  GET /health/live    - Liveness probe
  GET /health/ready   - Readiness probe

🔧 API Endpoints (Full Middleware Stack):
  GET /api/test       - Simple test endpoint
  GET /api/heavy      - Heavy operation (2s delay)
  GET /api/users      - User list (bulkhead: user-service)
  GET /api/flaky      - Flaky endpoint (demos retry)
  GET /api/large      - Large response (demos compression)
```

## 📍 Endpoints Especializados

### 🏥 Health Checks
```bash
# Status geral
curl http://localhost:8080/health

# Liveness probe
curl http://localhost:8080/health/live

# Readiness probe  
curl http://localhost:8080/health/ready
```

### 🔧 API Test
```bash
# Endpoint básico com todos os middlewares
curl http://localhost:8080/api/test
```

**Resposta:**
```json
{
  "message": "All middleware working!",
  "timestamp": "2025-07-20T15:04:05Z",
  "method": "GET",
  "path": "/api/test",
  "user_agent": "curl/7.68.0",
  "middleware_features": [
    "CORS enabled",
    "Request logged", 
    "Timeout protected",
    "Rate limited",
    "Bulkhead isolated",
    "Retry enabled",
    "Response compressed"
  ]
}
```

### ⚡ Heavy Operations (Bulkhead Demo)
```bash
# Operação pesada (bulkhead separado)
curl http://localhost:8080/api/heavy
```

**Características:**
- 🕐 Demora 2 segundos intencionalmente
- 🛡️ Isolado em bulkhead `heavy-operations`
- 📊 Não afeta outros endpoints

### 👥 Users Service (Bulkhead Demo)
```bash
# Serviço de usuários (bulkhead separado)
curl http://localhost:8080/api/users
```

**Características:**
- 📋 Retorna lista de 10 usuários
- 🛡️ Isolado em bulkhead `user-service`
- 💾 Dados gerados dinamicamente

### 🔄 Retry Demo
```bash
# Endpoint que falha 2 de 3 vezes
curl http://localhost:8080/api/flaky
curl http://localhost:8080/api/flaky
curl http://localhost:8080/api/flaky
```

**Comportamento:**
- ❌ Primeiras 2 tentativas: HTTP 500
- ✅ 3ª tentativa: HTTP 200
- 🔄 Demonstra retry automático

### 📦 Compression Demo
```bash
# Resposta grande para demonstrar compressão
curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large
```

**Características:**
- 📊 1000 items de dados
- 📦 ~50KB sem compressão
- 🗜️ ~8KB com gzip (84% redução)

## 🔍 Observando os Middlewares

### 📝 Logs em Tempo Real

Ao fazer requests, você verá logs coloridos:

```
📝 [15:04:05] GET /api/test -> 200 (45ms)
📝 [15:04:06] GET /api/heavy -> 200 (2.1s)
🔄 Retrying /api/flaky (attempt 2) after 200ms
📝 [15:04:07] GET /api/flaky -> 200 (350ms)
```

### 🚦 Rate Limiting

Execute multiple requests rapidamente:

```bash
# Teste rate limiting (20 req/min)
for i in {1..25}; do
    echo "Request $i:"
    curl -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -s http://localhost:8080/api/test
done
```

**Resultado esperado:**
- ✅ Primeiros 20: HTTP 200
- ❌ Próximos 5: HTTP 429 (Too Many Requests)

### 🛡️ Bulkhead Isolation

Teste concorrência:

```bash
# Terminal 1: Heavy operations
for i in {1..5}; do
    curl http://localhost:8080/api/heavy &
done

# Terminal 2: User service (não afetado)
curl http://localhost:8080/api/users
```

**Comportamento:**
- Heavy operations ficam na fila (max 10 concurrent)
- User service responde normalmente
- Isolamento perfeito entre recursos

### ⏱️ Timeout Protection

Teste com delay forçado:

```bash
# Operação que demora muito (simular)
curl http://localhost:8080/api/heavy
```

**Proteção:**
- ⏰ Timeout automático em 30s
- 🛡️ Previne requests infinitos
- 📊 Logs de timeout quando aplicável

## 🔧 Configurações Detalhadas

### CORS Configuration
```go
corsConfig := cors.DefaultConfig()
corsConfig.AllowedOrigins = []string{"*"}
corsConfig.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
corsConfig.AllowedHeaders = []string{"Content-Type", "Authorization", "X-User-ID"}
```

### Rate Limiting Configuration
```go
rateLimitConfig := ratelimit.DefaultConfig()
rateLimitConfig.Limit = 20           // 20 requests
rateLimitConfig.Window = time.Minute // per minute
rateLimitConfig.SkipPaths = []string{"/health"}
```

### Bulkhead Configuration
```go
bulkheadConfig := bulkhead.DefaultConfig()
bulkheadConfig.MaxConcurrent = 10
bulkheadConfig.QueueSize = 20
bulkheadConfig.ResourceKey = func(r *http.Request) string {
    switch {
    case r.URL.Path == "/api/heavy":
        return "heavy-operations"
    case r.URL.Path == "/api/users":
        return "user-service"
    default:
        return "default"
    }
}
```

### Retry Configuration
```go
retryConfig := retry.DefaultConfig()
retryConfig.MaxRetries = 2
retryConfig.InitialDelay = 100 * time.Millisecond
retryConfig.OnRetry = func(r *http.Request, attempt int, delay time.Duration) {
    fmt.Printf("🔄 Retrying %s (attempt %d) after %v\n", r.URL.Path, attempt, delay)
}
```

## 🧪 Cenários de Teste

### 1. Teste de Carga Completo
```bash
#!/bin/bash
echo "🧪 Teste de Carga Completo"
echo "========================="

# Teste rate limiting
echo "1. Testando Rate Limiting..."
for i in {1..25}; do
    STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/api/test)
    echo "Request $i: $STATUS"
done

# Teste bulkhead
echo -e "\n2. Testando Bulkhead..."
for i in {1..3}; do
    curl http://localhost:8080/api/heavy &
done
curl http://localhost:8080/api/users
wait

# Teste retry
echo -e "\n3. Testando Retry..."
for i in {1..3}; do
    curl http://localhost:8080/api/flaky
done

# Teste compressão
echo -e "\n4. Testando Compressão..."
echo "Sem compressão: $(curl -s http://localhost:8080/api/large | wc -c) bytes"
echo "Com gzip: $(curl -s -H 'Accept-Encoding: gzip' http://localhost:8080/api/large | wc -c) bytes"
```

### 2. Teste de CORS
```bash
# Teste origins diferentes
curl -H 'Origin: http://localhost:3000' http://localhost:8080/api/test
curl -H 'Origin: https://example.com' http://localhost:8080/api/test

# Teste preflight
curl -X OPTIONS \
     -H 'Origin: http://localhost:3000' \
     -H 'Access-Control-Request-Method: POST' \
     http://localhost:8080/api/test
```

### 3. Teste de Health Checks
```bash
# Verificar todos os health checks
echo "Health Status:"
curl -s http://localhost:8080/health | jq .

echo -e "\nLiveness:"
curl -s http://localhost:8080/health/live | jq .

echo -e "\nReadiness:"
curl -s http://localhost:8080/health/ready | jq .
```

## 📊 Monitoramento

### Métricas Observadas

O exemplo mostra em tempo real:
- **Request Rate** - Através dos logs
- **Response Times** - Tempo de cada request
- **Error Rate** - Status codes diferentes de 2xx
- **Retry Attempts** - Tentativas de retry
- **Rate Limit Hits** - Requests bloqueados
- **Bulkhead Usage** - Concorrência por recurso

### Dashboard Simulado

```
📊 Middleware Metrics (Last 1 min):
─────────────────────────────────────
🔵 Total Requests: 45
🟢 Successful (2xx): 38 (84%)
🟡 Rate Limited (429): 5 (11%)
🔴 Server Error (5xx): 2 (5%)

⏱️  Average Response Time: 156ms
🔄 Total Retries: 3
🛡️  Bulkhead Queued: 2
📦 Compression Ratio: 78%
```

## 🎯 Cenários de Produção

### Load Balancer Health Checks
```bash
# Kubernetes liveness probe
curl -f http://localhost:8080/health/live

# Kubernetes readiness probe
curl -f http://localhost:8080/health/ready
```

### API Gateway Integration
```bash
# Com rate limiting per user
curl -H 'X-User-ID: user123' http://localhost:8080/api/test

# Com API key
curl -H 'Authorization: Bearer token123' http://localhost:8080/api/test
```

### CDN/Proxy Compatibility
```bash
# Teste com proxy headers
curl -H 'X-Forwarded-For: 192.168.1.100' \
     -H 'X-Real-IP: 192.168.1.100' \
     http://localhost:8080/api/test
```

## 🔧 Customização

### Adicionando Novo Middleware
```go
// Adicionar à chain
chain.Add(yourCustomMiddleware.NewMiddleware(config))
```

### Modificando Ordem
```go
// A ordem importa! 
// 1. CORS (primeiro - preflight)
// 2. Logging (capturar tudo)
// 3. Timeout (proteção)
// 4. Rate Limiting (controle)
// 5. Bulkhead (isolamento)
// 6. Retry (recuperação)
// 7. Compression (último - resposta final)
```

### Configuração por Ambiente
```go
func getConfigForEnv(env string) Config {
    switch env {
    case "development":
        return developmentConfig()
    case "staging":
        return stagingConfig()
    case "production":
        return productionConfig()
    default:
        return defaultConfig()
    }
}
```

## 🚨 Troubleshooting

### Performance Issues
- **Sintoma:** Requests muito lentos
- **Verificar:** Logs de timeout, retry attempts
- **Solução:** Ajustar timeouts, otimizar endpoints

### Rate Limiting Excessivo
- **Sintoma:** Muitos 429 errors
- **Verificar:** Configuração de limite e janela
- **Solução:** Aumentar limites ou usar chaves mais específicas

### Bulkhead Saturation
- **Sintoma:** Requests em fila
- **Verificar:** Logs de concorrência
- **Solução:** Aumentar MaxConcurrent ou otimizar operações

### Memory Usage
- **Sintoma:** Alto uso de memória
- **Verificar:** Rate limiting storage, bulkhead queues
- **Solução:** Implementar cleanup, ajustar configurações

## 📚 Próximos Passos

1. **Métricas:** Integrar com Prometheus/Grafana
2. **Tracing:** Adicionar OpenTelemetry
3. **Alertas:** Configurar alertas baseados em métricas
4. **Dashboard:** Criar dashboard de monitoramento
5. **Tests:** Implementar testes de carga automatizados

## 🎓 Aprendizado

Este exemplo ensina:
- **Arquitetura** de middlewares
- **Observabilidade** em aplicações
- **Resilience patterns** (bulkhead, retry, timeout)
- **Performance optimization** (compression, rate limiting)
- **Security best practices** (CORS, input validation)

## 📖 Documentação Relacionada

- [Middleware Chain Documentation](../../middleware/README.md)
- [Health Checks Guide](../health/README.md)
- [Rate Limiting Guide](../ratelimit/README.md)
- [CORS Configuration](../cors/README.md)
- [Compression Optimization](../compression/README.md)
