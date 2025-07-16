package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func main() {
	// Example of using the PostgreSQL provider in a generic way
	ctx := context.Background()

	// Create configuration using builder pattern
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	// Cast to specific config type to access options
	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		// Apply configuration options
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(20),
			postgresql.WithMinConns(5),
			postgresql.WithMaxConnLifetime(30*time.Minute),
			postgresql.WithMultiTenant(true),
			postgresql.WithReadReplicas(
				[]string{
					"postgres://user:password@replica1:5432/testdb",
					"postgres://user:password@replica2:5432/testdb",
				},
				interfaces.LoadBalanceModeRoundRobin,
			),
			postgresql.WithFailover(
				[]string{
					"postgres://user:password@fallback1:5432/testdb",
					"postgres://user:password@fallback2:5432/testdb",
				},
				3,
			),
		)
		if err != nil {
			log.Fatalf("Failed to apply configuration: %v", err)
		}
	}

	// Create provider using factory
	provider, err := postgresql.NewPGXProvider()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Provider: %s v%s\n", provider.Name(), provider.Version())
	fmt.Printf("Supported features: %v\n", provider.GetSupportedFeatures())

	// Create connection pool
	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Test basic operations
	if err := basicOperations(ctx, pool); err != nil {
		log.Fatalf("Basic operations failed: %v", err)
	}

	// Test transaction operations
	if err := transactionOperations(ctx, pool); err != nil {
		log.Fatalf("Transaction operations failed: %v", err)
	}

	// Test batch operations
	if err := batchOperations(ctx, pool); err != nil {
		log.Fatalf("Batch operations failed: %v", err)
	}

	// Show pool statistics
	stats := pool.Stats()
	fmt.Printf("Pool Stats: Acquired=%d, Total=%d, Idle=%d\n",
		stats.AcquiredConns, stats.TotalConns, stats.IdleConns)

	fmt.Println("Example completed successfully!")
}

func basicOperations(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n--- Basic Operations ---")

	// Acquire connection from pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Ping test
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	fmt.Println("✓ Ping successful")

	// Create test table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	fmt.Println("✓ Table created")

	// Insert data
	result, err := conn.Exec(ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"John Doe", "john@example.com")
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}
	fmt.Printf("✓ Inserted %d row(s)\n", result.RowsAffected())

	// Query single row
	var name, email string
	err = conn.QueryOne(ctx, &struct {
		Name  string `db:"name"`
		Email string `db:"email"`
	}{}, "SELECT name, email FROM users WHERE email = $1", "john@example.com")
	if err != nil {
		return fmt.Errorf("failed to query single row: %w", err)
	}
	fmt.Printf("✓ Queried user: %s <%s>\n", name, email)

	// Query all rows
	var users []struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}
	err = conn.QueryAll(ctx, &users, "SELECT id, name, email FROM users LIMIT 10")
	if err != nil {
		return fmt.Errorf("failed to query all rows: %w", err)
	}
	fmt.Printf("✓ Queried %d user(s)\n", len(users))

	// Count rows
	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		return fmt.Errorf("failed to count rows: %w", err)
	}
	fmt.Printf("✓ Total users: %d\n", count)

	return nil
}

func transactionOperations(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n--- Transaction Operations ---")

	// Acquire connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert multiple users in transaction
	_, err = tx.Exec(ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"Alice Smith", "alice@example.com")
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to insert in transaction: %w", err)
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"Bob Johnson", "bob@example.com")
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to insert in transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("✓ Transaction completed successfully")
	return nil
}

func batchOperations(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n--- Batch Operations ---")

	// Use pool's AcquireFunc for automatic connection management
	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Create batch
		batch := &struct {
			queries []batchQuery
		}{
			queries: []batchQuery{
				{"INSERT INTO users (name, email) VALUES ($1, $2)", []interface{}{"User1", "user1@example.com"}},
				{"INSERT INTO users (name, email) VALUES ($1, $2)", []interface{}{"User2", "user2@example.com"}},
				{"INSERT INTO users (name, email) VALUES ($1, $2)", []interface{}{"User3", "user3@example.com"}},
			},
		}

		// This is a simplified example - in real implementation, you'd use the actual IBatch interface
		for _, query := range batch.queries {
			_, err := conn.Exec(ctx, query.sql, query.args...)
			if err != nil {
				return fmt.Errorf("failed to execute batch query: %w", err)
			}
		}

		fmt.Println("✓ Batch operations completed")
		return nil
	})
}

type batchQuery struct {
	sql  string
	args []interface{}
}
