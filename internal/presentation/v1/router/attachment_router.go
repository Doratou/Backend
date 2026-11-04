package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegisterAttachmentRoutes wires the attachment upload, download, and management endpoints onto the router group.
func RegisterAttachmentRoutes(
	rg *gin.RouterGroup,
	attachmentHandler *handler.AttachmentHandler,
	authMiddleware gin.HandlerFunc,
) {
	if attachmentHandler == nil {
		return
	}

	attachments := rg.Group("/attachments")
	{
		// Public read & access endpoints
		attachments.GET("", attachmentHandler.List)
		attachments.GET("/:id", attachmentHandler.GetByID)
		attachments.GET("/:id/url", attachmentHandler.GetURL)
		attachments.GET("/:id/download", attachmentHandler.DownloadStream)

		// Protected endpoints (require valid access token)
		if authMiddleware != nil {
			attachments.POST("/upload", authMiddleware, attachmentHandler.Upload)
			attachments.POST("/presigned-upload", authMiddleware, attachmentHandler.CreatePresignedUpload)
			attachments.POST("/confirm-upload", authMiddleware, attachmentHandler.ConfirmPresignedUpload)
			attachments.PATCH("/:id", authMiddleware, attachmentHandler.Update)
			attachments.DELETE("/:id", authMiddleware, attachmentHandler.Delete)
		} else {
			attachments.POST("/upload", attachmentHandler.Upload)
			attachments.POST("/presigned-upload", attachmentHandler.CreatePresignedUpload)
			attachments.POST("/confirm-upload", attachmentHandler.ConfirmPresignedUpload)
			attachments.PATCH("/:id", attachmentHandler.Update)
			attachments.DELETE("/:id", attachmentHandler.Delete)
		}
	}
}
