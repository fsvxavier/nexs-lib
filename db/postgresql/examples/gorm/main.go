package main

import (
	"context"
	"log"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
)

// User represents a user entity for GORM
type User struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"size:100;not null" json:"name"`
	Email string `gorm:"size:100;uniqueIndex;not null" json:"email"`
}

func main() {
	// Configuration specifically for GORM driver
	cfg := config.NewConfig(
		config.WithDriver(interfaces.DriverGORM),
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("testdb"),
		config.WithUser("postgres"),
		config.WithPassword("password"),
		config.WithSSLMode("disable"),
		config.WithMaxOpenConns(25),
		config.WithMaxIdleConns(5),
	)

	// Create GORM provider directly
	provider, err := gorm.NewProvider(cfg)
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

	log.Println("=== GORM Driver Specific Features ===")

	// Get a connection
	pool := provider.Pool()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal("Failed to acquire connection:", err)
	}
	defer conn.Release(ctx)

	// Auto migrate the schema
	log.Println("\n=== GORM Auto Migration ===")

	// Note: In a real GORM implementation, we would access the *gorm.DB instance
	// For this example, we'll use raw SQL to create the table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS gorm_users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL
	)`

	err = conn.Exec(ctx, createTableSQL)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}
	log.Println("Table created/verified successfully")

	// GORM-style operations through our interface
	log.Println("\n=== GORM-style CRUD Operations ===")

	// Create users
	insertSQL := "INSERT INTO gorm_users (name, email) VALUES ($1, $2)"

	err = conn.Exec(ctx, insertSQL, "Alice Johnson", "alice@gorm.com")
	if err != nil {
		log.Printf("Failed to create user Alice: %v", err)
	} else {
		log.Println("User Alice created successfully")
	}

	err = conn.Exec(ctx, insertSQL, "Bob Smith", "bob@gorm.com")
	if err != nil {
		log.Printf("Failed to create user Bob: %v", err)
	} else {
		log.Println("User Bob created successfully")
	}

	// Query single user (Find equivalent)
	log.Println("\n=== Finding Single User ===")

	var user User
	selectSQL := "SELECT id, name, email FROM gorm_users WHERE email = $1"
	err = conn.QueryOne(ctx, &user, selectSQL, "alice@gorm.com")
	if err != nil {
		log.Printf("Failed to find user: %v", err)
	} else {
		log.Printf("Found user: ID=%d, Name=%s, Email=%s", user.ID, user.Name, user.Email)
	}

	// Query all users (Find all equivalent)
	log.Println("\n=== Finding All Users ===")

	var users []User
	selectAllSQL := "SELECT id, name, email FROM gorm_users ORDER BY id"
	err = conn.QueryAll(ctx, &users, selectAllSQL)
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	} else {
		log.Printf("Found %d users:", len(users))
		for _, u := range users {
			log.Printf("  - ID: %d, Name: %s, Email: %s", u.ID, u.Name, u.Email)
		}
	}

	// Update user
	log.Println("\n=== Updating User ===")

	updateSQL := "UPDATE gorm_users SET name = $1 WHERE email = $2"
	err = conn.Exec(ctx, updateSQL, "Alice Updated", "alice@gorm.com")
	if err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		log.Println("User updated successfully")
	}

	// Count users
	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM gorm_users")
	if err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		log.Printf("Total users: %d", *count)
	}

	// Transaction example
	log.Println("\n=== GORM Transaction ===")

	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
	} else {
		// Create multiple users in transaction
		err = tx.Exec(ctx, insertSQL, "Charlie Brown", "charlie@gorm.com")
		if err != nil {
			log.Printf("Failed to create Charlie: %v", err)
			tx.Rollback(ctx)
		} else {
			err = tx.Exec(ctx, insertSQL, "Diana Prince", "diana@gorm.com")
			if err != nil {
				log.Printf("Failed to create Diana: %v", err)
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

	// Final count
	finalCount, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM gorm_users")
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

	log.Println("\n=== GORM Example completed ===")
	log.Println("Note: This example uses raw SQL through the unified interface.")
	log.Println("In practice, you would access the underlying *gorm.DB for ORM features.")
}
