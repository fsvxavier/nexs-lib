package goredis

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/nitishm/go-rejson/v4/rjs"
)

func (c *client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	return err
}
func (c *client) JSONSet(ctx context.Context, key, path string, obj interface{}, opts ...SetOption) (interface{}, error) {
	ops := make([]rjs.SetOption, len(opts))
	for i, opt := range opts {
		ops[i] = rjs.SetOption(opt)
	}
	res, err := c.handler.SetContext(ctx).JSONSet(key, path, obj, ops...)
	return res, wrapError(err)
}
func (c *client) JSONGet(ctx context.Context, key, path string, opts ...GetOption) (interface{}, error) {
	ops := make([]rjs.GetOption, len(opts))
	for i, opt := range opts {
		ops[i] = getOptions[opt]
	}
	res, err := c.handler.SetContext(ctx).JSONGet(key, path, ops...)
	return res, wrapError(err)
}
func (c *client) JSONMGet(ctx context.Context, path string, keys ...string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONMGet(path, keys...)
	return res, wrapError(err)
}
func (c *client) JSONDel(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONDel(key, path)
	return res, wrapError(err)
}
func (c *client) JSONType(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONType(key, path)
	return res, wrapError(err)
}
func (c *client) JSONNumIncrBy(ctx context.Context, key, path string, number int) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONNumIncrBy(key, path, number)
	return res, wrapError(err)
}
func (c *client) JSONNumMultBy(ctx context.Context, key, path string, number int) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONNumMultBy(key, path, number)
	return res, wrapError(err)
}
func (c *client) JSONStrAppend(ctx context.Context, key, path string, jsonstring string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONStrAppend(key, path, jsonstring)
	return res, wrapError(err)
}
func (c *client) JSONStrLen(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONStrLen(key, path)
	return res, wrapError(err)
}
func (c *client) JSONArrAppend(ctx context.Context, key, path string, values ...interface{}) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrAppend(key, path, values...)
	return res, wrapError(err)
}
func (c *client) JSONArrLen(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrLen(key, path)
	return res, wrapError(err)
}
func (c *client) JSONArrPop(ctx context.Context, key, path string, index int) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrPop(key, path, index)
	return res, wrapError(err)
}
func (c *client) JSONArrIndex(ctx context.Context, key, path string, jsonValue interface{}, optionalRange ...int) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrIndex(key, path, jsonValue, optionalRange...)
	return res, wrapError(err)
}
func (c *client) JSONArrTrim(ctx context.Context, key, path string, start, end int) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrTrim(key, path, start, end)
	return res, wrapError(err)
}
func (c *client) JSONArrInsert(ctx context.Context, key, path string, index int, values ...interface{}) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONArrInsert(key, path, index, values...)
	return res, wrapError(err)
}
func (c *client) JSONObjKeys(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONObjKeys(key, path)
	return res, wrapError(err)
}
func (c *client) JSONObjLen(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONObjLen(key, path)
	return res, wrapError(err)
}
func (c *client) JSONDebug(ctx context.Context, subCmd DebugSubCommand, key, path string) (interface{}, error) {
	cmd := rjs.DebugSubCommand(subCmd)
	res, err := c.handler.SetContext(ctx).JSONDebug(cmd, key, path)
	return res, wrapError(err)
}
func (c *client) JSONForget(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONForget(key, path)
	return res, wrapError(err)
}
func (c *client) JSONResp(ctx context.Context, key, path string) (interface{}, error) {
	res, err := c.handler.SetContext(ctx).JSONResp(key, path)
	return res, wrapError(err)
}

func (c *client) JSONIncrBy(ctx context.Context, key string, path string, number interface{}) (interface{}, error) {
	name, args, err := rjs.CommandBuilder(rjs.ReJSONCommandNUMINCRBY, key, path, number)
	if err != nil {
		return nil, wrapError(err)
	}
	args = append([]interface{}{name}, args...)

	res, err := c.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, wrapError(err)
	}

	return rjs.StringToBytes(res), nil
}

func (c *client) MSetJSON(ctx context.Context, params ...MSetParam) (res interface{}, err error) { // nolint: lll
	p := map[string]interface{}{}
	for _, param := range params {
		b, err := json.Marshal(param.Obj)
		if err != nil {
			return nil, wrapError(err)
		}

		p[param.Key] = string(b)

	}

	res, err = c.client.MSet(ctx, p).Result()
	return res, wrapError(err)
}

func (c *client) MGetJSON(ctx context.Context, v interface{}, keys ...string) (err error) {
	cmd := c.client.MGet(ctx, keys...)

	var res []interface{}
	res, err = cmd.Result()
	if err != nil {
		return wrapError(err)
	}

	var buf strings.Builder
	buf.WriteRune('[')
	items := len(res)
	for i := 0; i < items; i++ {
		if i != 0 {
			buf.WriteRune(',')
		}
		if res[i] == nil {
			buf.WriteString("null")
			continue
		}
		buf.WriteString(res[i].(string))
	}
	buf.WriteRune(']')

	d := json.NewDecoder(strings.NewReader(buf.String()))
	d.UseNumber()
	err = d.Decode(v)
	return wrapError(err)
}

func (c *client) GetJSON(ctx context.Context, key string, v interface{}) (err error) {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return wrapError(err)
	}

	d := json.NewDecoder(strings.NewReader(result))
	d.UseNumber()
	err = d.Decode(v)

	return wrapError(err)
}

func (c *client) Del(ctx context.Context, keys ...string) (res interface{}, err error) {
	result, err := c.client.Del(ctx, keys...).Result()
	return result, wrapError(err)
}

func (c *client) ExpireDate(ctx context.Context, key string, expireDate time.Time) (interface{}, error) {
	res, err := c.client.ExpireAt(ctx, key, expireDate).Result()
	return res, wrapError(err)
}
