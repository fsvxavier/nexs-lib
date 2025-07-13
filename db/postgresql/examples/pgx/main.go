package main

import (
	"context"
	"log"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
	// Configuration specifically for PGX driver
	cfg := config.NewConfig(
		config.WithDriver(interfaces.DriverPGX),
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("testdb"),
		config.WithUser("postgres"),
		config.WithPassword("password"),
		config.WithSSLMode("disable"),
		config.WithMaxOpenConns(25),
		config.WithMaxIdleConns(5),
	)

	// Create PGX provider directly
	provider, err := pgx.NewProvider(cfg)
	if err != nil {
		log.Fatal("Failed to create provider:", err)
	}

	// Connect to database
	err = provider.Connect()
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer provider.Close()

	ctx := context.Background()

	log.Println("=== PGX Driver Specific Features ===")

	// Get a connection
	pool := provider.Pool()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal("Failed to acquire connection:", err)
	}
	defer conn.Release(ctx)

	// PGX specific batch operations
	log.Println("\n=== PGX Batch Operations ===")

	// Create table for demo
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS pgx_users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL
	)`

	err = conn.Exec(ctx, createTableSQL)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}

	// Create a new batch using the PGX-specific function
	batch := pgx.NewBatch()

	// Queue multiple operations
	batch.Queue("INSERT INTO pgx_users (name, email) VALUES ($1, $2)", "John Doe", "john@pgx.com")
	batch.Queue("INSERT INTO pgx_users (name, email) VALUES ($1, $2)", "Jane Smith", "jane@pgx.com")
	batch.Queue("INSERT INTO pgx_users (name, email) VALUES ($1, $2)", "Bob Wilson", "bob@pgx.com")

	log.Printf("Batch contains %d operations", batch.Len())

	// Send the batch
	results, err := conn.SendBatch(ctx, batch)
	if err != nil {
		log.Printf("Failed to send batch: %v", err)
		return
	}
	defer results.Close()

	// Process batch results
	log.Println("Processing batch results...")
	for i := 0; i < 3; i++ {
		err = results.Exec()
		if err != nil {
			log.Printf("Batch operation %d failed: %v", i+1, err)
		} else {
			log.Printf("Batch operation %d completed successfully", i+1)
		}
	}

	// Query the inserted data
	log.Println("\n=== Querying Results ===")

	type User struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	var users []User
	err = conn.QueryAll(ctx, &users, "SELECT id, name, email FROM pgx_users ORDER BY id")
	if err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		log.Printf("Found %d users:", len(users))
		for _, user := range users {
			log.Printf("  - ID: %d, Name: %s, Email: %s", user.ID, user.Name, user.Email)
		}
	}

	// Test advanced PGX features
	log.Println("\n=== Advanced PGX Features ===")

	// Transaction with options
	txOpts := interfaces.TxOptions{
		IsolationLevel: "READ_COMMITTED",
	}

	tx, err := conn.BeginTransactionWithOptions(ctx, txOpts)
	if err != nil {
		log.Printf("Failed to begin transaction with options: %v", err)
	} else {
		// Update a user in transaction
		err = tx.Exec(ctx, "UPDATE pgx_users SET name = $1 WHERE email = $2", "John Updated", "john@pgx.com")
		if err != nil {
			log.Printf("Failed to update user: %v", err)
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				log.Printf("Failed to commit transaction: %v", err)
			} else {
				log.Println("Transaction with options completed successfully")
			}
		}
	}

	// Count records
	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM pgx_users")
	if err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		log.Printf("Total users in database: %d", *count)
	}

	// Pool statistics
	stats := pool.Stats()
	log.Printf("\nPool Statistics:")
	log.Printf("  - Max Connections: %d", stats.MaxConns)
	log.Printf("  - Total Connections: %d", stats.TotalConns)
	log.Printf("  - Idle Connections: %d", stats.IdleConns)
	log.Printf("  - Acquire Count: %d", stats.AcquireCount)

	log.Println("\n=== PGX Example completed ===")
}
