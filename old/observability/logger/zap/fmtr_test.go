package zap

import (
	"testing"
)

func Test_formatFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "should return full path when no exchange pattern found",
			filename: "/home/user/project/file.go",
			want:     "/home/user/project/file.go",
		},
		{
			name:     "should return only exchange path when pattern found",
			filename: "/home/user/project/exchange/service/file.go",
			want:     "exchange/service/file.go",
		},
		{
			name:     "should handle empty string",
			filename: "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFilename(tt.filename); got != tt.want {
				t.Errorf("formatFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
