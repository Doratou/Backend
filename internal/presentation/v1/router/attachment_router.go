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
	optionalAuthMiddleware gin.HandlerFunc,
) {

	attachments := rg.Group("/attachments")
	{
		// 1. Batch / Bulk routes (Registered first to avoid :id collisions)
		batch := attachments.Group("/batch")
		{
			// Batch download & URL endpoints (require auth for access verification)
			batch.POST("/urls", authMiddleware, attachmentHandler.GetBatchURLs)
			batch.POST("/download", authMiddleware, attachmentHandler.DownloadZipStream)

			// Protected batch upload endpoints
			batch.POST("/upload", authMiddleware, attachmentHandler.UploadBatch)
			batch.POST("/presigned-upload", authMiddleware, attachmentHandler.CreatePresignedUploadBatch)
			batch.POST("/confirm-upload", authMiddleware, attachmentHandler.ConfirmPresignedUploadBatch)
			batch.DELETE("", authMiddleware, attachmentHandler.DeleteBatch)
		}

		// 2. Relation query endpoints (Registered before :id to prevent wildcard collisions)
		attachments.GET("/group/:group_id", authMiddleware, attachmentHandler.GetByGroup)
		attachments.GET("/expense/:expense_id", authMiddleware, attachmentHandler.GetByExpense)

		// 3. Single-item access endpoints (Public by default, Context Guard enforced for private files)
		attachments.GET("/:id", optionalAuthMiddleware, attachmentHandler.GetByID)
		attachments.GET("/:id/download", optionalAuthMiddleware, attachmentHandler.Download)

		// 4. Protected single-item upload & delete endpoints
		attachments.POST("/upload", authMiddleware, attachmentHandler.Upload)
		attachments.POST("/presigned-upload", authMiddleware, attachmentHandler.CreatePresignedUpload)
		attachments.POST("/confirm-upload", authMiddleware, attachmentHandler.ConfirmPresignedUpload)
		attachments.DELETE("/:id", authMiddleware, attachmentHandler.Delete)
	}
}
