package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/middleware"
)

// mockProvider implements a minimal Provider interface for testing
type mockProvider struct {
	languages []language.Tag
}

func (p *mockProvider) LoadTranslations(path string, format string) error { return nil }
func (p *mockProvider) Translate(key string, templateData map[string]interface{}) (string, error) {
	return "", nil
}
func (p *mockProvider) TranslatePlural(key string, count interface{}, templateData map[string]interface{}) (string, error) {
	return "", nil
}
func (p *mockProvider) SetLanguages(langs ...string) error {
	p.languages = make([]language.Tag, len(langs))
	for i, lang := range langs {
		tag, err := language.Parse(lang)
		if err != nil {
			return err
		}
		p.languages[i] = tag
	}
	return nil
}
func (p *mockProvider) GetLanguages() []language.Tag { return p.languages }

func TestMiddleware(t *testing.T) {
	provider := &mockProvider{}
	defaultLang := language.MustParse("pt-BR")

	config := middleware.Config{
		Provider:        provider,
		HeaderName:      "Accept-Language",
		QueryParam:      "lang",
		DefaultLanguage: defaultLang,
	}

	middleware := middleware.New(config)

	tests := []struct {
		name           string
		header         string
		query          string
		expectedLangs  []string
		expectProvider bool
	}{
		{
			name:           "Header language",
			header:         "es-MX,es;q=0.9,en;q=0.8",
			query:          "",
			expectedLangs:  []string{"es-MX", "es", "en"},
			expectProvider: true,
		},
		{
			name:           "Query parameter",
			header:         "",
			query:          "fr",
			expectedLangs:  []string{"fr"},
			expectProvider: true,
		},
		{
			name:           "Query overrides header",
			header:         "es-MX,es;q=0.9",
			query:          "fr",
			expectedLangs:  []string{"fr"},
			expectProvider: true,
		},
		{
			name:           "Invalid language",
			header:         "invalid",
			query:          "",
			expectedLangs:  []string{"pt-BR"},
			expectProvider: true,
		},
		{
			name:           "No language specified",
			header:         "",
			query:          "",
			expectedLangs:  []string{"pt-BR"},
			expectProvider: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with test headers
			r := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				r.Header.Set("Accept-Language", tt.header)
			}
			if tt.query != "" {
				q := r.URL.Query()
				q.Set("lang", tt.query)
				r.URL.RawQuery = q.Encode()
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a test handler that checks the context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get provider from context directly
				v := r.Context().Value(middleware.LocalizerKey)
				if v == nil && tt.expectProvider {
					t.Error("Expected provider in context but got none")
					return
				}
				if v != nil && !tt.expectProvider {
					t.Error("Expected no provider in context but got one")
					return
				}
				if v != nil {
					provider := v.(interfaces.Provider)
					langs := provider.GetLanguages()
					if len(langs) != len(tt.expectedLangs) {
						t.Errorf("Expected %d languages, got %d", len(tt.expectedLangs), len(langs))
						return
					}
					for i, expected := range tt.expectedLangs {
						expectedTag, _ := language.Parse(expected)
						if langs[i] != expectedTag {
							t.Errorf("Expected language %v at position %d, got %v", expectedTag, i, langs[i])
						}
					}
				}
			})

			// Run the middleware
			middleware(handler).ServeHTTP(w, r)

			// Check response status
			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestMustGetProvider(t *testing.T) {
	t.Run("PanicsOnMissingProvider", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected MustGetProvider to panic")
			}
		}()

		r := httptest.NewRequest("GET", "/", nil)
		middleware.MustGetProvider(r.Context())
	})
}
