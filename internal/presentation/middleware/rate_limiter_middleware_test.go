package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

type dummyRedisCache struct {
	cmd goredis.Cmdable
}

func (d *dummyRedisCache) GetRDB() goredis.Cmdable    { return d.cmd }
func (d *dummyRedisCache) GetClient() *goredis.Client { return nil }
func (d *dummyRedisCache) WithCmdable(c goredis.Cmdable) redis.Cache {
	return &dummyRedisCache{cmd: c}
}
func (d *dummyRedisCache) GetRedisConfig() redis.Config { return redis.Config{} }

func init() {
	gin.SetMode(gin.TestMode)
}

func TestDefaultKeyFunc(t *testing.T) {
	t.Run("authenticated user returns user:<id>", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

		claims := &jwt.Claims{
			UserID: 42,
			Role:   enum.UserRoleUser,
		}
		tokencontext.SetClaims(c, claims)

		key := DefaultKeyFunc(c)
		expected := "user:42"
		if key != expected {
			t.Fatalf("expected key %q, got %q", expected, key)
		}
	})

	t.Run("unauthenticated client returns ip:<ip>", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		c.Request.RemoteAddr = "203.0.113.195:12345"

		key := DefaultKeyFunc(c)
		expected := "ip:203.0.113.195"
		if key != expected {
			t.Fatalf("expected key %q, got %q", expected, key)
		}
	})
}

func TestIPKeyFunc(t *testing.T) {
	t.Run("strictly uses IP even when user is authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		c.Request.RemoteAddr = "198.51.100.1:54321"

		claims := &jwt.Claims{
			UserID: 99,
			Role:   enum.UserRoleUser,
		}
		tokencontext.SetClaims(c, claims)

		key := IPKeyFunc(c)
		expected := "ip:198.51.100.1"
		if key != expected {
			t.Fatalf("expected key %q, got %q", expected, key)
		}
	})
}

func TestRedisLimiter_FailOpenWhenRedisUnavailable(t *testing.T) {
	// Create an unreachable redis client (e.g. invalid port)
	client := goredis.NewClient(&goredis.Options{
		Addr: "127.0.0.1:65534", // unavailable port
	})
	dummyCache := &dummyRedisCache{cmd: client}

	limiter := NewRedisLimiter(dummyCache)

	router := gin.New()
	router.Use(limiter.PerSecond(5))
	router.GET("/test-fail-open", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-fail-open", nil)
	router.ServeHTTP(w, req)

	// Since Redis is down, fail-open should allow the request through with 200 OK
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d (fail-open), got %d", http.StatusOK, w.Code)
	}
}
