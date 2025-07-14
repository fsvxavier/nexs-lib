# Global PostgreSQL Example

This example demonstrates the comprehensive usage of the nexs-lib PostgreSQL module with a generic approach that works with any supported driver.

## Features Demonstrated

- **Provider Management**: Creating and managing a PostgreSQL provider
- **Configuration**: Setting up database configuration with various options
- **Connection Pooling**: Creating and managing connection pools
- **Single Connections**: Creating standalone database connections
- **Basic Operations**: INSERT, SELECT, UPDATE, DELETE operations
- **Transactions**: Transaction management with commit/rollback
- **Savepoints**: Using savepoints within transactions
- **Batch Operations**: Executing multiple queries in a batch
- **Health Checks**: Verifying database connectivity
- **Metrics**: Monitoring pool and provider statistics
- **Error Handling**: Proper error handling throughout

## Prerequisites

- PostgreSQL database running on localhost:5432
- Database named `nexs_lib_example`
- User `postgres` with password `password`

## Configuration Options

The example demonstrates various configuration options:

```go
cfg := config.NewConfig(
    config.WithHost("localhost"),                    // Database host
    config.WithPort(5432),                          // Database port
    config.WithDatabase("nexs_lib_example"),         // Database name
    config.WithUsername("postgres"),                 // Username
    config.WithPassword("password"),                 // Password
    config.WithMaxConns(20),                        // Maximum connections
    config.WithMinConns(2),                         // Minimum connections
    config.WithMaxConnLifetime(30*time.Minute),     // Max connection lifetime
    config.WithMaxConnIdleTime(5*time.Minute),      // Max idle time
    config.WithApplicationName("nexs-lib-example"),  // Application name
    config.WithQueryTimeout(30*time.Second),        // Query timeout
    config.WithConnectTimeout(10*time.Second),      // Connection timeout
)
```

## Running the Example

```bash
cd examples/global
go run main.go
```

## Expected Output

The example will:

1. Create a PostgreSQL provider
2. Set up configuration and validate it
3. Create a connection pool
4. Test database connectivity
5. Show pool statistics
6. Demonstrate basic SQL operations
7. Show transaction usage with savepoints
8. Demonstrate batch operations
9. Display provider metrics
10. Clean up resources

## Operations Demonstrated

### Basic SQL Operations
- Creating tables
- Inserting data
- Querying data
- Counting records

### Transaction Management
```go
tx, err := conn.BeginTransaction(ctx)
if err != nil {
    return err
}

// Perform operations
err = tx.QueryOne(ctx, &result, query, args...)

// Create savepoint
err = tx.Savepoint(ctx, "checkpoint1")

// Rollback to savepoint if needed
err = tx.RollbackToSavepoint(ctx, "checkpoint1")

// Commit or rollback
if err != nil {
    tx.Rollback(ctx)
} else {
    tx.Commit(ctx)
}
```

### Batch Operations
```go
batch := pgx.NewBatch()
batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "User 1", "user1@example.com")
batch.Queue("INSERT INTO users (name, email) VALUES ($1, $2)", "User 2", "user2@example.com")

results, err := conn.SendBatch(ctx, batch)
defer results.Close()

// Process results
err = results.Exec() // First operation
err = results.Exec() // Second operation
```

### Pool Statistics
```go
stats := pool.Stats()
fmt.Printf("Max Connections: %d\n", stats.MaxConns)
fmt.Printf("Active Connections: %d\n", stats.AcquiredConns)
fmt.Printf("Idle Connections: %d\n", stats.IdleConns)
```

## Error Handling

The example demonstrates proper error handling patterns:

- Configuration validation
- Connection failures
- Query errors
- Transaction rollbacks
- Resource cleanup

## Best Practices Shown

1. **Resource Management**: Proper cleanup with defer statements
2. **Context Usage**: Using context for timeouts and cancellation
3. **Error Handling**: Comprehensive error checking
4. **Transaction Safety**: Proper transaction commit/rollback
5. **Configuration Validation**: Validating configuration before use
6. **Health Monitoring**: Regular health checks and metrics
7. **Connection Lifecycle**: Proper acquisition and release of connections

## Customization

You can modify the configuration to:

- Use different database credentials
- Adjust connection pool settings
- Change timeout values
- Add custom runtime parameters
- Enable TLS/SSL
- Configure multi-tenancy settings

## Troubleshooting

**Connection Failures**: The example handles database connection failures gracefully and will continue to demonstrate the API even when a database is not available.

**Permission Errors**: Ensure the PostgreSQL user has the necessary permissions to create tables and perform operations.

**Port Conflicts**: If PostgreSQL is running on a different port, update the configuration accordingly.
