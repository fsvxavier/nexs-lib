package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrawBox(t *testing.T) {
	tests := []struct {
		input    string
		align    AlignType
		expected string
		width    int
		err      bool
	}{
		{"", Center, "┌──────────────────┐\n│                  │\n└──────────────────┘", 20, false},
		{"Hello World", Center, "┌──────────────────┐\n│   Hello World    │\n└──────────────────┘", 20, false},
		{"\nHello World\n", Center, "┌──────────────────┐\n│                  │\n│   Hello World    │\n│                  │\n└──────────────────┘", 20, false},
		{"résumé", Left, "┌────────┐\n│résumé  │\n└────────┘", 10, false},
		{"Hello World", Left, "", 2, true},
		{"Hello\n\n\nWorld", Left, "┌────────┐\n│Hello   │\n│        │\n│        │\n│World   │\n└────────┘", 10, false},
	}

	for i, test := range tests {
		output, err := DrawBox(test.input, test.width, test.align)
		if test.err {
			assert.Errorf(t, err, "Test case %d is not successful\n", i)
		} else {
			assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
		}
	}
}

func TestDrawCustomBox(t *testing.T) {
	test9Slice := SimpleBox9Slice()
	test9Slice.Top = "-৹_৹"
	test9Slice.Left = "Ͼ|"
	test9Slice.Right = "|Ͽ"
	test9Slice.Bottom = "৹-৹"

	tests := []struct {
		input    string
		align    AlignType
		expected string
		width    int
		err      bool
	}{
		{"", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|                |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"Hello World", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|  Hello World   |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"\nHello World\n", Center, "+-৹_৹-৹_৹-৹_৹-৹_৹-৹+\nϾ|                |Ͽ\nϾ|  Hello World   |Ͽ\nϾ|                |Ͽ\n+৹-৹৹-৹৹-৹৹-৹৹-৹৹-৹+", 20, false},
		{"résumé", Left, "+-৹_৹-৹_৹+\nϾ|résumé|Ͽ\n+৹-৹৹-৹৹-+", 10, false},
		{"résumé", Left, "+-৹_৹-৹_+\nϾ|résum|Ͽ\nϾ|é    |Ͽ\n+৹-৹৹-৹৹+", 9, false},
		{"Hello World", Left, "", 2, true},
		{"Hello\n\n\nWorld", Right, "+-৹_৹-৹_৹+\nϾ| Hello|Ͽ\nϾ|      |Ͽ\nϾ|      |Ͽ\nϾ| World|Ͽ\n+৹-৹৹-৹৹-+", 10, false},
	}

	for i, test := range tests {
		output, err := DrawCustomBox(test.input, test.width, test.align, &test9Slice, "\n")
		if test.err {
			assert.Errorf(t, err, "Test case %d is not successful, expecting error\n", i)
		} else {
			assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
		}
	}
}

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
