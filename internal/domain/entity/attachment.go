package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// Attachment represents uploaded media or files (avatars, covers, receipts, etc.).
type Attachment struct {
	BaseEntity
	UploadedByUserID   *int64              `gorm:"column:uploaded_by_user_id" json:"uploaded_by_user_id,omitempty"`
	GroupID            *int64              `gorm:"column:group_id" json:"group_id,omitempty"`
	ExpenseID          *int64              `gorm:"column:expense_id" json:"expense_id,omitempty"`
	Type               enum.AttachmentType `gorm:"column:type;type:attachment_type;not null" json:"type"`
	OriginalFilename   *string             `gorm:"column:original_filename;size:255" json:"original_filename,omitempty"`
	ObjectKey          *string             `gorm:"column:object_key;not null" json:"object_key,omitempty"`
	ThumbnailObjectKey *string             `gorm:"column:thumbnail_object_key" json:"thumbnail_object_key,omitempty"`
	MimeType           *string             `gorm:"column:mime_type;size:100" json:"mime_type,omitempty"`
	FileSizeBytes      *int64              `gorm:"column:file_size_bytes" json:"file_size_bytes,omitempty"`
	FileHash           *string             `gorm:"column:file_hash;size:64" json:"file_hash,omitempty"`
	Width              *int                `gorm:"column:width" json:"width,omitempty"`
	Height             *int                `gorm:"column:height" json:"height,omitempty"`
	DurationSeconds    *int                `gorm:"column:duration_seconds" json:"duration_seconds,omitempty"`
	IsPublic           bool                `gorm:"column:is_public;default:false;not null" json:"is_public"`
	UploadedAt         *time.Time          `gorm:"column:uploaded_at" json:"uploaded_at,omitempty"`
}

func (Attachment) TableName() string {
	return "attachments"
}
