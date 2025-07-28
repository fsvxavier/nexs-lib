package pagination

import (
	"context"
	"net/url"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

func TestEnhancedPaginationService(t *testing.T) {
	cfg := config.NewDefaultConfig()
	service := NewPaginationService(cfg)

	tests := []struct {
		name        string
		params      url.Values
		sortFields  []string
		expectError bool
	}{
		{
			name: "Valid pagination parameters",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
				"sort":  []string{"id"},
				"order": []string{"ASC"},
			},
			sortFields:  []string{"id", "name"},
			expectError: false,
		},
		{
			name: "Invalid page parameter",
			params: url.Values{
				"page":  []string{"0"},
				"limit": []string{"10"},
			},
			sortFields:  []string{"id"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ParseRequest(tt.params, tt.sortFields...)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestHooksSupport(t *testing.T) {
	cfg := config.NewDefaultConfig()
	service := NewPaginationService(cfg)

	// Test adding hooks
	hookExecuted := false
	testHook := &testHook{
		executed: &hookExecuted,
	}

	service.AddHook("pre-validation", testHook)

	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"10"},
	}

	_, err := service.ParseRequest(params)
	if err != nil {
		t.Fatalf("ParseRequest failed: %v", err)
	}

	if !hookExecuted {
		t.Error("Hook was not executed")
	}
}

func TestQueryBuilderPoolSupport(t *testing.T) {
	pool := NewDefaultQueryBuilderPool(5)

	// Test pool operations
	builder1 := pool.Get()
	if builder1 == nil {
		t.Error("Failed to get builder from pool")
	}

	builder2 := pool.Get()
	if builder2 == nil {
		t.Error("Failed to get second builder from pool")
	}

	// Test query building
	params := &interfaces.PaginationParams{
		Page:      1,
		Limit:     10,
		SortField: "id",
		SortOrder: "ASC",
	}

	query := builder1.BuildQuery("SELECT * FROM users", params)
	expected := "SELECT * FROM users ORDER BY id ASC LIMIT 10 OFFSET 0"
	if query != expected {
		t.Errorf("Expected query: %s, got: %s", expected, query)
	}

	// Return builders to pool
	pool.Put(builder1)
	pool.Put(builder2)

	// Test pool size
	if pool.Size() < 2 {
		t.Errorf("Expected pool size >= 2, got: %d", pool.Size())
	}
}

func TestPoolStats(t *testing.T) {
	cfg := config.NewDefaultConfig()
	service := NewPaginationService(cfg)

	// Test pool stats
	stats := service.GetPoolStats()
	if stats["enabled"] != true {
		t.Error("Pool should be enabled by default")
	}

	// Disable pool
	service.SetPoolEnabled(false)
	stats = service.GetPoolStats()
	if stats["enabled"] != false {
		t.Error("Pool should be disabled after SetPoolEnabled(false)")
	}
}

// testHook implements the Hook interface for testing
type testHook struct {
	executed *bool
}

func (h *testHook) Execute(ctx context.Context, data interface{}) error {
	*h.executed = true
	return nil
}
