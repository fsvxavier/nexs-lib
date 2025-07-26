package strutil

import (
	"strings"
	"testing"
)

func TestRandom(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
		wantLen int
		wantErr bool
	}{
		{
			name:    "alphanumeric 8 chars",
			length:  8,
			charset: CharsetAlphanumeric,
			wantLen: 8,
			wantErr: false,
		},
		{
			name:    "numeric 10 chars",
			length:  10,
			charset: CharsetNumeric,
			wantLen: 10,
			wantErr: false,
		},
		{
			name:    "hex 16 chars",
			length:  16,
			charset: CharsetHex,
			wantLen: 16,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			charset: CharsetAlphanumeric,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "negative length",
			length:  -5,
			charset: CharsetAlphanumeric,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "empty charset",
			length:  5,
			charset: "",
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Random(tt.length, tt.charset)

			if len(got) != tt.wantLen {
				t.Errorf("Random() length = %v, want %v", len(got), tt.wantLen)
			}

			// Verify all characters are from the charset
			if tt.wantLen > 0 {
				for _, char := range got {
					if !strings.ContainsRune(tt.charset, char) {
						t.Errorf("Random() contains invalid character %c", char)
					}
				}
			}
		})
	}
}

func TestRandomUniqueness(t *testing.T) {
	// Test that random strings are actually random
	length := 10
	charset := CharsetAlphanumeric
	generated := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result := Random(length, charset)
		if generated[result] {
			t.Errorf("Random() generated duplicate string: %s", result)
		}
		generated[result] = true
	}
}

func TestRemoveAccents(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "empty string",
			text: "",
			want: "",
		},
		{
			name: "no accents",
			text: "hello world",
			want: "hello world",
		},
		{
			name: "french accents",
			text: "café naïve résumé",
			want: "cafe naive resume",
		},
		{
			name: "spanish accents",
			text: "niño señor",
			want: "nino senor",
		},
		{
			name: "german umlauts",
			text: "Müller Größe",
			want: "Muller Grosse",
		},
		{
			name: "mixed case accents",
			text: "CAFÉ café",
			want: "CAFE cafe",
		},
		{
			name: "portuguese accents",
			text: "São Paulo coração",
			want: "Sao Paulo coracao",
		},
		{
			name: "cedilla",
			text: "français Ça va",
			want: "francais Ca va",
		},
		{
			name: "multiple accent types",
			text: "àáâãäåèéêëìíîïòóôõöùúûü",
			want: "aaaaaaeeeeiiiiooooouuuu",
		},
		{
			name: "preserve non-accented special chars",
			text: "café@test.com",
			want: "cafe@test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveAccents(tt.text); got != tt.want {
				t.Errorf("RemoveAccents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "empty string",
			text: "",
			want: "",
		},
		{
			name: "simple text",
			text: "Hello World",
			want: "hello-world",
		},
		{
			name: "with special characters",
			text: "Hello, World!",
			want: "hello-world",
		},
		{
			name: "with accents",
			text: "Café & Bar",
			want: "cafe-bar",
		},
		{
			name: "multiple spaces",
			text: "hello   world   test",
			want: "hello-world-test",
		},
		{
			name: "leading and trailing spaces",
			text: "  hello world  ",
			want: "hello-world",
		},
		{
			name: "numbers and letters",
			text: "Product 123 v2.0",
			want: "product-123-v2-0",
		},
		{
			name: "only special characters",
			text: "!@#$%^&*()",
			want: "",
		},
		{
			name: "mixed languages",
			text: "Hello 世界",
			want: "hello",
		},
		{
			name: "consecutive separators",
			text: "hello---world___test",
			want: "hello-world-test",
		},
		{
			name: "url-like string",
			text: "example.com/path?param=value",
			want: "example-com-path-param-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.text); got != tt.want {
				t.Errorf("Slugify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "empty string",
			text: "",
			want: nil,
		},
		{
			name: "simple words",
			text: "hello world test",
			want: []string{"hello", "world", "test"},
		},
		{
			name: "punctuation separated",
			text: "hello,world;test",
			want: []string{"hello", "world", "test"},
		},
		{
			name: "contractions",
			text: "don't can't won't",
			want: []string{"don't", "can't", "won't"},
		},
		{
			name: "mixed separators",
			text: "word1!word2@word3#word4",
			want: []string{"word1", "word2", "word3", "word4"},
		},
		{
			name: "numbers",
			text: "test123 456test word789word",
			want: []string{"test123", "456test", "word789word"},
		},
		{
			name: "unicode letters",
			text: "hello 世界 test",
			want: []string{"hello", "test"},
		},
		{
			name: "consecutive separators",
			text: "hello!!!world",
			want: []string{"hello", "world"},
		},
		{
			name: "leading and trailing separators",
			text: "!!!hello world!!!",
			want: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Words(tt.text)
			if len(got) != len(tt.want) {
				t.Errorf("Words() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, word := range got {
				if word != tt.want[i] {
					t.Errorf("Words()[%d] = %v, want %v", i, word, tt.want[i])
				}
			}
		})
	}
}

func TestReplaceMultiple(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		replacements map[string]string
		want         string
	}{
		{
			name: "empty text",
			text: "",
			replacements: map[string]string{
				"hello": "hi",
			},
			want: "",
		},
		{
			name:         "no replacements",
			text:         "hello world",
			replacements: map[string]string{},
			want:         "hello world",
		},
		{
			name: "simple replacements",
			text: "hello world",
			replacements: map[string]string{
				"hello": "hi",
				"world": "earth",
			},
			want: "hi earth",
		},
		{
			name: "overlapping replacements",
			text: "hello hello world",
			replacements: map[string]string{
				"hello": "hi",
				"world": "earth",
			},
			want: "hi hi earth",
		},
		{
			name: "case sensitive",
			text: "Hello world",
			replacements: map[string]string{
				"hello": "hi",
				"world": "earth",
			},
			want: "Hello earth",
		},
		{
			name: "empty replacement",
			text: "remove this word",
			replacements: map[string]string{
				" this": "",
			},
			want: "remove word",
		},
		{
			name: "no matches",
			text: "hello world",
			replacements: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
			want: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceMultiple(tt.text, tt.replacements); got != tt.want {
				t.Errorf("ReplaceMultiple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "empty string",
			text: "",
			want: "",
		},
		{
			name: "single spaces",
			text: "hello world",
			want: "hello world",
		},
		{
			name: "multiple spaces",
			text: "hello    world",
			want: "hello world",
		},
		{
			name: "leading spaces",
			text: "   hello world",
			want: "hello world",
		},
		{
			name: "trailing spaces",
			text: "hello world   ",
			want: "hello world",
		},
		{
			name: "tabs and newlines",
			text: "hello\t\tworld\n\ntest",
			want: "hello world test",
		},
		{
			name: "mixed whitespace",
			text: "  hello\t\n  world  \n\n  test  ",
			want: "hello world test",
		},
		{
			name: "only whitespace",
			text: "   \t\n  ",
			want: "",
		},
		{
			name: "unicode spaces",
			text: "hello\u00A0\u2000world",
			want: "hello world",
		},
		{
			name: "preserve single spaces",
			text: "a b c",
			want: "a b c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.text); got != tt.want {
				t.Errorf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractWords(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		minLength     int
		caseSensitive bool
		want          []string
	}{
		{
			name:          "empty string",
			text:          "",
			minLength:     1,
			caseSensitive: true,
			want:          nil,
		},
		{
			name:          "filter by length",
			text:          "a bb ccc dddd",
			minLength:     3,
			caseSensitive: true,
			want:          []string{"ccc", "dddd"},
		},
		{
			name:          "case insensitive",
			text:          "Hello WORLD Test",
			minLength:     1,
			caseSensitive: false,
			want:          []string{"hello", "world", "test"},
		},
		{
			name:          "case sensitive",
			text:          "Hello WORLD Test",
			minLength:     1,
			caseSensitive: true,
			want:          []string{"Hello", "WORLD", "Test"},
		},
		{
			name:          "zero min length",
			text:          "a bb ccc",
			minLength:     0,
			caseSensitive: true,
			want:          []string{"a", "bb", "ccc"},
		},
		{
			name:          "high min length",
			text:          "short words only",
			minLength:     10,
			caseSensitive: true,
			want:          []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractWords(tt.text, tt.minLength, tt.caseSensitive)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractWords() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, word := range got {
				if word != tt.want[i] {
					t.Errorf("ExtractWords()[%d] = %v, want %v", i, word, tt.want[i])
				}
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "empty string",
			text: "",
			want: 0,
		},
		{
			name: "single word",
			text: "hello",
			want: 1,
		},
		{
			name: "multiple words",
			text: "hello world test",
			want: 3,
		},
		{
			name: "punctuation separated",
			text: "one,two;three!four",
			want: 4,
		},
		{
			name: "with contractions",
			text: "don't count as two",
			want: 4,
		},
		{
			name: "only separators",
			text: "!@#$%^&*()",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountWords(tt.text); got != tt.want {
				t.Errorf("CountWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "empty string",
			text: "",
			want: 0,
		},
		{
			name: "single line",
			text: "hello world",
			want: 1,
		},
		{
			name: "multiple lines LF",
			text: "line1\nline2\nline3",
			want: 3,
		},
		{
			name: "multiple lines CRLF",
			text: "line1\r\nline2\r\nline3",
			want: 3,
		},
		{
			name: "multiple lines CR",
			text: "line1\rline2\rline3",
			want: 3,
		},
		{
			name: "mixed line endings",
			text: "line1\nline2\r\nline3\r",
			want: 4,
		},
		{
			name: "empty lines",
			text: "line1\n\nline3",
			want: 3,
		},
		{
			name: "trailing newline",
			text: "line1\nline2\n",
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountLines(tt.text); got != tt.want {
				t.Errorf("CountLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateWords(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		maxWords    int
		addEllipsis bool
		want        string
	}{
		{
			name:        "empty string",
			text:        "",
			maxWords:    3,
			addEllipsis: true,
			want:        "",
		},
		{
			name:        "fewer words than max",
			text:        "one two",
			maxWords:    5,
			addEllipsis: true,
			want:        "one two",
		},
		{
			name:        "exact word count",
			text:        "one two three",
			maxWords:    3,
			addEllipsis: true,
			want:        "one two three",
		},
		{
			name:        "truncate with ellipsis",
			text:        "one two three four five",
			maxWords:    3,
			addEllipsis: true,
			want:        "one two three...",
		},
		{
			name:        "truncate without ellipsis",
			text:        "one two three four five",
			maxWords:    3,
			addEllipsis: false,
			want:        "one two three",
		},
		{
			name:        "zero max words",
			text:        "one two three",
			maxWords:    0,
			addEllipsis: true,
			want:        "",
		},
		{
			name:        "negative max words",
			text:        "one two three",
			maxWords:    -1,
			addEllipsis: true,
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateWords(tt.text, tt.maxWords, tt.addEllipsis); got != tt.want {
				t.Errorf("TruncateWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "empty filename",
			filename: "",
			want:     "",
		},
		{
			name:     "valid filename",
			filename: "document.txt",
			want:     "document.txt",
		},
		{
			name:     "with invalid characters",
			filename: "file/name?.txt",
			want:     "file_name_.txt",
		},
		{
			name:     "windows path",
			filename: "C:\\Users\\file.txt",
			want:     "C__Users_file.txt",
		},
		{
			name:     "all invalid characters",
			filename: "/<>:\"\\|?*",
			want:     "_________",
		},
		{
			name:     "mixed valid and invalid",
			filename: "report_2023<final>.pdf",
			want:     "report_2023_final_.pdf",
		},
		{
			name:     "unicode characters",
			filename: "café_file.txt",
			want:     "café_file.txt",
		},
		{
			name:     "control characters",
			filename: "file\x00\x1f.txt",
			want:     "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanFilename(tt.filename); got != tt.want {
				t.Errorf("CleanFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name          string
		s             string
		prefixes      []string
		caseSensitive bool
		want          bool
	}{
		{
			name:          "empty string",
			s:             "",
			prefixes:      []string{"test"},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "empty prefixes",
			s:             "hello",
			prefixes:      []string{},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "case sensitive match",
			s:             "Hello World",
			prefixes:      []string{"Hello", "Hi"},
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "case sensitive no match",
			s:             "Hello World",
			prefixes:      []string{"hello", "hi"},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "case insensitive match",
			s:             "Hello World",
			prefixes:      []string{"hello", "hi"},
			caseSensitive: false,
			want:          true,
		},
		{
			name:          "multiple prefixes",
			s:             "test string",
			prefixes:      []string{"hello", "test", "world"},
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "no match",
			s:             "hello world",
			prefixes:      []string{"test", "foo", "bar"},
			caseSensitive: true,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPrefix(tt.s, tt.prefixes, tt.caseSensitive); got != tt.want {
				t.Errorf("HasPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSuffix(t *testing.T) {
	tests := []struct {
		name          string
		s             string
		suffixes      []string
		caseSensitive bool
		want          bool
	}{
		{
			name:          "empty string",
			s:             "",
			suffixes:      []string{"test"},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "empty suffixes",
			s:             "hello",
			suffixes:      []string{},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "case sensitive match",
			s:             "Hello World",
			suffixes:      []string{"World", "Earth"},
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "case sensitive no match",
			s:             "Hello World",
			suffixes:      []string{"world", "earth"},
			caseSensitive: true,
			want:          false,
		},
		{
			name:          "case insensitive match",
			s:             "Hello World",
			suffixes:      []string{"world", "earth"},
			caseSensitive: false,
			want:          true,
		},
		{
			name:          "multiple suffixes",
			s:             "test.txt",
			suffixes:      []string{".pdf", ".txt", ".doc"},
			caseSensitive: true,
			want:          true,
		},
		{
			name:          "no match",
			s:             "hello world",
			suffixes:      []string{"test", "foo", "bar"},
			caseSensitive: true,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasSuffix(tt.s, tt.suffixes, tt.caseSensitive); got != tt.want {
				t.Errorf("HasSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "empty string",
			s:    "",
			want: true,
		},
		{
			name: "valid ASCII",
			s:    "hello world",
			want: true,
		},
		{
			name: "valid UTF-8",
			s:    "hello 世界 🌍",
			want: true,
		},
		{
			name: "invalid UTF-8",
			s:    string([]byte{0xff, 0xfe, 0xfd}),
			want: false,
		},
		{
			name: "partial UTF-8 sequence",
			s:    string([]byte{0xc2}),
			want: false,
		},
		{
			name: "overlong encoding",
			s:    string([]byte{0xc0, 0x80}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUTF8(tt.s); got != tt.want {
				t.Errorf("IsValidUTF8() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests for text processing functions
func BenchmarkRandom(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Random(32, CharsetAlphanumeric)
	}
}

func BenchmarkRemoveAccents(b *testing.B) {
	text := "café naïve résumé señor müller größe"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RemoveAccents(text)
	}
}

func BenchmarkSlugify(b *testing.B) {
	text := "Hello World! This is a test string with Accénts & Special Characters."
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Slugify(text)
	}
}

func BenchmarkWords(b *testing.B) {
	text := strings.Repeat("word1 word2, word3; word4! ", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Words(text)
	}
}

func BenchmarkNormalize(b *testing.B) {
	text := strings.Repeat("hello    world\t\n  test  ", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Normalize(text)
	}
}

// Edge case tests
func TestTextProcessorEdgeCases(t *testing.T) {
	t.Run("very long strings", func(t *testing.T) {
		longText := strings.Repeat("word ", 100000)

		// Should not panic or cause excessive memory usage
		words := Words(longText)
		if len(words) == 0 {
			t.Error("Long text should produce words")
		}

		normalized := Normalize(longText)
		if normalized == "" {
			t.Error("Normalization should not result in empty string")
		}
	})

	t.Run("unicode edge cases", func(t *testing.T) {
		unicodeTexts := []string{
			"🌍🌎🌏",     // Emoji
			"中文测试",    // Chinese
			"العربية", // Arabic
			"हिन्दी",  // Hindi
		}

		for _, text := range unicodeTexts {
			// Should handle Unicode gracefully
			Words(text)
			Slugify(text)
			CleanFilename(text)
			IsValidUTF8(text)
		}
	})

	t.Run("performance with repeated operations", func(t *testing.T) {
		text := "Hello World! Test String with Various Characters 123 @#$"

		// Multiple operations should be efficient
		for i := 0; i < 1000; i++ {
			result := Slugify(RemoveAccents(Normalize(text)))
			if result == "" {
				t.Error("Chained operations should not result in empty string")
				break
			}
		}
	})
}
