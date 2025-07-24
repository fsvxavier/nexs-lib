package decimal

import (
	"fmt"
	"testing"
)

type TestDecimal func(s string) (*Decimal, error)

func TestSub(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("2000000000.00000000"),
			inputB: Ctx.NewFromString("1000000000.00000000"),
			want:   "1000000000.00000000",
		},
		{
			inputA: Ctx.NewFromString("1000000000000.00000002"),
			inputB: Ctx.NewFromString("1000000000000.00000001"),
			want:   "0.00000001",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			inputB: Ctx.NewFromString("0000000000001.00000001"),
			want:   "9999999999997.00000000",
		},
		{
			inputA: Ctx.NewFromString("1.00000001"),
			inputB: Ctx.NewFromString("1.00000001"),
			want:   "0.00000000",
		},
		{
			inputA: Ctx.NewFromString("2.002"),
			inputB: Ctx.NewFromString("1.001"),
			want:   "1.001",
		},
		{
			inputA: Ctx.NewFromString("0.00000002"),
			inputB: Ctx.NewFromString("0.00000001"),
			want:   "0.00000001",
		},
	}

	for _, tt := range tests {
		t.Run("TestSub", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Sub(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Sub() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("1000000000.00000000"),
			inputB: Ctx.NewFromString("1000000000.00000000"),
			want:   "2000000000.00000000",
		},
		{
			inputA: Ctx.NewFromString("1000000000000.00000001"),
			inputB: Ctx.NewFromString("1000000000000.00000001"),
			want:   "2000000000000.00000002",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			inputB: Ctx.NewFromString("0000000000001.00000001"),
			want:   "9999999999999.00000002",
		},
		{
			inputA: Ctx.NewFromString("1.00000001"),
			inputB: Ctx.NewFromString("1.00000001"),
			want:   "2.00000002",
		},
		{
			inputA: Ctx.NewFromString("1.001"),
			inputB: Ctx.NewFromString("1.001"),
			want:   "2.002",
		},
		{
			inputA: Ctx.NewFromString("0.00000001"),
			inputB: Ctx.NewFromString("0.00000001"),
			want:   "0.00000002",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Add(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Add() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("22.00000000"),
			inputB: Ctx.NewFromString("1.00000000"),
			want:   "22.0000000000000000",
		},
		{
			inputA: Ctx.NewFromString("2.00000001"),
			inputB: Ctx.NewFromString("1000000000000.00000001"),
			want:   "2000000010000.00000002",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			inputB: Ctx.NewFromString("0000000000001.00000001"),
			want:   "10000000099997.9999999",
		},
		{
			inputA: Ctx.NewFromString("1.00000001"),
			inputB: Ctx.NewFromString("1.00000001"),
			want:   "1.0000000200000001",
		},
		{
			inputA: Ctx.NewFromString("1.001"),
			inputB: Ctx.NewFromString("1.001"),
			want:   "1.002001",
		},
		{
			inputA: Ctx.NewFromString("0.00000002"),
			inputB: Ctx.NewFromString("2.00000001"),
			want:   "0.0000000400000002",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Mul(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Mul() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestQuo(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("22.00000000"),
			inputB: Ctx.NewFromString("1.00000000"),
			want:   "22.0000000000000000000",
		},
		{
			inputA: Ctx.NewFromString("2.00000001"),
			inputB: Ctx.NewFromString("100.00000001"),
			want:   "0.0200000000979999999902",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			inputB: Ctx.NewFromString("1.00000001"),
			want:   "9999999899998.00100002",
		},
		{
			inputA: Ctx.NewFromString("1.00000001"),
			inputB: Ctx.NewFromString("1.00000001"),
			want:   "1.00000000000000000000",
		},
		{
			inputA: Ctx.NewFromString("2.001"),
			inputB: Ctx.NewFromString("1.001"),
			want:   "1.99900099900099900099",
		},
		{
			inputA: Ctx.NewFromString("4.00000002"),
			inputB: Ctx.NewFromString("2.00000001"),
			want:   "2.00000000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Quo(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Mul() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("1000000000.00000000"),
			want:   "1000000000.00000000",
		},
		{
			inputA: Ctx.NewFromString("1000000000000.00000001"),
			want:   "1000000000000.00000001",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			want:   "9999999999998.00000001",
		},
		{
			inputA: Ctx.NewFromString("-1.00000001"),
			want:   "1.00000001",
		},
		{
			inputA: Ctx.NewFromString("-1.001"),
			want:   "1.001",
		},
		{
			inputA: Ctx.NewFromString("0.00000001"),
			want:   "0.00000001",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Abs(got, tt.inputA)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Abs() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: Ctx.NewFromString("1000000000.00000000"),
			want:   "-1000000000.00000000",
		},
		{
			inputA: Ctx.NewFromString("1000000000000.00000001"),
			want:   "-1000000000000.00000001",
		},
		{
			inputA: Ctx.NewFromString("9999999999998.00000001"),
			want:   "-9999999999998.00000001",
		},
		{
			inputA: Ctx.NewFromString("1.00000001"),
			want:   "-1.00000001",
		},
		{
			inputA: Ctx.NewFromString("1.001"),
			want:   "-1.001",
		},
		{
			inputA: Ctx.NewFromString("0.00000001"),
			want:   "-0.00000001",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := Ctx.NewFromString("0")
			err := Ctx.Neg(got, tt.inputA)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Neg() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: Ctx.NewFromString("100.54"),
			want:  "100.54",
		},
		{
			input: Ctx.NewFromString("100.5"),
			want:  "100.5",
		},
		{
			input: Ctx.NewFromString("10999.39"),
			want:  "10999.39",
		},
		{
			input: Ctx.NewFromString("10999.39989"),
			want:  "10999.39989",
		},
		{
			input: Ctx.NewFromString("100000009.50000000"),
			want:  "100000009.50000000",
		},
		{
			input: Ctx.NewFromString("100.499"),
			want:  "100.499",
		},
		{
			input: Ctx.NewFromString("100.5000000099"),
			want:  "100.50000000",
		},
		{
			input: Ctx.NewFromString("99999999.999999999"),
			want:  "99999999.99999999",
		},
	}

	for _, tt := range tests {
		t.Run("TestTruncate", func(t *testing.T) {
			truncated := tt.input.Truncate(MAX_PRECISION, MIN_EXPONENT, DEFAULT_ROUNDING)

			if truncated.String() != tt.want {
				t.Errorf("Decimal.Truncate() = %v, want %v", truncated.String(), tt.want)
			}
		})
	}
}

func TestNewFromFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{
			input: 10000000.00000001,
			want:  "10000000.00000001",
		},
		{
			input: 1.001,
			want:  "1.001",
		},
		{
			input: -1.00000001,
			want:  "-1.00000001",
		},
		{
			input: 99999999998.0,
			want:  "99999999998",
		},
		{
			input: 0.0,
			want:  "0",
		},
		{
			input: 100.54,
			want:  "100.54",
		},
	}

	for _, tt := range tests {
		t.Run("TestNewFromFloat", func(t *testing.T) {
			got := Ctx.NewFromFloat(tt.input)

			if got == nil || got.String() != tt.want {
				t.Errorf("Decimal.NewFromFloat(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewFromInt(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{
			input: 10000000,
			want:  "10000000",
		},
		{
			input: 1,
			want:  "1",
		},
		{
			input: -1,
			want:  "-1",
		},
		{
			input: 99999999998,
			want:  "99999999998",
		},
		{
			input: 0,
			want:  "0",
		},
		{
			input: 9223372036854775807, // max int64
			want:  "9223372036854775807",
		},
		{
			input: -9223372036854775808, // min int64
			want:  "-9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run("TestNewFromInt", func(t *testing.T) {
			got := Ctx.NewFromInt(tt.input)

			if got == nil || got.String() != tt.want {
				t.Errorf("Decimal.NewFromInt(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestComparison(t *testing.T) {
	tests := []struct {
		a   string
		b   string
		eq  bool
		gt  bool
		lt  bool
		gte bool
		lte bool
	}{
		{
			a:   "100.00",
			b:   "100.00",
			eq:  true,
			gt:  false,
			lt:  false,
			gte: true,
			lte: true,
		},
		{
			a:   "100.01",
			b:   "100.00",
			eq:  false,
			gt:  true,
			lt:  false,
			gte: true,
			lte: false,
		},
		{
			a:   "99.99",
			b:   "100.00",
			eq:  false,
			gt:  false,
			lt:  true,
			gte: false,
			lte: true,
		},
		{
			a:   "0.00000001",
			b:   "0.00000001",
			eq:  true,
			gt:  false,
			lt:  false,
			gte: true,
			lte: true,
		},
		{
			a:   "-1.00000001",
			b:   "1.00000001",
			eq:  false,
			gt:  false,
			lt:  true,
			gte: false,
			lte: true,
		},
		{
			a:   "9999999999999.99999999",
			b:   "9999999999999.99999998",
			eq:  false,
			gt:  true,
			lt:  false,
			gte: true,
			lte: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.a, tt.b), func(t *testing.T) {
			a := Ctx.NewFromString(tt.a)
			b := Ctx.NewFromString(tt.b)

			if got := Ctx.IsEqual(a, b); got != tt.eq {
				t.Errorf("IsEqual(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.eq)
			}

			if got := Ctx.IsGreaterThan(a, b); got != tt.gt {
				t.Errorf("IsGreaterThan(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.gt)
			}

			if got := Ctx.IsLessThan(a, b); got != tt.lt {
				t.Errorf("IsLessThan(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.lt)
			}

			if got := Ctx.IsGreaterThanOrEqual(a, b); got != tt.gte {
				t.Errorf("IsGreaterThanOrEqual(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.gte)
			}

			if got := Ctx.IsLessThanOrEqual(a, b); got != tt.lte {
				t.Errorf("IsLessThanOrEqual(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.lte)
			}
		})
	}
}
