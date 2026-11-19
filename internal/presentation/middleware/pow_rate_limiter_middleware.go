package middleware

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	infraredis "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/Doratou/Backend/pkg/pow"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// PoW request headers.  Clients that have solved a PoW challenge must include
// all five headers in the follow-up request so the server can verify and unblock.
const (
	HeaderPoWSalt       = "X-PoW-Salt"
	HeaderPoWTimestamp  = "X-PoW-Timestamp"
	HeaderPoWHMAC       = "X-PoW-HMAC"
	HeaderPoWDifficulty = "X-PoW-Difficulty"
	HeaderPoWNonce      = "X-PoW-Nonce"
)

// TwoLayerLimiterConfig holds the per-instance tuning parameters.
// All values are typically sourced from bootstrap.PoWConfig.
type TwoLayerLimiterConfig struct {
	// L1Limit is the outer (permissive) ceiling in requests per second.
	// Exceeding it returns 429 with no PoW challenge.
	L1Limit int

	// L2Limit is the inner (strict) ceiling in requests per second.
	// L2Limit must be strictly lower than L1Limit.
	// Exceeding it blocks the client in Redis and returns 423 with a challenge.
	L2Limit int

	// Difficulty is the PoW difficulty forwarded to pow.Server.GenerateChallenge.
	Difficulty int
}

// TwoLayerLimiter implements the two-layer rate limiting strategy:
//
//	Layer 1 (L1): coarse limit — blocked → 429, no challenge.
//	PoW check:    if the client is Redis-blocked, verify or issue a challenge.
//	Layer 2 (L2): fine limit  — blocked → write Redis block + 423 with challenge.
//
// The PoW mechanism is fully stateless: challenges are HMAC-signed so the
// server verifies them without any additional Redis round-trip.
type TwoLayerLimiter struct {
	rateLimiter  *redis_rate.Limiter
	powBlockRepo cache.PoWBlockRepository
	powServer    *pow.Server
	cfg          TwoLayerLimiterConfig
}

// NewTwoLayerLimiter wires up all dependencies and returns a ready limiter.
func NewTwoLayerLimiter(
	redisClient infraredis.Cache,
	powBlockRepo cache.PoWBlockRepository,
	powServer *pow.Server,
	cfg TwoLayerLimiterConfig,
) *TwoLayerLimiter {
	return &TwoLayerLimiter{
		rateLimiter:  redis_rate.NewLimiter(redisClient.GetRDB()),
		powBlockRepo: powBlockRepo,
		powServer:    powServer,
		cfg:          cfg,
	}
}

// PerSecond returns a Gin middleware that enforces both rate-limit layers on a
// per-second basis.  keyFn determines how clients are identified; pass nil to
// use DefaultKeyFunc (user ID when authenticated, IP otherwise).
//
// Request flow:
//
//  1. L1 check   → 429 if exceeded (plain rate-limit, no challenge).
//  2. PoW check  → if the client has an active Redis block:
//     a. No PoW headers     → 423 + fresh challenge.
//     b. Invalid PoW        → 422 + fresh challenge.
//     c. Valid PoW          → unblock + continue to L2.
//  3. L2 check   → block in Redis + 423 + fresh challenge if exceeded.
//  4. c.Next()   → handler runs normally.
func (tl *TwoLayerLimiter) PerSecond(keyFn KeyFunc) gin.HandlerFunc {
	if keyFn == nil {
		keyFn = DefaultKeyFunc
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		clientID := keyFn(c)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		method := c.Request.Method

		// Shared key suffix used across all Redis keys for this request.
		keySuffix := fmt.Sprintf("%s:%s:%s", method, route, clientID)

		l1Key := "rl_l1:" + keySuffix
		l2Key := "rl_l2:" + keySuffix
		blockKey := keySuffix // PoWBlockRepository adds its own prefix.

		// ── Layer 1 ──────────────────────────────────────────────────────────
		l1Res, err := tl.rateLimiter.Allow(ctx, l1Key, redis_rate.PerSecond(tl.cfg.L1Limit))
		if err != nil {
			// Fail-open: Redis is unavailable, let the request through.
			log.Printf("[TWO_LAYER_RL] L1 Redis error (fail-open): key=%q err=%v", l1Key, err)
			c.Next()
			return
		}

		// Inject standard rate-limit response headers for L1.
		tl.injectRateLimitHeaders(c, tl.cfg.L1Limit, l1Res, "L1")

		if l1Res.Allowed == 0 {
			c.Header("Retry-After", strconv.Itoa(atLeastOne(int(l1Res.RetryAfter.Seconds()))))
			log.Printf("[TWO_LAYER_RL] L1 exceeded: key=%q", l1Key)
			respondError(c, exception.ErrRateLimitExceeded)
			return
		}

		// ── PoW block check ──────────────────────────────────────────────────
		blocked, err := tl.powBlockRepo.IsBlocked(ctx, blockKey)
		if err != nil {
			// Fail-open: cannot read block state, let the request through.
			log.Printf("[TWO_LAYER_RL] PoW block check error (fail-open): key=%q err=%v", blockKey, err)
		}

		if blocked {
			if !tl.hasPowHeaders(c) {
				// Client is blocked but sent no PoW solution → issue a challenge.
				log.Printf("[TWO_LAYER_RL] blocked client has no PoW headers: key=%q", blockKey)
				tl.respondWithChallenge(c, exception.ErrPoWChallengeRequired)
				return
			}

			if !tl.verifyPoWHeaders(c) {
				// Client sent PoW headers but the solution is wrong or expired.
				log.Printf("[TWO_LAYER_RL] invalid PoW solution: key=%q", blockKey)
				tl.respondWithChallenge(c, exception.ErrPoWChallengeInvalid)
				return
			}

			// Valid PoW → remove the block and fall through to the L2 check.
			// The client may still trigger L2 again if they are hammering the API.
			if err := tl.powBlockRepo.Unblock(ctx, blockKey); err != nil {
				log.Printf("[TWO_LAYER_RL] Unblock error (non-fatal): key=%q err=%v", blockKey, err)
			}
			log.Printf("[TWO_LAYER_RL] PoW accepted, client unblocked: key=%q", blockKey)
		}

		// ── Layer 2 ──────────────────────────────────────────────────────────
		l2Res, err := tl.rateLimiter.Allow(ctx, l2Key, redis_rate.PerSecond(tl.cfg.L2Limit))
		if err != nil {
			// Fail-open: Redis is unavailable, let the request through.
			log.Printf("[TWO_LAYER_RL] L2 Redis error (fail-open): key=%q err=%v", l2Key, err)
			c.Next()
			return
		}

		// Inject standard rate-limit response headers for L2.
		tl.injectRateLimitHeaders(c, tl.cfg.L2Limit, l2Res, "L2")

		if l2Res.Allowed == 0 {
			// Block the client in Redis and return a fresh PoW challenge.
			if err := tl.powBlockRepo.Block(ctx, blockKey); err != nil {
				log.Printf("[TWO_LAYER_RL] Block error (non-fatal): key=%q err=%v", blockKey, err)
			}
			log.Printf("[TWO_LAYER_RL] L2 exceeded, client blocked: key=%q", blockKey)
			tl.respondWithChallenge(c, exception.ErrPoWChallengeRequired)
			return
		}

		c.Next()
	}
}

// hasPowHeaders reports whether all five PoW solution headers are present.
func (tl *TwoLayerLimiter) hasPowHeaders(c *gin.Context) bool {
	return c.GetHeader(HeaderPoWSalt) != "" &&
		c.GetHeader(HeaderPoWTimestamp) != "" &&
		c.GetHeader(HeaderPoWHMAC) != "" &&
		c.GetHeader(HeaderPoWDifficulty) != "" &&
		c.GetHeader(HeaderPoWNonce) != ""
}

// verifyPoWHeaders parses the five PoW request headers and delegates
// verification to pow.Server.VerifyChallenge.  Returns false on any
// parse error or failed verification.
func (tl *TwoLayerLimiter) verifyPoWHeaders(c *gin.Context) bool {
	salt := c.GetHeader(HeaderPoWSalt)
	hmacHex := c.GetHeader(HeaderPoWHMAC)
	nonce := c.GetHeader(HeaderPoWNonce)

	timestamp, err := strconv.ParseInt(c.GetHeader(HeaderPoWTimestamp), 10, 64)
	if err != nil {
		return false
	}

	difficulty, err := strconv.Atoi(c.GetHeader(HeaderPoWDifficulty))
	if err != nil {
		return false
	}

	return tl.powServer.VerifyChallenge(salt, timestamp, hmacHex, difficulty, nonce)
}

// respondWithChallenge aborts the request with the given AppError status/code
// and attaches a freshly generated PoW challenge to the response body.
func (tl *TwoLayerLimiter) respondWithChallenge(c *gin.Context, appErr *exception.AppError) {
	challenge := tl.powServer.GenerateChallenge(tl.cfg.Difficulty)
	c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
		"success":   false,
		"code":      appErr.Code,
		"error":     appErr.Message,
		"challenge": challenge,
	})
}

// injectRateLimitHeaders writes X-RateLimit-<Layer>-{Limit,Remaining,Reset}
// headers so clients can observe the state of each layer independently.
func (tl *TwoLayerLimiter) injectRateLimitHeaders(c *gin.Context, limit int, res *redis_rate.Result, layer string) {
	prefix := "X-RateLimit-" + layer + "-"
	c.Header(prefix+"Limit", strconv.Itoa(limit))
	c.Header(prefix+"Remaining", strconv.Itoa(res.Remaining))
	c.Header(prefix+"Reset", strconv.Itoa(atLeastOne(int(res.ResetAfter.Seconds()))))
}

// atLeastOne clamps v to a minimum of 1, used for Retry-After / Reset headers
// to avoid sending 0 which some HTTP clients interpret incorrectly.
func atLeastOne(v int) int {
	if v < 1 {
		return 1
	}
	return v
}
