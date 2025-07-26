package internal

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "full validation error",
			err: &ValidationError{
				Field:   "length",
				Value:   "invalid",
				Reason:  "too short",
				UIDType: interfaces.UIDTypeULID,
			},
			expected: "validation failed for ulid (length): invalid - too short",
		},
		{
			name: "minimal validation error",
			err: &ValidationError{
				Field:  "format",
				Value:  "test",
				Reason: "invalid format",
			},
			expected: "validation failed for  (format): test - invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "parse error with cause",
			err: &ParseError{
				Input:   "invalid-input",
				UIDType: interfaces.UIDTypeUUIDV4,
				Reason:  "malformed",
				Cause:   assert.AnError,
			},
			expected: "parse failed for uuid_v4: invalid-input - malformed (caused by: assert.AnError general error for testing)",
		},
		{
			name: "parse error without cause",
			err: &ParseError{
				Input:   "test-input",
				UIDType: interfaces.UIDTypeULID,
				Reason:  "format error",
			},
			expected: "parse failed for ulid: test-input - format error",
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

func TestConversionError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConversionError
		expected string
	}{
		{
			name: "conversion error with cause",
			err: &ConversionError{
				SourceType: interfaces.UIDTypeULID,
				TargetType: interfaces.UIDTypeUUIDV4,
				Reason:     "incompatible types",
				Cause:      assert.AnError,
			},
			expected: "conversion failed from ulid to uuid_v4: incompatible types (caused by: assert.AnError general error for testing)",
		},
		{
			name: "conversion error without cause",
			err: &ConversionError{
				SourceType: interfaces.UIDTypeUUIDV1,
				TargetType: interfaces.UIDTypeUUIDV7,
				Reason:     "unsupported conversion",
			},
			expected: "conversion failed from uuid_v1 to uuid_v7: unsupported conversion",
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

func TestFormatDetector_DetectFormat(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name        string
		input       string
		expected    interfaces.UIDType
		expectError bool
	}{
		{
			name:     "valid ULID",
			input:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expected: interfaces.UIDTypeULID,
		},
		{
			name:     "valid ULID lowercase",
			input:    "01arz3ndektsv4rrffq69g5fav",
			expected: interfaces.UIDTypeULID,
		},
		{
			name:     "valid UUID v4",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: interfaces.UIDTypeUUIDV4,
		},
		{
			name:     "valid UUID v1",
			input:    "550e8400-e29b-11d4-a716-446655440000",
			expected: interfaces.UIDTypeUUIDV1,
		},
		{
			name:     "valid UUID v6",
			input:    "550e8400-e29b-61d4-a716-446655440000",
			expected: interfaces.UIDTypeUUIDV6,
		},
		{
			name:     "valid UUID v7",
			input:    "550e8400-e29b-71d4-a716-446655440000",
			expected: interfaces.UIDTypeUUIDV7,
		},
		{
			name:     "hex UUID without hyphens",
			input:    "550e8400e29b41d4a716446655440000",
			expected: interfaces.UIDTypeUUIDV4,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "invalid-uuid",
			expectError: true,
		},
		{
			name:        "too short",
			input:       "123",
			expectError: true,
		},
		{
			name:        "too long",
			input:       "550e8400-e29b-41d4-a716-446655440000-extra",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.DetectFormat(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBytesToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "single byte",
			input:    []byte{0xFF},
			expected: "ff",
		},
		{
			name:     "multiple bytes",
			input:    []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			expected: "0123456789abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToHex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexToBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectError bool
	}{
		{
			name:        "empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:     "single byte",
			input:    "ff",
			expected: []byte{0xFF},
		},
		{
			name:     "multiple bytes",
			input:    "0123456789abcdef",
			expected: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
		},
		{
			name:     "uppercase hex",
			input:    "0123456789ABCDEF",
			expected: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
		},
		{
			name:     "hex with hyphens",
			input:    "01-23-45-67",
			expected: []byte{0x01, 0x23, 0x45, 0x67},
		},
		{
			name:        "odd length",
			input:       "123",
			expectError: true,
		},
		{
			name:        "invalid hex characters",
			input:       "gg",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HexToBytes(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatUUIDString(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    string
		expectError bool
	}{
		{
			name:     "valid UUID bytes",
			input:    []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			expected: "01234567-89ab-cdef-0123-456789abcdef",
		},
		{
			name:        "invalid length - too short",
			input:       []byte{0x01, 0x23, 0x45},
			expectError: true,
		},
		{
			name:        "invalid length - too long",
			input:       make([]byte, 20),
			expectError: true,
		},
		{
			name:        "empty bytes",
			input:       []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatUUIDString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateUUIDString(t *testing.T) {
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
			name:        "invalid format - no hyphens",
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
			err := ValidateUUIDString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateULIDString(t *testing.T) {
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
			name:  "valid ULID lowercase",
			input: "01arz3ndektsv4rrffq69g5fav",
		},
		{
			name:        "invalid length - too short",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5F",
			expectError: true,
		},
		{
			name:        "invalid length - too long",
			input:       "01ARZ3NDEKTSV4RRFFQ69G5FAVX",
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
			err := ValidateULIDString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       time.Time
		expectError bool
	}{
		{
			name:  "valid current time",
			input: time.Now(),
		},
		{
			name:        "valid unix epoch",
			input:       time.Unix(0, 0).UTC(),
			expectError: false,
		},
		{
			name:  "valid future time",
			input: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "zero time",
			input:       time.Time{},
			expectError: true,
		},
		{
			name:        "before unix epoch",
			input:       time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
			expectError: true,
		},
		{
			name:        "too far in future",
			input:       time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimestamp(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecureRandomBytes(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{
			name:   "valid length 16",
			length: 16,
		},
		{
			name:   "valid length 1",
			length: 1,
		},
		{
			name:   "valid length 32",
			length: 32,
		},
		{
			name:        "zero length",
			length:      0,
			expectError: true,
		},
		{
			name:        "negative length",
			length:      -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SecureRandomBytes(tt.length)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.length)

				// Test randomness by generating multiple times
				if tt.length > 0 {
					result2, err2 := SecureRandomBytes(tt.length)
					require.NoError(t, err2)
					assert.NotEqual(t, result, result2, "Random bytes should be different")
				}
			}
		})
	}
}

func TestExtractUUIDVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    int
		expectError bool
	}{
		{
			name:     "version 1",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 1,
		},
		{
			name:     "version 4",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 4,
		},
		{
			name:     "version 6",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 6,
		},
		{
			name:     "version 7",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 7,
		},
		{
			name:        "invalid length",
			input:       []byte{0x00, 0x00},
			expectError: true,
		},
		{
			name:        "invalid version",
			input:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractUUIDVersion(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Zero(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractUUIDVariant(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    string
		expectError bool
	}{
		{
			name:     "RFC 4122 variant",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "rfc4122",
		},
		{
			name:     "reserved NCS variant",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "reserved_ncs",
		},
		{
			name:     "reserved Microsoft variant",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "reserved_microsoft",
		},
		{
			name:     "reserved future variant",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "reserved_future",
		},
		{
			name:        "invalid length",
			input:       []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractUUIDVariant(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsContextCanceled(t *testing.T) {
	tests := []struct {
		name        string
		ctx         func() context.Context
		expectError bool
	}{
		{
			name: "normal context",
			ctx: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name: "canceled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectError: true,
		},
		{
			name: "timeout context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				time.Sleep(time.Millisecond) // Ensure timeout
				return ctx
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsContextCanceled(tt.ctx())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
