package schema

import (
	"context"
	"time"
)

// ruleBuilder implements the RuleBuilder interface
type ruleBuilder struct {
	rules    []Rule
	required bool
}

// NewRuleBuilder creates a new rule builder
func NewRuleBuilder() RuleBuilder {
	return &ruleBuilder{
		rules: make([]Rule, 0),
	}
}

// Required marks the field as required
func (rb *ruleBuilder) Required() RuleBuilder {
	rb.required = true
	rb.rules = append(rb.rules, NewRequiredRule())
	return rb
}

// Optional marks the field as optional (default behavior)
func (rb *ruleBuilder) Optional() RuleBuilder {
	rb.required = false
	return rb
}

// String adds string validation rules
func (rb *ruleBuilder) String() StringRuleBuilder {
	return &stringRuleBuilder{
		parent: rb,
	}
}

// Number adds numeric validation rules
func (rb *ruleBuilder) Number() NumberRuleBuilder {
	return &numberRuleBuilder{
		parent: rb,
	}
}

// DateTime adds date/time validation rules
func (rb *ruleBuilder) DateTime() DateTimeRuleBuilder {
	return &dateTimeRuleBuilder{
		parent: rb,
	}
}

// Custom adds a custom validation rule
func (rb *ruleBuilder) Custom(rule Rule) RuleBuilder {
	rb.rules = append(rb.rules, rule)
	return rb
}

// Build creates the final rule
func (rb *ruleBuilder) Build() Rule {
	if len(rb.rules) == 0 {
		return NewCustomRule("empty", "no validation rules", func(ctx context.Context, value interface{}) error {
			return nil // Always pass if no rules
		})
	}

	if len(rb.rules) == 1 {
		return rb.rules[0]
	}

	// Combine multiple rules into a composite rule
	return NewCompositeRule(rb.rules...)
}

// stringRuleBuilder implements StringRuleBuilder
type stringRuleBuilder struct {
	parent *ruleBuilder
	rules  []Rule
}

// MinLength sets minimum string length
func (srb *stringRuleBuilder) MinLength(min int) StringRuleBuilder {
	srb.rules = append(srb.rules, NewMinLengthRule(min))
	return srb
}

// MaxLength sets maximum string length
func (srb *stringRuleBuilder) MaxLength(max int) StringRuleBuilder {
	srb.rules = append(srb.rules, NewMaxLengthRule(max))
	return srb
}

// Pattern validates against a regex pattern
func (srb *stringRuleBuilder) Pattern(pattern string) StringRuleBuilder {
	srb.rules = append(srb.rules, NewPatternRule(pattern))
	return srb
}

// Email validates email format
func (srb *stringRuleBuilder) Email() StringRuleBuilder {
	srb.rules = append(srb.rules, NewEmailRule())
	return srb
}

// URL validates URL format
func (srb *stringRuleBuilder) URL() StringRuleBuilder {
	srb.rules = append(srb.rules, NewURLRule())
	return srb
}

// UUID validates UUID format
func (srb *stringRuleBuilder) UUID() StringRuleBuilder {
	srb.rules = append(srb.rules, NewUUIDRule())
	return srb
}

// Custom adds custom string validation
func (srb *stringRuleBuilder) Custom(validator func(string) error) StringRuleBuilder {
	rule := NewCustomRule("custom_string", "custom validation failed", func(ctx context.Context, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("", "value must be a string", "type_error", value)
		}
		return validator(str)
	})
	srb.rules = append(srb.rules, rule)
	return srb
}

// Build creates the final string rule
func (srb *stringRuleBuilder) Build() Rule {
	// Add string rules to parent
	srb.parent.rules = append(srb.parent.rules, srb.rules...)
	return srb.parent.Build()
}

// numberRuleBuilder implements NumberRuleBuilder
type numberRuleBuilder struct {
	parent *ruleBuilder
	rules  []Rule
}

// Min sets minimum value
func (nrb *numberRuleBuilder) Min(min float64) NumberRuleBuilder {
	nrb.rules = append(nrb.rules, NewMinValueRule(min))
	return nrb
}

// Max sets maximum value
func (nrb *numberRuleBuilder) Max(max float64) NumberRuleBuilder {
	nrb.rules = append(nrb.rules, NewMaxValueRule(max))
	return nrb
}

// Range sets min and max values
func (nrb *numberRuleBuilder) Range(min, max float64) NumberRuleBuilder {
	nrb.rules = append(nrb.rules, NewMinValueRule(min))
	nrb.rules = append(nrb.rules, NewMaxValueRule(max))
	return nrb
}

// Positive ensures the number is positive
func (nrb *numberRuleBuilder) Positive() NumberRuleBuilder {
	nrb.rules = append(nrb.rules, NewMinValueRule(0.000001)) // Greater than 0
	return nrb
}

// NonNegative ensures the number is >= 0
func (nrb *numberRuleBuilder) NonNegative() NumberRuleBuilder {
	nrb.rules = append(nrb.rules, NewMinValueRule(0))
	return nrb
}

// Integer ensures the number is an integer
func (nrb *numberRuleBuilder) Integer() NumberRuleBuilder {
	rule := NewCustomRule("integer", "value must be an integer", func(ctx context.Context, value interface{}) error {
		num, err := toFloat64(value)
		if err != nil {
			return err
		}
		if num != float64(int64(num)) {
			return NewValidationError("", "value must be an integer", "integer_error", value)
		}
		return nil
	})
	nrb.rules = append(nrb.rules, rule)
	return nrb
}

// Decimal validates decimal with specific precision
func (nrb *numberRuleBuilder) Decimal(precision int) NumberRuleBuilder {
	rule := NewCustomRule("decimal", "invalid decimal precision", func(ctx context.Context, value interface{}) error {
		// This would integrate with the existing decimal package
		// For now, just validate it's a valid number
		_, err := toFloat64(value)
		return err
	})
	nrb.rules = append(nrb.rules, rule)
	return nrb
}

// Custom adds custom numeric validation
func (nrb *numberRuleBuilder) Custom(validator func(float64) error) NumberRuleBuilder {
	rule := NewCustomRule("custom_number", "custom validation failed", func(ctx context.Context, value interface{}) error {
		num, err := toFloat64(value)
		if err != nil {
			return err
		}
		return validator(num)
	})
	nrb.rules = append(nrb.rules, rule)
	return nrb
}

// Build creates the final number rule
func (nrb *numberRuleBuilder) Build() Rule {
	// Add number rules to parent
	nrb.parent.rules = append(nrb.parent.rules, nrb.rules...)
	return nrb.parent.Build()
}

// dateTimeRuleBuilder implements DateTimeRuleBuilder
type dateTimeRuleBuilder struct {
	parent *ruleBuilder
	rules  []Rule
}

// Format validates against specific date format
func (dtrb *dateTimeRuleBuilder) Format(format string) DateTimeRuleBuilder {
	dtrb.rules = append(dtrb.rules, NewDateTimeFormatRule(format))
	return dtrb
}

// RFC3339 validates RFC3339 format
func (dtrb *dateTimeRuleBuilder) RFC3339() DateTimeRuleBuilder {
	dtrb.rules = append(dtrb.rules, NewDateTimeFormatRule(time.RFC3339))
	return dtrb
}

// ISO8601 validates ISO8601 format
func (dtrb *dateTimeRuleBuilder) ISO8601() DateTimeRuleBuilder {
	dtrb.rules = append(dtrb.rules, NewDateTimeFormatRule("2006-01-02T15:04:05.999Z07:00"))
	return dtrb
}

// Before ensures date is before specified date
func (dtrb *dateTimeRuleBuilder) Before(date string) DateTimeRuleBuilder {
	rule := NewCustomRule("before", "date must be before "+date, func(ctx context.Context, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("", "value must be a string", "type_error", value)
		}

		valueTime, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}

		beforeTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return err
		}

		if !valueTime.Before(beforeTime) {
			return NewValidationError("", "date must be before "+date, "before_error", value)
		}

		return nil
	})
	dtrb.rules = append(dtrb.rules, rule)
	return dtrb
}

// After ensures date is after specified date
func (dtrb *dateTimeRuleBuilder) After(date string) DateTimeRuleBuilder {
	rule := NewCustomRule("after", "date must be after "+date, func(ctx context.Context, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("", "value must be a string", "type_error", value)
		}

		valueTime, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}

		afterTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return err
		}

		if !valueTime.After(afterTime) {
			return NewValidationError("", "date must be after "+date, "after_error", value)
		}

		return nil
	})
	dtrb.rules = append(dtrb.rules, rule)
	return dtrb
}

// Range ensures date is within specified range
func (dtrb *dateTimeRuleBuilder) Range(start, end string) DateTimeRuleBuilder {
	dtrb.After(start)
	dtrb.Before(end)
	return dtrb
}

// Custom adds custom date/time validation
func (dtrb *dateTimeRuleBuilder) Custom(validator func(string) error) DateTimeRuleBuilder {
	rule := NewCustomRule("custom_datetime", "custom validation failed", func(ctx context.Context, value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("", "value must be a string", "type_error", value)
		}
		return validator(str)
	})
	dtrb.rules = append(dtrb.rules, rule)
	return dtrb
}

// Build creates the final date/time rule
func (dtrb *dateTimeRuleBuilder) Build() Rule {
	// Add datetime rules to parent
	dtrb.parent.rules = append(dtrb.parent.rules, dtrb.rules...)
	return dtrb.parent.Build()
}

// CompositeRule combines multiple rules
type CompositeRule struct {
	*baseRule
	rules []Rule
}

// NewCompositeRule creates a new composite rule
func NewCompositeRule(rules ...Rule) *CompositeRule {
	return &CompositeRule{
		baseRule: &baseRule{
			name:    "composite",
			message: "composite validation failed",
		},
		rules: rules,
	}
}

// Validate executes all contained rules
func (cr *CompositeRule) Validate(ctx context.Context, value interface{}) error {
	for _, rule := range cr.rules {
		if err := rule.Validate(ctx, value); err != nil {
			return err
		}
	}
	return nil
}
