package DB

import (
	"context"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
)

type InviteRepository interface {
	CreateLink(ctx context.Context, groupID, creatorUserID int64, invite *entity.LinkInvite) error
	FindLinkByID(ctx context.Context, groupID, inviteID int64) (*entity.LinkInvite, error)
	FindLinksByGroupID(ctx context.Context, groupID int64) ([]entity.LinkInvite, error)
	FindLinkByTokenHash(ctx context.Context, tokenHash string) (*entity.LinkInvite, error)
	RevokeLink(ctx context.Context, groupID, inviteID int64) error
	UpdateLinkPassword(ctx context.Context, groupID, inviteID int64, passwordHash string) error
	UpdateLink(ctx context.Context, invite *entity.LinkInvite) error
	JoinGroupWithLink(ctx context.Context, inviteID, userID int64) (int64, error)
	CreateDirect(ctx context.Context, invite *entity.DirectInvite, rejectCooldown time.Duration) error
	RespondDirect(ctx context.Context, inviteID, recipientID int64, accept bool) (*entity.DirectInvite, error)
	FindDirectByRecipient(ctx context.Context, recipientID int64) ([]entity.DirectInvite, error)
	FindDirectByGroup(ctx context.Context, groupID int64) ([]entity.DirectInvite, error)
}
