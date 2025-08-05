package advanced

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   float64
		locale   language.Tag
		format   NumberFormat
		expected string
	}{
		{
			name:   "simple number",
			number: 1234.56,
			locale: language.English,
			format: NumberFormat{
				DecimalSeparator:  ".",
				ThousandSeparator: ",",
				MinDecimals:       2,
				MaxDecimals:       2,
			},
			expected: "1,234.56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numberCache.Store(tt.locale.String(), tt.format)
			result := formatNumber(tt.number, tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		locale   language.Tag
		format   CurrencyFormat
		expected string
	}{
		{
			name:     "USD amount",
			amount:   1234.56,
			currency: "USD",
			locale:   language.English,
			format: CurrencyFormat{
				Symbol:        "$",
				Position:      "prefix",
				DecimalDigits: 2,
				GroupingSep:   ",",
				DecimalSep:    ".",
			},
			expected: "$1,234.56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheKey := tt.locale.String() + ":" + tt.currency
			currencyCache.Store(cacheKey, tt.format)
			result := formatCurrency(tt.amount, tt.currency, tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPluralForm(t *testing.T) {
	tests := []struct {
		name     string
		number   float64
		locale   language.Tag
		rules    PluralRules
		expected string
	}{
		{
			name:   "singular form",
			number: 1,
			locale: language.English,
			rules: PluralRules{
				OneForm:   "one",
				OtherForm: "other",
				PluralFuncs: map[string]func(n float64) bool{
					"one": func(n float64) bool { return n == 1 },
				},
			},
			expected: "one",
		},
		{
			name:   "plural form",
			number: 2,
			locale: language.English,
			rules: PluralRules{
				OneForm:   "one",
				OtherForm: "other",
				PluralFuncs: map[string]func(n float64) bool{
					"one": func(n float64) bool { return n == 1 },
				},
			},
			expected: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluralRulesCache.Store(tt.locale.String(), tt.rules)
			result := getPluralForm(tt.number, tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupCaches(t *testing.T) {
	// Add some entries to caches
	templateCache.Store("key1", struct {
		template  interface{}
		timestamp time.Time
	}{
		template:  "test",
		timestamp: time.Now().Add(-25 * time.Hour),
	})

	numberCache.Store("key1", NumberFormat{})
	numberCache.Store("key2", NumberFormat{})

	// Run cleanup
	CleanupCaches()

	// Verify template cache cleanup
	_, exists := templateCache.Load("key1")
	assert.False(t, exists, "Old template entry should be removed")

	// Verify size-based cleanup
	size := 0
	numberCache.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	assert.Equal(t, 0, size, "Number cache should be cleared when too large")
}

func BenchmarkFormatNumber(b *testing.B) {
	locale := language.English
	number := 1234.56
	format := NumberFormat{
		DecimalSeparator:  ".",
		ThousandSeparator: ",",
		MinDecimals:       2,
		MaxDecimals:       2,
	}
	numberCache.Store(locale.String(), format)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatNumber(number, locale)
	}
}

func BenchmarkFormatCurrency(b *testing.B) {
	locale := language.English
	amount := 1234.56
	currency := "USD"
	format := CurrencyFormat{
		Symbol:        "$",
		Position:      "prefix",
		DecimalDigits: 2,
	}
	cacheKey := locale.String() + ":" + currency
	currencyCache.Store(cacheKey, format)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatCurrency(amount, currency, locale)
	}
}

func BenchmarkGetPluralForm(b *testing.B) {
	locale := language.English
	number := 1.0
	rules := PluralRules{
		OneForm:   "one",
		OtherForm: "other",
		PluralFuncs: map[string]func(n float64) bool{
			"one": func(n float64) bool { return n == 1 },
		},
	}
	pluralRulesCache.Store(locale.String(), rules)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getPluralForm(number, locale)
	}
}
