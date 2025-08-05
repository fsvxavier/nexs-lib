package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: "811c9dc5", // Expected hash for empty map
		},
		{
			name: "single key",
			input: map[string]interface{}{
				"name": "John",
			},
			expected: hashMap(map[string]interface{}{"name": "John"}),
		},
		{
			name: "multiple keys",
			input: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expected: hashMap(map[string]interface{}{"name": "John", "age": 30}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashMap(tt.input)
			if tt.input == nil || len(tt.input) == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				// For non-empty maps, verify that the hash is consistent
				assert.Equal(t, tt.expected, hashMap(tt.input))
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestGetPluralKey(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:     "singular",
			count:    1,
			expected: "one",
		},
		{
			name:     "plural zero",
			count:    0,
			expected: "other",
		},
		{
			name:     "plural positive",
			count:    2,
			expected: "other",
		},
		{
			name:     "plural negative",
			count:    -1,
			expected: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPluralKey(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkHashMap(b *testing.B) {
	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashMap(data)
	}
}

func BenchmarkGetPluralKey(b *testing.B) {
	counts := []int{0, 1, 2, -1}
	index := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getPluralKey(counts[index%len(counts)])
		index++
	}
}
