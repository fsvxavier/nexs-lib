# PostgreSQL Provider

Este pacote fornece uma implementação robusta e extensível para conexões PostgreSQL, utilizando **padrões de design modernos** e práticas de Clean Architecture.

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
	if err != nil {
		log.Fatalf("Erro ao criar pool: %v", err)
	}
	defer pool.Close()

	// Adquirir conexão do pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("Erro ao adquirir conexão: %v", err)
	}
	defer conn.Close(ctx)

	// Executar consulta
	var result string
	err = conn.QueryOne(ctx, &result, "SELECT current_timestamp")
	if err != nil {
		log.Fatalf("Erro na consulta: %v", err)
	}

	fmt.Println("Data/hora atual:", result)

	// Exemplo de transação
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Fatalf("Erro ao iniciar transação: %v", err)
	}

	// Em caso de erro, fazer rollback
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Executar comandos na transação
	err = tx.Exec(ctx, "INSERT INTO logs (message) VALUES ($1)", "Teste de transação")
	if err != nil {
		return
	}

	// Confirmar transação
	err = tx.Commit(ctx)
	if err != nil {
		return
	}

	fmt.Println("Transação concluída com sucesso")
}
```

## Usando Batch (Lote de Comandos)

```go
batch, _ := postgresql.NewBatch(postgresql.PGX)
batch.Queue("INSERT INTO users (name) VALUES ($1)", "Alice")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "Bob")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "Charlie")

results, err := conn.SendBatch(ctx, batch)
if err != nil {
	log.Fatalf("Erro ao enviar lote: %v", err)
}
defer results.Close()

// Executar cada comando no lote
for i := 0; i < 3; i++ {
	err = results.Exec()
	if err != nil {
		log.Fatalf("Erro no comando %d: %v", i+1, err)
	}
}
```

## Tratamento de Erros Específicos

```go
err := conn.QueryOne(ctx, &user, "SELECT * FROM users WHERE id = $1", 1)
if postgresql.IsEmptyResultError(err) {
	fmt.Println("Usuário não encontrado")
	return
}
if err != nil {
	log.Fatalf("Erro na consulta: %v", err)
}
```

## Testes

Para executar os testes unitários:

```
go test ./db/postgresql/...
```

Para testes com verificação de race conditions:

```
go test -race ./db/postgresql/...
```

Para benchmarks:

```
go test -bench=. ./db/postgresql/...
```

## Notas de Implementação

- A implementação pgx oferece melhor desempenho e funcionalidades mais avançadas
- A implementação pq oferece compatibilidade com o pacote database/sql padrão
- Para operações de alta performance, recomenda-se o uso do provider pgx
- O tratamento de multi-tenancy deve ser configurado conforme a necessidade específica do aplicativo
