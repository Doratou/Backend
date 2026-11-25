package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/application/service"
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"gorm.io/gorm"
)

type mockNotificationRepo struct {
	notifications []*entity.Notification
	readStates    map[int64]int64 // userID -> lastReadID
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{
		notifications: make([]*entity.Notification, 0),
		readStates:    make(map[int64]int64),
	}
}

func (m *mockNotificationRepo) Create(ctx context.Context, notif *entity.Notification) error {
	notif.ID = int64(len(m.notifications) + 1)
	notif.CreatedAt = time.Now()
	m.notifications = append(m.notifications, notif)
	return nil
}

func (m *mockNotificationRepo) CreateBatch(ctx context.Context, notifs []*entity.Notification) error {
	for _, n := range notifs {
		_ = m.Create(ctx, n)
	}
	return nil
}

func (m *mockNotificationRepo) FindByID(ctx context.Context, id int64) (*entity.Notification, error) {
	for _, n := range m.notifications {
		if n.ID == id && !n.DeletedAt.Valid {
			return n, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockNotificationRepo) ListByRecipient(ctx context.Context, filter domain_repo.NotificationFilter) ([]*entity.Notification, error) {
	lastRead := m.readStates[filter.RecipientID]
	var results []*entity.Notification
	for i := len(m.notifications) - 1; i >= 0; i-- {
		n := m.notifications[i]
		if n.RecipientID != filter.RecipientID || n.DeletedAt.Valid {
			continue
		}
		if filter.CursorID > 0 && n.ID >= filter.CursorID {
			continue
		}
		if filter.GroupID != nil && (n.GroupID == nil || *n.GroupID != *filter.GroupID) {
			continue
		}
		isRead := n.ID <= lastRead
		if filter.Status == "unread" && isRead {
			continue
		}
		if filter.Status == "read" && !isRead {
			continue
		}
		copyNotif := *n
		copyNotif.IsRead = isRead
		results = append(results, &copyNotif)
		if len(results) >= filter.Limit {
			break
		}
	}
	return results, nil
}

func (m *mockNotificationRepo) ListMissedSince(ctx context.Context, recipientID int64, sinceID int64, limit int) ([]*entity.Notification, error) {
	lastRead := m.readStates[recipientID]
	var results []*entity.Notification
	for _, n := range m.notifications {
		if n.RecipientID == recipientID && !n.DeletedAt.Valid && n.ID > sinceID {
			copyNotif := *n
			copyNotif.IsRead = n.ID <= lastRead
			results = append(results, &copyNotif)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockNotificationRepo) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	return m.GetUnreadCountByGroup(ctx, userID, nil)
}

func (m *mockNotificationRepo) GetUnreadCountByGroup(ctx context.Context, userID int64, groupID *int64) (int64, error) {
	lastRead := m.readStates[userID]
	var count int64
	for _, n := range m.notifications {
		if n.RecipientID == userID && !n.DeletedAt.Valid && n.ID > lastRead {
			if groupID != nil && (n.GroupID == nil || *n.GroupID != *groupID) {
				continue
			}
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationRepo) GetLastReadID(ctx context.Context, userID int64) (int64, error) {
	return m.readStates[userID], nil
}

func (m *mockNotificationRepo) Update(ctx context.Context, notif *entity.Notification) error {
	return nil
}

func (m *mockNotificationRepo) UpdateLastError(ctx context.Context, id int64, lastErr *string) error {
	return nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, userID int64, upToID int64) error {
	if upToID > m.readStates[userID] {
		m.readStates[userID] = upToID
	}
	return nil
}

func (m *mockNotificationRepo) SoftDelete(ctx context.Context, id int64) error {
	for _, n := range m.notifications {
		if n.ID == id {
			n.DeletedAt.Valid = true
			n.DeletedAt.Time = time.Now()
			return nil
		}
	}
	return nil
}

func (m *mockNotificationRepo) SoftDeleteAllByRecipient(ctx context.Context, recipientID int64) error {
	for _, n := range m.notifications {
		if n.RecipientID == recipientID {
			n.DeletedAt.Valid = true
			n.DeletedAt.Time = time.Now()
		}
	}
	return nil
}

func (m *mockNotificationRepo) SoftDeleteReadByRecipient(ctx context.Context, recipientID int64) error {
	lastRead := m.readStates[recipientID]
	for _, n := range m.notifications {
		if n.RecipientID == recipientID && n.ID <= lastRead {
			n.DeletedAt.Valid = true
			n.DeletedAt.Time = time.Now()
		}
	}
	return nil
}

func (m *mockNotificationRepo) HardDelete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockNotificationRepo) PurgeSoftDeletedBefore(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

type mockGroupRepo struct {
	groups  map[int64]*entity.Group
	members map[int64][]entity.GroupMember
}

func (m *mockGroupRepo) CreateWithOwner(ctx context.Context, group *entity.Group) error { return nil }
func (m *mockGroupRepo) FindByID(ctx context.Context, id int64) (*entity.Group, int64, error) {
	g, ok := m.groups[id]
	if !ok {
		return nil, 0, gorm.ErrRecordNotFound
	}
	return g, int64(len(m.members[id])), nil
}
func (m *mockGroupRepo) ListActiveMembers(ctx context.Context, groupID int64) ([]entity.GroupMember, error) {
	return m.members[groupID], nil
}
func (m *mockGroupRepo) FindActiveMemberByID(ctx context.Context, groupID, memberID int64) (*entity.GroupMember, error) {
	return nil, nil
}
func (m *mockGroupRepo) SetMemberInactive(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (m *mockGroupRepo) SetMemberInactiveByID(ctx context.Context, groupID, memberID int64) error {
	return nil
}
func (m *mockGroupRepo) TransferOwnership(ctx context.Context, groupID, newOwnerID int64) error {
	return nil
}
func (m *mockGroupRepo) FindByMember(ctx context.Context, userID int64, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) FindAll(ctx context.Context, status *enum.GroupStatus, limit, offset int) ([]entity.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return true, nil
}
func (m *mockGroupRepo) Update(ctx context.Context, group *entity.Group) error      { return nil }
func (m *mockGroupRepo) UpdateCover(ctx context.Context, group *entity.Group) error { return nil }
func (m *mockGroupRepo) SetStatus(ctx context.Context, groupID int64, status enum.GroupStatus) error {
	return nil
}
func (m *mockGroupRepo) Delete(ctx context.Context, groupID int64) error { return nil }
func (m *mockGroupRepo) CloneWithMembers(ctx context.Context, sourceGroupID int64, newGroup *entity.Group) error {
	return nil
}

type publishedRecord struct {
	payload payload.WSPayload
	opts    domain_broker.PublishOptions
}

type mockPublisher struct {
	published []payload.WSPayload
	records   []publishedRecord
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, msg domain_broker.Message, opts ...domain_broker.PublishOption) error {
	return nil
}

func (m *mockPublisher) PublishJSON(ctx context.Context, exchange, routingKey string, p any, opts ...domain_broker.PublishOption) error {
	var pubOpts domain_broker.PublishOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&pubOpts)
		}
	}
	if wsPayload, ok := p.(payload.WSPayload); ok {
		m.published = append(m.published, wsPayload)
		m.records = append(m.records, publishedRecord{payload: wsPayload, opts: pubOpts})
	}
	return nil
}

var testServiceConfig = service.NotificationServiceConfig{
	RealtimeTTL: time.Minute,
	StandardTTL: 5 * time.Minute,
	NormalTTL:   15 * time.Minute,
}

func TestNotificationService_GetAndMarkAsRead(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	// Dispatch 3 notifications
	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventFriendRequestReceived, map[string]string{"msg": "hi 1"})
	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventFriendRequestReceived, map[string]string{"msg": "hi 2"})
	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventFriendRequestReceived, map[string]string{"msg": "hi 3"})

	// Check unread count
	unread, err := srv.GetUnreadCount(ctx, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error getting unread count: %v", err)
	}
	if unread.UnreadCount != 3 {
		t.Fatalf("expected unread count 3, got %d", unread.UnreadCount)
	}

	// Mark up to ID 2 as read
	markRes, err := srv.MarkAsRead(ctx, 10, dto.MarkAsReadRequest{UpToID: 2})
	if err != nil {
		t.Fatalf("unexpected error marking as read: %v", err)
	}
	if markRes.LastReadID != 2 || markRes.UnreadCount != 1 {
		t.Fatalf("expected last_read_id=2 and unread_count=1, got last_read_id=%d and unread_count=%d", markRes.LastReadID, markRes.UnreadCount)
	}

	// Get notifications
	list, err := srv.GetNotifications(ctx, 10, dto.GetNotificationsQuery{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error listing notifications: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list.Items))
	}
	// Item 3 should be unread, Item 2 and 1 should be read
	if list.Items[0].ID != 3 || list.Items[0].IsRead {
		t.Fatalf("expected item 3 to be unread, got id=%d is_read=%v", list.Items[0].ID, list.Items[0].IsRead)
	}
	if list.Items[1].ID != 2 || !list.Items[1].IsRead {
		t.Fatalf("expected item 2 to be read, got id=%d is_read=%v", list.Items[1].ID, list.Items[1].IsRead)
	}
}

func TestNotificationService_SyncNotifications(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventTokenPriceUpdated, map[string]int{"price": 100})
	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventTokenPriceUpdated, map[string]int{"price": 110})
	_ = srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventTokenPriceUpdated, map[string]int{"price": 120})

	// Client has received up to ID 1, needs sync since ID 1
	syncRes, err := srv.SyncNotifications(ctx, 10, dto.SyncNotificationsQuery{SinceID: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error syncing: %v", err)
	}
	if len(syncRes.Items) != 2 {
		t.Fatalf("expected 2 missed items, got %d", len(syncRes.Items))
	}
	if syncRes.Items[0].ID != 2 || syncRes.Items[1].ID != 3 {
		t.Fatalf("expected items 2 and 3 in ASC order, got %d and %d", syncRes.Items[0].ID, syncRes.Items[1].ID)
	}
	if syncRes.LatestID != 3 {
		t.Fatalf("expected latestID=3, got %d", syncRes.LatestID)
	}
}

func TestNotificationService_CreateGroupAnnouncement_ExcludesOwner_AndPublishesToRMQ(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups: map[int64]*entity.Group{
			1: {
				Title:           "Test Group",
				CreatedByUserID: 100, // User 100 is creator/owner
			},
		},
		members: map[int64][]entity.GroupMember{
			1: {
				{UserID: 100}, // Owner
				{UserID: 200}, // Member 1
				{UserID: 300}, // Member 2
			},
		},
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	res, err := srv.CreateGroupAnnouncement(ctx, 100, 1, dto.GroupAnnouncementRequest{
		Title:   "Important Announcement",
		Message: "Meeting tomorrow at 5 PM",
	})
	if err != nil {
		t.Fatalf("unexpected error creating announcement: %v", err)
	}

	// Should dispatch to 2 members (200, 300), excluding owner 100
	if res.DispatchedCount != 2 {
		t.Fatalf("expected 2 dispatched notifications, got %d", res.DispatchedCount)
	}

	// Check that owner 100 has 0 notifications
	ownerList, _ := srv.GetNotifications(ctx, 100, dto.GetNotificationsQuery{})
	if len(ownerList.Items) != 0 {
		t.Fatalf("expected owner to have 0 notifications, got %d", len(ownerList.Items))
	}

	// Check that member 200 has 1 notification
	memberList, _ := srv.GetNotifications(ctx, 200, dto.GetNotificationsQuery{})
	if len(memberList.Items) != 1 {
		t.Fatalf("expected member 200 to have 1 notification, got %d", len(memberList.Items))
	}

	// Verify RabbitMQ publish was triggered with TargetGroups and Realtime delivery
	if len(pub.published) != 1 {
		t.Fatalf("expected 1 event published to RMQ, got %d", len(pub.published))
	}
	event := pub.published[0]
	if event.Event != payload.WSEventGroupAnnouncement {
		t.Fatalf("expected event %s, got %s", payload.WSEventGroupAnnouncement, event.Event)
	}
	if event.TargetType != payload.TargetGroups {
		t.Fatalf("expected target type %s, got %s", payload.TargetGroups, event.TargetType)
	}
	if len(event.TargetIDs) != 1 || event.TargetIDs[0] != 1 {
		t.Fatalf("expected target IDs [1], got %v", event.TargetIDs)
	}
	if event.SenderID != 100 {
		t.Fatalf("expected sender ID 100, got %d", event.SenderID)
	}

	rec := pub.records[0]
	if rec.opts.Urgency != domain_broker.UrgencyRealtime {
		t.Fatalf("expected announcement urgency Realtime (%d), got %d", domain_broker.UrgencyRealtime, rec.opts.Urgency)
	}
	if rec.opts.TTL != time.Minute {
		t.Fatalf("expected announcement TTL 1m, got %v", rec.opts.TTL)
	}
}

func TestNotificationService_DispatchToUser_PublishesToRMQ(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	var actorID int64 = 42
	err := srv.DispatchToUser(ctx, 10, &actorID, nil, payload.WSEventFriendRequestReceived, map[string]string{"from": "Ali"})
	if err != nil {
		t.Fatalf("unexpected error dispatching to user: %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.published))
	}
	ev := pub.published[0]
	if ev.TargetType != payload.TargetUsers {
		t.Fatalf("expected target type users, got %s", ev.TargetType)
	}
	if len(ev.TargetIDs) != 1 || ev.TargetIDs[0] != 10 {
		t.Fatalf("expected target IDs [10], got %v", ev.TargetIDs)
	}
	if ev.SenderID != 42 {
		t.Fatalf("expected sender ID 42, got %d", ev.SenderID)
	}

	rec := pub.records[0]
	if rec.opts.Urgency != domain_broker.UrgencyRealtime {
		t.Fatalf("expected friend_request urgency Realtime, got %d", rec.opts.Urgency)
	}
	if rec.opts.TTL != time.Minute {
		t.Fatalf("expected friend_request TTL 1m, got %v", rec.opts.TTL)
	}
}

func TestNotificationService_EventUrgencyAndTTL_Defaults(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	// 1. UrgencyRealtime event (e.g. SettlementPaid)
	if err := srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventSettlementPaid, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2. UrgencyStandard event (e.g. PollCreated)
	if err := srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventPollCreated, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3. UrgencyNormal event (e.g. TokenPriceUpdated)
	if err := srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventTokenPriceUpdated, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.records) != 3 {
		t.Fatalf("expected 3 published records, got %d", len(pub.records))
	}

	// Verify SettlementPaid -> Realtime & 1m
	if pub.records[0].opts.Urgency != domain_broker.UrgencyRealtime {
		t.Errorf("SettlementPaid: expected UrgencyRealtime (%d), got %d", domain_broker.UrgencyRealtime, pub.records[0].opts.Urgency)
	}
	if pub.records[0].opts.TTL != time.Minute {
		t.Errorf("SettlementPaid: expected TTL 1m, got %v", pub.records[0].opts.TTL)
	}

	// Verify PollCreated -> Standard & 5m
	if pub.records[1].opts.Urgency != domain_broker.UrgencyStandard {
		t.Errorf("PollCreated: expected UrgencyStandard (%d), got %d", domain_broker.UrgencyStandard, pub.records[1].opts.Urgency)
	}
	if pub.records[1].opts.TTL != 5*time.Minute {
		t.Errorf("PollCreated: expected TTL 5m, got %v", pub.records[1].opts.TTL)
	}

	// Verify TokenPriceUpdated -> Normal & 15m
	if pub.records[2].opts.Urgency != domain_broker.UrgencyNormal {
		t.Errorf("TokenPriceUpdated: expected UrgencyNormal (%d), got %d", domain_broker.UrgencyNormal, pub.records[2].opts.Urgency)
	}
	if pub.records[2].opts.TTL != 15*time.Minute {
		t.Errorf("TokenPriceUpdated: expected TTL 15m, got %v", pub.records[2].opts.TTL)
	}
}

func TestNotificationService_CallerOverride(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	// Override TokenPriceUpdated (default Normal/15m) to Realtime with custom 30s TTL
	err := srv.DispatchToUser(ctx, 10, nil, nil, payload.WSEventTokenPriceUpdated, nil,
		adapter.WithDispatchUrgency(domain_broker.UrgencyRealtime),
		adapter.WithDispatchTTL(30*time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(pub.records))
	}
	if pub.records[0].opts.Urgency != domain_broker.UrgencyRealtime {
		t.Errorf("override: expected UrgencyRealtime, got %d", pub.records[0].opts.Urgency)
	}
	if pub.records[0].opts.TTL != 30*time.Second {
		t.Errorf("override: expected TTL 30s, got %v", pub.records[0].opts.TTL)
	}
}

func TestNotificationService_DispatchToGroup_AutoResolvesMembersAndPublishesOnce(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups: map[int64]*entity.Group{
			5: {Title: "Trip Group"},
		},
		members: map[int64][]entity.GroupMember{
			5: {
				{UserID: 100},
				{UserID: 200},
				{UserID: 300},
			},
		},
	}
	pub := &mockPublisher{}
	srv := service.NewNotificationService(repo, groupRepo, nil, pub, testServiceConfig)

	actorID := int64(100)
	err := srv.DispatchToGroup(ctx, 5, &actorID, nil, payload.WSEventPollCreated, payload.PollCreatedData{
		PollID:   1,
		GroupID:  5,
		Question: "Where to go?",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify DB rows saved for ALL 3 members including actorID
	for _, uid := range []int64{100, 200, 300} {
		list, lErr := srv.GetNotifications(ctx, uid, dto.GetNotificationsQuery{Limit: 10})
		if lErr != nil {
			t.Fatalf("unexpected error fetching notifications for user %d: %v", uid, lErr)
		}
		if len(list.Items) != 1 {
			t.Fatalf("expected 1 notification for user %d, got %d", uid, len(list.Items))
		}
	}

	// Verify RabbitMQ received ONLY ONE message (TargetGroups), not duplicate TargetUsers + TargetGroups
	if len(pub.records) != 1 {
		t.Fatalf("expected exactly 1 RabbitMQ publish, got %d", len(pub.records))
	}
	ev := pub.records[0].payload
	if ev.TargetType != payload.TargetGroups {
		t.Fatalf("expected target type %s, got %s", payload.TargetGroups, ev.TargetType)
	}
	if len(ev.TargetIDs) != 1 || ev.TargetIDs[0] != 5 {
		t.Fatalf("expected target group ID [5], got %v", ev.TargetIDs)
	}
	if ev.SenderID != 100 {
		t.Fatalf("expected sender ID 100, got %d", ev.SenderID)
	}
}
