package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordwrap(t *testing.T) {
	tests := []struct {
		input          string
		expected       string
		colLen         int
		breakLongWords bool
	}{

		{"Lorem ipsum dolor sit amet", "Lorem ipsum dolor sit amet", 0, false},
		{"Lorem ipsum\ndolor\nsit amet", "Lorem ipsum\ndolor\nsit amet", 15, false},
		{"Lorem ipsum dolor sit amet", "Lorem ipsum\ndolor sit amet", 15, false},
		{"Lorem ipsum     \n   dolor sit amet\n    consectetur", "Lorem ipsum    \n\n   dolor sit\namet\n    consectetur", 15, false},
		{"Lorem ipsum, dolor sit amet.", "Lorem ipsum,\ndolor sit amet.", 15, false},
		{"Lorem ipsum, dolor sit amet.\n", "Lorem ipsum,\ndolor sit amet.\n", 15, false},
		{"\nLorem ipsum, dolor sit amet.\n", "\nLorem ipsum,\ndolor sit amet.\n", 15, false},
		{"\n   Lorem ipsum, dolor sit amet.\n", "\n   Lorem ipsum,\ndolor sit amet.\n", 15, false},
		{"Lorem ipsum, dolor sit amet", "Lorem\nipsum,\ndolor\nsit\namet", 1, false},
		{"Lorem ipsum, dolor sit amet", "L\no\nr\ne\nm\ni\np\ns\nu\nm\n,\nd\no\nl\no\nr\ns\ni\nt\na\nm\ne\nt", 1, true},
		{"Loremipsum, dolorsitamet", "Loremi\npsum,\ndolors\nitamet", 6, true},
		{"Lorem ipsum, dolor sit amet", "Lor\nem\nips\num,\ndol\nor\nsit\name\nt", 3, true},
		{"Το Lorem Ipsum είναι απλά ένα κείμενο χωρίς", "Το Lorem Ipsum\nείναι απλά ένα\nκείμενο χωρίς", 15, false},
		{"", "", 15, false},
		{"                        ", "               \n        ", 15, false},
		{"Lorem ipsum   dolor sit amet", "Lorem ipsum  \ndolor sit amet", 15, false},
		{"Lorem ipsum,   dolor sit amet.", "Lorem ipsum,  \ndolor sit amet.", 15, false},
		{"   Lorem ipsum,   dolor sit amet.", "   Lorem ipsum,\n  dolor sit\namet.", 15, false},
		{"Lorem ipsum,dolor sit amet.", "Lorem\nipsum,dolor sit\namet.", 15, false},
		{"Lorem ipsum,dolor sit amet.   ", "Lorem\nipsum,dolor sit\namet.   ", 15, false},
		{"Lorem ipsum,dolor sit amet.   ", "Lorem\nipsum,d\nolor\nsit\namet.  ", 7, true},
	}

	for i, test := range tests {
		output := WordWrap(test.input, test.colLen, test.breakLongWords)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}

}

func TestWriteString(t *testing.T) {
	tests := []struct {
		inputStr    string
		initialBuf  []byte
		expectedBuf []byte
		inputLength int
		expectedLen int
	}{
		{" World", []byte("Hello"), []byte("Hello World"), 6, 11},
		{"GoLang", []byte(""), []byte("GoLang"), 6, 6},
		{" String", []byte("Test"), []byte("Test String"), 7, 11},
		{"", []byte(""), []byte(""), 0, 0},
	}

	for i, test := range tests {
		b := &strBuffer{buf: test.initialBuf, length: len(test.initialBuf)}
		b.WriteString(test.inputStr, test.inputLength)
		assert.Equalf(t, test.expectedBuf, b.buf, "Test case %d failed: expected buffer %v, got %v", i, test.expectedBuf, b.buf)
		assert.Equalf(t, test.expectedLen, b.length, "Test case %d failed: expected length %v, got %v", i, test.expectedLen, b.length)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		expectedStr string
		initialBuf  []byte
	}{
		{"Hello", []byte("Hello")},
		{"", []byte("")},
		{"GoLang", []byte("GoLang")},
		{"Test String", []byte("Test String")},
	}

	for i, test := range tests {
		b := &strBuffer{buf: test.initialBuf}
		output := b.String()
		assert.Equalf(t, test.expectedStr, output, "Test case %d failed: expected string %v, got %v", i, test.expectedStr, output)
	}
}

func ExampleWordWrap() {
	fmt.Println(WordWrap("Lorem ipsum, dolor sit amet.", 15, false))
	// Output:
	// Lorem ipsum,
	// dolor sit amet.

}
