package valkeyglide

import (
	"context"
	"fmt"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/options"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// ClusterClient implementa interfaces.IClient para clientes cluster usando valkey-glide.
type ClusterClient struct {
	client *glide.ClusterClient
	config *valkeyconfig.Config
}

// Close fecha a conexão com o cliente.
func (c *ClusterClient) Close() error {
	if c.client == nil {
		return nil
	}
	c.client.Close()
	return nil
}

// Ping testa a conectividade com o servidor.
func (c *ClusterClient) Ping(ctx context.Context) error {
	result, err := c.client.Ping(ctx)
	if err != nil {
		return err
	}

	if result != "PONG" {
		return fmt.Errorf("resposta inesperada do ping: %s", result)
	}

	return nil
}

// Get obtém o valor de uma chave.
func (c *ClusterClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", fmt.Errorf("key not found")
	}

	return result.Value(), nil
}

// Set define o valor de uma chave.
func (c *ClusterClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Converter value para string
	valueStr := fmt.Sprintf("%v", value)

	if expiration > 0 {
		setOptions := options.NewSetOptions().SetExpiry(options.NewExpiryIn(expiration))
		result, err := c.client.SetWithOptions(ctx, key, valueStr, *setOptions)
		if err != nil {
			return err
		}

		if result.IsNil() {
			return fmt.Errorf("comando SET falhou")
		}

		if result.Value() != "OK" {
			return fmt.Errorf("resposta inesperada do SET: %s", result.Value())
		}
	} else {
		result, err := c.client.Set(ctx, key, valueStr)
		if err != nil {
			return err
		}

		if result != "OK" {
			return fmt.Errorf("resposta inesperada do SET: %s", result)
		}
	}

	return nil
}

// Del remove uma ou mais chaves.
func (c *ClusterClient) Del(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Del(ctx, keys)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exists verifica se uma ou mais chaves existem.
func (c *ClusterClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Exists(ctx, keys)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Expire define expiração para uma chave.
func (c *ClusterClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	success, err := c.client.Expire(ctx, key, expiration)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("falha ao definir expiração para chave: %s", key)
	}

	return nil
}

// TTL retorna o tempo de vida restante de uma chave.
func (c *ClusterClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	seconds, err := c.client.TTL(ctx, key)
	if err != nil {
		return 0, err
	}

	if seconds == -1 {
		return -1, nil // Chave existe mas não tem expiração
	}
	if seconds == -2 {
		return -2, nil // Chave não existe
	}

	return time.Duration(seconds) * time.Second, nil
}

// Incr incrementa o valor de uma chave por 1.
func (c *ClusterClient) Incr(ctx context.Context, key string) (int64, error) {
	value, err := c.client.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// IncrBy incrementa o valor de uma chave por um número especificado.
func (c *ClusterClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := c.client.IncrBy(ctx, key, value)
	if err != nil {
		return 0, err
	}

	return newValue, nil
}

// Decr decrementa o valor de uma chave por 1.
func (c *ClusterClient) Decr(ctx context.Context, key string) (int64, error) {
	value, err := c.client.Decr(ctx, key)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// DecrBy decrementa o valor de uma chave por um número especificado.
func (c *ClusterClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := c.client.DecrBy(ctx, key, value)
	if err != nil {
		return 0, err
	}

	return newValue, nil
}

// HGet obtém o valor de um campo em um hash.
func (c *ClusterClient) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := c.client.HGet(ctx, key, field)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", fmt.Errorf("key not found")
	}

	return result.Value(), nil
}

// HSet define o valor de um campo em um hash.
func (c *ClusterClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	if len(values)%2 != 0 {
		return fmt.Errorf("número de argumentos deve ser par (field1, value1, field2, value2, ...)")
	}

	// Converter para map[string]string
	fieldValueMap := make(map[string]string)
	for i := 0; i < len(values); i += 2 {
		field, ok := values[i].(string)
		if !ok {
			return fmt.Errorf("campo deve ser string: %T", values[i])
		}

		value, ok := values[i+1].(string)
		if !ok {
			return fmt.Errorf("valor deve ser string: %T", values[i+1])
		}

		fieldValueMap[field] = value
	}

	_, err := c.client.HSet(ctx, key, fieldValueMap)
	return err
}

// HDel remove um ou mais campos de um hash.
func (c *ClusterClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	count, err := c.client.HDel(ctx, key, fields)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// HExists verifica se um campo existe em um hash.
func (c *ClusterClient) HExists(ctx context.Context, key, field string) (bool, error) {
	exists, err := c.client.HExists(ctx, key, field)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// HGetAll obtém todos os campos e valores de um hash.
func (c *ClusterClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	hashMap, err := c.client.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	return hashMap, nil
}

// HKeys obtém todos os campos de um hash.
func (c *ClusterClient) HKeys(ctx context.Context, key string) ([]string, error) {
	keys, err := c.client.HKeys(ctx, key)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// HVals obtém todos os valores de um hash.
func (c *ClusterClient) HVals(ctx context.Context, key string) ([]string, error) {
	values, err := c.client.HVals(ctx, key)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// HLen retorna o número de campos em um hash.
func (c *ClusterClient) HLen(ctx context.Context, key string) (int64, error) {
	length, err := c.client.HLen(ctx, key)
	if err != nil {
		return 0, err
	}

	return length, nil
}

// Pipeline cria um novo pipeline para execução em lote.
func (c *ClusterClient) Pipeline() interfaces.IPipeline {
	return newPipeline(c.client)
}

// Transaction cria uma nova transação para execução atômica.
func (c *ClusterClient) Transaction() interfaces.ITransaction {
	return newTransaction(c.client)
}

// Config retorna a configuração do cliente.
func (c *ClusterClient) Config() interface{} {
	return c.config
}

// Métodos obrigatórios da interface IClient

// LPush adiciona elementos ao início de uma lista.
func (c *ClusterClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	// TODO: Implementar LPush
	return 0, fmt.Errorf("LPush not implemented yet")
}

// RPush adiciona elementos ao final de uma lista.
func (c *ClusterClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	// TODO: Implementar RPush
	return 0, fmt.Errorf("RPush not implemented yet")
}

// LPop remove e retorna o primeiro elemento de uma lista.
func (c *ClusterClient) LPop(ctx context.Context, key string) (string, error) {
	// TODO: Implementar LPop
	return "", fmt.Errorf("LPop not implemented yet")
}

// RPop remove e retorna o último elemento de uma lista.
func (c *ClusterClient) RPop(ctx context.Context, key string) (string, error) {
	// TODO: Implementar RPop
	return "", fmt.Errorf("RPop not implemented yet")
}

// LLen retorna o tamanho de uma lista.
func (c *ClusterClient) LLen(ctx context.Context, key string) (int64, error) {
	// TODO: Implementar LLen
	return 0, fmt.Errorf("LLen not implemented yet")
}

// SAdd adiciona membros a um set.
func (c *ClusterClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// TODO: Implementar SAdd
	return 0, fmt.Errorf("SAdd not implemented yet")
}

// SRem remove membros de um set.
func (c *ClusterClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// TODO: Implementar SRem
	return 0, fmt.Errorf("SRem not implemented yet")
}

// SMembers retorna todos os membros de um set.
func (c *ClusterClient) SMembers(ctx context.Context, key string) ([]string, error) {
	// TODO: Implementar SMembers
	return nil, fmt.Errorf("SMembers not implemented yet")
}

// SIsMember verifica se um valor é membro de um set.
func (c *ClusterClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	// TODO: Implementar SIsMember
	return false, fmt.Errorf("SIsMember not implemented yet")
}

// ZAdd adiciona membros a um sorted set.
func (c *ClusterClient) ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// TODO: Implementar ZAdd
	return 0, fmt.Errorf("ZAdd not implemented yet")
}

// ZRem remove membros de um sorted set.
func (c *ClusterClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// TODO: Implementar ZRem
	return 0, fmt.Errorf("ZRem not implemented yet")
}

// ZRange retorna elementos de um sorted set por range de índice.
func (c *ClusterClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	// TODO: Implementar ZRange
	return nil, fmt.Errorf("ZRange not implemented yet")
}

// ZScore retorna o score de um membro em um sorted set.
func (c *ClusterClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	// TODO: Implementar ZScore
	return 0, fmt.Errorf("ZScore not implemented yet")
}

// TxPipeline cria uma transação (mesmo que Transaction).
func (c *ClusterClient) TxPipeline() interfaces.ITransaction {
	return c.Transaction()
}

// Eval executa um script Lua.
func (c *ClusterClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	// TODO: Implementar Eval
	return nil, fmt.Errorf("Eval not implemented yet")
}

// EvalSha executa um script Lua por SHA.
func (c *ClusterClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	// TODO: Implementar EvalSha
	return nil, fmt.Errorf("EvalSha not implemented yet")
}

// ScriptLoad carrega um script Lua.
func (c *ClusterClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	// TODO: Implementar ScriptLoad
	return "", fmt.Errorf("ScriptLoad not implemented yet")
}

// Subscribe inscreve em canais para Pub/Sub.
func (c *ClusterClient) Subscribe(ctx context.Context, channels ...string) (interfaces.IPubSub, error) {
	// TODO: Implementar Subscribe
	return nil, fmt.Errorf("Subscribe not implemented yet")
}

// Publish publica uma mensagem em um canal.
func (c *ClusterClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	// TODO: Implementar Publish
	return 0, fmt.Errorf("Publish not implemented yet")
}

// XAdd adiciona uma entrada a um stream.
func (c *ClusterClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	// TODO: Implementar XAdd
	return "", fmt.Errorf("XAdd not implemented yet")
}

// XRead lê entradas de streams.
func (c *ClusterClient) XRead(ctx context.Context, streams map[string]string) ([]interfaces.XMessage, error) {
	// TODO: Implementar XRead
	return nil, fmt.Errorf("XRead not implemented yet")
}

// XReadGroup lê entradas de streams com grupo de consumidores.
func (c *ClusterClient) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]interfaces.XMessage, error) {
	// TODO: Implementar XReadGroup
	return nil, fmt.Errorf("XReadGroup not implemented yet")
}

// Scan escaneia chaves.
func (c *ClusterClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	// TODO: Implementar Scan
	return nil, 0, fmt.Errorf("Scan not implemented yet")
}

// HScan escaneia campos de um hash.
func (c *ClusterClient) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	// TODO: Implementar HScan
	return nil, 0, fmt.Errorf("HScan not implemented yet")
}

// IsHealthy verifica se o cliente está saudável.
func (c *ClusterClient) IsHealthy(ctx context.Context) bool {
	err := c.Ping(ctx)
	return err == nil
}
