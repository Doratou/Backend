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

type groupNotificationDecorator struct {
	next      adapter.GroupService
	notifSvc  adapter.NotificationService
	groupRepo domain_db.GroupRepository
	userRepo  domain_db.UserRepository
}

// NewGroupNotificationDecorator wraps an adapter.GroupService to dispatch notifications when
// members leave voluntarily or are kicked by a group owner/admin.
func NewGroupNotificationDecorator(
	next adapter.GroupService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
	userRepo domain_db.UserRepository,
) adapter.GroupService {
	return &groupNotificationDecorator{
		next:      next,
		notifSvc:  notifSvc,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (d *groupNotificationDecorator) LeaveGroup(ctx context.Context, userID, groupID int64) error {
	if err := d.next.LeaveGroup(ctx, userID, groupID); err != nil {
		return err
	}

	if d.notifSvc != nil {
		groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
		userName, _ := resolveUserProfile(ctx, d.userRepo, userID)

		dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventGroupMemberLeft, payload.GroupMemberLeftData{
			GroupID:   groupID,
			GroupName: groupName,
			UserID:    strconv.FormatInt(userID, 10),
			UserName:  userName,
			Reason:    "left",
			LeftAt:    time.Now().UTC(),
		})
	}

	return nil
}

func (d *groupNotificationDecorator) RemoveMember(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error {
	// Capture the member's UserID before they are marked inactive.
	var kickedUserID int64
	if d.notifSvc != nil && d.groupRepo != nil {
		if member, err := d.groupRepo.FindActiveMemberByID(ctx, groupID, memberID); err == nil && member != nil {
			kickedUserID = member.UserID
		}
	}

	if err := d.next.RemoveMember(ctx, userID, role, groupID, memberID); err != nil {
		return err
	}

	if d.notifSvc != nil && kickedUserID > 0 {
		groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
		kickedUserName, _ := resolveUserProfile(ctx, d.userRepo, kickedUserID)
		data := payload.GroupMemberLeftData{
			GroupID:   groupID,
			GroupName: groupName,
			UserID:    strconv.FormatInt(kickedUserID, 10),
			UserName:  kickedUserName,
			Reason:    "kicked",
			LeftAt:    time.Now().UTC(),
		}

		// Notify remaining active group members and directly notify the kicked user.
		dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventGroupMemberLeft, data)
		dispatchUserNotification(ctx, d.notifSvc, kickedUserID, &userID, &groupID, payload.WSEventGroupMemberLeft, data)
	}

	return nil
}

func (d *groupNotificationDecorator) ChangeStatus(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.ChangeGroupStatusRequest) (*dto.GroupResponse, error) {
	return d.next.ChangeStatus(ctx, userID, role, groupID, req)
}

func (d *groupNotificationDecorator) Create(ctx context.Context, userID int64, req dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	return d.next.Create(ctx, userID, req)
}

func (d *groupNotificationDecorator) ListMine(ctx context.Context, userID int64, filter dto.ListUserGroupsFilter, limit, offset int) ([]dto.GroupResponse, int64, error) {
	return d.next.ListMine(ctx, userID, filter, limit, offset)
}

func (d *groupNotificationDecorator) ListAll(ctx context.Context, status *string, limit, offset int) ([]dto.GroupResponse, int64, error) {
	return d.next.ListAll(ctx, status, limit, offset)
}

func (d *groupNotificationDecorator) Get(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*dto.GroupResponse, error) {
	return d.next.Get(ctx, userID, role, groupID)
}

func (d *groupNotificationDecorator) ListMembers(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupMemberResponse, error) {
	return d.next.ListMembers(ctx, userID, role, groupID)
}

func (d *groupNotificationDecorator) TransferOwnership(ctx context.Context, userID int64, role enum.UserRole, groupID, newOwnerID int64) error {
	return d.next.TransferOwnership(ctx, userID, role, groupID, newOwnerID)
}

func (d *groupNotificationDecorator) Update(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	return d.next.Update(ctx, userID, role, groupID, req)
}

func (d *groupNotificationDecorator) ChangeCover(ctx context.Context, userID int64, role enum.UserRole, groupID int64, coverAttachmentID int64) (*dto.GroupResponse, error) {
	return d.next.ChangeCover(ctx, userID, role, groupID, coverAttachmentID)
}

func (d *groupNotificationDecorator) Archive(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	return d.next.Archive(ctx, userID, role, groupID)
}

func (d *groupNotificationDecorator) Delete(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	return d.next.Delete(ctx, userID, role, groupID)
}

func (d *groupNotificationDecorator) Clone(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CloneGroupRequest) (*dto.GroupResponse, error) {
	return d.next.Clone(ctx, userID, role, groupID, req)
}

func (d *groupNotificationDecorator) IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return d.next.IsActiveMember(ctx, groupID, userID)
}
