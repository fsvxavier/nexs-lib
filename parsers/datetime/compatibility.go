// Package datetime provides compatibility layer with _old/parse/datetime
package datetime

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Compatibility constants and variables from old module
var (
	// ErrAmbiguousMMDD indicates ambiguous month/day parsing
	ErrAmbiguousMMDD = errors.New("ambiguous M/D vs D/M")
)

// ParserOption defines options for the parser (compatibility type)
type ParserOption func(*compatibilityParser) error

// compatibilityParser maintains state for advanced parsing
type compatibilityParser struct {
	datestr                    string
	loc                        *time.Location
	preferMonthFirst           bool
	retryAmbiguousDateWithSwap bool
	ambiguousMD                bool
	format                     []byte
	t                          time.Time
	// Additional state fields for advanced parsing
	year, month, day        int
	hour, min, sec, nsec    int
	offsetSign              int
	offsetHours, offsetMins int
}

// PreferMonthFirst sets month-first preference for ambiguous dates
func PreferMonthFirst(preferMonthFirst bool) ParserOption {
	return func(p *compatibilityParser) error {
		p.preferMonthFirst = preferMonthFirst
		return nil
	}
}

// RetryAmbiguousDateWithSwap enables retry with day/month swap for ambiguous dates
func RetryAmbiguousDateWithSwap(retryAmbiguousDateWithSwap bool) ParserOption {
	return func(p *compatibilityParser) error {
		p.retryAmbiguousDateWithSwap = retryAmbiguousDateWithSwap
		return nil
	}
}

// ParseFormatToGo converts custom format strings to Go time format
func ParseFormatToGo(format string) string {
	formatGo := strings.ToUpper(format)
	formatGo = strings.ReplaceAll(formatGo, "YYYY", "2006")
	formatGo = strings.ReplaceAll(formatGo, "MM", "01")
	formatGo = strings.ReplaceAll(formatGo, "DD", "02")
	formatGo = strings.ReplaceAll(formatGo, "HH", "15")
	formatGo = strings.ReplaceAll(formatGo, "II", "04")
	formatGo = strings.ReplaceAll(formatGo, "SS", "05")
	return formatGo
}

// ParseDateTime converts datetime string with format and timezone
func ParseDateTime(datetime, format, timezone string) (string, error) {
	formatGo := ParseFormatToGo(format)

	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	// Try to parse with new system first
	parsedResult, err := ParseDatetime(datetime)
	if err != nil {
		// Fallback to advanced parsing
		regularizeDate, err := ParseAny(datetime)
		if err != nil {
			return "", fmt.Errorf("error parsing date: %w", err)
		}
		return regularizeDate.In(loc).Format(formatGo), nil
	}

	return parsedResult.Time.In(loc).Format(formatGo), nil
}

// ParseAny parses unknown date format with automatic detection (compatibility function)
func ParseAny(datestr string, opts ...ParserOption) (time.Time, error) {
	p, err := parseTime(datestr, nil, opts...)
	if err != nil {
		return time.Time{}, err
	}
	return p.parse()
}

// ParseIn parses with specific location (compatibility function)
func ParseIn(datestr string, loc *time.Location, opts ...ParserOption) (time.Time, error) {
	p, err := parseTime(datestr, loc, opts...)
	if err != nil {
		return time.Time{}, err
	}
	return p.parse()
}

// ParseLocal parses using local timezone (compatibility function)
func ParseLocal(datestr string, opts ...ParserOption) (time.Time, error) {
	p, err := parseTime(datestr, time.Local, opts...)
	if err != nil {
		return time.Time{}, err
	}
	return p.parse()
}

// MustParse parses and panics on error (compatibility function)
func MustParse(datestr string, opts ...ParserOption) time.Time {
	p, err := parseTime(datestr, nil, opts...)
	if err != nil {
		panic(err.Error())
	}
	t, err := p.parse()
	if err != nil {
		panic(err.Error())
	}
	return t
}

// ParseFormat returns layout string for the given date (compatibility function)
func ParseFormat(datestr string, opts ...ParserOption) (string, error) {
	p, err := parseTime(datestr, nil, opts...)
	if err != nil {
		return "", err
	}
	_, err = p.parse()
	if err != nil {
		return "", err
	}
	return string(p.format), nil
}

// ParseStrict parses with strict ambiguity checking (compatibility function)
func ParseStrict(datestr string, opts ...ParserOption) (time.Time, error) {
	p, err := parseTime(datestr, nil, opts...)
	if err != nil {
		return time.Time{}, err
	}
	if p.ambiguousMD {
		return time.Time{}, ErrAmbiguousMMDD
	}
	return p.parse()
}

// parseTime creates a new parser and performs initial parsing
func parseTime(datestr string, loc *time.Location, opts ...ParserOption) (*compatibilityParser, error) {
	p := newCompatibilityParser(datestr, loc, opts...)

	// Try new parser first for common formats
	newParser := NewParser()
	if result, err := newParser.ParseString(context.Background(), datestr); err == nil {
		// Convert to compatibility parser result
		p.t = result.Time
		p.format = []byte(result.Layout)
		return p, nil
	}

	// Fallback to advanced parsing
	err := p.parseAdvanced()
	return p, err
}

// newCompatibilityParser creates a new compatibility parser
func newCompatibilityParser(dateStr string, loc *time.Location, opts ...ParserOption) *compatibilityParser {
	p := &compatibilityParser{
		datestr:                    dateStr,
		loc:                        loc,
		preferMonthFirst:           true,
		retryAmbiguousDateWithSwap: false,
	}

	if p.loc == nil {
		p.loc = time.UTC
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// parse returns the parsed time
func (p *compatibilityParser) parse() (time.Time, error) {
	if !p.t.IsZero() {
		return p.t, nil
	}

	// Build time from parsed components
	if p.year == 0 {
		p.year = time.Now().Year()
	}
	if p.month == 0 {
		p.month = 1
	}
	if p.day == 0 {
		p.day = 1
	}

	// Handle timezone offset
	var loc *time.Location = p.loc
	if p.offsetSign != 0 {
		offsetSeconds := p.offsetSign * (p.offsetHours*3600 + p.offsetMins*60)
		loc = time.FixedZone("", offsetSeconds)
	}

	t := time.Date(p.year, time.Month(p.month), p.day, p.hour, p.min, p.sec, p.nsec, loc)
	p.t = t
	return t, nil
}

// parseAdvanced performs advanced parsing similar to old module
func (p *compatibilityParser) parseAdvanced() error {
	datestr := strings.TrimSpace(p.datestr)
	if len(datestr) == 0 {
		return errors.New("empty date string")
	}

	// Try various parsing strategies
	strategies := []func() error{
		p.parseISO8601,
		p.parseRFC3339,
		p.parseCommonFormats,
		p.parseUSFormats,
		p.parseEuropeanFormats,
		p.parseUnixTimestamp,
		p.parseRelativeFormats,
		p.parseTextFormats,
	}

	for _, strategy := range strategies {
		if err := strategy(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to parse date: %s", datestr)
}

// parseISO8601 handles ISO 8601 formats
func (p *compatibilityParser) parseISO8601() error {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05.000000Z07:00",
		"2006-01-02T15:04:05.000000000Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05.000000000Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, p.datestr, p.loc); err == nil {
			p.t = t
			p.format = []byte(format)
			return nil
		}
	}

	return errors.New("not ISO 8601 format")
}

// parseRFC3339 handles RFC 3339 formats
func (p *compatibilityParser) parseRFC3339() error {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, p.datestr, p.loc); err == nil {
			p.t = t
			p.format = []byte(format)
			return nil
		}
	}

	return errors.New("not RFC format")
}

// parseCommonFormats handles common date formats
func (p *compatibilityParser) parseCommonFormats() error {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
		"02-01-2006 15:04:05",
		"02-01-2006 15:04",
		"02-01-2006",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006",
		"01-02-2006 15:04:05",
		"01-02-2006 15:04",
		"01-02-2006",
		"01/02/2006 15:04:05",
		"01/02/2006 15:04",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, p.datestr, p.loc); err == nil {
			p.t = t
			p.format = []byte(format)
			return nil
		}
	}

	return errors.New("not common format")
}

// parseUSFormats handles US-specific formats
func (p *compatibilityParser) parseUSFormats() error {
	// Handle formats like "Jan 2, 2006", "January 2, 2006 15:04:05"
	formats := []string{
		"Jan 2, 2006 15:04:05 PM",
		"Jan 2, 2006 15:04:05",
		"Jan 2, 2006 3:04:05 PM",
		"Jan 2, 2006 3:04 PM",
		"Jan 2, 2006",
		"January 2, 2006 15:04:05 PM",
		"January 2, 2006 15:04:05",
		"January 2, 2006 3:04:05 PM",
		"January 2, 2006 3:04 PM",
		"January 2, 2006",
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan _2 15:04:05 2006",
		"Mon Jan 2 15:04:05 MST 2006",
		"Mon Jan _2 15:04:05 MST 2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, p.datestr, p.loc); err == nil {
			p.t = t
			p.format = []byte(format)
			return nil
		}
	}

	return errors.New("not US format")
}

// parseEuropeanFormats handles European date formats
func (p *compatibilityParser) parseEuropeanFormats() error {
	formats := []string{
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		"2.1.2006 15:04:05",
		"2.1.2006 15:04",
		"2.1.2006",
		"02-01-2006",
		"2-1-2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, p.datestr, p.loc); err == nil {
			p.t = t
			p.format = []byte(format)
			return nil
		}
	}

	return errors.New("not European format")
}

// parseUnixTimestamp handles Unix timestamps
func (p *compatibilityParser) parseUnixTimestamp() error {
	// Try parsing as Unix timestamp (seconds or milliseconds)
	if timestamp, err := strconv.ParseInt(p.datestr, 10, 64); err == nil {
		// Reject negative timestamps and very large values
		if timestamp < 0 || timestamp > 9999999999999 { // Reasonable bounds
			return errors.New("invalid timestamp range")
		}

		var t time.Time
		if timestamp > 1e12 { // Likely milliseconds
			t = time.Unix(timestamp/1000, (timestamp%1000)*1e6)
		} else { // Likely seconds
			t = time.Unix(timestamp, 0)
		}

		// Convert to specified location
		p.t = t.In(p.loc)
		p.format = []byte("unix")
		return nil
	}

	return errors.New("not unix timestamp")
}

// parseRelativeFormats handles relative date formats
func (p *compatibilityParser) parseRelativeFormats() error {
	now := time.Now()
	lower := strings.ToLower(strings.TrimSpace(p.datestr))

	switch lower {
	case "now", "today":
		p.t = now
		p.format = []byte("relative")
		return nil
	case "yesterday":
		p.t = now.AddDate(0, 0, -1)
		p.format = []byte("relative")
		return nil
	case "tomorrow":
		p.t = now.AddDate(0, 0, 1)
		p.format = []byte("relative")
		return nil
	}

	return errors.New("not relative format")
}

// parseTextFormats handles text-based date formats
func (p *compatibilityParser) parseTextFormats() error {
	// Normalize input for text parsing
	normalized := p.normalizeTextInput()

	// Try to extract date components using regex
	patterns := []struct {
		regex  *regexp.Regexp
		parser func([]string) error
	}{
		{
			regexp.MustCompile(`(?i)(\d{1,2})\s+(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{2,4})`),
			p.parseDayMonthYear,
		},
		{
			regexp.MustCompile(`(?i)(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{1,2}),?\s+(\d{2,4})`),
			p.parseMonthDayYear,
		},
		{
			regexp.MustCompile(`(?i)(\d{2,4})\s+(jan|january|feb|february|mar|march|apr|april|may|jun|june|jul|july|aug|august|sep|september|oct|october|nov|november|dec|december)\s+(\d{1,2})`),
			p.parseYearMonthDay,
		},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(normalized); matches != nil {
			return pattern.parser(matches)
		}
	}

	return errors.New("not text format")
}

// normalizeTextInput normalizes text input for parsing
func (p *compatibilityParser) normalizeTextInput() string {
	text := strings.TrimSpace(p.datestr)

	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Handle common abbreviations
	replacements := map[string]string{
		"sept": "september",
		"sep":  "september",
	}

	for old, new := range replacements {
		text = regexp.MustCompile(`(?i)\b`+old+`\b`).ReplaceAllString(text, new)
	}

	return text
}

// parseDayMonthYear parses "15 January 2023" format
func (p *compatibilityParser) parseDayMonthYear(matches []string) error {
	if len(matches) < 4 {
		return errors.New("invalid day month year format")
	}

	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return err
	}

	month := p.parseMonthName(matches[2])
	if month == 0 {
		return errors.New("invalid month name")
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return err
	}

	// Handle 2-digit years
	if year < 100 {
		if year < 50 {
			year += 2000
		} else {
			year += 1900
		}
	}

	// Validate the date (especially for leap years)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc)
	if t.Month() != time.Month(month) || t.Day() != day {
		return errors.New("invalid date")
	}

	p.t = t
	p.format = []byte("day month year")
	return nil
}

// parseYearMonthDay parses "2023 January 15" format
func (p *compatibilityParser) parseYearMonthDay(matches []string) error {
	if len(matches) < 4 {
		return errors.New("invalid year month day format")
	}

	year, err := strconv.Atoi(matches[1])
	if err != nil {
		return err
	}

	month := p.parseMonthName(matches[2])
	if month == 0 {
		return errors.New("invalid month name")
	}

	day, err := strconv.Atoi(matches[3])
	if err != nil {
		return err
	}

	// Validate the date (especially for leap years)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc)
	if t.Month() != time.Month(month) || t.Day() != day {
		return errors.New("invalid date")
	}

	p.t = t
	p.format = []byte("year month day")
	return nil
}

// parseMonthDayYear parses "January 15, 2023" format
func (p *compatibilityParser) parseMonthDayYear(matches []string) error {
	if len(matches) < 4 {
		return errors.New("invalid month day year format")
	}

	month := p.parseMonthName(matches[1])
	if month == 0 {
		return errors.New("invalid month name")
	}

	day, err := strconv.Atoi(matches[2])
	if err != nil {
		return err
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return err
	}

	// Handle 2-digit years
	if year < 100 {
		if year < 50 {
			year += 2000
		} else {
			year += 1900
		}
	}

	// Validate the date (especially for leap years)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, p.loc)
	if t.Month() != time.Month(month) || t.Day() != day {
		return errors.New("invalid date")
	}

	p.t = t
	p.format = []byte("month day year")
	return nil
}

// parseMonthName converts month name to number
func (p *compatibilityParser) parseMonthName(name string) int {
	monthMap := map[string]int{
		"january": 1, "jan": 1,
		"february": 2, "feb": 2,
		"march": 3, "mar": 3,
		"april": 4, "apr": 4,
		"may":  5,
		"june": 6, "jun": 6,
		"july": 7, "jul": 7,
		"august": 8, "aug": 8,
		"september": 9, "sep": 9, "sept": 9,
		"october": 10, "oct": 10,
		"november": 11, "nov": 11,
		"december": 12, "dec": 12,
	}

	return monthMap[strings.ToLower(name)]
}

// ToTime converts ParsedDateTime to time.Time (adapter function)
func ToTime(pdt *ParsedDateTime) time.Time {
	if pdt == nil {
		return time.Time{}
	}
	return pdt.Time
}

// FromTime converts time.Time to ParsedDateTime (adapter function)
func FromTime(t time.Time) *ParsedDateTime {
	return &ParsedDateTime{
		Time:      t,
		Layout:    time.RFC3339,
		IsUTC:     t.Location() == time.UTC,
		Precision: "second",
		Original:  t.Format(time.RFC3339),
	}
}

// ToParsedDateTime converts time.Time to ParsedDateTime with layout
func ToParsedDateTime(t time.Time, layout string) *ParsedDateTime {
	return &ParsedDateTime{
		Time:      t,
		Layout:    layout,
		IsUTC:     t.Location() == time.UTC,
		Precision: detectPrecisionFromLayout(layout),
		Original:  t.Format(layout),
	}
}

// detectPrecisionFromLayout detects precision from Go time layout
func detectPrecisionFromLayout(layout string) string {
	if strings.Contains(layout, ".000000000") || strings.Contains(layout, ".999999999") {
		return "nanosecond"
	}
	if strings.Contains(layout, ":05") {
		return "second"
	}
	if strings.Contains(layout, ":04") {
		return "minute"
	}
	if strings.Contains(layout, "15") {
		return "hour"
	}
	if strings.Contains(layout, "02") || strings.Contains(layout, "2") {
		return "day"
	}
	if strings.Contains(layout, "01") || strings.Contains(layout, "1") {
		return "month"
	}
	return "year"
}

// Additional utility functions for enhanced compatibility

// ParseAnyFormat tries to parse with any of the provided formats
func ParseAnyFormat(datestr string, formats []string, loc *time.Location) (time.Time, string, error) {
	if loc == nil {
		loc = time.UTC
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, datestr, loc); err == nil {
			return t, format, nil
		}
	}

	// Fallback to advanced parsing
	t, err := ParseAny(datestr)
	if err != nil {
		return time.Time{}, "", err
	}

	// Try to detect the format
	layout, _ := ParseFormat(datestr)
	return t, layout, nil
}

// IsValidDate checks if a date string can be parsed
func IsValidDate(datestr string) bool {
	_, err := ParseAny(datestr)
	return err == nil
}

// GetSupportedFormats returns list of supported date formats
func GetSupportedFormats() []string {
	return []string{
		// ISO 8601
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05.000000Z07:00",
		"2006-01-02T15:04:05.000000000Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",

		// RFC formats
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,

		// Common formats
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",

		// US formats
		"01/02/2006 15:04:05",
		"01/02/2006 15:04",
		"01/02/2006",
		"01-02-2006 15:04:05",
		"01-02-2006 15:04",
		"01-02-2006",

		// European formats
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006",
		"02-01-2006 15:04:05",
		"02-01-2006 15:04",
		"02-01-2006",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",

		// Text formats
		"Jan 2, 2006 15:04:05",
		"Jan 2, 2006",
		"January 2, 2006 15:04:05",
		"January 2, 2006",
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan _2 15:04:05 2006",
	}
}
