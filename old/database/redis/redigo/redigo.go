package redis

import (
	"context"
	"errors"
	"time"

	redigotrace "github.com/DataDog/dd-trace-go/contrib/gomodule/redigo/v2"
	rgo "github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

type Redigo struct {
	Pool    *rgo.Pool
	Cluster *redisc.Cluster
	cfg     *redigoConfig
}

type IRedigo interface {
	Acquire(ctx context.Context) (conn rgo.Conn, err error)
	Close() (err error)
	Get(ctx context.Context, key string) (string, error)
	HSet(ctx context.Context, hash, key, val string) error
	HGet(ctx context.Context, hash, key string) (string, error)
	Set(ctx context.Context, key, val string, exp time.Duration) error
	WithDatabase(ctx context.Context, db int) error
}

func NewRedigo(ctx context.Context, options ...RedigoConfig) (rdbg Redigo, err error) {
	cfg := GetConfig()
	for _, opt := range options {
		opt(cfg)
	}

	rdbg.cfg = cfg

	if len(rdbg.cfg.addresses) > 1 {
		cluster, errCluster := rdbg.createCluster()
		if errCluster != nil {
			return Redigo{}, errCluster
		}

		// initialize its mapping
		errRefresh := cluster.Refresh()
		if errRefresh != nil {
			return Redigo{}, errRefresh
		}

		rdbg.Cluster = cluster
	} else {
		pool, errPool := rdbg.createPool(ctx)
		if errPool != nil {
			return Redigo{}, errPool
		}
		rdbg.Pool = pool
	}

	return rdbg, err
}

func (rdbg *Redigo) createCluster() (cluster *redisc.Cluster, err error) {
	// create the cluster
	cluster = &redisc.Cluster{
		StartupNodes: rdbg.cfg.addresses,
		DialOptions: []rgo.DialOption{
			rgo.DialConnectTimeout(5 * time.Second),
		},
		CreatePool: rdbg.creatingPool,
	}

	return cluster, nil
}

func (rdbg *Redigo) creatingPool(addr string, opts ...rgo.DialOption) (*rgo.Pool, error) {
	return &rgo.Pool{
		MaxIdle:         rdbg.cfg.maxIdle,
		MaxActive:       rdbg.cfg.maxActive,
		IdleTimeout:     rdbg.cfg.idleTimeout,
		MaxConnLifetime: rdbg.cfg.maxConnLifetime,
		DialContext: func(ctx context.Context) (conn rgo.Conn, err error) {
			connDial, errDial := redigotrace.DialContext(ctx, "tcp", addr,
				redigotrace.WithContextConnection(),
				rgo.DialConnectTimeout(5*time.Second),
				rgo.DialUseTLS(rdbg.cfg.usageTLS),
				rgo.DialTLSConfig(rdbg.cfg.tlsConfig),
				redigotrace.WithService(rdbg.cfg.traceServiceName),
				rgo.DialTLSSkipVerify(rdbg.cfg.skipVerify),
				rgo.DialDatabase(rdbg.cfg.database),
			)
			return connDial, errDial
		},
		TestOnBorrow: func(c rgo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}, nil
}

func (rdbg *Redigo) createPool(ctx context.Context) (pool *rgo.Pool, err error) {
	pool = &rgo.Pool{
		MaxIdle:         rdbg.cfg.maxIdle,
		MaxActive:       rdbg.cfg.maxActive,
		IdleTimeout:     rdbg.cfg.idleTimeout,
		MaxConnLifetime: rdbg.cfg.maxConnLifetime,
		DialContext: func(ctx context.Context) (conn rgo.Conn, err error) {
			connDial, errDial := redigotrace.DialContext(ctx, "tcp", rdbg.cfg.addresses[0],
				redigotrace.WithContextConnection(),
				rgo.DialConnectTimeout(5*time.Second),
				rgo.DialUseTLS(rdbg.cfg.usageTLS),
				rgo.DialTLSConfig(rdbg.cfg.tlsConfig),
				redigotrace.WithService(rdbg.cfg.traceServiceName),
				rgo.DialTLSSkipVerify(rdbg.cfg.skipVerify),
				rgo.DialDatabase(rdbg.cfg.database),
			)
			return connDial, errDial
		},
		TestOnBorrow: func(c rgo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return pool, pool.Get().Err()
}

func (rdbg *Redigo) Acquire(ctx context.Context) (conn rgo.Conn, err error) {
	if rdbg.Pool == nil {
		return nil, errors.New("Error while connecting to redis")
	}

	if len(rdbg.cfg.addresses) > 1 {
		return rdbg.Cluster.Get(), nil
	} else {
		return rdbg.Pool.GetContext(ctx)
	}
}

func (rdbg *Redigo) Close() (err error) {
	err = rdbg.Pool.Close()
	if err != nil {
		return errors.New("Error while close conn to redis")
	}
	return nil
}

func (rdbg *Redigo) Set(ctx context.Context, key, val string, exp time.Duration) error {
	_, err := rdbg.Pool.Get().Do("SET", key, val, exp, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rdbg *Redigo) Get(ctx context.Context, key string) (string, error) {
	val, err := rgo.String(rdbg.Pool.Get().Do("GET", key, ctx))
	if err != nil {
		return "", err
	}
	return val, err
}

func (rdbg *Redigo) HSet(ctx context.Context, hash, key, val string) error {
	_, err := rdbg.Pool.Get().Do("HSET", hash, key, val, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rdbg *Redigo) HGet(ctx context.Context, hash, key string) (string, error) {
	val, err := rgo.String(rdbg.Pool.Get().Do("HGET", hash, key, ctx))
	if err != nil {
		return "", err
	}
	return val, err
}

func (rdbg *Redigo) WithDatabase(ctx context.Context, db int) error {
	_, err := rdbg.Pool.Get().Do("SELECT", db, ctx)
	if err != nil {
		return err
	}
	return nil
}
