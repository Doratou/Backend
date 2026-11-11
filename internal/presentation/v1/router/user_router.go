package router

import (
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
	users := rg.Group("/users", authMiddleware)
	{
		users.GET("/me", userHandler.GetMe)
		users.GET("/by-friend-code/:friendCode", userHandler.FindByFriendCode)
	}
}
