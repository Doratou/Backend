package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type LuckyDrawRepository interface {
	Create(ctx context.Context, draw *entity.LuckyDraw) error
	FindByID(ctx context.Context, groupID, drawID int64) (*entity.LuckyDraw, error)
	FindByIDForUpdate(ctx context.Context, groupID, drawID int64) (*entity.LuckyDraw, error)
	FindByGroupID(ctx context.Context, groupID int64, status *enum.LuckyDrawStatus, limit, offset int) ([]*entity.LuckyDraw, int64, error)
	FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error)
	FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error)
	UpdateDraft(ctx context.Context, draw *entity.LuckyDraw) error
	Complete(ctx context.Context, draw *entity.LuckyDraw) error
}
