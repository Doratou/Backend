package decorator_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/application/service/decorator"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

// --- Shared Test Doubles ---

type dispatchCall struct {
	kind             string // "user" | "group"
	recipientID      int64
	groupID          *int64
	actorID          *int64
	recipientUserIDs []int64
	event            payload.WSEvent
	data             any
}

type mockNotifService struct {
	adapter.NotificationService
	calls []dispatchCall
}

func (m *mockNotifService) DispatchToUser(ctx context.Context, recipientID int64, actorID *int64, groupID *int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	m.calls = append(m.calls, dispatchCall{
		kind:        "user",
		recipientID: recipientID,
		actorID:     actorID,
		groupID:     groupID,
		event:       event,
		data:        data,
	})
	return nil
}

func (m *mockNotifService) DispatchToGroup(ctx context.Context, groupID int64, actorID *int64, recipientUserIDs []int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	gid := groupID
	m.calls = append(m.calls, dispatchCall{
		kind:             "group",
		groupID:          &gid,
		actorID:          actorID,
		recipientUserIDs: recipientUserIDs,
		event:            event,
		data:             data,
	})
	return nil
}

func (m *mockNotifService) Broadcast(ctx context.Context, actorID *int64, event payload.WSEvent, data any, opts ...adapter.DispatchOption) error {
	m.calls = append(m.calls, dispatchCall{
		kind:    "broadcast",
		actorID: actorID,
		event:   event,
		data:    data,
	})
	return nil
}

type mockUserRepo struct {
	domain_db.UserRepository
	users map[int64]*entity.User
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

type mockGroupRepo struct {
	domain_db.GroupRepository
	groups        map[int64]*entity.Group
	activeMembers map[int64]*entity.GroupMember
}

func (m *mockGroupRepo) FindByID(ctx context.Context, id int64) (*entity.Group, int64, error) {
	if g, ok := m.groups[id]; ok {
		return g, 3, nil
	}
	return nil, 0, errors.New("group not found")
}

func (m *mockGroupRepo) FindActiveMemberByID(ctx context.Context, groupID, memberID int64) (*entity.GroupMember, error) {
	if m.activeMembers != nil {
		if member, ok := m.activeMembers[memberID]; ok && member.GroupID == groupID {
			return member, nil
		}
	}
	return nil, errors.New("member not found")
}

type mockFriendshipRepo struct {
	domain_db.FriendshipRepository
	items map[int64]*entity.Friendship
}

func (m *mockFriendshipRepo) FindByID(ctx context.Context, id int64, lock bool) (*entity.Friendship, error) {
	if f, ok := m.items[id]; ok {
		return f, nil
	}
	return nil, errors.New("friendship not found")
}

type mockChecklistRepo struct {
	domain_db.ChecklistRepository
	items map[int64]*entity.ChecklistItem
}

func (m *mockChecklistRepo) FindItemForUpdate(ctx context.Context, groupID, itemID int64) (*entity.ChecklistItem, error) {
	if item, ok := m.items[itemID]; ok {
		return item, nil
	}
	return nil, errors.New("item not found")
}

// --- Service Stubs ---

type stubFriendshipService struct {
	adapter.FriendshipService
	sendFn   func(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error)
	acceptFn func(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error)
	rejectFn func(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error)
}

func (s *stubFriendshipService) SendRequest(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error) {
	return s.sendFn(ctx, userID, req)
}
func (s *stubFriendshipService) AcceptRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	return s.acceptFn(ctx, userID, requestID)
}
func (s *stubFriendshipService) RejectRequest(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
	return s.rejectFn(ctx, userID, requestID)
}

type stubSettlementService struct {
	adapter.SettlementService
	createFn func(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error)
}

func (s *stubSettlementService) Create(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error) {
	return s.createFn(ctx, userID, groupID, req)
}

type stubPollService struct {
	adapter.PollService
	createFn func(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error)
	closeFn  func(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error)
	getFn    func(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error)
}

func (s *stubPollService) Create(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error) {
	return s.createFn(ctx, userID, groupID, req)
}
func (s *stubPollService) Close(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error) {
	return s.closeFn(ctx, userID, groupID, pollID)
}
func (s *stubPollService) Get(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error) {
	return s.getFn(ctx, userID, groupID, pollID)
}

type stubLuckyDrawService struct {
	adapter.LuckyDrawService
	createFn func(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error)
	runFn    func(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error)
	getFn    func(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error)
}

func (s *stubLuckyDrawService) Create(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
	return s.createFn(ctx, userID, groupID, req)
}
func (s *stubLuckyDrawService) Run(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	return s.runFn(ctx, userID, groupID, drawID)
}
func (s *stubLuckyDrawService) Get(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
	return s.getFn(ctx, userID, groupID, drawID)
}

type stubChecklistService struct {
	adapter.ChecklistService
	createFn func(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error)
	claimFn  func(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error)
	deleteFn func(ctx context.Context, userID, groupID, itemID int64) error
}

func (s *stubChecklistService) CreateItem(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error) {
	return s.createFn(ctx, userID, groupID, req)
}
func (s *stubChecklistService) SetClaim(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error) {
	return s.claimFn(ctx, userID, groupID, itemID, claimed)
}
func (s *stubChecklistService) DeleteItem(ctx context.Context, userID, groupID, itemID int64) error {
	return s.deleteFn(ctx, userID, groupID, itemID)
}

// --- Tests ---

func TestFriendshipNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	avID := int64(99)
	users := &mockUserRepo{users: map[int64]*entity.User{
		10: {Name: "Ali", AvatarAttachmentID: &avID},
		20: {Name: "Reza"},
	}}
	fRepo := &mockFriendshipRepo{items: map[int64]*entity.Friendship{
		55: {BaseEntity: entity.BaseEntity{ID: 55}, RequesterUserID: 10, ReceiverUserID: 20},
	}}
	notif := &mockNotifService{}

	now := time.Now().UTC()
	stub := &stubFriendshipService{
		sendFn: func(ctx context.Context, userID int64, req dto.CreateFriendshipRequest) (*dto.FriendRequestResponse, error) {
			if req.ReceiverUserID == 999 {
				return nil, errors.New("boom")
			}
			return &dto.FriendRequestResponse{ID: 55, RequesterUserID: userID, ReceiverUserID: req.ReceiverUserID, CreatedAt: now}, nil
		},
		acceptFn: func(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
			return &dto.FriendRequestStatusResponse{ID: requestID, Status: enum.FriendshipStatusAccepted, RespondedAt: &now}, nil
		},
		rejectFn: func(ctx context.Context, userID, requestID int64) (*dto.FriendRequestStatusResponse, error) {
			return &dto.FriendRequestStatusResponse{ID: requestID, Status: enum.FriendshipStatusRejected, RespondedAt: &now}, nil
		},
	}

	svc := decorator.NewFriendshipNotificationDecorator(stub, notif, fRepo, users)

	// Error case should not dispatch
	if _, err := svc.SendRequest(ctx, 10, dto.CreateFriendshipRequest{ReceiverUserID: 999}); err == nil {
		t.Fatal("expected error")
	}
	if len(notif.calls) != 0 {
		t.Fatalf("expected 0 dispatches on error, got %d", len(notif.calls))
	}

	// SendRequest -> WSEventFriendRequestReceived to receiver 20
	if _, err := svc.SendRequest(ctx, 10, dto.CreateFriendshipRequest{ReceiverUserID: 20}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// AcceptRequest -> WSEventFriendRequestResponded ("accepted") to requester 10
	if _, err := svc.AcceptRequest(ctx, 20, 55); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// RejectRequest -> WSEventFriendRequestResponded ("rejected") to requester 10
	if _, err := svc.RejectRequest(ctx, 20, 55); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 3 {
		t.Fatalf("expected 3 dispatches, got %d", len(notif.calls))
	}

	c0 := notif.calls[0]
	if c0.recipientID != 20 || c0.event != payload.WSEventFriendRequestReceived {
		t.Errorf("unexpected SendRequest dispatch: %+v", c0)
	}
	d0 := c0.data.(payload.FriendRequestReceivedData)
	if d0.SenderName != "Ali" || d0.SenderAvatarID != 99 {
		t.Errorf("unexpected SendRequest payload: %+v", d0)
	}

	c1 := notif.calls[1]
	d1 := c1.data.(payload.FriendRequestRespondedData)
	if c1.recipientID != 10 || c1.event != payload.WSEventFriendRequestResponded || d1.Status != "accepted" || d1.ResponderName != "Reza" {
		t.Errorf("unexpected AcceptRequest dispatch: %+v payload=%+v", c1, d1)
	}

	c2 := notif.calls[2]
	d2 := c2.data.(payload.FriendRequestRespondedData)
	if c2.recipientID != 10 || c2.event != payload.WSEventFriendRequestResponded || d2.Status != "rejected" {
		t.Errorf("unexpected RejectRequest dispatch: %+v payload=%+v", c2, d2)
	}
}

func TestSettlementNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	groups := &mockGroupRepo{groups: map[int64]*entity.Group{7: {Title: "North Trip"}}}
	notif := &mockNotifService{}
	paidAt := time.Now().UTC()

	stub := &stubSettlementService{
		createFn: func(ctx context.Context, userID, groupID int64, req dto.CreateSettlementRequest) (*dto.SettlementResponse, error) {
			return &dto.SettlementResponse{
				ID:          101,
				GroupID:     groupID,
				FromMember:  dto.SettlementMemberResponse{MemberID: 1, UserID: userID, Name: "Sara"},
				ToMember:    dto.SettlementMemberResponse{MemberID: 2, UserID: 88, Name: "Omid"},
				AmountToman: req.AmountToman,
				PaidAt:      paidAt,
			}, nil
		},
	}

	svc := decorator.NewSettlementNotificationDecorator(stub, notif, groups)
	if _, err := svc.Create(ctx, 11, 7, dto.CreateSettlementRequest{ToMemberID: 2, AmountToman: 500000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 1 {
		t.Fatalf("expected 1 dispatch, got %d", len(notif.calls))
	}
	c := notif.calls[0]
	if c.recipientID != 88 || c.event != payload.WSEventSettlementPaid {
		t.Fatalf("expected creditor 88 and settlement.paid, got %+v", c)
	}
	data := c.data.(payload.SettlementPaidData)
	if data.GroupName != "North Trip" || data.DebtorName != "Sara" || data.CreditorName != "Omid" || data.AmountToman != 500000 {
		t.Errorf("unexpected payload: %+v", data)
	}
}

func TestPollNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	groups := &mockGroupRepo{groups: map[int64]*entity.Group{3: {Title: "Campus"}}}
	users := &mockUserRepo{users: map[int64]*entity.User{1: {Name: "Mina"}}}
	notif := &mockNotifService{}
	closedAt := time.Now().UTC()

	stub := &stubPollService{
		createFn: func(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error) {
			return &dto.PollResponse{
				ID:       40,
				GroupID:  groupID,
				Question: req.Question,
				Options: []dto.PollOptionResponse{
					{ID: 1, Title: "Pizza"},
					{ID: 2, Title: "Burger"},
				},
			}, nil
		},
		closeFn: func(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error) {
			return &dto.PollStatusResponse{ID: pollID, Status: enum.PollStatusInactive, ClosedAt: &closedAt}, nil
		},
		getFn: func(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error) {
			return &dto.PollResponse{
				ID:         pollID,
				GroupID:    groupID,
				Question:   "Dinner?",
				TotalVotes: 4,
				Options: []dto.PollOptionResponse{
					{ID: 1, Title: "Pizza", VoteCount: 3},
					{ID: 2, Title: "Burger", VoteCount: 1},
				},
			}, nil
		},
	}

	svc := decorator.NewPollNotificationDecorator(stub, notif, groups, users)
	if _, err := svc.Create(ctx, 1, 3, dto.CreatePollRequest{Question: "Dinner?", Options: []string{"Pizza", "Burger"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Close(ctx, 1, 3, 40); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 2 {
		t.Fatalf("expected 2 dispatches, got %d", len(notif.calls))
	}

	created := notif.calls[0].data.(payload.PollCreatedData)
	if notif.calls[0].event != payload.WSEventPollCreated || created.GroupName != "Campus" || created.CreatorName != "Mina" || len(created.Options) != 2 {
		t.Errorf("unexpected PollCreated payload: %+v", created)
	}

	closed := notif.calls[1].data.(payload.PollClosedData)
	if notif.calls[1].event != payload.WSEventPollClosed || closed.TotalVotes != 4 || len(closed.WinningOptions) != 1 || closed.WinningOptions[0].Text != "Pizza" || closed.Results[0].Percentage != 75 {
		t.Errorf("unexpected PollClosed payload: %+v", closed)
	}
}

func TestLuckyDrawNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	groups := &mockGroupRepo{groups: map[int64]*entity.Group{5: {Title: "Friends"}}}
	users := &mockUserRepo{users: map[int64]*entity.User{
		1: {Name: "Nima"},
		2: {Name: "Tara"},
	}}
	notif := &mockNotifService{}
	completedAt := time.Now().UTC()

	drawStatus := enum.LuckyDrawStatusDraft
	stub := &stubLuckyDrawService{
		createFn: func(ctx context.Context, userID, groupID int64, req dto.CreateLuckyDrawRequest) (*dto.LuckyDrawResponse, error) {
			return &dto.LuckyDrawResponse{
				LuckyDrawListItem: dto.LuckyDrawListItem{ID: 12, GroupID: groupID, Title: req.Title, Status: enum.LuckyDrawStatusDraft, CreatedAt: completedAt},
				Candidates: []dto.LuckyDrawCandidateResponse{
					{Type: enum.LuckyDrawCandidateTypeUser, Value: json.RawMessage("2")},
					{Type: enum.LuckyDrawCandidateTypeText, Value: json.RawMessage(`"Coffee"`)},
				},
			}, nil
		},
		getFn: func(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
			return &dto.LuckyDrawResponse{LuckyDrawListItem: dto.LuckyDrawListItem{ID: drawID, GroupID: groupID, Status: drawStatus}}, nil
		},
		runFn: func(ctx context.Context, userID, groupID, drawID int64) (*dto.LuckyDrawResponse, error) {
			drawStatus = enum.LuckyDrawStatusCompleted
			return &dto.LuckyDrawResponse{
				LuckyDrawListItem: dto.LuckyDrawListItem{
					ID: drawID, GroupID: groupID, Title: "Who pays?", Status: enum.LuckyDrawStatusCompleted, CompletedAt: &completedAt,
					Result: &dto.LuckyDrawResult{
						Index:                      0,
						LuckyDrawCandidateResponse: dto.LuckyDrawCandidateResponse{Type: enum.LuckyDrawCandidateTypeUser, Value: json.RawMessage("2")},
					},
				},
			}, nil
		},
	}

	svc := decorator.NewLuckyDrawNotificationDecorator(stub, notif, groups, users)
	if _, err := svc.Create(ctx, 1, 5, dto.CreateLuckyDrawRequest{Title: "Who pays?"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Run(ctx, 1, 5, 12); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Idempotent second Run call must NOT dispatch duplicate notification
	if _, err := svc.Run(ctx, 1, 5, 12); err != nil {
		t.Fatalf("unexpected error on repeat run: %v", err)
	}

	if len(notif.calls) != 2 {
		t.Fatalf("expected 2 dispatches (create + first run), got %d", len(notif.calls))
	}
	dCreated := notif.calls[0].data.(payload.DrawCreatedData)
	if len(dCreated.Candidates) != 2 || dCreated.Candidates[0] != "Tara" || dCreated.Candidates[1] != "Coffee" {
		t.Errorf("unexpected DrawCreated candidates: %+v", dCreated.Candidates)
	}
	dResult := notif.calls[1].data.(payload.DrawResultData)
	if dResult.WinnerID != "2" || dResult.WinnerName != "Tara" {
		t.Errorf("unexpected DrawResult winner: %+v", dResult)
	}
}

func TestChecklistNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	groups := &mockGroupRepo{groups: map[int64]*entity.Group{9: {Title: "Hiking"}}}
	users := &mockUserRepo{users: map[int64]*entity.User{1: {Name: "Kian"}}}
	cRepo := &mockChecklistRepo{items: map[int64]*entity.ChecklistItem{
		50: {BaseEntity: entity.BaseEntity{ID: 50}, AssignedToMemberID: nil},
	}}
	notif := &mockNotifService{}
	now := time.Now().UTC()

	stub := &stubChecklistService{
		createFn: func(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error) {
			return &dto.ChecklistItemResponse{
				ID:        51,
				Title:     req.Title,
				Type:      req.Type,
				CreatedBy: &dto.ChecklistMemberResponse{UserID: userID, Name: "Kian"},
				UpdatedAt: now,
			}, nil
		},
		claimFn: func(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error) {
			if claimed {
				mid := int64(10)
				cRepo.items[itemID].AssignedToMemberID = &mid
			} else {
				cRepo.items[itemID].AssignedToMemberID = nil
			}
			return &dto.ChecklistItemResponse{ID: itemID, UpdatedAt: now}, nil
		},
		deleteFn: func(ctx context.Context, userID, groupID, itemID int64) error {
			return nil
		},
	}

	svc := decorator.NewChecklistNotificationDecorator(stub, notif, groups, users, cRepo)

	// 1. Create PERSONAL item -> notifies
	if _, err := svc.CreateItem(ctx, 1, 9, dto.CreateChecklistItemRequest{Title: "Water", Type: enum.ChecklistItemTypePersonal}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2. Claim item (nil -> claimed) -> notifies
	if _, err := svc.SetClaim(ctx, 1, 9, 50, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3. No-op claim (already claimed -> claimed) -> must NOT notify
	if _, err := svc.SetClaim(ctx, 1, 9, 50, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4. Delete item -> notifies
	if err := svc.DeleteItem(ctx, 1, 9, 51); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 3 {
		t.Fatalf("expected 3 checklist.updated dispatches, got %d", len(notif.calls))
	}
	for i, c := range notif.calls {
		if c.event != payload.WSEventChecklistUpdated {
			t.Errorf("call %d: expected checklist.updated, got %s", i, c.event)
		}
		d := c.data.(payload.ChecklistUpdatedData)
		if d.GroupName != "Hiking" || d.UpdatedByName != "Kian" {
			t.Errorf("call %d: unexpected payload %+v", i, d)
		}
	}
}
