package adapter

import (
	"context"
	"io"
	"mime/multipart"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
)

// AttachmentService defines the application boundary for handling media and attachments.
type AttachmentService interface {
	// Upload processes a multipart file, validates mime/size/magic bytes, uploads to storage, and persists metadata.
	Upload(ctx context.Context, userID int64, fileHeader *multipart.FileHeader, req dto.UploadAttachmentRequest) (*dto.AttachmentResponse, error)

	// CreatePresignedUpload generates a presigned PUT URL enabling the client to upload directly to storage.
	CreatePresignedUpload(ctx context.Context, userID int64, req dto.PresignedUploadRequest) (*dto.PresignedUploadResponse, error)

	// ConfirmPresignedUpload verifies the uploaded object in storage, calculates stats, and creates a database record.
	ConfirmPresignedUpload(ctx context.Context, userID int64, req dto.ConfirmPresignedUploadRequest) (*dto.AttachmentResponse, error)

	// GetByID retrieves attachment metadata and generates an active access URL.
	GetByID(ctx context.Context, id int64) (*dto.AttachmentResponse, error)

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
}
