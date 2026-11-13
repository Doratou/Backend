package redis

import (
	"context"
	"time"

	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"

	cache_repo "github.com/Doratou/Backend/internal/domain/repository/cache"
)

const powBlockKeyPrefix = "pow_block:"

type powBlockRedisRepository struct {
	cache    redisInfra.Cache
	blockTTL time.Duration
}

func NewPoWBlockRepository(cache redisInfra.Cache, blockTTL time.Duration) cache_repo.PoWBlockRepository {
	return &powBlockRedisRepository{
		cache:    cache,
		blockTTL: blockTTL,
	}
}

// IsBlocked returns true if a block key exists in Redis for the given client.
func (r *powBlockRedisRepository) IsBlocked(ctx context.Context, key string) (bool, error) {
	n, err := r.cache.GetRDB().Exists(ctx, powBlockKeyPrefix+key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Block writes a block entry for the given client key with the configured TTL.
// If the key already exists (e.g. repeated violations), the TTL is refreshed.
func (r *powBlockRedisRepository) Block(ctx context.Context, key string) error {
	// Value "1" is a placeholder; only the key's existence matters.
	return r.cache.GetRDB().Set(ctx, powBlockKeyPrefix+key, "1", r.blockTTL).Err()
}

// Unblock deletes the block entry so the client can proceed immediately.
// DEL on a non-existent key is safe and returns 0, which we silently ignore.
func (r *powBlockRedisRepository) Unblock(ctx context.Context, key string) error {
	return r.cache.GetRDB().Del(ctx, powBlockKeyPrefix+key).Err()
}
