package strutl

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Empty string", "", 0},
		{"Single word", "hello", 1},
		{"Two words", "hello world", 2},
		{"Multiple words", "one two three four", 4},
		{"With punctuation", "hello, world! How are you?", 5},
		{"With apostrophe", "don't can't won't", 3},
		{"With hyphen", "well-known self-service", 2},
		{"With underscore", "variable_name another_var", 2},
		{"Numbers", "123 456 789", 3},
		{"Mixed", "Hello123 world-wide test_case", 3},
		{"Unicode", "こんにちは 世界", 2},
		{"Extra spaces", "  hello   world  ", 2},
		{"Only punctuation", ".,;:!?", 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CountWords(test.input)
			if result != test.expected {
				t.Errorf("CountWords(%q) = %d; want %d", test.input, result, test.expected)
			}
		})
	}
}

func TestWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty string", "", []string{}},
		{"Single word", "hello", []string{"hello"}},
		{"Two words", "hello world", []string{"hello", "world"}},
		{"With punctuation", "hello, world!", []string{"hello", "world"}},
		{"With apostrophe", "don't can't", []string{"don't", "can't"}},
		{"With hyphen", "well-known", []string{"well-known"}},
		{"With underscore", "variable_name", []string{"variable_name"}},
		{"Numbers", "123 456", []string{"123", "456"}},
		{"Mixed", "Hello123 world-wide", []string{"Hello123", "world-wide"}},
		{"Unicode", "こんにちは 世界", []string{"こんにちは", "世界"}},
		{"Extra spaces", "  hello   world  ", []string{"hello", "world"}},
		{"Only punctuation", ".,;:!?", []string{}},
		{"Complex", "O'Neil's micro-service is user_friendly!", []string{"O'Neil's", "micro-service", "is", "user_friendly"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Words(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Words(%q) = %v; want %v", test.input, result, test.expected)
			}

			// Verifica se a contagem de palavras está correta
			if len(result) != CountWords(test.input) {
				t.Errorf("len(Words(%q)) = %d; CountWords(%q) = %d - counts should be equal",
					test.input, len(result), test.input, CountWords(test.input))
			}
		})
	}
}

// Exemplos para documentação
func ExampleCountWords() {
	fmt.Println(CountWords("hello world"))
	fmt.Println(CountWords("one,two three"))
	fmt.Println(CountWords(""))
	fmt.Println(CountWords("hello-world"))
	// Output:
	// 2
	// 3
	// 0
	// 1
}

func ExampleWords() {
	fmt.Println(Words("hello world"))
	fmt.Println(Words("O'Neil's micro-service"))
	fmt.Println(Words("word1, word2; word3"))
	// Output:
	// [hello world]
	// [O'Neil's micro-service]
	// [word1 word2 word3]
}

// Benchmarks
func BenchmarkCountWords(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Short", "hello world"},
		{"Medium", "The quick brown fox jumps over the lazy dog"},
		{"Long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
		{"With Punctuation", "Hello, world! This is a test: count-these_words."},
		{"Unicode", "こんにちは 世界 これは テスト です"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = CountWords(bc.input)
			}
		})
	}
}

func BenchmarkWords(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Short", "hello world"},
		{"Medium", "The quick brown fox jumps over the lazy dog"},
		{"Long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
		{"With Punctuation", "Hello, world! This is a test: count-these_words."},
		{"Unicode", "こんにちは 世界 これは テスト です"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Words(bc.input)
			}
		})
	}
}
