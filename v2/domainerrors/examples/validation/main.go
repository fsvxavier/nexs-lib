package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// User represents a user model for validation
type User struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Age      int       `json:"age"`
	Phone    string    `json:"phone"`
	Country  string    `json:"country"`
	Password string    `json:"password"`
	Profile  *Profile  `json:"profile,omitempty"`
	Created  time.Time `json:"created"`
}

// Profile represents user profile
type Profile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
	Website   string `json:"website"`
}

func main() {
	fmt.Println("✅ Domain Errors v2 - Validation Examples")
	fmt.Println("==========================================")

	basicValidationExample()
	structuredValidationExample()
	businessRuleValidationExample()
	nestedValidationExample()
	customValidatorExample()
	validationChainingExample()
	validationWithContextExample()
}

// basicValidationExample demonstrates simple field validation
func basicValidationExample() {
	fmt.Println("\n📝 Basic Validation Example:")

	// Create validation error with multiple fields
	validationErr := domainerrors.NewValidationError("User registration failed", map[string][]string{
		"email":    {"invalid format", "domain not allowed"},
		"age":      {"must be 18 or older"},
		"password": {"too short", "missing special characters"},
	})

	// Add more validation errors dynamically
	validationErr.AddField("username", "already taken")
	validationErr.AddField("username", "contains invalid characters")
	validationErr.AddField("phone", "invalid country code")

	fmt.Printf("  Error: %s\n", validationErr.Error())
	fmt.Printf("  All Fields: %+v\n", validationErr.Fields())

	// Check specific field
	if validationErr.HasField("email") {
		fmt.Printf("  Email Errors: %v\n", validationErr.FieldErrors("email"))
	}

	// Count total errors
	totalErrors := 0
	for _, errors := range validationErr.Fields() {
		totalErrors += len(errors)
	}
	fmt.Printf("  Total Validation Errors: %d\n", totalErrors)
}

// structuredValidationExample shows advanced validation with categories
func structuredValidationExample() {
	fmt.Println("\n🏗️ Structured Validation Example:")

	user := User{
		Email:    "invalid-email",
		Username: "ab",
		Age:      16,
		Phone:    "+1234567890123456", // too long
		Country:  "XX",                // invalid country
		Password: "weak",
	}

	validator := NewUserValidator()
	err := validator.Validate(user)

	if validationErr, ok := err.(interfaces.ValidationErrorInterface); ok {
		fmt.Printf("  Validation Result: %s\n", validationErr.Error())
		fmt.Printf("  Fields with errors: %d\n", len(validationErr.Fields()))

		for field, errors := range validationErr.Fields() {
			fmt.Printf("    %s: %v\n", field, errors)
		}
	}
}

// businessRuleValidationExample demonstrates business rule validation
func businessRuleValidationExample() {
	fmt.Println("\n💼 Business Rule Validation Example:")

	user := User{
		Email:    "test@competitor.com",
		Username: "admin", // reserved username
		Age:      25,
		Country:  "restricted", // restricted country
		Password: "ValidPass123!",
	}

	businessErr := validateBusinessRules(user)
	if businessErr != nil {
		fmt.Printf("  Business Rule Error: %s\n", businessErr.Error())
		fmt.Printf("  Error Type: %s\n", businessErr.Type())
		fmt.Printf("  Details: %+v\n", businessErr.Details())
		fmt.Printf("  Tags: %v\n", businessErr.Tags())
	}
}

// nestedValidationExample shows validation of nested structures
func nestedValidationExample() {
	fmt.Println("\n🔗 Nested Validation Example:")

	user := User{
		Email:    "valid@example.com",
		Username: "validuser",
		Age:      25,
		Password: "ValidPass123!",
		Profile: &Profile{
			FirstName: "", // required
			LastName:  "User",
			Bio:       strings.Repeat("a", 1001), // too long
			Website:   "invalid-url",
		},
	}

	nestedErr := validateNestedStructure(user)
	if nestedErr != nil {
		fmt.Printf("  Nested Validation Error: %s\n", nestedErr.Error())

		if validationErr, ok := nestedErr.(interfaces.ValidationErrorInterface); ok {
			for field, errors := range validationErr.Fields() {
				fmt.Printf("    %s: %v\n", field, errors)
			}
		}
	}
}

// customValidatorExample demonstrates custom validation functions
func customValidatorExample() {
	fmt.Println("\n🎯 Custom Validator Example:")

	validators := []struct {
		name      string
		validator func(User) error
		user      User
	}{
		{
			name:      "Email Domain Validator",
			validator: validateEmailDomain,
			user:      User{Email: "test@blacklisted.com"},
		},
		{
			name:      "Password Strength Validator",
			validator: validatePasswordStrength,
			user:      User{Password: "weak"},
		},
		{
			name:      "Age Range Validator",
			validator: validateAgeRange,
			user:      User{Age: 150},
		},
	}

	for _, v := range validators {
		fmt.Printf("\n  %s:\n", v.name)
		if err := v.validator(v.user); err != nil {
			if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
				fmt.Printf("    Error: %s\n", domainErr.Error())
				fmt.Printf("    Code: %s\n", domainErr.Code())
				fmt.Printf("    Severity: %v\n", domainErr.Severity())
			}
		} else {
			fmt.Printf("    ✅ Validation passed\n")
		}
	}
}

// validationChainingExample shows chaining multiple validations
func validationChainingExample() {
	fmt.Println("\n⛓️ Validation Chaining Example:")

	user := User{
		Email:    "test@example.com",
		Username: "admin", // reserved
		Age:      16,      // too young
		Password: "weak",  // weak password
	}

	// Chain multiple validators
	validators := []func(User) error{
		validateEmailDomain,
		validatePasswordStrength,
		validateAgeRange,
		validateReservedUsername,
	}

	var allErrors []error
	for _, validator := range validators {
		if err := validator(user); err != nil {
			allErrors = append(allErrors, err)
		}
	}

	if len(allErrors) > 0 {
		// Combine all validation errors
		combinedErr := combineValidationErrors(allErrors)
		fmt.Printf("  Combined Validation Error: %s\n", combinedErr.Error())

		if validationErr, ok := combinedErr.(interfaces.ValidationErrorInterface); ok {
			fmt.Printf("  Total fields with errors: %d\n", len(validationErr.Fields()))
			for field, errors := range validationErr.Fields() {
				fmt.Printf("    %s: %v\n", field, errors)
			}
		}
	}
}

// validationWithContextExample shows context-aware validation
func validationWithContextExample() {
	fmt.Println("\n🌐 Context-Aware Validation Example:")

	contexts := []struct {
		name    string
		context map[string]interface{}
		user    User
	}{
		{
			name: "Admin Registration",
			context: map[string]interface{}{
				"role":        "admin",
				"permissions": []string{"read", "write", "admin"},
			},
			user: User{Username: "regularuser", Email: "user@example.com"},
		},
		{
			name: "Premium Account",
			context: map[string]interface{}{
				"account_type": "premium",
				"features":     []string{"advanced_analytics", "priority_support"},
			},
			user: User{Email: "premium@example.com", Age: 17}, // age restriction for premium
		},
	}

	for _, ctx := range contexts {
		fmt.Printf("\n  %s:\n", ctx.name)
		err := validateWithContext(ctx.user, ctx.context)
		if err != nil {
			if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
				fmt.Printf("    Error: %s\n", domainErr.Error())
				fmt.Printf("    Context Details: %+v\n", domainErr.Details())
			}
		} else {
			fmt.Printf("    ✅ Context validation passed\n")
		}
	}
}

// UserValidator represents a comprehensive user validator
type UserValidator struct {
	emailRegex    *regexp.Regexp
	usernameRegex *regexp.Regexp
	phoneRegex    *regexp.Regexp
}

// NewUserValidator creates a new user validator
func NewUserValidator() *UserValidator {
	return &UserValidator{
		emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`),
		phoneRegex:    regexp.MustCompile(`^\+[1-9]\d{1,14}$`),
	}
}

// Validate performs comprehensive user validation
func (v *UserValidator) Validate(user User) error {
	fields := make(map[string][]string)

	// Email validation
	if user.Email == "" {
		fields["email"] = append(fields["email"], "required")
	} else if !v.emailRegex.MatchString(user.Email) {
		fields["email"] = append(fields["email"], "invalid format")
	}

	// Username validation
	if user.Username == "" {
		fields["username"] = append(fields["username"], "required")
	} else if !v.usernameRegex.MatchString(user.Username) {
		fields["username"] = append(fields["username"], "invalid format (3-20 alphanumeric characters)")
	}

	// Age validation
	if user.Age <= 0 {
		fields["age"] = append(fields["age"], "required")
	} else if user.Age < 13 {
		fields["age"] = append(fields["age"], "must be at least 13 years old")
	} else if user.Age > 120 {
		fields["age"] = append(fields["age"], "invalid age")
	}

	// Phone validation
	if user.Phone != "" && !v.phoneRegex.MatchString(user.Phone) {
		fields["phone"] = append(fields["phone"], "invalid international format")
	}

	// Password validation
	if user.Password == "" {
		fields["password"] = append(fields["password"], "required")
	} else {
		if len(user.Password) < 8 {
			fields["password"] = append(fields["password"], "minimum 8 characters")
		}
		if !regexp.MustCompile(`[A-Z]`).MatchString(user.Password) {
			fields["password"] = append(fields["password"], "must contain uppercase letter")
		}
		if !regexp.MustCompile(`[a-z]`).MatchString(user.Password) {
			fields["password"] = append(fields["password"], "must contain lowercase letter")
		}
		if !regexp.MustCompile(`\d`).MatchString(user.Password) {
			fields["password"] = append(fields["password"], "must contain number")
		}
		if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(user.Password) {
			fields["password"] = append(fields["password"], "must contain special character")
		}
	}

	if len(fields) > 0 {
		return domainerrors.NewValidationError("User validation failed", fields)
	}

	return nil
}

// Business rule validation functions
func validateBusinessRules(user User) interfaces.DomainErrorInterface {
	err := domainerrors.NewBuilder().
		WithCode("BIZ001").
		WithMessage("Business rule validation failed").
		WithType(string(types.ErrorTypeBusinessRule)).
		WithSeverity(interfaces.Severity(types.SeverityMedium))

	violations := []string{}

	// Check for competitor email domains
	if strings.Contains(user.Email, "@competitor.com") {
		violations = append(violations, "competitor email domains not allowed")
		err = err.WithDetail("email_domain", "competitor.com")
	}

	// Check for reserved usernames
	reservedUsernames := []string{"admin", "root", "system", "api"}
	for _, reserved := range reservedUsernames {
		if user.Username == reserved {
			violations = append(violations, "username is reserved")
			err = err.WithDetail("reserved_username", user.Username)
			break
		}
	}

	// Check for restricted countries
	restrictedCountries := []string{"restricted", "blocked"}
	for _, restricted := range restrictedCountries {
		if user.Country == restricted {
			violations = append(violations, "country not supported")
			err = err.WithDetail("restricted_country", user.Country)
			break
		}
	}

	if len(violations) > 0 {
		err = err.WithDetail("violations", violations).
			WithTag("business_rules").
			WithTag("validation")
		return err.Build()
	}

	return nil
}

// Custom validation functions
func validateEmailDomain(user User) error {
	blacklistedDomains := []string{"blacklisted.com", "spam.com", "fake.com"}
	for _, domain := range blacklistedDomains {
		if strings.Contains(user.Email, "@"+domain) {
			return domainerrors.NewBuilder().
				WithCode("VAL001").
				WithMessage("Email domain not allowed").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("domain", domain).
				WithDetail("email", user.Email).
				WithTag("email_validation").
				Build()
		}
	}
	return nil
}

func validatePasswordStrength(user User) error {
	score := calculatePasswordStrength(user.Password)
	if score < 3 {
		return domainerrors.NewBuilder().
			WithCode("VAL002").
			WithMessage("Password strength insufficient").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("strength_score", score).
			WithDetail("minimum_score", 3).
			WithTag("password_validation").
			Build()
	}
	return nil
}

func validateAgeRange(user User) error {
	if user.Age < 13 || user.Age > 120 {
		return domainerrors.NewBuilder().
			WithCode("VAL003").
			WithMessage("Age out of valid range").
			WithType(string(types.ErrorTypeValidation)).
			WithSeverity(interfaces.Severity(types.SeverityLow)).
			WithDetail("age", user.Age).
			WithDetail("min_age", 13).
			WithDetail("max_age", 120).
			WithTag("age_validation").
			Build()
	}
	return nil
}

func validateReservedUsername(user User) error {
	reserved := []string{"admin", "root", "system", "api", "www", "mail"}
	for _, r := range reserved {
		if user.Username == r {
			return domainerrors.NewBuilder().
				WithCode("VAL004").
				WithMessage("Username is reserved").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("username", user.Username).
				WithDetail("reserved_usernames", reserved).
				WithTag("username_validation").
				Build()
		}
	}
	return nil
}

// Helper functions
func calculatePasswordStrength(password string) int {
	score := 0
	if len(password) >= 8 {
		score++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`\d`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		score++
	}
	return score
}

func validateNestedStructure(user User) error {
	fields := make(map[string][]string)

	if user.Profile != nil {
		if user.Profile.FirstName == "" {
			fields["profile.first_name"] = append(fields["profile.first_name"], "required")
		}
		if len(user.Profile.Bio) > 1000 {
			fields["profile.bio"] = append(fields["profile.bio"], "maximum 1000 characters")
		}
		if user.Profile.Website != "" {
			if !regexp.MustCompile(`^https?://`).MatchString(user.Profile.Website) {
				fields["profile.website"] = append(fields["profile.website"], "must be a valid URL")
			}
		}
	}

	if len(fields) > 0 {
		return domainerrors.NewValidationError("Nested validation failed", fields)
	}
	return nil
}

func combineValidationErrors(errors []error) error {
	combinedFields := make(map[string][]string)

	for _, err := range errors {
		if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
			// Extract field information from error details or message
			field := extractFieldFromError(domainErr)
			message := domainErr.Message()
			combinedFields[field] = append(combinedFields[field], message)
		}
	}

	return domainerrors.NewValidationError("Multiple validation failures", combinedFields)
}

func extractFieldFromError(err interfaces.DomainErrorInterface) string {
	// Extract field name from error code or details
	if field, ok := err.Details()["field"].(string); ok {
		return field
	}

	// Map error codes to fields
	codeToField := map[string]string{
		"VAL001": "email",
		"VAL002": "password",
		"VAL003": "age",
		"VAL004": "username",
	}

	if field, ok := codeToField[err.Code()]; ok {
		return field
	}

	return "unknown"
}

func validateWithContext(user User, context map[string]interface{}) error {
	if role, ok := context["role"].(string); ok && role == "admin" {
		// Admin users need strong usernames
		if len(user.Username) < 8 {
			return domainerrors.NewBuilder().
				WithCode("CTX001").
				WithMessage("Admin username must be at least 8 characters").
				WithType(string(types.ErrorTypeValidation)).
				WithDetail("role", role).
				WithDetail("username_length", len(user.Username)).
				Build()
		}
	}

	if accountType, ok := context["account_type"].(string); ok && accountType == "premium" {
		// Premium accounts require users to be 18+
		if user.Age < 18 {
			return domainerrors.NewBuilder().
				WithCode("CTX002").
				WithMessage("Premium accounts require age 18+").
				WithType(string(types.ErrorTypeBusinessRule)).
				WithDetail("account_type", accountType).
				WithDetail("user_age", user.Age).
				WithDetail("required_age", 18).
				Build()
		}
	}

	return nil
}
