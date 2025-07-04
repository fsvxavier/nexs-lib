package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/parsers/datetime"
	"github.com/fsvxavier/nexs-lib/parsers/duration"
	"github.com/fsvxavier/nexs-lib/parsers/environment"
)

func main() {
	fmt.Println("Testing Parsers Library")

	// Test DateTime
	ctx := context.Background()
	dtParser := datetime.NewParser()

	if date, err := dtParser.Parse(ctx, "2023-01-15T10:30:45Z"); err == nil {
		fmt.Printf("DateTime parsed: %s\n", date.Format("2006-01-02 15:04:05"))
	} else {
		log.Printf("DateTime error: %v", err)
	}

	// Test Duration
	if d, err := duration.Parse("1h30m"); err == nil {
		fmt.Printf("Duration parsed: %v\n", d)
	} else {
		log.Printf("Duration error: %v", err)
	}

	// Test Environment
	env := environment.NewParser()
	port := env.GetInt("PORT", 8080)
	fmt.Printf("Port: %d\n", port)
}
