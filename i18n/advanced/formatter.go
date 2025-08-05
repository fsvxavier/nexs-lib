package advanced

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/language"
)

var (
	templateCache    sync.Map
	numberCache      sync.Map
	currencyCache    sync.Map
	pluralRulesCache sync.Map
)

type NumberFormat struct {
	DecimalSeparator  string
	ThousandSeparator string
	MinDecimals       int
	MaxDecimals       int
}

type CurrencyFormat struct {
	Symbol        string
	Position      string // "prefix" or "suffix"
	Format        string
	DecimalDigits int
	IncludeSymbol bool
	UseGrouping   bool
	GroupingSep   string
	DecimalSep    string
	GroupingSize  int
	SecondarySize int
}

type PluralRules struct {
	ZeroForm    string
	OneForm     string
	TwoForm     string
	FewForm     string
	ManyForm    string
	OtherForm   string
	PluralFuncs map[string]func(n float64) bool
}

// formatNumber formats a number according to locale rules
func formatNumber(number float64, locale language.Tag) string {
	if format, ok := numberCache.Load(locale.String()); ok {
		return applyNumberFormat(number, format.(NumberFormat))
	}

	// Load format for locale
	format := loadNumberFormat(locale)
	numberCache.Store(locale.String(), format)
	return applyNumberFormat(number, format)
}

// formatCurrency formats a monetary value according to locale rules
func formatCurrency(amount float64, currency string, locale language.Tag) string {
	cacheKey := locale.String() + ":" + currency
	format := loadCurrencyFormat(locale, currency)
	if f, ok := currencyCache.Load(cacheKey); ok {
		format = f.(CurrencyFormat)
	}
	result := applyCurrencyFormat(amount, format)
	return result
}

// getPluralForm returns the correct plural form for a number in a given locale
func getPluralForm(number float64, locale language.Tag) string {
	if rules, ok := pluralRulesCache.Load(locale.String()); ok {
		return applyPluralRules(number, rules.(PluralRules))
	}

	// Load rules for locale
	rules := loadPluralRules(locale)
	pluralRulesCache.Store(locale.String(), rules)
	return applyPluralRules(number, rules)
}

// applyNumberFormat applies formatting rules to a number
func applyNumberFormat(number float64, format NumberFormat) string {
	// Format number with specified decimal places
	str := fmt.Sprintf("%.*f", format.MaxDecimals, number)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	integer := parts[0]

	// Add thousand separators if specified
	if format.ThousandSeparator != "" {
		for i := len(integer) - 3; i > 0; i -= 3 {
			integer = integer[:i] + format.ThousandSeparator + integer[i:]
		}
	}

	// Handle decimal part if present
	if len(parts) > 1 {
		// Trim trailing zeros if under MinDecimals
		decimal := parts[1]
		for len(decimal) > format.MinDecimals && decimal[len(decimal)-1] == '0' {
			decimal = decimal[:len(decimal)-1]
		}
		if len(decimal) > 0 {
			return integer + format.DecimalSeparator + decimal
		}
	}
	return integer
}

// applyCurrencyFormat applies currency formatting rules
func applyCurrencyFormat(amount float64, format CurrencyFormat) string {
	// Format number with fixed decimal places
	str := fmt.Sprintf("%.*f", format.DecimalDigits, amount)
	parts := strings.Split(str, ".")
	integer := parts[0]

	// Add grouping separators if specified
	if format.GroupingSep != "" {
		for i := len(integer) - 3; i > 0; i -= 3 {
			integer = integer[:i] + format.GroupingSep + integer[i:]
		}
	}

	// Build final string with decimal part
	result := integer
	if len(parts) > 1 {
		result = result + format.DecimalSep + parts[1]
	}

	// Add currency symbol
	if format.Symbol != "" {
		if format.Position == "prefix" {
			return format.Symbol + result
		} else if format.Position == "suffix" {
			return result + format.Symbol
		}
	}

	return result
}

// applyPluralRules determines which plural form to use
func applyPluralRules(number float64, rules PluralRules) string {
	// Check each form in order
	if fn, ok := rules.PluralFuncs["zero"]; ok && fn(number) {
		return rules.ZeroForm
	}
	if fn, ok := rules.PluralFuncs["one"]; ok && fn(number) {
		return rules.OneForm
	}
	if fn, ok := rules.PluralFuncs["two"]; ok && fn(number) {
		return rules.TwoForm
	}
	if fn, ok := rules.PluralFuncs["few"]; ok && fn(number) {
		return rules.FewForm
	}
	if fn, ok := rules.PluralFuncs["many"]; ok && fn(number) {
		return rules.ManyForm
	}
	return rules.OtherForm
}

// loadNumberFormat loads number formatting rules for a locale
func loadNumberFormat(locale language.Tag) NumberFormat {
	// Add locale-specific rules
	switch locale.String() {
	case "fr":
		return NumberFormat{
			DecimalSeparator:  ",",
			ThousandSeparator: " ",
			MinDecimals:       2,
			MaxDecimals:       2,
		}
	default:
		return NumberFormat{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			MinDecimals:       2,
			MaxDecimals:       2,
		}
	}
}

// loadCurrencyFormat loads currency formatting rules for a locale
func loadCurrencyFormat(locale language.Tag, currency string) CurrencyFormat {
	// Add locale and currency specific rules
	switch locale.String() {
	case "fr":
		format := CurrencyFormat{
			DecimalDigits: 2,
			IncludeSymbol: true,
			UseGrouping:   true,
			GroupingSep:   " ",
			DecimalSep:    ",",
			GroupingSize:  3,
		}
		if currency == "EUR" {
			format.Symbol = " €"
			format.Position = "suffix"
		} else {
			format.Symbol = currency
			format.Position = "suffix"
		}
		return format
	case "en":
		format := CurrencyFormat{
			DecimalDigits: 2,
			IncludeSymbol: true,
			UseGrouping:   true,
			GroupingSep:   ",",
			DecimalSep:    ".",
			GroupingSize:  3,
		}
		if currency == "USD" {
			format.Symbol = "$"
			format.Position = "prefix"
		} else {
			format.Symbol = currency
			format.Position = "prefix"
		}
		return format
	default:
		return CurrencyFormat{
			Symbol:        "$",
			Position:      "prefix",
			DecimalDigits: 2,
			IncludeSymbol: true,
			UseGrouping:   true,
			GroupingSep:   ",",
			DecimalSep:    ".",
			GroupingSize:  3,
		}
	}
}

// loadPluralRules loads plural rules for a locale
func loadPluralRules(locale language.Tag) PluralRules {
	switch locale.String() {
	case "ar":
		return PluralRules{
			ZeroForm:  "zero",
			OneForm:   "one",
			TwoForm:   "two",
			FewForm:   "few",
			ManyForm:  "many",
			OtherForm: "other",
			PluralFuncs: map[string]func(n float64) bool{
				"zero": func(n float64) bool { return n == 0 },
				"one":  func(n float64) bool { return n == 1 },
				"two":  func(n float64) bool { return n == 2 },
				"few":  func(n float64) bool { return n >= 3 && n <= 10 },
				"many": func(n float64) bool { return n >= 11 && n <= 99 },
			},
		}
	default:
		return PluralRules{
			OneForm:   "one",
			OtherForm: "other",
			PluralFuncs: map[string]func(n float64) bool{
				"one": func(n float64) bool { return n == 1 },
			},
		}
	}
}

// CleanupCaches removes expired entries from all caches
func CleanupCaches() {
	// Clean template cache
	templateCache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(struct {
			template  interface{}
			timestamp time.Time
		}); ok && time.Since(entry.timestamp) > 24*time.Hour {
			templateCache.Delete(key)
		}
		return true
	})

	// Reset all format caches
	numberCache = sync.Map{}
	currencyCache = sync.Map{}
	pluralRulesCache = sync.Map{}
}

func cacheSize() int {
	size := 0
	numberCache.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	return size
}
