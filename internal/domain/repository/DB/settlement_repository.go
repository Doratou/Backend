package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type SettlementFilter struct {
	Status    *enum.SettlementStatus
	Direction enum.SettlementDirection
}

type SettlementRepository interface {
	FindGroupByID(ctx context.Context, id int64) (*entity.Group, error)
	FindMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error)
	FindMembersByIDs(ctx context.Context, groupID int64, ids []int64) ([]*entity.GroupMember, error)
	Create(ctx context.Context, payment *entity.Settlement) error
	FindByID(ctx context.Context, groupID, id int64, lock bool) (*entity.Settlement, error)
	FindGroupID(ctx context.Context, id int64) (int64, error)
	List(ctx context.Context, groupID, memberID int64, filter SettlementFilter, limit, offset int) ([]*entity.Settlement, int64, error)
	Respond(ctx context.Context, groupID, id int64, status enum.SettlementStatus) error
	DeletePending(ctx context.Context, groupID, id int64) error
	FindAttachmentsByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error)
	BindAttachments(ctx context.Context, groupID, id int64, ids []int64) error
	NetDebt(ctx context.Context, groupID, memberID int64) (string, error)
}
