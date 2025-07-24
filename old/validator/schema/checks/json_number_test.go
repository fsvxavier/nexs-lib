package checks

import (
	"encoding/json"
	"testing"
)

func TestJsonNumber_IsFormat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "valid json number",
			input: json.Number("123"),
			want:  true,
		},
		{
			name:  "invalid input - string",
			input: "123",
			want:  false,
		},
		{
			name:  "invalid input - int",
			input: 123,
			want:  false,
		},
		{
			name:  "invalid input - nil",
			input: nil,
			want:  false,
		},
	}

	j := JsonNumber{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := j.IsFormat(tt.input); got != tt.want {
				t.Errorf("JsonNumber.IsFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
