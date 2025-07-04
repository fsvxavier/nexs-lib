package datetime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "RFC3339",
			input:    "2023-01-15T10:30:45Z",
			expected: "2023-01-15 10:30:45 +0000 UTC",
		},
		{
			name:     "ISO date",
			input:    "2023-01-15",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:     "US format",
			input:    "01/15/2023",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:     "Text format",
			input:    "January 15, 2023",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:     "Text format short",
			input:    "Jan 15, 2023",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestParser_ParseWithOptions(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("with custom location", func(t *testing.T) {
		loc, _ := time.LoadLocation("America/New_York")
		result, err := parser.ParseWithOptions(ctx, "2023-01-15 10:30:45",
			parsers.WithLocation(loc))

		require.NoError(t, err)
		assert.Equal(t, loc, result.Location())
	})

	t.Run("with custom formats", func(t *testing.T) {
		result, err := parser.ParseWithOptions(ctx, "15/01/2023 14:30",
			parsers.WithCustomFormats("02/01/2006 15:04"))

		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
	})

	t.Run("with strict mode", func(t *testing.T) {
		_, err := parser.ParseWithOptions(ctx, "sometime in january",
			parsers.WithStrictMode(true))

		assert.Error(t, err)
	})
}

func TestParser_MustParse(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		result := parser.MustParse(ctx, "2023-01-15T10:30:45Z")
		assert.Equal(t, 2023, result.Year())
	})

	t.Run("invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			parser.MustParse(ctx, "invalid-date")
		})
	})
}

func TestParser_TryParse(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		result, ok := parser.TryParse(ctx, "2023-01-15")
		assert.True(t, ok)
		assert.Equal(t, 2023, result.Year())
	})

	t.Run("invalid input", func(t *testing.T) {
		_, ok := parser.TryParse(ctx, "invalid-date")
		assert.False(t, ok)
	})
}

func TestParser_ParseInLocation(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	loc, _ := time.LoadLocation("America/New_York")
	result, err := parser.ParseInLocation(ctx, "2023-01-15 10:30:45", loc)

	require.NoError(t, err)
	assert.Equal(t, loc, result.Location())
}

func TestParser_ParseToUTC(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	result, err := parser.ParseToUTC(ctx, "2023-01-15T10:30:45-05:00")

	require.NoError(t, err)
	assert.Equal(t, time.UTC, result.Location())
	assert.Equal(t, 15, result.Hour()) // Should be converted to UTC
}

func TestParser_FlexibleParsing(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "relative - today",
			input:    "today",
			expected: time.Now().Format("2006-01-02"),
		},
		{
			name:     "relative - yesterday",
			input:    "yesterday",
			expected: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		},
		{
			name:     "relative - tomorrow",
			input:    "tomorrow",
			expected: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		},
		{
			name:     "text with comma",
			input:    "March 15, 2023",
			expected: "2023-03-15",
		},
		{
			name:     "US format",
			input:    "03/15/2023",
			expected: "2023-03-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)

			require.NoError(t, err)

			if strings.Contains(tt.name, "relative") {
				// For relative dates, just check that we got a valid result
				assert.False(t, result.IsZero())
			} else {
				assert.Equal(t, tt.expected, result.Format("2006-01-02"))
			}
		})
	}
}

func TestParser_DateOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		order    parsers.DateOrder
		expected string
	}{
		{
			name:     "MDY order",
			input:    "03/15/2023",
			order:    parsers.DateOrderMDY,
			expected: "2023-03-15",
		},
		{
			name:     "DMY order",
			input:    "15/03/2023",
			order:    parsers.DateOrderDMY,
			expected: "2023-03-15",
		},
		{
			name:     "YMD order",
			input:    "2023/03/15",
			order:    parsers.DateOrderYMD,
			expected: "2023-03-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(parsers.WithDateOrder(tt.order))
			result, err := parser.Parse(ctx, tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02"))
		})
	}
}

func TestParser_ErrorHandling(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("parse error details", func(t *testing.T) {
		_, err := parser.Parse(ctx, "invalid-date-format")

		require.Error(t, err)

		var parseErr *parsers.ParseError
		assert.ErrorAs(t, err, &parseErr)
		assert.Equal(t, parsers.ErrorTypeInvalidFormat, parseErr.Type)
		assert.Equal(t, "invalid-date-format", parseErr.Input)
		assert.NotEmpty(t, parseErr.Suggestions)
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := parser.Parse(cancelledCtx, "2023-01-15")

		require.Error(t, err)
		var parseErr *parsers.ParseError
		assert.ErrorAs(t, err, &parseErr)
		assert.Equal(t, parsers.ErrorTypeTimeout, parseErr.Type)
	})
}

func TestParser_Caching(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	// Parse the same input multiple times
	input := "2023-01-15T10:30:45Z"

	result1, err1 := parser.Parse(ctx, input)
	require.NoError(t, err1)

	result2, err2 := parser.Parse(ctx, input)
	require.NoError(t, err2)

	assert.Equal(t, result1, result2)

	// Verify the format was cached
	assert.Contains(t, parser.formatCache, input)
}

func TestParser_GetSupportedFormats(t *testing.T) {
	parser := NewParser()

	formats := parser.GetSupportedFormats()

	assert.NotEmpty(t, formats)
	assert.Contains(t, formats, time.RFC3339)
	assert.Contains(t, formats, "2006-01-02")
}

func TestParser_ParseFormat(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "RFC3339",
			input: "2023-01-15T10:30:45Z",
		},
		{
			name:  "ISO date",
			input: "2023-01-15",
		},
		{
			name:  "US format MM/DD/YYYY",
			input: "01/15/2023",
		},
		{
			name:  "European format DD.MM.YYYY",
			input: "15.01.2023",
		},
		{
			name:  "Text format long",
			input: "January 15, 2023",
		},
		{
			name:  "Text format short",
			input: "Jan 15, 2023",
		},
		{
			name:  "SQL datetime",
			input: "2023-01-15 10:30:45",
		},
		{
			name:  "Time only",
			input: "15:30:45",
		},
		{
			name:    "Invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := parser.ParseFormat(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, format, "should return a valid format")

				// Verify the detected format actually works
				_, parseErr := time.Parse(format, tt.input)
				assert.NoError(t, parseErr, "detected format should be able to parse the input")
			}
		})
	}
}

func TestParser_DetectFormat(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	// Test that DetectFormat is an alias for ParseFormat
	input := "2023-01-15T10:30:45Z"

	format1, err1 := parser.ParseFormat(ctx, input)
	format2, err2 := parser.DetectFormat(ctx, input)

	assert.Equal(t, err1, err2)
	assert.Equal(t, format1, format2)
}

func TestParser_ParseWithFormat(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		format   string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid format",
			input:    "15/01/2023",
			format:   "02/01/2006",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:     "Custom format",
			input:    "2023年01月15日",
			format:   "2006年01月02日",
			expected: "2023-01-15 00:00:00 +0000 UTC",
		},
		{
			name:    "Wrong format",
			input:   "15/01/2023",
			format:  "2006-01-02",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "",
			format:  "2006-01-02",
			wantErr: true,
		},
		{
			name:    "Empty format",
			input:   "2023-01-15",
			format:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseWithFormat(ctx, tt.input, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestParser_FormatCaching(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()
	input := "2023-01-15T10:30:45Z"

	// First call should detect and cache the format
	format1, err1 := parser.ParseFormat(ctx, input)
	require.NoError(t, err1)

	// Second call should use cached format
	format2, err2 := parser.ParseFormat(ctx, input)
	require.NoError(t, err2)

	assert.Equal(t, format1, format2)

	// Verify the format is actually cached
	cachedFormat, exists := parser.formatCache[input]
	assert.True(t, exists)
	assert.Equal(t, format1, cachedFormat)
}

// Package-level function tests

func TestParse(t *testing.T) {
	result, err := Parse("2023-01-15T10:30:45Z")

	require.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.January, result.Month())
	assert.Equal(t, 15, result.Day())
}

// Test ParseAny function for compatibility with old library
func TestParseAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected format: RFC3339 for comparison
		wantErr  bool
	}{
		{
			name:     "ISO8601 date",
			input:    "2023-01-15T10:30:45Z",
			expected: "2023-01-15T10:30:45Z",
			wantErr:  false,
		},
		{
			name:     "US date format",
			input:    "01/15/2023",
			expected: "2023-01-15T00:00:00Z",
			wantErr:  false,
		},
		{
			name:     "European date format",
			input:    "15/01/2023",
			expected: "2023-01-15T00:00:00Z",
			wantErr:  false,
		},
		{
			name:     "Date with time",
			input:    "January 15, 2023 10:30 AM",
			expected: "2023-01-15T10:30:00Z",
			wantErr:  false,
		},
		{
			name:     "Unix timestamp",
			input:    "1673778645",
			expected: "2023-01-15T10:30:45Z",
			wantErr:  false,
		},
		{
			name:    "Invalid date",
			input:   "not a date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAny(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAny() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseAny() unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Convert expected to time for comparison
			expected, err := time.Parse(time.RFC3339, tt.expected)
			if err != nil {
				t.Fatalf("Failed to parse expected time %q: %v", tt.expected, err)
			}

			// Compare times (allowing for some flexibility in precision)
			if !result.Equal(expected) && result.Unix() != expected.Unix() {
				t.Errorf("ParseAny() = %v, expected %v", result, expected)
			}
		})
	}
}

// Test ParseAny with options for compatibility
func TestParseAnyWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     []ParserOption
		expected string
		wantErr  bool
	}{
		{
			name:     "Prefer month first - US format",
			input:    "02/03/2023", // Ambiguous: could be Feb 3 or Mar 2
			opts:     []ParserOption{PreferMonthFirst(true)},
			expected: "2023-02-03T00:00:00Z", // Should interpret as Feb 3
			wantErr:  false,
		},
		{
			name:     "Prefer day first - European format",
			input:    "02/03/2023", // Ambiguous: could be Feb 3 or Mar 2
			opts:     []ParserOption{PreferMonthFirst(false)},
			expected: "2023-03-02T00:00:00Z", // Should interpret as Mar 2
			wantErr:  false,
		},
		{
			name:     "With retry on ambiguous dates",
			input:    "13/02/2023", // Day > 12, so should work with retry
			opts:     []ParserOption{PreferMonthFirst(true), RetryAmbiguousDateWithSwap(true)},
			expected: "2023-02-13T00:00:00Z",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAny(tt.input, tt.opts...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAny() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseAny() unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Convert expected to time for comparison
			expected, err := time.Parse(time.RFC3339, tt.expected)
			if err != nil {
				t.Fatalf("Failed to parse expected time %q: %v", tt.expected, err)
			}

			// Compare dates (ignore time for date-only inputs)
			if !result.Equal(expected) && result.Format("2006-01-02") != expected.Format("2006-01-02") {
				t.Errorf("ParseAny() = %v, expected %v", result, expected)
			}
		})
	}
}

// Test ParseIn function
func TestParseIn(t *testing.T) {
	// Create a test timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Cannot load America/New_York timezone")
	}

	tests := []struct {
		name     string
		input    string
		location *time.Location
		wantErr  bool
	}{
		{
			name:     "Parse in specific timezone",
			input:    "2023-01-15 10:30:00",
			location: loc,
			wantErr:  false,
		},
		{
			name:     "Parse with UTC",
			input:    "2023-01-15T10:30:45Z",
			location: time.UTC,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIn(tt.input, tt.location)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseIn() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseIn() unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Verify that the location is correctly applied
			if result.Location() != tt.location {
				t.Errorf("ParseIn() location = %v, expected %v", result.Location(), tt.location)
			}
		})
	}
}

// Test ParseLocal function
func TestParseLocal(t *testing.T) {
	// Store original timezone and restore after test
	original := time.Local
	defer func() { time.Local = original }()

	// Set a test timezone
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Skip("Cannot load America/Los_Angeles timezone")
	}
	time.Local = loc

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Parse with local timezone",
			input:   "2023-01-15 10:30:00",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLocal(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLocal() expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseLocal() unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Verify that the local timezone is used
			if result.Location() != time.Local {
				t.Errorf("ParseLocal() location = %v, expected %v", result.Location(), time.Local)
			}
		})
	}
}

// Test MustParseAny function
func TestMustParseAny(t *testing.T) {
	// Test successful parse
	result := MustParseAny("2023-01-15T10:30:45Z")
	expected, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:45Z")

	if !result.Equal(expected) {
		t.Errorf("MustParseAny() = %v, expected %v", result, expected)
	}

	// Test panic on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParseAny() expected panic for invalid input")
		}
	}()

	MustParseAny("invalid date")
}
