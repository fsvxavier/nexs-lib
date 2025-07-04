package apd

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
)

func TestApdProvider(t *testing.T) {
	provider := NewProvider()

	t.Run("NewFromString", func(t *testing.T) {
		dec, err := provider.NewFromString("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())

		// Teste com valor inválido
		_, err = provider.NewFromString("invalid")
		assert.Error(t, err)
	})

	t.Run("NewFromFloat", func(t *testing.T) {
		dec, err := provider.NewFromFloat(123.456)
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())
	})

	t.Run("NewFromInt", func(t *testing.T) {
		dec, err := provider.NewFromInt(123)
		assert.NoError(t, err)
		assert.Equal(t, "123", dec.String())
	})

	t.Run("NewProviderWithContext", func(t *testing.T) {
		provider := NewProviderWithContext(10, 5, -5, apd.RoundHalfUp)
		dec, err := provider.NewFromString("123.456789")
		assert.NoError(t, err)

		// Usando o contexto para arredondar
		rounded := dec.Round(3)
		assert.Equal(t, "123.457", rounded.String())
	})
}

func TestApdDecimal(t *testing.T) {
	provider := NewProvider()

	t.Run("String", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")
		assert.Equal(t, "123.456", dec.String())
	})

	t.Run("Float64", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")
		val, err := dec.Float64()
		assert.NoError(t, err)
		assert.Equal(t, 123.456, val)
	})

	t.Run("Int64", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")
		val, err := dec.Int64()
		assert.NoError(t, err)
		assert.Equal(t, int64(123), val)

		// Teste com valor grande
		dec, _ = provider.NewFromString("9223372036854775807") // max int64
		val, err = dec.Int64()
		assert.NoError(t, err)
		assert.Equal(t, int64(9223372036854775807), val)
	})

	t.Run("IsZero", func(t *testing.T) {
		dec, _ := provider.NewFromString("0")
		assert.True(t, dec.IsZero())

		dec, _ = provider.NewFromString("1")
		assert.False(t, dec.IsZero())
	})

	t.Run("IsNegative", func(t *testing.T) {
		dec, _ := provider.NewFromString("-1")
		assert.True(t, dec.IsNegative())

		dec, _ = provider.NewFromString("1")
		assert.False(t, dec.IsNegative())
	})

	t.Run("IsPositive", func(t *testing.T) {
		dec, _ := provider.NewFromString("1")
		assert.True(t, dec.IsPositive())

		dec, _ = provider.NewFromString("-1")
		assert.False(t, dec.IsPositive())
	})

	t.Run("Equals", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("123.456")
		dec3, _ := provider.NewFromString("123.457")

		assert.True(t, dec1.Equals(dec2))
		assert.False(t, dec1.Equals(dec3))
	})

	t.Run("GreaterThan", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.457")
		dec2, _ := provider.NewFromString("123.456")

		assert.True(t, dec1.GreaterThan(dec2))
		assert.False(t, dec2.GreaterThan(dec1))
	})

	t.Run("LessThan", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("123.457")

		assert.True(t, dec1.LessThan(dec2))
		assert.False(t, dec2.LessThan(dec1))
	})

	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("123.456")
		dec3, _ := provider.NewFromString("123.455")

		assert.True(t, dec1.GreaterThanOrEqual(dec2))
		assert.True(t, dec1.GreaterThanOrEqual(dec3))
		assert.False(t, dec3.GreaterThanOrEqual(dec1))
	})

	t.Run("LessThanOrEqual", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("123.456")
		dec3, _ := provider.NewFromString("123.457")

		assert.True(t, dec1.LessThanOrEqual(dec2))
		assert.True(t, dec1.LessThanOrEqual(dec3))
		assert.False(t, dec3.LessThanOrEqual(dec1))
	})

	t.Run("Add", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("10.544")

		result := dec1.Add(dec2)
		assert.Equal(t, "134", result.String()[:3])
	})

	t.Run("Sub", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("23.456")

		result := dec1.Sub(dec2)
		assert.Equal(t, "100", result.String())
	})

	t.Run("Mul", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("2")

		result := dec1.Mul(dec2)
		assert.Equal(t, "246.912", result.String())
	})

	t.Run("Div", func(t *testing.T) {
		dec1, _ := provider.NewFromString("123.456")
		dec2, _ := provider.NewFromString("2")

		result, err := dec1.Div(dec2)
		assert.NoError(t, err)
		assert.Equal(t, "61.728", result.String())

		// Teste de divisão por zero
		dec3, _ := provider.NewFromString("0")
		_, err = dec1.Div(dec3)
		assert.Error(t, err)
	})

	t.Run("Abs", func(t *testing.T) {
		dec, _ := provider.NewFromString("-123.456")
		result := dec.Abs()
		assert.Equal(t, "123.456", result.String())
	})

	t.Run("Round", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456789")
		result := dec.Round(2)
		assert.Equal(t, "123.46", result.String())
	})

	t.Run("Truncate", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456789")
		result := dec.Truncate(2)
		assert.Equal(t, "123.45", result.String())
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")
		data, err := dec.MarshalJSON()
		assert.NoError(t, err)

		var result string
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, "123.456", result)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		dec, _ := provider.NewFromString("0")
		err := dec.UnmarshalJSON([]byte(`"123.456"`))
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())

		// Teste com JSON number
		dec, _ = provider.NewFromString("0")
		err = dec.UnmarshalJSON([]byte(`123.456`))
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())
	})
}

// Teste de race condition
func TestRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	provider := NewProvider()
	dec, _ := provider.NewFromString("100")

	var wg sync.WaitGroup
	iterations := 1000

	// Teste concorrente de operações Add
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			addend, _ := provider.NewFromString("1")
			dec.Add(addend)
		}()
	}
	wg.Wait()

	// Teste concorrente de operações Sub
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			subtrahend, _ := provider.NewFromString("1")
			dec.Sub(subtrahend)
		}()
	}
	wg.Wait()
}

func BenchmarkApdNewFromString(b *testing.B) {
	provider := NewProvider()
	for i := 0; i < b.N; i++ {
		provider.NewFromString("123.456789")
	}
}

func BenchmarkApdNewFromFloat(b *testing.B) {
	provider := NewProvider()
	for i := 0; i < b.N; i++ {
		provider.NewFromFloat(123.456789)
	}
}

func BenchmarkApdAdd(b *testing.B) {
	provider := NewProvider()
	dec1, _ := provider.NewFromString("123.456789")
	dec2, _ := provider.NewFromString("876.543211")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec1.Add(dec2)
	}
}

func BenchmarkApdMul(b *testing.B) {
	provider := NewProvider()
	dec1, _ := provider.NewFromString("123.456789")
	dec2, _ := provider.NewFromString("876.543211")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec1.Mul(dec2)
	}
}

func BenchmarkApdDiv(b *testing.B) {
	provider := NewProvider()
	dec1, _ := provider.NewFromString("123.456789")
	dec2, _ := provider.NewFromString("876.543211")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec1.Div(dec2)
	}
}
