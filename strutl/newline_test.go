package strutl

import (
	"fmt"
	"runtime"
	"testing"
)

func TestOSNewLine(t *testing.T) {
	expected := "\n"
	if runtime.GOOS == "windows" {
		expected = "\r\n"
	}

	result := OSNewLine()
	if result != expected {
		t.Errorf("OSNewLine() = %q; want %q", result, expected)
	}
}

func ExampleOSNewLine() {
	// Nota: a saída depende do sistema operacional
	fmt.Println("Nova linha:" + OSNewLine() + "segunda linha")
	// No Linux/Mac:
	// Output:
	// Nova linha:
	// segunda linha
}

// Benchmark de OSNewLine
func BenchmarkOSNewLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = OSNewLine()
	}
}
