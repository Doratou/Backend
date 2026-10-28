package redis

import (
	"context"
	"fmt"
	"net"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host      string
	Port      string
	Password  string
	RDBNumber int
}

// Cache defines the interface for interacting with Redis, supporting both direct and transactional operations.
type Cache interface {
	GetRDB() redis.Cmdable
	GetClient() *redis.Client
	WithCmdable(c redis.Cmdable) Cache
	GetRedisConfig() Config
}

type RedisDatabase struct {
	rdb *redis.Client
	cmd redis.Cmdable
	cfg Config
}

func NewRedisDatabase(ctx context.Context, cfg *Config) (Cache, error) {
	address := net.JoinHostPort(cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:                  address,
		Password:              cfg.Password,
		DB:                    cfg.RDBNumber,
		ContextTimeoutEnabled: true,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisDatabase{rdb: rdb, cfg: *cfg}, nil
}

func (r *RedisDatabase) GetRDB() redis.Cmdable {
	if r.cmd != nil {
		return r.cmd
	}
	return r.rdb
}

func (r *RedisDatabase) GetClient() *redis.Client {
	return r.rdb
}

func (r *RedisDatabase) WithCmdable(c redis.Cmdable) Cache {
	return &RedisDatabase{rdb: r.rdb, cmd: c, cfg: r.cfg}
}

func (r *RedisDatabase) GetRedisConfig() Config {
	return r.cfg
}
