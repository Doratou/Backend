package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes wires the authentication and OTP verification endpoints onto the router group.
func RegisterAuthRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		// Public OTP endpoints
		auth.POST("/otp/send", authHandler.SendOTP)
		auth.POST("/otp/verify", authHandler.VerifyOTP)

		// Public registration & login endpoints
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/login/email", authHandler.LoginWithEmail)
		auth.POST("/login/phone", authHandler.LoginWithPhone)
		auth.POST("/password/reset", authHandler.ResetPassword)

		// Token refresh
		auth.POST("/refresh", authHandler.Refresh)

		// Protected endpoints (require valid access token)
		if authMiddleware != nil {
			auth.POST("/logout", authMiddleware, authHandler.Logout)
			auth.POST("/password/change", authMiddleware, authHandler.ChangePassword)
		} else {
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/password/change", authHandler.ChangePassword)
		}
	}
}
