package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type GroupService interface {
	Create(ctx context.Context, userID int64, req dto.CreateGroupRequest) (*dto.GroupResponse, error)
	ListMine(ctx context.Context, userID int64, status *string, limit, offset int) ([]dto.GroupResponse, int64, error)
	ListAll(ctx context.Context, status *string, limit, offset int) ([]dto.GroupResponse, int64, error)
	Get(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*dto.GroupResponse, error)
	ListMembers(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupMemberResponse, error)
	RemoveMember(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error
	LeaveGroup(ctx context.Context, userID, groupID int64) error
	TransferOwnership(ctx context.Context, userID int64, role enum.UserRole, groupID, newOwnerID int64) error
	CreateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CreateGroupInviteRequest) (*dto.GroupInviteResponse, error)
	DeactivateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) error
	ChangeInvitePassword(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, password string) error
	UpdateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, req dto.UpdateGroupInviteRequest) (*dto.GroupInviteResponse, error)
	JoinGroupByInvite(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error)
	ListInvites(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupInviteResponse, error)
	Update(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.UpdateGroupRequest) (*dto.GroupResponse, error)
	ChangeCover(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error
	Archive(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error
	Delete(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error
	Clone(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CloneGroupRequest) (*dto.GroupResponse, error)
}
