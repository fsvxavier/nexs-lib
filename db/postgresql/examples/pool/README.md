# Pool PostgreSQL Provider Example

Este exemplo demonstra o gerenciamento avançado de pools de conexão PostgreSQL com monitoramento, operações concorrentes e lifecycle management.

## 📋 Funcionalidades Demonstradas

- ✅ **Gerenciamento de Pool**: Criação e configuração de pools de conexão
- ✅ **Estatísticas em Tempo Real**: Monitoramento contínuo do pool
- ✅ **Operações Concorrentes**: Múltiplas operações simultâneas
- ✅ **Lifecycle Management**: Criação, aquecimento e encerramento graceful
- ✅ **Auto Resource Management**: Uso do AcquireFunc para liberação automática
- ✅ **Health Checks**: Verificação de saúde do pool
- ✅ **Performance Monitoring**: Métricas detalhadas de performance

## 🚀 Pré-requisitos

1. **PostgreSQL Database**:
   ```bash
   # Usando Docker
   docker run --name postgres-pool \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=testdb \
     -p 5432:5432 -d postgres:15
   ```

2. **Dependências Go**:
   ```bash
   go mod tidy
   ```

## ⚙️ Configuração do Pool

O exemplo demonstra configurações avançadas de pool:

```go
cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

err := cfg.ApplyOptions(
    postgresql.WithMaxConns(20),          // Máximo de 20 conexões
    postgresql.WithMinConns(5),           // Mínimo de 5 conexões
    postgresql.WithMaxConnLifetime(30*time.Minute),  // Vida útil das conexões
    postgresql.WithMaxConnIdleTime(5*time.Minute),   // Timeout de idle
)
```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/pool

# Executar o exemplo
go run main.go
```

## 📊 Saída Esperada

```
=== Basic Pool Example ===
📊 Initial Pool Stats:
  - Total Connections: 5
  - Acquired Connections: 0
  - Idle Connections: 5
  - Max Connections: 20
🔌 Connection acquired from pool
✅ Query executed successfully, result: 1
✅ Pool operations completed successfully
📊 Updated Pool Stats:
  - Total Connections: 5
  - Acquired Connections: 0
  - Idle Connections: 5

=== Pool Statistics Example ===
📈 Monitoring pool statistics for 5 seconds...
⏰ [14:30:01] Total: 5, Acquired: 0, Idle: 5, Constructing: 0
⏰ [14:30:02] Total: 5, Acquired: 0, Idle: 5, Constructing: 0
...
✅ Pool monitoring completed

=== Concurrent Operations Example ===
🚀 Starting 10 workers with 5 operations each
✅ Worker 0 operation 0 completed
✅ Worker 1 operation 0 completed
...
📊 Concurrent operations summary:
  - Successful operations: 50
  - Failed operations: 0
📊 Final Pool Stats:
  - Total Connections: 10
  - Acquired Connections: 0
  - Idle Connections: 10

=== Pool Lifecycle Example ===
🔄 Creating pool...
📊 Pool created - Initial connections: 5
🔥 Warming up pool with operations...
✅ Warmup operation 0 completed
✅ Warmup operation 1 completed
✅ Warmup operation 2 completed
📊 Pool warmed up - Total connections: 5, Idle: 5
🏥 Performing pool health check...
✅ Pool health check passed
🔄 Closing pool gracefully...
✅ Pool closed successfully
Pool examples completed!
```

## 📝 Conceitos Demonstrados

### 1. Configuração Avançada de Pool
```go
postgresql.WithMaxConns(20),
postgresql.WithMinConns(5),
postgresql.WithMaxConnLifetime(30*time.Minute),
postgresql.WithMaxConnIdleTime(5*time.Minute),
```

### 2. Uso do AcquireFunc (Recomendado)
```go
err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
    // Operações com a conexão
    // Liberação automática quando sai da função
    return nil
})
```

### 3. Monitoramento de Estatísticas
```go
stats := pool.Stats()
fmt.Printf("Total: %d, Acquired: %d, Idle: %d", 
    stats.TotalConns, stats.AcquiredConns, stats.IdleConns)
```

### 4. Operações Concorrentes
```go
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func(workerID int) {
        defer wg.Done()
        // Operações paralelas usando o pool
    }(i)
}
wg.Wait()
```

## 🔧 Parâmetros de Tuning

### Pool Size
- **MinConns**: Conexões mínimas sempre ativas
- **MaxConns**: Limite máximo de conexões
- **Regra**: MinConns = CPU cores, MaxConns = 2-4x CPU cores

### Timeouts
- **MaxConnLifetime**: Tempo máximo de vida de uma conexão
- **MaxConnIdleTime**: Tempo antes de fechar conexão idle
- **AcquireTimeout**: Timeout para adquirir conexão

### Para Alta Concorrência
```go
postgresql.WithMaxConns(100),
postgresql.WithMinConns(10),
postgresql.WithMaxConnLifetime(1*time.Hour),
postgresql.WithMaxConnIdleTime(10*time.Minute),
postgresql.WithAcquireTimeout(30*time.Second),
```

## 📈 Métricas de Performance

O exemplo monitora:
- **Total Connections**: Conexões totais no pool
- **Acquired Connections**: Conexões em uso
- **Idle Connections**: Conexões disponíveis
- **Constructing Connections**: Conexões sendo criadas
- **Successful Operations**: Operações bem-sucedidas
- **Failed Operations**: Operações que falharam

## 🐛 Troubleshooting

### Pool Esgotado
```
failed to acquire connection: pool exhausted
```
**Solução**: Aumente MaxConns ou implemente timeout/retry.

### Conexões Idle Excessivas
```
too many idle connections
```
**Solução**: Diminua MinConns ou MaxConnIdleTime.

### Vazamentos de Conexão
```
connection leak detected
```
**Solução**: Sempre use AcquireFunc ou release manual das conexões.

## 📚 Próximos Passos

Após dominar pools de conexão, explore:

1. **[Transaction Example](../transaction/)**: Transações avançadas
2. **[Advanced Example](../advanced/)**: Hooks e middleware
3. **[Multi-tenant Example](../multitenant/)**: Suporte multi-tenant

## 🔍 Debugging

Para debug detalhado do pool:
```bash
export LOG_LEVEL=debug
export POOL_DEBUG=true
```

Isso mostrará:
- Criação/destruição de conexões
- Estatísticas detalhadas do pool
- Timeouts e erros de aquisição
- Padrões de uso concorrente
