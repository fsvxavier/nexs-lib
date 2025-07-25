// Package datetime provides advanced date and time parsing functionality.
package datetime

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

// ParsedDateTime represents a parsed date/time with metadata.
type ParsedDateTime struct {
	Time      time.Time
	Layout    string
	IsUTC     bool
	Precision string // year, month, day, hour, minute, second, nanosecond
	Original  string
}

// Parser implements datetime parsing with format detection.
type Parser struct {
	config *interfaces.ParserConfig
}

// NewParser creates a new datetime parser with default configuration.
func NewParser() *Parser {
	return &Parser{
		config: interfaces.DefaultConfig(),
	}
}

// NewParserWithConfig creates a new datetime parser with custom configuration.
func NewParserWithConfig(config *interfaces.ParserConfig) *Parser {
	return &Parser{
		config: config,
	}
}

// Parse implements interfaces.Parser.
func (p *Parser) Parse(ctx context.Context, data []byte) (*ParsedDateTime, error) {
	return p.ParseString(ctx, string(data))
}

// ParseString parses a datetime string with automatic format detection.
func (p *Parser) ParseString(ctx context.Context, input string) (*ParsedDateTime, error) {
	if err := p.validateInput(input); err != nil {
		return nil, err
	}

	input = strings.TrimSpace(input)
	original := input

	// Try common formats first for performance
	commonFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"01/02/2006 15:04",
		"02/01/2006",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"Jan 2, 2006",
		"Jan 2, 2006 15:04:05",
		"January 2, 2006",
		"January 2, 2006 15:04:05",
	}

	for _, format := range commonFormats {
		if t, err := time.Parse(format, input); err == nil {
			return &ParsedDateTime{
				Time:      t,
				Layout:    format,
				IsUTC:     t.Location() == time.UTC,
				Precision: p.detectPrecision(format),
				Original:  original,
			}, nil
		}
	}

	// Advanced parsing with preprocessing
	return p.parseAdvanced(ctx, input, original)
}

// parseAdvanced performs advanced parsing with normalization and pattern matching.
func (p *Parser) parseAdvanced(ctx context.Context, input, original string) (*ParsedDateTime, error) {
	// Normalize the input
	normalized := p.normalizeInput(input)

	// Pattern-based parsing
	patterns := []struct {
		regex  *regexp.Regexp
		format string
	}{
		{regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?$`), time.RFC3339Nano},
		{regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`), "2006-01-02 15:04:05"},
		{regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`), "2006-01-02"},
		{regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`), "01/02/2006"},
		{regexp.MustCompile(`^\d{2}/\d{2}/\d{4} \d{2}:\d{2}$`), "01/02/2006 15:04"},
	}

	for _, pattern := range patterns {
		if pattern.regex.MatchString(normalized) {
			if t, err := time.Parse(pattern.format, normalized); err == nil {
				return &ParsedDateTime{
					Time:      t,
					Layout:    pattern.format,
					IsUTC:     t.Location() == time.UTC,
					Precision: p.detectPrecision(pattern.format),
					Original:  original,
				}, nil
			}
		}
	}

	return nil, &interfaces.ParseError{
		Type:    interfaces.ErrorTypeSyntax,
		Message: fmt.Sprintf("unable to parse datetime: %s", original),
	}
}

// normalizeInput normalizes input for better parsing.
func (p *Parser) normalizeInput(input string) string {
	// Remove extra whitespace
	input = strings.TrimSpace(input)
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	// Convert common separators
	input = strings.ReplaceAll(input, ".", "/")
	input = strings.ReplaceAll(input, "-", "/")

	// Handle month names
	monthMap := map[string]string{
		"january": "01", "jan": "01",
		"february": "02", "feb": "02",
		"march": "03", "mar": "03",
		"april": "04", "apr": "04",
		"may":  "05",
		"june": "06", "jun": "06",
		"july": "07", "jul": "07",
		"august": "08", "aug": "08",
		"september": "09", "sep": "09", "sept": "09",
		"october": "10", "oct": "10",
		"november": "11", "nov": "11",
		"december": "12", "dec": "12",
	}

	for month, num := range monthMap {
		input = regexp.MustCompile(`(?i)\b`+month+`\b`).ReplaceAllString(input, num)
	}

	return input
}

// detectPrecision determines the precision level of the parsed datetime.
func (p *Parser) detectPrecision(layout string) string {
	if strings.Contains(layout, ".000000000") || strings.Contains(layout, ".999999999") {
		return "nanosecond"
	}
	if strings.Contains(layout, ":05") {
		return "second"
	}
	if strings.Contains(layout, ":04") {
		return "minute"
	}
	if strings.Contains(layout, "15") {
		return "hour"
	}
	if strings.Contains(layout, "02") || strings.Contains(layout, "2") {
		return "day"
	}
	if strings.Contains(layout, "01") || strings.Contains(layout, "1") {
		return "month"
	}
	return "year"
}

// validateInput validates the input string.
func (p *Parser) validateInput(input string) error {
	if len(input) == 0 {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input datetime string is empty",
		}
	}

	if p.config.MaxSize > 0 && int64(len(input)) > p.config.MaxSize {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSize,
			Message: fmt.Sprintf("datetime string length %d exceeds maximum %d", len(input), p.config.MaxSize),
		}
	}

	// Basic validation - must contain digits
	hasDigit := false
	for _, r := range input {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}

	if !hasDigit {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "datetime string must contain digits",
		}
	}

	return nil
}

// Formatter implements datetime formatting.
type Formatter struct {
	defaultLayout string
}

// NewFormatter creates a new datetime formatter.
func NewFormatter() *Formatter {
	return &Formatter{
		defaultLayout: time.RFC3339,
	}
}

// NewFormatterWithLayout creates a new datetime formatter with custom layout.
func NewFormatterWithLayout(layout string) *Formatter {
	return &Formatter{
		defaultLayout: layout,
	}
}

// Format implements interfaces.Formatter.
func (f *Formatter) Format(ctx context.Context, data *ParsedDateTime) ([]byte, error) {
	if data == nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "data cannot be nil",
		}
	}

	layout := f.defaultLayout
	if data.Layout != "" {
		layout = data.Layout
	}

	formatted := data.Time.Format(layout)
	return []byte(formatted), nil
}

// FormatString implements interfaces.Formatter.
func (f *Formatter) FormatString(ctx context.Context, data *ParsedDateTime) (string, error) {
	result, err := f.Format(ctx, data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// FormatWriter is not commonly used for datetime, returns error.
func (f *Formatter) FormatWriter(ctx context.Context, data *ParsedDateTime, writer interface{}) error {
	return &interfaces.ParseError{
		Type:    interfaces.ErrorTypeValidation,
		Message: "FormatWriter not supported for datetime formatter",
	}
}

// Utility functions

// ParseDatetime parses a datetime string with automatic format detection.
func ParseDatetime(input string) (*ParsedDateTime, error) {
	parser := NewParser()
	return parser.ParseString(context.Background(), input)
}

// ParseDatetimeInLocation parses a datetime string in a specific location.
func ParseDatetimeInLocation(input string, location *time.Location) (*ParsedDateTime, error) {
	parsed, err := ParseDatetime(input)
	if err != nil {
		return nil, err
	}

	// Convert to specified location
	parsed.Time = parsed.Time.In(location)
	return parsed, nil
}

// ParseDatetimeWithFormat parses a datetime string with a specific format.
func ParseDatetimeWithFormat(input, format string) (*ParsedDateTime, error) {
	t, err := time.Parse(format, input)
	if err != nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: fmt.Sprintf("failed to parse datetime with format %s: %v", format, err),
			Cause:   err,
		}
	}

	parser := NewParser()
	return &ParsedDateTime{
		Time:      t,
		Layout:    format,
		IsUTC:     t.Location() == time.UTC,
		Precision: parser.detectPrecision(format),
		Original:  input,
	}, nil
}

// FormatDatetime formats a time.Time using the specified layout.
func FormatDatetime(t time.Time, layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return t.Format(layout)
}
