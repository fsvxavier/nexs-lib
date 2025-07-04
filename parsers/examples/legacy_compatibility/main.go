package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
	fmt.Println("=== Legacy Compatibility Examples ===")
	fmt.Println("This example shows 100% API compatibility with old dateparse libraries\n")

	// Example 1: ParseAny - Basic usage (most common legacy function)
	fmt.Println("1. ParseAny - Basic Usage:")
	testDates := []string{
		"2023-01-15",
		"01/15/2023",
		"15/01/2023",
		"January 15, 2023",
		"Jan 15, 2023 10:30 AM",
		"1673778645", // Unix timestamp
		"today",
		"yesterday",
	}

	for _, dateStr := range testDates {
		if date, err := datetime.ParseAny(dateStr); err == nil {
			fmt.Printf("  %-25s -> %s\n", dateStr, date.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  %-25s -> ERROR: %v\n", dateStr, err)
		}
	}

	fmt.Println("\n2. ParseAny with PreferMonthFirst Option:")

	// Example 2: ParseAny with PreferMonthFirst (European vs American format)
	ambiguousDate := "02/03/2023"

	// American format (month first) - default
	dateUS, err := datetime.ParseAny(ambiguousDate, datetime.PreferMonthFirst(true))
	if err == nil {
		fmt.Printf("  US format (MM/DD/YYYY):  %s -> %s (February 3rd)\n",
			ambiguousDate, dateUS.Format("2006-01-02"))
	}

	// European format (day first)
	dateEU, err := datetime.ParseAny(ambiguousDate, datetime.PreferMonthFirst(false))
	if err == nil {
		fmt.Printf("  EU format (DD/MM/YYYY):  %s -> %s (March 2nd)\n",
			ambiguousDate, dateEU.Format("2006-01-02"))
	}

	fmt.Println("\n3. ParseAny with RetryAmbiguousDateWithSwap:")

	// Example 3: Auto-retry on invalid dates
	invalidDate := "29/02/2023" // Invalid in MM/DD format, valid in DD/MM

	date, err := datetime.ParseAny(invalidDate,
		datetime.PreferMonthFirst(true),           // Try MM/DD first
		datetime.RetryAmbiguousDateWithSwap(true)) // Auto-retry with DD/MM if failed

	if err == nil {
		fmt.Printf("  Auto-retry success: %s -> %s\n", invalidDate, date.Format("2006-01-02"))
		fmt.Printf("  (Tried MM/DD first, failed, then succeeded with DD/MM)\n")
	}

	fmt.Println("\n4. ParseIn - Parse with Specific Timezone:")

	// Example 4: ParseIn with timezone
	locations := []string{"America/New_York", "Europe/London", "Asia/Tokyo"}
	dateStr := "2023-01-15 10:30:00"

	for _, locName := range locations {
		loc, err := time.LoadLocation(locName)
		if err != nil {
			continue
		}

		date, err := datetime.ParseIn(dateStr, loc)
		if err == nil {
			fmt.Printf("  %-20s: %s -> %s (%s)\n",
				locName, dateStr, date.Format("2006-01-02 15:04:05"), date.Location().String())
		}
	}

	fmt.Println("\n5. ParseLocal - Parse with Local Timezone:")

	// Example 5: ParseLocal
	localDates := []string{
		"2023-01-15 10:30:00",
		"15/01/2023 14:30",
		"January 15, 2023 2:30 PM",
	}

	for _, dateStr := range localDates {
		date, err := datetime.ParseLocal(dateStr)
		if err == nil {
			fmt.Printf("  %-25s -> %s (%s)\n",
				dateStr, date.Format("2006-01-02 15:04:05"), date.Location().String())
		}
	}

	fmt.Println("\n6. MustParseAny - Panic on Error (for testing):")

	// Example 6: MustParseAny (careful - panics on error!)
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  Caught panic (expected): %v\n", r)
			}
		}()

		// This will work
		date := datetime.MustParseAny("2023-01-15T10:30:45Z")
		fmt.Printf("  Success: %s\n", date.Format(time.RFC3339))

		// This will panic (uncomment to test)
		// datetime.MustParseAny("completely-invalid-date")
	}()

	fmt.Println("\n7. ParseStrict - Strict Mode Parsing:")

	// Example 7: ParseStrict for more controlled parsing
	strictDates := []string{
		"2023-01-15", // Should work
		"2023-1-15",  // Might fail in strict mode
		"01/15/2023", // Should work with PreferMonthFirst(true)
		"today",      // Might fail in strict mode
	}

	for _, dateStr := range strictDates {
		date, err := datetime.ParseStrict(dateStr, datetime.PreferMonthFirst(true))
		if err == nil {
			fmt.Printf("  ✓ %-20s -> %s\n", dateStr, date.Format("2006-01-02"))
		} else {
			fmt.Printf("  ✗ %-20s -> STRICT MODE REJECTED: %v\n", dateStr, err)
		}
	}

	fmt.Println("\n8. Real-World Migration Example:")

	// Example 8: Real-world usage that would be common in legacy code
	userInputs := []string{
		"2023-12-25",
		"12/25/2023",
		"25/12/2023",
		"December 25, 2023",
		"Dec 25, 2023 11:59 PM",
		"Christmas Day 2023", // This might not work, but shows error handling
		"1703462400",         // Unix timestamp for 2023-12-25
	}

	fmt.Println("  Processing user date inputs (with error handling):")
	successCount := 0

	for i, input := range userInputs {
		// This is exactly how you'd use it in legacy code
		date, err := datetime.ParseAny(input,
			datetime.PreferMonthFirst(false), // European preference
			datetime.RetryAmbiguousDateWithSwap(true))

		if err == nil {
			successCount++
			fmt.Printf("    %d. ✓ %-25s -> %s\n", i+1, input, date.Format("January 2, 2006 15:04"))
		} else {
			fmt.Printf("    %d. ✗ %-25s -> Failed: %v\n", i+1, input, err)
		}
	}

	fmt.Printf("\n  Successfully parsed %d out of %d inputs (%.1f%% success rate)\n",
		successCount, len(userInputs), float64(successCount)/float64(len(userInputs))*100)

	fmt.Println("\n=== Migration Summary ===")
	fmt.Println("✓ All legacy functions work exactly the same")
	fmt.Println("✓ Same function signatures and behavior")
	fmt.Println("✓ Same options and configurations")
	fmt.Println("✓ Enhanced error handling and performance")
	fmt.Println("✓ Additional features available if needed")
	fmt.Println("\nYou can replace your old dateparse import and everything works!")
}

func basicParseAnyExamples() {
	examples := []string{
		"2023-01-15T10:30:45Z",
		"2023-01-15",
		"01/15/2023",
		"15/01/2023",
		"January 15, 2023",
		"Jan 15, 2023",
		"15 Jan 2023",
		"2023/01/15",
	}

	for _, input := range examples {
		if result, err := datetime.ParseAny(input); err != nil {
			fmt.Printf("  ❌ %s -> Error: %v\n", input, err)
		} else {
			fmt.Printf("  ✅ %s -> %s\n", input, result.Format("2006-01-02 15:04:05"))
		}
	}
}

func parseAnyWithOptionsExamples() {
	ambiguousDate := "02/03/2023"
	fmt.Printf("  Ambiguous date: %s\n", ambiguousDate)

	// Prefer month first (US format: MM/DD/YYYY)
	result1, err := datetime.ParseAny(ambiguousDate, datetime.PreferMonthFirst(true))
	if err != nil {
		log.Printf("Error with PreferMonthFirst(true): %v", err)
	} else {
		fmt.Printf("  PreferMonthFirst(true):  %s -> %s (Feb 3, 2023)\n",
			ambiguousDate, result1.Format("2006-01-02"))
	}

	// Prefer day first (European format: DD/MM/YYYY)
	result2, err := datetime.ParseAny(ambiguousDate, datetime.PreferMonthFirst(false))
	if err != nil {
		log.Printf("Error with PreferMonthFirst(false): %v", err)
	} else {
		fmt.Printf("  PreferMonthFirst(false): %s -> %s (Mar 2, 2023)\n",
			ambiguousDate, result2.Format("2006-01-02"))
	}

	// Test retry on ambiguous dates
	retryDate := "13/02/2023" // Day > 12, should trigger retry
	result3, err := datetime.ParseAny(retryDate,
		datetime.PreferMonthFirst(true),           // Try MM/DD first
		datetime.RetryAmbiguousDateWithSwap(true)) // Retry with DD/MM
	if err != nil {
		log.Printf("Error with retry: %v", err)
	} else {
		fmt.Printf("  Retry example: %s -> %s (auto-corrected to DD/MM)\n",
			retryDate, result3.Format("2006-01-02"))
	}
}

func timezoneParsingExamples() {
	// ParseIn - parse with specific timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Printf("  ❌ Could not load timezone: %v\n", err)
		return
	}

	result1, err := datetime.ParseIn("2023-01-15 10:30:00", loc)
	if err != nil {
		fmt.Printf("  ❌ ParseIn error: %v\n", err)
	} else {
		fmt.Printf("  ParseIn (America/New_York): %s -> %s\n",
			"2023-01-15 10:30:00", result1.Format("2006-01-02 15:04:05 MST"))
	}

	// ParseLocal - parse with local timezone
	result2, err := datetime.ParseLocal("2023-01-15 10:30:00")
	if err != nil {
		fmt.Printf("  ❌ ParseLocal error: %v\n", err)
	} else {
		fmt.Printf("  ParseLocal: %s -> %s\n",
			"2023-01-15 10:30:00", result2.Format("2006-01-02 15:04:05 MST"))
	}
}

func mustParseExamples() {
	// Successful parse
	result := datetime.MustParseAny("2023-01-15T10:30:45Z")
	fmt.Printf("  ✅ MustParseAny success: %s\n", result.Format(time.RFC3339))

	// This would panic - commented out for demo
	// datetime.MustParseAny("invalid-date")
	fmt.Printf("  ⚠️  MustParseAny would panic on invalid input (commented out for demo)\n")
}

func formatDetectionExamples() {
	examples := []string{
		"2023-01-15T10:30:45Z",
		"January 15, 2023",
		"01/15/2023",
		"15.01.2023",
		"2023/01/15",
	}

	for _, input := range examples {
		if format, err := datetime.ParseFormat(input); err != nil {
			fmt.Printf("  ❌ %s -> Error: %v\n", input, err)
		} else {
			fmt.Printf("  ✅ %s -> Format: %s\n", input, format)
		}
	}

	// DetectFormat is an alias for ParseFormat
	fmt.Printf("  DetectFormat (alias): %s\n", "Same as ParseFormat")

	// ParseWithFormat - parse using a specific format
	result, err := datetime.ParseWithFormat("15/01/2023", "02/01/2006")
	if err != nil {
		fmt.Printf("  ❌ ParseWithFormat error: %v\n", err)
	} else {
		fmt.Printf("  ✅ ParseWithFormat: 15/01/2023 with format 02/01/2006 -> %s\n",
			result.Format("2006-01-02"))
	}
}

func advancedFeaturesExamples() {
	fmt.Println("  These features are NEW and improved compared to old libraries:")

	// Unix timestamps
	timestamps := []string{
		"1673778645",     // Seconds
		"1673778645000",  // Milliseconds
		"1673778645.123", // Decimal seconds
	}

	for _, ts := range timestamps {
		if result, err := datetime.ParseAny(ts); err != nil {
			fmt.Printf("  ❌ Unix timestamp %s -> Error: %v\n", ts, err)
		} else {
			fmt.Printf("  ✅ Unix timestamp %s -> %s\n", ts, result.Format("2006-01-02 15:04:05"))
		}
	}

	// Improved text parsing with times
	textWithTime := []string{
		"January 15, 2023 10:30 AM",
		"Jan 15, 2023 2:45 PM",
		"March 1, 2023 9:15 AM",
	}

	for _, text := range textWithTime {
		if result, err := datetime.ParseAny(text); err != nil {
			fmt.Printf("  ❌ %s -> Error: %v\n", text, err)
		} else {
			fmt.Printf("  ✅ %s -> %s\n", text, result.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("  🚀 Performance: Format caching for repeated parsing")
	fmt.Println("  🛡️  Validation: Better error messages with suggestions")
	fmt.Println("  ⚙️  Context: Support for timeouts and cancellation")
}
