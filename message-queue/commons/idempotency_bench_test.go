package commons

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func BenchmarkIdempotencyManager_IsProcessed(b *testing.B) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	// Pré-popular com algumas chaves
	for i := 0; i < 1000; i++ {
		manager.MarkAsProcessed(ctx, fmt.Sprintf("key-%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			manager.IsProcessed(ctx, key)
			i++
		}
	})
}

func BenchmarkIdempotencyManager_MarkAsProcessed(b *testing.B) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			manager.MarkAsProcessed(ctx, key)
			i++
		}
	})
}

func BenchmarkIdempotencyManager_Mixed(b *testing.B) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)

			// 80% read, 20% write
			if i%5 == 0 {
				manager.MarkAsProcessed(ctx, key)
			} else {
				manager.IsProcessed(ctx, key)
			}
			i++
		}
	})
}

func BenchmarkIdempotencyManager_ConcurrentAccess(b *testing.B) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%100) // Reutilizar chaves para simular concorrência

			if i%2 == 0 {
				manager.MarkAsProcessed(ctx, key)
			} else {
				manager.IsProcessed(ctx, key)
			}
			i++
		}
	})
}
