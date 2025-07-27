package httpclient

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientManager(t *testing.T) {
	t.Parallel()

	manager := NewClientManager()

	t.Run("GetOrCreateClient creates new client", func(t *testing.T) {
		client, err := manager.GetOrCreateClient(
			"test-client",
			interfaces.ProviderNetHTTP,
			nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-client", client.GetID())
	})

	t.Run("GetOrCreateClient returns existing client", func(t *testing.T) {
		// Create first client
		client1, err := manager.GetOrCreateClient(
			"reuse-client",
			interfaces.ProviderNetHTTP,
			nil,
		)
		require.NoError(t, err)

		// Get the same client again
		client2, err := manager.GetOrCreateClient(
			"reuse-client",
			interfaces.ProviderNetHTTP,
			nil,
		)
		require.NoError(t, err)

		// Should be the same instance
		assert.Equal(t, client1.GetID(), client2.GetID())
	})

	t.Run("GetClient retrieves existing client", func(t *testing.T) {
		// Create client
		_, err := manager.GetOrCreateClient(
			"get-client",
			interfaces.ProviderNetHTTP,
			nil,
		)
		require.NoError(t, err)

		// Retrieve client
		client, exists := manager.GetClient("get-client")
		assert.True(t, exists)
		assert.NotNil(t, client)
		assert.Equal(t, "get-client", client.GetID())
	})

	t.Run("GetClient returns false for non-existing client", func(t *testing.T) {
		client, exists := manager.GetClient("non-existing")
		assert.False(t, exists)
		assert.Nil(t, client)
	})

	t.Run("ListClients returns all client names", func(t *testing.T) {
		// Create multiple clients
		_, err := manager.GetOrCreateClient("client1", interfaces.ProviderNetHTTP, nil)
		require.NoError(t, err)
		_, err = manager.GetOrCreateClient("client2", interfaces.ProviderFiber, nil)
		require.NoError(t, err)

		clients := manager.ListClients()
		assert.Contains(t, clients, "client1")
		assert.Contains(t, clients, "client2")
	})

	t.Run("RemoveClient removes client", func(t *testing.T) {
		// Create client
		_, err := manager.GetOrCreateClient("remove-me", interfaces.ProviderNetHTTP, nil)
		require.NoError(t, err)

		// Verify it exists
		_, exists := manager.GetClient("remove-me")
		assert.True(t, exists)

		// Remove it
		manager.RemoveClient("remove-me")

		// Verify it's gone
		_, exists = manager.GetClient("remove-me")
		assert.False(t, exists)
	})

	t.Run("Shutdown clears all clients", func(t *testing.T) {
		// Create some clients
		_, err := manager.GetOrCreateClient("shutdown1", interfaces.ProviderNetHTTP, nil)
		require.NoError(t, err)
		_, err = manager.GetOrCreateClient("shutdown2", interfaces.ProviderNetHTTP, nil)
		require.NoError(t, err)

		// Verify they exist
		clients := manager.ListClients()
		assert.Contains(t, clients, "shutdown1")
		assert.Contains(t, clients, "shutdown2")

		// Shutdown
		err = manager.Shutdown()
		require.NoError(t, err)

		// Verify they're gone
		clients = manager.ListClients()
		assert.NotContains(t, clients, "shutdown1")
		assert.NotContains(t, clients, "shutdown2")
	})
}

func TestGlobalManager(t *testing.T) {
	t.Parallel()

	t.Run("GetManager returns singleton instance", func(t *testing.T) {
		manager1 := GetManager()
		manager2 := GetManager()

		// Should be the same instance
		assert.Equal(t, manager1, manager2)
	})
}

func TestConvenienceFunctions(t *testing.T) {
	t.Parallel()

	t.Run("New creates client with baseURL", func(t *testing.T) {
		client, err := New(interfaces.ProviderNetHTTP, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://example.com", client.GetConfig().BaseURL)
	})

	t.Run("NewNamed creates named client", func(t *testing.T) {
		client, err := NewNamed("my-api", interfaces.ProviderNetHTTP, "https://api.example.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "my-api", client.GetID())

		// Should be able to retrieve it
		retrieved, exists := GetNamedClient("my-api")
		assert.True(t, exists)
		assert.Equal(t, client.GetID(), retrieved.GetID())
	})

	t.Run("GetNamedClient returns false for non-existing", func(t *testing.T) {
		client, exists := GetNamedClient("non-existing-named")
		assert.False(t, exists)
		assert.Nil(t, client)
	})
}

func TestConnectionReuse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection reuse test in short mode")
	}

	t.Run("Multiple clients reuse connections", func(t *testing.T) {
		// Create multiple instances of the same named client
		client1, err := NewNamed("reuse-test", interfaces.ProviderNetHTTP, "https://httpbin.org")
		require.NoError(t, err)

		client2, err := NewNamed("reuse-test", interfaces.ProviderNetHTTP, "https://httpbin.org")
		require.NoError(t, err)

		// Should be the same client instance
		assert.Equal(t, client1.GetID(), client2.GetID())

		ctx := context.Background()

		// Make requests with both clients
		resp1, err := client1.Get(ctx, "/get")
		require.NoError(t, err)
		assert.Equal(t, 200, resp1.StatusCode)

		resp2, err := client2.Get(ctx, "/get")
		require.NoError(t, err)
		assert.Equal(t, 200, resp2.StatusCode)

		// Check metrics show both requests
		metrics := client1.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalRequests, int64(2))
	})
}

func TestOptimizeConfigForReuse(t *testing.T) {
	t.Parallel()

	t.Run("Optimizes config for connection reuse", func(t *testing.T) {
		originalConfig := &interfaces.Config{
			BaseURL:             "https://example.com",
			MaxIdleConns:        0,    // Will be optimized
			IdleConnTimeout:     0,    // Will be optimized
			DisableKeepAlives:   true, // Will be optimized
			TLSHandshakeTimeout: 0,    // Will be optimized
		}

		optimized := optimizeConfigForReuse(originalConfig)

		// Should have optimal values
		assert.Equal(t, 100, optimized.MaxIdleConns)
		assert.Equal(t, 90*time.Second, optimized.IdleConnTimeout)
		assert.False(t, optimized.DisableKeepAlives)
		assert.Equal(t, 10*time.Second, optimized.TLSHandshakeTimeout)

		// Should preserve original values
		assert.Equal(t, originalConfig.BaseURL, optimized.BaseURL)
	})

	t.Run("Preserves existing optimal values", func(t *testing.T) {
		originalConfig := &interfaces.Config{
			BaseURL:             "https://example.com",
			MaxIdleConns:        50,
			IdleConnTimeout:     60 * time.Second,
			DisableKeepAlives:   false,
			TLSHandshakeTimeout: 5 * time.Second,
		}

		optimized := optimizeConfigForReuse(originalConfig)

		// Should preserve existing values
		assert.Equal(t, 50, optimized.MaxIdleConns)
		assert.Equal(t, 60*time.Second, optimized.IdleConnTimeout)
		assert.False(t, optimized.DisableKeepAlives)
		assert.Equal(t, 5*time.Second, optimized.TLSHandshakeTimeout)
	})
}

func TestClientExtendedInterface(t *testing.T) {
	t.Parallel()

	client, err := New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	require.NoError(t, err)

	t.Run("GetConfig returns configuration", func(t *testing.T) {
		config := client.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "https://httpbin.org", config.BaseURL)
	})

	t.Run("GetID returns client ID", func(t *testing.T) {
		id := client.GetID()
		assert.NotEmpty(t, id)
	})

	t.Run("IsHealthy returns health status", func(t *testing.T) {
		healthy := client.IsHealthy()
		assert.True(t, healthy)
	})

	t.Run("GetMetrics returns metrics", func(t *testing.T) {
		metrics := client.GetMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.TotalRequests, int64(0))
	})
}

func TestConcurrentClientAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	t.Run("Concurrent access to same named client", func(t *testing.T) {
		const numGoroutines = 10
		const clientName = "concurrent-test"

		results := make(chan string, numGoroutines)

		// Launch multiple goroutines trying to get the same client
		for i := 0; i < numGoroutines; i++ {
			go func() {
				client, err := NewNamed(clientName, interfaces.ProviderNetHTTP, "https://httpbin.org")
				if err != nil {
					results <- ""
					return
				}
				results <- client.GetID()
			}()
		}

		// Collect results
		var clientIDs []string
		for i := 0; i < numGoroutines; i++ {
			id := <-results
			if id != "" {
				clientIDs = append(clientIDs, id)
			}
		}

		// All should have the same client ID
		assert.Greater(t, len(clientIDs), 0)
		firstID := clientIDs[0]
		for _, id := range clientIDs {
			assert.Equal(t, firstID, id)
		}
	})
}
