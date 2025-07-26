package strutil

import (
	"strings"
	"testing"
)

func TestAlign(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		align int
		width int
		want  string
	}{
		{
			name:  "left align",
			text:  "hello",
			align: AlignLeft,
			width: 10,
			want:  "hello     ",
		},
		{
			name:  "right align",
			text:  "hello",
			align: AlignRight,
			width: 10,
			want:  "     hello",
		},
		{
			name:  "center align even padding",
			text:  "hello",
			align: AlignCenter,
			width: 11,
			want:  "   hello   ",
		},
		{
			name:  "center align odd padding",
			text:  "hello",
			align: AlignCenter,
			width: 10,
			want:  "  hello   ",
		},
		{
			name:  "text longer than width",
			text:  "hello world",
			align: AlignCenter,
			width: 5,
			want:  "hello world",
		},
		{
			name:  "zero width",
			text:  "hello",
			align: AlignLeft,
			width: 0,
			want:  "hello",
		},
		{
			name:  "negative width",
			text:  "hello",
			align: AlignLeft,
			width: -5,
			want:  "hello",
		},
		{
			name:  "text equals width",
			text:  "hello",
			align: AlignCenter,
			width: 5,
			want:  "hello",
		},
		{
			name:  "invalid align",
			text:  "hello",
			align: 99,
			width: 10,
			want:  "hello",
		},
		{
			name:  "empty text",
			text:  "",
			align: AlignCenter,
			width: 5,
			want:  "     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Align(tt.text, tt.align, tt.width); got != tt.want {
				t.Errorf("Align() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{
			name:  "normal case",
			text:  "hello",
			width: 10,
			want:  "  hello   ",
		},
		{
			name:  "empty text",
			text:  "",
			width: 5,
			want:  "     ",
		},
		{
			name:  "text longer than width",
			text:  "hello world",
			width: 5,
			want:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Center(tt.text, tt.width); got != tt.want {
				t.Errorf("Center() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		pad    string
		length int
		want   string
	}{
		{
			name:   "zero padding",
			str:    "42",
			pad:    "0",
			length: 5,
			want:   "00042",
		},
		{
			name:   "space padding",
			str:    "hello",
			pad:    " ",
			length: 10,
			want:   "     hello",
		},
		{
			name:   "multi-character padding",
			str:    "test",
			pad:    "ab",
			length: 8,
			want:   "ababtest",
		},
		{
			name:   "no padding needed",
			str:    "hello",
			pad:    "0",
			length: 5,
			want:   "hello",
		},
		{
			name:   "string longer than length",
			str:    "hello world",
			pad:    "0",
			length: 5,
			want:   "hello world",
		},
		{
			name:   "empty pad",
			str:    "hello",
			pad:    "",
			length: 10,
			want:   "hello",
		},
		{
			name:   "empty string",
			str:    "",
			pad:    "0",
			length: 3,
			want:   "000",
		},
		{
			name:   "zero length",
			str:    "hello",
			pad:    "0",
			length: 0,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadLeft(tt.str, tt.length, tt.pad); got != tt.want {
				t.Errorf("PadLeft() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		pad    string
		length int
		want   string
	}{
		{
			name:   "zero padding",
			str:    "42",
			pad:    "0",
			length: 5,
			want:   "42000",
		},
		{
			name:   "space padding",
			str:    "hello",
			pad:    " ",
			length: 10,
			want:   "hello     ",
		},
		{
			name:   "multi-character padding",
			str:    "test",
			pad:    "xy",
			length: 8,
			want:   "testxyxy",
		},
		{
			name:   "no padding needed",
			str:    "hello",
			pad:    "0",
			length: 5,
			want:   "hello",
		},
		{
			name:   "string longer than length",
			str:    "hello world",
			pad:    "0",
			length: 5,
			want:   "hello world",
		},
		{
			name:   "empty pad",
			str:    "hello",
			pad:    "",
			length: 10,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadRight(tt.str, tt.length, tt.pad); got != tt.want {
				t.Errorf("PadRight() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPadBoth(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		pad    string
		length int
		want   string
	}{
		{
			name:   "even padding",
			str:    "42",
			pad:    "0",
			length: 6,
			want:   "004200",
		},
		{
			name:   "odd padding - left priority",
			str:    "42",
			pad:    "0",
			length: 5,
			want:   "00420",
		},
		{
			name:   "space padding",
			str:    "hello",
			pad:    " ",
			length: 11,
			want:   "   hello   ",
		},
		{
			name:   "multi-character padding",
			str:    "hi",
			pad:    "xy",
			length: 8,
			want:   "xyxhixyx",
		},
		{
			name:   "no padding needed",
			str:    "hello",
			pad:    "0",
			length: 5,
			want:   "hello",
		},
		{
			name:   "string longer than length",
			str:    "hello world",
			pad:    "0",
			length: 5,
			want:   "hello world",
		},
		{
			name:   "empty pad",
			str:    "hello",
			pad:    "",
			length: 10,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadBoth(tt.str, tt.pad, tt.length); got != tt.want {
				t.Errorf("PadBoth() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIndent(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			text:   "hello",
			prefix: "  ",
			want:   "  hello",
		},
		{
			name:   "multi line",
			text:   "line1\nline2\nline3",
			prefix: "  ",
			want:   "  line1\n  line2\n  line3",
		},
		{
			name:   "empty lines preserved",
			text:   "line1\n\nline3",
			prefix: "  ",
			want:   "  line1\n\n  line3",
		},
		{
			name:   "empty text",
			text:   "",
			prefix: "  ",
			want:   "",
		},
		{
			name:   "empty prefix",
			text:   "hello\nworld",
			prefix: "",
			want:   "hello\nworld",
		},
		{
			name:   "tab prefix",
			text:   "function() {\n  return true;\n}",
			prefix: "\t",
			want:   "\tfunction() {\n\t  return true;\n\t}",
		},
		{
			name:   "whitespace only line",
			text:   "line1\n   \nline3",
			prefix: ">>",
			want:   ">>line1\n>>   \n>>line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Indent(tt.text, tt.prefix); got != tt.want {
				t.Errorf("Indent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		tabSize int
		want    string
	}{
		{
			name:    "single tab",
			text:    "a\tb",
			tabSize: 4,
			want:    "a   b",
		},
		{
			name:    "multiple tabs",
			text:    "a\tb\tc",
			tabSize: 4,
			want:    "a   b   c",
		},
		{
			name:    "tab at beginning",
			text:    "\thello",
			tabSize: 4,
			want:    "    hello",
		},
		{
			name:    "tab size 8",
			text:    "a\tb",
			tabSize: 8,
			want:    "a       b",
		},
		{
			name:    "tab size 1",
			text:    "a\tb",
			tabSize: 1,
			want:    "a b",
		},
		{
			name:    "no tabs",
			text:    "hello world",
			tabSize: 4,
			want:    "hello world",
		},
		{
			name:    "newlines reset column",
			text:    "abc\tdef\ngh\ti",
			tabSize: 4,
			want:    "abc def\ngh  i",
		},
		{
			name:    "empty text",
			text:    "",
			tabSize: 4,
			want:    "",
		},
		{
			name:    "zero tab size",
			text:    "a\tb",
			tabSize: 0,
			want:    "a\tb",
		},
		{
			name:    "negative tab size",
			text:    "a\tb",
			tabSize: -1,
			want:    "a\tb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandTabs(tt.text, tt.tabSize); got != tt.want {
				t.Errorf("ExpandTabs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		lineWidth int
		want      string
	}{
		{
			name:      "simple wrap",
			text:      "This is a long sentence that needs wrapping",
			lineWidth: 10,
			want:      "This is a\nlong\nsentence\nthat needs\nwrapping",
		},
		{
			name:      "no wrap needed",
			text:      "Short text",
			lineWidth: 20,
			want:      "Short text",
		},
		{
			name:      "existing newlines preserved",
			text:      "Line 1\nLine 2 is longer and needs wrapping",
			lineWidth: 10,
			want:      "Line 1\nLine 2 is\nlonger and\nneeds\nwrapping",
		},
		{
			name:      "empty text",
			text:      "",
			lineWidth: 10,
			want:      "",
		},
		{
			name:      "zero width",
			text:      "hello world",
			lineWidth: 0,
			want:      "hello world",
		},
		{
			name:      "negative width",
			text:      "hello world",
			lineWidth: -5,
			want:      "hello world",
		},
		{
			name:      "single long word",
			text:      "supercalifragilisticexpialidocious",
			lineWidth: 10,
			want:      "supercalifragilisticexpialidocious",
		},
		{
			name:      "exact width fit",
			text:      "hello world",
			lineWidth: 11,
			want:      "hello world",
		},
		{
			name:      "multiple spaces",
			text:      "word1    word2",
			lineWidth: 10,
			want:      "word1\nword2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WordWrap(tt.text, tt.lineWidth); got != tt.want {
				t.Errorf("WordWrap() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDrawBox(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		border rune
		want   string
	}{
		{
			name:   "simple box",
			text:   "Hello",
			border: '*',
			want:   "*********\n* Hello *\n*********",
		},
		{
			name:   "multi-line box",
			text:   "Line 1\nLine 2",
			border: '+',
			want:   "++++++++++\n+ Line 1 +\n+ Line 2 +\n++++++++++",
		},
		{
			name:   "empty text",
			text:   "",
			border: '#',
			want:   "",
		},
		{
			name:   "single character",
			text:   "A",
			border: '=',
			want:   "=====\n= A =\n=====",
		},
		{
			name:   "different width lines",
			text:   "Short\nLonger line",
			border: '|',
			want:   "|||||||||||||||\n| Short       |\n| Longer line |\n|||||||||||||||",
		},
		{
			name:   "unicode border",
			text:   "Test",
			border: '█',
			want:   "████████\n█ Test █\n████████",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DrawBoxSimple(tt.text, tt.border); got != tt.want {
				t.Errorf("DrawBoxSimple() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTile(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		width   int
		want    string
	}{
		{
			name:    "simple pattern",
			pattern: "-=",
			width:   10,
			want:    "-=-=-=-=-=",
		},
		{
			name:    "pattern longer than width",
			pattern: "abcdef",
			width:   3,
			want:    "abc",
		},
		{
			name:    "pattern equals width",
			pattern: "test",
			width:   4,
			want:    "test",
		},
		{
			name:    "single character pattern",
			pattern: "*",
			width:   5,
			want:    "*****",
		},
		{
			name:    "empty pattern",
			pattern: "",
			width:   5,
			want:    "",
		},
		{
			name:    "zero width",
			pattern: "abc",
			width:   0,
			want:    "",
		},
		{
			name:    "negative width",
			pattern: "abc",
			width:   -5,
			want:    "",
		},
		{
			name:    "unicode pattern",
			pattern: "🎵🎶",
			width:   6,
			want:    "🎵🎶🎵🎶🎵🎶",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Tile(tt.pattern, tt.width); got != tt.want {
				t.Errorf("Tile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		want      string
	}{
		{
			name:      "text needs truncation",
			text:      "This is a long text that needs to be summarized",
			maxLength: 15,
			want:      "This is a...",
		},
		{
			name:      "text shorter than max",
			text:      "Short text",
			maxLength: 20,
			want:      "Short text",
		},
		{
			name:      "text equals max length",
			text:      "Exact length",
			maxLength: 12,
			want:      "Exact length",
		},
		{
			name:      "empty text",
			text:      "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "zero max length",
			text:      "hello",
			maxLength: 0,
			want:      "",
		},
		{
			name:      "negative max length",
			text:      "hello",
			maxLength: -5,
			want:      "",
		},
		{
			name:      "max length too small for ellipsis",
			text:      "hello world",
			maxLength: 2,
			want:      "he",
		},
		{
			name:      "word boundary truncation",
			text:      "word1 word2 word3 word4",
			maxLength: 15,
			want:      "word1 word2...",
		},
		{
			name:      "no good word boundary",
			text:      "verylongwordwithoutspaces",
			maxLength: 10,
			want:      "verylon...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Summary(tt.text, tt.maxLength, "..."); got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark tests for formatting functions
func BenchmarkAlign(b *testing.B) {
	text := "Hello World"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Align(text, AlignCenter, 20)
	}
}

func BenchmarkPadLeft(b *testing.B) {
	text := "test"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PadLeft(text, 20, "0")
	}
}

func BenchmarkIndent(b *testing.B) {
	text := strings.Repeat("line of code\n", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Indent(text, "  ")
	}
}

func BenchmarkWordWrap(b *testing.B) {
	text := strings.Repeat("word ", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		WordWrap(text, 50)
	}
}

func BenchmarkExpandTabs(b *testing.B) {
	text := strings.Repeat("line\twith\ttabs\n", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ExpandTabs(text, 4)
	}
}

// Edge case tests
// Test the newly added alignment functions
func TestAlignLeftText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "normal case",
			text: "  hello  ",
			want: "hello  ",
		},
		{
			name: "multiline with leading spaces",
			text: "  line1  \n   line2   ",
			want: "line1  \nline2   ",
		},
		{
			name: "empty text",
			text: "",
			want: "",
		},
		{
			name: "only spaces",
			text: "   ",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlignLeftText(tt.text); got != tt.want {
				t.Errorf("AlignLeftText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlignRightText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{
			name:  "normal case",
			text:  "hello",
			width: 10,
			want:  "     hello",
		},
		{
			name:  "multiline",
			text:  "hello\nworld",
			width: 8,
			want:  "   hello\n   world",
		},
		{
			name:  "empty text",
			text:  "",
			width: 5,
			want:  "     ",
		},
		{
			name:  "text with spaces",
			text:  "  hello  ",
			width: 10,
			want:  "     hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlignRightText(tt.text, tt.width); got != tt.want {
				t.Errorf("AlignRightText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlignCenterText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{
			name:  "even padding",
			text:  "hello",
			width: 11,
			want:  "   hello   ",
		},
		{
			name:  "odd padding",
			text:  "hello",
			width: 10,
			want:  "  hello   ",
		},
		{
			name:  "multiline",
			text:  "hi\nbye",
			width: 6,
			want:  "  hi  \n bye  ",
		},
		{
			name:  "text with spaces",
			text:  "  hello  ",
			width: 10,
			want:  "  hello   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlignCenterText(tt.text, tt.width); got != tt.want {
				t.Errorf("AlignCenterText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{
			name:  "normal case",
			text:  "hello",
			width: 11,
			want:  "   hello   ",
		},
		{
			name:  "text longer than width",
			text:  "very long text",
			width: 5,
			want:  "very long text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CenterText(tt.text, tt.width); got != tt.want {
				t.Errorf("CenterText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWordWrapWithBreak(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		width          int
		breakLongWords bool
		want           string
	}{
		{
			name:           "break long words enabled",
			text:           "supercalifragilisticexpialidocious",
			width:          10,
			breakLongWords: true,
			want:           "supercalif\nragilistic\nexpialidoc\nious",
		},
		{
			name:           "break long words disabled",
			text:           "supercalifragilisticexpialidocious",
			width:          10,
			breakLongWords: false,
			want:           "supercalifragilisticexpialidocious",
		},
		{
			name:           "normal wrapping",
			text:           "hello world how are you",
			width:          10,
			breakLongWords: true,
			want:           "hello\nworld how\nare you",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WordWrapWithBreak(tt.text, tt.width, tt.breakLongWords); got != tt.want {
				t.Errorf("WordWrapWithBreak() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBox9Slice(t *testing.T) {
	// Test DefaultBox9Slice
	defaultBox := DefaultBox9Slice()
	if defaultBox.TopLeft != "┌" || defaultBox.TopRight != "┐" {
		t.Errorf("DefaultBox9Slice() incorrect values")
	}

	// Test SimpleBox9Slice
	simpleBox := SimpleBox9Slice()
	if simpleBox.TopLeft != "+" || simpleBox.TopRight != "+" {
		t.Errorf("SimpleBox9Slice() incorrect values")
	}
}

func TestDrawCustomBox(t *testing.T) {
	tests := []struct {
		name    string
		content string
		width   int
		align   AlignType
		chars   Box9Slice
		want    string
	}{
		{
			name:    "simple box with default chars",
			content: "hello",
			width:   10,
			align:   LeftAlign,
			chars:   DefaultBox9Slice(),
			want:    "┌────────┐\n│hello   │\n└────────┘",
		},
		{
			name:    "simple box with ASCII chars",
			content: "hello",
			width:   10,
			align:   LeftAlign,
			chars:   SimpleBox9Slice(),
			want:    "+--------+\n|hello   |\n+--------+",
		},
		{
			name:    "center aligned box",
			content: "hi",
			width:   8,
			align:   CenterAlign,
			chars:   DefaultBox9Slice(),
			want:    "┌──────┐\n│  hi  │\n└──────┘",
		},
		{
			name:    "right aligned box",
			content: "hi",
			width:   8,
			align:   RightAlign,
			chars:   DefaultBox9Slice(),
			want:    "┌──────┐\n│    hi│\n└──────┘",
		},
		{
			name:    "multiline content",
			content: "line1\nline2",
			width:   10,
			align:   LeftAlign,
			chars:   DefaultBox9Slice(),
			want:    "┌────────┐\n│line1   │\n│line2   │\n└────────┘",
		},
		{
			name:    "width too small",
			content: "hello",
			width:   2,
			align:   LeftAlign,
			chars:   DefaultBox9Slice(),
			want:    "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DrawCustomBox(tt.content, tt.width, tt.align, tt.chars); got != tt.want {
				t.Errorf("DrawCustomBox() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDrawBoxWithAlign(t *testing.T) {
	tests := []struct {
		name    string
		content string
		width   int
		align   AlignType
		want    string
	}{
		{
			name:    "left aligned",
			content: "hello",
			width:   10,
			align:   LeftAlign,
			want:    "┌────────┐\n│hello   │\n└────────┘",
		},
		{
			name:    "center aligned",
			content: "hi",
			width:   8,
			align:   CenterAlign,
			want:    "┌──────┐\n│  hi  │\n└──────┘",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DrawBoxWithAlign(tt.content, tt.width, tt.align); got != tt.want {
				t.Errorf("DrawBoxWithAlign() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatterEdgeCases(t *testing.T) {

	t.Run("unicode handling", func(t *testing.T) {
		unicodeText := "Hello 世界 🌍"

		// Test that functions handle Unicode correctly
		result := Center(unicodeText, 20)
		if len(result) != 20 {
			t.Errorf("Center with Unicode should produce exact width, got %d", len(result))
		}

		// Test padding with Unicode
		padResult := PadLeft("test", 10, "🎵")
		if padResult == "test" {
			t.Error("PadLeft should have added Unicode padding")
		}
	})

	t.Run("very long strings", func(t *testing.T) {
		longText := strings.Repeat("word ", 10000)

		// Should not panic or cause excessive memory usage
		WordWrap(longText, 80)
		Summary(longText, 100, "...")
		Indent(longText, "  ")
	})

	t.Run("memory efficiency", func(t *testing.T) {
		// Test that string builders are used efficiently
		text := strings.Repeat("test line\n", 1000)

		result := Indent(text, "  ")
		if !strings.Contains(result, "  test line") {
			t.Error("Indent should preserve content with prefix")
		}
	})
}
