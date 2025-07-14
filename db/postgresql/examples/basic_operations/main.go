package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
	ctx := context.Background()

	// Criar provedor PGX
	provider := pgx.NewProvider()
	defer func() {
		if err := provider.Close(); err != nil {
			log.Printf("Erro ao fechar provedor: %v", err)
		}
	}()

	// Configurar banco de dados
	cfg := config.NewConfig(
		config.WithHost(getEnv("DB_HOST", "localhost")),
		config.WithPort(getEnvInt("DB_PORT", 5432)),
		config.WithDatabase(getEnv("DB_NAME", "example")),
		config.WithUsername(getEnv("DB_USER", "postgres")),
		config.WithPassword(getEnv("DB_PASSWORD", "password")),
		config.WithMaxConns(10),
		config.WithMinConns(2),
		config.WithConnectTimeout(30*time.Second),
		config.WithQueryTimeout(30*time.Second),
	)

	// Criar pool de conexões
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Fatalf("Erro ao criar pool: %v", err)
	}
	defer pool.Close()

	// Testar conexão
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Erro ao conectar com banco: %v", err)
	}

	fmt.Println("✅ Conectado ao PostgreSQL com sucesso!")

	// Criar tabelas
	if err := createTables(ctx, pool); err != nil {
		log.Fatalf("Erro ao criar tabelas: %v", err)
	}

	// Demonstrar operações CRUD
	if err := demonstrateCRUD(ctx, pool); err != nil {
		log.Fatalf("Erro nas operações CRUD: %v", err)
	}

	// Demonstrar transações
	if err := demonstrateTransactions(ctx, pool); err != nil {
		log.Fatalf("Erro nas transações: %v", err)
	}

	// Demonstrar operações em lote
	if err := demonstrateBatchOperations(ctx, pool); err != nil {
		log.Fatalf("Erro nas operações em lote: %v", err)
	}

	fmt.Println("🎉 Todos os exemplos executados com sucesso!")
}

// createTables cria as tabelas necessárias para os exemplos
func createTables(ctx context.Context, pool postgresql.IPool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	queries := []string{
		`DROP TABLE IF EXISTS orders CASCADE`,
		`DROP TABLE IF EXISTS products CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
		`
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(150) UNIQUE NOT NULL,
			age INTEGER CHECK (age > 0),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE TABLE products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
			stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE TABLE orders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			total DECIMAL(10,2) NOT NULL CHECK (total >= 0),
			status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_orders_user_id ON orders(user_id);
		CREATE INDEX idx_orders_status ON orders(status);
		`,
	}

	for _, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("erro ao executar query: %w", err)
		}
	}

	fmt.Println("✅ Tabelas criadas com sucesso!")
	return nil
}

// demonstrateCRUD demonstra operações básicas de CRUD
func demonstrateCRUD(ctx context.Context, pool postgresql.IPool) error {
	fmt.Println("\n🔄 Demonstrando operações CRUD...")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	// CREATE - Inserir usuários
	users := []User{
		{Name: "João Silva", Email: "joao@example.com", Age: 30},
		{Name: "Maria Santos", Email: "maria@example.com", Age: 25},
		{Name: "Pedro Costa", Email: "pedro@example.com", Age: 35},
	}

	for i, user := range users {
		var id int
		row := conn.QueryRow(ctx,
			"INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id",
			user.Name, user.Email, user.Age)

		if err := row.Scan(&id); err != nil {
			return fmt.Errorf("erro ao inserir usuário: %w", err)
		}
		users[i].ID = id
		fmt.Printf("   ✅ Usuário inserido: ID=%d, Nome=%s\n", id, user.Name)
	}

	// READ - Consultar usuários
	fmt.Println("\n📖 Consultando usuários...")
	rows, err := conn.Query(ctx, "SELECT id, name, email, age, created_at FROM users ORDER BY id")
	if err != nil {
		return fmt.Errorf("erro ao consultar usuários: %w", err)
	}
	defer rows.Close()

	var retrievedUsers []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt); err != nil {
			return fmt.Errorf("erro ao escanear usuário: %w", err)
		}
		retrievedUsers = append(retrievedUsers, user)
		fmt.Printf("   📋 ID: %d | Nome: %s | Email: %s | Idade: %d\n",
			user.ID, user.Name, user.Email, user.Age)
	}

	// UPDATE - Atualizar usuário
	fmt.Println("\n✏️ Atualizando usuário...")
	if len(retrievedUsers) > 0 {
		userToUpdate := retrievedUsers[0]
		newAge := userToUpdate.Age + 1

		if err := conn.Exec(ctx,
			"UPDATE users SET age = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			newAge, userToUpdate.ID); err != nil {
			return fmt.Errorf("erro ao atualizar usuário: %w", err)
		}
		fmt.Printf("   ✅ Usuário ID=%d atualizado. Nova idade: %d\n", userToUpdate.ID, newAge)
	}

	// DELETE - Remover último usuário
	fmt.Println("\n🗑️ Removendo último usuário...")
	if len(retrievedUsers) > 0 {
		lastUser := retrievedUsers[len(retrievedUsers)-1]

		if err := conn.Exec(ctx, "DELETE FROM users WHERE id = $1", lastUser.ID); err != nil {
			return fmt.Errorf("erro ao remover usuário: %w", err)
		}
		fmt.Printf("   ✅ Usuário ID=%d removido com sucesso\n", lastUser.ID)
	}

	// Verificar contagem final
	var count int
	row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("erro ao contar usuários: %w", err)
	}
	fmt.Printf("   📊 Total de usuários após operações: %d\n", count)

	return nil
}

// demonstrateTransactions demonstra o uso de transações
func demonstrateTransactions(ctx context.Context, pool postgresql.IPool) error {
	fmt.Println("\n💳 Demonstrando transações...")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	// Iniciar transação
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Função para rollback em caso de erro
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Inserir produto
	var productID int
	row := tx.QueryRow(ctx,
		"INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id",
		"Notebook Gamer", "Notebook para jogos de alta performance", 2500.00, 10)

	if err := row.Scan(&productID); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("erro ao inserir produto: %w", err)
	}
	fmt.Printf("   ✅ Produto inserido na transação: ID=%d\n", productID)

	// Buscar um usuário para criar pedido
	var userID int
	userRow := tx.QueryRow(ctx, "SELECT id FROM users LIMIT 1")
	if err := userRow.Scan(&userID); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Criar pedido
	var orderID int
	orderRow := tx.QueryRow(ctx,
		"INSERT INTO orders (user_id, total, status) VALUES ($1, $2, $3) RETURNING id",
		userID, 2500.00, "pending")

	if err := orderRow.Scan(&orderID); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("erro ao inserir pedido: %w", err)
	}
	fmt.Printf("   ✅ Pedido inserido na transação: ID=%d\n", orderID)

	// Atualizar estoque do produto
	if err := tx.Exec(ctx,
		"UPDATE products SET stock = stock - 1 WHERE id = $1",
		productID); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("erro ao atualizar estoque: %w", err)
	}
	fmt.Printf("   ✅ Estoque atualizado na transação\n")

	// Simular condição de erro (comentado para não falhar o exemplo)
	// return fmt.Errorf("erro simulado - transação será revertida")

	// Confirmar transação
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	fmt.Printf("   🎯 Transação confirmada com sucesso!\n")
	return nil
}

// demonstrateBatchOperations demonstra operações em lote
func demonstrateBatchOperations(ctx context.Context, pool postgresql.IPool) error {
	fmt.Println("\n📦 Demonstrando operações em lote...")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	// Criar batch de produtos
	batch := &simpleBatch{}

	products := []Product{
		{Name: "Mouse Gamer", Description: "Mouse com sensor óptico", Price: 150.00, Stock: 50},
		{Name: "Teclado Mecânico", Description: "Teclado com switches azuis", Price: 300.00, Stock: 30},
		{Name: "Monitor 4K", Description: "Monitor 27 polegadas 4K", Price: 800.00, Stock: 15},
		{Name: "Headset Gaming", Description: "Fone com microfone", Price: 200.00, Stock: 25},
		{Name: "Webcam HD", Description: "Câmera para streaming", Price: 120.00, Stock: 40},
	}

	for _, product := range products {
		batch.Queue(
			"INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4)",
			product.Name, product.Description, product.Price, product.Stock)
	}

	fmt.Printf("   📊 Preparando batch com %d produtos...\n", batch.Len())

	// Executar batch
	batchResults, err := conn.SendBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("erro ao executar batch: %w", err)
	}
	defer batchResults.Close()

	// Processar resultados do batch
	for i := 0; i < len(products); i++ {
		if err := batchResults.Exec(); err != nil {
			return fmt.Errorf("erro no resultado do batch %d: %w", i, err)
		}
	}

	fmt.Printf("   ✅ Batch executado com sucesso! %d produtos inseridos\n", len(products))

	// Verificar produtos inseridos
	var productCount int
	row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM products")
	if err := row.Scan(&productCount); err != nil {
		return fmt.Errorf("erro ao contar produtos: %w", err)
	}
	fmt.Printf("   📊 Total de produtos no banco: %d\n", productCount)

	return nil
}

// simpleBatch implementação simples de IBatch para o exemplo
type simpleBatch struct {
	queries []string
	args    [][]interface{}
}

func (b *simpleBatch) Queue(query string, args ...interface{}) {
	b.queries = append(b.queries, query)
	b.args = append(b.args, args)
}

func (b *simpleBatch) Len() int {
	return len(b.queries)
}

func (b *simpleBatch) Clear() {
	b.queries = b.queries[:0]
	b.args = b.args[:0]
}

// Funções utilitárias para variáveis de ambiente
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d"); err == nil && intValue == 1 {
			var result int
			fmt.Sscanf(value, "%d", &result)
			return result
		}
	}
	return defaultValue
}
