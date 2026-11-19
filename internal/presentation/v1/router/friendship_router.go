package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterFriendshipRoutes(
	rg *gin.RouterGroup,
	friendshipHandler *handler.FriendshipHandler,
	authMiddleware gin.HandlerFunc,
) {
	if friendshipHandler == nil {
		panic("friendship handler is required")
	}
	if authMiddleware == nil {
		panic("friendship routes require authentication middleware")
	}

	friendships := rg.Group("/friendships", authMiddleware)
	{
		friendships.POST("/requests", friendshipHandler.SendRequest)
		friendships.GET("/requests", friendshipHandler.ListRequests)
		friendships.PATCH("/requests/:id/accept", friendshipHandler.AcceptRequest)
		friendships.PATCH("/requests/:id/reject", friendshipHandler.RejectRequest)
		friendships.DELETE("/requests/:id", friendshipHandler.CancelRequest)
		friendships.GET("", friendshipHandler.ListFriends)
		friendships.DELETE("/:id", friendshipHandler.RemoveFriend)
	}
}
