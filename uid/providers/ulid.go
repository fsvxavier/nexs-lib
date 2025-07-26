// Package providers implements concrete UID providers for different UID types.
// This package contains the actual implementations of UID generators, parsers, and converters.
package providers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/fsvxavier/nexs-lib/uid/config"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/internal"
)

// ULIDProvider implements the Provider interface for ULID generation and manipulation.
// It provides thread-safe operations for generating, parsing, and converting ULIDs.
type ULIDProvider struct {
	config    *config.ULIDConfig
	entropy   io.Reader
	monotonic *ulid.MonotonicEntropy
	mutex     sync.RWMutex
	name      string
	version   string
}

// NewULIDProvider creates a new ULID provider with the specified configuration.
func NewULIDProvider(cfg *config.ULIDConfig) (*ULIDProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ULID configuration: %w", err)
	}

	provider := &ULIDProvider{
		config:  cfg,
		name:    cfg.Name,
		version: "1.0.0",
	}

	// Initialize entropy source
	if cfg.CustomEntropy {
		provider.entropy = rand.Reader
	} else {
		provider.entropy = ulid.DefaultEntropy()
	}

	// Initialize monotonic entropy if enabled
	if cfg.MonotonicMode {
		provider.monotonic = ulid.Monotonic(provider.entropy, 0)
		provider.entropy = provider.monotonic
	}

	return provider, nil
}

// Generate creates a new ULID with the default configuration.
func (p *ULIDProvider) Generate(ctx context.Context) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	return p.GenerateWithTimestamp(ctx, time.Now())
}

// GenerateWithTimestamp creates a new ULID with a specific timestamp.
func (p *ULIDProvider) GenerateWithTimestamp(ctx context.Context, timestamp time.Time) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if err := internal.ValidateTimestamp(timestamp); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Generate ULID
	ulidValue, err := ulid.New(ulid.Timestamp(timestamp), p.entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ULID: %w", err)
	}

	// Create UIDData
	uidData := &interfaces.UIDData{
		Raw:       ulidValue.String(),
		Canonical: ulidValue.String(),
		Bytes:     ulidValue.Bytes(),
		Hex:       hex.EncodeToString(ulidValue.Bytes()),
		Type:      interfaces.UIDTypeULID,
		Timestamp: &timestamp,
	}

	return uidData, nil
}

// GetType returns the UID type this generator produces.
func (p *ULIDProvider) GetType() interfaces.UIDType {
	return interfaces.UIDTypeULID
}

// IsTimestampSupported indicates if the generator supports custom timestamps.
func (p *ULIDProvider) IsTimestampSupported() bool {
	return true
}

// Parse attempts to parse a string into UIDData.
func (p *ULIDProvider) Parse(ctx context.Context, input string) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if input == "" {
		return nil, &internal.ParseError{
			Input:   input,
			UIDType: interfaces.UIDTypeULID,
			Reason:  "empty input",
		}
	}

	// Validate ULID format
	if err := internal.ValidateULIDString(input); err != nil {
		// Try to parse as hex if ULID validation fails
		return p.parseFromHex(ctx, input)
	}

	// Parse ULID
	ulidValue, err := ulid.Parse(input)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   input,
			UIDType: interfaces.UIDTypeULID,
			Reason:  "failed to parse ULID",
			Cause:   err,
		}
	}

	// Extract timestamp
	timestamp := time.UnixMilli(int64(ulidValue.Time()))

	// Create UIDData
	uidData := &interfaces.UIDData{
		Raw:       input,
		Canonical: ulidValue.String(),
		Bytes:     ulidValue.Bytes(),
		Hex:       hex.EncodeToString(ulidValue.Bytes()),
		Type:      interfaces.UIDTypeULID,
		Timestamp: &timestamp,
	}

	return uidData, nil
}

// parseFromHex attempts to parse a hex string as ULID bytes.
func (p *ULIDProvider) parseFromHex(ctx context.Context, hexStr string) (*interfaces.UIDData, error) {
	bytes, err := internal.HexToBytes(hexStr)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   hexStr,
			UIDType: interfaces.UIDTypeULID,
			Reason:  "invalid hex format",
			Cause:   err,
		}
	}

	return p.ParseBytes(ctx, bytes)
}

// ParseBytes attempts to parse a byte slice into UIDData.
func (p *ULIDProvider) ParseBytes(ctx context.Context, input []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(input) != internal.UIDLength16 {
		return nil, &internal.ParseError{
			Input:   fmt.Sprintf("%d bytes", len(input)),
			UIDType: interfaces.UIDTypeULID,
			Reason:  "invalid byte length for ULID",
		}
	}

	// Create ULID from bytes
	var ulidValue ulid.ULID
	if err := ulidValue.UnmarshalBinary(input); err != nil {
		return nil, &internal.ParseError{
			Input:   hex.EncodeToString(input),
			UIDType: interfaces.UIDTypeULID,
			Reason:  "failed to unmarshal ULID bytes",
			Cause:   err,
		}
	}

	// Extract timestamp
	timestamp := time.UnixMilli(int64(ulidValue.Time()))

	// Create UIDData
	uidData := &interfaces.UIDData{
		Raw:       hex.EncodeToString(input),
		Canonical: ulidValue.String(),
		Bytes:     input,
		Hex:       hex.EncodeToString(input),
		Type:      interfaces.UIDTypeULID,
		Timestamp: &timestamp,
	}

	return uidData, nil
}

// Validate checks if a string represents a valid ULID.
func (p *ULIDProvider) Validate(ctx context.Context, input string) error {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return err
	}

	return internal.ValidateULIDString(input)
}

// ValidateBytes checks if a byte slice represents a valid ULID.
func (p *ULIDProvider) ValidateBytes(ctx context.Context, input []byte) error {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return err
	}

	if len(input) != internal.UIDLength16 {
		return &internal.ValidationError{
			Field:   "length",
			Value:   fmt.Sprintf("%d", len(input)),
			Reason:  "ULID must be 16 bytes",
			UIDType: interfaces.UIDTypeULID,
		}
	}

	// Try to create ULID from bytes
	var ulidValue ulid.ULID
	if err := ulidValue.UnmarshalBinary(input); err != nil {
		return &internal.ValidationError{
			Field:   "bytes",
			Value:   hex.EncodeToString(input),
			Reason:  "invalid ULID bytes",
			UIDType: interfaces.UIDTypeULID,
		}
	}

	return nil
}

// GetSupportedTypes returns the UID types this parser can handle.
func (p *ULIDProvider) GetSupportedTypes() []interfaces.UIDType {
	return []interfaces.UIDType{interfaces.UIDTypeULID}
}

// ToCanonical converts a UID to its canonical string representation.
func (p *ULIDProvider) ToCanonical(ctx context.Context, uid *interfaces.UIDData) (string, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return "", err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return "", &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: interfaces.UIDTypeULID,
			Reason:     "unsupported source type for ULID provider",
		}
	}

	return uid.Canonical, nil
}

// ToHex converts a UID to its hexadecimal string representation.
func (p *ULIDProvider) ToHex(ctx context.Context, uid *interfaces.UIDData) (string, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return "", err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return "", &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: interfaces.UIDTypeULID,
			Reason:     "unsupported source type for ULID provider",
		}
	}

	return uid.Hex, nil
}

// ToBytes converts a UID to its byte slice representation.
func (p *ULIDProvider) ToBytes(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return nil, &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: interfaces.UIDTypeULID,
			Reason:     "unsupported source type for ULID provider",
		}
	}

	return uid.Bytes, nil
}

// ConvertType attempts to convert between compatible UID types.
func (p *ULIDProvider) ConvertType(ctx context.Context, uid *interfaces.UIDData, targetType interfaces.UIDType) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	// ULID provider only supports ULID type
	if targetType != interfaces.UIDTypeULID {
		return nil, &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: targetType,
			Reason:     "ULID provider only supports ULID type",
		}
	}

	if uid.Type == interfaces.UIDTypeULID {
		// Already the target type, return copy
		return &interfaces.UIDData{
			Raw:       uid.Raw,
			Canonical: uid.Canonical,
			Bytes:     uid.Bytes,
			Hex:       uid.Hex,
			Type:      uid.Type,
			Timestamp: uid.Timestamp,
		}, nil
	}

	return nil, &internal.ConversionError{
		SourceType: uid.Type,
		TargetType: targetType,
		Reason:     "unsupported conversion",
	}
}

// GetSupportedConversions returns a map of source to target type conversions supported.
func (p *ULIDProvider) GetSupportedConversions() map[interfaces.UIDType][]interfaces.UIDType {
	return map[interfaces.UIDType][]interfaces.UIDType{
		interfaces.UIDTypeULID: {interfaces.UIDTypeULID},
	}
}

// GetName returns a human-readable name for this provider.
func (p *ULIDProvider) GetName() string {
	return p.name
}

// GetVersion returns the version of this provider implementation.
func (p *ULIDProvider) GetVersion() string {
	return p.version
}

// IsThreadSafe indicates if this provider can be used concurrently.
func (p *ULIDProvider) IsThreadSafe() bool {
	return p.config.ThreadSafe
}

// ExtractTimestamp extracts the timestamp from a ULID string.
func (p *ULIDProvider) ExtractTimestamp(ctx context.Context, ulidStr string) (time.Time, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return time.Time{}, err
	}

	uidData, err := p.Parse(ctx, ulidStr)
	if err != nil {
		return time.Time{}, err
	}

	if uidData.Timestamp == nil {
		return time.Time{}, &internal.ParseError{
			Input:   ulidStr,
			UIDType: interfaces.UIDTypeULID,
			Reason:  "no timestamp in ULID data",
		}
	}

	return *uidData.Timestamp, nil
}

// IsValidULID checks if a string represents a valid ULID and contains a valid timestamp.
func (p *ULIDProvider) IsValidULID(ctx context.Context, input string) bool {
	if err := p.Validate(ctx, input); err != nil {
		return false
	}

	// Additional check for valid timestamp
	timestamp, err := p.ExtractTimestamp(ctx, input)
	if err != nil {
		return false
	}

	// Check if timestamp is reasonable (not before Unix epoch)
	return timestamp.Year() >= internal.UnixEpochYear
}

// MarshalText marshals the UID to a text representation.
func (p *ULIDProvider) MarshalText(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for ULID provider",
		}
	}

	return []byte(uid.Canonical), nil
}

// MarshalBinary marshals the UID to a binary representation.
func (p *ULIDProvider) MarshalBinary(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for ULID provider",
		}
	}

	if len(uid.Bytes) == 0 {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "no binary data available",
		}
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(uid.Bytes))
	copy(result, uid.Bytes)
	return result, nil
}

// MarshalJSON marshals the UID to JSON format.
func (p *ULIDProvider) MarshalJSON(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if uid.Type != interfaces.UIDTypeULID {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for ULID provider",
		}
	}

	return uid.MarshalJSON()
}

// UnmarshalText unmarshals a UID from text representation.
func (p *ULIDProvider) UnmarshalText(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   interfaces.UIDTypeULID,
			Reason: "empty text data",
		}
	}

	return p.Parse(ctx, string(data))
}

// UnmarshalBinary unmarshals a UID from binary representation.
func (p *ULIDProvider) UnmarshalBinary(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   interfaces.UIDTypeULID,
			Reason: "empty binary data",
		}
	}

	return p.ParseBytes(ctx, data)
}

// UnmarshalJSON unmarshals a UID from JSON format.
func (p *ULIDProvider) UnmarshalJSON(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   interfaces.UIDTypeULID,
			Reason: "empty JSON data",
		}
	}

	var uidData interfaces.UIDData
	if err := uidData.UnmarshalJSON(data); err != nil {
		return nil, &interfaces.UnmarshalError{
			Type:   interfaces.UIDTypeULID,
			Reason: "failed to unmarshal JSON",
			Cause:  err,
		}
	}

	// Validate that it's actually a ULID
	if uidData.Type != interfaces.UIDTypeULID {
		return nil, &interfaces.UnmarshalError{
			Type:   interfaces.UIDTypeULID,
			Reason: "JSON contains non-ULID data",
		}
	}

	// Re-parse to ensure consistency
	return p.Parse(ctx, uidData.Canonical)
}
