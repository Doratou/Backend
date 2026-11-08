package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPollRoutes(
	rg *gin.RouterGroup,
	pollHandler *handler.PollHandler,
	authMiddleware gin.HandlerFunc,
) {
	if pollHandler == nil {
		panic("poll handler is required")
	}
	if authMiddleware == nil {
		panic("poll routes require authentication middleware")
	}

	polls := rg.Group("/groups/:groupID/polls", authMiddleware)
	{
		polls.POST("", pollHandler.Create)
		polls.GET("", pollHandler.List)
		polls.GET("/:pollID", pollHandler.Get)
		polls.PUT("/:pollID/vote", pollHandler.Vote)
		polls.DELETE("/:pollID/vote", pollHandler.WithdrawVote)
		polls.GET("/:pollID/voters", pollHandler.ListVoters)
		polls.PATCH("/:pollID/close", pollHandler.Close)
		polls.DELETE("/:pollID", pollHandler.Delete)
	}
}
