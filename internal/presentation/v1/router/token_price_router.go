package router

import (
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterTokenPriceRoutes mounts admin token price management endpoints.
func RegisterTokenPriceRoutes(rg *gin.RouterGroup, h *handler.TokenPriceHandler, authMiddleware gin.HandlerFunc) {
	if h == nil || authMiddleware == nil {
		return
	}

	adminToken := rg.Group("/admin/token-price")
	adminToken.Use(authMiddleware, middleware.RequireRole(enum.UserRoleAdmin))
	{
		adminToken.PUT("", h.UpdatePrice)
	}
}
