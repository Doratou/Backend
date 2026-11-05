package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterInviteRoutes(rg *gin.RouterGroup, h *handler.InviteHandler, authMiddleware gin.HandlerFunc) {
	invites := rg.Group("/invites")
	{
		direct := invites.Group("/direct")
		{
			direct.POST("", authMiddleware, h.SendDirect)
			direct.GET("/me", authMiddleware, h.ListMyDirectInvites)
			direct.GET("/groups/:groupID", authMiddleware, h.ListGroupDirectInvites)
			direct.POST("/:inviteID/respond", authMiddleware, h.RespondToDirect)
		}

		links := invites.Group("/links")
		{
			links.POST("/groups/:groupID", authMiddleware, h.CreateLink)
			links.GET("/groups/:groupID", authMiddleware, h.ListGroupLinks)
			links.DELETE("/groups/:groupID/:inviteID", authMiddleware, h.DeactivateLink)
			links.PATCH("/groups/:groupID/:inviteID", authMiddleware, h.UpdateLink)
			links.PATCH("/groups/:groupID/:inviteID/password", authMiddleware, h.ChangeLinkPassword)

			links.POST("/:token/join", authMiddleware, h.JoinByLink)
		}
	}
}
