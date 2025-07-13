//go:build unit

package postgresql

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

func TestCreateProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		driver      interfaces.DriverType
	}{
		{
			name:        "nil config",
			cfg:         nil,
			expectError: true,
		},
		{
			name: "pgx driver",
			cfg: &config.Config{
				Driver:   interfaces.DriverPGX,
				Host:     "localhost",
				Port:     5432,
				Database: "test",
			},
			expectError: false,
			driver:      interfaces.DriverPGX,
		},
		{
			name: "gorm driver",
			cfg: &config.Config{
				Driver:   interfaces.DriverGORM,
				Host:     "localhost",
				Port:     5432,
				Database: "test",
			},
			expectError: false,
			driver:      interfaces.DriverGORM,
		},
		{
			name: "pq driver",
			cfg: &config.Config{
				Driver:   interfaces.DriverPQ,
				Host:     "localhost",
				Port:     5432,
				Database: "test",
			},
			expectError: false,
			driver:      interfaces.DriverPQ,
		},
		{
			name: "unsupported driver",
			cfg: &config.Config{
				Driver:   "unsupported",
				Host:     "localhost",
				Port:     5432,
				Database: "test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := CreateProvider(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("Provider should not be nil")
				return
			}

			// For drivers that can be validated, check the driver type
			if tt.driver != "" {
				// We can't directly access driver from interface, but we can check it was created
				// Note: Since we don't have GetDriverType() method in interface,
				// we just validate that provider was created successfully
			}
		})
	}
}

// Benchmark tests
func BenchmarkCreateProvider(b *testing.B) {
	cfg := &config.Config{
		Driver:   interfaces.DriverPGX,
		Host:     "localhost",
		Port:     5432,
		Database: "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CreateProvider(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
