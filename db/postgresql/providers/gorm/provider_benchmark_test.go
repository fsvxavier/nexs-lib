package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm/mocks"
	"github.com/golang/mock/gomock"
)

// BenchmarkProvider_Operations benchmarks provider operations
func BenchmarkProvider_Operations(b *testing.B) {
	ctx := context.Background()
	provider := NewProvider()

	b.Run("NewProvider", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewProvider()
		}
	})

	b.Run("Provider_Metadata", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.Type()
			_ = provider.Name()
			_ = provider.Version()
		}
	})

	b.Run("Provider_Health_Check", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.IsHealthy(ctx)
		}
	})

	b.Run("Provider_Get_Metrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.GetMetrics(ctx)
		}
	})
}

// BenchmarkProvider_WithMocks benchmarks operations with mocks
func BenchmarkProvider_WithMocks(b *testing.B) {
	ctx := context.Background()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	b.Run("Mock_Connection_Query", func(b *testing.B) {
		mockConn := mocks.NewMockIConn(ctrl)
		mockRows := mocks.NewMockIRows(ctrl)

		// Setup expectations for benchmark iterations
		mockConn.EXPECT().Query(ctx, "SELECT * FROM users", gomock.Any()).Return(mockRows, nil).Times(b.N)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = mockConn.Query(ctx, "SELECT * FROM users")
		}
	})

	b.Run("Mock_Connection_Exec", func(b *testing.B) {
		mockConn := mocks.NewMockIConn(ctrl)

		// Setup expectations for benchmark iterations
		mockConn.EXPECT().Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test").Return(nil).Times(b.N)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mockConn.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		}
	})

	b.Run("Mock_Transaction_Operations", func(b *testing.B) {
		mockTx := mocks.NewMockITransaction(ctrl)

		// Setup expectations for benchmark iterations
		mockTx.EXPECT().Exec(ctx, "UPDATE users SET updated_at = NOW()", gomock.Any()).Return(nil).Times(b.N)
		mockTx.EXPECT().Commit(ctx).Return(nil).Times(b.N)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mockTx.Exec(ctx, "UPDATE users SET updated_at = NOW()")
			_ = mockTx.Commit(ctx)
		}
	})
}

// BenchmarkProvider_ConfigValidation benchmarks config validation
func BenchmarkProvider_ConfigValidation(b *testing.B) {
	ctx := context.Background()
	provider := NewProvider()

	validConfig := &config.Config{
		Host:               "localhost",
		Port:               5432,
		Database:           "testdb",
		Username:           "testuser",
		Password:           "testpass",
		MaxConns:           10,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 30,
		QueryTimeout:       time.Second * 30,
		ApplicationName:    "test-app",
		SearchPath:         []string{"public"},
		Timezone:           "UTC",
		DefaultSchema:      "public",
		MultiTenantEnabled: false,
	}

	b.Run("Config_Validation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// This will fail connection but will exercise validation logic
			_, _ = provider.CreatePool(ctx, validConfig)
		}
	})

	b.Run("Nil_Config_Validation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = provider.CreatePool(ctx, nil)
		}
	})
}
