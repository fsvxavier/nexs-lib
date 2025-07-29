// Package valkeyglide implementa um provider para o driver valkey-glide.
// Este provider oferece implementação completa da interface IClient
// com suporte a todas as operações Valkey de forma desacoplada.
package valkeyglide

import (
	"crypto/tls"
	"fmt"
	"net"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// Provider implementa interfaces.IProvider para o driver valkey-glide.
type Provider struct{}

// NewProvider cria uma nova instância do provider Valkey-Glide.
func NewProvider() interfaces.IProvider {
	return &Provider{}
}

// Name retorna o nome do provider.
func (p *Provider) Name() string {
	return "valkey-glide"
}

// NewClient cria um novo cliente Valkey usando valkey-glide.
func (p *Provider) NewClient(configInterface interface{}) (interfaces.IClient, error) {
	cfg, ok := configInterface.(*valkeyconfig.Config)
	if !ok {
		return nil, fmt.Errorf("configuração deve ser do tipo *config.Config")
	}

	if err := p.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	// Criar cliente baseado no modo
	if cfg.ClusterMode {
		clusterClient, err := p.createClusterClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar cliente cluster: %w", err)
		}

		client := &ClusterClient{
			client: clusterClient,
			config: cfg,
		}

		return client, nil
	} else {
		standaloneClient, err := p.createStandaloneClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar cliente standalone: %w", err)
		}

		client := &Client{
			client: standaloneClient,
			config: cfg,
		}

		return client, nil
	}
}

// ValidateConfig valida a configuração para o provider valkey-glide.
func (p *Provider) ValidateConfig(configInterface interface{}) error {
	cfg, ok := configInterface.(*valkeyconfig.Config)
	if !ok {
		return fmt.Errorf("configuração deve ser do tipo *config.Config")
	}

	return cfg.Validate()
}

// DefaultConfig retorna a configuração padrão para valkey-glide.
func (p *Provider) DefaultConfig() interface{} {
	cfg := valkeyconfig.DefaultConfig()
	cfg.Provider = "valkey-glide"
	return cfg
}

// createStandaloneClient cria um cliente standalone.
func (p *Provider) createStandaloneClient(cfg *valkeyconfig.Config) (*glide.Client, error) {
	clientConfig := config.NewClientConfiguration()

	// Configurar endereços
	if cfg.URI != "" {
		// Parse URI para host e port
		// Implementação simplificada
		clientConfig.WithAddress(&config.NodeAddress{
			Host: cfg.Host,
			Port: cfg.Port,
		})
	} else {
		clientConfig.WithAddress(&config.NodeAddress{
			Host: cfg.Host,
			Port: cfg.Port,
		})
	}

	// Configurar credenciais
	if cfg.Password != "" {
		credentials := config.NewServerCredentialsWithDefaultUsername(cfg.Password)
		clientConfig.WithCredentials(credentials)
	}

	// Configurar configuração avançada com timeouts
	advancedConfig := config.NewAdvancedClientConfiguration()
	advancedConfig.WithConnectionTimeout(cfg.DialTimeout)
	// TODO: Adicionar suporte para ReadTimeout quando disponível na API
	clientConfig.WithAdvancedConfiguration(advancedConfig)

	// TODO: Configurar TLS quando disponível na API
	// if cfg.TLSEnabled {
	//     tlsConfig := p.createTLSConfig(cfg)
	//     clientConfig.WithTLS(tlsConfig)
	// }

	// Configurar DB
	if cfg.DB > 0 {
		clientConfig.WithDatabaseId(cfg.DB)
	}

	// TODO: Configurar retry policy quando disponível na API
	// if cfg.MaxRetries > 0 {
	//     backoffStrategy := config.NewBackoffStrategy(
	//         cfg.MaxRetries,
	//         2, // factor
	//         2, // exponentBase
	//     )
	//     clientConfig.WithReconnectBackoffStrategy(backoffStrategy)
	// }

	client, err := glide.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente standalone: %w", err)
	}

	return client, nil
}

// createClusterClient cria um cliente cluster.
func (p *Provider) createClusterClient(cfg *valkeyconfig.Config) (*glide.ClusterClient, error) {
	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("endereços do cluster não podem estar vazios")
	}

	clusterConfig := config.NewClusterClientConfiguration()

	// Configurar endereços do cluster
	for _, addr := range cfg.Addrs {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("endereço inválido %s: %w", addr, err)
		}

		portInt, err := parsePort(port)
		if err != nil {
			return nil, fmt.Errorf("porta inválida %s: %w", port, err)
		}

		clusterConfig.WithAddress(&config.NodeAddress{
			Host: host,
			Port: portInt,
		})
	}

	// Configurar credenciais
	if cfg.Password != "" {
		credentials := config.NewServerCredentialsWithDefaultUsername(cfg.Password)
		clusterConfig.WithCredentials(credentials)
	}

	// Configurar configuração avançada com timeouts
	advancedConfig := config.NewAdvancedClusterClientConfiguration()
	advancedConfig.WithConnectionTimeout(cfg.DialTimeout)
	clusterConfig.WithAdvancedConfiguration(advancedConfig)

	// Configurar request timeout
	clusterConfig.WithRequestTimeout(cfg.ReadTimeout)

	// Configurar TLS se habilitado
	if cfg.TLSEnabled {
		clusterConfig.WithUseTLS(true)
	}

	// Configurar reconnect strategy se especificado
	if cfg.MaxRetries > 0 {
		backoffStrategy := config.NewBackoffStrategy(
			cfg.MaxRetries,
			2, // baseFactor
			2, // exponentBase
		)
		clusterConfig.WithReconnectStrategy(backoffStrategy)
	}

	client, err := glide.NewClusterClient(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente cluster: %w", err)
	}

	return client, nil
}

// createTLSConfig cria configuração TLS baseada na config.
func (p *Provider) createTLSConfig(cfg *valkeyconfig.Config) *tls.Config {
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

// parsePort converte string para int, validando range de porta.
func parsePort(portStr string) (int, error) {
	port := 0
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("porta deve estar entre 1 e 65535")
	}
	return port, nil
}
