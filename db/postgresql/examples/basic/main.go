package main

import (
	"context"
	"log"
	"os"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// User represents a simple user entity
type User struct {
	ID    int    `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

func main() {
	// Create configuration
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

	// Create provider
	provider, err := postgresql.CreateProvider(cfg)
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

	// Example 1: Using Pool to acquire connections
	log.Println("=== Example 1: Pool Operations ===")

	pool := provider.Pool()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to acquire connection: %v", err)
	} else {
		defer conn.Release(ctx)

		// Ping the connection
		if err := conn.Ping(ctx); err != nil {
			log.Printf("Ping failed: %v", err)
		} else {
			log.Println("Connection ping successful")
		}

		// Get pool stats
		stats := pool.Stats()
		log.Printf("Pool stats - Max: %d, Total: %d, Idle: %d",
			stats.MaxConns, stats.TotalConns, stats.IdleConns)
	}

	// Example 2: Basic CRUD operations
	log.Println("\n=== Example 2: CRUD Operations ===")

	// Create table (in real scenario, use migrations)
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL
	)`

	if conn != nil {
		err = conn.Exec(ctx, createTableSQL)
		if err != nil {
			log.Printf("Failed to create table: %v", err)
		} else {
			log.Println("Table created successfully")
		}

		// Insert a user
		insertSQL := "INSERT INTO users (name, email) VALUES ($1, $2)"
		err = conn.Exec(ctx, insertSQL, "John Doe", "john@example.com")
		if err != nil {
			log.Printf("Failed to insert user: %v", err)
		} else {
			log.Println("User inserted successfully")
		}

		// Query single user
		var user User
		selectSQL := "SELECT id, name, email FROM users WHERE email = $1"
		err = conn.QueryOne(ctx, &user, selectSQL, "john@example.com")
		if err != nil {
			log.Printf("Failed to query user: %v", err)
		} else {
			log.Printf("Found user: %+v", user)
		}

		// Query all users
		var users []User
		selectAllSQL := "SELECT id, name, email FROM users"
		err = conn.QueryAll(ctx, &users, selectAllSQL)
		if err != nil {
			log.Printf("Failed to query users: %v", err)
		} else {
			log.Printf("Found %d users", len(users))
		}

		// Count users
		count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users")
		if err != nil {
			log.Printf("Failed to count users: %v", err)
		} else {
			log.Printf("Total users: %d", *count)
		}
	}

	// Example 3: Transaction operations
	log.Println("\n=== Example 3: Transaction Operations ===")

	if conn != nil {
		tx, err := conn.BeginTransaction(ctx)
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
		} else {
			// Insert multiple users in transaction
			insertSQL := "INSERT INTO users (name, email) VALUES ($1, $2)"

			err = tx.Exec(ctx, insertSQL, "Alice Smith", "alice@example.com")
			if err != nil {
				log.Printf("Failed to insert Alice: %v", err)
				tx.Rollback(ctx)
			} else {
				err = tx.Exec(ctx, insertSQL, "Bob Wilson", "bob@example.com")
				if err != nil {
					log.Printf("Failed to insert Bob: %v", err)
					tx.Rollback(ctx)
				} else {
					err = tx.Commit(ctx)
					if err != nil {
						log.Printf("Failed to commit transaction: %v", err)
					} else {
						log.Println("Transaction committed successfully")
					}
				}
			}
		}
	}

	// Example 4: Batch operations (for drivers that support it)
	log.Println("\n=== Example 4: Batch Operations ===")

	if conn != nil {
		// Note: Batch creation depends on the driver implementation
		// For PGX driver, we would use pgx.NewBatch()
		// This is a simplified example showing the concept

		// For now, just show individual operations
		insertSQL := "INSERT INTO users (name, email) VALUES ($1, $2)"

		err = conn.Exec(ctx, insertSQL, "User 1", "user1@example.com")
		if err != nil {
			log.Printf("Failed to insert User 1: %v", err)
		}

		err = conn.Exec(ctx, insertSQL, "User 2", "user2@example.com")
		if err != nil {
			log.Printf("Failed to insert User 2: %v", err)
		}

		err = conn.Exec(ctx, insertSQL, "User 3", "user3@example.com")
		if err != nil {
			log.Printf("Failed to insert User 3: %v", err)
		}

		log.Println("Individual operations completed")
	}

	log.Println("\n=== Example completed ===")
}

// init sets up example environment if needed
func init() {
	// Set default values from environment if available
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "testdb")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "password")
	}
}
