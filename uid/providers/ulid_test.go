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

func TestULIDProvider_Generate(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)
	require.NotNil(t, uid)

	assert.Equal(t, interfaces.UIDTypeULID, uid.Type)
	assert.Len(t, uid.Raw, 26)
	assert.Len(t, uid.Canonical, 26)
	assert.Len(t, uid.Bytes, 16)
	assert.Len(t, uid.Hex, 32)
	assert.NotNil(t, uid.Timestamp)
}

func TestULIDProvider_GenerateWithTimestamp(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	uid, err := provider.GenerateWithTimestamp(ctx, timestamp)
	require.NoError(t, err)
	require.NotNil(t, uid)

	assert.Equal(t, interfaces.UIDTypeULID, uid.Type)
	assert.Equal(t, timestamp.UnixMilli(), uid.Timestamp.UnixMilli())
}

func TestULIDProvider_Parse(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate a ULID first
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

func TestULIDProvider_ParseBytes(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate a ULID first
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Parse from bytes
	parsed, err := provider.ParseBytes(ctx, original.Bytes)
	require.NoError(t, err)

	assert.Equal(t, original.Type, parsed.Type)
	assert.Equal(t, original.Bytes, parsed.Bytes)
	assert.Equal(t, original.Hex, parsed.Hex)
}

func TestULIDProvider_Validate(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "valid ULID",
			input: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		},
		{
			name:        "invalid length",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5F",
			expectError: true,
		},
		{
			name:        "invalid characters",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5F@V",
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

func TestULIDProvider_MarshalText(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	text, err := provider.MarshalText(ctx, uid)
	require.NoError(t, err)

	assert.Equal(t, uid.Canonical, string(text))

	// Test with wrong type
	wrongUID := &interfaces.UIDData{Type: interfaces.UIDTypeUUIDV4}
	_, err = provider.MarshalText(ctx, wrongUID)
	assert.Error(t, err)
}

func TestULIDProvider_MarshalBinary(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	binary, err := provider.MarshalBinary(ctx, uid)
	require.NoError(t, err)

	assert.Equal(t, uid.Bytes, binary)
	assert.NotSame(t, &uid.Bytes, &binary) // Should be a copy

	// Test with no binary data
	emptyUID := &interfaces.UIDData{Type: interfaces.UIDTypeULID}
	_, err = provider.MarshalBinary(ctx, emptyUID)
	assert.Error(t, err)
}

func TestULIDProvider_MarshalJSON(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
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

func TestULIDProvider_UnmarshalText(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate original ULID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to text
	text, err := provider.MarshalText(ctx, original)
	require.NoError(t, err)

	// Unmarshal from text
	unmarshaled, err := provider.UnmarshalText(ctx, text)
	require.NoError(t, err)

	assert.Equal(t, original.Canonical, unmarshaled.Canonical)
	assert.Equal(t, original.Type, unmarshaled.Type)

	// Test with empty data
	_, err = provider.UnmarshalText(ctx, []byte{})
	assert.Error(t, err)
}

func TestULIDProvider_UnmarshalBinary(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate original ULID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to binary
	binary, err := provider.MarshalBinary(ctx, original)
	require.NoError(t, err)

	// Unmarshal from binary
	unmarshaled, err := provider.UnmarshalBinary(ctx, binary)
	require.NoError(t, err)

	assert.Equal(t, original.Bytes, unmarshaled.Bytes)
	assert.Equal(t, original.Type, unmarshaled.Type)

	// Test with empty data
	_, err = provider.UnmarshalBinary(ctx, []byte{})
	assert.Error(t, err)
}

func TestULIDProvider_UnmarshalJSON(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate original ULID
	original, err := provider.Generate(ctx)
	require.NoError(t, err)

	// Marshal to JSON
	jsonData, err := provider.MarshalJSON(ctx, original)
	require.NoError(t, err)

	// Unmarshal from JSON
	unmarshaled, err := provider.UnmarshalJSON(ctx, jsonData)
	require.NoError(t, err)

	assert.Equal(t, original.Canonical, unmarshaled.Canonical)
	assert.Equal(t, original.Type, unmarshaled.Type)

	// Test with empty data
	_, err = provider.UnmarshalJSON(ctx, []byte{})
	assert.Error(t, err)

	// Test with wrong type in JSON
	wrongTypeJSON := `{"type":"uuid_v4","canonical":"test"}`
	_, err = provider.UnmarshalJSON(ctx, []byte(wrongTypeJSON))
	assert.Error(t, err)
}

func TestULIDProvider_GetType(t *testing.T) {
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	assert.Equal(t, interfaces.UIDTypeULID, provider.GetType())
}

func TestULIDProvider_IsTimestampSupported(t *testing.T) {
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	assert.True(t, provider.IsTimestampSupported())
}

func TestULIDProvider_GetSupportedTypes(t *testing.T) {
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	types := provider.GetSupportedTypes()
	assert.Equal(t, []interfaces.UIDType{interfaces.UIDTypeULID}, types)
}

func TestULIDProvider_IsThreadSafe(t *testing.T) {
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	assert.True(t, provider.IsThreadSafe())
}

func TestULIDProvider_ExtractTimestamp(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate ULID with known timestamp
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	uid, err := provider.GenerateWithTimestamp(ctx, timestamp)
	require.NoError(t, err)

	// Extract timestamp
	extracted, err := provider.ExtractTimestamp(ctx, uid.Canonical)
	require.NoError(t, err)

	assert.Equal(t, timestamp.UnixMilli(), extracted.UnixMilli())
}

func TestULIDProvider_IsValidULID(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	// Generate valid ULID
	uid, err := provider.Generate(ctx)
	require.NoError(t, err)

	assert.True(t, provider.IsValidULID(ctx, uid.Canonical))
	assert.False(t, provider.IsValidULID(ctx, "invalid-ulid"))
}

func TestULIDProvider_ConcurrentGeneration(t *testing.T) {
	ctx := context.Background()
	provider, err := NewULIDProvider(config.DefaultULIDConfig())
	require.NoError(t, err)

	const numGoroutines = 100
	const numPerGoroutine = 10

	results := make(chan *interfaces.UIDData, numGoroutines*numPerGoroutine)
	errors := make(chan error, numGoroutines*numPerGoroutine)

	// Start multiple goroutines generating ULIDs
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
				t.Fatalf("Duplicate ULID generated: %s", uid.Canonical)
			}
			generatedUIDs[uid.Canonical] = true

			// Validate the UID
			assert.Equal(t, interfaces.UIDTypeULID, uid.Type)
			assert.Len(t, uid.Canonical, 26)

		case err := <-errors:
			t.Fatalf("Error generating ULID: %v", err)
		}
	}

	assert.Len(t, generatedUIDs, numGoroutines*numPerGoroutine)
}
