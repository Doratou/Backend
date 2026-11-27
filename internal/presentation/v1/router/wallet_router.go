package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterWalletRoutes(rg *gin.RouterGroup, h *handler.WalletHandler, authMiddleware gin.HandlerFunc) {
	if authMiddleware == nil {
		panic("wallet routes require authentication middleware")
	}
	rg.GET("/users/me/wallet", authMiddleware, h.Summary)
	wallet := rg.Group("/wallet", authMiddleware)
	{
		wallet.GET("/offers", h.Offers)
		wallet.GET("/transactions", h.ListTransactions)
		wallet.GET("/transactions/:transactionID", h.Transaction)
		wallet.POST("/purchases", h.Purchase)
	}
}
