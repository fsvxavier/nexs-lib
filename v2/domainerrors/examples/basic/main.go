package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🔧 Domain Errors v2 - Basic Usage Examples")
	fmt.Println("==========================================")

	basicErrorCreation()
	builderPatternExample()
	validationErrorExample()
	jsonSerializationExample()
	errorTypesExample()
}

// basicErrorCreation demonstrates simple error creation
func basicErrorCreation() {
	fmt.Println("\n📝 Basic Error Creation:")

	// Simple error
	err1 := domainerrors.New("E001", "User not found")
	fmt.Printf("  Simple Error: %s\n", err1.Error())

	// Error with details using builder
	err2 := domainerrors.NewBuilder().
		WithCode("E002").
		WithMessage("Invalid data").
		WithDetail("field", "email").
		WithDetail("value", "invalid-email@").
		Build()
	fmt.Printf("  With Details: %s\n", err2.Error())
	fmt.Printf("  Details: %+v\n", err2.Details())

	// Error with tags using builder
	err3 := domainerrors.NewBuilder().
		WithCode("E003").
		WithMessage("Processing failed").
		WithTag("critical").
		WithTag("payment").
		Build()
	fmt.Printf("  With Tags: %s\n", err3.Error())
	fmt.Printf("  Tags: %v\n", err3.Tags())
}

// builderPatternExample shows fluent error construction
func builderPatternExample() {
	fmt.Println("\n🏗️  Builder Pattern Example:")

	ctx := context.Background()

	err := domainerrors.NewBuilder().
		WithCode("USR001").
		WithMessage("User validation failed").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("user_id", "12345").
		WithDetail("email", "user@example.com").
		WithTag("validation").
		WithTag("user_management").
		WithContext(ctx).
		Build()

	fmt.Printf("  Builder Error: %s\n", err.Error())
	fmt.Printf("  Type: %s\n", err.Type())
	fmt.Printf("  Details: %+v\n", err.Details())
	fmt.Printf("  Tags: %v\n", err.Tags())
}

// validationErrorExample demonstrates validation-specific errors
func validationErrorExample() {
	fmt.Println("\n✅ Validation Error Example:")

	fields := map[string][]string{
		"email":    {"invalid format", "required"},
		"age":      {"must be positive", "required"},
		"password": {"too short", "missing special character"},
	}

	validationErr := domainerrors.NewValidationError("User data validation failed", fields)

	fmt.Printf("  Validation Error: %s\n", validationErr.Error())
	fmt.Printf("  Validated Fields: %+v\n", validationErr.Fields())

	// Add more fields
	validationErr.AddField("phone", "invalid country code")
	fmt.Printf("  After adding field: %+v\n", validationErr.Fields())
}

// jsonSerializationExample shows JSON serialization capabilities
func jsonSerializationExample() {
	fmt.Println("\n📄 JSON Serialization Example:")

	err := domainerrors.NewBuilder().
		WithCode("API001").
		WithMessage("Request processing failed").
		WithType(string(types.ErrorTypeBadRequest)).
		WithDetail("endpoint", "/api/users").
		WithDetail("method", "POST").
		WithDetail("timestamp", time.Now().Format(time.RFC3339)).
		WithTag("api").
		WithTag("http").
		Build()

	// Serialize to JSON
	jsonData, jsonErr := json.MarshalIndent(err, "", "  ")
	if jsonErr != nil {
		fmt.Printf("  JSON serialization failed: %v\n", jsonErr)
		return
	}

	fmt.Printf("  JSON Representation:\n%s\n", string(jsonData))

	// Deserialize from JSON
	var deserializedErr domainerrors.DomainError
	if unmarshalErr := json.Unmarshal(jsonData, &deserializedErr); unmarshalErr != nil {
		fmt.Printf("  JSON deserialization failed: %v\n", unmarshalErr)
		return
	}

	fmt.Printf("  Deserialized Error: %s\n", deserializedErr.Error())
	fmt.Printf("  Deserialized Type: %s\n", deserializedErr.Type())
}

// errorTypesExample demonstrates different error types
func errorTypesExample() {
	fmt.Println("\n🏷️  Error Types Example:")

	errorExamples := []struct {
		name string
		err  error
	}{
		{
			"NotFound",
			domainerrors.NewBuilder().
				WithCode("E404").
				WithMessage("Resource not found").
				WithType(string(types.ErrorTypeNotFound)).
				Build(),
		},
		{
			"Validation",
			domainerrors.NewBuilder().
				WithCode("E400").
				WithMessage("Invalid input").
				WithType(string(types.ErrorTypeValidation)).
				Build(),
		},
		{
			"BusinessRule",
			domainerrors.NewBuilder().
				WithCode("E422").
				WithMessage("Business rule violation").
				WithType(string(types.ErrorTypeBusinessRule)).
				Build(),
		},
		{
			"Authentication",
			domainerrors.NewBuilder().
				WithCode("E401").
				WithMessage("Authentication failed").
				WithType(string(types.ErrorTypeAuthentication)).
				Build(),
		},
		{
			"Authorization",
			domainerrors.NewBuilder().
				WithCode("E403").
				WithMessage("Access denied").
				WithType(string(types.ErrorTypeAuthorization)).
				Build(),
		},
	}

	for _, example := range errorExamples {
		if domainErr, ok := example.err.(*domainerrors.DomainError); ok {
			fmt.Printf("  %s: %s (Type: %s)\n",
				example.name,
				domainErr.Error(),
				domainErr.Type())
		}
	}
}
