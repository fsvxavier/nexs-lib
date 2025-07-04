package strutl

import (
	"fmt"
	"testing"
)

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		old      string
		new      string
		expected string
	}{
		{"Empty string", "", "x", "y", ""},
		{"Empty old", "hello", "", "x", "hello"},
		{"Empty new", "hello", "l", "", "heo"},
		{"Simple case", "hello world", "world", "golang", "hello golang"},
		{"Multiple replacements", "hello hello", "hello", "hi", "hi hi"},
		{"No matches", "hello", "world", "golang", "hello"},
		{"UTF8 string", "こんにちは世界", "世界", "ワールド", "こんにちはワールド"},
		{"UTF8 replacement", "hello world", "world", "世界", "hello 世界"},
		{"Replace with empty", "hello world", "world", "", "hello "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplaceAll(test.s, test.old, test.new)
			if result != test.expected {
				t.Errorf("ReplaceAll(%q, %q, %q) = %q; want %q",
					test.s, test.old, test.new, result, test.expected)
			}
		})
	}
}

func TestReplaceFirst(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		old      string
		new      string
		expected string
	}{
		{"Empty string", "", "x", "y", ""},
		{"Empty old", "hello", "", "x", "hello"},
		{"Empty new", "hello", "l", "", "helo"},
		{"Simple case", "hello world", "world", "golang", "hello golang"},
		{"Multiple occurrences", "hello hello", "hello", "hi", "hi hello"},
		{"No matches", "hello", "world", "golang", "hello"},
		{"UTF8 string", "こんにちは世界こんにちは世界", "世界", "ワールド", "こんにちはワールドこんにちは世界"},
		{"UTF8 replacement", "hello world", "world", "世界", "hello 世界"},
		{"Replace with empty", "hello world", "world", "", "hello "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplaceFirst(test.s, test.old, test.new)
			if result != test.expected {
				t.Errorf("ReplaceFirst(%q, %q, %q) = %q; want %q",
					test.s, test.old, test.new, result, test.expected)
			}
		})
	}
}

func TestReplaceLast(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		old      string
		new      string
		expected string
	}{
		{"Empty string", "", "x", "y", ""},
		{"Empty old", "hello", "", "x", "hello"},
		{"Empty new", "hello", "l", "", "helo"},
		{"Simple case", "hello world", "world", "golang", "hello golang"},
		{"Multiple occurrences", "hello hello", "hello", "hi", "hello hi"},
		{"No matches", "hello", "world", "golang", "hello"},
		{"UTF8 string", "こんにちは世界こんにちは世界", "世界", "ワールド", "こんにちは世界こんにちはワールド"},
		{"UTF8 replacement", "hello world", "world", "世界", "hello 世界"},
		{"Replace with empty", "hello world", "world", "", "hello "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplaceLast(test.s, test.old, test.new)
			if result != test.expected {
				t.Errorf("ReplaceLast(%q, %q, %q) = %q; want %q",
					test.s, test.old, test.new, result, test.expected)
			}
		})
	}
}

func TestReplaceNonAlphanumeric(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		replacement string
		expected    string
	}{
		{"Empty string", "", "-", ""},
		{"No non-alphanumeric", "hello123", "-", "hello123"},
		{"Simple case", "hello, world!", "-", "hello--world-"},
		{"Email address", "test@example.com", "_", "test_example_com"},
		{"URL", "https://example.com/path?q=1", "_", "https___example_com_path_q_1"},
		{"Mixed characters", "abc123!@#$%^", "_", "abc123______"},
		{"UTF8 alphanumeric", "こんにちは123", "-", "こんにちは123"},
		{"UTF8 with punctuation", "こんにちは、世界！", "-", "こんにちは-世界-"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplaceNonAlphanumeric(test.s, test.replacement)
			if result != test.expected {
				t.Errorf("ReplaceNonAlphanumeric(%q, %q) = %q; want %q",
					test.s, test.replacement, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleReplaceAll() {
	fmt.Println(ReplaceAll("hello world", "world", "golang"))
	fmt.Println(ReplaceAll("hello hello", "hello", "hi"))
	fmt.Println(ReplaceAll("test test test", "test", "exam"))
	// Output:
	// hello golang
	// hi hi
	// exam exam exam
}

func ExampleReplaceFirst() {
	fmt.Println(ReplaceFirst("hello hello", "hello", "hi"))
	fmt.Println(ReplaceFirst("test test test", "test", "exam"))
	// Output:
	// hi hello
	// exam test test
}

func ExampleReplaceLast() {
	fmt.Println(ReplaceLast("hello hello", "hello", "hi"))
	fmt.Println(ReplaceLast("test test test", "test", "exam"))
	// Output:
	// hello hi
	// test test exam
}

func ExampleReplaceNonAlphanumeric() {
	fmt.Println(ReplaceNonAlphanumeric("hello, world!", "-"))
	fmt.Println(ReplaceNonAlphanumeric("test@example.com", "_"))
	// Output:
	// hello--world-
	// test_example_com
}

// Benchmarks
func BenchmarkReplaceAll(b *testing.B) {
	benchCases := []struct {
		name string
		s    string
		old  string
		new  string
	}{
		{"Short ASCII", "hello world", "world", "golang"},
		{"Long ASCII", "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog.", "fox", "cat"},
		{"Short UTF8", "こんにちは世界", "世界", "ワールド"},
		{"No match", "hello world", "xyz", "abc"},
		{"Multiple replacements", "test test test test", "test", "exam"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ReplaceAll(bc.s, bc.old, bc.new)
			}
		})
	}
}

func BenchmarkReplaceFirst(b *testing.B) {
	benchCases := []struct {
		name string
		s    string
		old  string
		new  string
	}{
		{"Short ASCII", "hello world", "world", "golang"},
		{"Long ASCII", "The quick brown fox jumps over the lazy dog. The fox is quick.", "fox", "cat"},
		{"Short UTF8", "こんにちは世界", "世界", "ワールド"},
		{"No match", "hello world", "xyz", "abc"},
		{"Multiple matches", "test test test test", "test", "exam"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ReplaceFirst(bc.s, bc.old, bc.new)
			}
		})
	}
}

func BenchmarkReplaceLast(b *testing.B) {
	benchCases := []struct {
		name string
		s    string
		old  string
		new  string
	}{
		{"Short ASCII", "hello world", "world", "golang"},
		{"Long ASCII", "The quick brown fox jumps over the lazy dog. The fox is quick.", "fox", "cat"},
		{"Short UTF8", "こんにちは世界", "世界", "ワールド"},
		{"No match", "hello world", "xyz", "abc"},
		{"Multiple matches", "test test test test", "test", "exam"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ReplaceLast(bc.s, bc.old, bc.new)
			}
		})
	}
}

func BenchmarkReplaceNonAlphanumeric(b *testing.B) {
	benchCases := []struct {
		name        string
		s           string
		replacement string
	}{
		{"No special chars", "helloworld123", "-"},
		{"Some special chars", "hello, world!", "-"},
		{"Many special chars", "!@#$%^&*()", "-"},
		{"Mixed content", "test@example.com, user@domain.org", "_"},
		{"UTF8 content", "こんにちは、世界！", "-"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ReplaceNonAlphanumeric(bc.s, bc.replacement)
			}
		})
	}
}
