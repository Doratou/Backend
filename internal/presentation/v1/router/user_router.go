package router

import (
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/presentation/middleware"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler, authMiddleware gin.HandlerFunc) {
	if userHandler == nil {
		panic("user handler is required")
	}
	if authMiddleware == nil {
		panic("user routes require authentication middleware")
	}

	// ==========================================
	// User Routes (Protected)
	// ==========================================
	users := rg.Group("/users", authMiddleware)
	{
		users.GET("/me", userHandler.GetMe)
		users.PATCH("/me", userHandler.UpdateMe)
		users.PATCH("/me/email", userHandler.ChangeEmail)
		users.PATCH("/me/phone", userHandler.ChangePhone)
		users.DELETE("/me", userHandler.DeleteMe)

		users.GET("/me/wallet", userHandler.GetWallet)
		users.POST("/me/wallet/top-up", userHandler.TopUpWallet)

		users.GET("/by-friend-code/:friendCode", userHandler.FindByFriendCode)
		users.POST("/me/friend-code/regenerate", userHandler.RegenerateFriendCode)
		users.GET("/:id", userHandler.GetUserByID)
	}

	// ==========================================
	// Admin User Management Routes (Protected + ADMIN Role)
	// ==========================================
	adminUsers := rg.Group("/admin/users", authMiddleware, middleware.RequireRole(enum.UserRoleAdmin))
	{
		adminUsers.GET("", userHandler.AdminListUsers)
		adminUsers.GET("/:id", userHandler.AdminGetUserByID)
		adminUsers.PATCH("/:id/role", userHandler.AdminUpdateUserRole)
	}
}
