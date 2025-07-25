// Package main demonstrates environment variable parsing functionality
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fsvxavier/nexs-lib/parsers/env"
)

func main() {
	fmt.Println("=== Environment Parser Basic Usage Examples ===")
	fmt.Println()

	// Set up some example environment variables
	setupExampleVars()

	demonstrateBasicParsing()
	demonstrateTypeConversion()
	demonstrateDefaultValues()
	demonstrateValidation()
}

func setupExampleVars() {
	// Set up example environment variables
	os.Setenv("APP_NAME", "MyApplication")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_TIMEOUT", "30.5")
	os.Setenv("APP_TAGS", "web,api,service,golang")
	os.Setenv("APP_CONFIG", "key1=value1,key2=value2,key3=value3")
	os.Setenv("APP_ENABLED", "1")
	os.Setenv("APP_WORKERS", "4")
}

func demonstrateBasicParsing() {
	fmt.Println("1. Basic Environment Variable Parsing:")

	parser := env.NewParser()
	ctx := context.Background()

	// Parse basic string values
	envVars := []string{
		"APP_NAME",
		"APP_PORT",
		"APP_DEBUG",
		"APP_TIMEOUT",
		"APP_TAGS",
		"NONEXISTENT_VAR", // This will show error handling
	}

	for _, envVar := range envVars {
		result, err := parser.ParseString(ctx, envVar)
		if err != nil {
			fmt.Printf("  %-20s -> ERROR: %v\n", envVar, err)
		} else {
			fmt.Printf("  %-20s -> %q\n", envVar, result.Value)
		}
	}

	fmt.Println()
}

func demonstrateTypeConversion() {
	fmt.Println("2. Type Conversion:")

	parser := env.NewParser()
	ctx := context.Background()

	// Parse as different types
	fmt.Println("  Converting APP_PORT to integer:")
	portResult, err := parser.ParseString(ctx, "APP_PORT")
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    String value: %q\n", portResult.Value)
		// Note: Type conversion would be handled by specific parser methods
		// This is a simplified example showing the string parsing
	}

	fmt.Println("  Converting APP_DEBUG to boolean:")
	debugResult, err := parser.ParseString(ctx, "APP_DEBUG")
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    String value: %q\n", debugResult.Value)
		fmt.Printf("    As boolean: %t\n", debugResult.Value == "true")
	}

	fmt.Println("  Converting APP_TIMEOUT to float:")
	timeoutResult, err := parser.ParseString(ctx, "APP_TIMEOUT")
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    String value: %q\n", timeoutResult.Value)
		// Would typically use ParseFloat method for actual conversion
	}

	fmt.Println("  Converting APP_TAGS to slice:")
	tagsResult, err := parser.ParseString(ctx, "APP_TAGS")
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    String value: %q\n", tagsResult.Value)
		if str, ok := tagsResult.Value.(string); ok {
			fmt.Printf("    As slice: %v\n", parseCommaSeparated(str))
		}
	}

	fmt.Println()
}

func demonstrateDefaultValues() {
	fmt.Println("3. Default Values:")

	// Demonstrate handling of missing environment variables
	defaultCases := []struct {
		envVar       string
		defaultValue string
	}{
		{"EXISTING_VAR", "default_if_missing"},
		{"MISSING_VAR", "this_is_the_default"},
		{"APP_NAME", "DefaultAppName"},
		{"UNDEFINED_PORT", "3000"},
	}

	for _, tc := range defaultCases {
		value := os.Getenv(tc.envVar)
		if value == "" {
			value = tc.defaultValue
			fmt.Printf("  %-20s -> %q (using default)\n", tc.envVar, value)
		} else {
			fmt.Printf("  %-20s -> %q (from env)\n", tc.envVar, value)
		}
	}

	fmt.Println()
}

func demonstrateValidation() {
	fmt.Println("4. Environment Variable Validation:")

	// Show validation of different formats
	validationCases := []struct {
		envVar      string
		description string
		validator   func(string) bool
	}{
		{"APP_PORT", "Valid port number", isValidPort},
		{"APP_DEBUG", "Valid boolean", isValidBool},
		{"APP_TIMEOUT", "Valid float", isValidFloat},
		{"APP_TAGS", "Valid comma-separated list", isValidCommaSeparated},
	}

	for _, vc := range validationCases {
		value := os.Getenv(vc.envVar)
		isValid := vc.validator(value)
		status := "✓ Valid"
		if !isValid {
			status = "✗ Invalid"
		}
		fmt.Printf("  %-25s: %s (%q)\n", vc.description, status, value)
	}

	fmt.Println()

	// Show all current environment variables matching our pattern
	fmt.Println("5. All APP_* Environment Variables:")
	for _, env := range os.Environ() {
		if len(env) > 4 && env[:4] == "APP_" {
			parts := parseEnvLine(env)
			if len(parts) == 2 {
				fmt.Printf("  %-20s = %q\n", parts[0], parts[1])
			}
		}
	}
}

// Helper functions for demonstration

func parseCommaSeparated(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func isValidPort(value string) bool {
	if value == "" {
		return false
	}
	// Simple port validation
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(value) > 0 && len(value) <= 5
}

func isValidBool(value string) bool {
	return value == "true" || value == "false" || value == "1" || value == "0"
}

func isValidFloat(value string) bool {
	if value == "" {
		return false
	}
	dotCount := 0
	for _, r := range value {
		if r == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isValidCommaSeparated(value string) bool {
	return value != "" && len(parseCommaSeparated(value)) > 0
}

func parseEnvLine(env string) []string {
	for i, r := range env {
		if r == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
