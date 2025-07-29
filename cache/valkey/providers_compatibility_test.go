// Package valkey - Testes de compatibilidade entre providers
// Este arquivo implementa testes abrangentes para validar que todos os providers
// (valkey-go e valkey-glide) implementam a interface IClient de forma consistente.
package valkey

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
	valkeyglide "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-glide"
	valkeygo "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-go"
)

// ProviderTestSuite define a estrutura para testes de compatibilidade.
type ProviderTestSuite struct {
	Name     string
	Provider interfaces.IProvider
	Config   *config.Config
}

// getAllProviders retorna todos os providers disponíveis para teste.
func getAllProviders() []ProviderTestSuite {
	return []ProviderTestSuite{
		{
			Name:     "valkey-go",
			Provider: valkeygo.NewProvider(),
			Config: &config.Config{
				Provider:     "valkey-go",
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
			},
		},
		{
			Name:     "valkey-glide",
			Provider: valkeyglide.NewProvider(),
			Config: &config.Config{
				Provider:     "valkey-glide",
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
			},
		},
	}
}

// TestProviderCompatibility_BasicOperations testa operações básicas de todos os providers.
func TestProviderCompatibility_BasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			// Criar cliente
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			// Teste de conectividade
			err = client.Ping(ctx)
			if err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			// Teste básico Set/Get
			testKey := fmt.Sprintf("test:compatibility:%s:%d", provider.Name, time.Now().UnixNano())
			testValue := "test_value_123"

			// Set
			err = client.Set(ctx, testKey, testValue, 0)
			require.NoError(t, err, "Set deveria funcionar em %s", provider.Name)

			// Get
			result, err := client.Get(ctx, testKey)
			require.NoError(t, err, "Get deveria funcionar em %s", provider.Name)
			assert.Equal(t, testValue, result, "Valor recuperado deveria ser igual em %s", provider.Name)

			// Del
			count, err := client.Del(ctx, testKey)
			require.NoError(t, err, "Del deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(1), count, "Del deveria retornar 1 em %s", provider.Name)

			// Verificar que foi deletado
			_, err = client.Get(ctx, testKey)
			assert.Error(t, err, "Get de chave deletada deveria falhar em %s", provider.Name)
		})
	}
}

// TestProviderCompatibility_StringOperations testa operações de string.
func TestProviderCompatibility_StringOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			testKey := fmt.Sprintf("test:string:%s:%d", provider.Name, time.Now().UnixNano())

			// Teste Set com TTL
			err = client.Set(ctx, testKey, "value_with_ttl", 100*time.Millisecond)
			require.NoError(t, err, "Set com TTL deveria funcionar em %s", provider.Name)

			// Verificar que existe
			result, err := client.Get(ctx, testKey)
			require.NoError(t, err, "Get deveria funcionar imediatamente em %s", provider.Name)
			assert.Equal(t, "value_with_ttl", result)

			// Verificar TTL
			ttl, err := client.TTL(ctx, testKey)
			require.NoError(t, err, "TTL deveria funcionar em %s", provider.Name)
			assert.True(t, ttl > 0 && ttl <= 100*time.Millisecond, "TTL deveria estar entre 0 e 100ms em %s", provider.Name)

			// Aguardar expiração
			time.Sleep(150 * time.Millisecond)

			// Verificar que expirou
			_, err = client.Get(ctx, testKey)
			assert.Error(t, err, "Get de chave expirada deveria falhar em %s", provider.Name)

			// Teste Exists
			key1 := fmt.Sprintf("test:exists1:%s:%d", provider.Name, time.Now().UnixNano())
			key2 := fmt.Sprintf("test:exists2:%s:%d", provider.Name, time.Now().UnixNano())

			// Criar apenas uma chave
			err = client.Set(ctx, key1, "value1", 0)
			require.NoError(t, err)

			count, err := client.Exists(ctx, key1, key2)
			require.NoError(t, err, "Exists deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(1), count, "Exists deveria retornar 1 em %s", provider.Name)

			// Limpeza
			client.Del(ctx, key1)
		})
	}
}

// TestProviderCompatibility_HashOperations testa operações de hash.
func TestProviderCompatibility_HashOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			hashKey := fmt.Sprintf("test:hash:%s:%d", provider.Name, time.Now().UnixNano())

			// HSet
			err = client.HSet(ctx, hashKey, "field1", "value1", "field2", "value2")
			require.NoError(t, err, "HSet deveria funcionar em %s", provider.Name)

			// HGet
			value, err := client.HGet(ctx, hashKey, "field1")
			require.NoError(t, err, "HGet deveria funcionar em %s", provider.Name)
			assert.Equal(t, "value1", value, "HGet deveria retornar valor correto em %s", provider.Name)

			// HExists
			exists, err := client.HExists(ctx, hashKey, "field1")
			require.NoError(t, err, "HExists deveria funcionar em %s", provider.Name)
			assert.True(t, exists, "HExists deveria retornar true em %s", provider.Name)

			exists, err = client.HExists(ctx, hashKey, "field_inexistente")
			require.NoError(t, err, "HExists para campo inexistente deveria funcionar em %s", provider.Name)
			assert.False(t, exists, "HExists deveria retornar false para campo inexistente em %s", provider.Name)

			// HGetAll
			allFields, err := client.HGetAll(ctx, hashKey)
			require.NoError(t, err, "HGetAll deveria funcionar em %s", provider.Name)
			assert.Equal(t, map[string]string{
				"field1": "value1",
				"field2": "value2",
			}, allFields, "HGetAll deveria retornar todos os campos em %s", provider.Name)

			// HDel
			count, err := client.HDel(ctx, hashKey, "field1")
			require.NoError(t, err, "HDel deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(1), count, "HDel deveria retornar 1 em %s", provider.Name)

			// Verificar que foi deletado
			_, err = client.HGet(ctx, hashKey, "field1")
			assert.Error(t, err, "HGet de campo deletado deveria falhar em %s", provider.Name)

			// Limpeza
			client.Del(ctx, hashKey)
		})
	}
}

// TestProviderCompatibility_ListOperations testa operações de lista.
func TestProviderCompatibility_ListOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			listKey := fmt.Sprintf("test:list:%s:%d", provider.Name, time.Now().UnixNano())

			// LPush
			count, err := client.LPush(ctx, listKey, "item1", "item2")
			require.NoError(t, err, "LPush deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(2), count, "LPush deveria retornar 2 em %s", provider.Name)

			// RPush
			count, err = client.RPush(ctx, listKey, "item3", "item4")
			require.NoError(t, err, "RPush deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(4), count, "RPush deveria retornar 4 em %s", provider.Name)

			// LLen
			length, err := client.LLen(ctx, listKey)
			require.NoError(t, err, "LLen deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(4), length, "LLen deveria retornar 4 em %s", provider.Name)

			// LPop
			item, err := client.LPop(ctx, listKey)
			require.NoError(t, err, "LPop deveria funcionar em %s", provider.Name)
			assert.Equal(t, "item2", item, "LPop deveria retornar item2 em %s", provider.Name)

			// RPop
			item, err = client.RPop(ctx, listKey)
			require.NoError(t, err, "RPop deveria funcionar em %s", provider.Name)
			assert.Equal(t, "item4", item, "RPop deveria retornar item4 em %s", provider.Name)

			// Verificar tamanho final
			length, err = client.LLen(ctx, listKey)
			require.NoError(t, err, "LLen final deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(2), length, "LLen final deveria retornar 2 em %s", provider.Name)

			// Limpeza
			client.Del(ctx, listKey)
		})
	}
}

// TestProviderCompatibility_SetOperations testa operações de set.
func TestProviderCompatibility_SetOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			setKey := fmt.Sprintf("test:set:%s:%d", provider.Name, time.Now().UnixNano())

			// SAdd
			count, err := client.SAdd(ctx, setKey, "member1", "member2", "member3")
			require.NoError(t, err, "SAdd deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(3), count, "SAdd deveria retornar 3 em %s", provider.Name)

			// SIsMember
			isMember, err := client.SIsMember(ctx, setKey, "member1")
			require.NoError(t, err, "SIsMember deveria funcionar em %s", provider.Name)
			assert.True(t, isMember, "SIsMember deveria retornar true em %s", provider.Name)

			isMember, err = client.SIsMember(ctx, setKey, "member_inexistente")
			require.NoError(t, err, "SIsMember para membro inexistente deveria funcionar em %s", provider.Name)
			assert.False(t, isMember, "SIsMember deveria retornar false para membro inexistente em %s", provider.Name)

			// SMembers
			members, err := client.SMembers(ctx, setKey)
			require.NoError(t, err, "SMembers deveria funcionar em %s", provider.Name)
			assert.Len(t, members, 3, "SMembers deveria retornar 3 membros em %s", provider.Name)
			assert.Contains(t, members, "member1", "SMembers deveria conter member1 em %s", provider.Name)
			assert.Contains(t, members, "member2", "SMembers deveria conter member2 em %s", provider.Name)
			assert.Contains(t, members, "member3", "SMembers deveria conter member3 em %s", provider.Name)

			// SRem
			count, err = client.SRem(ctx, setKey, "member1")
			require.NoError(t, err, "SRem deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(1), count, "SRem deveria retornar 1 em %s", provider.Name)

			// Verificar que foi removido
			isMember, err = client.SIsMember(ctx, setKey, "member1")
			require.NoError(t, err, "SIsMember após SRem deveria funcionar em %s", provider.Name)
			assert.False(t, isMember, "SIsMember deveria retornar false após SRem em %s", provider.Name)

			// Limpeza
			client.Del(ctx, setKey)
		})
	}
}

// TestProviderCompatibility_SortedSetOperations testa operações de sorted set.
func TestProviderCompatibility_SortedSetOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			zsetKey := fmt.Sprintf("test:zset:%s:%d", provider.Name, time.Now().UnixNano())

			// ZAdd
			count, err := client.ZAdd(ctx, zsetKey, 1.0, "member1", 2.0, "member2", 3.0, "member3")
			require.NoError(t, err, "ZAdd deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(3), count, "ZAdd deveria retornar 3 em %s", provider.Name)

			// ZScore
			score, err := client.ZScore(ctx, zsetKey, "member2")
			require.NoError(t, err, "ZScore deveria funcionar em %s", provider.Name)
			assert.Equal(t, 2.0, score, "ZScore deveria retornar 2.0 em %s", provider.Name)

			// ZRange
			members, err := client.ZRange(ctx, zsetKey, 0, -1)
			require.NoError(t, err, "ZRange deveria funcionar em %s", provider.Name)
			assert.Equal(t, []string{"member1", "member2", "member3"}, members, "ZRange deveria retornar membros ordenados em %s", provider.Name)

			// ZRem
			count, err = client.ZRem(ctx, zsetKey, "member2")
			require.NoError(t, err, "ZRem deveria funcionar em %s", provider.Name)
			assert.Equal(t, int64(1), count, "ZRem deveria retornar 1 em %s", provider.Name)

			// Verificar que foi removido
			_, err = client.ZScore(ctx, zsetKey, "member2")
			assert.Error(t, err, "ZScore de membro removido deveria falhar em %s", provider.Name)

			// Limpeza
			client.Del(ctx, zsetKey)
		})
	}
}

// TestProviderCompatibility_ErrorHandling testa tratamento de erros consistente.
func TestProviderCompatibility_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			if err := client.Ping(ctx); err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			nonExistentKey := fmt.Sprintf("test:nonexistent:%s:%d", provider.Name, time.Now().UnixNano())

			// Get de chave inexistente
			_, err = client.Get(ctx, nonExistentKey)
			assert.Error(t, err, "Get de chave inexistente deveria retornar erro em %s", provider.Name)

			// HGet de campo inexistente
			_, err = client.HGet(ctx, nonExistentKey, "field")
			assert.Error(t, err, "HGet de campo inexistente deveria retornar erro em %s", provider.Name)

			// LPop de lista vazia
			_, err = client.LPop(ctx, nonExistentKey)
			assert.Error(t, err, "LPop de lista vazia deveria retornar erro em %s", provider.Name)

			// ZScore de membro inexistente
			_, err = client.ZScore(ctx, nonExistentKey, "member")
			assert.Error(t, err, "ZScore de membro inexistente deveria retornar erro em %s", provider.Name)
		})
	}
}

// TestProviderCompatibility_HealthCheck testa verificações de saúde.
func TestProviderCompatibility_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando testes de compatibilidade em modo short")
	}

	providers := getAllProviders()
	ctx := context.Background()

	for _, provider := range providers {
		t.Run(provider.Name, func(t *testing.T) {
			client, err := provider.Provider.NewClient(provider.Config)
			if err != nil {
				t.Skipf("Não foi possível criar cliente %s: %v", provider.Name, err)
				return
			}
			defer client.Close()

			// Ping
			err = client.Ping(ctx)
			if err != nil {
				t.Skipf("Servidor Valkey não disponível para %s: %v", provider.Name, err)
				return
			}

			// IsHealthy quando conectado
			healthy := client.IsHealthy(ctx)
			assert.True(t, healthy, "IsHealthy deveria retornar true quando conectado em %s", provider.Name)

			// Teste de Close
			err = client.Close()
			assert.NoError(t, err, "Close deveria funcionar em %s", provider.Name)

			// IsHealthy após close (pode variar entre providers)
			healthy = client.IsHealthy(ctx)
			// Não fazemos assert aqui porque o comportamento pode variar entre providers após close
			t.Logf("IsHealthy após close em %s: %t", provider.Name, healthy)
		})
	}
}
