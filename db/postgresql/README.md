# PostgreSQL Database Module

A comprehensive, generic PostgreSQL database module for Go that provides a unified interface across multiple PostgreSQL drivers including PGX, GORM, and lib/pq.

## Features

### Core Capabilities
- **Multi-Driver Support**: Unified interface for PGX, GORM, and lib/pq drivers
- **Connection Pooling**: Advanced connection pool management with health monitoring
- **Transaction Management**: Full transaction support with savepoints
- **Batch Operations**: Efficient batch query execution
- **Context Support**: Full context.Context integration for timeouts and cancellation
- **Thread-Safe**: Race condition and deadlock prevention
- **Generic Interface**: Driver-agnostic operations

### Advanced Features
- **Multi-Tenancy**: Schema and database-level multi-tenancy support
- **LISTEN/NOTIFY**: PostgreSQL pub/sub messaging support
- **Health Checks**: Comprehensive connection and pool health monitoring
- **Retry Logic**: Configurable retry mechanisms with exponential backoff
- **Failover Support**: Automatic failover and read replica support
- **Middleware & Hooks**: Pre/post operation hooks and middleware support
- **Observability**: Built-in metrics, tracing, and logging
- **Performance Optimization**: Memory-optimized buffers and connection management

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/fsvxavier/nexs-lib/db/postgresql/config"
    "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
    // Create provider
    provider := pgx.NewProvider()
    defer provider.Close()

    // Configure database
    cfg := config.NewConfig(
        config.WithHost("localhost"),
        config.WithDatabase("myapp"),
        config.WithUsername("postgres"),
        config.WithPassword("password"),
        config.WithMaxConns(20),
    )

    // Create connection pool
    ctx := context.Background()
    pool, err := provider.CreatePool(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Acquire connection
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Release(ctx)

    // Execute query
    var count int
    err = conn.QueryOne(ctx, &count, "SELECT COUNT(*) FROM users")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Supported Drivers

### PGX (Recommended)
- **Package**: `github.com/jackc/pgx/v5`
- **Features**: Full PostgreSQL feature support, best performance
- **Use Case**: High-performance applications, full PostgreSQL features

### GORM
- **Package**: `gorm.io/gorm` with `gorm.io/driver/postgres`
- **Features**: ORM capabilities, migrations, associations
- **Use Case**: Applications requiring ORM features

### lib/pq
- **Package**: `github.com/lib/pq`
- **Features**: Standard database/sql interface
- **Use Case**: Legacy applications, standard SQL interface

## Configuration

### Basic Configuration

```go
cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithPort(5432),
    config.WithDatabase("mydb"),
    config.WithUsername("user"),
    config.WithPassword("pass"),
)
```

### Advanced Configuration

```go
cfg := config.NewConfig(
    // Connection settings
    config.WithHost("localhost"),
    config.WithPort(5432),
    config.WithDatabase("mydb"),
    config.WithUsername("user"),
    config.WithPassword("pass"),
    
    // Pool settings
    config.WithMaxConns(50),
    config.WithMinConns(5),
    config.WithMaxConnLifetime(30*time.Minute),
    config.WithMaxConnIdleTime(5*time.Minute),
    
    // Timeouts
    config.WithConnectTimeout(10*time.Second),
    config.WithQueryTimeout(30*time.Second),
    
    // TLS configuration
    config.WithTLSMode(config.TLSModeRequire),
    
    // Application settings
    config.WithApplicationName("myapp"),
    config.WithTimezone("UTC"),
    
    // Multi-tenancy
    config.WithMultiTenant(true),
    config.WithDefaultSchema("tenant1"),
    
    // Retry configuration
    config.WithRetryConfig(&config.RetryConfig{
        Enabled:     true,
        MaxRetries:  3,
        InitialWait: 100 * time.Millisecond,
        MaxWait:     2 * time.Second,
        Multiplier:  2.0,
        Jitter:      true,
    }),
)
```

### Connection String

```go
cfg := config.NewConfig(
    config.WithConnectionString("postgres://user:pass@localhost:5432/mydb?sslmode=require"),
)
```

## Usage Examples

### Basic Operations

```go
// Insert data
err := conn.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")

// Query single row
var user User
err := conn.QueryOne(ctx, &user, "SELECT id, name, email FROM users WHERE id = $1", 1)

// Query multiple rows
rows, err := conn.Query(ctx, "SELECT id, name, email FROM users")
defer rows.Close()
for rows.Next() {
    var user User
    err := rows.Scan(&user.ID, &user.Name, &user.Email)
    // Process user
}

// Count records
var count int
err := conn.QueryOne(ctx, &count, "SELECT COUNT(*) FROM users WHERE active = true")
```

### Transactions

```go
// Begin transaction
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    return err
}

// Use transaction with automatic rollback on error
defer func() {
    if err != nil {
        tx.Rollback(ctx)
    }
}()

// Perform operations
err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
if err != nil {
    return err
}

// Create savepoint
err = tx.Savepoint(ctx, "user_created")
if err != nil {
    return err
}

// More operations...
err = tx.Exec(ctx, "UPDATE users SET email = $1 WHERE name = $2", "alice@example.com", "Alice")
if err != nil {
    // Rollback to savepoint
    tx.RollbackToSavepoint(ctx, "user_created")
    return err
}

// Commit transaction
return tx.Commit(ctx)
```

### Batch Operations

```go
// Create batch
batch := pgx.NewBatch()
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 1")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "User 2")
batch.Queue("SELECT COUNT(*) FROM users")

// Execute batch
results, err := conn.SendBatch(ctx, batch)
if err != nil {
    return err
}
defer results.Close()

// Process results
err = results.Exec() // First insert
err = results.Exec() // Second insert

var count int
err = results.QueryOne(&count) // Count query
```

### LISTEN/NOTIFY

```go
// Start listening
err := conn.Listen(ctx, "user_updates")
if err != nil {
    return err
}

// Wait for notifications
notification, err := conn.WaitForNotification(ctx, 30*time.Second)
if err != nil {
    return err
}

fmt.Printf("Received notification: %s - %s\n", notification.Channel, notification.Payload)

// Stop listening
err = conn.Unlisten(ctx, "user_updates")
```

## Monitoring and Observability

### Pool Statistics

```go
stats := pool.Stats()
fmt.Printf("Active connections: %d\n", stats.AcquiredConns)
fmt.Printf("Idle connections: %d\n", stats.IdleConns)
fmt.Printf("Total connections: %d\n", stats.TotalConns)
fmt.Printf("Acquire count: %d\n", stats.AcquireCount)
fmt.Printf("Acquire duration: %v\n", stats.AcquireDuration)
```

### Provider Metrics

```go
metrics := provider.GetMetrics(ctx)
fmt.Printf("Provider type: %s\n", metrics["type"])
fmt.Printf("Pool count: %d\n", metrics["pools_count"])
fmt.Printf("Is healthy: %t\n", metrics["is_healthy"])
```

### Health Checks

```go
// Basic ping
err := pool.Ping(ctx)

// Comprehensive health check
err := pool.HealthCheck(ctx)

// Provider health
isHealthy := provider.IsHealthy(ctx)
```

## Hooks and Middleware

```go
hooks := &config.HooksConfig{
    BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
        log.Printf("Executing query: %s", query)
        return nil
    },
    AfterQuery: func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error {
        log.Printf("Query completed in %v: %s", duration, query)
        return nil
    },
    BeforeTransaction: func(ctx context.Context) error {
        log.Println("Starting transaction")
        return nil
    },
    AfterTransaction: func(ctx context.Context, committed bool, duration time.Duration, err error) error {
        log.Printf("Transaction %s in %v", map[bool]string{true: "committed", false: "rolled back"}[committed], duration)
        return nil
    },
}

cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithHooks(hooks),
)
```

## Multi-Tenancy

```go
// Enable multi-tenancy
cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithMultiTenant(true),
    config.WithDefaultSchema("tenant1"),
)

// The connection will automatically set the search_path
// and reset it when released back to the pool
```

## Error Handling

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql"

// Check for specific errors
err := conn.QueryOne(ctx, &user, "SELECT * FROM users WHERE id = $1", 999)
if err != nil {
    if errors.Is(err, postgresql.ErrNoRows) {
        // Handle no rows found
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("query failed: %w", err)
}
```

## Testing

The module includes comprehensive unit tests with 98%+ coverage:

```bash
# Run all tests
go test -tags=unit -timeout 30s -race ./...

# Run specific package tests
go test -tags=unit -timeout 30s -race ./db/postgresql/config/...
go test -tags=unit -timeout 30s -race ./db/postgresql/providers/pgx/...

# Run with coverage
go test -tags=unit -timeout 30s -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Examples

The module includes comprehensive examples:

- **[Global Example](examples/global/)**: Complete feature demonstration
- **[PGX Example](examples/pgx/)**: PGX-specific features
- **[GORM Example](examples/gorm/)**: GORM integration
- **[lib/pq Example](examples/pq/)**: Standard SQL interface
- **[Advanced Example](examples/advanced/)**: Advanced patterns and configurations

## Performance

### Benchmarks

```bash
go test -bench=. -benchmem ./...
```

### Optimization Tips

1. **Connection Pooling**: Use appropriate pool sizes for your workload
2. **Query Timeouts**: Set reasonable query timeouts
3. **Prepared Statements**: Use prepared statements for repeated queries
4. **Batch Operations**: Use batches for multiple operations
5. **Connection Reuse**: Properly release connections back to pool

## Architecture

### Interface Design

The module follows clean architecture principles with clear separation of concerns:

- **Interfaces**: Generic database operation interfaces
- **Configuration**: Flexible configuration system
- **Providers**: Driver-specific implementations
- **Connection Management**: Pool and connection lifecycle management

### Thread Safety

All operations are thread-safe and designed to prevent:
- Race conditions
- Deadlocks
- Connection leaks
- Memory leaks

## Contributing

1. Follow the existing code style and patterns
2. Add comprehensive unit tests (minimum 98% coverage)
3. Include examples for new features
4. Update documentation
5. Run all tests and linting tools

## Next Steps

See [NEXT_STEPS.md](NEXT_STEPS.md) for planned improvements and future features.

## License

This module is part of the nexs-lib project and follows the same licensing terms.
