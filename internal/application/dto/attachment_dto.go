package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// UploadAttachmentRequest contains parameters sent along with a multipart file.
type UploadAttachmentRequest struct {
	Type      enum.AttachmentType `form:"type" binding:"required"`
	GroupID   *int64              `form:"group_id"`
	ExpenseID *int64              `form:"expense_id"`
	IsPublic  *bool               `form:"is_public"`
}

// PresignedUploadRequest requests a direct PUT upload URL for client-side uploads.
type PresignedUploadRequest struct {
	Filename      string              `json:"filename" binding:"required"`
	Type          enum.AttachmentType `json:"type" binding:"required"`
	FileSizeBytes int64               `json:"file_size_bytes" binding:"required,gt=0"`
	MimeType      string              `json:"mime_type" binding:"required"`
	GroupID       *int64              `json:"group_id"`
	ExpenseID     *int64              `json:"expense_id"`
	IsPublic      *bool               `json:"is_public"`
}

// PresignedUploadResponse returns the direct upload URL and the designated object key.
type PresignedUploadResponse struct {
	UploadURL        string `json:"upload_url"`
	ObjectKey        string `json:"object_key"`
	ExpiresInSeconds int64  `json:"expires_in_seconds"`
}

// ConfirmPresignedUploadRequest is called after a client completes direct upload to storage.
type ConfirmPresignedUploadRequest struct {
	ObjectKey        string              `json:"object_key" binding:"required"`
	OriginalFilename string              `json:"original_filename" binding:"required"`
	Type             enum.AttachmentType `json:"type" binding:"required"`
	GroupID          *int64              `json:"group_id"`
	ExpenseID        *int64              `json:"expense_id"`
	IsPublic         *bool               `json:"is_public"`
}

// UpdateAttachmentRequest modifies an attachment's metadata.
type UpdateAttachmentRequest struct {
	OriginalFilename *string              `json:"original_filename" binding:"omitempty,min=1,max=255"`
	GroupID          *int64               `json:"group_id"`
	ExpenseID        *int64               `json:"expense_id"`
	Type             *enum.AttachmentType `json:"type"`
	IsPublic         *bool                `json:"is_public"`
}

// AttachmentFilterRequest contains query parameters for filtering attachments.
type AttachmentFilterRequest struct {
	Type             *enum.AttachmentType `form:"type"`
	GroupID          *int64               `form:"group_id"`
	ExpenseID        *int64               `form:"expense_id"`
	UploadedByUserID *int64               `form:"uploaded_by_user_id"`
	IsPublic         *bool                `form:"is_public"`
	Search           string               `form:"search"`
	Page             int                  `form:"page,default=1"`
	PageSize         int                  `form:"page_size,default=20"`
}

// AttachmentResponse provides full attachment details and an access URL.
type AttachmentResponse struct {
	ID                 int64               `json:"id"`
	UploadedByUserID   *int64              `json:"uploaded_by_user_id,omitempty"`
	GroupID            *int64              `json:"group_id,omitempty"`
	ExpenseID          *int64              `json:"expense_id,omitempty"`
	Type               enum.AttachmentType `json:"type"`
	OriginalFilename   *string             `json:"original_filename,omitempty"`
	ObjectKey          string              `json:"object_key"`
	ThumbnailObjectKey *string             `json:"thumbnail_object_key,omitempty"`
	MimeType           *string             `json:"mime_type,omitempty"`
	FileSizeBytes      *int64              `json:"file_size_bytes,omitempty"`
	FileHash           *string             `json:"file_hash,omitempty"`
	Width              *int                `json:"width,omitempty"`
	Height             *int                `json:"height,omitempty"`
	DurationSeconds    *int                `json:"duration_seconds,omitempty"`
	IsPublic           bool                `json:"is_public"`
	AccessURL          string              `json:"access_url"`
	UploadedAt         *time.Time          `json:"uploaded_at,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// AttachmentListResponse encapsulates paginated attachment items.
type AttachmentListResponse struct {
	Items      []*AttachmentResponse `json:"items"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// FileStreamMeta holds response headers for proxy streaming an attachment download.
type FileStreamMeta struct {
	Filename    string
	ContentType string
	Size        int64
}

// ToAttachmentResponse converts an entity.Attachment to an AttachmentResponse DTO.
func ToAttachmentResponse(entity *entity.Attachment, accessURL string) *AttachmentResponse {
	if entity == nil {
		return nil
	}

	return &AttachmentResponse{
		ID:                 entity.ID,
		UploadedByUserID:   entity.UploadedByUserID,
		GroupID:            entity.GroupID,
		ExpenseID:          entity.ExpenseID,
		Type:               entity.Type,
		OriginalFilename:   entity.OriginalFilename,
		ObjectKey:          *entity.ObjectKey,
		ThumbnailObjectKey: entity.ThumbnailObjectKey,
		MimeType:           entity.MimeType,
		FileSizeBytes:      entity.FileSizeBytes,
		FileHash:           entity.FileHash,
		Width:              entity.Width,
		Height:             entity.Height,
		DurationSeconds:    entity.DurationSeconds,
		IsPublic:           entity.IsPublic,
		AccessURL:          accessURL,
		UploadedAt:         entity.UploadedAt,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}
