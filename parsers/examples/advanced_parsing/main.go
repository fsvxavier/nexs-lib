package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
	"github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
	fmt.Println("=== Advanced DateTime Parsing Examples ===")
	fmt.Println("This example demonstrates advanced features and modern capabilities")

	// Example 1: Modern Parser with Configuration
	fmt.Println("1. Modern Parser Configuration:")
	modernParserExamples()
	fmt.Println()

	// Example 2: Format Detection and Analysis
	fmt.Println("2. Format Detection and Analysis:")
	formatDetectionExamples()
	fmt.Println()

	// Example 3: Context and Performance
	fmt.Println("3. Context Handling and Performance:")
	contextAndPerformanceExamples()
	fmt.Println()

	// Example 4: Advanced Unix Timestamp Support
	fmt.Println("4. Advanced Unix Timestamp Support:")
	unixTimestampExamples()
	fmt.Println()

	// Example 5: Enhanced Text Parsing
	fmt.Println("5. Enhanced Text Parsing:")
	enhancedTextParsingExamples()
	fmt.Println()

	// Example 6: Error Handling and Debugging
	fmt.Println("6. Error Handling and Debugging:")
	errorHandlingExamples()
	fmt.Println()

	// Example 7: Custom Formats and Options
	fmt.Println("7. Custom Formats and Options:")
	customFormatsExamples()
	fmt.Println()

	// Example 8: Performance Benchmarking
	fmt.Println("8. Performance Optimization:")
	performanceOptimizationExamples()
}

func modernParserExamples() {
	// Create a modern parser with custom configuration
	parser := datetime.NewParser(
		parsers.WithLocation(time.UTC),
		parsers.WithDateOrder(parsers.DateOrderDMY), // European preference
		parsers.WithStrictMode(false),
		parsers.WithCustomFormats(
			"02/01/2006 15:04:05", // DD/MM/YYYY HH:MM:SS
			"2006.01.02",          // Dot-separated dates
			"02-Jan-2006",         // Military style
		),
	)

	ctx := context.Background()

	testInputs := []string{
		"15/03/2023 14:30:45",    // Should use DD/MM with our config
		"2023.03.15",             // Dot format
		"15-Mar-2023",            // Military style
		"March 15, 2023 2:30 PM", // Standard text format
	}

	for _, input := range testInputs {
		if date, err := parser.Parse(ctx, input); err == nil {
			fmt.Printf("  %-25s -> %s\n", input, date.Format(time.RFC3339))
		} else {
			fmt.Printf("  %-25s -> ERROR: %v\n", input, err)
		}
	}
}

func formatDetectionExamples() {
	parser := datetime.NewParser()
	ctx := context.Background()

	inputs := []string{
		"2023-01-15T10:30:45Z",
		"January 15, 2023 10:30 AM",
		"15/01/2023",
		"01-15-2023",
		"2023.01.15",
		"1673778645",
		"1673778645.123",
	}

	fmt.Println("  Format detection for various inputs:")
	for _, input := range inputs {
		format, err := parser.ParseFormat(ctx, input)
		if err == nil {
			fmt.Printf("    %-25s -> Format: %s\n", input, format)

			// Verify the format works
			if _, parseErr := time.Parse(format, input); parseErr == nil {
				fmt.Printf("    %25s    ✓ Format verified\n", "")
			} else {
				fmt.Printf("    %25s    ✗ Format verification failed\n", "")
			}
		} else {
			fmt.Printf("    %-25s -> Detection failed: %v\n", input, err)
		}
	}
}

func contextAndPerformanceExamples() {
	parser := datetime.NewParser()

	// Example with timeout context
	fmt.Println("  Testing context cancellation:")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate a slow operation by using a complex input
	complexInput := "some very complex date string that might take time to parse"

	start := time.Now()
	_, err := parser.Parse(ctx, complexInput)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("    Parse failed as expected: %v (took %v)\n", err, elapsed)
	}

	// Example with caching performance
	fmt.Println("  Testing format caching performance:")
	input := "2023-01-15T10:30:45Z"

	// First parse (cache miss)
	start = time.Now()
	parser.Parse(context.Background(), input)
	firstParse := time.Since(start)

	// Second parse (cache hit)
	start = time.Now()
	parser.Parse(context.Background(), input)
	secondParse := time.Since(start)

	fmt.Printf("    First parse (cache miss): %v\n", firstParse)
	fmt.Printf("    Second parse (cache hit): %v\n", secondParse)
	if secondParse < firstParse {
		fmt.Printf("    ✓ Caching improved performance by %.1fx\n", float64(firstParse)/float64(secondParse))
	}
}

func unixTimestampExamples() {
	parser := datetime.NewParser()
	ctx := context.Background()

	// Various Unix timestamp formats
	timestamps := []string{
		"1673778645",        // Seconds
		"1673778645123",     // Milliseconds
		"1673778645123456",  // Microseconds
		"1673778645.123",    // Decimal seconds
		"1673778645.123456", // Decimal microseconds
	}

	fmt.Println("  Unix timestamp parsing:")
	for _, ts := range timestamps {
		if date, err := parser.Parse(ctx, ts); err == nil {
			fmt.Printf("    %-20s -> %s\n", ts, date.Format(time.RFC3339Nano))
		} else {
			fmt.Printf("    %-20s -> ERROR: %v\n", ts, err)
		}
	}

	// Show current timestamp for reference
	now := time.Now()
	fmt.Printf("\n  Current time for reference:\n")
	fmt.Printf("    Current time: %s\n", now.Format(time.RFC3339))
	fmt.Printf("    Unix seconds: %d\n", now.Unix())
	fmt.Printf("    Unix millis:  %d\n", now.UnixMilli())
	fmt.Printf("    Unix micros:  %d\n", now.UnixMicro())
}

func enhancedTextParsingExamples() {
	parser := datetime.NewParser()
	ctx := context.Background()

	// Advanced text formats
	textFormats := []string{
		"January 15, 2023 10:30 AM",
		"15th of March 2023",
		"March 15th, 2023 at 2:30 PM",
		"Mon, 15 Jan 2023 10:30:45",
		"Monday, January 15, 2023",
		"15-Jan-2023 14:30:45",
		"2023年01月15日", // Japanese format (if supported)
	}

	fmt.Println("  Enhanced text parsing:")
	for _, text := range textFormats {
		if date, err := parser.Parse(ctx, text); err == nil {
			fmt.Printf("    ✓ %-35s -> %s\n", text, date.Format("January 2, 2006 15:04"))
		} else {
			fmt.Printf("    ✗ %-35s -> %v\n", text, err)
		}
	}

	// Relative date parsing
	fmt.Println("\n  Relative date parsing:")
	relativeDates := []string{
		"today",
		"yesterday",
		"tomorrow",
		"2 days ago",
		"next week",
		"last month",
	}

	for _, relative := range relativeDates {
		if date, err := parser.Parse(ctx, relative); err == nil {
			fmt.Printf("    ✓ %-15s -> %s\n", relative, date.Format("2006-01-02"))
		} else {
			fmt.Printf("    ✗ %-15s -> %v\n", relative, err)
		}
	}
}

func errorHandlingExamples() {
	parser := datetime.NewParser()
	ctx := context.Background()

	// Test various error conditions
	invalidInputs := []string{
		"",                  // Empty input
		"not-a-date",        // Invalid format
		"2023-13-45",        // Invalid date values
		"25:99:99",          // Invalid time values
		"February 30, 2023", // Invalid date
	}

	fmt.Println("  Error handling and suggestions:")
	for _, input := range invalidInputs {
		_, err := parser.Parse(ctx, input)
		if err != nil {
			if parseErr, ok := err.(*parsers.ParseError); ok {
				fmt.Printf("    Input: '%s'\n", input)
				fmt.Printf("      Type: %s\n", parseErr.Type)
				fmt.Printf("      Message: %s\n", parseErr.Message)
				if len(parseErr.Suggestions) > 0 {
					fmt.Printf("      Suggestions: %v\n", parseErr.Suggestions)
				}
				fmt.Printf("\n")
			} else {
				fmt.Printf("    %-20s -> Generic error: %v\n", input, err)
			}
		}
	}
}

func customFormatsExamples() {
	// Create parser with custom formats
	parser := datetime.NewParser(
		parsers.WithCustomFormats(
			"2006-01-02T15:04:05.000Z07:00",  // ISO with milliseconds
			"02/01/06",                       // Short year format
			"Jan _2, 2006",                   // Space-padded day
			"Monday, 02-Jan-06 15:04:05 MST", // RFC850 style
		),
	)

	ctx := context.Background()

	customInputs := []string{
		"2023-01-15T10:30:45.123Z",
		"15/01/23",
		"Jan 15, 2023",
		"Monday, 15-Jan-23 10:30:45 UTC",
	}

	fmt.Println("  Custom format parsing:")
	for _, input := range customInputs {
		if date, err := parser.Parse(ctx, input); err == nil {
			fmt.Printf("    ✓ %-35s -> %s\n", input, date.Format(time.RFC3339))
		} else {
			fmt.Printf("    ✗ %-35s -> %v\n", input, err)
		}
	}

	// Parse with specific format
	fmt.Println("\n  Parse with specific format:")
	specificInput := "15/01/2023 14:30"
	specificFormat := "02/01/2006 15:04"

	if date, err := parser.ParseWithFormat(ctx, specificInput, specificFormat); err == nil {
		fmt.Printf("    Input: %s\n", specificInput)
		fmt.Printf("    Format: %s\n", specificFormat)
		fmt.Printf("    Result: %s\n", date.Format(time.RFC3339))
	}
}

func performanceOptimizationExamples() {
	// Test parsing performance with different configurations
	inputs := []string{
		"2023-01-15T10:30:45Z",
		"01/15/2023",
		"January 15, 2023",
		"1673778645",
	}

	// Test with caching enabled (default)
	fmt.Println("  Performance comparison:")

	parser := datetime.NewParser()
	ctx := context.Background()

	// Warm up cache
	for _, input := range inputs {
		parser.Parse(ctx, input)
	}

	// Measure cached performance
	start := time.Now()
	iterations := 1000

	for i := 0; i < iterations; i++ {
		for _, input := range inputs {
			parser.Parse(ctx, input)
		}
	}

	cachedTime := time.Since(start)
	fmt.Printf("    %d iterations with caching: %v\n", iterations*len(inputs), cachedTime)
	fmt.Printf("    Average per parse: %v\n", cachedTime/time.Duration(iterations*len(inputs)))

	// Show supported formats
	fmt.Println("\n  Supported formats:")
	formats := parser.GetSupportedFormats()
	for i, format := range formats {
		if i < 10 { // Show first 10 formats
			fmt.Printf("    %s\n", format)
		} else if i == 10 {
			fmt.Printf("    ... and %d more formats\n", len(formats)-10)
			break
		}
	}

	fmt.Println("\n=== Advanced Features Summary ===")
	fmt.Println("✓ Modern parser configuration with functional options")
	fmt.Println("✓ Automatic format detection and caching")
	fmt.Println("✓ Context support for cancellation and timeouts")
	fmt.Println("✓ Enhanced Unix timestamp support (seconds, millis, micros, decimal)")
	fmt.Println("✓ Advanced text parsing with natural language support")
	fmt.Println("✓ Comprehensive error handling with suggestions")
	fmt.Println("✓ Custom format support and specific format parsing")
	fmt.Println("✓ Performance optimization with intelligent caching")
}
