package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Doratou/Backend/internal/domain/exception"
	cache_repo "github.com/Doratou/Backend/internal/domain/repository/cache"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/redis/go-redis/v9"
)

type oauthStateRepository struct {
	cache redisInfra.Cache
}

// NewOAuthStateRepository creates a new Redis implementation of cache_repo.OAuthStateRepository.
func NewOAuthStateRepository(cache redisInfra.Cache) cache_repo.OAuthStateRepository {
	return &oauthStateRepository{cache: cache}
}

func (r *oauthStateRepository) getRDB(ctx context.Context) redis.Cmdable {
	return redisInfra.ExtractTrxOrCache(ctx, r.cache).GetRDB()
}

func (r *oauthStateRepository) stateKey(state string) string {
	return fmt.Sprintf("doratou:oauth:state:%s", state)
}

func (r *oauthStateRepository) SaveState(ctx context.Context, state, nonce string, ttl time.Duration) error {
	key := r.stateKey(state)
	if err := r.getRDB(ctx).Set(ctx, key, nonce, ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *oauthStateRepository) ConsumeState(ctx context.Context, state string) (string, error) {
	key := r.stateKey(state)
	nonce, err := r.getRDB(ctx).GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", exception.ErrOAuthStateInvalid
		}
		return "", exception.Wrap(exception.ErrInternal, err)
	}
	return nonce, nil
}
