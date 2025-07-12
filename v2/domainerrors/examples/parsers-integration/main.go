package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🔗 Domain Errors v2 - Parsers Integration Examples")
	fmt.Println("==================================================")

	datetimeParsingExample()
	durationParsingExample()
	environmentParsingExample()
	validationIntegrationExample()
	parseErrorMappingExample()
	parseErrorRecoveryExample()
	parseErrorAggregationExample()
	parseErrorContextExample()
}

// datetimeParsingExample demonstrates datetime parsing with error handling
func datetimeParsingExample() {
	fmt.Println("\n📅 Datetime Parsing Example:")

	parser := NewDateTimeParser()

	testCases := []struct {
		input    string
		format   string
		expected bool
	}{
		{"2024-01-15", "2006-01-02", true},
		{"2024-13-45", "2006-01-02", false}, // Invalid date
		{"15/01/2024", "02/01/2006", true},
		{"invalid-date", "2006-01-02", false},
		{"2024-01-15T10:30:00Z", time.RFC3339, true},
		{"2024-01-15T25:70:80Z", time.RFC3339, false}, // Invalid time
	}

	fmt.Println("  Datetime Parsing Results:")
	for _, tc := range testCases {
		result, err := parser.ParseDateTime(tc.input, tc.format)

		if tc.expected && err == nil {
			fmt.Printf("    ✅ '%s' -> %s\n", tc.input, result.Format("2006-01-02 15:04:05"))
		} else if !tc.expected && err != nil {
			fmt.Printf("    ❌ '%s' -> Error: %s\n", tc.input, err.Error())
			fmt.Printf("       Code: %s, Type: %s\n", err.Code(), err.Type())
			if details := err.Details(); len(details) > 0 {
				fmt.Printf("       Details: %+v\n", details)
			}
		} else {
			fmt.Printf("    ⚠️ Unexpected result for '%s'\n", tc.input)
		}
	}
}

// durationParsingExample demonstrates duration parsing with error handling
func durationParsingExample() {
	fmt.Println("\n⏱️ Duration Parsing Example:")

	parser := NewDurationParser()

	testCases := []struct {
		input    string
		expected bool
	}{
		{"1h30m", true},
		{"2h", true},
		{"45m", true},
		{"30s", true},
		{"1h30m45s", true},
		{"invalid", false},
		{"1x30y", false}, // Invalid units
		{"-5h", false},   // Negative duration not allowed
		{"25h", false},   // More than 24 hours not allowed in this context
	}

	fmt.Println("  Duration Parsing Results:")
	for _, tc := range testCases {
		result, err := parser.ParseDuration(tc.input)

		if tc.expected && err == nil {
			fmt.Printf("    ✅ '%s' -> %v\n", tc.input, result)
		} else if !tc.expected && err != nil {
			fmt.Printf("    ❌ '%s' -> Error: %s\n", tc.input, err.Error())
			fmt.Printf("       Code: %s, Severity: %s\n", err.Code(), err.Severity())
			if suggestions := err.Details()["suggestions"]; suggestions != nil {
				fmt.Printf("       Suggestions: %v\n", suggestions)
			}
		} else {
			fmt.Printf("    ⚠️ Unexpected result for '%s'\n", tc.input)
		}
	}
}

// environmentParsingExample demonstrates environment variable parsing
func environmentParsingExample() {
	fmt.Println("\n🌍 Environment Parsing Example:")

	parser := NewEnvironmentParser()

	// Simulate environment variables
	envVars := map[string]string{
		"APP_PORT":             "8080",
		"APP_DEBUG":            "true",
		"APP_TIMEOUT":          "30s",
		"APP_MAX_RETRIES":      "3",
		"APP_RATE_LIMIT":       "100",
		"APP_INVALID_PORT":     "not-a-number",
		"APP_INVALID_BOOL":     "maybe",
		"APP_INVALID_DURATION": "forever",
	}

	// Set up configuration requirements
	requirements := map[string]EnvRequirement{
		"APP_PORT": {
			Type:     "int",
			Required: true,
			Min:      1,
			Max:      65535,
		},
		"APP_DEBUG": {
			Type:     "bool",
			Required: false,
			Default:  "false",
		},
		"APP_TIMEOUT": {
			Type:     "duration",
			Required: true,
		},
		"APP_MAX_RETRIES": {
			Type:     "int",
			Required: true,
			Min:      0,
			Max:      10,
		},
		"APP_RATE_LIMIT": {
			Type:     "int",
			Required: true,
			Min:      1,
		},
		"APP_INVALID_PORT": {
			Type:     "int",
			Required: true,
			Min:      1,
			Max:      65535,
		},
		"APP_INVALID_BOOL": {
			Type:     "bool",
			Required: true,
		},
		"APP_INVALID_DURATION": {
			Type:     "duration",
			Required: true,
		},
		"APP_MISSING_REQUIRED": {
			Type:     "string",
			Required: true,
		},
	}

	fmt.Println("  Environment Parsing Results:")
	config, errors := parser.ParseEnvironment(envVars, requirements)

	if len(errors) == 0 {
		fmt.Println("    ✅ All environment variables parsed successfully")
		for key, value := range config {
			fmt.Printf("    %s = %v\n", key, value)
		}
	} else {
		fmt.Printf("    ❌ Found %d parsing errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("    %s: %s\n", err.Code(), err.Error())
			fmt.Printf("      Type: %s, Severity: %s\n", err.Type(), err.Severity())
			if field := err.Details()["field"]; field != nil {
				fmt.Printf("      Field: %s\n", field)
			}
			if expected := err.Details()["expected_type"]; expected != nil {
				fmt.Printf("      Expected Type: %s\n", expected)
			}
		}
	}
}

// validationIntegrationExample shows integration with validation systems
func validationIntegrationExample() {
	fmt.Println("\n✅ Validation Integration Example:")

	validator := NewIntegratedValidator()

	// Test user data with parsing and validation
	userData := map[string]string{
		"name":       "John Doe",
		"email":      "john.doe@example.com",
		"age":        "25",
		"birth_date": "1999-01-15",
		"phone":      "+1-555-123-4567",
		"salary":     "75000.50",
		"start_date": "2024-01-01",
		"department": "engineering",
	}

	invalidUserData := map[string]string{
		"name":       "",
		"email":      "invalid-email",
		"age":        "not-a-number",
		"birth_date": "invalid-date",
		"phone":      "123",
		"salary":     "not-a-salary",
		"start_date": "2025-13-45",
		"department": "",
	}

	fmt.Println("  Valid User Data:")
	validUser, err := validator.ValidateAndParseUser(userData)
	if err == nil {
		fmt.Printf("    ✅ User parsed successfully: %+v\n", validUser)
	} else {
		fmt.Printf("    ❌ Validation failed: %s\n", err.Error())
	}

	fmt.Println("\n  Invalid User Data:")
	_, err = validator.ValidateAndParseUser(invalidUserData)
	if err != nil {
		if validationErr, ok := err.(interfaces.ValidationErrorInterface); ok {
			fmt.Printf("    ❌ Validation failed with multiple field errors:\n")
			// Show all validation details
			fmt.Printf("    Error Message: %s\n", validationErr.Error())
			fmt.Printf("    Validation Code: %s\n", validationErr.Code())
		} else {
			fmt.Printf("    ❌ Validation failed: %s\n", err.Error())
		}
	}
}

// parseErrorMappingExample demonstrates mapping parse errors to domain errors
func parseErrorMappingExample() {
	fmt.Println("\n🗺️ Parse Error Mapping Example:")

	mapper := NewParseErrorMapper()

	// Register error mappings
	mapper.RegisterMapping("strconv", "INT_PARSE_ERROR", func(err error, context map[string]interface{}) interfaces.DomainErrorInterface {
		return factory.GetDefaultFactory().Builder().
			WithCode("INT_PARSE_ERROR").
			WithMessage("Failed to parse integer value").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityLow)).
			WithDetail("original_error", err.Error()).
			WithDetail("field", context["field"]).
			WithDetail("value", context["value"]).
			WithTag("parsing").
			WithTag("integer").
			Build()
	})

	mapper.RegisterMapping("time", "DATETIME_PARSE_ERROR", func(err error, context map[string]interface{}) interfaces.DomainErrorInterface {
		return factory.GetDefaultFactory().Builder().
			WithCode("DATETIME_PARSE_ERROR").
			WithMessage("Failed to parse datetime value").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("original_error", err.Error()).
			WithDetail("field", context["field"]).
			WithDetail("value", context["value"]).
			WithDetail("expected_format", context["format"]).
			WithTag("parsing").
			WithTag("datetime").
			Build()
	})

	// Test error mapping
	testCases := []struct {
		errType string
		context map[string]interface{}
	}{
		{
			errType: "strconv",
			context: map[string]interface{}{
				"field": "age",
				"value": "not-a-number",
			},
		},
		{
			errType: "time",
			context: map[string]interface{}{
				"field":  "birth_date",
				"value":  "invalid-date",
				"format": "2006-01-02",
			},
		},
	}

	fmt.Println("  Parse Error Mapping Results:")
	for _, tc := range testCases {
		originalErr := fmt.Errorf("simulated %s error", tc.errType)
		domainErr := mapper.MapError(tc.errType, originalErr, tc.context)

		fmt.Printf("    %s Error:\n", tc.errType)
		fmt.Printf("      Code: %s\n", domainErr.Code())
		fmt.Printf("      Message: %s\n", domainErr.Error())
		fmt.Printf("      Type: %s\n", domainErr.Type())
		fmt.Printf("      Details: %+v\n", domainErr.Details())
		fmt.Printf("      Tags: %v\n", domainErr.Tags())
	}
}

// parseErrorRecoveryExample demonstrates error recovery strategies
func parseErrorRecoveryExample() {
	fmt.Println("\n🔄 Parse Error Recovery Example:")

	recoverer := NewParseErrorRecoverer()

	testCases := []struct {
		name      string
		input     string
		parseType string
	}{
		{"Recoverable Integer", "123abc", "int"},
		{"Recoverable Float", "45.67xyz", "float"},
		{"Recoverable Boolean", "TRUE", "bool"},
		{"Recoverable Date", "2024/01/15", "date"},
		{"Non-recoverable", "completely-invalid", "int"},
	}

	fmt.Println("  Parse Error Recovery Results:")
	for _, tc := range testCases {
		fmt.Printf("    %s ('%s'):\n", tc.name, tc.input)

		result, recovered, err := recoverer.RecoverableParse(tc.input, tc.parseType)

		if err == nil {
			fmt.Printf("      ✅ Parsed successfully: %v\n", result)
		} else if recovered {
			fmt.Printf("      🔄 Recovered with fallback: %v\n", result)
			fmt.Printf("      Original error: %s\n", err.Error())
		} else {
			fmt.Printf("      ❌ Could not recover: %s\n", err.Error())
			fmt.Printf("      Suggestions: %v\n", err.Details()["suggestions"])
		}
	}
}

// parseErrorAggregationExample demonstrates aggregating multiple parse errors
func parseErrorAggregationExample() {
	fmt.Println("\n📊 Parse Error Aggregation Example:")

	aggregator := NewParseErrorAggregator()

	// Simulate parsing multiple fields with errors
	fieldData := map[string]string{
		"user_id":    "not-a-number",
		"email":      "invalid-email",
		"age":        "too-old",
		"birth_date": "invalid-date",
		"salary":     "not-a-salary",
		"active":     "maybe",
	}

	fieldTypes := map[string]string{
		"user_id":    "int",
		"email":      "email",
		"age":        "int",
		"birth_date": "date",
		"salary":     "float",
		"active":     "bool",
	}

	fmt.Println("  Parse Error Aggregation Results:")

	for field, value := range fieldData {
		expectedType := fieldTypes[field]
		err := aggregator.ParseField(field, value, expectedType)
		if err != nil {
			aggregator.AddError(err)
		}
	}

	if aggregator.HasErrors() {
		aggregatedErr := aggregator.BuildAggregatedError()
		fmt.Printf("    ❌ Aggregated Parse Error:\n")
		fmt.Printf("      Total Fields with Errors: %d\n", len(aggregator.GetErrors()))
		fmt.Printf("      Error Summary: %s\n", aggregatedErr.Error())

		if validationErr, ok := aggregatedErr.(interfaces.ValidationErrorInterface); ok {
			fmt.Println("      Field Error Summary:")
			fmt.Printf("        Validation Message: %s\n", validationErr.Error())
			fmt.Printf("        Validation Code: %s\n", validationErr.Code())
		}
	} else {
		fmt.Println("    ✅ All fields parsed successfully")
	}
}

// parseErrorContextExample demonstrates contextual parse error information
func parseErrorContextExample() {
	fmt.Println("\n🎯 Parse Error Context Example:")

	contextParser := NewContextualParser()

	// Simulate parsing configuration file
	configData := `
	server:
	  port: "not-a-number"
	  host: "localhost"
	  timeout: "invalid-duration"
	  
	database:
	  max_connections: "too-many"
	  retry_attempts: "3"
	  connection_timeout: "30s"
	  
	features:
	  debug_mode: "maybe"
	  cache_enabled: "true"
	  rate_limit: "100"
	`

	fmt.Println("  Contextual Parse Results:")
	config, errors := contextParser.ParseConfiguration(configData)

	if len(errors) == 0 {
		fmt.Println("    ✅ Configuration parsed successfully")
		fmt.Printf("    Config: %+v\n", config)
	} else {
		fmt.Printf("    ❌ Found %d parsing errors in configuration:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("      %s: %s\n", err.Code(), err.Error())
			fmt.Printf("        Section: %s\n", err.Details()["section"])
			fmt.Printf("        Key: %s\n", err.Details()["key"])
			fmt.Printf("        Line: %v\n", err.Details()["line"])
			fmt.Printf("        Expected Type: %s\n", err.Details()["expected_type"])
			if context := err.Details()["context"]; context != nil {
				fmt.Printf("        Context: %s\n", context)
			}
		}
	}
}

// Parser Implementations

type DateTimeParser struct {
	factory interfaces.ErrorFactory
}

func NewDateTimeParser() *DateTimeParser {
	return &DateTimeParser{
		factory: factory.GetDefaultFactory(),
	}
}

func (p *DateTimeParser) ParseDateTime(input, format string) (time.Time, interfaces.DomainErrorInterface) {
	parsed, err := time.Parse(format, input)
	if err != nil {
		return time.Time{}, p.factory.Builder().
			WithCode("DATETIME_PARSE_ERROR").
			WithMessage("Failed to parse datetime").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("input", input).
			WithDetail("format", format).
			WithDetail("parse_error", err.Error()).
			WithTag("parsing").
			WithTag("datetime").
			Build()
	}
	return parsed, nil
}

type DurationParser struct {
	factory interfaces.ErrorFactory
}

func NewDurationParser() *DurationParser {
	return &DurationParser{
		factory: factory.GetDefaultFactory(),
	}
}

func (p *DurationParser) ParseDuration(input string) (time.Duration, interfaces.DomainErrorInterface) {
	duration, err := time.ParseDuration(input)
	if err != nil {
		suggestions := []string{"1h30m", "45m", "30s", "2h15m30s"}
		return 0, p.factory.Builder().
			WithCode("DURATION_PARSE_ERROR").
			WithMessage("Failed to parse duration").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityLow)).
			WithDetail("input", input).
			WithDetail("parse_error", err.Error()).
			WithDetail("suggestions", suggestions).
			WithTag("parsing").
			WithTag("duration").
			Build()
	}

	// Business rule: no negative durations
	if duration < 0 {
		return 0, p.factory.Builder().
			WithCode("NEGATIVE_DURATION_ERROR").
			WithMessage("Negative durations are not allowed").
			WithType(string(types.ErrorTypeBusinessRule)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("input", input).
			WithDetail("parsed_duration", duration.String()).
			WithTag("parsing").
			WithTag("business_rule").
			Build()
	}

	// Business rule: max 24 hours in this context
	if duration > 24*time.Hour {
		return 0, p.factory.Builder().
			WithCode("DURATION_TOO_LONG_ERROR").
			WithMessage("Duration exceeds maximum allowed (24h)").
			WithType(string(types.ErrorTypeBusinessRule)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("input", input).
			WithDetail("parsed_duration", duration.String()).
			WithDetail("max_allowed", "24h").
			WithTag("parsing").
			WithTag("business_rule").
			Build()
	}

	return duration, nil
}

type EnvironmentParser struct {
	factory interfaces.ErrorFactory
}

type EnvRequirement struct {
	Type     string
	Required bool
	Default  string
	Min      interface{}
	Max      interface{}
}

func NewEnvironmentParser() *EnvironmentParser {
	return &EnvironmentParser{
		factory: factory.GetDefaultFactory(),
	}
}

func (p *EnvironmentParser) ParseEnvironment(envVars map[string]string, requirements map[string]EnvRequirement) (map[string]interface{}, []interfaces.DomainErrorInterface) {
	config := make(map[string]interface{})
	var errors []interfaces.DomainErrorInterface

	for key, req := range requirements {
		value, exists := envVars[key]

		if !exists {
			if req.Required {
				err := p.factory.Builder().
					WithCode("ENV_VAR_MISSING").
					WithMessage(fmt.Sprintf("Required environment variable '%s' is missing", key)).
					WithType(string(types.ErrorTypeConfiguration)).
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("variable", key).
					WithDetail("type", req.Type).
					WithTag("environment").
					WithTag("missing").
					Build()
				errors = append(errors, err)
			} else if req.Default != "" {
				value = req.Default
			} else {
				continue
			}
		}

		parsed, err := p.parseValue(value, req.Type, key, req)
		if err != nil {
			errors = append(errors, err)
		} else {
			config[key] = parsed
		}
	}

	return config, errors
}

func (p *EnvironmentParser) parseValue(value, valueType, key string, req EnvRequirement) (interface{}, interfaces.DomainErrorInterface) {
	switch valueType {
	case "string":
		return value, nil

	case "int":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, p.factory.Builder().
				WithCode("ENV_VAR_INVALID_INT").
				WithMessage(fmt.Sprintf("Environment variable '%s' is not a valid integer", key)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("variable", key).
				WithDetail("value", value).
				WithDetail("expected_type", "integer").
				WithDetail("parse_error", err.Error()).
				WithTag("environment").
				WithTag("parsing").
				Build()
		}

		if req.Min != nil {
			if min, ok := req.Min.(int); ok && parsed < min {
				return nil, p.factory.Builder().
					WithCode("ENV_VAR_BELOW_MIN").
					WithMessage(fmt.Sprintf("Environment variable '%s' value %d is below minimum %d", key, parsed, min)).
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("variable", key).
					WithDetail("value", parsed).
					WithDetail("min", min).
					WithTag("environment").
					WithTag("validation").
					Build()
			}
		}

		if req.Max != nil {
			if max, ok := req.Max.(int); ok && parsed > max {
				return nil, p.factory.Builder().
					WithCode("ENV_VAR_ABOVE_MAX").
					WithMessage(fmt.Sprintf("Environment variable '%s' value %d is above maximum %d", key, parsed, max)).
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("variable", key).
					WithDetail("value", parsed).
					WithDetail("max", max).
					WithTag("environment").
					WithTag("validation").
					Build()
			}
		}

		return parsed, nil

	case "bool":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, p.factory.Builder().
				WithCode("ENV_VAR_INVALID_BOOL").
				WithMessage(fmt.Sprintf("Environment variable '%s' is not a valid boolean", key)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("variable", key).
				WithDetail("value", value).
				WithDetail("expected_type", "boolean").
				WithDetail("valid_values", []string{"true", "false", "1", "0"}).
				WithTag("environment").
				WithTag("parsing").
				Build()
		}
		return parsed, nil

	case "duration":
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return nil, p.factory.Builder().
				WithCode("ENV_VAR_INVALID_DURATION").
				WithMessage(fmt.Sprintf("Environment variable '%s' is not a valid duration", key)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("variable", key).
				WithDetail("value", value).
				WithDetail("expected_type", "duration").
				WithDetail("examples", []string{"1h", "30m", "45s", "1h30m"}).
				WithTag("environment").
				WithTag("parsing").
				Build()
		}
		return parsed, nil

	default:
		return nil, p.factory.Builder().
			WithCode("ENV_VAR_UNKNOWN_TYPE").
			WithMessage(fmt.Sprintf("Unknown type '%s' for environment variable '%s'", valueType, key)).
			WithType(string(types.ErrorTypeConfiguration)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("variable", key).
			WithDetail("unknown_type", valueType).
			WithTag("environment").
			WithTag("configuration").
			Build()
	}
}

// User validation example
type User struct {
	Name       string
	Email      string
	Age        int
	BirthDate  time.Time
	Phone      string
	Salary     float64
	StartDate  time.Time
	Department string
}

type IntegratedValidator struct {
	factory interfaces.ErrorFactory
}

func NewIntegratedValidator() *IntegratedValidator {
	return &IntegratedValidator{
		factory: factory.GetDefaultFactory(),
	}
}

func (v *IntegratedValidator) ValidateAndParseUser(data map[string]string) (*User, interfaces.DomainErrorInterface) {
	user := &User{}
	fieldErrors := make(map[string][]string)

	// Parse and validate name
	if name := data["name"]; name == "" {
		fieldErrors["name"] = []string{"name is required"}
	} else {
		user.Name = name
	}

	// Parse and validate email
	if email := data["email"]; email == "" {
		fieldErrors["email"] = []string{"email is required"}
	} else if !strings.Contains(email, "@") {
		fieldErrors["email"] = []string{"email must contain @"}
	} else {
		user.Email = email
	}

	// Parse and validate age
	if ageStr := data["age"]; ageStr == "" {
		fieldErrors["age"] = []string{"age is required"}
	} else if age, err := strconv.Atoi(ageStr); err != nil {
		fieldErrors["age"] = []string{"age must be a valid integer"}
	} else if age < 18 || age > 65 {
		fieldErrors["age"] = []string{"age must be between 18 and 65"}
	} else {
		user.Age = age
	}

	// Parse birth date
	if birthDateStr := data["birth_date"]; birthDateStr != "" {
		if birthDate, err := time.Parse("2006-01-02", birthDateStr); err != nil {
			fieldErrors["birth_date"] = []string{"birth_date must be in YYYY-MM-DD format"}
		} else {
			user.BirthDate = birthDate
		}
	}

	// Validate phone
	if phone := data["phone"]; phone == "" {
		fieldErrors["phone"] = []string{"phone is required"}
	} else if len(phone) < 10 {
		fieldErrors["phone"] = []string{"phone must be at least 10 characters"}
	} else {
		user.Phone = phone
	}

	// Parse salary
	if salaryStr := data["salary"]; salaryStr != "" {
		if salary, err := strconv.ParseFloat(salaryStr, 64); err != nil {
			fieldErrors["salary"] = []string{"salary must be a valid number"}
		} else if salary < 0 {
			fieldErrors["salary"] = []string{"salary cannot be negative"}
		} else {
			user.Salary = salary
		}
	}

	// Parse start date
	if startDateStr := data["start_date"]; startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err != nil {
			fieldErrors["start_date"] = []string{"start_date must be in YYYY-MM-DD format"}
		} else {
			user.StartDate = startDate
		}
	}

	// Validate department
	if department := data["department"]; department == "" {
		fieldErrors["department"] = []string{"department is required"}
	} else {
		user.Department = department
	}

	if len(fieldErrors) > 0 {
		return nil, factory.GetDefaultFactory().NewValidation("User validation failed", fieldErrors)
	}

	return user, nil
}

// Parse Error Mapper
type ParseErrorMapper struct {
	mappings map[string]func(error, map[string]interface{}) interfaces.DomainErrorInterface
}

func NewParseErrorMapper() *ParseErrorMapper {
	return &ParseErrorMapper{
		mappings: make(map[string]func(error, map[string]interface{}) interfaces.DomainErrorInterface),
	}
}

func (m *ParseErrorMapper) RegisterMapping(errorType string, code string, mapper func(error, map[string]interface{}) interfaces.DomainErrorInterface) {
	m.mappings[errorType] = mapper
}

func (m *ParseErrorMapper) MapError(errorType string, err error, context map[string]interface{}) interfaces.DomainErrorInterface {
	if mapper, exists := m.mappings[errorType]; exists {
		return mapper(err, context)
	}

	return factory.GetDefaultFactory().Builder().
		WithCode("UNMAPPED_PARSE_ERROR").
		WithMessage("Unmapped parse error occurred").
		WithType(string(types.ErrorTypeInternal)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("error_type", errorType).
		WithDetail("original_error", err.Error()).
		WithDetail("context", context).
		WithTag("parsing").
		WithTag("unmapped").
		Build()
}

// Parse Error Recoverer
type ParseErrorRecoverer struct {
	factory interfaces.ErrorFactory
}

func NewParseErrorRecoverer() *ParseErrorRecoverer {
	return &ParseErrorRecoverer{
		factory: factory.GetDefaultFactory(),
	}
}

func (r *ParseErrorRecoverer) RecoverableParse(input, parseType string) (interface{}, bool, interfaces.DomainErrorInterface) {
	switch parseType {
	case "int":
		// Try to extract numbers from string
		cleaned := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, input)

		if cleaned != "" {
			if value, err := strconv.Atoi(cleaned); err == nil {
				recoverErr := r.factory.Builder().
					WithCode("INT_PARSE_RECOVERED").
					WithMessage("Integer parsed with recovery").
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityLow)).
					WithDetail("original_input", input).
					WithDetail("cleaned_input", cleaned).
					WithDetail("recovered_value", value).
					WithTag("parsing").
					WithTag("recovery").
					Build()
				return value, true, recoverErr
			}
		}

	case "float":
		// Try to extract float from string
		cleaned := strings.Map(func(r rune) rune {
			if (r >= '0' && r <= '9') || r == '.' {
				return r
			}
			return -1
		}, input)

		if cleaned != "" {
			if value, err := strconv.ParseFloat(cleaned, 64); err == nil {
				recoverErr := r.factory.Builder().
					WithCode("FLOAT_PARSE_RECOVERED").
					WithMessage("Float parsed with recovery").
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityLow)).
					WithDetail("original_input", input).
					WithDetail("cleaned_input", cleaned).
					WithDetail("recovered_value", value).
					WithTag("parsing").
					WithTag("recovery").
					Build()
				return value, true, recoverErr
			}
		}

	case "bool":
		// Try common boolean representations
		lower := strings.ToLower(strings.TrimSpace(input))
		switch lower {
		case "true", "1", "yes", "on", "enabled":
			recoverErr := r.factory.Builder().
				WithCode("BOOL_PARSE_RECOVERED").
				WithMessage("Boolean parsed with recovery").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityLow)).
				WithDetail("original_input", input).
				WithDetail("normalized_input", lower).
				WithDetail("recovered_value", true).
				WithTag("parsing").
				WithTag("recovery").
				Build()
			return true, true, recoverErr
		case "false", "0", "no", "off", "disabled":
			recoverErr := r.factory.Builder().
				WithCode("BOOL_PARSE_RECOVERED").
				WithMessage("Boolean parsed with recovery").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityLow)).
				WithDetail("original_input", input).
				WithDetail("normalized_input", lower).
				WithDetail("recovered_value", false).
				WithTag("parsing").
				WithTag("recovery").
				Build()
			return false, true, recoverErr
		}

	case "date":
		// Try different date formats
		formats := []string{
			"2006-01-02",
			"2006/01/02",
			"02/01/2006",
			"01-02-2006",
			"Jan 2, 2006",
		}

		for _, format := range formats {
			if date, err := time.Parse(format, input); err == nil {
				recoverErr := r.factory.Builder().
					WithCode("DATE_PARSE_RECOVERED").
					WithMessage("Date parsed with recovery").
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityLow)).
					WithDetail("original_input", input).
					WithDetail("format_used", format).
					WithDetail("recovered_value", date.Format("2006-01-02")).
					WithTag("parsing").
					WithTag("recovery").
					Build()
				return date, true, recoverErr
			}
		}
	}

	// Could not recover
	suggestions := r.getSuggestions(parseType)
	err := r.factory.Builder().
		WithCode("PARSE_NOT_RECOVERABLE").
		WithMessage(fmt.Sprintf("Could not parse or recover %s value", parseType)).
		WithType(string(types.ErrorTypeValidation)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("input", input).
		WithDetail("parse_type", parseType).
		WithDetail("suggestions", suggestions).
		WithTag("parsing").
		WithTag("not_recoverable").
		Build()

	return nil, false, err
}

func (r *ParseErrorRecoverer) getSuggestions(parseType string) []string {
	switch parseType {
	case "int":
		return []string{"123", "456", "789"}
	case "float":
		return []string{"123.45", "67.89", "0.12"}
	case "bool":
		return []string{"true", "false", "1", "0"}
	case "date":
		return []string{"2024-01-15", "2024/01/15", "15/01/2024"}
	default:
		return []string{}
	}
}

// Parse Error Aggregator
type ParseErrorAggregator struct {
	errors  []interfaces.DomainErrorInterface
	factory interfaces.ErrorFactory
}

func NewParseErrorAggregator() *ParseErrorAggregator {
	return &ParseErrorAggregator{
		errors:  make([]interfaces.DomainErrorInterface, 0),
		factory: factory.GetDefaultFactory(),
	}
}

func (a *ParseErrorAggregator) ParseField(field, value, expectedType string) interfaces.DomainErrorInterface {
	switch expectedType {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return a.factory.Builder().
				WithCode("FIELD_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Field '%s' could not be parsed as integer", field)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", field).
				WithDetail("value", value).
				WithDetail("expected_type", expectedType).
				WithDetail("parse_error", err.Error()).
				WithTag("parsing").
				WithTag("field").
				Build()
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return a.factory.Builder().
				WithCode("FIELD_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Field '%s' could not be parsed as float", field)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", field).
				WithDetail("value", value).
				WithDetail("expected_type", expectedType).
				WithDetail("parse_error", err.Error()).
				WithTag("parsing").
				WithTag("field").
				Build()
		}
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return a.factory.Builder().
				WithCode("FIELD_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Field '%s' could not be parsed as boolean", field)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", field).
				WithDetail("value", value).
				WithDetail("expected_type", expectedType).
				WithDetail("parse_error", err.Error()).
				WithTag("parsing").
				WithTag("field").
				Build()
		}
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return a.factory.Builder().
				WithCode("FIELD_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Field '%s' could not be parsed as date", field)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", field).
				WithDetail("value", value).
				WithDetail("expected_type", expectedType).
				WithDetail("expected_format", "YYYY-MM-DD").
				WithDetail("parse_error", err.Error()).
				WithTag("parsing").
				WithTag("field").
				Build()
		}
	case "email":
		if !strings.Contains(value, "@") {
			return a.factory.Builder().
				WithCode("FIELD_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Field '%s' is not a valid email", field)).
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", field).
				WithDetail("value", value).
				WithDetail("expected_type", expectedType).
				WithDetail("validation_rule", "must contain @").
				WithTag("parsing").
				WithTag("field").
				Build()
		}
	}

	return nil
}

func (a *ParseErrorAggregator) AddError(err interfaces.DomainErrorInterface) {
	a.errors = append(a.errors, err)
}

func (a *ParseErrorAggregator) HasErrors() bool {
	return len(a.errors) > 0
}

func (a *ParseErrorAggregator) GetErrors() []interfaces.DomainErrorInterface {
	return a.errors
}

func (a *ParseErrorAggregator) BuildAggregatedError() interfaces.ValidationErrorInterface {
	fieldErrors := make(map[string][]string)

	for _, err := range a.errors {
		if field, ok := err.Details()["field"].(string); ok {
			fieldErrors[field] = append(fieldErrors[field], err.Error())
		}
	}

	return a.factory.NewValidation("Multiple field parsing errors", fieldErrors)
}

// Contextual Parser
type ContextualParser struct {
	factory interfaces.ErrorFactory
}

func NewContextualParser() *ContextualParser {
	return &ContextualParser{
		factory: factory.GetDefaultFactory(),
	}
}

func (p *ContextualParser) ParseConfiguration(configData string) (map[string]interface{}, []interfaces.DomainErrorInterface) {
	config := make(map[string]interface{})
	var errors []interfaces.DomainErrorInterface

	// Simulate parsing YAML-like configuration
	lines := strings.Split(configData, "\n")
	currentSection := ""

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasSuffix(line, ":") && !strings.Contains(line, "\"") {
			// Section header
			currentSection = strings.TrimSuffix(line, ":")
			config[currentSection] = make(map[string]interface{})
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

				// Try to parse the value
				parsedValue, err := p.parseConfigValue(key, value, currentSection, lineNum+1)
				if err != nil {
					errors = append(errors, err)
				} else {
					if currentSection != "" {
						if sectionMap, ok := config[currentSection].(map[string]interface{}); ok {
							sectionMap[key] = parsedValue
						}
					} else {
						config[key] = parsedValue
					}
				}
			}
		}
	}

	return config, errors
}

func (p *ContextualParser) parseConfigValue(key, value, section string, line int) (interface{}, interfaces.DomainErrorInterface) {
	// Define expected types for known configuration keys
	expectedTypes := map[string]string{
		"port":               "int",
		"max_connections":    "int",
		"retry_attempts":     "int",
		"rate_limit":         "int",
		"timeout":            "duration",
		"connection_timeout": "duration",
		"debug_mode":         "bool",
		"cache_enabled":      "bool",
		"host":               "string",
	}

	expectedType, exists := expectedTypes[key]
	if !exists {
		expectedType = "string" // Default to string
	}

	context := fmt.Sprintf("line %d in section '%s'", line, section)

	switch expectedType {
	case "int":
		if parsed, err := strconv.Atoi(value); err != nil {
			return nil, p.factory.Builder().
				WithCode("CONFIG_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Configuration key '%s' has invalid integer value", key)).
				WithType(string(types.ErrorTypeConfiguration)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("section", section).
				WithDetail("key", key).
				WithDetail("value", value).
				WithDetail("line", line).
				WithDetail("expected_type", "integer").
				WithDetail("context", context).
				WithDetail("parse_error", err.Error()).
				WithTag("configuration").
				WithTag("parsing").
				Build()
		} else {
			return parsed, nil
		}

	case "bool":
		if parsed, err := strconv.ParseBool(value); err != nil {
			return nil, p.factory.Builder().
				WithCode("CONFIG_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Configuration key '%s' has invalid boolean value", key)).
				WithType(string(types.ErrorTypeConfiguration)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("section", section).
				WithDetail("key", key).
				WithDetail("value", value).
				WithDetail("line", line).
				WithDetail("expected_type", "boolean").
				WithDetail("context", context).
				WithDetail("valid_values", []string{"true", "false"}).
				WithTag("configuration").
				WithTag("parsing").
				Build()
		} else {
			return parsed, nil
		}

	case "duration":
		if parsed, err := time.ParseDuration(value); err != nil {
			return nil, p.factory.Builder().
				WithCode("CONFIG_PARSE_ERROR").
				WithMessage(fmt.Sprintf("Configuration key '%s' has invalid duration value", key)).
				WithType(string(types.ErrorTypeConfiguration)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("section", section).
				WithDetail("key", key).
				WithDetail("value", value).
				WithDetail("line", line).
				WithDetail("expected_type", "duration").
				WithDetail("context", context).
				WithDetail("examples", []string{"30s", "5m", "1h"}).
				WithTag("configuration").
				WithTag("parsing").
				Build()
		} else {
			return parsed, nil
		}

	default:
		return value, nil
	}
}
