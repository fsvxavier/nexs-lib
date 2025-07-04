package strutl

import (
	"fmt"
	"testing"
)

func TestWordwrap(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected       string
		colLen         int
		breakLongWords bool
	}{
		{"Zero column length", "Lorem ipsum dolor sit amet", "Lorem ipsum dolor sit amet", 0, false},
		{"With newlines", "Lorem ipsum\ndolor\nsit amet", "Lorem ipsum\ndolor\nsit amet", 15, false},
		{"Basic wrapping", "Lorem ipsum dolor sit amet", "Lorem ipsum\ndolor sit amet", 15, false},
		{"With spaces and newlines", "Lorem ipsum     \n   dolor sit amet\n    consectetur", "Lorem ipsum    \n\n   dolor sit\namet\n    consectetur", 15, false},
		{"With punctuation", "Lorem ipsum, dolor sit amet.", "Lorem ipsum,\ndolor sit amet.", 15, false},
		{"Ending with newline", "Lorem ipsum, dolor sit amet.\n", "Lorem ipsum,\ndolor sit amet.\n", 15, false},
		{"Starting with newline", "\nLorem ipsum, dolor sit amet.\n", "\nLorem ipsum,\ndolor sit amet.\n", 15, false},
		{"With indentation", "\n   Lorem ipsum, dolor sit amet.\n", "\n   Lorem ipsum,\ndolor sit amet.\n", 15, false},
		{"Single character width", "Lorem ipsum, dolor sit amet", "Lorem\nipsum,\ndolor\nsit\namet", 1, false},
		{"Break long words enabled", "Lorem ipsum, dolor sit amet", "L\no\nr\ne\nm\ni\np\ns\nu\nm\n,\nd\no\nl\no\nr\ns\ni\nt\na\nm\ne\nt", 1, true},
		{"Break long words with width 6", "Loremipsum, dolorsitamet", "Loremi\npsum,\ndolors\nitamet", 6, true},
		{"Break long words with width 3", "Lorem ipsum, dolor sit amet", "Lor\nem\nips\num,\ndol\nor\nsit\name\nt", 3, true},
		{"Unicode support", "Το Lorem Ipsum είναι απλά ένα κείμενο χωρίς", "Το Lorem Ipsum\nείναι απλά ένα\nκείμενο χωρίς", 15, false},
		{"Empty string", "", "", 15, false},
		{"Only spaces", "                        ", "               \n        ", 15, false},
		{"With multiple spaces", "Lorem ipsum   dolor sit amet", "Lorem ipsum  \ndolor sit amet", 15, false},
		{"Spaces after comma", "Lorem ipsum,   dolor sit amet.", "Lorem ipsum,  \ndolor sit amet.", 15, false},
		{"Leading spaces", "   Lorem ipsum,   dolor sit amet.", "   Lorem ipsum,\n  dolor sit\namet.", 15, false},
		{"No spaces after comma", "Lorem ipsum,dolor sit amet.", "Lorem\nipsum,dolor sit\namet.", 15, false},
		{"Trailing spaces", "Lorem ipsum,dolor sit amet.   ", "Lorem\nipsum,dolor sit\namet.   ", 15, false},
		{"Break long with spaces", "Lorem ipsum,dolor sit amet.   ", "Lorem\nipsum,d\nolor\nsit\namet.  ", 7, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := WordWrap(test.input, test.colLen, test.breakLongWords)
			if result != test.expected {
				t.Errorf("WordWrap(%q, %d, %v) = %q; want %q",
					test.input, test.colLen, test.breakLongWords, result, test.expected)
			}
		})
	}
}

func TestStrBuffer_WriteString(t *testing.T) {
	tests := []struct {
		name        string
		inputStr    string
		initialBuf  []byte
		expectedBuf []byte
		inputLength int
		expectedLen int
	}{
		{"Append to existing", " World", []byte("Hello"), []byte("Hello World"), 6, 11},
		{"New buffer", "GoLang", []byte(""), []byte("GoLang"), 6, 6},
		{"Another append", " String", []byte("Test"), []byte("Test String"), 7, 11},
		{"Empty string", "", []byte(""), []byte(""), 0, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := &strBuffer{buf: test.initialBuf, length: len(test.initialBuf)}
			b.WriteString(test.inputStr, test.inputLength)

			// Verificar o conteúdo do buffer
			if string(b.buf) != string(test.expectedBuf) {
				t.Errorf("Buffer = %q; want %q", string(b.buf), string(test.expectedBuf))
			}

			// Verificar o tamanho
			if b.length != test.expectedLen {
				t.Errorf("Length = %d; want %d", b.length, test.expectedLen)
			}
		})
	}
}

func TestStrBuffer_String(t *testing.T) {
	tests := []struct {
		name        string
		expectedStr string
		initialBuf  []byte
	}{
		{"Regular string", "Hello", []byte("Hello")},
		{"Empty string", "", []byte("")},
		{"GoLang", "GoLang", []byte("GoLang")},
		{"With space", "Test String", []byte("Test String")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := &strBuffer{buf: test.initialBuf}
			output := b.String()
			if output != test.expectedStr {
				t.Errorf("String() = %q; want %q", output, test.expectedStr)
			}
		})
	}
}

// Exemplos para documentação
func ExampleWordWrap() {
	fmt.Println(WordWrap("Lorem ipsum, dolor sit amet.", 15, false))
	// Output:
	// Lorem ipsum,
	// dolor sit amet.
}

// Benchmark de WordWrap
func BenchmarkWordWrap(b *testing.B) {
	benchCases := []struct {
		name          string
		input         string
		colLen        int
		breakLongWord bool
	}{
		{"ShortText", "Lorem ipsum dolor sit amet", 20, false},
		{"MediumText", "Lorem ipsum dolor sit amet, consectetur adipiscing elit", 15, false},
		{"LongText", "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 10, false},
		{"BreakLongWords", "Loremipsumdolorsitametconsecteturadipiscing", 10, true},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = WordWrap(bc.input, bc.colLen, bc.breakLongWord)
			}
		})
	}
}
