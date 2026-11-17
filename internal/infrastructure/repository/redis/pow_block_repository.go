package redis

import (
	"context"
	"time"

	infraredis "github.com/Doratou/Backend/internal/infrastructure/redis"
)

const powBlockKeyPrefix = "pow_block:"

// PoWBlockRedisRepository implements cache.PoWBlockRepository using Redis.
//
// Each block entry is stored as a simple string key with a TTL so Redis
// garbage-collects expired blocks automatically — no background sweep needed.
type PoWBlockRedisRepository struct {
	cache    infraredis.Cache
	blockTTL time.Duration
}

// NewPoWBlockRepository creates a PoWBlockRedisRepository.
//
//   - cache:    the shared Redis connection.
//   - blockTTL: how long a block entry lives before it expires on its own.
func NewPoWBlockRepository(cache infraredis.Cache, blockTTL time.Duration) *PoWBlockRedisRepository {
	return &PoWBlockRedisRepository{
		cache:    cache,
		blockTTL: blockTTL,
	}
}

// IsBlocked returns true if a block key exists in Redis for the given client.
func (r *PoWBlockRedisRepository) IsBlocked(ctx context.Context, key string) (bool, error) {
	n, err := r.cache.GetRDB().Exists(ctx, powBlockKeyPrefix+key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Block writes a block entry for the given client key with the configured TTL.
// If the key already exists (e.g. repeated violations), the TTL is refreshed.
func (r *PoWBlockRedisRepository) Block(ctx context.Context, key string) error {
	// Value "1" is a placeholder; only the key's existence matters.
	return r.cache.GetRDB().Set(ctx, powBlockKeyPrefix+key, "1", r.blockTTL).Err()
}

// Unblock deletes the block entry so the client can proceed immediately.
// DEL on a non-existent key is safe and returns 0, which we silently ignore.
func (r *PoWBlockRedisRepository) Unblock(ctx context.Context, key string) error {
	return r.cache.GetRDB().Del(ctx, powBlockKeyPrefix+key).Err()
}
