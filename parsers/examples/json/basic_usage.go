// Package main demonstrates basic JSON parsing functionality
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/parsers/json"
)

// User represents a user structure for demonstrations
type User struct {
	Name   string   `json:"name"`
	Age    int      `json:"age"`
	Email  string   `json:"email"`
	Skills []string `json:"skills"`
	Active bool     `json:"active"`
}

func main() {
	fmt.Println("=== JSON Parser Basic Usage Examples ===\n")

	demonstrateBasicParsing()
	demonstrateTypeSafeParsing()
	demonstrateFormatting()
	demonstrateValidation()
}

func demonstrateBasicParsing() {
	fmt.Println("1. Basic JSON Parsing:")

	// Parse JSON string to interface{}
	jsonString := `{
		"name": "Alice",
		"age": 30,
		"email": "alice@example.com",
		"skills": ["Go", "Python", "JavaScript"],
		"active": true
	}`

	result, err := json.ParseJSONString(jsonString)
	if err != nil {
		log.Printf("Error parsing JSON string: %v", err)
		return
	}

	fmt.Printf("Parsed JSON: %+v\n", result)

	// Parse JSON bytes
	jsonBytes := []byte(`[1, 2, 3, 4, 5]`)
	arrayResult, err := json.ParseJSONBytes(jsonBytes)
	if err != nil {
		log.Printf("Error parsing JSON bytes: %v", err)
		return
	}

	fmt.Printf("Parsed JSON array: %+v\n", arrayResult)

	// Parse with context
	parser := json.NewParser()
	ctx := context.Background()

	contextResult, err := parser.ParseString(ctx, jsonString)
	if err != nil {
		log.Printf("Error parsing with context: %v", err)
		return
	}

	fmt.Printf("Parsed with context: %+v\n\n", contextResult)
}

func demonstrateTypeSafeParsing() {
	fmt.Println("2. Type-Safe JSON Parsing:")

	// Data as map (simulating JSON input)
	userData := map[string]interface{}{
		"name":   "Bob",
		"age":    25,
		"email":  "bob@example.com",
		"skills": []interface{}{"Go", "Docker", "Kubernetes"},
		"active": true,
	}

	// Parse to specific type
	user, err := json.ParseJSONToType[User](userData)
	if err != nil {
		log.Printf("Error parsing to User type: %v", err)
		return
	}

	fmt.Printf("Parsed User: %+v\n", user)

	// Parse JSON string to specific type
	jsonString := `{
		"name": "Charlie",
		"age": 35,
		"email": "charlie@example.com",
		"skills": ["Python", "Machine Learning"],
		"active": false
	}`

	// Convert string to map first, then to type
	tempResult, err := json.ParseJSONString(jsonString)
	if err != nil {
		log.Printf("Error parsing JSON string: %v", err)
		return
	}

	user2, err := json.ParseJSONToType[User](tempResult)
	if err != nil {
		log.Printf("Error converting to User type: %v", err)
		return
	}

	fmt.Printf("Parsed User from JSON string: %+v\n\n", user2)
}

func demonstrateFormatting() {
	fmt.Println("3. JSON Formatting:")

	data := map[string]interface{}{
		"name": "Alice",
		"details": map[string]interface{}{
			"age":    30,
			"skills": []string{"Go", "Python"},
		},
	}

	// Basic formatting
	formatter := json.NewFormatter()
	ctx := context.Background()

	formatted, err := formatter.FormatString(ctx, data)
	if err != nil {
		log.Printf("Error formatting JSON: %v", err)
		return
	}

	fmt.Printf("Compact JSON: %s\n", formatted)

	// Pretty formatting with indentation
	prettyFormatter := json.NewFormatterWithIndent("  ")
	prettyFormatted, err := prettyFormatter.FormatString(ctx, data)
	if err != nil {
		log.Printf("Error pretty formatting JSON: %v", err)
		return
	}

	fmt.Printf("Pretty JSON:\n%s\n", prettyFormatted)

	// Using utility functions
	compactJSON := `{
		"name": "Alice",
		"age": 30,
		"skills": [
			"Go",
			"Python"
		]
	}`

	compact, err := json.CompactJSON(compactJSON)
	if err != nil {
		log.Printf("Error compacting JSON: %v", err)
		return
	}

	fmt.Printf("Compacted: %s\n", compact)

	pretty, err := json.PrettyJSON(compact, "  ")
	if err != nil {
		log.Printf("Error prettifying JSON: %v", err)
		return
	}

	fmt.Printf("Prettified:\n%s\n\n", pretty)
}

func demonstrateValidation() {
	fmt.Println("4. JSON Validation:")

	validJSON := `{"name": "Alice", "age": 30}`
	invalidJSON := `{"name": "Alice", "age": 30`

	// Validate JSON string
	err := json.ValidateJSONString(validJSON)
	if err != nil {
		fmt.Printf("Valid JSON failed validation: %v\n", err)
	} else {
		fmt.Println("✓ Valid JSON passed validation")
	}

	err = json.ValidateJSONString(invalidJSON)
	if err != nil {
		fmt.Printf("✓ Invalid JSON correctly failed validation: %v\n", err)
	} else {
		fmt.Println("✗ Invalid JSON incorrectly passed validation")
	}

	// Validate data
	validData := map[string]interface{}{"name": "Alice"}
	invalidData := make(chan int) // channels can't be marshaled to JSON

	err = json.ValidateJSONData(validData)
	if err != nil {
		fmt.Printf("Valid data failed validation: %v\n", err)
	} else {
		fmt.Println("✓ Valid data passed validation")
	}

	err = json.ValidateJSONData(invalidData)
	if err != nil {
		fmt.Printf("✓ Invalid data correctly failed validation: %v\n", err)
	} else {
		fmt.Println("✗ Invalid data incorrectly passed validation")
	}

	fmt.Println()
}
