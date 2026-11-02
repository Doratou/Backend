package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type Friendship struct {
	BaseEntity
	RequesterUserID int64                 `gorm:"column:requester_user_id;not null" json:"requester_user_id"`
	ReceiverUserID  int64                 `gorm:"column:receiver_user_id;not null" json:"receiver_user_id"`
	RequesterUser   *User                 `gorm:"foreignKey:RequesterUserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"requester_user,omitempty"`
	ReceiverUser    *User                 `gorm:"foreignKey:ReceiverUserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"receiver_user,omitempty"`
	Status          enum.FriendshipStatus `gorm:"column:status;type:friendship_status;not null;default:'PENDING'" json:"status"`
	RequestedAt     time.Time             `gorm:"column:requested_at;not null;default:CURRENT_TIMESTAMP" json:"requested_at"`
	RespondedAt     *time.Time            `gorm:"column:responded_at" json:"responded_at"`
}

func (Friendship) TableName() string {
	return "friendships"
}
