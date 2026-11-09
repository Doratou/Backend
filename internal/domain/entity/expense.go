package entity

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
)

// Expense represents a shared expenditure record in a group.
type Expense struct {
	BaseEntity
	GroupID           int64            `gorm:"column:group_id;not null" json:"group_id"`
	CreatedByMemberID int64            `gorm:"column:created_by_member_id;not null" json:"created_by_member_id"`
	Title             string           `gorm:"column:title;size:200;not null" json:"title"`
	Description       *string          `gorm:"column:description;type:text" json:"description,omitempty"`
	TotalAmountToman  int64            `gorm:"column:total_amount_toman;not null;default:0" json:"total_amount_toman"`
	ExpenseType       enum.ExpenseType `gorm:"column:expense_type;type:expense_type;not null;default:'NORMAL'" json:"expense_type"`
	OccurredAt        *time.Time       `gorm:"column:occurred_at" json:"occurred_at,omitempty"`
	Group             *Group           `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"group,omitempty"`
	CreatedByMember   *GroupMember     `gorm:"foreignKey:CreatedByMemberID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"created_by_member,omitempty"`
	Attachments       []Attachment     `gorm:"foreignKey:ExpenseID" json:"attachments,omitempty"`
}

func (Expense) TableName() string {
	return "expenses"
}
