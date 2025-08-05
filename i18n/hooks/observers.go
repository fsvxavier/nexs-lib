package hooks

import (
	"log"
	"sync"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// LoggingObserver implements Observer interface for logging translation events
type LoggingObserver struct {
	logger *log.Logger
	mu     sync.RWMutex
}

// NewLoggingObserver creates a new logging observer
func NewLoggingObserver(logger *log.Logger) *LoggingObserver {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingObserver{logger: logger}
}

// OnTranslationLoaded implements Observer.OnTranslationLoaded
func (o *LoggingObserver) OnTranslationLoaded(provider interfaces.Provider, path string) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	o.logger.Printf("Translation file loaded: %s", path)
}

// OnTranslationNotFound implements Observer.OnTranslationNotFound
func (o *LoggingObserver) OnTranslationNotFound(provider interfaces.Provider, key string) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	o.logger.Printf("Translation not found for key: %s", key)
}

// OnTranslationError implements Observer.OnTranslationError
func (o *LoggingObserver) OnTranslationError(provider interfaces.Provider, key string, err error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	o.logger.Printf("Translation error for key %s: %v", key, err)
}

// MetricsObserver implements Observer interface for collecting translation metrics
type MetricsObserver struct {
	mu              sync.RWMutex
	loadedFiles     int
	notFoundKeys    map[string]int
	translationErrs map[string]int
}

// NewMetricsObserver creates a new metrics observer
func NewMetricsObserver() *MetricsObserver {
	return &MetricsObserver{
		notFoundKeys:    make(map[string]int),
		translationErrs: make(map[string]int),
	}
}

// OnTranslationLoaded implements Observer.OnTranslationLoaded
func (o *MetricsObserver) OnTranslationLoaded(provider interfaces.Provider, path string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.loadedFiles++
}

// OnTranslationNotFound implements Observer.OnTranslationNotFound
func (o *MetricsObserver) OnTranslationNotFound(provider interfaces.Provider, key string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.notFoundKeys[key]++
}

// OnTranslationError implements Observer.OnTranslationError
func (o *MetricsObserver) OnTranslationError(provider interfaces.Provider, key string, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.translationErrs[key]++
}

// GetMetrics returns the current metrics
func (o *MetricsObserver) GetMetrics() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return map[string]interface{}{
		"loaded_files":     o.loadedFiles,
		"not_found_keys":   o.notFoundKeys,
		"translation_errs": o.translationErrs,
	}
}
