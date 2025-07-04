package strutl

import (
	"fmt"
	"testing"
)

func TestToCamel(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Underscore separator", "test_case", "TestCase"},
		{"Dot separator", "test.case", "TestCase"},
		{"Single word", "test", "Test"},
		{"Already camel case", "TestCase", "TestCase"},
		{"Whitespace", " test  case ", "TestCase"},
		{"Empty string", "", ""},
		{"Multiple underscores", "many_many_words", "ManyManyWords"},
		{"Mixed separators", "AnyKind of_string", "AnyKindOfString"},
		{"Hyphen separator", "odd-fix", "OddFix"},
		{"With numbers", "numbers2And55with000", "Numbers2And55With000"},
		{"Default acronym", "ID", "Id"},
		{"Constant case", "CONSTANT_CASE", "ConstantCase"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToCamel(tc.input)
			if result != tc.expected {
				t.Errorf("ToCamel(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToLowerCamel(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Hyphen separator", "foo-bar", "fooBar"},
		{"Already camel case", "TestCase", "testCase"},
		{"Empty string", "", ""},
		{"Mixed separators", "AnyKind of_string", "anyKindOfString"},
		{"Multiple separators", "AnyKind.of-string", "anyKindOfString"},
		{"Default acronym", "ID", "id"},
		{"Whitespace", "some string", "someString"},
		{"Leading whitespace", " some string", "someString"},
		{"Constant case", "CONSTANT_CASE", "constantCase"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToLowerCamel(tc.input)
			if result != tc.expected {
				t.Errorf("ToLowerCamel(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCustomAcronyms(t *testing.T) {
	tests := []struct {
		name          string
		acronymKey    string
		acronymValue  string
		expectedCamel string
		expectedLower string
	}{
		{
			name:          "API Custom Acronym",
			acronymKey:    "API",
			acronymValue:  "api",
			expectedCamel: "Api",
			expectedLower: "api",
		},
		{
			name:          "ABCDACME Custom Acroynm",
			acronymKey:    "ABCDACME",
			acronymValue:  "AbcdAcme",
			expectedCamel: "AbcdAcme",
			expectedLower: "abcdAcme",
		},
		{
			name:          "PostgreSQL Custom Acronym",
			acronymKey:    "PostgreSQL",
			acronymValue:  "PostgreSQL",
			expectedCamel: "PostgreSQL",
			expectedLower: "postgreSQL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ConfigureAcronym(test.acronymKey, test.acronymValue)

			// Test ToCamel
			camelResult := ToCamel(test.acronymKey)
			if camelResult != test.expectedCamel {
				t.Errorf("ToCamel(%q) = %q; want %q", test.acronymKey, camelResult, test.expectedCamel)
			}

			// Test ToLowerCamel
			lowerResult := ToLowerCamel(test.acronymKey)
			if lowerResult != test.expectedLower {
				t.Errorf("ToLowerCamel(%q) = %q; want %q", test.acronymKey, lowerResult, test.expectedLower)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Camel case", "TestCase", "test_case"},
		{"Spaces", "Test Case", "test_case"},
		{"Already snake case", "test_case", "test_case"},
		{"Hyphen case", "test-case", "test_case"},
		{"Mixed separators", "test.case-example", "test_case_example"},
		{"Numbers", "test2Case", "test2_case"},
		{"Acronyms", "HTTPRequest", "http_request"},
		{"Empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToSnake(tc.input)
			if result != tc.expected {
				t.Errorf("ToSnake(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToScreamingSnake(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Camel case", "TestCase", "TEST_CASE"},
		{"Spaces", "Test Case", "TEST_CASE"},
		{"Already snake case", "test_case", "TEST_CASE"},
		{"Already screaming snake", "TEST_CASE", "TEST_CASE"},
		{"Hyphen case", "test-case", "TEST_CASE"},
		{"Mixed case", "testCase", "TEST_CASE"},
		{"Numbers", "test2Case", "TEST2_CASE"},
		{"Empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToScreamingSnake(tc.input)
			if result != tc.expected {
				t.Errorf("ToScreamingSnake(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Camel case", "TestCase", "test-case"},
		{"Spaces", "Test Case", "test-case"},
		{"Snake case", "test_case", "test-case"},
		{"Already kebab case", "test-case", "test-case"},
		{"Mixed separators", "test.case_example", "test-case-example"},
		{"Numbers", "test2Case", "test2-case"},
		{"Acronyms", "HTTPRequest", "http-request"},
		{"Empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToKebab(tc.input)
			if result != tc.expected {
				t.Errorf("ToKebab(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToScreamingKebab(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Camel case", "TestCase", "TEST-CASE"},
		{"Spaces", "Test Case", "TEST-CASE"},
		{"Snake case", "test_case", "TEST-CASE"},
		{"Kebab case", "test-case", "TEST-CASE"},
		{"Already screaming kebab", "TEST-CASE", "TEST-CASE"},
		{"Mixed case", "testCase", "TEST-CASE"},
		{"Numbers", "test2Case", "TEST2-CASE"},
		{"Empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToScreamingKebab(tc.input)
			if result != tc.expected {
				t.Errorf("ToScreamingKebab(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// Exemplos para a documentação
func ExampleToCamel() {
	fmt.Println(ToCamel("test_case"))
	fmt.Println(ToCamel("test.case"))
	fmt.Println(ToCamel("TestCase"))
	fmt.Println(ToCamel(" test  case "))
	// Output:
	// TestCase
	// TestCase
	// TestCase
	// TestCase
}

func ExampleToLowerCamel() {
	fmt.Println(ToLowerCamel("TestCase"))
	fmt.Println(ToLowerCamel("test_case"))
	fmt.Println(ToLowerCamel("test-case"))
	fmt.Println(ToLowerCamel("TEST_CASE"))
	// Output:
	// testCase
	// testCase
	// testCase
	// testCase
}

func ExampleToSnake() {
	fmt.Println(ToSnake("TestCase"))
	fmt.Println(ToSnake("testCase"))
	fmt.Println(ToSnake("test-case"))
	// Output:
	// test_case
	// test_case
	// test_case
}

func ExampleToKebab() {
	fmt.Println(ToKebab("TestCase"))
	fmt.Println(ToKebab("testCase"))
	fmt.Println(ToKebab("test_case"))
	// Output:
	// test-case
	// test-case
	// test-case
}

// Benchmarks para medir performance

var benchString = "AnyKind_of-string.with multiple_Separators"

func BenchmarkToCamel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToCamel(benchString)
	}
}

func BenchmarkToLowerCamel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToLowerCamel(benchString)
	}
}

func BenchmarkToSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToSnake(benchString)
	}
}

func BenchmarkToScreamingSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToScreamingSnake(benchString)
	}
}

func BenchmarkToKebab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToKebab(benchString)
	}
}

func BenchmarkToScreamingKebab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToScreamingKebab(benchString)
	}
}
