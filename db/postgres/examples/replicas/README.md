# Read Replicas Example

Este exemplo demonstra como usar Read Replicas com PostgreSQL.

## Funcionalidades Demonstradas

- **Configuração de Read Replicas**: Conectar com primary e réplicas
- **Balanceamento de Carga**: Distribuir queries entre réplicas
- **Failover**: Recuperação automática quando uma réplica falha
- **Separação Read/Write**: Writes no primary, reads nas réplicas

## Pré-requisitos

1. **Infraestrutura**: Execute o script de infraestrutura para configurar PostgreSQL com réplicas:
   ```bash
   ./infrastructure/manage.sh start
   ```

2. **Aguardar**: Aguarde alguns segundos para que as réplicas sejam configuradas.

## Como Executar

```bash
# Executar com a infraestrutura do projeto
./infrastructure/manage.sh example replicas

# Ou executar manualmente
cd db/postgres/examples/replicas
go run main.go
```

## Variáveis de Ambiente

```bash
# Configurações das conexões
export NEXS_DB_PRIMARY_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
export NEXS_DB_REPLICA1_DSN="postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
export NEXS_DB_REPLICA2_DSN="postgres://nexs_user:nexs_password@localhost:5434/nexs_testdb"
```

## Estrutura do Exemplo

### 1. Configuração das Conexões
```go
primaryDSN := getEnvOrDefault("NEXS_DB_PRIMARY_DSN", "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
replica1DSN := getEnvOrDefault("NEXS_DB_REPLICA1_DSN", "postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb")
replica2DSN := getEnvOrDefault("NEXS_DB_REPLICA2_DSN", "postgres://nexs_user:nexs_password@localhost:5434/nexs_testdb")
```

### 2. Uso Real com Conexões
```go
// Criar pools de conexão
primaryPool, err := postgres.ConnectPoolWithConfig(ctx, primaryCfg)
replica1Pool, err := postgres.ConnectPoolWithConfig(ctx, replica1Cfg)
replica2Pool, err := postgres.ConnectPoolWithConfig(ctx, replica2Cfg)

// Escrever no primary
primaryConn, err := primaryPool.Acquire(ctx)
_, err = primaryConn.Exec(ctx, "INSERT INTO table VALUES ...")

// Ler das réplicas
replica1Conn, err := replica1Pool.Acquire(ctx)
err = replica1Conn.QueryRow(ctx, "SELECT * FROM table").Scan(&result)
```

### 3. Balanceamento de Carga Manual
```go
replicas := []string{replica1DSN, replica2DSN}

for i := 0; i < 6; i++ {
    replicaIndex := i % len(replicas)
    conn, err := postgres.Connect(ctx, replicas[replicaIndex])
    // Executar query na réplica selecionada
}
```

### 4. Failover Básico
```go
replicas := []struct {
    dsn  string
    name string
}{
    {replica1DSN, "replica1"},
    {replica2DSN, "replica2"},
}

for _, replica := range replicas {
    conn, err := postgres.Connect(ctx, replica.dsn)
    if err != nil {
        continue // Tentar próxima réplica
    }
    
    err = conn.Ping(ctx)
    if err == nil {
        // Réplica está saudável
        break
    }
}
```

## Recursos Avançados

### Monitoramento de Saúde
- Health checks automáticos
- Detecção de falhas
- Recuperação automática

### Estratégias de Balanceamento
- **Round Robin**: Distribuição circular
- **Random**: Seleção aleatória
- **Weighted**: Baseado em pesos
- **Latency**: Baseado na latência

### Preferências de Leitura
- **Secondary**: Apenas réplicas
- **Secondary Preferred**: Réplicas preferenciais
- **Nearest**: Réplica mais próxima

## Arquitetura

```
Primary Database (localhost:5432)
├── Writes: INSERT, UPDATE, DELETE
├── Reads: SELECT (quando réplicas indisponíveis)
└── Replication: Streaming replication

Replica 1 (localhost:5433)
├── Reads: SELECT queries
├── Replication: Streaming from primary
└── Health: Monitored

Replica 2 (localhost:5434)
├── Reads: SELECT queries
├── Replication: Streaming from primary
└── Health: Monitored
```

## Benefícios

1. **Escalabilidade**: Distribuir carga de leitura
2. **Disponibilidade**: Redundância para reads
3. **Performance**: Queries paralelas
4. **Isolamento**: Separar workloads analíticos

## Próximos Passos

- Veja o exemplo `advanced/` para configurações avançadas
- Veja o exemplo `performance/` para otimizações
- Consulte a documentação sobre connection pooling
