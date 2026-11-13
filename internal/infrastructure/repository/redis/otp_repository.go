package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	cache_repo "github.com/Doratou/Backend/internal/domain/repository/cache"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/redis/go-redis/v9"
)

type otpRepository struct {
	cache redisInfra.Cache
}

func NewOTPRepository(cache redisInfra.Cache) cache_repo.OTPRepository {
	return &otpRepository{cache: cache}
}

func (r *otpRepository) getRDB(ctx context.Context) redis.Cmdable {
	return redisInfra.ExtractTrxOrCache(ctx, r.cache).GetRDB()
}

func (r *otpRepository) otpKey(otpType enum.OTPType, identifier string) string {
	return fmt.Sprintf("doratou:otp:%s:%s", otpType, identifier)
}

func (r *otpRepository) verifiedKey(otpType enum.OTPType, identifier string) string {
	return fmt.Sprintf("doratou:otp_verified:%s:%s", otpType, identifier)
}

func (r *otpRepository) SetOTP(ctx context.Context, otpType enum.OTPType, identifier, code string, ttl time.Duration) error {
	key := r.otpKey(otpType, identifier)
	if err := r.getRDB(ctx).Set(ctx, key, code, ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *otpRepository) GetOTP(ctx context.Context, otpType enum.OTPType, identifier string) (string, error) {
	key := r.otpKey(otpType, identifier)
	val, err := r.getRDB(ctx).Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", exception.ErrOTPNotFound
		}
		return "", exception.Wrap(exception.ErrInternal, err)
	}
	return val, nil
}

func (r *otpRepository) DeleteOTP(ctx context.Context, otpType enum.OTPType, identifier string) error {
	key := r.otpKey(otpType, identifier)
	if err := r.getRDB(ctx).Del(ctx, key).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *otpRepository) GetOTPTTL(ctx context.Context, otpType enum.OTPType, identifier string) (time.Duration, error) {
	key := r.otpKey(otpType, identifier)
	ttl, err := r.getRDB(ctx).TTL(ctx, key).Result()
	if err != nil {
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	if ttl <= 0 {
		return 0, exception.ErrOTPNotFound
	}
	return ttl, nil
}

func (r *otpRepository) SetVerified(ctx context.Context, otpType enum.OTPType, identifier string, ttl time.Duration) error {
	key := r.verifiedKey(otpType, identifier)
	if err := r.getRDB(ctx).Set(ctx, key, "1", ttl).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *otpRepository) IsVerified(ctx context.Context, otpType enum.OTPType, identifier string) (bool, error) {
	key := r.verifiedKey(otpType, identifier)
	exists, err := r.getRDB(ctx).Exists(ctx, key).Result()
	if err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return exists > 0, nil
}

func (r *otpRepository) ClearVerified(ctx context.Context, otpType enum.OTPType, identifier string) error {
	key := r.verifiedKey(otpType, identifier)
	if err := r.getRDB(ctx).Del(ctx, key).Err(); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}
