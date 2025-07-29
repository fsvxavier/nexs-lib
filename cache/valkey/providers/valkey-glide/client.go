package valkeyglide

import (
	"context"
	"fmt"
	"strconv"
	"time"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/options"

	valkeyconfig "github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// Client implementa interfaces.IClient para clientes standalone usando valkey-glide.
type Client struct {
	client *glide.Client
	config *valkeyconfig.Config
}

// Close fecha a conexão com o cliente.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	c.client.Close()
	return nil
}

// Ping testa a conectividade com o servidor.
func (c *Client) Ping(ctx context.Context) error {
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
func (c *Client) Get(ctx context.Context, key string) (string, error) {
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
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
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
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Del(ctx, keys)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exists verifica se uma ou mais chaves existem.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Exists(ctx, keys)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Expire define expiração para uma chave.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
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
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
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
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	value, err := c.client.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// IncrBy incrementa o valor de uma chave por um número especificado.
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := c.client.IncrBy(ctx, key, value)
	if err != nil {
		return 0, err
	}

	return newValue, nil
}

// Decr decrementa o valor de uma chave por 1.
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	value, err := c.client.Decr(ctx, key)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// DecrBy decrementa o valor de uma chave por um número especificado.
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := c.client.DecrBy(ctx, key, value)
	if err != nil {
		return 0, err
	}

	return newValue, nil
}

// HGet obtém o valor de um campo em um hash.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
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
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
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
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	count, err := c.client.HDel(ctx, key, fields)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// HExists verifica se um campo existe em um hash.
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	exists, err := c.client.HExists(ctx, key, field)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// HGetAll obtém todos os campos e valores de um hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	hashMap, err := c.client.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	return hashMap, nil
}

// HKeys obtém todos os campos de um hash.
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	keys, err := c.client.HKeys(ctx, key)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// HVals obtém todos os valores de um hash.
func (c *Client) HVals(ctx context.Context, key string) ([]string, error) {
	values, err := c.client.HVals(ctx, key)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// HLen retorna o número de campos em um hash.
func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	length, err := c.client.HLen(ctx, key)
	if err != nil {
		return 0, err
	}

	return length, nil
}

// Pipeline cria um novo pipeline para execução em lote.
func (c *Client) Pipeline() interfaces.IPipeline {
	return newPipeline(c.client)
}

// Transaction cria uma nova transação para execução atômica.
func (c *Client) Transaction() interfaces.ITransaction {
	return newTransaction(c.client)
}

// Config retorna a configuração do cliente.
func (c *Client) Config() interface{} {
	return c.config
}

// Métodos obrigatórios da interface IClient

// LPush adiciona elementos ao início de uma lista.
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	count, err := c.client.LPush(ctx, key, stringValues)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// RPush adiciona elementos ao final de uma lista.
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	count, err := c.client.RPush(ctx, key, stringValues)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// LPop remove e retorna o primeiro elemento de uma lista.
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	result, err := c.client.LPop(ctx, key)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", fmt.Errorf("list is empty or key does not exist")
	}

	return result.Value(), nil
}

// RPop remove e retorna o último elemento de uma lista.
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	result, err := c.client.RPop(ctx, key)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", fmt.Errorf("list is empty or key does not exist")
	}

	return result.Value(), nil
}

// LLen retorna o tamanho de uma lista.
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	count, err := c.client.LLen(ctx, key)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SAdd adiciona membros a um set.
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	count, err := c.client.SAdd(ctx, key, stringMembers)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SRem remove membros de um set.
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	count, err := c.client.SRem(ctx, key, stringMembers)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SMembers retorna todos os membros de um set.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	membersMap, err := c.client.SMembers(ctx, key)
	if err != nil {
		return nil, err
	}

	// Converter map[string]struct{} para []string
	members := make([]string, 0, len(membersMap))
	for member := range membersMap {
		members = append(members, member)
	}

	return members, nil
}

// SIsMember verifica se um valor é membro de um set.
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	memberStr := fmt.Sprintf("%v", member)
	isMember, err := c.client.SIsMember(ctx, key, memberStr)
	if err != nil {
		return false, err
	}

	return isMember, nil
}

// ZAdd adiciona membros a um sorted set.
func (c *Client) ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if len(members)%2 != 0 {
		return 0, fmt.Errorf("número de argumentos deve ser par (score1, member1, score2, member2, ...)")
	}

	// Converter para map[string]float64
	scoreMembers := make(map[string]float64)
	for i := 0; i < len(members); i += 2 {
		score, ok := members[i].(float64)
		if !ok {
			// Tentar converter de int ou string
			switch v := members[i].(type) {
			case int:
				score = float64(v)
			case int64:
				score = float64(v)
			case string:
				var err error
				if score, err = strconv.ParseFloat(v, 64); err != nil {
					return 0, fmt.Errorf("score deve ser numérico: %T", members[i])
				}
			default:
				return 0, fmt.Errorf("score deve ser numérico: %T", members[i])
			}
		}

		member := fmt.Sprintf("%v", members[i+1])
		scoreMembers[member] = score
	}

	count, err := c.client.ZAdd(ctx, key, scoreMembers)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ZRem remove membros de um sorted set.
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	count, err := c.client.ZRem(ctx, key, stringMembers)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ZRange retorna elementos de um sorted set por range de índice.
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	rangeQuery := options.NewRangeByIndexQuery(start, stop)
	result, err := c.client.ZRange(ctx, key, rangeQuery)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ZScore retorna o score de um membro em um sorted set.
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	result, err := c.client.ZScore(ctx, key, member)
	if err != nil {
		return 0, err
	}

	if result.IsNil() {
		return 0, fmt.Errorf("member not found in sorted set")
	}

	return result.Value(), nil
}

// TxPipeline cria uma transação (mesmo que Transaction).
func (c *Client) TxPipeline() interfaces.ITransaction {
	return c.Transaction()
}

// Eval executa um script Lua (implementação básica).
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	// Para compatibilidade básica, vamos usar InvokeScript
	// Isso pode ser expandido futuramente conforme necessidade
	return nil, fmt.Errorf("Eval não suportado completamente no valkey-glide - use InvokeScript diretamente")
}

// EvalSha executa um script Lua por SHA (implementação básica).
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	// Para compatibilidade básica
	return nil, fmt.Errorf("EvalSha não suportado completamente no valkey-glide - use InvokeScript diretamente")
}

// ScriptLoad carrega um script Lua (implementação básica).
func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	// Para compatibilidade básica
	return "", fmt.Errorf("ScriptLoad não suportado completamente no valkey-glide - use InvokeScript diretamente")
}

// Subscribe inscreve em canais para Pub/Sub (implementação básica).
func (c *Client) Subscribe(ctx context.Context, channels ...string) (interfaces.IPubSub, error) {
	// Para compatibilidade básica - implementação completa requer estrutura mais complexa
	return nil, fmt.Errorf("Subscribe não implementado ainda no valkey-glide provider")
}

// Publish publica uma mensagem em um canal.
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	messageStr := fmt.Sprintf("%v", message)
	count, err := c.client.Publish(ctx, channel, messageStr)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// XAdd adiciona uma entrada a um stream (implementação básica).
func (c *Client) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	// Para compatibilidade básica - valkey-glide pode ter API diferente para streams
	return "", fmt.Errorf("XAdd não implementado ainda no valkey-glide provider")
}

// XRead lê entradas de streams (implementação básica).
func (c *Client) XRead(ctx context.Context, streams map[string]string) ([]interfaces.XMessage, error) {
	// Para compatibilidade básica - valkey-glide pode ter API diferente para streams
	return nil, fmt.Errorf("XRead não implementado ainda no valkey-glide provider")
}

// XReadGroup lê entradas de streams com grupo de consumidores (implementação básica).
func (c *Client) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]interfaces.XMessage, error) {
	// Para compatibilidade básica - valkey-glide pode ter API diferente para streams
	return nil, fmt.Errorf("XReadGroup não implementado ainda no valkey-glide provider")
}

// Scan escaneia chaves (implementação básica).
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	// Para compatibilidade básica - valkey-glide pode ter API diferente
	return nil, 0, fmt.Errorf("Scan não implementado ainda no valkey-glide provider")
}

// HScan escaneia campos de um hash (implementação básica).
func (c *Client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	// Para compatibilidade básica - valkey-glide pode ter API diferente
	return nil, 0, fmt.Errorf("HScan não implementado ainda no valkey-glide provider")
}

// IsHealthy verifica se o cliente está saudável.
func (c *Client) IsHealthy(ctx context.Context) bool {
	err := c.Ping(ctx)
	return err == nil
}
