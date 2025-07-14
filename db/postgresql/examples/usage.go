package examples

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	gormProvider "github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
	pgxProvider "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
	pqProvider "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// ExampleUsage demonstrates how to use all three PostgreSQL providers
func ExampleUsage() {
	ctx := context.Background()

	// Common configuration
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		Database:        "testdb",
		Username:        "testuser",
		Password:        "testpass",
		TLSMode:         config.TLSModeDisable,
		ApplicationName: "example-app",
		MaxConns:        10,
		MinConns:        2,
	}

	// Example 1: Using PGX Provider (Recommended for performance)
	fmt.Println("=== PGX Provider Example ===")
	pgxProv := pgxProvider.NewProvider()
	pgxPool, err := pgxProv.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create PGX pool: %v", err)
	} else {
		defer pgxPool.Close()

		conn, err := pgxPool.Acquire(ctx)
		if err != nil {
			log.Printf("Failed to acquire PGX connection: %v", err)
		} else {
			defer conn.Release(ctx)

			// Example query
			var count int
			countPtr, err := conn.QueryCount(ctx, "SELECT 1")
			if err != nil {
				log.Printf("PGX query error: %v", err)
			} else {
				count = *countPtr
				fmt.Printf("PGX Query result: %d\n", count)
			}

			// Example transaction
			tx, err := conn.BeginTransaction(ctx)
			if err != nil {
				log.Printf("PGX transaction error: %v", err)
			} else {
				defer tx.Rollback(ctx)

				err = tx.Exec(ctx, "SELECT 1")
				if err != nil {
					log.Printf("PGX transaction exec error: %v", err)
				} else {
					err = tx.Commit(ctx)
					if err != nil {
						log.Printf("PGX commit error: %v", err)
					} else {
						fmt.Println("PGX Transaction completed successfully")
					}
				}
			}
		}

		fmt.Printf("PGX Provider: %s, Version: %s\n", pgxProv.Name(), pgxProv.Version())
		fmt.Printf("PGX Pool Stats: %+v\n", pgxPool.Stats())
	}

	// Example 2: Using GORM Provider (Best for ORM features)
	fmt.Println("\n=== GORM Provider Example ===")
	gormProv := gormProvider.NewProvider()
	gormPool, err := gormProv.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create GORM pool: %v", err)
	} else {
		defer gormPool.Close()

		conn, err := gormPool.Acquire(ctx)
		if err != nil {
			log.Printf("Failed to acquire GORM connection: %v", err)
		} else {
			defer conn.Release(ctx)

			// Example query
			var count int
			countPtr, err := conn.QueryCount(ctx, "SELECT 1")
			if err != nil {
				log.Printf("GORM query error: %v", err)
			} else {
				count = *countPtr
				fmt.Printf("GORM Query result: %d\n", count)
			}

			// Example transaction
			tx, err := conn.BeginTransaction(ctx)
			if err != nil {
				log.Printf("GORM transaction error: %v", err)
			} else {
				defer tx.Rollback(ctx)

				err = tx.Exec(ctx, "SELECT 1")
				if err != nil {
					log.Printf("GORM transaction exec error: %v", err)
				} else {
					err = tx.Commit(ctx)
					if err != nil {
						log.Printf("GORM commit error: %v", err)
					} else {
						fmt.Println("GORM Transaction completed successfully")
					}
				}
			}
		}

		fmt.Printf("GORM Provider: %s, Version: %s\n", gormProv.Name(), gormProv.Version())
		fmt.Printf("GORM Pool Stats: %+v\n", gormPool.Stats())
	}

	// Example 3: Using lib/pq Provider (Standard library compatible)
	fmt.Println("\n=== lib/pq Provider Example ===")
	pqProv := pqProvider.NewProvider()
	pqPool, err := pqProv.CreatePool(ctx, cfg)
	if err != nil {
		log.Printf("Failed to create lib/pq pool: %v", err)
	} else {
		defer pqPool.Close()

		conn, err := pqPool.Acquire(ctx)
		if err != nil {
			log.Printf("Failed to acquire lib/pq connection: %v", err)
		} else {
			defer conn.Release(ctx)

			// Example query
			var count int
			countPtr, err := conn.QueryCount(ctx, "SELECT 1")
			if err != nil {
				log.Printf("lib/pq query error: %v", err)
			} else {
				count = *countPtr
				fmt.Printf("lib/pq Query result: %d\n", count)
			}

			// Example transaction
			tx, err := conn.BeginTransaction(ctx)
			if err != nil {
				log.Printf("lib/pq transaction error: %v", err)
			} else {
				defer tx.Rollback(ctx)

				err = tx.Exec(ctx, "SELECT 1")
				if err != nil {
					log.Printf("lib/pq transaction exec error: %v", err)
				} else {
					err = tx.Commit(ctx)
					if err != nil {
						log.Printf("lib/pq commit error: %v", err)
					} else {
						fmt.Println("lib/pq Transaction completed successfully")
					}
				}
			}
		}

		fmt.Printf("lib/pq Provider: %s, Version: %s\n", pqProv.Name(), pqProv.Version())
		fmt.Printf("lib/pq Pool Stats: %+v\n", pqPool.Stats())
	}

	// Example 4: Provider Selection Based on Use Case
	fmt.Println("\n=== Provider Selection Guide ===")
	fmt.Println("PGX: Best for high performance, low-level PostgreSQL features")
	fmt.Println("GORM: Best for ORM features, migrations, model relationships")
	fmt.Println("lib/pq: Best for standard library compatibility, simple use cases")

	// Example metrics
	fmt.Println("\n=== Provider Metrics ===")
	providers := []postgresql.IProvider{pgxProv, gormProv, pqProv}

	for _, provider := range providers {
		metrics := provider.GetMetrics(ctx)
		fmt.Printf("%s Metrics: %+v\n", provider.Name(), metrics)
	}
}

// ExampleFactoryPattern demonstrates a factory pattern for provider selection
func ExampleFactoryPattern() {
	ctx := context.Background()

	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		TLSMode:  config.TLSModeDisable,
	}

	// Factory function to create providers
	createProvider := func(providerType postgresql.ProviderType) (postgresql.IProvider, error) {
		switch providerType {
		case postgresql.ProviderTypePGX:
			return pgxProvider.NewProvider(), nil
		case postgresql.ProviderTypeGORM:
			return gormProvider.NewProvider(), nil
		case postgresql.ProviderTypePQ:
			return pqProvider.NewProvider(), nil
		default:
			return nil, fmt.Errorf("unknown provider type: %s", providerType)
		}
	}

	// Example usage with factory
	providerTypes := []postgresql.ProviderType{
		postgresql.ProviderTypePGX,
		postgresql.ProviderTypeGORM,
		postgresql.ProviderTypePQ,
	}

	for _, providerType := range providerTypes {
		provider, err := createProvider(providerType)
		if err != nil {
			log.Printf("Failed to create provider %s: %v", providerType, err)
			continue
		}

		pool, err := provider.CreatePool(ctx, cfg)
		if err != nil {
			log.Printf("Failed to create pool for %s: %v", provider.Name(), err)
			continue
		}
		defer pool.Close()

		fmt.Printf("Successfully created pool for provider: %s (Type: %s, Version: %s)\n",
			provider.Name(), provider.Type(), provider.Version())
	}
}

// ExampleAdvancedFeatures demonstrates advanced features of each provider
func ExampleAdvancedFeatures() {
	fmt.Println("=== Advanced Features Comparison ===")

	// PGX Advanced Features
	fmt.Println("\nPGX Advanced Features:")
	fmt.Println("- Native PostgreSQL types support")
	fmt.Println("- LISTEN/NOTIFY support")
	fmt.Println("- Batch operations")
	fmt.Println("- Connection pooling with detailed statistics")
	fmt.Println("- Prepared statements with automatic cleanup")

	// GORM Advanced Features
	fmt.Println("\nGORM Advanced Features:")
	fmt.Println("- ORM with associations and relationships")
	fmt.Println("- Automatic migrations")
	fmt.Println("- Hooks and callbacks")
	fmt.Println("- Soft deletes")
	fmt.Println("- Query builder with method chaining")

	// lib/pq Advanced Features
	fmt.Println("\nlib/pq Advanced Features:")
	fmt.Println("- Standard database/sql interface")
	fmt.Println("- LISTEN/NOTIFY support")
	fmt.Println("- SSL/TLS configuration")
	fmt.Println("- Compatible with most SQL builders")
	fmt.Println("- Stable and well-tested codebase")

	// Feature matrix
	fmt.Println("\n=== Feature Matrix ===")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "Feature", "PGX", "GORM", "lib/pq")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "Performance", "High", "Medium", "Medium")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "ORM Features", "No", "Yes", "No")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "Batch Operations", "Yes", "No", "No")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "LISTEN/NOTIFY", "Yes", "No", "Yes")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "Std Lib Compat", "No", "Partial", "Yes")
	fmt.Printf("%-20s | %-6s | %-6s | %-6s\n", "Learning Curve", "Medium", "Low", "Low")
}
