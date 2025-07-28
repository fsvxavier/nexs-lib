package pagination

import (
	"context"
	"net/url"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

func TestPaginationServiceWithJSONSchema(t *testing.T) {
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
		{
			name: "Invalid limit parameter",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"0"},
			},
			sortFields:  []string{"id"},
			expectError: true,
		},
		{
			name: "Invalid sort field",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
				"sort":  []string{"invalid_field"},
			},
			sortFields:  []string{"id", "name"},
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

func TestPaginationServiceHooks(t *testing.T) {
	cfg := config.NewDefaultConfig()
	service := NewPaginationService(cfg)

	// Test hook execution
	hookExecuted := false
	testHook := &BasicHook{
		name: "test-hook",
		executor: func(ctx context.Context, data interface{}) error {
			hookExecuted = true
			return nil
		},
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

func TestQueryBuilderPool(t *testing.T) {
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

func TestLazyValidator(t *testing.T) {
	cfg := config.NewDefaultConfig()
	validator := &LazyValidator{
		config:          cfg,
		loadedFields:    make(map[string]bool),
		validationRules: make(map[string]interface{}),
		jsonValidator:   nil,
	}

	fields := []string{"id", "name", "created_at"}

	// Test loading
	if validator.IsLoaded(fields) {
		t.Error("Fields should not be loaded initially")
	}

	err := validator.LoadValidator(fields)
	if err != nil {
		t.Fatalf("Failed to load validator: %v", err)
	}

	if !validator.IsLoaded(fields) {
		t.Error("Fields should be loaded after LoadValidator")
	}

	// Test validation
	params := &interfaces.PaginationParams{
		Page:      1,
		Limit:     10,
		SortField: "id",
		SortOrder: "ASC",
	}

	err = validator.ValidateParams(params, fields)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}
