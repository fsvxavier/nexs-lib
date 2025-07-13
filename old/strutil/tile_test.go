package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTile(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
		length   int
	}{
		{"", "", 10},
		{"-", "----------", 10},
		{"-.", "", -1},
		{"-.", "", 0},
		{"-.", "-.", 2},
		{"-.", "-", 1},
		{"-.", "-.-", 3},
		{"-.", "-.-.", 4},
		{"-৹", "", 0},
		{"-৹", "-৹", 2},
		{"-৹", "-", 1},
		{"-৹", "-৹-", 3},
		{"-৹", "-৹-৹", 4},
		{".-", ".-.-", 4},
		{".-=", ".-=.", 4},
		{".-", ".", 1},
		{"-৹_৹", "-৹_৹-৹_৹-৹_৹-৹_৹", 16},
		{"-৹_৹", "-৹_৹-৹_৹-৹_৹-৹_৹-", 17},
		{"-৹_৹", "-৹_৹-৹_৹-৹_৹-৹_৹-৹", 18},
		{"-৹_৹", "-৹_৹-৹_৹-৹_৹-৹_৹-৹_", 19},
		{"-৹_৹", "-৹_৹-৹_৹-৹_৹-৹_৹-৹_৹", 20},
	}

	for i, test := range tests {
		output := Tile(test.pattern, test.length)
		assert.Equalf(t, test.expected, output, "Test case %d is not successful\n", i)
	}
}

func BenchmarkTile(b *testing.B) {
	var s string
	for n := 0; n < b.N; n++ {
		s = Tile("-.", 10)
	}
	_ = s
}
