package strutil

import (
	"strings"
	"testing"
	"time"
)

// TestConfigureAcronym tests acronym configuration functionality
func TestConfigureAcronym(t *testing.T) {
	// Clear any existing acronyms
	ClearAcronyms()

	tests := []struct {
		name        string
		acronym     string
		replacement string
		wantExists  bool
	}{
		{
			name:        "valid acronym",
			acronym:     "URL",
			replacement: "url",
			wantExists:  true,
		},
		{
			name:        "empty acronym",
			acronym:     "",
			replacement: "test",
			wantExists:  false,
		},
		{
			name:        "overwrite existing",
			acronym:     "API",
			replacement: "api",
			wantExists:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigureAcronym(tt.acronym, tt.replacement)

			got, exists := GetAcronym(tt.acronym)
			if exists != tt.wantExists {
				t.Errorf("GetAcronym() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists && got != tt.replacement {
				t.Errorf("GetAcronym() = %v, want %v", got, tt.replacement)
			}
		})
	}
}

func TestGetAcronym(t *testing.T) {
	ClearAcronyms()
	ConfigureAcronym("ID", "id")
	ConfigureAcronym("HTTP", "http")

	tests := []struct {
		name       string
		acronym    string
		wantValue  string
		wantExists bool
	}{
		{
			name:       "existing acronym",
			acronym:    "ID",
			wantValue:  "id",
			wantExists: true,
		},
		{
			name:       "non-existing acronym",
			acronym:    "XYZ",
			wantValue:  "",
			wantExists: false,
		},
		{
			name:       "empty acronym",
			acronym:    "",
			wantValue:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := GetAcronym(tt.acronym)
			if got != tt.wantValue {
				t.Errorf("GetAcronym() value = %v, want %v", got, tt.wantValue)
			}
			if exists != tt.wantExists {
				t.Errorf("GetAcronym() exists = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestRemoveAcronym(t *testing.T) {
	ClearAcronyms()
	ConfigureAcronym("TEST", "test")

	// Verify it exists
	if _, exists := GetAcronym("TEST"); !exists {
		t.Fatal("Failed to configure test acronym")
	}

	// Remove it
	RemoveAcronym("TEST")

	// Verify it's gone
	if _, exists := GetAcronym("TEST"); exists {
		t.Error("Acronym should have been removed")
	}

	// Test removing non-existent acronym (should not panic)
	RemoveAcronym("NONEXISTENT")

	// Test removing empty acronym (should not panic)
	RemoveAcronym("")
}

func TestClearAcronyms(t *testing.T) {
	// Add some acronyms
	ConfigureAcronym("A", "a")
	ConfigureAcronym("B", "b")
	ConfigureAcronym("C", "c")

	// Clear all
	ClearAcronyms()

	// Verify they're all gone
	acronyms := []string{"A", "B", "C"}
	for _, acronym := range acronyms {
		if _, exists := GetAcronym(acronym); exists {
			t.Errorf("Acronym %s should have been cleared", acronym)
		}
	}
}

func TestIsEmpty(t *testing.T) {
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
			name: "whitespace only",
			s:    "   \t\n  ",
			want: true,
		},
		{
			name: "single space",
			s:    " ",
			want: true,
		},
		{
			name: "non-empty string",
			s:    "hello",
			want: false,
		},
		{
			name: "string with content and whitespace",
			s:    "  hello  ",
			want: false,
		},
		{
			name: "single character",
			s:    "a",
			want: false,
		},
		{
			name: "unicode whitespace",
			s:    "\u00A0\u2000\u2001", // Non-breaking space and other Unicode spaces
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.s); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsASCII(t *testing.T) {
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
			name: "ASCII only",
			s:    "Hello World 123!@#",
			want: true,
		},
		{
			name: "with unicode",
			s:    "Hello 世界",
			want: false,
		},
		{
			name: "with accents",
			s:    "café",
			want: false,
		},
		{
			name: "control characters",
			s:    "hello\x1b[31m",
			want: false,
		},
		{
			name: "newline and tab",
			s:    "line1\nline2\t",
			want: true,
		},
		{
			name: "extended ASCII",
			s:    string([]byte{128, 255}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsASCII(tt.s); got != tt.want {
				t.Errorf("IsASCII() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeSubstring(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		start int
		end   int
		want  string
	}{
		{
			name:  "normal case",
			s:     "hello world",
			start: 0,
			end:   5,
			want:  "hello",
		},
		{
			name:  "empty string",
			s:     "",
			start: 0,
			end:   5,
			want:  "",
		},
		{
			name:  "start out of bounds",
			s:     "hello",
			start: 10,
			end:   15,
			want:  "",
		},
		{
			name:  "end out of bounds",
			s:     "hello",
			start: 0,
			end:   10,
			want:  "hello",
		},
		{
			name:  "negative start",
			s:     "hello",
			start: -1,
			end:   3,
			want:  "hel",
		},
		{
			name:  "negative end",
			s:     "hello",
			start: 1,
			end:   -1,
			want:  "ello",
		},
		{
			name:  "start equals end",
			s:     "hello",
			start: 2,
			end:   2,
			want:  "",
		},
		{
			name:  "start greater than end",
			s:     "hello",
			start: 3,
			end:   1,
			want:  "",
		},
		{
			name:  "unicode string",
			s:     "hello 世界",
			start: 6,
			end:   8,
			want:  "世界",
		},
		{
			name:  "emoji string",
			s:     "😀😃😄😁",
			start: 1,
			end:   3,
			want:  "😃😄",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeSubstring(tt.s, tt.start, tt.end); got != tt.want {
				t.Errorf("SafeSubstring() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
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
			name: "single character",
			s:    "a",
			want: "a",
		},
		{
			name: "simple string",
			s:    "hello",
			want: "olleh",
		},
		{
			name: "palindrome",
			s:    "radar",
			want: "radar",
		},
		{
			name: "unicode string",
			s:    "hello 世界",
			want: "界世 olleh",
		},
		{
			name: "emoji string",
			s:    "😀😃😄",
			want: "😄😃😀",
		},
		{
			name: "mixed characters",
			s:    "abc123XYZ",
			want: "ZYX321cba",
		},
		{
			name: "with spaces",
			s:    "a b c",
			want: "c b a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.s); got != tt.want {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLen verifies that our Len function works correctly with Unicode
func TestLen(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{
			name: "empty string",
			s:    "",
			want: 0,
		},
		{
			name: "ASCII string",
			s:    "hello",
			want: 5,
		},
		{
			name: "unicode string",
			s:    "hello 世界",
			want: 8,
		},
		{
			name: "emoji string",
			s:    "😀😃😄😁",
			want: 4,
		},
		{
			name: "mixed string",
			s:    "café 🌮",
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Len(tt.s); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests for performance verification
func BenchmarkIsASCII(b *testing.B) {
	testString := strings.Repeat("Hello World 123!@#", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsASCII(testString)
	}
}

func BenchmarkSafeSubstring(b *testing.B) {
	testString := strings.Repeat("Hello World Unicode 世界", 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SafeSubstring(testString, 10, 100)
	}
}

func BenchmarkReverse(b *testing.B) {
	testString := strings.Repeat("Hello World", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Reverse(testString)
	}
}

func BenchmarkGetAcronym(b *testing.B) {
	ConfigureAcronym("API", "api")
	ConfigureAcronym("URL", "url")
	ConfigureAcronym("HTTP", "http")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetAcronym("API")
	}
}

// TestConcurrency tests thread safety of acronym operations
func TestConcurrency(t *testing.T) {
	ClearAcronyms()

	// Number of goroutines
	numGoroutines := 100
	numOpsPerGoroutine := 100

	// Channel to signal completion
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOpsPerGoroutine; j++ {
				// Configure acronym
				acronym := string(rune('A' + (id+j)%26))
				replacement := strings.ToLower(acronym)
				ConfigureAcronym(acronym, replacement)

				// Get acronym
				GetAcronym(acronym)

				// Remove acronym occasionally
				if j%10 == 0 {
					RemoveAcronym(acronym)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete with timeout
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-timeout.C:
			t.Fatal("Test timed out - possible deadlock or performance issue")
		}
	}
}

// Edge case tests
func TestEdgeCases(t *testing.T) {
	t.Run("very long string", func(t *testing.T) {
		longString := strings.Repeat("a", 1000000) // 1MB string

		// Test that these don't panic or cause memory issues
		result := IsASCII(longString)
		if !result {
			t.Error("Expected long ASCII string to be detected as ASCII")
		}

		reversed := Reverse(longString)
		if len(reversed) != len(longString) {
			t.Error("Reversed string should have same length")
		}
	})

	t.Run("unicode edge cases", func(t *testing.T) {
		unicodeStrings := []string{
			"\uFEFF", // BOM
			"\u0000", // NULL
			"\u001F", // Control character
			"\u007F", // DEL
			"\u0080", // First non-ASCII
			"\uFFFF", // Valid Unicode
		}

		for _, s := range unicodeStrings {
			// Should not panic
			IsASCII(s)
			Reverse(s)
			SafeSubstring(s, 0, 1)
		}
	})
}
