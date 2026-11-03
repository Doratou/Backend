package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// GroupMember links a user to a group. A unique database index prevents duplicate membership rows.
type GroupMember struct {
	ID       int64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID  int64                  `gorm:"column:group_id;not null;uniqueIndex:idx_group_members_group_user" json:"group_id"`
	UserID   int64                  `gorm:"column:user_id;not null;uniqueIndex:idx_group_members_group_user;index:idx_group_members_user_status" json:"user_id"`
	Status   enum.GroupMemberStatus `gorm:"column:status;size:15;not null;default:ACTIVE;index:idx_group_members_user_status" json:"status"`
	JoinedAt time.Time              `gorm:"column:joined_at;not null;default:CURRENT_TIMESTAMP" json:"joined_at"`
	LeftAt   *time.Time             `gorm:"column:left_at" json:"left_at,omitempty"`
}

func (GroupMember) TableName() string { return "group_members" }
