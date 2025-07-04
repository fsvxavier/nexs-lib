package datetime

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
)

// ParserOption defines a function signature for compatibility with old library
// This maintains backwards compatibility with the old API
type ParserOption func(*CompatibilityConfig) error

// CompatibilityConfig holds configuration options for backwards compatibility
type CompatibilityConfig struct {
	PreferMonthFirst           bool
	RetryAmbiguousDateWithSwap bool
	Location                   *time.Location
}

// PreferMonthFirst creates an option for setting month preference
func PreferMonthFirst(preferMonthFirst bool) ParserOption {
	return func(c *CompatibilityConfig) error {
		c.PreferMonthFirst = preferMonthFirst
		return nil
	}
}

// RetryAmbiguousDateWithSwap creates an option for retry behavior
func RetryAmbiguousDateWithSwap(retryAmbiguousDateWithSwap bool) ParserOption {
	return func(c *CompatibilityConfig) error {
		c.RetryAmbiguousDateWithSwap = retryAmbiguousDateWithSwap
		return nil
	}
}

// Parser implements the DateTimeParser interface
type Parser struct {
	config        *parsers.Config
	formatCache   map[string]string
	customFormats []string
}

// NewParser creates a new datetime parser with default configuration
func NewParser(opts ...parsers.Option) *Parser {
	config := parsers.DefaultConfig()
	for _, opt := range opts {
		opt.Apply(config)
	}

	p := &Parser{
		config:      config,
		formatCache: make(map[string]string),
	}

	p.initializeFormats()
	return p
}

// Parse parses a datetime string with the default configuration
func (p *Parser) Parse(ctx context.Context, input string) (time.Time, error) {
	return p.ParseWithOptions(ctx, input)
}

// ParseWithOptions parses a datetime string with additional options
func (p *Parser) ParseWithOptions(ctx context.Context, input string, opts ...parsers.Option) (time.Time, error) {
	if ctx.Err() != nil {
		return time.Time{}, parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "context cancelled")
	}

	// Apply options to a temporary config
	config := *p.config
	for _, opt := range opts {
		opt.Apply(&config)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, parsers.NewInvalidValueError(input, "empty input")
	}

	// Try to parse using cached format first
	if format, exists := p.formatCache[input]; exists {
		if t, err := time.Parse(format, input); err == nil {
			return t.In(config.DefaultLocation), nil
		}
		// Remove from cache if it no longer works
		delete(p.formatCache, input)
	}

	// Try standard formats
	formats := p.getAllFormats(&config)

	var lastErr error
	for _, format := range formats {
		select {
		case <-ctx.Done():
			return time.Time{}, parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "parsing timeout")
		default:
		}

		if t, err := time.ParseInLocation(format, input, config.DefaultLocation); err == nil {
			// Cache successful format
			p.formatCache[input] = format
			return t, nil
		} else {
			lastErr = err
		}
	}

	// Try flexible parsing as fallback
	if !config.StrictMode {
		if t, err := p.flexibleParse(input, &config); err == nil {
			return t, nil
		}
	}

	// Return detailed error with suggestions
	parseErr := parsers.NewInvalidFormatError(input, "datetime")
	if lastErr != nil {
		parseErr.Cause = lastErr
	}

	suggestions := p.generateSuggestions(input)
	for _, suggestion := range suggestions {
		parseErr.WithSuggestion(suggestion)
	}

	return time.Time{}, parseErr
}

// MustParse parses a datetime string and panics on error
func (p *Parser) MustParse(ctx context.Context, input string) time.Time {
	t, err := p.Parse(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("failed to parse datetime '%s': %v", input, err))
	}
	return t
}

// TryParse attempts to parse a datetime string, returning success status
func (p *Parser) TryParse(ctx context.Context, input string) (time.Time, bool) {
	t, err := p.Parse(ctx, input)
	return t, err == nil
}

// ParseFormat detects and returns the format string that can parse the input
func (p *Parser) ParseFormat(ctx context.Context, input string) (string, error) {
	if ctx.Err() != nil {
		return "", parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "context cancelled")
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", parsers.NewInvalidValueError(input, "empty input")
	}

	// Check cache first
	if format, exists := p.formatCache[input]; exists {
		return format, nil
	}

	// Try all formats to find which one works
	formats := p.getAllFormats(p.config)
	for _, format := range formats {
		select {
		case <-ctx.Done():
			return "", parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "format detection timeout")
		default:
		}

		if _, err := time.ParseInLocation(format, input, p.config.DefaultLocation); err == nil {
			// Cache successful format
			p.formatCache[input] = format
			return format, nil
		}
	}

	// Try flexible parsing to detect dynamic format
	if !p.config.StrictMode {
		if detectedFormat := p.detectFlexibleFormat(input); detectedFormat != "" {
			p.formatCache[input] = detectedFormat
			return detectedFormat, nil
		}
	}

	return "", parsers.NewInvalidFormatError(input, "datetime").
		WithSuggestion("try formats like: '2006-01-02 15:04:05', 'Jan 2, 2006', '01/02/2006'")
}

// DetectFormat is an alias for ParseFormat for backward compatibility
func (p *Parser) DetectFormat(ctx context.Context, input string) (string, error) {
	return p.ParseFormat(ctx, input)
}

// ParseWithFormat parses a datetime using a specific format
func (p *Parser) ParseWithFormat(ctx context.Context, input, format string) (time.Time, error) {
	if ctx.Err() != nil {
		return time.Time{}, parsers.WrapError(ctx.Err(), parsers.ErrorTypeTimeout, input, "context cancelled")
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, parsers.NewInvalidValueError(input, "empty input")
	}

	if format == "" {
		return time.Time{}, parsers.NewInvalidValueError(format, "empty format")
	}

	t, err := time.ParseInLocation(format, input, p.config.DefaultLocation)
	if err != nil {
		return time.Time{}, parsers.WrapError(err, parsers.ErrorTypeInvalidFormat, input,
			fmt.Sprintf("failed to parse with format '%s'", format))
	}

	return t, nil
}

// ParseInLocation parses a datetime string in a specific location
func (p *Parser) ParseInLocation(ctx context.Context, input string, loc *time.Location) (time.Time, error) {
	return p.ParseWithOptions(ctx, input, parsers.WithLocation(loc))
}

// ParseToUTC parses a datetime string and converts to UTC
func (p *Parser) ParseToUTC(ctx context.Context, input string) (time.Time, error) {
	t, err := p.Parse(ctx, input)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// SetDefaultLocation sets the default timezone location
func (p *Parser) SetDefaultLocation(loc *time.Location) {
	p.config.DefaultLocation = loc
}

// GetSupportedFormats returns all supported datetime formats
func (p *Parser) GetSupportedFormats() []string {
	return append(p.getStandardFormats(), p.customFormats...)
}

// initializeFormats sets up the default formats
func (p *Parser) initializeFormats() {
	p.customFormats = []string{
		// Add any custom formats here
		"2006-01-02T15:04:05.999999999Z07:00", // RFC3339Nano with timezone
		"02/01/2006 15:04:05",                 // DD/MM/YYYY HH:MM:SS
		"01/02/2006 15:04:05",                 // MM/DD/YYYY HH:MM:SS
		"2006/01/02 15:04:05",                 // YYYY/MM/DD HH:MM:SS
	}
}

// getStandardFormats returns the standard Go time formats
func (p *Parser) getStandardFormats() []string {
	return []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"15:04:05",
		"15:04",
		// Common date formats
		"01/02/2006", // MM/DD/YYYY (US)
		"02/01/2006", // DD/MM/YYYY (European)
		"2006/01/02", // YYYY/MM/DD (ISO-ish)
		"01-02-2006", // MM-DD-YYYY
		"02-01-2006", // DD-MM-YYYY
		"2006-01-02", // YYYY-MM-DD (ISO)
		"01.02.2006", // MM.DD.YYYY
		"02.01.2006", // DD.MM.YYYY
		"2006.01.02", // YYYY.MM.DD
		// Date with time
		"01/02/2006 15:04:05",
		"02/01/2006 15:04:05",
		"2006/01/02 15:04:05",
		"01-02-2006 15:04:05",
		"02-01-2006 15:04:05",
		"January 2, 2006",
		"January 2, 2006 15:04:05",
		"Jan 2, 2006",
		"Jan 2, 2006 15:04:05",
		"2 January 2006",
		"2 Jan 2006",
		"Monday, January 2, 2006",
		"Mon, Jan 2, 2006",
	}
}

// getAllFormats returns all available formats including custom ones
func (p *Parser) getAllFormats(config *parsers.Config) []string {
	formats := []string{}

	// Add timezone and RFC formats first (unambiguous)
	formats = append(formats,
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	)

	// Add ambiguous date formats based on DateOrder preference
	switch config.DateOrder {
	case parsers.DateOrderMDY:
		// Prefer US format (MM/DD/YYYY)
		formats = append(formats,
			"01/02/2006", // MM/DD/YYYY
			"01-02-2006", // MM-DD-YYYY
			"01.02.2006", // MM.DD.YYYY
			"01/02/2006 15:04:05",
			"01-02-2006 15:04:05",
			"02/01/2006", // DD/MM/YYYY (fallback)
			"02-01-2006", // DD-MM-YYYY (fallback)
			"02.01.2006", // DD.MM.YYYY (fallback)
			"02/01/2006 15:04:05",
			"02-01-2006 15:04:05",
		)
	case parsers.DateOrderDMY:
		// Prefer European format (DD/MM/YYYY)
		formats = append(formats,
			"02/01/2006", // DD/MM/YYYY
			"02-01-2006", // DD-MM-YYYY
			"02.01.2006", // DD.MM.YYYY
			"02/01/2006 15:04:05",
			"02-01-2006 15:04:05",
			"01/02/2006", // MM/DD/YYYY (fallback)
			"01-02-2006", // MM-DD-YYYY (fallback)
			"01.02.2006", // MM.DD.YYYY (fallback)
			"01/02/2006 15:04:05",
			"01-02-2006 15:04:05",
		)
	default:
		// Add both with US preference as default
		formats = append(formats,
			"01/02/2006", // MM/DD/YYYY
			"02/01/2006", // DD/MM/YYYY
			"01-02-2006", // MM-DD-YYYY
			"02-01-2006", // DD-MM-YYYY
			"01.02.2006", // MM.DD.YYYY
			"02.01.2006", // DD.MM.YYYY
			"01/02/2006 15:04:05",
			"02/01/2006 15:04:05",
			"01-02-2006 15:04:05",
			"02-01-2006 15:04:05",
		)
	}

	// Add other common formats
	formats = append(formats,
		"2006/01/02", // YYYY/MM/DD
		"2006.01.02", // YYYY.MM.DD
		"2006/01/02 15:04:05",
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		"15:04:05",
		"15:04",
		"January 2, 2006",
		"January 2, 2006 15:04:05",
		"Jan 2, 2006",
		"Jan 2, 2006 15:04:05",
		"2 January 2006",
		"2 Jan 2006",
		"Monday, January 2, 2006",
		"Mon, Jan 2, 2006",
	)

	// Add custom formats
	formats = append(formats, p.customFormats...)
	formats = append(formats, config.CustomFormats...)

	return formats
}

// flexibleParse attempts to parse using pattern recognition
func (p *Parser) flexibleParse(input string, config *parsers.Config) (time.Time, error) {
	// Normalize input
	normalized := p.normalizeInput(input)

	// Try to detect and parse various patterns
	patterns := []func(string, *parsers.Config) (time.Time, error){
		p.parseUnixTimestamp, // Try Unix timestamp first
		p.parseISO8601Variants,
		p.parseNumericFormats,
		p.parseTextualFormats,
		p.parseRelativeFormats,
	}

	for _, pattern := range patterns {
		if t, err := pattern(normalized, config); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse flexible format")
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

// parseISO8601Variants attempts to parse ISO8601-like formats
func (p *Parser) parseISO8601Variants(input string, config *parsers.Config) (time.Time, error) {
	// Handle various ISO8601 variants
	isoPatterns := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05.000000000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, pattern := range isoPatterns {
		if t, err := time.ParseInLocation(pattern, input, config.DefaultLocation); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("not an ISO8601 variant")
}

// parseNumericFormats attempts to parse numeric date formats
func (p *Parser) parseNumericFormats(input string, config *parsers.Config) (time.Time, error) {
	// Handle various numeric formats based on separators
	separators := []string{"/", "-", ".", " "}

	for _, sep := range separators {
		if strings.Contains(input, sep) {
			parts := strings.Split(input, sep)
			if len(parts) >= 3 {
				if t, err := p.parseNumericParts(parts, config); err == nil {
					return t, nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("not a numeric format")
}

// parseNumericParts parses numeric date parts considering date order preference
func (p *Parser) parseNumericParts(parts []string, config *parsers.Config) (time.Time, error) {
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("insufficient parts")
	}

	// Convert parts to integers
	nums := make([]int, len(parts))
	for i, part := range parts {
		var err error
		nums[i], err = strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Determine year, month, day based on config.DateOrder
	var year, month, day int

	switch config.DateOrder {
	case parsers.DateOrderMDY: // MM/DD/YYYY
		month, day, year = nums[0], nums[1], nums[2]
	case parsers.DateOrderDMY: // DD/MM/YYYY
		day, month, year = nums[0], nums[1], nums[2]
	case parsers.DateOrderYMD: // YYYY/MM/DD
		year, month, day = nums[0], nums[1], nums[2]
	default:
		// Auto-detect based on values
		if nums[0] > 31 || nums[0] > 12 && nums[1] <= 12 && nums[2] <= 31 {
			year, month, day = nums[0], nums[1], nums[2] // YYYY/MM/DD
		} else if nums[2] > 31 || nums[2] > 12 && nums[0] <= 12 && nums[1] <= 31 {
			month, day, year = nums[0], nums[1], nums[2] // MM/DD/YYYY
		} else {
			day, month, year = nums[0], nums[1], nums[2] // DD/MM/YYYY
		}
	}

	// Handle 2-digit years
	if year < 100 {
		if year < 50 {
			year += 2000
		} else {
			year += 1900
		}
	}

	// Parse time if more parts available
	hour, minute, second := 0, 0, 0
	if len(nums) > 3 {
		hour = nums[3]
	}
	if len(nums) > 4 {
		minute = nums[4]
	}
	if len(nums) > 5 {
		second = nums[5]
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, config.DefaultLocation), nil
}

// parseTextualFormats attempts to parse formats with month names
func (p *Parser) parseTextualFormats(input string, config *parsers.Config) (time.Time, error) {
	// Define month mappings
	months := map[string]time.Month{
		"january": time.January, "jan": time.January,
		"february": time.February, "feb": time.February,
		"march": time.March, "mar": time.March,
		"april": time.April, "apr": time.April,
		"may":  time.May,
		"june": time.June, "jun": time.June,
		"july": time.July, "jul": time.July,
		"august": time.August, "aug": time.August,
		"september": time.September, "sep": time.September, "sept": time.September,
		"october": time.October, "oct": time.October,
		"november": time.November, "nov": time.November,
		"december": time.December, "dec": time.December,
	}

	// Convert input to lowercase for matching
	lower := strings.ToLower(input)
	words := strings.Fields(lower)

	var month time.Month
	var day, year int
	var hour, minute int
	var found bool
	var isAM, isPM bool

	// Find month
	for _, word := range words {
		if m, exists := months[strings.TrimSuffix(word, ",")]; exists {
			month = m
			found = true
			break
		}
	}

	if !found {
		return time.Time{}, fmt.Errorf("no month found")
	}

	// Check for AM/PM
	for _, word := range words {
		if strings.Contains(word, "am") {
			isAM = true
		} else if strings.Contains(word, "pm") {
			isPM = true
		}
	}

	// Extract numbers and time patterns
	numbers := make([]int, 0)
	for _, word := range words {
		// Check for time patterns like "10:30"
		if timeMatch := regexp.MustCompile(`(\d{1,2}):(\d{2})`).FindStringSubmatch(word); len(timeMatch) == 3 {
			if h, err := strconv.Atoi(timeMatch[1]); err == nil {
				if m, err := strconv.Atoi(timeMatch[2]); err == nil {
					hour = h
					minute = m
					// Handle 12-hour format
					if isPM && hour != 12 {
						hour += 12
					} else if isAM && hour == 12 {
						hour = 0
					}
				}
			}
		} else {
			// Clean punctuation and extract numbers
			cleaned := strings.Trim(word, ",.!?;:")
			if num, err := strconv.Atoi(cleaned); err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	if len(numbers) < 2 {
		return time.Time{}, fmt.Errorf("insufficient numeric values")
	}

	// Determine day and year
	for _, num := range numbers {
		if num <= 31 && day == 0 {
			day = num
		} else if num > 31 || (num > 12 && year == 0) {
			year = num
		}
	}

	// Handle missing values
	if day == 0 {
		day = 1
	}
	if year == 0 {
		year = time.Now().Year()
	}

	// Handle 2-digit years
	if year < 100 {
		if year < 50 {
			year += 2000
		} else {
			year += 1900
		}
	}

	return time.Date(year, month, day, hour, minute, 0, 0, config.DefaultLocation), nil
}

// parseRelativeFormats attempts to parse relative date expressions
func (p *Parser) parseRelativeFormats(input string, config *parsers.Config) (time.Time, error) {
	now := time.Now().In(config.DefaultLocation)
	lower := strings.ToLower(strings.TrimSpace(input))

	switch lower {
	case "now", "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	}

	// Handle "X days ago", "X weeks ago", etc.
	relativePattern := regexp.MustCompile(`^(\d+)\s+(days?|weeks?|months?|years?)\s+ago$`)
	if matches := relativePattern.FindStringSubmatch(lower); len(matches) == 3 {
		num, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		switch {
		case strings.HasPrefix(unit, "day"):
			return now.AddDate(0, 0, -num), nil
		case strings.HasPrefix(unit, "week"):
			return now.AddDate(0, 0, -num*7), nil
		case strings.HasPrefix(unit, "month"):
			return now.AddDate(0, -num, 0), nil
		case strings.HasPrefix(unit, "year"):
			return now.AddDate(-num, 0, 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("not a relative format")
}

// generateSuggestions creates helpful suggestions for parsing errors
func (p *Parser) generateSuggestions(input string) []string {
	suggestions := make([]string, 0)

	// Analyze input to provide specific suggestions
	if strings.Contains(input, "/") {
		suggestions = append(suggestions, "try format: MM/DD/YYYY or DD/MM/YYYY")
	}
	if strings.Contains(input, "-") {
		suggestions = append(suggestions, "try format: YYYY-MM-DD or DD-MM-YYYY")
	}
	if regexp.MustCompile(`\d{4}`).MatchString(input) {
		suggestions = append(suggestions, "ensure year is in YYYY format")
	}
	if regexp.MustCompile(`[a-zA-Z]`).MatchString(input) {
		suggestions = append(suggestions, "try format: 'January 2, 2006' or 'Jan 2 2006'")
	}

	// Add general suggestions
	suggestions = append(suggestions, "supported formats include RFC3339, ISO8601, and common date formats")

	return suggestions
}

// detectFlexibleFormat attempts to detect the format of a date string through pattern analysis
func (p *Parser) detectFlexibleFormat(input string) string {
	// Common format patterns based on content analysis
	patterns := map[*regexp.Regexp]string{
		// ISO formats - more specific patterns first
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`):               "2006-01-02T15:04:05Z",
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`):        "2006-01-02T15:04:05.000Z",
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$`):        "2006-01-02T15:04:05.000000Z",
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.[\d]{1,9}Z$`):    "2006-01-02T15:04:05.999999999Z",
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$`): time.RFC3339,
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`):                "2006-01-02 15:04:05",
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`):                                  "2006-01-02",

		// US formats
		regexp.MustCompile(`^(0?[1-9]|1[0-2])/(0?[1-9]|[12]\d|3[01])/\d{4} \d{1,2}:\d{2}:\d{2}$`): "1/2/2006 15:04:05",
		regexp.MustCompile(`^(0?[1-9]|1[0-2])/(0?[1-9]|[12]\d|3[01])/\d{4}$`):                     "1/2/2006",
		regexp.MustCompile(`^(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])/\d{4}$`):                       "01/02/2006",

		// European formats
		regexp.MustCompile(`^(0?[1-9]|[12]\d|3[01])\.(0?[1-9]|1[0-2])\.\d{4}$`): "2.1.2006",
		regexp.MustCompile(`^(0[1-9]|[12]\d|3[01])\.(0[1-9]|1[0-2])\.\d{4}$`):   "02.01.2006",

		// Text formats - case insensitive
		regexp.MustCompile(`^[A-Za-z]{3} (0?[1-9]|[12]\d|3[01]), \d{4}$`): "Jan 2, 2006",
		regexp.MustCompile(`^[A-Za-z]{3} (0?[1-9]|[12]\d|3[01]) \d{4}$`):  "Jan 2 2006",
		regexp.MustCompile(`^[A-Za-z]+ (0?[1-9]|[12]\d|3[01]), \d{4}$`):   "January 2, 2006",
		regexp.MustCompile(`^(0?[1-9]|[12]\d|3[01]) [A-Za-z]{3} \d{4}$`):  "2 Jan 2006",
		regexp.MustCompile(`^(0?[1-9]|[12]\d|3[01]) [A-Za-z]+ \d{4}$`):    "2 January 2006",

		// Time only
		regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`): "15:04:05",
		regexp.MustCompile(`^\d{1,2}:\d{2}$`):     "15:04",

		// Kitchen format
		regexp.MustCompile(`^\d{1,2}:\d{2}:\d{2}\s?(AM|PM|am|pm)$`): time.Kitchen,
		regexp.MustCompile(`^\d{1,2}:\d{2}\s?(AM|PM|am|pm)$`):       time.Kitchen,
	}

	// Test each pattern
	for pattern, format := range patterns {
		if pattern.MatchString(input) {
			// Verify the format actually works
			if _, err := time.Parse(format, input); err == nil {
				return format
			}
		}
	}

	// Fallback: try standard time layouts
	standardLayouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	}

	for _, layout := range standardLayouts {
		if _, err := time.Parse(layout, input); err == nil {
			return layout
		}
	}

	return "" // Unable to detect
}

// Package-level convenience functions

// Parse parses a datetime string using the default parser
func Parse(input string) (time.Time, error) {
	parser := NewParser()
	return parser.Parse(context.Background(), input)
}

// ParseToUTC parses a datetime string and converts to UTC
func ParseToUTC(input string) (time.Time, error) {
	parser := NewParser()
	t, err := parser.Parse(context.Background(), input)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// ParseFormat detects and returns the format string that can parse the input
func ParseFormat(input string) (string, error) {
	parser := NewParser()
	return parser.ParseFormat(context.Background(), input)
}

// DetectFormat is an alias for ParseFormat for backward compatibility
func DetectFormat(input string) (string, error) {
	return ParseFormat(input)
}

// ParseWithFormat parses a datetime using a specific format
func ParseWithFormat(input, format string) (time.Time, error) {
	parser := NewParser()
	return parser.ParseWithFormat(context.Background(), input, format)
}

// Convenience functions for backwards compatibility with old library API

// ParseAnySimple parses a datetime string using the default configuration (simple version)
func ParseAnySimple(input string) (time.Time, error) {
	return ParseAny(input)
}

// MustParseSimple parses a datetime string and panics on error (simple version)
func MustParseSimple(input string) time.Time {
	return MustParseAny(input)
}

// ParseInSimple parses a datetime string in a specific location (simple version)
func ParseInSimple(input string, loc *time.Location) (time.Time, error) {
	parser := NewParser(parsers.WithLocation(loc))
	return parser.Parse(context.Background(), input)
}

// ParseLocalSimple parses a datetime string using the local timezone (simple version)
func ParseLocalSimple(input string) (time.Time, error) {
	return ParseInSimple(input, time.Local)
}

// ParseAny parses an unknown date format, detecting the layout automatically.
// This function maintains 100% compatibility with the old library API.
// Normal parse with equivalent timezone rules as time.Parse().
// NOTE: please see readme on mmdd vs ddmm ambiguous dates.
func ParseAny(datestr string, opts ...ParserOption) (time.Time, error) {
	// Create compatibility config with defaults matching old library
	config := &CompatibilityConfig{
		PreferMonthFirst:           true,
		RetryAmbiguousDateWithSwap: false,
		Location:                   time.UTC,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return time.Time{}, err
		}
	}

	// Convert to new parser options
	var parserOpts []parsers.Option

	if config.Location != nil {
		parserOpts = append(parserOpts, parsers.WithLocation(config.Location))
	}

	// Set date order based on prefer month first
	if config.PreferMonthFirst {
		parserOpts = append(parserOpts, parsers.WithDateOrder(parsers.DateOrderMDY))
	} else {
		parserOpts = append(parserOpts, parsers.WithDateOrder(parsers.DateOrderDMY))
	}

	// Create parser and parse
	parser := NewParser(parserOpts...)
	ctx := context.Background()

	result, err := parser.ParseWithOptions(ctx, datestr, parserOpts...)

	// Handle ambiguous date retry logic if enabled
	if err != nil && config.RetryAmbiguousDateWithSwap &&
		(strings.Contains(err.Error(), "month out of range") ||
			strings.Contains(err.Error(), "day out of range")) {

		// Retry with opposite month preference
		retryOpts := append([]parsers.Option(nil), parserOpts...)

		// Toggle date order
		if config.PreferMonthFirst {
			retryOpts = append(retryOpts, parsers.WithDateOrder(parsers.DateOrderDMY))
		} else {
			retryOpts = append(retryOpts, parsers.WithDateOrder(parsers.DateOrderMDY))
		}

		retryParser := NewParser(retryOpts...)
		if retryResult, retryErr := retryParser.ParseWithOptions(ctx, datestr, retryOpts...); retryErr == nil {
			return retryResult, nil
		}
	}

	return result, err
}

// ParseIn parses with location, equivalent to time.ParseInLocation() timezone/offset rules.
// Using location arg, if timezone/offset info exists in the datestring, it uses the given
// location rules for any zone interpretation.
func ParseIn(datestr string, loc *time.Location, opts ...ParserOption) (time.Time, error) {
	// Add location to the options
	locationOpts := make([]ParserOption, 0, len(opts)+1)
	locationOpts = append(locationOpts, func(c *CompatibilityConfig) error {
		c.Location = loc
		return nil
	})
	locationOpts = append(locationOpts, opts...)

	return ParseAny(datestr, locationOpts...)
}

// ParseLocal parses given an unknown date format, detect the layout, using time.Local.
// Set Location to time.Local. Same as ParseIn Location but lazily uses the global
// time.Local variable for Location argument.
func ParseLocal(datestr string, opts ...ParserOption) (time.Time, error) {
	return ParseIn(datestr, time.Local, opts...)
}

// MustParseAny parses a date, and panics if it can't be parsed. Used for testing.
func MustParseAny(datestr string, opts ...ParserOption) time.Time {
	t, err := ParseAny(datestr, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse date '%s': %v", datestr, err))
	}
	return t
}

// ParseStrict parses with strict mode enabled for better error reporting
func ParseStrict(datestr string, opts ...ParserOption) (time.Time, error) {
	// Convert compatibility options to new parser options
	config := &CompatibilityConfig{
		PreferMonthFirst:           true,
		RetryAmbiguousDateWithSwap: false,
		Location:                   time.UTC,
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return time.Time{}, err
		}
	}

	var parserOpts []parsers.Option
	parserOpts = append(parserOpts, parsers.WithStrictMode(true))

	if config.Location != nil {
		parserOpts = append(parserOpts, parsers.WithLocation(config.Location))
	}

	if config.PreferMonthFirst {
		parserOpts = append(parserOpts, parsers.WithDateOrder(parsers.DateOrderMDY))
	} else {
		parserOpts = append(parserOpts, parsers.WithDateOrder(parsers.DateOrderDMY))
	}

	parser := NewParser(parserOpts...)
	return parser.ParseWithOptions(context.Background(), datestr, parserOpts...)
}

// parseUnixTimestamp attempts to parse Unix timestamps
func (p *Parser) parseUnixTimestamp(input string, config *parsers.Config) (time.Time, error) {
	// Check if input looks like a Unix timestamp (all digits, reasonable length)
	if !regexp.MustCompile(`^\d{10}(\.\d+)?$|^\d{13}$`).MatchString(input) {
		return time.Time{}, fmt.Errorf("not a unix timestamp")
	}

	// Try to parse as Unix timestamp
	if timestamp, err := strconv.ParseInt(input, 10, 64); err == nil {
		// Handle both seconds and milliseconds timestamps
		if len(input) == 13 {
			// Milliseconds timestamp
			return time.Unix(timestamp/1000, (timestamp%1000)*1000000).In(config.DefaultLocation), nil
		} else if len(input) == 10 {
			// Seconds timestamp
			return time.Unix(timestamp, 0).In(config.DefaultLocation), nil
		}
	}

	// Try to parse as float (seconds with decimal)
	if timestamp, err := strconv.ParseFloat(input, 64); err == nil {
		sec := int64(timestamp)
		nsec := int64((timestamp - float64(sec)) * 1e9)
		return time.Unix(sec, nsec).In(config.DefaultLocation), nil
	}

	return time.Time{}, fmt.Errorf("invalid unix timestamp format")
}
