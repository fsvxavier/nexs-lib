package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
	// Create a new PGX provider
	provider := pgx.NewProvider()
	defer provider.Close()

	// Create configuration
	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("nexs_lib_example"),
		config.WithUsername("postgres"),
		config.WithPassword("password"),
		config.WithMaxConns(20),
		config.WithMinConns(2),
		config.WithMaxConnLifetime(30*time.Minute),
		config.WithMaxConnIdleTime(5*time.Minute),
		config.WithApplicationName("nexs-lib-example"),
		config.WithQueryTimeout(30*time.Second),
		config.WithConnectTimeout(10*time.Second),
	)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	ctx := context.Background()

	// Create a connection pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Test pool health
	if err := pool.Ping(ctx); err != nil {
		log.Printf("Pool ping failed (this is expected if no database is running): %v", err)
		// Continue with demonstration even if database is not available
	} else {
		log.Println("Pool is healthy!")
	}

	// Demonstrate pool statistics
	stats := pool.Stats()
	fmt.Printf("Pool Statistics:\n")
	fmt.Printf("  Max Connections: %d\n", stats.MaxConns)
	fmt.Printf("  Total Connections: %d\n", stats.TotalConns)
	fmt.Printf("  Idle Connections: %d\n", stats.IdleConns)
	fmt.Printf("  Acquired Connections: %d\n", stats.AcquiredConns)
	fmt.Printf("  Acquire Count: %d\n", stats.AcquireCount)

	// Demonstrate connection acquisition (will fail without database, but shows the API)
	fmt.Println("\nAttempting to acquire connection...")
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to acquire connection (expected without database): %v", err)
	} else {
		defer conn.Release(ctx)
		log.Println("Connection acquired successfully!")

		// Demonstrate basic operations
		demonstrateBasicOperations(ctx, conn)
	}

	// Demonstrate provider metrics
	fmt.Println("\nProvider Metrics:")
	metrics := provider.GetMetrics(ctx)
	for key, value := range metrics {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Demonstrate single connection (not from pool)
	fmt.Println("\nCreating single connection...")
	singleConn, err := provider.CreateConnection(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create single connection (expected without database): %v", err)
	} else {
		defer singleConn.Release(ctx)
		log.Println("Single connection created successfully!")
	}

	fmt.Println("\nExample completed successfully!")
}

func demonstrateBasicOperations(ctx context.Context, conn postgresql.IConn) {
	fmt.Println("\n--- Demonstrating Basic Operations ---")

	// Create a sample table
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	fmt.Println("Creating table...")
	if err := conn.Exec(ctx, createTableQuery); err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}
	fmt.Println("Table created successfully!")

	// Insert sample data
	insertQuery := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`

	fmt.Println("Inserting sample users...")
	users := []struct {
		name  string
		email string
	}{
		{"Alice Johnson", "alice@example.com"},
		{"Bob Smith", "bob@example.com"},
		{"Charlie Brown", "charlie@example.com"},
	}

	for _, user := range users {
		var userID int
		err := conn.QueryOne(ctx, &userID, insertQuery, user.name, user.email)
		if err != nil {
			log.Printf("Failed to insert user %s: %v", user.name, err)
			continue
		}
		fmt.Printf("  Inserted user: %s (ID: %d)\n", user.name, userID)
	}

	// Count total users
	var totalUsers int
	err := conn.QueryOne(ctx, &totalUsers, "SELECT COUNT(*) FROM users")
	if err != nil {
		log.Printf("Failed to count users: %v", err)
		return
	}
	fmt.Printf("Total users in database: %d\n", totalUsers)

	// Query all users
	fmt.Println("Querying all users...")
	rows, err := conn.Query(ctx, "SELECT id, name, email FROM users ORDER BY id")
	if err != nil {
		log.Printf("Failed to query users: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		fmt.Printf("  User %d: %s (%s)\n", id, name, email)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	// Demonstrate transaction
	fmt.Println("Demonstrating transaction...")
	demonstrateTransaction(ctx, conn)

	// Demonstrate batch operations
	fmt.Println("Demonstrating batch operations...")
	demonstrateBatch(ctx, conn)

	// Cleanup
	fmt.Println("Cleaning up...")
	if err := conn.Exec(ctx, "DROP TABLE IF EXISTS users"); err != nil {
		log.Printf("Failed to drop table: %v", err)
	} else {
		fmt.Println("Table dropped successfully!")
	}
}

func demonstrateTransaction(ctx context.Context, conn postgresql.IConn) {
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// Insert a user in transaction
	insertQuery := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
	var userID int
	err = tx.QueryOne(ctx, &userID, insertQuery, "Transaction User", "tx@example.com")
	if err != nil {
		log.Printf("Failed to insert user in transaction: %v", err)
		tx.Rollback(ctx)
		return
	}

	fmt.Printf("  Inserted user in transaction (ID: %d)\n", userID)

	// Create a savepoint
	if err := tx.Savepoint(ctx, "user_inserted"); err != nil {
		log.Printf("Failed to create savepoint: %v", err)
		tx.Rollback(ctx)
		return
	}

	// Try to insert duplicate (should fail)
	err = tx.QueryOne(ctx, &userID, insertQuery, "Duplicate User", "tx@example.com")
	if err != nil {
		fmt.Printf("  Expected error for duplicate email: %v\n", err)
		// Rollback to savepoint
		if err := tx.RollbackToSavepoint(ctx, "user_inserted"); err != nil {
			log.Printf("Failed to rollback to savepoint: %v", err)
			tx.Rollback(ctx)
			return
		}
		fmt.Println("  Rolled back to savepoint")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}
	fmt.Println("  Transaction committed successfully!")
}

func demonstrateBatch(ctx context.Context, conn postgresql.IConn) {
	// Create a batch
	batch := pgx.NewBatch()

	// Add multiple queries to batch
	batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "Batch User 1", "batch1@example.com")
	batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "Batch User 2", "batch2@example.com")
	batch.Queue("SELECT COUNT(*) FROM users")

	fmt.Printf("  Batch has %d queued operations\n", batch.Len())

	// Send batch
	results, err := conn.SendBatch(ctx, batch)
	if err != nil {
		log.Printf("Failed to send batch: %v", err)
		return
	}
	defer results.Close()

	// Process first insert
	if err := results.Exec(); err != nil {
		log.Printf("Failed to execute first batch operation: %v", err)
		return
	}
	fmt.Println("  First batch insert executed")

	// Process second insert
	if err := results.Exec(); err != nil {
		log.Printf("Failed to execute second batch operation: %v", err)
		return
	}
	fmt.Println("  Second batch insert executed")

	// Process count query
	var count int
	if err := results.QueryOne(&count); err != nil {
		log.Printf("Failed to execute batch count query: %v", err)
		return
	}
	fmt.Printf("  Batch count result: %d users\n", count)

	fmt.Println("  Batch operations completed successfully!")
}
