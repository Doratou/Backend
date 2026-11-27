package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type stubGroupRepo struct {
	domainrepo.GroupRepository
	gotUserID int64
	gotFilter domainrepo.GroupListFilter
	gotLimit  int
	gotOffset int
	items     []domainrepo.GroupWithMemberCount
	total     int64
}

func (s *stubGroupRepo) FindByMember(_ context.Context, userID int64, filter domainrepo.GroupListFilter, limit, offset int) ([]domainrepo.GroupWithMemberCount, int64, error) {
	s.gotUserID = userID
	s.gotFilter = filter
	s.gotLimit = limit
	s.gotOffset = offset
	return s.items, s.total, nil
}

func strPtr(v string) *string { return &v }

func TestJalaliToGregorian(t *testing.T) {
	cases := []struct {
		jy, jm, jd int
		gy, gm, gd int
	}{
		{1403, 1, 1, 2024, 3, 20},
		{1404, 1, 1, 2025, 3, 21},
		{1405, 1, 1, 2026, 3, 21},
		{1406, 1, 1, 2027, 3, 21},
	}
	for _, tc := range cases {
		gy, gm, gd := jalaliToGregorian(tc.jy, tc.jm, tc.jd)
		if gy != tc.gy || gm != tc.gm || gd != tc.gd {
			t.Errorf("jalaliToGregorian(%d,%d,%d) = (%d,%d,%d), want (%d,%d,%d)",
				tc.jy, tc.jm, tc.jd, gy, gm, gd, tc.gy, tc.gm, tc.gd)
		}
	}
}

func TestListMineFiltersAndSorting(t *testing.T) {
	repo := &stubGroupRepo{
		items: []domainrepo.GroupWithMemberCount{
			{
				Group: entity.Group{
					BaseEntity:      entity.BaseEntity{ID: 10, CreatedAt: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
					Title:           "Shiraz Trip",
					CreatedByUserID: 7,
					Status:          enum.GroupStatusSettling,
					IsPremium:       true,
				},
				MemberCount: 4,
			},
		},
		total: 1,
	}
	svc := NewGroupService(repo, nil, nil)

	res, total, err := svc.ListMine(context.Background(), 7, dto.ListUserGroupsFilter{
		Status:    strPtr("settling"),
		IsPremium: strPtr("true"),
		Role:      strPtr("owner"),
		Year:      strPtr("1404"),
		SortBy:    strPtr("most_members"),
	}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(res) != 1 || res[0].MemberCount != 4 {
		t.Fatalf("unexpected result: total=%d res=%+v", total, res)
	}
	if repo.gotFilter.Status == nil || *repo.gotFilter.Status != enum.GroupStatusSettling {
		t.Errorf("expected status SETTLING, got %v", repo.gotFilter.Status)
	}
	if repo.gotFilter.IsPremium == nil || !*repo.gotFilter.IsPremium {
		t.Errorf("expected is_premium true, got %v", repo.gotFilter.IsPremium)
	}
	if repo.gotFilter.Role == nil || *repo.gotFilter.Role != enum.GroupMemberRoleOwner {
		t.Errorf("expected role OWNER, got %v", repo.gotFilter.Role)
	}
	if repo.gotFilter.SortBy != enum.GroupSortByMostMembers {
		t.Errorf("expected sort_by MOST_MEMBERS, got %v", repo.gotFilter.SortBy)
	}
	if repo.gotFilter.CreatedFrom == nil || repo.gotFilter.CreatedTo == nil {
		t.Fatalf("expected year bounds to be set")
	}
	wantFrom := time.Date(2025, 3, 20, 20, 30, 0, 0, time.UTC) // 1404-01-01 00:00:00 +03:30
	wantTo := time.Date(2026, 3, 20, 20, 30, 0, 0, time.UTC)   // 1405-01-01 00:00:00 +03:30
	if !repo.gotFilter.CreatedFrom.Equal(wantFrom) || !repo.gotFilter.CreatedTo.Equal(wantTo) {
		t.Errorf("got range [%v, %v), want [%v, %v)", repo.gotFilter.CreatedFrom, repo.gotFilter.CreatedTo, wantFrom, wantTo)
	}
}

func TestListMineInvalidFilters(t *testing.T) {
	svc := NewGroupService(&stubGroupRepo{}, nil, nil)
	cases := []struct {
		name   string
		filter dto.ListUserGroupsFilter
		want   error
	}{
		{"invalid status", dto.ListUserGroupsFilter{Status: strPtr("UNKNOWN")}, exception.ErrGroupInvalidStatus},
		{"invalid is_premium", dto.ListUserGroupsFilter{IsPremium: strPtr("yes")}, exception.ErrGroupInvalidIsPremium},
		{"invalid role", dto.ListUserGroupsFilter{Role: strPtr("ADMIN")}, exception.ErrGroupInvalidRole},
		{"invalid year non-numeric", dto.ListUserGroupsFilter{Year: strPtr("abc")}, exception.ErrGroupInvalidYear},
		{"invalid year out of range", dto.ListUserGroupsFilter{Year: strPtr("1700")}, exception.ErrGroupInvalidYear},
		{"invalid sort_by", dto.ListUserGroupsFilter{SortBy: strPtr("RANDOM")}, exception.ErrGroupInvalidSortBy},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.ListMine(context.Background(), 1, tc.filter, 20, 0)
			if !errors.Is(err, tc.want) {
				t.Errorf("got err=%v, want %v", err, tc.want)
			}
		})
	}
}

type stubGroupMembersRepo struct {
	domainrepo.GroupRepository
	group   *entity.Group
	members []entity.GroupMember
}

func (s *stubGroupMembersRepo) FindByID(_ context.Context, id int64) (*entity.Group, int64, error) {
	if s.group != nil && s.group.ID == id {
		return s.group, int64(len(s.members)), nil
	}
	return nil, 0, exception.ErrGroupNotFound
}

func (s *stubGroupMembersRepo) IsActiveMember(_ context.Context, groupID, userID int64) (bool, error) {
	for _, m := range s.members {
		if m.GroupID == groupID && m.UserID == userID && m.Status == enum.GroupMemberStatusActive {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubGroupMembersRepo) ListActiveMembers(_ context.Context, groupID int64) ([]entity.GroupMember, error) {
	var active []entity.GroupMember
	for _, m := range s.members {
		if m.GroupID == groupID && m.Status == enum.GroupMemberStatusActive {
			active = append(active, m)
		}
	}
	return active, nil
}

type stubAttachmentResolver struct {
	urls map[int64]string
	err  error
}

func (s *stubAttachmentResolver) ResolveAccessURLs(_ context.Context, _ int64, ids []int64) (map[int64]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	res := make(map[int64]string)
	for _, id := range ids {
		if url, ok := s.urls[id]; ok {
			res[id] = url
		}
	}
	return res, nil
}

func TestGroupMembersAvatarResolution(t *testing.T) {
	avatarID := int64(101)
	avatarURL := "https://storage.example.com/attachments/profile_image/1/avatar.jpg"

	group := &entity.Group{
		BaseEntity:      entity.BaseEntity{ID: 10},
		Title:           "Test Group",
		CreatedByUserID: 1,
		Status:          enum.GroupStatusActive,
	}

	members := []entity.GroupMember{
		{
			ID:       1,
			GroupID:  10,
			UserID:   1,
			Status:   enum.GroupMemberStatusActive,
			JoinedAt: time.Now(),
			User: &entity.User{
				Name:               "Owner User",
				AvatarAttachmentID: &avatarID,
			},
		},
		{
			ID:       2,
			GroupID:  10,
			UserID:   2,
			Status:   enum.GroupMemberStatusActive,
			JoinedAt: time.Now(),
			User: &entity.User{
				Name:               "Regular Member",
				AvatarAttachmentID: nil,
			},
		},
	}

	repo := &stubGroupMembersRepo{group: group, members: members}
	resolver := &stubAttachmentResolver{
		urls: map[int64]string{avatarID: avatarURL},
	}

	svc := NewGroupService(repo, nil, resolver)

	t.Run("ListMembers returns avatars", func(t *testing.T) {
		res, err := svc.ListMembers(context.Background(), 1, enum.UserRoleUser, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 2 {
			t.Fatalf("expected 2 members, got %d", len(res))
		}

		// Member 1 (Owner with avatar)
		if res[0].AvatarAttachmentID == nil || *res[0].AvatarAttachmentID != avatarID {
			t.Errorf("member 1 avatar_attachment_id = %v, want %d", res[0].AvatarAttachmentID, avatarID)
		}
		if res[0].AvatarURL == nil || *res[0].AvatarURL != avatarURL {
			t.Errorf("member 1 avatar_url = %v, want %s", res[0].AvatarURL, avatarURL)
		}
		if res[0].Role != string(enum.GroupMemberRoleOwner) {
			t.Errorf("member 1 role = %s, want OWNER", res[0].Role)
		}

		// Member 2 (Member without avatar)
		if res[1].AvatarAttachmentID != nil {
			t.Errorf("member 2 avatar_attachment_id = %v, want nil", res[1].AvatarAttachmentID)
		}
		if res[1].AvatarURL != nil {
			t.Errorf("member 2 avatar_url = %v, want nil", res[1].AvatarURL)
		}
		if res[1].Role != string(enum.GroupMemberRoleMember) {
			t.Errorf("member 2 role = %s, want MEMBER", res[1].Role)
		}
	})

	t.Run("Get returns group with member avatars", func(t *testing.T) {
		res, err := svc.Get(context.Background(), 2, enum.UserRoleUser, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Members) != 2 {
			t.Fatalf("expected 2 members, got %d", len(res.Members))
		}
		if res.Members[0].AvatarURL == nil || *res.Members[0].AvatarURL != avatarURL {
			t.Errorf("member 0 avatar_url = %v, want %s", res.Members[0].AvatarURL, avatarURL)
		}
		if res.Members[1].AvatarURL != nil {
			t.Errorf("member 1 avatar_url = %v, want nil", res.Members[1].AvatarURL)
		}
	})

	t.Run("Storage unavailable gracefully leaves avatar_url nil", func(t *testing.T) {
		failingResolver := &stubAttachmentResolver{
			err: exception.ErrStorageUnavailable,
		}
		svcStorageErr := NewGroupService(repo, nil, failingResolver)
		res, err := svcStorageErr.ListMembers(context.Background(), 1, enum.UserRoleUser, 10)
		if err != nil {
			t.Fatalf("expected no error when storage is unavailable, got %v", err)
		}
		if res[0].AvatarAttachmentID == nil || *res[0].AvatarAttachmentID != avatarID {
			t.Errorf("expected avatar_attachment_id to still be present")
		}
		if res[0].AvatarURL != nil {
			t.Errorf("expected avatar_url to be nil when storage is unavailable, got %v", res[0].AvatarURL)
		}
	})
}
