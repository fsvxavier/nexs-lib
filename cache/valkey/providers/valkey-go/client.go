package valkeygo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// Client implementa interfaces.IClient usando valkey-go.
type Client struct {
	client valkey.Client
	config *config.Config
}

// String commands

// Get implementa interfaces.IClient.Get.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	cmd := c.client.B().Get().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return "", fmt.Errorf("key not found")
		}
		return "", result.Error()
	}

	return result.ToString()
}

// Set implementa interfaces.IClient.Set.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	cmd := c.client.B().Set().Key(key).Value(fmt.Sprintf("%v", value))

	var builtCmd valkey.Completed
	if expiration > 0 {
		builtCmd = cmd.Ex(expiration).Build()
	} else {
		builtCmd = cmd.Build()
	}

	result := c.client.Do(ctx, builtCmd)
	return result.Error()
}

// Del implementa interfaces.IClient.Del.
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	cmd := c.client.B().Del().Key(keys...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// Exists implementa interfaces.IClient.Exists.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	cmd := c.client.B().Exists().Key(keys...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// TTL implementa interfaces.IClient.TTL.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	cmd := c.client.B().Ttl().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	ttlSeconds, err := result.AsInt64()
	if err != nil {
		return 0, err
	}

	// -1 significa sem expiração, -2 significa chave não existe
	if ttlSeconds == -1 {
		return -1, nil
	}
	if ttlSeconds == -2 {
		return 0, fmt.Errorf("key not found")
	}

	return time.Duration(ttlSeconds) * time.Second, nil
}

// Expire implementa interfaces.IClient.Expire.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	cmd := c.client.B().Expire().Key(key).Seconds(int64(expiration.Seconds())).Build()
	result := c.client.Do(ctx, cmd)
	return result.Error()
}

// Hash commands

// HGet implementa interfaces.IClient.HGet.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	cmd := c.client.B().Hget().Key(key).Field(field).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return "", fmt.Errorf("field not found")
		}
		return "", result.Error()
	}

	return result.ToString()
}

// HSet implementa interfaces.IClient.HSet.
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	if len(values)%2 != 0 {
		return fmt.Errorf("número ímpar de argumentos para HSET")
	}

	cmd := c.client.B().Hset().Key(key).FieldValue()

	for i := 0; i < len(values); i += 2 {
		field := fmt.Sprintf("%v", values[i])
		value := fmt.Sprintf("%v", values[i+1])
		cmd = cmd.FieldValue(field, value)
	}

	result := c.client.Do(ctx, cmd.Build())
	return result.Error()
}

// HDel implementa interfaces.IClient.HDel.
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	cmd := c.client.B().Hdel().Key(key).Field(fields...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// HExists implementa interfaces.IClient.HExists.
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	cmd := c.client.B().Hexists().Key(key).Field(field).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return false, result.Error()
	}

	exists, err := result.AsInt64()
	return exists == 1, err
}

// HGetAll implementa interfaces.IClient.HGetAll.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cmd := c.client.B().Hgetall().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	return result.AsStrMap()
}

// List commands

// LPush implementa interfaces.IClient.LPush.
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}

	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	cmd := c.client.B().Lpush().Key(key).Element(stringValues...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// RPush implementa interfaces.IClient.RPush.
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}

	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = fmt.Sprintf("%v", v)
	}

	cmd := c.client.B().Rpush().Key(key).Element(stringValues...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// LPop implementa interfaces.IClient.LPop.
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	cmd := c.client.B().Lpop().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return "", fmt.Errorf("list is empty")
		}
		return "", result.Error()
	}

	return result.ToString()
}

// RPop implementa interfaces.IClient.RPop.
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	cmd := c.client.B().Rpop().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return "", fmt.Errorf("list is empty")
		}
		return "", result.Error()
	}

	return result.ToString()
}

// LLen implementa interfaces.IClient.LLen.
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	cmd := c.client.B().Llen().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// Set commands

// SAdd implementa interfaces.IClient.SAdd.
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if len(members) == 0 {
		return 0, nil
	}

	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	cmd := c.client.B().Sadd().Key(key).Member(stringMembers...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// SRem implementa interfaces.IClient.SRem.
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if len(members) == 0 {
		return 0, nil
	}

	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	cmd := c.client.B().Srem().Key(key).Member(stringMembers...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// SMembers implementa interfaces.IClient.SMembers.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	cmd := c.client.B().Smembers().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	return result.AsStrSlice()
}

// SIsMember implementa interfaces.IClient.SIsMember.
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	memberStr := fmt.Sprintf("%v", member)
	cmd := c.client.B().Sismember().Key(key).Member(memberStr).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return false, result.Error()
	}

	exists, err := result.AsInt64()
	return exists == 1, err
}

// Sorted Set commands

// ZAdd implementa interfaces.IClient.ZAdd.
func (c *Client) ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if len(members) == 0 || len(members)%2 != 0 {
		return 0, fmt.Errorf("número inválido de argumentos para ZADD")
	}

	cmd := c.client.B().Zadd().Key(key).ScoreMember()

	for i := 0; i < len(members); i += 2 {
		score, err := parseFloat64(members[i])
		if err != nil {
			return 0, fmt.Errorf("score inválido: %w", err)
		}
		member := fmt.Sprintf("%v", members[i+1])
		cmd = cmd.ScoreMember(score, member)
	}

	result := c.client.Do(ctx, cmd.Build())

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// ZRem implementa interfaces.IClient.ZRem.
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if len(members) == 0 {
		return 0, nil
	}

	stringMembers := make([]string, len(members))
	for i, m := range members {
		stringMembers[i] = fmt.Sprintf("%v", m)
	}

	cmd := c.client.B().Zrem().Key(key).Member(stringMembers...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// ZRange implementa interfaces.IClient.ZRange.
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cmd := c.client.B().Zrange().Key(key).Min(strconv.FormatInt(start, 10)).Max(strconv.FormatInt(stop, 10)).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	return result.AsStrSlice()
}

// ZScore implementa interfaces.IClient.ZScore.
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	cmd := c.client.B().Zscore().Key(key).Member(member).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		if valkey.IsValkeyNil(result.Error()) {
			return 0, fmt.Errorf("member not found")
		}
		return 0, result.Error()
	}

	return result.AsFloat64()
}

// Scripts - implementações básicas
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	stringArgs := make([]string, len(args))
	for i, arg := range args {
		stringArgs[i] = fmt.Sprintf("%v", arg)
	}

	cmd := c.client.B().Eval().Script(script).Numkeys(int64(len(keys))).Key(keys...).Arg(stringArgs...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	// Retornar resultado bruto - pode ser melhorado
	return result, nil
}

func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	stringArgs := make([]string, len(args))
	for i, arg := range args {
		stringArgs[i] = fmt.Sprintf("%v", arg)
	}

	cmd := c.client.B().Evalsha().Sha1(sha1).Numkeys(int64(len(keys))).Key(keys...).Arg(stringArgs...).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	return result, nil
}

func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	cmd := c.client.B().ScriptLoad().Script(script).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return "", result.Error()
	}

	return result.ToString()
}

// Pub/Sub, Streams, Scan - implementações básicas

// Subscribe implementa interfaces.IClient.Subscribe.
func (c *Client) Subscribe(ctx context.Context, channels ...string) (interfaces.IPubSub, error) {
	// Implementação básica - seria expandida
	return nil, fmt.Errorf("Subscribe não implementado nesta versão")
}

// Publish implementa interfaces.IClient.Publish.
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	cmd := c.client.B().Publish().Channel(channel).Message(fmt.Sprintf("%v", message)).Build()
	result := c.client.Do(ctx, cmd)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return result.AsInt64()
}

// XAdd implementa interfaces.IClient.XAdd.
func (c *Client) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	cmd := c.client.B().Xadd().Key(stream).Id("*").FieldValue()
	for field, value := range values {
		cmd = cmd.FieldValue(field, fmt.Sprintf("%v", value))
	}

	result := c.client.Do(ctx, cmd.Build())

	if result.Error() != nil {
		return "", result.Error()
	}

	return result.ToString()
}

// XRead implementa interfaces.IClient.XRead.
func (c *Client) XRead(ctx context.Context, streams map[string]string) ([]interfaces.XMessage, error) {
	// Implementação básica - seria expandida
	return nil, fmt.Errorf("XRead não implementado nesta versão")
}

// XReadGroup implementa interfaces.IClient.XReadGroup.
func (c *Client) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]interfaces.XMessage, error) {
	// Implementação básica - seria expandida
	return nil, fmt.Errorf("XReadGroup não implementado nesta versão")
}

// Scan implementa interfaces.IClient.Scan.
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	cmd := c.client.B().Scan().Cursor(cursor)

	var builtCmd valkey.Completed
	if match != "" && count > 0 {
		builtCmd = cmd.Match(match).Count(count).Build()
	} else if match != "" {
		builtCmd = cmd.Match(match).Build()
	} else if count > 0 {
		builtCmd = cmd.Count(count).Build()
	} else {
		builtCmd = cmd.Build()
	}

	result := c.client.Do(ctx, builtCmd)

	if result.Error() != nil {
		return nil, 0, result.Error()
	}

	// Implementação básica - seria melhorada para parsear o resultado do SCAN
	return []string{}, 0, fmt.Errorf("Scan parsing não implementado nesta versão")
}

// HScan implementa interfaces.IClient.HScan.
func (c *Client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	cmd := c.client.B().Hscan().Key(key).Cursor(cursor)

	var builtCmd valkey.Completed
	if match != "" && count > 0 {
		builtCmd = cmd.Match(match).Count(count).Build()
	} else if match != "" {
		builtCmd = cmd.Match(match).Build()
	} else if count > 0 {
		builtCmd = cmd.Count(count).Build()
	} else {
		builtCmd = cmd.Build()
	}

	result := c.client.Do(ctx, builtCmd)

	if result.Error() != nil {
		return nil, 0, result.Error()
	}

	// Implementação básica - seria melhorada
	return []string{}, 0, fmt.Errorf("HScan parsing não implementado nesta versão")
}

// Connection management

// Ping implementa interfaces.IClient.Ping.
func (c *Client) Ping(ctx context.Context) error {
	cmd := c.client.B().Ping().Build()
	result := c.client.Do(ctx, cmd)
	return result.Error()
}

// Close implementa interfaces.IClient.Close.
func (c *Client) Close() error {
	c.client.Close()
	return nil
}

// IsHealthy implementa interfaces.IClient.IsHealthy.
func (c *Client) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// Pipeline implementa interfaces.IClient.Pipeline.
func (c *Client) Pipeline() interfaces.IPipeline {
	return &Pipeline{
		client: c.client,
		cmds:   make([]valkey.Completed, 0),
	}
}

// TxPipeline implementa interfaces.IClient.TxPipeline.
func (c *Client) TxPipeline() interfaces.ITransaction {
	return &Transaction{
		client: c.client,
		cmds:   make([]valkey.Completed, 0),
	}
}

// parseFloat64 converte interface{} para float64.
func parseFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}
