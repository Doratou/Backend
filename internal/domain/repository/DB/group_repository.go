package DB

import (
	"context"
	"github.com/Doratou/Backend/internal/domain/enum"

	"github.com/Doratou/Backend/internal/domain/entity"
)

type GroupRepository interface {
	CreateWithOwner(ctx context.Context, group *entity.Group) error
	FindByID(ctx context.Context, id int64) (*entity.Group, int64, error)
	ListActiveMembers(ctx context.Context, groupID int64) ([]entity.GroupMember, error)
	FindActiveMemberByID(ctx context.Context, groupID, memberID int64) (*entity.GroupMember, error)
	SetMemberInactive(ctx context.Context, groupID, userID int64) error
	SetMemberInactiveByID(ctx context.Context, groupID, memberID int64) error
	TransferOwnership(ctx context.Context, groupID, newOwnerID int64) error
	CreateInvite(ctx context.Context, groupID, creatorUserID int64, invite *entity.GroupInvite) error
	FindInviteByID(ctx context.Context, groupID, inviteID int64) (*entity.GroupInvite, error)
	FindInvitesByGroupID(ctx context.Context, groupID int64) ([]entity.GroupInvite, error)
	FindInviteByTokenHash(ctx context.Context, tokenHash string) (*entity.GroupInvite, error)
	RevokeInvite(ctx context.Context, groupID, inviteID int64) error
	UpdateInvitePassword(ctx context.Context, groupID, inviteID int64, passwordHash string) error
	UpdateInvite(ctx context.Context, invite *entity.GroupInvite) error
	JoinGroupWithInvite(ctx context.Context, inviteID, userID int64) (int64, error)
	FindByMember(ctx context.Context, userID int64, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error)
	FindAll(ctx context.Context, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error)
	IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error)
	Update(ctx context.Context, group *entity.Group) error
	SetStatus(ctx context.Context, groupID int64, status enum.GroupStatus) error
	Delete(ctx context.Context, groupID int64) error
	CloneWithMembers(ctx context.Context, sourceGroupID int64, newGroup *entity.Group) error
}
