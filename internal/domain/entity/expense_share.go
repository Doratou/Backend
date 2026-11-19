package entity

type ExpenseShare struct {
	BaseEntity
	ExpenseID        int64        `gorm:"column:expense_id;not null;uniqueIndex:uq_expense_shares_expense_member" json:"expense_id"`
	GroupID          int64        `gorm:"column:group_id;not null" json:"-"`
	MemberID         int64        `gorm:"column:member_id;not null;uniqueIndex:uq_expense_shares_expense_member" json:"member_id"`
	ShareAmountToman int64        `gorm:"column:share_amount_toman;not null" json:"share_amount_toman"`
	SplitValue       int64        `gorm:"column:split_value;not null" json:"-"`
	Member           *GroupMember `gorm:"foreignKey:MemberID" json:"-"`
}

func (ExpenseShare) TableName() string { return "expense_shares" }
