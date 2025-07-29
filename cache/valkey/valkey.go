// Package valkey fornece um cliente genérico para Valkey com suporte a múltiplos drivers.
// O módulo implementa o padrão Factory para permitir intercambiabilidade entre drivers
// mantendo uma interface consistente e desacoplada.
package valkey

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/hooks"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// Client representa o cliente principal do Valkey.
// Implementa a interface IClient e atua como wrapper para o provider específico.
type Client struct {
	provider       interfaces.IProvider
	client         interfaces.IClient
	config         *config.Config
	hooks          *hooks.CompositeHook
	healthChecker  interfaces.IHealthChecker
	retryPolicy    interfaces.IRetryPolicy
	circuitBreaker interfaces.ICircuitBreaker
	mu             sync.RWMutex
	closed         bool
}

// Manager gerencia múltiplas instâncias de clientes Valkey.
type Manager struct {
	providers map[string]interfaces.IProvider
	clients   map[string]*Client
	mu        sync.RWMutex
}

// NewManager cria um novo gerenciador de clientes Valkey.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]interfaces.IProvider),
		clients:   make(map[string]*Client),
	}
}

// RegisterProvider registra um provider no gerenciador.
func (m *Manager) RegisterProvider(name string, provider interfaces.IProvider) error {
	if name == "" {
		return fmt.Errorf("nome do provider não pode ser vazio")
	}
	if provider == nil {
		return fmt.Errorf("provider não pode ser nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	return nil
}

// NewClient cria um novo cliente Valkey usando a configuração especificada.
func (m *Manager) NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	m.mu.RLock()
	provider, exists := m.providers[cfg.Provider]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' não registrado", cfg.Provider)
	}

	client := &Client{
		provider: provider,
		config:   cfg.Copy(),
		hooks:    hooks.NewCompositeHook(),
	}

	// Criar cliente do provider
	providerClient, err := provider.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente do provider: %w", err)
	}

	client.client = providerClient

	// Configurar hooks padrão se habilitados
	if cfg.LogLevel != "silent" {
		loggingHook := hooks.NewLoggingHook()
		client.hooks.AddExecutionHook(loggingHook)
		client.hooks.AddConnectionHook(loggingHook)
		client.hooks.AddPipelineHook(loggingHook)
		client.hooks.AddRetryHook(loggingHook)
	}

	// Configurar métricas
	metricsHook := hooks.NewMetricsHook()
	client.hooks.AddExecutionHook(metricsHook)
	client.hooks.AddConnectionHook(metricsHook)
	client.hooks.AddPipelineHook(metricsHook)
	client.hooks.AddRetryHook(metricsHook)

	// Configurar retry policy
	if cfg.MaxRetries > 0 {
		client.retryPolicy = NewExponentialBackoffRetryPolicy(
			cfg.MaxRetries,
			cfg.MinRetryBackoff,
			cfg.MaxRetryBackoff,
		)
	}

	// Configurar circuit breaker
	if cfg.CircuitBreakerEnabled {
		client.circuitBreaker = NewCircuitBreaker(
			cfg.CircuitBreakerThreshold,
			cfg.CircuitBreakerTimeout,
			cfg.CircuitBreakerMaxRequests,
		)
	}

	return client, nil
}

// NewClientFromEnv cria um cliente usando configuração de variáveis de ambiente.
func (m *Manager) NewClientFromEnv() (*Client, error) {
	cfg := config.LoadFromEnv()
	return m.NewClient(cfg)
}

// GetClient retorna um cliente existente pelo nome ou cria um novo se não existir.
func (m *Manager) GetClient(name string, cfg *config.Config) (*Client, error) {
	m.mu.RLock()
	client, exists := m.clients[name]
	m.mu.RUnlock()

	if exists && !client.IsClosed() {
		return client, nil
	}

	// Criar novo cliente
	newClient, err := m.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.clients[name] = newClient
	m.mu.Unlock()

	return newClient, nil
}

// CloseAll fecha todos os clientes gerenciados.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao fechar cliente '%s': %w", name, err))
		}
	}

	m.clients = make(map[string]*Client)

	if len(errs) > 0 {
		return fmt.Errorf("erros ao fechar clientes: %v", errs)
	}

	return nil
}

// AddHook adiciona um hook ao cliente.
func (c *Client) AddHook(hook interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("cliente está fechado")
	}

	switch h := hook.(type) {
	case hooks.ExecutionHook:
		c.hooks.AddExecutionHook(h)
	case hooks.ConnectionHook:
		c.hooks.AddConnectionHook(h)
	case hooks.PipelineHook:
		c.hooks.AddPipelineHook(h)
	case hooks.RetryHook:
		c.hooks.AddRetryHook(h)
	default:
		return fmt.Errorf("tipo de hook não suportado: %T", hook)
	}

	return nil
}

// IsClosed verifica se o cliente está fechado.
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// executeWithHooks executa uma operação com hooks e retry.
func (c *Client) executeWithHooks(ctx context.Context, cmd string, args []interface{}, fn func() (interface{}, error)) (interface{}, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, fmt.Errorf("cliente está fechado")
	}
	hooks := c.hooks
	retryPolicy := c.retryPolicy
	circuitBreaker := c.circuitBreaker
	c.mu.RUnlock()

	// Aplicar circuit breaker se configurado
	if circuitBreaker != nil {
		return circuitBreaker.Execute(ctx, func() (interface{}, error) {
			return c.executeWithRetry(ctx, cmd, args, fn, hooks, retryPolicy)
		})
	}

	return c.executeWithRetry(ctx, cmd, args, fn, hooks, retryPolicy)
}

// executeWithRetry executa uma operação com retry e hooks.
func (c *Client) executeWithRetry(ctx context.Context, cmd string, args []interface{}, fn func() (interface{}, error), hooks *hooks.CompositeHook, retryPolicy interfaces.IRetryPolicy) (interface{}, error) {
	var lastErr error
	maxAttempts := 1

	if retryPolicy != nil {
		maxAttempts = 4 // MaxRetries + 1 tentativa inicial
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && retryPolicy != nil {
			if !retryPolicy.ShouldRetry(attempt, lastErr) {
				break
			}

			delay := retryPolicy.NextDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			ctx = hooks.BeforeRetry(ctx, attempt, lastErr)
		}

		// Before execution hook
		start := time.Now()
		ctx = hooks.BeforeExecution(ctx, cmd, args)

		// Execute operation
		result, err := fn()
		duration := time.Since(start)

		// After execution hook
		hooks.AfterExecution(ctx, cmd, args, result, err, duration)

		if err == nil {
			if attempt > 0 {
				hooks.AfterRetry(ctx, attempt, true, nil)
			}
			return result, nil
		}

		lastErr = err

		if attempt > 0 {
			hooks.AfterRetry(ctx, attempt, false, err)
		}

		// Se não há retry policy, falha imediatamente
		if retryPolicy == nil {
			break
		}
	}

	return nil, lastErr
}

// String commands implementation

// Get implementa IClient.Get.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.executeWithHooks(ctx, "GET", []interface{}{key}, func() (interface{}, error) {
		return c.client.Get(ctx, c.prefixKey(key))
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Set implementa IClient.Set.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := c.executeWithHooks(ctx, "SET", []interface{}{key, value, expiration}, func() (interface{}, error) {
		return nil, c.client.Set(ctx, c.prefixKey(key), value, expiration)
	})
	return err
}

// Del implementa IClient.Del.
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	result, err := c.executeWithHooks(ctx, "DEL", []interface{}{keys}, func() (interface{}, error) {
		return c.client.Del(ctx, prefixedKeys...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Exists implementa IClient.Exists.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	result, err := c.executeWithHooks(ctx, "EXISTS", []interface{}{keys}, func() (interface{}, error) {
		return c.client.Exists(ctx, prefixedKeys...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// TTL implementa IClient.TTL.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := c.executeWithHooks(ctx, "TTL", []interface{}{key}, func() (interface{}, error) {
		return c.client.TTL(ctx, c.prefixKey(key))
	})
	if err != nil {
		return 0, err
	}
	return result.(time.Duration), nil
}

// Expire implementa IClient.Expire.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := c.executeWithHooks(ctx, "EXPIRE", []interface{}{key, expiration}, func() (interface{}, error) {
		return nil, c.client.Expire(ctx, c.prefixKey(key), expiration)
	})
	return err
}

// Hash commands implementation

// HGet implementa IClient.HGet.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := c.executeWithHooks(ctx, "HGET", []interface{}{key, field}, func() (interface{}, error) {
		return c.client.HGet(ctx, c.prefixKey(key), field)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// HSet implementa IClient.HSet.
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	_, err := c.executeWithHooks(ctx, "HSET", append([]interface{}{key}, values...), func() (interface{}, error) {
		return nil, c.client.HSet(ctx, c.prefixKey(key), values...)
	})
	return err
}

// HDel implementa IClient.HDel.
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	result, err := c.executeWithHooks(ctx, "HDEL", append([]interface{}{key}, interfaceSlice(fields)...), func() (interface{}, error) {
		return c.client.HDel(ctx, c.prefixKey(key), fields...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// HExists implementa IClient.HExists.
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	result, err := c.executeWithHooks(ctx, "HEXISTS", []interface{}{key, field}, func() (interface{}, error) {
		return c.client.HExists(ctx, c.prefixKey(key), field)
	})
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}

// HGetAll implementa IClient.HGetAll.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := c.executeWithHooks(ctx, "HGETALL", []interface{}{key}, func() (interface{}, error) {
		return c.client.HGetAll(ctx, c.prefixKey(key))
	})
	if err != nil {
		return nil, err
	}
	return result.(map[string]string), nil
}

// prefixKey adiciona o prefixo configurado à chave.
func (c *Client) prefixKey(key string) string {
	if c.config.KeyPrefix == "" {
		return key
	}
	return c.config.KeyPrefix + key
}

// interfaceSlice converte []string para []interface{}.
func interfaceSlice(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// Implementação das demais interfaces será continuada nos próximos arquivos
// devido ao limite de tamanho. As implementações seguem o mesmo padrão com
// hooks, retry e circuit breaker.

// List commands (continuação das implementações)
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "LPUSH", append([]interface{}{key}, values...), func() (interface{}, error) {
		return c.client.LPush(ctx, c.prefixKey(key), values...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "RPUSH", append([]interface{}{key}, values...), func() (interface{}, error) {
		return c.client.RPush(ctx, c.prefixKey(key), values...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	result, err := c.executeWithHooks(ctx, "LPOP", []interface{}{key}, func() (interface{}, error) {
		return c.client.LPop(ctx, c.prefixKey(key))
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	result, err := c.executeWithHooks(ctx, "RPOP", []interface{}{key}, func() (interface{}, error) {
		return c.client.RPop(ctx, c.prefixKey(key))
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	result, err := c.executeWithHooks(ctx, "LLEN", []interface{}{key}, func() (interface{}, error) {
		return c.client.LLen(ctx, c.prefixKey(key))
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Set commands
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "SADD", append([]interface{}{key}, members...), func() (interface{}, error) {
		return c.client.SAdd(ctx, c.prefixKey(key), members...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "SREM", append([]interface{}{key}, members...), func() (interface{}, error) {
		return c.client.SRem(ctx, c.prefixKey(key), members...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	result, err := c.executeWithHooks(ctx, "SMEMBERS", []interface{}{key}, func() (interface{}, error) {
		return c.client.SMembers(ctx, c.prefixKey(key))
	})
	if err != nil {
		return nil, err
	}
	return result.([]string), nil
}

func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	result, err := c.executeWithHooks(ctx, "SISMEMBER", []interface{}{key, member}, func() (interface{}, error) {
		return c.client.SIsMember(ctx, c.prefixKey(key), member)
	})
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}

// Sorted Set commands
func (c *Client) ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "ZADD", append([]interface{}{key}, members...), func() (interface{}, error) {
		return c.client.ZAdd(ctx, c.prefixKey(key), members...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "ZREM", append([]interface{}{key}, members...), func() (interface{}, error) {
		return c.client.ZRem(ctx, c.prefixKey(key), members...)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result, err := c.executeWithHooks(ctx, "ZRANGE", []interface{}{key, start, stop}, func() (interface{}, error) {
		return c.client.ZRange(ctx, c.prefixKey(key), start, stop)
	})
	if err != nil {
		return nil, err
	}
	return result.([]string), nil
}

func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	result, err := c.executeWithHooks(ctx, "ZSCORE", []interface{}{key, member}, func() (interface{}, error) {
		return c.client.ZScore(ctx, c.prefixKey(key), member)
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// Pipeline and Transaction
func (c *Client) Pipeline() interfaces.IPipeline {
	return c.client.Pipeline()
}

func (c *Client) TxPipeline() interfaces.ITransaction {
	return c.client.TxPipeline()
}

// Scripts
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	return c.executeWithHooks(ctx, "EVAL", append([]interface{}{script}, append(interfaceSlice(keys), args...)...), func() (interface{}, error) {
		return c.client.Eval(ctx, script, prefixedKeys, args...)
	})
}

func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.prefixKey(key)
	}

	return c.executeWithHooks(ctx, "EVALSHA", append([]interface{}{sha1}, append(interfaceSlice(keys), args...)...), func() (interface{}, error) {
		return c.client.EvalSha(ctx, sha1, prefixedKeys, args...)
	})
}

func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	result, err := c.executeWithHooks(ctx, "SCRIPT LOAD", []interface{}{script}, func() (interface{}, error) {
		return c.client.ScriptLoad(ctx, script)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Pub/Sub
func (c *Client) Subscribe(ctx context.Context, channels ...string) (interfaces.IPubSub, error) {
	prefixedChannels := make([]string, len(channels))
	for i, channel := range channels {
		prefixedChannels[i] = c.prefixKey(channel)
	}

	result, err := c.executeWithHooks(ctx, "SUBSCRIBE", interfaceSlice(channels), func() (interface{}, error) {
		return c.client.Subscribe(ctx, prefixedChannels...)
	})
	if err != nil {
		return nil, err
	}
	return result.(interfaces.IPubSub), nil
}

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	result, err := c.executeWithHooks(ctx, "PUBLISH", []interface{}{channel, message}, func() (interface{}, error) {
		return c.client.Publish(ctx, c.prefixKey(channel), message)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Streams
func (c *Client) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	result, err := c.executeWithHooks(ctx, "XADD", []interface{}{stream, values}, func() (interface{}, error) {
		return c.client.XAdd(ctx, c.prefixKey(stream), values)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (c *Client) XRead(ctx context.Context, streams map[string]string) ([]interfaces.XMessage, error) {
	prefixedStreams := make(map[string]string)
	for stream, id := range streams {
		prefixedStreams[c.prefixKey(stream)] = id
	}

	result, err := c.executeWithHooks(ctx, "XREAD", []interface{}{streams}, func() (interface{}, error) {
		return c.client.XRead(ctx, prefixedStreams)
	})
	if err != nil {
		return nil, err
	}
	return result.([]interfaces.XMessage), nil
}

func (c *Client) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]interfaces.XMessage, error) {
	prefixedStreams := make(map[string]string)
	for stream, id := range streams {
		prefixedStreams[c.prefixKey(stream)] = id
	}

	result, err := c.executeWithHooks(ctx, "XREADGROUP", []interface{}{group, consumer, streams}, func() (interface{}, error) {
		return c.client.XReadGroup(ctx, group, consumer, prefixedStreams)
	})
	if err != nil {
		return nil, err
	}
	return result.([]interfaces.XMessage), nil
}

// Scan
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	prefixedMatch := match
	if match != "" && c.config.KeyPrefix != "" {
		prefixedMatch = c.config.KeyPrefix + match
	}

	keys, nextCursor, err := c.client.Scan(ctx, cursor, prefixedMatch, count)
	if err != nil {
		return nil, 0, err
	}

	// Remove prefix from returned keys
	if c.config.KeyPrefix != "" {
		unprefixedKeys := make([]string, len(keys))
		for i, key := range keys {
			if len(key) > len(c.config.KeyPrefix) && key[:len(c.config.KeyPrefix)] == c.config.KeyPrefix {
				unprefixedKeys[i] = key[len(c.config.KeyPrefix):]
			} else {
				unprefixedKeys[i] = key
			}
		}
		keys = unprefixedKeys
	}

	return keys, nextCursor, nil
}

func (c *Client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.client.HScan(ctx, c.prefixKey(key), cursor, match, count)
}

// Connection management
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.executeWithHooks(ctx, "PING", []interface{}{}, func() (interface{}, error) {
		return nil, c.client.Ping(ctx)
	})
	return err
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.client.Close()
}

// Health check
func (c *Client) IsHealthy(ctx context.Context) bool {
	if c.IsClosed() {
		return false
	}
	return c.client.IsHealthy(ctx)
}

// DefaultManager é a instância global padrão do gerenciador.
var DefaultManager = NewManager()

// NewClient cria um cliente usando o gerenciador padrão.
func NewClient(cfg *config.Config) (*Client, error) {
	return DefaultManager.NewClient(cfg)
}

// NewClientFromEnv cria um cliente usando configuração de ambiente com o gerenciador padrão.
func NewClientFromEnv() (*Client, error) {
	return DefaultManager.NewClientFromEnv()
}
