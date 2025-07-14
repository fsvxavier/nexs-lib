package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// User represents a user entity for demonstration
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func main() {
	// Configuration for all providers
	cfg := &config.Config{
		Host:            os.Getenv("DB_HOST"),
		Port:            5432,
		Database:        os.Getenv("DB_NAME"),
		Username:        os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		TLSMode:         config.TLSModeDisable,
		ApplicationName: "nexs-lib-demo",
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: time.Minute * 30,
		ConnectTimeout:  time.Second * 10,
		QueryTimeout:    time.Second * 30,
		RuntimeParams: map[string]string{
			"statement_timeout": "30s",
			"lock_timeout":      "10s",
		},
	}

	// Fallback to localhost if env vars not set
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Database == "" {
		cfg.Database = "testdb"
	}
	if cfg.Username == "" {
		cfg.Username = "testuser"
	}
	if cfg.Password == "" {
		cfg.Password = "testpass"
	}

	ctx := context.Background()

	fmt.Println("🚀 PostgreSQL Providers Demo")
	fmt.Println("============================")

	// Demonstrate each provider
	demonstratePGX(ctx, cfg)
	demonstrateGORM(ctx, cfg)
	demonstrateLibPQ(ctx, cfg)

	// Performance comparison
	performanceComparison(ctx, cfg)

	// Best practices
	showBestPractices()
}

func demonstratePGX(ctx context.Context, cfg *config.Config) {
	fmt.Println("\n🔥 PGX Provider - High Performance")
	fmt.Println("-----------------------------------")

	provider := pgx.NewProvider()
	fmt.Printf("Provider: %s v%s\n", provider.Name(), provider.Version())

	// Create pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create PGX pool: %v", err)
		return
	}
	defer pool.Close()

	// Get connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to get PGX connection: %v", err)
		return
	}
	defer conn.Close()

	// Create table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users_pgx (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	if err := conn.Exec(ctx, createTableSQL); err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}

	// Insert user
	insertSQL := `INSERT INTO users_pgx (name, email) VALUES ($1, $2) RETURNING id, created_at`
	row := conn.QueryOne(ctx, insertSQL, "John Doe (PGX)", "john.pgx@example.com")

	var user User
	if err := row.Scan(&user.ID, &user.CreatedAt); err != nil {
		log.Printf("Failed to scan user: %v", err)
		return
	}

	user.Name = "John Doe (PGX)"
	user.Email = "john.pgx@example.com"

	fmt.Printf("✅ Created user: ID=%d, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)

	// Transaction example
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// Update in transaction
	updateSQL := `UPDATE users_pgx SET name = $1 WHERE id = $2`
	if err := tx.Exec(ctx, updateSQL, "John Updated (PGX)", user.ID); err != nil {
		tx.Rollback(ctx)
		log.Printf("Failed to update user: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Println("✅ Transaction committed successfully")

	// Cleanup
	conn.Exec(ctx, "DROP TABLE IF EXISTS users_pgx")
}

func demonstrateGORM(ctx context.Context, cfg *config.Config) {
	fmt.Println("\n🛠️  GORM Provider - ORM Features")
	fmt.Println("----------------------------------")

	provider := gorm.NewProvider()
	fmt.Printf("Provider: %s v%s\n", provider.Name(), provider.Version())

	// Create pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create GORM pool: %v", err)
		return
	}
	defer pool.Close()

	// Get connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to get GORM connection: %v", err)
		return
	}
	defer func() {
		// Connection will be returned to pool automatically
	}()

	// Create table (GORM style)
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users_gorm (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	if err := conn.Exec(ctx, createTableSQL); err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}

	// Insert user
	insertSQL := `INSERT INTO users_gorm (name, email) VALUES ($1, $2) RETURNING id, created_at`
	row := conn.QueryOne(ctx, insertSQL, "Jane Smith (GORM)", "jane.gorm@example.com")

	var user User
	if err := row.Scan(&user.ID, &user.CreatedAt); err != nil {
		log.Printf("Failed to scan user: %v", err)
		return
	}

	user.Name = "Jane Smith (GORM)"
	user.Email = "jane.gorm@example.com"

	fmt.Printf("✅ Created user: ID=%d, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)

	// Transaction with GORM
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// Batch operations example
	batchSQL := `INSERT INTO users_gorm (name, email) VALUES ($1, $2), ($3, $4)`
	if err := tx.Exec(ctx, batchSQL, "User 1", "user1@gorm.com", "User 2", "user2@gorm.com"); err != nil {
		tx.Rollback(ctx)
		log.Printf("Failed to insert batch: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Println("✅ Batch operations completed successfully")

	// Cleanup
	conn.Exec(ctx, "DROP TABLE IF EXISTS users_gorm")
}

func demonstrateLibPQ(ctx context.Context, cfg *config.Config) {
	fmt.Println("\n📚 lib/pq Provider - Standard Compatibility")
	fmt.Println("-------------------------------------------")

	provider := pq.NewProvider()
	fmt.Printf("Provider: %s v%s\n", provider.Name(), provider.Version())

	// Create pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create lib/pq pool: %v", err)
		return
	}
	defer pool.Close()

	// Get connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to get lib/pq connection: %v", err)
		return
	}
	defer func() {
		// Connection will be returned to pool automatically
	}()

	// Create table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users_pq (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	if err := conn.Exec(ctx, createTableSQL); err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}

	// Insert user
	insertSQL := `INSERT INTO users_pq (name, email) VALUES ($1, $2) RETURNING id, created_at`
	row := conn.QueryOne(ctx, insertSQL, "Bob Wilson (lib/pq)", "bob.pq@example.com")

	var user User
	if err := row.Scan(&user.ID, &user.CreatedAt); err != nil {
		log.Printf("Failed to scan user: %v", err)
		return
	}

	user.Name = "Bob Wilson (lib/pq)"
	user.Email = "bob.pq@example.com"

	fmt.Printf("✅ Created user: ID=%d, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)

	// Prepared statement example
	querySQL := `SELECT name, email FROM users_pq WHERE id = $1`
	queryRow := conn.QueryOne(ctx, querySQL, user.ID)

	var name, email string
	if err := queryRow.Scan(&name, &email); err != nil {
		log.Printf("Failed to query user: %v", err)
		return
	}

	fmt.Printf("✅ Queried user: Name=%s, Email=%s\n", name, email)

	// Cleanup
	conn.Exec(ctx, "DROP TABLE IF EXISTS users_pq")
}

func performanceComparison(ctx context.Context, cfg *config.Config) {
	fmt.Println("\n⚡ Performance Comparison")
	fmt.Println("-------------------------")

	providers := map[string]interface{}{
		"PGX":    pgx.NewProvider(),
		"GORM":   gorm.NewProvider(),
		"lib/pq": pq.NewProvider(),
	}

	for name, p := range providers {
		fmt.Printf("\n📊 %s Performance:\n", name)

		start := time.Now()

		// Create pool
		pool, err := p.(interface {
			CreatePool(context.Context, *config.Config) (interface{}, error)
		}).CreatePool(ctx, cfg)

		if err != nil {
			fmt.Printf("❌ Failed to create pool: %v\n", err)
			continue
		}

		poolTime := time.Since(start)
		fmt.Printf("   Pool Creation: %v\n", poolTime)

		// Close pool
		if closer, ok := pool.(interface{ Close() }); ok {
			closer.Close()
		}

		fmt.Printf("   ✅ Provider initialized successfully\n")
	}
}

func showBestPractices() {
	fmt.Println("\n💡 Best Practices & Recommendations")
	fmt.Println("====================================")

	practices := []struct {
		title       string
		description string
		providers   []string
	}{
		{
			title:       "High Performance Applications",
			description: "Use PGX for maximum performance and native PostgreSQL features",
			providers:   []string{"PGX"},
		},
		{
			title:       "ORM-based Applications",
			description: "Use GORM for rapid development with ORM features",
			providers:   []string{"GORM"},
		},
		{
			title:       "Standard Library Compatibility",
			description: "Use lib/pq for compatibility with existing database/sql code",
			providers:   []string{"lib/pq"},
		},
		{
			title:       "Connection Pooling",
			description: "Configure appropriate pool sizes based on your workload",
			providers:   []string{"PGX", "GORM", "lib/pq"},
		},
		{
			title:       "Error Handling",
			description: "Always handle connection and query errors appropriately",
			providers:   []string{"PGX", "GORM", "lib/pq"},
		},
		{
			title:       "Transaction Management",
			description: "Use transactions for data consistency and rollback capability",
			providers:   []string{"PGX", "GORM", "lib/pq"},
		},
	}

	for i, practice := range practices {
		fmt.Printf("%d. %s\n", i+1, practice.title)
		fmt.Printf("   %s\n", practice.description)
		fmt.Printf("   Providers: %v\n\n", practice.providers)
	}

	fmt.Println("🔧 Configuration Tips:")
	fmt.Println("   • Set appropriate timeouts for your use case")
	fmt.Println("   • Monitor connection pool metrics")
	fmt.Println("   • Use TLS in production environments")
	fmt.Println("   • Configure runtime parameters for optimization")
	fmt.Println("   • Test with your expected workload")
}
