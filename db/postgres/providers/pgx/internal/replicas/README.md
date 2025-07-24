# Sistema de Read Replicas PostgreSQL

O sistema de Read Replicas foi desenvolvido para fornecer balanceamento de carga inteligente e alta disponibilidade para operações de leitura em bancos de dados PostgreSQL.

## Funcionalidades

### 🔄 Estratégias de Balanceamento de Carga

1. **Round-Robin**: Distribui requisições de forma circular entre as réplicas
2. **Random**: Distribui requisições aleatoriamente
3. **Weighted**: Distribui baseado em pesos configurados para cada réplica
4. **Latency-based**: Distribui baseado na latência medida de cada réplica

### 📊 Preferências de Leitura

- **Primary**: Força leitura apenas no primary
- **Secondary**: Força leitura apenas nas réplicas
- **SecondaryPreferred**: Prefere réplicas, mas pode usar primary como fallback
- **Nearest**: Usa a conexão com menor latência

### 🏥 Health Checking Automático

- Verificação periódica de saúde das réplicas
- Detecção automática de falhas
- Recuperação automática quando réplicas voltam ao ar
- Métricas de latência e taxa de sucesso

### 🔧 Gerenciamento de Status

- **Healthy**: Réplica funcionando normalmente
- **Unhealthy**: Réplica com problemas
- **Recovering**: Réplica se recuperando
- **Maintenance**: Réplica em manutenção

## Arquitetura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │  Replica Pool   │    │ Load Balancer   │
│                 │────│                 │────│                 │
│  Read Queries   │    │  Read Routing   │    │   Strategies    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                 │
                                 │
                    ┌─────────────────┐
                    │ Replica Manager │
                    │                 │
                    │ Health Checks   │
                    │ Statistics      │
                    │ Failover        │
                    └─────────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                        │                        │
  ┌───────────┐           ┌───────────┐           ┌───────────┐
  │ Replica 1 │           │ Replica 2 │           │ Replica 3 │
  │           │           │           │           │           │
  │ Weight: 10│           │ Weight: 20│           │ Weight: 30│
  │ Status: ✅│           │ Status: ✅│           │ Status: 🔧│
  └───────────┘           └───────────┘           └───────────┘
```

## Uso Básico

### 1. Configuração do Replica Manager

```go
import (
    "context"
    "time"
    "github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
    "github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/replicas"
)

// Configurar replica manager
config := replicas.ReplicaManagerConfig{
    LoadBalancingStrategy: interfaces.LoadBalancingRoundRobin,
    ReadPreference:       interfaces.ReadPreferenceSecondaryPreferred,
    HealthCheckInterval:  30 * time.Second,
    HealthCheckTimeout:   5 * time.Second,
}

replicaManager := replicas.NewReplicaManager(config)
```

### 2. Adicionar Réplicas

```go
ctx := context.Background()

// Adicionar réplicas com diferentes pesos
err := replicaManager.AddReplica(ctx, "replica-1", "postgres://replica1:5432/mydb", 10)
err = replicaManager.AddReplica(ctx, "replica-2", "postgres://replica2:5432/mydb", 20)
err = replicaManager.AddReplica(ctx, "replica-3", "postgres://replica3:5432/mydb", 30)

// Iniciar gerenciador
err = replicaManager.Start(ctx)
```

### 3. Seleção de Réplicas

```go
// Selecionar réplica com preferência
replica, err := replicaManager.SelectReplica(ctx, interfaces.ReadPreferenceSecondary)

// Selecionar réplica com estratégia específica
replica, err := replicaManager.SelectReplicaWithStrategy(ctx, interfaces.LoadBalancingWeighted)
```

### 4. Usando ReplicaPool

```go
// Criar pool de réplicas
builder := replicas.NewReplicaPoolBuilder(primaryPool)
replicaPool, err := builder.
    SetLoadBalancingStrategy(interfaces.LoadBalancingLatency).
    SetReadPreference(interfaces.ReadPreferenceSecondaryPreferred).
    AddReplica("replica-1", "postgres://replica1:5432/mydb", 10).
    AddReplica("replica-2", "postgres://replica2:5432/mydb", 20).
    Build(ctx)

// Usar pool para leituras
conn, err := replicaPool.AcquireRead(ctx, interfaces.ReadPreferenceSecondary)
defer conn.Release()

// Usar pool para escritas (sempre primary)
conn, err := replicaPool.AcquireWrite(ctx)
defer conn.Release()
```

## Configuração Avançada

### Health Check Personalizado

```go
replicaManager.SetHealthCheckInterval(15 * time.Second)
replicaManager.SetHealthCheckTimeout(3 * time.Second)
```

### Callbacks de Eventos

```go
// Callback para mudanças de saúde
replicaManager.OnReplicaHealthChange(func(replica interfaces.IReplicaInfo, oldStatus, newStatus interfaces.ReplicaStatus) {
    log.Printf("Réplica %s mudou de %s para %s", replica.GetID(), oldStatus, newStatus)
})

// Callback para failover
replicaManager.OnReplicaFailover(func(from, to interfaces.IReplicaInfo) {
    log.Printf("Failover de %s para %s", from.GetID(), to.GetID())
})
```

### Manutenção de Réplicas

```go
// Colocar réplica em manutenção
err := replicaManager.SetReplicaMaintenance("replica-1", true)

// Drenar réplica gradualmente
err := replicaManager.DrainReplica(ctx, "replica-1", 30*time.Second)

// Remover réplica
err := replicaManager.RemoveReplica(ctx, "replica-1")
```

## Monitoramento e Estatísticas

### Estatísticas Gerais

```go
stats := replicaManager.GetStats()

fmt.Printf("Total de réplicas: %d\n", stats.GetTotalReplicas())
fmt.Printf("Réplicas saudáveis: %d\n", stats.GetHealthyReplicas())
fmt.Printf("Total de queries: %d\n", stats.GetTotalQueries())
fmt.Printf("Taxa de sucesso: %.2f%%\n", stats.GetSuccessfulQueries() / stats.GetTotalQueries() * 100)
fmt.Printf("Latência média: %v\n", stats.GetAvgLatency())
fmt.Printf("Uptime: %v\n", stats.GetUptime())
```

### Estatísticas por Réplica

```go
replicas := replicaManager.ListReplicas()
for _, replica := range replicas {
    fmt.Printf("Réplica %s:\n", replica.GetID())
    fmt.Printf("  Status: %s\n", replica.GetStatus())
    fmt.Printf("  Taxa de sucesso: %.2f%%\n", replica.GetSuccessRate())
    fmt.Printf("  Latência média: %v\n", replica.GetAvgLatency())
    fmt.Printf("  Conexões ativas: %d\n", replica.GetConnectionCount())
}
```

### Export de Métricas

```go
// Export para JSON
jsonData, err := stats.ToJSON()

// Export para Map (útil para Prometheus)
metricsMap := stats.ToMap()
```

## Integração com Provider PGX

O sistema está integrado ao provider PGX e pode ser acessado através do método `GetReplicaManager()`:

```go
provider := pgxprovider.NewProvider()
replicaManager := provider.GetReplicaManager()

// Configurar réplicas
// ...
```

## Estratégias de Balanceamento Detalhadas

### Round-Robin
- Distribui requisições sequencialmente
- Ideal para réplicas com capacidade similar
- Baixa latência de seleção

### Random
- Distribui requisições aleatoriamente  
- Boa para distribuição uniforme
- Simples e eficiente

### Weighted
- Distribui baseado em pesos configurados
- Ideal para réplicas com capacidades diferentes
- Permite controle fino da distribuição

### Latency-based
- Seleciona réplica com menor latência
- Ideal para otimização de performance
- Adapta-se automaticamente às condições de rede

## Tolerância a Falhas

O sistema oferece várias camadas de tolerância a falhas:

1. **Health Check Automático**: Detecção proativa de falhas
2. **Failover Automático**: Remoção automática de réplicas não saudáveis
3. **Recuperação Automática**: Reintegração automática após recuperação
4. **Fallback para Primary**: Quando nenhuma réplica está disponível

## Performance

O sistema é otimizado para alta performance:

- **Operações thread-safe**: Uso de sync.RWMutex para leitura concorrente
- **Atomic operations**: Contadores usando operações atômicas
- **Pool de conexões**: Reutilização eficiente de conexões
- **Caching**: Cache de estatísticas e métricas

## Exemplo Completo

Veja o exemplo completo em `example/main.go` para uma demonstração prática do sistema.

## Contribuindo

Para contribuir com melhorias no sistema de réplicas:

1. Mantenha os testes com cobertura mínima de 98%
2. Documente todas as funcionalidades públicas
3. Siga os padrões de código Go
4. Adicione benchmarks para código crítico

## Roadmap

- [ ] Suporte a múltiplas regions
- [ ] Integração com service discovery
- [ ] Métricas Prometheus nativas
- [ ] Suporte a réplicas de diferentes versões
- [ ] Balanceamento baseado em CPU/Memory
