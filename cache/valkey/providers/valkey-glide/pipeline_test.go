package valkeyglide

import (
	"context"
	"testing"
	"time"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
)

func TestPipeline_NewPipeline(t *testing.T) {
	config := &valkeyconfig.Config{
		Host: "localhost",
		Port: 6379,
	}

	client := &Client{
		config: config,
	}

	pipeInterface := newPipeline(client)

	if pipeInterface == nil {
		t.Error("Pipeline should not be nil")
	}

	// Cast para acessar campos internos para teste
	pipe, ok := pipeInterface.(*pipeline)
	if !ok {
		t.Error("Pipeline should be of concrete type *pipeline")
	}

	if len(pipe.commands) != 0 {
		t.Error("Pipeline commands should be empty initially")
	}
}

func TestPipeline_Get(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "get simple key",
			key:  "test_key",
		},
		{
			name: "get key with special chars",
			key:  "test:key:123",
		},
		{
			name: "get empty key",
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)

			cmd := pipeInterface.Get(tt.key)
			if cmd == nil {
				t.Error("Get command should not be nil")
			}

			// Cast para verificar comando foi adicionado
			pipe := pipeInterface.(*pipeline)
			if len(pipe.commands) != 1 {
				t.Errorf("Pipeline should have 1 command, got %d", len(pipe.commands))
			}
		})
	}
}

func TestPipeline_Set(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      interface{}
		expiration time.Duration
	}{
		{
			name:       "set string value",
			key:        "test_key",
			value:      "test_value",
			expiration: 0,
		},
		{
			name:       "set with expiration",
			key:        "test_key",
			value:      "test_value",
			expiration: time.Minute,
		},
		{
			name:       "set integer value",
			key:        "int_key",
			value:      42,
			expiration: 0,
		},
		{
			name:       "set bool value",
			key:        "bool_key",
			value:      true,
			expiration: 0,
		},
		{
			name:       "set float value",
			key:        "float_key",
			value:      3.14,
			expiration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)

			cmd := pipeInterface.Set(tt.key, tt.value, tt.expiration)
			if cmd == nil {
				t.Error("Set command should not be nil")
			}

			pipe := pipeInterface.(*pipeline)
			if len(pipe.commands) != 1 {
				t.Errorf("Pipeline should have 1 command, got %d", len(pipe.commands))
			}
		})
	}
}

func TestPipeline_Del(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{
			name: "delete single key",
			keys: []string{"key1"},
		},
		{
			name: "delete multiple keys",
			keys: []string{"key1", "key2", "key3"},
		},
		{
			name: "delete no keys",
			keys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)

			cmd := pipeInterface.Del(tt.keys...)
			if cmd == nil {
				t.Error("Del command should not be nil")
			}

			pipe := pipeInterface.(*pipeline)
			if len(pipe.commands) != 1 {
				t.Errorf("Pipeline should have 1 command, got %d", len(pipe.commands))
			}
		})
	}
}

func TestPipeline_HGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		field string
	}{
		{
			name:  "get hash field",
			key:   "hash_key",
			field: "field1",
		},
		{
			name:  "get hash field with special chars",
			key:   "hash:key:123",
			field: "field:1",
		},
		{
			name:  "get empty field",
			key:   "hash_key",
			field: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)

			cmd := pipeInterface.HGet(tt.key, tt.field)
			if cmd == nil {
				t.Error("HGet command should not be nil")
			}

			pipe := pipeInterface.(*pipeline)
			if len(pipe.commands) != 1 {
				t.Errorf("Pipeline should have 1 command, got %d", len(pipe.commands))
			}
		})
	}
}

func TestPipeline_HSet(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		values []interface{}
	}{
		{
			name:   "set single hash field",
			key:    "hash_key",
			values: []interface{}{"field1", "value1"},
		},
		{
			name:   "set multiple hash fields",
			key:    "hash_key",
			values: []interface{}{"field1", "value1", "field2", "value2"},
		},
		{
			name:   "set hash with mixed types",
			key:    "hash_key",
			values: []interface{}{"field1", "value1", "field2", 42, "field3", true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)

			cmd := pipeInterface.HSet(tt.key, tt.values...)
			if cmd == nil {
				t.Error("HSet command should not be nil")
			}

			pipe := pipeInterface.(*pipeline)
			if len(pipe.commands) != 1 {
				t.Errorf("Pipeline should have 1 command, got %d", len(pipe.commands))
			}
		})
	}
}

func TestPipeline_MultipleCommands(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	// Adicionar múltiplos comandos
	pipeInterface.Set("key1", "value1", 0)
	pipeInterface.Get("key2")
	pipeInterface.Del("key3")
	pipeInterface.HSet("hash1", "field1", "value1")
	pipeInterface.HGet("hash2", "field2")

	pipe := pipeInterface.(*pipeline)
	expectedCommands := 5
	if len(pipe.commands) != expectedCommands {
		t.Errorf("Pipeline should have %d commands, got %d", expectedCommands, len(pipe.commands))
	}
}

func TestPipeline_Exec(t *testing.T) {
	tests := []struct {
		name        string
		setupPipe   func(*pipeline)
		wantErr     bool
		expectCount int
	}{
		{
			name: "exec empty pipeline",
			setupPipe: func(p *pipeline) {
				// No commands
			},
			wantErr:     false,
			expectCount: 0,
		},
		{
			name: "exec pipeline with commands",
			setupPipe: func(p *pipeline) {
				p.Set("key1", "value1", 0)
				p.Get("key2")
			},
			wantErr:     false, // Mock test - não executa real
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			pipeInterface := newPipeline(client)
			pipe := pipeInterface.(*pipeline)

			tt.setupPipe(pipe)

			ctx := context.Background()

			// Como estamos testando sem client real, apenas verificamos estrutura
			if len(pipe.commands) != tt.expectCount {
				t.Errorf("Expected %d commands, got %d", tt.expectCount, len(pipe.commands))
			}

			// Em um teste real com mock, executaríamos:
			// results, err := pipeInterface.Exec(ctx)
			// if (err != nil) != tt.wantErr {
			//     t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			// }
			_ = ctx // Evita warning unused
		})
	}
}

func TestPipeline_Discard(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	// Adicionar alguns comandos
	pipeInterface.Set("key1", "value1", 0)
	pipeInterface.Get("key2")

	pipe := pipeInterface.(*pipeline)
	if len(pipe.commands) != 2 {
		t.Errorf("Pipeline should have 2 commands before discard")
	}

	err := pipeInterface.Discard()
	if err != nil {
		t.Errorf("Discard() error = %v, want nil", err)
	}

	if len(pipe.commands) != 0 {
		t.Errorf("Pipeline should have 0 commands after discard, got %d", len(pipe.commands))
	}
}

func TestPipeline_ChainedCommands(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	// Teste de comandos encadeados
	cmd1 := pipeInterface.Set("key1", "value1", 0)
	cmd2 := pipeInterface.Get("key1")
	cmd3 := pipeInterface.Del("key1")

	if cmd1 == nil || cmd2 == nil || cmd3 == nil {
		t.Error("All chained commands should not be nil")
	}

	pipe := pipeInterface.(*pipeline)
	if len(pipe.commands) != 3 {
		t.Errorf("Pipeline should have 3 commands, got %d", len(pipe.commands))
	}
}

// Benchmark tests para Pipeline
func BenchmarkPipeline_Set(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeInterface.Set("benchmark_key", "value", 0)
	}
}

func BenchmarkPipeline_Get(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeInterface.Get("benchmark_key")
	}
}

func BenchmarkPipeline_HSet(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	pipeInterface := newPipeline(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeInterface.HSet("hash_key", "field", "value")
	}
}

func BenchmarkPipeline_MixedCommands(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeInterface := newPipeline(client)
		pipeInterface.Set("key1", "value1", 0)
		pipeInterface.Get("key2")
		pipeInterface.HSet("hash1", "field1", "value1")
		pipeInterface.Del("key3")
	}
}
