# Performance PostgreSQL Provider Example

Este exemplo demonstra técnicas avançadas de otimização de performance, benchmarking e monitoramento para operações PostgreSQL, incluindo pool optimization, query tuning e análise de memória.

## 🎯 Modo de Operação

Este exemplo foi **otimizado para funcionar tanto com quanto sem banco de dados PostgreSQL**:

- ✅ **Com Banco**: Executa benchmarks reais com métricas precisas
- ✅ **Sem Banco**: Executa simulações educativas com dicas de performance
- ✅ **Graceful Degradation**: Nunca falha ou gera panic, sempre fornece valor educativo

## 📋 Funcionalidades Demonstradas

- ✅ **Pool Performance**: Otimização de pools de conexão com diferentes configurações
- ✅ **Query Optimization**: Técnicas de otimização de queries e prepared statements
- ✅ **Batch Operations**: Performance de operações em lote vs individuais
- ✅ **Concurrent Benchmarks**: Benchmarks de concorrência multi-worker
- ✅ **Memory Optimization**: Otimização de uso de memória e análise de GC
- ✅ **Connection Lifecycle**: Otimização do ciclo de vida de conexões
- ✅ **Performance Metrics**: Coleta e análise de métricas detalhadas
- ✅ **Error Recovery**: Tratamento robusto de falhas de conectividade

## 🚀 Execução Rápida

**Sem configuração** (modo simulação):
```bash
go run main.go
```

**Com banco PostgreSQL** (benchmarks reais):
```bash
# 1. Inicie o PostgreSQL
docker run --name postgres-performance \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 -d postgres:15

# 2. Execute o exemplo
go run main.go
```

## ⚙️ Configuração

1. **Atualize a string de conexão** no arquivo `main.go`:
   ```go
   cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
   ```

2. **Configure parâmetros de performance**:
   ```go
   postgresql.WithMaxConns(50),              // Pool size otimizado
   postgresql.WithMinConns(10),              // Conexões mínimas
   postgresql.WithMaxConnLifetime(1*time.Hour),   // Vida útil das conexões
   postgresql.WithMaxConnIdleTime(10*time.Minute), // Timeout idle
   ```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/performance

# Executar o exemplo
go run main.go
```

## � Pré-requisitos (Opcional)

**PostgreSQL Database** (para benchmarks reais):
```bash
# Usando Docker com configurações otimizadas
docker run --name postgres-performance \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  --shm-size=256m \
  -d postgres:15 \
  -c shared_buffers=128MB \
  -c max_connections=200 \
  -c work_mem=4MB
```

**Configuração de Conexão** (atualize se necessário):
```go
cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
```

## �📊 Modos de Execução

### Modo Simulação (Padrão - Sem Banco)
- ✅ **Sempre funciona** - não requer configuração
- ✅ **Educativo** - fornece dicas e simulações realistas
- ✅ **Seguro** - nunca falha ou gera panic
- ✅ **Rápido** - execução instantânea

**Exemplo de saída:**
```
🚀 Starting PostgreSQL Performance Examples
💡 Note: These examples require a running PostgreSQL database
🔧 If database is not available, examples will run in simulation mode

=== Connection Pool Performance Example ===

🏊 Testing Small Pool (Max: 5, Min: 2)
  🔍 Testing pool connectivity...
  💡 Pool created but database connection failed: connection test failed with panic
  📊 Simulating pool performance metrics...
  📈 Simulated Results for Small Pool:
    - Total Operations: 100
    - Simulated Average Time: 1ms
    - Estimated Operations/sec: 100000000.00
    - Pool Configuration: Max=5, Min=2
```

### Modo Real (Com Banco PostgreSQL)
- ✅ **Benchmarks reais** - métricas precisas
- ✅ **Performance real** - testa operações no banco
- ✅ **Análise completa** - todos os aspectos de performance

**Exemplo de saída:**
```
=== Connection Pool Performance Example ===

🏊 Testing Small Pool (Max: 5, Min: 2)
  🔍 Testing pool connectivity...
  📊 Running 100 acquisition operations...
    ⏱️  Operation 0: 1.2ms
    ⏱️  Operation 20: 0.8ms
    ⏱️  Operation 40: 0.9ms
    ⏱️  Operation 60: 1.1ms
    ⏱️  Operation 80: 0.7ms
  📈 Results for Small Pool:
    - Total Operations: 100
    - Total Time: 234ms
    - Average Time: 2.1ms
    - Min Time: 0.5ms
    - Max Time: 8.3ms
    - Error Count: 0
    - Slow Operations: 5
    - Operations/sec: 427.35
    - Final Pool Size: 5
    - Idle Connections: 5

🏊 Testing Medium Pool (Max: 20, Min: 5)
  📊 Running 100 acquisition operations...
  📈 Results for Medium Pool:
    - Total Operations: 100
    - Total Time: 156ms
    - Average Time: 1.4ms
    - Min Time: 0.4ms
    - Max Time: 4.2ms
    - Error Count: 0
    - Slow Operations: 0
    - Operations/sec: 641.03
    - Final Pool Size: 8
    - Idle Connections: 8

🏊 Testing Large Pool (Max: 50, Min: 10)
  📊 Running 100 acquisition operations...
  📈 Results for Large Pool:
    - Total Operations: 100
    - Total Time: 142ms
    - Average Time: 1.2ms
    - Min Time: 0.3ms
    - Max Time: 3.1ms
    - Error Count: 0
    - Slow Operations: 0
    - Operations/sec: 704.23
    - Final Pool Size: 12
    - Idle Connections: 12

=== Query Optimization Example ===
📊 Testing different query patterns...

  🔍 Testing: Simple Select
    - Average: 0.8ms
    - Min: 0.5ms
    - Max: 2.1ms
    - Errors: 0
    - Rating: ⚡ Excellent

  🔍 Testing: Parameterized Query
    - Average: 0.9ms
    - Min: 0.6ms
    - Max: 1.8ms
    - Errors: 0
    - Rating: ⚡ Excellent

  🔍 Testing: Complex Calculation
    - Average: 15.2ms
    - Min: 12.4ms
    - Max: 22.1ms
    - Errors: 0
    - Rating: ⚠️  Acceptable

  🔍 Testing: System Query
    - Average: 3.2ms
    - Min: 2.1ms
    - Max: 5.8ms
    - Errors: 0
    - Rating: ✅ Good

📝 Testing prepared statement vs regular query...
  🔄 Running 50 regular queries...
    Regular Query Average: 1.2ms
    ℹ️  Prepared statement optimization would require driver-specific implementation
    💡 Generally 2-5x faster for repeated queries

=== Batch Operations Performance Example ===
📊 Comparing individual inserts vs batch operations...

  🔄 Testing 1000 individual inserts...
    ⏱️  Inserted 100 records...
    ⏱️  Inserted 200 records...
    ⏱️  Inserted 300 records...
    ⏱️  Inserted 400 records...
    ⏱️  Inserted 500 records...
    ⏱️  Inserted 600 records...
    ⏱️  Inserted 700 records...
    ⏱️  Inserted 800 records...
    ⏱️  Inserted 900 records...
    ✅ Individual inserts completed in: 2.543s
    📈 Rate: 393.25 inserts/sec

  📦 Testing batch insert with VALUES clause...
    ⏱️  Batch inserted 500 records...
    ✅ Batch inserts completed in: 421ms
    📈 Rate: 2375.30 inserts/sec

  📊 Performance Comparison:
    - Individual: 2.543s (393.25 ops/sec)
    - Batch: 421ms (2375.30 ops/sec)
    - Speedup: 6.04x faster
    - Rating: ⚡ Excellent optimization
    ✅ Verified 1000 records in table

=== Concurrent Operations Benchmark ===
📊 Testing concurrency levels with 50 operations per worker...

  🚀 Testing 1 concurrent workers...
    📈 Results for 1 workers:
      - Total Operations: 50
      - Total Time: 156ms
      - Throughput: 320.51 ops/sec
      - Average Latency: 3.1ms
      - Min Latency: 1.2ms
      - Max Latency: 8.4ms
      - Error Rate: 0.00%
      - Pool Connections: 2
      - Active Connections: 0
      - Rating: ⚠️  Acceptable

  🚀 Testing 5 concurrent workers...
    📈 Results for 5 workers:
      - Total Operations: 250
      - Total Time: 234ms
      - Throughput: 1068.38 ops/sec
      - Average Latency: 4.6ms
      - Min Latency: 1.1ms
      - Max Latency: 12.3ms
      - Error Rate: 0.00%
      - Pool Connections: 5
      - Active Connections: 0
      - Rating: ⚡ Excellent

  🚀 Testing 10 concurrent workers...
    📈 Results for 10 workers:
      - Total Operations: 500
      - Total Time: 312ms
      - Throughput: 1602.56 ops/sec
      - Average Latency: 6.2ms
      - Min Latency: 1.3ms
      - Max Latency: 18.7ms
      - Error Rate: 0.00%
      - Pool Connections: 10
      - Active Connections: 0
      - Rating: ⚡ Excellent

  🚀 Testing 20 concurrent workers...
    📈 Results for 20 workers:
      - Total Operations: 1000
      - Total Time: 428ms
      - Throughput: 2336.45 ops/sec
      - Average Latency: 8.5ms
      - Min Latency: 1.4ms
      - Max Latency: 25.6ms
      - Error Rate: 0.00%
      - Pool Connections: 20
      - Active Connections: 0
      - Rating: ⚡ Excellent

  🚀 Testing 50 concurrent workers...
    📈 Results for 50 workers:
      - Total Operations: 2500
      - Total Time: 1.234s
      - Throughput: 2025.93 ops/sec
      - Average Latency: 24.6ms
      - Min Latency: 2.1ms
      - Max Latency: 85.3ms
      - Error Rate: 0.00%
      - Pool Connections: 50
      - Active Connections: 0
      - Rating: ⚡ Excellent

=== Memory Optimization Example ===
📊 Memory usage analysis...
  📈 Baseline memory: 2048 KB

  🔍 Testing large result set handling...
    ✅ Processed 10000 rows in 245ms
    📈 Memory after large query: 3842 KB (delta: +1794 KB)

  🏊 Testing connection pool memory efficiency...
    ✅ Completed 20 pool operations in 68ms
    📈 Memory after pool test: 3956 KB (delta: +1908 KB)

  💡 Memory Optimization Tips:
    - Use connection pooling to reuse connections
    - Process large result sets in batches
    - Close rows and statements explicitly
    - Use prepared statements for repeated queries
    - Monitor and tune pool sizes

  📊 Current Pool Statistics:
    - Total Connections: 12
    - Idle Connections: 12
    - Acquired Connections: 0
    - Max Connections: 50

=== Connection Lifecycle Optimization Example ===
📊 Analyzing connection lifecycle performance...

  🔍 Testing: Rapid Acquire/Release
    Description: Many short-lived connections
    ✅ Test completed in: 245ms
    📈 Pool Changes:
      - Initial connections: 12
      - Final connections: 12
      - Connection delta: 0
      - Final idle: 12

  🔍 Testing: Long-held Connection
    Description: One longer-lived connection
    ✅ Test completed in: 156ms
    📈 Pool Changes:
      - Initial connections: 12
      - Final connections: 12
      - Connection delta: 0
      - Final idle: 12

  🏥 Testing connection health checking...
    ✅ Health checked 5 connections in 23ms

  💡 Connection Lifecycle Optimization Tips:
    - Use AcquireFunc for automatic connection management
    - Configure appropriate MaxConnLifetime
    - Set reasonable MaxConnIdleTime
    - Monitor connection pool metrics
    - Implement health checks for long-lived connections
    - Use prepared statements to reduce connection overhead

Performance examples completed!
```

## �️ Robustez e Tratamento de Erros

### Características de Robustez
- ✅ **Zero Panic**: Nunca falha com panic, mesmo com problemas de conectividade
- ✅ **Graceful Degradation**: Funciona em modo simulação quando banco não disponível  
- ✅ **Error Recovery**: Captura e trata todos os tipos de erro automaticamente
- ✅ **Connection Testing**: Testa conectividade antes de executar benchmarks
- ✅ **Resource Cleanup**: Limpa recursos automaticamente em caso de falha

### Implementação de Error Recovery
```go
// Todas as funções usam este padrão para evitar panic
var testErr error
func() {
    defer func() {
        if r := recover(); r != nil {
            testErr = fmt.Errorf("connection test failed with panic: %v", r)
        }
    }()
    
    testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
        return conn.HealthCheck(ctx)
    })
}()

if testErr != nil {
    // Fallback para modo simulação com dicas educativas
    fmt.Printf("💡 Would require database: %v\n", testErr)
    return nil
}
```

### Benefícios da Abordagem Robusta
- **Para Desenvolvedores**: Pode executar exemplos sem configurar banco
- **Para Produção**: Código resiliente que nunca falha inesperadamente
- **Para Aprendizado**: Sempre fornece valor educativo, mesmo sem infraestrutura

## �📝 Conceitos Demonstrados

### 1. Pool Performance Optimization
```go
// Configuração otimizada baseada no workload
postgresql.WithMaxConns(50),                    // CPU cores * 2-4
postgresql.WithMinConns(10),                    // CPU cores
postgresql.WithMaxConnLifetime(1*time.Hour),    // Evitar connection leaks
postgresql.WithMaxConnIdleTime(10*time.Minute), // Liberar recursos idle
```

### 2. Métricas de Performance
```go
type PerformanceMetrics struct {
    TotalQueries      int64
    TotalDuration     time.Duration
    MinDuration       time.Duration
    MaxDuration       time.Duration
    ErrorCount        int64
    SlowQueries       int64
}

func (pm *PerformanceMetrics) RecordQuery(duration time.Duration, err error) {
    pm.TotalQueries++
    pm.TotalDuration += duration
    if duration > pm.slowQueryThreshold {
        pm.SlowQueries++
    }
}
```

### 3. Batch Operations
```go
// Batch insert usando VALUES clause
values := ""
args := make([]interface{}, 0, batchSize*2)

for j := 0; j < batchSize; j++ {
    if j > 0 {
        values += ","
    }
    values += fmt.Sprintf("($%d, $%d)", j*2+1, j*2+2)
    args = append(args, data1, data2)
}

query := fmt.Sprintf("INSERT INTO table (col1, col2) VALUES %s", values)
_, err := conn.Exec(ctx, query, args...)
```

### 4. Concurrent Benchmarking
```go
func benchmarkConcurrency(workers int, operationsPerWorker int) {
    var wg sync.WaitGroup
    metrics := NewPerformanceMetrics()
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for j := 0; j < operationsPerWorker; j++ {
                // Measure operation
                start := time.Now()
                err := performOperation()
                duration := time.Since(start)
                metrics.RecordQuery(duration, err)
            }
        }(i)
    }
    wg.Wait()
}
```

### 5. Memory Optimization
```go
// Monitoramento de memória
var m runtime.MemStats
runtime.GC()
runtime.ReadMemStats(&m)
baselineAlloc := m.Alloc

// Após operações
runtime.GC()
runtime.ReadMemStats(&m)
currentAlloc := m.Alloc
memoryDelta := currentAlloc - baselineAlloc
```

## 🎯 Benchmarks e Métricas

### Pool Performance
| Pool Size | Operations/sec | Avg Latency | Memory Usage |
|-----------|---------------|-------------|--------------|
| Small (5) | 427 ops/sec | 2.1ms | Low |
| Medium (20) | 641 ops/sec | 1.4ms | Medium |
| Large (50) | 704 ops/sec | 1.2ms | High |

### Query Patterns
| Pattern | Avg Time | Rating | Use Case |
|---------|----------|--------|----------|
| Simple SELECT | 0.8ms | ⚡ Excellent | Lookups |
| Parameterized | 0.9ms | ⚡ Excellent | Dynamic queries |
| Complex calc | 15.2ms | ⚠️ Acceptable | Analytics |
| System query | 3.2ms | ✅ Good | Metadata |

### Batch vs Individual
| Method | Rate | Speedup | Memory |
|--------|------|---------|--------|
| Individual | 393 ops/sec | 1x | Low |
| Batch | 2375 ops/sec | 6x | Medium |

### Concurrency Scaling
| Workers | Throughput | Avg Latency | Efficiency |
|---------|------------|-------------|------------|
| 1 | 320 ops/sec | 3.1ms | 100% |
| 5 | 1068 ops/sec | 4.6ms | 213% |
| 10 | 1603 ops/sec | 6.2ms | 160% |
| 20 | 2336 ops/sec | 8.5ms | 117% |
| 50 | 2026 ops/sec | 24.6ms | 41% |

## 🔧 Tuning Guidelines

### Pool Size Tuning
```go
// Regra geral para pool size
MaxConns = CPU_CORES * 2  // Para workloads balanceados
MaxConns = CPU_CORES * 4  // Para workloads I/O intensivos
MinConns = CPU_CORES      // Conexões sempre ativas

// Para aplicações específicas
WebApp:    MaxConns = 20-50
API:       MaxConns = 50-100
Analytics: MaxConns = 10-20
Batch:     MaxConns = 5-10
```

### Query Optimization
```go
// Thresholds de performance
ExcellentQuery := 1 * time.Millisecond   // < 1ms
GoodQuery      := 10 * time.Millisecond  // < 10ms
AcceptableQuery := 100 * time.Millisecond // < 100ms
SlowQuery      := 1 * time.Second        // > 1s
```

### Memory Optimization
```go
// Guidelines de memória
LowMemoryUsage    := 10 * 1024 * 1024  // < 10MB
MediumMemoryUsage := 50 * 1024 * 1024  // < 50MB
HighMemoryUsage   := 100 * 1024 * 1024 // < 100MB
```

## 🚨 Alertas de Performance

### Configuração de Alertas
```go
// Thresholds para alertas
SlowQueryThreshold     := 100 * time.Millisecond
HighErrorRateThreshold := 5.0  // 5%
PoolExhaustionThreshold := 90.0 // 90% do pool
MemoryLeakThreshold    := 500 * 1024 * 1024 // 500MB
```

### Métricas Críticas
- **Latência média > 100ms**: Investigar queries lentas
- **Taxa de erro > 5%**: Verificar conectividade/queries
- **Pool usage > 90%**: Aumentar pool ou otimizar uso
- **Memory growth > 500MB**: Investigar vazamentos

## 📊 Profiling Tools

### Built-in Profiling
```bash
# Habilitar profiling
export PROFILE_CPU=true
export PROFILE_MEMORY=true
export PROFILE_BLOCK=true

# Executar com profiling
go run -race main.go
```

### External Tools
```bash
# pgbench para benchmarking PostgreSQL
pgbench -c 10 -j 2 -t 1000 testdb

# pg_stat_statements para análise de queries
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY total_time DESC;
```

## 🔍 Debugging Performance

Para debug detalhado:
```bash
export LOG_LEVEL=debug
export PERFORMANCE_DEBUG=true
export POOL_METRICS=true
export QUERY_TIMING=true
export MEMORY_PROFILING=true
```

Logs incluirão:
- Timing de cada operação
- Estatísticas detalhadas do pool
- Uso de memória por operação
- Identificação de gargalos
- Recomendações de otimização
