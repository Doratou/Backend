package router

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// Dependencies bundles all handlers and presentation dependencies needed to wire routes.
type Dependencies struct {
	AuthHandler    *handler.AuthHandler
	CaptchaHandler *handler.CaptchaHandler
	CaptchaService adapter.CaptchaService
	CaptchaEnabled bool
	TokenMaker     jwt.TokenMaker
	TokenRepo      cache.AuthnTokenRepository
}

// New builds the fully configured Gin engine with global middlewares and versioned routes.
func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	var authMiddleware gin.HandlerFunc
	if deps.TokenMaker != nil && deps.TokenRepo != nil {
		authMiddleware = middleware.Auth(deps.TokenMaker, deps.TokenRepo)
	}

	var captchaMiddleware gin.HandlerFunc
	if deps.CaptchaService != nil {
		captchaMiddleware = middleware.CaptchaGuard(deps.CaptchaService, deps.CaptchaEnabled)
	}

	v1 := engine.Group("/api/v1")
	RegisterAuthRoutes(v1, deps.AuthHandler, authMiddleware, captchaMiddleware)
	RegisterCaptchaRoutes(v1, deps.CaptchaHandler)

	return engine
}
