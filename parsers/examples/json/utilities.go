// Package main demonstrates JSON utility functions
package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/parsers/json"
)

func main() {
	runUtilitiesExamples()
}

func runUtilitiesExamples() {
	fmt.Println("=== JSON Utilities Examples ===\n")

	demonstrateMergeJSON()
	demonstrateExtractPath()
	demonstrateJSONValidation()
	demonstrateDataTransformation()
}

func demonstrateMergeJSON() {
	fmt.Println("1. JSON Merge Operations:")

	// Merge multiple JSON objects
	user := map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}

	profile := `{
		"age": 31,
		"location": "New York",
		"skills": ["Go", "Python", "JavaScript"]
	}`

	settings := map[string]interface{}{
		"theme":         "dark",
		"notifications": true,
		"language":      "en",
	}

	merged, err := json.MergeJSON(user, profile, settings)
	if err != nil {
		log.Printf("Error merging JSON: %v", err)
		return
	}

	fmt.Printf("Original user: %+v\n", user)
	fmt.Printf("Profile JSON: %s\n", profile)
	fmt.Printf("Settings: %+v\n", settings)
	fmt.Printf("Merged result: %+v\n\n", merged)

	// Demonstrate overriding values
	base := map[string]interface{}{
		"name":   "Default",
		"status": "inactive",
		"config": map[string]interface{}{
			"debug": false,
			"level": 1,
		},
	}

	override := `{
		"status": "active",
		"config": {
			"debug": true,
			"level": 2,
			"new_feature": true
		}
	}`

	overridden, err := json.MergeJSON(base, override)
	if err != nil {
		log.Printf("Error overriding JSON: %v", err)
		return
	}

	fmt.Printf("Base configuration: %+v\n", base)
	fmt.Printf("Override configuration: %s\n", override)
	fmt.Printf("Final configuration: %+v\n\n", overridden)
}

func demonstrateExtractPath() {
	fmt.Println("2. JSON Path Extraction:")

	// Complex nested data structure
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   12345,
			"name": "Alice Johnson",
			"profile": map[string]interface{}{
				"email":    "alice@example.com",
				"location": "New York",
				"preferences": map[string]interface{}{
					"theme":    "dark",
					"language": "en",
					"notifications": map[string]interface{}{
						"email": true,
						"push":  false,
						"sms":   true,
					},
				},
			},
			"skills": []interface{}{"Go", "Python", "JavaScript", "Docker"},
		},
		"metadata": map[string]interface{}{
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-15T12:30:00Z",
			"version":    "1.0.0",
		},
	}

	// Extract various paths
	paths := []string{
		"user.name",
		"user.profile.email",
		"user.profile.preferences.theme",
		"user.profile.preferences.notifications.email",
		"metadata.version",
		"user.skills",
	}

	fmt.Println("Extracting paths from complex data:")
	for _, path := range paths {
		value, err := json.ExtractPath(data, path)
		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", path, err)
		} else {
			fmt.Printf("  %s: %+v\n", path, value)
		}
	}

	// Try invalid paths
	fmt.Println("\nTesting invalid paths:")
	invalidPaths := []string{
		"user.profile.nonexistent",
		"user.name.invalid",
		"nonexistent.path",
	}

	for _, path := range invalidPaths {
		value, err := json.ExtractPath(data, path)
		if err != nil {
			fmt.Printf("  %s: Expected error - %v\n", path, err)
		} else {
			fmt.Printf("  %s: Unexpected success - %+v\n", path, value)
		}
	}

	fmt.Println()
}

func demonstrateJSONValidation() {
	fmt.Println("3. JSON Validation:")

	// Test various JSON strings
	testCases := []struct {
		name string
		json string
	}{
		{"Valid object", `{"name": "Alice", "age": 30}`},
		{"Valid array", `[1, 2, 3, "test"]`},
		{"Valid nested", `{"user": {"name": "Alice", "skills": ["Go", "Python"]}}`},
		{"Invalid - missing quote", `{"name": Alice, "age": 30}`},
		{"Invalid - trailing comma", `{"name": "Alice", "age": 30,}`},
		{"Invalid - missing brace", `{"name": "Alice", "age": 30`},
		{"Empty object", `{}`},
		{"Empty array", `[]`},
		{"Null value", `null`},
		{"Boolean", `true`},
		{"Number", `42`},
		{"String", `"hello"`},
	}

	fmt.Println("Validating JSON strings:")
	for _, tc := range testCases {
		err := json.ValidateJSONString(tc.json)
		status := "✓ Valid"
		if err != nil {
			status = fmt.Sprintf("✗ Invalid: %v", err)
		}
		fmt.Printf("  %-20s: %s\n", tc.name, status)
	}

	// Test data validation
	fmt.Println("\nValidating Go data structures:")
	dataTestCases := []struct {
		name string
		data interface{}
	}{
		{"Valid map", map[string]interface{}{"name": "Alice"}},
		{"Valid slice", []interface{}{1, 2, 3}},
		{"Valid struct", struct{ Name string }{Name: "Alice"}},
		{"Invalid channel", make(chan int)},
		{"Invalid function", func() {}},
		{"Valid pointer", &struct{ Name string }{Name: "Alice"}},
	}

	for _, tc := range dataTestCases {
		err := json.ValidateJSONData(tc.data)
		status := "✓ Valid"
		if err != nil {
			status = fmt.Sprintf("✗ Invalid: %v", err)
		}
		fmt.Printf("  %-20s: %s\n", tc.name, status)
	}

	fmt.Println()
}

func demonstrateDataTransformation() {
	fmt.Println("4. Data Transformation:")

	// Convert between different representations
	originalData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "active": true},
		{"id": 2, "name": "Bob", "active": false},
		{"id": 3, "name": "Charlie", "active": true},
	}

	fmt.Printf("Original data: %+v\n", originalData)

	// Convert to JSONL
	jsonlString, err := json.ConvertToJSONL([]interface{}{
		originalData[0], originalData[1], originalData[2],
	})
	if err != nil {
		log.Printf("Error converting to JSONL: %v", err)
		return
	}

	fmt.Printf("Converted to JSONL:\n%s\n", jsonlString)

	// Parse JSONL back
	parsedObjects, err := json.ParseJSONL(jsonlString)
	if err != nil {
		log.Printf("Error parsing JSONL: %v", err)
		return
	}

	fmt.Printf("Parsed back from JSONL: %+v\n", parsedObjects)

	// Format for display
	formatter := json.NewFormatterWithIndent("  ")
	prettyJSON, err := formatter.FormatString(nil, originalData)
	if err != nil {
		log.Printf("Error formatting JSON: %v", err)
		return
	}

	fmt.Printf("Pretty formatted:\n%s\n", prettyJSON)

	// Compact the pretty JSON
	compactJSON, err := json.CompactJSON(prettyJSON)
	if err != nil {
		log.Printf("Error compacting JSON: %v", err)
		return
	}

	fmt.Printf("Compacted: %s\n", compactJSON)

	// Working with type-safe conversion
	fmt.Println("\n5. Type-Safe Transformations:")

	type User struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}

	// Convert map to struct
	userMap := map[string]interface{}{
		"id":     100,
		"name":   "Diana",
		"active": true,
	}

	user, err := json.ParseJSONToType[User](userMap)
	if err != nil {
		log.Printf("Error converting to User: %v", err)
		return
	}

	fmt.Printf("Converted map to struct: %+v\n", user)

	// Convert struct back to map
	userAsMap, err := json.ParseJSONToType[map[string]interface{}](user)
	if err != nil {
		log.Printf("Error converting to map: %v", err)
		return
	}

	fmt.Printf("Converted struct to map: %+v\n", userAsMap)
}
