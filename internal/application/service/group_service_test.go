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
	svc := NewGroupService(repo, nil)

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
	svc := NewGroupService(&stubGroupRepo{}, nil)
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
