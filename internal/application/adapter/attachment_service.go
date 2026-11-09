package adapter

import (
	"context"
	"io"
	"mime/multipart"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// AttachmentService defines the application boundary for handling media and attachments.
type AttachmentService interface {
	// --- Single-item operations ---

	// Upload processes a multipart file, validates mime/size/magic bytes, uploads to storage, and persists metadata.
	Upload(ctx context.Context, userID int64, fileHeader *multipart.FileHeader, req dto.UploadAttachmentRequest) (*dto.AttachmentResponse, error)

	// CreatePresignedUpload generates a presigned PUT URL enabling the client to upload directly to storage.
	CreatePresignedUpload(ctx context.Context, userID int64, req dto.PresignedUploadRequest) (*dto.PresignedUploadResponse, error)

	// ConfirmPresignedUpload verifies the uploaded object in storage, calculates stats, and creates a database record.
	ConfirmPresignedUpload(ctx context.Context, userID int64, req dto.ConfirmPresignedUploadRequest) (*dto.AttachmentResponse, error)

	// GetByID retrieves attachment metadata and generates an active access URL.
	GetByID(ctx context.Context, id int64) (*dto.AttachmentResponse, error)

	// GetByIDWithRelations retrieves attachment metadata with associated User, Group, and Expense preloaded.
	GetByIDWithRelations(ctx context.Context, id int64) (*dto.AttachmentResponse, error)

	// GetByGroup retrieves attachments associated with a specific group, with optional type filtering.
	GetByGroup(ctx context.Context, groupID int64, attType *enum.AttachmentType, page, pageSize int) (*dto.AttachmentListResponse, error)

	// GetByExpense retrieves attachments tied to a specific expense.
	GetByExpense(ctx context.Context, expenseID int64) ([]*dto.AttachmentResponse, error)

	// GetByObjectKey retrieves attachment metadata by its object storage key.
	GetByObjectKey(ctx context.Context, objectKey string) (*dto.AttachmentResponse, error)

	// List queries attachments based on dynamic filters and pagination.
	List(ctx context.Context, filter dto.AttachmentFilterRequest) (*dto.AttachmentListResponse, error)

	// Update modifies attachment metadata (requires ownership or administrative permission).
	Update(ctx context.Context, userID int64, id int64, req dto.UpdateAttachmentRequest) (*dto.AttachmentResponse, error)

	// Delete deletes an attachment. If hardDelete is true, the object in storage is also purged.
	Delete(ctx context.Context, userID int64, id int64, hardDelete bool) error

	// GetFreshAccessURL generates a fresh presigned URL or direct public URL.
	GetFreshAccessURL(ctx context.Context, id int64, expiry time.Duration) (string, error)

	// DownloadStream retrieves a raw file stream from storage for reverse-proxy downloading.
	DownloadStream(ctx context.Context, id int64) (io.ReadCloser, *dto.FileStreamMeta, error)

	// --- Batch / Bulk operations ---

	// UploadBatch uploads multiple multipart files concurrently, with partial-success tolerance.
	UploadBatch(ctx context.Context, userID int64, files []*multipart.FileHeader, defaultType enum.AttachmentType, groupID, expenseID *int64, isPublic *bool) (*dto.BatchUploadResponse, error)

	// CreatePresignedUploadBatch generates direct upload PUT URLs for multiple items.
	CreatePresignedUploadBatch(ctx context.Context, userID int64, req dto.BatchPresignedUploadRequest) (*dto.BatchPresignedUploadResponse, error)

	// ConfirmPresignedUploadBatch confirms and creates database records for multiple uploaded items.
	ConfirmPresignedUploadBatch(ctx context.Context, userID int64, req dto.BatchConfirmUploadRequest) (*dto.BatchUploadResponse, error)

	// GetBatchURLs generates active access URLs for multiple attachment IDs.
	GetBatchURLs(ctx context.Context, ids []int64, expiry time.Duration) (*dto.BatchURLsResponse, error)

	// DownloadZipStream streams an on-the-fly zip archive containing the requested attachments.
	DownloadZipStream(ctx context.Context, ids []int64, w io.Writer) error

	// UpdateBatch updates common metadata fields for multiple attachments owned by the user.
	UpdateBatch(ctx context.Context, userID int64, req dto.BatchUpdateAttachmentRequest) error

	// DeleteBatch deletes multiple attachments (soft or hard).
	DeleteBatch(ctx context.Context, userID int64, ids []int64, hardDelete bool) error
}
