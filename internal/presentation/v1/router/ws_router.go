package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterWSRoutes mounts the WebSocket upgrade endpoint onto the versioned router group.
func RegisterWSRoutes(rg *gin.RouterGroup, wsHandler *handler.WSHandler) {
	if wsHandler == nil {
		return
	}
	rg.GET("/ws", wsHandler.Connect)
}
