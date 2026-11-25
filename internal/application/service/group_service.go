package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type groupService struct {
	groups         domainrepo.GroupRepository
	attachmentRepo domainrepo.AttachmentRepository
}

func NewGroupService(groups domainrepo.GroupRepository, attachmentRepo domainrepo.AttachmentRepository) *groupService {
	return &groupService{groups: groups, attachmentRepo: attachmentRepo}
}

func (s *groupService) Create(ctx context.Context, userID int64, req dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	group := &entity.Group{
		Title:             req.Title,
		Description:       req.Description,
		CreatedByUserID:   userID,
		CoverAttachmentID: req.CoverAttachmentID,
		Status:            enum.GroupStatusActive,
	}

	if req.CoverAttachmentID != nil {
		att, err := s.attachmentRepo.FindByID(ctx, *req.CoverAttachmentID)
		if err != nil {
			return nil, err
		}
		if att.Type != enum.AttachmentTypeGroupCover {
			return nil, exception.ErrInvalidFileType
		}
	}

	if err := s.groups.CreateWithOwner(ctx, group); err != nil {
		return nil, err
	}
	res := toGroupResponse(group, 1)
	return &res, nil
}

func (s *groupService) ListMine(ctx context.Context, userID int64, filter dto.ListUserGroupsFilter, limit, offset int) ([]dto.GroupResponse, int64, error) {
	repoFilter, err := parseGroupListFilter(filter)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := s.groups.FindByMember(ctx, userID, repoFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupResponse, 0, len(items))
	for i := range items {
		result = append(result, toGroupResponse(&items[i].Group, items[i].MemberCount))
	}
	return result, total, nil
}

func (s *groupService) ListAll(ctx context.Context, status *string, limit, offset int) ([]dto.GroupResponse, int64, error) {
	groupStatus, err := parseGroupStatus(status)
	if err != nil {
		return nil, 0, err
	}
	groups, total, err := s.groups.FindAll(ctx, groupStatus, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupResponse, 0, len(groups))
	for i := range groups {
		result = append(result, toGroupResponse(&groups[i], 0))
	}
	return result, total, nil
}

func parseGroupStatus(status *string) (*enum.GroupStatus, error) {
	if status == nil {
		return nil, nil
	}
	value := strings.ToUpper(*status)
	if !enum.GroupStatus(value).IsValid() {
		return nil, exception.ErrGroupInvalidStatus
	}
	parsed := enum.GroupStatus(value)
	return &parsed, nil
}

func parseGroupListFilter(filter dto.ListUserGroupsFilter) (domainrepo.GroupListFilter, error) {
	status, err := parseGroupStatus(filter.Status)
	if err != nil {
		return domainrepo.GroupListFilter{}, err
	}
	isPremium, err := parseGroupIsPremium(filter.IsPremium)
	if err != nil {
		return domainrepo.GroupListFilter{}, err
	}
	role, err := parseGroupMemberRole(filter.Role)
	if err != nil {
		return domainrepo.GroupListFilter{}, err
	}
	createdFrom, createdTo, err := parseGroupYearRange(filter.Year)
	if err != nil {
		return domainrepo.GroupListFilter{}, err
	}
	sortBy, err := parseGroupSortBy(filter.SortBy)
	if err != nil {
		return domainrepo.GroupListFilter{}, err
	}
	return domainrepo.GroupListFilter{
		Status:      status,
		IsPremium:   isPremium,
		Role:        role,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		SortBy:      sortBy,
	}, nil
}

func parseGroupIsPremium(raw *string) (*bool, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(strings.ToLower(*raw))
	if err != nil {
		return nil, exception.ErrGroupInvalidIsPremium
	}
	return &parsed, nil
}

func parseGroupMemberRole(raw *string) (*enum.GroupMemberRole, error) {
	if raw == nil {
		return nil, nil
	}
	value := enum.GroupMemberRole(strings.ToUpper(*raw))
	if !value.IsValid() {
		return nil, exception.ErrGroupInvalidRole
	}
	return &value, nil
}

func parseGroupSortBy(raw *string) (enum.GroupSortBy, error) {
	if raw == nil {
		return enum.GroupSortByNewest, nil
	}
	value := enum.GroupSortBy(strings.ToUpper(*raw))
	if !value.IsValid() {
		return "", exception.ErrGroupInvalidSortBy
	}
	return value, nil
}

func parseGroupYearRange(raw *string) (*time.Time, *time.Time, error) {
	if raw == nil {
		return nil, nil, nil
	}
	year, err := strconv.Atoi(*raw)
	if err != nil {
		return nil, nil, exception.ErrGroupInvalidYear
	}

	if year >= 1300 && year <= 1599 {
		loc, locErr := time.LoadLocation("Asia/Tehran")
		if locErr != nil {
			loc = time.FixedZone("IRST", 3*3600+30*60)
		}
		sy, sm, sd := jalaliToGregorian(year, 1, 1)
		ey, em, ed := jalaliToGregorian(year+1, 1, 1)
		start := time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, loc).UTC()
		end := time.Date(ey, time.Month(em), ed, 0, 0, 0, 0, loc).UTC()
		return &start, &end, nil
	}

	if year >= 1900 && year <= 2200 {
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		return &start, &end, nil
	}

	return nil, nil, exception.ErrGroupInvalidYear
}

// jalaliToGregorian converts a Jalali (Solar Hijri) date to Gregorian (year, month, day).
func jalaliToGregorian(jy, jm, jd int) (int, int, int) {
	jy += 1595
	days := -355668 + (365 * jy) + ((jy / 33) * 8) + (((jy % 33) + 3) / 4) + jd
	if jm < 7 {
		days += (jm - 1) * 31
	} else {
		days += ((jm - 7) * 30) + 186
	}
	gy := 400 * (days / 146097)
	days %= 146097
	if days > 36524 {
		days--
		gy += 100 * (days / 36524)
		days %= 36524
		if days >= 365 {
			days++
		}
	}
	gy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		gy += (days - 1) / 365
		days = (days - 1) % 365
	}
	gd := days + 1
	salA := [13]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if (gy%4 == 0 && gy%100 != 0) || (gy%400 == 0) {
		salA[2] = 29
	}
	var gm int
	for gm = 1; gm <= 12 && gd > salA[gm]; gm++ {
		gd -= salA[gm]
	}
	return gy, gm, gd
}

func (s *groupService) ownedGroup(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*entity.Group, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return nil, exception.ErrGroupForbidden
	}
	return group, nil
}

func (s *groupService) Update(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	if req.Title == nil && req.Description == nil && req.CoverAttachmentID == nil {
		return nil, exception.ErrGroupNoFieldsToUpdate
	}
	group, err := s.ownedGroup(ctx, userID, role, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status == enum.GroupStatusClosed {
		return nil, exception.ErrGroupClosed
	}
	if req.Title != nil {
		group.Title = *req.Title
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	if req.CoverAttachmentID != nil {
		att, err := s.attachmentRepo.FindByID(ctx, *req.CoverAttachmentID)
		if err != nil {
			return nil, err
		}
		if att.Type != enum.AttachmentTypeGroupCover {
			return nil, exception.ErrInvalidFileType
		}
		group.CoverAttachmentID = req.CoverAttachmentID
	}
	if err := s.groups.Update(ctx, group); err != nil {
		return nil, err
	}
	updated, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	result := toGroupResponse(updated, count)
	return &result, nil
}

func (s *groupService) ChangeCover(ctx context.Context, userID int64, role enum.UserRole, groupID int64, coverAttachmentID int64) (*dto.GroupResponse, error) {
	group, err := s.ownedGroup(ctx, userID, role, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status == enum.GroupStatusClosed {
		return nil, exception.ErrGroupClosed
	}
	att, err := s.attachmentRepo.FindByID(ctx, coverAttachmentID)
	if err != nil {
		return nil, err
	}
	if att.Type != enum.AttachmentTypeGroupCover {
		return nil, exception.ErrInvalidFileType
	}

	group.CoverAttachmentID = &coverAttachmentID
	if err := s.groups.UpdateCover(ctx, group); err != nil {
		return nil, err
	}

	updated, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	result := toGroupResponse(updated, count)
	return &result, nil
}

func (s *groupService) Archive(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	return s.groups.SetStatus(ctx, groupID, enum.GroupStatusSettling)
}

func (s *groupService) Delete(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	return s.groups.Delete(ctx, groupID)
}

func (s *groupService) Clone(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CloneGroupRequest) (*dto.GroupResponse, error) {
	source, err := s.ownedGroup(ctx, userID, role, groupID)
	if err != nil {
		return nil, err
	}
	if source.Status == enum.GroupStatusActive {
		return nil, exception.ErrGroupMustBeInactiveToClone
	}
	clone := &entity.Group{Title: req.Title, Description: req.Description, CreatedByUserID: userID, Status: enum.GroupStatusActive}
	if err := s.groups.CloneWithMembers(ctx, groupID, clone); err != nil {
		return nil, err
	}
	result := toGroupResponse(clone, 1)
	return &result, nil
}

func (s *groupService) Get(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*dto.GroupResponse, error) {
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeGroupAccess(ctx, userID, role, groupID); err != nil {
		return nil, err
	}
	return s.buildGroupDetailResponse(ctx, group, count)
}

func (s *groupService) ListMembers(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupMemberResponse, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeGroupAccess(ctx, userID, role, groupID); err != nil {
		return nil, err
	}
	members, err := s.groups.ListActiveMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return toGroupMemberResponses(members, group.CreatedByUserID), nil
}

func (s *groupService) RemoveMember(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	member, err := s.groups.FindActiveMemberByID(ctx, groupID, memberID)
	if err != nil {
		return err
	}
	if member.UserID == group.CreatedByUserID {
		return exception.ErrGroupOwnerCannotBeRemoved
	}
	return s.groups.SetMemberInactiveByID(ctx, groupID, memberID)
}

func (s *groupService) LeaveGroup(ctx context.Context, userID, groupID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatedByUserID == userID {
		return exception.ErrGroupOwnerCannotLeave
	}
	return s.groups.SetMemberInactive(ctx, groupID, userID)
}

func (s *groupService) IsActiveMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return s.groups.IsActiveMember(ctx, groupID, userID)
}

func (s *groupService) TransferOwnership(ctx context.Context, userID int64, role enum.UserRole, groupID, newOwnerID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	return s.groups.TransferOwnership(ctx, groupID, newOwnerID)
}

func (s *groupService) authorizeGroupAccess(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if role == enum.UserRoleAdmin {
		return nil
	}
	isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return exception.ErrGroupForbidden
	}
	return nil
}

// =================================================================
// RESPONSE BUILDERS & MAPPERS
// =================================================================
func (s *groupService) buildGroupDetailResponse(ctx context.Context, group *entity.Group, count int64) (*dto.GroupResponse, error) {
	result := toGroupResponse(group, count)
	members, err := s.groups.ListActiveMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	result.Members = toGroupMemberResponses(members, group.CreatedByUserID)
	return &result, nil
}

func toGroupMemberResponses(members []entity.GroupMember, ownerID int64) []dto.GroupMemberResponse {
	responses := make([]dto.GroupMemberResponse, 0, len(members))
	for _, member := range members {
		name := ""
		if member.User != nil {
			name = member.User.Name
		}
		role := enum.GroupMemberRoleMember
		if member.UserID == ownerID {
			role = enum.GroupMemberRoleOwner
		}
		responses = append(responses, dto.GroupMemberResponse{
			MemberID: member.ID, UserID: member.UserID, Name: name, Role: string(role), JoinedAt: member.JoinedAt,
		})
	}
	return responses
}

func toGroupResponse(group *entity.Group, count int64) dto.GroupResponse {
	return dto.GroupResponse{ID: group.ID, Title: group.Title, Description: group.Description,
		CreatedByUserID: group.CreatedByUserID, CoverAttachmentID: group.CoverAttachmentID,
		Status: string(group.Status), IsPremium: group.IsPremium, ArchivedAt: group.ArchivedAt, ClosedAt: group.ClosedAt,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt, MemberCount: count}
}

func (s *groupService) ChangeStatus(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.ChangeGroupStatusRequest) (*dto.GroupResponse, error) {
	if !req.Status.IsValid() {
		return nil, exception.ErrGroupInvalidStatus
	}
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return nil, err
	}
	if err := s.groups.SetStatus(ctx, groupID, req.Status); err != nil {
		return nil, err
	}
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	res := toGroupResponse(group, count)
	return &res, nil
}
