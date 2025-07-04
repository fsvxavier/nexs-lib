package strutl

import (
	"testing"
)

func TestLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ASCII string", "hello", 5},
		{"Japanese characters", "こんにちは", 5},
		{"Empty string", "", 0},
		{"Single character", "a", 1},
		{"Emoji", "😊", 1},
		{"Mixed characters", "a世界b", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Len(test.input)
			if result != test.expected {
				t.Errorf("Len(%q) = %d; want %d", test.input, result, test.expected)
			}
		})
	}
}

func TestConfigureAcronym(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"Simple acronym", "ID", "id"},
		{"Complex acronym", "API", "api"},
		{"Long acronym", "POSTGRESQL", "postgresql"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Configure the acronym
			ConfigureAcronym(test.key, test.value)

			// Retrieve and verify
			val, ok := GetAcronym(test.key)
			if !ok {
				t.Errorf("GetAcronym failed to retrieve key %s", test.key)
			}
			if val != test.value {
				t.Errorf("GetAcronym(%q) = %q; want %q", test.key, val, test.value)
			}
		})
	}
}

func TestGetAcronym(t *testing.T) {
	// Test for non-existent acronym
	_, ok := GetAcronym("NONEXISTENT")
	if ok {
		t.Error("GetAcronym returned true for non-existent acronym")
	}

	// Set an acronym and retrieve it
	ConfigureAcronym("TEST", "test")
	val, ok := GetAcronym("TEST")
	if !ok || val != "test" {
		t.Errorf("GetAcronym failed, got %v (ok=%v), want %v", val, ok, "test")
	}
}

func BenchmarkLen(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
	}{
		{"ASCII", "hello world"},
		{"Unicode", "こんにちは世界"},
		{"Mixed", "hello世界"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Len(bc.input)
			}
		})
	}
}

func BenchmarkConfigureAcronym(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConfigureAcronym("BENCH", "bench")
	}
}

func BenchmarkGetAcronym(b *testing.B) {
	ConfigureAcronym("BENCH", "bench")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = GetAcronym("BENCH")
	}
}
