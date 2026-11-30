package decorator

import (
	"context"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type inviteNotificationDecorator struct {
	next      adapter.InviteService
	notifSvc  adapter.NotificationService
	groupRepo domain_db.GroupRepository
	userRepo  domain_db.UserRepository
}

// NewInviteNotificationDecorator wraps an adapter.InviteService to dispatch notifications when
// direct group invites are sent and when users join a group via direct invite or invite link.
func NewInviteNotificationDecorator(
	next adapter.InviteService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
	userRepo domain_db.UserRepository,
) adapter.InviteService {
	return &inviteNotificationDecorator{
		next:      next,
		notifSvc:  notifSvc,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (d *inviteNotificationDecorator) SendDirect(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error) {
	results, err := d.next.SendDirect(ctx, userID, role, req)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil {
		groupName := resolveGroupTitle(ctx, d.groupRepo, req.GroupID)
		senderName, senderAvatarID := resolveUserProfile(ctx, d.userRepo, userID)
		senderIDStr := strconv.FormatInt(userID, 10)

		for _, r := range results {
			if !r.Success || r.Invite == nil {
				continue
			}
			dispatchUserNotification(ctx, d.notifSvc, r.UserID, &userID, &req.GroupID, payload.WSEventGroupInviteReceived, payload.GroupInviteReceivedData{
				InviteID:       r.Invite.ID,
				GroupID:        req.GroupID,
				GroupName:      groupName,
				SenderID:       senderIDStr,
				SenderName:     senderName,
				SenderAvatarID: senderAvatarID,
				InvitedAt:      r.Invite.CreatedAt,
			})
		}
	}

	return results, nil
}

func (d *inviteNotificationDecorator) RespondToDirect(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error) {
	res, err := d.next.RespondToDirect(ctx, userID, inviteID, response)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil && response == string(enum.DirectInviteStatusAccepted) && res.Group != nil {
		var userName string
		if res.Recipient != nil {
			userName = res.Recipient.Name
		}
		if userName == "" {
			userName, _ = resolveUserProfile(ctx, d.userRepo, userID)
		}
		joinedAt := time.Now().UTC()
		if res.RespondedAt != nil {
			joinedAt = *res.RespondedAt
		}

		dispatchGroupNotification(ctx, d.notifSvc, res.Group.ID, &userID, payload.WSEventGroupMemberJoined, payload.GroupMemberJoinedData{
			GroupID:   res.Group.ID,
			GroupName: res.Group.Title,
			UserID:    strconv.FormatInt(userID, 10),
			UserName:  userName,
			JoinedAt:  joinedAt,
		})
	}

	return res, nil
}

func (d *inviteNotificationDecorator) JoinByLink(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error) {
	res, err := d.next.JoinByLink(ctx, userID, token, password)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil {
		var userName string
		var joinedAt time.Time
		for _, m := range res.Members {
			if m.UserID == userID {
				userName = m.Name
				joinedAt = m.JoinedAt
				break
			}
		}
		if userName == "" {
			userName, _ = resolveUserProfile(ctx, d.userRepo, userID)
		}
		if joinedAt.IsZero() {
			joinedAt = time.Now().UTC()
		}

		dispatchGroupNotification(ctx, d.notifSvc, res.ID, &userID, payload.WSEventGroupMemberJoined, payload.GroupMemberJoinedData{
			GroupID:   res.ID,
			GroupName: res.Title,
			UserID:    strconv.FormatInt(userID, 10),
			UserName:  userName,
			JoinedAt:  joinedAt,
		})
	}

	return res, nil
}

func (d *inviteNotificationDecorator) ListMyDirectInvites(ctx context.Context, userID int64) ([]dto.DirectInviteResponse, error) {
	return d.next.ListMyDirectInvites(ctx, userID)
}

func (d *inviteNotificationDecorator) ListGroupDirectInvites(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.DirectInviteResponse, error) {
	return d.next.ListGroupDirectInvites(ctx, userID, role, groupID)
}

func (d *inviteNotificationDecorator) CreateLink(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CreateLinkInviteRequest) (*dto.LinkInviteResponse, error) {
	return d.next.CreateLink(ctx, userID, role, groupID, req)
}

func (d *inviteNotificationDecorator) DeactivateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) error {
	return d.next.DeactivateLink(ctx, userID, role, groupID, inviteID)
}

func (d *inviteNotificationDecorator) ChangeLinkPassword(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, password string) error {
	return d.next.ChangeLinkPassword(ctx, userID, role, groupID, inviteID, password)
}

func (d *inviteNotificationDecorator) UpdateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, req dto.UpdateLinkInviteRequest) (*dto.LinkInviteResponse, error) {
	return d.next.UpdateLink(ctx, userID, role, groupID, inviteID, req)
}

func (d *inviteNotificationDecorator) ListGroupLinks(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.LinkInviteResponse, error) {
	return d.next.ListGroupLinks(ctx, userID, role, groupID)
}
