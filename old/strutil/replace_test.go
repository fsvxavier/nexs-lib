package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAllToOne(t *testing.T) {
	tests := []struct {
		input    string
		to       string
		expected string
		from     []string
	}{
		{"", "b", "", []string{"a"}},
		{"lorem", "-", "-l-o-r-e-m-", []string{""}},
		{"lorem", "", "r", []string{"lo", "em"}},
		{"a b c a c f", "-", "----c---c-f", []string{" ", "a", "b"}},
	}

	for i, test := range tests {
		output := ReplaceAllToOne(test.input, test.from, test.to)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExampleReplaceAllToOne() {
	fmt.Println(ReplaceAllToOne("lorem", []string{"lo", "em"}, "x"))
	// Output: xrx
}

func TestSplice(t *testing.T) {
	tests := []struct {
		input     string
		newStr    string
		expected  string
		start     int
		end       int
		mustPanic bool
	}{
		{"lorem", "x", "xrem", 0, 2, false},
		{"lorem", "", "rem", 0, 2, false},
		{"", "x", "", 0, 2, false},
		{"lorem", "x", "lorex", 4, 5, false},
		{"lorem", "ipsum", "loipsumem", 2, 3, false},
		{"lorem", "x", "", 5, 6, true},
		{"lorem", "x", "", 4, 4, true},
		{"lorem", "x", "", 4, 3, true},
	}

	for i, test := range tests {
		if test.mustPanic {
			assert.Panicsf(t, func() {
				_ = Splice(test.input, test.newStr, test.start, test.end)
			}, "Test case %d is not successful\n", i)
		} else {
			output := Splice(test.input, test.newStr, test.start, test.end)
			assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
		}
	}
}

func ExampleSplice() {
	fmt.Println(Splice("Lorem", "ipsum", 2, 3))
	// Output: Loipsumem
}
