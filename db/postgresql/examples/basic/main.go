package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func main() {
	// Basic PostgreSQL provider example
	ctx := context.Background()

	// 1. Create basic configuration
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	// 2. Create provider using factory
	provider, err := postgresql.NewPGXProvider()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// 3. Example: Basic connection
	if err := basicConnectionExample(ctx, provider, cfg); err != nil {
		log.Printf("Basic connection example failed: %v", err)
	}

	// 4. Example: Simple query
	if err := simpleQueryExample(ctx, provider, cfg); err != nil {
		log.Printf("Simple query example failed: %v", err)
	}

	// 5. Example: Basic transaction
	if err := basicTransactionExample(ctx, provider, cfg); err != nil {
		log.Printf("Basic transaction example failed: %v", err)
	}

	fmt.Println("Basic examples completed!")
}

func basicConnectionExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("=== Basic Connection Example ===")

	// Create single connection
	conn, err := provider.NewConn(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Connection creation would require actual database: %v\n", err)
		return nil // Don't fail the example
	}
	defer conn.Close(ctx)

	// Health check
	if err := conn.HealthCheck(ctx); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
	} else {
		fmt.Println("✅ Connection health check passed")
	}

	// Get connection statistics
	stats := conn.Stats()
	fmt.Printf("📊 Connection Stats - Total Queries: %d, Failed Queries: %d\n",
		stats.TotalQueries, stats.FailedQueries)

	return nil
}

func simpleQueryExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Simple Query Example ===")

	// Create connection
	conn, err := provider.NewConn(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Query example would require actual database: %v\n", err)
		return nil
	}
	defer conn.Close(ctx)

	// Simple SELECT query
	row := conn.QueryRow(ctx, "SELECT 1 as number, 'Hello' as greeting")
	var number int
	var greeting string

	if err := row.Scan(&number, &greeting); err != nil {
		fmt.Printf("Query scan failed: %v\n", err)
	} else {
		fmt.Printf("✅ Query result: number=%d, greeting=%s\n", number, greeting)
	}

	// Query with parameters
	row = conn.QueryRow(ctx, "SELECT $1 + $2 as sum", 10, 20)
	var sum int

	if err := row.Scan(&sum); err != nil {
		fmt.Printf("Parameterized query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Parameterized query result: sum=%d\n", sum)
	}

	return nil
}

func basicTransactionExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Basic Transaction Example ===")

	conn, err := provider.NewConn(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Transaction example would require actual database: %v\n", err)
		return nil
	}
	defer conn.Close(ctx)

	// Begin transaction
	tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
		IsoLevel:   interfaces.TxIsoLevelReadCommitted,
		AccessMode: interfaces.TxAccessModeReadWrite,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is handled properly
	defer func() {
		if err != nil {
			fmt.Println("🔄 Rolling back transaction due to error")
			tx.Rollback(ctx)
		} else {
			fmt.Println("✅ Committing transaction")
			tx.Commit(ctx)
		}
	}()

	// Execute operations in transaction
	_, err = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT UNIQUE)")
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	var userID int64
	row := tx.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", "john@example.com")
	err = row.Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	fmt.Printf("✅ Created user with ID: %d\n", userID)
	return nil
}
