package strutl

import (
	"testing"
)

func TestTile(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		length   int
		expected string
	}{
		{"Empty pattern", "", 5, ""},
		{"Zero length", "abc", 0, ""},
		{"Negative length", "abc", -1, ""},
		{"Single char pattern", "a", 5, "aaaaa"},
		{"Pattern shorter than length", "ab", 5, "ababa"},
		{"Pattern equal to length", "abc", 3, "abc"},
		{"Pattern longer than length", "abcdef", 3, "abc"},
		{"Pattern with Unicode", "世界", 4, "世界世界"},
		{"Pattern with Unicode truncated", "世界", 3, "世界世"},
		{"Special characters", "-_-", 7, "-_--_--"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Tile(test.pattern, test.length)
			if result != test.expected {
				t.Errorf("Tile(%q, %d) = %q; want %q",
					test.pattern, test.length, result, test.expected)
			}

			// Verifica se o comprimento do resultado está correto
			if Len(result) != test.length && test.length > 0 && test.pattern != "" {
				t.Errorf("Len(Tile(%q, %d)) = %d; want %d",
					test.pattern, test.length, Len(result), test.length)
			}
		})
	}
}

func BenchmarkTile(b *testing.B) {
	benchCases := []struct {
		name    string
		pattern string
		length  int
	}{
		{"SingleChar", "x", 50},
		{"MultiChar", "abc", 50},
		{"Unicode", "世界", 50},
		{"LongPattern", "abcdefghijklmnopqrstuvwxyz", 100},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Tile(bc.pattern, bc.length)
			}
		})
	}
}
