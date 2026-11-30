package decorator_test

import (
	"context"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/decorator"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

// --- Service Stubs ---

type stubAdminReportService struct {
	adapter.AdminReportService
	reviewFn func(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error)
}

func (s *stubAdminReportService) ReviewReport(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error) {
	return s.reviewFn(ctx, id, reviewerUserID, req)
}

type stubInviteService struct {
	adapter.InviteService
	sendDirectFn func(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error)
	respondFn    func(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error)
	joinByLinkFn func(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error)
}

func (s *stubInviteService) SendDirect(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error) {
	return s.sendDirectFn(ctx, userID, role, req)
}
func (s *stubInviteService) RespondToDirect(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error) {
	return s.respondFn(ctx, userID, inviteID, response)
}
func (s *stubInviteService) JoinByLink(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error) {
	return s.joinByLinkFn(ctx, userID, token, password)
}

type stubGroupService struct {
	adapter.GroupService
	leaveFn  func(ctx context.Context, userID, groupID int64) error
	removeFn func(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error
}

func (s *stubGroupService) LeaveGroup(ctx context.Context, userID, groupID int64) error {
	return s.leaveFn(ctx, userID, groupID)
}
func (s *stubGroupService) RemoveMember(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error {
	return s.removeFn(ctx, userID, role, groupID, memberID)
}

type stubTokenPriceService struct {
	adapter.TokenPriceService
	updateFn func(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error)
}

func (s *stubTokenPriceService) UpdatePrice(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error) {
	return s.updateFn(ctx, adminUserID, req)
}

// --- Tests ---

func TestAdminReportNotificationDecorator_TerminalOnly(t *testing.T) {
	ctx := context.Background()
	users := &mockUserRepo{users: map[int64]*entity.User{99: {Name: "Moderator"}}}
	notif := &mockNotifService{}
	reporterID := int64(42)
	resolvedAt := time.Now().UTC()

	stub := &stubAdminReportService{
		reviewFn: func(ctx context.Context, id int64, reviewerUserID int64, req dto.ReviewReportRequest) (*dto.ReportResponse, error) {
			resp := req.Response
			return &dto.ReportResponse{
				ID:             id,
				ReporterUserID: &reporterID,
				Reason:         enum.ReportReasonSpam,
				Status:         req.Status,
				Response:       &resp,
				ReviewedByUser: &dto.ReportUserSummary{ID: reviewerUserID, Name: "Moderator"},
				ResolvedAt:     &resolvedAt,
			}, nil
		},
	}

	svc := decorator.NewAdminReportNotificationDecorator(stub, notif, users)

	// 1. Non-terminal status UNDER_REVIEW -> must NOT dispatch
	if _, err := svc.ReviewReport(ctx, 10, 99, dto.ReviewReportRequest{Status: enum.ReportStatusUnderReview, Response: "Looking into it"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notif.calls) != 0 {
		t.Fatalf("expected 0 dispatches for non-terminal status, got %d", len(notif.calls))
	}

	// 2. Terminal status RESOLVED -> must dispatch admin_report.answered
	if _, err := svc.ReviewReport(ctx, 10, 99, dto.ReviewReportRequest{Status: enum.ReportStatusResolved, Response: "Issue resolved"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notif.calls) != 1 {
		t.Fatalf("expected 1 dispatch for RESOLVED, got %d", len(notif.calls))
	}
	c := notif.calls[0]
	if c.recipientID != 42 || c.event != payload.WSEventAdminReportAnswered {
		t.Fatalf("unexpected dispatch target/event: %+v", c)
	}
	d := c.data.(payload.AdminReportAnsweredData)
	if d.Answer != "Issue resolved" || d.AdminName != "Moderator" || d.Subject != string(enum.ReportReasonSpam) {
		t.Errorf("unexpected payload: %+v", d)
	}
}

func TestInviteNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	avID := int64(77)
	groups := &mockGroupRepo{groups: map[int64]*entity.Group{3: {Title: "Shiraz Trip"}}}
	users := &mockUserRepo{users: map[int64]*entity.User{
		1: {Name: "Omid", AvatarAttachmentID: &avID},
		2: {Name: "Nazanin"},
		4: {Name: "Kourosh"},
	}}
	notif := &mockNotifService{}
	now := time.Now().UTC()

	stub := &stubInviteService{
		sendDirectFn: func(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error) {
			return []dto.DirectInviteSendResult{
				{UserID: 2, Success: true, Invite: &dto.DirectInviteResponse{ID: 501, CreatedAt: now}},
				{UserID: 3, Success: false, Error: &dto.DirectInviteError{Code: "NOT_FRIEND"}},
			}, nil
		},
		respondFn: func(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error) {
			return &dto.DirectInviteResponse{
				ID:          inviteID,
				Status:      response,
				RespondedAt: &now,
				Group:       &dto.DirectInviteGroup{ID: 3, Title: "Shiraz Trip"},
				Recipient:   &dto.DirectInviteUser{ID: userID, Name: "Nazanin"},
			}, nil
		},
		joinByLinkFn: func(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error) {
			return &dto.GroupResponse{
				ID:    3,
				Title: "Shiraz Trip",
				Members: []dto.GroupMemberResponse{
					{UserID: userID, Name: "Kourosh", JoinedAt: now},
				},
			}, nil
		},
	}

	svc := decorator.NewInviteNotificationDecorator(stub, notif, groups, users)

	// 1. SendDirect -> only user 2 succeeded
	_, err := svc.SendDirect(ctx, 1, enum.UserRoleUser, dto.CreateDirectInviteRequest{GroupID: 3, UserIDs: []int64{2, 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2. RespondToDirect REJECTED -> no group.member_joined
	if _, err := svc.RespondToDirect(ctx, 2, 501, string(enum.DirectInviteStatusRejected)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3. RespondToDirect ACCEPTED -> group.member_joined
	if _, err := svc.RespondToDirect(ctx, 2, 501, string(enum.DirectInviteStatusAccepted)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4. JoinByLink -> group.member_joined
	if _, err := svc.JoinByLink(ctx, 4, "tok", "pass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 3 {
		t.Fatalf("expected 3 dispatches, got %d", len(notif.calls))
	}

	inv := notif.calls[0]
	invData := inv.data.(payload.GroupInviteReceivedData)
	if inv.recipientID != 2 || inv.event != payload.WSEventGroupInviteReceived || invData.GroupName != "Shiraz Trip" || invData.SenderName != "Omid" || invData.SenderAvatarID != 77 {
		t.Errorf("unexpected invite dispatch: %+v data=%+v", inv, invData)
	}

	j1 := notif.calls[1]
	j1Data := j1.data.(payload.GroupMemberJoinedData)
	if j1.kind != "group" || *j1.groupID != 3 || j1.event != payload.WSEventGroupMemberJoined || j1Data.UserName != "Nazanin" {
		t.Errorf("unexpected direct join dispatch: %+v data=%+v", j1, j1Data)
	}

	j2 := notif.calls[2]
	j2Data := j2.data.(payload.GroupMemberJoinedData)
	if j2.kind != "group" || *j2.groupID != 3 || j2.event != payload.WSEventGroupMemberJoined || j2Data.UserName != "Kourosh" {
		t.Errorf("unexpected link join dispatch: %+v data=%+v", j2, j2Data)
	}
}

func TestGroupNotificationDecorator_LeaveAndKick(t *testing.T) {
	ctx := context.Background()
	groups := &mockGroupRepo{
		groups: map[int64]*entity.Group{8: {Title: "Mountain Club"}},
		activeMembers: map[int64]*entity.GroupMember{
			15: {ID: 15, GroupID: 8, UserID: 30},
		},
	}
	users := &mockUserRepo{users: map[int64]*entity.User{
		20: {Name: "Babak"},
		30: {Name: "Ramin"},
	}}
	notif := &mockNotifService{}

	stub := &stubGroupService{
		leaveFn: func(ctx context.Context, userID, groupID int64) error {
			return nil
		},
		removeFn: func(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error {
			// Simulate deactivation inside RemoveMember
			delete(groups.activeMembers, memberID)
			return nil
		},
	}

	svc := decorator.NewGroupNotificationDecorator(stub, notif, groups, users)

	// 1. LeaveGroup -> notifies group with reason "left"
	if err := svc.LeaveGroup(ctx, 20, 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2. RemoveMember -> notifies BOTH remaining group AND kicked user 30 with reason "kicked"
	if err := svc.RemoveMember(ctx, 1, enum.UserRoleUser, 8, 15); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.calls) != 3 {
		t.Fatalf("expected 3 dispatches (1 leave + 1 group kick + 1 kicked user), got %d", len(notif.calls))
	}

	left := notif.calls[0]
	leftData := left.data.(payload.GroupMemberLeftData)
	if left.kind != "group" || left.event != payload.WSEventGroupMemberLeft || leftData.Reason != "left" || leftData.UserName != "Babak" {
		t.Errorf("unexpected leave dispatch: %+v data=%+v", left, leftData)
	}

	kickGroup := notif.calls[1]
	kickGroupData := kickGroup.data.(payload.GroupMemberLeftData)
	if kickGroup.kind != "group" || *kickGroup.groupID != 8 || kickGroup.event != payload.WSEventGroupMemberLeft || kickGroupData.Reason != "kicked" || kickGroupData.UserName != "Ramin" {
		t.Errorf("unexpected kick group dispatch: %+v data=%+v", kickGroup, kickGroupData)
	}

	kickUser := notif.calls[2]
	kickUserData := kickUser.data.(payload.GroupMemberLeftData)
	if kickUser.kind != "user" || kickUser.recipientID != 30 || kickUser.event != payload.WSEventGroupMemberLeft || kickUserData.Reason != "kicked" {
		t.Errorf("unexpected kick user dispatch: %+v data=%+v", kickUser, kickUserData)
	}
}

func TestTokenPriceNotificationDecorator(t *testing.T) {
	ctx := context.Background()
	notif := &mockNotifService{}
	now := time.Now().UTC()
	current := int64(10000)

	stub := &stubTokenPriceService{
		updateFn: func(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error) {
			old := current
			current = req.PriceToman
			return &dto.TokenPriceResponse{OldPriceToman: old, NewPriceToman: current, UpdatedAt: now}, nil
		},
	}

	svc := decorator.NewTokenPriceNotificationDecorator(stub, notif)

	// 1. Unchanged price -> must NOT broadcast
	if _, err := svc.UpdatePrice(ctx, 1, dto.UpdateTokenPriceRequest{PriceToman: 10000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notif.calls) != 0 {
		t.Fatalf("expected 0 broadcasts when price unchanged, got %d", len(notif.calls))
	}

	// 2. Changed price -> broadcasts token.price_updated
	if _, err := svc.UpdatePrice(ctx, 1, dto.UpdateTokenPriceRequest{PriceToman: 12500}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notif.calls) != 1 {
		t.Fatalf("expected 1 broadcast, got %d", len(notif.calls))
	}
	c := notif.calls[0]
	d := c.data.(payload.TokenPriceUpdatedData)
	if c.kind != "broadcast" || c.event != payload.WSEventTokenPriceUpdated || d.OldPriceToman != 10000 || d.NewPriceToman != 12500 {
		t.Errorf("unexpected broadcast dispatch: %+v data=%+v", c, d)
	}
}
