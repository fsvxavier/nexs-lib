package interfaces

import (
	"context"
	"time"
)

// IConn representa uma conexão com o banco de dados
type IConn interface {
	// Query operations
	QueryRow(ctx context.Context, query string, args ...interface{}) IRow
	Query(ctx context.Context, query string, args ...interface{}) (IRows, error)
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error)

	// Execution operations
	Exec(ctx context.Context, query string, args ...interface{}) (ICommandTag, error)

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
	CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc ICopyFromSource) (int64, error)
	CopyTo(ctx context.Context, w ICopyToWriter, query string, args ...interface{}) error

	// Listen/Notify (PostgreSQL specific)
	Listen(ctx context.Context, channel string) error
	Unlisten(ctx context.Context, channel string) error
	WaitForNotification(ctx context.Context, timeout time.Duration) (*Notification, error)

	// Multi-tenancy support
	SetTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context) (string, error)

	// Hook management
	GetHookManager() IHookManager

	// Health check
	HealthCheck(ctx context.Context) error

	// Statistics
	Stats() ConnectionStats
}

// IPool representa um pool de conexões
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
	GetHookManager() IHookManager

	// Buffer pool management for optimization
	GetBufferPool() IBufferPool

	// Thread-safety monitoring
	GetSafetyMonitor() ISafetyMonitor
}

// ITransaction representa uma transação de banco de dados
type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// IRow representa um resultado de linha única
type IRow interface {
	Scan(dest ...any) error
}

// IRows representa um resultado de múltiplas linhas
type IRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
	CommandTag() ICommandTag
	FieldDescriptions() []IFieldDescription
	RawValues() [][]byte
}

// IBatch representa um lote de queries
type IBatch interface {
	Queue(query string, arguments ...any)
	QueueFunc(query string, arguments []any, callback func(IBatchResults) error)
	Len() int
	Clear()
	Reset()
}

// IBatchResults representa resultados de execução em lote
type IBatchResults interface {
	QueryRow() IRow
	Query() (IRows, error)
	Exec() (ICommandTag, error)
	Close() error
	Err() error
}

// ICommandTag representa metadados de resultado de comando
type ICommandTag interface {
	String() string
	RowsAffected() int64
	Insert() bool
	Update() bool
	Delete() bool
	Select() bool
}

// IFieldDescription representa metadados de campo
type IFieldDescription interface {
	Name() string
	TableOID() uint32
	TableAttributeNumber() uint16
	DataTypeOID() uint32
	DataTypeSize() int16
	TypeModifier() int32
	Format() int16
}

// ICopyFromSource representa uma fonte para operações COPY FROM
type ICopyFromSource interface {
	Next() bool
	Values() ([]interface{}, error)
	Err() error
}

// ICopyToWriter representa um writer para operações COPY TO
type ICopyToWriter interface {
	Write(row []interface{}) error
	Close() error
}

// IBufferPool gerencia buffers de memória para otimização
type IBufferPool interface {
	Get(size int) []byte
	Put(buf []byte)
	Stats() MemoryStats
	Reset()
}

// ISafetyMonitor fornece monitoramento de thread-safety
type ISafetyMonitor interface {
	CheckDeadlocks() []DeadlockInfo
	CheckRaceConditions() []RaceConditionInfo
	CheckLeaks() []LeakInfo
	IsHealthy() bool
}

// TxIsoLevel representa nível de isolamento de transação
type TxIsoLevel int

const (
	TxIsoLevelDefault TxIsoLevel = iota
	TxIsoLevelReadUncommitted
	TxIsoLevelReadCommitted
	TxIsoLevelRepeatableRead
	TxIsoLevelSerializable
)

// TxAccessMode representa modo de acesso de transação
type TxAccessMode int

const (
	TxAccessModeReadWrite TxAccessMode = iota
	TxAccessModeReadOnly
)

// TxDeferrableMode representa modo deferível de transação
type TxDeferrableMode int

const (
	TxDeferrableModeNotDeferrable TxDeferrableMode = iota
	TxDeferrableModeDeferrable
)

// TxOptions representa opções de transação
type TxOptions struct {
	IsoLevel       TxIsoLevel
	AccessMode     TxAccessMode
	DeferrableMode TxDeferrableMode
	BeginQuery     string
}

// Notification representa uma notificação PostgreSQL
type Notification struct {
	PID     uint32
	Channel string
	Payload string
}

// ConnectionStats representa estatísticas de conexão
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

// PoolStats representa estatísticas do pool
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

// MemoryStats representa estatísticas de uso de memória
type MemoryStats struct {
	BufferSize         int64
	AllocatedBuffers   int32
	PooledBuffers      int32
	TotalAllocations   int64
	TotalDeallocations int64
}

// DeadlockInfo representa informações de deadlock
type DeadlockInfo struct {
	Timestamp   time.Time
	Goroutines  []string
	StackTraces map[string]string
}

// RaceConditionInfo representa informações de race condition
type RaceConditionInfo struct {
	Timestamp time.Time
	Location  string
	Details   string
}

// LeakInfo representa informações de vazamento de recursos
type LeakInfo struct {
	Timestamp time.Time
	Resource  string
	Count     int64
	Details   string
}
