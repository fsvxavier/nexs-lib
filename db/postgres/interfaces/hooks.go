package interfaces

import (
	"context"
	"time"
)

// HookType representa o tipo de hook
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
	BeforeCommitHook
	AfterCommitHook
	BeforeRollbackHook
	AfterRollbackHook
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

// HookResult representa o resultado de uma execução de hook
type HookResult struct {
	Continue bool
	Error    error
	Data     map[string]interface{}
}

// ExecutionContext contém informações de contexto para hooks
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

// Hook representa uma função que pode ser executada em pontos específicos
type Hook func(ctx *ExecutionContext) *HookResult

// IHookManager gerencia registro e execução de hooks
type IHookManager interface {
	// RegisterHook registra um hook para um tipo específico
	RegisterHook(hookType HookType, hook Hook) error

	// RegisterCustomHook registra um hook customizado com um tipo customizado
	RegisterCustomHook(hookType HookType, name string, hook Hook) error

	// ExecuteHooks executa todos os hooks de um tipo específico
	ExecuteHooks(hookType HookType, ctx *ExecutionContext) error

	// UnregisterHook remove um hook
	UnregisterHook(hookType HookType) error

	// UnregisterCustomHook remove um hook customizado
	UnregisterCustomHook(hookType HookType, name string) error

	// ListHooks retorna todos os hooks registrados
	ListHooks() map[HookType][]Hook
}

// IRetryManager gerencia operações de retry
type IRetryManager interface {
	Execute(ctx context.Context, operation func() error) error
	ExecuteWithConn(ctx context.Context, pool IPool, operation func(conn IConn) error) error
	UpdateConfig(config RetryConfig) error
	GetStats() RetryStats
}

// IFailoverManager gerencia operações de failover
type IFailoverManager interface {
	Execute(ctx context.Context, operation func(conn IConn) error) error
	MarkNodeDown(nodeID string) error
	MarkNodeUp(nodeID string) error
	GetHealthyNodes() []string
	GetUnhealthyNodes() []string
	GetStats() FailoverStats
}

// RetryStats representa estatísticas de operações de retry
type RetryStats struct {
	TotalAttempts  int64
	SuccessfulOps  int64
	FailedOps      int64
	TotalRetries   int64
	AverageRetries float64
	LastRetryTime  time.Time
}

// FailoverStats representa estatísticas de operações de failover
type FailoverStats struct {
	TotalFailovers      int64
	SuccessfulFailovers int64
	FailedFailovers     int64
	CurrentActiveNode   string
	DownNodes           []string
	LastFailoverTime    time.Time
}
