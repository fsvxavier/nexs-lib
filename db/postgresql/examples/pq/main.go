package main

import (
	"context"
	"log"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// User represents a user entity
type User struct {
	ID    int    `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

func main() {
	// Configuration specifically for lib/pq driver
	cfg := config.NewConfig(
		config.WithDriver(interfaces.DriverPQ),
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("testdb"),
		config.WithUser("postgres"),
		config.WithPassword("password"),
		config.WithSSLMode("disable"),
		config.WithMaxOpenConns(25),
		config.WithMaxIdleConns(5),
	)

	// Create PQ provider directly
	provider, err := pq.NewProvider(cfg)
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

	log.Println("=== Lib/PQ Driver Features ===")

	// Get a connection
	pool := provider.Pool()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal("Failed to acquire connection:", err)
	}
	defer conn.Release(ctx)

	// Create table for demo
	log.Println("\n=== Table Creation ===")

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS pq_users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL
	)`

	err = conn.Exec(ctx, createTableSQL)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}
	log.Println("Table created successfully")

	// Basic CRUD operations using lib/pq
	log.Println("\n=== CRUD Operations ===")

	// Insert users
	insertSQL := "INSERT INTO pq_users (name, email) VALUES ($1, $2)"

	err = conn.Exec(ctx, insertSQL, "John Doe", "john@pq.com")
	if err != nil {
		log.Printf("Failed to insert John: %v", err)
	} else {
		log.Println("Inserted John successfully")
	}

	err = conn.Exec(ctx, insertSQL, "Jane Smith", "jane@pq.com")
	if err != nil {
		log.Printf("Failed to insert Jane: %v", err)
	} else {
		log.Println("Inserted Jane successfully")
	}

	// Query single user
	log.Println("\n=== Querying Single User ===")

	var user User
	selectSQL := "SELECT id, name, email FROM pq_users WHERE email = $1"
	err = conn.QueryOne(ctx, &user, selectSQL, "john@pq.com")
	if err != nil {
		log.Printf("Failed to query user: %v", err)
	} else {
		log.Printf("Found user: ID=%d, Name=%s, Email=%s", user.ID, user.Name, user.Email)
	}

	// Query all users
	log.Println("\n=== Querying All Users ===")

	var users []User
	selectAllSQL := "SELECT id, name, email FROM pq_users ORDER BY id"
	err = conn.QueryAll(ctx, &users, selectAllSQL)
	if err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		log.Printf("Found %d users:", len(users))
		for _, u := range users {
			log.Printf("  - ID: %d, Name: %s, Email: %s", u.ID, u.Name, u.Email)
		}
	}

	// Update user
	log.Println("\n=== Updating User ===")

	updateSQL := "UPDATE pq_users SET name = $1 WHERE email = $2"
	err = conn.Exec(ctx, updateSQL, "John Updated", "john@pq.com")
	if err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		log.Println("User updated successfully")
	}

	// Count users
	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM pq_users")
	if err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		log.Printf("Total users: %d", *count)
	}

	// Transaction example
	log.Println("\n=== Transaction Operations ===")

	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
	} else {
		// Insert multiple users in transaction
		err = tx.Exec(ctx, insertSQL, "Alice Wilson", "alice@pq.com")
		if err != nil {
			log.Printf("Failed to insert Alice: %v", err)
			tx.Rollback(ctx)
		} else {
			err = tx.Exec(ctx, insertSQL, "Bob Johnson", "bob@pq.com")
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

	// Test row-level operations
	log.Println("\n=== Row Operations ===")

	row, err := conn.QueryRow(ctx, "SELECT name FROM pq_users WHERE email = $1", "alice@pq.com")
	if err != nil {
		log.Printf("Failed to query row: %v", err)
	} else {
		var name string
		err = row.Scan(&name)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
		} else {
			log.Printf("User name from row scan: %s", name)
		}
	}

	// Test multiple rows
	log.Println("\n=== Multiple Rows Operations ===")

	rows, err := conn.QueryRows(ctx, "SELECT name, email FROM pq_users ORDER BY id LIMIT 2")
	if err != nil {
		log.Printf("Failed to query rows: %v", err)
	} else {
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name, email string
			err = rows.Scan(&name, &email)
			if err != nil {
				log.Printf("Failed to scan row: %v", err)
				break
			}
			count++
			log.Printf("Row %d: Name=%s, Email=%s", count, name, email)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Rows iteration error: %v", err)
		}
	}

	// Final count
	finalCount, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM pq_users")
	if err != nil {
		log.Printf("Failed to get final count: %v", err)
	} else {
		log.Printf("Final user count: %d", *finalCount)
	}

	// Pool statistics
	stats := pool.Stats()
	log.Printf("\nPool Statistics:")
	log.Printf("  - Max Connections: %d", stats.MaxConns)
	log.Printf("  - Total Connections: %d", stats.TotalConns)
	log.Printf("  - Idle Connections: %d", stats.IdleConns)

	log.Println("\n=== Lib/PQ Example completed ===")
	log.Println("Note: This example demonstrates standard SQL operations using lib/pq driver.")
}
