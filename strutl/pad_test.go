package strutl

import (
	"fmt"
	"testing"
)

func TestPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		width    int
		pad      string
		expected string
	}{
		{"Empty string", "", 5, " ", "     "},
		{"No padding needed", "hello", 3, " ", "hello"},
		{"Single space padding", "hello", 7, " ", "  hello"},
		{"Zero padding", "123", 5, "0", "00123"},
		{"Multi-char padding", "abc", 9, "xy", "xyxyxyabc"},
		{"Unicode string", "世界", 4, "-", "--世界"},
		{"Unicode padding", "hello", 7, "世", "世世hello"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PadLeft(test.str, test.width, test.pad)
			if result != test.expected {
				t.Errorf("PadLeft(%q, %d, %q) = %q; want %q",
					test.str, test.width, test.pad, result, test.expected)
			}

			// Verifica o comprimento se esperamos que tenha preenchimento
			if Len(test.str) < test.width {
				if Len(result) != test.width {
					t.Errorf("Len(PadLeft(%q, %d, %q)) = %d; want %d",
						test.str, test.width, test.pad, Len(result), test.width)
				}
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		width    int
		pad      string
		expected string
	}{
		{"Empty string", "", 5, " ", "     "},
		{"No padding needed", "hello", 3, " ", "hello"},
		{"Single space padding", "hello", 7, " ", "hello  "},
		{"Zero padding", "123", 5, "0", "12300"},
		{"Multi-char padding", "abc", 9, "xy", "abcxyxyx"},
		{"Unicode string", "世界", 4, "-", "世界--"},
		{"Unicode padding", "hello", 7, "世", "hello世世"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PadRight(test.str, test.width, test.pad)
			if result != test.expected {
				t.Errorf("PadRight(%q, %d, %q) = %q; want %q",
					test.str, test.width, test.pad, result, test.expected)
			}

			// Verifica o comprimento se esperamos que tenha preenchimento
			if Len(test.str) < test.width {
				if Len(result) != test.width {
					t.Errorf("Len(PadRight(%q, %d, %q)) = %d; want %d",
						test.str, test.width, test.pad, Len(result), test.width)
				}
			}
		})
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		width    int
		leftPad  string
		rightPad string
		expected string
	}{
		{"Empty string", "", 6, " ", " ", "      "},
		{"No padding needed", "hello", 3, " ", " ", "hello"},
		{"Equal padding", "hello", 9, " ", " ", "  hello  "},
		{"Unequal width (odd)", "hello", 10, " ", " ", "  hello   "},
		{"Different pad chars", "hello", 9, "-", "+", "--hello++"},
		{"Empty left pad", "hello", 9, "", "+", "hello++++"},
		{"Empty right pad", "hello", 9, "-", "", "----hello"},
		{"Unicode string", "世界", 6, "-", "+", "--世界++"},
		{"Unicode padding", "hello", 9, "世", "界", "世世hello界界"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Pad(test.str, test.width, test.leftPad, test.rightPad)
			if result != test.expected {
				t.Errorf("Pad(%q, %d, %q, %q) = %q; want %q",
					test.str, test.width, test.leftPad, test.rightPad, result, test.expected)
			}

			// Verifica o comprimento se esperamos que tenha preenchimento
			if Len(test.str) < test.width {
				if Len(result) != test.width {
					t.Errorf("Len(Pad(%q, %d, %q, %q)) = %d; want %d",
						test.str, test.width, test.leftPad, test.rightPad, Len(result), test.width)
				}
			}
		})
	}
}

// Exemplos para documentação
func ExamplePadLeft() {
	fmt.Println(PadLeft("hello", 10, " "))
	fmt.Println(PadLeft("123", 5, "0"))
	// Output:
	//      hello
	// 00123
}

func ExamplePadRight() {
	fmt.Println(PadRight("hello", 10, " "))
	fmt.Println(PadRight("123", 5, "0"))
	// Output:
	// hello
	// 12300
}

func ExamplePad() {
	fmt.Println(Pad("hello", 10, " ", " "))
	fmt.Println(Pad("center", 10, "-", "-"))
	// Output:
	//   hello
	// --center--
}

// Benchmarks
func BenchmarkPadLeft(b *testing.B) {
	benchCases := []struct {
		name  string
		str   string
		width int
		pad   string
	}{
		{"Small ASCII", "hello", 10, " "},
		{"Large ASCII", "hello", 50, " "},
		{"Multi-char pad", "hello", 20, "xy"},
		{"Unicode string", "世界", 10, " "},
		{"Unicode pad", "hello", 10, "世"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = PadLeft(bc.str, bc.width, bc.pad)
			}
		})
	}
}

func BenchmarkPadRight(b *testing.B) {
	benchCases := []struct {
		name  string
		str   string
		width int
		pad   string
	}{
		{"Small ASCII", "hello", 10, " "},
		{"Large ASCII", "hello", 50, " "},
		{"Multi-char pad", "hello", 20, "xy"},
		{"Unicode string", "世界", 10, " "},
		{"Unicode pad", "hello", 10, "世"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = PadRight(bc.str, bc.width, bc.pad)
			}
		})
	}
}

func BenchmarkPad(b *testing.B) {
	benchCases := []struct {
		name     string
		str      string
		width    int
		leftPad  string
		rightPad string
	}{
		{"Small ASCII", "hello", 10, " ", " "},
		{"Large ASCII", "hello", 50, " ", " "},
		{"Different pads", "hello", 20, "-", "+"},
		{"Unicode string", "世界", 10, " ", " "},
		{"Unicode pads", "hello", 10, "世", "界"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Pad(bc.str, bc.width, bc.leftPad, bc.rightPad)
			}
		})
	}
}
