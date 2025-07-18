package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	// Basic PostgreSQL provider example
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Create basic configuration
	cfg := postgres.NewDefaultConfig("postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")

	// 2. Example: Basic connection
	fmt.Println("Starting basic connection example...")
	if err := basicConnectionExample(ctx, cfg); err != nil {
		log.Printf("Basic connection example failed: %v", err)
	}

	// 3. Example: Pool connection (simplified)
	fmt.Println("Starting pool connection example...")
	if err := poolConnectionExample(ctx, cfg); err != nil {
		log.Printf("Pool connection example failed: %v", err)
	}

	// 4. Example: Simple query
	fmt.Println("Starting simple query example...")
	if err := simpleQueryExample(ctx, cfg); err != nil {
		log.Printf("Simple query example failed: %v", err)
	}

	// 5. Example: Basic transaction
	fmt.Println("Starting basic transaction example...")
	if err := basicTransactionExample(ctx, cfg); err != nil {
		log.Printf("Basic transaction example failed: %v", err)
	}

	fmt.Println("Basic examples completed!")
}

func basicConnectionExample(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("=== Basic Connection Example ===")

	// Create single connection
	conn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
	if err != nil {
		fmt.Printf("Note: Connection creation would require actual database: %v\n", err)
		return nil // Don't fail the example
	}
	defer conn.Close(ctx)

	// Test connection
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✓ Connection established successfully!")
	return nil
}

func poolConnectionExample(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("=== Pool Connection Example ===")

	// Create connection pool
	pool, err := postgres.ConnectPool(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
	if err != nil {
		fmt.Printf("Note: Pool creation would require actual database: %v\n", err)
		return nil // Don't fail the example
	}

	// Acquire connection from pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}

	// Test connection
	if err := conn.Ping(ctx); err != nil {
		conn.Release()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✓ Pool connection established successfully!")

	// Release connection back to pool
	conn.Release()

	// Close pool in a separate goroutine to avoid blocking
	go func() {
		time.Sleep(100 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

func simpleQueryExample(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("=== Simple Query Example ===")

	// Create connection
	conn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
	if err != nil {
		fmt.Printf("Note: Connection creation would require actual database: %v\n", err)
		return nil // Don't fail the example
	}
	defer conn.Close(ctx)

	// Example query
	query := "SELECT version()"

	// QueryRow example
	var version string
	err = conn.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to query database version: %w", err)
	}

	fmt.Printf("✓ Database version: %s\n", version)
	return nil
}

func basicTransactionExample(ctx context.Context, cfg postgres.IConfig) error {
	fmt.Println("=== Basic Transaction Example ===")

	// Create connection
	conn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
	if err != nil {
		fmt.Printf("Note: Connection creation would require actual database: %v\n", err)
		return nil // Don't fail the example
	}
	defer conn.Close(ctx)

	// Try alternative approach - execute transaction as single command
	fmt.Println("  Executando transação como comando único...")
	_, err = conn.Exec(ctx, "BEGIN; SELECT 1; COMMIT;")
	if err != nil {
		fmt.Printf("  Aviso: Transação não disponível nesta versão: %v\n", err)
		// Try simple autocommit instead
		fmt.Println("  Testando execução simples...")
		_, err = conn.Exec(ctx, "SELECT 1")
		if err != nil {
			return fmt.Errorf("failed to execute simple query: %w", err)
		}
	}

	fmt.Println("✓ Transaction (or simple query) completed successfully!")
	return nil
}
