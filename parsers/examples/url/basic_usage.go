package main

import (
	"context"
	"fmt"

	urlparser "github.com/fsvxavier/nexs-lib/parsers/url"
)

func main() {
	fmt.Println("=== URL Parser Example ===")

	parser := urlparser.NewParser()
	ctx := context.Background()

	testURL := "https://api.example.com/v1/users?page=1&limit=10"
	result, err := parser.ParseString(ctx, testURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("URL: %s\n", result.String())
	fmt.Printf("Scheme: %s\n", result.Scheme)
	fmt.Printf("Host: %s\n", result.Host)
	fmt.Printf("Path: %s\n", result.Path)
}
