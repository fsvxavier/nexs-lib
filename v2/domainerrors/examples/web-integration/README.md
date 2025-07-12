# Web Integration Examples

Este exemplo demonstra a integração completa do Domain Errors v2 com aplicações web e HTTP, incluindo tratamento de erros em APIs REST, GraphQL, middlewares, clientes HTTP, webhooks, rate limiting e CORS.

## 🎯 Funcionalidades Demonstradas

### 1. HTTP Error Handling
- Mapeamento automático de tipos de erro para códigos HTTP
- Respostas de erro padronizadas com metadata completa
- Suporte a diferentes severidades e tipos de erro

### 2. REST API Error Standards
- Endpoints com tratamento de erro padronizado
- Validação de entrada com múltiplos erros
- Respostas estruturadas seguindo boas práticas REST

### 3. Middleware Error Handling
- Pipeline de middlewares com tratamento de erro robusto
- Autenticação, rate limiting, validação e logging
- Identificação do middleware que falhou

### 4. HTTP Client Error Handling
- Timeouts, conexões falhadas e erros de servidor
- Classificação de erros como "retryable"
- Metadata detalhada para debugging

### 5. GraphQL Error Handling
- Erros de sintaxe, validação e execução
- Paths e extensions conforme especificação GraphQL
- Tratamento de erros de autorização em campos

### 6. Webhook Error Handling
- Delivery com retry automático
- Timeouts e falhas de autenticação
- Tracking de tentativas e próximo retry

### 7. Rate Limiting
- Múltiplos endpoints com limites diferentes
- Estatísticas detalhadas de uso
- Headers de retry informativos

### 8. CORS Handling
- Validação completa de política CORS
- Geração automática de headers apropriados
- Suporte a preflight requests

## 🏗️ Arquitetura

### HTTP Error Handler
```go
type HTTPErrorHandler struct {
    factory interfaces.ErrorFactory
}

// Mapeia tipos de erro para códigos HTTP
func (h *HTTPErrorHandler) GetHTTPStatusCode(err interfaces.DomainErrorInterface) int
func (h *HTTPErrorHandler) CreateErrorResponse(err interfaces.DomainErrorInterface) map[string]interface{}
```

### REST API
```go
type RESTAPI struct {
    factory interfaces.ErrorFactory
}

// Endpoints com tratamento padronizado
func (api *RESTAPI) GetUser(userID string) *APIResponse
func (api *RESTAPI) CreateUser(data map[string]interface{}) *APIResponse
```

### Middleware Chain
```go
type ErrorMiddleware struct {
    factory interfaces.ErrorFactory
}

// Pipeline de middlewares
func (m *ErrorMiddleware) ProcessRequest(req *HTTPRequest, middlewares []Middleware) *MiddlewareResponse
```

### Rate Limiter
```go
type RateLimiter struct {
    limits     map[string]RateLimit
    counters   map[string]map[string]*RateCounter
    statistics map[string]*RateLimitStats
}
```

## 🔄 Padrões de Erro Implementados

### 1. Error Type Mapping
- `ErrorTypeValidation` → HTTP 400
- `ErrorTypeAuthentication` → HTTP 401
- `ErrorTypeAuthorization` → HTTP 403
- `ErrorTypeNotFound` → HTTP 404
- `ErrorTypeConflict` → HTTP 409
- `ErrorTypeRateLimit` → HTTP 429
- `ErrorTypeExternalService` → HTTP 502
- `ErrorTypeTimeout` → HTTP 504
- `ErrorTypeDatabase` → HTTP 500

### 2. Structured Error Responses
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID '123' not found",
    "type": "not_found",
    "severity": "low",
    "details": {
      "user_id": "123",
      "available_actions": ["create_user", "list_users"]
    },
    "timestamp": "2025-07-12T10:30:00Z",
    "trace_id": "trace_1720781234567890"
  }
}
```

### 3. GraphQL Error Format
```json
{
  "errors": [
    {
      "message": "Access denied: insufficient permissions",
      "path": ["user", "sensitiveData"],
      "extensions": {
        "code": "AUTHORIZATION_ERROR",
        "required_role": "admin",
        "current_role": "user"
      }
    }
  ]
}
```

## 📊 Métricas e Observabilidade

### Rate Limiter Statistics
- Total de requests por endpoint
- Taxa de sucesso e bloqueios
- Requests permitidos vs bloqueados
- Análise de performance por cliente

### CORS Policy Monitoring
- Origins bloqueadas vs permitidas
- Métodos e headers rejeitados
- Estatísticas de preflight requests

### Webhook Delivery Tracking
- Tentativas de delivery por webhook
- Timeouts e falhas por endpoint
- Schedule de retry automático

## 🎮 Como Executar

```bash
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/v2/domainerrors/examples/web-integration
go run main.go
```

## 📈 Performance

### Benchmarks Esperados
- HTTP Error Mapping: ~100ns por operação
- Middleware Chain: ~1μs para 4 middlewares
- Rate Limiter Check: ~500ns por verificação
- CORS Validation: ~300ns por request
- Error Response Creation: ~2μs incluindo JSON marshal

### Memory Efficiency
- Zero allocations para error mappings
- Object pooling para responses frequentes
- Garbage collection otimizada para alta throughput

## 🔧 Configuração Avançada

### Custom Error Mappings
```go
handler := NewHTTPErrorHandler()
// Adicionar mapeamentos customizados
handler.AddMapping("BUSINESS_RULE_ERROR", 422)
```

### Rate Limiter Policies
```go
limiter := NewRateLimiter()
limiter.SetLimit("api", 100, time.Hour)      // 100/hour
limiter.SetLimit("upload", 5, time.Minute)   // 5/minute
```

### CORS Configuration
```go
cors := NewCORSHandler()
cors.SetPolicy(CORSPolicy{
    AllowedOrigins: []string{"*.example.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowCredentials: true,
    MaxAge: 3600,
})
```

## 🎯 Casos de Uso Empresariais

### 1. API Gateway Integration
- Error standardization across microservices
- Centralized rate limiting and CORS
- Distributed tracing with correlation IDs

### 2. E-commerce Platform
- Product catalog with proper error handling
- Payment processing with retry logic
- Order management with webhook notifications

### 3. SaaS Application
- Multi-tenant rate limiting
- Feature access control with authorization errors
- Webhook delivery for integrations

### 4. Mobile Backend
- Optimized error responses for mobile clients
- Network resilience with retryable errors
- Push notification webhooks

## 🔍 Debugging e Troubleshooting

### Error Tracing
Cada erro inclui:
- Trace ID para correlação
- Timestamp preciso
- Stack trace quando relevante
- Context metadata completo

### Log Integration
```go
// Middleware de logging automático
func (m *ErrorMiddleware) LoggingMiddleware(req *HTTPRequest) interfaces.DomainErrorInterface {
    // Log structured com error metadata
    return nil
}
```

### Health Checks
```go
// Endpoint de health check com status detalhado
func (api *RESTAPI) HealthCheck() *APIResponse {
    // Verifica dependências e retorna status
}
```

Este exemplo serve como referência para implementação enterprise de tratamento de erros em aplicações web modernas, seguindo as melhores práticas da indústria.
