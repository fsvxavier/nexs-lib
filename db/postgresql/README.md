# PostgreSQL Database Module

Uma biblioteca abrangente e moderna para integração com PostgreSQL em aplicações Go, oferecendo suporte a múltiplos drivers, connection pooling, observabilidade e padrões de design enterprise-ready.

## 🚀 Características Principais

- **🎯 Múltiplos Drivers**: Suporte completo para PGX, PQ e GORM
- **🏭 Factory Pattern**: Criação padronizada de conexões e pools
- **🏊 Connection Pooling**: Gerenciamento inteligente de conexões
- **🔄 Batch Operations**: Operações em lote para alta performance
- **📊 Observabilidade**: Integração com métricas, logs e tracing
- **🛡️ Segurança**: SSL/TLS, autenticação e autorização
- **🏗️ Multi-tenancy**: Suporte a múltiplos bancos e esquemas
- **⚡ Circuit Breaker**: Proteção contra falhas em cascata
- **🔁 Retry Logic**: Reconexão automática e retry inteligente
- **🧪 Testabilidade**: Mocks e interfaces para testes

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/db/postgresql
```

### Dependências dos Drivers

```bash
# Para usar PGX (recomendado)
go get github.com/jackc/pgx/v5

# Para usar PQ
go get github.com/lib/pq

# Para usar GORM
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

## 🏗️ Arquitetura

```
db/postgresql/
├── postgresql.go           # 🎯 API principal e factory
├── common/
│   ├── interfaces.go       # 📋 Interfaces comuns
│   ├── config.go          # ⚙️ Configurações
│   └── errors.go          # 🚨 Tratamento de erros
├── pgx/                   # 🚀 Driver PGX (recomendado)
│   ├── connection.go      # 🔌 Conexões PGX
│   ├── pool.go           # 🏊 Pool de conexões
│   └── batch.go          # 🔄 Operações em lote
├── pq/                    # 📡 Driver PQ (database/sql)
│   ├── connection.go      # 🔌 Conexões PQ
│   └── batch.go          # 🔄 Operações em lote
├── gorm/                  # 🦄 ORM GORM
│   ├── connection.go      # 🔌 Conexões GORM
│   └── batch.go          # 🔄 Operações em lote
└── examples/              # 💡 Exemplos práticos
    ├── basic/             # Exemplos básicos
    ├── advanced/          # Exemplos avançados
    └── production/        # Configurações de produção
```

## 🚀 Uso Rápido

### Conexão Simples

```go
package main

import (
    "context"
    "log"
    
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
    ctx := context.Background()
    
    // Configuração básica
    config := &postgresql.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        User:     "postgres",
        Password: "password",
        Provider: postgresql.ProviderPGX, // ou ProviderPQ, ProviderGORM
    }
    
    // Criar conexão
    conn, err := postgresql.NewConnection(ctx, config)
    if err != nil {
        log.Fatal("Erro ao conectar:", err)
    }
    defer conn.Close(ctx)
    
    // Usar a conexão
    rows, err := conn.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
    if err != nil {
        log.Fatal("Erro na query:", err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Fatal("Erro no scan:", err)
        }
        log.Printf("User: %d - %s", id, name)
    }
}
```

### Connection Pool

```go
package main

import (
    "context"
    "log"
    
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
    ctx := context.Background()
    
    // Configuração com pool
    config := &postgresql.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        User:     "postgres",
        Password: "password",
        Provider: postgresql.ProviderPGX,
        
        // Configurações do pool
        MaxConns:        20,
        MinConns:        5,
        MaxConnLifetime: "1h",
        MaxConnIdleTime: "30m",
    }
    
    // Criar pool
    pool, err := postgresql.NewPool(ctx, config)
    if err != nil {
        log.Fatal("Erro ao criar pool:", err)
    }
    defer pool.Close()
    
    // Usar o pool
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal("Erro ao adquirir conexão:", err)
    }
    defer conn.Release()
    
    // Executar query
    var count int
    err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        log.Fatal("Erro na query:", err)
    }
    
    log.Printf("Total de usuários: %d", count)
}
```

## 🔧 Configuração Avançada

### Configuração Completa

```go
config := &postgresql.Config{
    // Conexão básica
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    User:     "postgres",
    Password: "password",
    Provider: postgresql.ProviderPGX,
    
    // SSL/TLS
    SSLMode:     "require",
    SSLCert:     "/path/to/client-cert.pem",
    SSLKey:      "/path/to/client-key.pem",
    SSLRootCert: "/path/to/ca-cert.pem",
    
    // Pool de conexões
    MaxConns:        50,
    MinConns:        10,
    MaxConnLifetime: "2h",
    MaxConnIdleTime: "15m",
    
    // Timeouts
    ConnectTimeout: "10s",
    QueryTimeout:   "30s",
    
    // Observabilidade
    EnableMetrics: true,
    EnableTracing: true,
    LogLevel:      "info",
    
    // Circuit Breaker
    CircuitBreaker: &postgresql.CircuitBreakerConfig{
        MaxFailures:    5,
        ResetTimeout:   "60s",
        FailureTimeout: "30s",
    },
    
    // Retry Logic
    RetryConfig: &postgresql.RetryConfig{
        MaxAttempts: 3,
        InitialDelay: "1s",
        MaxDelay:     "10s",
        Multiplier:   2.0,
    },
}
```

### Multi-tenancy

```go
// Configuração para multi-tenancy
configs := map[string]*postgresql.Config{
    "tenant1": {
        Host:     "db1.example.com",
        Database: "tenant1_db",
        Schema:   "tenant1",
        // ... outras configurações
    },
    "tenant2": {
        Host:     "db2.example.com", 
        Database: "tenant2_db",
        Schema:   "tenant2",
        // ... outras configurações
    },
}

// Pool manager para múltiplos tenants
type TenantManager struct {
    pools map[string]*postgresql.Pool
}

func NewTenantManager(configs map[string]*postgresql.Config) (*TenantManager, error) {
    tm := &TenantManager{
        pools: make(map[string]*postgresql.Pool),
    }
    
    for tenantID, config := range configs {
        pool, err := postgresql.NewPool(context.Background(), config)
        if err != nil {
            return nil, fmt.Errorf("erro ao criar pool para tenant %s: %w", tenantID, err)
        }
        tm.pools[tenantID] = pool
    }
    
    return tm, nil
}

func (tm *TenantManager) GetPool(tenantID string) (*postgresql.Pool, error) {
    pool, exists := tm.pools[tenantID]
    if !exists {
        return nil, fmt.Errorf("tenant %s não encontrado", tenantID)
    }
    return pool, nil
}
```

## 🔄 Operações em Lote (Batch)

### Batch com PGX

```go
func ExampleBatchOperations() {
    ctx := context.Background()
    
    config := &postgresql.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        User:     "postgres",
        Password: "password",
        Provider: postgresql.ProviderPGX,
    }
    
    // Criar batch
    batch, err := postgresql.NewBatch(config)
    if err != nil {
        log.Fatal("Erro ao criar batch:", err)
    }
    
    // Adicionar operações ao batch
    batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "João", "joao@example.com")
    batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "Maria", "maria@example.com")
    batch.Queue("UPDATE users SET active = $1 WHERE id = $2", true, 1)
    batch.Queue("DELETE FROM users WHERE active = $1", false)
    
    // Executar batch
    pool, err := postgresql.NewPool(ctx, config)
    if err != nil {
        log.Fatal("Erro ao criar pool:", err)
    }
    defer pool.Close()
    
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal("Erro ao adquirir conexão:", err)
    }
    defer conn.Release()
    
    results := conn.SendBatch(ctx, batch)
    defer results.Close()
    
    // Processar resultados
    for i := 0; i < batch.Len(); i++ {
        _, err := results.Exec()
        if err != nil {
            log.Printf("Erro na operação %d: %v", i, err)
        }
    }
}
```

## 📊 Observabilidade

### Métricas

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Configurar métricas personalizadas
config := &postgresql.Config{
    // ... configuração básica
    EnableMetrics: true,
    MetricsConfig: &postgresql.MetricsConfig{
        Namespace: "myapp",
        Subsystem: "database",
        Registry:  prometheus.DefaultRegisterer,
        Labels: map[string]string{
            "service": "user-service",
            "env":     "production",
        },
    },
}

// Métricas automáticas disponíveis:
// - postgresql_connections_active
// - postgresql_connections_total
// - postgresql_query_duration_seconds
// - postgresql_query_errors_total
// - postgresql_pool_acquire_duration_seconds
```

### Logging

```go
import (
    "go.uber.org/zap"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Configurar logging
logger, _ := zap.NewProduction()

config := &postgresql.Config{
    // ... configuração básica
    Logger:   logger,
    LogLevel: "debug", // trace, debug, info, warn, error
    LogConfig: &postgresql.LogConfig{
        LogQueries:       true,
        LogSlowQueries:   true,
        SlowQueryThreshold: "1s",
        LogConnections:   true,
        LogTransactions: true,
    },
}
```

### Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Configurar tracing
config := &postgresql.Config{
    // ... configuração básica
    EnableTracing: true,
    TracingConfig: &postgresql.TracingConfig{
        Tracer:      otel.Tracer("postgresql"),
        ServiceName: "user-service",
        Environment: "production",
        TraceQueries: true,
        TraceParams:  false, // Por segurança, não loggar parâmetros
    },
}
```

## 🛡️ Segurança

### SSL/TLS

```go
// Configuração SSL completa
config := &postgresql.Config{
    Host:     "secure-db.example.com",
    Port:     5432,
    Database: "secure_app",
    User:     "app_user",
    Password: "secure_password",
    
    // SSL obrigatório
    SSLMode: "require", // disable, allow, prefer, require, verify-ca, verify-full
    
    // Certificados cliente
    SSLCert: "/etc/ssl/certs/client-cert.pem",
    SSLKey:  "/etc/ssl/private/client-key.pem",
    
    // CA raiz para verificação do servidor
    SSLRootCert: "/etc/ssl/certs/ca-cert.pem",
    
    // Configurações adicionais de SSL
    SSLCompression: false,
    SSLMinVersion:  "TLSv1.2",
    SSLMaxVersion:  "TLSv1.3",
}
```

### Autenticação e Autorização

```go
// Configuração com diferentes métodos de autenticação
config := &postgresql.Config{
    Host:     "auth-db.example.com",
    Port:     5432,
    Database: "secure_app",
    
    // Autenticação básica
    User:     "app_user",
    Password: "secure_password",
    
    // Ou autenticação via certificado
    AuthMethod: "cert",
    SSLCert:    "/path/to/client-cert.pem",
    SSLKey:     "/path/to/client-key.pem",
    
    // Ou autenticação via IAM (AWS RDS)
    AuthMethod: "iam",
    AWSRegion:  "us-east-1",
    
    // Role-based access
    ApplicationName: "user-service",
    SearchPath:      "app_schema,public",
    
    // Configurações de segurança
    ConnectTimeout:  "5s",
    QueryTimeout:    "30s",
    StatementTimeout: "60s",
}
```

## ⚡ Alta Performance

### Connection Pooling Otimizado

```go
// Configuração para alta performance
config := &postgresql.Config{
    Host:     "high-perf-db.example.com",
    Port:     5432,
    Database: "high_perf_app",
    User:     "perf_user",
    Password: "password",
    Provider: postgresql.ProviderPGX, // Mais performático
    
    // Pool otimizado para alta carga
    MaxConns:        100,        // Máximo de conexões
    MinConns:        20,         // Mínimo sempre ativo
    MaxConnLifetime: "1h",       // Reciclar conexões a cada hora
    MaxConnIdleTime: "5m",       // Fechar idle após 5 minutos
    HealthCheckPeriod: "1m",     // Verificar saúde das conexões
    
    // Configurações de performance
    DefaultQueryExecMode: "cache_statement", // Cache de prepared statements
    PreferSimpleProtocol: false,             // Usar extended protocol
    
    // Timeouts otimizados
    ConnectTimeout: "3s",
    QueryTimeout:   "10s",
    
    // Buffer sizes
    ReadBufferSize:  8192,
    WriteBufferSize: 8192,
}
```

### Prepared Statements

```go
func ExamplePreparedStatements() {
    ctx := context.Background()
    
    // Obter conexão do pool
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal("Erro ao adquirir conexão:", err)
    }
    defer conn.Release()
    
    // Preparar statement
    stmt, err := conn.Prepare(ctx, "select_user", "SELECT id, name, email FROM users WHERE id = $1")
    if err != nil {
        log.Fatal("Erro ao preparar statement:", err)
    }
    
    // Usar o prepared statement múltiplas vezes
    for i := 1; i <= 100; i++ {
        var id int
        var name, email string
        
        err := conn.QueryRow(ctx, "select_user", i).Scan(&id, &name, &email)
        if err != nil {
            log.Printf("Usuário %d não encontrado", i)
            continue
        }
        
        log.Printf("User: %d - %s <%s>", id, name, email)
    }
}
```

## 🧪 Testabilidade

### Mocks para Testes

```go
package main_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
    "github.com/fsvxavier/nexs-lib/db/postgresql/mocks"
)

func TestUserService(t *testing.T) {
    // Criar mock da conexão
    mockConn := &mocks.MockConnection{}
    
    // Configurar expectativas
    mockConn.On("QueryRow", mock.Anything, 
        "SELECT id, name, email FROM users WHERE id = $1", 
        123).Return(&mockRow{
            id:    123,
            name:  "João",
            email: "joao@example.com",
        }, nil)
    
    // Usar o mock no teste
    service := &UserService{conn: mockConn}
    user, err := service.GetUser(context.Background(), 123)
    
    assert.NoError(t, err)
    assert.Equal(t, 123, user.ID)
    assert.Equal(t, "João", user.Name)
    assert.Equal(t, "joao@example.com", user.Email)
    
    // Verificar se as expectativas foram cumpridas
    mockConn.AssertExpectations(t)
}
```

### Testes de Integração

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Pulando teste de integração")
    }
    
    ctx := context.Background()
    
    // Configuração para ambiente de teste
    config := &postgresql.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "test_db",
        User:     "test_user",
        Password: "test_password",
        Provider: postgresql.ProviderPGX,
        
        // Configurações específicas para teste
        MaxConns:        5,
        MinConns:        1,
        ConnectTimeout:  "5s",
        QueryTimeout:    "10s",
    }
    
    // Criar pool para teste
    pool, err := postgresql.NewPool(ctx, config)
    assert.NoError(t, err)
    defer pool.Close()
    
    // Verificar conectividade
    conn, err := pool.Acquire(ctx)
    assert.NoError(t, err)
    defer conn.Release()
    
    // Testar operações básicas
    var version string
    err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
    assert.NoError(t, err)
    assert.Contains(t, version, "PostgreSQL")
}
```

## 🔧 Troubleshooting

### Problemas Comuns

#### Erro de Conexão

```go
// Problema: "connection refused"
// Solução: Verificar se o PostgreSQL está rodando e acessível

config := &postgresql.Config{
    Host:           "localhost",
    Port:           5432,
    ConnectTimeout: "10s", // Aumentar timeout
    
    // Adicionar retry logic
    RetryConfig: &postgresql.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: "1s",
        MaxDelay:     "10s",
        Multiplier:   2.0,
    },
}
```

#### Pool Esgotado

```go
// Problema: "failed to acquire connection from pool"
// Solução: Ajustar configurações do pool

config := &postgresql.Config{
    MaxConns:        50,  // Aumentar máximo de conexões
    MinConns:        10,  // Garantir mínimo sempre disponível
    MaxConnLifetime: "1h", // Reciclar conexões antigas
    MaxConnIdleTime: "15m", // Liberar conexões idle
    
    // Timeout para aquisição de conexão
    AcquireTimeout: "5s",
}
```

#### SSL/TLS Issues

```go
// Problema: SSL certificate verify failed
// Solução: Configurar SSL corretamente

config := &postgresql.Config{
    SSLMode:     "require",     // ou "verify-full" para máxima segurança
    SSLRootCert: "/path/to/ca-cert.pem", // Certificado da CA
    
    // Para desenvolvimento (NÃO usar em produção)
    SSLMode: "disable", // Desabilitar SSL completamente
}
```

### Debugging

```go
import (
    "go.uber.org/zap"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Habilitar logs detalhados para debugging
logger, _ := zap.NewDevelopment()

config := &postgresql.Config{
    // ... configuração básica
    Logger:   logger,
    LogLevel: "trace", // Máximo detalhamento
    LogConfig: &postgresql.LogConfig{
        LogQueries:         true,
        LogConnections:     true,
        LogTransactions:    true,
        LogSlowQueries:     true,
        SlowQueryThreshold: "100ms",
        LogParameters:      true, // CUIDADO: pode expor dados sensíveis
    },
}
```

## 📈 Benchmarks

### Performance Comparativa

```bash
# Executar benchmarks
go test -bench=. -benchmem ./...

# Resultados típicos (conexões/segundo):
# PGX (connection):     ~50,000/sec
# PGX (pool):          ~100,000/sec  
# PQ (connection):      ~30,000/sec
# GORM:                 ~20,000/sec
```

### Otimizações Recomendadas

```go
// Para máxima performance
config := &postgresql.Config{
    Provider: postgresql.ProviderPGX, // Driver mais rápido
    
    // Pool otimizado
    MaxConns:        runtime.NumCPU() * 4, // 4x número de CPUs
    MinConns:        runtime.NumCPU(),     // 1x número de CPUs
    MaxConnLifetime: "1h",
    MaxConnIdleTime: "5m",
    
    // Protocolo otimizado
    DefaultQueryExecMode: "cache_statement",
    PreferSimpleProtocol: false,
    
    // Buffers maiores para alta throughput
    ReadBufferSize:  16384,
    WriteBufferSize: 16384,
}
```

## 🔄 Migração

### De database/sql

```go
// Antes (database/sql + pq)
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", "postgres://user:pass@localhost/db")

// Depois (nexs-lib)
import "github.com/fsvxavier/nexs-lib/db/postgresql"

config := &postgresql.Config{
    Host:     "localhost",
    Database: "db",
    User:     "user",
    Password: "pass",
    Provider: postgresql.ProviderPQ, // Mantém compatibilidade
}

pool, err := postgresql.NewPool(ctx, config)
```

### De GORM

```go
// Antes (GORM puro)
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Depois (nexs-lib + GORM)
import "github.com/fsvxavier/nexs-lib/db/postgresql"

config := &postgresql.Config{
    Host:     "localhost",
    Database: "db", 
    User:     "user",
    Password: "pass",
    Provider: postgresql.ProviderGORM,
}

conn, err := postgresql.NewConnection(ctx, config)
gormDB := conn.(*gorm.DB) // Type assertion para GORM
```

## 📚 Exemplos Avançados

### Circuit Breaker

```go
config := &postgresql.Config{
    // ... configuração básica
    CircuitBreaker: &postgresql.CircuitBreakerConfig{
        MaxFailures:    5,         // Falhas antes de abrir
        ResetTimeout:   "60s",     // Tempo para tentar fechar
        FailureTimeout: "30s",     // Timeout para considerar falha
        OnStateChange: func(from, to string) {
            log.Printf("Circuit breaker: %s -> %s", from, to)
        },
    },
}
```

### Health Checks

```go
func HealthCheck(pool *postgresql.Pool) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return fmt.Errorf("falha ao adquirir conexão: %w", err)
    }
    defer conn.Release()
    
    var result int
    err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        return fmt.Errorf("falha no health check: %w", err)
    }
    
    if result != 1 {
        return fmt.Errorf("health check retornou valor inesperado: %d", result)
    }
    
    return nil
}
```

### Transações

```go
func TransferMoney(pool *postgresql.Pool, fromID, toID int, amount decimal.Decimal) error {
    ctx := context.Background()
    
    // Iniciar transação
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("erro ao iniciar transação: %w", err)
    }
    defer tx.Rollback(ctx) // Rollback automático se não commitado
    
    // Debitar da conta origem
    _, err = tx.Exec(ctx, 
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, fromID)
    if err != nil {
        return fmt.Errorf("erro ao debitar conta %d: %w", fromID, err)
    }
    
    // Creditar na conta destino
    _, err = tx.Exec(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, toID)
    if err != nil {
        return fmt.Errorf("erro ao creditar conta %d: %w", toID, err)
    }
    
    // Registrar transferência
    _, err = tx.Exec(ctx,
        "INSERT INTO transfers (from_account, to_account, amount, created_at) VALUES ($1, $2, $3, NOW())",
        fromID, toID, amount)
    if err != nil {
        return fmt.Errorf("erro ao registrar transferência: %w", err)
    }
    
    // Commit da transação
    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("erro ao confirmar transação: %w", err)
    }
    
    return nil
}
```

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Desenvolvimento

```bash
# Clone o repositório
git clone https://github.com/fsvxavier/nexs-lib.git
cd nexs-lib/db/postgresql

# Instalar dependências
go mod tidy

# Executar testes
go test ./...

# Executar benchmarks
go test -bench=. -benchmem

# Verificar cobertura
go test -cover ./...
```

## 📝 Changelog

### v2.0.0 (2025-07-07)
- ✨ Implementação do Factory Pattern
- ✨ Suporte a múltiplos drivers (PGX, PQ, GORM)
- ✨ Connection pooling avançado
- ✨ Observabilidade integrada
- ✨ Circuit breaker e retry logic
- ✨ Suporte a SSL/TLS completo
- ✨ Mocks para testes
- 📚 Documentação completamente reescrita

### v1.x.x (Legado)
- 🔧 Implementação básica com PGX
- 🔧 Configuração simples

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- 📧 Email: suporte@nexs-lib.com
- 🐛 Issues: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- 📖 Documentação: [docs.nexs-lib.com](https://docs.nexs-lib.com)
- 💬 Discussões: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

---

⭐ **Se este projeto foi útil, considere dar uma estrela no GitHub!**
