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

type captchaRepository struct {
	cache redisInfra.Cache
}

// NewCaptchaRepository initializes a Redis-backed CaptchaRepository.
func NewCaptchaRepository(cache redisInfra.Cache) cache_repo.CaptchaRepository {
	return &captchaRepository{cache: cache}
}

func (r *captchaRepository) getRDB(ctx context.Context) redis.Cmdable {
	return redisInfra.ExtractTrxOrCache(ctx, r.cache).GetRDB()
}

func (r *captchaRepository) captchaKey(captchaID string) string {
	return fmt.Sprintf("doratou:captcha:%s", captchaID)
}

func (r *captchaRepository) Save(ctx context.Context, captchaID, code string, ttl time.Duration) error {
	key := r.captchaKey(captchaID)
	if err := r.getRDB(ctx).Set(ctx, key, code, ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *captchaRepository) Get(ctx context.Context, captchaID string) (string, error) {
	key := r.captchaKey(captchaID)
	val, err := r.getRDB(ctx).Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", exception.ErrCaptchaNotFound
		}
		return "", exception.Wrap(exception.ErrInternal, err)
	}
	return val, nil
}

func (r *captchaRepository) Delete(ctx context.Context, captchaID string) error {
	key := r.captchaKey(captchaID)
	if err := r.getRDB(ctx).Del(ctx, key).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *captchaRepository) Exists(ctx context.Context, captchaID string) (bool, error) {
	key := r.captchaKey(captchaID)
	count, err := r.getRDB(ctx).Exists(ctx, key).Result()
	if err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}
