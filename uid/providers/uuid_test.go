package providers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/config"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDProvider_Generate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		version int
		uidType interfaces.UIDType
	}{
		{
			name:    "UUID v1",
			version: 1,
			uidType: interfaces.UIDTypeUUIDV1,
		},
		{
			name:    "UUID v4",
			version: 4,
			uidType: interfaces.UIDTypeUUIDV4,
		},
		{
			name:    "UUID v6",
			version: 6,
			uidType: interfaces.UIDTypeUUIDV6,
		},
		{
			name:    "UUID v7",
			version: 7,
			uidType: interfaces.UIDTypeUUIDV7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.UUIDConfig{
				ProviderConfig: *config.DefaultProviderConfig(tt.uidType),
				Version:        tt.version,
			}

			provider, err := NewUUIDProvider(cfg)
			require.NoError(t, err)

			uid, err := provider.Generate(ctx)
			require.NoError(t, err)
			require.NotNil(t, uid)

			assert.Equal(t, tt.uidType, uid.Type)
			assert.Len(t, uid.Raw, 36)       // UUID with hyphens
			assert.Len(t, uid.Canonical, 36) // UUID with hyphens
			assert.Len(t, uid.Bytes, 16)     // 128 bits
			assert.Len(t, uid.Hex, 32)       // Hex without hyphens
			assert.NotNil(t, uid.Version)
			assert.Equal(t, tt.version, *uid.Version)
		})
	}
}

func TestUUIDProvider_GenerateWithTimestamp(t *testing.T) {
	ctx := context.Background()
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		version            int
		uidType            interfaces.UIDType
		timestampSupported bool
	}{
		{
			name:               "UUID v1 with timestamp",
			version:            1,
			uidType:            interfaces.UIDTypeUUIDV1,
			timestampSupported: true,
		},
		{
			name:               "UUID v4 ignores timestamp",
			version:            4,
			uidType:            interfaces.UIDTypeUUIDV4,
			timestampSupported: false,
		},
		{
			name:               "UUID v6 with timestamp",
			version:            6,
			uidType:            interfaces.UIDTypeUUIDV6,
			timestampSupported: true,
		},
		{
			name:               "UUID v7 with timestamp",
			version:            7,
			uidType:            interfaces.UIDTypeUUIDV7,
			timestampSupported: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.UUIDConfig{
				ProviderConfig: *config.DefaultProviderConfig(tt.uidType),
				Version:        tt.version,
			}

			provider, err := NewUUIDProvider(cfg)
			require.NoError(t, err)

			uid, err := provider.GenerateWithTimestamp(ctx, timestamp)
			require.NoError(t, err)
			require.NotNil(t, uid)

			assert.Equal(t, tt.uidType, uid.Type)
			assert.Equal(t, tt.timestampSupported, provider.IsTimestampSupported())
		})
	}
}

func TestUUIDProvider_Parse(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate a UUID first
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Parse it back
	parsed, err := provider.Parse(ctx, original.Canonical)
	require.NoError(t, err)

	assert.Equal(t, original.Canonical, parsed.Canonical)
	assert.Equal(t, original.Type, parsed.Type)
	assert.Equal(t, original.Hex, parsed.Hex)
	assert.Equal(t, original.Bytes, parsed.Bytes)
}

func TestUUIDProvider_ParseBytes(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate a UUID first
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Parse from bytes
	parsed, err := provider.ParseBytes(ctx, original.Bytes)
	require.NoError(t, err)

	assert.Equal(t, original.Bytes, parsed.Bytes)
	assert.Equal(t, original.Hex, parsed.Hex)
}

func TestUUIDProvider_Validate(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "valid UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:        "invalid length",
			input:       "550e8400-e29b-41d4-a716",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "550e8400e29b41d4a716446655440000",
			expectError: true,
		},
		{
			name:        "invalid characters",
			input:       "550e8400-e29b-41d4-a716-44665544000g",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Validate(ctx, tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUUIDProvider_MarshalText(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	text, err := provider.MarshalText(ctx, uid)
	require.NoError(t, err)

	assert.Equal(t, uid.Canonical, string(text))

	// Test with wrong type
	wrongUID := &interfaces.UIDData{Type: interfaces.UIDTypeULID}
	_, err = provider.MarshalText(ctx, wrongUID)
	assert.Error(t, err)
}

func TestUUIDProvider_MarshalBinary(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	binary, err := provider.MarshalBinary(ctx, uid)
	require.NoError(t, err)

	assert.Equal(t, uid.Bytes, binary)
	assert.NotSame(t, &uid.Bytes, &binary) // Should be a copy

	// Test with no binary data
	emptyUID := &interfaces.UIDData{Type: interfaces.UIDTypeUUIDV4}
	_, err = provider.MarshalBinary(ctx, emptyUID)
	assert.Error(t, err)
}

func TestUUIDProvider_MarshalJSON(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	jsonData, err := provider.MarshalJSON(ctx, uid)
	require.NoError(t, err)

	var parsed interfaces.UIDData
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	assert.Equal(t, uid.Type, parsed.Type)
	assert.Equal(t, uid.Canonical, parsed.Canonical)
}

func TestUUIDProvider_UnmarshalText(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate original UUID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to text
	text, err := provider.MarshalText(ctx, original)
	require.NoError(t, err)

	// Unmarshal from text
	unmarshaled, err := provider.UnmarshalText(ctx, text)
	require.NoError(t, err)

	assert.Equal(t, original.Canonical, unmarshaled.Canonical)

	// Test with empty data
	_, err = provider.UnmarshalText(ctx, []byte{})
	assert.Error(t, err)
}

func TestUUIDProvider_UnmarshalBinary(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate original UUID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to binary
	binary, err := provider.MarshalBinary(ctx, original)
	require.NoError(t, err)

	// Unmarshal from binary
	unmarshaled, err := provider.UnmarshalBinary(ctx, binary)
	require.NoError(t, err)

	assert.Equal(t, original.Bytes, unmarshaled.Bytes)

	// Test with empty data
	_, err = provider.UnmarshalBinary(ctx, []byte{})
	assert.Error(t, err)
}

func TestUUIDProvider_UnmarshalJSON(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate original UUID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to JSON
	jsonData, err := provider.MarshalJSON(ctx, original)
	require.NoError(t, err)

	// Unmarshal from JSON
	unmarshaled, err := provider.UnmarshalJSON(ctx, jsonData)
	require.NoError(t, err)

	assert.Equal(t, original.Canonical, unmarshaled.Canonical)

	// Test with empty data
	_, err = provider.UnmarshalJSON(ctx, []byte{})
	assert.Error(t, err)

	// Test with unsupported type in JSON
	wrongTypeJSON := `{"type":"ulid","canonical":"test"}`
	_, err = provider.UnmarshalJSON(ctx, []byte(wrongTypeJSON))
	assert.Error(t, err)
}

func TestUUIDProvider_GetType(t *testing.T) {
	tests := []struct {
		name     string
		version  int
		expected interfaces.UIDType
	}{
		{
			name:     "UUID v1",
			version:  1,
			expected: interfaces.UIDTypeUUIDV1,
		},
		{
			name:     "UUID v4",
			version:  4,
			expected: interfaces.UIDTypeUUIDV4,
		},
		{
			name:     "UUID v6",
			version:  6,
			expected: interfaces.UIDTypeUUIDV6,
		},
		{
			name:     "UUID v7",
			version:  7,
			expected: interfaces.UIDTypeUUIDV7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.UUIDConfig{
				ProviderConfig: *config.DefaultProviderConfig(tt.expected),
				Version:        tt.version,
			}

			provider, err := NewUUIDProvider(cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, provider.GetType())
		})
	}
}

func TestUUIDProvider_IsTimestampSupported(t *testing.T) {
	tests := []struct {
		name      string
		version   int
		uidType   interfaces.UIDType
		supported bool
	}{
		{
			name:      "UUID v1 supports timestamp",
			version:   1,
			uidType:   interfaces.UIDTypeUUIDV1,
			supported: true,
		},
		{
			name:      "UUID v4 does not support timestamp",
			version:   4,
			uidType:   interfaces.UIDTypeUUIDV4,
			supported: false,
		},
		{
			name:      "UUID v6 supports timestamp",
			version:   6,
			uidType:   interfaces.UIDTypeUUIDV6,
			supported: true,
		},
		{
			name:      "UUID v7 supports timestamp",
			version:   7,
			uidType:   interfaces.UIDTypeUUIDV7,
			supported: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.UUIDConfig{
				ProviderConfig: *config.DefaultProviderConfig(tt.uidType),
				Version:        tt.version,
			}

			provider, err := NewUUIDProvider(cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.supported, provider.IsTimestampSupported())
		})
	}
}

func TestUUIDProvider_GetSupportedTypes(t *testing.T) {
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	types := provider.GetSupportedTypes()
	expected := []interfaces.UIDType{
		interfaces.UIDTypeUUIDV1,
		interfaces.UIDTypeUUIDV4,
		interfaces.UIDTypeUUIDV6,
		interfaces.UIDTypeUUIDV7,
	}

	assert.Equal(t, expected, types)
}

func TestUUIDProvider_IsThreadSafe(t *testing.T) {
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	assert.True(t, provider.IsThreadSafe())
}

func TestUUIDProvider_ExtractTimestamp(t *testing.T) {
	ctx := context.Background()

	// Test with UUID v7 (has millisecond timestamp)
	cfg := config.DefaultUUIDV7Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate UUID v7
	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Extract timestamp (UUID v7 should have a timestamp)
	if uid.Timestamp != nil {
		extracted, err := provider.ExtractTimestamp(ctx, uid.Canonical)
		require.NoError(t, err)

		// Timestamps should be close (within a second)
		timeDiff := extracted.Sub(*uid.Timestamp)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		assert.Less(t, timeDiff, time.Second)
	}
}

func TestUUIDProvider_IsValidUUID(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	// Generate valid UUID
	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	assert.True(t, provider.IsValidUUID(ctx, uid.Canonical))
	assert.False(t, provider.IsValidUUID(ctx, "invalid-uuid"))
}

func TestUUIDProvider_ConcurrentGeneration(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultUUIDV4Config()
	provider, err := NewUUIDProvider(cfg)
	require.NoError(t, err)

	const numGoroutines = 100
	const numPerGoroutine = 10

	results := make(chan *interfaces.UIDData, numGoroutines*numPerGoroutine)
	errors := make(chan error, numGoroutines*numPerGoroutine)

	// Start multiple goroutines generating UUIDs
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numPerGoroutine; j++ {
				uid, err := provider.Generate(ctx)
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
	for i := 0; i < numGoroutines*numPerGoroutine; i++ {
		select {
		case uid := <-results:
			// Check for duplicates
			if generatedUIDs[uid.Canonical] {
				t.Fatalf("Duplicate UUID generated: %s", uid.Canonical)
			}
			generatedUIDs[uid.Canonical] = true

			// Validate the UID
			assert.Equal(t, interfaces.UIDTypeUUIDV4, uid.Type)
			assert.Len(t, uid.Canonical, 36)

		case err := <-errors:
			t.Fatalf("Error generating UUID: %v", err)
		}
	}

	assert.Len(t, generatedUIDs, numGoroutines*numPerGoroutine)
}
