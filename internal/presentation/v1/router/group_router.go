package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterGroupRoutes(rg *gin.RouterGroup, groupHandler *handler.GroupHandler, authMiddleware gin.HandlerFunc) {
	groups := rg.Group("/groups")
	{
		groups.POST("", authMiddleware, groupHandler.Create)
		groups.GET("", authMiddleware, groupHandler.ListMine)
		groups.GET("/admin", authMiddleware, groupHandler.ListAll)
		groups.GET("/:groupID", authMiddleware, groupHandler.Get)
		groups.PATCH("/:groupID", authMiddleware, groupHandler.Update)
		groups.PATCH("/:groupID/cover", authMiddleware, groupHandler.ChangeCover)
		groups.POST("/:groupID/archive", authMiddleware, groupHandler.Archive)
		groups.DELETE("/:groupID", authMiddleware, groupHandler.Delete)
		groups.POST("/:groupID/clone", authMiddleware, groupHandler.Clone)
	}
}
