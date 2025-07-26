package providers

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/fsvxavier/nexs-lib/uid/config"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/internal"
)

// UUIDProvider implements the Provider interface for UUID generation and manipulation.
// It provides thread-safe operations for generating, parsing, and converting UUIDs.
type UUIDProvider struct {
	config  *config.UUIDConfig
	mutex   sync.RWMutex
	name    string
	version string
}

// NewUUIDProvider creates a new UUID provider with the specified configuration.
func NewUUIDProvider(cfg *config.UUIDConfig) (*UUIDProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid UUID configuration: %w", err)
	}

	provider := &UUIDProvider{
		config:  cfg,
		name:    cfg.Name,
		version: "1.0.0",
	}

	return provider, nil
}

// Generate creates a new UUID with the default configuration.
func (p *UUIDProvider) Generate(ctx context.Context) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	switch p.config.Version {
	case 1:
		return p.generateV1(ctx, time.Now())
	case 4:
		return p.generateV4(ctx)
	case 6:
		return p.generateV6(ctx, time.Now())
	case 7:
		return p.generateV7(ctx, time.Now())
	default:
		return nil, fmt.Errorf("unsupported UUID version: %d", p.config.Version)
	}
}

// GenerateWithTimestamp creates a new UUID with a specific timestamp.
func (p *UUIDProvider) GenerateWithTimestamp(ctx context.Context, timestamp time.Time) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if err := internal.ValidateTimestamp(timestamp); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	switch p.config.Version {
	case 1:
		return p.generateV1(ctx, timestamp)
	case 6:
		return p.generateV6(ctx, timestamp)
	case 7:
		return p.generateV7(ctx, timestamp)
	default:
		// For versions that don't support timestamps, ignore the timestamp
		return p.Generate(ctx)
	}
}

// generateV1 generates a UUID version 1 (timestamp-based).
func (p *UUIDProvider) generateV1(ctx context.Context, timestamp time.Time) (*interfaces.UIDData, error) {
	uuidValue, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v1: %w", err)
	}
	return p.createUIDData(uuidValue, interfaces.UIDTypeUUIDV1, &timestamp), nil
}

// generateV4 generates a UUID version 4 (random).
func (p *UUIDProvider) generateV4(ctx context.Context) (*interfaces.UIDData, error) {
	uuidValue := uuid.New()
	return p.createUIDData(uuidValue, interfaces.UIDTypeUUIDV4, nil), nil
}

// generateV6 generates a UUID version 6 (timestamp-ordered).
func (p *UUIDProvider) generateV6(ctx context.Context, timestamp time.Time) (*interfaces.UIDData, error) {
	uuidValue, err := uuid.NewV6()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v6: %w", err)
	}
	return p.createUIDData(uuidValue, interfaces.UIDTypeUUIDV6, &timestamp), nil
}

// generateV7 generates a UUID version 7 (timestamp and random).
func (p *UUIDProvider) generateV7(ctx context.Context, timestamp time.Time) (*interfaces.UIDData, error) {
	uuidValue, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v7: %w", err)
	}
	return p.createUIDData(uuidValue, interfaces.UIDTypeUUIDV7, &timestamp), nil
}

// createUIDData creates a UIDData structure from a UUID.
func (p *UUIDProvider) createUIDData(uuidValue uuid.UUID, uidType interfaces.UIDType, timestamp *time.Time) *interfaces.UIDData {
	bytes, _ := uuidValue.MarshalBinary()
	version, _ := internal.ExtractUUIDVersion(bytes)
	variant, _ := internal.ExtractUUIDVariant(bytes)

	return &interfaces.UIDData{
		Raw:       uuidValue.String(),
		Canonical: uuidValue.String(),
		Bytes:     bytes,
		Hex:       hex.EncodeToString(bytes),
		Type:      uidType,
		Timestamp: timestamp,
		Version:   &version,
		Variant:   &variant,
	}
}

// GetType returns the UID type this generator produces.
func (p *UUIDProvider) GetType() interfaces.UIDType {
	switch p.config.Version {
	case 1:
		return interfaces.UIDTypeUUIDV1
	case 4:
		return interfaces.UIDTypeUUIDV4
	case 6:
		return interfaces.UIDTypeUUIDV6
	case 7:
		return interfaces.UIDTypeUUIDV7
	default:
		return interfaces.UIDTypeUUIDV4
	}
}

// IsTimestampSupported indicates if the generator supports custom timestamps.
func (p *UUIDProvider) IsTimestampSupported() bool {
	return p.config.Version == 1 || p.config.Version == 6 || p.config.Version == 7
}

// Parse attempts to parse a string into UIDData.
func (p *UUIDProvider) Parse(ctx context.Context, input string) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if input == "" {
		return nil, &internal.ParseError{
			Input:   input,
			UIDType: p.GetType(),
			Reason:  "empty input",
		}
	}

	// Try to parse as UUID first
	if err := internal.ValidateUUIDString(input); err == nil {
		return p.parseUUIDString(ctx, input)
	}

	// Try to parse as hex if UUID validation fails
	return p.parseFromHex(ctx, input)
}

// parseUUIDString parses a standard UUID string.
func (p *UUIDProvider) parseUUIDString(ctx context.Context, input string) (*interfaces.UIDData, error) {
	uuidValue, err := uuid.Parse(input)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   input,
			UIDType: p.GetType(),
			Reason:  "failed to parse UUID",
			Cause:   err,
		}
	}

	bytes, _ := uuidValue.MarshalBinary()
	version, err := internal.ExtractUUIDVersion(bytes)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   input,
			UIDType: p.GetType(),
			Reason:  "failed to extract UUID version",
			Cause:   err,
		}
	}

	variant, _ := internal.ExtractUUIDVariant(bytes)

	// Determine the correct UID type based on version
	var uidType interfaces.UIDType
	switch version {
	case 1:
		uidType = interfaces.UIDTypeUUIDV1
	case 4:
		uidType = interfaces.UIDTypeUUIDV4
	case 6:
		uidType = interfaces.UIDTypeUUIDV6
	case 7:
		uidType = interfaces.UIDTypeUUIDV7
	default:
		uidType = interfaces.UIDTypeUUIDV4
	}

	// Extract timestamp for time-based UUIDs
	var timestamp *time.Time
	if version == 1 || version == 6 || version == 7 {
		if ts, err := p.extractTimestampFromUUID(bytes, version); err == nil {
			timestamp = &ts
		}
	}

	return &interfaces.UIDData{
		Raw:       input,
		Canonical: uuidValue.String(),
		Bytes:     bytes,
		Hex:       hex.EncodeToString(bytes),
		Type:      uidType,
		Timestamp: timestamp,
		Version:   &version,
		Variant:   &variant,
	}, nil
}

// parseFromHex attempts to parse a hex string as UUID bytes.
func (p *UUIDProvider) parseFromHex(ctx context.Context, hexStr string) (*interfaces.UIDData, error) {
	bytes, err := internal.HexToBytes(hexStr)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   hexStr,
			UIDType: p.GetType(),
			Reason:  "invalid hex format",
			Cause:   err,
		}
	}

	return p.ParseBytes(ctx, bytes)
}

// ParseBytes attempts to parse a byte slice into UIDData.
func (p *UUIDProvider) ParseBytes(ctx context.Context, input []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(input) != internal.UIDLength16 {
		return nil, &internal.ParseError{
			Input:   fmt.Sprintf("%d bytes", len(input)),
			UIDType: p.GetType(),
			Reason:  "invalid byte length for UUID",
		}
	}

	// Create UUID from bytes
	var uuidValue uuid.UUID
	if err := uuidValue.UnmarshalBinary(input); err != nil {
		return nil, &internal.ParseError{
			Input:   hex.EncodeToString(input),
			UIDType: p.GetType(),
			Reason:  "failed to unmarshal UUID bytes",
			Cause:   err,
		}
	}

	version, err := internal.ExtractUUIDVersion(input)
	if err != nil {
		return nil, &internal.ParseError{
			Input:   hex.EncodeToString(input),
			UIDType: p.GetType(),
			Reason:  "failed to extract UUID version",
			Cause:   err,
		}
	}

	variant, _ := internal.ExtractUUIDVariant(input)

	// Determine the correct UID type based on version
	var uidType interfaces.UIDType
	switch version {
	case 1:
		uidType = interfaces.UIDTypeUUIDV1
	case 4:
		uidType = interfaces.UIDTypeUUIDV4
	case 6:
		uidType = interfaces.UIDTypeUUIDV6
	case 7:
		uidType = interfaces.UIDTypeUUIDV7
	default:
		uidType = interfaces.UIDTypeUUIDV4
	}

	// Extract timestamp for time-based UUIDs
	var timestamp *time.Time
	if version == 1 || version == 6 || version == 7 {
		if ts, err := p.extractTimestampFromUUID(input, version); err == nil {
			timestamp = &ts
		}
	}

	return &interfaces.UIDData{
		Raw:       hex.EncodeToString(input),
		Canonical: uuidValue.String(),
		Bytes:     input,
		Hex:       hex.EncodeToString(input),
		Type:      uidType,
		Timestamp: timestamp,
		Version:   &version,
		Variant:   &variant,
	}, nil
}

// extractTimestampFromUUID extracts timestamp from time-based UUIDs.
func (p *UUIDProvider) extractTimestampFromUUID(bytes []byte, version int) (time.Time, error) {
	switch version {
	case 1:
		// UUID v1: 60-bit timestamp in 100-nanosecond intervals since 1582-10-15
		timeLow := binary.BigEndian.Uint32(bytes[0:4])
		timeMid := binary.BigEndian.Uint16(bytes[4:6])
		timeHigh := binary.BigEndian.Uint16(bytes[6:8]) & 0x0FFF // Remove version bits

		timestamp := uint64(timeHigh)<<48 | uint64(timeMid)<<32 | uint64(timeLow)

		// Convert from 100-nanosecond intervals since 1582-10-15 to Unix time
		const uuidEpoch = 122192928000000000 // 100-nanosecond intervals between 1582-10-15 and 1970-01-01
		unixNanos := (int64(timestamp) - uuidEpoch) * 100

		return time.Unix(0, unixNanos), nil

	case 6:
		// UUID v6: similar to v1 but reordered for better sorting
		timeHigh := binary.BigEndian.Uint32(bytes[0:4]) & 0x0FFFFFFF // Remove version bits
		timeMid := binary.BigEndian.Uint16(bytes[4:6])
		timeLow := binary.BigEndian.Uint16(bytes[6:8])

		timestamp := uint64(timeHigh)<<28 | uint64(timeMid)<<12 | uint64(timeLow)

		// Convert from 100-nanosecond intervals since 1582-10-15 to Unix time
		const uuidEpoch = 122192928000000000
		unixNanos := (int64(timestamp) - uuidEpoch) * 100

		return time.Unix(0, unixNanos), nil

	case 7:
		// UUID v7: 48-bit timestamp in milliseconds since Unix epoch
		timestampMs := binary.BigEndian.Uint64(bytes[0:8]) >> 16 // Take upper 48 bits

		return time.UnixMilli(int64(timestampMs)), nil

	default:
		return time.Time{}, fmt.Errorf("UUID version %d does not contain timestamp", version)
	}
}

// Validate checks if a string represents a valid UUID.
func (p *UUIDProvider) Validate(ctx context.Context, input string) error {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return err
	}

	return internal.ValidateUUIDString(input)
}

// ValidateBytes checks if a byte slice represents a valid UUID.
func (p *UUIDProvider) ValidateBytes(ctx context.Context, input []byte) error {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return err
	}

	if len(input) != internal.UIDLength16 {
		return &internal.ValidationError{
			Field:   "length",
			Value:   fmt.Sprintf("%d", len(input)),
			Reason:  "UUID must be 16 bytes",
			UIDType: p.GetType(),
		}
	}

	// Try to create UUID from bytes
	var uuidValue uuid.UUID
	if err := uuidValue.UnmarshalBinary(input); err != nil {
		return &internal.ValidationError{
			Field:   "bytes",
			Value:   hex.EncodeToString(input),
			Reason:  "invalid UUID bytes",
			UIDType: p.GetType(),
		}
	}

	return nil
}

// GetSupportedTypes returns the UID types this parser can handle.
func (p *UUIDProvider) GetSupportedTypes() []interfaces.UIDType {
	return []interfaces.UIDType{
		interfaces.UIDTypeUUIDV1,
		interfaces.UIDTypeUUIDV4,
		interfaces.UIDTypeUUIDV6,
		interfaces.UIDTypeUUIDV7,
	}
}

// ToCanonical converts a UID to its canonical string representation.
func (p *UUIDProvider) ToCanonical(ctx context.Context, uid *interfaces.UIDData) (string, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return "", err
	}

	if !p.isSupportedType(uid.Type) {
		return "", &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: p.GetType(),
			Reason:     "unsupported source type for UUID provider",
		}
	}

	return uid.Canonical, nil
}

// ToHex converts a UID to its hexadecimal string representation.
func (p *UUIDProvider) ToHex(ctx context.Context, uid *interfaces.UIDData) (string, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return "", err
	}

	if !p.isSupportedType(uid.Type) {
		return "", &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: p.GetType(),
			Reason:     "unsupported source type for UUID provider",
		}
	}

	return strings.ReplaceAll(uid.Hex, "-", ""), nil
}

// ToBytes converts a UID to its byte slice representation.
func (p *UUIDProvider) ToBytes(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if !p.isSupportedType(uid.Type) {
		return nil, &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: p.GetType(),
			Reason:     "unsupported source type for UUID provider",
		}
	}

	return uid.Bytes, nil
}

// ConvertType attempts to convert between compatible UID types.
func (p *UUIDProvider) ConvertType(ctx context.Context, uid *interfaces.UIDData, targetType interfaces.UIDType) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if !p.isSupportedType(targetType) {
		return nil, &internal.ConversionError{
			SourceType: uid.Type,
			TargetType: targetType,
			Reason:     "unsupported target type for UUID provider",
		}
	}

	if uid.Type == targetType {
		// Already the target type, return copy
		return &interfaces.UIDData{
			Raw:       uid.Raw,
			Canonical: uid.Canonical,
			Bytes:     uid.Bytes,
			Hex:       uid.Hex,
			Type:      uid.Type,
			Timestamp: uid.Timestamp,
			Version:   uid.Version,
			Variant:   uid.Variant,
		}, nil
	}

	// For UUID types, we can only change the type metadata, not the actual UUID
	if p.isSupportedType(uid.Type) {
		return &interfaces.UIDData{
			Raw:       uid.Raw,
			Canonical: uid.Canonical,
			Bytes:     uid.Bytes,
			Hex:       uid.Hex,
			Type:      targetType,
			Timestamp: uid.Timestamp,
			Version:   uid.Version,
			Variant:   uid.Variant,
		}, nil
	}

	return nil, &internal.ConversionError{
		SourceType: uid.Type,
		TargetType: targetType,
		Reason:     "unsupported conversion",
	}
}

// GetSupportedConversions returns a map of source to target type conversions supported.
func (p *UUIDProvider) GetSupportedConversions() map[interfaces.UIDType][]interfaces.UIDType {
	supportedTypes := p.GetSupportedTypes()
	conversions := make(map[interfaces.UIDType][]interfaces.UIDType)

	for _, sourceType := range supportedTypes {
		conversions[sourceType] = supportedTypes
	}

	return conversions
}

// GetName returns a human-readable name for this provider.
func (p *UUIDProvider) GetName() string {
	return p.name
}

// GetVersion returns the version of this provider implementation.
func (p *UUIDProvider) GetVersion() string {
	return p.version
}

// IsThreadSafe indicates if this provider can be used concurrently.
func (p *UUIDProvider) IsThreadSafe() bool {
	return p.config.ThreadSafe
}

// isSupportedType checks if the given UID type is supported by this provider.
func (p *UUIDProvider) isSupportedType(uidType interfaces.UIDType) bool {
	supportedTypes := p.GetSupportedTypes()
	for _, supported := range supportedTypes {
		if supported == uidType {
			return true
		}
	}
	return false
}

// ExtractTimestamp extracts the timestamp from a time-based UUID string.
func (p *UUIDProvider) ExtractTimestamp(ctx context.Context, uuidStr string) (time.Time, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return time.Time{}, err
	}

	uidData, err := p.Parse(ctx, uuidStr)
	if err != nil {
		return time.Time{}, err
	}

	if uidData.Timestamp == nil {
		return time.Time{}, &internal.ParseError{
			Input:   uuidStr,
			UIDType: uidData.Type,
			Reason:  "UUID does not contain timestamp",
		}
	}

	return *uidData.Timestamp, nil
}

// IsValidUUID checks if a string represents a valid UUID.
func (p *UUIDProvider) IsValidUUID(ctx context.Context, input string) bool {
	return p.Validate(ctx, input) == nil
}

// MarshalText marshals the UID to a text representation.
func (p *UUIDProvider) MarshalText(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if !p.isSupportedType(uid.Type) {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for UUID provider",
		}
	}

	return []byte(uid.Canonical), nil
}

// MarshalBinary marshals the UID to a binary representation.
func (p *UUIDProvider) MarshalBinary(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if !p.isSupportedType(uid.Type) {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for UUID provider",
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
func (p *UUIDProvider) MarshalJSON(ctx context.Context, uid *interfaces.UIDData) ([]byte, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if !p.isSupportedType(uid.Type) {
		return nil, &interfaces.MarshalError{
			Type:   uid.Type,
			Reason: "unsupported type for UUID provider",
		}
	}

	return uid.MarshalJSON()
}

// UnmarshalText unmarshals a UID from text representation.
func (p *UUIDProvider) UnmarshalText(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   p.GetType(),
			Reason: "empty text data",
		}
	}

	return p.Parse(ctx, string(data))
}

// UnmarshalBinary unmarshals a UID from binary representation.
func (p *UUIDProvider) UnmarshalBinary(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   p.GetType(),
			Reason: "empty binary data",
		}
	}

	return p.ParseBytes(ctx, data)
}

// UnmarshalJSON unmarshals a UID from JSON format.
func (p *UUIDProvider) UnmarshalJSON(ctx context.Context, data []byte) (*interfaces.UIDData, error) {
	if err := internal.IsContextCanceled(ctx); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &interfaces.UnmarshalError{
			Type:   p.GetType(),
			Reason: "empty JSON data",
		}
	}

	var uidData interfaces.UIDData
	if err := uidData.UnmarshalJSON(data); err != nil {
		return nil, &interfaces.UnmarshalError{
			Type:   p.GetType(),
			Reason: "failed to unmarshal JSON",
			Cause:  err,
		}
	}

	// Validate that it's actually a supported UUID type
	if !p.isSupportedType(uidData.Type) {
		return nil, &interfaces.UnmarshalError{
			Type:   p.GetType(),
			Reason: "JSON contains unsupported UUID type",
		}
	}

	// Re-parse to ensure consistency
	return p.Parse(ctx, uidData.Canonical)
}
