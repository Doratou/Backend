package router

import (
	"net/http"

	"github.com/Doratou/Backend/internal/domain/repository/cache"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// Dependencies bundles all handlers and presentation dependencies needed to wire routes.
type Dependencies struct {
	AuthHandler *handler.AuthHandler
	TokenMaker  jwt.TokenMaker
	TokenRepo   cache.AuthnTokenRepository
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

	v1 := engine.Group("/api/v1")
	if deps.AuthHandler != nil {
		RegisterAuthRoutes(v1, deps.AuthHandler, authMiddleware)
	}

	return engine
}
