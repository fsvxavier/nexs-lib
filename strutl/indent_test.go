package strutl

import (
	"fmt"
	"strings"
	"testing"
)

func TestIndent(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		left     string
		expected string
	}{
		{"Empty string", "", "", ""},
		{"Empty string with indent", "", "  ", ""},
		{"Empty indent", "hello", "", "hello"},
		{"Single line", "hello", "  ", "  hello"},
		{"Multiple lines", "linha1\nlinha2", "  ", "  linha1\n  linha2"},
		{"With empty lines", "a\n\nb", "-", "-a\n-\n-b"},
		{"Tab indent", "teste", "\t", "\tteste"},
		{"Multiple char indent", "abc", ">> ", ">> abc"},
		{"Unicode indent", "world", "世界 ", "世界 world"},
		{"Unicode text", "こんにちは", "- ", "- こんにちは"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Indent(test.str, test.left)
			if result != test.expected {
				t.Errorf("Indent(%q, %q) = %q; want %q",
					test.str, test.left, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleIndent() {
	fmt.Println(Indent("linha1\nlinha2", "  "))
	fmt.Println("---")
	fmt.Println(Indent("texto", "> "))
	fmt.Println("---")
	fmt.Println(Indent("a\n\nb", "-"))
	// Output:
	//   linha1
	//   linha2
	// ---
	// > texto
	// ---
	// -a
	// -
	// -b
}

// Benchmarks
func BenchmarkIndent(b *testing.B) {
	benchCases := []struct {
		name string
		str  string
		left string
	}{
		{"SingleLine", "hello world", "  "},
		{"MultiLine", "line1\nline2\nline3\nline4\nline5", "    "},
		{"LongText", strings.Repeat("This is a longer line of text to benchmark indentation.\n", 50), "> "},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Indent(bc.str, bc.left)
			}
		})
	}
}
