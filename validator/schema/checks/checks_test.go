package checks

import (
	"testing"
)

func TestDateTimeChecker(t *testing.T) {
	checker := DateTimeChecker{}

	validCases := []string{
		"14:30:45",                       // time.TimeOnly
		"14:30:45+02:00",                 // RFC3339TimeOnlyFormat
		"2023-12-25",                     // time.DateOnly
		"2023-12-25T14:30:45Z",           // time.RFC3339
		"2023-12-25T14:30:45.123456789Z", // time.RFC3339Nano
		"2023-12-25T14:30:45.123Z",       // ISO8601DateTimeFormat
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to be valid datetime format", testCase)
		}
	}

	invalidCases := []interface{}{
		"invalid-date",
		123,
		nil,
		"",
		"2006-13-01", // Invalid month
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid datetime format", testCase)
		}
	}

	if checker.FormatName() != "date_time" {
		t.Errorf("Expected format name to be 'date_time', got %s", checker.FormatName())
	}
}

func TestISO8601DateChecker(t *testing.T) {
	checker := ISO8601DateChecker{}

	validCases := []string{
		"2006-01-02",
		"2023-12-25",
		"1999-01-01",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to be valid ISO8601 date format", testCase)
		}
	}

	invalidCases := []interface{}{
		"2006-13-01", // Invalid month
		"invalid-date",
		123,
		nil,
		"",
		"2006-01-02T15:04:05Z", // This is datetime, not just date
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid ISO8601 date format", testCase)
		}
	}

	if checker.FormatName() != "iso_8601_date" {
		t.Errorf("Expected format name to be 'iso_8601_date', got %s", checker.FormatName())
	}
}

func TestStringChecker(t *testing.T) {
	checker := StringChecker{}

	validCases := []interface{}{
		"hello",
		"",
		"test string",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be valid string", testCase)
		}
	}

	invalidCases := []interface{}{
		123,
		nil,
		true,
		[]string{},
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid string", testCase)
		}
	}

	if checker.FormatName() != "string" {
		t.Errorf("Expected format name to be 'string', got %s", checker.FormatName())
	}
}

func TestEmptyStringChecker(t *testing.T) {
	checker := EmptyStringChecker{}

	// Valid (empty) cases
	validCases := []interface{}{
		"",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be valid empty string", testCase)
		}
	}

	// Invalid (non-empty or non-string) cases
	invalidCases := []interface{}{
		"hello",
		" ",
		123,
		nil,
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid empty string", testCase)
		}
	}

	if checker.FormatName() != "empty_string" {
		t.Errorf("Expected format name to be 'empty_string', got %s", checker.FormatName())
	}
}

func TestStrongNameChecker(t *testing.T) {
	checker := StrongNameChecker{}

	validCases := []string{
		"ValidName",
		"valid_name",
		"validName123",
		"Valid-Name",
		"a",
		"A1_B2-C3",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to be valid strong name", testCase)
		}
	}

	invalidCases := []interface{}{
		"",           // Empty
		"123invalid", // Starts with number
		"_invalid",   // Starts with underscore
		"-invalid",   // Starts with hyphen
		"invalid space",
		123,
		nil,
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid strong name", testCase)
		}
	}

	if checker.FormatName() != "strong_name" {
		t.Errorf("Expected format name to be 'strong_name', got %s", checker.FormatName())
	}
}

func TestTextMatchChecker(t *testing.T) {
	checker := TextMatchChecker{}

	validCases := []string{
		"hello world",
		"text_with_underscore",
		"TEXT",
		"text",
		"",
		"   ",
		"a b c d e",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to be valid text match", testCase)
		}
	}

	invalidCases := []interface{}{
		"text123",     // Contains numbers
		"text-dash",   // Contains dash
		"text@symbol", // Contains symbol
		123,
		nil,
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid text match", testCase)
		}
	}

	if checker.FormatName() != "text_match" {
		t.Errorf("Expected format name to be 'text_match', got %s", checker.FormatName())
	}
}

func TestTextMatchWithNumberChecker(t *testing.T) {
	checker := TextMatchWithNumberChecker{}

	validCases := []string{
		"hello world",
		"text123",
		"text_with_underscore",
		"TEXT123",
		"",
		"   ",
		"a1 b2 c3",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to be valid text match with number", testCase)
		}
	}

	invalidCases := []interface{}{
		"text-dash",   // Contains dash
		"text@symbol", // Contains symbol
		"text0",       // Contains 0 (only 1-9 allowed)
		123,
		nil,
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to be invalid text match with number", testCase)
		}
	}

	if checker.FormatName() != "text_match_with_number" {
		t.Errorf("Expected format name to be 'text_match_with_number', got %s", checker.FormatName())
	}
}

func TestTextMatchCustomChecker(t *testing.T) {
	// Test with valid regex
	checker := NewTextMatchCustom("^[0-9]+$") // Only digits

	validCases := []string{
		"123",
		"0",
		"999999",
	}

	for _, testCase := range validCases {
		if !checker.IsFormat(testCase) {
			t.Errorf("Expected %s to match custom pattern", testCase)
		}
	}

	invalidCases := []interface{}{
		"123abc",
		"abc",
		"",
		123,
		nil,
	}

	for _, testCase := range invalidCases {
		if checker.IsFormat(testCase) {
			t.Errorf("Expected %v to not match custom pattern", testCase)
		}
	}

	// Test with invalid regex
	invalidChecker := NewTextMatchCustom("[invalid")
	if invalidChecker.IsFormat("anything") {
		t.Error("Invalid regex checker should never match")
	}

	if checker.FormatName() != "text_match_custom" {
		t.Errorf("Expected format name to be 'text_match_custom', got %s", checker.FormatName())
	}
}
