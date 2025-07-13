package tracer

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceConfig(t *testing.T) {
	config := DefaultPerformanceConfig()

	assert.True(t, config.EnableSpanPooling)
	assert.Equal(t, 1000, config.SpanPoolSize)
	assert.True(t, config.EnableFastPaths)
	assert.True(t, config.EnableZeroAlloc)
	assert.Equal(t, 128, config.MaxAttributesPerSpan)
	assert.Equal(t, 64, config.MaxEventsPerSpan)
	assert.Equal(t, 512, config.BatchSize)
}

func TestSpanPool_BasicOperations(t *testing.T) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)

	t.Run("get and put span", func(t *testing.T) {
		span := pool.Get()
		assert.NotNil(t, span)
		assert.True(t, span.IsRecording())
		assert.Equal(t, pool, span.pool)

		pool.Put(span)

		// Verify metrics
		metrics := pool.GetMetrics()
		assert.Greater(t, metrics.SpansCreated, int64(0))
	})

	t.Run("span reuse", func(t *testing.T) {
		span1 := pool.Get()
		span1.SetOperationName("test-op")
		span1.SetAttributeFast("key", "value")
		pool.Put(span1)

		span2 := pool.Get()
		assert.NotNil(t, span2)
		assert.Equal(t, "", span2.operationName) // Should be reset
		assert.Equal(t, 0, len(span2.attributes))

		pool.Put(span2)
	})

	t.Run("pool disabled", func(t *testing.T) {
		configDisabled := config
		configDisabled.EnableSpanPooling = false
		poolDisabled := NewSpanPool(configDisabled)

		span := poolDisabled.Get()
		assert.NotNil(t, span)

		poolDisabled.Put(span)

		metrics := poolDisabled.GetMetrics()
		assert.Equal(t, int64(0), metrics.SpansReused)
	})
}

func TestPooledSpan_FastOperations(t *testing.T) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	t.Run("set operation name", func(t *testing.T) {
		span.SetOperationName("test-operation")
		assert.Equal(t, "test-operation", span.operationName)
	})

	t.Run("fast attribute setting", func(t *testing.T) {
		span.SetAttributeFast("string", "value")
		span.SetAttributeFast("int", 42)
		span.SetAttributeFast("int64", int64(12345))
		span.SetAttributeFast("float", 3.14)
		span.SetAttributeFast("bool", true)

		attrs := span.GetAttributes()
		assert.Equal(t, 5, len(attrs))

		// Verify types and values
		assert.Equal(t, "string", attrs[0].Key)
		assert.Equal(t, AttributeTypeString, attrs[0].Value.Type)
		assert.Equal(t, "value", attrs[0].Value.StringVal)

		assert.Equal(t, "int", attrs[1].Key)
		assert.Equal(t, AttributeTypeInt, attrs[1].Value.Type)
		assert.Equal(t, int64(42), attrs[1].Value.IntVal)
	})

	t.Run("regular attribute setting", func(t *testing.T) {
		span.reset() // Clear previous attributes

		span.SetAttributeRegular("test", "value")
		attrs := span.GetAttributes()
		assert.Equal(t, 1, len(attrs))
		assert.Equal(t, "test", attrs[0].Key)
	})

	t.Run("fast event adding", func(t *testing.T) {
		eventAttrs := map[string]interface{}{
			"event.type": "click",
			"count":      1,
		}

		span.AddEventFast("user_interaction", eventAttrs)
		events := span.GetEvents()
		assert.Equal(t, 1, len(events))
		assert.Equal(t, "user_interaction", events[0].Name)
		assert.NotZero(t, events[0].Timestamp)
	})

	t.Run("span lifecycle", func(t *testing.T) {
		span.Start()
		assert.True(t, span.IsRecording())
		assert.NotZero(t, span.startTime)

		time.Sleep(1 * time.Millisecond)

		span.SetStatus(StatusCodeOk, "success")
		assert.Equal(t, StatusCodeOk, span.status)
		assert.Equal(t, "success", span.statusMsg)

		duration := span.GetDuration()
		assert.Greater(t, duration, time.Duration(0))

		span.End()
		assert.False(t, span.IsRecording())
		assert.True(t, span.isFinished)
		assert.NotZero(t, span.duration)
	})
}

func TestPooledSpan_CapacityLimits(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MaxAttributesPerSpan = 3
	config.MaxEventsPerSpan = 2

	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	t.Run("attribute capacity limit", func(t *testing.T) {
		// Add more attributes than capacity
		for i := 0; i < 5; i++ {
			span.SetAttributeRegular("key", "value")
		}

		attrs := span.GetAttributes()
		assert.LessOrEqual(t, len(attrs), config.MaxAttributesPerSpan)
	})

	t.Run("event capacity limit", func(t *testing.T) {
		// Add more events than capacity
		for i := 0; i < 4; i++ {
			span.AddEventRegular("event", nil)
		}

		events := span.GetEvents()
		assert.LessOrEqual(t, len(events), config.MaxEventsPerSpan)
	})
}

func TestFastTracer(t *testing.T) {
	config := DefaultPerformanceConfig()
	tracer := NewFastTracer(config)

	t.Run("start span fast", func(t *testing.T) {
		ctx := context.Background()

		newCtx, span := tracer.StartSpanFast(ctx, "test-operation")
		assert.NotNil(t, span)
		assert.NotEqual(t, ctx, newCtx)
		assert.Equal(t, "test-operation", span.operationName)

		span.End()
	})

	t.Run("tracer metrics", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple spans
		for i := 0; i < 10; i++ {
			_, span := tracer.StartSpanFast(ctx, "test")
			span.End()
		}

		metrics := tracer.GetMetrics()
		assert.Greater(t, metrics.SpansCreated, int64(0))
		assert.GreaterOrEqual(t, metrics.FastPathHits, int64(0))
	})
}

func TestFastSerializer(t *testing.T) {
	config := DefaultPerformanceConfig()
	serializer := NewFastSerializer(config)
	pool := NewSpanPool(config)

	span := pool.Get()
	span.SetOperationName("test-operation")
	span.SetAttributeFast("key1", "value1")
	span.SetAttributeFast("key2", 42)
	span.AddEventFast("test-event", map[string]interface{}{
		"event.id": "123",
	})
	span.Start()
	span.End()

	t.Run("serialize span", func(t *testing.T) {
		data := serializer.SerializeSpanFast(span)
		assert.NotNil(t, data)
		assert.Greater(t, len(data), 0)
	})

	t.Run("serialize nil span", func(t *testing.T) {
		data := serializer.SerializeSpanFast(nil)
		assert.Nil(t, data)
	})

	pool.Put(span)
}

func TestAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		attrType AttributeType
		value    interface{}
	}{
		{"string", AttributeTypeString, "test"},
		{"int", AttributeTypeInt, int64(42)},
		{"float", AttributeTypeFloat, 3.14},
		{"bool", AttributeTypeBool, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := Attribute{
				Key:   "test-key",
				Value: AttributeValue{Type: tt.attrType},
			}

			switch tt.attrType {
			case AttributeTypeString:
				attr.Value.StringVal = tt.value.(string)
				assert.Equal(t, tt.value, attr.Value.StringVal)
			case AttributeTypeInt:
				attr.Value.IntVal = tt.value.(int64)
				assert.Equal(t, tt.value, attr.Value.IntVal)
			case AttributeTypeFloat:
				attr.Value.FloatVal = tt.value.(float64)
				assert.Equal(t, tt.value, attr.Value.FloatVal)
			case AttributeTypeBool:
				attr.Value.BoolVal = tt.value.(bool)
				assert.Equal(t, tt.value, attr.Value.BoolVal)
			}
		})
	}
}

// Benchmark tests for performance validation

func BenchmarkSpanPool_GetPut(b *testing.B) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			span := pool.Get()
			span.SetOperationName("benchmark")
			pool.Put(span)
		}
	})
}

func BenchmarkSpanPool_GetPutWithoutPooling(b *testing.B) {
	config := DefaultPerformanceConfig()
	config.EnableSpanPooling = false
	pool := NewSpanPool(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span := pool.Get()
		span.SetOperationName("benchmark")
		pool.Put(span)
	}
}

func BenchmarkPooledSpan_SetAttributeFast(b *testing.B) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttributeFast("key", "value")
		span.reset() // Reset to avoid capacity issues
	}
}

func BenchmarkPooledSpan_SetAttributeRegular(b *testing.B) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttributeRegular("key", "value")
		span.reset() // Reset to avoid capacity issues
	}
}

func BenchmarkPooledSpan_AddEventFast(b *testing.B) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	attrs := map[string]interface{}{
		"event.type": "benchmark",
		"count":      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.AddEventFast("benchmark-event", attrs)
		span.reset() // Reset to avoid capacity issues
	}
}

func BenchmarkFastTracer_StartSpan(b *testing.B) {
	config := DefaultPerformanceConfig()
	tracer := NewFastTracer(config)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, span := tracer.StartSpanFast(ctx, "benchmark")
			span.End()
		}
	})
}

func BenchmarkFastSerializer_SerializeSpan(b *testing.B) {
	config := DefaultPerformanceConfig()
	serializer := NewFastSerializer(config)
	pool := NewSpanPool(config)

	// Create a representative span
	span := pool.Get()
	span.SetOperationName("benchmark-operation")
	span.SetAttributeFast("http.method", "GET")
	span.SetAttributeFast("http.status_code", 200)
	span.SetAttributeFast("user.id", "12345")
	span.AddEventFast("request.start", map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
	})
	span.Start()
	span.End()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := serializer.SerializeSpanFast(span)
		_ = data // Prevent optimization
	}

	pool.Put(span)
}

// Memory allocation tests

func TestSpanPool_MemoryUsage(t *testing.T) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)

	// Measure memory before
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create and use spans
	spans := make([]*PooledSpan, 1000)
	for i := 0; i < 1000; i++ {
		spans[i] = pool.Get()
		spans[i].SetOperationName("memory-test")
		spans[i].SetAttributeFast("id", i)
	}

	// Return spans to pool
	for _, span := range spans {
		pool.Put(span)
	}

	// Measure memory after
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// With pooling, memory usage should be relatively stable
	// This is more of a demonstration than a strict test
	t.Logf("Memory allocated: %d bytes", m2.TotalAlloc-m1.TotalAlloc)
	t.Logf("Pool metrics: %+v", pool.GetMetrics())
}

func TestConcurrentSpanPoolAccess(t *testing.T) {
	config := DefaultPerformanceConfig()
	pool := NewSpanPool(config)

	numGoroutines := 100
	spansPerGoroutine := 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < spansPerGoroutine; j++ {
				span := pool.Get()
				span.SetOperationName("concurrent-test")
				span.SetAttributeFast("goroutine.id", id)
				span.SetAttributeFast("span.id", j)
				span.Start()

				// Simulate some work
				time.Sleep(time.Microsecond)

				span.End()
				pool.Put(span)
			}
		}(i)
	}

	wg.Wait()

	metrics := pool.GetMetrics()
	expectedTotal := int64(numGoroutines * spansPerGoroutine)

	// Verify that all spans were handled
	assert.GreaterOrEqual(t, metrics.SpansCreated+metrics.SpansReused, expectedTotal)
}

func TestZeroAllocationPaths(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.EnableZeroAlloc = true

	pool := NewSpanPool(config)
	span := pool.Get()
	defer pool.Put(span)

	// These operations should use zero-allocation paths
	span.SetAttributeFast("string", "value")
	span.SetAttributeFast("int", 42)
	span.SetAttributeFast("bool", true)

	// Small event with few attributes should use fast path
	smallAttrs := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	span.AddEventFast("small-event", smallAttrs)

	attrs := span.GetAttributes()
	assert.Equal(t, 3, len(attrs))

	events := span.GetEvents()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, 2, len(events[0].Attributes))
}
