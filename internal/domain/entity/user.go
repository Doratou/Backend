package entity

import (
	"github.com/Doratou/Backend/internal/domain/enum"
)

// User represents an authenticated account in Doratou.
type User struct {
	BaseEntity
	Name               string          `gorm:"column:name;size:100;not null" json:"name"`
	Email              *string         `gorm:"column:email;size:255;uniqueIndex" json:"email,omitempty"`
	Phone              *string         `gorm:"column:phone;size:16;uniqueIndex" json:"phone,omitempty"`
	PasswordHash       string          `gorm:"column:password_hash;not null" json:"-"`
	IsEmailVerified    bool            `gorm:"column:is_email_verified;not null;default:false" json:"is_email_verified"`
	IsPhoneVerified    bool            `gorm:"column:is_phone_verified;not null;default:false" json:"is_phone_verified"`
	AvatarAttachmentID *int64          `gorm:"column:avatar_attachment_id" json:"avatar_attachment_id,omitempty"`
	AvatarAttachment   *Attachment     `gorm:"foreignKey:AvatarAttachmentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"avatar_attachment,omitempty"`
	SystemRole         enum.UserRole   `gorm:"column:system_role;type:user_role;not null;default:'USER'" json:"system_role"`
	Status             enum.UserStatus `gorm:"column:status;type:user_status;not null;default:'ACTIVE'" json:"status"`
	Wallet             int64           `gorm:"column:wallet;not null;default:0" json:"wallet"`
	TelegramChatID     *int64          `gorm:"column:telegram_chat_id" json:"telegram_chat_id,omitempty"`
	TelegramUserID     *int64          `gorm:"column:telegram_user_id" json:"telegram_user_id,omitempty"`
}

func (User) TableName() string {
	return "users"
}
