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

	attachments := rg.Group("/attachments")
	{
		// 1. Batch / Bulk routes (Registered first to avoid :id collisions)
		batch := attachments.Group("/batch")
		{
			// Public batch read/download endpoints
			batch.POST("/urls", attachmentHandler.GetBatchURLs)
			batch.POST("/download", attachmentHandler.DownloadZipStream)

			// Protected batch endpoints
			batch.POST("/upload", authMiddleware, attachmentHandler.UploadBatch)
			batch.POST("/presigned-upload", authMiddleware, attachmentHandler.CreatePresignedUploadBatch)
			batch.POST("/confirm-upload", authMiddleware, attachmentHandler.ConfirmPresignedUploadBatch)
			batch.PATCH("", authMiddleware, attachmentHandler.UpdateBatch)
			batch.DELETE("", authMiddleware, attachmentHandler.DeleteBatch)

		}

		// 2. Single-item endpoints
		attachments.GET("", attachmentHandler.List)
		attachments.GET("/:id", attachmentHandler.GetByID)
		attachments.GET("/:id/url", attachmentHandler.GetURL)
		attachments.GET("/:id/download", attachmentHandler.DownloadStream)

		// Protected single-item endpoints (require valid access token)
		attachments.POST("/upload", authMiddleware, attachmentHandler.Upload)
		attachments.POST("/presigned-upload", authMiddleware, attachmentHandler.CreatePresignedUpload)
		attachments.POST("/confirm-upload", authMiddleware, attachmentHandler.ConfirmPresignedUpload)
		attachments.PATCH("/:id", authMiddleware, attachmentHandler.Update)
		attachments.DELETE("/:id", authMiddleware, attachmentHandler.Delete)

	}
}
