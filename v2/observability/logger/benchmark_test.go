package logger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/slog"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/zap"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/zerolog"
)

// BenchmarkLogger_SyncVsAsync compara performance síncrona vs assíncrona
func BenchmarkLogger_SyncVsAsync(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	b.Run("Sync", func(b *testing.B) {
		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
			// Sem configuração async = modo síncrono
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message",
					interfaces.String("key1", "value1"),
					interfaces.Int("key2", 42),
				)
			}
		})
	})

	b.Run("Async", func(b *testing.B) {
		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
			Async: &interfaces.AsyncConfig{
				Enabled:       true,
				BufferSize:    10000,
				Workers:       4,
				FlushInterval: 100 * time.Millisecond,
			},
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message",
					interfaces.String("key1", "value1"),
					interfaces.Int("key2", 42),
				)
			}
		})
	})
}

// BenchmarkProviders_Comparison compara performance entre providers
func BenchmarkProviders_Comparison(b *testing.B) {
	ctx := context.Background()

	b.Run("Zap", func(b *testing.B) {
		provider := zap.NewProvider()
		err := provider.Configure(interfaces.Config{
			Level:       interfaces.InfoLevel,
			Format:      interfaces.JSONFormat,
			ServiceName: "benchmark-test",
			Environment: "test",
		})
		if err != nil {
			b.Fatalf("Failed to configure provider: %v", err)
		}

		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message",
					interfaces.String("provider", "zap"),
					interfaces.Int("iteration", b.N),
				)
			}
		})
	})

	b.Run("Zerolog", func(b *testing.B) {
		provider := zerolog.NewProvider()
		err := provider.Configure(interfaces.Config{
			Level:       interfaces.InfoLevel,
			Format:      interfaces.JSONFormat,
			ServiceName: "benchmark-test",
			Environment: "test",
		})
		if err != nil {
			b.Fatalf("Failed to configure provider: %v", err)
		}

		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message",
					interfaces.String("provider", "zerolog"),
					interfaces.Int("iteration", b.N),
				)
			}
		})
	})

	b.Run("Slog", func(b *testing.B) {
		provider := slog.NewProvider()
		err := provider.Configure(interfaces.Config{
			Level:       interfaces.InfoLevel,
			Format:      interfaces.JSONFormat,
			ServiceName: "benchmark-test",
			Environment: "test",
		})
		if err != nil {
			b.Fatalf("Failed to configure provider: %v", err)
		}

		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message",
					interfaces.String("provider", "slog"),
					interfaces.Int("iteration", b.N),
				)
			}
		})
	})
}

// BenchmarkSampling_Performance testa performance do sampling
func BenchmarkSampling_Performance(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	b.Run("NoSampling", func(b *testing.B) {
		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
			// Sem sampling
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message")
			}
		})
	})

	b.Run("SamplingEnabled", func(b *testing.B) {
		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
			Sampling: &interfaces.SamplingConfig{
				Enabled:    true,
				Initial:    10,
				Thereafter: 100,
				Tick:       time.Second,
				Levels: []interfaces.Level{
					interfaces.InfoLevel,
					interfaces.DebugLevel,
				},
			},
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message")
			}
		})
	})

	b.Run("SamplingAggressive", func(b *testing.B) {
		config := interfaces.Config{
			Level:       interfaces.InfoLevel,
			ServiceName: "benchmark-test",
			Environment: "test",
			Sampling: &interfaces.SamplingConfig{
				Enabled:    true,
				Initial:    1,
				Thereafter: 1000,
				Tick:       time.Second,
				Levels: []interfaces.Level{
					interfaces.InfoLevel,
					interfaces.DebugLevel,
				},
			},
		}

		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark message")
			}
		})
	})
}

// BenchmarkMemoryAllocation testa alocações de memória
func BenchmarkMemoryAllocation(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		ServiceName: "benchmark-test",
		Environment: "test",
	}

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	b.Run("SimpleMessage", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, "simple message")
		}
	})

	b.Run("WithFields", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, "message with fields",
				interfaces.String("key1", "value1"),
				interfaces.Int("key2", 42),
				interfaces.Bool("key3", true),
			)
		}
	})

	b.Run("WithManyFields", func(b *testing.B) {
		fields := make([]interfaces.Field, 20)
		for i := 0; i < 20; i++ {
			fields[i] = interfaces.String(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, "message with many fields", fields...)
		}
	})

	b.Run("WithContext", func(b *testing.B) {
		ctxWithValues := context.WithValue(context.Background(), "trace_id", "trace-123")
		ctxWithValues = context.WithValue(ctxWithValues, "span_id", "span-456")
		ctxWithValues = context.WithValue(ctxWithValues, "user_id", "user-789")

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctxWithValues, "message with context",
				interfaces.String("key1", "value1"),
			)
		}
	})
}

// BenchmarkLevelFiltering testa performance do filtro de nível
func BenchmarkLevelFiltering(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.WarnLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	config := interfaces.Config{
		Level:       interfaces.WarnLevel,
		ServiceName: "benchmark-test",
		Environment: "test",
	}

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	b.Run("FilteredDebug", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Essas mensagens devem ser filtradas rapidamente
				logger.Debug(ctx, "debug message that will be filtered")
			}
		})
	})

	b.Run("FilteredInfo", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Essas mensagens devem ser filtradas rapidamente
				logger.Info(ctx, "info message that will be filtered")
			}
		})
	})

	b.Run("AllowedWarn", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Essas mensagens passam pelo filtro
				logger.Warn(ctx, "warn message that will be logged")
			}
		})
	})
}

// BenchmarkWorkerScaling testa performance com diferentes números de workers
func BenchmarkWorkerScaling(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	for _, workers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			config := interfaces.Config{
				Level:       interfaces.InfoLevel,
				ServiceName: "benchmark-test",
				Environment: "test",
				Async: &interfaces.AsyncConfig{
					Enabled:       true,
					BufferSize:    10000,
					Workers:       workers,
					FlushInterval: 100 * time.Millisecond,
				},
			}

			logger := NewCoreLogger(provider, config)
			defer logger.Close()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(ctx, "benchmark message",
						interfaces.Int("workers", workers),
					)
				}
			})
		})
	}
}

// BenchmarkBufferSizes testa performance com diferentes tamanhos de buffer
func BenchmarkBufferSizes(b *testing.B) {
	provider := zap.NewProvider()
	err := provider.Configure(interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "benchmark-test",
		Environment: "test",
	})
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	for _, bufferSize := range []int{100, 1000, 5000, 10000, 50000} {
		b.Run(fmt.Sprintf("Buffer_%d", bufferSize), func(b *testing.B) {
			config := interfaces.Config{
				Level:       interfaces.InfoLevel,
				ServiceName: "benchmark-test",
				Environment: "test",
				Async: &interfaces.AsyncConfig{
					Enabled:       true,
					BufferSize:    bufferSize,
					Workers:       4,
					FlushInterval: 100 * time.Millisecond,
				},
			}

			logger := NewCoreLogger(provider, config)
			defer logger.Close()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(ctx, "benchmark message",
						interfaces.Int("buffer_size", bufferSize),
					)
				}
			})
		})
	}
}
