# Exemplo Structured Logging - Logger v2

Este exemplo demonstra o uso de logging estruturado com campos tipados.

## Funcionalidades Demonstradas

- ✅ Campos estruturados tipados (String, Int, Float64, Bool, Time, Duration)
- ✅ Logging de operações comerciais
- ✅ Métricas de performance estruturadas  
- ✅ Logging de erros com contexto rico
- ✅ Arrays e objetos complexos
- ✅ Logger hierárquico com campos comuns
- ✅ Campos pré-definidos com WithFields

## Como Executar

```bash
cd structured
go run main.go
```

## Output Esperado

O exemplo produzirá logs em formato JSON estruturado mostrando:
1. Campos tipados corretamente formatados
2. Logs de operações comerciais
3. Métricas de performance
4. Tratamento de erros com contexto
5. Estruturas complexas (arrays, objetos)

## Código Principal

```go
// Campos básicos tipados
logger.Info(ctx, "Usuário autenticado",
    interfaces.String("user_id", "user123"),
    interfaces.Int("age", 30),
    interfaces.Bool("is_admin", false),
)

// Operações comerciais
logger.Info(ctx, "Pedido criado",
    interfaces.String("order_id", "ord_789"),
    interfaces.Float64("total_amount", 299.99),
    interfaces.Time("created_at", time.Now()),
)

// Erros com contexto
logger.Error(ctx, "Falha no processamento",
    interfaces.ErrorNamed("error", businessErr),
    interfaces.String("error_code", "INSUFFICIENT_FUNDS"),
)

// Logger com campos comuns
requestLogger := logger.WithFields(
    interfaces.String("request_id", "req_abc123"),
    interfaces.String("method", "POST"),
)
```

## Tipos de Campo Disponíveis

| Função | Tipo Go | Descrição |
|--------|---------|-----------|
| `String(key, value)` | string | Texto |
| `Int(key, value)` | int | Número inteiro |
| `Int64(key, value)` | int64 | Número inteiro 64-bit |
| `Float64(key, value)` | float64 | Número decimal |
| `Bool(key, value)` | bool | Booleano |
| `Time(key, value)` | time.Time | Timestamp |
| `Duration(key, value)` | time.Duration | Duração |
| `ErrorNamed(key, err)` | error | Erro nomeado |
| `Array(key, value)` | []interface{} | Array/slice |
| `Object(key, value)` | map[string]interface{} | Objeto/map |

## Benefícios do Structured Logging

1. **Consultabilidade**: Logs facilmente pesquisáveis
2. **Análise**: Agregação e métricas automáticas
3. **Debugging**: Contexto rico para troubleshooting
4. **Monitoramento**: Alertas baseados em campos específicos
5. **Compliance**: Auditoria e rastreabilidade

## Padrões Recomendados

### IDs e Identificadores
```go
interfaces.String("user_id", userID)
interfaces.String("request_id", requestID)
interfaces.String("trace_id", traceID)
```

### Métricas e Performance
```go
interfaces.Duration("response_time", duration)
interfaces.Int("status_code", statusCode)
interfaces.Float64("cpu_usage", cpuPercent)
```

### Erros e Falhas
```go
interfaces.ErrorNamed("error", err)
interfaces.String("error_code", "BUSINESS_ERROR")
interfaces.String("error_type", "validation")
```

### Operações Comerciais
```go
interfaces.String("operation", "payment_processing")
interfaces.Float64("amount", 299.99)
interfaces.String("currency", "BRL")
```

## Próximos Passos

Após dominar este exemplo:
- [Context-Aware](../context-aware/) - Para propagação de contexto
- [Middleware](../middleware/) - Para transformação automática
- [Microservices](../microservices/) - Para sistemas distribuídos
