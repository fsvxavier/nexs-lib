package duration

import (
	"context"
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
		expected time.Duration
		wantErr  bool
	}{
		// Standard Go formats
		{
			name:     "standard seconds",
			input:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "standard minutes",
			input:    "15m",
			expected: 15 * time.Minute,
		},
		{
			name:     "standard hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "combined standard",
			input:    "1h30m45s",
			expected: time.Hour + 30*time.Minute + 45*time.Second,
		},

		// Extended formats
		{
			name:     "days",
			input:    "2d",
			expected: 2 * Day,
		},
		{
			name:     "weeks",
			input:    "1w",
			expected: Week,
		},
		{
			name:     "combined with days",
			input:    "1d12h30m",
			expected: Day + 12*time.Hour + 30*time.Minute,
		},
		{
			name:     "weeks and days",
			input:    "1w2d",
			expected: Week + 2*Day,
		},

		// Decimal numbers
		{
			name:     "decimal hours",
			input:    "2.5h",
			expected: time.Duration(2.5 * float64(time.Hour)),
		},
		{
			name:     "decimal days",
			input:    "1.5d",
			expected: time.Duration(1.5 * float64(Day)),
		},

		// Zero duration
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "zero seconds",
			input:    "0s",
			expected: 0,
		},

		// Negative durations
		{
			name:     "negative seconds",
			input:    "-30s",
			expected: -30 * time.Second,
		},
		{
			name:     "negative hours",
			input:    "-2h",
			expected: -2 * time.Hour,
		},

		// Error cases
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "not-a-duration",
			wantErr: true,
		},
		{
			name:    "unknown unit",
			input:   "30x",
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
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseExtended(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "days",
			input:    "3d",
			expected: 3 * Day,
		},
		{
			name:     "weeks",
			input:    "2w",
			expected: 2 * Week,
		},
		{
			name:     "months",
			input:    "1mo",
			expected: Month,
		},
		{
			name:     "years",
			input:    "1y",
			expected: Year,
		},
		{
			name:     "complex combination",
			input:    "1w2d3h4m5s",
			expected: Week + 2*Day + 3*time.Hour + 4*time.Minute + 5*time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseExtended(ctx, tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseVerboseFormat(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "hours and minutes",
			input:    "2 hours 30 minutes",
			expected: 2*time.Hour + 30*time.Minute,
		},
		{
			name:     "days and hours",
			input:    "1 day 12 hours",
			expected: Day + 12*time.Hour,
		},
		{
			name:     "single unit",
			input:    "45 seconds",
			expected: 45 * time.Second,
		},
		{
			name:     "decimal number",
			input:    "2.5 hours",
			expected: time.Duration(2.5 * float64(time.Hour)),
		},
		{
			name:     "weeks and days",
			input:    "1 week 3 days",
			expected: Week + 3*Day,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseRelativeFormat(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "half an hour",
			input:    "half an hour",
			expected: 30 * time.Minute,
		},
		{
			name:     "quarter hour",
			input:    "quarter hour",
			expected: 15 * time.Minute,
		},
		{
			name:     "a minute",
			input:    "a minute",
			expected: time.Minute,
		},
		{
			name:     "an hour",
			input:    "an hour",
			expected: time.Hour,
		},
		{
			name:     "a day",
			input:    "a day",
			expected: Day,
		},
		{
			name:     "half a day",
			input:    "half a day",
			expected: 12 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseWithOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("with custom units", func(t *testing.T) {
		customUnits := map[string]time.Duration{
			"fortnight": 14 * Day,
			"jiffy":     10 * time.Millisecond,
		}

		parser := NewParser(parsers.WithCustomUnits(customUnits))

		result, err := parser.Parse(ctx, "2fortnight")
		require.NoError(t, err)
		assert.Equal(t, 28*Day, result)

		result, err = parser.Parse(ctx, "5jiffy")
		require.NoError(t, err)
		assert.Equal(t, 50*time.Millisecond, result)
	})

	t.Run("with ignore case", func(t *testing.T) {
		parser := NewParser(parsers.WithIgnoreCase(true))

		result, err := parser.Parse(ctx, "2 HOURS 30 MINUTES")
		require.NoError(t, err)
		assert.Equal(t, 2*time.Hour+30*time.Minute, result)
	})
}

func TestParser_MustParse(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		result := parser.MustParse(ctx, "1h30m")
		assert.Equal(t, time.Hour+30*time.Minute, result)
	})

	t.Run("invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			parser.MustParse(ctx, "invalid-duration")
		})
	})
}

func TestParser_TryParse(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		result, ok := parser.TryParse(ctx, "45m")
		assert.True(t, ok)
		assert.Equal(t, 45*time.Minute, result)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, ok := parser.TryParse(ctx, "invalid-duration")
		assert.False(t, ok)
	})
}

func TestParser_GetSupportedUnits(t *testing.T) {
	parser := NewParser()

	units := parser.GetSupportedUnits()

	assert.NotEmpty(t, units)
	assert.Contains(t, units, "s")
	assert.Contains(t, units, "m")
	assert.Contains(t, units, "h")
	assert.Contains(t, units, "d")
	assert.Contains(t, units, "w")
	assert.Equal(t, time.Second, units["s"])
	assert.Equal(t, Day, units["d"])
	assert.Equal(t, Week, units["w"])
}

func TestParser_ErrorHandling(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	t.Run("parse error details", func(t *testing.T) {
		_, err := parser.Parse(ctx, "invalid-duration-format")

		require.Error(t, err)

		var parseErr *parsers.ParseError
		assert.ErrorAs(t, err, &parseErr)
		assert.Equal(t, parsers.ErrorTypeInvalidFormat, parseErr.Type)
		assert.Equal(t, "invalid-duration-format", parseErr.Input)
		assert.NotEmpty(t, parseErr.Suggestions)
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := parser.Parse(cancelledCtx, "1h")

		require.Error(t, err)
		var parseErr *parsers.ParseError
		assert.ErrorAs(t, err, &parseErr)
		assert.Equal(t, parsers.ErrorTypeTimeout, parseErr.Type)
	})
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 24*time.Hour, Day)
	assert.Equal(t, 7*Day, Week)
	assert.Equal(t, 30*Day, Month)
	assert.Equal(t, 365*Day, Year)
}

// Package-level function tests

func TestParse(t *testing.T) {
	result, err := Parse("1h30m")

	require.NoError(t, err)
	assert.Equal(t, time.Hour+30*time.Minute, result)
}

func TestMustParse(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		result := MustParse("45m")
		assert.Equal(t, 45*time.Minute, result)
	})

	t.Run("invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse("invalid-duration")
		})
	})
}

func TestParseExtended(t *testing.T) {
	result, err := ParseExtended("1w2d3h")

	require.NoError(t, err)
	assert.Equal(t, Week+2*Day+3*time.Hour, result)
}

func TestGetSupportedUnits(t *testing.T) {
	units := GetSupportedUnits()

	assert.NotEmpty(t, units)
	assert.Contains(t, units, "d")
	assert.Contains(t, units, "w")
}

// Benchmark tests

func BenchmarkParser_Parse(b *testing.B) {
	parser := NewParser()
	ctx := context.Background()
	input := "1h30m45s"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(ctx, input)
	}
}

func BenchmarkParser_ParseExtended(b *testing.B) {
	parser := NewParser()
	ctx := context.Background()
	input := "1w2d3h4m5s"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseExtended(ctx, input)
	}
}

func BenchmarkParser_ParseVerbose(b *testing.B) {
	parser := NewParser()
	ctx := context.Background()
	input := "2 hours 30 minutes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(ctx, input)
	}
}

func BenchmarkStandardParse(b *testing.B) {
	input := "1h30m45s"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = time.ParseDuration(input)
	}
}
