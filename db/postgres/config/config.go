package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// DefaultConfig implementa IConfig com otimizações de memória
type DefaultConfig struct {
	// Usar campos privados para controle de acesso
	connectionString   string
	poolConfig         interfaces.PoolConfig
	tlsConfig          interfaces.TLSConfig
	retryConfig        interfaces.RetryConfig
	hookConfig         interfaces.HookConfig
	failoverConfig     interfaces.FailoverConfig
	readReplicaConfig  interfaces.ReadReplicaConfig
	multiTenantEnabled bool

	// Mutex para thread-safety
	mu sync.RWMutex

	// Cache para validação - otimização de memória
	validationCache map[string]bool
	cacheMu         sync.RWMutex
}

// NewDefaultConfig cria uma nova configuração padrão com otimizações
func NewDefaultConfig(connectionString string) interfaces.IConfig {
	return &DefaultConfig{
		connectionString: connectionString,
		poolConfig: interfaces.PoolConfig{
			MaxConns:          30,
			MinConns:          5,
			MaxConnLifetime:   time.Hour,
			MaxConnIdleTime:   time.Minute * 30,
			HealthCheckPeriod: time.Minute * 5,
			ConnectTimeout:    time.Second * 30,
			LazyConnect:       true,
		},
		tlsConfig: interfaces.TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: false,
		},
		retryConfig: interfaces.RetryConfig{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			MaxInterval:     time.Second * 5,
			Multiplier:      2.0,
			RandomizeWait:   true,
		},
		hookConfig: interfaces.HookConfig{
			EnabledHooks: []interfaces.HookType{},
			CustomHooks:  make(map[string]interfaces.HookType),
			HookTimeout:  time.Second * 5,
		},
		failoverConfig: interfaces.FailoverConfig{
			Enabled:             false,
			FallbackNodes:       []string{},
			HealthCheckInterval: time.Second * 30,
			RetryInterval:       time.Second * 5,
			MaxFailoverAttempts: 3,
		},
		readReplicaConfig: interfaces.ReadReplicaConfig{
			Enabled:             false,
			ConnectionStrings:   []string{},
			LoadBalanceMode:     interfaces.LoadBalanceModeRoundRobin,
			HealthCheckInterval: time.Second * 30,
		},
		multiTenantEnabled: false,
		validationCache:    make(map[string]bool),
	}
}

// GetConnectionString retorna a string de conexão
func (c *DefaultConfig) GetConnectionString() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connectionString
}

// GetPoolConfig retorna configuração do pool
func (c *DefaultConfig) GetPoolConfig() interfaces.PoolConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.poolConfig
}

// GetTLSConfig retorna configuração TLS
func (c *DefaultConfig) GetTLSConfig() interfaces.TLSConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tlsConfig
}

// GetRetryConfig retorna configuração de retry
func (c *DefaultConfig) GetRetryConfig() interfaces.RetryConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.retryConfig
}

// GetHookConfig retorna configuração de hooks
func (c *DefaultConfig) GetHookConfig() interfaces.HookConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hookConfig
}

// GetFailoverConfig retorna configuração de failover
func (c *DefaultConfig) GetFailoverConfig() interfaces.FailoverConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.failoverConfig
}

// GetReadReplicaConfig retorna configuração de read replica
func (c *DefaultConfig) GetReadReplicaConfig() interfaces.ReadReplicaConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.readReplicaConfig
}

// IsMultiTenantEnabled retorna se multi-tenant está habilitado
func (c *DefaultConfig) IsMultiTenantEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.multiTenantEnabled
}

// Validate valida a configuração com cache para otimização
func (c *DefaultConfig) Validate() error {
	// Criar uma chave de cache baseada no estado da config
	cacheKey := c.getCacheKey()

	// Verificar cache primeiro
	c.cacheMu.RLock()
	if cached, exists := c.validationCache[cacheKey]; exists {
		c.cacheMu.RUnlock()
		if cached {
			return nil
		}
		return fmt.Errorf("configuration validation failed (cached)")
	}
	c.cacheMu.RUnlock()

	// Validar configuração
	if err := c.validateInternal(); err != nil {
		c.cacheMu.Lock()
		c.validationCache[cacheKey] = false
		c.cacheMu.Unlock()
		return err
	}

	// Armazenar resultado válido no cache
	c.cacheMu.Lock()
	c.validationCache[cacheKey] = true
	c.cacheMu.Unlock()

	return nil
}

// validateInternal executa validação interna
func (c *DefaultConfig) validateInternal() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.connectionString == "" {
		return fmt.Errorf("connection string is required")
	}

	if c.poolConfig.MaxConns <= 0 {
		return fmt.Errorf("max connections must be greater than 0")
	}

	if c.poolConfig.MinConns < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}

	if c.poolConfig.MinConns > c.poolConfig.MaxConns {
		return fmt.Errorf("min connections cannot be greater than max connections")
	}

	if c.retryConfig.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.retryConfig.InitialInterval <= 0 {
		return fmt.Errorf("initial interval must be positive")
	}

	if c.retryConfig.MaxInterval <= 0 {
		return fmt.Errorf("max interval must be positive")
	}

	if c.retryConfig.Multiplier <= 1.0 {
		return fmt.Errorf("multiplier must be greater than 1.0")
	}

	return nil
}

// getCacheKey gera uma chave de cache baseada no estado da configuração
func (c *DefaultConfig) getCacheKey() string {
	return fmt.Sprintf("%s_%d_%d_%v_%v",
		c.connectionString,
		c.poolConfig.MaxConns,
		c.poolConfig.MinConns,
		c.retryConfig.MaxRetries,
		c.multiTenantEnabled,
	)
}

// ConfigOption representa uma opção de configuração
type ConfigOption func(*DefaultConfig)

// WithConnectionString define a string de conexão
func WithConnectionString(connectionString string) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.connectionString = connectionString
		c.clearCache()
	}
}

// WithMaxConns define o número máximo de conexões
func WithMaxConns(maxConns int32) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.poolConfig.MaxConns = maxConns
		c.clearCache()
	}
}

// WithMinConns define o número mínimo de conexões
func WithMinConns(minConns int32) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.poolConfig.MinConns = minConns
		c.clearCache()
	}
}

// WithMaxConnLifetime define o tempo de vida máximo da conexão
func WithMaxConnLifetime(lifetime time.Duration) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.poolConfig.MaxConnLifetime = lifetime
		c.clearCache()
	}
}

// WithMaxConnIdleTime define o tempo máximo de idle da conexão
func WithMaxConnIdleTime(idleTime time.Duration) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.poolConfig.MaxConnIdleTime = idleTime
		c.clearCache()
	}
}

// WithMultiTenant habilita/desabilita multi-tenant
func WithMultiTenant(enabled bool) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.multiTenantEnabled = enabled
		c.clearCache()
	}
}

// WithTLS configura TLS
func WithTLS(enabled bool, insecureSkipVerify bool) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.tlsConfig.Enabled = enabled
		c.tlsConfig.InsecureSkipVerify = insecureSkipVerify
		c.clearCache()
	}
}

// WithRetry configura retry
func WithRetry(maxRetries int, initialInterval, maxInterval time.Duration, multiplier float64) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.retryConfig.MaxRetries = maxRetries
		c.retryConfig.InitialInterval = initialInterval
		c.retryConfig.MaxInterval = maxInterval
		c.retryConfig.Multiplier = multiplier
		c.clearCache()
	}
}

// WithFailover configura failover
func WithFailover(enabled bool, fallbackNodes []string) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.failoverConfig.Enabled = enabled
		c.failoverConfig.FallbackNodes = make([]string, len(fallbackNodes))
		copy(c.failoverConfig.FallbackNodes, fallbackNodes)
		c.clearCache()
	}
}

// WithReadReplicas configura read replicas
func WithReadReplicas(enabled bool, connectionStrings []string, loadBalanceMode interfaces.LoadBalanceMode) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.readReplicaConfig.Enabled = enabled
		c.readReplicaConfig.ConnectionStrings = make([]string, len(connectionStrings))
		copy(c.readReplicaConfig.ConnectionStrings, connectionStrings)
		c.readReplicaConfig.LoadBalanceMode = loadBalanceMode
		c.clearCache()
	}
}

// WithEnabledHooks configura hooks habilitados
func WithEnabledHooks(hooks []interfaces.HookType) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.hookConfig.EnabledHooks = make([]interfaces.HookType, len(hooks))
		copy(c.hookConfig.EnabledHooks, hooks)
		c.clearCache()
	}
}

// WithCustomHook adiciona um hook customizado
func WithCustomHook(name string, hookType interfaces.HookType) ConfigOption {
	return func(c *DefaultConfig) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.hookConfig.CustomHooks == nil {
			c.hookConfig.CustomHooks = make(map[string]interfaces.HookType)
		}
		c.hookConfig.CustomHooks[name] = hookType
		c.clearCache()
	}
}

// clearCache limpa o cache de validação - deve ser chamado com lock
func (c *DefaultConfig) clearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.validationCache = make(map[string]bool)
}

// Apply aplica opções de configuração
func (c *DefaultConfig) Apply(options ...ConfigOption) {
	for _, option := range options {
		option(c)
	}
}
