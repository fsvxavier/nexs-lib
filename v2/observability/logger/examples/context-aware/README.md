# Exemplo Context-Aware Logging - Logger v2

Este exemplo demonstra como extrair e propagar automaticamente informações de contexto através de uma aplicação.

## Funcionalidades Demonstradas

- ✅ Extração automática de trace_id, span_id, user_id
- ✅ Propagação de contexto em sub-operações
- ✅ Middleware pattern para enriquecimento de contexto
- ✅ Logging concorrente com contexto preservado
- ✅ Simulação de distributed tracing
- ✅ Context propagation através de goroutines

## Como Executar

```bash
cd context-aware
go run main.go
```

## Output Esperado

Logs JSON estruturados com contexto rico incluindo:
- IDs de trace distribuído
- Correlação entre operações
- Contexto preservado em operações concorrentes
- Hierarquia de spans e operações

## Conceitos de Contexto

### Tipos de Context Keys
```go
type contextKey string

const (
    TraceIDKey   contextKey = "trace_id"
    SpanIDKey    contextKey = "span_id" 
    UserIDKey    contextKey = "user_id"
    RequestIDKey contextKey = "request_id"
)
```

### Extração de Contexto
```go
func getTraceID(ctx context.Context) string {
    if val := ctx.Value(TraceIDKey); val != nil {
        return val.(string)
    }
    return ""
}
```

### Logger Context-Aware
```go
contextLogger := baseLogger.
    WithTraceID(getTraceID(ctx)).
    WithSpanID(getSpanID(ctx)).
    WithFields(
        interfaces.String("user_id", getUserID(ctx)),
        interfaces.String("request_id", getRequestID(ctx)),
    )
```

## Padrões Implementados

### 1. **Sub-operations com Context**
Cada operação cria seu próprio span mantendo o trace ID:
```go
authCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
authLogger := logger.WithSpanID(getSpanID(authCtx))
```

### 2. **Middleware Pattern**
Enriquecimento automático de contexto:
```go
// Middleware 1: Request ID
ctx = context.WithValue(ctx, RequestIDKey, generateRequestID())

// Middleware 2: Authentication  
ctx = context.WithValue(ctx, UserIDKey, extractUserID(req))
```

### 3. **Concurrent Operations**
Contexto preservado em goroutines:
```go
go func() {
    workerCtx := context.WithValue(ctx, SpanIDKey, generateSpanID())
    workerLogger := logger.WithSpanID(getSpanID(workerCtx))
    workerLogger.Info(workerCtx, "Worker iniciado")
}()
```

## Benefícios

1. **Observabilidade**: Correlação automática entre logs
2. **Debugging**: Rastreamento de operações distribuídas  
3. **Performance**: Identificação de gargalos por trace
4. **Compliance**: Auditoria completa de operações
5. **Troubleshooting**: Context rico para resolução de problemas

## Próximos Passos

- [Async](../async/) - Para aplicações de alta performance
- [Microservices](../microservices/) - Para sistemas distribuídos
- [Web-App](../web-app/) - Para aplicações HTTP
