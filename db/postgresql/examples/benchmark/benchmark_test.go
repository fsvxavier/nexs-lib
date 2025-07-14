package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// BenchmarkProviders compares the performance of all three providers
func BenchmarkProviders(b *testing.B) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		Database:        "testdb",
		Username:        "testuser",
		Password:        "testpass",
		TLSMode:         config.TLSModeDisable,
		ApplicationName: "benchmark-test",
		MaxConns:        10,
		MinConns:        2,
		ConnectTimeout:  time.Second * 5,
		QueryTimeout:    time.Second * 10,
	}

	ctx := context.Background()

	// Benchmark PGX
	b.Run("PGX", func(b *testing.B) {
		provider := pgx.NewProvider()
		for i := 0; i < b.N; i++ {
			pool, err := provider.CreatePool(ctx, cfg)
			if err != nil {
				log.Printf("Failed to create PGX pool: %v", err)
				continue
			}

			// Test ping
			if err := pool.Ping(ctx); err != nil {
				log.Printf("Failed to ping PGX: %v", err)
			}

			pool.Close()
		}
	})

	// Benchmark GORM
	b.Run("GORM", func(b *testing.B) {
		provider := gorm.NewProvider()
		for i := 0; i < b.N; i++ {
			pool, err := provider.CreatePool(ctx, cfg)
			if err != nil {
				log.Printf("Failed to create GORM pool: %v", err)
				continue
			}

			// Test ping
			if err := pool.Ping(ctx); err != nil {
				log.Printf("Failed to ping GORM: %v", err)
			}

			pool.Close()
		}
	})

	// Benchmark lib/pq
	b.Run("lib/pq", func(b *testing.B) {
		provider := pq.NewProvider()
		for i := 0; i < b.N; i++ {
			pool, err := provider.CreatePool(ctx, cfg)
			if err != nil {
				log.Printf("Failed to create lib/pq pool: %v", err)
				continue
			}

			// Test ping
			if err := pool.Ping(ctx); err != nil {
				log.Printf("Failed to ping lib/pq: %v", err)
			}

			pool.Close()
		}
	})
}

// Example demonstrates simple usage
func Example() {
	fmt.Println("PostgreSQL Providers Benchmark")
	fmt.Println("Use: go test -bench=. ./...")
	// Output: PostgreSQL Providers Benchmark
	// Use: go test -bench=. ./...
}
