package strutil

import (
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
