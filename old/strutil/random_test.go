package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var randCounter = 0

type dummyRandReader struct {
}

func (r *dummyRandReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(randCounter)
		randCounter++
	}
	return len(p), err
}

func TestRandom(t *testing.T) {
	oldReader := randReader
	randReader = &dummyRandReader{}
	tests := []struct {
		input    string
		expected string
		length   int
	}{
		{"abcdefghij", "abcde", 5},
		{"abcdefghij", "a", 1},
		{"abc", "abcabc", 6},
		{"", "", 5},
		{"aaa", "aaaaa", 5},
	}

	for i, test := range tests {
		randCounter = 0
		output, _ := Random(test.input, test.length)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
	randReader = oldReader
}

func ExampleRandom() {
	fmt.Println(Random("abcdefghik", 5))
}
