# Middleware Example

Este exemplo demonstra como usar middlewares no HTTP Client para interceptar e modificar requisições/respostas.

## Funcionalidades Demonstradas

### 1. **LoggingMiddleware**
- Registra detalhes de requisições e respostas
- Adiciona IDs únicos para rastreamento
- Mede tempo de resposta
- Registra erros com detalhes

### 2. **AuthMiddleware**
- Adiciona headers de autenticação automaticamente
- Injeta tokens Bearer
- Adiciona versão da API

### 3. **RateLimitMiddleware**
- Implementa rate limiting com token bucket
- Controla número de requisições por janela de tempo
- Bloqueia requisições quando limite é atingido

## Como Executar

```bash
cd httpclient/examples/middleware
go run main.go
```

## Saída Esperada

```
🚀 Middleware Example
=====================

📋 Middlewares added:
  1. Rate Limit Middleware (5 requests per 10s)
  2. Auth Middleware (adds Bearer token)
  3. Logging Middleware (logs requests/responses)

1️⃣ Making GET request to /get...
[HTTP] 📤 [REQUEST] GET https://httpbin.org/get
[HTTP] 📥 [RESPONSE] GET https://httpbin.org/get - 200 (took 234ms)
   ✅ Status: 200

2️⃣ Making POST request to /post...
[HTTP] 📤 [REQUEST] POST https://httpbin.org/post
[HTTP] 📥 [RESPONSE] POST https://httpbin.org/post - 200 (took 345ms)
   ✅ Status: 200

3️⃣ Testing rate limiting with rapid requests...
   Request 1: Status 200 (took 156ms)
   Request 2: Status 200 (took 234ms)
   Request 3: Status 200 (took 123ms)
   Request 4: Status 200 (took 456ms)
   Request 5: Status 200 (took 234ms)
   Request 6: Status 200 (took 2.1s)  # Rate limited
   Request 7: Status 200 (took 2.3s)  # Rate limited

4️⃣ Removing rate limit middleware...
Making request without rate limiting...
   ✅ Status: 200 (should be faster)

🎉 Middleware example completed!

💡 Key Features Demonstrated:
  • Chaining multiple middlewares
  • Request/response logging
  • Authentication header injection
  • Rate limiting with token bucket
  • Dynamic middleware removal
  • Context propagation
```

## Conceitos Importantes

### **Ordem de Execução**
Os middlewares são executados em **ordem reversa** ao serem adicionados:
1. RateLimitMiddleware (primeiro na execução)
2. AuthMiddleware
3. LoggingMiddleware (último na execução)

### **Interface de Middleware**
```go
type Middleware interface {
    Process(ctx context.Context, req *Request, next func(context.Context, *Request) (*Response, error)) (*Response, error)
}
```

### **Padrão Chain of Responsibility**
Cada middleware pode:
- Modificar a requisição antes de chamar `next()`
- Processar a resposta após chamar `next()`
- Interromper a cadeia retornando sem chamar `next()`
- Implementar funcionalidades transversais (logging, auth, cache, etc.)

## Casos de Uso

- **Logging**: Auditoria e debugging
- **Autenticação**: Injeção automática de tokens
- **Rate Limiting**: Controle de tráfego
- **Caching**: Cache de respostas
- **Retry**: Tentativas automáticas
- **Metrics**: Coleta de métricas
- **Security**: Validação de headers de segurança
