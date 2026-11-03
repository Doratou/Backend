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

type groupService struct{ groups domainrepo.GroupRepository }

func NewGroupService(groups domainrepo.GroupRepository) *groupService {
	return &groupService{groups: groups}
}

func (s *groupService) Create(ctx context.Context, userID int64, req dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	group := &entity.Group{Title: req.Title, Description: req.Description, CreatedByUserID: userID,
		Status: enum.GroupStatusActive}
	// TODO: Connect the group cover upload/attachment service when it is implemented.
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
		return nil, exception.ErrBadRequest
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

func (s *groupService) ChangeCover(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	// TODO: Implement cover image changes when the attachment service is available.
	return nil
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

func (s *groupService) Clone(ctx context.Context, userID, groupID int64, req dto.CloneGroupRequest) (*dto.GroupResponse, error) {
	if _, _, err := s.groups.FindByID(ctx, groupID); err != nil {
		return nil, err
	}
	isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, exception.ErrGroupForbidden
	}
	clone := &entity.Group{Title: req.Title, Description: req.Description, CreatedByUserID: userID, Status: enum.GroupStatusActive}
	if err := s.groups.CloneWithMembers(ctx, groupID, clone); err != nil {
		return nil, err
	}
	result := toGroupResponse(clone, 1)
	return &result, nil
}

func (s *groupService) Get(ctx context.Context, userID, groupID int64) (*dto.GroupResponse, error) {
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	// Groups are only visible to active members.
	isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, exception.ErrGroupForbidden
	}
	result := toGroupResponse(group, count)
	return &result, nil
}

func toGroupResponse(group *entity.Group, count int64) dto.GroupResponse {
	return dto.GroupResponse{ID: group.ID, Title: group.Title, Description: group.Description,
		CreatedByUserID: group.CreatedByUserID, CoverAttachmentID: group.CoverAttachmentID,
		Status: string(group.Status), IsPremium: group.IsPremium, ArchivedAt: group.ArchivedAt,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt, MemberCount: count}
}
