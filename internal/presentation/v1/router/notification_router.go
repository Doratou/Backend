package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(rg *gin.RouterGroup, notificationHandler *handler.NotificationHandler, authMiddleware gin.HandlerFunc) {
	if notificationHandler == nil {
		panic("notification handler is required")
	}
	if authMiddleware == nil {
		panic("notification routes require authentication middleware")
	}

	notifications := rg.Group("/notifications", authMiddleware)
	{
		notifications.GET("", notificationHandler.GetNotifications)
		notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
		notifications.GET("/sync", notificationHandler.SyncNotifications)
		notifications.GET("/:id", notificationHandler.GetNotificationByID)
		notifications.POST("/read", notificationHandler.MarkAsRead)
		notifications.POST("/:id/read", notificationHandler.MarkSingleAsRead)
		notifications.DELETE("/:id", notificationHandler.DeleteNotification)
		notifications.DELETE("", notificationHandler.ClearNotifications)
	}

	groups := rg.Group("/groups", authMiddleware)
	{
		groups.POST("/:groupID/announcements", notificationHandler.CreateGroupAnnouncement)
	}
}
