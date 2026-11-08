package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// LinkInvite stores a lookup hash and an encrypted copy of the invite token.
type LinkInvite struct {
	ID                int64      `gorm:"primaryKey;autoIncrement"`
	GroupID           int64      `gorm:"column:group_id;not null;index"`
	CreatedByMemberID int64      `gorm:"column:created_by_member_id;not null"`
	TokenHash         string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex"`
	TokenCiphertext   string     `gorm:"column:token_ciphertext;type:text"`
	PasswordHash      string     `gorm:"column:password_hash;type:text;not null"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	MaxUses           *int       `gorm:"column:max_uses"`
	UsedCount         int        `gorm:"column:used_count;not null;default:0"`
	RevokedAt         *time.Time `gorm:"column:revoked_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

func (LinkInvite) TableName() string { return "group_invites" }

// DirectInvite is a direct, owner-to-friend invitation to join a group.
type DirectInvite struct {
	ID              int64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID         int64                   `gorm:"column:group_id;not null;index" json:"group_id"`
	SenderUserID    int64                   `gorm:"column:sender_user_id;not null;index" json:"sender_user_id"`
	RecipientUserID int64                   `gorm:"column:recipient_user_id;not null;index" json:"recipient_user_id"`
	Status          enum.DirectInviteStatus `gorm:"column:status;type:group_invite_status;not null;default:'PENDING';index" json:"status"`
	CreatedAt       time.Time               `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	RespondedAt     *time.Time              `gorm:"column:responded_at" json:"responded_at,omitempty"`
	Group           *Group                  `gorm:"foreignKey:GroupID" json:"-"`
	Sender          *User                   `gorm:"foreignKey:SenderUserID" json:"-"`
	Recipient       *User                   `gorm:"foreignKey:RecipientUserID" json:"-"`
}

func (DirectInvite) TableName() string { return "group_member_invites" }
