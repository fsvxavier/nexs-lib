# Advanced Formatting Features

This package provides advanced formatting features for the i18n package, including:

- Number formatting
- Currency formatting
- Plural rules

## Number Formatting

```go
// Format a number according to locale rules
formatted := formatNumber(1234567.89, language.English)
// Output: 1,234,567.89

// French locale
formatted = formatNumber(1234567.89, language.French)
// Output: 1 234 567,89
```

Number formatting supports:
- Configurable decimal and thousand separators
- Minimum and maximum decimal places
- Per-locale formatting rules

## Currency Formatting

```go
// Format amount in US dollars
formatted := formatCurrency(1234.56, "USD", language.English)
// Output: $1,234.56

// Format amount in Euros
formatted = formatCurrency(1234.56, "EUR", language.French)
// Output: 1 234,56 €
```

Currency formatting supports:
- Currency symbols
- Position (prefix/suffix)
- Grouping separators
- Decimal digits
- Per-locale and per-currency formatting rules

## Plural Rules

```go
// Get plural form for English
form := getPluralForm(1, language.English)  // "one"
form = getPluralForm(2, language.English)   // "other"

// Arabic (more complex rules)
form = getPluralForm(0, language.Arabic)    // "zero"
form = getPluralForm(1, language.Arabic)    // "one"
form = getPluralForm(2, language.Arabic)    // "two"
form = getPluralForm(5, language.Arabic)    // "few"
```

Plural rules support:
- Language-specific plural forms
- Zero, one, two, few, many, other forms
- Custom plural functions per locale

## Performance Optimizations

The package includes several optimizations:
- Thread-safe caching using `sync.Map`
- TTL-based cache expiration
- Maximum cache size limits
- Pre-compiled plural rules
- Optimized number and currency formatters

## Using the Cache

The package automatically handles caching of:
- Number formats
- Currency formats
- Plural rules
- Templates (with TTL)

Cache cleanup is automatic but can be triggered manually:

```go
CleanupCaches() // Removes expired entries and enforces size limits
```

## Thread Safety

All operations are thread-safe and can be used concurrently. The package uses `sync.Map` internally to ensure thread-safe access to cached data.
