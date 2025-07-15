package postgresql

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

// IConn represents a generic database connection interface
type IConn interface {
	// Query operations
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error)
	Query(ctx context.Context, query string, args ...interface{}) (IRows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error)

	// Execution operations
	Exec(ctx context.Context, query string, args ...interface{}) error

	// Batch operations
	SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error)

	// Transaction operations
	BeginTransaction(ctx context.Context) (ITransaction, error)
	BeginTransactionWithOptions(ctx context.Context, opts TxOptions) (ITransaction, error)

	// Prepared statements
	Prepare(ctx context.Context, name, query string) error

	// Hooks for middleware
	BeforeReleaseHook(ctx context.Context) error
	AfterAcquireHook(ctx context.Context) error

	// Connection management
	Release(ctx context.Context)
	Ping(ctx context.Context) error

	// PostgreSQL specific features
	Listen(ctx context.Context, channel string) error
	Unlisten(ctx context.Context, channel string) error
	WaitForNotification(ctx context.Context, timeout time.Duration) (*Notification, error)
}

// IPool represents a generic connection pool interface
type IPool interface {
	// Connection management
	Acquire(ctx context.Context) (IConn, error)
	AcquireWithTimeout(ctx context.Context, timeout time.Duration) (IConn, error)
	Close()

	// Health checks
	Ping(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// Pool statistics
	Stats() PoolStats

	// Utility methods
	GetConnWithNotPresent(ctx context.Context, conn IConn) (IConn, func(), error)
}

// ITransaction represents a generic transaction interface
type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Savepoint(ctx context.Context, name string) error
	RollbackToSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
}

// TxOptions represents transaction options
type TxOptions struct {
	IsoLevel       IsoLevel
	AccessMode     AccessMode
	DeferrableMode DeferrableMode
	ReadOnly       bool
}

// IsoLevel represents transaction isolation levels
type IsoLevel int

const (
	IsoLevelDefault IsoLevel = iota
	IsoLevelReadUncommitted
	IsoLevelReadCommitted
	IsoLevelRepeatableRead
	IsoLevelSerializable
)

// AccessMode represents transaction access modes
type AccessMode int

const (
	AccessModeDefault AccessMode = iota
	AccessModeReadWrite
	AccessModeReadOnly
)

// DeferrableMode represents transaction deferrable modes
type DeferrableMode int

const (
	DeferrableModeDefault DeferrableMode = iota
	DeferrableModeNotDeferrable
	DeferrableModeDeferrable
)

// IBatch represents a generic batch interface
type IBatch interface {
	Queue(query string, arguments ...interface{})
	Len() int
	Clear()
}

// IBatchResults represents batch execution results
type IBatchResults interface {
	QueryOne(dst interface{}) error
	QueryAll(dst interface{}) error
	Exec() error
	Close()
	Err() error
}

// IRow represents a single row result
type IRow interface {
	Scan(dest ...interface{}) error
}

// IRows represents multiple rows result
type IRows interface {
	Scan(dest ...interface{}) error
	Close()
	Next() bool
	Err() error
	RawValues() [][]byte
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	AcquireCount            int64
	AcquireDuration         time.Duration
	AcquiredConns           int32
	CanceledAcquireCount    int64
	ConstructingConns       int32
	EmptyAcquireCount       int64
	IdleConns               int32
	MaxConns                int32
	MinConns                int32
	TotalConns              int32
	NewConnsCount           int64
	MaxLifetimeDestroyCount int64
	MaxIdleDestroyCount     int64
}

// Notification represents a PostgreSQL LISTEN/NOTIFY notification
type Notification struct {
	PID     uint32
	Channel string
	Payload string
}

// ProviderType represents the database driver type
type ProviderType string

const (
	ProviderTypePGX  ProviderType = "pgx"
	ProviderTypeGORM ProviderType = "gorm"
)

// IProvider represents a generic provider interface
type IProvider interface {
	// Provider information
	Type() ProviderType
	Name() string
	Version() string

	// Pool management
	CreatePool(ctx context.Context, config *config.Config) (IPool, error)
	CreateConnection(ctx context.Context, config *config.Config) (IConn, error)

	// Health and diagnostics
	IsHealthy(ctx context.Context) bool
	GetMetrics(ctx context.Context) map[string]interface{}

	// Cleanup
	Close() error
}
