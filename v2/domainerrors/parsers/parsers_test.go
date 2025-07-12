package parsers

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func TestBasicParsersCreation(t *testing.T) {
	parsers := []interface{}{
		NewPostgreSQLErrorParser(),
		NewMySQLErrorParser(),
		NewNetworkErrorParser(),
		NewTimeoutErrorParser(),
		NewSQLErrorParser(),
		NewGRPCErrorParser(),
		NewHTTPErrorParser(),
		NewRedisErrorParser(),
		NewMongoDBErrorParser(),
		NewAWSErrorParser(),
		NewPGXErrorParser(),
		NewEnhancedPostgreSQLErrorParser(),
	}

	for i, parser := range parsers {
		if parser == nil {
			t.Errorf("Parser %d should not be nil", i)
		}
	}
}

func TestDistributedRegistryCreation(t *testing.T) {
	registry := NewDistributedParserRegistry()
	if registry == nil {
		t.Error("Registry should not be nil")
	}

	global := GetGlobalParserRegistry()
	if global == nil {
		t.Error("Global registry should not be nil")
	}
}

func TestPostgreSQLParser(t *testing.T) {
	parser := NewPostgreSQLErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"postgres error", fmt.Errorf("pq: duplicate key"), true},
		{"SQLSTATE error", fmt.Errorf("ERROR: test (SQLSTATE 23505)"), true},
		{"postgres connection", fmt.Errorf("postgres: connection failed"), true},
		{"non-postgres", fmt.Errorf("mysql: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestMySQLParser(t *testing.T) {
	parser := NewMySQLErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"mysql error", fmt.Errorf("Error 1062: Duplicate entry"), true},
		{"mysql connection", fmt.Errorf("mysql: connection lost"), true},
		{"non-mysql", fmt.Errorf("postgres: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestNetworkParser(t *testing.T) {
	parser := NewNetworkErrorParser()

	netErr := &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}
	dnsErr := &net.DNSError{Err: "no such host", Name: "invalid.domain"}
	urlErr := &url.Error{Op: "Get", URL: "http://example.com", Err: fmt.Errorf("connection refused")}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"net OpError", netErr, true},
		{"DNS error", dnsErr, true},
		{"URL error", urlErr, true},
		{"non-network", fmt.Errorf("database error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeNetwork) {
					t.Errorf("Expected network type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestTimeoutParser(t *testing.T) {
	parser := NewTimeoutErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"context deadline", context.DeadlineExceeded, true},
		{"timeout string", fmt.Errorf("operation timeout"), true},
		{"deadline exceeded", fmt.Errorf("deadline exceeded"), true},
		{"connection timeout", fmt.Errorf("connection timeout"), true},
		{"non-timeout", fmt.Errorf("validation error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeTimeout) {
					t.Errorf("Expected timeout type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestSQLParser(t *testing.T) {
	parser := NewSQLErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"sql.ErrNoRows", sql.ErrNoRows, true},
		{"sql.ErrTxDone", sql.ErrTxDone, true},
		{"sql.ErrConnDone", sql.ErrConnDone, true},
		{"sql string", fmt.Errorf("sql: syntax error"), true},
		{"database string", fmt.Errorf("database connection failed"), true},
		{"query string", fmt.Errorf("query timeout"), true},
		{"transaction string", fmt.Errorf("transaction failed"), true},
		{"non-sql", fmt.Errorf("validation error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestGRPCParser(t *testing.T) {
	parser := NewGRPCErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"grpc string", fmt.Errorf("grpc: connection failed"), true},
		{"rpc error", fmt.Errorf("rpc error: code = NotFound"), true},
		{"status code", fmt.Errorf("status code = Unavailable"), true},
		{"non-grpc", fmt.Errorf("database error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeNetwork) {
					t.Errorf("Expected network type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestHTTPParser(t *testing.T) {
	parser := NewHTTPErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"http string", fmt.Errorf("http: client error"), true},
		{"HTTP status", fmt.Errorf("HTTP 404: Not Found"), true},
		{"status code", fmt.Errorf("status: 500"), true},
		{"non-http", fmt.Errorf("database error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeNetwork) {
					t.Errorf("Expected network type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestRedisParser(t *testing.T) {
	parser := NewRedisErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"redis string", fmt.Errorf("redis: connection refused"), true},
		{"WRONGTYPE", fmt.Errorf("WRONGTYPE Operation against wrong type"), true},
		{"ERR", fmt.Errorf("ERR invalid command"), true},
		{"redigo", fmt.Errorf("redigo: connection failed"), true},
		{"non-redis", fmt.Errorf("postgres: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestMongoDBParser(t *testing.T) {
	parser := NewMongoDBErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"mongo string", fmt.Errorf("mongo: connection failed"), true},
		{"E11000", fmt.Errorf("E11000 duplicate key error (11000)"), true},
		{"bson error", fmt.Errorf("bson: invalid format"), true},
		{"collection", fmt.Errorf("collection not found"), true},
		{"non-mongo", fmt.Errorf("postgres: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestAWSParser(t *testing.T) {
	parser := NewAWSErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"aws string", fmt.Errorf("aws: service error"), true},
		{"s3 error", fmt.Errorf("s3: bucket not found"), true},
		{"dynamodb", fmt.Errorf("dynamodb: item not found"), true},
		{"throttling", fmt.Errorf("throttling: rate exceeded"), true},
		{"lambda", fmt.Errorf("lambda: function timeout"), true},
		{"non-aws", fmt.Errorf("postgres: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeCloud) {
					t.Errorf("Expected cloud type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestPGXParser(t *testing.T) {
	parser := NewPGXErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"pgx string", fmt.Errorf("pgx: connection failed"), true},
		{"pgxpool", fmt.Errorf("pgxpool: all connections busy"), true},
		{"non-pgx", fmt.Errorf("mysql: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestEnhancedPostgreSQLParser(t *testing.T) {
	parser := NewEnhancedPostgreSQLErrorParser()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"postgres error", fmt.Errorf("pq: duplicate key DETAIL: Key already exists"), true},
		{"complex SQLSTATE", fmt.Errorf("ERROR: foreign key constraint (SQLSTATE 23503) DETAIL: Referenced row not found"), true},
		{"non-postgres", fmt.Errorf("mysql: error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.CanParse(tt.err); got != tt.want {
				t.Errorf("CanParse() = %v, want %v", got, tt.want)
			}

			if tt.want && tt.err != nil {
				parsed := parser.Parse(tt.err)
				if parsed.Type != string(types.ErrorTypeDatabase) {
					t.Errorf("Expected database type, got %v", parsed.Type)
				}
			}
		})
	}
}

func TestRegistryOperations(t *testing.T) {
	registry := NewDistributedParserRegistry()

	t.Run("register parser", func(t *testing.T) {
		err := registry.RegisterParser("test", NewPostgreSQLErrorParser(), 100)
		if err != nil {
			t.Errorf("RegisterParser failed: %v", err)
		}
	})

	t.Run("list parsers", func(t *testing.T) {
		parsers := registry.ListParsers()
		if len(parsers) == 0 {
			t.Error("Should have parsers registered")
		}
	})

	t.Run("disable parser", func(t *testing.T) {
		err := registry.DisableParser("test")
		if err != nil {
			t.Errorf("DisableParser failed: %v", err)
		}
	})

	t.Run("enable parser", func(t *testing.T) {
		err := registry.EnableParser("test")
		if err != nil {
			t.Errorf("EnableParser failed: %v", err)
		}
	})

	t.Run("set priority", func(t *testing.T) {
		err := registry.SetPriority("test", 200)
		if err != nil {
			t.Errorf("SetPriority failed: %v", err)
		}
	})

	t.Run("parse with registry", func(t *testing.T) {
		ctx := context.Background()
		testErr := fmt.Errorf("pq: duplicate key")

		parsed, parserName, err := registry.Parse(ctx, testErr)
		if err != nil {
			t.Errorf("Parse failed: %v", err)
		}
		if parsed.Code == "" && parsed.Message == "" {
			t.Error("Parsed result should have content")
		}
		if parserName == "" {
			t.Error("Parser name should not be empty")
		}
	})

	t.Run("get metrics", func(t *testing.T) {
		metrics, err := registry.GetMetrics("postgresql")
		if err != nil {
			t.Errorf("GetMetrics failed: %v", err)
		}
		if metrics == nil {
			t.Error("Metrics should not be nil")
		}
	})
}

func TestParserErrorHandling(t *testing.T) {
	registry := NewDistributedParserRegistry()

	t.Run("register parser with empty name", func(t *testing.T) {
		err := registry.RegisterParser("", NewPostgreSQLErrorParser(), 100)
		if err == nil {
			t.Error("Should fail with empty name")
		}
	})

	t.Run("disable non-existent parser", func(t *testing.T) {
		err := registry.DisableParser("non-existent")
		if err == nil {
			t.Error("Should fail with non-existent parser")
		}
	})

	t.Run("parse with nil error", func(t *testing.T) {
		ctx := context.Background()
		_, _, err := registry.Parse(ctx, nil)
		if err == nil {
			t.Error("Should fail with nil error")
		}
	})

	t.Run("get metrics for non-existent parser", func(t *testing.T) {
		_, err := registry.GetMetrics("non-existent")
		if err == nil {
			t.Error("Should fail for non-existent parser")
		}
	})
}

func TestGlobalRegistry(t *testing.T) {
	ctx := context.Background()
	testErr := fmt.Errorf("pq: duplicate key")

	parsed, parserName, err := ParseWithGlobalRegistry(ctx, testErr)
	if err != nil {
		t.Errorf("ParseWithGlobalRegistry failed: %v", err)
	}
	if parsed.Code == "" && parsed.Message == "" {
		t.Error("Parsed result should have content")
	}
	if parserName == "" {
		t.Error("Parser name should not be empty")
	}
}

// Testes adicionais para edge cases e concorrência
func TestEdgeCases(t *testing.T) {
	t.Run("empty error messages", func(t *testing.T) {
		parser := NewPostgreSQLErrorParser()
		emptyErr := fmt.Errorf("")

		if parser.CanParse(emptyErr) {
			t.Error("Should not parse empty error")
		}
	})

	t.Run("very long error messages", func(t *testing.T) {
		longMsg := make([]byte, 10000)
		for i := range longMsg {
			longMsg[i] = 'a'
		}

		longErr := fmt.Errorf("postgres error: %s", string(longMsg))
		parser := NewPostgreSQLErrorParser()

		if !parser.CanParse(longErr) {
			t.Error("Should parse long error messages")
		}

		parsed := parser.Parse(longErr)
		if parsed.Code == "" && parsed.Message == "" {
			t.Error("Should return parsed result")
		}
	})

	t.Run("nested error structures", func(t *testing.T) {
		innerErr := fmt.Errorf("pq: connection lost")
		wrappedErr := fmt.Errorf("database operation failed: %w", innerErr)

		parser := NewPostgreSQLErrorParser()
		if !parser.CanParse(wrappedErr) {
			t.Error("Should parse wrapped errors")
		}
	})
}

func TestConcurrentParsing(t *testing.T) {
	registry := NewDistributedParserRegistry()
	ctx := context.Background()

	// Registrar alguns parsers
	registry.RegisterParser("postgresql", NewPostgreSQLErrorParser(), 100)
	registry.RegisterParser("mysql", NewMySQLErrorParser(), 90)
	registry.RegisterParser("network", NewNetworkErrorParser(), 80)

	errors := []error{
		fmt.Errorf("pq: duplicate key"),
		fmt.Errorf("mysql: connection lost"),
		fmt.Errorf("network: timeout"),
		fmt.Errorf("redis: connection refused"),
	}

	done := make(chan bool, len(errors))

	for _, err := range errors {
		go func(testErr error) {
			defer func() { done <- true }()

			parsed, parserName, parseErr := registry.Parse(ctx, testErr)
			if parseErr != nil {
				t.Errorf("Parse failed: %v", parseErr)
				return
			}

			if parsed.Code == "" && parsed.Message == "" {
				t.Error("Parsed result should have content")
				return
			}

			if parserName == "" {
				t.Error("Parser name should not be empty")
			}
		}(err)
	}

	// Aguardar todas as goroutines terminarem
	for i := 0; i < len(errors); i++ {
		<-done
	}
}
