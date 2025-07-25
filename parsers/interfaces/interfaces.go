// Package interfaces provides core interfaces for the parsers module.
package interfaces

import (
	"context"
	"io"
	"time"
)

// Parser defines the core parsing interface.
type Parser[T any] interface {
	// Parse parses input data into type T.
	Parse(ctx context.Context, data []byte) (*T, error)
	// ParseString parses string input into type T.
	ParseString(ctx context.Context, input string) (*T, error)
	// ParseReader parses data from reader into type T.
	ParseReader(ctx context.Context, reader io.Reader) (*T, error)
	// Validate validates the parsed result.
	Validate(ctx context.Context, result *T) error
}

// StreamParser defines interface for streaming parsers.
type StreamParser[T any] interface {
	// ParseStream parses data from a stream with callback for each parsed item.
	ParseStream(ctx context.Context, reader io.Reader, callback func(*T) error) error
}

// Formatter defines interface for data formatting.
type Formatter[T any] interface {
	// Format formats data to bytes.
	Format(ctx context.Context, data *T) ([]byte, error)
	// FormatString formats data to string.
	FormatString(ctx context.Context, data *T) (string, error)
	// FormatWriter formats data to writer.
	FormatWriter(ctx context.Context, data *T, writer io.Writer) error
}

// Transformer defines interface for data transformation.
type Transformer[From, To any] interface {
	// Transform transforms data from one type to another.
	Transform(ctx context.Context, from *From) (*To, error)
}

// Validator defines interface for data validation.
type Validator[T any] interface {
	// Validate validates data structure.
	Validate(ctx context.Context, data *T) error
}

// ParserConfig defines configuration for parsers.
type ParserConfig struct {
	// Timeout for parsing operations.
	Timeout time.Duration
	// MaxSize limits the maximum input size.
	MaxSize int64
	// StrictMode enables strict parsing rules.
	StrictMode bool
	// AllowComments enables comment parsing where applicable.
	AllowComments bool
	// Encoding specifies the character encoding.
	Encoding string
}

// DefaultConfig returns default parser configuration.
func DefaultConfig() *ParserConfig {
	return &ParserConfig{
		Timeout:       30 * time.Second,
		MaxSize:       10 * 1024 * 1024, // 10MB
		StrictMode:    true,
		AllowComments: false,
		Encoding:      "utf-8",
	}
}

// ParseError represents parsing errors.
type ParseError struct {
	// Type of error.
	Type ErrorType
	// Message describes the error.
	Message string
	// Line where error occurred (if applicable).
	Line int
	// Column where error occurred (if applicable).
	Column int
	// Offset in bytes where error occurred.
	Offset int64
	// Context provides additional error context.
	Context string
	// Cause is the underlying error.
	Cause error
}

// Error implements error interface.
func (e *ParseError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return e.Message + " at line " + string(rune(e.Line)) + ", column " + string(rune(e.Column))
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ErrorType defines types of parsing errors.
type ErrorType int

const (
	// ErrorTypeUnknown represents unknown error.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeSyntax represents syntax error.
	ErrorTypeSyntax
	// ErrorTypeValidation represents validation error.
	ErrorTypeValidation
	// ErrorTypeTimeout represents timeout error.
	ErrorTypeTimeout
	// ErrorTypeSize represents size limit error.
	ErrorTypeSize
	// ErrorTypeEncoding represents encoding error.
	ErrorTypeEncoding
	// ErrorTypeIO represents I/O error.
	ErrorTypeIO
)

// String returns string representation of error type.
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeSyntax:
		return "syntax"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeSize:
		return "size"
	case ErrorTypeEncoding:
		return "encoding"
	case ErrorTypeIO:
		return "io"
	default:
		return "unknown"
	}
}

// Result represents parsing result with metadata.
type Result[T any] struct {
	// Data contains the parsed data.
	Data *T
	// Metadata contains parsing metadata.
	Metadata *Metadata
	// Warnings contains non-fatal parsing warnings.
	Warnings []string
}

// Metadata contains information about parsing operation.
type Metadata struct {
	// ParsedAt indicates when parsing was completed.
	ParsedAt time.Time
	// Duration indicates how long parsing took.
	Duration time.Duration
	// BytesProcessed indicates number of bytes processed.
	BytesProcessed int64
	// LinesProcessed indicates number of lines processed.
	LinesProcessed int
	// ItemsProcessed indicates number of items processed.
	ItemsProcessed int
	// ParserType indicates which parser was used.
	ParserType string
	// Version indicates parser version.
	Version string
}
