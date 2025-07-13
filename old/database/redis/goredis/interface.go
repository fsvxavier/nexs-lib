package goredis

import (
	"context"
	"time"
)

type JsonBasic interface {
	Ping(ctx context.Context) error
	JSONSet(ctx context.Context, key, path string, obj interface{}, opts ...SetOption) (interface{}, error)
	JSONGet(ctx context.Context, key, path string, opts ...GetOption) (interface{}, error)
	JSONMGet(ctx context.Context, path string, keys ...string) (interface{}, error)
	JSONDel(ctx context.Context, key, path string) (interface{}, error)
	JSONType(ctx context.Context, key, path string) (interface{}, error)
	JSONDebug(ctx context.Context, subCmd DebugSubCommand, key, path string) (interface{}, error)
	JSONForget(ctx context.Context, key, path string) (interface{}, error)
	JSONResp(ctx context.Context, key, path string) (interface{}, error)
	JSONIncrBy(ctx context.Context, key string, path string, number interface{}) (interface{}, error)
	ExpireDate(ctx context.Context, key string, expireDate time.Time) (interface{}, error)
	MSetJSON(ctx context.Context, params ...MSetParam) (res interface{}, err error)
	GetJSON(ctx context.Context, key string, v interface{}) (err error)
	MGetJSON(ctx context.Context, v interface{}, keys ...string) (err error)
	Del(ctx context.Context, keys ...string) (res interface{}, err error)
}

type JsonObj interface {
	JSONObjKeys(ctx context.Context, key, path string) (interface{}, error)
	JSONObjLen(ctx context.Context, key, path string) (interface{}, error)
}

type JsonNum interface {
	JSONNumIncrBy(ctx context.Context, key, path string, number int) (interface{}, error)
	JSONNumMultBy(ctx context.Context, key, path string, number int) (interface{}, error)
}

type JsonStr interface {
	JSONStrAppend(ctx context.Context, key, path string, jsonstring string) (interface{}, error)
	JSONStrLen(ctx context.Context, key, path string) (interface{}, error)
}

type JsonArray interface {
	JSONArrAppend(ctx context.Context, key, path string, values ...interface{}) (interface{}, error)
	JSONArrLen(ctx context.Context, key, path string) (interface{}, error)
	JSONArrPop(ctx context.Context, key, path string, index int) (interface{}, error)
	JSONArrIndex(ctx context.Context, key, path string, jsonValue interface{}, optionalRange ...int) (interface{}, error)
	JSONArrTrim(ctx context.Context, key, path string, start, end int) (interface{}, error)
	JSONArrInsert(ctx context.Context, key, path string, index int, values ...interface{}) (interface{}, error)
}

// Json agrega todas as interfaces
type Json interface {
	JsonBasic
	JsonArray
	JsonNum
	JsonStr
	JsonObj
}
