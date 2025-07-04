# Provider GORM para PostgreSQL

Este provider permite o uso do GORM (ORM para Go) com o PostgreSQL.

## Características

- Implementa as interfaces comuns do provider PostgreSQL
- Suporta todas as funcionalidades do GORM
- Integra-se com o sistema de pool de conexões e transações
- Compatível com multi-tenancy

## Como usar

### Inicialização básica

```go
package main

import (
	"context"
	"log"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

func main() {
	// Configuração
	config := common.DefaultConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.Database = "meu_banco"
	config.User = "postgres"
	config.Password = "senha"

	ctx := context.Background()

	// Criar uma conexão com o GORM
	conn, err := postgresql.NewConnection(ctx, postgresql.GORM, config)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close(ctx)

	// Uso básico de SQL
	type Usuario struct {
		ID   int
		Nome string
	}

	// Consulta usando a interface comum
	var usuarios []Usuario
	err = conn.QueryAll(ctx, &usuarios, "SELECT id, nome FROM usuarios ORDER BY nome")
	if err != nil {
		log.Fatalf("Erro na consulta: %v", err)
	}

	for _, u := range usuarios {
		log.Printf("ID: %d, Nome: %s", u.ID, u.Nome)
	}
}
```

### Uso avançado com modelos GORM

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/gorm"
	"gorm.io/gorm"
)

// Produto é um modelo GORM
type Produto struct {
	gorm.Model
	Nome        string
	Preco       float64
	Descricao   string
	CategoriaID uint
	Categoria   Categoria
}

// Categoria é um modelo GORM
type Categoria struct {
	gorm.Model
	Nome string
}

func main() {
	// Configuração
	config := common.DefaultConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.Database = "meu_banco"
	config.User = "postgres"
	config.Password = "senha"

	ctx := context.Background()

	// Criar uma conexão com o GORM
	conn, err := postgresql.NewConnection(ctx, postgresql.GORM, config)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close(ctx)

	// Obter a instância DB do GORM para usar funcionalidades avançadas
	db, err := gorm.GormDB(conn)
	if err != nil {
		log.Fatalf("Erro ao obter DB GORM: %v", err)
	}

	// Auto-migração dos modelos
	err = db.AutoMigrate(&Categoria{}, &Produto{})
	if err != nil {
		log.Fatalf("Erro na migração: %v", err)
	}

	// Criar categorias
	categorias := []Categoria{
		{Nome: "Eletrônicos"},
		{Nome: "Livros"},
	}
	
	for _, c := range categorias {
		if err := db.Create(&c).Error; err != nil {
			log.Fatalf("Erro ao criar categoria: %v", err)
		}
	}

	// Buscar produtos usando GORM
	var produtos []Produto
	err = db.Preload("Categoria").Find(&produtos).Error
	if err != nil {
		log.Fatalf("Erro ao buscar produtos: %v", err)
	}

	for _, p := range produtos {
		log.Printf("Produto: %s, Preço: %.2f, Categoria: %s", 
			p.Nome, p.Preco, p.Categoria.Nome)
	}
}
```

## Transações

```go
// Exemplo de transação
tx, err := conn.BeginTransaction(ctx)
if err != nil {
	log.Fatalf("Erro ao iniciar transação: %v", err)
}

// Em caso de erro, fazer rollback
defer func() {
	if r := recover(); r != nil {
		tx.Rollback(ctx)
		panic(r)
	}
}()

// Operações dentro da transação
err = tx.Exec(ctx, "INSERT INTO produtos (nome, preco) VALUES ($1, $2)", "Novo Produto", 99.99)
if err != nil {
	tx.Rollback(ctx)
	log.Fatalf("Erro ao inserir produto: %v", err)
}

// Confirmar a transação
if err := tx.Commit(ctx); err != nil {
	log.Fatalf("Erro ao fazer commit: %v", err)
}
```
