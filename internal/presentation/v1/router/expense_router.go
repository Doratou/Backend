package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterExpenseRoutes(rg *gin.RouterGroup, expenseHandler *handler.ExpenseHandler, authMiddleware gin.HandlerFunc) {
	if expenseHandler == nil {
		panic("expense handler is required")
	}
	if authMiddleware == nil {
		panic("expense routes require authentication middleware")
	}
	expenses := rg.Group("/groups/:groupID/expenses", authMiddleware)
	expenses.POST("", expenseHandler.Create)
	expenses.GET("", expenseHandler.List)
	expenses.GET("/participants", expenseHandler.ListParticipants)
	expenses.GET("/:expenseID", expenseHandler.Get)
	expenses.PUT("/:expenseID", expenseHandler.Update)
	expenses.PATCH("/:expenseID", expenseHandler.Patch)
	expenses.DELETE("/:expenseID", expenseHandler.Delete)
}
