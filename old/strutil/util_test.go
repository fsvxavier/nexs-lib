package strutil

import (
	"testing"
)

func TestLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"こんにちは", 5},
		{"", 0},
		{"a", 1},
		{"😊", 1},
	}

	for _, test := range tests {
		result := Len(test.input)
		if result != test.expected {
			t.Errorf("Len(%q) = %d; want %d", test.input, result, test.expected)
		}
	}
}

func TestConfigureAcronym(t *testing.T) {
	ConfigureAcronym("ID", "id")
	val, ok := uppercaseAcronym.Load("ID")
	if !ok || val != "id" {
		t.Errorf("ConfigureAcronym failed, got %v, want %v", val, "id")
	}
}
