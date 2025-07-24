# Middleware Examples

Este diretório contém exemplos demonstrando o uso dos middlewares do sistema httpserver da nexs-lib.

## 📁 Estrutura dos Exemplos

```
examples/middleware/
├── README.md             # Este arquivo
├── advanced/             # Exemplo avançado
│   ├── main.go
│   └── README.md
├── simple/               # Exemplo básico
│   ├── main.go
│   └── README.md
├── health/               # Health checks
│   └── main.go
├── ratelimit/            # Rate limiting
│   └── main.go
├── cors/                 # CORS support
│   └── main.go
├── compression/          # Response compression
│   └── main.go
└── complete/             # Todos os middlewares juntos
    └── main.go
```

## 🚀 Executando os Exemplos

### Exemplo Básico (Simple)
```bash
cd simple/
go run main.go
```

Demonstra:
- ✅ CORS básico
- ✅ Logging de requests
- ✅ Rate limiting simples

### Health Checks
```bash
cd health/
go run main.go
```

Demonstra:
- 🏥 Liveness, readiness e startup probes
- 🔍 Verificações personalizadas
- 📊 Endpoints de health check

**Endpoints:**
- `GET /health` - Status geral
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /health/startup` - Startup probe

### Rate Limiting
```bash
cd ratelimit/
go run main.go
```

Demonstra:
- 🚦 Token bucket algorithm
- 🪟 Sliding window algorithm
- 🔑 Chaves personalizadas (IP, usuário)
- ⚠️ Respostas customizadas para rate limit

**Testes:**
```bash
# Teste básico
curl http://localhost:8080/api/test

# Teste rate limiting (execute várias vezes rapidamente)
for i in {1..15}; do curl http://localhost:8080/api/test; done

# Teste com usuário específico
curl -H 'X-User-ID: user123' http://localhost:8080/strict
```

### CORS Support
```bash
cd cors/
go run main.go
```

Demonstra:
- 🌐 Configuração de origens permitidas
- 🔧 Headers e métodos customizados
- 🍪 Suporte a credenciais
- ✈️ Requests preflight

**Testes:**
```bash
# Teste com origem permitida
curl -H 'Origin: http://localhost:3000' http://localhost:8080/api/test

# Teste com origem não permitida
curl -H 'Origin: https://malicious.com' http://localhost:8080/api/test

# Teste preflight request
curl -X OPTIONS \
     -H 'Origin: http://localhost:3000' \
     -H 'Access-Control-Request-Method: POST' \
     -H 'Access-Control-Request-Headers: Content-Type' \
     http://localhost:8080/api/test
```

### Compression
```bash
cd compression/
go run main.go
```

Demonstra:
- 📦 Compressão gzip e deflate
- 📏 Tamanho mínimo para compressão
- 📄 Tipos MIME suportados
- 🚫 Caminhos excluídos

**Testes:**
```bash
# Teste com compressão
curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large

# Teste sem compressão
curl http://localhost:8080/api/large

# Verificar headers de resposta
curl -I -H 'Accept-Encoding: gzip' http://localhost:8080/api/large
```

### Exemplo Completo
```bash
cd complete/
go run main.go
```

Demonstra **TODOS** os middlewares funcionando juntos:
- 🌐 CORS
- 📝 Logging
- ⏱️ Timeout
- 🚦 Rate limiting
- 🛡️ Bulkhead pattern
- 🔄 Retry policies
- 📦 Compression
- 🏥 Health checks

**Endpoints Especiais:**
- `/api/heavy` - Operação pesada (bulkhead separado)
- `/api/users` - Serviço de usuários (bulkhead separado)
- `/api/flaky` - Endpoint instável (demonstra retry)
- `/api/large` - Resposta grande (demonstra compressão)

## 📊 Monitoramento

### Logs de Request
Todos os exemplos incluem logging detalhado que mostra:
```
📝 [15:04:05] GET /api/test -> 200 (45ms)
🔄 Retrying /api/flaky (attempt 2) after 200ms
```

### Health Check Response
```json
{
  "status": "healthy",
  "checks": {
    "ping": {
      "status": "healthy",
      "message": "Ping successful"
    },
    "database": {
      "status": "healthy",
      "message": "Database connection active",
      "metadata": {
        "connections": 5,
        "latency_ms": 50
      }
    }
  }
}
```

### Rate Limit Headers
```
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 15
X-RateLimit-Reset: 1642694400
```

## 🔧 Configuração

### Middleware Chain Order
A ordem dos middlewares é importante:

1. **CORS** - Primeiro, para lidar com preflight requests
2. **Logging** - Para capturar todas as requests
3. **Timeout** - Proteção contra requests lentos
4. **Rate Limiting** - Controle de tráfego
5. **Bulkhead** - Isolamento de recursos
6. **Retry** - Políticas de retry
7. **Compression** - Último, para comprimir a resposta final

### Configurações Customizáveis

Cada middleware suporta configuração extensa:

```go
// Rate Limiting
rateLimitConfig := ratelimit.Config{
    Enabled:   true,
    SkipPaths: []string{"/health"},
    Limit:     100,
    Window:    time.Minute,
    Algorithm: ratelimit.TokenBucket,
    KeyGenerator: func(r *http.Request) string {
        return r.Header.Get("X-User-ID")
    },
}

// CORS
corsConfig := cors.Config{
    Enabled:          true,
    AllowedOrigins:   []string{"https://myapp.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}

// Compression
compressionConfig := compression.Config{
    Enabled: true,
    Level:   6,
    MinSize: 1024,
    Types:   []string{"application/json", "text/html"},
}
```

## 🧪 Testando

### Teste de Carga
```bash
# Teste rate limiting
for i in {1..50}; do
    curl -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -o /dev/null -s http://localhost:8080/api/test
done
```

### Teste de Compressão
```bash
# Compare tamanhos de resposta
echo "Sem compressão:"
curl -s http://localhost:8080/api/large | wc -c

echo "Com compressão:"
curl -s -H 'Accept-Encoding: gzip' http://localhost:8080/api/large | wc -c
```

### Teste de Retry
```bash
# Execute várias vezes para ver retry em ação
curl http://localhost:8080/api/flaky
curl http://localhost:8080/api/flaky
curl http://localhost:8080/api/flaky
```

## 📚 Documentação Adicional

- [Middleware Chain](../../middleware/README.md) - Documentação do sistema de chain
- [Health Checks](../../middleware/health/README.md) - Guia de health checks
- [Rate Limiting](../../middleware/ratelimit/README.md) - Algoritmos de rate limiting
- [CORS](../../middleware/cors/README.md) - Configuração de CORS
- [Compression](../../middleware/compression/README.md) - Compressão de respostas

## 🤝 Contribuindo

Para adicionar novos exemplos:

1. Crie um novo diretório com nome descritivo
2. Adicione `main.go` com exemplo funcional
3. Inclua comentários explicativos
4. Atualize este README
5. Teste a compilação com `go build`

## 🐛 Troubleshooting

### Problemas Comuns

**"Rate limit exceeded" imediatamente:**
- Verifique se outro processo está consumindo o rate limit
- Ajuste o limite ou janela de tempo

**CORS não funciona:**
- Verifique se a origem está na lista `AllowedOrigins`
- Para desenvolvimento, use `["*"]`

**Compressão não ativa:**
- Verifique se o cliente envia `Accept-Encoding`
- Confirme se o Content-Type está na lista `Types`
- Verifique se a resposta atende o `MinSize`

**Health checks falham:**
- Verifique conectividade de rede para checks externos
- Confirme se os serviços dependentes estão rodando
