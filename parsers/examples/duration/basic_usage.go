// Package main demonstrates basic duration parsing functionality
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/duration"
)

func main() {
	fmt.Println("=== Duration Parser Basic Usage Examples ===")
	fmt.Println()

	demonstrateBasicParsing()
	demonstrateExtendedUnits()
	demonstrateFormatting()
	demonstrateConversions()
	demonstrateValidation()
}

func demonstrateBasicParsing() {
	fmt.Println("1. Basic Duration Parsing:")

	parser := duration.NewParser()
	ctx := context.Background()

	// Test standard Go duration formats
	standardInputs := []string{
		"1h",
		"30m",
		"45s",
		"500ms",
		"1h30m",
		"2h45m30s",
		"1h30m45s500ms",
		"0.5h",
		"1.5m",
		"2.5s",
	}

	for _, input := range standardInputs {
		result, err := parser.ParseString(ctx, input)
		if err != nil {
			fmt.Printf("  %-20s -> ERROR: %v\n", input, err)
		} else {
			fmt.Printf("  %-20s -> %v (%s)\n",
				input, result.Duration, result.Duration.String())
		}
	}

	fmt.Println()
}

func demonstrateExtendedUnits() {
	fmt.Println("2. Extended Units (Days and Weeks):")

	parser := duration.NewParser()
	ctx := context.Background()

	// Test extended units that are not in standard Go
	extendedInputs := []string{
		"1d",         // 1 day
		"1w",         // 1 week
		"2d12h",      // 2 days and 12 hours
		"1w3d",       // 1 week and 3 days
		"2w1d12h30m", // 2 weeks, 1 day, 12 hours, 30 minutes
		"0.5d",       // Half day (12 hours)
		"1.5w",       // 1.5 weeks
		"10d",        // 10 days
	}

	for _, input := range extendedInputs {
		result, err := parser.ParseString(ctx, input)
		if err != nil {
			fmt.Printf("  %-15s -> ERROR: %v\n", input, err)
		} else {
			fmt.Printf("  %-15s -> %v\n", input, result.Duration)
			fmt.Printf("  %-15s    Units found: %v\n", "", result.Units)
			fmt.Printf("  %-15s    In hours: %.1f\n", "", result.Duration.Hours())
		}
		fmt.Println()
	}
}

func demonstrateFormatting() {
	fmt.Println("3. Duration Formatting:")

	parser := duration.NewParser()
	ctx := context.Background()

	// Parse some durations and show different formatting
	testDurations := []string{
		"1h30m45s",
		"2d12h30m",
		"1w3d6h",
		"90m",
		"36h",
	}

	for _, input := range testDurations {
		result, err := parser.ParseString(ctx, input)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", input, err)
			continue
		}

		fmt.Printf("Original: %s\n", input)
		fmt.Printf("  String():        %s\n", result.Duration.String())
		fmt.Printf("  Nanoseconds():   %d ns\n", result.Duration.Nanoseconds())
		fmt.Printf("  Microseconds():  %d µs\n", result.Duration.Microseconds())
		fmt.Printf("  Milliseconds():  %d ms\n", result.Duration.Milliseconds())
		fmt.Printf("  Seconds():       %.2f s\n", result.Duration.Seconds())
		fmt.Printf("  Minutes():       %.2f m\n", result.Duration.Minutes())
		fmt.Printf("  Hours():         %.2f h\n", result.Duration.Hours())

		// Calculate days and weeks manually
		days := result.Duration.Hours() / 24
		weeks := days / 7
		fmt.Printf("  Days:            %.2f d\n", days)
		fmt.Printf("  Weeks:           %.2f w\n", weeks)
		fmt.Println()
	}
}

func demonstrateConversions() {
	fmt.Println("4. Duration Conversions:")

	parser := duration.NewParser()
	ctx := context.Background()

	// Parse a complex duration
	input := "1w2d3h4m5s"
	result, err := parser.ParseString(ctx, input)
	if err != nil {
		log.Printf("Error parsing duration: %v", err)
		return
	}

	fmt.Printf("Original duration: %s\n", input)
	fmt.Printf("Parsed as: %v\n", result.Duration)
	fmt.Println()

	// Show breakdown
	totalNs := result.Duration.Nanoseconds()

	weeks := totalNs / (7 * 24 * int64(time.Hour))
	remaining := totalNs % (7 * 24 * int64(time.Hour))

	days := remaining / (24 * int64(time.Hour))
	remaining = remaining % (24 * int64(time.Hour))

	hours := remaining / int64(time.Hour)
	remaining = remaining % int64(time.Hour)

	minutes := remaining / int64(time.Minute)
	remaining = remaining % int64(time.Minute)

	seconds := remaining / int64(time.Second)

	fmt.Println("Breakdown:")
	fmt.Printf("  %d weeks\n", weeks)
	fmt.Printf("  %d days\n", days)
	fmt.Printf("  %d hours\n", hours)
	fmt.Printf("  %d minutes\n", minutes)
	fmt.Printf("  %d seconds\n", seconds)
	fmt.Println()

	// Show various representations
	fmt.Println("Different representations:")
	fmt.Printf("  Total seconds:      %.0f\n", result.Duration.Seconds())
	fmt.Printf("  Total minutes:      %.1f\n", result.Duration.Minutes())
	fmt.Printf("  Total hours:        %.2f\n", result.Duration.Hours())
	fmt.Printf("  Total days:         %.3f\n", result.Duration.Hours()/24)
	fmt.Printf("  Total weeks:        %.4f\n", result.Duration.Hours()/(24*7))

	// Compare with different time units
	fmt.Println("\nComparisons:")
	fmt.Printf("  Equivalent to %d hours and %.0f minutes\n",
		int(result.Duration.Hours()),
		result.Duration.Minutes()-float64(int(result.Duration.Hours()))*60)

	workingDays := result.Duration.Hours() / 8 // 8-hour work days
	fmt.Printf("  Equivalent to %.1f working days (8h each)\n", workingDays)

	fmt.Println()
}

func demonstrateValidation() {
	fmt.Println("5. Duration Validation:")

	parser := duration.NewParser()
	ctx := context.Background()

	// Test valid and invalid inputs
	testCases := []struct {
		input       string
		description string
	}{
		{"1h30m", "Valid standard format"},
		{"2d12h", "Valid with days"},
		{"1w3d6h", "Valid with weeks"},
		{"abc", "Invalid - not a duration"},
		{"1x", "Invalid - unknown unit"},
		{"", "Invalid - empty string"},
		{"-1h", "Invalid - negative duration"},
		{"1h30x45s", "Invalid - mixed valid/invalid"},
		{"1.5h30m", "Valid - decimal hours with minutes"},
		{"0s", "Valid - zero duration"},
	}

	for _, tc := range testCases {
		result, err := parser.ParseString(ctx, tc.input)
		if err != nil {
			fmt.Printf("  %-30s: ✗ %v\n", tc.description, err)
		} else {
			fmt.Printf("  %-30s: ✓ %v\n", tc.description, result.Duration)
		}
	}

	fmt.Println()
}
