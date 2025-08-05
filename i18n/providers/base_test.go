package providers_test

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

func TestBaseProviderLanguages(t *testing.T) {
	p := providers.NewBaseProvider()

	// Test SetLanguages
	t.Run("SetLanguages", func(t *testing.T) {
		tests := []struct {
			name        string
			langs       []string
			expectError bool
		}{
			{
				name:        "Valid languages",
				langs:       []string{"pt-BR", "en", "es"},
				expectError: false,
			},
			{
				name:        "Invalid language",
				langs:       []string{"invalid"},
				expectError: true,
			},
			{
				name:        "Empty language list",
				langs:       []string{},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := p.SetLanguages(tt.langs...)
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if !tt.expectError {
					langs := p.GetLanguages()
					if len(langs) != len(tt.langs) {
						t.Errorf("Expected %d languages, got %d", len(tt.langs), len(langs))
					}

					// Verify languages are correctly parsed
					for i, langStr := range tt.langs {
						if !tt.expectError {
							expected, _ := language.Parse(langStr)
							if !langs[i].Equal(expected) {
								t.Errorf("Language at index %d: expected %v, got %v", i, expected, langs[i])
							}
						}
					}
				}
			})
		}
	})

	// Test GetLanguages immutability
	t.Run("GetLanguagesImmutability", func(t *testing.T) {
		err := p.SetLanguages("pt-BR", "en")
		if err != nil {
			t.Fatalf("Failed to set languages: %v", err)
		}

		langs1 := p.GetLanguages()
		langs2 := p.GetLanguages()

		if len(langs1) != len(langs2) {
			t.Errorf("Expected same length, got %d and %d", len(langs1), len(langs2))
		}

		// Modify first slice
		if len(langs1) > 0 {
			langs1[0] = language.French
		}

		// Verify second slice is unchanged
		langs3 := p.GetLanguages()
		if len(langs2) > 0 && len(langs3) > 0 {
			if langs2[0] != langs3[0] {
				t.Error("GetLanguages should return a copy of the languages")
			}
		}
	})
}
