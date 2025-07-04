package strutl

import (
	"fmt"
	"testing"
)

func TestDrawBox(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		align    AlignType
		expected string
		width    int
		wantErr  bool
	}{
		{"Empty string", "", Center, "┌──────────────────┐\n│                  │\n└──────────────────┘", 20, false},
		{"Basic text", "Hello World", Center, "┌──────────────────┐\n│   Hello World    │\n└──────────────────┘", 20, false},
		{"With newlines", "\nHello World\n", Center, "┌──────────────────┐\n│                  │\n│   Hello World    │\n│                  │\n└──────────────────┘", 20, false},
		{"Unicode content", "résumé", Left, "┌────────┐\n│résumé  │\n└────────┘", 10, false},
		{"Width too small", "Hello World", Left, "", 2, true},
		{"Multiple newlines", "Hello\n\n\nWorld", Left, "┌────────┐\n│Hello   │\n│        │\n│        │\n│World   │\n└────────┘", 10, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DrawBox(test.input, test.width, test.align)

			if test.wantErr {
				if err == nil {
					t.Errorf("DrawBox(%q, %d, %v) esperava erro, mas não obteve nenhum", test.input, test.width, test.align)
				}
				return
			}

			if err != nil {
				t.Errorf("DrawBox(%q, %d, %v) obteve erro inesperado: %v", test.input, test.width, test.align, err)
				return
			}

			if result != test.expected {
				t.Errorf("DrawBox(%q, %d, %v) =\n%q\nEsperado:\n%q", test.input, test.width, test.align, result, test.expected)
			}
		})
	}
}

func TestDrawCustomBox(t *testing.T) {
	test9Slice := SimpleBox9Slice()
	test9Slice.Top = "-৹_৹"
	test9Slice.Left = "Ͼ|"
	test9Slice.Right = "|Ͽ"
	test9Slice.Bottom = "৹-৹"

	tests := []struct {
		name     string
		input    string
		align    AlignType
		expected string
		width    int
		wantErr  bool
	}{
		{"Empty string", "", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|                |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"Basic text", "Hello World", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|  Hello World   |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"With newlines", "\nHello World\n", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|                |Ͽ\nϾ|  Hello World   |Ͽ\nϾ|                |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"Unicode content", "résumé", Left, "+-৹_৹-৹_৹+\nϾ|résumé|Ͽ\n+৹-৹৹-৹৹-+", 10, false},
		{"Unicode with wrap", "résumé", Left, "+-৹_৹-৹_+\nϾ|résum|Ͽ\nϾ|é    |Ͽ\n+৹-৹৹-৹৹+", 9, false},
		{"Width too small", "Hello World", Left, "", 2, true},
		{"Multiple newlines with right align", "Hello\n\n\nWorld", Right, "+-৹_৹-৹_৹+\nϾ| Hello|Ͽ\nϾ|      |Ͽ\nϾ|      |Ͽ\nϾ| World|Ͽ\n+৹-৹৹-৹৹-+", 10, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DrawCustomBox(test.input, test.width, test.align, &test9Slice, "\n")

			if test.wantErr {
				if err == nil {
					t.Errorf("DrawCustomBox(%q, %d, %v, ...) esperava erro, mas não obteve nenhum", test.input, test.width, test.align)
				}
				return
			}

			if err != nil {
				t.Errorf("DrawCustomBox(%q, %d, %v, ...) obteve erro inesperado: %v", test.input, test.width, test.align, err)
				return
			}

			if result != test.expected {
				t.Errorf("DrawCustomBox(%q, %d, %v, ...) =\n%q\nEsperado:\n%q", test.input, test.width, test.align, result, test.expected)
			}
		})
	}
}

// Exemplos para documentação
func ExampleDrawBox() {
	output, _ := DrawBox("Hello World", 20, Center)
	fmt.Println(output)
	// Output:
	// ┌──────────────────┐
	// │   Hello World    │
	// └──────────────────┘
}

func ExampleDrawBox_long() {
	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.`
	output, _ := DrawBox(text, 30, Left)
	fmt.Println(output)
	// Output:
	// ┌────────────────────────────┐
	// │Lorem ipsum dolor sit amet, │
	// │consectetur adipiscing elit,│
	// │sed do eiusmod tempor       │
	// │incididunt ut labore et     │
	// │dolore magna aliqua. Ut enim│
	// │ad minim veniam, quis       │
	// │nostrud exercitation ullamco│
	// │laboris nisi ut aliquip ex  │
	// │ea commodo consequat.       │
	// └────────────────────────────┘
}

func ExampleDrawCustomBox() {
	defaultBox9Slice := DefaultBox9Slice()
	output, _ := DrawCustomBox("Hello World", 20, Center, &defaultBox9Slice, "\n")
	fmt.Println(output)
	// Output:
	// ┌──────────────────┐
	// │   Hello World    │
	// └──────────────────┘
}

// Benchmarks
func BenchmarkDrawBox(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
		width int
		align AlignType
	}{
		{"Short", "Hello World", 20, Center},
		{"Medium", "This is a medium length text that will need to be wrapped", 30, Left},
		{"Long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 40, Right},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = DrawBox(bc.input, bc.width, bc.align)
			}
		})
	}
}

func BenchmarkDrawCustomBox(b *testing.B) {
	slice := SimpleBox9Slice()

	benchCases := []struct {
		name  string
		input string
		width int
		align AlignType
	}{
		{"Short", "Hello World", 20, Center},
		{"Medium", "This is a medium length text that will need to be wrapped", 30, Left},
		{"Long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 40, Right},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = DrawCustomBox(bc.input, bc.width, bc.align, &slice, "\n")
			}
		})
	}
}
