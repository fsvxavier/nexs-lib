# Performance Examples

Este exemplo demonstra análises detalhadas de performance do Domain Errors v2, incluindo benchmarks de criação de erros, uso de memória, concorrência, serialização, chains de erro e técnicas de otimização.

## 🎯 Funcionalidades Demonstradas

### 1. Error Creation Benchmarks
- Performance de criação de erros simples vs complexos
- Análise de tempo médio por operação
- Classificação de performance (Excellent/Good/Acceptable/Slow)
- Operações por segundo (ops/sec)

### 2. Memory Usage Analysis
- Padrões de alocação de memória
- Tracking de allocations/deallocations
- Bytes por operação
- Análise de garbage collection

### 3. Concurrency Benchmarks
- Thread-safety testing
- Performance com múltiplas goroutines
- Scaling under load
- Race condition detection

### 4. Serialization Performance
- JSON serialization benchmarks
- String conversion performance
- Size vs speed trade-offs
- Different complexity levels

### 5. Error Chain Performance
- Chain creation benchmarks
- Chain traversal performance
- Memory impact of chaining
- Scalability analysis

### 6. Builder Pattern Optimization
- Fluent vs step-by-step builders
- Builder reuse patterns
- Memory allocation patterns
- Performance classification

### 7. Factory Pattern Benchmarks
- Factory access patterns
- Caching strategies
- Direct vs indirect creation
- Singleton performance

### 8. Optimization Techniques
- Object pooling simulation
- String interning benefits
- Lazy initialization patterns
- Memory pre-allocation

## 🏗️ Arquitetura de Benchmarks

### Performance Thresholds
```go
// Error Creation Performance Classifications
Excellent: ≤100ns per operation
Good:      ≤1μs per operation  
Acceptable: ≤10μs per operation
Slow:      >10μs per operation

// Memory Efficiency Classifications
Excellent: ≤500 bytes per operation
Good:      ≤1KB per operation
Acceptable: ≤5KB per operation
High Usage: >5KB per operation

// Concurrency Performance Classifications
Excellent: ≥100K ops/sec
Good:      ≥50K ops/sec
Acceptable: ≥10K ops/sec
Slow:      <10K ops/sec
```

### Benchmark Framework
```go
type BenchmarkSuite struct {
    iterations   int
    warmupRuns   int
    scenarios    []BenchmarkScenario
    results      []BenchmarkResult
}

// Automated performance classification
func classifyPerformance(duration time.Duration, threshold time.Duration) PerformanceClass
```

## 📊 Resultados Esperados

### Error Creation Performance
- **Simple Error**: ~50-100ns (⚡ EXCELLENT)
- **Error with Type**: ~100-200ns (⚡ EXCELLENT)
- **Error with Severity**: ~150-300ns (✅ GOOD)
- **Error with Details**: ~500ns-1μs (✅ GOOD)
- **Complex Error**: ~1-3μs (⚠️ ACCEPTABLE)

### Memory Allocation
- **Simple Error**: ~200-400 bytes (⚡ EXCELLENT)
- **Medium Complex**: ~400-800 bytes (✅ GOOD)
- **High Complex**: ~1-2KB (⚠️ ACCEPTABLE)
- **Error Chains**: ~Linear growth per link

### Concurrency Scaling
- **10 goroutines**: >200K ops/sec (⚡ EXCELLENT)
- **100 goroutines**: >150K ops/sec (⚡ EXCELLENT)
- **1000 goroutines**: >100K ops/sec (⚡ EXCELLENT)

### Serialization Performance
- **JSON Simple**: ~1-2μs (⚡ EXCELLENT)
- **JSON Medium**: ~3-5μs (✅ GOOD)
- **JSON Complex**: ~10-20μs (⚠️ ACCEPTABLE)
- **String Conversion**: ~100-500ns (⚡ EXCELLENT)

## 🎮 Como Executar

### Executar Benchmarks Interativos
```bash
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/v2/domainerrors/examples/performance
go run main.go
```

### Executar Benchmarks Go
```bash
# Benchmarks básicos
go test -bench=.

# Benchmarks com profiling de memória
go test -bench=. -benchmem

# Benchmarks específicos
go test -bench=BenchmarkSimpleErrorCreation
go test -bench=BenchmarkComplexErrorCreation
go test -bench=BenchmarkConcurrentErrorCreation

# Benchmark de longa duração para estabilidade
go test -bench=. -benchtime=10s

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof
```

### Análise de Profiles
```bash
# Analisar CPU profile
go tool pprof cpu.prof

# Analisar memory profile
go tool pprof mem.prof

# Web interface para profiles
go tool pprof -http=:8080 cpu.prof
```

## 📈 Técnicas de Otimização

### 1. Object Pooling
```go
var errorDetailsPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]interface{})
    },
}

// Reutilizar maps para details
details := errorDetailsPool.Get().(map[string]interface{})
defer func() {
    for k := range details {
        delete(details, k)
    }
    errorDetailsPool.Put(details)
}()
```

### 2. String Interning
```go
var commonErrorCodes = map[string]string{
    "VALIDATION_ERROR": "VALIDATION_ERROR",
    "NOT_FOUND":       "NOT_FOUND",
    "INTERNAL_ERROR":  "INTERNAL_ERROR",
}

// Usar strings internalizadas para códigos comuns
if internedCode, exists := commonErrorCodes[code]; exists {
    code = internedCode
}
```

### 3. Lazy Initialization
```go
type OptimizedError struct {
    code        string
    message     string
    details     map[string]interface{}
    detailsOnce sync.Once
}

func (e *OptimizedError) Details() map[string]interface{} {
    e.detailsOnce.Do(func() {
        e.details = make(map[string]interface{})
        // Populate details only when needed
    })
    return e.details
}
```

### 4. Pre-allocation
```go
// Pre-alocar capacidade para maps conhecidos
details := make(map[string]interface{}, expectedSize)
tags := make([]string, 0, expectedTagCount)
```

### 5. Builder Optimization
```go
// Reutilização de builders (com cuidado)
type BuilderPool struct {
    pool sync.Pool
}

func (bp *BuilderPool) Get() *ErrorBuilder {
    if builder := bp.pool.Get(); builder != nil {
        return builder.(*ErrorBuilder).Reset()
    }
    return NewErrorBuilder()
}
```

## 🔧 Configuração de Performance

### Production Settings
```go
// Configurações otimizadas para produção
const (
    DefaultDetailsCapacity = 8    // Capacidade inicial para details
    DefaultTagsCapacity    = 4    // Capacidade inicial para tags
    MaxChainDepth         = 20   // Máxima profundidade de chains
    PoolSize              = 100  // Tamanho dos object pools
)
```

### Development Settings
```go
// Configurações para desenvolvimento (mais verbose)
const (
    EnableDetailedProfiling = true
    EnableMemoryTracking   = true
    EnableConcurrencyCheck = true
    BenchmarkIterations    = 1000
)
```

## 📊 Monitoring de Performance

### Métricas Key Performance Indicators (KPIs)
- **Error Creation Rate**: errors/second
- **Memory Allocation Rate**: bytes/second
- **GC Pressure**: GC cycles/minute
- **Concurrency Efficiency**: scaling factor
- **Serialization Overhead**: serialization time/error size

### Performance Alerts
```go
// Thresholds para alertas
const (
    ErrorCreationThreshold = 5 * time.Microsecond  // > 5μs
    MemoryUsageThreshold   = 10 * 1024             // > 10KB per error
    ConcurrencyThreshold   = 50000                 // < 50K ops/sec
    SerializationThreshold = 50 * time.Microsecond // > 50μs
)
```

### Continuous Performance Testing
```yaml
# CI/CD pipeline performance tests
performance_tests:
  - name: "Error Creation Benchmark"
    command: "go test -bench=BenchmarkSimpleErrorCreation -benchtime=5s"
    threshold: "1000 ns/op"
    
  - name: "Memory Usage Test"
    command: "go test -bench=. -benchmem"
    threshold: "1000 B/op"
    
  - name: "Concurrency Test"
    command: "go test -bench=BenchmarkConcurrentErrorCreation -benchtime=10s"
    threshold: "100000 ops/sec"
```

## 🎯 Performance Goals

### Target Performance Metrics
- **99th Percentile Error Creation**: <1μs
- **Memory Efficiency**: <1KB per error
- **Concurrency Scaling**: Linear até 1000 goroutines
- **JSON Serialization**: <10μs para erros complexos
- **Chain Traversal**: O(n) com n = profundidade da chain

### Regression Detection
```go
// Automated performance regression detection
type PerformanceBaseline struct {
    ErrorCreation   time.Duration
    MemoryUsage     int64
    Serialization   time.Duration
    Concurrency     float64
}

func detectRegression(current, baseline PerformanceBaseline) bool {
    threshold := 0.1 // 10% degradation threshold
    return current.ErrorCreation > baseline.ErrorCreation*(1+threshold)
}
```

## 🔍 Profiling e Debugging

### CPU Profiling
```bash
# Profile específico para error creation
go test -bench=BenchmarkSimpleErrorCreation -cpuprofile=error_creation_cpu.prof

# Analisar hotspots
go tool pprof -top error_creation_cpu.prof
go tool pprof -web error_creation_cpu.prof
```

### Memory Profiling
```bash
# Profile de alocações
go test -bench=. -memprofile=mem.prof -memprofilerate=1

# Analisar vazamentos de memória
go tool pprof -alloc_space mem.prof
go tool pprof -inuse_space mem.prof
```

### Trace Analysis
```bash
# Trace de execução
go test -bench=BenchmarkConcurrentErrorCreation -trace=trace.out

# Analisar trace
go tool trace trace.out
```

Este exemplo fornece uma base abrangente para análise e otimização de performance do Domain Errors v2, permitindo identificar gargalos e implementar melhorias baseadas em dados reais de performance.
