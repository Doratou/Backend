package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type Poll struct {
	BaseEntity
	GroupID           int64               `gorm:"column:group_id;not null" json:"group_id"`
	CreatedByMemberID int64               `gorm:"column:created_by_member_id;not null" json:"created_by_member_id"`
	Group             *Group              `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	CreatedByMember   *GroupMember        `gorm:"foreignKey:CreatedByMemberID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Question          string              `gorm:"column:question;size:500;not null" json:"question"`
	Visibility        enum.PollVisibility `gorm:"column:visibility;type:poll_visibility;not null;default:'PRIVATE'" json:"visibility"`
	Status            enum.PollStatus     `gorm:"column:status;type:poll_status;not null;default:'ACTIVE'" json:"status"`
	ClosedAt          *time.Time          `gorm:"column:closed_at" json:"closed_at"`
	Options           []*PollOption       `gorm:"foreignKey:PollID" json:"options,omitempty"`
}

func (Poll) TableName() string {
	return "polls"
}
