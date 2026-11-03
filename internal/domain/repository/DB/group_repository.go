package DB

import (
	"context"
	"github.com/Doratou/Backend/internal/domain/enum"

	"github.com/Doratou/Backend/internal/domain/entity"
)

type GroupRepository interface {
	CreateWithOwner(ctx context.Context, group *entity.Group) error
	FindByID(ctx context.Context, id int64) (*entity.Group, int64, error)
	FindByMember(ctx context.Context, userID int64, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error)
	FindAll(ctx context.Context, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error)
	IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error)
	Update(ctx context.Context, group *entity.Group) error
	SetStatus(ctx context.Context, groupID int64, status enum.GroupStatus) error
	Delete(ctx context.Context, groupID int64) error
	CloneWithMembers(ctx context.Context, sourceGroupID int64, newGroup *entity.Group) error
}
