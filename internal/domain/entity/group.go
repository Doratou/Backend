package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// Group is a shared space created by a user.
type Group struct {
	BaseEntity
	Title             string           `gorm:"column:title;size:150;not null" json:"title"`
	Description       *string          `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedByUserID   int64            `gorm:"column:created_by_user_id;not null;index" json:"created_by_user_id"`
	CoverAttachmentID *int64           `gorm:"column:cover_attachment_id" json:"cover_attachment_id,omitempty"`
	CoverAttachment   *Attachment      `gorm:"foreignKey:CoverAttachmentID" json:"cover_attachment,omitempty"`
	Status            enum.GroupStatus `gorm:"column:status;type:group_status;not null;default:ACTIVE" json:"status"`
	IsPremium         bool             `gorm:"column:is_premium;not null;default:false" json:"is_premium"`
	ArchivedAt        *time.Time       `gorm:"column:archived_at" json:"archived_at,omitempty"`
	Members           []GroupMember    `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

func (Group) TableName() string { return "groups" }
