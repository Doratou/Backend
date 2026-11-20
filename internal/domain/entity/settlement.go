package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

type Settlement struct {
	BaseEntity
	GroupID         int64                 `gorm:"column:group_id;not null" json:"group_id"`
	FromMemberID    int64                 `gorm:"column:from_member_id;not null" json:"from_member_id"`
	ToMemberID      int64                 `gorm:"column:to_member_id;not null" json:"to_member_id"`
	AmountToman     int64                 `gorm:"column:amount_toman;not null" json:"amount_toman"`
	Status          enum.SettlementStatus `gorm:"column:status;type:settlement_status;not null;default:'PENDING'" json:"status"`
	Note            *string               `gorm:"column:note;type:text" json:"note"`
	PaidAt          time.Time             `gorm:"column:paid_at;not null" json:"paid_at"`
	ConfirmedAt     *time.Time            `gorm:"column:confirmed_at" json:"confirmed_at"`
	RejectedAt      *time.Time            `gorm:"column:rejected_at" json:"rejected_at"`
	CanceledAt      *time.Time            `gorm:"column:canceled_at" json:"canceled_at"`
	BalanceRevision *string               `gorm:"column:balance_revision" json:"-"`
	FromMember      *GroupMember          `gorm:"foreignKey:FromMemberID" json:"-"`
	ToMember        *GroupMember          `gorm:"foreignKey:ToMemberID" json:"-"`
	Attachments     []Attachment          `gorm:"foreignKey:SettlementID" json:"-"`
}

func (Settlement) TableName() string { return "settlements" }
