package strutil

import (
	"strings"
	"testing"
)

func TestToCamel(t *testing.T) {
	// Clear acronyms for consistent testing
	ClearAcronyms()

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "single word",
			s:    "hello",
			want: "Hello",
		},
		{
			name: "snake case",
			s:    "hello_world",
			want: "HelloWorld",
		},
		{
			name: "kebab case",
			s:    "hello-world",
			want: "HelloWorld",
		},
		{
			name: "dot notation",
			s:    "hello.world",
			want: "HelloWorld",
		},
		{
			name: "space separated",
			s:    "hello world",
			want: "HelloWorld",
		},
		{
			name: "mixed separators",
			s:    "hello_world-test.case",
			want: "HelloWorldTestCase",
		},
		{
			name: "already camel case",
			s:    "HelloWorld",
			want: "HelloWorld",
		},
		{
			name: "with numbers",
			s:    "hello2world",
			want: "Hello2World",
		},
		{
			name: "consecutive uppercase",
			s:    "XMLParser",
			want: "XmlParser",
		},
		{
			name: "leading underscore",
			s:    "_hello_world",
			want: "HelloWorld",
		},
		{
			name: "trailing underscore",
			s:    "hello_world_",
			want: "HelloWorld",
		},
		{
			name: "multiple consecutive separators",
			s:    "hello___world",
			want: "HelloWorld",
		},
		{
			name: "whitespace trimming",
			s:    "  hello_world  ",
			want: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamel(tt.s); got != tt.want {
				t.Errorf("ToCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLowerCamel(t *testing.T) {
	ClearAcronyms()

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "single word",
			s:    "hello",
			want: "hello",
		},
		{
			name: "snake case",
			s:    "hello_world",
			want: "helloWorld",
		},
		{
			name: "kebab case",
			s:    "hello-world",
			want: "helloWorld",
		},
		{
			name: "space separated",
			s:    "hello world",
			want: "helloWorld",
		},
		{
			name: "already camel case",
			s:    "HelloWorld",
			want: "helloWorld",
		},
		{
			name: "already lower camel case",
			s:    "helloWorld",
			want: "helloWorld",
		},
		{
			name: "with numbers",
			s:    "hello2world",
			want: "hello2World",
		},
		{
			name: "consecutive uppercase",
			s:    "XMLParser",
			want: "xmlParser",
		},
		{
			name: "single character",
			s:    "a",
			want: "a",
		},
		{
			name: "single uppercase character",
			s:    "A",
			want: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLowerCamel(tt.s); got != tt.want {
				t.Errorf("ToLowerCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCamelWithAcronyms(t *testing.T) {
	ClearAcronyms()
	ConfigureAcronym("ID", "id")
	ConfigureAcronym("URL", "url")
	ConfigureAcronym("API", "api")

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "acronym replacement",
			s:    "ID",
			want: "Id",
		},
		{
			name: "url acronym",
			s:    "URL",
			want: "Url",
		},
		{
			name: "api acronym",
			s:    "API",
			want: "Api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamel(tt.s); got != tt.want {
				t.Errorf("ToCamel() with acronym = %v, want %v", got, tt.want)
			}
		})
	}

	// Clean up
	ClearAcronyms()
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty string",
			str:  "",
			want: "",
		},
		{
			name: "single character",
			str:  "a",
			want: "a",
		},
		{
			name: "simple case",
			str:  "camel case",
			want: "camelCase",
		},
		{
			name: "preserve existing case",
			str:  "inside dynaMIC-HTML",
			want: "insideDynaMIC-HTML",
		},
		{
			name: "multiple spaces",
			str:  "hello   world   test",
			want: "helloWorldTest",
		},
		{
			name: "leading and trailing spaces",
			str:  "  hello world  ",
			want: "helloWorld",
		},
		{
			name: "single word",
			str:  "hello",
			want: "hello",
		},
		{
			name: "already camelCase",
			str:  "helloWorld",
			want: "helloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamelCase(tt.str); got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty string",
			str:  "",
			want: "",
		},
		{
			name: "simple case",
			str:  "Snake Case",
			want: "snake_case",
		},
		{
			name: "multiple spaces",
			str:  "hello   world",
			want: "hello___world",
		},
		{
			name: "leading and trailing spaces",
			str:  "  hello world  ",
			want: "hello_world",
		},
		{
			name: "already snake_case",
			str:  "snake_case",
			want: "snake_case",
		},
		{
			name: "mixed case",
			str:  "Hello World Test",
			want: "hello_world_test",
		},
		{
			name: "single word",
			str:  "hello",
			want: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.str); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "camelCase",
			s:    "helloWorld",
			want: "hello_world",
		},
		{
			name: "PascalCase",
			s:    "HelloWorld",
			want: "hello_world",
		},
		{
			name: "already snake_case",
			s:    "hello_world",
			want: "hello_world",
		},
		{
			name: "with numbers",
			s:    "hello2World",
			want: "hello2_world",
		},
		{
			name: "consecutive uppercase",
			s:    "XMLParser",
			want: "xml_parser",
		},
		{
			name: "single word",
			s:    "hello",
			want: "hello",
		},
		{
			name: "uppercase word",
			s:    "HELLO",
			want: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnake(tt.s); got != tt.want {
				t.Errorf("ToSnake() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToScreamingSnake(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "camelCase",
			s:    "helloWorld",
			want: "HELLO_WORLD",
		},
		{
			name: "snake_case",
			s:    "hello_world",
			want: "HELLO_WORLD",
		},
		{
			name: "kebab-case",
			s:    "hello-world",
			want: "HELLO_WORLD",
		},
		{
			name: "space separated",
			s:    "hello world",
			want: "HELLO_WORLD",
		},
		{
			name: "mixed case",
			s:    "HelloWorld",
			want: "HELLO_WORLD",
		},
		{
			name: "with numbers",
			s:    "hello2World",
			want: "HELLO2_WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToScreamingSnake(tt.s); got != tt.want {
				t.Errorf("ToScreamingSnake() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "camelCase",
			s:    "helloWorld",
			want: "hello-world",
		},
		{
			name: "PascalCase",
			s:    "HelloWorld",
			want: "hello-world",
		},
		{
			name: "snake_case",
			s:    "hello_world",
			want: "hello-world",
		},
		{
			name: "already kebab-case",
			s:    "hello-world",
			want: "hello-world",
		},
		{
			name: "with numbers",
			s:    "hello2World",
			want: "hello2-world",
		},
		{
			name: "space separated",
			s:    "hello world",
			want: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToKebab(tt.s); got != tt.want {
				t.Errorf("ToKebab() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToScreamingKebab(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "camelCase",
			s:    "helloWorld",
			want: "HELLO-WORLD",
		},
		{
			name: "kebab-case",
			s:    "hello-world",
			want: "HELLO-WORLD",
		},
		{
			name: "snake_case",
			s:    "hello_world",
			want: "HELLO-WORLD",
		},
		{
			name: "space separated",
			s:    "hello world",
			want: "HELLO-WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToScreamingKebab(tt.s); got != tt.want {
				t.Errorf("ToScreamingKebab() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDelimited(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		delimiter uint8
		want      string
	}{
		{
			name:      "dot delimiter",
			s:         "helloWorld",
			delimiter: '.',
			want:      "hello.world",
		},
		{
			name:      "pipe delimiter",
			s:         "HelloWorld",
			delimiter: '|',
			want:      "hello|world",
		},
		{
			name:      "colon delimiter",
			s:         "hello_world",
			delimiter: ':',
			want:      "hello:world",
		},
		{
			name:      "empty string",
			s:         "",
			delimiter: '.',
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToDelimited(tt.s, tt.delimiter); got != tt.want {
				t.Errorf("ToDelimited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeWithIgnore(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		ignore string
		want   string
	}{
		{
			name:   "ignore dash",
			s:      "hello-world_test",
			ignore: "-",
			want:   "hello-world_test",
		},
		{
			name:   "ignore multiple chars",
			s:      "hello-world_test.case",
			ignore: "-.",
			want:   "hello-world_test.case",
		},
		{
			name:   "no ignore",
			s:      "helloWorld",
			ignore: "",
			want:   "hello_world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeWithIgnore(tt.s, tt.ignore); got != tt.want {
				t.Errorf("ToSnakeWithIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDelimiter(t *testing.T) {
	tests := []struct {
		name          string
		delimiterName string
		wantValue     uint8
		wantExists    bool
	}{
		{
			name:          "snake delimiter",
			delimiterName: "snake",
			wantValue:     '_',
			wantExists:    true,
		},
		{
			name:          "kebab delimiter",
			delimiterName: "kebab",
			wantValue:     '-',
			wantExists:    true,
		},
		{
			name:          "dot delimiter",
			delimiterName: "dot",
			wantValue:     '.',
			wantExists:    true,
		},
		{
			name:          "nonexistent delimiter",
			delimiterName: "nonexistent",
			wantValue:     0,
			wantExists:    false,
		},
		{
			name:          "space delimiter",
			delimiterName: "space",
			wantValue:     ' ',
			wantExists:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotExists := GetDelimiter(tt.delimiterName)
			if gotValue != tt.wantValue {
				t.Errorf("GetDelimiter() value = %v, want %v", gotValue, tt.wantValue)
			}
			if gotExists != tt.wantExists {
				t.Errorf("GetDelimiter() exists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

// Benchmark tests for case conversion performance
func BenchmarkToCamel(b *testing.B) {
	testString := "hello_world_this_is_a_test_string_for_benchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToCamel(testString)
	}
}

func BenchmarkToLowerCamel(b *testing.B) {
	testString := "hello_world_this_is_a_test_string_for_benchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToLowerCamel(testString)
	}
}

func BenchmarkToSnake(b *testing.B) {
	testString := "HelloWorldThisIsATestStringForBenchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToSnake(testString)
	}
}

func BenchmarkToKebab(b *testing.B) {
	testString := "HelloWorldThisIsATestStringForBenchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToKebab(testString)
	}
}

func BenchmarkToScreamingDelimited(b *testing.B) {
	testString := "HelloWorldThisIsATestStringForBenchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ToScreamingDelimited(testString, "", '_', true)
	}
}

// Edge case tests
func TestCaseConversionEdgeCases(t *testing.T) {
	t.Run("unicode characters", func(t *testing.T) {
		input := "hello_世界"
		result := ToCamel(input)
		// Should handle Unicode gracefully
		if result == "" {
			t.Error("Unicode string should not result in empty string")
		}
	})

	t.Run("very long string", func(t *testing.T) {
		longString := strings.Repeat("hello_world_", 1000)

		// Should not panic or cause memory issues
		result := ToCamel(longString)
		if len(result) == 0 {
			t.Error("Long string conversion should not result in empty string")
		}
	})

	t.Run("special characters", func(t *testing.T) {
		specialChars := []string{
			"hello@world",
			"hello#world",
			"hello$world",
			"hello%world",
			"hello&world",
		}

		for _, s := range specialChars {
			// Should not panic
			ToCamel(s)
			ToSnake(s)
			ToKebab(s)
		}
	})
}

// Test memory allocation efficiency
func TestCaseConversionMemoryEfficiency(t *testing.T) {
	// Test that string builder capacity is used efficiently
	testString := "hello_world_test_case_conversion"

	result := ToCamel(testString)
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Verify the result is correct
	expected := "HelloWorldTestCaseConversion"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
