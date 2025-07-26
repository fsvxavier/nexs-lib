package uid

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/config"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)
	assert.NotNil(t, factory)

	// Test with invalid config
	invalidCfg := &config.FactoryConfig{}
	_, err = NewFactory(invalidCfg)
	assert.Error(t, err)
}

func TestFactory_CreateProvider(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		uidType interfaces.UIDType
		wantErr bool
	}{
		{
			name:    "ULID provider",
			uidType: interfaces.UIDTypeULID,
			wantErr: false,
		},
		{
			name:    "UUID v4 provider",
			uidType: interfaces.UIDTypeUUIDV4,
			wantErr: false,
		},
		{
			name:    "UUID v1 provider",
			uidType: interfaces.UIDTypeUUIDV1,
			wantErr: false,
		},
		{
			name:    "UUID v6 provider",
			uidType: interfaces.UIDTypeUUIDV6,
			wantErr: false,
		},
		{
			name:    "UUID v7 provider",
			uidType: interfaces.UIDTypeUUIDV7,
			wantErr: false,
		},
		{
			name:    "unknown provider",
			uidType: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(ctx, tt.uidType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.uidType, provider.GetType())
			}
		})
	}
}

func TestFactory_GetSupportedTypes(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	types := factory.GetSupportedTypes()
	expectedTypes := []interfaces.UIDType{
		interfaces.UIDTypeULID,
		interfaces.UIDTypeUUIDV1,
		interfaces.UIDTypeUUIDV4,
		interfaces.UIDTypeUUIDV6,
		interfaces.UIDTypeUUIDV7,
	}

	assert.ElementsMatch(t, expectedTypes, types)
}

func TestNewUIDManager(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)
	assert.NotNil(t, manager)

	// Test with default manager
	manager2, err := NewDefaultUIDManager()
	require.NoError(t, err)
	assert.NotNil(t, manager2)
}

func TestUIDManager_Generate(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()

	tests := []struct {
		name    string
		uidType interfaces.UIDType
		wantErr bool
	}{
		{
			name:    "generate ULID",
			uidType: interfaces.UIDTypeULID,
			wantErr: false,
		},
		{
			name:    "generate UUID v4",
			uidType: interfaces.UIDTypeUUIDV4,
			wantErr: false,
		},
		{
			name:    "generate UUID v1",
			uidType: interfaces.UIDTypeUUIDV1,
			wantErr: false,
		},
		{
			name:    "generate UUID v6",
			uidType: interfaces.UIDTypeUUIDV6,
			wantErr: false,
		},
		{
			name:    "generate UUID v7",
			uidType: interfaces.UIDTypeUUIDV7,
			wantErr: false,
		},
		{
			name:    "unknown type",
			uidType: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, err := manager.Generate(ctx, tt.uidType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, uid)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, uid)
				assert.Equal(t, tt.uidType, uid.Type)
				assert.NotEmpty(t, uid.Raw)
				assert.NotEmpty(t, uid.Canonical)
				assert.NotEmpty(t, uid.Bytes)
				assert.NotEmpty(t, uid.Hex)
			}
		})
	}
}

func TestUIDManager_GenerateWithTimestamp(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		uidType           interfaces.UIDType
		supportsTimestamp bool
		wantErr           bool
	}{
		{
			name:              "ULID with timestamp",
			uidType:           interfaces.UIDTypeULID,
			supportsTimestamp: true,
			wantErr:           false,
		},
		{
			name:              "UUID v1 with timestamp",
			uidType:           interfaces.UIDTypeUUIDV1,
			supportsTimestamp: true,
			wantErr:           false,
		},
		{
			name:              "UUID v4 ignores timestamp",
			uidType:           interfaces.UIDTypeUUIDV4,
			supportsTimestamp: false,
			wantErr:           false,
		},
		{
			name:              "UUID v6 with timestamp",
			uidType:           interfaces.UIDTypeUUIDV6,
			supportsTimestamp: true,
			wantErr:           false,
		},
		{
			name:              "UUID v7 with timestamp",
			uidType:           interfaces.UIDTypeUUIDV7,
			supportsTimestamp: true,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, err := manager.GenerateWithTimestamp(ctx, tt.uidType, timestamp)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, uid)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, uid)
				assert.Equal(t, tt.uidType, uid.Type)

				if tt.supportsTimestamp && uid.Timestamp != nil {
					// For timestamp-based UIDs, verify timestamp is close to expected
					timeDiff := uid.Timestamp.Sub(timestamp).Abs()
					assert.Less(t, timeDiff, time.Second, "Timestamp should be close to expected value")
				}
			}
		})
	}
}

func TestUIDManager_GenerateDefault(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()

	uid, err := manager.GenerateDefault(ctx)
	require.NoError(t, err)
	assert.NotNil(t, uid)
	assert.Equal(t, factory.GetDefaultType(), uid.Type)
}

func TestUIDManager_Parse(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()

	// Generate a ULID first
	originalUID, err := manager.Generate(ctx, interfaces.UIDTypeULID)
	require.NoError(t, err)

	// Parse it back
	parsedUID, err := manager.Parse(ctx, originalUID.Raw)
	require.NoError(t, err)
	assert.Equal(t, originalUID.Raw, parsedUID.Raw)
	assert.Equal(t, originalUID.Type, parsedUID.Type)

	// Test with invalid input
	_, err = manager.Parse(ctx, "invalid-uid")
	assert.Error(t, err)
}

func TestUIDManager_Validate(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()

	// Generate a valid ULID
	uid, err := manager.Generate(ctx, interfaces.UIDTypeULID)
	require.NoError(t, err)

	// Validate it
	err = manager.Validate(ctx, uid.Raw)
	assert.NoError(t, err)

	// Test with invalid input
	err = manager.Validate(ctx, "invalid-uid")
	assert.Error(t, err)
}

func TestUIDManager_GetSupportedTypes(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	types := manager.GetSupportedTypes()
	expectedTypes := []interfaces.UIDType{
		interfaces.UIDTypeULID,
		interfaces.UIDTypeUUIDV1,
		interfaces.UIDTypeUUIDV4,
		interfaces.UIDTypeUUIDV6,
		interfaces.UIDTypeUUIDV7,
	}

	assert.ElementsMatch(t, expectedTypes, types)
}

func TestUIDManager_MarshalIntegration(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()

	// Test with ULID
	uid, err := manager.Generate(ctx, interfaces.UIDTypeULID)
	require.NoError(t, err)

	// Test marshal/unmarshal cycle
	textData, err := uid.MarshalText()
	require.NoError(t, err)

	newUID := &interfaces.UIDData{}
	err = newUID.UnmarshalText(textData)
	require.NoError(t, err)

	assert.Equal(t, uid.Canonical, newUID.Canonical)
	assert.Equal(t, uid.Raw, newUID.Raw)
}

func TestUIDManager_ConcurrentOperations(t *testing.T) {
	cfg := config.DefaultFactoryConfig()
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	manager := NewUIDManager(factory)

	ctx := context.Background()
	const numGoroutines = 100
	const numOpsPerGoroutine = 10

	results := make(chan *interfaces.UIDData, numGoroutines*numOpsPerGoroutine)
	errors := make(chan error, numGoroutines*numOpsPerGoroutine)

	// Launch concurrent generators
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numOpsPerGoroutine; j++ {
				uid, err := manager.Generate(ctx, interfaces.UIDTypeULID)
				if err != nil {
					errors <- err
					return
				}
				results <- uid
			}
		}()
	}

	// Collect results
	generatedUIDs := make(map[string]bool)
	for i := 0; i < numGoroutines*numOpsPerGoroutine; i++ {
		select {
		case uid := <-results:
			// Check for uniqueness
			assert.False(t, generatedUIDs[uid.Raw], "UID should be unique: %s", uid.Raw)
			generatedUIDs[uid.Raw] = true

		case err := <-errors:
			t.Fatalf("Unexpected error during concurrent generation: %v", err)

		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations to complete")
		}
	}

	assert.Len(t, generatedUIDs, numGoroutines*numOpsPerGoroutine, "All UIDs should be unique")
}
