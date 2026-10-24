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

type authnTokenRepository struct {
	cache redisInfra.Cache
}

func NewAuthnTokenRepository(cache redisInfra.Cache) cache_repo.AuthnTokenRepository {
	return &authnTokenRepository{cache: cache}
}

func (r *authnTokenRepository) getRDB(ctx context.Context) redis.Cmdable {
	return redisInfra.ExtractTrxOrCache(ctx, r.cache).GetRDB()
}

func (r *authnTokenRepository) refreshSessionKey(userID, tokenID string) string {
	return fmt.Sprintf("doratou:refresh_token:%s:%s", userID, tokenID)
}

func (r *authnTokenRepository) blacklistKey(jti string) string {
	return fmt.Sprintf("doratou:blacklist:access_token:%s", jti)
}

func (r *authnTokenRepository) SaveRefreshSession(ctx context.Context, userID, tokenID, refreshToken string, ttl time.Duration) error {
	key := r.refreshSessionKey(userID, tokenID)
	if err := r.getRDB(ctx).Set(ctx, key, refreshToken, ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, fmt.Errorf("redis set refresh session: %w", err))
	}
	return nil
}

func (r *authnTokenRepository) GetRefreshSession(ctx context.Context, userID, tokenID string) (string, error) {
	key := r.refreshSessionKey(userID, tokenID)
	val, err := r.getRDB(ctx).Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", exception.ErrTokenNotFound
		}
		return "", exception.Wrap(exception.ErrInternal, fmt.Errorf("redis get refresh session: %w", err))
	}
	return val, nil
}

func (r *authnTokenRepository) DeleteRefreshSession(ctx context.Context, userID, tokenID string) (string, error) {
	key := r.refreshSessionKey(userID, tokenID)
	if err := r.getRDB(ctx).Del(ctx, key).Err(); err != nil {
		return key, exception.Wrap(exception.ErrInternal, fmt.Errorf("redis delete refresh session: %w", err))
	}
	return key, nil
}

func (r *authnTokenRepository) DeleteUserRefreshSessions(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("doratou:refresh_token:%s:*", userID)
	rdb := r.getRDB(ctx)

	// In transaction context or standard context, scan keys
	// Note: go-redis Scan works on standard client
	var keys []string
	iter := r.cache.GetRDB().Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, fmt.Errorf("redis scan refresh sessions: %w", err))
	}

	if len(keys) > 0 {
		if err := rdb.Del(ctx, keys...).Err(); err != nil {
			return exception.Wrap(exception.ErrInternal, fmt.Errorf("redis delete user refresh sessions: %w", err))
		}
	}

	return nil
}

func (r *authnTokenRepository) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := r.blacklistKey(jti)
	if err := r.getRDB(ctx).Set(ctx, key, "revoked", ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, fmt.Errorf("redis blacklist access token: %w", err))
	}
	return nil
}

func (r *authnTokenRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := r.blacklistKey(jti)
	count, err := r.getRDB(ctx).Exists(ctx, key).Result()
	if err != nil {
		return false, exception.Wrap(exception.ErrInternal, fmt.Errorf("redis exists check failed: %w", err))
	}
	return count > 0, nil
}
