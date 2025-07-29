// Package valkeygo implementa um provider para o driver valkey-go/v9.
// Este provider oferece implementação completa da interface IClient
// com suporte a todas as operações Valkey de forma desacoplada.
package valkeygo

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// Provider implementa interfaces.IProvider para o driver valkey-go.
type Provider struct{}

// NewProvider cria uma nova instância do provider Valkey-Go.
func NewProvider() interfaces.IProvider {
	return &Provider{}
}

// Name retorna o nome do provider.
func (p *Provider) Name() string {
	return "valkey-go"
}

// NewClient cria um novo cliente Valkey usando valkey-go.
func (p *Provider) NewClient(configInterface interface{}) (interfaces.IClient, error) {
	cfg, ok := configInterface.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("configuração deve ser do tipo *config.Config")
	}

	if err := p.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	// Criar cliente baseado no modo
	var valkeyClient valkey.Client
	var err error

	if cfg.ClusterMode {
		valkeyClient, err = p.createClusterClient(cfg)
	} else if cfg.SentinelMode {
		valkeyClient, err = p.createSentinelClient(cfg)
	} else {
		valkeyClient, err = p.createStandaloneClient(cfg)
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente valkey: %w", err)
	}

	client := &Client{
		client: valkeyClient,
		config: cfg,
	}

	return client, nil
}

// ValidateConfig valida a configuração para o provider valkey-go.
func (p *Provider) ValidateConfig(configInterface interface{}) error {
	cfg, ok := configInterface.(*config.Config)
	if !ok {
		return fmt.Errorf("configuração deve ser do tipo *config.Config")
	}

	return cfg.Validate()
}

// DefaultConfig retorna a configuração padrão para valkey-go.
func (p *Provider) DefaultConfig() interface{} {
	cfg := config.DefaultConfig()
	cfg.Provider = "valkey-go"
	return cfg
}

// createStandaloneClient cria um cliente standalone.
func (p *Provider) createStandaloneClient(cfg *config.Config) (valkey.Client, error) {
	var addr string
	if cfg.URI != "" {
		addr = cfg.URI
	} else {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	options := valkey.ClientOption{
		InitAddress: []string{addr},
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
		Dialer: net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: time.Second * 30,
		},
		ConnWriteTimeout: cfg.WriteTimeout,
	}

	// Configurar TLS se habilitado
	if cfg.TLSEnabled {
		options.TLSConfig = p.createTLSConfig(cfg)
	}

	client, err := valkey.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente standalone: %w", err)
	}

	return client, nil
}

// createClusterClient cria um cliente cluster.
func (p *Provider) createClusterClient(cfg *config.Config) (valkey.Client, error) {
	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("endereços do cluster não podem estar vazios")
	}

	options := valkey.ClientOption{
		InitAddress: cfg.Addrs,
		Password:    cfg.Password,
		Dialer: net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: time.Second * 30,
		},
		ConnWriteTimeout: cfg.WriteTimeout,
		ClusterOption: valkey.ClusterOption{
			ShardsRefreshInterval: time.Minute * 10, // Default refresh interval
		},
	}

	// Configurar TLS se habilitado
	if cfg.TLSEnabled {
		options.TLSConfig = p.createTLSConfig(cfg)
	}

	client, err := valkey.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente cluster: %w", err)
	}

	return client, nil
}

// createSentinelClient cria um cliente sentinel.
func (p *Provider) createSentinelClient(cfg *config.Config) (valkey.Client, error) {
	if len(cfg.SentinelAddrs) == 0 {
		return nil, fmt.Errorf("endereços do sentinel não podem estar vazios")
	}

	if cfg.SentinelMasterName == "" {
		return nil, fmt.Errorf("nome do master sentinel não pode estar vazio")
	}

	options := valkey.ClientOption{
		InitAddress: cfg.SentinelAddrs,
		Password:    cfg.Password,
		Dialer: net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: time.Second * 30,
		},
		ConnWriteTimeout: cfg.WriteTimeout,
		Sentinel: valkey.SentinelOption{
			MasterSet: cfg.SentinelMasterName,
			Password:  cfg.SentinelPassword,
		},
	}

	// Configurar TLS se habilitado
	if cfg.TLSEnabled {
		options.TLSConfig = p.createTLSConfig(cfg)
		options.Sentinel.TLSConfig = p.createTLSConfig(cfg)
	}

	client, err := valkey.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente sentinel: %w", err)
	}

	return client, nil
}

// createTLSConfig cria configuração TLS baseada na config.
func (p *Provider) createTLSConfig(cfg *config.Config) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
	}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return tlsConfig
}
