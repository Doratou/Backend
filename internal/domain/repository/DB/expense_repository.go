package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
)

type ExpenseRepository interface {
	CreateWithShares(ctx context.Context, expense *entity.Expense) error
	UpdateWithShares(ctx context.Context, expense *entity.Expense) error
	DeleteByID(ctx context.Context, groupID, expenseID int64) error
	FindByID(ctx context.Context, groupID, expenseID int64) (*entity.Expense, error)
	FindByGroupID(ctx context.Context, groupID int64, limit, offset int) ([]*entity.Expense, int64, error)
	FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error)
	FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error)
	FindMembersByIDs(ctx context.Context, groupID int64, memberIDs []int64) ([]*entity.GroupMember, error)
	FindParticipants(ctx context.Context, groupID int64, limit, offset int) ([]*entity.GroupMember, int64, error)
	FindGroupIDByExpenseID(ctx context.Context, expenseID int64) (int64, error)
	FindAttachmentsByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error)
	ReplaceAttachments(ctx context.Context, groupID, expenseID int64, ids []int64) error
}
