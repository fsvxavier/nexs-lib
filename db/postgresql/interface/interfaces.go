package interfaces

import (
	"context"
	"time"
)

// HookType represents the type of hook
type HookType int

const (
	// Connection hooks
	BeforeConnectionHook HookType = iota
	AfterConnectionHook
	BeforeReleaseHook
	AfterReleaseHook

	// Operation hooks
	BeforeQueryHook
	AfterQueryHook
	BeforeExecHook
	AfterExecHook
	BeforeTransactionHook
	AfterTransactionHook
	BeforeBatchHook
	AfterBatchHook

	// Pool hooks
	BeforeAcquireHook
	AfterAcquireHook

	// Error hooks
	OnErrorHook

	// Custom hooks (starting from 1000 to avoid conflicts)
	CustomHookBase HookType = 1000
)

// HookResult represents the result of a hook execution
type HookResult struct {
	Continue bool
	Error    error
	Data     map[string]interface{}
}

// ExecutionContext contains context information for hooks
type ExecutionContext struct {
	Context      context.Context
	Operation    string
	Query        string
	Args         []interface{}
	StartTime    time.Time
	Duration     time.Duration
	Error        error
	RowsAffected int64
	Metadata     map[string]interface{}
}

// Hook represents a function that can be executed at specific points
type Hook func(ctx *ExecutionContext) *HookResult

// HookManager manages hooks registration and execution
type HookManager interface {
	// RegisterHook registers a hook for a specific type
	RegisterHook(hookType HookType, hook Hook) error

	// RegisterCustomHook registers a custom hook with a custom type
	RegisterCustomHook(hookType HookType, name string, hook Hook) error

	// ExecuteHooks executes all hooks of a specific type
	ExecuteHooks(hookType HookType, ctx *ExecutionContext) error

	// UnregisterHook removes a hook
	UnregisterHook(hookType HookType) error

	// UnregisterCustomHook removes a custom hook
	UnregisterCustomHook(hookType HookType, name string) error

	// ListHooks returns all registered hooks
	ListHooks() map[HookType][]Hook
}

// IRow represents a single row result
type IRow interface {
	Scan(dest ...any) error
}

// IRows represents multiple rows result
type IRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
	CommandTag() CommandTag
	FieldDescriptions() []FieldDescription
	RawValues() [][]byte
}

// IBatch represents a batch of queries
type IBatch interface {
	Queue(query string, arguments ...any)
	QueueFunc(query string, arguments []any, callback func(IBatchResults) error)
	Len() int
	Clear()
	Reset()
}

// IBatchResults represents batch execution results
type IBatchResults interface {
	QueryRow() IRow
	Query() (IRows, error)
	Exec() (CommandTag, error)
	Close() error
	Err() error
}

// CommandTag represents command execution result metadata
type CommandTag interface {
	String() string
	RowsAffected() int64
	Insert() bool
	Update() bool
	Delete() bool
	Select() bool
}

// FieldDescription represents field metadata
type FieldDescription interface {
	Name() string
	TableOID() uint32
	TableAttributeNumber() uint16
	DataTypeOID() uint32
	DataTypeSize() int16
	TypeModifier() int32
	Format() int16
}

// TxIsoLevel represents transaction isolation level
type TxIsoLevel int

const (
	TxIsoLevelDefault TxIsoLevel = iota
	TxIsoLevelReadUncommitted
	TxIsoLevelReadCommitted
	TxIsoLevelRepeatableRead
	TxIsoLevelSerializable
)

// TxAccessMode represents transaction access mode
type TxAccessMode int

const (
	TxAccessModeReadWrite TxAccessMode = iota
	TxAccessModeReadOnly
)

// TxDeferrableMode represents transaction deferrable mode
type TxDeferrableMode int

const (
	TxDeferrableModeNotDeferrable TxDeferrableMode = iota
	TxDeferrableModeDeferrable
)

// TxOptions represents transaction options
type TxOptions struct {
	IsoLevel       TxIsoLevel
	AccessMode     TxAccessMode
	DeferrableMode TxDeferrableMode
	BeginQuery     string
}

// ITransaction represents a database transaction
type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// IConn represents a database connection
type IConn interface {
	// Query operations
	QueryRow(ctx context.Context, query string, args ...interface{}) IRow
	Query(ctx context.Context, query string, args ...interface{}) (IRows, error)
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error)

	// Execution operations
	Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error)

	// Batch operations
	SendBatch(ctx context.Context, batch IBatch) IBatchResults

	// Transaction operations
	Begin(ctx context.Context) (ITransaction, error)
	BeginTx(ctx context.Context, txOptions TxOptions) (ITransaction, error)

	// Connection management
	Release()
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	IsClosed() bool

	// Prepared statements
	Prepare(ctx context.Context, name, query string) error
	Deallocate(ctx context.Context, name string) error

	// Copy operations (PostgreSQL specific)
	CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc CopyFromSource) (int64, error)
	CopyTo(ctx context.Context, w CopyToWriter, query string, args ...interface{}) error

	// Listen/Notify (PostgreSQL specific)
	Listen(ctx context.Context, channel string) error
	Unlisten(ctx context.Context, channel string) error
	WaitForNotification(ctx context.Context, timeout time.Duration) (*Notification, error)

	// Multi-tenancy support
	SetTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context) (string, error)

	// Hook management
	GetHookManager() HookManager

	// Health check
	HealthCheck(ctx context.Context) error

	// Statistics
	Stats() ConnectionStats
}

// CopyFromSource represents a source for COPY FROM operations
type CopyFromSource interface {
	Next() bool
	Values() ([]interface{}, error)
	Err() error
}

// CopyToWriter represents a writer for COPY TO operations
type CopyToWriter interface {
	Write(row []interface{}) error
	Close() error
}

// Notification represents a PostgreSQL notification
type Notification struct {
	PID     uint32
	Channel string
	Payload string
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	TotalQueries       int64
	TotalExecs         int64
	TotalTransactions  int64
	TotalBatches       int64
	FailedQueries      int64
	FailedExecs        int64
	FailedTransactions int64
	AverageQueryTime   time.Duration
	AverageExecTime    time.Duration
	LastActivity       time.Time
	CreatedAt          time.Time
	MemoryUsage        MemoryStats
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	BufferSize         int64
	AllocatedBuffers   int32
	PooledBuffers      int32
	TotalAllocations   int64
	TotalDeallocations int64
}

// IPool represents a connection pool
type IPool interface {
	// Connection management
	Acquire(ctx context.Context) (IConn, error)
	AcquireFunc(ctx context.Context, f func(IConn) error) error

	// Pool management
	Close()
	Reset()

	// Pool information
	Stats() PoolStats
	Config() PoolConfig

	// Health check
	Ping(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// Hook management
	GetHookManager() HookManager

	// Buffer pool management for optimization
	GetBufferPool() BufferPool

	// Thread-safety monitoring
	GetSafetyMonitor() SafetyMonitor
}

// BufferPool manages memory buffers for optimization
type BufferPool interface {
	Get(size int) []byte
	Put(buf []byte)
	Stats() MemoryStats
	Reset()
}

// SafetyMonitor provides thread-safety monitoring
type SafetyMonitor interface {
	CheckDeadlocks() []DeadlockInfo
	CheckRaceConditions() []RaceConditionInfo
	CheckLeaks() []LeakInfo
	IsHealthy() bool
}

// DeadlockInfo represents deadlock information
type DeadlockInfo struct {
	Timestamp   time.Time
	Goroutines  []string
	StackTraces map[string]string
}

// RaceConditionInfo represents race condition information
type RaceConditionInfo struct {
	Timestamp time.Time
	Location  string
	Details   string
}

// LeakInfo represents resource leak information
type LeakInfo struct {
	Timestamp time.Time
	Resource  string
	Count     int64
	Details   string
}

// PoolStats represents pool statistics
type PoolStats struct {
	AcquireCount            int64
	AcquireDuration         time.Duration
	AcquiredConns           int32
	CanceledAcquireCount    int64
	ConstructingConns       int32
	EmptyAcquireCount       int64
	IdleConns               int32
	MaxConns                int32
	TotalConns              int32
	NewConnsCount           int64
	MaxLifetimeDestroyCount int64
	MaxIdleDestroyCount     int64
}

// PoolConfig represents pool configuration
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
	LazyConnect       bool
}

// Provider represents a database provider
type Provider interface {
	// Provider information
	Name() string
	Version() string
	SupportsFeature(feature string) bool

	// Connection and pool creation
	NewPool(ctx context.Context, config Config) (IPool, error)
	NewConn(ctx context.Context, config Config) (IConn, error)

	// Configuration validation
	ValidateConfig(config Config) error

	// Provider-specific features
	GetDriverName() string
	GetSupportedFeatures() []string
}

// Config represents database configuration
type Config interface {
	GetConnectionString() string
	GetPoolConfig() PoolConfig
	GetTLSConfig() TLSConfig
	GetRetryConfig() RetryConfig
	GetHookConfig() HookConfig
	IsMultiTenantEnabled() bool
	GetReadReplicaConfig() ReadReplicaConfig
	GetFailoverConfig() FailoverConfig
	Validate() error
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
	ServerName         string
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomizeWait   bool
}

// HookConfig represents hook configuration
type HookConfig struct {
	EnabledHooks []HookType
	CustomHooks  map[string]HookType
	HookTimeout  time.Duration
}

// ReadReplicaConfig represents read replica configuration
type ReadReplicaConfig struct {
	Enabled             bool
	ConnectionStrings   []string
	LoadBalanceMode     LoadBalanceMode
	HealthCheckInterval time.Duration
}

// LoadBalanceMode represents load balance mode for read replicas
type LoadBalanceMode int

const (
	LoadBalanceModeRoundRobin LoadBalanceMode = iota
	LoadBalanceModeRandom
	LoadBalanceModeWeighted
)

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled             bool
	FallbackNodes       []string
	HealthCheckInterval time.Duration
	RetryInterval       time.Duration
	MaxFailoverAttempts int
}

// PostgreSQLProvider represents the main PostgreSQL provider interface
type PostgreSQLProvider interface {
	Provider

	// PostgreSQL specific features
	NewListenConn(ctx context.Context, config Config) (IConn, error)
	CreateSchema(ctx context.Context, conn IConn, schemaName string) error
	DropSchema(ctx context.Context, conn IConn, schemaName string) error
	ListSchemas(ctx context.Context, conn IConn) ([]string, error)

	// Multi-database support
	CreateDatabase(ctx context.Context, conn IConn, dbName string) error
	DropDatabase(ctx context.Context, conn IConn, dbName string) error
	ListDatabases(ctx context.Context, conn IConn) ([]string, error)

	// Retry and failover support
	WithRetry(ctx context.Context, operation func() error) error
	WithFailover(ctx context.Context, operation func(conn IConn) error) error
	GetRetryManager() RetryManager
	GetFailoverManager() FailoverManager
}

// RetryManager manages retry operations
type RetryManager interface {
	Execute(ctx context.Context, operation func() error) error
	ExecuteWithConn(ctx context.Context, pool IPool, operation func(conn IConn) error) error
	UpdateConfig(config RetryConfig) error
	GetStats() RetryStats
}

// FailoverManager manages failover operations
type FailoverManager interface {
	Execute(ctx context.Context, operation func(conn IConn) error) error
	MarkNodeDown(nodeID string) error
	MarkNodeUp(nodeID string) error
	GetHealthyNodes() []string
	GetUnhealthyNodes() []string
	GetStats() FailoverStats
}

// RetryStats represents retry operation statistics
type RetryStats struct {
	TotalAttempts  int64
	SuccessfulOps  int64
	FailedOps      int64
	TotalRetries   int64
	AverageRetries float64
	LastRetryTime  time.Time
}

// FailoverStats represents failover operation statistics
type FailoverStats struct {
	TotalFailovers      int64
	SuccessfulFailovers int64
	FailedFailovers     int64
	CurrentActiveNode   string
	DownNodes           []string
	LastFailoverTime    time.Time
}

// ProviderType represents the type of database provider
type ProviderType string

const (
	ProviderTypePGX ProviderType = "pgx"
)

// ProviderFactory creates providers
type ProviderFactory interface {
	CreateProvider(providerType ProviderType) (PostgreSQLProvider, error)
	RegisterProvider(providerType ProviderType, provider PostgreSQLProvider) error
	ListProviders() []ProviderType
	GetProvider(providerType ProviderType) (PostgreSQLProvider, bool)
}
