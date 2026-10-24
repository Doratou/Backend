package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes wires the authentication and OTP verification endpoints onto the router group.
func RegisterAuthRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/otp/send", authHandler.SendOTP)
		auth.POST("/otp/verify", authHandler.VerifyOTP)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)

		auth.POST("/logout", authMiddleware, authHandler.Logout)
		auth.POST("/logout", authHandler.Logout)
	}
}
