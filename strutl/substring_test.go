package strutl

import (
	"fmt"
	"testing"
)

func TestSubstring(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{"Empty string", "", 0, 0, ""},
		{"Negative start", "hello", -1, 5, ""},
		{"Start > end", "hello", 3, 2, ""},
		{"Start > length", "hello", 10, 15, ""},
		{"End > length", "hello", 0, 10, "hello"},
		{"ASCII substring", "hello world", 0, 5, "hello"},
		{"ASCII substring mid", "hello world", 6, 11, "world"},
		{"ASCII substring with end 0", "hello world", 6, 0, "world"},
		{"ASCII substring with negative end", "hello world", 0, -6, "hello"},
		{"UTF8 substring", "こんにちは世界", 0, 3, "こんに"},
		{"UTF8 substring mid", "こんにちは世界", 3, 5, "ちは"},
		{"UTF8 substring end", "こんにちは世界", 5, 7, "世界"},
		{"UTF8 substring with negative end", "こんにちは世界", 0, -2, "こんにちは"},
		{"Mixed ASCII and UTF8", "hello 世界", 0, 6, "hello "},
		{"Mixed ASCII and UTF8 end", "hello 世界", 6, 8, "世界"},
		{"Emoji substring", "😊🚀✨", 0, 2, "😊🚀"},
		{"Complete substring", "hello", 0, 5, "hello"},
		{"Complete UTF8 substring", "こんにちは", 0, 5, "こんにちは"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Substring(test.input, test.start, test.end)
			if result != test.expected {
				t.Errorf("Substring(%q, %d, %d) = %q; want %q",
					test.input, test.start, test.end, result, test.expected)
			}
		})
	}
}

func TestSubstringAfter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected string
	}{
		{"Empty string", "", "-", ""},
		{"Empty separator", "abc", "", "abc"},
		{"Separator not found", "abc-def-ghi", ".", ""},
		{"Simple case", "abc-def-ghi", "-", "def-ghi"},
		{"Multiple separators", "abc-def-ghi", "-", "def-ghi"},
		{"UTF8 string", "こんにちは-世界", "-", "世界"},
		{"UTF8 separator", "hello世界bye", "世界", "bye"},
		{"Separator at end", "hello-", "-", ""},
		{"Separator at start", "-hello", "-", "hello"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SubstringAfter(test.input, test.sep)
			if result != test.expected {
				t.Errorf("SubstringAfter(%q, %q) = %q; want %q",
					test.input, test.sep, result, test.expected)
			}
		})
	}
}

func TestSubstringAfterLast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected string
	}{
		{"Empty string", "", "-", ""},
		{"Empty separator", "abc", "", "abc"},
		{"Separator not found", "abc-def-ghi", ".", ""},
		{"Simple case", "abc-def-ghi", "-", "ghi"},
		{"Single separator", "abc-def", "-", "def"},
		{"UTF8 string", "こんにちは-世界-!", "-", "!"},
		{"UTF8 separator", "hello世界bye世界end", "世界", "end"},
		{"Separator at end", "hello-", "-", ""},
		{"Separator at start", "-hello", "-", "hello"},
		{"Multiple identical separators", "a--b", "-", "b"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SubstringAfterLast(test.input, test.sep)
			if result != test.expected {
				t.Errorf("SubstringAfterLast(%q, %q) = %q; want %q",
					test.input, test.sep, result, test.expected)
			}
		})
	}
}

func TestSubstringBefore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected string
	}{
		{"Empty string", "", "-", ""},
		{"Empty separator", "abc", "", ""},
		{"Separator not found", "abc-def-ghi", ".", "abc-def-ghi"},
		{"Simple case", "abc-def-ghi", "-", "abc"},
		{"Multiple separators", "abc-def-ghi", "-", "abc"},
		{"UTF8 string", "こんにちは-世界", "-", "こんにちは"},
		{"UTF8 separator", "hello世界bye", "世界", "hello"},
		{"Separator at end", "hello-", "-", "hello"},
		{"Separator at start", "-hello", "-", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SubstringBefore(test.input, test.sep)
			if result != test.expected {
				t.Errorf("SubstringBefore(%q, %q) = %q; want %q",
					test.input, test.sep, result, test.expected)
			}
		})
	}
}

func TestSubstringBeforeLast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected string
	}{
		{"Empty string", "", "-", ""},
		{"Empty separator", "abc", "", ""},
		{"Separator not found", "abc-def-ghi", ".", "abc-def-ghi"},
		{"Simple case", "abc-def-ghi", "-", "abc-def"},
		{"Single separator", "abc-def", "-", "abc"},
		{"UTF8 string", "こんにちは-世界-!", "-", "こんにちは-世界"},
		{"UTF8 separator", "hello世界bye世界end", "世界", "hello世界bye"},
		{"Separator at end", "hello-", "-", "hello"},
		{"Separator at start", "-hello", "-", ""},
		{"Multiple identical separators", "a--b", "-", "a-"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SubstringBeforeLast(test.input, test.sep)
			if result != test.expected {
				t.Errorf("SubstringBeforeLast(%q, %q) = %q; want %q",
					test.input, test.sep, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleSubstring() {
	fmt.Println(Substring("Hello, 世界!", 0, 5))
	fmt.Println(Substring("Hello, 世界!", 7, 9))
	fmt.Println(Substring("Hello, 世界!", 0, -2))
	// Output:
	// Hello
	// 世界
	// Hello, 世
}

func ExampleSubstringAfter() {
	fmt.Println(SubstringAfter("abc-def-ghi", "-"))
	fmt.Println(SubstringAfter("abc-def-ghi", "."))
	fmt.Println(SubstringAfter("abc", ""))
	// Output:
	// def-ghi
	//
	// abc
}

func ExampleSubstringAfterLast() {
	fmt.Println(SubstringAfterLast("abc-def-ghi", "-"))
	fmt.Println(SubstringAfterLast("abc-def-ghi", "."))
	fmt.Println(SubstringAfterLast("abc", ""))
	// Output:
	// ghi
	//
	// abc
}

func ExampleSubstringBefore() {
	fmt.Println(SubstringBefore("abc-def-ghi", "-"))
	fmt.Println(SubstringBefore("abc-def-ghi", "."))
	fmt.Println(SubstringBefore("abc", ""))
	// Output:
	// abc
	// abc-def-ghi
	//
}

func ExampleSubstringBeforeLast() {
	fmt.Println(SubstringBeforeLast("abc-def-ghi", "-"))
	fmt.Println(SubstringBeforeLast("abc-def-ghi", "."))
	fmt.Println(SubstringBeforeLast("abc", ""))
	// Output:
	// abc-def
	// abc-def-ghi
	//
}

// Benchmarks
var benchStrings = []struct {
	name  string
	input string
}{
	{"ASCII", "hello world this is a test string for benchmarking"},
	{"Unicode", "こんにちは世界これはベンチマーク用のテスト文字列です"},
	{"Mixed", "Hello 世界 this is こんにちは a mixed ベンチマーク string テスト"},
}

func BenchmarkSubstring(b *testing.B) {
	for _, bs := range benchStrings {
		mid := Len(bs.input) / 2

		b.Run(bs.name+"_Start", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Substring(bs.input, 0, mid)
			}
		})

		b.Run(bs.name+"_Middle", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Substring(bs.input, mid/2, mid*3/2)
			}
		})

		b.Run(bs.name+"_End", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Substring(bs.input, mid, 0)
			}
		})
	}
}

func BenchmarkSubstringAfter(b *testing.B) {
	for _, bs := range benchStrings {
		if len(bs.input) < 10 {
			continue
		}

		sep := Substring(bs.input, 5, 10)

		b.Run(bs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SubstringAfter(bs.input, sep)
			}
		})
	}
}

func BenchmarkSubstringAfterLast(b *testing.B) {
	for _, bs := range benchStrings {
		if len(bs.input) < 10 {
			continue
		}

		sep := Substring(bs.input, 5, 6) // Uma letra que deve aparecer várias vezes

		b.Run(bs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SubstringAfterLast(bs.input, sep)
			}
		})
	}
}

func BenchmarkSubstringBefore(b *testing.B) {
	for _, bs := range benchStrings {
		if len(bs.input) < 10 {
			continue
		}

		sep := Substring(bs.input, 15, 20)

		b.Run(bs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SubstringBefore(bs.input, sep)
			}
		})
	}
}

func BenchmarkSubstringBeforeLast(b *testing.B) {
	for _, bs := range benchStrings {
		if len(bs.input) < 10 {
			continue
		}

		sep := Substring(bs.input, 5, 6) // Uma letra que deve aparecer várias vezes

		b.Run(bs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SubstringBeforeLast(bs.input, sep)
			}
		})
	}
}
