package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type notificationService struct {
	repo      domain_repo.NotificationRepository
	groupRepo domain_repo.GroupRepository
	dbTrx     database.TrxManager
	brokerPub domain_broker.Publisher
}

// NewNotificationService creates a new instance of NotificationService.
func NewNotificationService(
	repo domain_repo.NotificationRepository,
	groupRepo domain_repo.GroupRepository,
	dbTrx database.TrxManager,
	brokerPub domain_broker.Publisher,
) adapter.NotificationService {
	return &notificationService{
		repo:      repo,
		groupRepo: groupRepo,
		dbTrx:     dbTrx,
		brokerPub: brokerPub,
	}
}

// publishWS encapsulates real-time event dispatching to RabbitMQ with standard TTL and urgency.
func (s *notificationService) publishWS(ctx context.Context, p payload.WSPayload) {
	if s.brokerPub == nil {
		return
	}
	p.Timestamp = time.Now()
	if err := s.brokerPub.PublishJSON(ctx, domain_broker.NotificationExchange, domain_broker.WSRoutingKey, p,
		domain_broker.WithTTL(60*time.Second),
		domain_broker.WithUrgency(domain_broker.UrgencyRealtime),
	); err != nil {
		log.Printf("[NOTIFICATION PUBLISHER WARNING] failed to publish WS event %q to rabbitmq: %v", p.Event, err)
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

	lastReadID, _ := s.repo.GetLastReadID(ctx, userID)
	notif.IsRead = notif.ID <= lastReadID

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

	notifs := make([]*entity.Notification, 0, len(members))
	for _, m := range members {
		if m.UserID == ownerUserID {
			continue // Exclude owner
		}
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
	}

	// Dispatch real-time WebSocket event via RabbitMQ
	s.publishWS(ctx, payload.WSPayload{
		Event:      payload.WSEventGroupAnnouncement,
		TargetType: payload.TargetGroups,
		TargetIDs:  []int64{groupID},
		Data:       announcementData,
		SenderID:   ownerUserID,
	})

	return &dto.GroupAnnouncementResponse{
		DispatchedCount: len(notifs),
		GroupID:         groupID,
	}, nil
}

func (s *notificationService) DispatchToUser(ctx context.Context, recipientID int64, actorID *int64, groupID *int64, event payload.WSEvent, data any) error {
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
	s.publishWS(ctx, payload.WSPayload{
		Event:      event,
		TargetType: payload.TargetUsers,
		TargetIDs:  []int64{recipientID},
		Data:       data,
		SenderID:   senderID,
	})

	return nil
}

func (s *notificationService) DispatchToUsers(ctx context.Context, recipientIDs []int64, actorID *int64, groupID *int64, event payload.WSEvent, data any) error {
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

	if err := s.repo.CreateBatch(ctx, notifs); err != nil {
		return err
	}

	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	s.publishWS(ctx, payload.WSPayload{
		Event:      event,
		TargetType: payload.TargetUsers,
		TargetIDs:  recipientIDs,
		Data:       data,
		SenderID:   senderID,
	})

	return nil
}

func (s *notificationService) DispatchToGroup(ctx context.Context, groupID int64, actorID *int64, recipientMemberIDs []int64, event payload.WSEvent, data any) error {
	if len(recipientMemberIDs) == 0 {
		return nil
	}
	if err := s.DispatchToUsers(ctx, recipientMemberIDs, actorID, &groupID, event, data); err != nil {
		return err
	}

	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	s.publishWS(ctx, payload.WSPayload{
		Event:      event,
		TargetType: payload.TargetGroups,
		TargetIDs:  []int64{groupID},
		Data:       data,
		SenderID:   senderID,
	})

	return nil
}

func (s *notificationService) Broadcast(ctx context.Context, actorID *int64, event payload.WSEvent, data any) error {
	var senderID int64
	if actorID != nil {
		senderID = *actorID
	}
	s.publishWS(ctx, payload.WSPayload{
		Event:      event,
		TargetType: payload.TargetBroadcast,
		Data:       data,
		SenderID:   senderID,
	})
	return nil
}
