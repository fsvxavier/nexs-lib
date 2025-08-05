package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
)

// BasicProvider é a implementação básica do Provider
type BasicProvider struct {
	languages    []language.Tag
	translations map[string]map[string]string
}

// NewBasicProvider cria uma nova instância do BasicProvider
func NewBasicProvider() *BasicProvider {
	return &BasicProvider{
		translations: make(map[string]map[string]string),
	}
}

// replaceVariables substitui as variáveis no texto traduzido
func (p *BasicProvider) replaceVariables(translation string, data map[string]interface{}) string {
	if data == nil {
		return translation
	}

	result := translation
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		if str, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, str)
		} else {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
		}
	}
	return result
}

// Translate implementa Provider.Translate
func (p *BasicProvider) Translate(key string, data map[string]interface{}) (string, error) {
	if len(p.languages) == 0 {
		return "", fmt.Errorf("no languages configured")
	}

	langKey := p.languages[0].String()
	if translations, ok := p.translations[langKey]; ok {
		if translation, ok := translations[key]; ok {
			return p.replaceVariables(translation, data), nil
		}
		return "", fmt.Errorf("translation not found for key: %s", key)
	}
	return "", fmt.Errorf("translations not found for language: %s", langKey)
}

// TranslatePlural implementa Provider.TranslatePlural
func (p *BasicProvider) TranslatePlural(key string, count interface{}, data map[string]interface{}) (string, error) {
	if len(p.languages) == 0 {
		return "", fmt.Errorf("no languages configured")
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	// Adiciona o count aos dados para substituição
	data["count"] = count

	langKey := p.languages[0].String()
	if translations, ok := p.translations[langKey]; ok {
		// Tenta encontrar a tradução específica para o número
		pluralKey := fmt.Sprintf("%s.%v", key, count)
		if translation, ok := translations[pluralKey]; ok {
			return p.replaceVariables(translation, data), nil
		}

		// Se não encontrou uma tradução específica, usa a forma plural genérica
		otherKey := fmt.Sprintf("%s.other", key)
		if translation, ok := translations[otherKey]; ok {
			return p.replaceVariables(translation, data), nil
		}

		// Como último recurso, tenta a chave base
		if translation, ok := translations[key]; ok {
			return p.replaceVariables(translation, data), nil
		}

		return "", fmt.Errorf("translation not found for key: %s", pluralKey)
	}
	return "", fmt.Errorf("translations not found for language: %s", langKey)
}

// LoadTranslations implementa Provider.LoadTranslations
func (p *BasicProvider) LoadTranslations(path string, format string) error {
	// Por enquanto, suporta apenas JSON
	if format != "json" {
		return fmt.Errorf("unsupported translation format: %s", format)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read translations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// O nome do arquivo deve ser o código do idioma (ex: en.json, pt.json)
		lang := strings.TrimSuffix(file.Name(), ".json")

		content, err := ioutil.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", file.Name(), err)
		}

		translations := make(map[string]string)
		if err := json.Unmarshal(content, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", file.Name(), err)
		}

		p.translations[lang] = translations
	}

	return nil
}

// GetLanguages implementa Provider.GetLanguages
func (p *BasicProvider) GetLanguages() []language.Tag {
	return p.languages
}

// SetLanguages implementa Provider.SetLanguages
func (p *BasicProvider) SetLanguages(languages ...string) error {
	tags := make([]language.Tag, 0, len(languages))
	for _, lang := range languages {
		tag, err := language.Parse(lang)
		if err != nil {
			return fmt.Errorf("invalid language code %s: %w", lang, err)
		}
		tags = append(tags, tag)
	}
	p.languages = tags
	return nil
}
