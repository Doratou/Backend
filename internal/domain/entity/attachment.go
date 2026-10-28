package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// Attachment represents uploaded media or files (avatars, covers, receipts, etc.).
type Attachment struct {
	BaseEntity
	UploadedByUserID *int64              `gorm:"column:uploaded_by_user_id" json:"uploaded_by_user_id,omitempty"`
	GroupID          *int64              `gorm:"column:group_id" json:"group_id,omitempty"`
	ExpenseID        *int64              `gorm:"column:expense_id" json:"expense_id,omitempty"`
	Type             enum.AttachmentType `gorm:"column:type;type:attachment_type;not null" json:"type"`
	OriginalFilename *string             `gorm:"column:original_filename;size:255" json:"original_filename,omitempty"`
	ObjectKey        *string             `gorm:"column:object_key" json:"object_key,omitempty"`
	MimeType         *string             `gorm:"column:mime_type;size:100" json:"mime_type,omitempty"`
	FileSizeBytes    *int64              `gorm:"column:file_size_bytes" json:"file_size_bytes,omitempty"`
	UploadedAt       *time.Time          `gorm:"column:uploaded_at" json:"uploaded_at,omitempty"`
}

func (Attachment) TableName() string {
	return "attachments"
}
