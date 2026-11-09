package service

import (
	"context"
	"strings"

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

func (s *groupService) ListMine(ctx context.Context, userID int64, status *string, limit, offset int) ([]dto.GroupResponse, int64, error) {
	groupStatus, err := parseGroupStatus(status)
	if err != nil {
		return nil, 0, err
	}
	groups, total, err := s.groups.FindByMember(ctx, userID, groupStatus, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupResponse, 0, len(groups))
	for i := range groups {
		result = append(result, toGroupResponse(&groups[i], 0))
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
	if value != string(enum.GroupStatusActive) && value != string(enum.GroupStatusInactive) {
		return nil, exception.ErrGroupInvalidStatus
	}
	parsed := enum.GroupStatus(value)
	return &parsed, nil
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

	att, err := s.attachmentRepo.FindByID(ctx, coverAttachmentID)
	if err != nil {
		return nil, err
	}
	if att.Type != enum.AttachmentTypeGroupCover {
		return nil, exception.ErrInvalidFileType
	}

	group.CoverAttachmentID = &coverAttachmentID
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

func (s *groupService) Archive(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	return s.groups.SetStatus(ctx, groupID, enum.GroupStatusInactive)
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
	if source.Status != enum.GroupStatusInactive {
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
		Status: string(group.Status), IsPremium: group.IsPremium, ArchivedAt: group.ArchivedAt,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt, MemberCount: count}
}
