package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterChecklistRoutes(rg *gin.RouterGroup, checklistHandler *handler.ChecklistHandler, authMiddleware gin.HandlerFunc) {
	if checklistHandler == nil {
		panic("checklist handler is required")
	}
	if authMiddleware == nil {
		panic("checklist routes require authentication middleware")
	}

	checklist := rg.Group("/groups/:groupID/checklist", authMiddleware)
	checklist.GET("", checklistHandler.Get)
	checklist.POST("/items", checklistHandler.CreateItem)
	checklist.PUT("/items/:itemID/claim", checklistHandler.ClaimItem)
	checklist.DELETE("/items/:itemID/claim", checklistHandler.ReleaseItem)
	checklist.DELETE("/items/:itemID", checklistHandler.DeleteItem)
}
