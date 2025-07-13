package checks

import (
	"testing"
)

func TestTextMatch_IsFormat(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{"hello", true},
		{"Hello World", true},
		{"hello_world", true},
		{"hello123", false},
		{"123", false},
		{"hello world!", false},
		{123, false},
		{nil, false},
	}

	for _, test := range tests {
		tm := TextMatch{}
		result := tm.IsFormat(test.input)
		if result != test.expected {
			t.Errorf("IsFormat(%v) = %v; expected %v", test.input, result, test.expected)
		}
	}
}
func TestTextMatchWithNumber_IsFormat(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{"hello", true},
		{"Hello World", true},
		{"hello_world", true},
		{"hello123", true},
		{"123", true},
		{"hello world!", false},
		{123, false},
		{nil, false},
	}

	for _, test := range tests {
		tm := TextMatchWithNumber{}
		result := tm.IsFormat(test.input)
		if result != test.expected {
			t.Errorf("IsFormat(%v) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestTextMatchCustom_IsFormat(t *testing.T) {
	tests := []struct {
		regex    string
		input    interface{}
		expected bool
	}{
		{`^[a-zA-Z0-9_ ]*$`, "hello", true},
		{`^[a-zA-Z0-9_ ]*$`, "Hello World", true},
		{`^[a-zA-Z0-9_ ]*$`, "hello_world", true},
		{`^[a-zA-Z0-9_ -]*$`, "hello-world", true},
		{`^[a-zA-Z0-9_ ]*$`, "hello123", true},
		{`^[a-zA-Z0-9_ ]*$`, "123", true},
		{`^[a-zA-Z0-9_ ]*$`, "hello world!", false},
		{`^[0-9]+$`, "12345", true},
		{`^[0-9]+$`, "abc123", false},
		{`^[a-z]+$`, "abc", true},
		{`^[a-z]+$`, "ABC", false},
		{`^\d{3}-\d{2}-\d{4}$`, "123-45-6789", true},
		{`^\d{3}-\d{2}-\d{4}$`, "123456789", false},
		{`.*`, 123, false},
		{`.*`, nil, false},
	}

	for _, test := range tests {
		tm := NewTextMatchCustom(test.regex)
		result := tm.IsFormat(test.input)
		if result != test.expected {
			t.Errorf("Regex: %q, IsFormat(%v) = %v; expected %v", test.regex, test.input, result, test.expected)
		}
	}
}
