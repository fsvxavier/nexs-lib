package strutl

import (
	"fmt"
	"testing"
)

func TestAlign(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		alignTo  AlignType
		width    int
		expected string
	}{
		{"Center", "hello", Center, 10, "  hello   "},
		{"Left", "  hello", Left, 10, "hello"},
		{"Right", "hello", Right, 10, "     hello"},
		{"Default", "hello", "invalid", 10, "hello"},
		{"Empty string center", "", Center, 5, "     "},
		{"Empty string left", "  ", Left, 5, ""},
		{"Empty string right", "", Right, 5, "     "},
		{"Multiline center", "hello\nworld", Center, 10, "  hello   \n  world   "},
		{"Multiline left", "  hello\n  world", Left, 10, "hello\nworld"},
		{"Multiline right", "hello\nworld", Right, 10, "     hello\n     world"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Align(test.str, test.alignTo, test.width)
			if result != test.expected {
				t.Errorf("Align(%q, %v, %d) = %q; want %q",
					test.str, test.alignTo, test.width, result, test.expected)
			}
		})
	}
}

func TestAlignLeft(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Space only", "   ", ""},
		{"Single line", "  hello  ", "hello  "},
		{"Multiple lines", "  linha1\n    linha2", "linha1\nlinha2"},
		{"Mixed spaces", "\t hello\n   world  ", "hello\nworld  "},
		{"Already left aligned", "hello\nworld", "hello\nworld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AlignLeft(test.input)
			if result != test.expected {
				t.Errorf("AlignLeft(%q) = %q; want %q",
					test.input, result, test.expected)
			}
		})
	}
}

func TestAlignRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"Empty string", "", 5, "     "},
		{"Single line", "hello", 10, "     hello"},
		{"Single line with spaces", "  hello  ", 10, "     hello"},
		{"Multiple lines", "linha1\nlinha2", 10, "     linha1\n     linha2"},
		{"Width smaller than text", "hello world", 5, "hello world"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AlignRight(test.input, test.width)
			if result != test.expected {
				t.Errorf("AlignRight(%q, %d) = %q; want %q",
					test.input, test.width, result, test.expected)
			}
		})
	}
}

func TestAlignCenter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"Empty string", "", 5, "     "},
		{"Single line", "hello", 11, "   hello   "},
		{"Single line with spaces", "  hello  ", 10, "  hello   "},
		{"Multiple lines", "linha1\nlinha2", 10, "  linha1  \n  linha2  "},
		{"Width smaller than text", "hello world", 5, "hello world"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AlignCenter(test.input, test.width)
			if result != test.expected {
				t.Errorf("AlignCenter(%q, %d) = %q; want %q",
					test.input, test.width, result, test.expected)
			}
		})
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"Empty string", "", 5, "     "},
		{"Basic centering", "hello", 11, "   hello   "},
		{"Even padding", "center", 10, "  center  "},
		{"Uneven padding", "text", 10, "   text   "},
		{"Width smaller than text", "hello world", 5, "hello world"},
		{"Unicode text", "世界", 6, "  世界  "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CenterText(test.input, test.width)
			if result != test.expected {
				t.Errorf("CenterText(%q, %d) = %q; want %q",
					test.input, test.width, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleAlign() {
	fmt.Println(Align("hello", Center, 10))
	fmt.Println(Align("hello", Left, 10))
	fmt.Println(Align("hello", Right, 10))
	// Output:
	//   hello
	// hello
	//      hello
}

func ExampleAlignLeft() {
	fmt.Println(AlignLeft("    hello    "))
	fmt.Println(AlignLeft("  linha1\n    linha2"))
	// Output:
	// hello
	// linha1
	// linha2
}

func ExampleAlignRight() {
	fmt.Println(AlignRight("hello", 10))
	fmt.Println(AlignRight("linha1\nlinha2", 10))
	// Output:
	//      hello
	//      linha1
	//      linha2
}

func ExampleAlignCenter() {
	fmt.Println(AlignCenter("hello", 11))
	fmt.Println(AlignCenter("linha1\nlinha2", 10))
	// Output:
	//    hello
	//   linha1
	//   linha2
}

func ExampleCenterText() {
	fmt.Println(CenterText("hello", 11))
	fmt.Println(CenterText("center", 10))
	// Output:
	//    hello
	//   center
}

// Benchmarks
func BenchmarkAlign(b *testing.B) {
	benchCases := []struct {
		name    string
		str     string
		alignTo AlignType
		width   int
	}{
		{"Center", "hello world", Center, 20},
		{"Left", "  hello world  ", Left, 20},
		{"Right", "hello world", Right, 20},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Align(bc.str, bc.alignTo, bc.width)
			}
		})
	}
}

func BenchmarkAlignLeft(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
	}{
		{"SingleLine", "  hello world  "},
		{"MultiLine", "  line1\n  line2\n  line3"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = AlignLeft(bc.input)
			}
		})
	}
}

func BenchmarkAlignRight(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
		width int
	}{
		{"SingleLine", "hello world", 20},
		{"MultiLine", "line1\nline2\nline3", 20},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = AlignRight(bc.input, bc.width)
			}
		})
	}
}

func BenchmarkAlignCenter(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
		width int
	}{
		{"SingleLine", "hello world", 20},
		{"MultiLine", "line1\nline2\nline3", 20},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = AlignCenter(bc.input, bc.width)
			}
		})
	}
}

func BenchmarkCenterText(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
		width int
	}{
		{"Short", "hello", 20},
		{"Medium", "hello world", 30},
		{"Unicode", "こんにちは", 20},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = CenterText(bc.input, bc.width)
			}
		})
	}
}
