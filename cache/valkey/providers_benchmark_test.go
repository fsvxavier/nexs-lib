// Package valkey - Benchmarks comparativos entre providers
// Este arquivo implementa benchmarks abrangentes para comparar performance
// entre diferentes providers (valkey-go e valkey-glide).
package valkey

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
	valkeyglide "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-glide"
	valkeygo "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-go"
)

// setupBenchmarkClient cria um cliente para benchmarks.
func setupBenchmarkClient(b *testing.B, providerName string) interfaces.IClient {
	baseConfig := &config.Config{
		Provider:     providerName,
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

	var provider interfaces.IProvider
	switch providerName {
	case "valkey-go":
		provider = valkeygo.NewProvider()
	case "valkey-glide":
		provider = valkeyglide.NewProvider()
	default:
		b.Fatalf("Provider desconhecido: %s", providerName)
	}

	client, err := provider.NewClient(baseConfig)
	if err != nil {
		b.Skipf("Não foi possível criar cliente %s: %v", providerName, err)
	}

	// Verificar conectividade
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		b.Skipf("Servidor Valkey não disponível para %s: %v", providerName, err)
	}

	return client
}

// BenchmarkProviders_SetOperation compara performance de operações SET.
func BenchmarkProviders_SetOperation(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()
			value := "benchmark_value_123456789"

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("benchmark:set:%s:%d", providerName, i)
					if err := client.Set(ctx, key, value, 0); err != nil {
						b.Errorf("Erro no SET: %v", err)
					}
					i++
				}
			})
		})
	}
}

// BenchmarkProviders_GetOperation compara performance de operações GET.
func BenchmarkProviders_GetOperation(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()
			key := fmt.Sprintf("benchmark:get:%s", providerName)
			value := "benchmark_value_123456789"

			// Preparar dados
			if err := client.Set(ctx, key, value, 0); err != nil {
				b.Fatalf("Erro ao preparar dados: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if _, err := client.Get(ctx, key); err != nil {
						b.Errorf("Erro no GET: %v", err)
					}
				}
			})

			// Limpeza
			client.Del(ctx, key)
		})
	}
}

// BenchmarkProviders_SetGetCycle compara performance de ciclos SET/GET.
func BenchmarkProviders_SetGetCycle(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()
			value := "benchmark_cycle_value"

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("benchmark:cycle:%s:%d", providerName, i)

					// SET
					if err := client.Set(ctx, key, value, 0); err != nil {
						b.Errorf("Erro no SET: %v", err)
						continue
					}

					// GET
					if _, err := client.Get(ctx, key); err != nil {
						b.Errorf("Erro no GET: %v", err)
					}

					i++
				}
			})
		})
	}
}

// BenchmarkProviders_HashOperations compara performance de operações de hash.
func BenchmarkProviders_HashOperations(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					hashKey := fmt.Sprintf("benchmark:hash:%s:%d", providerName, i)
					field := "field1"
					value := "hash_value_123"

					// HSET
					if err := client.HSet(ctx, hashKey, field, value); err != nil {
						b.Errorf("Erro no HSET: %v", err)
						continue
					}

					// HGET
					if _, err := client.HGet(ctx, hashKey, field); err != nil {
						b.Errorf("Erro no HGET: %v", err)
					}

					i++
				}
			})
		})
	}
}

// BenchmarkProviders_ListOperations compara performance de operações de lista.
func BenchmarkProviders_ListOperations(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()
			value := "list_item_123"

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					listKey := fmt.Sprintf("benchmark:list:%s:%d", providerName, i)

					// LPUSH
					if _, err := client.LPush(ctx, listKey, value); err != nil {
						b.Errorf("Erro no LPUSH: %v", err)
						continue
					}

					// LPOP
					if _, err := client.LPop(ctx, listKey); err != nil {
						b.Errorf("Erro no LPOP: %v", err)
					}

					i++
				}
			})
		})
	}
}

// BenchmarkProviders_PingOperation compara performance de operações PING.
func BenchmarkProviders_PingOperation(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := client.Ping(ctx); err != nil {
					b.Errorf("Erro no PING: %v", err)
				}
			}
		})
	}
}

// BenchmarkProviders_MultipleOperations compara performance com múltiplas operações.
func BenchmarkProviders_MultipleOperations(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					baseKey := fmt.Sprintf("benchmark:multi:%s:%d", providerName, i)

					// String operations
					stringKey := baseKey + ":string"
					if err := client.Set(ctx, stringKey, "value", 0); err != nil {
						b.Errorf("Erro no SET string: %v", err)
						continue
					}

					// Hash operations
					hashKey := baseKey + ":hash"
					if err := client.HSet(ctx, hashKey, "field", "value"); err != nil {
						b.Errorf("Erro no HSET: %v", err)
						continue
					}

					// List operations
					listKey := baseKey + ":list"
					if _, err := client.LPush(ctx, listKey, "item"); err != nil {
						b.Errorf("Erro no LPUSH: %v", err)
						continue
					}

					// Set operations
					setKey := baseKey + ":set"
					if _, err := client.SAdd(ctx, setKey, "member"); err != nil {
						b.Errorf("Erro no SADD: %v", err)
						continue
					}

					// Read operations
					if _, err := client.Get(ctx, stringKey); err != nil {
						b.Errorf("Erro no GET: %v", err)
					}

					if _, err := client.HGet(ctx, hashKey, "field"); err != nil {
						b.Errorf("Erro no HGET: %v", err)
					}

					i++
				}
			})
		})
	}
}

// BenchmarkProviders_BatchOperations compara performance com operações em lote.
func BenchmarkProviders_BatchOperations(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}
	batchSizes := []int{10, 50, 100}

	for _, providerName := range providers {
		for _, batchSize := range batchSizes {
			b.Run(fmt.Sprintf("%s-batch-%d", providerName, batchSize), func(b *testing.B) {
				client := setupBenchmarkClient(b, providerName)
				defer client.Close()

				ctx := context.Background()

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					// Simular operações em lote
					for j := 0; j < batchSize; j++ {
						key := fmt.Sprintf("benchmark:batch:%s:%d:%d", providerName, i, j)
						value := fmt.Sprintf("batch_value_%d", j)

						if err := client.Set(ctx, key, value, 0); err != nil {
							b.Errorf("Erro no SET batch: %v", err)
							break
						}
					}
				}
			})
		}
	}
}

// BenchmarkProviders_ConcurrentConnections testa performance com múltiplas conexões.
func BenchmarkProviders_ConcurrentConnections(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			// Criar múltiplos clientes
			clients := make([]interfaces.IClient, 10)
			for i := range clients {
				clients[i] = setupBenchmarkClient(b, providerName)
				defer clients[i].Close()
			}

			ctx := context.Background()
			value := "concurrent_value"

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				clientIndex := 0
				i := 0
				for pb.Next() {
					client := clients[clientIndex%len(clients)]
					key := fmt.Sprintf("benchmark:concurrent:%s:%d", providerName, i)

					if err := client.Set(ctx, key, value, 0); err != nil {
						b.Errorf("Erro no SET concurrent: %v", err)
					}

					clientIndex++
					i++
				}
			})
		})
	}
}

// BenchmarkProviders_ExpireOperations compara performance com TTL.
func BenchmarkProviders_ExpireOperations(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()
			value := "expire_value"
			ttl := 1 * time.Hour // TTL alto para não expirar durante o teste

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("benchmark:expire:%s:%d", providerName, i)

					// SET com TTL
					if err := client.Set(ctx, key, value, ttl); err != nil {
						b.Errorf("Erro no SET com TTL: %v", err)
						continue
					}

					// Verificar TTL
					if _, err := client.TTL(ctx, key); err != nil {
						b.Errorf("Erro no TTL: %v", err)
					}

					i++
				}
			})
		})
	}
}

// BenchmarkProviders_LargeValues compara performance com valores grandes.
func BenchmarkProviders_LargeValues(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}
	valueSizes := []int{1024, 10240, 102400} // 1KB, 10KB, 100KB

	for _, providerName := range providers {
		for _, valueSize := range valueSizes {
			b.Run(fmt.Sprintf("%s-size-%d", providerName, valueSize), func(b *testing.B) {
				client := setupBenchmarkClient(b, providerName)
				defer client.Close()

				ctx := context.Background()
				value := string(make([]byte, valueSize))

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					key := fmt.Sprintf("benchmark:large:%s:%d", providerName, i)

					if err := client.Set(ctx, key, value, 0); err != nil {
						b.Errorf("Erro no SET large value: %v", err)
						continue
					}

					if _, err := client.Get(ctx, key); err != nil {
						b.Errorf("Erro no GET large value: %v", err)
					}
				}
			})
		}
	}
}

// BenchmarkProviders_NumberConversion testa performance com conversão de números.
func BenchmarkProviders_NumberConversion(b *testing.B) {
	providers := []string{"valkey-go", "valkey-glide"}

	for _, providerName := range providers {
		b.Run(providerName, func(b *testing.B) {
			client := setupBenchmarkClient(b, providerName)
			defer client.Close()

			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("benchmark:number:%s:%d", providerName, i)
					value := i * 123 // Valor numérico

					// SET com número
					if err := client.Set(ctx, key, value, 0); err != nil {
						b.Errorf("Erro no SET number: %v", err)
						continue
					}

					// GET e verificar
					result, err := client.Get(ctx, key)
					if err != nil {
						b.Errorf("Erro no GET number: %v", err)
						continue
					}

					// Converter de volta para número
					if _, err := strconv.Atoi(result); err != nil {
						b.Errorf("Erro ao converter number: %v", err)
					}

					i++
				}
			})
		})
	}
}
