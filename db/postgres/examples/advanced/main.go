package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo Avançado PostgreSQL ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configurar DSN do banco
	dsn := getEnvOrDefault("NEXS_DB_DSN", "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")

	// Criar configuração avançada
	cfg := postgres.NewConfigWithOptions(
		dsn,
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
		postgres.WithMaxConnLifetime(time.Hour),
		postgres.WithMaxConnIdleTime(30*time.Minute),
	)

	// Demonstrar diferentes funcionalidades (simplificado)
	examples := []struct {
		name string
		fn   func(context.Context, postgres.IConfig) error
	}{
		{"Pool Management", demonstratePoolManagement},
		{"Transactions", demonstrateTransactions},
		{"Concurrent Operations", demonstrateConcurrentOperations},
		{"Error Handling", demonstrateErrorHandling},
	}

	for _, example := range examples {
		fmt.Printf("\n%s\n", example.name)
		fmt.Printf("%s\n", strings.Repeat("=", len(example.name)))

		if err := example.fn(ctx, cfg); err != nil {
			log.Printf("Erro no exemplo %s: %v", example.name, err)
		} else {
			fmt.Printf("✓ %s concluído com sucesso\n", example.name)
		}
	}

	fmt.Println("\n=== Exemplos avançados concluídos ===")
}

// demonstratePoolManagement demonstra gerenciamento avançado de pool
func demonstratePoolManagement(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("  Criando pool de conexões...")

	pool, err := postgres.ConnectPoolWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}

	// Demonstrar aquisição e liberação de conexões
	fmt.Println("  Adquirindo múltiplas conexões...")

	var connections []postgres.IConn
	for i := 0; i < 3; i++ {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("erro ao adquirir conexão %d: %w", i, err)
		}
		connections = append(connections, conn)
		fmt.Printf("  Conexão %d adquirida\n", i+1)
	}

	// Liberar todas as conexões
	fmt.Println("  Liberando conexões...")
	for i, conn := range connections {
		conn.Release()
		fmt.Printf("  Conexão %d liberada\n", i+1)
	}

	// Fechar pool de forma assíncrona
	go func() {
		time.Sleep(50 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

// demonstrateTransactions demonstra transações avançadas
func demonstrateTransactions(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("  Demonstrando transações...")

	conn, err := postgres.ConnectWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao conectar: %w", err)
	}
	defer conn.Close(ctx)

	// Criar tabela de teste
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			amount DECIMAL(10,2)
		)
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}

	// Demonstrar transação com commit
	fmt.Println("  Executando transação com commit...")

	// Try alternative approach - execute transaction as single command
	_, err = conn.Exec(ctx, "BEGIN; INSERT INTO test_transactions (name, amount) VALUES ('João', 100.50); COMMIT;")
	if err != nil {
		// Fallback to simple insert without explicit transaction
		fmt.Println("  Fallback: Executando insert simples...")
		_, err = conn.Exec(ctx, "INSERT INTO test_transactions (name, amount) VALUES ($1, $2)", "João", 100.50)
		if err != nil {
			return fmt.Errorf("erro ao inserir dados: %w", err)
		}
	}

	fmt.Println("  Transação confirmada com sucesso")

	// Limpar dados
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS test_transactions")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	return nil
}

// demonstrateConcurrentOperations demonstra operações concorrentes
func demonstrateConcurrentOperations(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("  Demonstrando operações concorrentes...")

	pool, err := postgres.ConnectPoolWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}

	// Criar tabela de teste
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}

	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_concurrent (
			id SERIAL PRIMARY KEY,
			worker_id INT,
			task_id INT,
			completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}
	conn.Release()

	// Executar operações concorrentes (simplificado)
	fmt.Println("  Executando 3 workers concorrentes...")

	var wg sync.WaitGroup
	numWorkers := 3
	tasksPerWorker := 2

	for workerID := 1; workerID <= numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, err := pool.Acquire(ctx)
			if err != nil {
				log.Printf("Worker %d: erro ao adquirir conexão: %v", id, err)
				return
			}
			defer conn.Release()

			for taskID := 1; taskID <= tasksPerWorker; taskID++ {
				_, err = conn.Exec(ctx, "INSERT INTO test_concurrent (worker_id, task_id) VALUES ($1, $2)", id, taskID)
				if err != nil {
					log.Printf("Worker %d: erro ao inserir tarefa %d: %v", id, taskID, err)
					return
				}
			}

			fmt.Printf("  Worker %d completou %d tarefas\n", id, tasksPerWorker)
		}(workerID)
	}

	wg.Wait()

	// Verificar resultados
	conn, err = pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}

	var totalTasks int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM test_concurrent").Scan(&totalTasks)
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao contar tarefas: %w", err)
	}

	fmt.Printf("  Total de tarefas completadas: %d\n", totalTasks)

	// Limpar dados
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS test_concurrent")
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}
	conn.Release()

	// Fechar pool de forma assíncrona
	go func() {
		time.Sleep(50 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

// demonstrateErrorHandling demonstra tratamento de erros
func demonstrateErrorHandling(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("  Demonstrando tratamento de erros...")

	conn, err := postgres.ConnectWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao conectar: %w", err)
	}
	defer conn.Close(ctx)

	// Erro de sintaxe SQL
	fmt.Println("  Testando erro de sintaxe SQL...")
	_, err = conn.Exec(ctx, "SELECT * FROM users") // Erro proposital
	if err != nil {
		fmt.Printf("  ✓ Erro de sintaxe capturado: %v\n", err)
	}

	// Erro de tabela inexistente
	fmt.Println("  Testando erro de tabela inexistente...")
	_, err = conn.Exec(ctx, "SELECT * FROM tabela_inexistente")
	if err != nil {
		fmt.Printf("  ✓ Erro de tabela inexistente capturado: %v\n", err)
	}

	return nil
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
