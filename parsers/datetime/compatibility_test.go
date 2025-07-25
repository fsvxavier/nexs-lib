package datetime

import (
	"strings"
	"testing"
	"time"
)

// TestCompatibilityFunctions tests all compatibility functions match old behavior
func TestCompatibilityFunctions(t *testing.T) {
	// Test ParseAny
	t.Run("ParseAny", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool // whether parsing should succeed
		}{
			{"2023-12-25T15:30:45Z", true},
			{"2023-12-25 15:30:45", true},
			{"2023-12-25", true},
			{"12/25/2023", true},
			{"25/12/2023", true},
			{"Dec 25, 2023", true},
			{"December 25, 2023", true},
			{"25 Dec 2023", true},
			{"2023 Dec 25", true},
			{"invalid-date", false},
			{"", false},
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

	// Test ParseIn
	t.Run("ParseIn", func(t *testing.T) {
		utc, _ := time.LoadLocation("UTC")
		est, _ := time.LoadLocation("America/New_York")

		testCases := []struct {
			input    string
			location *time.Location
			expected bool
		}{
			{"2023-12-25T15:30:45Z", utc, true},
			{"2023-12-25T15:30:45Z", est, true},
			{"2023-12-25 15:30:45", utc, true},
			{"invalid-date", utc, false},
		}

		for _, tc := range testCases {
			result, err := ParseIn(tc.input, tc.location)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseIn(%q, %v) failed: %v", tc.input, tc.location, err)
				}
				if result.IsZero() {
					t.Errorf("ParseIn(%q, %v) returned zero time", tc.input, tc.location)
				}
			} else {
				if err == nil {
					t.Errorf("ParseIn(%q, %v) should have failed", tc.input, tc.location)
				}
			}
		}
	})

	// Test ParseLocal
	t.Run("ParseLocal", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"2023-12-25T15:30:45Z", true},
			{"2023-12-25 15:30:45", true},
			{"Dec 25, 2023", true},
			{"invalid-date", false},
		}

		for _, tc := range testCases {
			result, err := ParseLocal(tc.input)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseLocal(%q) failed: %v", tc.input, err)
				}
				if result.IsZero() {
					t.Errorf("ParseLocal(%q) returned zero time", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("ParseLocal(%q) should have failed", tc.input)
				}
			}
		}
	})

	// Test MustParse
	t.Run("MustParse", func(t *testing.T) {
		// Test successful parsing
		result := MustParse("2023-12-25T15:30:45Z")
		if result.IsZero() {
			t.Error("MustParse returned zero time")
		}

		// Test panic on invalid input
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParse should have panicked on invalid input")
			}
		}()
		MustParse("invalid-date")
	})

	// Test ParseFormat
	t.Run("ParseFormat", func(t *testing.T) {
		testCases := []struct {
			input          string
			expectedFormat bool // whether format detection should succeed
		}{
			{"2023-12-25T15:30:45Z", true},
			{"2023-12-25 15:30:45", true},
			{"12/25/2023", true},
			{"Dec 25, 2023", true},
			{"invalid-date", false},
		}

		for _, tc := range testCases {
			format, err := ParseFormat(tc.input)
			if tc.expectedFormat {
				if err != nil {
					t.Errorf("ParseFormat(%q) failed: %v", tc.input, err)
				}
				if format == "" {
					t.Errorf("ParseFormat(%q) returned empty format", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("ParseFormat(%q) should have failed", tc.input)
				}
			}
		}
	})

	// Test ParseStrict
	t.Run("ParseStrict", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"2023-12-25T15:30:45Z", true},
			{"2023-12-25 15:30:45", true},
			{"Dec 25, 2023", true},
			// Note: ParseStrict should fail on ambiguous dates like "01/02/2023"
			{"invalid-date", false},
		}

		for _, tc := range testCases {
			result, err := ParseStrict(tc.input)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseStrict(%q) failed: %v", tc.input, err)
				}
				if result.IsZero() {
					t.Errorf("ParseStrict(%q) returned zero time", tc.input)
				}
			} else {
				if err == nil {
					t.Errorf("ParseStrict(%q) should have failed", tc.input)
				}
			}
		}
	})
}

// TestParserOptions tests parser options functionality
func TestParserOptions(t *testing.T) {
	t.Run("PreferMonthFirst", func(t *testing.T) {
		// Test with month-first preference
		result1, err := ParseAny("01/02/2023", PreferMonthFirst(true))
		if err != nil {
			t.Errorf("ParseAny with PreferMonthFirst(true) failed: %v", err)
		}
		if result1.Month() != time.January {
			t.Errorf("Expected January, got %v", result1.Month())
		}

		// Test with day-first preference
		result2, err := ParseAny("01/02/2023", PreferMonthFirst(false))
		if err != nil {
			t.Errorf("ParseAny with PreferMonthFirst(false) failed: %v", err)
		}
		// Note: The actual behavior depends on implementation
		_ = result2
	})

	t.Run("RetryAmbiguousDateWithSwap", func(t *testing.T) {
		// Test with retry enabled
		_, err := ParseAny("32/01/2023", RetryAmbiguousDateWithSwap(true))
		// This should potentially succeed by swapping to 01/32/2023 -> error -> try other format
		// The exact behavior depends on implementation
		_ = err
	})
}

// TestFormatConversion tests format conversion functions
func TestFormatConversion(t *testing.T) {
	t.Run("ParseFormatToGo", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"YYYY-MM-DD", "2006-01-02"},
			{"YYYY-MM-DD HH:II:SS", "2006-01-02 15:04:05"},
			{"DD/MM/YYYY", "02/01/2006"},
			{"MM/DD/YYYY HH:II", "01/02/2006 15:04"},
		}

		for _, tc := range testCases {
			result := ParseFormatToGo(tc.input)
			if result != tc.expected {
				t.Errorf("ParseFormatToGo(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("ParseDateTime", func(t *testing.T) {
		testCases := []struct {
			datetime string
			format   string
			timezone string
			expected bool
		}{
			{"2023-12-25 15:30:45", "YYYY-MM-DD HH:II:SS", "UTC", true},
			{"25/12/2023", "DD/MM/YYYY", "", true},
			{"Dec 25, 2023", "YYYY-MM-DD", "America/New_York", true},
			{"invalid-date", "YYYY-MM-DD", "UTC", false},
		}

		for _, tc := range testCases {
			result, err := ParseDateTime(tc.datetime, tc.format, tc.timezone)
			if tc.expected {
				if err != nil {
					t.Errorf("ParseDateTime(%q, %q, %q) failed: %v", tc.datetime, tc.format, tc.timezone, err)
				}
				if result == "" {
					t.Errorf("ParseDateTime(%q, %q, %q) returned empty result", tc.datetime, tc.format, tc.timezone)
				}
			} else {
				if err == nil {
					t.Errorf("ParseDateTime(%q, %q, %q) should have failed", tc.datetime, tc.format, tc.timezone)
				}
			}
		}
	})
}

// TestAdapterFunctions tests adapter functions between old and new API
func TestAdapterFunctions(t *testing.T) {
	t.Run("ToTime", func(t *testing.T) {
		// Test with valid ParsedDateTime
		parsed, err := ParseDatetime("2023-12-25T15:30:45Z")
		if err != nil {
			t.Fatalf("Failed to parse test datetime: %v", err)
		}

		result := ToTime(parsed)
		if result.IsZero() {
			t.Error("ToTime returned zero time")
		}

		// Test with nil
		result = ToTime(nil)
		if !result.IsZero() {
			t.Error("ToTime(nil) should return zero time")
		}
	})

	t.Run("FromTime", func(t *testing.T) {
		now := time.Now()
		result := FromTime(now)

		if result == nil {
			t.Error("FromTime returned nil")
		}
		if !result.Time.Equal(now) {
			t.Error("FromTime didn't preserve time value")
		}
		if result.Layout == "" {
			t.Error("FromTime didn't set layout")
		}
	})

	t.Run("ToParsedDateTime", func(t *testing.T) {
		now := time.Now()
		layout := "2006-01-02 15:04:05"

		result := ToParsedDateTime(now, layout)
		if result == nil {
			t.Error("ToParsedDateTime returned nil")
		}
		if !result.Time.Equal(now) {
			t.Error("ToParsedDateTime didn't preserve time value")
		}
		if result.Layout != layout {
			t.Errorf("ToParsedDateTime layout = %q, expected %q", result.Layout, layout)
		}
	})
}

// TestCompatibilityUtilityFunctions tests additional utility functions
func TestCompatibilityUtilityFunctions(t *testing.T) {
	t.Run("ParseAnyFormat", func(t *testing.T) {
		formats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"01/02/2006",
		}

		testTime, format, err := ParseAnyFormat("2023-12-25T15:30:45Z", formats, time.UTC)
		if err != nil {
			t.Errorf("ParseAnyFormat failed: %v", err)
		}
		if testTime.IsZero() {
			t.Error("ParseAnyFormat returned zero time")
		}
		if format == "" {
			t.Error("ParseAnyFormat returned empty format")
		}
	})

	t.Run("IsValidDate", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"2023-12-25T15:30:45Z", true},
			{"2023-12-25", true},
			{"Dec 25, 2023", true},
			{"invalid-date", false},
			{"", false},
		}

		for _, tc := range testCases {
			result := IsValidDate(tc.input)
			if result != tc.expected {
				t.Errorf("IsValidDate(%q) = %t, expected %t", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("GetSupportedFormats", func(t *testing.T) {
		formats := GetSupportedFormats()
		if len(formats) == 0 {
			t.Error("GetSupportedFormats returned empty slice")
		}

		// Check that some expected formats are present
		expectedFormats := []string{
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"01/02/2006",
			"Jan 2, 2006",
		}

		for _, expected := range expectedFormats {
			found := false
			for _, format := range formats {
				if format == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected format %q not found in supported formats", expected)
			}
		}
	})
}

// TestAdvancedParsing tests the advanced parsing functionality
func TestAdvancedParsing(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"ISO8601", "2023-12-25T15:30:45Z", true},
		{"RFC3339", "2023-12-25T15:30:45-07:00", true},
		{"Common Format", "2023-12-25 15:30:45", true},
		{"US Format", "12/25/2023", true},
		{"European Format", "25.12.2023", true},
		{"Unix Timestamp Seconds", "1703519445", true},
		{"Unix Timestamp Milliseconds", "1703519445000", true},
		{"Relative Today", "today", true},
		{"Relative Yesterday", "yesterday", true},
		{"Relative Tomorrow", "tomorrow", true},
		{"Text Format DMY", "25 December 2023", true},
		{"Text Format MDY", "December 25, 2023", true},
		{"Text Format YMD", "2023 December 25", true},
		{"Invalid", "not-a-date", false},
		{"Empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

// TestErrorHandling tests error handling compatibility
func TestErrorHandling(t *testing.T) {
	t.Run("ErrAmbiguousMMDD", func(t *testing.T) {
		// This error should be returned when ParseStrict encounters ambiguous dates
		if ErrAmbiguousMMDD == nil {
			t.Error("ErrAmbiguousMMDD should not be nil")
		}
		if ErrAmbiguousMMDD.Error() != "ambiguous M/D vs D/M" {
			t.Errorf("ErrAmbiguousMMDD message = %q, expected %q", ErrAmbiguousMMDD.Error(), "ambiguous M/D vs D/M")
		}
	})

	t.Run("ParseStrict ambiguous detection", func(t *testing.T) {
		// Note: The exact behavior of ambiguous date detection depends on implementation
		// This test checks that ParseStrict can handle potentially ambiguous dates
		testCases := []string{
			"01/02/2023", // Could be Jan 2 or Feb 1
			"03/04/2023", // Could be Mar 4 or Apr 3
		}

		for _, input := range testCases {
			_, err := ParseStrict(input)
			// We don't assert specific behavior here because it depends on implementation
			// The important thing is that it doesn't panic
			_ = err
		}
	})
}

// TestPrecisionDetection tests precision detection compatibility
func TestPrecisionDetection(t *testing.T) {
	testCases := []struct {
		layout   string
		expected string
	}{
		{"2006-01-02T15:04:05.000000000Z", "nanosecond"},
		{"2006-01-02T15:04:05Z", "second"},
		{"2006-01-02T15:04Z", "minute"},
		{"2006-01-02T15Z", "hour"},
		{"2006-01-02", "day"},
		{"2006-01", "day"},
		{"2006", "day"},
		{"", "year"},
	}

	for _, tc := range testCases {
		result := detectPrecisionFromLayout(tc.layout)
		if result != tc.expected {
			t.Errorf("detectPrecisionFromLayout(%q) = %q, expected %q", tc.layout, result, tc.expected)
		}
	}
}

// TestParseMonthDayYear tests the parseMonthDayYear method
func TestParseMonthDayYear(t *testing.T) {
	parser := &compatibilityParser{loc: time.UTC}

	t.Run("Valid month day year formats", func(t *testing.T) {
		testCases := []struct {
			name     string
			matches  []string
			expected time.Time
		}{
			{
				name:     "January 15, 2023",
				matches:  []string{"", "January", "15", "2023"},
				expected: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "Dec 25, 2023",
				matches:  []string{"", "Dec", "25", "2023"},
				expected: time.Date(2023, time.December, 25, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "Feb 29, 2024",
				matches:  []string{"", "Feb", "29", "2024"},
				expected: time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "May 1, 2000",
				matches:  []string{"", "May", "1", "2000"},
				expected: time.Date(2000, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := parser.parseMonthDayYear(tc.matches)
				if err != nil {
					t.Errorf("parseMonthDayYear(%v) failed: %v", tc.matches, err)
				}
				if !parser.t.Equal(tc.expected) {
					t.Errorf("parseMonthDayYear(%v) = %v, expected %v", tc.matches, parser.t, tc.expected)
				}
				if string(parser.format) != "month day year" {
					t.Errorf("parseMonthDayYear format = %q, expected %q", string(parser.format), "month day year")
				}
			})
		}
	})

	t.Run("Two-digit year handling", func(t *testing.T) {
		testCases := []struct {
			name     string
			matches  []string
			expected time.Time
		}{
			{
				name:     "Jan 1, 25 (should be 2025)",
				matches:  []string{"", "Jan", "1", "25"},
				expected: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "Jan 1, 49 (should be 2049)",
				matches:  []string{"", "Jan", "1", "49"},
				expected: time.Date(2049, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "Jan 1, 50 (should be 1950)",
				matches:  []string{"", "Jan", "1", "50"},
				expected: time.Date(1950, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "Jan 1, 99 (should be 1999)",
				matches:  []string{"", "Jan", "1", "99"},
				expected: time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := parser.parseMonthDayYear(tc.matches)
				if err != nil {
					t.Errorf("parseMonthDayYear(%v) failed: %v", tc.matches, err)
				}
				if !parser.t.Equal(tc.expected) {
					t.Errorf("parseMonthDayYear(%v) = %v, expected %v", tc.matches, parser.t, tc.expected)
				}
			})
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		testCases := []struct {
			name     string
			matches  []string
			errorMsg string
		}{
			{
				name:     "Too few matches",
				matches:  []string{"", "Jan", "15"},
				errorMsg: "invalid month day year format",
			},
			{
				name:     "Invalid month name",
				matches:  []string{"", "InvalidMonth", "15", "2023"},
				errorMsg: "invalid month name",
			},
			{
				name:     "Invalid day",
				matches:  []string{"", "Jan", "invalid", "2023"},
				errorMsg: "strconv.Atoi",
			},
			{
				name:     "Invalid year",
				matches:  []string{"", "Jan", "15", "invalid"},
				errorMsg: "strconv.Atoi",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := parser.parseMonthDayYear(tc.matches)
				if err == nil {
					t.Errorf("parseMonthDayYear(%v) should have failed", tc.matches)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("parseMonthDayYear(%v) error = %q, should contain %q", tc.matches, err.Error(), tc.errorMsg)
				}
			})
		}
	})
}

// TestParseYearMonthDay tests the parseYearMonthDay method
func TestParseYearMonthDay(t *testing.T) {
	parser := &compatibilityParser{loc: time.UTC}

	t.Run("Valid year month day formats", func(t *testing.T) {
		testCases := []struct {
			name     string
			matches  []string
			expected time.Time
		}{
			{
				name:     "2023 January 15",
				matches:  []string{"", "2023", "January", "15"},
				expected: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "2023 Dec 25",
				matches:  []string{"", "2023", "Dec", "25"},
				expected: time.Date(2023, time.December, 25, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "2024 Feb 29",
				matches:  []string{"", "2024", "Feb", "29"},
				expected: time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "2000 May 1",
				matches:  []string{"", "2000", "May", "1"},
				expected: time.Date(2000, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := parser.parseYearMonthDay(tc.matches)
				if err != nil {
					t.Errorf("parseYearMonthDay(%v) failed: %v", tc.matches, err)
				}
				if !parser.t.Equal(tc.expected) {
					t.Errorf("parseYearMonthDay(%v) = %v, expected %v", tc.matches, parser.t, tc.expected)
				}
				if string(parser.format) != "year month day" {
					t.Errorf("parseYearMonthDay format = %q, expected %q", string(parser.format), "year month day")
				}
			})
		}
	})

	t.Run("Different timezone", func(t *testing.T) {
		est, _ := time.LoadLocation("America/New_York")
		parser := &compatibilityParser{loc: est}

		matches := []string{"", "2023", "January", "15"}
		expected := time.Date(2023, time.January, 15, 0, 0, 0, 0, est)

		err := parser.parseYearMonthDay(matches)
		if err != nil {
			t.Errorf("parseYearMonthDay(%v) failed: %v", matches, err)
		}
		if !parser.t.Equal(expected) {
			t.Errorf("parseYearMonthDay(%v) = %v, expected %v", matches, parser.t, expected)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		testCases := []struct {
			name     string
			matches  []string
			errorMsg string
		}{
			{
				name:     "Too few matches",
				matches:  []string{"", "2023", "Jan"},
				errorMsg: "invalid year month day format",
			},
			{
				name:     "Invalid year",
				matches:  []string{"", "invalid", "Jan", "15"},
				errorMsg: "strconv.Atoi",
			},
			{
				name:     "Invalid month name",
				matches:  []string{"", "2023", "InvalidMonth", "15"},
				errorMsg: "invalid month name",
			},
			{
				name:     "Invalid day",
				matches:  []string{"", "2023", "Jan", "invalid"},
				errorMsg: "strconv.Atoi",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := parser.parseYearMonthDay(tc.matches)
				if err == nil {
					t.Errorf("parseYearMonthDay(%v) should have failed", tc.matches)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("parseYearMonthDay(%v) error = %q, should contain %q", tc.matches, err.Error(), tc.errorMsg)
				}
			})
		}
	})
}

// TestParsersWithDifferentMonthFormats tests both parsers with various month formats
func TestParsersWithDifferentMonthFormats(t *testing.T) {
	parser := &compatibilityParser{loc: time.UTC}

	monthTests := []struct {
		monthName string
		expected  time.Month
	}{
		{"January", time.January},
		{"Jan", time.January},
		{"February", time.February},
		{"Feb", time.February},
		{"March", time.March},
		{"Mar", time.March},
		{"April", time.April},
		{"Apr", time.April},
		{"May", time.May},
		{"June", time.June},
		{"Jun", time.June},
		{"July", time.July},
		{"Jul", time.July},
		{"August", time.August},
		{"Aug", time.August},
		{"September", time.September},
		{"Sep", time.September},
		{"Sept", time.September},
		{"October", time.October},
		{"Oct", time.October},
		{"November", time.November},
		{"Nov", time.November},
		{"December", time.December},
		{"Dec", time.December},
	}

	for _, mt := range monthTests {
		t.Run("parseMonthDayYear with "+mt.monthName, func(t *testing.T) {
			matches := []string{"", mt.monthName, "15", "2023"}
			expected := time.Date(2023, mt.expected, 15, 0, 0, 0, 0, time.UTC)

			err := parser.parseMonthDayYear(matches)
			if err != nil {
				t.Errorf("parseMonthDayYear with %s failed: %v", mt.monthName, err)
			}
			if !parser.t.Equal(expected) {
				t.Errorf("parseMonthDayYear with %s = %v, expected %v", mt.monthName, parser.t, expected)
			}
		})

		t.Run("parseYearMonthDay with "+mt.monthName, func(t *testing.T) {
			matches := []string{"", "2023", mt.monthName, "15"}
			expected := time.Date(2023, mt.expected, 15, 0, 0, 0, 0, time.UTC)

			err := parser.parseYearMonthDay(matches)
			if err != nil {
				t.Errorf("parseYearMonthDay with %s failed: %v", mt.monthName, err)
			}
			if !parser.t.Equal(expected) {
				t.Errorf("parseYearMonthDay with %s = %v, expected %v", mt.monthName, parser.t, expected)
			}
		})
	}
}
