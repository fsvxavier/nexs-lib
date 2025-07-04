package strutl

import (
	"fmt"
	"testing"
)

var randCounter = 0

type dummyRandReader struct {
}

func (r *dummyRandReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(randCounter)
		randCounter++
	}
	return len(p), err
}

func TestRandom(t *testing.T) {
	oldReader := randReader
	randReader = &dummyRandReader{}

	tests := []struct {
		name     string
		input    string
		expected string
		length   int
	}{
		{"5 chars from 10", "abcdefghij", "abcde", 5},
		{"Single char", "abcdefghij", "a", 1},
		{"Repeating from small set", "abc", "abcabc", 6},
		{"Empty set", "", "", 5},
		{"Same char set", "aaa", "aaaaa", 5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			randCounter = 0
			output, err := Random(test.input, test.length)

			if err != nil {
				t.Errorf("Random(%q, %d) retornou erro: %v", test.input, test.length, err)
				return
			}

			if output != test.expected {
				t.Errorf("Random(%q, %d) = %q; want %q",
					test.input, test.length, output, test.expected)
			}
		})
	}

	randReader = oldReader
}

func TestRandom_Real(t *testing.T) {
	// Este teste não é determinístico, então verificamos apenas comprimento e caracteres usados
	tests := []struct {
		name   string
		strSet string
		length int
	}{
		{"Letras", "abcdefghijklmnopqrstuvwxyz", 10},
		{"Alfanuméricos", "abcdefghijklmnopqrstuvwxyz0123456789", 15},
		{"Especiais", "!@#$%^&*()_+-=", 8},
		{"Zero length", "abc", 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Random(test.strSet, test.length)

			if err != nil {
				t.Errorf("Random(%q, %d) retornou erro: %v", test.strSet, test.length, err)
				return
			}

			// Verificar comprimento
			if len(result) != test.length {
				t.Errorf("Random(%q, %d) retornou string com comprimento %d, esperava %d",
					test.strSet, test.length, len(result), test.length)
			}

			// Verificar se todos os caracteres são do conjunto
			if test.length > 0 && test.strSet != "" {
				setMap := make(map[rune]bool)
				for _, r := range test.strSet {
					setMap[r] = true
				}

				for _, r := range result {
					if !setMap[r] {
						t.Errorf("Random(%q, %d) retornou string com caractere %q que não está no conjunto",
							test.strSet, test.length, r)
						break
					}
				}
			}
		})
	}
}

// Exemplo para documentação
func ExampleRandom() {
	// Nota: Devido à natureza aleatória, não podemos ter uma saída determinística
	result, err := Random("abcdefghik", 5)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
	fmt.Printf("String gerada com comprimento %d: %s\n", len(result), result)
}

// Benchmarks
func BenchmarkRandom(b *testing.B) {
	benchCases := []struct {
		name   string
		strSet string
		length int
	}{
		{"SmallSet", "abc", 10},
		{"MediumSet", "abcdefghijklmnopqrstuvwxyz", 10},
		{"LargeSet", "abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_+=", 10},
		{"LongOutput", "abcdefghijklmnopqrstuvwxyz", 100},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Random(bc.strSet, bc.length)
			}
		})
	}
}
