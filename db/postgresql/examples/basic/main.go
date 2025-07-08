package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// User representa um usuário para os exemplos
type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

func main() {
	// Configuração do banco de dados
	config := postgresql.WithConfig(
		postgresql.WithHost(getEnv("DB_HOST", "localhost")),
		postgresql.WithPort(getEnvInt("DB_PORT", 5432)),
		postgresql.WithDatabase(getEnv("DB_NAME", "testdb")),
		postgresql.WithUser(getEnv("DB_USER", "postgres")),
		postgresql.WithPassword(getEnv("DB_PASSWORD", "postgres")),
		postgresql.WithSSLMode(getEnv("DB_SSLMODE", "disable")),
		postgresql.WithMaxConns(10),
		postgresql.WithMinConns(2),
		postgresql.WithMaxConnLifetime(time.Minute*30),
		postgresql.WithMaxConnIdleTime(time.Minute*5),
		postgresql.WithTraceEnabled(true),
		postgresql.WithQueryLogEnabled(true),
	)

	ctx := context.Background()

	// Exemplo 1: Conexão direta
	fmt.Println("=== Exemplo 1: Conexão Direta ===")
	if err := exemploConexaoDireta(ctx, config); err != nil {
		log.Printf("Erro no exemplo 1: %v", err)
	}

	// Exemplo 2: Pool de conexões
	fmt.Println("\n=== Exemplo 2: Pool de Conexões ===")
	if err := exemploPoolConexoes(ctx, config); err != nil {
		log.Printf("Erro no exemplo 2: %v", err)
	}

	// Exemplo 3: Diferentes providers
	fmt.Println("\n=== Exemplo 3: Diferentes Providers ===")
	if err := exemploProviders(ctx, config); err != nil {
		log.Printf("Erro no exemplo 3: %v", err)
	}

	// Exemplo 4: Operações batch
	fmt.Println("\n=== Exemplo 4: Operações Batch ===")
	if err := exemploOperacoesBatch(ctx, config); err != nil {
		log.Printf("Erro no exemplo 4: %v", err)
	}
}

func exemploConexaoDireta(ctx context.Context, config *common.Config) error {
	// Cria conexão direta usando PGX
	conn, err := postgresql.NewConnection(ctx, postgresql.PGX, config)
	if err != nil {
		return fmt.Errorf("falha ao criar conexão: %w", err)
	}
	defer conn.Close(ctx)

	// Testa conexão
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("falha ao pingar o banco: %w", err)
	}

	fmt.Println("✓ Conexão direta estabelecida com sucesso")

	// Cria tabela de exemplo
	if err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users_basic (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("falha ao criar tabela: %w", err)
	}

	// Insere dados
	if err := conn.Exec(ctx,
		"INSERT INTO users_basic (name, email, age) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
		"João Silva", "joao@example.com", 30); err != nil {
		return fmt.Errorf("falha ao inserir dados: %w", err)
	}

	// Consulta dados
	var user User
	if err := conn.QueryOne(ctx, &user,
		"SELECT id, name, email, age FROM users_basic WHERE email = $1",
		"joao@example.com"); err != nil {
		return fmt.Errorf("falha ao consultar dados: %w", err)
	}

	fmt.Printf("✓ Usuário encontrado: %+v\n", user)

	return nil
}

func exemploPoolConexoes(ctx context.Context, config *common.Config) error {
	// Cria pool de conexões
	pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)
	if err != nil {
		return fmt.Errorf("falha ao criar pool: %w", err)
	}
	defer pool.Close()

	// Testa pool
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("falha ao pingar o pool: %w", err)
	}

	fmt.Println("✓ Pool de conexões criado com sucesso")

	// Obtém estatísticas do pool
	stats := pool.Stats()
	fmt.Printf("✓ Estatísticas do pool: Total=%d, Ativo=%d, Idle=%d\n",
		stats.TotalConns, stats.AcquiredConns, stats.IdleConns)

	// Adquire conexão do pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("falha ao adquirir conexão: %w", err)
	}
	defer conn.Close(ctx)

	// Cria tabela de exemplo
	if err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users_pool (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("falha ao criar tabela: %w", err)
	}

	// Insere múltiplos registros
	users := []User{
		{Name: "Maria Santos", Email: "maria@example.com", Age: 25},
		{Name: "Pedro Costa", Email: "pedro@example.com", Age: 35},
		{Name: "Ana Oliveira", Email: "ana@example.com", Age: 28},
	}

	for _, user := range users {
		if err := conn.Exec(ctx,
			"INSERT INTO users_pool (name, email, age) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
			user.Name, user.Email, user.Age); err != nil {
			return fmt.Errorf("falha ao inserir usuário %s: %w", user.Name, err)
		}
	}

	// Consulta todos os usuários
	var allUsers []User
	if err := conn.QueryAll(ctx, &allUsers,
		"SELECT id, name, email, age FROM users_pool ORDER BY id"); err != nil {
		return fmt.Errorf("falha ao consultar usuários: %w", err)
	}

	fmt.Printf("✓ Encontrados %d usuários no pool\n", len(allUsers))
	for _, user := range allUsers {
		fmt.Printf("  - %s (%s), %d anos\n", user.Name, user.Email, user.Age)
	}

	return nil
}

func exemploProviders(ctx context.Context, config *common.Config) error {
	providers := []postgresql.ProviderType{
		postgresql.PGX,
		postgresql.PQ,
		postgresql.GORM,
	}

	for _, provider := range providers {
		fmt.Printf("🔧 Testando provider: %s\n", provider)

		// Testa criação de batch
		batch, err := postgresql.NewBatch(provider)
		if err != nil {
			log.Printf("❌ Erro ao criar batch para %s: %v", provider, err)
			continue
		}

		// Adiciona operações ao batch
		batch.Queue("SELECT 1")
		batch.Queue("SELECT $1", 42)
		batch.Queue("SELECT $1, $2", "test", 123)

		fmt.Printf("✓ Batch criado para %s\n", provider)

		// Tenta criar conexão (pode falhar sem DB real)
		conn, err := postgresql.NewConnection(ctx, provider, config)
		if err != nil {
			log.Printf("⚠️  Conexão para %s falhou (esperado sem DB): %v", provider, err)
			continue
		}
		defer conn.Close(ctx)

		if err := conn.Ping(ctx); err != nil {
			log.Printf("⚠️  Ping para %s falhou: %v", provider, err)
			continue
		}

		fmt.Printf("✓ Conexão bem-sucedida para %s\n", provider)
	}

	return nil
}

func exemploOperacoesBatch(ctx context.Context, config *common.Config) error {
	// Cria conexão
	conn, err := postgresql.NewConnection(ctx, postgresql.PGX, config)
	if err != nil {
		return fmt.Errorf("falha ao criar conexão: %w", err)
	}
	defer conn.Close(ctx)

	// Cria tabela para batch
	if err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users_batch (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("falha ao criar tabela: %w", err)
	}

	// Cria batch
	batch, err := postgresql.NewBatch(postgresql.PGX)
	if err != nil {
		return fmt.Errorf("falha ao criar batch: %w", err)
	}

	// Adiciona operações ao batch
	batch.Queue("INSERT INTO users_batch (name, email, age) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
		"Carlos Silva", "carlos@example.com", 40)
	batch.Queue("INSERT INTO users_batch (name, email, age) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
		"Lucia Santos", "lucia@example.com", 32)
	batch.Queue("INSERT INTO users_batch (name, email, age) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
		"Roberto Costa", "roberto@example.com", 45)

	// Executa batch
	batchResults, err := conn.SendBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("falha ao executar batch: %w", err)
	}
	defer batchResults.Close()

	// Processa resultados do batch
	for i := 0; i < 3; i++ {
		if err := batchResults.Exec(); err != nil {
			log.Printf("⚠️  Erro na operação %d do batch: %v", i+1, err)
		} else {
			fmt.Printf("✓ Operação %d do batch executada com sucesso\n", i+1)
		}
	}

	// Conta total de registros
	var count int
	if countPtr, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users_batch"); err != nil {
		return fmt.Errorf("falha ao contar registros: %w", err)
	} else {
		count = *countPtr
	}

	fmt.Printf("✓ Total de registros na tabela batch: %d\n", count)

	return nil
}

// Funções auxiliares para variáveis de ambiente
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
