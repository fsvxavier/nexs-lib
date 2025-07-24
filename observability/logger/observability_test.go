package logger

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("RecordLog", func(t *testing.T) {
		// Registra logs de diferentes níveis
		collector.RecordLog(interfaces.InfoLevel, 10*time.Millisecond)
		collector.RecordLog(interfaces.ErrorLevel, 15*time.Millisecond)
		collector.RecordLog(interfaces.InfoLevel, 20*time.Millisecond)

		metrics := collector.GetMetrics()

		// Verifica contadores
		if count := metrics.GetLogCount(interfaces.InfoLevel); count != 2 {
			t.Errorf("Expected 2 info logs, got %d", count)
		}
		if count := metrics.GetLogCount(interfaces.ErrorLevel); count != 1 {
			t.Errorf("Expected 1 error log, got %d", count)
		}
		if total := metrics.GetTotalLogCount(); total != 3 {
			t.Errorf("Expected 3 total logs, got %d", total)
		}

		// Verifica tempo de processamento
		if avgTime := metrics.GetProcessingTimeByLevel(interfaces.InfoLevel); avgTime == 0 {
			t.Error("Expected non-zero average processing time for info level")
		}
	})

	t.Run("RecordError", func(t *testing.T) {
		collector.RecordError(errors.New("test error"))
		collector.RecordError(nil) // Não deve incrementar

		metrics := collector.GetMetrics()
		if rate := metrics.GetErrorRate(); rate == 0 {
			t.Error("Expected non-zero error rate")
		}
	})

	t.Run("RecordSample", func(t *testing.T) {
		collector.RecordSample(true)
		collector.RecordSample(false)
		collector.RecordSample(true)

		metrics := collector.GetMetrics()
		if rate := metrics.GetSamplingRate(); rate != 2.0/3.0 {
			t.Errorf("Expected sampling rate of 0.67, got %f", rate)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		collector.RecordLog(interfaces.InfoLevel, 10*time.Millisecond)

		metrics := collector.GetMetrics()
		metrics.Reset()

		if total := metrics.GetTotalLogCount(); total != 0 {
			t.Errorf("Expected 0 total logs after reset, got %d", total)
		}
	})

	t.Run("Export", func(t *testing.T) {
		collector.RecordLog(interfaces.InfoLevel, 10*time.Millisecond)

		metrics := collector.GetMetrics()
		exported := metrics.Export()

		if exported == nil {
			t.Error("Expected non-nil exported metrics")
		}
		if _, exists := exported["log_counts"]; !exists {
			t.Error("Expected log_counts in exported metrics")
		}
	})
}

func TestSpecificHooks(t *testing.T) {
	t.Run("MetricsHook", func(t *testing.T) {
		collector := NewMetricsCollector()
		hook := NewMetricsHook(collector)

		entry := &interfaces.LogEntry{
			Level:   interfaces.InfoLevel,
			Message: "test message",
		}

		err := hook.Execute(context.Background(), entry)
		if err != nil {
			t.Errorf("Unexpected error in metrics hook: %v", err)
		}

		// Verifica se métrica foi registrada
		metrics := collector.GetMetrics()
		if count := metrics.GetLogCount(interfaces.InfoLevel); count != 1 {
			t.Errorf("Expected 1 info log, got %d", count)
		}
	})

	t.Run("ValidationHook", func(t *testing.T) {
		validator := func(entry *interfaces.LogEntry) error {
			if entry.Message == "" {
				return errors.New("message cannot be empty")
			}
			return nil
		}

		hook := NewValidationHook(validator)

		// Teste com mensagem válida
		entry := &interfaces.LogEntry{
			Level:   interfaces.InfoLevel,
			Message: "test message",
		}

		err := hook.Execute(context.Background(), entry)
		if err != nil {
			t.Errorf("Unexpected error in validation hook: %v", err)
		}

		// Teste com mensagem inválida
		entry.Message = ""
		err = hook.Execute(context.Background(), entry)
		if err == nil {
			t.Error("Expected validation error")
		}
	})

	t.Run("FilterHook", func(t *testing.T) {
		filter := func(entry *interfaces.LogEntry) bool {
			return entry.Level >= interfaces.WarnLevel
		}

		hook := NewFilterHook(filter)

		// Teste com log que passa no filtro
		entry := &interfaces.LogEntry{
			Level:   interfaces.ErrorLevel,
			Message: "error message",
		}

		err := hook.Execute(context.Background(), entry)
		if err != nil {
			t.Errorf("Unexpected error in filter hook: %v", err)
		}

		// Teste com log que não passa no filtro
		entry.Level = interfaces.InfoLevel
		err = hook.Execute(context.Background(), entry)
		if err == nil {
			t.Error("Expected filter error")
		}
	})

	t.Run("TransformHook", func(t *testing.T) {
		transformer := func(entry *interfaces.LogEntry) error {
			entry.Message = "[TRANSFORMED] " + entry.Message
			return nil
		}

		hook := NewTransformHook(transformer)

		entry := &interfaces.LogEntry{
			Level:   interfaces.InfoLevel,
			Message: "original message",
		}

		err := hook.Execute(context.Background(), entry)
		if err != nil {
			t.Errorf("Unexpected error in transform hook: %v", err)
		}

		if entry.Message != "[TRANSFORMED] original message" {
			t.Errorf("Expected transformed message, got: %s", entry.Message)
		}
	})
}

func TestHookManager(t *testing.T) {
	manager := NewHookManager()

	t.Run("RegisterHook", func(t *testing.T) {
		hook := NewBaseHook("test_hook")

		err := manager.RegisterHook(interfaces.BeforeHook, hook)
		if err != nil {
			t.Errorf("Unexpected error registering hook: %v", err)
		}

		// Tenta registrar hook com mesmo nome
		err = manager.RegisterHook(interfaces.BeforeHook, hook)
		if err == nil {
			t.Error("Expected error when registering duplicate hook")
		}

		// Tenta registrar hook nil
		err = manager.RegisterHook(interfaces.BeforeHook, nil)
		if err == nil {
			t.Error("Expected error when registering nil hook")
		}
	})

	t.Run("ClearHooks", func(t *testing.T) {
		hook := NewBaseHook("test_hook")
		manager.RegisterHook(interfaces.BeforeHook, hook)

		manager.ClearHooks(interfaces.BeforeHook)

		hooks := manager.ListHooks(interfaces.BeforeHook)
		if len(hooks) != 0 {
			t.Errorf("Expected 0 hooks after clear, got %d", len(hooks))
		}
	})
}
