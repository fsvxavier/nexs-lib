package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// ErrorEnvelope represents a serializable error envelope
type ErrorEnvelope struct {
	Error     ErrorInfo `json:"error" xml:"error"`
	Timestamp string    `json:"timestamp" xml:"timestamp"`
	RequestID string    `json:"request_id,omitempty" xml:"request_id,omitempty"`
	TraceID   string    `json:"trace_id,omitempty" xml:"trace_id,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code       string                 `json:"code" xml:"code"`
	Type       string                 `json:"type" xml:"type"`
	Message    string                 `json:"message" xml:"message"`
	Details    map[string]interface{} `json:"details,omitempty" xml:"details,omitempty"`
	HTTPStatus int                    `json:"http_status" xml:"http_status"`
	Cause      string                 `json:"cause,omitempty" xml:"cause,omitempty"`
}

// SerializeDomainError converts a domain error to a serializable format
func SerializeDomainError(err error, requestID, traceID string) *ErrorEnvelope {
	envelope := &ErrorEnvelope{
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: requestID,
		TraceID:   traceID,
	}

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		envelope.Error = ErrorInfo{
			Code:       domainErr.Code,
			Type:       string(domainErr.Type),
			Message:    domainErr.Message,
			HTTPStatus: domainErr.HTTPStatus(),
		}

		if domainErr.Cause != nil {
			envelope.Error.Cause = domainErr.Cause.Error()
		}

		// Extract specific details based on error type
		details := make(map[string]interface{})

		switch domainErr.Type {
		case domainerrors.ErrorTypeValidation:
			if validationErr, ok := err.(*domainerrors.ValidationError); ok {
				details["fields"] = validationErr.Fields
			}
		case domainerrors.ErrorTypeBusinessRule:
			if businessErr, ok := err.(*domainerrors.BusinessError); ok {
				details["business_code"] = businessErr.BusinessCode
				details["rules"] = businessErr.Rules
			}
		case domainerrors.ErrorTypeNotFound:
			if notFoundErr, ok := err.(*domainerrors.NotFoundError); ok {
				details["resource_type"] = notFoundErr.ResourceType
				details["resource_id"] = notFoundErr.ResourceID
			}
		case domainerrors.ErrorTypeConflict:
			if conflictErr, ok := err.(*domainerrors.ConflictError); ok {
				details["resource"] = conflictErr.Resource
				details["conflict_reason"] = conflictErr.ConflictReason
			}
		case domainerrors.ErrorTypeTimeout:
			if timeoutErr, ok := err.(*domainerrors.TimeoutError); ok {
				details["operation"] = timeoutErr.Operation
				details["duration"] = timeoutErr.Duration.String()
				details["timeout"] = timeoutErr.Timeout.String()
			}
		case domainerrors.ErrorTypeRateLimit:
			if rateLimitErr, ok := err.(*domainerrors.RateLimitError); ok {
				details["limit"] = rateLimitErr.Limit
				details["remaining"] = rateLimitErr.Remaining
				details["reset_time"] = rateLimitErr.ResetTime
				details["window"] = rateLimitErr.Window
			}
		case domainerrors.ErrorTypeExternalService:
			if extErr, ok := err.(*domainerrors.ExternalServiceError); ok {
				details["service"] = extErr.Service
				details["endpoint"] = extErr.Endpoint
				details["status_code"] = extErr.StatusCode
			}
		case domainerrors.ErrorTypeDatabase:
			if dbErr, ok := err.(*domainerrors.DatabaseError); ok {
				details["operation"] = dbErr.Operation
				details["table"] = dbErr.Table
				details["query"] = dbErr.Query
			}
		case domainerrors.ErrorTypeAuthentication:
			if authErr, ok := err.(*domainerrors.AuthenticationError); ok {
				details["scheme"] = authErr.Scheme
				details["token"] = authErr.Token
			}
		case domainerrors.ErrorTypeAuthorization:
			if authzErr, ok := err.(*domainerrors.AuthorizationError); ok {
				details["permission"] = authzErr.Permission
				details["resource"] = authzErr.Resource
			}
		case domainerrors.ErrorTypeServer:
			if serverErr, ok := err.(*domainerrors.ServerError); ok {
				details["request_id"] = serverErr.RequestID
				details["correlation_id"] = serverErr.CorrelationID
				details["component"] = serverErr.Component
			}
		}

		if len(details) > 0 {
			envelope.Error.Details = details
		}
	} else {
		// Handle standard errors
		envelope.Error = ErrorInfo{
			Code:       "UNKNOWN_ERROR",
			Type:       "unknown",
			Message:    err.Error(),
			HTTPStatus: 500,
		}
	}

	return envelope
}

// DeserializeDomainError recreates a domain error from serialized format
func DeserializeDomainError(envelope *ErrorEnvelope) error {
	errorInfo := envelope.Error

	// Basic error information
	code := errorInfo.Code
	message := errorInfo.Message

	// Create base cause error if present
	var cause error
	if errorInfo.Cause != "" {
		cause = fmt.Errorf("%s", errorInfo.Cause)
	}

	// Recreate specific error type based on type field
	switch errorInfo.Type {
	case "validation":
		validationErr := domainerrors.NewValidationError(code, message, cause)
		if fields, ok := errorInfo.Details["fields"].(map[string]interface{}); ok {
			for field, msg := range fields {
				if msgStr, ok := msg.(string); ok {
					validationErr.WithField(field, msgStr)
				}
			}
		}
		return validationErr

	case "business_rule":
		businessErr := domainerrors.NewBusinessError(code, message)
		if businessCode, ok := errorInfo.Details["business_code"].(string); ok {
			businessErr.BusinessCode = businessCode
		}
		if rules, ok := errorInfo.Details["rules"].([]interface{}); ok {
			for _, rule := range rules {
				if ruleStr, ok := rule.(string); ok {
					businessErr.WithRule(ruleStr)
				}
			}
		}
		return businessErr

	case "not_found":
		notFoundErr := domainerrors.NewNotFoundError(code, message, cause)
		if resourceType, ok := errorInfo.Details["resource_type"].(string); ok {
			if resourceID, ok := errorInfo.Details["resource_id"].(string); ok {
				notFoundErr.WithResource(resourceType, resourceID)
			}
		}
		return notFoundErr

	case "conflict":
		conflictErr := domainerrors.NewConflictError(code, message)
		if resource, ok := errorInfo.Details["resource"].(string); ok {
			if reason, ok := errorInfo.Details["conflict_reason"].(string); ok {
				conflictErr.WithConflictingResource(resource, reason)
			}
		}
		return conflictErr

	case "timeout":
		timeoutErr := domainerrors.NewTimeoutError(code, "", message, cause)
		if operation, ok := errorInfo.Details["operation"].(string); ok {
			timeoutErr.Operation = operation
		}
		if durationStr, ok := errorInfo.Details["duration"].(string); ok {
			if duration, err := time.ParseDuration(durationStr); err == nil {
				timeoutErr.Duration = duration
			}
		}
		if timeoutStr, ok := errorInfo.Details["timeout"].(string); ok {
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				timeoutErr.Timeout = timeout
			}
		}
		return timeoutErr

	case "rate_limit":
		rateLimitErr := domainerrors.NewRateLimitError(code, message)
		if limit, ok := errorInfo.Details["limit"].(float64); ok {
			rateLimitErr.Limit = int(limit)
		}
		if remaining, ok := errorInfo.Details["remaining"].(float64); ok {
			rateLimitErr.Remaining = int(remaining)
		}
		if resetTime, ok := errorInfo.Details["reset_time"].(string); ok {
			rateLimitErr.ResetTime = resetTime
		}
		if window, ok := errorInfo.Details["window"].(string); ok {
			rateLimitErr.Window = window
		}
		return rateLimitErr

	case "external_service":
		extErr := domainerrors.NewExternalServiceError(code, "", message, cause)
		if service, ok := errorInfo.Details["service"].(string); ok {
			extErr.Service = service
		}
		if endpoint, ok := errorInfo.Details["endpoint"].(string); ok {
			extErr.Endpoint = endpoint
		}
		if statusCode, ok := errorInfo.Details["status_code"].(float64); ok {
			extErr.StatusCode = int(statusCode)
		}
		return extErr

	case "database":
		dbErr := domainerrors.NewDatabaseError(code, message, cause)
		if operation, ok := errorInfo.Details["operation"].(string); ok {
			if table, ok := errorInfo.Details["table"].(string); ok {
				dbErr.WithOperation(operation, table)
			}
		}
		if query, ok := errorInfo.Details["query"].(string); ok {
			dbErr.WithQuery(query)
		}
		return dbErr

	case "authentication":
		authErr := domainerrors.NewAuthenticationError(code, message, cause)
		if scheme, ok := errorInfo.Details["scheme"].(string); ok {
			authErr.WithScheme(scheme)
		}
		if token, ok := errorInfo.Details["token"].(string); ok {
			authErr.Token = token
		}
		return authErr

	case "authorization":
		authzErr := domainerrors.NewAuthorizationError(code, message, cause)
		if permission, ok := errorInfo.Details["permission"].(string); ok {
			if resource, ok := errorInfo.Details["resource"].(string); ok {
				authzErr.WithPermission(permission, resource)
			}
		}
		return authzErr

	case "server":
		serverErr := domainerrors.NewServerError(code, message, cause)
		if requestID, ok := errorInfo.Details["request_id"].(string); ok {
			if correlationID, ok := errorInfo.Details["correlation_id"].(string); ok {
				serverErr.WithRequestInfo(requestID, correlationID)
			}
		}
		if component, ok := errorInfo.Details["component"].(string); ok {
			serverErr.WithComponent(component)
		}
		return serverErr

	default:
		// Return a generic domain error
		return domainerrors.NewServerError(code, message, cause)
	}
}

// JSONSerializationExample demonstrates JSON serialization/deserialization
func JSONSerializationExample() {
	fmt.Println("=== JSON Serialization Example ===")
	fmt.Println()

	// Create various domain errors
	errors := []error{
		domainerrors.NewValidationError("INVALID_EMAIL", "Invalid email format", nil).
			WithField("email", "must be a valid email address").
			WithField("format", "example@domain.com"),

		domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Insufficient funds for transaction").
			WithRule("Account balance must be greater than transaction amount").
			WithRule("Daily transaction limit: $5000"),

		domainerrors.NewNotFoundError("USER_NOT_FOUND", "User not found", nil).
			WithResource("user", "12345"),

		domainerrors.NewRateLimitError("RATE_LIMIT_EXCEEDED", "API rate limit exceeded").
			WithRateLimit(100, 50, time.Now().Add(time.Hour).Format(time.RFC3339), "3600s"),

		domainerrors.NewDatabaseError("DB_CONNECTION_FAILED", "Database connection failed",
			fmt.Errorf("connection timeout")).
			WithOperation("SELECT", "users").
			WithQuery("SELECT * FROM users WHERE id = ?"),
	}

	for i, err := range errors {
		fmt.Printf("--- Error %d: %s ---\n", i+1, err.Error())

		// Serialize to JSON
		envelope := SerializeDomainError(err, fmt.Sprintf("req-%d", i+1), fmt.Sprintf("trace-%d", i+1))
		jsonData, _ := json.MarshalIndent(envelope, "", "  ")
		fmt.Printf("JSON Serialization:\n%s\n", string(jsonData))

		// Deserialize from JSON
		var deserializedEnvelope ErrorEnvelope
		json.Unmarshal(jsonData, &deserializedEnvelope)
		recreatedErr := DeserializeDomainError(&deserializedEnvelope)
		fmt.Printf("Deserialized Error: %s\n", recreatedErr.Error())

		// Verify error type is preserved
		if domainErr, ok := recreatedErr.(*domainerrors.DomainError); ok {
			fmt.Printf("Error Type: %s, Code: %s\n", domainErr.Type, domainErr.Code)
		}

		fmt.Println()
	}
}

// XMLSerializationExample demonstrates XML serialization/deserialization
func XMLSerializationExample() {
	fmt.Println("=== XML Serialization Example ===")
	fmt.Println()

	// Create a complex error
	conflictErr := domainerrors.NewConflictError("EMAIL_ALREADY_EXISTS", "Email address already exists").
		WithConflictingResource("user", "email already registered")

	// Serialize to XML
	envelope := SerializeDomainError(conflictErr, "req-xml-001", "trace-xml-001")
	xmlData, _ := xml.MarshalIndent(envelope, "", "  ")
	fmt.Printf("XML Serialization:\n%s\n", string(xmlData))

	// Deserialize from XML
	var deserializedEnvelope ErrorEnvelope
	xml.Unmarshal(xmlData, &deserializedEnvelope)
	recreatedErr := DeserializeDomainError(&deserializedEnvelope)
	fmt.Printf("Deserialized Error: %s\n", recreatedErr.Error())

	if domainErr, ok := recreatedErr.(*domainerrors.DomainError); ok {
		fmt.Printf("Error Type: %s, Code: %s\n", domainErr.Type, domainErr.Code)
	}

	fmt.Println()
}

// ErrorCollectionExample demonstrates serializing multiple errors
func ErrorCollectionExample() {
	fmt.Println("=== Error Collection Serialization Example ===")
	fmt.Println()

	// Collection of errors
	type ErrorCollection struct {
		Errors    []ErrorEnvelope `json:"errors" xml:"errors"`
		Count     int             `json:"count" xml:"count"`
		Timestamp string          `json:"timestamp" xml:"timestamp"`
	}

	// Create multiple validation errors
	validationErrors := []error{
		domainerrors.NewValidationError("INVALID_NAME", "Name is required", nil).
			WithField("name", "cannot be empty"),
		domainerrors.NewValidationError("INVALID_AGE", "Age must be positive", nil).
			WithField("age", "must be greater than 0"),
		domainerrors.NewValidationError("INVALID_EMAIL", "Invalid email format", nil).
			WithField("email", "must be a valid email address"),
	}

	// Create error collection
	collection := ErrorCollection{
		Count:     len(validationErrors),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	for i, err := range validationErrors {
		envelope := SerializeDomainError(err, fmt.Sprintf("req-%d", i+1), fmt.Sprintf("trace-%d", i+1))
		collection.Errors = append(collection.Errors, *envelope)
	}

	// Serialize collection to JSON
	jsonData, _ := json.MarshalIndent(collection, "", "  ")
	fmt.Printf("JSON Error Collection:\n%s\n", string(jsonData))

	// Deserialize and recreate errors
	var deserializedCollection ErrorCollection
	json.Unmarshal(jsonData, &deserializedCollection)

	fmt.Printf("Deserialized %d errors:\n", deserializedCollection.Count)
	for i, envelope := range deserializedCollection.Errors {
		recreatedErr := DeserializeDomainError(&envelope)
		fmt.Printf("  %d. %s\n", i+1, recreatedErr.Error())
	}

	fmt.Println()
}

// CompactSerializationExample demonstrates compact error format
func CompactSerializationExample() {
	fmt.Println("=== Compact Serialization Example ===")
	fmt.Println()

	// Compact error format for high-volume scenarios
	type CompactError struct {
		C string `json:"c"`           // Code
		M string `json:"m"`           // Message
		T string `json:"t"`           // Type
		S int    `json:"s"`           // Status
		D string `json:"d,omitempty"` // Details (JSON string)
	}

	// Create a complex error
	businessErr := domainerrors.NewBusinessError("ACCOUNT_LOCKED", "Account is locked due to security policy").
		WithRule("Account locked after 5 failed login attempts").
		WithRule("Contact support to unlock")

	// Serialize to compact format
	envelope := SerializeDomainError(businessErr, "req-compact", "trace-compact")
	compact := CompactError{
		C: envelope.Error.Code,
		M: envelope.Error.Message,
		T: envelope.Error.Type,
		S: envelope.Error.HTTPStatus,
	}

	if len(envelope.Error.Details) > 0 {
		detailsJSON, _ := json.Marshal(envelope.Error.Details)
		compact.D = string(detailsJSON)
	}

	compactJSON, _ := json.Marshal(compact)
	fmt.Printf("Compact Format: %s\n", string(compactJSON))

	// Compare sizes
	fullJSON, _ := json.Marshal(envelope)
	fmt.Printf("Full Format Size: %d bytes\n", len(fullJSON))
	fmt.Printf("Compact Format Size: %d bytes\n", len(compactJSON))
	fmt.Printf("Space Savings: %.1f%%\n", float64(len(fullJSON)-len(compactJSON))/float64(len(fullJSON))*100)

	fmt.Println()
}

func main() {
	fmt.Println("=== Domain Errors - Serialization Examples ===")
	fmt.Println()

	// Run serialization examples
	JSONSerializationExample()

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	XMLSerializationExample()

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	ErrorCollectionExample()

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	CompactSerializationExample()

	fmt.Println("=== Serialization Examples Complete ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("- JSON serialization/deserialization of domain errors")
	fmt.Println("- XML serialization/deserialization")
	fmt.Println("- Error collection handling")
	fmt.Println("- Compact serialization for high-volume scenarios")
	fmt.Println("- Preservation of error type and details across serialization")
	fmt.Println("- Request/trace ID tracking in error envelopes")
}
