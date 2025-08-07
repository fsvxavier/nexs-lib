package performance

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// BenchmarkErrorCreation testa performance de criação de erros
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("StandardError", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fmt.Errorf("test error %d", i)
		}
	})

	b.Run("PooledDomainError", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := NewPooledError(interfaces.ValidationError, "TEST_ERROR", "test error "+strconv.Itoa(i))
			err.Release() // Importante: liberar de volta ao pool
		}
	})

	b.Run("PooledErrorWithMetadata", func(b *testing.B) {
		metadata := map[string]interface{}{
			"user_id": 12345,
			"action":  "create_user",
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := NewPooledErrorWithMetadata(interfaces.ValidationError, "TEST_ERROR", "test error", metadata)
			err.Release()
		}
	})
}

// BenchmarkStackTraceCapture testa performance de captura de stack trace
func BenchmarkStackTraceCapture(b *testing.B) {
	b.Run("RuntimeStack", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf := make([]byte, 1024)
			_ = runtime.Stack(buf, false)
		}
	})

	b.Run("RuntimeCallers", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			pcs := make([]uintptr, 32)
			_ = runtime.Callers(0, pcs)
		}
	})

	b.Run("LazyStackTrace", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			lst := NewLazyStackTrace(1)
			_ = lst.HasFrames() // Apenas verifica, não captura detalhes
		}
	})

	b.Run("LazyStackTraceWithDetails", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			lst := NewLazyStackTrace(1)
			_ = lst.GetFrames() // Força captura de detalhes
		}
	})

	b.Run("OptimizedStackCapture", func(b *testing.B) {
		capture := NewOptimizedStackCapture()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			lst := capture.CaptureStackTrace(1)
			capture.ReleaseStackTrace(lst)
		}
	})
}

// BenchmarkStringOperations testa performance de operações com strings
func BenchmarkStringOperations(b *testing.B) {
	b.Run("StandardStringConcat", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			msg := "Error occurred: " + strconv.Itoa(i) + " at " + time.Now().String()
			_ = msg
		}
	})

	b.Run("StringBuilderConcat", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			builder.WriteString("Error occurred: ")
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString(" at ")
			builder.WriteString(time.Now().String())
			_ = builder.String()
		}
	})

	b.Run("StringInternCommon", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = InternString("VALIDATION_ERROR")
		}
	})

	b.Run("StringInternUncommon", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = InternString("UNCOMMON_ERROR_" + strconv.Itoa(i%100))
		}
	})
}

// BenchmarkPoolOperations testa performance das operações de pool
func BenchmarkPoolOperations(b *testing.B) {
	b.Run("ErrorPoolGetPut", func(b *testing.B) {
		pool := NewErrorPool()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := pool.GetDomainError()
			pool.PutDomainError(err)
		}
	})

	b.Run("MetadataPoolGetPut", func(b *testing.B) {
		pool := NewErrorPool()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			metadata := pool.GetMetadata()
			metadata["key"] = "value"
			pool.PutMetadata(metadata)
		}
	})

	b.Run("StackTracePoolGetPut", func(b *testing.B) {
		pool := NewStackTracePool()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			lst := pool.Get(1)
			pool.Put(lst)
		}
	})
}

// BenchmarkConcurrentOperations testa performance com concorrência
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentErrorCreation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := NewPooledError(interfaces.ValidationError, "TEST_ERROR", "concurrent test")
				err.Release()
			}
		})
	})

	b.Run("ConcurrentStackCapture", func(b *testing.B) {
		capture := NewOptimizedStackCapture()

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				lst := capture.CaptureStackTrace(1)
				capture.ReleaseStackTrace(lst)
			}
		})
	})

	b.Run("ConcurrentStringIntern", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = InternString("VALIDATION_ERROR")
			}
		})
	})
}

// BenchmarkMemoryUsage testa uso de memória
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("TraditionalErrorAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		errors := make([]error, b.N)

		for i := 0; i < b.N; i++ {
			errors[i] = fmt.Errorf("error %d with metadata: %s", i, "some data")
		}

		// Previne otimização
		if len(errors) != b.N {
			b.Fatal("unexpected length")
		}
	})

	b.Run("PooledErrorAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		errors := make([]*PooledDomainError, b.N)

		for i := 0; i < b.N; i++ {
			err := NewPooledError(interfaces.ValidationError, "TEST_ERROR", fmt.Sprintf("error %d", i))
			err.WithMetadata("data", "some data")
			errors[i] = err
		}

		// Libera de volta ao pool
		for _, err := range errors {
			err.Release()
		}

		// Previne otimização
		if len(errors) != b.N {
			b.Fatal("unexpected length")
		}
	})
}

// BenchmarkRealWorldScenarios testa cenários do mundo real
func BenchmarkRealWorldScenarios(b *testing.B) {
	b.Run("ValidationErrorFlow", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Simula fluxo de validação
			err := NewPooledError(interfaces.ValidationError, "INVALID_EMAIL", "Email format is invalid")
			err.WithMetadata("field", "email")
			err.WithMetadata("value", "invalid-email")
			err.WithMetadata("user_id", 12345)

			// Simula processamento
			_ = err.Error()
			_ = err.HTTPStatus()
			_ = err.Type()
			_ = err.Metadata()

			err.Release()
		}
	})

	b.Run("ExternalServiceErrorFlow", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Simula erro de serviço externo
			err := NewPooledError(interfaces.ExternalServiceError, "PAYMENT_SERVICE_TIMEOUT", "Payment service timeout")
			err.WithMetadata("service", "payment-api")
			err.WithMetadata("timeout", "30s")
			err.WithMetadata("attempt", i%3+1)

			// Simula captura de stack trace condicional
			if i%10 == 0 { // Apenas 10% das vezes
				lst := CaptureConditionalStackTrace(1)
				_ = lst.HasFrames()
				ReleaseStackTrace(lst)
			}

			err.Release()
		}
	})
}

// PerformanceProfiler ajuda a medir performance em tempo real
type PerformanceProfiler struct {
	measurements []PerformanceMeasurement
	enabled      bool
}

type PerformanceMeasurement struct {
	Operation  string
	Duration   time.Duration
	AllocSize  int64
	AllocCount int64
	Timestamp  time.Time
}

// NewPerformanceProfiler cria novo profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		measurements: make([]PerformanceMeasurement, 0, 1000),
		enabled:      true,
	}
}

// Measure mede performance de uma operação
func (pp *PerformanceProfiler) Measure(operation string, fn func()) {
	if !pp.enabled {
		fn()
		return
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()
	fn()
	duration := time.Since(start)

	runtime.ReadMemStats(&m2)

	measurement := PerformanceMeasurement{
		Operation:  operation,
		Duration:   duration,
		AllocSize:  int64(m2.TotalAlloc - m1.TotalAlloc),
		AllocCount: int64(m2.Mallocs - m1.Mallocs),
		Timestamp:  start,
	}

	pp.measurements = append(pp.measurements, measurement)
}

// GetMeasurements retorna todas as medições
func (pp *PerformanceProfiler) GetMeasurements() []PerformanceMeasurement {
	return pp.measurements
}

// Clear limpa todas as medições
func (pp *PerformanceProfiler) Clear() {
	pp.measurements = pp.measurements[:0]
}

// SetEnabled habilita/desabilita profiler
func (pp *PerformanceProfiler) SetEnabled(enabled bool) {
	pp.enabled = enabled
}

// GetStats retorna estatísticas das medições
func (pp *PerformanceProfiler) GetStats() map[string]interface{} {
	if len(pp.measurements) == 0 {
		return map[string]interface{}{}
	}

	var totalDuration time.Duration
	var totalAllocs int64
	var totalAllocSize int64

	for _, m := range pp.measurements {
		totalDuration += m.Duration
		totalAllocs += m.AllocCount
		totalAllocSize += m.AllocSize
	}

	count := int64(len(pp.measurements))

	return map[string]interface{}{
		"total_operations":    count,
		"average_duration_ns": totalDuration.Nanoseconds() / count,
		"total_duration_ns":   totalDuration.Nanoseconds(),
		"average_allocs":      totalAllocs / count,
		"total_allocs":        totalAllocs,
		"average_alloc_size":  totalAllocSize / count,
		"total_alloc_size":    totalAllocSize,
	}
}

// GlobalProfiler instância global do profiler
var GlobalProfiler = NewPerformanceProfiler()

// MeasureGlobal mede performance usando profiler global
func MeasureGlobal(operation string, fn func()) {
	GlobalProfiler.Measure(operation, fn)
}
