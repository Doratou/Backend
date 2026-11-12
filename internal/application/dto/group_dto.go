package dto

import "time"

type CreateGroupRequest struct {
	Title             string  `json:"title" binding:"required,min=1,max=150"`
	Description       *string `json:"description" binding:"omitempty,max=5000"`
	CoverAttachmentID *int64  `json:"cover_attachment_id"`
}

type UpdateGroupRequest struct {
	Title             *string `json:"title" binding:"omitempty,min=1,max=150"`
	Description       *string `json:"description" binding:"omitempty,max=5000"`
	CoverAttachmentID *int64  `json:"cover_attachment_id"`
}

type UpdateGroupCoverRequest struct {
	CoverAttachmentID int64 `json:"cover_attachment_id" binding:"required,gt=0"`
}

type CloneGroupRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=150"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
}

type GroupResponse struct {
	ID                int64                 `json:"id"`
	Title             string                `json:"title"`
	Description       *string               `json:"description,omitempty"`
	CreatedByUserID   int64                 `json:"created_by_user_id"`
	CoverAttachmentID *int64                `json:"cover_attachment_id,omitempty"`
	Status            string                `json:"status"`
	IsPremium         bool                  `json:"is_premium"`
	ArchivedAt        *time.Time            `json:"archived_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	MemberCount       int64                 `json:"member_count,omitempty"`
	Members           []GroupMemberResponse `json:"members,omitempty"`
}

type GroupMemberResponse struct {
	MemberID int64     `json:"member_id"`
	UserID   int64     `json:"user_id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type TransferGroupOwnershipRequest struct {
	NewOwnerID int64 `json:"new_owner_id" binding:"required,gt=0"`
}
