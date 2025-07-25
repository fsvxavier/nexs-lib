// Package duration provides enhanced duration parsing functionality.
package duration

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

// Additional durations, a day is considered to be 24 hours.
const (
	Day  time.Duration = time.Hour * 24
	Week               = Day * 7
)

// ParsedDuration represents a parsed duration with metadata.
type ParsedDuration struct {
	Duration time.Duration
	Original string
	Units    []string // units found in the string (ns, us, ms, s, m, h, d, w)
}

// Parser implements duration parsing with enhanced functionality.
type Parser struct {
	config  *interfaces.ParserConfig
	unitMap map[string]int64
}

// NewParser creates a new duration parser with default configuration.
func NewParser() *Parser {
	return &Parser{
		config: interfaces.DefaultConfig(),
		unitMap: map[string]int64{
			"ns": int64(time.Nanosecond),
			"us": int64(time.Microsecond),
			"µs": int64(time.Microsecond), // U+00B5 = micro symbol
			"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
			"ms": int64(time.Millisecond),
			"s":  int64(time.Second),
			"m":  int64(time.Minute),
			"h":  int64(time.Hour),
			"d":  int64(Day),
			"w":  int64(Week),
		},
	}
}

// NewParserWithConfig creates a new duration parser with custom configuration.
func NewParserWithConfig(config *interfaces.ParserConfig) *Parser {
	parser := NewParser()
	parser.config = config
	return parser
}

// Parse implements interfaces.Parser.
func (p *Parser) Parse(ctx context.Context, data []byte) (*ParsedDuration, error) {
	return p.ParseString(ctx, string(data))
}

// ParseString parses a duration string with enhanced unit support.
func (p *Parser) ParseString(ctx context.Context, input string) (*ParsedDuration, error) {
	if err := p.validateInput(input); err != nil {
		return nil, err
	}

	original := input
	input = strings.TrimSpace(input)

	// Try standard library first
	if d, err := time.ParseDuration(input); err == nil {
		return &ParsedDuration{
			Duration: d,
			Original: original,
			Units:    p.extractUnits(input),
		}, nil
	}

	// Enhanced parsing with days and weeks
	duration, err := p.parseEnhanced(input)
	if err != nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: fmt.Sprintf("unable to parse duration: %s", original),
			Cause:   err,
		}
	}

	return &ParsedDuration{
		Duration: duration,
		Original: original,
		Units:    p.extractUnits(input),
	}, nil
}

// parseEnhanced performs enhanced duration parsing with days and weeks.
func (p *Parser) parseEnhanced(s string) (time.Duration, error) {
	orig := s
	var d int64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, fmt.Errorf("time: invalid duration %q", orig)
	}

	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, fmt.Errorf("time: invalid duration %q", orig)
		}

		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, fmt.Errorf("time: invalid duration %q", orig)
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, s, err = leadingInt(s)
			if err != nil {
				return 0, fmt.Errorf("time: invalid duration %q", orig)
			}
			for n := pl - len(s); n > 0; n-- {
				scale *= 10
			}
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, fmt.Errorf("time: invalid duration %q", orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, fmt.Errorf("time: missing unit in duration %q", orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := p.unitMap[u]
		if !ok {
			return 0, fmt.Errorf("time: unknown unit %q in duration %q", u, orig)
		}
		if v > 1<<63/uint64(unit) {
			// overflow
			return 0, fmt.Errorf("time: invalid duration %q", orig)
		}
		v *= uint64(unit)
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, fmt.Errorf("time: invalid duration %q", orig)
			}
		}
		d += int64(v)
		if d > 1<<63-1 {
			return 0, fmt.Errorf("time: invalid duration %q", orig)
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x uint64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, "", fmt.Errorf("overflow")
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, "", fmt.Errorf("overflow")
		}
	}
	return x, s[i:], nil
}

// extractUnits extracts unit strings from the duration string.
func (p *Parser) extractUnits(s string) []string {
	var units []string
	for unit := range p.unitMap {
		if strings.Contains(s, unit) {
			units = append(units, unit)
		}
	}
	return units
}

// validateInput validates the input string.
func (p *Parser) validateInput(input string) error {
	if len(input) == 0 {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input duration string is empty",
		}
	}

	if p.config.MaxSize > 0 && int64(len(input)) > p.config.MaxSize {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSize,
			Message: fmt.Sprintf("duration string length %d exceeds maximum %d", len(input), p.config.MaxSize),
		}
	}

	// Basic validation - must contain digits and letters
	hasDigit := false
	hasLetter := false
	for _, r := range input {
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if hasDigit && hasLetter {
			break
		}
	}

	if !hasDigit {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "duration string must contain digits",
		}
	}

	if !hasLetter {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "duration string must contain unit letters",
		}
	}

	return nil
}

// Formatter implements duration formatting.
type Formatter struct {
	useShortUnits bool
}

// NewFormatter creates a new duration formatter.
func NewFormatter() *Formatter {
	return &Formatter{
		useShortUnits: true,
	}
}

// NewFormatterWithLongUnits creates a new duration formatter with long unit names.
func NewFormatterWithLongUnits() *Formatter {
	return &Formatter{
		useShortUnits: false,
	}
}

// Format implements interfaces.Formatter.
func (f *Formatter) Format(ctx context.Context, data *ParsedDuration) ([]byte, error) {
	if data == nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "data cannot be nil",
		}
	}

	formatted := f.formatDuration(data.Duration)
	return []byte(formatted), nil
}

// FormatString implements interfaces.Formatter.
func (f *Formatter) FormatString(ctx context.Context, data *ParsedDuration) (string, error) {
	result, err := f.Format(ctx, data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// formatDuration formats a duration with enhanced unit support.
func (f *Formatter) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	var result strings.Builder

	if d < 0 {
		result.WriteByte('-')
		d = -d
	}

	// Handle weeks
	if weeks := d / Week; weeks > 0 {
		result.WriteString(strconv.FormatInt(int64(weeks), 10))
		result.WriteByte('w')
		d -= weeks * Week
	}

	// Handle days
	if days := d / Day; days > 0 {
		result.WriteString(strconv.FormatInt(int64(days), 10))
		result.WriteByte('d')
		d -= days * Day
	}

	// Use standard formatting for the rest
	if d > 0 {
		standard := d.String()
		result.WriteString(standard)
	}

	if result.Len() == 0 || (result.Len() == 1 && result.String() == "-") {
		return "0s"
	}

	return result.String()
}

// FormatWriter is not commonly used for duration, returns error.
func (f *Formatter) FormatWriter(ctx context.Context, data *ParsedDuration, writer interface{}) error {
	return &interfaces.ParseError{
		Type:    interfaces.ErrorTypeValidation,
		Message: "FormatWriter not supported for duration formatter",
	}
}

// Utility functions

// ParseDuration parses a duration string with enhanced unit support.
func ParseDuration(input string) (*ParsedDuration, error) {
	parser := NewParser()
	return parser.ParseString(context.Background(), input)
}

// FormatDuration formats a time.Duration with enhanced unit support.
func FormatDuration(d time.Duration) string {
	formatter := NewFormatter()
	result, _ := formatter.FormatString(context.Background(), &ParsedDuration{Duration: d})
	return result
}

// ToDays converts a duration to days (as float64).
func ToDays(d time.Duration) float64 {
	return float64(d) / float64(Day)
}

// ToWeeks converts a duration to weeks (as float64).
func ToWeeks(d time.Duration) float64 {
	return float64(d) / float64(Week)
}

// FromDays creates a duration from days.
func FromDays(days float64) time.Duration {
	return time.Duration(days * float64(Day))
}

// FromWeeks creates a duration from weeks.
func FromWeeks(weeks float64) time.Duration {
	return time.Duration(weeks * float64(Week))
}
