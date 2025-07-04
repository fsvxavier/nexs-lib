package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
	"github.com/fsvxavier/nexs-lib/parsers/datetime"
	"github.com/fsvxavier/nexs-lib/parsers/duration"
	"github.com/fsvxavier/nexs-lib/parsers/environment"
)

func main() {
	fmt.Println("=== Parsers Library Examples ===")
	fmt.Println()

	datetimeExamples()
	durationExamples()
	environmentExamples()
}

func datetimeExamples() {
	fmt.Println("📅 DateTime Parser Examples:")
	fmt.Println("----------------------------")

	ctx := context.Background()
	parser := datetime.NewParser()

	// Various datetime formats
	formats := []string{
		"2023-01-15T10:30:45Z",
		"2023-01-15 10:30:45",
		"15/01/2023",
		"01/15/2023",
		"January 15, 2023",
		"Jan 15, 2023",
		"15 Jan 2023",
		"today",
		"yesterday",
		"tomorrow",
	}

	for _, format := range formats {
		if date, err := parser.Parse(ctx, format); err == nil {
			fmt.Printf("  ✅ %-25s -> %s\n", format, date.Format("2006-01-02 15:04:05 MST"))
		} else {
			fmt.Printf("  ❌ %-25s -> %v\n", format, err)
		}
	}

	// Parse with specific timezone
	fmt.Println("\n🌍 Timezone Examples:")
	loc, _ := time.LoadLocation("America/New_York")
	if date, err := parser.ParseInLocation(ctx, "2023-01-15 15:30:45", loc); err == nil {
		fmt.Printf("  ✅ New York time: %s\n", date.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("  ✅ UTC time:      %s\n", date.UTC().Format("2006-01-02 15:04:05 MST"))
	}

	fmt.Println()
}

func durationExamples() {
	fmt.Println("⏱️  Duration Parser Examples:")
	fmt.Println("-----------------------------")

	// Various duration formats
	formats := []string{
		"1h30m45s",           // Standard Go format
		"2d3h",               // Extended with days
		"1w2d",               // Weeks and days
		"1w2d3h4m5s",         // Complex combination
		"2.5h",               // Decimal hours
		"90m",                // Minutes that could be hours
		"2 hours 30 minutes", // Verbose format
		"half an hour",       // Natural language
		"quarter hour",       // Natural language
		"a day",              // Simple natural language
		"30 seconds",         // Simple format
	}

	for _, format := range formats {
		if d, err := duration.Parse(format); err == nil {
			fmt.Printf("  ✅ %-20s -> %-15s (%v)\n", format, d.String(), d)
		} else {
			fmt.Printf("  ❌ %-20s -> %v\n", format, err)
		}
	}

	// Custom units example
	fmt.Println("\n🔧 Custom Units Example:")
	customUnits := map[string]time.Duration{
		"fortnight": 14 * 24 * time.Hour,
		"jiffy":     10 * time.Millisecond,
	}

	parser := duration.NewParser(parsers.WithCustomUnits(customUnits))
	if d, err := parser.Parse(context.Background(), "1fortnight2jiffy"); err == nil {
		fmt.Printf("  ✅ 1fortnight2jiffy -> %v\n", d)
	}

	fmt.Println()
}

func environmentExamples() {
	fmt.Println("🌍 Environment Parser Examples:")
	fmt.Println("--------------------------------")

	// Set some example environment variables
	setupExampleEnv()

	// Create parser with configuration
	env := environment.NewParser(
		environment.WithPrefix("EXAMPLE"),
		environment.WithDefaults(map[string]string{
			"PORT":    "8080",
			"HOST":    "localhost",
			"DEBUG":   "false",
			"TIMEOUT": "30s",
		}),
	)

	// Basic type parsing
	fmt.Println("📋 Basic Types:")
	fmt.Printf("  Port:    %d\n", env.GetInt("PORT"))
	fmt.Printf("  Host:    %s\n", env.GetString("HOST"))
	fmt.Printf("  Debug:   %t\n", env.GetBool("DEBUG"))
	fmt.Printf("  Timeout: %v\n", env.GetDuration("TIMEOUT"))

	// Advanced types
	fmt.Println("\n📊 Advanced Types:")
	fmt.Printf("  Tags:    %v\n", env.GetSlice("TAGS", ","))
	fmt.Printf("  Config:  %v\n", env.GetMap("CONFIG", ",", "="))

	// Optional values (pointers)
	fmt.Println("\n🔍 Optional Values:")
	if maxConn := env.GetIntPtr("MAX_CONNECTIONS"); maxConn != nil {
		fmt.Printf("  Max Connections: %d\n", *maxConn)
	} else {
		fmt.Printf("  Max Connections: not set\n")
	}

	// Hierarchical configuration
	fmt.Println("\n🏗️  Hierarchical Config:")
	dbEnv := env.WithPrefix("DB")
	fmt.Printf("  DB Host: %s\n", dbEnv.GetString("HOST", "localhost"))
	fmt.Printf("  DB Port: %d\n", dbEnv.GetInt("PORT", 5432))
	fmt.Printf("  DB Name: %s\n", dbEnv.GetString("NAME", "myapp"))

	// Validation example
	fmt.Println("\n✅ Validation:")
	validatingEnv := environment.NewParser(
		environment.WithPrefix("EXAMPLE"),
		environment.WithRequired("API_KEY"), // This will fail
	)

	if err := validatingEnv.Validate(); err != nil {
		fmt.Printf("  ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ All required variables present\n")
	}

	fmt.Println()
}

func setupExampleEnv() {
	// Set up some example environment variables
	envVars := map[string]string{
		"EXAMPLE_PORT":    "9090",
		"EXAMPLE_DEBUG":   "true",
		"EXAMPLE_TIMEOUT": "45s",
		"EXAMPLE_TAGS":    "api,web,golang",
		"EXAMPLE_CONFIG":  "env=prod,region=us-east,cache=redis",
		"EXAMPLE_DB_HOST": "db.example.com",
		"EXAMPLE_DB_PORT": "5432",
		"EXAMPLE_DB_NAME": "production",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			log.Printf("Failed to set %s: %v", key, err)
		}
	}
}
