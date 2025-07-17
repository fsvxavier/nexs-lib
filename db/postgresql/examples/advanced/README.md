# Advanced PostgreSQL Provider Example

Este exemplo demonstra funcionalidades avançadas do provider PostgreSQL, incluindo hooks customizados, middleware, monitoramento, análise de performance e tratamento avançado de erros.

## 📋 Funcionalidades Demonstradas

- ✅ **Custom Hooks**: Hooks customizados para logging, timing, auditoria e detecção de queries lentas
- ✅ **Metrics Collection**: Coleta automática de métricas de performance
- ✅ **Performance Analysis**: Análise detalhada de performance de queries
- ✅ **Advanced Error Handling**: Categorização e análise avançada de erros
- ✅ **Connection Lifecycle**: Hooks de ciclo de vida de conexões
- ✅ **Security Auditing**: Auditoria de operações sensíveis
- ✅ **Real-time Monitoring**: Monitoramento em tempo real de operações

## 🚀 Pré-requisitos

1. **PostgreSQL Database**:
   ```bash
   # Usando Docker
   docker run --name postgres-advanced \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=testdb \
     -p 5432:5432 -d postgres:15
   ```

2. **Dependências Go**:
   ```bash
   go mod tidy
   ```

## ⚙️ Configuração

1. **Atualize a string de conexão** no arquivo `main.go`:
   ```go
   cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
   ```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/advanced

# Executar o exemplo
go run main.go
```

## 📊 Saída Esperada

```
=== Custom Hooks Demonstration ===
✅ Registered Query Logger hook
✅ Registered Timer hook
✅ Registered Error Logger hook
✅ Registered Slow Query Detector hook
✅ Registered Security Auditor hook

📊 Testing hooks with sample operations...

🔍 Executing test query 1...
🪝 [QUERY] query: SELECT 1 as test
🪝 [ARGS] []
⏱️  [TIMING] query took 2.5ms

🔍 Executing test query 2...
🪝 [QUERY] query: SELECT COUNT(*) FROM information_schema.tables
⏱️  [TIMING] query took 15.2ms

🔍 Executing test query 3...
🪝 [QUERY] query: SELECT pg_sleep(0.1), 'slow query' as result
🐌 [SLOW QUERY] query took 102.3ms (threshold: 50ms)
🐌 [SLOW QUERY SQL] SELECT pg_sleep(0.1), 'slow query' as result
⏱️  [TIMING] query took 102.3ms

⏭️  Skipping error query for demo purposes

📈 Collected Metrics:
  - Total Queries: 3
  - Total Duration: 120ms
  - Error Count: 0
  - Slow Queries: 1
  - Average Duration: 40ms

=== Metrics and Monitoring Demonstration ===
🔄 Simulating various database operations...
  Executing metadata_query...
  Executing count_query...
  Executing info_query...

📊 Operation Metrics:
  query: 3 operations, avg duration: 8.5ms

📈 Pool Statistics:
  - Total Connections: 2
  - Acquired Connections: 0
  - Idle Connections: 2
  - Max Connections: 10

=== Performance Analysis Demonstration ===
🔍 Executing queries with different performance characteristics...
  📝 Executing simple query...
    ⏱️  Completed in 1.2ms
  📝 Executing metadata query...
    ⏱️  Completed in 5.8ms
  📝 Executing calculation query...
    ⏱️  Completed in 12.4ms

📊 Performance Analysis Results:
  - SELECT 1: 1.2ms
  - SELECT schemaname, tablename FROM pg_tables...: 5.8ms
  - SELECT generate_series(1,100) as num: 12.4ms

📈 Summary:
  - Total Queries: 3
  - Average Duration: 6.5ms
  - Slow Queries (>10ms): 1
  - Performance Score: 66.7%

=== Advanced Error Handling Demonstration ===
🔍 Testing various error scenarios...
  🧪 Testing syntax_error...
    ❌ Expected error occurred: syntax error at or near "SELEC"
🔍 [ERROR ANALYSIS] Category: Syntax Error, Error: syntax error at or near "SELEC"
  🧪 Testing missing_table...
    ❌ Expected error occurred: relation "non_existent_table" does not exist
🔍 [ERROR ANALYSIS] Category: Unknown Error, Error: relation "non_existent_table" does not exist
  🧪 Testing invalid_function...
    ❌ Expected error occurred: function unknown_function() does not exist
🔍 [ERROR ANALYSIS] Category: Unknown Error, Error: function unknown_function() does not exist

📊 Error Analysis Summary:
  - Syntax Error: 1 occurrences
  - Unknown Error: 2 occurrences

=== Connection Lifecycle Hooks Demonstration ===
🔄 Testing connection lifecycle...

  🔄 Connection cycle 1:
🔗 [LIFECYCLE] Before connection: Operation: query
✅ [LIFECYCLE] After connection: Operation: query
🔓 [LIFECYCLE] Before release: Operation: query
🆓 [LIFECYCLE] After release: Operation: query
    ✅ Connection cycle 1 completed

  🔄 Connection cycle 2:
🔗 [LIFECYCLE] Before connection: Operation: query
✅ [LIFECYCLE] After connection: Operation: query
🔓 [LIFECYCLE] Before release: Operation: query
🆓 [LIFECYCLE] After release: Operation: query
    ✅ Connection cycle 2 completed

  🔄 Connection cycle 3:
🔗 [LIFECYCLE] Before connection: Operation: query
✅ [LIFECYCLE] After connection: Operation: query
🔓 [LIFECYCLE] Before release: Operation: query
🆓 [LIFECYCLE] After release: Operation: query
    ✅ Connection cycle 3 completed

📋 Connection Lifecycle Events:
  1. 🔗 [BEFORE CONNECTION] Operation: query
  2. ✅ [AFTER CONNECTION] Operation: query
  3. 🔓 [BEFORE RELEASE] Operation: query
  4. 🆓 [AFTER RELEASE] Operation: query
  5. 🔗 [BEFORE CONNECTION] Operation: query
  6. ✅ [AFTER CONNECTION] Operation: query
  7. 🔓 [BEFORE RELEASE] Operation: query
  8. 🆓 [AFTER RELEASE] Operation: query
  9. 🔗 [BEFORE CONNECTION] Operation: query
  10. ✅ [AFTER CONNECTION] Operation: query
  11. 🔓 [BEFORE RELEASE] Operation: query
  12. 🆓 [AFTER RELEASE] Operation: query

Advanced example completed successfully!
```

## 📝 Conceitos Demonstrados

### 1. Custom Hooks System
```go
// Hook para logging de queries
queryLogHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
    if ctx.Operation == "query" || ctx.Operation == "exec" {
        fmt.Printf("🪝 [QUERY] %s: %s\n", ctx.Operation, ctx.Query)
    }
    return &interfaces.HookResult{Continue: true}
}

// Registrar hook
hookManager.RegisterHook(interfaces.BeforeQueryHook, queryLogHook)
```

### 2. Metrics Collection
```go
type MetricsCollector struct {
    queryCount      int64
    totalDuration   time.Duration
    errorCount      int64
    slowQueries     int64
}

func (m *MetricsCollector) RecordQuery(duration time.Duration, err error) {
    m.queryCount++
    m.totalDuration += duration
    if err != nil {
        m.errorCount++
    }
    if duration > m.slowQueryThreshold {
        m.slowQueries++
    }
}
```

### 3. Performance Analysis
```go
type QueryPerformance struct {
    Query     string
    Duration  time.Duration
    Timestamp time.Time
}

// Hook para coleta de performance
perfHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
    performanceLog = append(performanceLog, QueryPerformance{
        Query:     ctx.Query,
        Duration:  ctx.Duration,
        Timestamp: time.Now(),
    })
    return &interfaces.HookResult{Continue: true}
}
```

### 4. Error Categorization
```go
type ErrorCategory int

const (
    ConnectionError ErrorCategory = iota
    SyntaxError
    ConstraintError
    TimeoutError
    UnknownError
)

func categorizeError(err error) ErrorCategory {
    errStr := err.Error()
    if strings.Contains(errStr, "syntax") {
        return SyntaxError
    }
    // ... outras categorizações
    return UnknownError
}
```

### 5. Security Auditing
```go
auditHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
    if containsSensitiveOperation(ctx.Query) {
        fmt.Printf("🔒 [AUDIT] Sensitive operation: %s\n", ctx.Query)
    }
    return &interfaces.HookResult{Continue: true}
}

func containsSensitiveOperation(query string) bool {
    sensitiveKeywords := []string{"DROP", "DELETE", "UPDATE", "ALTER"}
    // ... verificação
}
```

## 🎯 Tipos de Hooks Disponíveis

### Connection Hooks
- `BeforeConnectionHook`: Antes de estabelecer conexão
- `AfterConnectionHook`: Após estabelecer conexão
- `BeforeReleaseHook`: Antes de liberar conexão
- `AfterReleaseHook`: Após liberar conexão

### Operation Hooks
- `BeforeQueryHook`: Antes de executar query
- `AfterQueryHook`: Após executar query
- `BeforeExecHook`: Antes de executar comando
- `AfterExecHook`: Após executar comando
- `BeforeTransactionHook`: Antes de iniciar transação
- `AfterTransactionHook`: Após concluir transação

### Pool Hooks
- `BeforeAcquireHook`: Antes de adquirir conexão do pool
- `AfterAcquireHook`: Após adquirir conexão do pool

### Error Hooks
- `OnErrorHook`: Quando ocorre erro

### Custom Hooks
- `CustomHookBase + N`: Hooks customizados definidos pelo usuário

## 📈 Métricas Coletadas

### Performance Metrics
- **Query Count**: Número total de queries executadas
- **Total Duration**: Tempo total gasto em queries
- **Average Duration**: Tempo médio por query
- **Slow Queries**: Queries que excedem threshold definido
- **Error Rate**: Taxa de erro das operações

### Pool Metrics
- **Total Connections**: Conexões totais no pool
- **Acquired Connections**: Conexões em uso
- **Idle Connections**: Conexões disponíveis
- **Max Connections**: Limite máximo configurado

### Error Metrics
- **Error Categories**: Categorização de tipos de erro
- **Error Frequency**: Frequência de cada tipo de erro
- **Error Patterns**: Padrões de erro detectados

## 🔧 Configuração de Monitoramento

### Configuração de Thresholds
```go
slowQueryThreshold := 100 * time.Millisecond  // Queries lentas
errorRateThreshold := 5.0                     // Taxa de erro máxima (%)
connectionPoolWarning := 0.8                  // 80% do pool usado
```

### Configuração de Alertas
```go
// Configurar alertas para métricas críticas
if slowQueryRate > 10.0 {
    // Alerta de performance
}
if errorRate > errorRateThreshold {
    // Alerta de erro
}
if poolUsage > connectionPoolWarning {
    // Alerta de pool
}
```

## 🐛 Troubleshooting

### Hooks Não Funcionam
```
No hooks were executed
```
**Solução**: Verifique se o hook manager está implementado e registrado corretamente.

### Performance Degradada
```
High query latency detected
```
**Solução**: Use hooks de performance para identificar queries lentas.

### Muitos Erros
```
High error rate detected
```
**Solução**: Use hooks de erro para categorizar e analisar padrões.

## 📚 Próximos Passos

Após dominar hooks e middleware, explore:

1. **[Multi-tenant Example](../multitenant/)**: Suporte multi-tenant
2. **[Performance Example](../performance/)**: Otimização de performance
3. **[Production Example](../production/)**: Configuração para produção

## 🔍 Debugging Avançado

Para debug detalhado:
```bash
export LOG_LEVEL=debug
export HOOK_DEBUG=true
export METRICS_DEBUG=true
export PERFORMANCE_DEBUG=true
```

Logs incluirão:
- Execução detalhada de hooks
- Métricas em tempo real
- Análise de performance query-by-query
- Rastreamento de lifecycle de conexões
