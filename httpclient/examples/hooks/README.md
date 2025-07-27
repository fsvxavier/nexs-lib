# Hooks Example

Este exemplo demonstra como usar hooks no HTTP Client para interceptar eventos do ciclo de vida das requisições.

## Funcionalidades Demonstradas

### 1. **MetricsHook**
- Coleta métricas de performance
- Rastreia tempo de resposta
- Conta requisições lentas (>2s)
- Calcula taxa de sucesso
- Gera relatórios de métricas

### 2. **SecurityHook**
- Adiciona headers de segurança
- Valida headers de segurança na resposta
- Implementa políticas de segurança
- Detecta headers ausentes

### 3. **AuditHook**
- Registra todas as requisições para compliance
- Log estruturado para auditoria
- Rastreia método, URL e status
- Registra erros para investigação

## Como Executar

```bash
cd httpclient/examples/hooks
go run main.go
```

## Saída Esperada

```
🪝 Hooks Example
================

📋 Hooks added:
  1. Metrics Hook (tracks performance)
  2. Security Hook (adds security headers)
  3. Audit Hook (logs for compliance)

1️⃣ Making GET request...
📊 [METRICS] Starting request req-1640995200123: GET https://httpbin.org/get
🔒 [SECURITY] Added security headers to request
[AUDIT] AUDIT_REQUEST: method=GET url=https://httpbin.org/get user_agent=SecureHTTPClient/1.0
⚠️  [SECURITY] Missing security headers: [X-Frame-Options X-Content-Type-Options Strict-Transport-Security]
[AUDIT] AUDIT_RESPONSE: method=GET url=https://httpbin.org/get status=200
📊 [METRICS] Completed request req-1640995200123: 200 (took 234ms)
   ✅ GET Status: 200

2️⃣ Making POST request...
📊 [METRICS] Starting request req-1640995200456: POST https://httpbin.org/post
🔒 [SECURITY] Added security headers to request
[AUDIT] AUDIT_REQUEST: method=POST url=https://httpbin.org/post user_agent=SecureHTTPClient/1.0
⚠️  [SECURITY] Missing security headers: [X-Frame-Options X-Content-Type-Options Strict-Transport-Security]
[AUDIT] AUDIT_RESPONSE: method=POST url=https://httpbin.org/post status=200
📊 [METRICS] Completed request req-1640995200456: 200 (took 345ms)
   ✅ POST Status: 200

3️⃣ Making slow request (3 second delay)...
📊 [METRICS] Starting request req-1640995200789: GET https://httpbin.org/delay/3
🔒 [SECURITY] Added security headers to request
[AUDIT] AUDIT_REQUEST: method=GET url=https://httpbin.org/delay/3 user_agent=SecureHTTPClient/1.0
⚠️  [SECURITY] Missing security headers: [X-Frame-Options X-Content-Type-Options Strict-Transport-Security]
[AUDIT] AUDIT_RESPONSE: method=GET url=https://httpbin.org/delay/3 status=200
📊 [METRICS] Completed request req-1640995200789: 200 (took 3.234s)
   ✅ Slow request Status: 200

4️⃣ Making request that will fail...
📊 [METRICS] Starting request req-1640995204012: GET https://httpbin.org/status/404
🔒 [SECURITY] Added security headers to request
[AUDIT] AUDIT_REQUEST: method=GET url=https://httpbin.org/status/404 user_agent=SecureHTTPClient/1.0
⚠️  [SECURITY] Missing security headers: [X-Frame-Options X-Content-Type-Options Strict-Transport-Security]
[AUDIT] AUDIT_RESPONSE: method=GET url=https://httpbin.org/status/404 status=404
📊 [METRICS] Completed request req-1640995204012: 404 (took 156ms)
   ⚠️ Failed request Status: 404

📈 Request Metrics Summary:
  Total Requests: 7
  Errors: 0
  Success Rate: 100.0%
  Average Response Time: 1.156s
  Slow Requests (>2s): 1

6️⃣ Removing security hook and making another request...
📊 [METRICS] Starting request req-1640995206123: GET https://httpbin.org/get?no-security-hook=true
[AUDIT] AUDIT_REQUEST: method=GET url=https://httpbin.org/get?no-security-hook=true user_agent=
[AUDIT] AUDIT_RESPONSE: method=GET url=https://httpbin.org/get?no-security-hook=true status=200
📊 [METRICS] Completed request req-1640995206123: 200 (took 123ms)
   ✅ Status without security hook: 200

🎉 Hooks example completed!

💡 Key Features Demonstrated:
  • Request lifecycle hooks (before/after/error)
  • Performance metrics collection
  • Security header injection and validation
  • Audit logging for compliance
  • Hook removal and dynamic behavior
  • Concurrent request tracking
```

## Interface de Hook

```go
type Hook interface {
    BeforeRequest(ctx context.Context, req *Request) error
    AfterResponse(ctx context.Context, req *Request, resp *Response) error
    OnError(ctx context.Context, req *Request, err error) error
}
```

## Ciclo de Vida dos Hooks

1. **BeforeRequest**: Executado antes da requisição
   - Modificar headers
   - Adicionar metadados
   - Validar requisição
   - Iniciar timers

2. **AfterResponse**: Executado após resposta bem-sucedida
   - Coletar métricas
   - Validar resposta
   - Log de auditoria
   - Finalizar timers

3. **OnError**: Executado quando há erro
   - Log de erros
   - Métricas de falha
   - Cleanup de recursos
   - Notificações

## Casos de Uso

- **Metrics & Monitoring**: Coleta de métricas de performance
- **Security**: Validação e injeção de headers de segurança
- **Audit & Compliance**: Log de auditoria para regulamentações
- **Debugging**: Rastreamento detalhado de requisições
- **Alerting**: Notificações baseadas em eventos
- **Caching**: Invalidação de cache baseada em respostas
- **Circuit Breaker**: Implementação de circuit breaker patterns
