# Exemplo Pool de Conexões PostgreSQL

Este exemplo demonstra o uso e gerenciamento de pools de conexões PostgreSQL usando a biblioteca nexs-lib.

## Funcionalidades Demonstradas

### 1. Pool Básico
- Criação de pool com configuração padrão
- Aquisição e liberação de conexões
- Execução de queries básicas

### 2. Pool Configurado
- Configuração avançada de pool
- Definição de limites min/max
- Configuração de lifetimes
- Configuração de idle time

### 3. Pool com Métricas
- Monitoramento de conexões ativas
- Tracking de conexões disponíveis
- Métricas de utilização

### 4. Pool com Timeout
- Configuração de timeouts
- Tratamento de esgotamento de pool
- Recuperação após liberação

### 5. Pool com Limites
- Teste de limites máximos
- Distribuição de carga
- Execução concorrente

### 6. Pool com Lifecycle
- Ciclo de vida completo do pool
- Warmup de conexões
- Gestão de idle timeout
- Reutilização de pool

### 7. Pool com Monitoring
- Monitoramento em tempo real
- Métricas de utilização
- Tracking de fases

### 8. Pool com Load Testing
- Teste de carga com múltiplos workers
- Medição de throughput
- Análise de performance
- Estatísticas detalhadas

## Como Executar

### Usando Docker (Recomendado)
```bash
# Iniciar infraestrutura
./infraestructure/manage.sh start

# Executar exemplo de pool
./infraestructure/manage.sh example pool

# Parar infraestrutura
./infraestructure/manage.sh stop
```

### Execução Direta
```bash
# Configurar variáveis de ambiente
export NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"

# Executar
go run main.go
```

## Configurações de Pool

### Configuração Básica
```go
pool, err := postgres.ConnectPool(ctx, dsn)
```

### Configuração Avançada
```go
cfg := postgres.NewConfigWithOptions(
    dsn,
    postgres.WithMaxConns(20),           // Máximo de conexões
    postgres.WithMinConns(5),            // Mínimo de conexões
    postgres.WithMaxConnLifetime(30*time.Minute), // Tempo de vida máximo
    postgres.WithMaxConnIdleTime(10*time.Minute), // Tempo idle máximo
)

pool, err := postgres.ConnectPoolWithConfig(ctx, cfg)
```

## Parâmetros de Configuração

| Parâmetro | Descrição | Valor Padrão | Recomendação |
|-----------|-----------|--------------|--------------|
| MaxConns | Número máximo de conexões | 50 | Baseado na capacidade do servidor |
| MinConns | Número mínimo de conexões | 2 | 10-20% do MaxConns |
| MaxConnLifetime | Tempo de vida máximo da conexão | 1 hora | 30 minutos - 2 horas |
| MaxConnIdleTime | Tempo máximo idle | 30 minutos | 10-30 minutos |

## Saída Esperada

```
=== Exemplo Pool de Conexões PostgreSQL ===

Pool Básico
===========
  Criando pool básico...
  Adquirindo conexão...
  Resultado: Pool básico funcionando!
✓ Pool Básico concluído com sucesso

Pool Configurado
================
  Criando pool configurado...
  Configurações do pool:
    Max Conexões: 20
    Min Conexões: 5
    Max Lifetime: 30m0s
    Max Idle Time: 10m0s
  Testando múltiplas aquisições...
  Adquiridas 10 conexões com sucesso
  Todas as conexões liberadas
✓ Pool Configurado concluído com sucesso

[... outros exemplos ...]

=== Exemplos de pool concluídos ===
```

## Métricas e Monitoramento

### Métricas Básicas
- **Conexões Ativas**: Conexões em uso
- **Conexões Disponíveis**: Conexões prontas para uso
- **Utilização**: Percentual de uso do pool
- **Throughput**: Operações por segundo

### Monitoramento
```go
// Exemplo de monitoramento
printMetrics := func(activeConns int, phase string) {
    fmt.Printf("  [%s] Métricas:\n", phase)
    fmt.Printf("    Conexões ativas: %d\n", activeConns)
    fmt.Printf("    Conexões disponíveis: %d\n", maxConns-activeConns)
    fmt.Printf("    Utilização: %.1f%%\n", float64(activeConns)/maxConns*100)
}
```

## Boas Práticas

### 1. Dimensionamento
- Use min/max conexões apropriadas para sua carga
- Monitore uso real para ajustar limites
- Considere picos de tráfego

### 2. Timeouts
- Configure timeouts apropriados
- Implemente retry logic
- Monitore timeouts em produção

### 3. Lifecycle
- Configure lifetimes para rotação de conexões
- Implemente health checks
- Monitore conexões "mortas"

### 4. Monitoramento
- Implemente métricas de pool
- Monitore performance
- Configure alertas

## Teste de Carga

### Configuração
```go
numWorkers := 20
operationsPerWorker := 50
totalOperations := numWorkers * operationsPerWorker
```

### Métricas Medidas
- **Tempo Total**: Duração completa do teste
- **Throughput**: Operações por segundo
- **Latência Média**: Tempo médio por operação
- **Latência Min/Max**: Faixa de latência

### Resultados Esperados
```
Resultados do teste:
  Tempo total: 2.5s
  Operações realizadas: 1000
  Throughput: 400.00 ops/sec
  Tempo médio por operação: 2.5ms
  Tempo mínimo: 1.2ms
  Tempo máximo: 15.8ms
```

## Troubleshooting

### Problemas Comuns

1. **Pool Exhausted**
   - Aumentar MaxConns
   - Verificar vazamentos de conexão
   - Implementar timeouts

2. **Conexões Idle**
   - Ajustar MaxConnIdleTime
   - Implementar health checks
   - Monitorar padrões de uso

3. **Performance Degradada**
   - Verificar configurações de pool
   - Analisar queries lentas
   - Implementar connection warming

### Debugging
```go
// Adicionar logging detalhado
log.Printf("Pool stats: active=%d, idle=%d, total=%d", 
    activeConns, idleConns, totalConns)
```

## Próximos Passos

1. Implementar health checks automáticos
2. Adicionar métricas de observabilidade
3. Implementar retry policies
4. Adicionar circuit breakers
5. Implementar connection warming
