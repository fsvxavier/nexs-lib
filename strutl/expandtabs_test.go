package strutl

import (
	"fmt"
	"testing"
)

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		count    int
	}{
		{"Empty string", "", "", 2},
		{"Zero count", "\t", "", 0},
		{"Multiple tabs with newlines", "\t\n\t\n", "  \n  \n", 2},
		{"Adjacent tabs", "\t\t", "    ", 2},
		{"Tabs with content", "\tlorem\n\tipsum\n", "  lorem\n  ipsum\n", 2},
		{"Different tab width", "\ttest", "    test", 4},
		{"Mixed tabs and spaces", "  \ttest  \t  ", "    test      ", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExpandTabs(test.input, test.count)
			if result != test.expected {
				t.Errorf("ExpandTabs(%q, %d) = %q; want %q",
					test.input, test.count, result, test.expected)
			}
		})
	}
}

func ExampleExpandTabs() {
	fmt.Printf("%s", ExpandTabs("\tlorem\n\tipsum", 2))
	// Output:
	//   lorem
	//   ipsum
}

// Benchmarks
func BenchmarkExpandTabs(b *testing.B) {
	benchCases := []struct {
		name  string
		input string
		count int
	}{
		{"SingleTab", "\thello", 2},
		{"MultipleTabs", "\t\t\t\t\t", 4},
		{"Mixed", "line\twith\ttabs\n\tindented\tline", 2},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ExpandTabs(bc.input, bc.count)
			}
		})
	}
}
