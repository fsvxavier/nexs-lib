# PostgreSQL Provider

Este pacote fornece uma implementação robusta e extensível para conexões PostgreSQL, com suporte a dois drivers populares: 
- `pgx` (github.com/jackc/pgx/v5)
- `pq` (github.com/lib/pq)

## Características

- Interface unificada para ambos os drivers
- Suporte a conexão direta e pool de conexões
- Tratamento adequado de transações
- Operações em lote (batch)
- Configuração flexível com padrão de options
- Suporte a multi-tenancy
- Gerenciamento de erros específicos do PostgreSQL
- Testes unitários e de integração

## Uso Básico

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
	ctx := context.Background()

	// Criar configuração
	config := postgresql.WithConfig(
		postgresql.WithHost("localhost"),
		postgresql.WithPort(5432),
		postgresql.WithDatabase("mydb"),
		postgresql.WithUser("postgres"),
		postgresql.WithPassword("postgres"),
		postgresql.WithMaxConns(10),
		postgresql.WithMinConns(2),
		postgresql.WithMaxConnLifetime(time.Minute * 30),
		postgresql.WithMaxConnIdleTime(time.Minute * 10),
		postgresql.WithSSLMode("disable"),
	)

	// Criar pool de conexões usando pgx
	pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)
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
