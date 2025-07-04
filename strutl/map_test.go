package strutl

import (
	"fmt"
	"strings"
	"testing"
)

func TestMapLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fn       func(string) string
		expected string
	}{
		{"Empty string", "", strings.ToUpper, ""},
		{"Single line", "hello", strings.ToUpper, "HELLO"},
		{"Multiple lines", "line1\nline2\nline3", strings.ToUpper, "LINE1\nLINE2\nLINE3"},
		{"With empty lines", "\nline1\n\nline2", strings.ToUpper, "\nLINE1\n\nLINE2"},
		{
			"Custom function",
			"hello\nworld",
			func(s string) string { return "prefix-" + s },
			"prefix-hello\nprefix-world",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MapLines(test.input, test.fn)
			if result != test.expected {
				t.Errorf("MapLines(%q, fn) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestSplitAndMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		split    string
		fn       func(string) string
		expected string
	}{
		{"Empty string", "", ",", strings.ToUpper, ""},
		{"No split char", "hello", ",", strings.ToUpper, "HELLO"},
		{"Simple split", "a,b,c", ",", strings.ToUpper, "A,B,C"},
		{"Custom function", "tag1|tag2|tag3", "|", func(s string) string { return "prefix-" + s }, "prefix-tag1|prefix-tag2|prefix-tag3"},
		{"Empty parts", ",,", ",", strings.ToUpper, ",,"},
		{"With spaces", "a, b, c", ",", strings.TrimSpace, "a,b,c"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SplitAndMap(test.input, test.split, test.fn)
			if result != test.expected {
				t.Errorf("SplitAndMap(%q, %q, fn) = %q; want %q",
					test.input, test.split, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleMapLines() {
	result := MapLines("linha1\nlinha2\nlinha3", strings.ToUpper)
	fmt.Println(result)
	// Output:
	// LINHA1
	// LINHA2
	// LINHA3
}

func ExampleSplitAndMap() {
	result := SplitAndMap("tag1|tag2|tag3", "|", func(s string) string {
		return "prefix-" + s
	})
	fmt.Println(result)
	// Output: prefix-tag1|prefix-tag2|prefix-tag3
}

// Benchmarks
func BenchmarkMapLines(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{"Short", "line1\nline2"},
		{"Medium", "line1\nline2\nline3\nline4\nline5"},
		{"Long", strings.Repeat("line\n", 100)},
	}

	for _, input := range inputs {
		b.Run(input.name, func(b *testing.B) {
			fn := strings.ToUpper
			for i := 0; i < b.N; i++ {
				_ = MapLines(input.input, fn)
			}
		})
	}
}

func BenchmarkSplitAndMap(b *testing.B) {
	inputs := []struct {
		name  string
		input string
		split string
	}{
		{"Short", "a,b,c", ","},
		{"Medium", "item1,item2,item3,item4,item5", ","},
		{"Long", strings.Repeat("item,", 100) + "lastitem", ","},
		{"NoSplitChar", "hello", ","},
	}

	for _, input := range inputs {
		b.Run(input.name, func(b *testing.B) {
			fn := strings.ToUpper
			for i := 0; i < b.N; i++ {
				_ = SplitAndMap(input.input, input.split, fn)
			}
		})
	}
}
