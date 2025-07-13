package goredis

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"sync"

	redisWrapper "github.com/DataDog/dd-trace-go/contrib/redis/go-redis.v9/v2"
	"github.com/nitishm/go-rejson/v4"
	"github.com/nitishm/go-rejson/v4/rjs"
	goredis "github.com/redis/go-redis/v9"
)

var (
	instance Json
	once     sync.Once

	ErrKeyNotFound = errors.New("key not found")
)

type DebugSubCommand string
type SetOption string
type GetOption string

const (
	// DebugMemorySubcommand provide the corresponding MEMORY sub commands for JSONDebug
	DebugMemorySubcommand DebugSubCommand = "MEMORY"

	// DebugHelpSubcommand provide the corresponding HELP sub commands for JSONDebug
	DebugHelpSubcommand DebugSubCommand = "HELP"

	// JSONSET command Options
	SetOptionNX SetOption = "NX"
	SetOptionXX SetOption = "XX"

	// JSONGET command Options
	GetOptionSPACE    GetOption = "SPACE"
	GetOptionINDENT   GetOption = "INDENT"
	GetOptionNEWLINE  GetOption = "NEWLINE"
	GetOptionNOESCAPE GetOption = "NOESCAPE"
)

var (
	// Mapping getoptions commands
	getOptions = map[GetOption]rjs.GetOption{
		GetOptionSPACE:    rjs.GETOptionSPACE,
		GetOptionINDENT:   rjs.GETOptionINDENT,
		GetOptionNEWLINE:  rjs.GETOptionNEWLINE,
		GetOptionNOESCAPE: rjs.GETOptionNOESCAPE,
	}
)

type client struct {
	client     goredis.UniversalClient
	handler    *rejson.Handler
	withTracer bool
}

type MSetParam struct {
	Obj  interface{}
	Key  string
	Path string
}

// Verify interface compliance
var _ Json = (*client)(nil)

const (
	MAX_IDLE   = 1000
	MAX_ACTIVE = 1000
)

// NewClient creates a new Redis client instance with the provided options.
// It initializes the client with the default configuration and applies any additional options provided.
// The client is a singleton, ensuring that only one instance exists throughout the application lifecycle.
// The client supports both standard and cluster modes, with TLS support if enabled in the environment variables.
func NewClient(options ...GoRedisConfig) Json {

	cfg := &goRedisConfig{}

	DefaultsGoRedisConfig(cfg)
	for _, opt := range options {
		opt(cfg)
	}

	return GetInstance(cfg)
}

// Singleton implementation to garantee the existence of only one Redis client instance.
// This approach was chosen due to some designs previously used in the project.
func GetInstance(cfg *goRedisConfig) Json {

	once.Do(func() {
		opt := &goredis.UniversalOptions{
			Addrs:          cfg.Addresses,
			MaxIdleConns:   cfg.MaxIdleConns,
			PoolSize:       cfg.PoolSize,
			MinIdleConns:   cfg.MinIdleConns,
			Username:       cfg.Username,
			Password:       cfg.Password,
			DB:             cfg.DB,
			MaxRetries:     cfg.MaxRetries,
			DialTimeout:    cfg.DialTimeout,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			PoolTimeout:    cfg.PoolTimeout,
			MaxActiveConns: cfg.MaxActiveConns,
		}

		if cfg.UseTLS {
			opt.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		conn := goredis.NewUniversalClient(opt)
		if cfg.TraceEnabled {
			redisWrapper.WrapClient(
				conn,
				redisWrapper.WithService(cfg.TraceService),
			)
		}

		handler := rejson.NewReJSONHandler()
		handler.SetGoRedisClientWithContext(context.Background(), conn)
		instance = &client{
			client:     conn,
			handler:    handler,
			withTracer: cfg.TraceEnabled,
		}
	})

	return instance
}

func wrapError(err error) error {
	switch err {
	case nil:
		// No error
		return nil
	case goredis.Nil:
		return ErrKeyNotFound
	default:
		// On Update
		if strings.HasPrefix(err.Error(), "ERR could not perform this operation on a key that doesn't exist") {
			return ErrKeyNotFound
		}
		return err
	}
}
