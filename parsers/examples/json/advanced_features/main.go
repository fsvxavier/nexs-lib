// Package main demonstrates advanced JSON parsing features
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/parsers/json"
)

func main() {
	runAdvancedExamples()
}

func runAdvancedExamples() {
	fmt.Println("=== JSON Parser Advanced Features Examples ===\n")

	demonstrateCommentsSupport()
	demonstrateTrailingCommas()
	demonstrateCombinedFeatures()
	demonstrateCustomParser()
}

func demonstrateCommentsSupport() {
	fmt.Println("1. JSON with Comments Support:")

	jsonWithComments := `{
		// User information
		"name": "Alice",
		"age": 30,
		/* Contact details
		   Multi-line comment */
		"email": "alice@example.com",
		// Skills array
		"skills": ["Go", "Python", "JavaScript"]
	}`

	// Create advanced parser with comments enabled
	parser := json.NewAdvancedParser().WithComments(true)
	ctx := context.Background()

	result, err := parser.ParseAdvanced(ctx, jsonWithComments)
	if err != nil {
		log.Printf("Error parsing JSON with comments: %v", err)
		return
	}

	fmt.Printf("Parsed JSON with comments: %+v\n\n", result)
}

func demonstrateTrailingCommas() {
	fmt.Println("2. JSON with Trailing Commas Support:")

	jsonWithTrailingCommas := `{
		"name": "Bob",
		"age": 25,
		"skills": [
			"Go",
			"Docker",
			"Kubernetes",
		],
		"active": true,
	}`

	// Create advanced parser with trailing commas enabled
	parser := json.NewAdvancedParser().WithTrailingCommas(true)
	ctx := context.Background()

	result, err := parser.ParseAdvanced(ctx, jsonWithTrailingCommas)
	if err != nil {
		log.Printf("Error parsing JSON with trailing commas: %v", err)
		return
	}

	fmt.Printf("Parsed JSON with trailing commas: %+v\n\n", result)
}

func demonstrateCombinedFeatures() {
	fmt.Println("3. JSON with Comments AND Trailing Commas:")

	jsonWithBoth := `{
		// User identification
		"id": 123,
		"name": "Charlie",
		/* Personal information
		   Including age and location */
		"age": 35,
		"location": "São Paulo",
		// Professional skills
		"skills": [
			"Go",        // Primary language
			"Python",    // Secondary language
			"Docker",    // DevOps tool
		],
		// Account status
		"active": true,
	}`

	// Create advanced parser with both features enabled
	parser := json.NewAdvancedParser().
		WithComments(true).
		WithTrailingCommas(true)

	ctx := context.Background()

	result, err := parser.ParseAdvanced(ctx, jsonWithBoth)
	if err != nil {
		log.Printf("Error parsing JSON with combined features: %v", err)
		return
	}

	fmt.Printf("Parsed JSON with comments and trailing commas: %+v\n\n", result)
}

func demonstrateCustomParser() {
	fmt.Println("4. Custom Parser Configuration:")

	// Create parser with all advanced features
	parser := json.NewAdvancedParser().
		WithComments(true).
		WithTrailingCommas(true).
		WithStrictNumbers(true)

	fmt.Printf("Parser configuration:\n")
	fmt.Printf("- Advanced parser created with all features enabled\n")
	fmt.Printf("- Comments processing: enabled\n")
	fmt.Printf("- Trailing commas processing: enabled\n")
	fmt.Printf("- Strict numbers processing: enabled\n")

	// Test parsing with different input types
	ctx := context.Background()

	// String input
	jsonString := `{
		// Configuration
		"debug": true,
		"timeout": 30,
	}`

	stringResult, err := parser.ParseAdvanced(ctx, jsonString)
	if err != nil {
		log.Printf("Error parsing string: %v", err)
		return
	}
	fmt.Printf("String result: %+v\n", stringResult)

	// Bytes input
	jsonBytes := []byte(`{
		// Settings
		"retries": 3,
		"enabled": true,
	}`)

	bytesResult, err := parser.ParseAdvanced(ctx, jsonBytes)
	if err != nil {
		log.Printf("Error parsing bytes: %v", err)
		return
	}
	fmt.Printf("Bytes result: %+v\n", bytesResult)

	// Demonstrate comment removal
	fmt.Println("\n5. Comment Processing Examples:")

	testCases := []string{
		`{"name": "test"} // line comment`,
		`{"name": /* inline */ "test"}`,
		`{"comment": "this // is not a comment"}`,
		`{
			"name": "test",
			/* multi-line
			   comment */
			"value": 42
		}`,
	}

	for i, testCase := range testCases {
		fmt.Printf("Test case %d:\n", i+1)
		fmt.Printf("Input: %s\n", testCase)

		result, err := parser.ParseAdvanced(ctx, testCase)
		if err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Output: %+v\n", result)
		}
		fmt.Println()
	}
}
