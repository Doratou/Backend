package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

type checklistService struct {
	repo       domainrepo.ChecklistRepository
	dbTrx      database.TrxManager
	readTrx    database.TrxManager
	avatarURLs adapter.AttachmentURLResolver
}

func NewChecklistService(repo domainrepo.ChecklistRepository, dbTrx, readTrx database.TrxManager, avatarURLs adapter.AttachmentURLResolver) adapter.ChecklistService {
	return &checklistService{repo: repo, dbTrx: dbTrx, readTrx: readTrx, avatarURLs: avatarURLs}
}

func (s *checklistService) Get(ctx context.Context, userID, groupID int64, req dto.GetChecklistRequest) (*dto.ChecklistResponse, error) {
	if req.Filter == "" {
		req.Filter = enum.ChecklistFilterAll
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if !req.Filter.IsValid() || req.Page < 1 || req.Limit < 1 || req.Limit > 100 || req.Page-1 > math.MaxInt/req.Limit {
		return nil, exception.ErrBadRequest
	}

	var group *entity.Group
	var viewer *entity.GroupMember
	var items []entity.ChecklistItem
	var members []entity.GroupMember
	var counts []domainrepo.ChecklistItemCount
	err := s.withGroupAccess(ctx, userID, groupID, false, func(txCtx context.Context, authorizedGroup *entity.Group, member *entity.GroupMember) error {
		group, viewer = authorizedGroup, member
		var err error
		items, err = s.repo.ListItems(txCtx, groupID, viewer.ID, req.Filter, req.Limit, (req.Page-1)*req.Limit)
		if err != nil {
			return err
		}
		counts, err = s.repo.CountItems(txCtx, groupID)
		if err != nil {
			return err
		}
		members, err = s.repo.ListActiveMembers(txCtx, groupID)
		return err
	})
	if err != nil {
		return nil, err
	}

	profiles := make([]*entity.GroupMember, 0, len(members)+2*len(items))
	for i := range members {
		profiles = append(profiles, &members[i])
	}
	for _, item := range items {
		profiles = append(profiles, item.CreatedByMember, item.AssignedToMember)
	}
	urls, err := s.resolveAvatarURLs(ctx, userID, profiles)
	if err != nil {
		return nil, err
	}
	summary, claimedByMember := summarizeChecklist(counts, viewer.ID)
	response := &dto.ChecklistResponse{
		GroupID: groupID, ViewerMemberID: viewer.ID, IsFrozen: group.Status != enum.GroupStatusActive, Filter: req.Filter,
		Summary: summary, MemberClaims: make([]dto.ChecklistMemberClaimsResponse, 0, len(members)),
		Items:      make([]dto.ChecklistItemResponse, 0, len(items)),
		Pagination: dto.ChecklistPaginationResponse{Page: req.Page, Limit: req.Limit, Total: checklistFilterTotal(summary, req.Filter)},
	}
	for i := range members {
		member := &members[i]
		response.MemberClaims = append(response.MemberClaims, dto.ChecklistMemberClaimsResponse{
			ChecklistMemberResponse: *dto.ToChecklistMemberResponse(member, urls), ClaimedItemCount: claimedByMember[member.ID],
		})
	}
	for i := range items {
		response.Items = append(response.Items, dto.ToChecklistItemResponse(&items[i], group, viewer, urls))
	}
	return response, nil
}

func (s *checklistService) CreateItem(ctx context.Context, userID, groupID int64, req dto.CreateChecklistItemRequest) (*dto.ChecklistItemResponse, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || !utf8.ValidString(title) || strings.ContainsRune(title, '\x00') || utf8.RuneCountInString(title) > 200 || !req.Type.IsValid() {
		return nil, exception.ErrBadRequest
	}
	var response *dto.ChecklistItemResponse
	err := s.withGroupAccess(ctx, userID, groupID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		item := &entity.ChecklistItem{GroupID: group.ID, Title: title, Type: req.Type, CreatedByMemberID: member.ID, CreatedByMember: member}
		if err := s.repo.CreateItem(txCtx, item); err != nil {
			return err
		}
		var err error
		response, err = s.itemResponse(txCtx, group, member, item)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *checklistService) SetClaim(ctx context.Context, userID, groupID, itemID int64, claimed bool) (*dto.ChecklistItemResponse, error) {
	var response *dto.ChecklistItemResponse
	err := s.withItemForUpdate(ctx, userID, groupID, itemID, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, item *entity.ChecklistItem) error {
		if item.Type != enum.ChecklistItemTypeGroup {
			return exception.ErrChecklistItemNotClaimable
		}
		if err := s.setClaim(txCtx, item, member, claimed); err != nil {
			return err
		}
		var err error
		response, err = s.itemResponse(txCtx, group, member, item)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *checklistService) setClaim(ctx context.Context, item *entity.ChecklistItem, member *entity.GroupMember, claimed bool) error {
	if item.AssignedToMemberID != nil && *item.AssignedToMemberID != member.ID {
		if claimed {
			return exception.ErrChecklistItemClaimed
		}
		return exception.ErrChecklistAssigneeRequired
	}
	if claimed == (item.AssignedToMemberID != nil) {
		return nil
	}
	var assigneeID *int64
	var assignee *entity.GroupMember
	if claimed {
		assigneeID, assignee = &member.ID, member
	}
	if err := s.repo.SetAssignee(ctx, item, assigneeID); err != nil {
		return err
	}
	item.AssignedToMemberID, item.AssignedToMember = assigneeID, assignee
	return nil
}

func (s *checklistService) DeleteItem(ctx context.Context, userID, groupID, itemID int64) error {
	return s.withItemForUpdate(ctx, userID, groupID, itemID, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember, item *entity.ChecklistItem) error {
		if item.AssignedToMemberID != nil && group.CreatedByUserID != member.UserID {
			return exception.ErrChecklistOwnerRequired
		}
		return s.repo.DeleteItem(txCtx, item)
	})
}

func (s *checklistService) withGroupAccess(ctx context.Context, userID, groupID int64, write bool, fn func(context.Context, *entity.Group, *entity.GroupMember) error) error {
	if groupID <= 0 || userID <= 0 {
		return exception.ErrBadRequest
	}
	trx := s.readTrx
	if write {
		trx = s.dbTrx
	}
	return trx.WithTransaction(ctx, func(txCtx context.Context) error {
		group, err := s.repo.FindGroup(txCtx, groupID, write)
		if err != nil {
			return err
		}
		member, err := s.repo.FindActiveMember(txCtx, groupID, userID, write)
		if err != nil {
			return err
		}
		if member.User == nil {
			return exception.ErrUserNotFound
		}
		if member.User.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		if write && group.Status != enum.GroupStatusActive {
			return exception.ErrChecklistGroupInactive
		}
		return fn(txCtx, group, member)
	})
}

func (s *checklistService) withItemForUpdate(ctx context.Context, userID, groupID, itemID int64, fn func(context.Context, *entity.Group, *entity.GroupMember, *entity.ChecklistItem) error) error {
	if itemID <= 0 {
		return exception.ErrBadRequest
	}
	return s.withGroupAccess(ctx, userID, groupID, true, func(txCtx context.Context, group *entity.Group, member *entity.GroupMember) error {
		item, err := s.repo.FindItemForUpdate(txCtx, groupID, itemID)
		if err != nil {
			return err
		}
		return fn(txCtx, group, member, item)
	})
}

func (s *checklistService) itemResponse(ctx context.Context, group *entity.Group, viewer *entity.GroupMember, item *entity.ChecklistItem) (*dto.ChecklistItemResponse, error) {
	urls, err := s.resolveAvatarURLs(ctx, viewer.UserID, []*entity.GroupMember{item.CreatedByMember, item.AssignedToMember})
	if err != nil {
		return nil, err
	}
	response := dto.ToChecklistItemResponse(item, group, viewer, urls)
	return &response, nil
}

func (s *checklistService) resolveAvatarURLs(ctx context.Context, viewerUserID int64, members []*entity.GroupMember) (map[int64]string, error) {
	ids := make([]int64, 0, len(members))
	seen := make(map[int64]struct{}, len(members))
	for _, member := range members {
		if member == nil || member.User == nil || member.User.Status != enum.UserStatusActive || member.User.AvatarAttachmentID == nil {
			continue
		}
		id := *member.User.AvatarAttachmentID
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	urls, err := s.avatarURLs.ResolveAccessURLs(ctx, viewerUserID, ids)
	var appErr *exception.AppError
	if errors.As(err, &appErr) && appErr.Code == exception.ErrStorageUnavailable.Code {
		return nil, nil
	}
	return urls, err
}

func summarizeChecklist(counts []domainrepo.ChecklistItemCount, viewerID int64) (dto.ChecklistSummaryResponse, map[int64]int64) {
	var summary dto.ChecklistSummaryResponse
	claimedByMember := make(map[int64]int64)
	for _, count := range counts {
		switch {
		case count.Type == enum.ChecklistItemTypePersonal:
			summary.PersonalItemCount += count.ItemCount
		case count.AssignedToMemberID == nil:
			summary.UnclaimedItemCount += count.ItemCount
		default:
			summary.ClaimedItemCount += count.ItemCount
			claimedByMember[*count.AssignedToMemberID] += count.ItemCount
		}
	}
	summary.GroupItemCount = summary.ClaimedItemCount + summary.UnclaimedItemCount
	summary.TotalItems = summary.GroupItemCount + summary.PersonalItemCount
	summary.MyClaimedItemCount = claimedByMember[viewerID]
	if summary.GroupItemCount > 0 {
		summary.ProgressPercent = int(math.Round(100 * float64(summary.ClaimedItemCount) / float64(summary.GroupItemCount)))
	}
	return summary, claimedByMember
}

func checklistFilterTotal(summary dto.ChecklistSummaryResponse, filter enum.ChecklistFilter) int64 {
	switch filter {
	case enum.ChecklistFilterClaimed:
		return summary.ClaimedItemCount
	case enum.ChecklistFilterUnclaimed:
		return summary.UnclaimedItemCount
	case enum.ChecklistFilterMine:
		return summary.MyClaimedItemCount
	case enum.ChecklistFilterPersonal:
		return summary.PersonalItemCount
	default:
		return summary.TotalItems
	}
}
