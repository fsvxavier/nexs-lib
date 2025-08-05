package hooks_test

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"golang.org/x/text/language"

	"github.com/fsvxavier/nexs-lib/i18n/hooks"
)

// mockProvider implements a minimal Provider interface for testing
type mockProvider struct{}

func (p *mockProvider) LoadTranslations(path string, format string) error { return nil }
func (p *mockProvider) Translate(key string, templateData map[string]interface{}) (string, error) {
	return "", nil
}
func (p *mockProvider) TranslatePlural(key string, count interface{}, templateData map[string]interface{}) (string, error) {
	return "", nil
}
func (p *mockProvider) SetLanguages(langs ...string) error { return nil }
func (p *mockProvider) GetLanguages() []language.Tag       { return nil }

func TestLoggingObserver(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := hooks.NewLoggingObserver(logger)
	provider := &mockProvider{}

	// Test OnTranslationLoaded
	t.Run("OnTranslationLoaded", func(t *testing.T) {
		buf.Reset()
		path := "/path/to/translations.json"
		observer.OnTranslationLoaded(provider, path)
		expected := "Translation file loaded: /path/to/translations.json\n"
		if buf.String() != expected {
			t.Errorf("Expected log %q, got %q", expected, buf.String())
		}
	})

	// Test OnTranslationNotFound
	t.Run("OnTranslationNotFound", func(t *testing.T) {
		buf.Reset()
		key := "missing.key"
		observer.OnTranslationNotFound(provider, key)
		expected := "Translation not found for key: missing.key\n"
		if buf.String() != expected {
			t.Errorf("Expected log %q, got %q", expected, buf.String())
		}
	})

	// Test OnTranslationError
	t.Run("OnTranslationError", func(t *testing.T) {
		buf.Reset()
		key := "error.key"
		err := errors.New("test error")
		observer.OnTranslationError(provider, key, err)
		expected := "Translation error for key error.key: test error\n"
		if buf.String() != expected {
			t.Errorf("Expected log %q, got %q", expected, buf.String())
		}
	})
}

func TestMetricsObserver(t *testing.T) {
	observer := hooks.NewMetricsObserver()
	provider := &mockProvider{}

	// Test initial state
	t.Run("InitialState", func(t *testing.T) {
		metrics := observer.GetMetrics()
		if metrics["loaded_files"].(int) != 0 {
			t.Errorf("Expected 0 loaded files, got %d", metrics["loaded_files"])
		}
		if len(metrics["not_found_keys"].(map[string]int)) != 0 {
			t.Error("Expected empty not_found_keys map")
		}
		if len(metrics["translation_errs"].(map[string]int)) != 0 {
			t.Error("Expected empty translation_errs map")
		}
	})

	// Test metrics collection
	t.Run("MetricsCollection", func(t *testing.T) {
		// Load some files
		observer.OnTranslationLoaded(provider, "file1.json")
		observer.OnTranslationLoaded(provider, "file2.json")

		// Record some missing keys
		observer.OnTranslationNotFound(provider, "key1")
		observer.OnTranslationNotFound(provider, "key1") // duplicate
		observer.OnTranslationNotFound(provider, "key2")

		// Record some errors
		err := errors.New("test error")
		observer.OnTranslationError(provider, "key1", err)
		observer.OnTranslationError(provider, "key1", err) // duplicate

		metrics := observer.GetMetrics()

		// Check loaded files
		if metrics["loaded_files"].(int) != 2 {
			t.Errorf("Expected 2 loaded files, got %d", metrics["loaded_files"])
		}

		// Check not found keys
		notFoundKeys := metrics["not_found_keys"].(map[string]int)
		if notFoundKeys["key1"] != 2 {
			t.Errorf("Expected key1 not found 2 times, got %d", notFoundKeys["key1"])
		}
		if notFoundKeys["key2"] != 1 {
			t.Errorf("Expected key2 not found 1 time, got %d", notFoundKeys["key2"])
		}

		// Check translation errors
		translationErrs := metrics["translation_errs"].(map[string]int)
		if translationErrs["key1"] != 2 {
			t.Errorf("Expected key1 errors 2 times, got %d", translationErrs["key1"])
		}
	})

	// Test metrics immutability
	t.Run("MetricsImmutability", func(t *testing.T) {
		metrics1 := observer.GetMetrics()
		metrics2 := observer.GetMetrics()

		// Modify first metrics map
		metrics1["loaded_files"] = 999

		// Verify second map is unchanged
		if metrics2["loaded_files"].(int) == 999 {
			t.Error("GetMetrics should return a copy of the metrics")
		}
	})
}
