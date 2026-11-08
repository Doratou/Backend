package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type InviteService interface {
	SendDirect(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error)
	RespondToDirect(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error)
	ListMyDirectInvites(ctx context.Context, userID int64) ([]dto.DirectInviteResponse, error)
	ListGroupDirectInvites(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.DirectInviteResponse, error)
	CreateLink(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CreateLinkInviteRequest) (*dto.LinkInviteResponse, error)
	DeactivateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) error
	ChangeLinkPassword(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, password string) error
	UpdateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, req dto.UpdateLinkInviteRequest) (*dto.LinkInviteResponse, error)
	JoinByLink(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error)
	ListGroupLinks(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.LinkInviteResponse, error)
}
