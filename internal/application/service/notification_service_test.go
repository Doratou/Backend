package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/application/service"
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
func (m *mockGroupRepo) Update(ctx context.Context, group *entity.Group) error       { return nil }
func (m *mockGroupRepo) UpdateCover(ctx context.Context, group *entity.Group) error  { return nil }
func (m *mockGroupRepo) SetStatus(ctx context.Context, groupID int64, status enum.GroupStatus) error {
	return nil
}
func (m *mockGroupRepo) Delete(ctx context.Context, groupID int64) error { return nil }
func (m *mockGroupRepo) CloneWithMembers(ctx context.Context, sourceGroupID int64, newGroup *entity.Group) error {
	return nil
}

func TestNotificationService_GetAndMarkAsRead(t *testing.T) {
	ctx := context.Background()
	repo := newMockNotificationRepo()
	groupRepo := &mockGroupRepo{
		groups:  make(map[int64]*entity.Group),
		members: make(map[int64][]entity.GroupMember),
	}
	srv := service.NewNotificationService(repo, groupRepo, nil)

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
	srv := service.NewNotificationService(repo, groupRepo, nil)

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

func TestNotificationService_CreateGroupAnnouncement_ExcludesOwner(t *testing.T) {
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
	srv := service.NewNotificationService(repo, groupRepo, nil)

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
}
