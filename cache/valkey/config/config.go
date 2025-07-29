// Package config define as configurações para o módulo Valkey.
// Suporte a configuração via environment variables e struct com validação.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config define a configuração principal do Valkey.
type Config struct {
	// Provider especifica qual driver usar: "valkey-go" ou "valkey-glide"
	Provider string `json:"provider" yaml:"provider"`

	// Connection settings
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`

	// URI connection (alternativa a host/port)
	URI string `json:"uri" yaml:"uri"`

	// Pool settings
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxIdleConns int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxAge   time.Duration `json:"conn_max_age" yaml:"conn_max_age"`
	PoolTimeout  time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`

	// Timeouts
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// Retry settings
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	MinRetryBackoff time.Duration `json:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff" yaml:"max_retry_backoff"`

	// Cluster settings
	ClusterMode bool     `json:"cluster_mode" yaml:"cluster_mode"`
	Addrs       []string `json:"addrs" yaml:"addrs"`

	// Sentinel settings
	SentinelMode       bool     `json:"sentinel_mode" yaml:"sentinel_mode"`
	SentinelAddrs      []string `json:"sentinel_addrs" yaml:"sentinel_addrs"`
	SentinelMasterName string   `json:"sentinel_master_name" yaml:"sentinel_master_name"`
	SentinelPassword   string   `json:"sentinel_password" yaml:"sentinel_password"`

	// Multi-tenancy
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`

	// TLS settings
	TLSEnabled            bool   `json:"tls_enabled" yaml:"tls_enabled"`
	TLSCertFile           string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile            string `json:"tls_key_file" yaml:"tls_key_file"`
	TLSCACertFile         string `json:"tls_ca_cert_file" yaml:"tls_ca_cert_file"`
	TLSInsecureSkipVerify bool   `json:"tls_insecure_skip_verify" yaml:"tls_insecure_skip_verify"`

	// Circuit breaker settings
	CircuitBreakerEnabled     bool          `json:"circuit_breaker_enabled" yaml:"circuit_breaker_enabled"`
	CircuitBreakerThreshold   int           `json:"circuit_breaker_threshold" yaml:"circuit_breaker_threshold"`
	CircuitBreakerTimeout     time.Duration `json:"circuit_breaker_timeout" yaml:"circuit_breaker_timeout"`
	CircuitBreakerMaxRequests int           `json:"circuit_breaker_max_requests" yaml:"circuit_breaker_max_requests"`

	// Health check settings
	HealthCheckEnabled  bool          `json:"health_check_enabled" yaml:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout" yaml:"health_check_timeout"`

	// Logging
	LogLevel string `json:"log_level" yaml:"log_level"`
}

// DefaultConfig retorna uma configuração padrão.
func DefaultConfig() *Config {
	return &Config{
		Provider: "valkey-go",
		Host:     "localhost",
		Port:     6379,
		DB:       0,

		// Pool defaults
		PoolSize:     10,
		MinIdleConns: 1,
		MaxIdleConns: 3,
		ConnMaxAge:   30 * time.Minute,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,

		// Timeout defaults
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Retry defaults
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// Circuit breaker defaults
		CircuitBreakerEnabled:     true,
		CircuitBreakerThreshold:   5,
		CircuitBreakerTimeout:     60 * time.Second,
		CircuitBreakerMaxRequests: 3,

		// Health check defaults
		HealthCheckEnabled:  true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,

		LogLevel: "info",
	}
}

// LoadFromEnv carrega configuração de variáveis de ambiente.
func LoadFromEnv() *Config {
	config := DefaultConfig()

	// Provider
	if provider := os.Getenv("VALKEY_PROVIDER"); provider != "" {
		config.Provider = provider
	}

	// Connection
	if host := os.Getenv("VALKEY_HOST"); host != "" {
		config.Host = host
	}
	if port := os.Getenv("VALKEY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}
	if password := os.Getenv("VALKEY_PASSWORD"); password != "" {
		config.Password = password
	}
	if db := os.Getenv("VALKEY_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.DB = d
		}
	}
	if uri := os.Getenv("VALKEY_URI"); uri != "" {
		config.URI = uri
	}

	// Pool settings
	if poolSize := os.Getenv("VALKEY_POOL_SIZE"); poolSize != "" {
		if p, err := strconv.Atoi(poolSize); err == nil {
			config.PoolSize = p
		}
	}
	if minIdle := os.Getenv("VALKEY_MIN_IDLE_CONNS"); minIdle != "" {
		if m, err := strconv.Atoi(minIdle); err == nil {
			config.MinIdleConns = m
		}
	}
	if maxIdle := os.Getenv("VALKEY_MAX_IDLE_CONNS"); maxIdle != "" {
		if m, err := strconv.Atoi(maxIdle); err == nil {
			config.MaxIdleConns = m
		}
	}

	// Timeouts
	if dialTimeout := os.Getenv("VALKEY_DIAL_TIMEOUT"); dialTimeout != "" {
		if d, err := time.ParseDuration(dialTimeout); err == nil {
			config.DialTimeout = d
		}
	}
	if readTimeout := os.Getenv("VALKEY_READ_TIMEOUT"); readTimeout != "" {
		if r, err := time.ParseDuration(readTimeout); err == nil {
			config.ReadTimeout = r
		}
	}
	if writeTimeout := os.Getenv("VALKEY_WRITE_TIMEOUT"); writeTimeout != "" {
		if w, err := time.ParseDuration(writeTimeout); err == nil {
			config.WriteTimeout = w
		}
	}

	// Cluster mode
	if clusterMode := os.Getenv("VALKEY_CLUSTER_MODE"); clusterMode != "" {
		config.ClusterMode = strings.ToLower(clusterMode) == "true"
	}
	if addrs := os.Getenv("VALKEY_CLUSTER_ADDRS"); addrs != "" {
		config.Addrs = strings.Split(addrs, ",")
	}

	// Sentinel mode
	if sentinelMode := os.Getenv("VALKEY_SENTINEL_MODE"); sentinelMode != "" {
		config.SentinelMode = strings.ToLower(sentinelMode) == "true"
	}
	if sentinelAddrs := os.Getenv("VALKEY_SENTINEL_ADDRS"); sentinelAddrs != "" {
		config.SentinelAddrs = strings.Split(sentinelAddrs, ",")
	}
	if masterName := os.Getenv("VALKEY_SENTINEL_MASTER_NAME"); masterName != "" {
		config.SentinelMasterName = masterName
	}
	if sentinelPassword := os.Getenv("VALKEY_SENTINEL_PASSWORD"); sentinelPassword != "" {
		config.SentinelPassword = sentinelPassword
	}

	// Multi-tenancy
	if keyPrefix := os.Getenv("VALKEY_KEY_PREFIX"); keyPrefix != "" {
		config.KeyPrefix = keyPrefix
	}

	// TLS
	if tlsEnabled := os.Getenv("VALKEY_TLS_ENABLED"); tlsEnabled != "" {
		config.TLSEnabled = strings.ToLower(tlsEnabled) == "true"
	}
	if certFile := os.Getenv("VALKEY_TLS_CERT_FILE"); certFile != "" {
		config.TLSCertFile = certFile
	}
	if keyFile := os.Getenv("VALKEY_TLS_KEY_FILE"); keyFile != "" {
		config.TLSKeyFile = keyFile
	}
	if caFile := os.Getenv("VALKEY_TLS_CA_CERT_FILE"); caFile != "" {
		config.TLSCACertFile = caFile
	}
	if skipVerify := os.Getenv("VALKEY_TLS_INSECURE_SKIP_VERIFY"); skipVerify != "" {
		config.TLSInsecureSkipVerify = strings.ToLower(skipVerify) == "true"
	}

	// Log level
	if logLevel := os.Getenv("VALKEY_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	return config
}

// Validate verifica se a configuração é válida.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider não pode ser vazio")
	}

	if c.Provider != "valkey-go" && c.Provider != "valkey-glide" {
		return fmt.Errorf("provider deve ser 'valkey-go' ou 'valkey-glide', recebido: %s", c.Provider)
	}

	// Validar conexão
	if c.URI == "" {
		if c.Host == "" {
			return fmt.Errorf("host não pode ser vazio quando URI não está especificado")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return fmt.Errorf("port deve estar entre 1 e 65535, recebido: %d", c.Port)
		}
	}

	// Validar cluster mode
	if c.ClusterMode && len(c.Addrs) == 0 {
		return fmt.Errorf("addrs não pode ser vazio no modo cluster")
	}

	// Validar sentinel mode
	if c.SentinelMode {
		if len(c.SentinelAddrs) == 0 {
			return fmt.Errorf("sentinel_addrs não pode ser vazio no modo sentinel")
		}
		if c.SentinelMasterName == "" {
			return fmt.Errorf("sentinel_master_name não pode ser vazio no modo sentinel")
		}
	}

	// Validar pool settings
	if c.PoolSize <= 0 {
		return fmt.Errorf("pool_size deve ser maior que 0, recebido: %d", c.PoolSize)
	}
	if c.MinIdleConns < 0 {
		return fmt.Errorf("min_idle_conns não pode ser negativo, recebido: %d", c.MinIdleConns)
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns não pode ser negativo, recebido: %d", c.MaxIdleConns)
	}

	// Validar timeouts
	if c.DialTimeout <= 0 {
		return fmt.Errorf("dial_timeout deve ser maior que 0")
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout deve ser maior que 0")
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout deve ser maior que 0")
	}

	// Validar retry settings
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries não pode ser negativo, recebido: %d", c.MaxRetries)
	}

	// Validar TLS
	if c.TLSEnabled {
		if c.TLSCertFile != "" && c.TLSKeyFile == "" {
			return fmt.Errorf("tls_key_file deve ser especificado quando tls_cert_file está definido")
		}
		if c.TLSKeyFile != "" && c.TLSCertFile == "" {
			return fmt.Errorf("tls_cert_file deve ser especificado quando tls_key_file está definido")
		}
	}

	// Validar circuit breaker
	if c.CircuitBreakerEnabled {
		if c.CircuitBreakerThreshold <= 0 {
			return fmt.Errorf("circuit_breaker_threshold deve ser maior que 0")
		}
		if c.CircuitBreakerTimeout <= 0 {
			return fmt.Errorf("circuit_breaker_timeout deve ser maior que 0")
		}
		if c.CircuitBreakerMaxRequests <= 0 {
			return fmt.Errorf("circuit_breaker_max_requests deve ser maior que 0")
		}
	}

	return nil
}

// ConnectionString retorna a string de conexão baseada na configuração.
func (c *Config) ConnectionString() string {
	if c.URI != "" {
		return c.URI
	}

	if c.Password != "" {
		return fmt.Sprintf("valkey://:%s@%s:%d/%d", c.Password, c.Host, c.Port, c.DB)
	}

	return fmt.Sprintf("valkey://%s:%d/%d", c.Host, c.Port, c.DB)
}

// Copy cria uma cópia profunda da configuração.
func (c *Config) Copy() *Config {
	copy := *c

	// Copiar slices
	if len(c.Addrs) > 0 {
		copy.Addrs = make([]string, len(c.Addrs))
		for i, addr := range c.Addrs {
			copy.Addrs[i] = addr
		}
	}

	if len(c.SentinelAddrs) > 0 {
		copy.SentinelAddrs = make([]string, len(c.SentinelAddrs))
		for i, addr := range c.SentinelAddrs {
			copy.SentinelAddrs[i] = addr
		}
	}

	return &copy
}

// WithProvider define o provider.
func (c *Config) WithProvider(provider string) *Config {
	c.Provider = provider
	return c
}

// WithHost define o host.
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort define a porta.
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

// WithPassword define a senha.
func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

// WithDB define o banco de dados.
func (c *Config) WithDB(db int) *Config {
	c.DB = db
	return c
}

// WithURI define a URI de conexão.
func (c *Config) WithURI(uri string) *Config {
	c.URI = uri
	return c
}

// WithPoolSize define o tamanho do pool.
func (c *Config) WithPoolSize(size int) *Config {
	c.PoolSize = size
	return c
}

// WithKeyPrefix define o prefixo das chaves.
func (c *Config) WithKeyPrefix(prefix string) *Config {
	c.KeyPrefix = prefix
	return c
}

// WithClusterMode habilita/desabilita o modo cluster.
func (c *Config) WithClusterMode(enabled bool) *Config {
	c.ClusterMode = enabled
	return c
}

// WithClusterAddrs define os endereços do cluster.
func (c *Config) WithClusterAddrs(addrs []string) *Config {
	c.Addrs = addrs
	return c
}

// WithSentinelMode habilita/desabilita o modo sentinel.
func (c *Config) WithSentinelMode(enabled bool) *Config {
	c.SentinelMode = enabled
	return c
}

// WithSentinelAddrs define os endereços do sentinel.
func (c *Config) WithSentinelAddrs(addrs []string) *Config {
	c.SentinelAddrs = addrs
	return c
}

// WithSentinelMasterName define o nome do master no sentinel.
func (c *Config) WithSentinelMasterName(name string) *Config {
	c.SentinelMasterName = name
	return c
}

// WithTLS habilita/desabilita TLS.
func (c *Config) WithTLS(enabled bool) *Config {
	c.TLSEnabled = enabled
	return c
}

// WithCircuitBreaker habilita/desabilita circuit breaker.
func (c *Config) WithCircuitBreaker(enabled bool) *Config {
	c.CircuitBreakerEnabled = enabled
	return c
}

// WithHealthCheck habilita/desabilita health check.
func (c *Config) WithHealthCheck(enabled bool) *Config {
	c.HealthCheckEnabled = enabled
	return c
}
