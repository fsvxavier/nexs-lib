package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummary(t *testing.T) {
	tests := []struct {
		end      string
		input    string
		expected string
		colLen   int
	}{
		{"...", "", "", 15},
		{"...", "Lorem ipsum dolor sit amet", "Lorem ipsum dolor sit amet", 0},
		{"...", "Lorem ipsum dolor sit amet", "Lorem ipsum...", 12},
		{"...", "Lorem\nipsum dolor sit amet", "Lorem...", 12},
		{"...", "Lorem\tipsum\tdolor sit amet", "Lorem\tipsum...", 12},
		{"...", "Lorem ipsum dolor sit amet", "Lorem...", 10},
		{"...", "Lorem\nipsum dolor sit amet", "Lorem...", 10},
		{"...", "Lorem ipsum", "Lorem ipsum", 15},
		{"...", "Lorem ipsum", "Lorem...", 5},
		{"...", "Lorem ipsum", "Lore...", 4},
		{"...", "Lorem         ipsum", "Lorem...", 15},
	}

	for i, test := range tests {
		output := Summary(test.input, test.colLen, test.end)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func ExampleSummary() {
	fmt.Println(Summary("Lorem ipsum dolor sit amet.", 12, "..."))
	// Output: Lorem ipsum...
}
