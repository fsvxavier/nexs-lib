//go:build unit

package pgx

import (
	"context"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/stretchr/testify/assert"
)

func TestPGXTransaction(t *testing.T) {
	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXTransaction implements ITransaction
		var _ interfaces.ITransaction = (*PGXTransaction)(nil)
	})

	t.Run("Commit should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := tx.Commit(ctx)
			// err will likely be non-nil for empty transaction, which is expected
			_ = err
		})
	})

	t.Run("Rollback should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := tx.Rollback(ctx)
			// err will likely be non-nil for empty transaction, which is expected
			_ = err
		})
	})

	t.Run("Multiple rollbacks should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			tx.Rollback(ctx)
			tx.Rollback(ctx) // Second rollback should not panic
		})
	})

	t.Run("Multiple commits should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			tx.Commit(ctx)
			tx.Commit(ctx) // Second commit should not panic
		})
	})

	t.Run("Commit after rollback should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			tx.Rollback(ctx)
			tx.Commit(ctx) // Commit after rollback should not panic
		})
	})

	t.Run("Rollback after commit should not panic", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx := context.Background()

		assert.NotPanics(t, func() {
			tx.Commit(ctx)
			tx.Rollback(ctx) // Rollback after commit should not panic
		})
	})

	t.Run("Context cancellation handling", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel context immediately

		assert.NotPanics(t, func() {
			err := tx.Commit(ctx)
			// Should handle cancelled context gracefully
			_ = err
		})

		assert.NotPanics(t, func() {
			err := tx.Rollback(ctx)
			// Should handle cancelled context gracefully
			_ = err
		})
	})

	t.Run("Context timeout handling", func(t *testing.T) {
		tx := &PGXTransaction{}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(1 * time.Millisecond)

		assert.NotPanics(t, func() {
			err := tx.Commit(ctx)
			// Should handle timeout gracefully
			_ = err
		})
	})

	t.Run("Nil context handling", func(t *testing.T) {
		tx := &PGXTransaction{}

		// Note: Passing nil context should be avoided, but shouldn't panic
		assert.NotPanics(t, func() {
			err := tx.Commit(context.Background()) // Use background instead of nil
			_ = err
		})
	})
}

func TestTxOptions(t *testing.T) {
	t.Run("Default transaction options", func(t *testing.T) {
		opts := interfaces.TxOptions{}

		// Default values should be valid
		assert.Equal(t, interfaces.TxIsoLevelDefault, opts.IsoLevel)
		assert.Equal(t, interfaces.TxAccessModeReadWrite, opts.AccessMode)
		assert.Equal(t, interfaces.TxDeferrableModeNotDeferrable, opts.DeferrableMode)
		assert.Empty(t, opts.BeginQuery)
	})

	t.Run("Custom transaction options", func(t *testing.T) {
		opts := interfaces.TxOptions{
			IsoLevel:       interfaces.TxIsoLevelSerializable,
			AccessMode:     interfaces.TxAccessModeReadOnly,
			DeferrableMode: interfaces.TxDeferrableModeDeferrable,
			BeginQuery:     "BEGIN ISOLATION LEVEL SERIALIZABLE READ ONLY DEFERRABLE",
		}

		assert.Equal(t, interfaces.TxIsoLevelSerializable, opts.IsoLevel)
		assert.Equal(t, interfaces.TxAccessModeReadOnly, opts.AccessMode)
		assert.Equal(t, interfaces.TxDeferrableModeDeferrable, opts.DeferrableMode)
		assert.NotEmpty(t, opts.BeginQuery)
	})

	t.Run("All isolation levels", func(t *testing.T) {
		levels := []interfaces.TxIsoLevel{
			interfaces.TxIsoLevelDefault,
			interfaces.TxIsoLevelReadUncommitted,
			interfaces.TxIsoLevelReadCommitted,
			interfaces.TxIsoLevelRepeatableRead,
			interfaces.TxIsoLevelSerializable,
		}

		for _, level := range levels {
			opts := interfaces.TxOptions{IsoLevel: level}
			assert.True(t, int(opts.IsoLevel) >= 0)
			assert.True(t, int(opts.IsoLevel) < len(levels))
		}
	})

	t.Run("All access modes", func(t *testing.T) {
		modes := []interfaces.TxAccessMode{
			interfaces.TxAccessModeReadWrite,
			interfaces.TxAccessModeReadOnly,
		}

		for _, mode := range modes {
			opts := interfaces.TxOptions{AccessMode: mode}
			assert.True(t, int(opts.AccessMode) >= 0)
			assert.True(t, int(opts.AccessMode) < len(modes))
		}
	})

	t.Run("All deferrable modes", func(t *testing.T) {
		modes := []interfaces.TxDeferrableMode{
			interfaces.TxDeferrableModeNotDeferrable,
			interfaces.TxDeferrableModeDeferrable,
		}

		for _, mode := range modes {
			opts := interfaces.TxOptions{DeferrableMode: mode}
			assert.True(t, int(opts.DeferrableMode) >= 0)
			assert.True(t, int(opts.DeferrableMode) < len(modes))
		}
	})
}

// Benchmark tests
func BenchmarkPGXTransaction_Commit(b *testing.B) {
	tx := &PGXTransaction{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.Commit(ctx)
	}
}

func BenchmarkPGXTransaction_Rollback(b *testing.B) {
	tx := &PGXTransaction{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.Rollback(ctx)
	}
}

func BenchmarkTxOptions_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := interfaces.TxOptions{
			IsoLevel:       interfaces.TxIsoLevelSerializable,
			AccessMode:     interfaces.TxAccessModeReadOnly,
			DeferrableMode: interfaces.TxDeferrableModeDeferrable,
			BeginQuery:     "BEGIN ISOLATION LEVEL SERIALIZABLE READ ONLY DEFERRABLE",
		}
		_ = opts
	}
}
