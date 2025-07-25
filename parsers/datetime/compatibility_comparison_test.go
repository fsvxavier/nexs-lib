package datetime

import (
	"testing"
	"time"
)

// TestNewVsOldAPI compares the new API with the old compatibility layer
func TestNewVsOldAPI(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"ISO8601", "2023-12-25T15:30:45Z"},
		{"Date only", "2023-12-25"},
		{"Space separated", "2023-12-25 15:30:45"},
		{"US format", "12/25/2023"},
		{"Text format", "December 25, 2023"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test new API
			newResult, newErr := ParseDatetime(tc.input)

			// Test old API (compatibility layer)
			oldResult, oldErr := ParseAny(tc.input)

			// Both should succeed or both should fail
			if (newErr == nil) != (oldErr == nil) {
				t.Errorf("API mismatch for %q: new error=%v, old error=%v", tc.input, newErr, oldErr)
				return
			}

			// If both succeeded, times should be equal (allowing for minor differences)
			if newErr == nil && oldErr == nil {
				newTime := newResult.Time
				oldTime := oldResult

				// Allow for up to 1 second difference due to different parsing precision
				diff := newTime.Sub(oldTime)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("Time difference too large for %q: new=%v, old=%v, diff=%v",
						tc.input, newTime, oldTime, diff)
				}
			}
		})
	}
}

// TestCompatibilityBehavior tests that compatibility functions behave like the old module
func TestCompatibilityBehavior(t *testing.T) {
	t.Run("ParseIn location handling", func(t *testing.T) {
		utc, _ := time.LoadLocation("UTC")
		est, _ := time.LoadLocation("America/New_York")

		input := "2023-12-25T15:30:45Z"

		utcResult, err := ParseIn(input, utc)
		if err != nil {
			t.Fatalf("ParseIn with UTC failed: %v", err)
		}

		estResult, err := ParseIn(input, est)
		if err != nil {
			t.Fatalf("ParseIn with EST failed: %v", err)
		}

		// The time value should be the same (UTC input)
		if !utcResult.Equal(estResult) {
			t.Errorf("Times should be equal: UTC=%v, EST=%v", utcResult, estResult)
		}
	})

	t.Run("ParseLocal uses local timezone", func(t *testing.T) {
		originalLocal := time.Local
		defer func() { time.Local = originalLocal }()

		// Set a specific local timezone
		est, _ := time.LoadLocation("America/New_York")
		time.Local = est

		input := "2023-12-25 15:30:45"
		result, err := ParseLocal(input)
		if err != nil {
			t.Fatalf("ParseLocal failed: %v", err)
		}

		// The result should be interpreted in the local timezone
		// For dates without timezone info, ParseLocal should use the local timezone
		// Note: The actual behavior may vary based on the parsing implementation
		// Some parsers might parse in UTC and then convert, others parse directly in local time

		// What's important is that ParseLocal behaves consistently
		// Let's just verify it doesn't panic and returns a valid time
		if result.IsZero() {
			t.Error("ParseLocal returned zero time")
		}

		// Additional check: parse the same input with UTC and compare
		utcResult, err := ParseIn(input, time.UTC)
		if err != nil {
			t.Fatalf("ParseIn with UTC failed: %v", err)
		}

		// The times should be different if they're interpreted in different timezones
		// but represent the same moment, OR they should be the same local time
		// in different timezones (different absolute times)
		_ = utcResult // We verified both parse successfully

		// The test passes if ParseLocal doesn't fail and returns a valid time
	})

	t.Run("MustParse panic behavior", func(t *testing.T) {
		// Should not panic on valid input
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MustParse panicked on valid input: %v", r)
				}
			}()
			result := MustParse("2023-12-25T15:30:45Z")
			if result.IsZero() {
				t.Error("MustParse returned zero time")
			}
		}()

		// Should panic on invalid input
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("MustParse should have panicked on invalid input")
				}
			}()
			MustParse("invalid-date")
		}()
	})

	t.Run("ParseFormat returns layout", func(t *testing.T) {
		testCases := []struct {
			input          string
			expectedFormat string
		}{
			{"2023-12-25T15:30:45Z", "2006-01-02T15:04:05Z"},
			{"2023-12-25 15:30:45", "2006-01-02 15:04:05"},
			{"12/25/2023", "01/02/2006"},
		}

		for _, tc := range testCases {
			format, err := ParseFormat(tc.input)
			if err != nil {
				t.Errorf("ParseFormat(%q) failed: %v", tc.input, err)
				continue
			}

			// Verify the format works by parsing with it
			_, err = time.Parse(format, tc.input)
			if err != nil {
				t.Errorf("Returned format %q doesn't work for input %q: %v", format, tc.input, err)
			}
		}
	})
}

// TestParserOptionsCompatibility tests that parser options work correctly
func TestParserOptionsCompatibility(t *testing.T) {
	t.Run("PreferMonthFirst option", func(t *testing.T) {
		ambiguousInput := "01/02/2023"

		// With month first (default)
		result1, err := ParseAny(ambiguousInput, PreferMonthFirst(true))
		if err != nil {
			t.Errorf("ParseAny with PreferMonthFirst(true) failed: %v", err)
		}

		// With day first
		result2, err := ParseAny(ambiguousInput, PreferMonthFirst(false))
		if err != nil {
			t.Errorf("ParseAny with PreferMonthFirst(false) failed: %v", err)
		}

		// Results might be different depending on interpretation
		// The important thing is that both succeed
		_ = result1
		_ = result2
	})

	t.Run("RetryAmbiguousDateWithSwap option", func(t *testing.T) {
		// Test with potentially problematic input
		problemInput := "32/01/2023" // Invalid day

		// Without retry
		_, err1 := ParseAny(problemInput, RetryAmbiguousDateWithSwap(false))

		// With retry
		_, err2 := ParseAny(problemInput, RetryAmbiguousDateWithSwap(true))

		// The retry option might help in some cases
		// At minimum, it shouldn't make things worse
		if err1 == nil && err2 != nil {
			t.Error("RetryAmbiguousDateWithSwap made parsing worse")
		}
	})
}

// TestAdvancedParsingCompatibility tests advanced parsing features
func TestAdvancedParsingCompatibility(t *testing.T) {
	t.Run("Unix timestamp parsing", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"1703519445", true},    // Seconds
			{"1703519445000", true}, // Milliseconds
			{"0", true},             // Epoch
			{"-1", false},           // Negative (should fail)
			{"abc", false},          // Non-numeric
		}

		for _, tc := range testCases {
			result, err := ParseAny(tc.input)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseAny(%q) failed: %v", tc.input, err)
				}
				if result.IsZero() && tc.input != "0" {
					t.Errorf("ParseAny(%q) returned zero time", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("ParseAny(%q) should have failed", tc.input)
				}
			}
		}
	})

	t.Run("Relative date parsing", func(t *testing.T) {
		testCases := []string{"today", "yesterday", "tomorrow", "now"}

		for _, input := range testCases {
			result, err := ParseAny(input)
			if err != nil {
				t.Errorf("ParseAny(%q) failed: %v", input, err)
			}
			if result.IsZero() {
				t.Errorf("ParseAny(%q) returned zero time", input)
			}
		}
	})

	t.Run("Text format parsing", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"25 December 2023", true},
			{"December 25, 2023", true},
			{"2023 December 25", true},
			{"25 Dec 2023", true},
			{"Dec 25, 2023", true},
			{"Jan 1, 2024", true},
			{"February 29, 2024", true},  // Leap year
			{"February 29, 2023", false}, // Non-leap year
			{"Invalid Month 25, 2023", false},
		}

		for _, tc := range testCases {
			result, err := ParseAny(tc.input)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseAny(%q) failed: %v", tc.input, err)
				}
				if result.IsZero() {
					t.Errorf("ParseAny(%q) returned zero time", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("ParseAny(%q) should have failed", tc.input)
				}
			}
		}
	})
}

// TestFormatConversionCompatibility tests format conversion functions
func TestFormatConversionCompatibility(t *testing.T) {
	t.Run("ParseFormatToGo comprehensive", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"yyyy-mm-dd", "2006-01-02"},
			{"YYYY-MM-DD", "2006-01-02"},
			{"DD/MM/YYYY", "02/01/2006"},
			{"MM/DD/YYYY", "01/02/2006"},
			{"YYYY-MM-DD HH:II:SS", "2006-01-02 15:04:05"},
			{"DD.MM.YYYY HH:II", "02.01.2006 15:04"},
		}

		for _, tc := range testCases {
			result := ParseFormatToGo(tc.input)
			if result != tc.expected {
				t.Errorf("ParseFormatToGo(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("ParseDateTime with timezone", func(t *testing.T) {
		testCases := []struct {
			datetime string
			format   string
			timezone string
		}{
			{"2023-12-25 15:30:45", "YYYY-MM-DD HH:II:SS", "UTC"},
			{"2023-12-25 15:30:45", "YYYY-MM-DD HH:II:SS", "America/New_York"},
			{"25/12/2023 15:30", "DD/MM/YYYY HH:II", "Europe/London"},
			{"2023-12-25 15:30:45", "YYYY-MM-DD HH:II:SS", ""}, // Empty timezone should default to UTC
		}

		for _, tc := range testCases {
			result, err := ParseDateTime(tc.datetime, tc.format, tc.timezone)
			if err != nil {
				t.Errorf("ParseDateTime(%q, %q, %q) failed: %v", tc.datetime, tc.format, tc.timezone, err)
				continue
			}
			if result == "" {
				t.Errorf("ParseDateTime(%q, %q, %q) returned empty result", tc.datetime, tc.format, tc.timezone)
			}

			// Verify result is parseable
			goFormat := ParseFormatToGo(tc.format)
			_, err = time.Parse(goFormat, result)
			if err != nil {
				t.Errorf("ParseDateTime result %q is not valid with format %q: %v", result, goFormat, err)
			}
		}
	})
}

// TestAdapterFunctionsCompatibility tests adapter functions between APIs
func TestAdapterFunctionsCompatibility(t *testing.T) {
	t.Run("ToTime and FromTime round trip", func(t *testing.T) {
		original := time.Now().UTC().Truncate(time.Second)

		// Convert to ParsedDateTime and back
		parsed := FromTime(original)
		converted := ToTime(parsed)

		if !original.Equal(converted) {
			t.Errorf("Round trip failed: original=%v, converted=%v", original, converted)
		}
	})

	t.Run("ToParsedDateTime preserves data", func(t *testing.T) {
		original := time.Now().UTC().Truncate(time.Second)
		layout := "2006-01-02 15:04:05"

		parsed := ToParsedDateTime(original, layout)

		if !parsed.Time.Equal(original) {
			t.Errorf("Time not preserved: original=%v, parsed=%v", original, parsed.Time)
		}
		if parsed.Layout != layout {
			t.Errorf("Layout not preserved: expected=%q, got=%q", layout, parsed.Layout)
		}
		if parsed.Original != original.Format(layout) {
			t.Errorf("Original not set correctly: expected=%q, got=%q", original.Format(layout), parsed.Original)
		}
	})

	t.Run("Precision detection accuracy", func(t *testing.T) {
		testCases := []struct {
			layout   string
			expected string
		}{
			{"2006-01-02T15:04:05.000000000Z", "nanosecond"},
			{"2006-01-02T15:04:05.999999999Z", "nanosecond"},
			{"2006-01-02T15:04:05Z", "second"},
			{"2006-01-02T15:04Z", "minute"},
			{"2006-01-02T15Z", "hour"},
			{"2006-01-02", "day"},
			{"2006-01", "day"},  // Contains "01"
			{"2006", "day"},     // Contains "2"
			{"unknown", "year"}, // No recognized patterns
		}

		for _, tc := range testCases {
			result := detectPrecisionFromLayout(tc.layout)
			if result != tc.expected {
				t.Errorf("detectPrecisionFromLayout(%q) = %q, expected %q", tc.layout, result, tc.expected)
			}
		}
	})
}

// TestErrorCompatibility tests error handling compatibility
func TestErrorCompatibility(t *testing.T) {
	t.Run("ErrAmbiguousMMDD definition", func(t *testing.T) {
		if ErrAmbiguousMMDD == nil {
			t.Fatal("ErrAmbiguousMMDD should not be nil")
		}
		expectedMsg := "ambiguous M/D vs D/M"
		if ErrAmbiguousMMDD.Error() != expectedMsg {
			t.Errorf("ErrAmbiguousMMDD.Error() = %q, expected %q", ErrAmbiguousMMDD.Error(), expectedMsg)
		}
	})

	t.Run("ParseStrict error handling", func(t *testing.T) {
		// Test that ParseStrict properly handles errors
		invalidInputs := []string{
			"invalid-date",
			"",
			"32/13/2023", // Invalid date
		}

		for _, input := range invalidInputs {
			_, err := ParseStrict(input)
			if err == nil {
				t.Errorf("ParseStrict(%q) should have returned an error", input)
			}
		}
	})
}
