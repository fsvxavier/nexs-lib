package strutil

import (
	"strings"
	"testing"
)

func TestMapLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fn       func(string) string
		expected string
	}{
		{
			name:     "uppercase lines",
			input:    "hello\nworld\ntest",
			fn:       strings.ToUpper,
			expected: "HELLO\nWORLD\nTEST",
		},
		{
			name:     "trim lines",
			input:    " hello \n world \n test ",
			fn:       strings.TrimSpace,
			expected: "hello\nworld\ntest",
		},
		{
			name:     "empty string",
			input:    "",
			fn:       strings.ToUpper,
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello",
			fn:       strings.ToUpper,
			expected: "HELLO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapLines(tt.input, tt.fn)
			if result != tt.expected {
				t.Errorf("MapLines() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSplitAndMap(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator string
		fn        func(string) string
		expected  []string
	}{
		{
			name:      "uppercase parts",
			input:     "hello,world,test",
			separator: ",",
			fn:        strings.ToUpper,
			expected:  []string{"HELLO", "WORLD", "TEST"},
		},
		{
			name:      "trim parts",
			input:     " hello , world , test ",
			separator: ",",
			fn:        strings.TrimSpace,
			expected:  []string{"hello", "world", "test"},
		},
		{
			name:      "empty string",
			input:     "",
			separator: ",",
			fn:        strings.ToUpper,
			expected:  []string{""},
		},
		{
			name:      "no separator",
			input:     "hello",
			separator: ",",
			fn:        strings.ToUpper,
			expected:  []string{"HELLO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitAndMap(tt.input, tt.separator, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("SplitAndMap() length = %d, expected %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("SplitAndMap()[%d] = %q, expected %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestOSNewLine(t *testing.T) {
	result := OSNewLine()
	// Should return either "\n" (Unix/Mac) or "\r\n" (Windows)
	if result != "\n" && result != "\r\n" {
		t.Errorf("OSNewLine() = %q, expected either \"\\n\" or \"\\r\\n\"", result)
	}
}

func TestReplaceAllToOne(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		from     []string
		to       string
		expected string
	}{
		{
			name:     "multiple replacements",
			input:    "hello world test",
			from:     []string{"hello", "world"},
			to:       "hi",
			expected: "hi hi test",
		},
		{
			name:     "no matches",
			input:    "hello world",
			from:     []string{"foo", "bar"},
			to:       "hi",
			expected: "hello world",
		},
		{
			name:     "empty from",
			input:    "hello world",
			from:     []string{},
			to:       "hi",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			from:     []string{"hello"},
			to:       "hi",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceAllToOne(tt.input, tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("ReplaceAllToOne() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSplice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		newStr   string
		start    int
		end      int
		expected string
		panics   bool
	}{
		{
			name:     "replace middle",
			input:    "hello world",
			newStr:   "beautiful",
			start:    6,
			end:      11,
			expected: "hello beautiful",
		},
		{
			name:     "remove part",
			input:    "hello world",
			newStr:   "",
			start:    5,
			end:      11,
			expected: "hello",
		},
		{
			name:     "insert at position",
			input:    "helloworld",
			newStr:   " ",
			start:    5,
			end:      5,
			expected: "hello world",
		},
		{
			name:   "start out of range",
			input:  "hello",
			newStr: "hi",
			start:  10,
			end:    15,
			panics: true,
		},
		{
			name:   "end out of range",
			input:  "hello",
			newStr: "hi",
			start:  0,
			end:    15,
			panics: true,
		},
		{
			name:   "end before start",
			input:  "hello",
			newStr: "hi",
			start:  3,
			end:    1,
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Splice() should have panicked")
					}
				}()
			}
			result := Splice(tt.input, tt.newStr, tt.start, tt.end)
			if !tt.panics && result != tt.expected {
				t.Errorf("Splice() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSubstring(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		start     int
		end       int
		expected  string
		expectErr bool
	}{
		{
			name:     "normal substring",
			input:    "hello world",
			start:    0,
			end:      5,
			expected: "hello",
		},
		{
			name:     "end as zero (full length)",
			input:    "hello",
			start:    2,
			end:      0,
			expected: "llo",
		},
		{
			name:     "unicode support",
			input:    "héllö wörld",
			start:    0,
			end:      5,
			expected: "héllö",
		},
		{
			name:      "start out of range",
			input:     "hello",
			start:     10,
			end:       15,
			expectErr: true,
		},
		{
			name:      "end out of range",
			input:     "hello",
			start:     0,
			end:       15,
			expectErr: true,
		},
		{
			name:      "end before start",
			input:     "hello",
			start:     3,
			end:       1,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Substring(tt.input, tt.start, tt.end)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Substring() should have returned an error")
				}
			} else {
				if err != nil {
					t.Errorf("Substring() returned unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Substring() = %q, expected %q", result, tt.expected)
				}
			}
		})
	}
}

func TestMustSubstring(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
		panics   bool
	}{
		{
			name:     "normal substring",
			input:    "hello world",
			start:    0,
			end:      5,
			expected: "hello",
		},
		{
			name:     "unicode support",
			input:    "héllö wörld",
			start:    0,
			end:      5,
			expected: "héllö",
		},
		{
			name:   "start out of range",
			input:  "hello",
			start:  10,
			end:    15,
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MustSubstring() should have panicked")
					}
				}()
			}
			result := MustSubstring(tt.input, tt.start, tt.end)
			if !tt.panics && result != tt.expected {
				t.Errorf("MustSubstring() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkMapLines(b *testing.B) {
	input := "line1\nline2\nline3\nline4\nline5"
	fn := strings.ToUpper

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapLines(input, fn)
	}
}

func BenchmarkSplitAndMap(b *testing.B) {
	input := "part1,part2,part3,part4,part5"
	fn := strings.ToUpper

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SplitAndMap(input, ",", fn)
	}
}

func BenchmarkReplaceAllToOne(b *testing.B) {
	input := "hello world test hello world"
	from := []string{"hello", "world"}
	to := "hi"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReplaceAllToOne(input, from, to)
	}
}

func BenchmarkSubstring(b *testing.B) {
	input := "this is a test string for benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Substring(input, 5, 15)
	}
}
