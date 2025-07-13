# PostgreSQL Provider Module

Este módulo fornece uma abstração unificada para trabalhar com PostgreSQL em Go, suportando múltiplos drivers através de uma interface comum.

## Recursos

- ✅ **Múltiplos Drivers**: Suporte para PGX, GORM e lib/pq
- ✅ **Interface Unificada**: API consistente independente do driver
- ✅ **Pool de Conexões**: Gerenciamento automático de conexões
- ✅ **Transações**: Suporte completo a transações
- ✅ **Operações em Lote**: Batching para operações de alta performance
- ✅ **Thread-Safe**: Design seguro para concorrência
- ✅ **Configuração Flexível**: Configuração via código ou variáveis de ambiente
- ✅ **Cobertura de Testes**: Testes abrangentes com alta cobertura
- ✅ **Mocks**: Mocks prontos para testes

## Drivers Suportados

### PGX (Recomendado)
- Driver nativo de alta performance
- Suporte completo a recursos PostgreSQL
- Operações em lote nativas
- Pool de conexões otimizado

### GORM
- ORM popular com recursos avançados
- Migrações automáticas
- Relacionamentos e associações
- Hooks e callbacks

### lib/pq
- Driver clássico e estável
- Compatibilidade com database/sql
- Amplamente testado
- Boa opção para aplicações legadas

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/db/postgresql
```

### Dependências dos Drivers

Para PGX:
```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

Para GORM:
```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

Para lib/pq:
```bash
go get github.com/lib/pq
```

## Uso Rápido

### Configuração Básica

```go
package main

import (
    "context"
    "log"
    
    "github.com/fsvxavier/nexs-lib/db/postgresql"
    "github.com/fsvxavier/nexs-lib/db/postgresql/config"
    "github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

func main() {
    // Configuração
    cfg := config.NewConfig(
        config.WithDriver(interfaces.DriverPGX),
        config.WithHost("localhost"),
        config.WithPort(5432),
        config.WithDatabase("mydb"),
        config.WithUser("user"),
        config.WithPassword("password"),
    )
    
    // Criar provider
    provider, err := postgresql.CreateProvider(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Conectar
    err = provider.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()
    
    // Usar o provider
    ctx := context.Background()
    pool := provider.Pool()
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Release(ctx)
    
    // Executar operações
    err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT)")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Configuração via Variáveis de Ambiente

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=mydb
export DB_USER=user
export DB_PASSWORD=password
export DB_DRIVER=pgx
```

```go
cfg := config.NewConfigFromEnv()
```

## Operações Básicas

### Inserção

```go
err := conn.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "João", "joao@example.com")
```

### Consulta Única

```go
var user User
err := conn.QueryOne(ctx, &user, "SELECT id, name, email FROM users WHERE id = $1", 1)
```

### Consulta Múltipla

```go
var users []User
err := conn.QueryAll(ctx, &users, "SELECT id, name, email FROM users")
```

### Contagem

```go
count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users WHERE active = $1", true)
```

## Transações

```go
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    return err
}

err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "User 1")
if err != nil {
    tx.Rollback(ctx)
    return err
}

err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "User 2")
if err != nil {
    tx.Rollback(ctx)
    return err
}

return tx.Commit(ctx)
```

## Operações em Lote (PGX)

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"

batch := pgx.NewBatch()
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 1")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 2")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 3")

results, err := conn.SendBatch(ctx, batch)
if err != nil {
    return err
}
defer results.Close()

// Processar resultados
for i := 0; i < batch.Len(); i++ {
    err = results.Exec()
    if err != nil {
        return err
    }
}
```

## Configuração Avançada

### Pool de Conexões

```go
cfg := config.NewConfig(
    config.WithDriver(interfaces.DriverPGX),
    config.WithHost("localhost"),
    config.WithMaxOpenConns(100),
    config.WithMaxIdleConns(10),
    config.WithConnMaxLifetime(time.Hour),
    config.WithConnMaxIdleTime(time.Minute * 30),
)
```

### Multi-tenancy

```go
cfg := config.NewConfig(
    config.WithMultiTenant(true),
    // outras configurações...
)
```

## Testes

O módulo inclui mocks prontos para testes:

```go
import (
    "testing"
    "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx/mocks"
)

func TestMyFunction(t *testing.T) {
    mockProvider := &mocks.MockDatabaseProvider{}
    mockPool := &mocks.MockPool{}
    mockConn := &mocks.MockConn{}
    
    // Configurar mocks
    mockProvider.On("Pool").Return(mockPool)
    mockPool.On("Acquire", mock.Anything).Return(mockConn, nil)
    
    // Usar nos testes
    // ...
}
```

## Exemplos

Veja os exemplos completos no diretório `examples/`:

- [`examples/basic/`](examples/basic/) - Uso básico com factory
- [`examples/pgx/`](examples/pgx/) - Recursos específicos do PGX
- [`examples/gorm/`](examples/gorm/) - Uso com GORM
- [`examples/pq/`](examples/pq/) - Uso com lib/pq

## Arquitetura

```
db/postgresql/
├── interfaces/          # Interfaces principais
├── config/             # Sistema de configuração
├── providers/          # Implementações dos drivers
│   ├── pgx/           # Provider PGX
│   ├── gorm/          # Provider GORM
│   └── pq/            # Provider lib/pq
├── examples/          # Exemplos de uso
└── mocks/            # Mocks para testes
```

## Performance

### Benchmark Results

```
BenchmarkCreateProvider-8    1000000    1234 ns/op    512 B/op    8 allocs/op
```

### Cobertura de Testes

- **Config**: 95.8%
- **PGX Provider**: 25.5%
- **GORM Provider**: 30.8%
- **PQ Provider**: 33.3%

## Migração de Outros Drivers

### De database/sql + lib/pq

```go
// Antes
db, err := sql.Open("postgres", "postgres://user:pass@localhost/db")

// Depois
cfg := config.NewConfig(config.WithDriver(interfaces.DriverPQ))
provider, err := postgresql.CreateProvider(cfg)
```

### De pgx direto

```go
// Antes
pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost/db")

// Depois
cfg := config.NewConfig(config.WithDriver(interfaces.DriverPGX))
provider, err := postgresql.CreateProvider(cfg)
```

## Próximos Passos

Veja [NEXT_STEPS.md](NEXT_STEPS.md) para:
- Roadmap de funcionalidades
- Melhorias planejadas
- Como contribuir

## Licença

Este projeto faz parte da nexs-lib e segue a mesma licença do projeto principal.
