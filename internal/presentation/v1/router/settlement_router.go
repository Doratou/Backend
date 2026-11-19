package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterSettlementRoutes(rg *gin.RouterGroup, settlementHandler *handler.SettlementHandler, authMiddleware gin.HandlerFunc) {
	if settlementHandler == nil {
		panic("settlement handler is required")
	}
	if authMiddleware == nil {
		panic("settlement routes require authentication middleware")
	}
	settlements := rg.Group("/groups/:groupID/settlements", authMiddleware)
	settlements.POST("", settlementHandler.Create)
	settlements.GET("", settlementHandler.List)
	settlements.GET("/:settlementID", settlementHandler.Get)
	settlements.PATCH("/:settlementID/confirm", settlementHandler.Confirm)
	settlements.PATCH("/:settlementID/reject", settlementHandler.Reject)
	settlements.DELETE("/:settlementID", settlementHandler.Delete)
}
