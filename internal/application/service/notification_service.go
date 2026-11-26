package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

// NotificationServiceConfig holds optional delivery configuration for NotificationService.
type NotificationServiceConfig struct {
	RealtimeTTL time.Duration
	StandardTTL time.Duration
	NormalTTL   time.Duration
}

type notificationService struct {
	repo      domain_repo.NotificationRepository
	groupRepo domain_repo.GroupRepository
	dbTrx     database.TrxManager
	wsPub     domain_ws.Publisher
	cfg       NotificationServiceConfig
}

// NewNotificationService creates a new instance of NotificationService.
func NewNotificationService(
	repo domain_repo.NotificationRepository,
	groupRepo domain_repo.GroupRepository,
	dbTrx database.TrxManager,
	wsPub domain_ws.Publisher,
	cfg NotificationServiceConfig,
) adapter.NotificationService {
	if cfg.RealtimeTTL <= 0 {
		cfg.RealtimeTTL = 1 * time.Minute
	}
	if cfg.StandardTTL <= 0 {
		cfg.StandardTTL = 5 * time.Minute
	}
	if cfg.NormalTTL <= 0 {
		cfg.NormalTTL = 15 * time.Minute
	}
	return &notificationService{
		repo:      repo,
		groupRepo: groupRepo,
		dbTrx:     dbTrx,
		wsPub:     wsPub,
		cfg:       cfg,
	}
}

func marshalRelayData(data any) (json.RawMessage, error) {
	if data == nil {
		return nil, nil
	}
	switch v := data.(type) {
	case json.RawMessage:
		return v, nil
	case []byte:
		return json.RawMessage(v), nil
	default:
		raw, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}
}

// publishWS dispatches real-time WebSocket events directly through the Pub/Sub publisher (e.g. Redis).
func (s *notificationService) publishWS(ctx context.Context, msg domain_ws.RelayMessage, data any) {
	if s.wsPub == nil {
		return
	}

	rawData, err := marshalRelayData(data)
	if err != nil {
		log.Printf("[NOTIFICATION PUBLISHER WARNING] failed to marshal WS event %q data: %v", msg.Event, err)
		return
	}
	msg.Data = rawData
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}

	if err := s.wsPub.Publish(ctx, msg); err != nil {
		log.Printf("[NOTIFICATION PUBLISHER WARNING] failed to publish WS event %q to pubsub: %v", msg.Event, err)
	}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID int64, req dto.GetNotificationsQuery) (*dto.NotificationListResponse, error) {
	if userID <= 0 {
		return nil, exception.ErrBadRequest
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	filter := domain_repo.NotificationFilter{
		RecipientID: userID,
		CursorID:    req.CursorID,
		Limit:       limit,
		Status:      req.Status,
		GroupID:     req.GroupID,
	}

	entities, err := s.repo.ListByRecipient(ctx, filter)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.GetUnreadCountByGroup(ctx, userID, req.GroupID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotificationResponse, len(entities))
	var nextCursor int64
	for i, e := range entities {
		items[i] = dto.NotificationResponse{
			ID:          e.ID,
			RecipientID: e.RecipientID,
			ActorID:     e.ActorID,
			GroupID:     e.GroupID,
			Event:       e.Event,
			Data:        e.Data,
			IsRead:      e.IsRead,
			CreatedAt:   e.CreatedAt,
		}
	}

	if len(entities) > 0 {
		nextCursor = entities[len(entities)-1].ID
	}

	hasMore := len(entities) == limit

	return &dto.NotificationListResponse{
		Items:       items,
		UnreadCount: unreadCount,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID int64, groupID *int64) (*dto.UnreadCountResponse, error) {
	if userID <= 0 {
		return nil, exception.ErrBadRequest
	}

	count, err := s.repo.GetUnreadCountByGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	lastReadID, err := s.repo.GetLastReadID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.UnreadCountResponse{
		UnreadCount: count,
		LastReadID:  lastReadID,
	}, nil
}

func (s *notificationService) SyncNotifications(ctx context.Context, userID int64, req dto.SyncNotificationsQuery) (*dto.NotificationSyncResponse, error) {
	if userID <= 0 || req.SinceID < 0 {
		return nil, exception.ErrBadRequest
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	entities, err := s.repo.ListMissedSince(ctx, userID, req.SinceID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotificationResponse, len(entities))
	var latestID int64 = req.SinceID
	for i, e := range entities {
		items[i] = dto.NotificationResponse{
			ID:          e.ID,
			RecipientID: e.RecipientID,
			ActorID:     e.ActorID,
			GroupID:     e.GroupID,
			Event:       e.Event,
			Data:        e.Data,
			IsRead:      e.IsRead,
			CreatedAt:   e.CreatedAt,
		}
		if e.ID > latestID {
			latestID = e.ID
		}
	}

	hasMore := len(entities) == limit

	return &dto.NotificationSyncResponse{
		Items:    items,
		LatestID: latestID,
		HasMore:  hasMore,
	}, nil
}

func (s *notificationService) GetNotificationByID(ctx context.Context, userID, id int64) (*dto.NotificationResponse, error) {
	if userID <= 0 || id <= 0 {
		return nil, exception.ErrBadRequest
	}

	notif, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if notif.RecipientID != userID {
		return nil, exception.ErrNotificationNotFound
	}

	return &dto.NotificationResponse{
		ID:          notif.ID,
		RecipientID: notif.RecipientID,
		ActorID:     notif.ActorID,
		GroupID:     notif.GroupID,
		Event:       notif.Event,
		Data:        notif.Data,
		IsRead:      notif.IsRead,
		CreatedAt:   notif.CreatedAt,
	}, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID int64, req dto.MarkAsReadRequest) (*dto.MarkAsReadResponse, error) {
	if userID <= 0 {
		return nil, exception.ErrBadRequest
	}

	targetID := req.UpToID
	if targetID <= 0 {
		// If up_to_id is omitted/zero, mark all current notifications as read
		latest, err := s.repo.ListByRecipient(ctx, domain_repo.NotificationFilter{
			RecipientID: userID,
			Limit:       1,
		})
		if err != nil {
			return nil, err
		}
		if len(latest) > 0 {
			targetID = latest[0].ID
		} else {
			lastReadID, _ := s.repo.GetLastReadID(ctx, userID)
			return &dto.MarkAsReadResponse{
				LastReadID:  lastReadID,
				UnreadCount: 0,
			}, nil
		}
	}

	if err := s.repo.MarkAsRead(ctx, userID, targetID); err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.MarkAsReadResponse{
		LastReadID:  targetID,
		UnreadCount: unreadCount,
	}, nil
}

func (s *notificationService) MarkSingleAsRead(ctx context.Context, userID, id int64) (*dto.MarkAsReadResponse, error) {
	if userID <= 0 || id <= 0 {
		return nil, exception.ErrBadRequest
	}

	notif, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if notif.RecipientID != userID {
		return nil, exception.ErrNotificationNotFound
	}

	return s.MarkAsRead(ctx, userID, dto.MarkAsReadRequest{UpToID: id})
}

func (s *notificationService) DeleteNotification(ctx context.Context, userID, id int64) error {
	if userID <= 0 || id <= 0 {
		return exception.ErrBadRequest
	}

	notif, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if notif.RecipientID != userID {
		return exception.ErrNotificationNotFound
	}

	return s.repo.SoftDelete(ctx, id)
}

func (s *notificationService) ClearNotifications(ctx context.Context, userID int64, onlyRead bool) error {
	if userID <= 0 {
		return exception.ErrBadRequest
	}

	if onlyRead {
		return s.repo.SoftDeleteReadByRecipient(ctx, userID)
	}
	return s.repo.SoftDeleteAllByRecipient(ctx, userID)
}

func (s *notificationService) CreateGroupAnnouncement(ctx context.Context, ownerUserID, groupID int64, req dto.GroupAnnouncementRequest) (*dto.GroupAnnouncementResponse, error) {
	if ownerUserID <= 0 || groupID <= 0 {
		return nil, exception.ErrBadRequest
	}

	group, _, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, exception.Wrap(exception.ErrGroupNotFound, err)
	}

	if group.CreatedByUserID != ownerUserID {
		return nil, exception.ErrForbiddenRole
	}

	members, err := s.groupRepo.ListActiveMembers(ctx, groupID)
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	// Filter out the owner so the owner does not receive duplicate/self notifications
	announcementData := payload.GroupAnnouncementData{
		GroupID:    groupID,
		GroupName:  group.Title,
		SenderID:   fmt.Sprintf("%d", ownerUserID),
		SenderName: "Group Owner",
		Title:      req.Title,
		Message:    req.Message,
		SentAt:     time.Now(),
	}
	dataBytes, err := json.Marshal(announcementData)
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}

	recipientIDs := make([]int64, 0, len(members))
	notifs := make([]*entity.Notification, 0, len(members))
	for _, m := range members {
		if m.UserID == ownerUserID {
			continue // Exclude owner
		}
		recipientIDs = append(recipientIDs, m.UserID)
		notif := &entity.Notification{
			RecipientID: m.UserID,
			ActorID:     &ownerUserID,
			GroupID:     &groupID,
			Event:       payload.WSEventGroupAnnouncement,
			Data:        dataBytes,
		}
		notifs = append(notifs, notif)
	}

	if len(notifs) > 0 {
		if err := s.repo.CreateBatch(ctx, notifs); err != nil {
			return nil, err
		}

		// Dispatch real-time WebSocket event directly via Pub/Sub to active group members.
		s.publishWS(ctx, domain_ws.RelayMessage{
			Event:      string(payload.WSEventGroupAnnouncement),
			TargetType: domain_ws.RelayTargetUsers,
			TargetIDs:  recipientIDs,
			Channel:    fmt.Sprintf("group:%d", groupID),
			SenderID:   ownerUserID,
		}, announcementData)
	}

	return &dto.GroupAnnouncementResponse{
		DispatchedCount: len(notifs),
		GroupID:         groupID,
	}, nil
}

func (s *notificationService) DispatchToUser(ctx context.Context, recipientID int64, actorID *int64, groupID *int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	if recipientID <= 0 {
		return exception.ErrBadRequest
	}

	notif := &entity.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		GroupID:     groupID,
		Event:       event,
	}
	if err := notif.SetData(data); err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}

	if err := s.repo.Create(ctx, notif); err != nil {
		return err
	}

	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	var ch string
	if groupID != nil && *groupID > 0 {
		ch = fmt.Sprintf("group:%d", *groupID)
	}
	s.publishWS(ctx, domain_ws.RelayMessage{
		Event:      string(event),
		TargetType: domain_ws.RelayTargetUsers,
		TargetIDs:  []int64{recipientID},
		Channel:    ch,
		SenderID:   senderID,
	}, data)

	return nil
}

func (s *notificationService) createBatchNotifications(ctx context.Context, recipientIDs []int64, actorID *int64, groupID *int64, event payload.WSEvent, data any) error {
	if len(recipientIDs) == 0 {
		return nil
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}

	notifs := make([]*entity.Notification, len(recipientIDs))
	for i, rID := range recipientIDs {
		notifs[i] = &entity.Notification{
			RecipientID: rID,
			ActorID:     actorID,
			GroupID:     groupID,
			Event:       event,
			Data:        dataBytes,
		}
	}

	return s.repo.CreateBatch(ctx, notifs)
}

func (s *notificationService) DispatchToUsers(ctx context.Context, recipientIDs []int64, actorID *int64, groupID *int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	if len(recipientIDs) == 0 {
		return nil
	}

	if err := s.createBatchNotifications(ctx, recipientIDs, actorID, groupID, event, data); err != nil {
		return err
	}

	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	var ch string
	if groupID != nil && *groupID > 0 {
		ch = fmt.Sprintf("group:%d", *groupID)
	}
	s.publishWS(ctx, domain_ws.RelayMessage{
		Event:      string(event),
		TargetType: domain_ws.RelayTargetUsers,
		TargetIDs:  recipientIDs,
		Channel:    ch,
		SenderID:   senderID,
	}, data)

	return nil
}

func (s *notificationService) DispatchToGroup(ctx context.Context, groupID int64, actorID *int64, recipientUserIDs []int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	if groupID <= 0 {
		return exception.ErrBadRequest
	}

	if len(recipientUserIDs) == 0 {
		members, err := s.groupRepo.ListActiveMembers(ctx, groupID)
		if err != nil {
			return exception.Wrap(exception.ErrInternal, err)
		}
		recipientUserIDs = make([]int64, 0, len(members))
		for _, m := range members {
			recipientUserIDs = append(recipientUserIDs, m.UserID)
		}
	}

	if len(recipientUserIDs) == 0 {
		return nil
	}

	if err := s.createBatchNotifications(ctx, recipientUserIDs, actorID, &groupID, event, data); err != nil {
		return err
	}

	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	s.publishWS(ctx, domain_ws.RelayMessage{
		Event:      string(event),
		TargetType: domain_ws.RelayTargetUsers,
		TargetIDs:  recipientUserIDs,
		Channel:    fmt.Sprintf("group:%d", groupID),
		SenderID:   senderID,
	}, data)

	return nil
}

func (s *notificationService) Broadcast(ctx context.Context, actorID *int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	s.publishWS(ctx, domain_ws.RelayMessage{
		Event:      string(event),
		TargetType: domain_ws.RelayTargetBroadcast,
		SenderID:   senderID,
	}, data)
	return nil
}
