package entity

import "github.com/Doratou/Backend/internal/domain/enum"

type ChecklistItem struct {
	BaseEntity
	GroupID            int64                  `gorm:"column:group_id;not null"`
	Title              string                 `gorm:"column:title;size:200;not null"`
	Type               enum.ChecklistItemType `gorm:"column:type;type:checklist_item_type;not null"`
	CreatedByMemberID  int64                  `gorm:"column:created_by_member_id;not null"`
	AssignedToMemberID *int64                 `gorm:"column:assigned_to_member_id"`
	CreatedByMember    *GroupMember           `gorm:"foreignKey:CreatedByMemberID"`
	AssignedToMember   *GroupMember           `gorm:"foreignKey:AssignedToMemberID"`
}
