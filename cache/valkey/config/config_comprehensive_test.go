package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_CopySimple(t *testing.T) {
	original := &Config{
		Provider: "valkey-go",
		Host:     "localhost",
		Port:     6379,
		Password: "secret",
		DB:       5,
		URI:      "valkey://localhost:6379",
		PoolSize: 20,
		Addrs:    []string{"localhost:6379", "localhost:6380"},
	}

	copied := original.Copy()

	require.NotNil(t, copied)
	assert.Equal(t, original.Provider, copied.Provider)
	assert.Equal(t, original.Host, copied.Host)
	assert.Equal(t, original.Port, copied.Port)
	assert.Equal(t, original.Password, copied.Password)
	assert.Equal(t, original.DB, copied.DB)
	assert.Equal(t, original.URI, copied.URI)
	assert.Equal(t, original.PoolSize, copied.PoolSize)

	// Verify slices are copied (not shared)
	if original.Addrs != nil && copied.Addrs != nil {
		assert.Equal(t, original.Addrs, copied.Addrs)
		assert.NotSame(t, original.Addrs, copied.Addrs)
	}
}

func TestConfig_Copy_NilSlices(t *testing.T) {
	original := &Config{
		Provider: "valkey-go",
		Addrs:    nil,
	}

	copied := original.Copy()

	require.NotNil(t, copied)
	assert.Nil(t, copied.Addrs)
}

func TestConfig_Copy_EmptySlices(t *testing.T) {
	original := &Config{
		Provider: "valkey-go",
		Addrs:    []string{},
	}

	copied := original.Copy()

	require.NotNil(t, copied)
	assert.Empty(t, copied.Addrs)
}

func TestLoadFromEnv_Basic(t *testing.T) {
	t.Run("default_when_no_env_vars", func(t *testing.T) {
		// Clear environment
		os.Clearenv()

		config := LoadFromEnv()

		expected := DefaultConfig()
		assert.Equal(t, expected.Provider, config.Provider)
		assert.Equal(t, expected.Host, config.Host)
		assert.Equal(t, expected.Port, config.Port)
	})

	t.Run("basic_env_vars", func(t *testing.T) {
		// Clear environment
		os.Clearenv()

		// Set some environment variables
		os.Setenv("VALKEY_PROVIDER", "valkey-glide")
		os.Setenv("VALKEY_HOST", "example.com")
		os.Setenv("VALKEY_PORT", "6380")
		defer func() {
			os.Unsetenv("VALKEY_PROVIDER")
			os.Unsetenv("VALKEY_HOST")
			os.Unsetenv("VALKEY_PORT")
		}()

		config := LoadFromEnv()

		assert.Equal(t, "valkey-glide", config.Provider)
		assert.Equal(t, "example.com", config.Host)
		assert.Equal(t, 6380, config.Port)
	})
}

func TestConfig_Validate_Basic(t *testing.T) {
	t.Run("valid_default_config", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid_provider", func(t *testing.T) {
		config := DefaultConfig()
		config.Provider = "invalid-provider"
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("invalid_port", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = -1
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("invalid_pool_size", func(t *testing.T) {
		config := DefaultConfig()
		config.PoolSize = 0
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool_size")
	})
}



func BenchmarkConfig_Copy(b *testing.B) {
	config := &Config{
		Provider:      "valkey-go",
		Host:          "localhost",
		Port:          6379,
		Password:      "secret",
		DB:            5,
		PoolSize:      20,
		MinIdleConns:  5,
		Addrs:         []string{"localhost:6379", "localhost:6380", "localhost:6381"},
		SentinelAddrs: []string{"localhost:26379", "localhost:26380"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Copy()
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkLoadFromEnv(b *testing.B) {
	// Set some environment variables
	os.Setenv("VALKEY_PROVIDER", "valkey-go")
	os.Setenv("VALKEY_HOST", "localhost")
	os.Setenv("VALKEY_PORT", "6379")
	defer func() {
		os.Unsetenv("VALKEY_PROVIDER")
		os.Unsetenv("VALKEY_HOST")
		os.Unsetenv("VALKEY_PORT")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LoadFromEnv()
	}
}
