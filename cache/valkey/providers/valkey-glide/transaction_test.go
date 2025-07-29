package valkeyglide

import (
	"context"
	"testing"
	"time"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
)

func TestTransaction_NewTransaction(t *testing.T) {
	config := &valkeyconfig.Config{
		Host: "localhost",
		Port: 6379,
	}

	client := &Client{
		config: config,
	}

	txInterface := newTransaction(client)

	if txInterface == nil {
		t.Error("Transaction should not be nil")
	}

	// Cast para acessar campos internos para teste
	tx, ok := txInterface.(*transaction)
	if !ok {
		t.Error("Transaction should be of concrete type *transaction")
	}

	if len(tx.commands) != 0 {
		t.Error("Transaction commands should be empty initially")
	}
}

func TestTransaction_Get(t *testing.T) {
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
			txInterface := newTransaction(client)

			cmd := txInterface.Get(tt.key)
			if cmd == nil {
				t.Error("Get command should not be nil")
			}

			// Cast para verificar comando foi adicionado
			tx := txInterface.(*transaction)
			if len(tx.commands) != 1 {
				t.Errorf("Transaction should have 1 command, got %d", len(tx.commands))
			}
		})
	}
}

func TestTransaction_Set(t *testing.T) {
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
			txInterface := newTransaction(client)

			cmd := txInterface.Set(tt.key, tt.value, tt.expiration)
			if cmd == nil {
				t.Error("Set command should not be nil")
			}

			tx := txInterface.(*transaction)
			if len(tx.commands) != 1 {
				t.Errorf("Transaction should have 1 command, got %d", len(tx.commands))
			}
		})
	}
}

func TestTransaction_Del(t *testing.T) {
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
			txInterface := newTransaction(client)

			cmd := txInterface.Del(tt.keys...)
			if cmd == nil {
				t.Error("Del command should not be nil")
			}

			tx := txInterface.(*transaction)
			if len(tx.commands) != 1 {
				t.Errorf("Transaction should have 1 command, got %d", len(tx.commands))
			}
		})
	}
}

func TestTransaction_HGet(t *testing.T) {
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
			txInterface := newTransaction(client)

			cmd := txInterface.HGet(tt.key, tt.field)
			if cmd == nil {
				t.Error("HGet command should not be nil")
			}

			tx := txInterface.(*transaction)
			if len(tx.commands) != 1 {
				t.Errorf("Transaction should have 1 command, got %d", len(tx.commands))
			}
		})
	}
}

func TestTransaction_HSet(t *testing.T) {
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
			txInterface := newTransaction(client)

			cmd := txInterface.HSet(tt.key, tt.values...)
			if cmd == nil {
				t.Error("HSet command should not be nil")
			}

			tx := txInterface.(*transaction)
			if len(tx.commands) != 1 {
				t.Errorf("Transaction should have 1 command, got %d", len(tx.commands))
			}
		})
	}
}

func TestTransaction_Watch(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{
			name: "watch single key",
			keys: []string{"key1"},
		},
		{
			name: "watch multiple keys",
			keys: []string{"key1", "key2", "key3"},
		},
		{
			name: "watch no keys",
			keys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &valkeyconfig.Config{},
			}
			txInterface := newTransaction(client)

			ctx := context.Background()

			// Em um teste real, testaria a funcionalidade do WATCH
			// Por enquanto, apenas testamos a interface
			err := txInterface.Watch(ctx, tt.keys...)

			// Como não temos client real, esperamos que não cause panic
			// Em implementação real, testaria se o WATCH foi executado
			_ = err // Para evitar warning unused
		})
	}
}

func TestTransaction_Unwatch(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	ctx := context.Background()

	// Em um teste real, testaria a funcionalidade do UNWATCH
	err := txInterface.Unwatch(ctx)

	// Como não temos client real, apenas testamos que não causa panic
	_ = err // Para evitar warning unused
}

func TestTransaction_MultipleCommands(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	// Adicionar múltiplos comandos
	txInterface.Set("key1", "value1", 0)
	txInterface.Get("key2")
	txInterface.Del("key3")
	txInterface.HSet("hash1", "field1", "value1")
	txInterface.HGet("hash2", "field2")

	tx := txInterface.(*transaction)
	expectedCommands := 5
	if len(tx.commands) != expectedCommands {
		t.Errorf("Transaction should have %d commands, got %d", expectedCommands, len(tx.commands))
	}
}

func TestTransaction_Exec(t *testing.T) {
	tests := []struct {
		name        string
		setupTx     func(*transaction)
		wantErr     bool
		expectCount int
	}{
		{
			name: "exec empty transaction",
			setupTx: func(tx *transaction) {
				// No commands
			},
			wantErr:     false,
			expectCount: 0,
		},
		{
			name: "exec transaction with commands",
			setupTx: func(tx *transaction) {
				tx.Set("key1", "value1", 0)
				tx.Get("key2")
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
			txInterface := newTransaction(client)
			tx := txInterface.(*transaction)

			tt.setupTx(tx)

			ctx := context.Background()

			// Como estamos testando sem client real, apenas verificamos estrutura
			if len(tx.commands) != tt.expectCount {
				t.Errorf("Expected %d commands, got %d", tt.expectCount, len(tx.commands))
			}

			// Em um teste real com mock, executaríamos:
			// results, err := txInterface.Exec(ctx)
			// if (err != nil) != tt.wantErr {
			//     t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			// }
			_ = ctx // Evita warning unused
		})
	}
}

func TestTransaction_Discard(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	// Adicionar alguns comandos
	txInterface.Set("key1", "value1", 0)
	txInterface.Get("key2")

	tx := txInterface.(*transaction)
	if len(tx.commands) != 2 {
		t.Errorf("Transaction should have 2 commands before discard")
	}

	err := txInterface.Discard()
	if err != nil {
		t.Errorf("Discard() error = %v, want nil", err)
	}

	if len(tx.commands) != 0 {
		t.Errorf("Transaction should have 0 commands after discard, got %d", len(tx.commands))
	}
}

func TestTransaction_ChainedCommands(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	// Teste de comandos encadeados
	cmd1 := txInterface.Set("key1", "value1", 0)
	cmd2 := txInterface.Get("key1")
	cmd3 := txInterface.Del("key1")

	if cmd1 == nil || cmd2 == nil || cmd3 == nil {
		t.Error("All chained commands should not be nil")
	}

	tx := txInterface.(*transaction)
	if len(tx.commands) != 3 {
		t.Errorf("Transaction should have 3 commands, got %d", len(tx.commands))
	}
}

func TestTransaction_AtomicBehavior(t *testing.T) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	ctx := context.Background()

	// Simular comportamento transacional
	txInterface.Set("account:1", "100", 0)
	txInterface.Set("account:2", "50", 0)

	tx := txInterface.(*transaction)
	if len(tx.commands) != 2 {
		t.Errorf("Transaction should have 2 commands")
	}

	// Em um teste real, verificaríamos que todos os comandos são executados atomicamente
	// ou todos falham juntos
	_ = ctx
}

// Benchmark tests para Transaction
func BenchmarkTransaction_Set(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txInterface.Set("benchmark_key", "value", 0)
	}
}

func BenchmarkTransaction_Get(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txInterface.Get("benchmark_key")
	}
}

func BenchmarkTransaction_HSet(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txInterface.HSet("hash_key", "field", "value")
	}
}

func BenchmarkTransaction_MixedCommands(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txInterface := newTransaction(client)
		txInterface.Set("key1", "value1", 0)
		txInterface.Get("key2")
		txInterface.HSet("hash1", "field1", "value1")
		txInterface.Del("key3")
	}
}

func BenchmarkTransaction_Watch(b *testing.B) {
	client := &Client{
		config: &valkeyconfig.Config{},
	}
	txInterface := newTransaction(client)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txInterface.Watch(ctx, "watch_key")
	}
}
