package strutl

import (
	"fmt"
	"testing"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"ASCII string", "abc", "cba"},
		{"Single character", "a", "a"},
		{"Turkish characters", "çınar", "ranıç"},
		{"Mixed spaces", "    yağmur", "rumğay    "},
		{"Greek characters", "επαγγελματίες", "ςείταμλεγγαπε"},
		{"Mixed ASCII and Unicode", "hello 世界", "界世 olleh"},
		{"Emojis", "😊🚀✨", "✨🚀😊"},
		{"Numbers", "12345", "54321"},
		{"Special characters", "!@#$%^&*()", ")(*&^%$#@!"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := Reverse(test.input)
			if output != test.expected {
				t.Errorf("Reverse(%q) = %q; want %q", test.input, output, test.expected)
			}
		})
	}

	// Teste adicional: reverter duas vezes deve voltar ao valor original
	for _, test := range tests {
		t.Run("Double reverse "+test.name, func(t *testing.T) {
			output := Reverse(Reverse(test.input))
			if output != test.input {
				t.Errorf("Reverse(Reverse(%q)) = %q; want %q", test.input, output, test.input)
			}
		})
	}
}

func ExampleReverse() {
	fmt.Println(Reverse("Hello, 世界!"))
	fmt.Println(Reverse("abc123"))
	fmt.Println(Reverse("επαγγελματίες"))
	// Output:
	// !界世 ,olleH
	// 321cba
	// ςείταμλεγγαπε
}

var benchCases = []struct {
	name  string
	input string
}{
	{"ASCII", "It is not known exactly"},
	{"Unicode", "επαγγελματίες και επιχειρηματίες"},
	{"Mixed", "It is not known exactly - επαγγελματίες"},
	{"Complex", "😊 Hello, 世界! こんにちは 🚀"},
}

func BenchmarkReverse(b *testing.B) {
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = Reverse(bc.input)
			}
		})
	}
}

// BenchmarkReverseSize avalia o desempenho da função Reverse com strings de diferentes tamanhos
func BenchmarkReverseSize(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		// Cria uma string ASCII de tamanho específico
		s := make([]byte, size)
		for i := 0; i < size; i++ {
			s[i] = 'a' + byte(i%26)
		}
		asciiStr := string(s)

		b.Run(fmt.Sprintf("ASCII_%d", size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = Reverse(asciiStr)
			}
		})

		// Cria uma string com caracteres Unicode de tamanho específico
		// (usa repetição de caracteres para criar uma string mais previsível)
		unicodeRunes := []rune{'世', '界', 'こ', 'ん', 'に', 'ち', 'は'}
		runeSlice := make([]rune, size)
		for i := 0; i < size; i++ {
			runeSlice[i] = unicodeRunes[i%len(unicodeRunes)]
		}
		unicodeStr := string(runeSlice)

		b.Run(fmt.Sprintf("Unicode_%d", size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = Reverse(unicodeStr)
			}
		})
	}
}
