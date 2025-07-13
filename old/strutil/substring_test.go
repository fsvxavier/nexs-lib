package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustSubstring(t *testing.T) {
	tests := []struct {
		input     string
		expected  string
		start     int
		end       int
		mustPanic bool
	}{
		{"lorem", "l", 0, 1, false},
		{"", "", 0, 1, true},
		{"lorem", "lorem", 0, 5, false},
		{"lorem", "", 0, 10, true},
		{"lorem", "", -1, 4, true},
		{"lorem", "", 9, 10, true},
		{"lorem", "", 4, 3, true},
		{"Υπάρχουν", "πάρ", 1, 4, false},
		{"Υπάρχουν", "πάρχουν", 1, 0, false},
		{"Υπάρχουν", "", 1, 9, true},
		{"žůžo", "ůžo", 1, 4, false},
	}

	for i, test := range tests {
		if test.mustPanic {
			assert.Panicsf(t, func() {
				_ = MustSubstring(test.input, test.start, test.end)
			}, "Test case %d is not successful\n", i)
		} else {
			output := MustSubstring(test.input, test.start, test.end)
			assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
		}
	}

}

func ExampleMustSubstring() {
	fmt.Println(MustSubstring("Υπάρχουν", 1, 4))
	// Output: πάρ
}

func ExampleMustSubstring_tillTheEnd() {
	fmt.Println(MustSubstring("Υπάρχουν", 1, 0))
	// Output: πάρχουν
}
