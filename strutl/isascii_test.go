package strutl

import (
	"fmt"
	"testing"
)

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"ASCII letters", "abcdefghijklmnopqrstuvwxyz", true},
		{"ASCII uppercase", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", true},
		{"ASCII numbers", "0123456789", true},
		{"ASCII symbols", "!@#$%^&*()_+-=[]{}|;':,./<>?`~", true},
		{"ASCII control chars", "\n\t\r\x00\x1F\x7F", true},
		{"Non-ASCII Latin", "café", false},
		{"Portuguese", "olá mundo", false},
		{"Japanese", "こんにちは", false},
		{"Chinese", "世界", false},
		{"Emoji", "😊", false},
		{"Mixed ASCII and non-ASCII", "hello世界", false},
		{"Special UTF-8 chars", "—–''‹›«»", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsASCII(test.input)
			if result != test.expected {
				t.Errorf("IsASCII(%q) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleIsASCII() {
	fmt.Println(IsASCII("hello"))
	fmt.Println(IsASCII("olá"))
	fmt.Println(IsASCII("123ABC"))
	fmt.Println(IsASCII("世界"))
	// Output:
	// true
	// false
	// true
	// false
}

// Benchmarks
func BenchmarkIsASCII(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Short ASCII", "hello"},
		{"Long ASCII", "The quick brown fox jumps over the lazy dog. 0123456789!@#$%^&*()"},
		{"Short non-ASCII", "café"},
		{"Long non-ASCII", "こんにちは世界こんにちは世界"},
		{"Mixed content", "hello world café 123 世界"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = IsASCII(bc.input)
			}
		})
	}
}

// Testes de edge cases
func TestIsASCIIEdgeCases(t *testing.T) {
	// Caractere com valor exatamente 127 (DEL)
	if !IsASCII("\x7F") {
		t.Errorf("IsASCII(DEL) = false; want true")
	}

	// Caractere com valor exatamente 128 (primeiro não-ASCII)
	if IsASCII("\x80") {
		t.Errorf("IsASCII(\\x80) = true; want false")
	}

	// String longa com caracteres ASCII
	longASCII := ""
	for i := 0; i < 10000; i++ {
		longASCII += "a"
	}
	if !IsASCII(longASCII) {
		t.Errorf("IsASCII(longASCII) = false; want true")
	}

	// String longa com um único caractere não-ASCII no final
	longMixed := longASCII + "é"
	if IsASCII(longMixed) {
		t.Errorf("IsASCII(longMixed) = true; want false")
	}
}
