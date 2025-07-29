// Package main demonstra o uso e comparação entre providers valkey-go e valkey-glide.
// Este exemplo mostra como alternar entre providers e suas diferenças de performance.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
	valkeyglide "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-glide"
	valkeygo "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-go"
)

func main() {
	fmt.Println("=== Demonstração de Compatibilidade entre Providers ===")
	fmt.Println()

	// Configuração base (mesmo para ambos os providers)
	baseConfig := &config.Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		ClusterMode:  false,
		PoolSize:     10,
		MinIdleConns: 1,
		MaxIdleConns: 3,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// Testar ambos os providers
	providers := []struct {
		Name     string
		Provider interfaces.IProvider
		Config   *config.Config
	}{
		{
			Name:     "valkey-go",
			Provider: valkeygo.NewProvider(),
			Config: func() *config.Config {
				cfg := *baseConfig
				cfg.Provider = "valkey-go"
				return &cfg
			}(),
		},
		{
			Name:     "valkey-glide",
			Provider: valkeyglide.NewProvider(),
			Config: func() *config.Config {
				cfg := *baseConfig
				cfg.Provider = "valkey-glide"
				return &cfg
			}(),
		},
	}

	ctx := context.Background()

	for _, p := range providers {
		fmt.Printf("--- Testando Provider: %s ---\n", p.Name)

		// Criar cliente
		client, err := p.Provider.NewClient(p.Config)
		if err != nil {
			fmt.Printf("❌ Erro ao criar cliente %s: %v\n", p.Name, err)
			fmt.Printf("   (Certifique-se de que o Valkey está rodando em localhost:6379)\n")
			fmt.Println()
			continue
		}

		// Testar conectividade
		if err := client.Ping(ctx); err != nil {
			fmt.Printf("❌ Erro ao conectar com %s: %v\n", p.Name, err)
			client.Close()
			fmt.Println()
			continue
		}

		fmt.Printf("✅ Conectado com sucesso usando %s\n", p.Name)

		// Testes básicos
		testBasicOperations(ctx, client, p.Name)
		testHashOperations(ctx, client, p.Name)
		testListOperations(ctx, client, p.Name)
		testSetOperations(ctx, client, p.Name)
		testSortedSetOperations(ctx, client, p.Name)

		// Performance básica
		measurePerformance(ctx, client, p.Name)

		client.Close()
		fmt.Println()
	}

	fmt.Println("=== Demonstração Concluída ===")
}

func testBasicOperations(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  📝 Testando operações básicas (%s)...\n", providerName)

	key := fmt.Sprintf("test:basic:%s:%d", providerName, time.Now().UnixNano())

	// SET/GET
	if err := client.Set(ctx, key, "hello world", 0); err != nil {
		fmt.Printf("    ❌ SET falhou: %v\n", err)
		return
	}

	value, err := client.Get(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ GET falhou: %v\n", err)
		return
	}

	if value != "hello world" {
		fmt.Printf("    ❌ Valor incorreto: esperado 'hello world', obtido '%s'\n", value)
		return
	}

	// TTL
	if err := client.Expire(ctx, key, 10*time.Second); err != nil {
		fmt.Printf("    ⚠️ EXPIRE falhou: %v\n", err)
	} else {
		ttl, err := client.TTL(ctx, key)
		if err != nil {
			fmt.Printf("    ⚠️ TTL falhou: %v\n", err)
		} else {
			fmt.Printf("    ✅ TTL definido: %v\n", ttl)
		}
	}

	// DELETE
	count, err := client.Del(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ DEL falhou: %v\n", err)
		return
	}

	fmt.Printf("    ✅ Operações básicas OK (%d chaves deletadas)\n", count)
}

func testHashOperations(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  🗂️ Testando operações de Hash (%s)...\n", providerName)

	key := fmt.Sprintf("test:hash:%s:%d", providerName, time.Now().UnixNano())

	if err := client.HSet(ctx, key, "field1", "value1", "field2", "value2"); err != nil {
		fmt.Printf("    ❌ HSET falhou: %v\n", err)
		return
	}

	value, err := client.HGet(ctx, key, "field1")
	if err != nil {
		fmt.Printf("    ❌ HGET falhou: %v\n", err)
		return
	}

	if value != "value1" {
		fmt.Printf("    ❌ Valor incorreto: esperado 'value1', obtido '%s'\n", value)
		return
	}

	allFields, err := client.HGetAll(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ HGETALL falhou: %v\n", err)
		return
	}

	if len(allFields) != 2 {
		fmt.Printf("    ❌ Número incorreto de campos: esperado 2, obtido %d\n", len(allFields))
		return
	}

	client.Del(ctx, key)
	fmt.Printf("    ✅ Operações de Hash OK (%d campos)\n", len(allFields))
}

func testListOperations(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  📝 Testando operações de Lista (%s)...\n", providerName)

	key := fmt.Sprintf("test:list:%s:%d", providerName, time.Now().UnixNano())

	count, err := client.LPush(ctx, key, "item1", "item2")
	if err != nil {
		fmt.Printf("    ❌ LPUSH falhou: %v\n", err)
		return
	}

	length, err := client.LLen(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ LLEN falhou: %v\n", err)
		return
	}

	if length != count {
		fmt.Printf("    ❌ Tamanho incorreto: esperado %d, obtido %d\n", count, length)
		return
	}

	item, err := client.LPop(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ LPOP falhou: %v\n", err)
		return
	}

	if item != "item2" {
		fmt.Printf("    ❌ Item incorreto: esperado 'item2', obtido '%s'\n", item)
		return
	}

	client.Del(ctx, key)
	fmt.Printf("    ✅ Operações de Lista OK (último item: %s)\n", item)
}

func testSetOperations(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  🎲 Testando operações de Set (%s)...\n", providerName)

	key := fmt.Sprintf("test:set:%s:%d", providerName, time.Now().UnixNano())

	count, err := client.SAdd(ctx, key, "member1", "member2", "member3")
	if err != nil {
		fmt.Printf("    ❌ SADD falhou: %v\n", err)
		return
	}

	isMember, err := client.SIsMember(ctx, key, "member1")
	if err != nil {
		fmt.Printf("    ❌ SISMEMBER falhou: %v\n", err)
		return
	}

	if !isMember {
		fmt.Printf("    ❌ Membro não encontrado: member1\n")
		return
	}

	members, err := client.SMembers(ctx, key)
	if err != nil {
		fmt.Printf("    ❌ SMEMBERS falhou: %v\n", err)
		return
	}

	if len(members) != 3 {
		fmt.Printf("    ❌ Número incorreto de membros: esperado 3, obtido %d\n", len(members))
		return
	}

	client.Del(ctx, key)
	fmt.Printf("    ✅ Operações de Set OK (%d membros adicionados)\n", count)
}

func testSortedSetOperations(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  🏆 Testando operações de Sorted Set (%s)...\n", providerName)

	key := fmt.Sprintf("test:zset:%s:%d", providerName, time.Now().UnixNano())

	count, err := client.ZAdd(ctx, key, 1.0, "member1", 2.0, "member2", 3.0, "member3")
	if err != nil {
		fmt.Printf("    ❌ ZADD falhou: %v\n", err)
		return
	}

	score, err := client.ZScore(ctx, key, "member2")
	if err != nil {
		fmt.Printf("    ❌ ZSCORE falhou: %v\n", err)
		return
	}

	if score != 2.0 {
		fmt.Printf("    ❌ Score incorreto: esperado 2.0, obtido %f\n", score)
		return
	}

	members, err := client.ZRange(ctx, key, 0, -1)
	if err != nil {
		fmt.Printf("    ❌ ZRANGE falhou: %v\n", err)
		return
	}

	if len(members) != 3 {
		fmt.Printf("    ❌ Número incorreto de membros: esperado 3, obtido %d\n", len(members))
		return
	}

	client.Del(ctx, key)
	fmt.Printf("    ✅ Operações de Sorted Set OK (%d membros adicionados)\n", count)
}

func measurePerformance(ctx context.Context, client interfaces.IClient, providerName string) {
	fmt.Printf("  ⚡ Medindo performance básica (%s)...\n", providerName)

	operations := 100
	key := fmt.Sprintf("perf:test:%s", providerName)

	// Medir SET operations
	start := time.Now()
	for i := 0; i < operations; i++ {
		client.Set(ctx, fmt.Sprintf("%s:%d", key, i), fmt.Sprintf("value%d", i), 0)
	}
	setDuration := time.Since(start)

	// Medir GET operations
	start = time.Now()
	for i := 0; i < operations; i++ {
		client.Get(ctx, fmt.Sprintf("%s:%d", key, i))
	}
	getDuration := time.Since(start)

	// Limpeza
	for i := 0; i < operations; i++ {
		client.Del(ctx, fmt.Sprintf("%s:%d", key, i))
	}

	setOpsPerSec := float64(operations) / setDuration.Seconds()
	getOpsPerSec := float64(operations) / getDuration.Seconds()

	fmt.Printf("    📊 Performance (%d ops):\n", operations)
	fmt.Printf("       SET: %.0f ops/sec (%.2fms total)\n", setOpsPerSec, setDuration.Seconds()*1000)
	fmt.Printf("       GET: %.0f ops/sec (%.2fms total)\n", getOpsPerSec, getDuration.Seconds()*1000)
}
