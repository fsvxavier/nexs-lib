// Package main demonstrates special JSON format parsing
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/fsvxavier/nexs-lib/parsers/json"
)

func main() {
	runSpecialFormatsExamples()
}

func runSpecialFormatsExamples() {
	fmt.Println("=== JSON Special Formats Examples ===\n")

	demonstrateJSONL()
	demonstrateNDJSON()
	demonstrateJSON5()
	demonstrateStreamingParser()
}

func demonstrateJSONL() {
	fmt.Println("1. JSON Lines (JSONL) Parsing:")

	// JSON Lines format - one JSON object per line
	jsonlData := `{"name": "Alice", "age": 30, "city": "New York"}
{"name": "Bob", "age": 25, "city": "San Francisco"}
{"name": "Charlie", "age": 35, "city": "Chicago"}
{"name": "Diana", "age": 28, "city": "Boston"}`

	results, err := json.ParseJSONL(jsonlData)
	if err != nil {
		log.Printf("Error parsing JSONL: %v", err)
		return
	}

	fmt.Printf("Parsed %d objects from JSONL:\n", len(results))
	for i, result := range results {
		fmt.Printf("  Object %d: %+v\n", i+1, result)
	}

	// Validate JSONL format
	isValid := json.IsValidJSONL(jsonlData)
	fmt.Printf("JSONL validation result: %v\n\n", isValid)
}

func demonstrateNDJSON() {
	fmt.Println("2. Newline Delimited JSON (NDJSON) Parsing:")

	// NDJSON is same as JSONL
	ndjsonData := `{"event": "user_login", "user_id": 123, "timestamp": "2025-01-01T10:00:00Z"}
{"event": "page_view", "user_id": 123, "page": "/dashboard", "timestamp": "2025-01-01T10:01:00Z"}
{"event": "user_logout", "user_id": 123, "timestamp": "2025-01-01T10:30:00Z"}`

	results, err := json.ParseNDJSON(ndjsonData)
	if err != nil {
		log.Printf("Error parsing NDJSON: %v", err)
		return
	}

	fmt.Printf("Parsed %d events from NDJSON:\n", len(results))
	for i, result := range results {
		fmt.Printf("  Event %d: %+v\n", i+1, result)
	}

	// Convert objects back to JSONL format
	originalObjects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Product A", "price": 29.99},
		map[string]interface{}{"id": 2, "name": "Product B", "price": 39.99},
		map[string]interface{}{"id": 3, "name": "Product C", "price": 19.99},
	}

	jsonlString, err := json.ConvertToJSONL(originalObjects)
	if err != nil {
		log.Printf("Error converting to JSONL: %v", err)
		return
	}

	fmt.Printf("Converted objects to JSONL:\n%s\n\n", jsonlString)
}

func demonstrateJSON5() {
	fmt.Println("3. JSON5 Parsing (Basic Support):")

	// JSON5 with comments and trailing commas
	json5Data := `{
		// User configuration
		"name": "Alice",
		"preferences": {
			// Display settings
			"theme": "dark",
			"language": "pt-BR",
			// Feature flags
			"notifications": true,
		},
		// Contact information
		"contacts": [
			"alice@example.com",
			"alice.backup@example.com",
		],
	}`

	result, err := json.ParseJSON5(json5Data)
	if err != nil {
		log.Printf("Error parsing JSON5: %v", err)
		return
	}

	fmt.Printf("Parsed JSON5: %+v\n\n", result)

	// More complex JSON5 example
	complexJSON5 := `{
		// Application configuration
		"app": {
			"name": "MyApp",
			"version": "1.0.0",
			/* Database settings
			   Support multiple environments */
			"database": {
				"host": "localhost",
				"port": 5432,
				"ssl": true,
			},
		},
		// Features
		"features": [
			"authentication",
			"logging",
			"monitoring",
		],
	}`

	complexResult, err := json.ParseJSON5(complexJSON5)
	if err != nil {
		log.Printf("Error parsing complex JSON5: %v", err)
		return
	}

	fmt.Printf("Parsed complex JSON5: %+v\n\n", complexResult)
}

func demonstrateStreamingParser() {
	fmt.Println("4. Streaming JSON Parser:")

	// Simulate a stream of JSON objects
	jsonStream := `{"id": 1, "name": "Alice", "status": "active"}
{"id": 2, "name": "Bob", "status": "inactive"}
{"id": 3, "name": "Charlie", "status": "active"}
{"id": 4, "name": "Diana", "status": "pending"}`

	reader := strings.NewReader(jsonStream)
	parser := json.NewStreamParser(reader)

	fmt.Println("Processing stream of JSON objects:")

	objectCount := 0
	for parser.HasMore() {
		var obj map[string]interface{}
		err := parser.ParseNext(&obj)
		if err != nil {
			log.Printf("Error parsing next object: %v", err)
			break
		}

		objectCount++
		fmt.Printf("  Stream object %d: %+v\n", objectCount, obj)
	}

	fmt.Printf("Processed %d objects from stream\n\n", objectCount)

	// Demonstrate with structured data
	fmt.Println("5. Streaming with Structured Data:")

	type User struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	structuredStream := `{"id": 100, "name": "Admin", "status": "active"}
{"id": 101, "name": "Moderator", "status": "active"}
{"id": 102, "name": "User", "status": "inactive"}`

	structReader := strings.NewReader(structuredStream)
	structParser := json.NewStreamParser(structReader)

	fmt.Println("Processing structured stream:")

	userCount := 0
	for structParser.HasMore() {
		var user User
		err := structParser.ParseNext(&user)
		if err != nil {
			log.Printf("Error parsing user: %v", err)
			break
		}

		userCount++
		fmt.Printf("  User %d: ID=%d, Name=%s, Status=%s\n",
			userCount, user.ID, user.Name, user.Status)
	}

	fmt.Printf("Processed %d users from structured stream\n", userCount)
}
