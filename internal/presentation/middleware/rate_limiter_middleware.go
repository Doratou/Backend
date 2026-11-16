package middleware

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// KeyFunc extracts a unique client identifier from the Gin context.
type KeyFunc func(c *gin.Context) string

// DefaultKeyFunc extracts user ID if authenticated via tokencontext; otherwise falls back to ClientIP.
func DefaultKeyFunc(c *gin.Context) string {
	if userID, ok := tokencontext.GetUserID(c); ok {
		return fmt.Sprintf("user:%d", userID)
	}
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}
	return fmt.Sprintf("ip:%s", ip)
}

// IPKeyFunc strictly extracts the client IP address, regardless of authentication state.
func IPKeyFunc(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}
	return fmt.Sprintf("ip:%s", ip)
}

// RedisLimiter handles distributed token bucket rate limiting backed by Redis.
type RedisLimiter struct {
	limiter *redis_rate.Limiter
}

// NewRedisLimiter constructs a RedisLimiter from the application's Cache interface.
func NewRedisLimiter(redisClient redis.Cache) *RedisLimiter {
	return &RedisLimiter{
		limiter: redis_rate.NewLimiter(redisClient.GetRDB()),
	}
}

// PerSecond returns a Gin middleware restricting requests to `limit` per second.
func (rlm *RedisLimiter) PerSecond(limit int) gin.HandlerFunc {
	return rlm.buildMiddleware(redis_rate.PerSecond(limit), DefaultKeyFunc)
}

// PerMinute returns a Gin middleware restricting requests to `limit` per minute.
func (rlm *RedisLimiter) PerMinute(limit int) gin.HandlerFunc {
	return rlm.buildMiddleware(redis_rate.PerMinute(limit), DefaultKeyFunc)
}

// PerHour returns a Gin middleware restricting requests to `limit` per hour.
func (rlm *RedisLimiter) PerHour(limit int) gin.HandlerFunc {
	return rlm.buildMiddleware(redis_rate.PerHour(limit), DefaultKeyFunc)
}

// Custom returns a Gin middleware using an arbitrary redis_rate.Limit rule and DefaultKeyFunc.
func (rlm *RedisLimiter) Custom(limitRule redis_rate.Limit) gin.HandlerFunc {
	return rlm.buildMiddleware(limitRule, DefaultKeyFunc)
}

// WithCustomKey returns a Gin middleware using an arbitrary redis_rate.Limit and custom key extractor.
func (rlm *RedisLimiter) WithCustomKey(limitRule redis_rate.Limit, keyFn KeyFunc) gin.HandlerFunc {
	if keyFn == nil {
		keyFn = DefaultKeyFunc
	}
	return rlm.buildMiddleware(limitRule, keyFn)
}

// ByIP returns a Gin middleware restricting requests strictly by client IP address.
func (rlm *RedisLimiter) ByIP(limitRule redis_rate.Limit) gin.HandlerFunc {
	return rlm.buildMiddleware(limitRule, IPKeyFunc)
}

func (rlm *RedisLimiter) buildMiddleware(limitRule redis_rate.Limit, keyFn KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIdentifier := keyFn(c)

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		key := fmt.Sprintf("rate_limit:%s:%s:%s", c.Request.Method, route, clientIdentifier)

		res, err := rlm.limiter.Allow(c.Request.Context(), key, limitRule)
		if err != nil {
			// Fail-open strategy: log Redis error and allow the request through so the API stays available
			log.Printf("[RATE LIMITER ERROR] Redis check failed for key %q: %v", key, err)
			c.Next()
			return
		}

		// Inject standard rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limitRule.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		resetAfterSec := int(res.ResetAfter.Seconds())
		if resetAfterSec < 0 {
			resetAfterSec = 0
		}
		c.Header("X-RateLimit-Reset", strconv.Itoa(resetAfterSec))

		// If quota exceeded, set Retry-After and abort with 429
		if res.Allowed == 0 {
			retryAfterSec := int(res.RetryAfter.Seconds())
			if retryAfterSec < 1 {
				retryAfterSec = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfterSec))

			log.Printf("[RATE LIMIT EXCEEDED] method=%s route=%s client=%s limit=%d remaining=%d retry_after=%ds",
				c.Request.Method, route, clientIdentifier, limitRule.Rate, res.Remaining, retryAfterSec)

			respondError(c, exception.ErrRateLimitExceeded)
			return
		}

		c.Next()
	}
}
