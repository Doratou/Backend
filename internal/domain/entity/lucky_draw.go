package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type LuckyDrawCandidate struct {
	Title  string `json:"title"`
	Weight int64  `json:"weight"`
}

type LuckyDraw struct {
	BaseEntity
	GroupID           int64                `gorm:"column:group_id;not null" json:"group_id"`
	CreatedByMemberID int64                `gorm:"column:created_by_member_id;not null" json:"created_by_member_id"`
	Group             *Group               `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	CreatedByMember   *GroupMember         `gorm:"foreignKey:CreatedByMemberID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Title             string               `gorm:"column:title;size:200;not null" json:"title"`
	Candidates        []LuckyDrawCandidate `gorm:"column:candidates;type:jsonb;serializer:json;not null" json:"candidates"`
	Status            enum.LuckyDrawStatus `gorm:"column:status;type:lucky_draw_status;not null;default:'DRAFT'" json:"status"`
	ResultIndex       *int                 `gorm:"column:result_index" json:"result_index"`
	CompletedAt       *time.Time           `gorm:"column:completed_at" json:"completed_at"`
}

func (LuckyDraw) TableName() string { return "lucky_draws" }
