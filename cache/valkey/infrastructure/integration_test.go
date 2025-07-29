//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/cache/valkey"
	valkeyglide "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-glide"
)

// TestIntegrationStandalone testa integração com Valkey standalone
func TestIntegrationStandalone(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available, skipping integration test")
	}

	// Aguardar o serviço estar disponível
	err := WaitForService("localhost:6379", ServiceStartTimeout)
	require.NoError(t, err, "Standalone service should be available")

	// Criar manager e registrar provider
	manager := valkey.NewManager()
	provider := valkeyglide.NewProvider()
	err = manager.RegisterProvider("valkey-glide", provider)
	require.NoError(t, err)

	// Criar cliente com configuração de teste
	testConfig := NewTestConfig()
	client, err := manager.NewClient(testConfig.Standalone)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)
	defer cancel()

	// Testar operações básicas
	t.Run("basic operations", func(t *testing.T) {
		// Test ping
		err := client.Ping(ctx)
		assert.NoError(t, err)

		// Test set/get
		key := fmt.Sprintf("%sstandalone_%d", TestKeyPrefix, time.Now().UnixNano())
		value := fmt.Sprintf("%sstandalone_test", TestValuePrefix)

		err = client.Set(ctx, key, value, 60*time.Second)
		assert.NoError(t, err)

		result, err := client.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Test del
		deleted, err := client.Del(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		// Verify deleted
		_, err = client.Get(ctx, key)
		assert.Error(t, err) // Should error because key doesn't exist
	})

	t.Run("hash operations", func(t *testing.T) {
		hashKey := fmt.Sprintf("%shash_standalone_%d", TestKeyPrefix, time.Now().UnixNano())
		field := "field1"
		value := "value1"

		// Test hset/hget
		err := client.HSet(ctx, hashKey, field, value)
		assert.NoError(t, err)

		result, err := client.HGet(ctx, hashKey, field)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Test hgetall
		allFields, err := client.HGetAll(ctx, hashKey)
		assert.NoError(t, err)
		assert.Equal(t, value, allFields[field])

		// Cleanup
		_, err = client.Del(ctx, hashKey)
		assert.NoError(t, err)
	})
}

// TestIntegrationCluster testa integração com Valkey cluster
func TestIntegrationCluster(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available, skipping integration test")
	}

	// Aguardar o serviço estar disponível
	for i := 0; i < 6; i++ {
		port := 7000 + i
		err := WaitForService(fmt.Sprintf("localhost:%d", port), ServiceStartTimeout)
		require.NoError(t, err, "Cluster node %d should be available", i+1)
	}

	// Criar manager e registrar provider
	manager := valkey.NewManager()
	provider := valkeyglide.NewProvider()
	err := manager.RegisterProvider("valkey-glide", provider)
	require.NoError(t, err)

	// Criar cliente com configuração de cluster
	testConfig := NewTestConfig()
	client, err := manager.NewClient(testConfig.Cluster)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)
	defer cancel()

	// Testar operações de cluster
	t.Run("cluster operations", func(t *testing.T) {
		// Test ping
		err := client.Ping(ctx)
		assert.NoError(t, err)

		// Test distributed operations
		keys := make([]string, 10)
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("%scluster_%d_%d", TestKeyPrefix, time.Now().UnixNano(), i)
			value := fmt.Sprintf("%scluster_test_%d", TestValuePrefix, i)
			keys[i] = key

			err = client.Set(ctx, key, value, 60*time.Second)
			assert.NoError(t, err)
		}

		// Verify all keys can be retrieved
		for i, key := range keys {
			result, err := client.Get(ctx, key)
			assert.NoError(t, err)
			expectedValue := fmt.Sprintf("%scluster_test_%d", TestValuePrefix, i)
			assert.Equal(t, expectedValue, result)
		}

		// Cleanup
		deleted, err := client.Del(ctx, keys...)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(keys)), deleted)
	})
}

// TestIntegrationSentinel testa integração com Valkey sentinel
func TestIntegrationSentinel(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available, skipping integration test")
	}

	// Aguardar os serviços estarem disponíveis
	sentinelPorts := []int{26379, 26380, 26381}
	for _, port := range sentinelPorts {
		err := WaitForService(fmt.Sprintf("localhost:%d", port), ServiceStartTimeout)
		require.NoError(t, err, "Sentinel on port %d should be available", port)
	}

	// Criar manager e registrar provider
	manager := valkey.NewManager()
	provider := valkeyglide.NewProvider()
	err := manager.RegisterProvider("valkey-glide", provider)
	require.NoError(t, err)

	// Criar cliente com configuração de sentinel
	testConfig := NewTestConfig()
	client, err := manager.NewClient(testConfig.Sentinel)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)
	defer cancel()

	// Testar operações de sentinel
	t.Run("sentinel operations", func(t *testing.T) {
		// Test ping
		err := client.Ping(ctx)
		assert.NoError(t, err)

		// Test high availability operations
		key := fmt.Sprintf("%ssentinel_%d", TestKeyPrefix, time.Now().UnixNano())
		value := fmt.Sprintf("%ssentinel_test", TestValuePrefix)

		err = client.Set(ctx, key, value, 60*time.Second)
		assert.NoError(t, err)

		result, err := client.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Test cleanup
		deleted, err := client.Del(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})
}

// TestIntegrationPerformance testa performance básica
func TestIntegrationPerformance(t *testing.T) {
	if !IsDockerAvailable() {
		t.Skip("Docker not available, skipping integration test")
	}

	// Use standalone for performance test
	err := WaitForService("localhost:6379", ServiceStartTimeout)
	require.NoError(t, err)

	manager := valkey.NewManager()
	provider := valkeyglide.NewProvider()
	err = manager.RegisterProvider("valkey-glide", provider)
	require.NoError(t, err)

	testConfig := NewTestConfig()
	client, err := manager.NewClient(testConfig.Standalone)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), IntegrationTestTimeout)
	defer cancel()

	t.Run("bulk operations", func(t *testing.T) {
		start := time.Now()

		// Insert 1000 keys
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("%sperf_%d", TestKeyPrefix, i)
			value := fmt.Sprintf("%sperf_value_%d", TestValuePrefix, i)

			err = client.Set(ctx, key, value, 60*time.Second)
			assert.NoError(t, err)
		}

		insertDuration := time.Since(start)
		t.Logf("Inserted 1000 keys in %v", insertDuration)

		// Read all keys
		start = time.Now()
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("%sperf_%d", TestKeyPrefix, i)
			_, err := client.Get(ctx, key)
			assert.NoError(t, err)
		}

		readDuration := time.Since(start)
		t.Logf("Read 1000 keys in %v", readDuration)

		// Cleanup
		keys := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			keys[i] = fmt.Sprintf("%sperf_%d", TestKeyPrefix, i)
		}

		deleted, err := client.Del(ctx, keys...)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), deleted)
	})
}
