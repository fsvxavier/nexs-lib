package interfaces

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUIDData_MarshalText(t *testing.T) {
	uid := &UIDData{
		Raw:       "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Canonical: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Bytes:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		Hex:       "0102030405060708090a0b0c0d0e0f10",
		Type:      UIDTypeULID,
	}

	text, err := uid.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, uid.Canonical, string(text))
}

func TestUIDData_UnmarshalText(t *testing.T) {
	uid := &UIDData{}
	text := []byte("01ARZ3NDEKTSV4RRFFQ69G5FAV")

	err := uid.UnmarshalText(text)
	require.NoError(t, err)
	assert.Equal(t, string(text), uid.Canonical)
	assert.Equal(t, string(text), uid.Raw)
}

func TestUIDData_MarshalBinary(t *testing.T) {
	testBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	uid := &UIDData{
		Raw:       "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Canonical: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Bytes:     testBytes,
		Hex:       "0102030405060708090a0b0c0d0e0f10",
		Type:      UIDTypeULID,
	}

	binary, err := uid.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, testBytes, binary)

	// Test with empty bytes
	emptyUID := &UIDData{Type: UIDTypeULID}
	_, err = emptyUID.MarshalBinary()
	assert.Error(t, err)
	var marshalErr *MarshalError
	assert.ErrorAs(t, err, &marshalErr)
}

func TestUIDData_UnmarshalBinary(t *testing.T) {
	testBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	uid := &UIDData{}

	err := uid.UnmarshalBinary(testBytes)
	require.NoError(t, err)
	assert.Equal(t, testBytes, uid.Bytes)

	// Test with empty data
	emptyUID := &UIDData{}
	err = emptyUID.UnmarshalBinary([]byte{})
	assert.Error(t, err)
	var unmarshalErr *UnmarshalError
	assert.ErrorAs(t, err, &unmarshalErr)
}

func TestUIDData_MarshalJSON(t *testing.T) {
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	version := 4
	variant := "rfc4122"

	uid := &UIDData{
		Raw:       "550e8400-e29b-41d4-a716-446655440000",
		Canonical: "550e8400-e29b-41d4-a716-446655440000",
		Bytes:     []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
		Hex:       "550e8400e29b41d4a716446655440000",
		Type:      UIDTypeUUIDV4,
		Timestamp: &timestamp,
		Version:   &version,
		Variant:   &variant,
	}

	jsonData, err := uid.MarshalJSON()
	require.NoError(t, err)

	// Verify JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	assert.Equal(t, uid.Raw, parsed["raw"])
	assert.Equal(t, uid.Canonical, parsed["canonical"])
	assert.Equal(t, string(uid.Type), parsed["type"])
	assert.Equal(t, uid.Hex, parsed["hex"])
	assert.Equal(t, float64(version), parsed["version"])
	assert.Equal(t, variant, parsed["variant"])
}

func TestUIDData_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"raw": "550e8400-e29b-41d4-a716-446655440000",
		"canonical": "550e8400-e29b-41d4-a716-446655440000",
		"bytes": "VQ6EAOKbQdSnFkRmVUQAAA==",
		"hex": "550e8400e29b41d4a716446655440000",
		"type": "uuid_v4",
		"timestamp": "2023-01-01T12:00:00Z",
		"version": 4,
		"variant": "rfc4122"
	}`

	uid := &UIDData{}
	err := uid.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uid.Raw)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uid.Canonical)
	assert.Equal(t, "550e8400e29b41d4a716446655440000", uid.Hex)
	assert.Equal(t, UIDTypeUUIDV4, uid.Type)
	assert.NotNil(t, uid.Timestamp)
	assert.NotNil(t, uid.Version)
	assert.Equal(t, 4, *uid.Version)
	assert.NotNil(t, uid.Variant)
	assert.Equal(t, "rfc4122", *uid.Variant)
}

func TestUIDData_String(t *testing.T) {
	uid := &UIDData{
		Canonical: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
	}

	assert.Equal(t, uid.Canonical, uid.String())
}

func TestMarshalError(t *testing.T) {
	tests := []struct {
		name     string
		err      *MarshalError
		expected string
	}{
		{
			name: "marshal error with cause",
			err: &MarshalError{
				Type:   UIDTypeULID,
				Reason: "encoding failed",
				Cause:  assert.AnError,
			},
			expected: "marshal failed for ulid: encoding failed (caused by: assert.AnError general error for testing)",
		},
		{
			name: "marshal error without cause",
			err: &MarshalError{
				Type:   UIDTypeUUIDV4,
				Reason: "no data available",
			},
			expected: "marshal failed for uuid_v4: no data available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			}
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	tests := []struct {
		name     string
		err      *UnmarshalError
		expected string
	}{
		{
			name: "unmarshal error with cause",
			err: &UnmarshalError{
				Type:   UIDTypeULID,
				Reason: "decoding failed",
				Cause:  assert.AnError,
			},
			expected: "unmarshal failed for ulid: decoding failed (caused by: assert.AnError general error for testing)",
		},
		{
			name: "unmarshal error without cause",
			err: &UnmarshalError{
				Type:   UIDTypeUUIDV4,
				Reason: "invalid format",
			},
			expected: "unmarshal failed for uuid_v4: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			}
		})
	}
}

func TestUIDData_RoundTripSerialization(t *testing.T) {
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	version := 4
	variant := "rfc4122"

	original := &UIDData{
		Raw:       "550e8400-e29b-41d4-a716-446655440000",
		Canonical: "550e8400-e29b-41d4-a716-446655440000",
		Bytes:     []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00},
		Hex:       "550e8400e29b41d4a716446655440000",
		Type:      UIDTypeUUIDV4,
		Timestamp: &timestamp,
		Version:   &version,
		Variant:   &variant,
	}

	// Test JSON round trip
	t.Run("JSON round trip", func(t *testing.T) {
		jsonData, err := original.MarshalJSON()
		require.NoError(t, err)

		restored := &UIDData{}
		err = restored.UnmarshalJSON(jsonData)
		require.NoError(t, err)

		assert.Equal(t, original.Raw, restored.Raw)
		assert.Equal(t, original.Canonical, restored.Canonical)
		assert.Equal(t, original.Hex, restored.Hex)
		assert.Equal(t, original.Type, restored.Type)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Variant, restored.Variant)

		if original.Timestamp != nil && restored.Timestamp != nil {
			assert.Equal(t, original.Timestamp.Unix(), restored.Timestamp.Unix())
		}
	})

	// Test Text round trip
	t.Run("Text round trip", func(t *testing.T) {
		textData, err := original.MarshalText()
		require.NoError(t, err)

		restored := &UIDData{}
		err = restored.UnmarshalText(textData)
		require.NoError(t, err)

		assert.Equal(t, original.Canonical, restored.Canonical)
		assert.Equal(t, original.Canonical, restored.Raw)
	})

	// Test Binary round trip
	t.Run("Binary round trip", func(t *testing.T) {
		binaryData, err := original.MarshalBinary()
		require.NoError(t, err)

		restored := &UIDData{}
		err = restored.UnmarshalBinary(binaryData)
		require.NoError(t, err)

		assert.Equal(t, original.Bytes, restored.Bytes)
	})
}
