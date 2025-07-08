# PostgreSQL Provider - nexs-lib

Este módulo fornece uma implementação robusta e extensível para conexões PostgreSQL, utilizando **padrões de design modernos** e práticas de Clean Architecture.

## 🚀 Características

- **Padrões de Design Aplicados:**
  - Factory Pattern para criação de conexões
  - Strategy Pattern para diferentes providers
  - Dependency Injection e Inversion of Control
- **Suporte a múltiplos drivers:**
  - `pgx` (github.com/jackc/pgx/v5) - Recomendado para performance
  - `pq` (github.com/lib/pq) - Compatibilidade e estabilidade  
  - `gorm` - ORM completo com recursos avançados
- **Interface unificada** para todos os drivers
- **Validação robusta** de configurações
- **Suporte completo a:**
  - Conexão direta e pool de conexões
  - Transações com diferentes níveis de isolamento
  - Operações em lote (batch) otimizadas
  - Multi-tenancy
  - Observabilidade (logs, traces, métricas)
- **Testes abrangentes** com cobertura de **82.9%+**
- **Exemplos práticos** de uso

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/db/postgresql
```

## 🏗️ Arquitetura

```
postgresql/
├── factory.go              # Factory Pattern + Strategy Pattern  
├── strategy_pgx.go         # Estratégia para driver PGX
├── strategy_pq.go          # Estratégia para driver PQ
├── strategy_gorm.go        # Estratégia para driver GORM
├── postgresql.go           # Facade principal (Clean Interface)
├── common/                 # Interfaces e tipos comuns
├── examples/               # Exemplos práticos
└── tests/                  # Testes com 80%+ cobertura
```

## 💡 Uso Básico

### Configuração com Factory Pattern

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
	ctx := context.Background()

	// Configuração usando Options Pattern
	config := postgresql.WithConfig(
		postgresql.WithHost("localhost"),
		postgresql.WithPort(5432),
		postgresql.WithDatabase("mydb"),
		postgresql.WithUser("postgres"),
		postgresql.WithPassword("postgres"),
		postgresql.WithMaxConns(10),
		postgresql.WithMinConns(2),
		postgresql.WithMaxConnLifetime(time.Minute * 30),
		postgresql.WithSSLMode("disable"),
		postgresql.WithTraceEnabled(true),
	)

	// Factory cria conexão com Strategy Pattern
	pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Usar a conexão...
}
```

### Diferentes Providers

```go
// Provider PGX (Recomendado para performance)
pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)

// Provider PQ (Compatibilidade)
pool, err := postgresql.NewPool(ctx, postgresql.PQ, config)

// Provider GORM (ORM completo)
pool, err := postgresql.NewPool(ctx, postgresql.GORM, config)
```

### Operações com Transações

```go
// Adquire conexão do pool
conn, err := pool.Acquire(ctx)
if err != nil {
    log.Fatal(err)
}
defer conn.Close(ctx)

// Inicia transação
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx) // Rollback automático se não commitado

// Executa operações na transação
err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "João")
if err != nil {
    log.Fatal(err)
}

// Commit da transação
err = tx.Commit(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Operações em Lote (Batch)

```go
// Cria batch otimizado
batch, err := postgresql.NewBatch(postgresql.PGX)
if err != nil {
    log.Fatal(err)
}

// Adiciona operações ao batch
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User1")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User2")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User3")

// Executa batch
results, err := conn.SendBatch(ctx, batch)
if err != nil {
    log.Fatal(err)
}
defer results.Close()

// Processa resultados
for i := 0; i < 3; i++ {
    err = results.Exec()
    if err != nil {
        log.Printf("Erro na operação %d: %v", i, err)
    }
}
```

## 🔧 Configuração Avançada

### Customização de Strategy

```go
// Cria factory customizada
factory := postgresql.NewDatabaseFactory()

// Registra strategy personalizada
customStrategy := &MyCustomStrategy{}
factory.RegisterStrategy(postgresql.ProviderType("custom"), customStrategy)

// Usa strategy customizada
conn, err := factory.CreateConnection(ctx, postgresql.ProviderType("custom"), config)
```

### Configuração Completa

```go
config := postgresql.WithConfig(
    // Conexão básica
    postgresql.WithHost("localhost"),
    postgresql.WithPort(5432),
    postgresql.WithDatabase("mydb"),
    postgresql.WithUser("postgres"),
    postgresql.WithPassword("postgres"),
    
    // Pool de conexões
    postgresql.WithMaxConns(50),
    postgresql.WithMinConns(5),
    postgresql.WithMaxConnLifetime(time.Hour),
    postgresql.WithMaxConnIdleTime(time.Minute * 30),
    
    // Segurança
    postgresql.WithSSLMode("require"),
    
    // Observabilidade
    postgresql.WithTraceEnabled(true),
    postgresql.WithQueryLogEnabled(true),
    
    // Multi-tenancy
    postgresql.WithMultiTenantEnabled(true),
)
```

## 🧪 Testes

O módulo possui **82.9%+ de cobertura de testes**, incluindo:

```bash
# Executar todos os testes
go test -v ./...

# Executar com cobertura
go test -cover -v ./...

# Executar benchmarks
go test -bench=. -v ./...

# Executar testes de race condition
go test -race -v ./...
```

## 📚 Exemplos

Veja a pasta `examples/` para exemplos completos:

- `examples/basic/` - Uso básico e operações simples
- `examples/advanced/` - Transações, batch e patterns avançados
- `examples/testing/` - Como testar código usando este módulo

## 🔍 Tratamento de Erros

```go
// Verifica tipos específicos de erro
if postgresql.IsEmptyResultError(err) {
    // Nenhum resultado encontrado
    log.Println("Nenhum registro encontrado")
} else if postgresql.IsDuplicateKeyError(err) {
    // Violação de chave única
    log.Println("Registro duplicado")
} else {
    // Outros erros
    log.Printf("Erro: %v", err)
}
```

## 📊 Monitoramento

```go
// Estatísticas do pool
stats := pool.Stats()
fmt.Printf("Conexões: Total=%d, Ativo=%d, Idle=%d\n", 
    stats.TotalConns, stats.AcquiredConns, stats.IdleConns)

// Health check
if err := pool.Ping(ctx); err != nil {
    log.Printf("Database não está respondendo: %v", err)
}
```

## 🎯 Benefícios dos Padrões Aplicados

### Factory Pattern
- **Flexibilidade**: Fácil adição de novos providers
- **Desacoplamento**: Código cliente não depende de implementações específicas
- **Testabilidade**: Fácil mock e substituição para testes

### Strategy Pattern  
- **Extensibilidade**: Novos providers sem modificar código existente
- **Manutenibilidade**: Cada provider isolado em sua própria estratégia
- **Reutilização**: Strategies podem ser compostas e reutilizadas

### Dependency Injection
- **Testabilidade**: Fácil injeção de mocks e fakes
- **Flexibilidade**: Configuração dinâmica em runtime
- **Desacoplamento**: Baixa dependência entre componentes

## 🔄 Migração de Código Existente

```go
// Código antigo (ainda funciona)
pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)

// Código novo (com Factory explícita)
factory := postgresql.GetFactory()
pool, err := factory.CreatePool(ctx, postgresql.PGX, config)
```

## 🤝 Contribuindo

1. Implementar novos providers criando uma `Strategy`
2. Adicionar testes com cobertura mínima de 80%
3. Documentar com exemplos práticos
4. Seguir princípios SOLID e Clean Architecture

## 📄 Licença

Este projeto está sob a licença MIT.
