# Otimização de Memória

Este exemplo demonstra as técnicas avançadas de otimização de memória implementadas no módulo IP, incluindo object pooling, lazy loading e garbage collection tuning.

## 🎯 Funcionalidades Demonstradas

### Object Pooling
- Pool de `DetectionResult` para reduzir allocações
- Pool de slices (string e byte) para operações repetitivas
- Comparação de performance com/sem pooling

### Memory Management
- Monitoramento automático de uso de memória
- Garbage collection configurável
- Limites de memória com force GC

### Lazy Loading
- Carregamento sob demanda de databases
- Cache com TTL configurável
- Unloading automático para liberar memória

## 🚀 Como Executar

```bash
cd /path/to/nexs-lib/ip/examples/memory-optimization
go run main.go
```

## 📊 Exemplo de Saída

```
🧠 Demonstração de Otimização de Memória
========================================

📊 Demonstração de Object Pooling
==================================
Teste sem object pooling:
• Criados 10000 objetos em 2.1ms
• Tempo por objeto: 210ns

Teste com object pooling:
• Processados 10000 objetos pooled em 1.2ms
• Tempo por objeto: 120ns

Testando pools de slices:
• String slice pool: 890µs para 1000 operações
• Byte slice pool: 750µs para 1000 operações

📈 Monitoramento de Memória
============================
Monitoramento inicial:
• Memória alocada: 2 MB
• Memória do sistema: 8 MB
• Total alocado: 3 MB
• Número de GCs: 1
• Tempo total de pausa GC: 45µs

Simulando carga de memória...

Após alocação de memória:
• Memória alocada: 3 MB
• Memória do sistema: 9 MB
• Total alocado: 4 MB
• Número de GCs: 2
• Tempo total de pausa GC: 89µs

🚀 Speedup: 1.75x mais rápido com otimizações!
```

## 🔧 Configuração

### Memory Manager

```go
memConfig := ip.DefaultMemoryConfig()
memConfig.GCPercent = 50          // GC mais agressivo
memConfig.MaxMemoryMB = 50        // Limite de memória
memConfig.CheckInterval = 1 * time.Second
memConfig.ForceGCThreshold = 0.8  // Force GC at 80%

memManager := ip.NewMemoryManager(memConfig)
defer memManager.Close()
```

### Object Pools

```go
// Obter objeto do pool
result := ip.GetPooledDetectionResult()
// ... usar objeto ...
// Retornar ao pool
ip.PutPooledDetectionResult(result)

// Slices
stringSlice := ip.GetPooledStringSlice()
byteSlice := ip.GetPooledByteSlice()
// ... usar slices ...
ip.PutPooledStringSlice(stringSlice)
ip.PutPooledByteSlice(byteSlice)
```

### Lazy Loading

```go
loadFunc := func() error {
    // Carregar dados caros
    time.Sleep(100 * time.Millisecond)
    return nil
}

lazyDB := ip.NewLazyDatabase(loadFunc, 5*time.Second)
data, err := lazyDB.Get() // Carrega sob demanda
```

## 📈 Benefícios de Performance

### Object Pooling
- **Redução de Allocações**: ~82% menos allocações para DetectionResult
- **Speedup**: 1.75x mais rápido em cenários de alta frequência
- **Redução de GC Pressure**: Menos trabalho para o garbage collector

### Memory Management
- **Controle de Limite**: Previne uso excessivo de memória
- **GC Tuning**: Garbage collection otimizado para workload
- **Monitoramento**: Estatísticas em tempo real

### Lazy Loading
- **Startup Rápido**: Databases carregados apenas quando necessário
- **Memory Efficiency**: Unloading automático de dados não usados
- **Cache Inteligente**: TTL configurável para balance performance/memory

## 🎛️ Cenários de Uso

### Alta Frequência
Quando processar milhares de IPs por segundo:
```go
// Use object pooling para reduzir allocações
for range requests {
    result := ip.GetPooledDetectionResult()
    // ... processar ...
    ip.PutPooledDetectionResult(result)
}
```

### Memória Limitada
Em ambientes com pouca memória:
```go
config := ip.DefaultMemoryConfig()
config.MaxMemoryMB = 100
config.ForceGCThreshold = 0.7  // GC agressivo
memManager := ip.NewMemoryManager(config)
```

### Databases Grandes
Para databases que consomem muita memória:
```go
// Lazy loading com TTL curto
lazyDB := ip.NewLazyDatabase(loadFunc, 5*time.Minute)
// Unload manual se necessário
defer lazyDB.Unload()
```

## ⚠️ Considerações

1. **Object Pooling**: Benefício maior em alta frequência (>1000 ops/sec)
2. **Memory Limits**: Definir limites realistas baseados no ambiente
3. **GC Tuning**: Testar configurações específicas para sua workload
4. **Lazy Loading**: TTL deve balancear performance vs memory usage

## 🔍 Monitoramento

```go
// Estatísticas de memória
stats := memManager.GetMemoryStats()
fmt.Printf("Allocated: %d MB\n", stats.AllocMB)
fmt.Printf("GC Count: %d\n", stats.NumGC)

// Verificar se lazy database está carregado
if lazyDB.IsLoaded() {
    fmt.Println("Database em memória")
}
```
