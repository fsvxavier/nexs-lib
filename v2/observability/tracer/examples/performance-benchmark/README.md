# Performance & Benchmark Example

Este exemplo demonstra monitoramento de performance e benchmarks com o tracer.

## Recursos Demonstrados

- **Performance Monitoring**: Latência, throughput, percentis
- **Benchmarking**: Testes de performance comparativos
- **Memory Profiling**: Detecção de vazamentos de memória
- **Stress Testing**: Teste sob alta carga
- **Resource Monitoring**: CPU, memória, goroutines
- **SLA Tracking**: Monitoramento de SLOs/SLAs

## Como Executar

```bash
# Executar benchmarks
go test -bench=. -benchmem -timeout 60s

# Executar testes de performance
go run main.go

# Executar com profiling
go run main.go -cpuprofile cpu.prof -memprofile mem.prof
```

## Métricas Monitoradas

### Latência
- P50, P95, P99 latencies
- Média móvel de latência
- Detecção de outliers

### Throughput
- Requests per second
- Messages per second
- Transactions per minute

### Recursos
- Memory allocation rate
- GC frequency
- Goroutine count
- CPU utilization

### Business Metrics
- Error rate percentage
- Success rate
- Conversion rates
- Revenue impact
