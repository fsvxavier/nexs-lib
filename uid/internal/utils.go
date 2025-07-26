// Package internal provides internal utilities and helper functions for UID operations.
// This package is not part of the public API and should only be used internally.
package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
)

const (
	// Standard UID lengths in bytes
	UIDLength16 = 16 // Standard UUID/ULID length
	UIDLength32 = 32 // Extended UID length

	// String representation lengths
	UUIDStringLength = 36 // UUID with hyphens: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	UUIDHexLength    = 32 // UUID hex without hyphens
	ULIDStringLength = 26 // ULID base32 encoding

	// Timestamp boundaries
	UnixEpochYear   = 1970
	InvalidUnixYear = 1969
	MaxValidYear    = 2200
)

var (
	// Compiled regex patterns for validation
	uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	hexRegex  = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	ulidRegex = regexp.MustCompile(`^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$`)
)

// ValidationError represents a UID validation error with detailed information.
type ValidationError struct {
	Field   string
	Value   string
	Reason  string
	UIDType interfaces.UIDType
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s (%s): %s - %s", e.UIDType, e.Field, e.Value, e.Reason)
}

// ParseError represents a UID parsing error with detailed information.
type ParseError struct {
	Input   string
	UIDType interfaces.UIDType
	Reason  string
	Cause   error
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("parse failed for %s: %s - %s (caused by: %v)", e.UIDType, e.Input, e.Reason, e.Cause)
	}
	return fmt.Sprintf("parse failed for %s: %s - %s", e.UIDType, e.Input, e.Reason)
}

// Unwrap returns the underlying error.
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ConversionError represents a UID conversion error with detailed information.
type ConversionError struct {
	SourceType interfaces.UIDType
	TargetType interfaces.UIDType
	Reason     string
	Cause      error
}

// Error implements the error interface.
func (e *ConversionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("conversion failed from %s to %s: %s (caused by: %v)", e.SourceType, e.TargetType, e.Reason, e.Cause)
	}
	return fmt.Sprintf("conversion failed from %s to %s: %s", e.SourceType, e.TargetType, e.Reason)
}

// Unwrap returns the underlying error.
func (e *ConversionError) Unwrap() error {
	return e.Cause
}

// FormatDetector analyzes input to determine its likely UID format.
type FormatDetector struct{}

// NewFormatDetector creates a new format detector instance.
func NewFormatDetector() *FormatDetector {
	return &FormatDetector{}
}

// DetectFormat attempts to determine the UID format from input string.
func (d *FormatDetector) DetectFormat(input string) (interfaces.UIDType, error) {
	if input == "" {
		return "", &ValidationError{
			Field:  "input",
			Value:  input,
			Reason: "empty string",
		}
	}

	// Remove whitespace and convert to uppercase for ULID check
	cleaned := strings.TrimSpace(strings.ToUpper(input))

	// Check for ULID format (26 characters, base32)
	if len(cleaned) == ULIDStringLength && ulidRegex.MatchString(cleaned) {
		return interfaces.UIDTypeULID, nil
	}

	// Check for UUID format with hyphens
	if len(input) == UUIDStringLength && uuidRegex.MatchString(input) {
		// Try to determine UUID version from the version field
		versionChar := input[14] // Position of version in UUID string
		switch versionChar {
		case '1':
			return interfaces.UIDTypeUUIDV1, nil
		case '4':
			return interfaces.UIDTypeUUIDV4, nil
		case '6':
			return interfaces.UIDTypeUUIDV6, nil
		case '7':
			return interfaces.UIDTypeUUIDV7, nil
		default:
			// Default to v4 for unknown versions
			return interfaces.UIDTypeUUIDV4, nil
		}
	}

	// Check for hex format (32 characters, no hyphens)
	cleanedLower := strings.ToLower(strings.ReplaceAll(input, "-", ""))
	if len(cleanedLower) == UUIDHexLength && hexRegex.MatchString(cleanedLower) {
		// Default to UUID v4 for hex input
		return interfaces.UIDTypeUUIDV4, nil
	}

	return "", &ValidationError{
		Field:  "input",
		Value:  input,
		Reason: "unrecognized format",
	}
}

// BytesToHex converts a byte slice to a hexadecimal string.
func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToBytes converts a hexadecimal string to a byte slice.
func HexToBytes(hexStr string) ([]byte, error) {
	// Remove any hyphens and convert to lowercase
	cleaned := strings.ToLower(strings.ReplaceAll(hexStr, "-", ""))

	if len(cleaned)%2 != 0 {
		return nil, &ValidationError{
			Field:  "hex",
			Value:  hexStr,
			Reason: "odd length hex string",
		}
	}

	if !hexRegex.MatchString(cleaned) {
		return nil, &ValidationError{
			Field:  "hex",
			Value:  hexStr,
			Reason: "invalid hex characters",
		}
	}

	data, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, &ParseError{
			Input:  hexStr,
			Reason: "hex decode failed",
			Cause:  err,
		}
	}

	return data, nil
}

// FormatUUIDString formats a byte slice as a standard UUID string with hyphens.
func FormatUUIDString(data []byte) (string, error) {
	if len(data) != UIDLength16 {
		return "", &ValidationError{
			Field:  "data",
			Value:  fmt.Sprintf("%d bytes", len(data)),
			Reason: "invalid length for UUID",
		}
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		data[0:4], data[4:6], data[6:8], data[8:10], data[10:16]), nil
}

// ValidateUUIDString validates a UUID string format.
func ValidateUUIDString(input string) error {
	if len(input) != UUIDStringLength {
		return &ValidationError{
			Field:   "length",
			Value:   fmt.Sprintf("%d", len(input)),
			Reason:  "UUID string must be 36 characters",
			UIDType: interfaces.UIDTypeUUIDV4,
		}
	}

	if !uuidRegex.MatchString(input) {
		return &ValidationError{
			Field:   "format",
			Value:   input,
			Reason:  "invalid UUID format",
			UIDType: interfaces.UIDTypeUUIDV4,
		}
	}

	return nil
}

// ValidateULIDString validates a ULID string format.
func ValidateULIDString(input string) error {
	if len(input) != ULIDStringLength {
		return &ValidationError{
			Field:   "length",
			Value:   fmt.Sprintf("%d", len(input)),
			Reason:  "ULID string must be 26 characters",
			UIDType: interfaces.UIDTypeULID,
		}
	}

	// Convert to uppercase for validation
	upper := strings.ToUpper(input)
	if !ulidRegex.MatchString(upper) {
		return &ValidationError{
			Field:   "format",
			Value:   input,
			Reason:  "invalid ULID format",
			UIDType: interfaces.UIDTypeULID,
		}
	}

	return nil
}

// ValidateTimestamp validates that a timestamp is within reasonable bounds.
func ValidateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return &ValidationError{
			Field:  "timestamp",
			Value:  timestamp.String(),
			Reason: "zero timestamp",
		}
	}

	if timestamp.Year() < UnixEpochYear {
		return &ValidationError{
			Field:  "timestamp",
			Value:  timestamp.String(),
			Reason: "timestamp before Unix epoch",
		}
	}

	if timestamp.Year() > MaxValidYear {
		return &ValidationError{
			Field:  "timestamp",
			Value:  timestamp.String(),
			Reason: "timestamp too far in future",
		}
	}

	return nil
}

// SecureRandomBytes generates cryptographically secure random bytes.
func SecureRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, &ValidationError{
			Field:  "length",
			Value:  fmt.Sprintf("%d", length),
			Reason: "length must be positive",
		}
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return bytes, nil
}

// ExtractUUIDVersion extracts the version number from UUID bytes.
func ExtractUUIDVersion(data []byte) (int, error) {
	if len(data) != UIDLength16 {
		return 0, &ValidationError{
			Field:  "data",
			Value:  fmt.Sprintf("%d bytes", len(data)),
			Reason: "invalid length for UUID",
		}
	}

	// Version is stored in the upper 4 bits of the 7th byte (index 6)
	version := int((data[6] & 0xF0) >> 4)

	if version < 1 || version > 7 {
		return 0, &ValidationError{
			Field:   "version",
			Value:   fmt.Sprintf("%d", version),
			Reason:  "invalid UUID version",
			UIDType: interfaces.UIDTypeUUIDV4,
		}
	}

	return version, nil
}

// ExtractUUIDVariant extracts the variant from UUID bytes.
func ExtractUUIDVariant(data []byte) (string, error) {
	if len(data) != UIDLength16 {
		return "", &ValidationError{
			Field:  "data",
			Value:  fmt.Sprintf("%d bytes", len(data)),
			Reason: "invalid length for UUID",
		}
	}

	// Variant is stored in the upper bits of the 9th byte (index 8)
	variantBits := data[8] & 0xE0 // Upper 3 bits

	switch {
	case (variantBits & 0x80) == 0x00:
		return "reserved_ncs", nil
	case (variantBits & 0xC0) == 0x80:
		return "rfc4122", nil
	case (variantBits & 0xE0) == 0xC0:
		return "reserved_microsoft", nil
	case (variantBits & 0xE0) == 0xE0:
		return "reserved_future", nil
	default:
		return "unknown", &ValidationError{
			Field:   "variant",
			Value:   fmt.Sprintf("0x%02X", variantBits),
			Reason:  "invalid UUID variant",
			UIDType: interfaces.UIDTypeUUIDV4,
		}
	}
}

// IsContextCanceled checks if the context has been canceled and returns an appropriate error.
func IsContextCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
