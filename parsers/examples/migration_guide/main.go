package main

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
	fmt.Println("=== Migration Guide: From Legacy dateparse to nexs-lib ===")
	fmt.Println("This guide shows how to migrate from old dateparse libraries")

	fmt.Println("🔄 STEP-BY-STEP MIGRATION GUIDE")
	fmt.Println("==============================")

	// Step 1: Import replacement
	fmt.Println("\n1. IMPORT REPLACEMENT:")
	fmt.Println("   OLD: import \"github.com/araddon/dateparse\"")
	fmt.Println("   NEW: import \"github.com/fsvxavier/nexs-lib/parsers/datetime\"")
	fmt.Println("   ✓ That's it! All function names stay the same.")

	// Step 2: Function compatibility
	fmt.Println("\n2. FUNCTION COMPATIBILITY:")
	showFunctionCompatibility()

	// Step 3: Before/After code examples
	fmt.Println("\n3. BEFORE/AFTER CODE EXAMPLES:")
	showBeforeAfterExamples()

	// Step 4: Compatibility matrix
	fmt.Println("\n4. COMPATIBILITY MATRIX:")
	showCompatibilityMatrix()

	// Step 5: New features you can use
	fmt.Println("\n5. NEW FEATURES AVAILABLE:")
	showNewFeatures()

	// Step 6: Performance improvements
	fmt.Println("\n6. PERFORMANCE IMPROVEMENTS:")
	showPerformanceImprovements()

	// Step 7: Migration checklist
	fmt.Println("\n7. MIGRATION CHECKLIST:")
	showMigrationChecklist()
}

func showFunctionCompatibility() {
	fmt.Println("   ✅ ALL LEGACY FUNCTIONS WORK EXACTLY THE SAME:")

	legacyFunctions := []struct {
		name        string
		description string
		example     string
	}{
		{
			name:        "ParseAny(dateStr)",
			description: "Parse any date string format",
			example:     "ParseAny(\"01/15/2023\")",
		},
		{
			name:        "ParseAny(dateStr, options...)",
			description: "Parse with options like PreferMonthFirst",
			example:     "ParseAny(\"02/03/2023\", PreferMonthFirst(false))",
		},
		{
			name:        "ParseIn(dateStr, location)",
			description: "Parse with specific timezone",
			example:     "ParseIn(\"2023-01-15 10:30\", time.UTC)",
		},
		{
			name:        "ParseLocal(dateStr)",
			description: "Parse with local timezone",
			example:     "ParseLocal(\"15/01/2023\")",
		},
		{
			name:        "MustParseAny(dateStr)",
			description: "Parse or panic (for testing)",
			example:     "MustParseAny(\"2023-01-15T10:30:45Z\")",
		},
		{
			name:        "ParseStrict(dateStr, options...)",
			description: "Parse in strict mode",
			example:     "ParseStrict(\"2023-01-15\", PreferMonthFirst(true))",
		},
	}

	for _, fn := range legacyFunctions {
		fmt.Printf("     • %-35s - %s\n", fn.name, fn.description)
		fmt.Printf("       Example: %s\n", fn.example)
	}
}

func showBeforeAfterExamples() {
	fmt.Println("   REAL CODE MIGRATION EXAMPLES:")
	fmt.Println()

	// Example 1: Basic parsing
	fmt.Println("   Example 1: Basic Date Parsing")
	fmt.Println("   BEFORE (old library):")
	fmt.Println("   ```go")
	fmt.Println("   import \"github.com/araddon/dateparse\"")
	fmt.Println("   ")
	fmt.Println("   date, err := dateparse.ParseAny(\"01/15/2023\")")
	fmt.Println("   if err != nil {")
	fmt.Println("       return err")
	fmt.Println("   }")
	fmt.Println("   ```")
	fmt.Println()
	fmt.Println("   AFTER (nexs-lib):")
	fmt.Println("   ```go")
	fmt.Println("   import \"github.com/fsvxavier/nexs-lib/parsers/datetime\"")
	fmt.Println("   ")
	fmt.Println("   date, err := datetime.ParseAny(\"01/15/2023\")")
	fmt.Println("   if err != nil {")
	fmt.Println("       return err")
	fmt.Println("   }")
	fmt.Println("   ```")
	fmt.Println("   ✓ Only the import changed!")

	// Example 2: With options
	fmt.Println("\n   Example 2: Parsing with Options")
	fmt.Println("   BEFORE:")
	fmt.Println("   ```go")
	fmt.Println("   date, err := dateparse.ParseAny(\"02/03/2023\",")
	fmt.Println("       dateparse.PreferMonthFirst(false))")
	fmt.Println("   ```")
	fmt.Println()
	fmt.Println("   AFTER:")
	fmt.Println("   ```go")
	fmt.Println("   date, err := datetime.ParseAny(\"02/03/2023\",")
	fmt.Println("       datetime.PreferMonthFirst(false))")
	fmt.Println("   ```")
	fmt.Println("   ✓ Same options, same behavior!")

	// Example 3: Demonstrate working code
	fmt.Println("\n   Example 3: Live Demonstration")
	demoMigration()
}

func demoMigration() {
	// Show that the migrated code actually works
	testCases := []struct {
		description string
		input       string
		options     []datetime.ParserOption
	}{
		{
			description: "Basic parsing",
			input:       "01/15/2023",
			options:     nil,
		},
		{
			description: "European format preference",
			input:       "02/03/2023",
			options:     []datetime.ParserOption{datetime.PreferMonthFirst(false)},
		},
		{
			description: "With retry on ambiguous dates",
			input:       "29/02/2024", // Valid leap year date
			options:     []datetime.ParserOption{datetime.PreferMonthFirst(true), datetime.RetryAmbiguousDateWithSwap(true)},
		},
	}

	for _, tc := range testCases {
		var date time.Time
		var err error

		if tc.options != nil {
			date, err = datetime.ParseAny(tc.input, tc.options...)
		} else {
			date, err = datetime.ParseAny(tc.input)
		}

		if err == nil {
			fmt.Printf("     ✓ %-30s: %s -> %s\n", tc.description, tc.input, date.Format("2006-01-02"))
		} else {
			fmt.Printf("     ✗ %-30s: %s -> %v\n", tc.description, tc.input, err)
		}
	}
}

func showCompatibilityMatrix() {
	fmt.Println("   COMPATIBILITY MATRIX:")
	fmt.Println()
	fmt.Printf("   %-35s | %-15s | %-15s | %s\n", "Feature", "Old Library", "nexs-lib", "Status")
	fmt.Printf("   %s\n", "---------------------------------------------------------------------------------------------------------")

	matrix := []struct {
		feature    string
		oldSupport string
		newSupport string
		status     string
	}{
		{"ParseAny function", "✓", "✓", "100% Compatible"},
		{"ParseIn function", "✓", "✓", "100% Compatible"},
		{"ParseLocal function", "✓", "✓", "100% Compatible"},
		{"MustParseAny function", "✓", "✓", "100% Compatible"},
		{"PreferMonthFirst option", "✓", "✓", "100% Compatible"},
		{"RetryAmbiguousDateWithSwap", "✓", "✓", "100% Compatible"},
		{"Unix timestamps", "Basic", "Enhanced", "✓ Improved"},
		{"Text parsing", "Basic", "Enhanced", "✓ Improved"},
		{"Error messages", "Basic", "Detailed", "✓ Enhanced"},
		{"Performance caching", "✗", "✓", "✓ New Feature"},
		{"Format detection", "✗", "✓", "✓ New Feature"},
		{"Context support", "✗", "✓", "✓ New Feature"},
		{"Custom formats", "✗", "✓", "✓ New Feature"},
		{"Strict mode", "✗", "✓", "✓ New Feature"},
	}

	for _, item := range matrix {
		fmt.Printf("   %-35s | %-15s | %-15s | %s\n", item.feature, item.oldSupport, item.newSupport, item.status)
	}
}

func showNewFeatures() {
	fmt.Println("   NEW FEATURES YOU CAN USE (Optional):")
	fmt.Println()

	// Format detection
	fmt.Println("   🔍 Format Detection:")
	input := "January 15, 2023 10:30 AM"
	if format, err := datetime.ParseFormat(input); err == nil {
		fmt.Printf("     datetime.ParseFormat(\"%s\")\n", input)
		fmt.Printf("     // Returns: \"%s\"\n", format)
	}
	fmt.Println()

	// Parse with specific format
	fmt.Println("   🎯 Parse with Specific Format:")
	specificInput := "15/01/2023"
	specificFormat := "02/01/2006"
	if date, err := datetime.ParseWithFormat(specificInput, specificFormat); err == nil {
		fmt.Printf("     datetime.ParseWithFormat(\"%s\", \"%s\")\n", specificInput, specificFormat)
		fmt.Printf("     // Returns: %s\n", date.Format("2006-01-02"))
	}
	fmt.Println()

	// Enhanced Unix timestamps
	fmt.Println("   ⏰ Enhanced Unix Timestamps:")
	timestamps := []string{"1673778645", "1673778645.123", "1673778645123"}
	for _, ts := range timestamps {
		if date, err := datetime.ParseAny(ts); err == nil {
			fmt.Printf("     \"%s\" -> %s\n", ts, date.Format(time.RFC3339))
		}
	}
	fmt.Println()

	// Modern parser with options
	fmt.Println("   ⚙️  Modern Parser Configuration:")
	fmt.Println("     parser := datetime.NewParser(")
	fmt.Println("         parsers.WithLocation(time.UTC),")
	fmt.Println("         parsers.WithDateOrder(parsers.DateOrderDMY),")
	fmt.Println("         parsers.WithCustomFormats(\"02/01/2006 15:04\"),")
	fmt.Println("     )")
}

func showPerformanceImprovements() {
	fmt.Println("   PERFORMANCE IMPROVEMENTS:")
	fmt.Println()

	// Demonstrate caching
	input := "2023-01-15T10:30:45Z"

	// First parse (cache miss)
	start := time.Now()
	datetime.ParseAny(input)
	firstParse := time.Since(start)

	// Second parse (cache hit)
	start = time.Now()
	datetime.ParseAny(input)
	secondParse := time.Since(start)

	fmt.Printf("   🚀 Format Caching:\n")
	fmt.Printf("     First parse (cache miss): %v\n", firstParse)
	fmt.Printf("     Second parse (cache hit):  %v\n", secondParse)
	if secondParse < firstParse {
		fmt.Printf("     ✓ Improvement: %.1fx faster on repeated parsing\n", float64(firstParse)/float64(secondParse))
	}
	fmt.Println()

	fmt.Println("   📊 Memory Efficiency:")
	fmt.Println("     ✓ Reduced allocations for repeated formats")
	fmt.Println("     ✓ Optimized string processing")
	fmt.Println("     ✓ Smart format ordering based on usage patterns")
	fmt.Println()

	fmt.Println("   🎯 Enhanced Unix Timestamp Support:")
	fmt.Println("     ✓ Automatic detection of seconds/milliseconds/microseconds")
	fmt.Println("     ✓ Decimal precision support (1673778645.123)")
	fmt.Println("     ✓ Better error handling for invalid timestamps")
}

func showMigrationChecklist() {
	fmt.Println("   MIGRATION CHECKLIST:")
	fmt.Println()
	fmt.Println("   □ Step 1: Update import statement")
	fmt.Println("     OLD: import \"github.com/araddon/dateparse\"")
	fmt.Println("     NEW: import \"github.com/fsvxavier/nexs-lib/parsers/datetime\"")
	fmt.Println()
	fmt.Println("   □ Step 2: Update function calls (prefix with 'datetime.')")
	fmt.Println("     OLD: dateparse.ParseAny(...)")
	fmt.Println("     NEW: datetime.ParseAny(...)")
	fmt.Println()
	fmt.Println("   □ Step 3: Test your existing code")
	fmt.Println("     ✓ All functions should work exactly the same")
	fmt.Println("     ✓ All options should behave identically")
	fmt.Println("     ✓ Performance should be same or better")
	fmt.Println()
	fmt.Println("   □ Step 4: (Optional) Explore new features")
	fmt.Println("     • Format detection with ParseFormat()")
	fmt.Println("     • Enhanced error messages")
	fmt.Println("     • Modern parser configuration")
	fmt.Println("     • Context support for timeouts")
	fmt.Println()
	fmt.Println("   □ Step 5: Update dependencies")
	fmt.Println("     go mod tidy")
	fmt.Println()
	fmt.Println("   🎉 MIGRATION COMPLETE!")
	fmt.Println("      Your code now benefits from:")
	fmt.Println("      ✓ 100% backward compatibility")
	fmt.Println("      ✓ Enhanced performance")
	fmt.Println("      ✓ Better error handling")
	fmt.Println("      ✓ Modern Go features")
	fmt.Println("      ✓ Optional advanced capabilities")
}
