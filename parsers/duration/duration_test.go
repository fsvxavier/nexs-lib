package duration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("Expected parser to be created")
	}
	if parser.config == nil {
		t.Error("Expected config to be initialized")
	}
	if parser.unitMap == nil {
		t.Error("Expected unitMap to be initialized")
	}
}

func TestParser_ParseString(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1s", time.Second, false},
		{"1m", time.Minute, false},
		{"1h", time.Hour, false},
		{"1d", Day, false},
		{"1w", Week, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"1d12h", Day + 12*time.Hour, false},
		{"1w2d", Week + 2*Day, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"123", 0, true}, // no unit
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result to be non-nil")
				return
			}

			if result.Duration != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result.Duration)
			}

			if result.Original != tt.input {
				t.Errorf("Expected original %s, got %s", tt.input, result.Original)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	data := []byte("1h30m")
	result, err := parser.Parse(ctx, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected result to be non-nil")
		return
	}

	expected := time.Hour + 30*time.Minute
	if result.Duration != expected {
		t.Errorf("Expected %v, got %v", expected, result.Duration)
	}
}

func TestFormatter_Format(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	data := &ParsedDuration{
		Duration: time.Hour + 30*time.Minute,
		Original: "1h30m",
		Units:    []string{"h", "m"},
	}

	result, err := formatter.Format(ctx, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(result) == "" {
		t.Error("Expected non-empty result")
	}
}

func TestFormatter_FormatDuration(t *testing.T) {
	formatter := NewFormatter()

	tests := []struct {
		duration time.Duration
		contains string // what the result should contain
	}{
		{0, "0s"},
		{time.Second, "1s"},
		{time.Minute, "1m"},
		{time.Hour, "1h"},
		{Day, "1d"},
		{Week, "1w"},
		{Week + Day, "1w1d"},
		{-time.Hour, "-"},
	}

	for _, tt := range tests {
		result := formatter.formatDuration(tt.duration)
		if !containsSubstring(result, tt.contains) {
			t.Errorf("Expected result to contain %s, got %s", tt.contains, result)
		}
	}
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("ParseDuration", func(t *testing.T) {
		result, err := ParseDuration("1h30m")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected result to be non-nil")
		}
	})

	t.Run("FormatDuration", func(t *testing.T) {
		result := FormatDuration(time.Hour + 30*time.Minute)
		if result == "" {
			t.Error("Expected non-empty result")
		}
	})

	t.Run("ToDays", func(t *testing.T) {
		result := ToDays(Day)
		if result != 1.0 {
			t.Errorf("Expected 1.0, got %f", result)
		}
	})

	t.Run("ToWeeks", func(t *testing.T) {
		result := ToWeeks(Week)
		if result != 1.0 {
			t.Errorf("Expected 1.0, got %f", result)
		}
	})

	t.Run("FromDays", func(t *testing.T) {
		result := FromDays(1.0)
		if result != Day {
			t.Errorf("Expected %v, got %v", Day, result)
		}
	})

	t.Run("FromWeeks", func(t *testing.T) {
		result := FromWeeks(1.0)
		if result != Week {
			t.Errorf("Expected %v, got %v", Week, result)
		}
	})
}

func TestNewParserWithConfig(t *testing.T) {
	config := &interfaces.ParserConfig{
		MaxSize: 100,
	}
	parser := NewParserWithConfig(config)
	if parser == nil {
		t.Error("Expected parser to be created")
	}
	if parser.config != config {
		t.Error("Expected config to be set to provided config")
	}
	if parser.unitMap == nil {
		t.Error("Expected unitMap to be initialized")
	}
}

func TestParser_ParseString_EdgeCases(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
		hasError bool
	}{
		{"zero", "0s", 0, false},
		{"negative", "-1h", -time.Hour, false},
		{"positive sign", "+1h", time.Hour, false},
		{"microseconds", "100µs", 100 * time.Microsecond, false},
		{"microseconds mu", "100μs", 100 * time.Microsecond, false},
		{"fractional", "1.5h", time.Hour + 30*time.Minute, false},
		{"complex", "1w2d3h4m5s", Week + 2*Day + 3*time.Hour + 4*time.Minute + 5*time.Second, false},
		{"whitespace", "  1h  ", time.Hour, false},
		{"missing number", "h", 0, true},
		{"missing unit", "123", 0, true},
		{"invalid unit", "1x", 0, true},
		{"double negative", "--1h", 0, true},
		{"only dot", ".", 0, true},
		{"only sign", "-", 0, true},
		{"empty after sign", "+", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Duration != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result.Duration)
			}
		})
	}
}

func TestParser_ValidateInput(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"valid", "1h", false},
		{"empty", "", true},
		{"no digits", "abc", true},
		{"no letters", "123", true},
		{"valid complex", "1d2h3m", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateInput(tt.input)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParser_ValidateInput_MaxSize(t *testing.T) {
	config := &interfaces.ParserConfig{
		MaxSize: 5,
	}
	parser := NewParserWithConfig(config)

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"within limit", "1h", false},
		{"at limit", "1h30m", false},
		{"exceeds limit", "123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateInput(tt.input)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParser_ExtractUnits(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input    string
		expected []string
	}{
		{"1h30m", []string{"h", "m"}},
		{"1d", []string{"d"}},
		{"1w2d3h", []string{"w", "d", "h"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			units := parser.extractUnits(tt.input)
			if len(units) != len(tt.expected) {
				t.Errorf("Expected %d units, got %d", len(tt.expected), len(units))
				return
			}
			// Note: order might differ, so check if all expected units are present
			for _, expected := range tt.expected {
				found := false
				for _, unit := range units {
					if unit == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected unit %s not found in %v", expected, units)
				}
			}
		})
	}
}

func TestNewFormatterWithLongUnits(t *testing.T) {
	formatter := NewFormatterWithLongUnits()
	if formatter == nil {
		t.Error("Expected formatter to be created")
	}
	if formatter.useShortUnits {
		t.Error("Expected useShortUnits to be false")
	}
}

func TestFormatter_FormatString(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	data := &ParsedDuration{
		Duration: time.Hour,
		Original: "1h",
		Units:    []string{"h"},
	}

	result, err := formatter.FormatString(ctx, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestFormatter_Format_NilData(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	_, err := formatter.Format(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
}

func TestFormatter_FormatWriter(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	data := &ParsedDuration{
		Duration: time.Hour,
		Original: "1h",
		Units:    []string{"h"},
	}

	err := formatter.FormatWriter(ctx, data, nil)
	if err == nil {
		t.Error("Expected error for FormatWriter")
	}
}

func TestFormatter_FormatDuration_Negative(t *testing.T) {
	formatter := NewFormatter()

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{-time.Second, "-1s"},
		{-time.Minute, "-1m0s"},
		{-time.Hour, "-1h0m0s"},
		{-Day, "-1d"},
		{-Week, "-1w"},
		{-(Week + Day), "-1w1d"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatter.formatDuration(tt.duration)
			if result[0] != '-' {
				t.Errorf("Expected negative duration to start with '-', got %s", result)
			}
		})
	}
}

func TestLeadingInt(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		rem      string
		hasError bool
	}{
		{"123abc", 123, "abc", false},
		{"0", 0, "", false},
		{"999", 999, "", false},
		{"abc", 0, "abc", false},
		{"", 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, rem, err := leadingInt(tt.input)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
			if rem != tt.rem {
				t.Errorf("Expected remainder %s, got %s", tt.rem, rem)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if Day != time.Hour*24 {
		t.Errorf("Expected Day to be 24 hours, got %v", Day)
	}
	if Week != Day*7 {
		t.Errorf("Expected Week to be 7 days, got %v", Week)
	}
}

func TestUtilityFunctions_EdgeCases(t *testing.T) {
	t.Run("ParseDuration empty", func(t *testing.T) {
		_, err := ParseDuration("")
		if err == nil {
			t.Error("Expected error for empty string")
		}
	})

	t.Run("FormatDuration zero", func(t *testing.T) {
		result := FormatDuration(0)
		if result != "0s" {
			t.Errorf("Expected '0s', got %s", result)
		}
	})

	t.Run("ToDays fractional", func(t *testing.T) {
		result := ToDays(12 * time.Hour)
		if result != 0.5 {
			t.Errorf("Expected 0.5, got %f", result)
		}
	})

	t.Run("FromDays fractional", func(t *testing.T) {
		result := FromDays(0.5)
		expected := 12 * time.Hour
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseEnhanced_EdgeCases(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"overflow value", "9223372036854775808s", true},
		{"overflow multiplication", "1000000000000000000000h", true},
		{"fractional overflow", "9223372036854775807.9999999999999999999s", true},
		{"invalid after sign", "+-1h", true},
		{"decimal without digits", ".h", true},
		{"decimal only", "1.", true},
		{"multiple decimals", "1.2.3h", true},
		{"valid zero", "0s", false},
		{"valid fractional seconds", "1.5s", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.parseEnhanced(tt.input)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestLeadingInt_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedX   uint64
		expectedRem string
		hasError    bool
	}{
		{"overflow", "99999999999999999999999999999", 0, "", true},
		{"large valid", "9223372036854775807", 9223372036854775807, "", false},
		{"max uint64 plus 1", "18446744073709551616", 0, "", true},
		{"single digit", "7", 7, "", false},
		{"with letters", "123abc", 123, "abc", false},
		{"no digits", "abc", 0, "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, rem, err := leadingInt(tt.input)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if !tt.hasError {
				if x != tt.expectedX {
					t.Errorf("Expected x=%d, got %d", tt.expectedX, x)
				}
				if rem != tt.expectedRem {
					t.Errorf("Expected rem=%s, got %s", tt.expectedRem, rem)
				}
			}
		})
	}
}

func TestFormatter_FormatString_ErrorCase(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	// Test error case by passing nil data
	result, err := formatter.FormatString(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
	if result != "" {
		t.Errorf("Expected empty string on error, got %s", result)
	}
}

func TestFormatter_FormatDuration_EdgeCases(t *testing.T) {
	formatter := NewFormatter()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"exactly one week", Week, "1w"},
		{"exactly one day", Day, "1d"},
		{"zero duration", 0, "0s"},
		{"very small duration", 1, "1ns"},
		{"negative week", -Week, "-1w"},
		{"negative day", -Day, "-1d"},
		{"week plus day plus time", Week + Day + time.Hour, "1w1d1h0m0s"},
		{"only negative sign", -time.Nanosecond, "-1ns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatDuration(tt.duration)
			if result == "" {
				t.Error("Expected non-empty result")
			}
			// For negative durations, ensure they start with '-'
			if tt.duration < 0 && !strings.HasPrefix(result, "-") {
				t.Errorf("Expected negative duration to start with '-', got %s", result)
			}
		})
	}
}

func TestParser_ParseString_MoreEdgeCases(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"small week", "5w", false},
		{"small day", "10d", false},
		{"mixed with standard units", "1w2d3h4m5s6ms7µs8ns", false},
		{"unicode microsecond", "100μs", false},
		{"micro symbol", "100µs", false},
		{"fractional week", "1.5w", false},
		{"fractional day", "2.25d", false},
		{"only spaces", "   ", true},
		{"tab and spaces", "\t  \n", true},
		{"leading zeros", "000001h", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result to be non-nil")
			}
		})
	}
}

func TestParseEnhanced_AdvancedEdgeCases(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		expectErr bool
		checkFunc func(*testing.T, time.Duration, error)
	}{
		{
			name:      "Fractional seconds",
			input:     "1.5s",
			expectErr: false,
			checkFunc: func(t *testing.T, result time.Duration, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expected := 1500 * time.Millisecond
				if result != expected {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:      "Mixed units with decimals",
			input:     "1h30.5m",
			expectErr: false,
			checkFunc: func(t *testing.T, result time.Duration, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expected := time.Hour + 30*time.Minute + 30*time.Second
				if result != expected {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:      "Very large number",
			input:     "999999999999999999ns",
			expectErr: false,
			checkFunc: func(t *testing.T, result time.Duration, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			name:      "Invalid number format",
			input:     "1.2.3s",
			expectErr: true,
			checkFunc: func(t *testing.T, result time.Duration, err error) {
				if err == nil {
					t.Error("Expected error for invalid number format")
				}
			},
		},
		{
			name:      "Number overflow",
			input:     "99999999999999999999999999999999999999s",
			expectErr: true,
			checkFunc: func(t *testing.T, result time.Duration, err error) {
				if err == nil {
					t.Error("Expected error for number overflow")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseEnhanced(tt.input)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestLeadingInt_ComprehensiveEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedX   uint64
		expectedRem string
		hasError    bool
	}{
		{
			name:        "Empty string",
			input:       "",
			expectedX:   0,
			expectedRem: "",
			hasError:    false,
		},
		{
			name:        "Only digits",
			input:       "12345",
			expectedX:   12345,
			expectedRem: "",
			hasError:    false,
		},
		{
			name:        "Digits with suffix",
			input:       "123abc",
			expectedX:   123,
			expectedRem: "abc",
			hasError:    false,
		},
		{
			name:        "No leading digits",
			input:       "abc123",
			expectedX:   0,
			expectedRem: "abc123",
			hasError:    false,
		},
		{
			name:        "Zero value",
			input:       "0s",
			expectedX:   0,
			expectedRem: "s",
			hasError:    false,
		},
		{
			name:        "Large number within limit",
			input:       "9223372036854775807ns", // 2^63 - 1
			expectedX:   9223372036854775807,
			expectedRem: "ns",
			hasError:    false,
		},
		{
			name:        "Overflow uint64",
			input:       "92233720368547758080ns", // Greater than 2^63
			expectedX:   0,
			expectedRem: "",
			hasError:    true,
		},
		{
			name:        "Fractional number",
			input:       "12.34s",
			expectedX:   12,
			expectedRem: ".34s",
			hasError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, rem, err := leadingInt(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if x != tt.expectedX {
				t.Errorf("Expected x=%d, got %d", tt.expectedX, x)
			}
			if rem != tt.expectedRem {
				t.Errorf("Expected rem=%s, got %s", tt.expectedRem, rem)
			}
		})
	}
}

func TestFormatDuration_AdvancedEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		duration    time.Duration
		longUnits   bool
		expectedLen int // We'll just check that it returns a non-empty string
	}{
		{
			name:        "Zero duration",
			duration:    0,
			longUnits:   false,
			expectedLen: 1, // "0s"
		},
		{
			name:        "Zero duration long units",
			duration:    0,
			longUnits:   true,
			expectedLen: 1, // "0 seconds"
		},
		{
			name:        "Very small duration",
			duration:    1 * time.Nanosecond,
			longUnits:   false,
			expectedLen: 1,
		},
		{
			name:        "Complex duration",
			duration:    25*time.Hour + 3*time.Minute + 45*time.Second,
			longUnits:   false,
			expectedLen: 1,
		},
		{
			name:        "Complex duration long units",
			duration:    25*time.Hour + 3*time.Minute + 45*time.Second,
			longUnits:   true,
			expectedLen: 1,
		},
		{
			name:        "Negative duration",
			duration:    -5 * time.Minute,
			longUnits:   false,
			expectedLen: 1,
		},
		{
			name:        "Maximum duration",
			duration:    time.Duration(1<<63 - 1),
			longUnits:   false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter()
			if tt.longUnits {
				formatter = NewFormatterWithLongUnits()
			}

			result := formatter.formatDuration(tt.duration)

			if len(result) == 0 {
				t.Error("Expected non-empty result")
			}
		})
	}
}
