package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type GetChecklistRequest struct {
	Filter enum.ChecklistFilter `form:"filter,default=all" binding:"oneof=all claimed unclaimed mine personal"`
	Page   int                  `form:"page,default=1" binding:"min=1"`
	Limit  int                  `form:"limit,default=20" binding:"min=1,max=100"`
}

type CreateChecklistItemRequest struct {
	Title string                 `json:"title" binding:"required"`
	Type  enum.ChecklistItemType `json:"type" binding:"required,oneof=GROUP PERSONAL"`
}

type ChecklistMemberResponse struct {
	MemberID  int64   `json:"member_id"`
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type ChecklistMemberClaimsResponse struct {
	ChecklistMemberResponse
	ClaimedItemCount int64 `json:"claimed_item_count"`
}

type ChecklistItemResponse struct {
	ID         int64                    `json:"id"`
	Title      string                   `json:"title"`
	Type       enum.ChecklistItemType   `json:"type"`
	CreatedBy  *ChecklistMemberResponse `json:"created_by"`
	AssignedTo *ChecklistMemberResponse `json:"assigned_to"`
	CanClaim   bool                     `json:"can_claim"`
	CanRelease bool                     `json:"can_release"`
	CanDelete  bool                     `json:"can_delete"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

type ChecklistSummaryResponse struct {
	TotalItems         int64 `json:"total_items"`
	GroupItemCount     int64 `json:"group_item_count"`
	PersonalItemCount  int64 `json:"personal_item_count"`
	ClaimedItemCount   int64 `json:"claimed_item_count"`
	UnclaimedItemCount int64 `json:"unclaimed_item_count"`
	MyClaimedItemCount int64 `json:"my_claimed_item_count"`
	ProgressPercent    int   `json:"progress_percent"`
}

type ChecklistPaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type ChecklistResponse struct {
	GroupID        int64                           `json:"group_id"`
	ViewerMemberID int64                           `json:"viewer_member_id"`
	IsFrozen       bool                            `json:"is_frozen"`
	Filter         enum.ChecklistFilter            `json:"filter"`
	Summary        ChecklistSummaryResponse        `json:"summary"`
	MemberClaims   []ChecklistMemberClaimsResponse `json:"member_claims"`
	Items          []ChecklistItemResponse         `json:"items"`
	Pagination     ChecklistPaginationResponse     `json:"pagination"`
}

func ToChecklistMemberResponse(member *entity.GroupMember, avatarURLs map[int64]string) *ChecklistMemberResponse {
	if member == nil {
		return nil
	}
	response := &ChecklistMemberResponse{MemberID: member.ID, UserID: member.UserID}
	if member.User != nil {
		response.Name = member.User.Name
		if member.User.Status == enum.UserStatusActive && member.User.AvatarAttachmentID != nil {
			if url := avatarURLs[*member.User.AvatarAttachmentID]; url != "" {
				response.AvatarURL = &url
			}
		}
	}
	return response
}

func ToChecklistItemResponse(item *entity.ChecklistItem, group *entity.Group, viewer *entity.GroupMember, avatarURLs map[int64]string) ChecklistItemResponse {
	active := group.Status == enum.GroupStatusActive
	claimed := item.AssignedToMemberID != nil
	return ChecklistItemResponse{
		ID:         item.ID,
		Title:      item.Title,
		Type:       item.Type,
		CreatedBy:  ToChecklistMemberResponse(item.CreatedByMember, avatarURLs),
		AssignedTo: ToChecklistMemberResponse(item.AssignedToMember, avatarURLs),
		CanClaim:   active && item.Type == enum.ChecklistItemTypeGroup && !claimed,
		CanRelease: active && claimed && *item.AssignedToMemberID == viewer.ID,
		CanDelete:  active && (!claimed || group.CreatedByUserID == viewer.UserID),
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}
