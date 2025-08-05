// Package hooks provides implementation of hooks that observe and react to i18n events.
// Hooks implement the Hook interface and can be registered to execute at specific
// points in the translation provider lifecycle.
package hooks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// LoggingHook is a hook that logs all i18n events for debugging and monitoring purposes.
// It implements the Hook interface and provides configurable logging levels.
type LoggingHook struct {
	name     string
	priority int
	logger   Logger
	config   LoggingHookConfig
	mu       sync.RWMutex
}

// LoggingHookConfig contains configuration options for the logging hook.
type LoggingHookConfig struct {
	// LogLevel determines what events to log ("debug", "info", "warn", "error")
	LogLevel string `json:"log_level" yaml:"log_level"`

	// LogTranslations determines if translation events should be logged
	LogTranslations bool `json:"log_translations" yaml:"log_translations"`

	// LogErrors determines if error events should be logged
	LogErrors bool `json:"log_errors" yaml:"log_errors"`

	// LogLifecycle determines if lifecycle events should be logged
	LogLifecycle bool `json:"log_lifecycle" yaml:"log_lifecycle"`

	// IncludeTimestamp determines if timestamps should be included in logs
	IncludeTimestamp bool `json:"include_timestamp" yaml:"include_timestamp"`

	// IncludeContext determines if context values should be included in logs
	IncludeContext bool `json:"include_context" yaml:"include_context"`
}

// Logger defines the interface for logging operations.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger is a simple logger implementation using the standard log package.
type DefaultLogger struct{}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

// Info logs an info message.
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

// Error logs an error message.
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

// NewLoggingHook creates a new logging hook with the specified configuration.
func NewLoggingHook(name string, priority int, config LoggingHookConfig, logger Logger) (*LoggingHook, error) {
	if name == "" {
		return nil, fmt.Errorf("hook name cannot be empty")
	}

	if logger == nil {
		logger = &DefaultLogger{}
	}

	// Set defaults if not provided
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return &LoggingHook{
		name:     name,
		priority: priority,
		logger:   logger,
		config:   config,
	}, nil
}

// Name returns the unique name of the hook.
func (h *LoggingHook) Name() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.name
}

// Priority returns the execution priority of the hook.
func (h *LoggingHook) Priority() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.priority
}

// OnStart is called when a translation provider starts.
func (h *LoggingHook) OnStart(ctx context.Context, providerName string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.LogLifecycle {
		return nil
	}

	msg := h.formatMessage("Provider started", map[string]interface{}{
		"provider": providerName,
		"hook":     h.name,
	})

	h.logAtLevel("info", msg)
	return nil
}

// OnStop is called when a translation provider stops.
func (h *LoggingHook) OnStop(ctx context.Context, providerName string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.LogLifecycle {
		return nil
	}

	msg := h.formatMessage("Provider stopped", map[string]interface{}{
		"provider": providerName,
		"hook":     h.name,
	})

	h.logAtLevel("info", msg)
	return nil
}

// OnError is called when a translation provider encounters an error.
func (h *LoggingHook) OnError(ctx context.Context, providerName string, err error) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.LogErrors {
		return nil
	}

	msg := h.formatMessage("Provider error", map[string]interface{}{
		"provider": providerName,
		"hook":     h.name,
		"error":    err.Error(),
	})

	h.logAtLevel("error", msg)
	return nil
}

// OnTranslate is called when a translation is performed.
func (h *LoggingHook) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.LogTranslations {
		return nil
	}

	msg := h.formatMessage("Translation performed", map[string]interface{}{
		"provider": providerName,
		"hook":     h.name,
		"key":      key,
		"language": lang,
		"result":   result,
	})

	h.logAtLevel("debug", msg)
	return nil
}

// UpdateConfig updates the hook configuration.
func (h *LoggingHook) UpdateConfig(config LoggingHookConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	h.config = config
	return nil
}

// GetConfig returns the current hook configuration.
func (h *LoggingHook) GetConfig() LoggingHookConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// formatMessage formats a log message with additional context.
func (h *LoggingHook) formatMessage(msg string, context map[string]interface{}) string {
	result := msg

	if h.config.IncludeTimestamp {
		result = fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), result)
	}

	if h.config.IncludeContext && len(context) > 0 {
		result += " |"
		for key, value := range context {
			result += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return result
}

// logAtLevel logs a message at the specified level if it meets the configured log level.
func (h *LoggingHook) logAtLevel(level string, msg string) {
	if !h.shouldLog(level) {
		return
	}

	switch level {
	case "debug":
		h.logger.Debug(msg)
	case "info":
		h.logger.Info(msg)
	case "warn":
		h.logger.Warn(msg)
	case "error":
		h.logger.Error(msg)
	}
}

// shouldLog determines if a message at the given level should be logged.
func (h *LoggingHook) shouldLog(level string) bool {
	levelPriority := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	configPriority, exists := levelPriority[h.config.LogLevel]
	if !exists {
		configPriority = 1 // default to info
	}

	messagePriority, exists := levelPriority[level]
	if !exists {
		return false
	}

	return messagePriority >= configPriority
}

// MetricsHook is a hook that collects metrics about i18n operations.
// It tracks translation counts, error rates, and performance metrics.
type MetricsHook struct {
	name     string
	priority int
	config   MetricsHookConfig
	metrics  *MetricsCollector
	mu       sync.RWMutex
}

// MetricsHookConfig contains configuration options for the metrics hook.
type MetricsHookConfig struct {
	// CollectTranslationMetrics determines if translation metrics should be collected
	CollectTranslationMetrics bool `json:"collect_translation_metrics" yaml:"collect_translation_metrics"`

	// CollectErrorMetrics determines if error metrics should be collected
	CollectErrorMetrics bool `json:"collect_error_metrics" yaml:"collect_error_metrics"`

	// CollectPerformanceMetrics determines if performance metrics should be collected
	CollectPerformanceMetrics bool `json:"collect_performance_metrics" yaml:"collect_performance_metrics"`

	// MetricsInterval is the interval for collecting metrics
	MetricsInterval time.Duration `json:"metrics_interval" yaml:"metrics_interval"`
}

// MetricsCollector collects and stores metrics data.
type MetricsCollector struct {
	TranslationCount map[string]int64    `json:"translation_count"`
	ErrorCount       map[string]int64    `json:"error_count"`
	PerformanceData  []PerformanceMetric `json:"performance_data"`
	mu               sync.RWMutex
}

// PerformanceMetric represents a performance measurement.
type PerformanceMetric struct {
	Provider  string        `json:"provider"`
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewMetricsHook creates a new metrics hook with the specified configuration.
func NewMetricsHook(name string, priority int, config MetricsHookConfig) (*MetricsHook, error) {
	if name == "" {
		return nil, fmt.Errorf("hook name cannot be empty")
	}

	// Set defaults if not provided
	if config.MetricsInterval == 0 {
		config.MetricsInterval = 1 * time.Minute
	}

	return &MetricsHook{
		name:     name,
		priority: priority,
		config:   config,
		metrics: &MetricsCollector{
			TranslationCount: make(map[string]int64),
			ErrorCount:       make(map[string]int64),
			PerformanceData:  make([]PerformanceMetric, 0),
		},
	}, nil
}

// Name returns the unique name of the hook.
func (h *MetricsHook) Name() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.name
}

// Priority returns the execution priority of the hook.
func (h *MetricsHook) Priority() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.priority
}

// OnStart is called when a translation provider starts.
func (h *MetricsHook) OnStart(ctx context.Context, providerName string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.config.CollectPerformanceMetrics {
		h.metrics.mu.Lock()
		h.metrics.PerformanceData = append(h.metrics.PerformanceData, PerformanceMetric{
			Provider:  providerName,
			Operation: "start",
			Duration:  0,
			Timestamp: time.Now(),
		})
		h.metrics.mu.Unlock()
	}

	return nil
}

// OnStop is called when a translation provider stops.
func (h *MetricsHook) OnStop(ctx context.Context, providerName string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.config.CollectPerformanceMetrics {
		h.metrics.mu.Lock()
		h.metrics.PerformanceData = append(h.metrics.PerformanceData, PerformanceMetric{
			Provider:  providerName,
			Operation: "stop",
			Duration:  0,
			Timestamp: time.Now(),
		})
		h.metrics.mu.Unlock()
	}

	return nil
}

// OnError is called when a translation provider encounters an error.
func (h *MetricsHook) OnError(ctx context.Context, providerName string, err error) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.config.CollectErrorMetrics {
		h.metrics.mu.Lock()
		h.metrics.ErrorCount[providerName]++
		h.metrics.mu.Unlock()
	}

	return nil
}

// OnTranslate is called when a translation is performed.
func (h *MetricsHook) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.config.CollectTranslationMetrics {
		h.metrics.mu.Lock()
		h.metrics.TranslationCount[providerName]++
		h.metrics.mu.Unlock()
	}

	return nil
}

// GetMetrics returns the collected metrics.
func (h *MetricsHook) GetMetrics() *MetricsCollector {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()

	// Return a copy of the metrics to avoid concurrent modification
	copy := &MetricsCollector{
		TranslationCount: make(map[string]int64),
		ErrorCount:       make(map[string]int64),
		PerformanceData:  make([]PerformanceMetric, len(h.metrics.PerformanceData)),
	}

	for k, v := range h.metrics.TranslationCount {
		copy.TranslationCount[k] = v
	}

	for k, v := range h.metrics.ErrorCount {
		copy.ErrorCount[k] = v
	}

	copySlice := copy.PerformanceData[:0]
	for _, metric := range h.metrics.PerformanceData {
		copySlice = append(copySlice, metric)
	}
	copy.PerformanceData = copySlice

	return copy
}

// ResetMetrics resets all collected metrics.
func (h *MetricsHook) ResetMetrics() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.TranslationCount = make(map[string]int64)
	h.metrics.ErrorCount = make(map[string]int64)
	h.metrics.PerformanceData = make([]PerformanceMetric, 0)

	return nil
}

// ValidationHook is a hook that validates translation operations and results.
// It can check for common issues like missing translations, invalid parameters, etc.
type ValidationHook struct {
	name     string
	priority int
	config   ValidationHookConfig
	mu       sync.RWMutex
}

// ValidationHookConfig contains configuration options for the validation hook.
type ValidationHookConfig struct {
	// ValidateKeys determines if translation keys should be validated
	ValidateKeys bool `json:"validate_keys" yaml:"validate_keys"`

	// ValidateLanguages determines if language codes should be validated
	ValidateLanguages bool `json:"validate_languages" yaml:"validate_languages"`

	// ValidateResults determines if translation results should be validated
	ValidateResults bool `json:"validate_results" yaml:"validate_results"`

	// AllowedKeyPattern is a regex pattern for valid translation keys
	AllowedKeyPattern string `json:"allowed_key_pattern" yaml:"allowed_key_pattern"`

	// AllowedLanguages is a list of allowed language codes
	AllowedLanguages []string `json:"allowed_languages" yaml:"allowed_languages"`

	// MaxResultLength is the maximum allowed length for translation results
	MaxResultLength int `json:"max_result_length" yaml:"max_result_length"`
}

// NewValidationHook creates a new validation hook with the specified configuration.
func NewValidationHook(name string, priority int, config ValidationHookConfig) (*ValidationHook, error) {
	if name == "" {
		return nil, fmt.Errorf("hook name cannot be empty")
	}

	// Set defaults if not provided
	if config.MaxResultLength == 0 {
		config.MaxResultLength = 10000 // 10KB default
	}

	return &ValidationHook{
		name:     name,
		priority: priority,
		config:   config,
	}, nil
}

// Name returns the unique name of the hook.
func (h *ValidationHook) Name() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.name
}

// Priority returns the execution priority of the hook.
func (h *ValidationHook) Priority() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.priority
}

// OnStart is called when a translation provider starts.
func (h *ValidationHook) OnStart(ctx context.Context, providerName string) error {
	// No validation needed for start events
	return nil
}

// OnStop is called when a translation provider stops.
func (h *ValidationHook) OnStop(ctx context.Context, providerName string) error {
	// No validation needed for stop events
	return nil
}

// OnError is called when a translation provider encounters an error.
func (h *ValidationHook) OnError(ctx context.Context, providerName string, err error) error {
	// No validation needed for error events (they're already errors)
	return nil
}

// OnTranslate is called when a translation is performed.
func (h *ValidationHook) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Validate translation key
	if h.config.ValidateKeys {
		if err := h.validateKey(key); err != nil {
			return fmt.Errorf("invalid translation key: %w", err)
		}
	}

	// Validate language code
	if h.config.ValidateLanguages {
		if err := h.validateLanguage(lang); err != nil {
			return fmt.Errorf("invalid language code: %w", err)
		}
	}

	// Validate translation result
	if h.config.ValidateResults {
		if err := h.validateResult(result); err != nil {
			return fmt.Errorf("invalid translation result: %w", err)
		}
	}

	return nil
}

// validateKey validates a translation key according to the configured pattern.
func (h *ValidationHook) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("translation key cannot be empty")
	}

	// Additional key validation logic can be added here
	// For example, regex pattern matching if AllowedKeyPattern is set

	return nil
}

// validateLanguage validates a language code according to the allowed languages.
func (h *ValidationHook) validateLanguage(lang string) error {
	if lang == "" {
		return fmt.Errorf("language code cannot be empty")
	}

	if len(h.config.AllowedLanguages) > 0 {
		for _, allowedLang := range h.config.AllowedLanguages {
			if lang == allowedLang {
				return nil
			}
		}
		return fmt.Errorf("language code '%s' is not in allowed languages", lang)
	}

	return nil
}

// validateResult validates a translation result according to the configured constraints.
func (h *ValidationHook) validateResult(result string) error {
	if len(result) > h.config.MaxResultLength {
		return fmt.Errorf("translation result exceeds maximum length of %d characters", h.config.MaxResultLength)
	}

	return nil
}

// UpdateConfig updates the hook configuration.
func (h *ValidationHook) UpdateConfig(config ValidationHookConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if config.MaxResultLength == 0 {
		config.MaxResultLength = 10000
	}

	h.config = config
	return nil
}

// GetConfig returns the current hook configuration.
func (h *ValidationHook) GetConfig() ValidationHookConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}
