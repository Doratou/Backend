package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes wires the authentication and OTP verification endpoints onto the router group.
func RegisterAuthRoutes(
	rg *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	authMiddleware gin.HandlerFunc,
	captchaMiddleware gin.HandlerFunc,
) {
	auth := rg.Group("/auth")
	{
		// Public OTP endpoints (protected by CaptchaGuard if configured)
		if captchaMiddleware != nil {
			auth.POST("/otp/send", captchaMiddleware, authHandler.SendOTP)
		} else {
			auth.POST("/otp/send", authHandler.SendOTP)
		}
		auth.POST("/otp/verify", authHandler.VerifyOTP)

		// Public registration & login endpoints
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/login/otp", authHandler.LoginWithOTP)

		// Password reset (protected by CaptchaGuard if configured)
		if captchaMiddleware != nil {
			auth.POST("/password/reset", captchaMiddleware, authHandler.ResetPassword)
		} else {
			auth.POST("/password/reset", authHandler.ResetPassword)
		}

		// Token refresh
		auth.POST("/refresh", authHandler.Refresh)

		// Protected endpoints (require valid access token)
		auth.POST("/logout", authMiddleware, authHandler.Logout)
		auth.POST("/password/change", authMiddleware, authHandler.ChangePassword)
	}
}
