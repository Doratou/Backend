package router

import (
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAdminReportRoutes wires the user reporting and admin moderation endpoints onto the router group.
func RegisterAdminReportRoutes(
	rg *gin.RouterGroup,
	reportHandler *handler.AdminReportHandler,
	authMiddleware gin.HandlerFunc,
) {
	if reportHandler == nil {
		return
	}

	// =========================================================================
	// User Reporting Routes (Protected)
	// =========================================================================
	reports := rg.Group("/reports")
	reports.Use(authMiddleware)
	{
		reports.POST("", reportHandler.SubmitReport)
		reports.GET("/my", reportHandler.ListMyReports)
		reports.GET("/:id", reportHandler.GetReport)
		reports.PATCH("/:id", reportHandler.UpdateMyPendingReport)
		reports.DELETE("/:id", reportHandler.CancelMyReport)
	}

	// =========================================================================
	// Admin Moderation Routes (Protected by Auth and RBAC: ADMIN only)
	// =========================================================================
	adminReports := rg.Group("/admin/reports")
	adminReports.Use(authMiddleware, middleware.RequireRole(enum.UserRoleAdmin))
	{
		adminReports.GET("", reportHandler.AdminListReports)
		adminReports.GET("/stats", reportHandler.AdminGetStats)
		adminReports.GET("/:id", reportHandler.AdminGetReport)
		adminReports.PUT("/:id/review", reportHandler.AdminReviewReport)
		adminReports.DELETE("/:id", reportHandler.AdminDeleteReport)
		adminReports.POST("/:id/restore", reportHandler.AdminRestoreReport)
		adminReports.DELETE("/:id/hard", reportHandler.AdminHardDeleteReport)
	}
}
