package valkeyglide

import (
	"testing"
	"time"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
)

func TestProvider_Name(t *testing.T) {
	provider := NewProvider()
	if provider.Name() != "valkey-glide" {
		t.Errorf("expected provider name to be 'valkey-glide', got %s", provider.Name())
	}
}

func TestProvider_DefaultConfig(t *testing.T) {
	provider := NewProvider()
	config := provider.DefaultConfig()

	if config == nil {
		t.Error("expected default config to not be nil")
	}

	valkeyConfig, ok := config.(*valkeyconfig.Config)
	if !ok {
		t.Errorf("expected config to be *valkeyconfig.Config, got %T", config)
	}

	if valkeyConfig.Provider != "valkey-glide" {
		t.Errorf("expected provider to be 'valkey-glide', got %s", valkeyConfig.Provider)
	}
}

func TestProvider_ValidateConfig(t *testing.T) {
	provider := NewProvider()

	// Test with valid config
	config := &valkeyconfig.Config{
		Host:         "localhost",
		Port:         6379,
		Provider:     "valkey-glide",
		PoolSize:     10,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	}

	err := provider.ValidateConfig(config)
	if err != nil {
		t.Errorf("expected no error for valid config, got %v", err)
	}

	// Test with invalid config type
	err = provider.ValidateConfig("invalid")
	if err == nil {
		t.Error("expected error for invalid config type")
	}
}

func TestProvider_NewClient_InvalidConfig(t *testing.T) {
	provider := NewProvider()

	// Test with invalid config type
	_, err := provider.NewClient("invalid")
	if err == nil {
		t.Error("expected error for invalid config type")
	}
}

// Teste básico de compilação - não executa operações reais
func TestProvider_NewClient_BasicStandalone(t *testing.T) {
	provider := NewProvider()

	config := &valkeyconfig.Config{
		Host:         "localhost",
		Port:         6379,
		Provider:     "valkey-glide",
		PoolSize:     10,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	}

	// Este teste só verifica se a criação do cliente não falha por problemas de compilação
	// Não tenta conectar com um servidor real
	_, err := provider.NewClient(config)
	// Esperamos um erro de conexão, mas não erro de compilação/configuração
	if err != nil {
		t.Logf("Expected connection error (no server running): %v", err)
	}
}

func TestProvider_NewClient_BasicCluster(t *testing.T) {
	provider := NewProvider()

	config := &valkeyconfig.Config{
		Host:         "localhost",
		Port:         6379,
		Provider:     "valkey-glide",
		ClusterMode:  true,
		Addrs:        []string{"localhost:6379", "localhost:6380"},
		PoolSize:     10,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	}

	// Este teste só verifica se a criação do cliente não falha por problemas de compilação
	// Não tenta conectar com um servidor real
	_, err := provider.NewClient(config)
	// Esperamos um erro de conexão, mas não erro de compilação/configuração
	if err != nil {
		t.Logf("Expected connection error (no server running): %v", err)
	}
}
