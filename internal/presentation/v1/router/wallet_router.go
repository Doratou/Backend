package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterWalletRoutes(rg *gin.RouterGroup, h *handler.WalletHandler, authMiddleware gin.HandlerFunc) {
	if authMiddleware == nil {
		panic("wallet routes require authentication middleware")
	}
	wallet := rg.Group("/wallet", authMiddleware)
	{
		wallet.GET("/summary", h.Summary)
		wallet.GET("/offers", h.Offers)
		wallet.GET("/transactions", h.ListTransactions)
		wallet.GET("/transactions/:transactionID", h.Transaction)
		wallet.POST("/purchases", h.Purchase)
	}
}
