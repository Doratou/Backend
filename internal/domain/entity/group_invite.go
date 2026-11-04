package entity

import "time"

// GroupInvite stores a lookup hash and an encrypted copy of the invite token.
type GroupInvite struct {
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

func (GroupInvite) TableName() string { return "group_invites" }
