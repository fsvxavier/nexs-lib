// Package interfaces provides the core contracts for UID generation and manipulation.
// This package defines the main interfaces that all UID providers must implement,
// ensuring consistent behavior across different UID types (ULID, UUID, etc.).
package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// UIDType represents the type of UID that can be generated.
type UIDType string

const (
	// UIDTypeULID represents Universally Unique Lexicographically Sortable Identifier.
	UIDTypeULID UIDType = "ulid"
	// UIDTypeUUIDV1 represents UUID version 1 (timestamp-based).
	UIDTypeUUIDV1 UIDType = "uuid_v1"
	// UIDTypeUUIDV4 represents UUID version 4 (random).
	UIDTypeUUIDV4 UIDType = "uuid_v4"
	// UIDTypeUUIDV6 represents UUID version 6 (timestamp-ordered).
	UIDTypeUUIDV6 UIDType = "uuid_v6"
	// UIDTypeUUIDV7 represents UUID version 7 (timestamp and random).
	UIDTypeUUIDV7 UIDType = "uuid_v7"
)

// UIDData represents comprehensive UID information with multiple format representations.
// This structure provides all possible representations of a UID for maximum flexibility.
// It implements standard Go marshaling interfaces for seamless serialization.
type UIDData struct {
	// Raw is the original string representation of the UID.
	Raw string `json:"raw"`
	// Canonical is the standard formatted representation (e.g., with hyphens for UUID).
	Canonical string `json:"canonical"`
	// Bytes is the binary representation of the UID.
	Bytes []byte `json:"bytes"`
	// Hex is the hexadecimal string representation without separators.
	Hex string `json:"hex"`
	// Type indicates the UID type (ULID, UUID v1, v4, v6, v7, etc.).
	Type UIDType `json:"type"`
	// Timestamp is the embedded timestamp (if applicable to the UID type).
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Version is the UID version (applicable for UUIDs).
	Version *int `json:"version,omitempty"`
	// Variant is the UID variant (applicable for UUIDs).
	Variant *string `json:"variant,omitempty"`
}

// MarshalText implements encoding.TextMarshaler interface.
// Returns the canonical string representation of the UID.
func (u *UIDData) MarshalText() ([]byte, error) {
	return []byte(u.Canonical), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
// Parses the canonical string representation and updates the UIDData.
func (u *UIDData) UnmarshalText(text []byte) error {
	// This is a simplified implementation. In practice, you'd need
	// a way to detect the format and parse appropriately.
	u.Canonical = string(text)
	u.Raw = string(text)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
// Returns the binary representation of the UID.
func (u *UIDData) MarshalBinary() ([]byte, error) {
	if len(u.Bytes) == 0 {
		return nil, &MarshalError{
			Type:   u.Type,
			Reason: "no binary data available",
		}
	}
	return u.Bytes, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
// Parses the binary representation and updates the UIDData.
func (u *UIDData) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return &UnmarshalError{
			Type:   u.Type,
			Reason: "empty binary data",
		}
	}
	u.Bytes = make([]byte, len(data))
	copy(u.Bytes, data)
	return nil
}

// MarshalJSON implements json.Marshaler interface.
// Returns a JSON representation with all UID information.
func (u *UIDData) MarshalJSON() ([]byte, error) {
	type alias UIDData
	return json.Marshal((*alias)(u))
}

// UnmarshalJSON implements json.Unmarshaler interface.
// Parses JSON and updates the UIDData structure.
func (u *UIDData) UnmarshalJSON(data []byte) error {
	type alias UIDData
	return json.Unmarshal(data, (*alias)(u))
}

// String returns the canonical string representation of the UID.
// Implements the fmt.Stringer interface.
func (u *UIDData) String() string {
	return u.Canonical
}

// MarshalError represents an error during marshaling operations.
type MarshalError struct {
	Type   UIDType
	Reason string
	Cause  error
}

// Error implements the error interface.
func (e *MarshalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("marshal failed for %s: %s (caused by: %v)", e.Type, e.Reason, e.Cause)
	}
	return fmt.Sprintf("marshal failed for %s: %s", e.Type, e.Reason)
}

// Unwrap returns the underlying error.
func (e *MarshalError) Unwrap() error {
	return e.Cause
}

// UnmarshalError represents an error during unmarshaling operations.
type UnmarshalError struct {
	Type   UIDType
	Reason string
	Cause  error
}

// Error implements the error interface.
func (e *UnmarshalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("unmarshal failed for %s: %s (caused by: %v)", e.Type, e.Reason, e.Cause)
	}
	return fmt.Sprintf("unmarshal failed for %s: %s", e.Type, e.Reason)
}

// Unwrap returns the underlying error.
func (e *UnmarshalError) Unwrap() error {
	return e.Cause
}

// Generator defines the interface for UID generation operations.
// Implementations must provide thread-safe operations and handle context cancellation.
type Generator interface {
	// Generate creates a new UID with the default configuration.
	Generate(ctx context.Context) (*UIDData, error)

	// GenerateWithTimestamp creates a new UID with a specific timestamp.
	// Not all UID types support custom timestamps.
	GenerateWithTimestamp(ctx context.Context, timestamp time.Time) (*UIDData, error)

	// GetType returns the UID type this generator produces.
	GetType() UIDType

	// IsTimestampSupported indicates if the generator supports custom timestamps.
	IsTimestampSupported() bool
}

// Parser defines the interface for UID parsing and validation operations.
// Implementations must handle various input formats and provide detailed error information.
type Parser interface {
	// Parse attempts to parse a string into UIDData.
	// It should handle multiple input formats (canonical, hex, bytes).
	Parse(ctx context.Context, input string) (*UIDData, error)

	// ParseBytes attempts to parse a byte slice into UIDData.
	ParseBytes(ctx context.Context, input []byte) (*UIDData, error)

	// Validate checks if a string represents a valid UID of the supported type.
	Validate(ctx context.Context, input string) error

	// ValidateBytes checks if a byte slice represents a valid UID of the supported type.
	ValidateBytes(ctx context.Context, input []byte) error

	// GetSupportedTypes returns the UID types this parser can handle.
	GetSupportedTypes() []UIDType
}

// Converter defines the interface for UID format conversion operations.
// Implementations must provide lossless conversion between supported formats.
type Converter interface {
	// ToCanonical converts a UID to its canonical string representation.
	ToCanonical(ctx context.Context, uid *UIDData) (string, error)

	// ToHex converts a UID to its hexadecimal string representation.
	ToHex(ctx context.Context, uid *UIDData) (string, error)

	// ToBytes converts a UID to its byte slice representation.
	ToBytes(ctx context.Context, uid *UIDData) ([]byte, error)

	// ConvertType attempts to convert between compatible UID types.
	// Not all conversions are possible or meaningful.
	ConvertType(ctx context.Context, uid *UIDData, targetType UIDType) (*UIDData, error)

	// GetSupportedConversions returns a map of source to target type conversions supported.
	GetSupportedConversions() map[UIDType][]UIDType
}

// Marshaler defines the interface for UID serialization operations.
// Implementations must provide thread-safe marshaling to various formats.
type Marshaler interface {
	// MarshalText marshals the UID to a text representation (implements encoding.TextMarshaler).
	MarshalText(ctx context.Context, uid *UIDData) ([]byte, error)

	// MarshalBinary marshals the UID to a binary representation (implements encoding.BinaryMarshaler).
	MarshalBinary(ctx context.Context, uid *UIDData) ([]byte, error)

	// MarshalJSON marshals the UID to JSON format (implements json.Marshaler).
	MarshalJSON(ctx context.Context, uid *UIDData) ([]byte, error)
}

// Unmarshaler defines the interface for UID deserialization operations.
// Implementations must provide thread-safe unmarshaling from various formats.
type Unmarshaler interface {
	// UnmarshalText unmarshals a UID from text representation (implements encoding.TextUnmarshaler).
	UnmarshalText(ctx context.Context, data []byte) (*UIDData, error)

	// UnmarshalBinary unmarshals a UID from binary representation (implements encoding.BinaryUnmarshaler).
	UnmarshalBinary(ctx context.Context, data []byte) (*UIDData, error)

	// UnmarshalJSON unmarshals a UID from JSON format (implements json.Unmarshaler).
	UnmarshalJSON(ctx context.Context, data []byte) (*UIDData, error)
}

// Provider defines the comprehensive interface for UID operations.
// It combines generation, parsing, conversion, and serialization capabilities.
type Provider interface {
	Generator
	Parser
	Converter
	Marshaler
	Unmarshaler

	// GetName returns a human-readable name for this provider.
	GetName() string

	// GetVersion returns the version of this provider implementation.
	GetVersion() string

	// IsThreadSafe indicates if this provider can be used concurrently.
	IsThreadSafe() bool
}

// Factory defines the interface for creating UID providers.
// Implementations must provide thread-safe factory operations.
type Factory interface {
	// CreateProvider creates a new provider instance for the specified UID type.
	CreateProvider(ctx context.Context, uidType UIDType) (Provider, error)

	// GetSupportedTypes returns all UID types supported by this factory.
	GetSupportedTypes() []UIDType

	// GetDefaultType returns the default UID type for this factory.
	GetDefaultType() UIDType
}
