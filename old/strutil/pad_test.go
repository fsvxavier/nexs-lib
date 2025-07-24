package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPadLeft(t *testing.T) {
	tests := []struct {
		input    string
		pad      string
		expected string
		width    int
	}{
		{"lorem", "-", "-----lorem", 10},
		{"lorem", "-", "lorem", 5},
		{"lorem", ".-", ".lorem", 6},
		{"lorem", ".-", ".-.-lorem", 9},
		{"lorem", "", "lorem", 10},
		{"lorem", "-", "lorem", 0},
		{"lorem", "-", "lorem", 4},
		{"", "-", "----", 4},
		{"lorem", ".-=", ".lorem", 6},
	}

	for i, test := range tests {
		output := PadLeft(test.input, test.width, test.pad)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExamplePadLeft() {
	fmt.Println(PadLeft("lorem", 10, "-"))
	// Output: -----lorem
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		pad      string
		expected string
		width    int
	}{
		{"lorem", "-", "lorem-----", 10},
		{"lorem", "-", "lorem", 5},
		{"lorem", ".-", "lorem.", 6},
		{"lorem", ".-", "lorem.-.-", 9},
		{"lorem", "", "lorem", 10},
		{"lorem", "-", "lorem", 0},
		{"lorem", "-", "lorem", 4},
		{"", "-", "----", 4},
	}

	for i, test := range tests {
		output := PadRight(test.input, test.width, test.pad)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExamplePadRight() {
	fmt.Println(PadRight("lorem", 10, "-"))
	// Output: lorem-----
}
func TestPad(t *testing.T) {
	tests := []struct {
		input    string
		leftPad  string
		rightPad string
		expected string
		width    int
	}{
		{"lorem", "-", "-", "--lorem--", 9},
		{"lorem", ".-", "-.", ".-lorem-.-", 10},
		{"lorem", ".-", "-.", "lorem", 1},
		{"", ".-", "-.", ".--.", 4},
		{"lorem", "", "", "lorem", 10},
		{"lorem", "-", "", "-----lorem", 10},
	}

	for i, test := range tests {
		output := Pad(test.input, test.width, test.leftPad, test.rightPad)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExamplePad() {
	fmt.Println(Pad("lorem", 9, "-", "-"))
	// Output: --lorem--
}
