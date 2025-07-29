// Package interfaces define as interfaces principais do módulo Valkey.
// Estas interfaces garantem desacoplamento completo dos drivers e permitem
// implementação genérica com qualquer provider Valkey.
package interfaces

import (
	"context"
	"time"
)

// IClient define a interface principal para operações Valkey.
// Todos os providers devem implementar esta interface para garantir
// compatibilidade e intercambiabilidade.
type IClient interface {
	// String commands
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Hash commands
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)

	// List commands
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)

	// Set commands
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// Sorted Set commands
	ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZScore(ctx context.Context, key, member string) (float64, error)

	// Pipeline and Transaction
	Pipeline() IPipeline
	TxPipeline() ITransaction

	// Scripts
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error)
	ScriptLoad(ctx context.Context, script string) (string, error)

	// Pub/Sub
	Subscribe(ctx context.Context, channels ...string) (IPubSub, error)
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)

	// Streams
	XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error)
	XRead(ctx context.Context, streams map[string]string) ([]XMessage, error)
	XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]XMessage, error)

	// Scan
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error

	// Health check
	IsHealthy(ctx context.Context) bool
}

// IPipeline define a interface para operações pipeline.
// Permite execução em lote para otimização de performance.
type IPipeline interface {
	Get(key string) ICommand
	Set(key string, value interface{}, expiration time.Duration) ICommand
	Del(keys ...string) ICommand
	HGet(key, field string) ICommand
	HSet(key string, values ...interface{}) ICommand
	Exec(ctx context.Context) ([]interface{}, error)
	Discard() error
}

// ITransaction define a interface para transações.
// Garante atomicidade de operações usando MULTI/EXEC.
type ITransaction interface {
	Get(key string) ICommand
	Set(key string, value interface{}, expiration time.Duration) ICommand
	Del(keys ...string) ICommand
	HGet(key, field string) ICommand
	HSet(key string, values ...interface{}) ICommand
	Watch(ctx context.Context, keys ...string) error
	Unwatch(ctx context.Context) error
	Exec(ctx context.Context) ([]interface{}, error)
	Discard() error
}

// ICommand representa um comando individual em pipeline/transação.
type ICommand interface {
	Result() (interface{}, error)
	Err() error
	String() (string, error)
	Int64() (int64, error)
	Bool() (bool, error)
	Float64() (float64, error)
	Slice() ([]interface{}, error)
	StringSlice() ([]string, error)
	StringMap() (map[string]string, error)
}

// IPubSub define a interface para Pub/Sub operations.
type IPubSub interface {
	Subscribe(ctx context.Context, channels ...string) error
	Unsubscribe(ctx context.Context, channels ...string) error
	PSubscribe(ctx context.Context, patterns ...string) error
	PUnsubscribe(ctx context.Context, patterns ...string) error
	Receive(ctx context.Context) (interface{}, error)
	Close() error
}

// IScanner define a interface para operações de scan.
type IScanner interface {
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)
	SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)
	ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)
}

// IConn define a interface para conexão direta.
// Útil para comandos de baixo nível não cobertos pelas interfaces principais.
type IConn interface {
	Do(ctx context.Context, cmd string, args ...interface{}) (interface{}, error)
	Close() error
}

// XMessage representa uma mensagem em Valkey Streams.
type XMessage struct {
	ID     string
	Values map[string]interface{}
}

// Message representa uma mensagem de Pub/Sub.
type Message struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}

// PMessage representa uma mensagem de pattern subscription.
type PMessage struct {
	Channel string
	Pattern string
	Payload string
}

// Subscription representa uma subscription ativa.
type Subscription struct {
	Kind    string // "subscribe", "unsubscribe", "psubscribe", "punsubscribe"
	Channel string
	Count   int
}

// IProvider define a interface que todos os providers devem implementar.
// Esta interface é usada pelo factory pattern para criar clientes genéricos.
type IProvider interface {
	Name() string
	NewClient(config interface{}) (IClient, error)
	ValidateConfig(config interface{}) error
	DefaultConfig() interface{}
}

// IHealthChecker define interface para health checks.
type IHealthChecker interface {
	HealthCheck(ctx context.Context) error
	IsReady(ctx context.Context) bool
	IsLive(ctx context.Context) bool
}

// IMetrics define interface para coleta de métricas.
type IMetrics interface {
	IncrementCounter(name string, tags map[string]string)
	RecordDuration(name string, duration time.Duration, tags map[string]string)
	SetGauge(name string, value float64, tags map[string]string)
}

// IRetryPolicy define a interface para políticas de retry.
type IRetryPolicy interface {
	ShouldRetry(attempt int, err error) bool
	NextDelay(attempt int) time.Duration
}

// ICircuitBreaker define a interface para circuit breaker.
type ICircuitBreaker interface {
	Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)
	State() string
	Reset()
}
