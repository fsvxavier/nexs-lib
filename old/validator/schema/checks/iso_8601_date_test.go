package checks

import "testing"

func TestIsISO8601Date(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "valid ISO8601 date",
			input: "2023-05-10T15:04:05.999Z",
			want:  true,
		},
		{
			name:  "valid ISO8601 date with timezone",
			input: "2023-05-10T15:04:05.999-07:00",
			want:  true,
		},
		{
			name:  "invalid date format",
			input: "2023/05/10",
			want:  false,
		},
		{
			name:  "invalid string",
			input: "not a date",
			want:  false,
		},
		{
			name:  "non-string input",
			input: 123,
			want:  false,
		},
		{
			name:  "nil input",
			input: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsISO8601Date(tt.input); got != tt.want {
				t.Errorf("IsISO8601Date() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestIso8601Date_IsFormat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "valid ISO8601 date",
			input: "2023-05-10T15:04:05.999Z",
			want:  true,
		},
		{
			name:  "valid ISO8601 date with timezone",
			input: "2023-05-10T15:04:05.999-07:00",
			want:  true,
		},
		{
			name:  "invalid date format",
			input: "2023/05/10",
			want:  false,
		},
		{
			name:  "invalid string",
			input: "not a date",
			want:  false,
		},
		{
			name:  "non-string input",
			input: 123,
			want:  false,
		},
		{
			name:  "nil input",
			input: nil,
			want:  false,
		},
	}

	validator := Iso8601Date{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.IsFormat(tt.input); got != tt.want {
				t.Errorf("Iso8601Date.IsFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
