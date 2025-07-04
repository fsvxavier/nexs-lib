package duration

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
)

// Extended time units beyond the standard library
const (
	Day   time.Duration = 24 * time.Hour
	Week  time.Duration = 7 * Day
	Month time.Duration = 30 * Day  // Approximate
	Year  time.Duration = 365 * Day // Approximate
)

// Parser implements the DurationParser interface
type Parser struct {
	config    *parsers.Config
	unitMap   map[string]time.Duration
	unitCache map[string]time.Duration
}

// NewParser creates a new duration parser with default configuration
func NewParser(opts ...parsers.Option) *Parser {
	config := parsers.DefaultConfig()
	for _, opt := range opts {
		opt.Apply(config)
	}

	p := &Parser{
		config:    config,
		unitCache: make(map[string]time.Duration),
	}

	p.initializeUnits()
	return p
}

// Parse parses a duration string with the default configuration
func (p *Parser) Parse(ctx context.Context, input string) (time.Duration, error) {
	return p.ParseWithOptions(ctx, input)
}

// ParseWithOptions parses a duration string with additional options
func (p *Parser) ParseWithOptions(ctx context.Context, input string, opts ...parsers.Option) (time.Duration, error) {
	if ctx.Err() != nil {
		return 0, parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "context cancelled")
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return 0, parsers.NewInvalidValueError(input, "empty input")
	}

	// Try standard time.ParseDuration first
	if d, err := time.ParseDuration(input); err == nil {
		return d, nil
	}

	// Apply options to a temporary config
	config := *p.config
	for _, opt := range opts {
		opt.Apply(&config)
	}

	// Try extended parsing
	return p.parseExtended(ctx, input, &config)
}

// MustParse parses a duration string and panics on error
func (p *Parser) MustParse(ctx context.Context, input string) time.Duration {
	d, err := p.Parse(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("failed to parse duration '%s': %v", input, err))
	}
	return d
}

// TryParse attempts to parse a duration string, returning success status
func (p *Parser) TryParse(ctx context.Context, input string) (time.Duration, bool) {
	d, err := p.Parse(ctx, input)
	return d, err == nil
}

// ParseExtended parses extended duration formats with days, weeks, months, years
func (p *Parser) ParseExtended(ctx context.Context, input string) (time.Duration, error) {
	return p.parseExtended(ctx, input, p.config)
}

// GetSupportedUnits returns all supported duration units
func (p *Parser) GetSupportedUnits() map[string]time.Duration {
	result := make(map[string]time.Duration)
	for k, v := range p.unitMap {
		result[k] = v
	}
	return result
}

// initializeUnits sets up the unit mappings
func (p *Parser) initializeUnits() {
	p.unitMap = map[string]time.Duration{
		// Standard Go time units
		"ns": time.Nanosecond,
		"us": time.Microsecond,
		"µs": time.Microsecond, // U+00B5 = micro symbol
		"μs": time.Microsecond, // U+03BC = Greek letter mu
		"ms": time.Millisecond,
		"s":  time.Second,
		"m":  time.Minute,
		"h":  time.Hour,

		// Extended units
		"d":      Day,
		"day":    Day,
		"days":   Day,
		"w":      Week,
		"week":   Week,
		"weeks":  Week,
		"mo":     Month,
		"month":  Month,
		"months": Month,
		"y":      Year,
		"year":   Year,
		"years":  Year,

		// Alternative forms
		"sec":     time.Second,
		"second":  time.Second,
		"seconds": time.Second,
		"min":     time.Minute,
		"minute":  time.Minute,
		"minutes": time.Minute,
		"hr":      time.Hour,
		"hour":    time.Hour,
		"hours":   time.Hour,
	}

	// Add custom units from config
	for unit, duration := range p.config.CustomUnits {
		p.unitMap[unit] = duration
	}
}

// parseExtended handles extended duration parsing
func (p *Parser) parseExtended(ctx context.Context, input string, config *parsers.Config) (time.Duration, error) {
	// Handle special cases
	if input == "0" {
		return 0, nil
	}

	// Normalize input
	normalized := p.normalizeInput(input)

	// Try different parsing strategies
	strategies := []func(context.Context, string, *parsers.Config) (time.Duration, error){
		p.parseRegularFormat,
		p.parseVerboseFormat,
		p.parseNumericWithUnit,
		p.parseRelativeFormat,
	}

	var lastErr error
	for _, strategy := range strategies {
		select {
		case <-ctx.Done():
			return 0, parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "parsing timeout")
		default:
		}

		if d, err := strategy(ctx, normalized, config); err == nil {
			return d, nil
		} else {
			lastErr = err
		}
	}

	// Return detailed error
	parseErr := parsers.NewInvalidFormatError(input, "duration")
	if lastErr != nil {
		parseErr.Cause = lastErr
	}

	suggestions := p.generateSuggestions(input)
	for _, suggestion := range suggestions {
		parseErr.WithSuggestion(suggestion)
	}

	return 0, parseErr
}

// normalizeInput cleans and normalizes the input string
func (p *Parser) normalizeInput(input string) string {
	// Remove extra whitespace
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(input), " ")

	// Convert to lowercase for case-insensitive matching if configured
	if p.config.IgnoreCase {
		normalized = strings.ToLower(normalized)
	}

	return normalized
}

// parseRegularFormat handles standard duration format (e.g., "1h30m45s")
func (p *Parser) parseRegularFormat(ctx context.Context, input string, config *parsers.Config) (time.Duration, error) {
	// Pattern: [+-]?([0-9]*(\.[0-9]*)?[a-z]+)+
	pattern := regexp.MustCompile(`^[+-]?(?:(?:\d*\.?\d*[a-zA-Z]+)+)$`)
	if !pattern.MatchString(input) {
		return 0, fmt.Errorf("invalid regular format")
	}

	var totalDuration time.Duration
	negative := false

	// Handle sign
	if len(input) > 0 && (input[0] == '+' || input[0] == '-') {
		negative = input[0] == '-'
		input = input[1:]
	}

	// Extract number-unit pairs
	unitPattern := regexp.MustCompile(`(\d*\.?\d*)([a-zA-Z]+)`)
	matches := unitPattern.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("no valid number-unit pairs found")
	}

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		numberStr := match[1]
		unitStr := match[2]

		// Parse number
		var number float64
		var err error
		if numberStr == "" {
			number = 1 // Default to 1 if no number specified
		} else {
			number, err = strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number '%s': %v", numberStr, err)
			}
		}

		// Get unit duration
		unitDuration, exists := p.unitMap[strings.ToLower(unitStr)]
		if !exists {
			return 0, fmt.Errorf("unknown unit '%s'", unitStr)
		}

		// Add to total
		totalDuration += time.Duration(float64(unitDuration) * number)
	}

	if negative {
		totalDuration = -totalDuration
	}

	return totalDuration, nil
}

// parseVerboseFormat handles verbose format (e.g., "1 hour 30 minutes 45 seconds")
func (p *Parser) parseVerboseFormat(ctx context.Context, input string, config *parsers.Config) (time.Duration, error) {
	// Split into words
	words := strings.Fields(input)
	if len(words) < 2 {
		return 0, fmt.Errorf("insufficient words for verbose format")
	}

	var totalDuration time.Duration
	i := 0

	for i < len(words) {
		// Look for number followed by unit
		if i+1 >= len(words) {
			break
		}

		numberStr := words[i]
		unitStr := words[i+1]

		// Parse number
		number, err := strconv.ParseFloat(numberStr, 64)
		if err != nil {
			// Try next word pair
			i++
			continue
		}

		// Get unit duration (try both singular and plural)
		unitDuration, exists := p.unitMap[strings.ToLower(unitStr)]
		if !exists {
			// Try without 's' for plural
			if strings.HasSuffix(unitStr, "s") {
				unitDuration, exists = p.unitMap[strings.ToLower(unitStr[:len(unitStr)-1])]
			}
		}

		if !exists {
			// Try next word pair
			i++
			continue
		}

		// Add to total
		totalDuration += time.Duration(float64(unitDuration) * number)
		i += 2
	}

	if totalDuration == 0 {
		return 0, fmt.Errorf("no valid duration found in verbose format")
	}

	return totalDuration, nil
}

// parseNumericWithUnit handles simple numeric with unit (e.g., "30 minutes", "2.5 hours")
func (p *Parser) parseNumericWithUnit(ctx context.Context, input string, config *parsers.Config) (time.Duration, error) {
	// Pattern: number followed by unit
	pattern := regexp.MustCompile(`^([+-]?\d*\.?\d+)\s+([a-zA-Z]+)$`)
	matches := pattern.FindStringSubmatch(input)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid numeric with unit format")
	}

	numberStr := matches[1]
	unitStr := matches[2]

	// Parse number
	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number '%s': %v", numberStr, err)
	}

	// Get unit duration
	unitDuration, exists := p.unitMap[strings.ToLower(unitStr)]
	if !exists {
		return 0, fmt.Errorf("unknown unit '%s'", unitStr)
	}

	return time.Duration(float64(unitDuration) * number), nil
}

// parseRelativeFormat handles relative durations (e.g., "half an hour", "quarter day")
func (p *Parser) parseRelativeFormat(ctx context.Context, input string, config *parsers.Config) (time.Duration, error) {
	lower := strings.ToLower(input)

	// Define relative mappings
	relatives := map[string]time.Duration{
		"a second":      time.Second,
		"a minute":      time.Minute,
		"an hour":       time.Hour,
		"a day":         Day,
		"a week":        Week,
		"half a second": time.Second / 2,
		"half a minute": time.Minute / 2,
		"half an hour":  time.Hour / 2,
		"half a day":    Day / 2,
		"quarter hour":  time.Hour / 4,
		"quarter day":   Day / 4,
		"quarter week":  Week / 4,
	}

	if duration, exists := relatives[lower]; exists {
		return duration, nil
	}

	// Handle "X and a half" patterns
	halfPattern := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s+and\s+a\s+half\s+([a-zA-Z]+)$`)
	if matches := halfPattern.FindStringSubmatch(lower); len(matches) == 3 {
		number, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			unitStr := matches[2]
			if unitDuration, exists := p.unitMap[unitStr]; exists {
				return time.Duration(float64(unitDuration) * (number + 0.5)), nil
			}
		}
	}

	return 0, fmt.Errorf("unrecognized relative format")
}

// generateSuggestions creates helpful suggestions for parsing errors
func (p *Parser) generateSuggestions(input string) []string {
	suggestions := make([]string, 0)

	// Analyze input to provide specific suggestions
	if regexp.MustCompile(`\d`).MatchString(input) {
		suggestions = append(suggestions, "try format: '1h30m' or '1 hour 30 minutes'")
	}

	if strings.Contains(input, " ") {
		suggestions = append(suggestions, "try format: '2.5 hours' or '30 minutes'")
	}

	// Add unit suggestions
	suggestions = append(suggestions, "supported units: ns, us, ms, s, m, h, d, w")
	suggestions = append(suggestions, "examples: '1h30m', '2 days', '1.5 hours', 'half an hour'")

	return suggestions
}

// Package-level convenience functions

// Parse parses a duration string using the default parser
func Parse(input string) (time.Duration, error) {
	parser := NewParser()
	return parser.Parse(context.Background(), input)
}

// MustParse parses a duration string using the default parser and panics on error
func MustParse(input string) time.Duration {
	parser := NewParser()
	return parser.MustParse(context.Background(), input)
}

// ParseExtended parses extended duration formats with days, weeks, months, years
func ParseExtended(input string) (time.Duration, error) {
	parser := NewParser()
	return parser.ParseExtended(context.Background(), input)
}

// GetSupportedUnits returns all supported duration units
func GetSupportedUnits() map[string]time.Duration {
	parser := NewParser()
	return parser.GetSupportedUnits()
}
