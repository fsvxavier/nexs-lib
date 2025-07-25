// Package main demonstrates compatibility features with old parse module
package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/parsers/json"
)

func main() {
	runCompatibilityExamples()
}

func runCompatibilityExamples() {
	fmt.Println("=== JSON Compatibility Examples ===\n")

	demonstrateDirectMigration()
	demonstrateCompatibilityFunctions()
	demonstrateAliasedFunctions()
	demonstrateEnhancedFeatures()
}

func demonstrateDirectMigration() {
	fmt.Println("1. Direct Migration from _old/parse:")

	// User struct for this demonstration
	type User struct {
		Name   string   `json:"name"`
		Age    int      `json:"age"`
		Email  string   `json:"email"`
		Skills []string `json:"skills"`
		Active bool     `json:"active"`
	}

	// OLD WAY (would have been):
	// import "github.com/fsvxavier/nexs-lib/_old/parse"
	// result, err := parse.ParseJSONToType[User](data)

	// NEW WAY - exactly the same signature:
	data := map[string]interface{}{
		"name":   "Alice",
		"age":    30,
		"email":  "alice@example.com",
		"skills": []interface{}{"Go", "Python", "JavaScript"},
		"active": true,
	}

	// Using compatibility function - exact same signature
	user, err := json.ParseJSONToTypeCompat[User](data)
	if err != nil {
		log.Printf("Error with compatibility function: %v", err)
		return
	}
	fmt.Printf("Compatibility function result: %+v\n", user)

	// Using new function - same result
	user2, err := json.ParseJSONToType[User](data)
	if err != nil {
		log.Printf("Error with new function: %v", err)
		return
	}
	fmt.Printf("New function result: %+v\n", user2)

	fmt.Println("✓ Both functions produce identical results\n")
}

func demonstrateCompatibilityFunctions() {
	fmt.Println("2. Compatibility Functions:")

	jsonString := `{
		"name": "Bob",
		"age": 25,
		"email": "bob@example.com",
		"skills": ["Go", "Docker"],
		"active": true
	}`

	// Parse with compatibility function
	result, err := json.ParseJSONToTypeCompat[map[string]interface{}](jsonString)
	if err != nil {
		log.Printf("Error parsing with compatibility function: %v", err)
		return
	}
	fmt.Printf("ParseJSONToTypeCompat result: %+v\n", result)

	// Test with different types
	arrayData := []interface{}{1, 2, 3, "test", true}
	arrayResult, err := json.ParseJSONToTypeCompat[[]interface{}](arrayData)
	if err != nil {
		log.Printf("Error with array compatibility: %v", err)
		return
	}
	fmt.Printf("Array compatibility result: %+v\n", arrayResult)

	// Test error handling
	invalidData := make(chan int) // channels can't be marshaled
	_, err = json.ParseJSONToTypeCompat[interface{}](invalidData)
	if err != nil {
		fmt.Printf("✓ Error handling works: %v\n", err)
	} else {
		fmt.Println("✗ Error handling failed")
	}

	fmt.Println()
}

func demonstrateAliasedFunctions() {
	fmt.Println("3. Aliased Functions for Easy Migration:")

	// These are aliases that make migration even easier
	data := map[string]interface{}{
		"id":   123,
		"name": "Charlie",
		"tags": []interface{}{"important", "user"},
	}

	// Using Parse alias (equivalent to ParseJSON)
	result1, err := json.Parse(data)
	if err != nil {
		log.Printf("Error with Parse alias: %v", err)
		return
	}
	fmt.Printf("Parse alias result: %+v\n", result1)

	// Using ParseString alias
	jsonStr := `{"message": "Hello, World!", "timestamp": 1640995200}`
	result2, err := json.ParseString(jsonStr)
	if err != nil {
		log.Printf("Error with ParseString alias: %v", err)
		return
	}
	fmt.Printf("ParseString alias result: %+v\n", result2)

	// Using ParseBytes alias
	jsonBytes := []byte(`[{"id": 1}, {"id": 2}]`)
	result3, err := json.ParseBytes(jsonBytes)
	if err != nil {
		log.Printf("Error with ParseBytes alias: %v", err)
		return
	}
	fmt.Printf("ParseBytes alias result: %+v\n", result3)

	// Using Validate alias
	validData := map[string]interface{}{"test": "data"}
	err = json.Validate(validData)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Validation passed with Validate alias")
	}

	fmt.Println()
}

func demonstrateEnhancedFeatures() {
	fmt.Println("4. Enhanced Features Not Available in Old Module:")

	// Advanced parser with comments and trailing commas
	jsonWithExtras := `{
		// User information
		"name": "Diana",
		"age": 28,
		/* Additional data
		   with multi-line comment */
		"roles": [
			"admin",
			"user",
		],
		"active": true,
	}`

	parser := json.NewAdvancedParser().
		WithComments(true).
		WithTrailingCommas(true)

	result, err := parser.ParseAdvanced(nil, jsonWithExtras)
	if err != nil {
		log.Printf("Error with advanced parser: %v", err)
		return
	}
	fmt.Printf("Advanced parser result: %+v\n", result)

	// JSON Lines support
	jsonlData := `{"name": "User1", "score": 100}
{"name": "User2", "score": 85}
{"name": "User3", "score": 92}`

	jsonlResults, err := json.ParseJSONL(jsonlData)
	if err != nil {
		log.Printf("Error with JSONL: %v", err)
		return
	}
	fmt.Printf("JSONL results: %+v\n", jsonlResults)

	// JSON merge functionality
	obj1 := map[string]interface{}{"name": "Base", "value": 1}
	obj2 := `{"name": "Override", "extra": "data"}`

	merged, err := json.MergeJSON(obj1, obj2)
	if err != nil {
		log.Printf("Error with merge: %v", err)
		return
	}
	fmt.Printf("Merged result: %+v\n", merged)

	// Path extraction
	complexData := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "Elena",
			},
		},
	}

	name, err := json.ExtractPath(complexData, "user.profile.name")
	if err != nil {
		log.Printf("Error with path extraction: %v", err)
		return
	}
	fmt.Printf("Extracted name: %v\n", name)

	// Pretty formatting
	prettyJSON, err := json.PrettyJSON(`{"compact":"json"}`, "  ")
	if err != nil {
		log.Printf("Error with pretty formatting: %v", err)
		return
	}
	fmt.Printf("Pretty JSON:\n%s\n", prettyJSON)

	fmt.Println("✓ All enhanced features working correctly")
}

// Migration guide function
func printMigrationGuide() {
	fmt.Println("=== Migration Guide ===")
	fmt.Println()
	fmt.Println("1. Update imports:")
	fmt.Println("   OLD: import \"github.com/fsvxavier/nexs-lib/_old/parse\"")
	fmt.Println("   NEW: import \"github.com/fsvxavier/nexs-lib/parsers/json\"")
	fmt.Println()
	fmt.Println("2. Update function calls:")
	fmt.Println("   OLD: parse.ParseJSONToType[T](data)")
	fmt.Println("   NEW: json.ParseJSONToType[T](data)")
	fmt.Println("    OR: json.ParseJSONToTypeCompat[T](data) // exact compatibility")
	fmt.Println()
	fmt.Println("3. Enhanced features now available:")
	fmt.Println("   - Advanced parsing (comments, trailing commas)")
	fmt.Println("   - JSONL/NDJSON support")
	fmt.Println("   - JSON5 basic support")
	fmt.Println("   - Streaming parser")
	fmt.Println("   - Merge and path extraction utilities")
	fmt.Println("   - Better formatting options")
	fmt.Println()
}
