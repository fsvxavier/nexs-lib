package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"golang.org/x/text/language"
)

// JSONProvider implements Provider interface for JSON files
type JSONProvider struct {
	*BaseProvider
}

// NewJSONProvider creates a new JSON provider
func NewJSONProvider() *JSONProvider {
	return &JSONProvider{
		BaseProvider: NewBaseProvider(),
	}
}

// LoadTranslations implements Provider.LoadTranslations
func (p *JSONProvider) LoadTranslations(path string, format string) error {
	if format != "json" {
		return fmt.Errorf("unsupported format %q for JSON provider", format)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read translation file %q: %w", path, err)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse JSON translation file %q: %w", path, err)
	}

	// Extract language from filename (assuming format translations.LANG.json)
	langCode := extractLanguageFromPath(path)
	lang, err := language.Parse(langCode)
	if err != nil {
		return fmt.Errorf("invalid language code %q in file %q: %w", langCode, path, err)
	}

	// Add all translations
	for key, value := range translations {
		p.addTranslation(lang, key, value)
	}

	return nil
}

// Translate implements Provider.Translate
func (p *JSONProvider) Translate(key string, templateData map[string]interface{}) (string, error) {
	val, ok := p.getTranslation(key)
	if !ok {
		return key, nil // Return key as fallback
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("translation for key %q is not a string", key)
	}

	if len(templateData) == 0 {
		return str, nil
	}

	// Process template
	tmpl, err := template.New("translation").Parse(str)
	if err != nil {
		return "", fmt.Errorf("failed to parse template for key %q: %w", key, err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template for key %q: %w", key, err)
	}

	return result.String(), nil
}

// TranslatePlural implements Provider.TranslatePlural
func (p *JSONProvider) TranslatePlural(key string, count interface{}, templateData map[string]interface{}) (string, error) {
	// Get the translation object that contains plural forms
	val, ok := p.getTranslation(key)
	if !ok {
		return key, nil // Return key as fallback
	}

	// Check if the value is a map containing plural forms
	pluralForms, ok := val.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("translation for key %q is not a plural form object", key)
	}

	// Determine plural form
	pluralForm := getPluralForm(count)

	// Get the correct plural form
	formVal, ok := pluralForms[pluralForm]
	if !ok {
		// Try "other" as fallback
		formVal, ok = pluralForms["other"]
		if !ok {
			return "", fmt.Errorf("no plural form found for key %q", key)
		}
	}

	// Convert to string
	str, ok := formVal.(string)
	if !ok {
		return "", fmt.Errorf("plural form for key %q is not a string", key)
	}

	if templateData == nil {
		templateData = make(map[string]interface{})
	}
	templateData["Count"] = count

	// Process template
	tmpl, err := template.New("translation").Parse(str)
	if err != nil {
		return "", fmt.Errorf("failed to parse plural template for key %q: %w", key, err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, templateData); err != nil {
		return "", fmt.Errorf("failed to execute plural template for key %q: %w", key, err)
	}

	return result.String(), nil
}

// getPluralForm returns the plural form based on count
func getPluralForm(count interface{}) string {
	switch v := count.(type) {
	case int:
		if v == 1 {
			return "one"
		}
	case int64:
		if v == 1 {
			return "one"
		}
	case float64:
		if v == 1.0 {
			return "one"
		}
	}
	return "other"
}
