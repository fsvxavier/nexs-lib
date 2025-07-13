package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		count    int
	}{
		{"", "", 2},
		{"\t", "", 0},
		{"\t\n\t\n", "  \n  \n", 2},
		{"\t\t", "    ", 2},
		{"\tlorem\n\tipsum\n", "  lorem\n  ipsum\n", 2},
	}

	for i, test := range tests {
		output := ExpandTabs(test.input, test.count)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExampleExpandTabs() {
	fmt.Printf("%s", ExpandTabs("\tlorem\n\tipsum", 2))
	// Output:
	//   lorem
	//   ipsum
}
