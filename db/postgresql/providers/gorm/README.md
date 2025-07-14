# GORM Provider para PostgreSQL

Este é o provider GORM para o módulo PostgreSQL da nexs-lib. Ele implementa as interfaces padrão do postgresql usando GORM como ORM.

## Características

- **ORM**: GORM v1.30+
- **Driver**: postgres
- **Pool de Conexões**: Suportado via GORM
- **Transações**: Suportado
- **Savepoints**: Suportado
- **Prepared Statements**: Gerenciado pelo GORM
- **Batch Operations**: ❌ Não suportado
- **LISTEN/NOTIFY**: ❌ Não suportado

## Uso Básico

### Criar Provider

```go
provider := gorm.NewProvider()
```

### Criar Pool de Conexões

```go
cfg := &config.Config{
    Host:            "localhost",
    Port:            5432,
    Database:        "mydb",
    Username:        "user",
    Password:        "password",
    TLSMode:         config.TLSModeDisable,
    MaxConns:        10,
    MinConns:        2,
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: time.Minute * 30,
}

pool, err := provider.CreatePool(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()
```

### Usar Conexão

```go
conn, err := pool.Acquire(ctx)
if err != nil {
    log.Fatal(err)
}
defer conn.Release(ctx)

// Executar query
var users []User
err = conn.QueryAll(ctx, &users, "SELECT * FROM users WHERE active = ?", true)
if err != nil {
    log.Fatal(err)
}
```

### Usar Transações

```go
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    log.Fatal(err)
}

err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
if err != nil {
    tx.Rollback(ctx)
    log.Fatal(err)
}

err = tx.Commit(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Usar Savepoints

```go
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    log.Fatal(err)
}

// Criar savepoint
err = tx.BeginSavepoint(ctx, "sp1")
if err != nil {
    log.Fatal(err)
}

// Fazer operações...
err = tx.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Test")
if err != nil {
    // Rollback para savepoint
    tx.RollbackToSavepoint(ctx, "sp1")
} else {
    // Liberar savepoint
    tx.ReleaseSavepoint(ctx, "sp1")
}

tx.Commit(ctx)
```

## Configuração

O provider GORM suporta todas as opções de configuração padrão:

- **Host/Port/Database/Username/Password**: Credenciais básicas
- **ConnString**: String de conexão completa (opcional)
- **TLSMode**: Modo SSL/TLS
- **ConnectTimeout**: Timeout de conexão
- **MaxConns/MinConns**: Limites do pool
- **MaxConnLifetime/MaxConnIdleTime**: Tempos de vida das conexões
- **RuntimeParams**: Parâmetros adicionais do PostgreSQL

## Limitações

### Batch Operations
O GORM não suporta operações em batch tradicionais do PostgreSQL. Métodos como `SendBatch` retornarão erro.

### LISTEN/NOTIFY
O GORM não suporta as funcionalidades LISTEN/NOTIFY do PostgreSQL. Métodos como `Listen`, `Unlisten` e `WaitForNotification` retornarão erro.

### Prepared Statements
O GORM gerencia prepared statements internamente. O método `Prepare` não realiza operações específicas.

## Performance

### Benchmarks

```
BenchmarkProvider_CreatePool          377.5 ns/op    240 B/op    4 allocs/op
BenchmarkProvider_CreateConnection    357.0 ns/op    240 B/op    4 allocs/op
BenchmarkConnection_QueryOne          18430 ns/op    11293 B/op  80 allocs/op
BenchmarkConnection_QueryAll          108966 ns/op   6176 B/op   65 allocs/op
BenchmarkConnection_Exec              108103 ns/op   6027 B/op   55 allocs/op
BenchmarkTransaction_Operations       889979 ns/op   8133 B/op   74 allocs/op
BenchmarkPool_AcquireRelease          41.49 ns/op    24 B/op     1 allocs/op
```

### Otimizações

- **Pool de Conexões**: Configure adequadamente MaxConns e MinConns
- **Prepared Statements**: GORM gerencia automaticamente
- **Timeouts**: Configure ConnectTimeout adequadamente
- **SSL**: Use TLSModeDisable em ambientes locais para melhor performance

## Testes

### Executar Testes

```bash
# Testes unitários
go test -v -timeout 30s -race .

# Testes com cobertura
go test -v -coverprofile=coverage.out -timeout 30s -race .

# Benchmarks
go test -run=^$ -bench=. -benchmem -timeout 30s .

# Testes de integração (requer PostgreSQL)
go test -tags=integration -v ./...
```

### Cobertura Atual

**Cobertura: 41.8%**

Veja [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) para detalhes completos.

### Mock Testing

O provider inclui testes abrangentes usando SQLMock:

```go
func TestConnection_WithMock(t *testing.T) {
    sqlDB, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer sqlDB.Close()

    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: sqlDB,
    }), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)

    conn := &Conn{
        db:     gormDB,
        config: &config.Config{},
    }

    mock.ExpectExec(`INSERT INTO users`).
        WithArgs("John").
        WillReturnResult(sqlmock.NewResult(1, 1))

    err = conn.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "John")
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## Dependências

```go
require (
    github.com/DATA-DOG/go-sqlmock v1.5.0
    github.com/stretchr/testify v1.8.4
    gorm.io/driver/postgres v1.5.4
    gorm.io/gorm v1.25.5
)
```

## Estrutura do Código

```
├── provider.go              # Implementação do provider
├── pool.go                  # Implementação do pool de conexões
├── conn.go                  # Implementação das conexões
├── rows.go                  # Implementação de rows, row e transaction
├── provider_test.go         # Testes originais
├── provider_clean_test.go   # Testes limpos e funcionais
├── provider_comprehensive_test.go  # Testes abrangentes
├── provider_benchmark_test.go      # Benchmarks
├── provider_integration_test.go    # Testes de integração
└── TEST_COVERAGE_REPORT.md         # Relatório de cobertura
```

## Exemplo Completo

Veja os exemplos em `../examples/` para uso prático do provider GORM.

## Contribuição

Para contribuir com este provider:

1. Mantenha cobertura de testes acima de 98%
2. Execute todos os testes antes de submeter
3. Inclua benchmarks para novas funcionalidades
4. Documente limitações e comportamentos específicos do GORM
5. Use SQLMock para testes unitários quando apropriado

## Licença

Este código faz parte da nexs-lib e segue a mesma licença do projeto principal.
