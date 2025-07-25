// Package main demonstrates basic datetime parsing functionality
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
	fmt.Println("=== DateTime Parser Basic Usage Examples ===")
	fmt.Println()

	demonstrateBasicParsing()
	demonstrateFormatDetection()
	demonstrateTimezoneParsing()
	demonstrateFormatting()
	demonstratePrecisionControl()
}

func demonstrateBasicParsing() {
	fmt.Println("1. Basic DateTime Parsing:")

	parser := datetime.NewParser()
	ctx := context.Background()

	// Test different input formats
	inputs := []string{
		"2025-01-15 14:30:00",
		"2025-01-15T14:30:00Z",
		"15/01/2025 14:30",
		"Jan 15, 2025 2:30 PM",
		"2025-01-15",
		"14:30:00",
	}

	for _, input := range inputs {
		result, err := parser.ParseString(ctx, input)
		if err != nil {
			fmt.Printf("  %-25s -> ERROR: %v\n", input, err)
		} else {
			fmt.Printf("  %-25s -> %s (Layout: %s)\n",
				input, result.Time.Format(time.RFC3339), result.Layout)
		}
	}

	fmt.Println()
}

func demonstrateFormatDetection() {
	fmt.Println("2. Automatic Format Detection:")

	parser := datetime.NewParser()
	ctx := context.Background()

	// Various formats that should be auto-detected
	testCases := []struct {
		input       string
		description string
	}{
		{"2025-01-15T14:30:00Z", "ISO 8601 UTC"},
		{"2025-01-15T14:30:00-03:00", "ISO 8601 with timezone"},
		{"2025-01-15 14:30:00", "Standard format"},
		{"15/01/2025 14:30:00", "Brazilian format"},
		{"01/15/2025 2:30:00 PM", "US format with AM/PM"},
		{"15.01.2025 14:30", "European format"},
		{"Jan 15, 2025", "Month name format"},
		{"2025-01-15", "Date only"},
		{"14:30:00", "Time only"},
	}

	for _, tc := range testCases {
		result, err := parser.ParseString(ctx, tc.input)
		if err != nil {
			fmt.Printf("  %-30s: ERROR - %v\n", tc.description, err)
		} else {
			fmt.Printf("  %-30s: %s\n", tc.description,
				result.Time.Format("2006-01-02 15:04:05 MST"))
			fmt.Printf("  %-30s  Layout: %s, Precision: %s\n",
				"", result.Layout, result.Precision)
		}
	}

	fmt.Println()
}

func demonstrateTimezoneParsing() {
	fmt.Println("3. Timezone Handling:")

	parser := datetime.NewParser()
	ctx := context.Background()

	// Test timezone-aware parsing
	timezoneInputs := []string{
		"2025-01-15T14:30:00Z",      // UTC
		"2025-01-15T14:30:00-03:00", // São Paulo time
		"2025-01-15T14:30:00+09:00", // Tokyo time
		"2025-01-15 14:30:00 UTC",   // UTC explicit
		"2025-01-15 14:30:00 EST",   // EST
	}

	for _, input := range timezoneInputs {
		result, err := parser.ParseString(ctx, input)
		if err != nil {
			fmt.Printf("  %-30s -> ERROR: %v\n", input, err)
		} else {
			fmt.Printf("  %-30s -> %s (UTC: %v)\n",
				input,
				result.Time.Format("2006-01-02 15:04:05 MST"),
				result.IsUTC)

			// Show in different timezones
			sp, _ := time.LoadLocation("America/Sao_Paulo")
			tokyo, _ := time.LoadLocation("Asia/Tokyo")

			fmt.Printf("  %-30s    São Paulo: %s\n", "",
				result.Time.In(sp).Format("2006-01-02 15:04:05 MST"))
			fmt.Printf("  %-30s    Tokyo: %s\n", "",
				result.Time.In(tokyo).Format("2006-01-02 15:04:05 MST"))
		}
		fmt.Println()
	}
}

func demonstrateFormatting() {
	fmt.Println("4. DateTime Formatting:")

	parser := datetime.NewParser()
	ctx := context.Background()

	// Parse a datetime
	input := "2025-01-15T14:30:00Z"
	result, err := parser.ParseString(ctx, input)
	if err != nil {
		log.Printf("Error parsing datetime: %v", err)
		return
	}

	fmt.Printf("Original: %s\n", input)
	fmt.Printf("Parsed time: %s\n", result.Time.String())
	fmt.Println("\nFormatted in different styles:")

	// Different output formats
	formats := []struct {
		name   string
		layout string
	}{
		{"ISO 8601", time.RFC3339},
		{"RFC 1123", time.RFC1123},
		{"Brazilian", "02/01/2006 15:04:05"},
		{"US Format", "01/02/2006 3:04:05 PM"},
		{"European", "02.01.2006 15:04:05"},
		{"Custom", "Monday, January 2, 2006 at 3:04 PM"},
		{"Date Only", "2006-01-02"},
		{"Time Only", "15:04:05"},
		{"Year/Month", "2006-01"},
	}

	for _, format := range formats {
		formatted := result.Time.Format(format.layout)
		fmt.Printf("  %-15s: %s\n", format.name, formatted)
	}

	// Show timezone conversions
	fmt.Println("\nTimezone conversions:")
	locations := []string{
		"UTC",
		"America/Sao_Paulo",
		"America/New_York",
		"Europe/London",
		"Asia/Tokyo",
	}

	for _, locName := range locations {
		loc, err := time.LoadLocation(locName)
		if err != nil {
			fmt.Printf("  %-20s: ERROR - %v\n", locName, err)
			continue
		}

		converted := result.Time.In(loc)
		fmt.Printf("  %-20s: %s\n", locName,
			converted.Format("2006-01-02 15:04:05 MST"))
	}

	fmt.Println()
}

func demonstratePrecisionControl() {
	fmt.Println("5. Precision Control:")

	parser := datetime.NewParser()
	ctx := context.Background()

	input := "2025-01-15T14:30:45.123456789Z"

	result, err := parser.ParseString(ctx, input)
	if err != nil {
		log.Printf("Error parsing datetime: %v", err)
		return
	}

	fmt.Printf("Original: %s\n", input)
	fmt.Printf("Detected precision: %s\n", result.Precision)
	fmt.Printf("Full precision: %s\n", result.Time.Format(time.RFC3339Nano))

	// Show different precision levels
	fmt.Println("\nDifferent precision levels:")
	precisions := []struct {
		name   string
		layout string
	}{
		{"Year", "2006"},
		{"Month", "2006-01"},
		{"Day", "2006-01-02"},
		{"Hour", "2006-01-02 15"},
		{"Minute", "2006-01-02 15:04"},
		{"Second", "2006-01-02 15:04:05"},
		{"Millisecond", "2006-01-02 15:04:05.000"},
		{"Microsecond", "2006-01-02 15:04:05.000000"},
		{"Nanosecond", "2006-01-02 15:04:05.000000000"},
	}

	for _, prec := range precisions {
		formatted := result.Time.Format(prec.layout)
		fmt.Printf("  %-15s: %s\n", prec.name, formatted)
	}

	fmt.Println()
}
