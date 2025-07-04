package strutl

import (
	"fmt"
	"testing"
)

func TestSummary(t *testing.T) {
	tests := []struct {
		name     string
		end      string
		input    string
		expected string
		length   int
	}{
		{"Empty string", "...", "", "", 15},
		{"Zero length", "...", "Lorem ipsum dolor sit amet", "Lorem ipsum dolor sit amet", 0},
		{"Basic summary", "...", "Lorem ipsum dolor sit amet", "Lorem ipsum...", 12},
		{"With newline", "...", "Lorem\nipsum dolor sit amet", "Lorem...", 12},
		{"With tabs", "...", "Lorem\tipsum\tdolor sit amet", "Lorem\tipsum...", 12},
		{"Short length", "...", "Lorem ipsum dolor sit amet", "Lorem...", 10},
		{"Newline with short length", "...", "Lorem\nipsum dolor sit amet", "Lorem...", 10},
		{"String shorter than length", "...", "Lorem ipsum", "Lorem ipsum", 15},
		{"Exactly at word boundary", "...", "Lorem ipsum", "Lorem...", 5},
		{"Break within word", "...", "Lorem ipsum", "Lore...", 4},
		{"Extra spaces", "...", "Lorem         ipsum", "Lorem...", 15},
		{"Custom end string", "→", "Lorem ipsum dolor", "Lorem ipsum→", 11},
		{"Empty end string", "", "Lorem ipsum dolor", "Lorem ipsum", 11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Summary(test.input, test.length, test.end)
			if result != test.expected {
				t.Errorf("Summary(%q, %d, %q) = %q; want %q",
					test.input, test.length, test.end, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleSummary() {
	fmt.Println(Summary("Lorem ipsum dolor sit amet.", 12, "..."))
	// Output: Lorem ipsum...
}

func ExampleSummary_withNewline() {
	fmt.Println(Summary("Lorem\nipsum dolor sit amet.", 10, "..."))
	// Output: Lorem...
}

// Benchmarks
func BenchmarkSummary(b *testing.B) {
	benchCases := []struct {
		name   string
		input  string
		length int
		end    string
	}{
		{"ShortText", "Lorem ipsum dolor sit amet", 10, "..."},
		{"MediumText", "Lorem ipsum dolor sit amet, consectetur adipiscing elit", 20, "..."},
		{"LongText", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 30, "..."},
		{"WithNewline", "Lorem ipsum\ndolor sit amet, consectetur adipiscing elit", 15, "..."},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Summary(bc.input, bc.length, bc.end)
			}
		})
	}
}
