package strutl

import (
	"fmt"
	"testing"
)

func TestRemoveAccents(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"No accents", "hello world", "hello world"},
		{"Simple accents", "olá", "ola"},
		{"Multiple accents", "àéêöhello", "aeeohello"},
		{"Portuguese", "ação", "acao"},
		{"German", "über", "uber"},
		{"French", "déjà vu", "deja vu"},
		{"Spanish", "año", "ano"},
		{"Mixed case", "Olá Mundo", "Ola Mundo"},
		{"With numbers and symbols", "café #1", "cafe #1"},
		{"Complex characters", "ñ Ñ á Á é É í Í ó Ó ú Ú ü Ü", "n N a A e E i I o O u U u U"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := RemoveAccents(test.input)
			if result != test.expected {
				t.Errorf("RemoveAccents(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Simple string", "hello world", "hello-world"},
		{"With accents", "São Paulo", "sao-paulo"},
		{"With trim spaces", "  a b c  ", "a-b-c"},
		{"With percentages", "100%", "100"},
		{"With special chars", "hello! @world#", "hello-world"},
		{"With multiple spaces", "hello   world", "hello-world"},
		{"With multiple special chars", "hello!@#world", "hello-world"},
		{"With numbers", "hello 123 world", "hello-123-world"},
		{"Complex example", "Título do Artigo: Como fazer café? (guia prático)", "titulo-do-artigo-como-fazer-cafe-guia-pratico"},
		{"URL path", "/blog/2020/01/01/title/", "blog-2020-01-01-title"},
		{"Email address", "user@example.com", "user-example-com"},
		{"With hyphens", "already-slugged-string", "already-slugged-string"},
		{"Multiple consecutive non-alphanumeric chars", "a---b", "a-b"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Slugify(test.input)
			if result != test.expected {
				t.Errorf("Slugify(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}

	// Teste adicional: o resultado deve ser o mesmo ao slugificar duas vezes
	for _, test := range tests {
		t.Run("Double slugify "+test.name, func(t *testing.T) {
			once := Slugify(test.input)
			twice := Slugify(once)
			if once != twice {
				t.Errorf("Slugify(Slugify(%q)) = %q; want %q", test.input, twice, once)
			}
		})
	}
}

// Exemplos para documentação
func ExampleRemoveAccents() {
	fmt.Println(RemoveAccents("olá mundo"))
	fmt.Println(RemoveAccents("São Paulo"))
	fmt.Println(RemoveAccents("über"))
	// Output:
	// ola mundo
	// Sao Paulo
	// uber
}

func ExampleSlugify() {
	fmt.Println(Slugify("Hello World!"))
	fmt.Println(Slugify("São Paulo"))
	fmt.Println(Slugify("  a b c  "))
	// Output:
	// hello-world
	// sao-paulo
	// a-b-c
}

// Benchmarks
func BenchmarkRemoveAccents(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{"No accents", "hello world this is a test string for benchmarking"},
		{"Some accents", "olá mundo isso é uma string de teste para benchmark"},
		{"Many accents", "àéêöñÑáÁéÉíÍóÓúÚüÜ áéíóúâêîôû äëïöü àèìòù ãõñ"},
	}

	for _, input := range inputs {
		b.Run(input.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = RemoveAccents(input.input)
			}
		})
	}
}

func BenchmarkSlugify(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{"Simple", "hello world"},
		{"With accents", "São Paulo - Capital"},
		{"Complex", "Título do Artigo: Como fazer café? (guia prático)"},
		{"With special chars", "user@example.com #hashtag !important"},
	}

	for _, input := range inputs {
		b.Run(input.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Slugify(input.input)
			}
		})
	}
}
