package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type ChecklistItemCount struct {
	Type               enum.ChecklistItemType
	AssignedToMemberID *int64
	ItemCount          int64
}

type ChecklistRepository interface {
	FindGroup(ctx context.Context, groupID int64, lock bool) (*entity.Group, error)
	FindActiveMember(ctx context.Context, groupID, userID int64, lock bool) (*entity.GroupMember, error)
	ListActiveMembers(ctx context.Context, groupID int64) ([]entity.GroupMember, error)
	ListItems(ctx context.Context, groupID, memberID int64, filter enum.ChecklistFilter, limit, offset int) ([]entity.ChecklistItem, error)
	CountItems(ctx context.Context, groupID int64) ([]ChecklistItemCount, error)
	FindItemForUpdate(ctx context.Context, groupID, itemID int64) (*entity.ChecklistItem, error)
	CreateItem(ctx context.Context, item *entity.ChecklistItem) error
	SetAssignee(ctx context.Context, item *entity.ChecklistItem, memberID *int64) error
	DeleteItem(ctx context.Context, item *entity.ChecklistItem) error
}
