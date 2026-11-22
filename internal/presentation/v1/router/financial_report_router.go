package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterFinancialReportRoutes(rg *gin.RouterGroup, reportHandler *handler.FinancialReportHandler, authMiddleware gin.HandlerFunc) {
	if reportHandler == nil {
		panic("financial report handler is required")
	}
	if authMiddleware == nil {
		panic("financial report routes require authentication middleware")
	}
	reports := rg.Group("/groups/:groupID/reports", authMiddleware)
	reports.GET("/expenses/owed-by-me", reportHandler.OwedByMe)
	reports.GET("/expenses/owed-to-me", reportHandler.OwedToMe)
	reports.GET("/expenses/graph", reportHandler.ExpenseGraph)
	reports.GET("/settlements/graph", reportHandler.SettlementGraph)
	reports.GET("/settlements/me", reportHandler.MySettlements)
}
