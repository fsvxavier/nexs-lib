package providers

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// YAMLProvider implements Provider interface for YAML files
type YAMLProvider struct {
	*BaseProvider
}

// NewYAMLProvider creates a new YAML provider
func NewYAMLProvider() *YAMLProvider {
	return &YAMLProvider{
		BaseProvider: NewBaseProvider(),
	}
}

// LoadTranslations implements Provider.LoadTranslations
func (p *YAMLProvider) LoadTranslations(path string, format string) error {
	if format != "yaml" && format != "yml" {
		return fmt.Errorf("unsupported format %q for YAML provider", format)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read translation file %q: %w", path, err)
	}

	var translations map[string]interface{}
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse YAML translation file %q: %w", path, err)
	}

	// Extract language from filename
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
func (p *YAMLProvider) Translate(key string, templateData map[string]interface{}) (string, error) {
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
func (p *YAMLProvider) TranslatePlural(key string, count interface{}, templateData map[string]interface{}) (string, error) {
	// For YAML we support nested plural forms
	// e.g., users:
	//         one: "{{.Count}} user"
	//         other: "{{.Count}} users"

	val, ok := p.getTranslation(key)
	if !ok {
		return key, nil
	}

	pluralForms, ok := val.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("translation for key %q is not a plural form object", key)
	}

	// Get plural form based on count
	pluralForm := getPluralForm(count)
	str, ok := pluralForms[pluralForm].(string)
	if !ok {
		// Try "other" as fallback
		str, ok = pluralForms["other"].(string)
		if !ok {
			return "", fmt.Errorf("no valid plural form found for key %q", key)
		}
	}

	if templateData == nil {
		templateData = make(map[string]interface{})
	}
	templateData["Count"] = count

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
