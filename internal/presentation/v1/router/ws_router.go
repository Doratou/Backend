package router

import (
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterWSRoutes mounts the WebSocket upgrade endpoint and admin monitoring routes onto the versioned router group.
func RegisterWSRoutes(rg *gin.RouterGroup, wsHandler *handler.WSHandler, authMiddleware gin.HandlerFunc) {
	if wsHandler == nil {
		return
	}

	rg.GET("/ws", wsHandler.Connect)

	if authMiddleware != nil {
		adminWS := rg.Group("/admin/ws")
		adminWS.Use(authMiddleware, middleware.RequireRole(enum.UserRoleAdmin))
		{
			adminWS.GET("/stats", wsHandler.Stats)
			adminWS.GET("/snapshot", wsHandler.Snapshot)
		}
	}
}
