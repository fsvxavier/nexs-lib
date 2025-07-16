# Global PostgreSQL Provider Example

This example demonstrates how to use the PostgreSQL provider in a generic, driver-agnostic way.

## Features Demonstrated

- ✅ **Generic Provider Usage**: Using the provider through interfaces without coupling to specific drivers
- ✅ **Configuration Builder Pattern**: Using `With*` functions for flexible configuration
- ✅ **Connection Pool Management**: Creating and managing connection pools
- ✅ **Multi-tenancy Support**: Enabling multi-tenant configurations
- ✅ **Read Replicas**: Configuring read replica connections with load balancing
- ✅ **Failover Support**: Setting up automatic failover to backup nodes
- ✅ **Basic Operations**: CRUD operations using generic interfaces
- ✅ **Transaction Management**: Working with database transactions
- ✅ **Batch Operations**: Performing batch operations for efficiency
- ✅ **Middleware Integration**: Built-in logging, timing, and metrics middleware
- ✅ **Statistics and Monitoring**: Pool statistics and connection monitoring

## Configuration Options

The example shows how to configure the provider with:

```go
cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

err := cfg.ApplyOptions(
    postgresql.WithMaxConns(20),
    postgresql.WithMinConns(5),
    postgresql.WithMaxConnLifetime(30*time.Minute),
    postgresql.WithLogging(true),
    postgresql.WithTiming(true),
    postgresql.WithMetrics(true),
    postgresql.WithMultiTenant(true),
    postgresql.WithReadReplicas([]string{...}, interfaces.LoadBalanceModeRoundRobin),
    postgresql.WithFailover([]string{...}, 3),
)
```

## Running the Example

1. **Setup PostgreSQL Database**:
   ```bash
   # Using Docker
   docker run --name postgres-test -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
   
   # Create test database
   docker exec -it postgres-test psql -U postgres -c "CREATE DATABASE testdb;"
   ```

2. **Update Connection String**:
   ```go
   // Update the connection string in main.go
   cfg := postgresql.NewDefaultConfig("postgres://postgres:password@localhost:5432/testdb")
   ```

3. **Run the Example**:
   ```bash
   go run main.go
   ```

## Expected Output

```
Provider: postgresql-pgx v1.0.0
Supported features: [connection_pooling transactions prepared_statements batch_operations listen_notify copy_operations multi_tenancy read_replicas failover ssl_tls context_support hooks middlewares health_check statistics]

--- Basic Operations ---
✓ Ping successful
✓ Table created
✓ Inserted 1 row(s)
✓ Queried user: John Doe <john@example.com>
✓ Queried 3 user(s)
✓ Total users: 3

--- Transaction Operations ---
✓ Transaction completed successfully

--- Batch Operations ---
✓ Batch operations completed

Pool Stats: Acquired=0, Total=5, Idle=5
Example completed successfully!
```

## Key Concepts

### Generic Provider Usage

The example shows how to use the provider through interfaces, making it easy to switch between different database drivers without changing application code:

```go
var provider interfaces.PostgreSQLProvider
provider, err := postgresql.NewPGXProvider() // Could be switched to other providers
```

### Pool Management

Demonstrates proper connection pool lifecycle:

```go
pool, err := provider.NewPool(ctx, cfg)
defer pool.Close() // Important: always close the pool

// Use pool.Acquire() for manual connection management
conn, err := pool.Acquire(ctx)
defer conn.Release()

// Or use pool.AcquireFunc() for automatic management
pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
    // Connection is automatically released
    return conn.Ping(ctx)
})
```

### Configuration Flexibility

Shows the builder pattern for configuration:

```go
// Start with defaults
cfg := postgresql.NewDefaultConfig(connectionString)

// Apply specific options as needed
cfg.ApplyOptions(
    postgresql.WithMaxConns(20),
    postgresql.WithLogging(true),
    // ... other options
)
```

## Thread Safety

All operations in this example are thread-safe:
- Connection pool is thread-safe
- Individual connections can be used safely within a single goroutine
- Configuration and providers are immutable after creation

## Error Handling

The example demonstrates proper error handling patterns:
- Always check errors from database operations
- Use proper cleanup with `defer`
- Handle transaction rollbacks on errors

## Performance Considerations

- **Connection Pooling**: Reuses connections for efficiency
- **Batch Operations**: Reduces round trips for multiple operations
- **Prepared Statements**: Can be used for repeated queries (not shown in this basic example)
- **Read Replicas**: Distributes read load across multiple nodes

## Next Steps

After running this example, check out:
- [Advanced Example](../advanced/README.md) - More complex scenarios with hooks and custom middleware
- [PGX Specific Example](../pgx/README.md) - Driver-specific features and optimizations
