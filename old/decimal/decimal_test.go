package decimal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromIntDec(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "positive integer",
			input:    12345,
			expected: "12345",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "negative integer",
			input:    -9876,
			expected: "-9876",
		},
		{
			name:     "large number",
			input:    9223372036854775807, // max int64
			expected: "9223372036854775807",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFromInt(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestDecimal_IsEqualFromString(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("TestDecimal_IsEqualFromString", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsEqual(y); got != tt.want {
				t.Errorf("Decimal.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsEqualFromFloat(t *testing.T) {
	tests := []struct {
		x    float64
		y    float64
		want bool
	}{
		{
			x:    1.0,
			y:    1.0,
			want: true,
		},
		{
			x:    1.0,
			y:    2.0,
			want: false,
		},
		{
			x:    1.0,
			y:    0.000000000000000001,
			want: false,
		},
		{
			x:    0.000000000000000002,
			y:    0.000000000000000002,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("TestDecimal_IsEqualFromFloat", func(t *testing.T) {
			x := NewFromFloat(tt.x)
			y := NewFromFloat(tt.y)

			if got := x.IsEqual(y); got != tt.want {
				t.Errorf("Decimal.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsGreaterThan(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("IsGreaterThan", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsGreaterThan(y); got != tt.want {
				t.Errorf("Decimal.IsGreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsLessThan(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: true,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("IsLessThan", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsLessThan(y); got != tt.want {
				t.Errorf("Decimal.IsLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: true,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: false,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("IsGreaterThanOrEqual", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsGreaterThanOrEqual(y); got != tt.want {
				t.Errorf("Decimal.IsGreaterThanOrEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsLessThanOrEqual(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: true,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("IsLessThanOrEqual", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsLessThanOrEqual(y); got != tt.want {
				t.Errorf("Decimal.IsLessThanOrEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_UnmarshalJSON(t *testing.T) {
	type Out struct {
		Amount *Decimal `json:"amount"`
	}

	tests := []struct {
		want  *Decimal
		input string
	}{
		{
			input: `{"amount":100.40}`,
			want:  NewFromString("100.40"),
		},
		{
			input: `{"amount":314.16}`,
			want:  NewFromString("314.16"),
		},
		{
			input: `{"amount":-14.1}`,
			want:  NewFromString("-14.1"),
		},
		{
			input: `{"amount":0.00}`,
			want:  NewFromString("0.00"),
		},
		{
			input: `{"amount":0.99}`,
			want:  NewFromString("0.99"),
		},
		{
			input: `{"amount":0.000000000000000001}`,
			want:  NewFromString("0.000000000000000001"),
		},
		{
			input: `{"amount":0.55}`,
			want:  NewFromString("0.55"),
		},
		{
			input: `{"amount":0.00000001}`,
			want:  NewFromString("0.00000001"),
		},
	}
	for _, tt := range tests {
		t.Run("TestDecimal_UnmarshalJSON", func(t *testing.T) {
			var out Out
			err := Unmarshal([]byte(tt.input), &out)
			if err != nil {
				t.Errorf("Decimal.UnmarshalJSON() error = %v", err)
				return
			}

			if !out.Amount.IsEqual(tt.want) {
				t.Errorf("Decimal.UnmarshalJSON() = %v, want %v", out.Amount, tt.want)
			}
		})
	}
}
func TestDecimal_MarshalJSON(t *testing.T) {
	type Out struct {
		Amount *Decimal `json:"amount"`
	}

	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: NewFromString("100.40"),
			want:  `{"amount":100.40}`,
		},
		{
			input: NewFromString("314.16"),
			want:  `{"amount":314.16}`,
		},
		{
			input: NewFromString("-14.1"),
			want:  `{"amount":-14.1}`,
		},
		{
			input: NewFromString("0.00"),
			want:  `{"amount":0.00}`,
		},
		{
			input: NewFromString("0.99"),
			want:  `{"amount":0.99}`,
		},
		{
			input: NewFromString("0.000000000000000001"),
			want:  `{"amount":0.00000000}`,
		},
		{
			input: NewFromString("0.55000000"),
			want:  `{"amount":0.55000000}`,
		},
		{
			input: NewFromString("0.00000001"),
			want:  `{"amount":0.00000001}`,
		},
		{
			input: NewFromString("0.00000000"),
			want:  `{"amount":0.00000000}`,
		},
		{
			input: NewFromString("0"),
			want:  `{"amount":0}`,
		},
		{
			input: NewFromString("1000000000000.00000000"),
			want:  `{"amount":1000000000000.00000000}`,
		},
	}
	for _, tt := range tests {
		t.Run("TestDecimal_MarshalJSON", func(t *testing.T) {
			out := &Out{
				Amount: tt.input,
			}
			bout := Marshal(out)

			if tt.want != string(bout) {
				t.Errorf("Decimal.MarshalJSON() = %v, want %v, got %v", out.Amount, tt.want, string(bout))
			}
		})
	}
}

func Test_TruncateTwoDecimalPlaces(t *testing.T) {
	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: NewFromString("1200"),
			want:  "1200",
		},
		{
			input: NewFromString("12.00"),
			want:  "12.00",
		},
		{
			input: NewFromString("1299.88882"),
			want:  "1299.88",
		},
		{
			input: NewFromString("99999999.999999999"),
			want:  "99999999.99",
		},
		{
			input: NewFromString("99999999.9299999999"),
			want:  "99999999.92",
		},
		{
			input: NewFromString("1"),
			want:  "1",
		},
		{
			input: NewFromString("123.000329"),
			want:  "123.00",
		},
		{
			input: NewFromString("123.4599"),
			want:  "123.45",
		},
		{
			input: NewFromString("123.998"),
			want:  "123.99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Truncate(tt.input, -2)
			fmt.Println(got)
			if got.Text('f') != tt.want {
				t.Errorf("TruncateRaw() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TruncateTwoDecimalPlacesTrim(t *testing.T) {
	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: NewFromString("1200"),
			want:  "1200",
		},
		{
			input: NewFromString("12.00"),
			want:  "12",
		},
		{
			input: NewFromString("1299.88882"),
			want:  "1299.88",
		},
		{
			input: NewFromString("99999999.999999999"),
			want:  "99999999.99",
		},
		{
			input: NewFromString("99999999.9299999999"),
			want:  "99999999.92",
		},
		{
			input: NewFromString("1"),
			want:  "1",
		},
		{
			input: NewFromString("123.000329"),
			want:  "123",
		},
		{
			input: NewFromString("123.4599"),
			want:  "123.45",
		},
		{
			input: NewFromString("123.998"),
			want:  "123.99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := TrimZerosRight(Truncate(tt.input, -2))
			fmt.Println(got)
			if got.Text('f') != tt.want {
				t.Errorf("TruncateRaw() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TruncateTwoDecimalPlacesTrimPointer(t *testing.T) {
	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: NewFromString("1200"),
			want:  "1200",
		},
		{
			input: NewFromString("12.00"),
			want:  "12",
		},
		{
			input: NewFromString("1299.88882"),
			want:  "1299.88",
		},
		{
			input: NewFromString("99999999.999999999"),
			want:  "99999999.99",
		},
		{
			input: NewFromString("99999999.9299999999"),
			want:  "99999999.92",
		},
		{
			input: NewFromString("1"),
			want:  "1",
		},
		{
			input: NewFromString("123.000329"),
			want:  "123",
		},
		{
			input: NewFromString("123.4599"),
			want:  "123.45",
		},
		{
			input: NewFromString("123.998"),
			want:  "123.99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.Truncate(MAX_EXPONENT, -2, DEFAULT_ROUNDING).TrimZerosRight()
			fmt.Println(got)
			if got.Text('f') != tt.want {
				t.Errorf("TruncateRaw() got %v, want %v", got, tt.want)
			}
		})
	}
}
