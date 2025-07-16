# PGX Provider Implementation

This document describes the complete implementation of the PGX provider for the PostgreSQL interface.

## Overview

The PGX provider is a comprehensive implementation of the PostgreSQL interface using the `github.com/jackc/pgx/v5` driver. It provides connection pooling, transactions, batch operations, hooks, middleware, and many advanced features.

## Architecture

### File Structure

```
pgx/
├── pgx.go           # Main package file with documentation
├── provider.go      # Provider implementation and interface methods
├── pool.go          # Connection pool implementation  
├── conn.go          # Connection implementation
├── transaction.go   # Transaction implementation
├── batch.go         # Batch operations implementation
├── rows.go          # Row and result set implementations
├── errors.go        # Error definitions and utilities
├── provider_test.go # Comprehensive tests
└── README.md        # This documentation
```

### Core Components

#### 1. PGXProvider (`provider.go`)
- Implements `interfaces.PostgreSQLProvider`
- Provides factory methods for creating pools and connections
- Supports configuration validation
- Schema and database management operations

#### 2. PGXPool (`pool.go`)
- Implements `interfaces.IPool`
- Connection pool management using `pgxpool`
- Hook and middleware integration
- Health checks and statistics

#### 3. PGXConn (`conn.go`)
- Implements `interfaces.IConn`
- Direct connection management
- Query, exec, and transaction operations
- Copy operations (COPY FROM/TO)
- Listen/Notify support
- Multi-tenancy support

#### 4. PGXTransaction (`transaction.go`)
- Implements `interfaces.ITransaction`
- Transaction lifecycle management
- Nested transaction support (via savepoints)
- Hook and middleware integration
- Thread-safe operations

#### 5. PGXBatch (`batch.go`)
- Implements `interfaces.IBatch` and `interfaces.IBatchResults`
- Batch operation queuing and execution
- Result processing

#### 6. Row/Rows Types (`rows.go`)
- `PGXRow` implements `interfaces.IRow`
- `PGXRows` implements `interfaces.IRows`
- `PGXCommandTag` implements `interfaces.CommandTag`
- `PGXFieldDescription` implements `interfaces.FieldDescription`

## Features

### Supported Features

✅ **Connection Pooling** - Full pgxpool integration with configuration support  
✅ **Transactions** - Complete transaction lifecycle with commit/rollback  
✅ **Prepared Statements** - Statement preparation and deallocation  
✅ **Batch Operations** - Efficient batch query execution  
✅ **Copy Operations** - PostgreSQL COPY FROM/TO support  
✅ **Listen/Notify** - PostgreSQL asynchronous notifications  
✅ **Hooks** - Before/after operation hooks with configurable types  
✅ **Middleware** - Request/response middleware chain  
✅ **Multi-tenancy** - Per-connection tenant ID management  
✅ **Health Checks** - Connection and pool health validation  
✅ **Statistics** - Comprehensive operation statistics tracking  
✅ **SSL/TLS** - Secure connection support  
✅ **Error Handling** - Comprehensive error types and utilities  
✅ **Thread Safety** - Concurrent operation support with proper locking  

### Configuration Support

- **Pool Configuration**: Min/max connections, timeouts, health checks
- **TLS Configuration**: SSL/TLS settings and certificate management
- **Retry Configuration**: Retry policies with backoff strategies
- **Middleware Configuration**: Configurable middleware features
- **Hook Configuration**: Selective hook enablement and timeouts
- **Multi-tenant Configuration**: Tenant isolation settings
- **Failover Configuration**: High availability and failover support

## Usage Examples

### Basic Connection

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"

// Create a simple connection
config := createConfig() // implements interfaces.Config
conn, err := pgx.NewConn(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer conn.Close(ctx)

// Execute query
rows, err := conn.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %d - %s\n", id, name)
}
```

### Connection Pool

```go
// Create a connection pool
pool, err := pgx.NewPool(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// Acquire connection from pool
conn, err := pool.Acquire(ctx)
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Use connection...
```

### Transactions

```go
// Begin transaction
tx, err := conn.Begin(ctx)
if err != nil {
    log.Fatal(err)
}

// Execute operations
_, err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
if err != nil {
    tx.Rollback(ctx)
    log.Fatal(err)
}

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE user_id = $2", 100, userID)
if err != nil {
    tx.Rollback(ctx)
    log.Fatal(err)
}

// Commit transaction
if err := tx.Commit(ctx); err != nil {
    log.Fatal(err)
}
```

### Batch Operations

```go
// Create batch
batch := &pgx.PGXBatch{}
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 1")
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 2")
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 3")

// Send batch
results := conn.SendBatch(ctx, batch)
defer results.Close()

// Process results
for i := 0; i < batch.Len(); i++ {
    tag, err := results.Exec()
    if err != nil {
        log.Printf("Batch item %d failed: %v", i, err)
        continue
    }
    fmt.Printf("Inserted %d rows\n", tag.RowsAffected())
}
```

### Copy Operations

```go
// Copy data from CSV-like source
type CSVSource struct {
    rows [][]interface{}
    idx  int
}

func (c *CSVSource) Next() bool {
    return c.idx < len(c.rows)
}

func (c *CSVSource) Values() ([]interface{}, error) {
    if c.idx >= len(c.rows) {
        return nil, io.EOF
    }
    values := c.rows[c.idx]
    c.idx++
    return values, nil
}

func (c *CSVSource) Err() error {
    return nil
}

// Perform copy
source := &CSVSource{rows: [][]interface{}{
    {1, "Alice"},
    {2, "Bob"},
    {3, "Charlie"},
}}

rowsCopied, err := conn.CopyFrom(ctx, "users", []string{"id", "name"}, source)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Copied %d rows\n", rowsCopied)
```

### Listen/Notify

```go
// Listen for notifications
err := conn.Listen(ctx, "user_updates")
if err != nil {
    log.Fatal(err)
}

// Wait for notification
notification, err := conn.WaitForNotification(ctx, time.Minute)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Received notification: %s - %s\n", notification.Channel, notification.Payload)
```

## Error Handling

The provider includes comprehensive error handling with specific error types:

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"

// Check for connection errors
if pgx.IsConnError(err) {
    // Handle connection-related errors
}

// Check for specific error types
if errors.Is(err, pgx.ErrConnectionClosed) {
    // Connection was closed
}

if errors.Is(err, pgx.ErrInvalidConfig) {
    // Configuration was invalid
}
```

## Testing

The implementation includes comprehensive tests covering:

- Provider interface compliance
- Configuration validation
- Connection and pool creation
- Error handling
- Type implementations
- Performance benchmarks

Run tests with:

```bash
go test -v ./...
```

## Performance Considerations

### Connection Pooling
- Use connection pools for high-concurrency applications
- Configure appropriate min/max connections based on workload
- Monitor pool statistics for optimal sizing

### Batch Operations
- Use batches for multiple operations to reduce round trips
- Consider batch size limits for memory usage
- Handle partial batch failures appropriately

### Prepared Statements
- Use prepared statements for repeated queries
- Clean up prepared statements when no longer needed
- Consider statement caching strategies

### Statistics and Monitoring
- Enable statistics collection for performance insights
- Monitor connection health and query performance
- Use health checks for load balancer integration

## Dependencies

- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/jackc/pgx/v5/pgxpool` - Connection pooling
- `github.com/jackc/pgx/v5/pgconn` - Low-level connection handling

## Version Compatibility

- PostgreSQL 10+ (recommended 12+)
- Go 1.19+
- PGX v5.x

## Contributing

When contributing to the PGX provider:

1. Ensure all interfaces are properly implemented
2. Add comprehensive tests for new features
3. Update documentation and examples
4. Follow the existing code style and patterns
5. Consider backward compatibility when making changes

## Future Enhancements

Potential areas for future development:

- Connection load balancing across read replicas
- Advanced connection pooling strategies
- Performance optimization for high-throughput scenarios
- Additional PostgreSQL-specific features
- Enhanced observability and metrics
- Integration with tracing systems
