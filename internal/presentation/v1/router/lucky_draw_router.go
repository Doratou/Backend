package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterLuckyDrawRoutes(rg *gin.RouterGroup, drawHandler *handler.LuckyDrawHandler, authMiddleware gin.HandlerFunc) {
	if drawHandler == nil {
		panic("lucky draw handler is required")
	}
	if authMiddleware == nil {
		panic("lucky draw routes require authentication middleware")
	}
	draws := rg.Group("/groups/:groupID/draws", authMiddleware)
	{
		draws.POST("", drawHandler.Create)
		draws.GET("", drawHandler.List)
		draws.GET("/:drawID", drawHandler.Get)
		draws.PATCH("/:drawID", drawHandler.Update)
		draws.POST("/:drawID/run", drawHandler.Run)
	}
}
