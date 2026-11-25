package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// NotificationSeedDeps bundles the live, decorated application services used to emit
// genuine end-to-end notifications across every supported event category on startup.
type NotificationSeedDeps struct {
	DB                  database.Database
	FriendshipService   adapter.FriendshipService
	SettlementService   adapter.SettlementService
	PollService         adapter.PollService
	LuckyDrawService    adapter.LuckyDrawService
	ChecklistService    adapter.ChecklistService
	AdminReportService  adapter.AdminReportService
	GroupService        adapter.GroupService
	InviteService       adapter.InviteService
	TokenPriceService   adapter.TokenPriceService
	NotificationService adapter.NotificationService
}

// SeedNotificationEvents runs real application service workflows (backed by Postgres and RabbitMQ)
// so that at least one notification is produced and persisted for every event category.
// Activated only when NOTIFICATION_SEED_DEMO_EVENTS=true in .env.
func SeedNotificationEvents(ctx context.Context, deps NotificationSeedDeps) error {
	if deps.DB == nil {
		return fmt.Errorf("database is required for notification demo seeding")
	}
	gdb := deps.DB.GetGormDB().WithContext(ctx)

	alice, err := loadSeedUserByEmail(gdb, "owner@test.com", enum.UserRoleUser)
	if err != nil {
		return fmt.Errorf("resolve alice (owner@test.com): %w", err)
	}
	bob, err := loadSeedUserByEmail(gdb, "member@test.com", enum.UserRoleUser)
	if err != nil {
		return fmt.Errorf("resolve bob (member@test.com): %w", err)
	}
	sara, err := loadSeedUserByEmail(gdb, "sara@test.com", enum.UserRoleUser)
	if err != nil {
		return fmt.Errorf("resolve sara (sara@test.com): %w", err)
	}
	reza, err := loadSeedUserByEmail(gdb, "reza@test.com", enum.UserRoleUser)
	if err != nil {
		return fmt.Errorf("resolve reza (reza@test.com): %w", err)
	}
	nima, err := loadSeedUserByEmail(gdb, "nima@test.com", enum.UserRoleUser)
	if err != nil {
		return fmt.Errorf("resolve nima (nima@test.com): %w", err)
	}
	admin, err := loadSeedUserByEmail(gdb, "admin@test.com", enum.UserRoleAdmin)
	if err != nil {
		return fmt.Errorf("resolve admin (admin@test.com): %w", err)
	}

	// Reset demo friendship pairs so repeated runs always execute SendRequest / Accept / Reject cleanly.
	if err := resetFriendshipPair(gdb, alice.ID, reza.ID); err != nil {
		return fmt.Errorf("reset friendship (alice, reza): %w", err)
	}
	if err := resetFriendshipPair(gdb, sara.ID, nima.ID); err != nil {
		return fmt.Errorf("reset friendship (sara, nima): %w", err)
	}

	// 1. Friend Request: received + accepted + rejected
	fr1, err := deps.FriendshipService.SendRequest(ctx, alice.ID, dto.CreateFriendshipRequest{ReceiverUserID: reza.ID})
	if err != nil {
		return fmt.Errorf("seed friend_request.received (alice->reza): %w", err)
	}
	if _, err := deps.FriendshipService.AcceptRequest(ctx, reza.ID, fr1.ID); err != nil {
		return fmt.Errorf("seed friend_request.responded (accepted): %w", err)
	}
	fr2, err := deps.FriendshipService.SendRequest(ctx, sara.ID, dto.CreateFriendshipRequest{ReceiverUserID: nima.ID})
	if err != nil {
		return fmt.Errorf("seed friend_request.received (sara->nima): %w", err)
	}
	if _, err := deps.FriendshipService.RejectRequest(ctx, nima.ID, fr2.ID); err != nil {
		return fmt.Errorf("seed friend_request.responded (rejected): %w", err)
	}

	// 2. Fresh Group + Link Invite + Direct Invite + Joins + Leave + Kick
	groupRes, err := deps.GroupService.Create(ctx, alice.ID, dto.CreateGroupRequest{
		Title: fmt.Sprintf("گروه تست اعلان‌ها (%s)", time.Now().Format("15:04:05")),
	})
	if err != nil {
		return fmt.Errorf("seed demo group: %w", err)
	}

	const linkPassword = "SeedPass123!"
	linkRes, err := deps.InviteService.CreateLink(ctx, alice.ID, enum.UserRoleUser, groupRes.ID, dto.CreateLinkInviteRequest{
		Password: linkPassword,
	})
	if err != nil {
		return fmt.Errorf("seed link invite: %w", err)
	}
	if _, err := deps.InviteService.JoinByLink(ctx, bob.ID, linkRes.Token, linkPassword); err != nil {
		return fmt.Errorf("seed group.member_joined (bob via link): %w", err)
	}
	if _, err := deps.InviteService.JoinByLink(ctx, sara.ID, linkRes.Token, linkPassword); err != nil {
		return fmt.Errorf("seed group.member_joined (sara via link): %w", err)
	}

	directResults, err := deps.InviteService.SendDirect(ctx, alice.ID, enum.UserRoleUser, dto.CreateDirectInviteRequest{
		GroupID: groupRes.ID,
		UserIDs: []int64{reza.ID},
	})
	if err != nil {
		return fmt.Errorf("seed group.invite_received: %w", err)
	}
	if len(directResults) == 0 || !directResults[0].Success || directResults[0].Invite == nil {
		return fmt.Errorf("seed group.invite_received failed for reza: %+v", directResults)
	}
	if _, err := deps.InviteService.RespondToDirect(ctx, reza.ID, directResults[0].Invite.ID, string(enum.DirectInviteStatusAccepted)); err != nil {
		return fmt.Errorf("seed group.member_joined (reza via direct): %w", err)
	}

	// Resolve member IDs before leave/kick
	members, err := deps.GroupService.ListMembers(ctx, alice.ID, enum.UserRoleUser, groupRes.ID)
	if err != nil {
		return fmt.Errorf("list group members: %w", err)
	}
	var aliceMemberID, rezaMemberID int64
	for _, m := range members {
		if m.UserID == alice.ID {
			aliceMemberID = m.MemberID
		} else if m.UserID == reza.ID {
			rezaMemberID = m.MemberID
		}
	}
	if aliceMemberID == 0 || rezaMemberID == 0 {
		return fmt.Errorf("failed to resolve member IDs in group %d", groupRes.ID)
	}

	if err := deps.GroupService.LeaveGroup(ctx, sara.ID, groupRes.ID); err != nil {
		return fmt.Errorf("seed group.member_left (sara leave): %w", err)
	}
	if err := deps.GroupService.RemoveMember(ctx, alice.ID, enum.UserRoleUser, groupRes.ID, rezaMemberID); err != nil {
		return fmt.Errorf("seed group.member_left (reza kick): %w", err)
	}

	// 3. Settlement Paid (Bob pays Alice)
	if _, err := deps.SettlementService.Create(ctx, bob.ID, groupRes.ID, dto.CreateSettlementRequest{
		ToMemberID:  aliceMemberID,
		AmountToman: 150000,
	}); err != nil {
		return fmt.Errorf("seed settlement.paid: %w", err)
	}

	// 4. Poll Create + Vote + Close
	pollRes, err := deps.PollService.Create(ctx, alice.ID, groupRes.ID, dto.CreatePollRequest{
		Question:   "سفر بعدی کجا بریم؟",
		Visibility: enum.PollVisibilityPublic,
		Options:    []string{"شمال", "شیراز"},
	})
	if err != nil {
		return fmt.Errorf("seed poll.created: %w", err)
	}
	if len(pollRes.Options) > 0 {
		if _, err := deps.PollService.Vote(ctx, bob.ID, groupRes.ID, pollRes.ID, dto.VotePollRequest{
			OptionIDs: []int64{pollRes.Options[0].ID},
		}); err != nil {
			return fmt.Errorf("seed poll vote: %w", err)
		}
	}
	if _, err := deps.PollService.Close(ctx, alice.ID, groupRes.ID, pollRes.ID); err != nil {
		return fmt.Errorf("seed poll.closed: %w", err)
	}

	// 5. Lucky Draw Create + Run
	aliceVal, _ := json.Marshal(alice.ID)
	bobVal, _ := json.Marshal(bob.ID)
	drawRes, err := deps.LuckyDrawService.Create(ctx, alice.ID, groupRes.ID, dto.CreateLuckyDrawRequest{
		Title:    "قرعه‌کشی دنگ قهوه",
		DrawType: enum.LuckyDrawTypeEqual,
		Candidates: []dto.LuckyDrawCandidateInput{
			{Type: enum.LuckyDrawCandidateTypeUser, Value: aliceVal},
			{Type: enum.LuckyDrawCandidateTypeUser, Value: bobVal},
		},
	})
	if err != nil {
		return fmt.Errorf("seed draw.created: %w", err)
	}
	if _, err := deps.LuckyDrawService.Run(ctx, alice.ID, groupRes.ID, drawRes.ID); err != nil {
		return fmt.Errorf("seed draw.result: %w", err)
	}

	// 6. Checklist Updated (Create + Claim + Delete)
	itemRes, err := deps.ChecklistService.CreateItem(ctx, alice.ID, groupRes.ID, dto.CreateChecklistItemRequest{
		Title: "چادر مسافرتی",
		Type:  enum.ChecklistItemTypeGroup,
	})
	if err != nil {
		return fmt.Errorf("seed checklist.updated (create): %w", err)
	}
	if _, err := deps.ChecklistService.SetClaim(ctx, bob.ID, groupRes.ID, itemRes.ID, true); err != nil {
		return fmt.Errorf("seed checklist.updated (claim): %w", err)
	}
	if err := deps.ChecklistService.DeleteItem(ctx, alice.ID, groupRes.ID, itemRes.ID); err != nil {
		return fmt.Errorf("seed checklist.updated (delete): %w", err)
	}

	// 7. Admin Report Answered (Bob creates report -> Admin resolves it)
	repRes, err := deps.AdminReportService.CreateReport(ctx, bob.ID, dto.CreateReportRequest{
		ReportedUserID: &nima.ID,
		Reason:         enum.ReportReasonSpam,
		Description:    "گزارش آزمایشی جهت تولید واقعی اعلان پاسخ ادمین",
	})
	if err != nil {
		return fmt.Errorf("seed admin report create: %w", err)
	}
	if _, err := deps.AdminReportService.ReviewReport(ctx, repRes.ID, admin.ID, dto.ReviewReportRequest{
		Status:   enum.ReportStatusResolved,
		Response: "گزارش شما بررسی و تایید شد.",
	}); err != nil {
		return fmt.Errorf("seed admin_report.answered: %w", err)
	}

	// 8. Token Price Updated (guarantee price actually changes so Broadcast fires)
	var currentToken entity.TokenPrice
	_ = gdb.Where("id = ?", 1).First(&currentToken).Error
	newPrice := currentToken.PriceToman + 500
	if currentToken.PriceToman <= 0 {
		newPrice = 15000
	}
	if _, err := deps.TokenPriceService.UpdatePrice(ctx, admin.ID, dto.UpdateTokenPriceRequest{
		PriceToman: newPrice,
	}); err != nil {
		return fmt.Errorf("seed token.price_updated: %w", err)
	}

	// 9. Group Announcement (Owner manual broadcast)
	if _, err := deps.NotificationService.CreateGroupAnnouncement(ctx, alice.ID, groupRes.ID, dto.GroupAnnouncementRequest{
		Title:   "اطلاعیه تست واقعی اعلان‌ها",
		Message: "این اعلان به صورت واقعی پس از بالا آمدن سرویس‌ها برای تمام دسته‌ها تولید شد.",
	}); err != nil {
		return fmt.Errorf("seed group.announcement: %w", err)
	}

	log.Printf("[SEED NOTIFICATIONS] Successfully emitted REAL notifications for all 14 event categories (demo group_id=%d, token_price=%d)", groupRes.ID, newPrice)
	return nil
}

func loadSeedUserByEmail(gdb *gorm.DB, email string, role enum.UserRole) (*entity.User, error) {
	var u entity.User
	err := gdb.Where("email = ? AND status = ?", email, enum.UserStatusActive).First(&u).Error
	if err == nil {
		return &u, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	err = gdb.Where("system_role = ? AND status = ?", role, enum.UserStatusActive).Order("id ASC").First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func resetFriendshipPair(gdb *gorm.DB, userA, userB int64) error {
	return gdb.Unscoped().
		Where("(requester_user_id = ? AND receiver_user_id = ?) OR (requester_user_id = ? AND receiver_user_id = ?)",
			userA, userB, userB, userA).
		Delete(&entity.Friendship{}).Error
}
