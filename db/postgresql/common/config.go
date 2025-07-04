package common

import (
	"fmt"
	"time"
)

// Config representa as configurações comuns para conexões PostgreSQL
type Config struct {
	// Configurações de conexão
	Host           string
	Port           int
	Database       string
	User           string
	Password       string
	SSLMode        string
	ConnectTimeout int // em segundos

	// Configurações de pool
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	// Recursos adicionais
	TraceEnabled       bool
	QueryLogEnabled    bool
	MultiTenantEnabled bool
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		SSLMode:         "disable",
		ConnectTimeout:  10,
		MaxConns:        20,
		MinConns:        2,
		MaxConnLifetime: time.Minute * 30,
		MaxConnIdleTime: time.Minute * 10,
		TraceEnabled:    false,
		QueryLogEnabled: false,
	}
}

// ConnectionString gera uma string de conexão a partir da configuração
func (c *Config) ConnectionString() string {
	connString := "host=" + c.Host + " port=" + IntToStr(c.Port) + " dbname=" + c.Database

	if c.User != "" {
		connString += " user=" + c.User
	}

	if c.Password != "" {
		connString += " password=" + c.Password
	}

	if c.SSLMode != "" {
		connString += " sslmode=" + c.SSLMode
	}

	if c.ConnectTimeout > 0 {
		connString += " connect_timeout=" + IntToStr(c.ConnectTimeout)
	}

	return connString
}

// IntToStr converte um inteiro para string
func IntToStr(value int) string {
	return fmt.Sprintf("%d", value)
}

// Option é um tipo de função para configuração fluente
type Option func(*Config)

// WithHost define o host do banco de dados
func WithHost(host string) Option {
	return func(c *Config) {
		if host != "" {
			c.Host = host
		}
	}
}

// WithPort define a porta do banco de dados
func WithPort(port int) Option {
	return func(c *Config) {
		if port > 0 {
			c.Port = port
		}
	}
}

// WithDatabase define o nome do banco de dados
func WithDatabase(database string) Option {
	return func(c *Config) {
		if database != "" {
			c.Database = database
		}
	}
}

// WithUser define o usuário para conexão
func WithUser(user string) Option {
	return func(c *Config) {
		if user != "" {
			c.User = user
		}
	}
}

// WithPassword define a senha para conexão
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithSSLMode define o modo SSL
func WithSSLMode(sslMode string) Option {
	return func(c *Config) {
		if sslMode != "" {
			c.SSLMode = sslMode
		}
	}
}

// WithMaxConns define o número máximo de conexões no pool
func WithMaxConns(maxConns int32) Option {
	return func(c *Config) {
		if maxConns > 0 {
			c.MaxConns = maxConns
		}
	}
}

// WithMinConns define o número mínimo de conexões no pool
func WithMinConns(minConns int32) Option {
	return func(c *Config) {
		if minConns > 0 {
			c.MinConns = minConns
		}
	}
}

// WithMaxConnLifetime define o tempo máximo de vida de uma conexão
func WithMaxConnLifetime(lifetime time.Duration) Option {
	return func(c *Config) {
		if lifetime > 0 {
			c.MaxConnLifetime = lifetime
		}
	}
}

// WithMaxConnIdleTime define o tempo máximo de ociosidade de uma conexão
func WithMaxConnIdleTime(idleTime time.Duration) Option {
	return func(c *Config) {
		if idleTime > 0 {
			c.MaxConnIdleTime = idleTime
		}
	}
}

// WithTraceEnabled habilita o rastreamento de queries
func WithTraceEnabled(enabled bool) Option {
	return func(c *Config) {
		c.TraceEnabled = enabled
	}
}

// WithQueryLogEnabled habilita o log de queries
func WithQueryLogEnabled(enabled bool) Option {
	return func(c *Config) {
		c.QueryLogEnabled = enabled
	}
}

// WithMultiTenantEnabled habilita o suporte a múltiplos tenants
func WithMultiTenantEnabled(enabled bool) Option {
	return func(c *Config) {
		c.MultiTenantEnabled = enabled
	}
}
