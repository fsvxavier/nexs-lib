// Package main demonstrates real-world JSON Schema validation scenarios
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/checks"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/hooks"
)

func main() {
	fmt.Println("=== Real-World JSON Schema Validation Scenarios ===\n")

	// Example 1: API Request Validation
	fmt.Println("1. API Request Validation:")
	apiRequestValidation()

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Example 2: Configuration File Validation
	fmt.Println("2. Configuration File Validation:")
	configFileValidation()

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Example 3: E-commerce Product Validation
	fmt.Println("3. E-commerce Product Validation:")
	ecommerceValidation()

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Example 4: User Registration with Complex Rules
	fmt.Println("4. User Registration Validation:")
	userRegistrationValidation()
}

func apiRequestValidation() {
	// Scenario: REST API endpoint that accepts user creation requests

	cfg := config.NewConfig()

	// Add request logging hook
	cfg.AddPreValidationHook(&hooks.LoggingHook{LogData: false}) // Don't log sensitive data

	// Add validation summary for monitoring
	cfg.AddPostValidationHook(&hooks.ValidationSummaryHook{LogSummary: true})

	// Add error limiting to prevent abuse
	cfg.AddErrorHook(&hooks.ErrorFilterHook{MaxErrors: 5})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// API schema for user creation
	apiSchema := []byte(`{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"firstName": {
						"type": "string",
						"minLength": 2,
						"maxLength": 50,
						"pattern": "^[a-zA-ZÀ-ÿ\\s]+$"
					},
					"lastName": {
						"type": "string",
						"minLength": 2,
						"maxLength": 50,
						"pattern": "^[a-zA-ZÀ-ÿ\\s]+$"
					},
					"email": {
						"type": "string",
						"format": "email",
						"maxLength": 100
					},
					"password": {
						"type": "string",
						"minLength": 8,
						"maxLength": 128
					},
					"birthDate": {
						"type": "string",
						"format": "date"
					},
					"preferences": {
						"type": "object",
						"properties": {
							"newsletter": {"type": "boolean"},
							"language": {
								"type": "string",
								"enum": ["en", "pt", "es", "fr"]
							}
						}
					}
				},
				"required": ["firstName", "lastName", "email", "password"]
			}
		},
		"required": ["user"],
		"additionalProperties": false
	}`)

	// Valid API request
	validRequest := map[string]interface{}{
		"user": map[string]interface{}{
			"firstName": "Maria",
			"lastName":  "Silva",
			"email":     "maria.silva@example.com",
			"password":  "SecurePass123!",
			"birthDate": "1990-05-15",
			"preferences": map[string]interface{}{
				"newsletter": true,
				"language":   "pt",
			},
		},
	}

	fmt.Println("📥 Validating API request (POST /api/users):")

	start := time.Now()
	errors, err := validator.ValidateFromBytes(apiSchema, validRequest)
	duration := time.Since(start)

	if err != nil {
		log.Printf("API validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Printf("✅ API request valid (processed in %v)\n", duration)
		fmt.Println("   → User creation would proceed")
	} else {
		fmt.Printf("❌ API request invalid (%d errors in %v):\n", len(errors), duration)
		for _, ve := range errors {
			fmt.Printf("   - %s: %s\n", ve.Field, ve.Message)
		}
		fmt.Println("   → HTTP 400 Bad Request would be returned")
	}

	// Invalid API request example
	invalidRequest := map[string]interface{}{
		"user": map[string]interface{}{
			"firstName": "A",            // Too short
			"lastName":  "",             // Empty
			"email":     "not-an-email", // Invalid format
			"password":  "123",          // Too short
			"birthDate": "invalid-date", // Invalid date
			"preferences": map[string]interface{}{
				"language": "invalid", // Not in enum
			},
		},
		"extraField": "not allowed", // Additional property
	}

	fmt.Println("\n📥 Validating invalid API request:")
	errors, err = validator.ValidateFromBytes(apiSchema, invalidRequest)
	if err != nil {
		log.Printf("API validation error: %v", err)
		return
	}

	if len(errors) > 0 {
		fmt.Printf("❌ API request invalid (%d errors):\n", len(errors))
		for i, ve := range errors {
			if i < 3 { // Show only first 3 errors for brevity
				fmt.Printf("   - %s: %s\n", ve.Field, ve.Message)
			}
		}
		if len(errors) > 3 {
			fmt.Printf("   ... and %d more errors\n", len(errors)-3)
		}
	}
}

func configFileValidation() {
	// Scenario: Application configuration file validation

	cfg := config.NewConfig().WithProvider(config.JSONSchemaProvider) // Use fast provider for config

	// Add specific checks for configuration
	cfg.AddCheck(&checks.RequiredFieldsCheck{
		RequiredFields: []string{"database", "server", "logging"},
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Configuration schema
	configSchema := []byte(`{
		"type": "object",
		"properties": {
			"database": {
				"type": "object",
				"properties": {
					"host": {"type": "string"},
					"port": {"type": "integer", "minimum": 1, "maximum": 65535},
					"name": {"type": "string"},
					"user": {"type": "string"},
					"password": {"type": "string"},
					"maxConnections": {"type": "integer", "minimum": 1},
					"timeout": {"type": "integer", "minimum": 1}
				},
				"required": ["host", "port", "name", "user"]
			},
			"server": {
				"type": "object",
				"properties": {
					"host": {"type": "string"},
					"port": {"type": "integer", "minimum": 1, "maximum": 65535},
					"tls": {
						"type": "object",
						"properties": {
							"enabled": {"type": "boolean"},
							"certFile": {"type": "string"},
							"keyFile": {"type": "string"}
						}
					}
				},
				"required": ["host", "port"]
			},
			"logging": {
				"type": "object",
				"properties": {
					"level": {
						"type": "string",
						"enum": ["debug", "info", "warn", "error"]
					},
					"format": {
						"type": "string",
						"enum": ["json", "text"]
					},
					"output": {
						"type": "string",
						"enum": ["stdout", "stderr", "file"]
					},
					"file": {"type": "string"}
				},
				"required": ["level"]
			}
		}
	}`)

	// Valid configuration
	configData := map[string]interface{}{
		"database": map[string]interface{}{
			"host":           "localhost",
			"port":           5432,
			"name":           "myapp",
			"user":           "dbuser",
			"password":       "secret",
			"maxConnections": 10,
			"timeout":        30,
		},
		"server": map[string]interface{}{
			"host": "0.0.0.0",
			"port": 8080,
			"tls": map[string]interface{}{
				"enabled":  true,
				"certFile": "/etc/ssl/cert.pem",
				"keyFile":  "/etc/ssl/key.pem",
			},
		},
		"logging": map[string]interface{}{
			"level":  "info",
			"format": "json",
			"output": "stdout",
		},
	}

	fmt.Println("⚙️  Validating application configuration:")

	errors, err := validator.ValidateFromBytes(configSchema, configData)
	if err != nil {
		log.Printf("Config validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Configuration is valid")
		fmt.Println("   → Application can start safely")

		// Show processed config
		configJSON, _ := json.MarshalIndent(configData, "   ", "  ")
		fmt.Printf("   Configuration loaded:\n%s\n", string(configJSON))
	} else {
		fmt.Printf("❌ Configuration is invalid (%d errors):\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("   - %s: %s\n", ve.Field, ve.Message)
		}
		fmt.Println("   → Application startup would fail")
	}
}

func ecommerceValidation() {
	// Scenario: E-commerce product catalog validation with complex business rules

	cfg := config.NewConfig()

	// Add business validation checks
	cfg.AddCheck(&checks.RequiredFieldsCheck{
		RequiredFields: []string{"sku", "name", "price", "category"},
	})

	cfg.AddCheck(&checks.EnumConstraintsCheck{
		Constraints: map[string][]interface{}{
			"category": {"electronics", "clothing", "books", "home", "sports"},
			"status":   {"active", "inactive", "discontinued"},
		},
	})

	// Add data normalization
	cfg.AddPreValidationHook(&hooks.DataNormalizationHook{
		TrimStrings:   true,
		LowerCaseKeys: false,
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Product schema
	productSchema := []byte(`{
		"type": "object",
		"properties": {
			"sku": {
				"type": "string",
				"pattern": "^[A-Z0-9]{3}-[A-Z0-9]{3}-[A-Z0-9]{3}$"
			},
			"name": {
				"type": "string",
				"minLength": 5,
				"maxLength": 100
			},
			"description": {
				"type": "string",
				"maxLength": 1000
			},
			"price": {
				"type": "number",
				"minimum": 0.01,
				"maximum": 999999.99
			},
			"currency": {
				"type": "string",
				"pattern": "^[A-Z]{3}$"
			},
			"category": {"type": "string"},
			"status": {"type": "string"},
			"inventory": {
				"type": "object",
				"properties": {
					"quantity": {"type": "integer", "minimum": 0},
					"warehouse": {"type": "string"},
					"reserved": {"type": "integer", "minimum": 0}
				},
				"required": ["quantity"]
			},
			"dimensions": {
				"type": "object",
				"properties": {
					"length": {"type": "number", "minimum": 0},
					"width": {"type": "number", "minimum": 0},
					"height": {"type": "number", "minimum": 0},
					"weight": {"type": "number", "minimum": 0}
				}
			},
			"images": {
				"type": "array",
				"items": {
					"type": "string",
					"format": "uri"
				},
				"minItems": 1,
				"maxItems": 10
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"uniqueItems": true
			}
		}
	}`)

	// Valid product
	product := map[string]interface{}{
		"sku":         "ELC-LAP-001",
		"name":        "Premium Gaming Laptop",
		"description": "High-performance gaming laptop with RGB keyboard",
		"price":       1299.99,
		"currency":    "USD",
		"category":    "electronics",
		"status":      "active",
		"inventory": map[string]interface{}{
			"quantity":  25,
			"warehouse": "US-WEST",
			"reserved":  3,
		},
		"dimensions": map[string]interface{}{
			"length": 35.5,
			"width":  24.5,
			"height": 2.3,
			"weight": 2.1,
		},
		"images": []interface{}{
			"https://example.com/images/laptop-1.jpg",
			"https://example.com/images/laptop-2.jpg",
		},
		"tags": []interface{}{"gaming", "laptop", "rgb", "premium"},
	}

	fmt.Println("🛒 Validating e-commerce product:")

	errors, err := validator.ValidateFromBytes(productSchema, product)
	if err != nil {
		log.Printf("Product validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Product is valid for catalog")
		fmt.Printf("   SKU: %s\n", product["sku"])
		fmt.Printf("   Name: %s\n", product["name"])
		fmt.Printf("   Price: $%.2f %s\n", product["price"], product["currency"])
		fmt.Printf("   Category: %s\n", product["category"])
		fmt.Printf("   Stock: %v units\n", product["inventory"].(map[string]interface{})["quantity"])
	} else {
		fmt.Printf("❌ Product validation failed (%d errors):\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("   - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func userRegistrationValidation() {
	// Scenario: Complex user registration with business rules and security checks

	cfg := config.NewConfig()

	// Add comprehensive checks
	cfg.AddCheck(&checks.RequiredFieldsCheck{
		RequiredFields: []string{"username", "email", "password", "termsAccepted"},
	})

	// Add date validation for age verification
	cfg.AddCheck(&checks.DateValidationCheck{
		DateFields:  []string{"birthDate"},
		AllowFuture: false,
		AllowPast:   true,
	})

	// Add pre-validation hooks
	cfg.AddPreValidationHook(&hooks.DataNormalizationHook{
		TrimStrings:   true,
		LowerCaseKeys: false,
	})

	// Add post-validation hooks
	cfg.AddPostValidationHook(&hooks.ErrorEnrichmentHook{
		AddContext:     true,
		AddSuggestions: true,
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// User registration schema with security requirements
	registrationSchema := []byte(`{
		"type": "object",
		"properties": {
			"username": {
				"type": "string",
				"minLength": 3,
				"maxLength": 20,
				"pattern": "^[a-zA-Z0-9_]+$"
			},
			"email": {
				"type": "string",
				"format": "email",
				"maxLength": 100
			},
			"password": {
				"type": "string",
				"minLength": 8,
				"maxLength": 128,
				"pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]"
			},
			"confirmPassword": {
				"type": "string"
			},
			"firstName": {
				"type": "string",
				"minLength": 2,
				"maxLength": 50
			},
			"lastName": {
				"type": "string",
				"minLength": 2,
				"maxLength": 50
			},
			"birthDate": {
				"type": "string",
				"format": "date"
			},
			"phone": {
				"type": "string",
				"pattern": "^\\+?[1-9]\\d{1,14}$"
			},
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"city": {"type": "string"},
					"state": {"type": "string"},
					"zipCode": {"type": "string"},
					"country": {"type": "string", "minLength": 2, "maxLength": 2}
				}
			},
			"marketingConsent": {"type": "boolean"},
			"termsAccepted": {"type": "boolean", "const": true},
			"privacyAccepted": {"type": "boolean", "const": true}
		},
		"additionalProperties": false
	}`)

	// Registration data with issues
	registrationData := map[string]interface{}{
		"username":        "  john_doe  ", // Will be trimmed
		"email":           "john@example.com",
		"password":        "weak",      // Too weak
		"confirmPassword": "different", // Doesn't match
		"firstName":       "John",
		"lastName":        "Doe",
		"birthDate":       "2010-01-01", // Too young
		"phone":           "555-0123",
		"address": map[string]interface{}{
			"street":  "123 Main St",
			"city":    "Anytown",
			"state":   "CA",
			"zipCode": "12345",
			"country": "US",
		},
		"marketingConsent": false,
		"termsAccepted":    true,
		"privacyAccepted":  true,
	}

	fmt.Println("👤 Validating user registration:")

	errors, err := validator.ValidateFromBytes(registrationSchema, registrationData)
	if err != nil {
		log.Printf("Registration validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Registration data is valid")
		fmt.Printf("   Welcome, %s %s!\n",
			registrationData["firstName"],
			registrationData["lastName"])
		fmt.Printf("   Username: %s\n", registrationData["username"])
		fmt.Printf("   Email: %s\n", registrationData["email"])
	} else {
		fmt.Printf("❌ Registration validation failed (%d errors):\n", len(errors))
		for i, ve := range errors {
			fmt.Printf("   %d. %s: %s\n", i+1, ve.Field, ve.Message)
			if ve.Description != "" {
				fmt.Printf("      → %s\n", ve.Description)
			}
		}
		fmt.Println("\n   🔒 Security notes:")
		fmt.Println("   - Password must contain uppercase, lowercase, number, and special character")
		fmt.Println("   - Birth date indicates user must be 18 or older")
		fmt.Println("   - Terms and privacy policy acceptance is mandatory")
	}

	fmt.Println("\n📊 Validation pipeline executed:")
	fmt.Println("   1. ✅ Data normalization (trimmed whitespace)")
	fmt.Println("   2. ✅ Required fields check")
	fmt.Println("   3. ✅ Date validation for age verification")
	fmt.Println("   4. ✅ JSON Schema validation")
	fmt.Println("   5. ✅ Error enrichment with suggestions")
}
