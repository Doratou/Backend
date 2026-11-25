package decorator

import (
	"context"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type friendshipNotificationDecorator struct {
	next           adapter.FriendshipService
	notifSvc       adapter.NotificationService
	friendshipRepo domain_db.FriendshipRepository
	userRepo       domain_db.UserRepository
}

// NewFriendshipNotificationDecorator wraps an adapter.FriendshipService to dispatch notifications
// on friend request creation, acceptance, and rejection.
func NewFriendshipNotificationDecorator(
	next adapter.FriendshipService,
	notifSvc adapter.NotificationService,
	friendshipRepo domain_db.FriendshipRepository,
	userRepo domain_db.UserRepository,
) adapter.FriendshipService {
	return &friendshipNotificationDecorator{
		next:           next,
		notifSvc:       notifSvc,
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

func (d *friendshipNotificationDecorator) SendRequest(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error) {
	res, err := d.next.SendRequest(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	senderName, senderAvatarID := resolveUserProfile(ctx, d.userRepo, userID)
	dispatchUserNotification(ctx, d.notifSvc, res.ReceiverUserID, &userID, nil, payload.WSEventFriendRequestReceived, payload.FriendRequestReceivedData{
		RequestID:      res.ID,
		SenderID:       strconv.FormatInt(userID, 10),
		SenderName:     senderName,
		SenderAvatarID: senderAvatarID,
		CreatedAt:      res.CreatedAt,
	})

	return res, nil
}

func (d *friendshipNotificationDecorator) ListRequests(ctx context.Context, userID int64, req dto.ListFriendRequestsRequest) (*dto.FriendRequestListResponse, error) {
	return d.next.ListRequests(ctx, userID, req)
}

func (d *friendshipNotificationDecorator) AcceptRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	res, err := d.next.AcceptRequest(ctx, userID, requestID)
	if err != nil {
		return nil, err
	}
	d.notifyRequestResponse(ctx, userID, requestID, "accepted", res.RespondedAt)
	return res, nil
}

func (d *friendshipNotificationDecorator) RejectRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	res, err := d.next.RejectRequest(ctx, userID, requestID)
	if err != nil {
		return nil, err
	}
	d.notifyRequestResponse(ctx, userID, requestID, "rejected", res.RespondedAt)
	return res, nil
}

func (d *friendshipNotificationDecorator) notifyRequestResponse(ctx context.Context, responderID, requestID int64, status string, respondedAt *time.Time) {
	if d.notifSvc == nil || d.friendshipRepo == nil {
		return
	}
	friendship, err := d.friendshipRepo.FindByID(ctx, requestID, false)
	if err != nil || friendship == nil {
		return
	}

	ts := time.Now().UTC()
	if respondedAt != nil {
		ts = *respondedAt
	}
	responderName, responderAvatarID := resolveUserProfile(ctx, d.userRepo, responderID)

	dispatchUserNotification(ctx, d.notifSvc, friendship.RequesterUserID, &responderID, nil, payload.WSEventFriendRequestResponded, payload.FriendRequestRespondedData{
		RequestID:         requestID,
		ResponderID:       strconv.FormatInt(responderID, 10),
		ResponderAvatarID: responderAvatarID,
		ResponderName:     responderName,
		Status:            status,
		RespondedAt:       ts,
	})
}

func (d *friendshipNotificationDecorator) CancelRequest(ctx context.Context, userID, requestID int64) error {
	return d.next.CancelRequest(ctx, userID, requestID)
}

func (d *friendshipNotificationDecorator) ListFriends(ctx context.Context, userID int64, req dto.ListFriendsRequest) (*dto.FriendListResponse, error) {
	return d.next.ListFriends(ctx, userID, req)
}

func (d *friendshipNotificationDecorator) ListMutualGroups(ctx context.Context, userID, friendshipID int64, req dto.ListFriendsRequest) (*dto.PaginatedResponse[dto.MutualGroupResponse], error) {
	return d.next.ListMutualGroups(ctx, userID, friendshipID, req)
}

func (d *friendshipNotificationDecorator) RemoveFriend(ctx context.Context, userID, friendshipID int64) error {
	return d.next.RemoveFriend(ctx, userID, friendshipID)
}
