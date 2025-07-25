package interfaces

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxSize != 10*1024*1024 {
		t.Errorf("Expected max size 10MB, got %d", config.MaxSize)
	}

	if !config.StrictMode {
		t.Error("Expected strict mode to be true")
	}

	if config.AllowComments {
		t.Error("Expected allow comments to be false")
	}

	if config.Encoding != "utf-8" {
		t.Errorf("Expected encoding utf-8, got %s", config.Encoding)
	}
}

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "Error with line and column",
			err: &ParseError{
				Type:    ErrorTypeSyntax,
				Message: "Invalid syntax",
				Line:    5,
				Column:  10,
			},
			expected: "Invalid syntax at line \x05, column \n",
		},
		{
			name: "Error without line and column",
			err: &ParseError{
				Type:    ErrorTypeValidation,
				Message: "Validation failed",
			},
			expected: "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	cause := &ParseError{Message: "original error"}
	err := &ParseError{
		Message: "wrapped error",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Expected unwrapped error to be the original cause")
	}

	errWithoutCause := &ParseError{Message: "no cause"}
	if errWithoutCause.Unwrap() != nil {
		t.Error("Expected nil when no cause is set")
	}
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeUnknown, "unknown"},
		{ErrorTypeSyntax, "syntax"},
		{ErrorTypeValidation, "validation"},
		{ErrorTypeTimeout, "timeout"},
		{ErrorTypeSize, "size"},
		{ErrorTypeEncoding, "encoding"},
		{ErrorTypeIO, "io"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResult_Structure(t *testing.T) {
	data := "test data"
	metadata := &Metadata{
		ParsedAt:       time.Now(),
		Duration:       100 * time.Millisecond,
		BytesProcessed: 1024,
		ParserType:     "json",
	}
	warnings := []string{"warning 1", "warning 2"}

	result := &Result[string]{
		Data:     &data,
		Metadata: metadata,
		Warnings: warnings,
	}

	if *result.Data != data {
		t.Errorf("Expected data %q, got %q", data, *result.Data)
	}

	if result.Metadata != metadata {
		t.Error("Expected metadata to match")
	}

	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

func TestMetadata_Structure(t *testing.T) {
	now := time.Now()
	duration := 250 * time.Millisecond

	metadata := &Metadata{
		ParsedAt:       now,
		Duration:       duration,
		BytesProcessed: 2048,
		LinesProcessed: 100,
		ItemsProcessed: 50,
		ParserType:     "csv",
		Version:        "1.0.0",
	}

	if !metadata.ParsedAt.Equal(now) {
		t.Error("Expected ParsedAt to match")
	}

	if metadata.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, metadata.Duration)
	}

	if metadata.BytesProcessed != 2048 {
		t.Errorf("Expected bytes processed 2048, got %d", metadata.BytesProcessed)
	}

	if metadata.LinesProcessed != 100 {
		t.Errorf("Expected lines processed 100, got %d", metadata.LinesProcessed)
	}

	if metadata.ItemsProcessed != 50 {
		t.Errorf("Expected items processed 50, got %d", metadata.ItemsProcessed)
	}

	if metadata.ParserType != "csv" {
		t.Errorf("Expected parser type csv, got %s", metadata.ParserType)
	}

	if metadata.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", metadata.Version)
	}
}
