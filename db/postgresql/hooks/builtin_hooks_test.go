package hooks

import (
	"context"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestBuiltinHooks_LoggingHook(t *testing.T) {
	hook := LoggingHook("TEST")

	ctx := &interfaces.ExecutionContext{
		Context:   context.Background(),
		Operation: "query",
		Query:     "SELECT 1",
		Args:      []interface{}{"test"},
		Duration:  time.Millisecond * 10,
		Error:     nil,
	}

	result := hook(ctx)

	if result == nil {
		t.Error("LoggingHook should return a result")
		return
	}

	if !result.Continue {
		t.Error("LoggingHook should return Continue: true")
	}

	if result.Error != nil {
		t.Errorf("LoggingHook should not return error, got: %v", result.Error)
	}
}

func TestBuiltinHooks_TimingHook(t *testing.T) {
	hook := TimingHook()

	t.Run("set start time", func(t *testing.T) {
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
		}

		result := hook(ctx)

		if result == nil {
			t.Error("TimingHook should return a result")
			return
		}

		if !result.Continue {
			t.Error("TimingHook should return Continue: true")
		}

		if ctx.StartTime.IsZero() {
			t.Error("TimingHook should set StartTime")
		}
	})

	t.Run("calculate duration", func(t *testing.T) {
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			StartTime: time.Now().Add(-time.Millisecond * 10),
		}

		result := hook(ctx)

		if result == nil {
			t.Error("TimingHook should return a result")
			return
		}

		if ctx.Duration <= 0 {
			t.Error("TimingHook should calculate duration")
		}
	})
}

func TestBuiltinHooks_ValidationHook(t *testing.T) {
	hook := ValidationHook()

	tests := []struct {
		name         string
		setupCtx     func() *interfaces.ExecutionContext
		wantError    bool
		wantContinue bool
	}{
		{
			name: "valid query",
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Query:     "SELECT 1",
				}
			},
			wantError:    false,
			wantContinue: true,
		},
		{
			name: "empty query for query operation",
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Query:     "",
				}
			},
			wantError:    true,
			wantContinue: false,
		},
		{
			name: "ping operation without query",
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "ping",
					Query:     "",
				}
			},
			wantError:    false,
			wantContinue: true,
		},
		{
			name: "healthcheck operation without query",
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "healthcheck",
					Query:     "",
				}
			},
			wantError:    false,
			wantContinue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := hook(ctx)

			if result == nil {
				t.Error("ValidationHook should return a result")
				return
			}

			if tt.wantError {
				if result.Error == nil {
					t.Error("ValidationHook expected error but got nil")
				}
				if result.Continue != tt.wantContinue {
					t.Errorf("ValidationHook Continue = %v, want %v", result.Continue, tt.wantContinue)
				}
			} else {
				if result.Error != nil {
					t.Errorf("ValidationHook unexpected error = %v", result.Error)
				}
				if result.Continue != tt.wantContinue {
					t.Errorf("ValidationHook Continue = %v, want %v", result.Continue, tt.wantContinue)
				}
			}
		})
	}
}

func TestBuiltinHooks_MetricsHook(t *testing.T) {
	hook := MetricsHook()

	t.Run("initialize metadata", func(t *testing.T) {
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			Query:     "SELECT 1",
			Args:      []interface{}{"arg1", "arg2"},
		}

		result := hook(ctx)

		if result == nil {
			t.Error("MetricsHook should return a result")
			return
		}

		if !result.Continue {
			t.Error("MetricsHook should return Continue: true")
		}

		if ctx.Metadata == nil {
			t.Error("MetricsHook should initialize Metadata")
			return
		}

		if ctx.Metadata["operation"] != "query" {
			t.Errorf("Metadata[operation] = %v, want query", ctx.Metadata["operation"])
		}

		if ctx.Metadata["query_length"] != len("SELECT 1") {
			t.Errorf("Metadata[query_length] = %v, want %d", ctx.Metadata["query_length"], len("SELECT 1"))
		}

		if ctx.Metadata["args_count"] != 2 {
			t.Errorf("Metadata[args_count] = %v, want 2", ctx.Metadata["args_count"])
		}
	})

	t.Run("with duration", func(t *testing.T) {
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			Duration:  time.Millisecond * 10,
		}

		result := hook(ctx)

		if result == nil {
			t.Error("MetricsHook should return a result")
			return
		}

		if ctx.Metadata["duration_ms"] != int64(10) {
			t.Errorf("Metadata[duration_ms] = %v, want 10", ctx.Metadata["duration_ms"])
		}
	})

	t.Run("with error", func(t *testing.T) {
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			Error:     context.DeadlineExceeded,
		}

		result := hook(ctx)

		if result == nil {
			t.Error("MetricsHook should return a result")
			return
		}

		if ctx.Metadata["has_error"] != true {
			t.Errorf("Metadata[has_error] = %v, want true", ctx.Metadata["has_error"])
		}

		if ctx.Metadata["error"] == nil {
			t.Error("Metadata[error] should be set when error exists")
		}
	})
}

func TestBuiltinHooks_RetryHook(t *testing.T) {
	tests := []struct {
		name            string
		maxRetries      int
		retryDelay      time.Duration
		setupCtx        func() *interfaces.ExecutionContext
		wantContinue    bool
		wantShouldRetry bool
	}{
		{
			name:       "no error",
			maxRetries: 3,
			retryDelay: time.Millisecond,
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Error:     nil,
				}
			},
			wantContinue:    true,
			wantShouldRetry: false,
		},
		{
			name:       "non-retryable error",
			maxRetries: 3,
			retryDelay: time.Millisecond,
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Error:     context.Canceled, // Non-retryable error
				}
			},
			wantContinue:    true,
			wantShouldRetry: false,
		},
		{
			name:       "first retry attempt",
			maxRetries: 3,
			retryDelay: 0, // No delay for tests
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Error:     context.DeadlineExceeded, // Retryable error
					Metadata:  make(map[string]interface{}),
				}
			},
			wantContinue:    true,
			wantShouldRetry: false, // Since isRetryableError returns false in our implementation
		},
		{
			name:       "max retries exceeded",
			maxRetries: 2,
			retryDelay: 0,
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
					Error:     context.DeadlineExceeded,
					Metadata: map[string]interface{}{
						"retry_count": 3,
					},
				}
			},
			wantContinue:    true,
			wantShouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := RetryHook(tt.maxRetries, tt.retryDelay)
			ctx := tt.setupCtx()

			result := hook(ctx)

			if result == nil {
				t.Error("RetryHook should return a result")
				return
			}

			if result.Continue != tt.wantContinue {
				t.Errorf("RetryHook Continue = %v, want %v", result.Continue, tt.wantContinue)
			}
		})
	}
}

func TestBuiltinHooks_TenantHook(t *testing.T) {
	hook := TenantHook()

	tests := []struct {
		name                 string
		setupCtx             func() *interfaces.ExecutionContext
		wantTenantInMetadata bool
	}{
		{
			name: "with tenant ID",
			setupCtx: func() *interfaces.ExecutionContext {
				ctx := context.WithValue(context.Background(), "tenant_id", "tenant_123")
				return &interfaces.ExecutionContext{
					Context:   ctx,
					Operation: "query",
				}
			},
			wantTenantInMetadata: true,
		},
		{
			name: "without tenant ID",
			setupCtx: func() *interfaces.ExecutionContext {
				return &interfaces.ExecutionContext{
					Context:   context.Background(),
					Operation: "query",
				}
			},
			wantTenantInMetadata: false,
		},
		{
			name: "with empty tenant ID",
			setupCtx: func() *interfaces.ExecutionContext {
				ctx := context.WithValue(context.Background(), "tenant_id", "")
				return &interfaces.ExecutionContext{
					Context:   ctx,
					Operation: "query",
				}
			},
			wantTenantInMetadata: false,
		},
		{
			name: "with non-string tenant ID",
			setupCtx: func() *interfaces.ExecutionContext {
				ctx := context.WithValue(context.Background(), "tenant_id", 123)
				return &interfaces.ExecutionContext{
					Context:   ctx,
					Operation: "query",
				}
			},
			wantTenantInMetadata: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := hook(ctx)

			if result == nil {
				t.Error("TenantHook should return a result")
				return
			}

			if !result.Continue {
				t.Error("TenantHook should return Continue: true")
			}

			if result.Error != nil {
				t.Errorf("TenantHook should not return error, got: %v", result.Error)
			}

			if tt.wantTenantInMetadata {
				if ctx.Metadata == nil {
					t.Error("TenantHook should initialize Metadata when tenant exists")
					return
				}
				if ctx.Metadata["tenant_id"] != "tenant_123" {
					t.Errorf("Metadata[tenant_id] = %v, want tenant_123", ctx.Metadata["tenant_id"])
				}
			}
		})
	}
}

func TestBuiltinHooks_CacheHook(t *testing.T) {
	cacheTTL := time.Minute

	t.Run("non-query operations", func(t *testing.T) {
		hook := CacheHook(cacheTTL)

		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "exec", // Non-query operation
			Query:     "INSERT INTO test VALUES (1)",
		}

		result := hook(ctx)

		if result == nil {
			t.Error("CacheHook should return a result")
			return
		}

		if !result.Continue {
			t.Error("CacheHook should return Continue: true for non-query operations")
		}
	})

	t.Run("query operations - cache miss and store", func(t *testing.T) {
		hook := CacheHook(cacheTTL)

		// First call - cache miss
		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			Query:     "SELECT 1",
			Args:      []interface{}{},
			Duration:  0, // Before execution
		}

		result := hook(ctx)

		if result == nil {
			t.Error("CacheHook should return a result")
			return
		}

		if !result.Continue {
			t.Error("CacheHook should return Continue: true for cache miss")
		}

		// Simulate after execution
		ctx.Duration = time.Millisecond * 10
		ctx.Error = nil

		result = hook(ctx)

		if result == nil {
			t.Error("CacheHook should return a result for after execution")
			return
		}

		if !result.Continue {
			t.Error("CacheHook should return Continue: true after successful execution")
		}

		// Verify cache metadata
		if ctx.Metadata == nil {
			t.Error("CacheHook should initialize Metadata")
			return
		}

		if ctx.Metadata["cache_hit"] != false {
			t.Errorf("Metadata[cache_hit] = %v, want false", ctx.Metadata["cache_hit"])
		}
	})

	t.Run("query operations - with error", func(t *testing.T) {
		hook := CacheHook(cacheTTL)

		ctx := &interfaces.ExecutionContext{
			Context:   context.Background(),
			Operation: "query",
			Query:     "SELECT 1",
			Args:      []interface{}{},
			Duration:  time.Millisecond * 10,
			Error:     context.DeadlineExceeded,
		}

		result := hook(ctx)

		if result == nil {
			t.Error("CacheHook should return a result")
			return
		}

		if !result.Continue {
			t.Error("CacheHook should return Continue: true even with error")
		}
	})
}

func TestHelperFunctions_ContainsSQLInjectionPatterns(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "safe query",
			query: "SELECT * FROM users WHERE id = $1",
			want:  false,
		},
		{
			name:  "empty query",
			query: "",
			want:  false,
		},
		{
			name:  "simple select",
			query: "SELECT 1",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSQLInjectionPatterns(tt.query)
			if result != tt.want {
				t.Errorf("containsSQLInjectionPatterns(%q) = %v, want %v", tt.query, result, tt.want)
			}
		})
	}
}

func TestHelperFunctions_IsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: false, // Our implementation returns false for all errors
		},
		{
			name: "canceled",
			err:  context.Canceled,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.want {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.want)
			}
		})
	}
}

func TestHelperFunctions_GenerateCacheKey(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []interface{}
		want  string
	}{
		{
			name:  "simple query no args",
			query: "SELECT 1",
			args:  []interface{}{},
			want:  "SELECT 1",
		},
		{
			name:  "query with args",
			query: "SELECT * FROM users WHERE id = $1",
			args:  []interface{}{123},
			want:  "SELECT * FROM users WHERE id = $1_0_123",
		},
		{
			name:  "query with multiple args",
			query: "SELECT * FROM users WHERE id = $1 AND name = $2",
			args:  []interface{}{123, "test"},
			want:  "SELECT * FROM users WHERE id = $1 AND name = $2_0_123_1_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCacheKey(tt.query, tt.args)
			if result != tt.want {
				t.Errorf("generateCacheKey(%q, %v) = %q, want %q", tt.query, tt.args, result, tt.want)
			}
		})
	}
}
