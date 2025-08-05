package hooks

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewLoggingHook(t *testing.T) {
	tests := []struct {
		name        string
		hookName    string
		priority    int
		config      LoggingHookConfig
		logger      Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid hook with default logger",
			hookName: "test-logging-hook",
			priority: 1,
			config: LoggingHookConfig{
				LogLevel:         "info",
				LogTranslations:  true,
				LogErrors:        true,
				LogLifecycle:     true,
				IncludeTimestamp: true,
				IncludeContext:   true,
			},
			logger:      nil, // should use default logger
			expectError: false,
		},
		{
			name:     "valid hook with custom logger",
			hookName: "test-logging-hook",
			priority: 2,
			config: LoggingHookConfig{
				LogLevel:        "debug",
				LogTranslations: false,
				LogErrors:       true,
				LogLifecycle:    false,
			},
			logger:      &DefaultLogger{},
			expectError: false,
		},
		{
			name:        "empty hook name",
			hookName:    "",
			priority:    1,
			config:      LoggingHookConfig{},
			logger:      nil,
			expectError: true,
			errorMsg:    "hook name cannot be empty",
		},
		{
			name:     "empty log level gets default",
			hookName: "test-hook",
			priority: 1,
			config: LoggingHookConfig{
				LogLevel: "", // should default to "info"
			},
			logger:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := NewLoggingHook(tt.hookName, tt.priority, tt.config, tt.logger)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if hook == nil {
					t.Error("expected hook to be created, but got nil")
					return
				}

				if hook.Name() != tt.hookName {
					t.Errorf("expected hook name '%s', got '%s'", tt.hookName, hook.Name())
				}

				if hook.Priority() != tt.priority {
					t.Errorf("expected priority %d, got %d", tt.priority, hook.Priority())
				}

				config := hook.GetConfig()
				expectedLogLevel := tt.config.LogLevel
				if expectedLogLevel == "" {
					expectedLogLevel = "info"
				}

				if config.LogLevel != expectedLogLevel {
					t.Errorf("expected log level '%s', got '%s'", expectedLogLevel, config.LogLevel)
				}
			}
		})
	}
}

func TestLoggingHookObserver(t *testing.T) {
	hook, err := NewLoggingHook("test-hook", 1, LoggingHookConfig{
		LogLevel:         "debug",
		LogTranslations:  true,
		LogErrors:        true,
		LogLifecycle:     true,
		IncludeTimestamp: false,
		IncludeContext:   false,
	}, &TestLogger{})
	if err != nil {
		t.Fatalf("failed to create logging hook: %v", err)
	}

	ctx := context.Background()
	providerName := "test-provider"

	// Test OnStart
	err = hook.OnStart(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStart, got: %v", err)
	}

	// Test OnStop
	err = hook.OnStop(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStop, got: %v", err)
	}

	// Test OnError
	testErr := fmt.Errorf("test error")
	err = hook.OnError(ctx, providerName, testErr)
	if err != nil {
		t.Errorf("expected no error from OnError, got: %v", err)
	}

	// Test OnTranslate
	err = hook.OnTranslate(ctx, providerName, "test.key", "en", "Test Result")
	if err != nil {
		t.Errorf("expected no error from OnTranslate, got: %v", err)
	}
}

func TestLoggingHookConfigUpdate(t *testing.T) {
	hook, err := NewLoggingHook("test-hook", 1, LoggingHookConfig{
		LogLevel: "info",
	}, &TestLogger{})
	if err != nil {
		t.Fatalf("failed to create logging hook: %v", err)
	}

	newConfig := LoggingHookConfig{
		LogLevel:         "debug",
		LogTranslations:  true,
		LogErrors:        false,
		LogLifecycle:     true,
		IncludeTimestamp: true,
		IncludeContext:   false,
	}

	err = hook.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("expected no error updating config, got: %v", err)
	}

	config := hook.GetConfig()
	if config.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got '%s'", config.LogLevel)
	}

	if !config.LogTranslations {
		t.Error("expected log translations to be true")
	}

	if config.LogErrors {
		t.Error("expected log errors to be false")
	}

	if !config.LogLifecycle {
		t.Error("expected log lifecycle to be true")
	}

	if !config.IncludeTimestamp {
		t.Error("expected include timestamp to be true")
	}

	if config.IncludeContext {
		t.Error("expected include context to be false")
	}
}

func TestLoggingHookLogLevels(t *testing.T) {
	hook, err := NewLoggingHook("test-hook", 1, LoggingHookConfig{
		LogLevel: "warn",
	}, &TestLogger{})
	if err != nil {
		t.Fatalf("failed to create logging hook: %v", err)
	}

	tests := []struct {
		level     string
		shouldLog bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", true},
		{"error", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := hook.shouldLog(tt.level)
			if result != tt.shouldLog {
				t.Errorf("expected shouldLog(%s) to be %v, got %v", tt.level, tt.shouldLog, result)
			}
		})
	}
}

func TestNewMetricsHook(t *testing.T) {
	tests := []struct {
		name        string
		hookName    string
		priority    int
		config      MetricsHookConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid hook with full config",
			hookName: "test-metrics-hook",
			priority: 1,
			config: MetricsHookConfig{
				CollectTranslationMetrics: true,
				CollectErrorMetrics:       true,
				CollectPerformanceMetrics: true,
				MetricsInterval:           5 * time.Minute,
			},
			expectError: false,
		},
		{
			name:        "empty hook name",
			hookName:    "",
			priority:    1,
			config:      MetricsHookConfig{},
			expectError: true,
			errorMsg:    "hook name cannot be empty",
		},
		{
			name:     "zero metrics interval gets default",
			hookName: "test-hook",
			priority: 1,
			config: MetricsHookConfig{
				MetricsInterval: 0, // should default to 1 minute
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := NewMetricsHook(tt.hookName, tt.priority, tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if hook == nil {
					t.Error("expected hook to be created, but got nil")
					return
				}

				if hook.Name() != tt.hookName {
					t.Errorf("expected hook name '%s', got '%s'", tt.hookName, hook.Name())
				}

				if hook.Priority() != tt.priority {
					t.Errorf("expected priority %d, got %d", tt.priority, hook.Priority())
				}
			}
		})
	}
}

func TestMetricsHookCollectMetrics(t *testing.T) {
	hook, err := NewMetricsHook("test-metrics-hook", 1, MetricsHookConfig{
		CollectTranslationMetrics: true,
		CollectErrorMetrics:       true,
		CollectPerformanceMetrics: true,
		MetricsInterval:           1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("failed to create metrics hook: %v", err)
	}

	ctx := context.Background()
	providerName := "test-provider"

	// Test OnStart (should collect performance metrics)
	err = hook.OnStart(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStart, got: %v", err)
	}

	// Test OnTranslate (should collect translation metrics)
	err = hook.OnTranslate(ctx, providerName, "test.key", "en", "Test Result")
	if err != nil {
		t.Errorf("expected no error from OnTranslate, got: %v", err)
	}

	// Test OnError (should collect error metrics)
	testErr := fmt.Errorf("test error")
	err = hook.OnError(ctx, providerName, testErr)
	if err != nil {
		t.Errorf("expected no error from OnError, got: %v", err)
	}

	// Test OnStop (should collect performance metrics)
	err = hook.OnStop(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStop, got: %v", err)
	}

	// Verify metrics were collected
	metrics := hook.GetMetrics()

	if metrics.TranslationCount[providerName] != 1 {
		t.Errorf("expected translation count 1, got %d", metrics.TranslationCount[providerName])
	}

	if metrics.ErrorCount[providerName] != 1 {
		t.Errorf("expected error count 1, got %d", metrics.ErrorCount[providerName])
	}

	if len(metrics.PerformanceData) != 2 { // start and stop events
		t.Errorf("expected 2 performance metrics, got %d", len(metrics.PerformanceData))
	}
}

func TestMetricsHookResetMetrics(t *testing.T) {
	hook, err := NewMetricsHook("test-metrics-hook", 1, MetricsHookConfig{
		CollectTranslationMetrics: true,
		CollectErrorMetrics:       true,
		CollectPerformanceMetrics: true,
	})
	if err != nil {
		t.Fatalf("failed to create metrics hook: %v", err)
	}

	ctx := context.Background()
	providerName := "test-provider"

	// Collect some metrics
	_ = hook.OnTranslate(ctx, providerName, "test.key", "en", "Test Result")
	_ = hook.OnError(ctx, providerName, fmt.Errorf("test error"))
	_ = hook.OnStart(ctx, providerName)

	// Verify metrics were collected
	metrics := hook.GetMetrics()
	if metrics.TranslationCount[providerName] == 0 {
		t.Error("expected translation count to be > 0 before reset")
	}

	// Reset metrics
	err = hook.ResetMetrics()
	if err != nil {
		t.Errorf("expected no error resetting metrics, got: %v", err)
	}

	// Verify metrics were reset
	metrics = hook.GetMetrics()
	if metrics.TranslationCount[providerName] != 0 {
		t.Errorf("expected translation count 0 after reset, got %d", metrics.TranslationCount[providerName])
	}

	if metrics.ErrorCount[providerName] != 0 {
		t.Errorf("expected error count 0 after reset, got %d", metrics.ErrorCount[providerName])
	}

	if len(metrics.PerformanceData) != 0 {
		t.Errorf("expected 0 performance metrics after reset, got %d", len(metrics.PerformanceData))
	}
}

func TestNewValidationHook(t *testing.T) {
	tests := []struct {
		name        string
		hookName    string
		priority    int
		config      ValidationHookConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid hook with full config",
			hookName: "test-validation-hook",
			priority: 1,
			config: ValidationHookConfig{
				ValidateKeys:      true,
				ValidateLanguages: true,
				ValidateResults:   true,
				AllowedKeyPattern: "^[a-z.]+$",
				AllowedLanguages:  []string{"en", "es", "pt"},
				MaxResultLength:   5000,
			},
			expectError: false,
		},
		{
			name:        "empty hook name",
			hookName:    "",
			priority:    1,
			config:      ValidationHookConfig{},
			expectError: true,
			errorMsg:    "hook name cannot be empty",
		},
		{
			name:     "zero max result length gets default",
			hookName: "test-hook",
			priority: 1,
			config: ValidationHookConfig{
				MaxResultLength: 0, // should default to 10000
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := NewValidationHook(tt.hookName, tt.priority, tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if hook == nil {
					t.Error("expected hook to be created, but got nil")
					return
				}

				if hook.Name() != tt.hookName {
					t.Errorf("expected hook name '%s', got '%s'", tt.hookName, hook.Name())
				}

				if hook.Priority() != tt.priority {
					t.Errorf("expected priority %d, got %d", tt.priority, hook.Priority())
				}
			}
		})
	}
}

func TestValidationHookValidation(t *testing.T) {
	hook, err := NewValidationHook("test-validation-hook", 1, ValidationHookConfig{
		ValidateKeys:      true,
		ValidateLanguages: true,
		ValidateResults:   true,
		AllowedLanguages:  []string{"en", "es", "pt"},
		MaxResultLength:   100,
	})
	if err != nil {
		t.Fatalf("failed to create validation hook: %v", err)
	}

	ctx := context.Background()
	providerName := "test-provider"

	tests := []struct {
		name        string
		key         string
		lang        string
		result      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid translation",
			key:         "test.key",
			lang:        "en",
			result:      "Test Result",
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			lang:        "en",
			result:      "Test Result",
			expectError: true,
			errorMsg:    "invalid translation key: translation key cannot be empty",
		},
		{
			name:        "empty language",
			key:         "test.key",
			lang:        "",
			result:      "Test Result",
			expectError: true,
			errorMsg:    "invalid language code: language code cannot be empty",
		},
		{
			name:        "invalid language",
			key:         "test.key",
			lang:        "fr",
			result:      "Test Result",
			expectError: true,
			errorMsg:    "invalid language code: language code 'fr' is not in allowed languages",
		},
		{
			name:        "result too long",
			key:         "test.key",
			lang:        "en",
			result:      string(make([]byte, 101)), // 101 characters, exceeds max of 100
			expectError: true,
			errorMsg:    "invalid translation result: translation result exceeds maximum length of 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hook.OnTranslate(ctx, providerName, tt.key, tt.lang, tt.result)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestValidationHookConfigUpdate(t *testing.T) {
	hook, err := NewValidationHook("test-hook", 1, ValidationHookConfig{
		ValidateKeys: false,
	})
	if err != nil {
		t.Fatalf("failed to create validation hook: %v", err)
	}

	newConfig := ValidationHookConfig{
		ValidateKeys:      true,
		ValidateLanguages: true,
		ValidateResults:   false,
		AllowedLanguages:  []string{"en", "pt"},
		MaxResultLength:   500,
	}

	err = hook.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("expected no error updating config, got: %v", err)
	}

	config := hook.GetConfig()
	if !config.ValidateKeys {
		t.Error("expected validate keys to be true")
	}

	if !config.ValidateLanguages {
		t.Error("expected validate languages to be true")
	}

	if config.ValidateResults {
		t.Error("expected validate results to be false")
	}

	if len(config.AllowedLanguages) != 2 {
		t.Errorf("expected 2 allowed languages, got %d", len(config.AllowedLanguages))
	}

	if config.MaxResultLength != 500 {
		t.Errorf("expected max result length 500, got %d", config.MaxResultLength)
	}
}

func TestValidationHookLifecycleEvents(t *testing.T) {
	hook, err := NewValidationHook("test-validation-hook", 1, ValidationHookConfig{})
	if err != nil {
		t.Fatalf("failed to create validation hook: %v", err)
	}

	ctx := context.Background()
	providerName := "test-provider"

	// Test OnStart (should not validate anything)
	err = hook.OnStart(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStart, got: %v", err)
	}

	// Test OnStop (should not validate anything)
	err = hook.OnStop(ctx, providerName)
	if err != nil {
		t.Errorf("expected no error from OnStop, got: %v", err)
	}

	// Test OnError (should not validate anything)
	testErr := fmt.Errorf("test error")
	err = hook.OnError(ctx, providerName, testErr)
	if err != nil {
		t.Errorf("expected no error from OnError, got: %v", err)
	}
}

// TestLogger is a test implementation of the Logger interface
type TestLogger struct {
	Messages []LogMessage
}

type LogMessage struct {
	Level   string
	Message string
	Args    []interface{}
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "debug", Message: msg, Args: args})
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "info", Message: msg, Args: args})
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "warn", Message: msg, Args: args})
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "error", Message: msg, Args: args})
}

func (l *TestLogger) Reset() {
	l.Messages = nil
}

func TestDefaultLogger(t *testing.T) {
	logger := &DefaultLogger{}

	// These tests just ensure the logger methods don't panic
	// The actual output goes to the standard log package
	logger.Debug("test debug message")
	logger.Info("test info message")
	logger.Warn("test warn message")
	logger.Error("test error message")

	// Test with format arguments
	logger.Debug("test debug message with arg: %s", "value")
	logger.Info("test info message with arg: %d", 42)
	logger.Warn("test warn message with arg: %v", true)
	logger.Error("test error message with arg: %f", 3.14)
}
